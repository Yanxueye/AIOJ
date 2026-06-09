package domain

import "time"

// SubmissionMessage 表示队列中的提交消息。
type SubmissionMessage struct {
	SubmissionID int64     `json:"submissionId"`
	UserID       int64     `json:"userId"`
	ProblemID    int64     `json:"problemId"`
	Language     string    `json:"language"`
	TraceID      string    `json:"traceId"`
	CreatedAt    time.Time `json:"createdAt"`
}
