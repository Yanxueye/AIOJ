package judger

import (
	"context"

	"remote_judge/internal/domain"
)

// Executor 定义判题执行能力，便于嵌入式与远程模式切换。
type Executor interface {
	Judge(ctx context.Context, req domain.JudgeRequest) (domain.JudgeResult, error)
	Health(ctx context.Context) error
}
