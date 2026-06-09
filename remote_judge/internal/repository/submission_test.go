package repository

import (
	"context"
	"testing"
	"time"

	"remote_judge/internal/domain"
)

// TestInMemorySubmissionCRUD verifies create, get, and update.
func TestInMemorySubmissionCRUD(t *testing.T) {
	t.Logf(">>> SubmissionRepo: create + getByID + update")
	repo := NewInMemorySubmissionRepository()

	sub := &domain.Submission{ID: 100001, UserID: 1, ProblemID: 1001, Language: "cpp17", Status: domain.StatusPending}
	if err := repo.Create(context.Background(), sub); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	got, err := repo.GetByID(context.Background(), 100001)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if got.ID != 100001 || got.Status != domain.StatusPending {
		t.Fatalf("unexpected submission: %+v", got)
	}
	t.Logf("    create+get: id=%d status=%s", got.ID, got.Status)

	got.Status = domain.StatusAccepted
	if err := repo.Update(context.Background(), got); err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	updated, _ := repo.GetByID(context.Background(), 100001)
	if updated.Status != domain.StatusAccepted {
		t.Fatalf("expected Accepted after update, got %s", updated.Status)
	}
	t.Logf("    update: status=%s", updated.Status)
}

// TestInMemorySubmissionGetNotFound verifies ErrNotFound.
func TestInMemorySubmissionGetNotFound(t *testing.T) {
	t.Logf(">>> SubmissionRepo: get non-existent -> ErrNotFound")
	repo := NewInMemorySubmissionRepository()
	_, err := repo.GetByID(context.Background(), 999999)
	t.Logf("    error=%v", err)
	if err != ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

// TestInMemorySubmissionUpdateNotFound verifies update with non-existent ID.
func TestInMemorySubmissionUpdateNotFound(t *testing.T) {
	t.Logf(">>> SubmissionRepo: update non-existent -> ErrNotFound")
	repo := NewInMemorySubmissionRepository()
	err := repo.Update(context.Background(), &domain.Submission{ID: 999999})
	t.Logf("    error=%v", err)
	if err != ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

// TestInMemorySubmissionListFilter verifies filtering by UserID.
func TestInMemorySubmissionListFilter(t *testing.T) {
	t.Logf(">>> SubmissionRepo: list with UserID filter")
	repo := NewInMemorySubmissionRepository()
	now := time.Now()
	for i := 0; i < 5; i++ {
		_ = repo.Create(context.Background(), &domain.Submission{
			ID: int64(200001 + i), UserID: 2, ProblemID: 1001 + int64(i),
			Language: "cpp17", Status: domain.StatusAccepted, CreatedAt: now,
		})
	}
	_ = repo.Create(context.Background(), &domain.Submission{
		ID: 300001, UserID: 99, ProblemID: 1001, Language: "cpp17",
		Status: domain.StatusPending, CreatedAt: now,
	})

	items, total, err := repo.List(context.Background(), domain.SubmissionFilter{UserID: 2, Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	t.Logf("    list: userID=2 -> total=%d returned=%d", total, len(items))
	if total != 5 || len(items) != 5 {
		t.Fatalf("expected 5 items, got total=%d len=%d", total, len(items))
	}
}

// TestInMemorySubmissionListPagination verifies pagination boundaries.
func TestInMemorySubmissionListPagination(t *testing.T) {
	t.Logf(">>> SubmissionRepo: pagination")
	repo := NewInMemorySubmissionRepository()
	now := time.Now()
	for i := 0; i < 10; i++ {
		_ = repo.Create(context.Background(), &domain.Submission{
			ID: int64(400001 + i), UserID: 1, ProblemID: 1001,
			Language: "cpp17", Status: domain.StatusAccepted, CreatedAt: now,
		})
	}

	items, total, _ := repo.List(context.Background(), domain.SubmissionFilter{UserID: 1, Page: 1, PageSize: 3})
	t.Logf("    page1 size=3: total=%d returned=%d", total, len(items))
	if len(items) != 3 {
		t.Fatalf("expected 3 items on page 1, got %d", len(items))
	}

	items, _, _ = repo.List(context.Background(), domain.SubmissionFilter{UserID: 1, Page: 5, PageSize: 3})
	t.Logf("    page5 size=3 (out of 10 items): returned=%d", len(items))
	if len(items) != 0 {
		t.Fatalf("expected 0 items on page beyond total, got %d", len(items))
	}
}

// TestInMemorySubmissionSaveAndGetCases verifies case result persistence.
func TestInMemorySubmissionSaveAndGetCases(t *testing.T) {
	t.Logf(">>> SubmissionRepo: save + get case results")
	repo := NewInMemorySubmissionRepository()
	cases := []domain.SubmissionCaseResult{
		{SubmissionID: 500001, CaseNo: 1, Status: domain.StatusAccepted, RuntimeMs: 10, MemoryKB: 1024},
		{SubmissionID: 500001, CaseNo: 2, Status: domain.StatusWrongAnswer, RuntimeMs: 15, MemoryKB: 2048},
	}
	if err := repo.SaveCaseResults(context.Background(), 500001, cases); err != nil {
		t.Fatalf("SaveCaseResults() error = %v", err)
	}
	got, err := repo.GetCaseResults(context.Background(), 500001)
	if err != nil {
		t.Fatalf("GetCaseResults() error = %v", err)
	}
	t.Logf("    saved=%d got=%d", len(cases), len(got))
	if len(got) != 2 || got[0].CaseNo != 1 || got[1].Status != domain.StatusWrongAnswer {
		t.Fatalf("unexpected case results: %+v", got)
	}
}

// TestInMemorySubmissionCountRecentByUser verifies rate-limit counting.
func TestInMemorySubmissionCountRecentByUser(t *testing.T) {
	t.Logf(">>> SubmissionRepo: countRecentByUser (rate limit window)")
	repo := NewInMemorySubmissionRepository()
	now := time.Now()
	for i := 0; i < 5; i++ {
		_ = repo.Create(context.Background(), &domain.Submission{
			ID: int64(600001 + i), UserID: 7, ProblemID: 1001,
			Language: "cpp17", Status: domain.StatusPending, CreatedAt: now.Add(-30 * time.Second),
		})
	}
	_ = repo.Create(context.Background(), &domain.Submission{
		ID: 700001, UserID: 7, ProblemID: 1001, Language: "cpp17", Status: domain.StatusPending,
		CreatedAt: now.Add(-5 * time.Minute),
	})

	count, err := repo.CountRecentByUser(context.Background(), 7, now.Add(-1*time.Minute))
	if err != nil {
		t.Fatalf("CountRecentByUser() error = %v", err)
	}
	t.Logf("    recent (1min): count=%d (out of 6 total for user 7)", count)
	if count != 5 {
		t.Fatalf("expected 5 recent submissions, got %d", count)
	}
}

// TestInMemorySubmissionNextID verifies monotonic ID generation.
func TestInMemorySubmissionNextID(t *testing.T) {
	t.Logf(">>> SubmissionRepo: NextID monotonic")
	repo := NewInMemorySubmissionRepository()
	id1, _ := repo.NextID(context.Background())
	id2, _ := repo.NextID(context.Background())
	t.Logf("    id1=%d id2=%d", id1, id2)
	if id2 != id1+1 {
		t.Fatalf("expected consecutive IDs, got %d and %d", id1, id2)
	}
}
