package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"remote_judge/internal/domain"
	"remote_judge/internal/queue"
	"remote_judge/internal/repository"
	"remote_judge/internal/stats"
)

// ErrBadRequest 表示无效的请求参数。
var ErrBadRequest = errors.New("bad request")

// ErrRateLimited 表示来自同一用户的提交过于频繁。
var ErrRateLimited = errors.New("rate limited")

// SubmissionService 处理提交创建流程。
type SubmissionService struct {
	submissions repository.SubmissionRepository
	problems    repository.ProblemRepository
	q           queue.Queue
	stats       *stats.Collector
	nowFunc     func() time.Time
}

// CreateSubmissionRequest 描述一次提交创建请求。
type CreateSubmissionRequest struct {
	UserID    int64
	ProblemID int64  `json:"problemId"`
	Language  string `json:"language"`
	Code      string `json:"code"`
}

// NewSubmissionService 创建提交服务。
func NewSubmissionService(submissions repository.SubmissionRepository, problems repository.ProblemRepository, q queue.Queue, collector *stats.Collector) *SubmissionService {
	return &SubmissionService{
		submissions: submissions,
		problems:    problems,
		q:           q,
		stats:       collector,
		nowFunc:     time.Now,
	}
}

// Create 存储新提交并将其发布到队列。
func (s *SubmissionService) Create(ctx context.Context, req CreateSubmissionRequest) (*domain.Submission, error) {
	if err := validateCreateRequest(req); err != nil {
		return nil, err
	}
	if _, err := s.problems.GetByID(ctx, req.ProblemID); err != nil {
		return nil, err
	}

	count, err := s.submissions.CountRecentByUser(ctx, req.UserID, s.nowFunc().Add(-1*time.Minute))
	if err != nil {
		return nil, err
	}
	if count >= 12 {
		return nil, ErrRateLimited
	}

	id, err := s.submissions.NextID(ctx)
	if err != nil {
		return nil, err
	}

	now := s.nowFunc()
	traceID := fmt.Sprintf("judge-%d-%d", req.ProblemID, id)
	sub := &domain.Submission{
		ID:         id,
		UserID:     req.UserID,
		ProblemID:  req.ProblemID,
		TraceID:    traceID,
		Language:   req.Language,
		Code:       req.Code,
		CodeLength: len([]byte(req.Code)),
		Status:     domain.StatusPending,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	if err := s.submissions.Create(ctx, sub); err != nil {
		return nil, err
	}

	msg := domain.SubmissionMessage{
		SubmissionID: sub.ID,
		UserID:       sub.UserID,
		ProblemID:    sub.ProblemID,
		Language:     sub.Language,
		TraceID:      traceID,
		CreatedAt:    now,
	}
	if err := s.q.Publish(ctx, msg); err != nil {
		return nil, err
	}
	if s.stats != nil {
		s.stats.RecordSubmission()
	}
	return sub, nil
}

// validateCreateRequest checks whether a submission request is acceptable.
func validateCreateRequest(req CreateSubmissionRequest) error {
	if req.UserID <= 0 || req.ProblemID <= 0 {
		return ErrBadRequest
	}
	if strings.TrimSpace(req.Code) == "" {
		return fmt.Errorf("%w: code required", ErrBadRequest)
	}
	if len(req.Code) > 128*1024 {
		return fmt.Errorf("%w: code too long", ErrBadRequest)
	}
	if strings.Contains(req.Code, "\x00") {
		return fmt.Errorf("%w: code contains null byte", ErrBadRequest)
	}
	if _, ok := domain.SupportedLanguages[req.Language]; !ok {
		return fmt.Errorf("%w: language not supported", ErrBadRequest)
	}
	return nil
}
