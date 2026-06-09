package queue

import (
	"context"
	"errors"
	"sync"

	"remote_judge/internal/domain"
)

// ErrQueueClosed 表示队列已经关闭。
var ErrQueueClosed = errors.New("queue closed")

// MemoryQueue 提供内存版队列。
type MemoryQueue struct {
	ch     chan domain.SubmissionMessage
	mu     sync.RWMutex
	closed bool
}

// NewMemoryQueue 创建一个内存队列。
func NewMemoryQueue(size int) *MemoryQueue {
	if size <= 0 {
		size = 128
	}
	return &MemoryQueue{ch: make(chan domain.SubmissionMessage, size)}
}

// Publish 发布一条提交消息。
func (q *MemoryQueue) Publish(ctx context.Context, msg domain.SubmissionMessage) error {
	q.mu.RLock()
	defer q.mu.RUnlock()
	if q.closed {
		return ErrQueueClosed
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	case q.ch <- msg:
		return nil
	}
}

// Consume 暴露消费通道。
func (q *MemoryQueue) Consume(context.Context) (<-chan domain.SubmissionMessage, error) {
	return q.ch, nil
}

// Close 关闭队列。
func (q *MemoryQueue) Close() error {
	q.mu.Lock()
	defer q.mu.Unlock()
	if q.closed {
		return nil
	}
	q.closed = true
	close(q.ch)
	return nil
}
