package models

import (
	"fmt"
	"time"
)

const (
	StatusPending    = "Pending"
	StatusAccepted   = "Accepted"
	StatusWrong      = "Wrong Answer"
	StatusTLE        = "Time Limit Exceeded"
	StatusRuntimeErr = "Runtime Error"
	StatusCompileErr = "Compilation Error"
)

type Submission struct {
	ID           uint64    `gorm:"primaryKey" json:"id"`
	UserID       uint64    `gorm:"index;not null" json:"-"`
	ProblemID    uint64    `gorm:"index;not null" json:"problemId"`
	ProblemTitle string    `gorm:"type:varchar(128)" json:"problemTitle"`
	Language     string    `gorm:"type:varchar(16)" json:"language"`
	Code         string    `gorm:"type:longtext" json:"code,omitempty"`
	CodeLength   int       `json:"codeLength"`
	Status       string    `gorm:"type:varchar(32);index;default:'Pending'" json:"status"`
	Runtime      int       `json:"runtime"`
	Memory       string    `gorm:"type:varchar(16)" json:"memory"`
	ErrorMessage string    `gorm:"type:text" json:"errorMessage,omitempty"`
	CreatedAt    time.Time `gorm:"index" json:"createdAt"`
}

func (Submission) TableName() string { return "submissions" }

// sprintFloat centralises the "%.1f" formatting used across models.
func sprintFloat(f float64) string { return fmt.Sprintf("%.1f", f) }
