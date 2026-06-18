package models

import "time"

// UserKnowledgeMastery tracks user's mastery progress per knowledge point.
type UserKnowledgeMastery struct {
	ID               uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID           uint64    `gorm:"uniqueIndex:idx_user_kp;not null" json:"userId"`
	KnowledgePointID int       `gorm:"uniqueIndex:idx_user_kp;not null" json:"knowledgePointId"`
	MasteryLevel     float64   `gorm:"default:0" json:"masteryLevel"` // 0-100
	ProblemsSolved   int       `gorm:"default:0" json:"problemsSolved"`
	TotalProblems    int       `gorm:"default:0" json:"totalProblems"`
	LastUpdatedAt    time.Time `json:"lastUpdatedAt"`
}

func (UserKnowledgeMastery) TableName() string { return "user_knowledge_mastery" }
