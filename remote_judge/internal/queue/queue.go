package queue

import (
	"context"

	"remote_judge/internal/domain"
)

// Queue 定义提交消息队列能力。
type Queue interface {
	Publish(ctx context.Context, msg domain.SubmissionMessage) error
	Consume(ctx context.Context) (<-chan domain.SubmissionMessage, error)
	Close() error
}
