package judger

import (
	"os"
	"path/filepath"
	"testing"
)

// TestWorkspacePoolAcquireRelease verifies pooled directory reuse.
func TestWorkspacePoolAcquireRelease(t *testing.T) {
	pool := NewWorkspacePool(1, 2)
	defer pool.Close()

	dir, err := pool.Acquire()
	if err != nil {
		t.Fatalf("Acquire() error = %v", err)
	}
	file := filepath.Join(dir, "main.cpp")
	if err := os.WriteFile(file, []byte("int main(){}"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	pool.Release(dir)

	reused, err := pool.Acquire()
	if err != nil {
		t.Fatalf("second Acquire() error = %v", err)
	}
	if reused != dir {
		t.Fatalf("expected reused dir %q, got %q", dir, reused)
	}
	if _, err := os.Stat(file); !os.IsNotExist(err) {
		t.Fatalf("expected workspace file to be cleared, stat err = %v", err)
	}
}
