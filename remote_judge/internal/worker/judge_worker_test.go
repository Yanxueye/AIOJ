package worker

import (
	"context"
	"testing"
	"time"

	"remote_judge/internal/domain"
	"remote_judge/internal/judger"
	"remote_judge/internal/queue"
	"remote_judge/internal/repository"
	"remote_judge/internal/sandbox"
	"remote_judge/internal/stats"
)

// TestJudgeWorkerHandle verifies a pending submission reaches a terminal state.
func TestJudgeWorkerHandle(t *testing.T) {
	t.Logf(">>> Worker: pending submission -> Accepted (mock)")
	subRepo := repository.NewInMemorySubmissionRepository()
	problemRepo := repository.NewInMemoryProblemRepository()
	q := queue.NewMemoryQueue(10)
	judgeSvc := judger.NewService(&sandbox.MockSandbox{})
	worker := NewJudgeWorker(q, subRepo, problemRepo, judgeSvc, stats.NewCollector(), 2)

	now := time.Now()
	sub := &domain.Submission{
		ID:         100001,
		UserID:     1,
		ProblemID:  1001,
		Language:   "cpp17",
		Code:       "dummy",
		CodeLength: 5,
		Status:     domain.StatusPending,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	if err := subRepo.Create(context.Background(), sub); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if err := worker.Start(context.Background()); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	if err := q.Publish(context.Background(), domain.SubmissionMessage{
		SubmissionID: sub.ID,
		UserID:       sub.UserID,
		ProblemID:    sub.ProblemID,
		Language:     sub.Language,
		CreatedAt:    now,
	}); err != nil {
		t.Fatalf("Publish() error = %v", err)
	}

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		got, err := subRepo.GetByID(context.Background(), sub.ID)
		if err == nil && domain.IsTerminalStatus(got.Status) {
			if got.Status != domain.StatusAccepted {
				t.Fatalf("unexpected terminal status: %s", got.Status)
			}
			cases, _ := subRepo.GetCaseResults(context.Background(), sub.ID)
			if len(cases) == 0 {
				t.Fatal("expected case results")
			}
				t.Logf("    status=%s | cases=%d", got.Status, len(cases))
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatal("worker did not complete in time")
}
