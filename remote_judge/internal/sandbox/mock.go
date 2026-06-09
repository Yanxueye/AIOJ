package sandbox

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// MockSandbox 提供无需 Docker 的测试沙箱。
type MockSandbox struct{}

// Compile 模拟编译阶段。
func (m *MockSandbox) Compile(_ context.Context, req ExecRequest) (ExecResult, error) {
	if containsCompileError(req.WorkDir) {
		return ExecResult{
			ExitCode: 1,
			Stderr:   "mock compile error",
			Runtime:  20 * time.Millisecond,
		}, nil
	}
	return ExecResult{ExitCode: 0, Runtime: 20 * time.Millisecond}, nil
}

// Run 模拟运行阶段。
func (m *MockSandbox) Run(_ context.Context, req ExecRequest) (ExecResult, error) {
	stdin := req.Stdin
	switch {
	case strings.Contains(req.WorkDir, "runtime_error"):
		return ExecResult{ExitCode: 1, Stderr: "mock runtime error", Runtime: 10 * time.Millisecond}, nil
	case strings.Contains(req.WorkDir, "tle"):
		return ExecResult{ExitCode: 124, Runtime: req.TimeLimit + 10*time.Millisecond, TimedOut: true}, nil
	case strings.Contains(req.WorkDir, "mle"):
		return ExecResult{ExitCode: 137, Runtime: 10 * time.Millisecond, MemoryKB: req.MemoryLimitMB*1024 + 1}, nil
	default:
		output := simulateProgram(stdin)
		return ExecResult{
			ExitCode:    0,
			Stdout:      output,
			StdoutBytes: len(output),
			Runtime:     10 * time.Millisecond,
			MemoryKB:    1024,
		}, nil
	}
}

// Health 返回 mock 沙箱健康状态。
func (m *MockSandbox) Health(context.Context) error {
	return nil
}

// simulateProgram 基于输入模拟简单程序输出。
func simulateProgram(stdin string) string {
	fields := strings.Fields(stdin)
	if len(fields) == 2 {
		if fields[0] == "1" && fields[1] == "2" {
			return "3\n"
		}
		if fields[0] == "10" && fields[1] == "20" {
			return "30\n"
		}
	}
	if stdin == "hello\n" {
		return "hello\n"
	}
	return stdin
}

// containsCompileError 通过源码内容判断是否模拟编译失败。
func containsCompileError(workDir string) bool {
	entries, err := os.ReadDir(workDir)
	if err != nil {
		return false
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		data, err := os.ReadFile(filepath.Join(workDir, entry.Name()))
		if err == nil && strings.Contains(string(data), "compile_error") {
			return true
		}
	}
	return false
}
