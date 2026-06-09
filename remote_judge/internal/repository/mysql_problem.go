package repository

import (
	"context"
	"database/sql"
	"errors"

	"remote_judge/internal/domain"
)

// MySQLProblemRepository 提供基于 MySQL 的题目仓储实现。
type MySQLProblemRepository struct {
	db *sql.DB
}

// NewMySQLProblemRepository 创建 MySQL 题目仓储。
func NewMySQLProblemRepository(db *sql.DB) *MySQLProblemRepository {
	return &MySQLProblemRepository{db: db}
}

// GetByID 读取题目元数据。
func (r *MySQLProblemRepository) GetByID(ctx context.Context, id int64) (*domain.Problem, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, title, time_limit_ms, memory_limit_mb, output_limit_kb
		FROM problems WHERE id=?`, id)
	var p domain.Problem
	err := row.Scan(&p.ID, &p.Title, &p.TimeLimitMs, &p.MemoryLimitMB, &p.OutputLimitKB)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// ListCases 读取题目的测试点。
func (r *MySQLProblemRepository) ListCases(ctx context.Context, problemID int64) ([]domain.TestCase, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT problem_id, case_no, input_text, expected_text
		FROM test_cases
		WHERE problem_id=?
		ORDER BY case_no ASC`, problemID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cases []domain.TestCase
	for rows.Next() {
		var item domain.TestCase
		if err := rows.Scan(&item.ProblemID, &item.CaseNo, &item.Input, &item.Expected); err != nil {
			return nil, err
		}
		cases = append(cases, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(cases) == 0 {
		return nil, ErrNotFound
	}
	return cases, nil
}
