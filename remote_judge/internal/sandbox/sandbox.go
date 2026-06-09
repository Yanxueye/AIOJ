package sandbox

import (
	"context"
	"time"
)

// ExecRequest 描述一次容器内执行请求。
type ExecRequest struct {
	Language      string
	Image         string
	WorkDir       string
	Command       []string
	Stdin         string
	TimeLimit     time.Duration
	MemoryLimitMB int
	OutputLimitKB int
}

// ExecResult 表示执行结果。
type ExecResult struct {
	ExitCode    int
	Stdout      string
	Stderr      string
	Runtime     time.Duration
	MemoryKB    int
	StdoutBytes int
	StderrBytes int
	TimedOut    bool
	Signal      string
	OOMKilled   bool
}

// Sandbox 定义代码执行沙箱接口。
type Sandbox interface {
	Compile(ctx context.Context, req ExecRequest) (ExecResult, error)
	Run(ctx context.Context, req ExecRequest) (ExecResult, error)
	Health(ctx context.Context) error
}
