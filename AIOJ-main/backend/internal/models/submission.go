package models

import (
	"fmt"
	"time"
)

const (
	StatusPending             = "Pending"
	StatusQueueing            = "Queueing"
	StatusCompiling           = "Compiling"
	StatusRunning             = "Running"
	StatusAccepted            = "Accepted"
	StatusWrong               = "Wrong Answer"
	StatusTLE                 = "Time Limit Exceeded"
	StatusRuntimeErr          = "Runtime Error"
	StatusCompileErr          = "Compile Error"
	StatusMLE                 = "Memory Limit Exceeded"
	StatusOLE                 = "Output Limit Exceeded"
	StatusSystemErr           = "System Error"
)

type SubmissionCaseResult struct {
	ID            uint64    `gorm:"primaryKey;autoIncrement" json:"-"`
	SubmissionID  uint64    `gorm:"index;not null" json:"submissionId"`
	CaseNo        int       `gorm:"not null" json:"caseNo"`
	Status        string    `gorm:"type:varchar(64);index" json:"status"`
	RuntimeMS     int       `json:"runtimeMs"`
	MemoryKB      int       `json:"memoryKb"`
	StdoutBytes   int       `json:"stdoutBytes"`
	StderrBytes   int       `json:"stderrBytes"`
	Signal        string    `gorm:"type:varchar(64)" json:"signal,omitempty"`
	StdoutPreview string    `gorm:"type:text" json:"stdoutPreview,omitempty"`
	StderrPreview string    `gorm:"type:text" json:"stderrPreview,omitempty"`
	CreatedAt     time.Time `json:"-"`
	UpdatedAt     time.Time `json:"-"`
}

func (SubmissionCaseResult) TableName() string { return "submission_case_results" }

type Submission struct {
	ID             uint64                 `gorm:"primaryKey" json:"id"`
	UserID         uint64                 `gorm:"index;not null" json:"-"`
	ProblemID      uint64                 `gorm:"index;not null" json:"problemId"`
	ProblemTitle   string                 `gorm:"type:varchar(128)" json:"problemTitle"`
	TraceID        string                 `gorm:"type:varchar(128);index" json:"traceId"`
	Source         string                 `gorm:"type:varchar(16);index;default:'submit'" json:"source"`
	Language       string                 `gorm:"type:varchar(16)" json:"language"`
	Code           string                 `gorm:"type:longtext" json:"code,omitempty"`
	CodeLength     int                    `json:"codeLength"`
	Status         string                 `gorm:"type:varchar(64);index;default:'Pending'" json:"status"`
	Runtime        int                    `json:"runtime"`
	RuntimeMS      int                    `gorm:"column:runtime_ms" json:"runtimeMs"`
	Memory         string                 `gorm:"type:varchar(16)" json:"memory"`
	MemoryKB       int                    `json:"memoryKb"`
	CompileOutput  string                 `gorm:"type:text" json:"compileOutput,omitempty"`
	ErrorMessage   string                 `gorm:"type:text" json:"errorMessage,omitempty"`
	QueueStartedAt *time.Time             `json:"queueStartedAt,omitempty"`
	JudgeStartedAt *time.Time             `json:"judgeStartedAt,omitempty"`
	FinishedAt     *time.Time             `json:"finishedAt,omitempty"`
	CreatedAt      time.Time              `gorm:"index" json:"createdAt"`
	UpdatedAt      time.Time              `json:"updatedAt"`
	CaseResults    []SubmissionCaseResult `gorm:"foreignKey:SubmissionID;references:ID" json:"caseResults,omitempty"`
}

func (Submission) TableName() string { return "submissions" }

// sprintFloat centralises the "%.1f" formatting used across models.
func sprintFloat(f float64) string { return fmt.Sprintf("%.1f", f) }
