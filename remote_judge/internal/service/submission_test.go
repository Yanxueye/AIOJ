package service

import (
	"context"
	"testing"
	"time"

	"remote_judge/internal/domain"
	"remote_judge/internal/queue"
	"remote_judge/internal/repository"
	"remote_judge/internal/stats"
)

// TestSubmissionServiceCreate verifies submission creation and queue publishing.
func TestSubmissionServiceCreate(t *testing.T) {
	t.Logf(">>> Submission: create + queue publish")
	subRepo := repository.NewInMemorySubmissionRepository()
	problemRepo := repository.NewInMemoryProblemRepository()
	q := queue.NewMemoryQueue(10)
	svc := NewSubmissionService(subRepo, problemRepo, q, stats.NewCollector())
	svc.nowFunc = func() time.Time { return time.Unix(100, 0) }

	sub, err := svc.Create(context.Background(), CreateSubmissionRequest{
		UserID:    1,
		ProblemID: 1001,
		Language:  "cpp17",
		Code:      "int main(){}",
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	t.Logf("    status=%s | id=%d | traceId=%s | queue=OK", sub.Status, sub.ID, sub.TraceID)
	if sub.Status != domain.StatusPending {
		t.Fatalf("unexpected status: %s", sub.Status)
	}
	if sub.TraceID == "" {
		t.Fatal("expected trace id")
	}

	ch, _ := q.Consume(context.Background())
	select {
	case msg := <-ch:
		if msg.SubmissionID != sub.ID {
			t.Fatalf("unexpected message submission id: %d", msg.SubmissionID)
		}
	default:
		t.Fatal("expected message published to queue")
	}
}

// TestSubmissionServiceRateLimit verifies rate limiting.
func TestSubmissionServiceRateLimit(t *testing.T) {
	t.Logf(">>> Submission: rate limit (12/50s -> 13th rejected)")
	subRepo := repository.NewInMemorySubmissionRepository()
	problemRepo := repository.NewInMemoryProblemRepository()
	q := queue.NewMemoryQueue(20)
	svc := NewSubmissionService(subRepo, problemRepo, q, stats.NewCollector())
	now := time.Now()
	svc.nowFunc = func() time.Time { return now }

	for i := 0; i < 12; i++ {
		id, _ := subRepo.NextID(context.Background())
		_ = subRepo.Create(context.Background(), &domain.Submission{
			ID:        id,
			UserID:    1,
			ProblemID: 1001,
			Language:  "cpp17",
			Code:      "x",
			CreatedAt: now.Add(-10 * time.Second),
			UpdatedAt: now.Add(-10 * time.Second),
		})
	}

	_, err := svc.Create(context.Background(), CreateSubmissionRequest{
		UserID:    1,
		ProblemID: 1001,
		Language:  "cpp17",
		Code:      "int main(){}",
	})
	t.Logf("    rate limited: %v", err == ErrRateLimited)
	if err != ErrRateLimited {
		t.Fatalf("expected ErrRateLimited, got %v", err)
	}
}

// TestSubmissionServiceRejectsNullByte verifies request validation hardening.
func TestSubmissionServiceRejectsNullByte(t *testing.T) {
	t.Logf(">>> Submission: reject null byte in code")
	subRepo := repository.NewInMemorySubmissionRepository()
	problemRepo := repository.NewInMemoryProblemRepository()
	q := queue.NewMemoryQueue(10)
	svc := NewSubmissionService(subRepo, problemRepo, q, stats.NewCollector())

	_, err := svc.Create(context.Background(), CreateSubmissionRequest{
		UserID:    1,
		ProblemID: 1001,
		Language:  "cpp17",
		Code:      "int main(){}\x00",
	})
	t.Logf("    rejected: %v", err)
	if err == nil || err == ErrRateLimited {
		t.Fatalf("expected bad request error, got %v", err)
	}
}

// BenchmarkSubmissionServiceCreate evaluates submission creation throughput.
func BenchmarkSubmissionServiceCreate(b *testing.B) {
	subRepo := repository.NewInMemorySubmissionRepository()
	problemRepo := repository.NewInMemoryProblemRepository()
	q := queue.NewMemoryQueue(4096)
	svc := NewSubmissionService(subRepo, problemRepo, q, stats.NewCollector())
	svc.nowFunc = time.Now
	ctx := context.Background()
	ch, _ := q.Consume(ctx)
	done := make(chan struct{})
	go func() {
		defer close(done)
		for range ch {
		}
	}()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := svc.Create(ctx, CreateSubmissionRequest{
			UserID:    int64((i % 1000) + 1 + ((i / 1000) * 100000)),
			ProblemID: 1001,
			Language:  "cpp17",
			Code:      "int main(){}",
		})
		if err != nil {
			b.Fatalf("Create() error = %v", err)
		}
	}
	b.StopTimer()
	_ = q.Close()
	<-done
}
