package sandbox

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// PooledSandbox 使用可复用容器池实现 Sandbox 接口。
// 池耗尽时自动回退到内部 DockerCLISandbox。
type PooledSandbox struct {
	inner *DockerCLISandbox
	pool  *ContainerPool
}

// NewPooledSandbox 创建一个容器池支撑的沙箱。
func NewPooledSandbox(inner *DockerCLISandbox, maxSize int) *PooledSandbox {
	return &PooledSandbox{
		inner: inner,
		pool:  NewContainerPool(maxSize),
	}
}

// Compile 通过内部沙箱在全新容器中执行编译命令。
// 编译不使用池化容器，确保后续运行步骤的内存测量不受编译峰值污染。

func (p *PooledSandbox) Compile(ctx context.Context, req ExecRequest) (ExecResult, error) {
	return p.inner.exec(ctx, req)
}

// Run 在池化容器中执行运行命令。
func (p *PooledSandbox) Run(ctx context.Context, req ExecRequest) (ExecResult, error) {
	return p.execPooled(ctx, req)
}

// Health 通过内部沙箱检查 Docker 守护进程可达性。
func (p *PooledSandbox) Health(ctx context.Context) error {
	return p.inner.Health(ctx)
}

// Close 关闭容器池。
func (p *PooledSandbox) Close(ctx context.Context) {
	p.pool.Close(ctx)
}

func (p *PooledSandbox) execPooled(ctx context.Context, req ExecRequest) (ExecResult, error) {
	runLimit := req.TimeLimit
	if runLimit <= 0 {
		runLimit = time.Second
	}
	setupTimeout := runLimit + 10*time.Second
	if setupTimeout < 15*time.Second {
		setupTimeout = 15 * time.Second
	}
	setupCtx, cancel := context.WithTimeout(ctx, setupTimeout)
	defer cancel()

	if err := p.inner.ensureImage(setupCtx, req.Image); err != nil {
		return ExecResult{}, err
	}

	// 尝试获取池化容器。超时时回退到内部沙箱。
	acquireCtx, acquireCancel := context.WithTimeout(setupCtx, 5*time.Second)
	item, err := p.pool.Acquire(acquireCtx, req.Image)
	acquireCancel()
	if err != nil {
		// 池已耗尽——回退到直接执行。
		return p.inner.exec(ctx, req)
	}

	if err := p.prepareWorkspace(setupCtx, item.ContainerID, req); err != nil {
		p.pool.Release(context.Background(), item)
		return ExecResult{}, err
	}

	// cgroup v2 memory.peak is a read-only monotonic counter — it tracks
	// the all-time maximum since the cgroup was created and CANNOT be reset
	// by writing to it.  For pooled (long-lived) containers this means the
	// wrapWithTimeMem uses /usr/bin/time -v to capture the per-process
	// RSS peak, which is independent of the pooled container's cgroup
	// history and is always accurate.
	memFile := "/workspace/.judge_mem_" + strconv.FormatInt(time.Now().UnixNano(), 36)
	wrappedCmd := wrapWithMemPeak(req.Command, memFile)

	execCtx, execCancel := context.WithTimeout(ctx, runLimit+2*time.Second)
	defer execCancel()

	start := time.Now()
	result, execErr := p.execInContainer(execCtx, item.ContainerID, wrappedCmd, req.Stdin)
	result.Runtime = time.Since(start)

	postCtx, postCancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer postCancel()

	// Read the per-process RSS peak (already in bytes) written by the
	// awk extraction inside wrapWithTimeMem.
	if data, err := execCmdOutput(postCtx, "docker", "exec", item.ContainerID, "cat", memFile); err == nil {
		if rssBytes, err := strconv.ParseInt(strings.TrimSpace(string(data)), 10, 64); err == nil {
			result.MemoryKB = int(rssBytes / 1024)
		}
	}
	_ = execCmd(postCtx, "docker", "exec", item.ContainerID, "rm", "-f", memFile, memFile+".raw")

	// 将输出文件收集回工作区。
	if result.ExitCode == 0 && req.WorkDir != "" {
		_ = p.inner.collectOutput(postCtx, item.ContainerID, req.WorkDir)
	}

	// 清洗后放回池中。
	p.pool.Release(context.Background(), item)

	if execCtx.Err() == context.DeadlineExceeded {
		result.TimedOut = true
		result.ExitCode = 124
		return result, nil
	}
	if execErr != nil && result.ExitCode == 0 {
		return result, execErr
	}
	return result, nil
}

func (p *PooledSandbox) prepareWorkspace(ctx context.Context, containerID string, req ExecRequest) error {
	// Clear container /workspace first.  cleanWorkspace on Release also does
	// this; the clear-on-acquire is a belt-and-suspenders guard in case the
	// release-time cleanup was skipped or incomplete.
	_ = execCmd(ctx, "docker", "exec", containerID, "sh", "-lc",
		"rm -rf /workspace/* /workspace/.[!.]* 2>/dev/null; true")
	if err := execCmd(ctx, "docker", "exec", containerID, "sh", "-lc", "mkdir -p /workspace"); err != nil {
		return err
	}
	return p.transferFiles(ctx, containerID, req.WorkDir)
}

func (p *PooledSandbox) transferFiles(ctx context.Context, containerID, workDir string) error {
	if workDir == "" {
		return nil
	}
	// Copy every file from the host workDir into the container, always
	// overwriting.  Stale files left by a prior run (e.g. a compiled binary
	// that cleanWorkspace missed) must never survive into this execution.
	entries, err := os.ReadDir(workDir)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		fileContent, err := os.ReadFile(filepath.Join(workDir, name))
		if err != nil {
			continue
		}
		// Use install(1) so CAP_FOWNER-lacking containers still get 0755.
		cmd := exec.CommandContext(ctx, "docker", "exec", "-i", containerID, "sh", "-lc",
			"install -m 755 /dev/null "+shellQuote("/workspace/"+name)+" && cat > "+shellQuote("/workspace/"+name))
		cmd.Stdin = bytes.NewReader(fileContent)
		var stderr bytes.Buffer
		cmd.Stderr = &stderr
		if err := cmd.Run(); err != nil {
			firstLine := strings.SplitN(strings.TrimSpace(stderr.String()), "\n", 2)[0]
			return errors.New(firstLine)
		}
	}
	return nil
}

func execCmdOutput(ctx context.Context, name string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	return cmd.Output()
}

func (p *PooledSandbox) execInContainer(ctx context.Context, containerID string, command []string, stdin string) (ExecResult, error) {
	args := []string{"exec"}
	if stdin != "" {
		args = append(args, "-i")
	}
	args = append(args, containerID)
	args = append(args, command...)
	cmd := exec.CommandContext(ctx, "docker", args...)
	if stdin != "" {
		cmd.Stdin = strings.NewReader(stdin)
	}
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()

	exitCode := 0
	if cmd.ProcessState != nil {
		exitCode = cmd.ProcessState.ExitCode()
	}
	result := ExecResult{
		ExitCode:    exitCode,
		Stdout:      stdout.String(),
		Stderr:      stderr.String(),
		StdoutBytes: stdout.Len(),
		StderrBytes: stderr.Len(),
	}
	if exitCode >= 128 {
		if sig := signalName(exitCode - 128); sig != "" {
			result.Signal = sig
		}
	}
	if ctx.Err() == context.DeadlineExceeded {
		result.TimedOut = true
		result.ExitCode = 124
		return result, nil
	}
	if err != nil && cmd.ProcessState == nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = err.Error()
		}
		return result, errors.New(strings.SplitN(msg, "\n", 2)[0])
	}
	return result, nil
}

func signalName(code int) string {
	switch code {
	case 9:
		return "SIGKILL"
	case 11:
		return "SIGSEGV"
	case 6:
		return "SIGABRT"
	case 8:
		return "SIGFPE"
	case 4:
		return "SIGILL"
	case 31:
		return "SIGSYS"
	default:
		return ""
	}
}
