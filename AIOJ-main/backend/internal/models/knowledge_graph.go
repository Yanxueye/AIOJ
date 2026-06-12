package models

import "time"

// UserKnowledgeGraph stores the AI-generated knowledge graph for a user
type UserKnowledgeGraph struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID    uint64    `gorm:"uniqueIndex;not null" json:"userId"`
	Scope     string    `gorm:"type:varchar(32);default:'recent'" json:"scope"`
	Nodes     string    `gorm:"type:text" json:"nodes"`  // JSON string
	Edges     string    `gorm:"type:text" json:"edges"`  // JSON string
	Summary   string    `gorm:"type:text" json:"summary"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func (UserKnowledgeGraph) TableName() string { return "user_knowledge_graphs" }
