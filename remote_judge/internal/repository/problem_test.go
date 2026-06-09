package repository

import (
	"context"
	"testing"
)

// TestInMemoryProblemGetByID verifies problem lookup.
func TestInMemoryProblemGetByID(t *testing.T) {
	t.Logf(">>> ProblemRepo: getByID (demo problem)")
	repo := NewInMemoryProblemRepository()
	problem, err := repo.GetByID(context.Background(), 1001)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	t.Logf("    id=%d title=%q limit=%dms/%dMB", problem.ID, problem.Title, problem.TimeLimitMs, problem.MemoryLimitMB)
	if problem.Title != "A+B Problem" {
		t.Fatalf("unexpected title: %s", problem.Title)
	}
}

// TestInMemoryProblemGetByIDNotFound verifies ErrNotFound.
func TestInMemoryProblemGetByIDNotFound(t *testing.T) {
	t.Logf(">>> ProblemRepo: getByID non-existent -> ErrNotFound")
	repo := NewInMemoryProblemRepository()
	_, err := repo.GetByID(context.Background(), 1)
	t.Logf("    error=%v", err)
	if err != ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

// TestInMemoryProblemListCases verifies test case listing.
func TestInMemoryProblemListCases(t *testing.T) {
	t.Logf(">>> ProblemRepo: listCases")
	repo := NewInMemoryProblemRepository()
	cases, err := repo.ListCases(context.Background(), 1001)
	if err != nil {
		t.Fatalf("ListCases() error = %v", err)
	}
	t.Logf("    problem=%d cases=%d", 1001, len(cases))
	if len(cases) != 2 || cases[0].CaseNo != 1 || cases[1].Expected != "30\n" {
		t.Fatalf("unexpected cases: %+v", cases)
	}
}

// TestInMemoryProblemListCasesNotFound verifies test case lookup for missing problem.
func TestInMemoryProblemListCasesNotFound(t *testing.T) {
	t.Logf(">>> ProblemRepo: listCases non-existent -> ErrNotFound")
	repo := NewInMemoryProblemRepository()
	_, err := repo.ListCases(context.Background(), 1)
	t.Logf("    error=%v", err)
	if err != ErrNotFound {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}
