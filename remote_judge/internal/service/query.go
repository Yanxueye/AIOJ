package service

import (
	"context"

	"remote_judge/internal/domain"
	"remote_judge/internal/repository"
)

// QueryService 提供提交查询能力。
type QueryService struct {
	submissions repository.SubmissionRepository
}

// NewQueryService 创建查询服务。
func NewQueryService(submissions repository.SubmissionRepository) *QueryService {
	return &QueryService{submissions: submissions}
}

// List 查询提交列表。
func (s *QueryService) List(ctx context.Context, filter domain.SubmissionFilter) ([]domain.Submission, int, error) {
	return s.submissions.List(ctx, filter)
}

// Get 查询单条提交。
func (s *QueryService) Get(ctx context.Context, id int64) (*domain.Submission, error) {
	return s.submissions.GetByID(ctx, id)
}

// Cases 查询单点结果。
func (s *QueryService) Cases(ctx context.Context, submissionID int64) ([]domain.SubmissionCaseResult, error) {
	return s.submissions.GetCaseResults(ctx, submissionID)
}
