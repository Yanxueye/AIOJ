package sandbox

import (
	"bytes"
	"context"
	"errors"
	"os/exec"
	"strings"
	"sync"
	"time"
)

// PoolItem 是池中维护的一个预创建且运行中的 Docker 容器。
type PoolItem struct {
	ContainerID string
	Image       string
	CreatedAt   time.Time
	LastUsedAt  time.Time
}

// ContainerPool 管理一个按镜像分组的可复用 Docker 容器池。
type ContainerPool struct {
	mu      sync.Mutex
	pools   map[string][]*PoolItem
	maxSize int
}

// NewContainerPool 创建一个容器池，指定每镜像最大池化容器数。
func NewContainerPool(maxSize int) *ContainerPool {
	if maxSize <= 0 {
		maxSize = 4
	}
	return &ContainerPool{
		pools:   make(map[string][]*PoolItem),
		maxSize: maxSize,
	}
}

// Acquire 获取或创建一个指定镜像的容器。
// 返回容器 ID。阻塞直到有容器可用或上下文超时，
// 以先发生者为准。
func (p *ContainerPool) Acquire(ctx context.Context, image string) (*PoolItem, error) {
	p.mu.Lock()
	// 尝试从现有池中取出容器。
	if pool, ok := p.pools[image]; ok && len(pool) > 0 {
		item := pool[len(pool)-1]
		p.pools[image] = pool[:len(pool)-1]
		p.mu.Unlock()

		// 验证容器是否仍然存活。
		if err := p.healthCheck(ctx, item.ContainerID); err != nil {
			_ = p.removeContainer(ctx, item.ContainerID)
			// 创建替代容器。
			return p.createAndPrepare(ctx, image)
		}
		item.LastUsedAt = time.Now()
		return item, nil
	}
	p.mu.Unlock()

	// 无可用容器——创建新容器。
	return p.createAndPrepare(ctx, image)
}

// Release 清洗容器后放回池中；若池已满或容器已死亡则丢弃。

func (p *ContainerPool) Release(ctx context.Context, item *PoolItem) {
	if item == nil || item.ContainerID == "" {
		return
	}

	// 清理容器内的 workspace 和 /tmp。
	p.cleanWorkspace(ctx, item.ContainerID)

	// 健康检查。若已死亡则丢弃。
	if err := p.healthCheck(ctx, item.ContainerID); err != nil {
		_ = p.removeContainer(ctx, item.ContainerID)
		return
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	pool := p.pools[item.Image]
	if len(pool) >= p.maxSize {
		// 池已满——丢弃此容器。
		_ = p.removeContainer(context.Background(), item.ContainerID)
		return
	}
	p.pools[item.Image] = append(pool, item)
}

// Close 移除所有池化容器。
func (p *ContainerPool) Close(ctx context.Context) {
	p.mu.Lock()
	defer p.mu.Unlock()

	for _, pool := range p.pools {
		for _, item := range pool {
			_ = p.removeContainer(ctx, item.ContainerID)
		}
	}
	p.pools = make(map[string][]*PoolItem)
}

func (p *ContainerPool) createAndPrepare(ctx context.Context, image string) (*PoolItem, error) {
	containerID, err := p.createPoolContainer(ctx, image)
	if err != nil {
		return nil, err
	}
	if err := p.startDetached(ctx, containerID); err != nil {
		_ = p.removeContainer(context.Background(), containerID)
		return nil, err
	}
	return &PoolItem{
		ContainerID: containerID,
		Image:       image,
		CreatedAt:   time.Now(),
		LastUsedAt:  time.Now(),
	}, nil
}

func (p *ContainerPool) createPoolContainer(ctx context.Context, image string) (string, error) {
	args := []string{
		"create",
		"-i",
		"--network", "none",
		"--cpus", "1",
		"--pids-limit", "64",
		"--tmpfs", "/tmp:size=64m",
		"--tmpfs", "/workspace:size=256m,uid=1000,gid=1000,mode=0775,exec",
		"--cap-drop", "ALL",
		"--security-opt", "no-new-privileges",
		"--user", "1000:1000",
		"-w", "/workspace",
	}

	// 为 Go 镜像挂载持久 Go 构建缓存。
	if strings.Contains(image, "go") {
		goCacheDir := osTempDir() + string('/') + "remote_judge_gocache"
		_ = osMkdirAll(goCacheDir)
		args = append(args, "-v", goCacheDir+":/go-cache")
	}

	seccompPath, cleanup, err := prepareSeccompProfile(ctx, SeccompProfileMode)
	if err != nil {
		return "", err
	}
	defer cleanup()
	if seccompPath != "" {
		args = append(args, "--security-opt", "seccomp="+seccompPath)
	}

	args = append(args, image)
	args = append(args, "sh", "-lc", "sleep 3600")

	cmd := exec.CommandContext(ctx, "docker", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", errors.New(strings.TrimSpace(nonEmpty(stderr.String(), err.Error())))
	}
	return strings.TrimSpace(stdout.String()), nil
}

func (p *ContainerPool) startDetached(ctx context.Context, containerID string) error {
	cmd := exec.CommandContext(ctx, "docker", "start", containerID)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return errors.New(strings.TrimSpace(nonEmpty(stderr.String(), err.Error())))
	}
	return nil
}

// cleanWorkspace 清除所有文件、杀死残留进程并重置
// 容器内的 cgroup 内存峰值，使下次执行获得干净的基线。
func (p *ContainerPool) cleanWorkspace(ctx context.Context, containerID string) {
	// 杀死运行用户的所有残留子进程。
	execCmd(ctx, "docker", "exec", containerID, "sh", "-lc",
		"kill -9 -1 2>/dev/null; true")

	// 删除 workspace 和 tmp 中的所有文件。
	execCmd(ctx, "docker", "exec", containerID, "sh", "-lc",
		"rm -rf /workspace/* /workspace/.[!.]* /tmp/* /tmp/.[!.]* 2>/dev/null; true")

	// 重置 cgroup 内存峰值，避免前次编译的高水位
	// 污染后续运行步骤的内存测量。必须以 root (-u 0) 身份运行，
	// 因为容器用户 (1000) 无法写入 cgroup 伪文件。
	// 尽力而为：echo + stderr 重定向 + 末尾 true 确保
	// 即使内核拒绝写入，调用本身也不会失败。
	execCmd(ctx, "docker", "exec", "-u", "0", containerID, "sh", "-lc",
		"echo 0 > /sys/fs/cgroup/memory.peak 2>/dev/null; echo 0 > /sys/fs/cgroup/memory/memory.max_usage_in_bytes 2>/dev/null; true")
}

// healthCheck 验证容器是否存活且可响应。
func (p *ContainerPool) healthCheck(ctx context.Context, containerID string) error {
	return execCmd(ctx, "docker", "exec", containerID, "sh", "-lc", "echo ok")
}

func (p *ContainerPool) removeContainer(ctx context.Context, containerID string) error {
	return execCmd(ctx, "docker", "rm", "-f", containerID)
}

func execCmd(ctx context.Context, name string, args ...string) error {
	cmd := exec.CommandContext(ctx, name, args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return errors.New(strings.TrimSpace(nonEmpty(stderr.String(), err.Error())))
	}
	return nil
}

func osTempDir() string {
	return "/tmp"
}

func osMkdirAll(path string) error {
	cmd := exec.Command("sh", "-c", "mkdir -p "+shellQuote(path))
	return cmd.Run()
}

