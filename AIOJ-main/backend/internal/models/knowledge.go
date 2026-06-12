package models

import "time"

type KnowledgePoint struct {
	ID          uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	Name        string    `gorm:"type:varchar(64);uniqueIndex;not null" json:"name"`
	Description string    `gorm:"type:text" json:"description"`
	Category    string    `gorm:"type:varchar(32);index" json:"category"`
	ParentID    *uint64   `gorm:"index" json:"parentId,omitempty"`
	OjWikiURL   string    `gorm:"type:varchar(256)" json:"ojWikiUrl,omitempty"`
	Color       string    `gorm:"type:varchar(16)" json:"color,omitempty"`
	Icon        string    `gorm:"type:varchar(32)" json:"icon,omitempty"`
	CreatedAt   time.Time `json:"-"`
	UpdatedAt   time.Time `json:"-"`
}

func (KnowledgePoint) TableName() string { return "knowledge_points" }

type ProblemKnowledgePoint struct {
	ID               uint64 `gorm:"primaryKey;autoIncrement" json:"id"`
	ProblemID        uint64 `gorm:"uniqueIndex:idx_problem_kp;not null" json:"problemId"`
	KnowledgePointID uint64 `gorm:"uniqueIndex:idx_problem_kp;not null" json:"knowledgePointId"`
}

func (ProblemKnowledgePoint) TableName() string { return "problem_knowledge_points" }

type UserKnowledgeMastery struct {
	ID               uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID           uint64    `gorm:"uniqueIndex:idx_user_kp;not null" json:"userId"`
	KnowledgePointID uint64    `gorm:"uniqueIndex:idx_user_kp;not null" json:"knowledgePointId"`
	MasteryLevel     float64   `gorm:"default:0" json:"masteryLevel"` // 0-100
	ProblemsSolved   int       `gorm:"default:0" json:"problemsSolved"`
	TotalProblems    int       `gorm:"default:0" json:"totalProblems"`
	LastUpdatedAt    time.Time `json:"lastUpdatedAt"`
}

func (UserKnowledgeMastery) TableName() string { return "user_knowledge_mastery" }
