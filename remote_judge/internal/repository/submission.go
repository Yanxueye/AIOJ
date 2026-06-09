package repository

import (
	"context"
	"errors"
	"sort"
	"sync"
	"time"

	"remote_judge/internal/domain"
)

// ErrNotFound 表示记录不存在。
var ErrNotFound = errors.New("not found")

// SubmissionRepository 定义提交仓储能力。
type SubmissionRepository interface {
	Create(ctx context.Context, sub *domain.Submission) error
	Update(ctx context.Context, sub *domain.Submission) error
	GetByID(ctx context.Context, id int64) (*domain.Submission, error)
	List(ctx context.Context, filter domain.SubmissionFilter) ([]domain.Submission, int, error)
	SaveCaseResults(ctx context.Context, submissionID int64, cases []domain.SubmissionCaseResult) error
	GetCaseResults(ctx context.Context, submissionID int64) ([]domain.SubmissionCaseResult, error)
	CountRecentByUser(ctx context.Context, userID int64, since time.Time) (int, error)
	NextID(ctx context.Context) (int64, error)
}

// InMemorySubmissionRepository 提供内存版提交仓储。
type InMemorySubmissionRepository struct {
	mu       sync.RWMutex
	nextID   int64
	items    map[int64]domain.Submission
	caseData map[int64][]domain.SubmissionCaseResult
}

// NewInMemorySubmissionRepository 创建提交仓储。
func NewInMemorySubmissionRepository() *InMemorySubmissionRepository {
	return &InMemorySubmissionRepository{
		nextID:   100000,
		items:    make(map[int64]domain.Submission),
		caseData: make(map[int64][]domain.SubmissionCaseResult),
	}
}

// NextID 生成下一个提交编号。
func (r *InMemorySubmissionRepository) NextID(context.Context) (int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.nextID++
	return r.nextID, nil
}

// Create 创建一条提交记录。
func (r *InMemorySubmissionRepository) Create(_ context.Context, sub *domain.Submission) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.items[sub.ID] = *sub
	return nil
}

// Update 更新一条提交记录。
func (r *InMemorySubmissionRepository) Update(_ context.Context, sub *domain.Submission) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.items[sub.ID]; !ok {
		return ErrNotFound
	}
	sub.UpdatedAt = time.Now()
	r.items[sub.ID] = *sub
	return nil
}

// GetByID 读取指定提交。
func (r *InMemorySubmissionRepository) GetByID(_ context.Context, id int64) (*domain.Submission, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	item, ok := r.items[id]
	if !ok {
		return nil, ErrNotFound
	}
	copyItem := item
	return &copyItem, nil
}

// List 按条件列出提交记录。
func (r *InMemorySubmissionRepository) List(_ context.Context, filter domain.SubmissionFilter) ([]domain.Submission, int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	items := make([]domain.Submission, 0, len(r.items))
	for _, item := range r.items {
		if item.UserID != filter.UserID {
			continue
		}
		if filter.ProblemID > 0 && item.ProblemID != filter.ProblemID {
			continue
		}
		if filter.Status != "" && string(item.Status) != filter.Status {
			continue
		}
		if filter.Language != "" && item.Language != filter.Language {
			continue
		}
		items = append(items, item)
	}

	sort.Slice(items, func(i, j int) bool {
		if filter.SortBy == "problemId" {
			if items[i].ProblemID == items[j].ProblemID {
				return items[i].ID > items[j].ID
			}
			return items[i].ProblemID < items[j].ProblemID
		}
		return items[i].ID > items[j].ID
	})

	total := len(items)
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.PageSize <= 0 {
		filter.PageSize = 20
	}
	start := (filter.Page - 1) * filter.PageSize
	if start >= total {
		return []domain.Submission{}, total, nil
	}
	end := start + filter.PageSize
	if end > total {
		end = total
	}
	return items[start:end], total, nil
}

// SaveCaseResults 保存单点结果。
func (r *InMemorySubmissionRepository) SaveCaseResults(_ context.Context, submissionID int64, cases []domain.SubmissionCaseResult) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	copyCases := make([]domain.SubmissionCaseResult, len(cases))
	copy(copyCases, cases)
	r.caseData[submissionID] = copyCases
	return nil
}

// GetCaseResults 读取单点结果。
func (r *InMemorySubmissionRepository) GetCaseResults(_ context.Context, submissionID int64) ([]domain.SubmissionCaseResult, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	cases := r.caseData[submissionID]
	copyCases := make([]domain.SubmissionCaseResult, len(cases))
	copy(copyCases, cases)
	return copyCases, nil
}

// CountRecentByUser 统计最近窗口内的提交次数。
func (r *InMemorySubmissionRepository) CountRecentByUser(_ context.Context, userID int64, since time.Time) (int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	count := 0
	for _, item := range r.items {
		if item.UserID == userID && item.CreatedAt.After(since) {
			count++
		}
	}
	return count, nil
}
