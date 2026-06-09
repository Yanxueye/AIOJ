package queue

import (
	"context"
	"testing"
	"time"

	"remote_judge/internal/domain"
)

// TestMemoryQueuePublishConsume verifies basic publish/consume.
func TestMemoryQueuePublishConsume(t *testing.T) {
	t.Logf(">>> MemoryQueue: publish -> consume")
	q := NewMemoryQueue(10)

	msg := domain.SubmissionMessage{SubmissionID: 1, UserID: 100, ProblemID: 1001}
	if err := q.Publish(context.Background(), msg); err != nil {
		t.Fatalf("Publish() error = %v", err)
	}

	ch, err := q.Consume(context.Background())
	if err != nil {
		t.Fatalf("Consume() error = %v", err)
	}
	select {
	case got := <-ch:
		if got.SubmissionID != 1 {
			t.Fatalf("expected submission 1, got %d", got.SubmissionID)
		}
		t.Logf("    consumed: submissionId=%d", got.SubmissionID)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("expected message within 100ms")
	}
}

// TestMemoryQueuePublishAfterClose verifies error on publish after close.
func TestMemoryQueuePublishAfterClose(t *testing.T) {
	t.Logf(">>> MemoryQueue: publish after close -> ErrQueueClosed")
	q := NewMemoryQueue(10)
	if err := q.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
	err := q.Publish(context.Background(), domain.SubmissionMessage{})
	t.Logf("    error=%v", err)
	if err != ErrQueueClosed {
		t.Fatalf("expected ErrQueueClosed, got %v", err)
	}
}

// TestMemoryQueueConsumeAfterCloseReturnsClosedChannel verifies the channel closes.
func TestMemoryQueueConsumeAfterCloseReturnsClosedChannel(t *testing.T) {
	t.Logf(">>> MemoryQueue: consume after close -> closed channel")
	q := NewMemoryQueue(10)
	q.Close()
	ch, _ := q.Consume(context.Background())
	_, ok := <-ch
	t.Logf("    channel closed: %v", !ok)
	if ok {
		t.Fatal("expected closed channel")
	}
}

// TestMemoryQueuePublishCancelledContext verifies publish respects context.
func TestMemoryQueuePublishCancelledContext(t *testing.T) {
	t.Logf(">>> MemoryQueue: cancelled context -> context error")
	q := NewMemoryQueue(1)
	_ = q.Publish(context.Background(), domain.SubmissionMessage{SubmissionID: 1})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := q.Publish(ctx, domain.SubmissionMessage{})
	t.Logf("    error=%v", err)
	if err == nil {
		t.Fatal("expected context error")
	}
}

// TestMemoryQueueDoubleCloseIsIdempotent verifies double close is safe.
func TestMemoryQueueDoubleCloseIsIdempotent(t *testing.T) {
	t.Logf(">>> MemoryQueue: double close -> no panic")
	q := NewMemoryQueue(10)
	if err := q.Close(); err != nil {
		t.Fatalf("first Close() error = %v", err)
	}
	if err := q.Close(); err != nil {
		t.Fatalf("second Close() error = %v", err)
	}
	t.Logf("    double close: ok")
}

// TestMemoryQueueZeroSizeDefaults verifies non-positive size falls back.
func TestMemoryQueueZeroSizeDefaults(t *testing.T) {
	t.Logf(">>> MemoryQueue: size=0 -> defaults to 128")
	q := NewMemoryQueue(0)
	msg := domain.SubmissionMessage{SubmissionID: 1}
	if err := q.Publish(context.Background(), msg); err != nil {
		t.Fatalf("Publish() error = %v", err)
	}
	t.Logf("    publish after zero-size init: ok")
}

// TestMemoryQueueFIFO verifies first-in-first-out ordering.
func TestMemoryQueueFIFO(t *testing.T) {
	t.Logf(">>> MemoryQueue: FIFO ordering")
	q := NewMemoryQueue(10)
	for i := 1; i <= 3; i++ {
		_ = q.Publish(context.Background(), domain.SubmissionMessage{SubmissionID: int64(i)})
	}
	ch, _ := q.Consume(context.Background())
	for i := 1; i <= 3; i++ {
		got := <-ch
		if int(got.SubmissionID) != i {
			t.Fatalf("expected %d, got %d (FIFO broken)", i, got.SubmissionID)
		}
	}
	t.Logf("    FIFO: 1->2->3 ok")
}
