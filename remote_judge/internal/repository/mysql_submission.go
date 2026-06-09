package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	_ "github.com/go-sql-driver/mysql"

	"remote_judge/internal/domain"
)

// MySQLSubmissionRepository 提供基于 MySQL 的提交仓储实现。
type MySQLSubmissionRepository struct {
	db *sql.DB
}

// NewMySQLSubmissionRepository 创建 MySQL 提交仓储。
func NewMySQLSubmissionRepository(db *sql.DB) *MySQLSubmissionRepository {
	return &MySQLSubmissionRepository{db: db}
}

// Create 创建一条提交记录。
func (r *MySQLSubmissionRepository) Create(ctx context.Context, sub *domain.Submission) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO submissions
		(id, user_id, problem_id, trace_id, language, code, code_length, status, runtime_ms, memory_kb, compile_output, error_message, queue_started_at, judge_started_at, finished_at, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		sub.ID, sub.UserID, sub.ProblemID, sub.TraceID, sub.Language, sub.Code, sub.CodeLength, sub.Status, sub.RuntimeMs, sub.MemoryKB,
		sub.CompileOutput, sub.ErrorMessage, sub.QueueStartedAt, sub.JudgeStartedAt, sub.FinishedAt, sub.CreatedAt, sub.UpdatedAt,
	)
	return err
}

// Update 更新一条提交记录。
func (r *MySQLSubmissionRepository) Update(ctx context.Context, sub *domain.Submission) error {
	sub.UpdatedAt = time.Now()
	result, err := r.db.ExecContext(ctx, `
		UPDATE submissions
		SET status=?, runtime_ms=?, memory_kb=?, compile_output=?, error_message=?, queue_started_at=?, judge_started_at=?, finished_at=?, updated_at=?
		WHERE id=?`,
		sub.Status, sub.RuntimeMs, sub.MemoryKB, sub.CompileOutput, sub.ErrorMessage, sub.QueueStartedAt, sub.JudgeStartedAt, sub.FinishedAt, sub.UpdatedAt, sub.ID,
	)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err == nil && affected == 0 {
		return ErrNotFound
	}
	return err
}

// GetByID 读取指定提交。
func (r *MySQLSubmissionRepository) GetByID(ctx context.Context, id int64) (*domain.Submission, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, user_id, problem_id, trace_id, language, code, code_length, status, runtime_ms, memory_kb, compile_output, error_message, queue_started_at, judge_started_at, finished_at, created_at, updated_at
		FROM submissions WHERE id=?`, id)
	var sub domain.Submission
	var status string
	err := row.Scan(
		&sub.ID, &sub.UserID, &sub.ProblemID, &sub.TraceID, &sub.Language, &sub.Code, &sub.CodeLength, &status, &sub.RuntimeMs, &sub.MemoryKB,
		&sub.CompileOutput, &sub.ErrorMessage, &sub.QueueStartedAt, &sub.JudgeStartedAt, &sub.FinishedAt, &sub.CreatedAt, &sub.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	sub.Status = domain.SubmissionStatus(status)
	return &sub, nil
}

// List 分页查询提交记录。
func (r *MySQLSubmissionRepository) List(ctx context.Context, filter domain.SubmissionFilter) ([]domain.Submission, int, error) {
	query := `
		SELECT id, user_id, problem_id, trace_id, language, code, code_length, status, runtime_ms, memory_kb, compile_output, error_message, queue_started_at, judge_started_at, finished_at, created_at, updated_at
		FROM submissions
		WHERE user_id=?`
	args := []any{filter.UserID}
	if filter.ProblemID > 0 {
		query += " AND problem_id=?"
		args = append(args, filter.ProblemID)
	}
	if filter.Status != "" {
		query += " AND status=?"
		args = append(args, filter.Status)
	}
	if filter.Language != "" {
		query += " AND language=?"
		args = append(args, filter.Language)
	}

	countQuery := "SELECT COUNT(*) FROM (" + query + ") AS t"
	var total int
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	orderBy := " ORDER BY id DESC"
	if filter.SortBy == "problemId" {
		orderBy = " ORDER BY problem_id ASC, id DESC"
	}
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.PageSize <= 0 {
		filter.PageSize = 20
	}
	query += orderBy + " LIMIT ? OFFSET ?"
	args = append(args, filter.PageSize, (filter.Page-1)*filter.PageSize)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var items []domain.Submission
	for rows.Next() {
		var sub domain.Submission
		var status string
		if err := rows.Scan(
			&sub.ID, &sub.UserID, &sub.ProblemID, &sub.TraceID, &sub.Language, &sub.Code, &sub.CodeLength, &status, &sub.RuntimeMs, &sub.MemoryKB,
			&sub.CompileOutput, &sub.ErrorMessage, &sub.QueueStartedAt, &sub.JudgeStartedAt, &sub.FinishedAt, &sub.CreatedAt, &sub.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		sub.Status = domain.SubmissionStatus(status)
		items = append(items, sub)
	}
	return items, total, rows.Err()
}

// SaveCaseResults 保存单测试点结果。
func (r *MySQLSubmissionRepository) SaveCaseResults(ctx context.Context, submissionID int64, cases []domain.SubmissionCaseResult) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, `DELETE FROM submission_case_results WHERE submission_id=?`, submissionID); err != nil {
		return err
	}
	for _, item := range cases {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO submission_case_results (submission_id, case_no, status, runtime_ms, memory_kb, stdout_bytes, stderr_bytes, signal_name, stdout_preview, stderr_preview)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			submissionID, item.CaseNo, item.Status, item.RuntimeMs, item.MemoryKB, item.StdoutBytes, item.StderrBytes, item.Signal, item.StdoutPreview, item.StderrPreview,
		); err != nil {
			return err
		}
	}
	return tx.Commit()
}

// GetCaseResults 读取单测试点结果。
func (r *MySQLSubmissionRepository) GetCaseResults(ctx context.Context, submissionID int64) ([]domain.SubmissionCaseResult, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT submission_id, case_no, status, runtime_ms, memory_kb, stdout_bytes, stderr_bytes, signal_name, stdout_preview, stderr_preview
		FROM submission_case_results
		WHERE submission_id=?
		ORDER BY case_no ASC`, submissionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []domain.SubmissionCaseResult
	for rows.Next() {
		var item domain.SubmissionCaseResult
		var status string
		if err := rows.Scan(&item.SubmissionID, &item.CaseNo, &status, &item.RuntimeMs, &item.MemoryKB, &item.StdoutBytes, &item.StderrBytes, &item.Signal, &item.StdoutPreview, &item.StderrPreview); err != nil {
			return nil, err
		}
		item.Status = domain.SubmissionStatus(status)
		items = append(items, item)
	}
	return items, rows.Err()
}

// CountRecentByUser 统计限流窗口内的提交数量。
func (r *MySQLSubmissionRepository) CountRecentByUser(ctx context.Context, userID int64, since time.Time) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM submissions WHERE user_id=? AND created_at>?`, userID, since).Scan(&count)
	return count, err
}

// NextID 生成下一个提交编号。
func (r *MySQLSubmissionRepository) NextID(ctx context.Context) (int64, error) {
	var next int64
	err := r.db.QueryRowContext(ctx, `SELECT COALESCE(MAX(id), 100000) + 1 FROM submissions`).Scan(&next)
	return next, err
}
