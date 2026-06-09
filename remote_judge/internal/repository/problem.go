package repository

import (
	"context"
	"sync"

	"remote_judge/internal/domain"
)

// ProblemRepository 定义题目仓储能力。
type ProblemRepository interface {
	GetByID(ctx context.Context, id int64) (*domain.Problem, error)
	ListCases(ctx context.Context, problemID int64) ([]domain.TestCase, error)
}

// InMemoryProblemRepository 提供内存版题库。
type InMemoryProblemRepository struct {
	mu       sync.RWMutex
	problems map[int64]domain.Problem
	cases    map[int64][]domain.TestCase
}

// NewInMemoryProblemRepository 创建题目仓储并预置演示题目。
func NewInMemoryProblemRepository() *InMemoryProblemRepository {
	return &InMemoryProblemRepository{
		problems: map[int64]domain.Problem{
			1001: {ID: 1001, Title: "A+B Problem", TimeLimitMs: 1000, MemoryLimitMB: 128, OutputLimitKB: 1024},
			1002: {ID: 1002, Title: "Echo", TimeLimitMs: 1000, MemoryLimitMB: 128, OutputLimitKB: 1024},
		},
		cases: map[int64][]domain.TestCase{
			1001: {
				{ProblemID: 1001, CaseNo: 1, Input: "1 2\n", Expected: "3\n"},
				{ProblemID: 1001, CaseNo: 2, Input: "10 20\n", Expected: "30\n"},
			},
			1002: {
				{ProblemID: 1002, CaseNo: 1, Input: "hello\n", Expected: "hello\n"},
			},
		},
	}
}

// GetByID 返回指定题目。
func (r *InMemoryProblemRepository) GetByID(_ context.Context, id int64) (*domain.Problem, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	problem, ok := r.problems[id]
	if !ok {
		return nil, ErrNotFound
	}
	copyItem := problem
	return &copyItem, nil
}

// ListCases 返回指定题目的测试点。
func (r *InMemoryProblemRepository) ListCases(_ context.Context, problemID int64) ([]domain.TestCase, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	cases, ok := r.cases[problemID]
	if !ok {
		return nil, ErrNotFound
	}
	copyCases := make([]domain.TestCase, len(cases))
	copy(copyCases, cases)
	return copyCases, nil
}
