package sandbox

import (
	"context"
	"os"
	"runtime"
	"testing"
)

// TestPrepareSeccompProfile verifies embedded seccomp materialization.
func TestPrepareSeccompProfile(t *testing.T) {
	path, cleanup, err := prepareSeccompProfile(context.Background(), "embedded")
	if err != nil {
		t.Fatalf("prepareSeccompProfile() error = %v", err)
	}
	defer cleanup()
	if path == "" {
		if runtime.GOOS != "windows" {
			t.Fatal("expected seccomp profile path")
		}
		t.Log("seccomp skipped on Windows (Docker Desktop cannot access host temp path)")
		return
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected profile file to exist: %v", err)
	}
}

// TestPrepareSeccompProfileDisabled verifies disabled seccomp mode.
func TestPrepareSeccompProfileDisabled(t *testing.T) {
	path, cleanup, err := prepareSeccompProfile(context.Background(), "")
	if err != nil {
		t.Fatalf("prepareSeccompProfile() error = %v", err)
	}
	defer cleanup()
	if path != "" {
		t.Fatalf("expected empty path, got %q", path)
	}
}
