package sandbox

import (
	"context"
	_ "embed"
	"os"
	"path/filepath"
	"runtime"
	"sync"
)

var (
	cachedSeccompPath string
	seccompOnce       sync.Once
)

//go:embed seccomp_profile.json
var embeddedSeccompProfile []byte

func initSeccompPath() {
	if runtime.GOOS == "windows" {
		return
	}
	dir, err := os.MkdirTemp("", "judge_seccomp_")
	if err != nil {
		return
	}
	path := filepath.Join(dir, "judge-profile.json")
	if err := os.WriteFile(path, embeddedSeccompProfile, 0o644); err != nil {
		_ = os.RemoveAll(dir)
		return
	}
	cachedSeccompPath = path
}

// prepareSeccompProfile returns the path to a materialized seccomp profile.
// 文件在每个进程生命周期内仅写入一次，跨所有容器复用。
func prepareSeccompProfile(_ context.Context, seccompMode string) (string, func(), error) {
	if seccompMode == "" || seccompMode != "embedded" {
		return "", func() {}, nil
	}
	if runtime.GOOS == "windows" {
		return "", func() {}, nil
	}
	seccompOnce.Do(initSeccompPath)
	return cachedSeccompPath, func() {}, nil
}
