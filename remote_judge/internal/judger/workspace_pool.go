package judger

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// WorkspacePool 复用判题工作区以减少重复的目录 I/O。
type WorkspacePool struct {
	mu       sync.Mutex
	baseDirs []string
	maxSize  int
}

// NewWorkspacePool 创建预分配目录的工作区对象池。
func NewWorkspacePool(initialSize, maxSize int) *WorkspacePool {
	if initialSize < 0 {
		initialSize = 0
	}
	if maxSize <= 0 {
		maxSize = 16
	}
	if initialSize > maxSize {
		initialSize = maxSize
	}
	p := &WorkspacePool{maxSize: maxSize}
	for i := 0; i < initialSize; i++ {
		if dir, err := os.MkdirTemp("", "judge_pool_"); err == nil {
			p.baseDirs = append(p.baseDirs, dir)
		}
	}
	return p
}

// Acquire 返回一个工作区目录，池为空时新建。
func (p *WorkspacePool) Acquire() (string, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	n := len(p.baseDirs)
	if n == 0 {
		return os.MkdirTemp("", "judge_pool_")
	}
	dir := p.baseDirs[n-1]
	p.baseDirs = p.baseDirs[:n-1]
	return dir, nil
}

// Release 清空工作区目录并放回池中。
func (p *WorkspacePool) Release(path string) {
	if path == "" {
		return
	}
	_ = clearWorkspace(path)

	p.mu.Lock()
	defer p.mu.Unlock()
	if len(p.baseDirs) >= p.maxSize {
		_ = os.RemoveAll(path)
		return
	}
	p.baseDirs = append(p.baseDirs, path)
}

// Close 移除所有池化目录。
func (p *WorkspacePool) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()
	for _, dir := range p.baseDirs {
		_ = os.RemoveAll(dir)
	}
	p.baseDirs = nil
}

func clearWorkspace(path string) error {
	entries, err := os.ReadDir(path)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		full := filepath.Join(path, entry.Name())
		if entry.IsDir() {
			_ = os.RemoveAll(full)
			continue
		}
		if isWorkspaceFile(entry.Name()) {
			_ = os.Remove(full)
		}
	}
	return nil
}

func isWorkspaceFile(name string) bool {
	allowed := []string{".cpp", ".go", ".py", ".out", ".txt", ".exe", ".obj", ".o"}
	for _, ext := range allowed {
		if strings.HasSuffix(name, ext) {
			return true
		}
	}
	return false
}
