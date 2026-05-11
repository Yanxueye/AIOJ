package models

import "time"

type Conversation struct {
	ID        string    `gorm:"primaryKey;type:varchar(40)" json:"id"`
	UserID    uint64    `gorm:"index" json:"-"`
	ProblemID *uint64   `gorm:"index" json:"problemId,omitempty"`
	Title     string    `gorm:"type:varchar(128)" json:"title"`
	CreatedAt time.Time `json:"createdAt"`
}

func (Conversation) TableName() string { return "conversations" }

type Message struct {
	ID             uint64    `gorm:"primaryKey;autoIncrement" json:"-"`
	ConversationID string    `gorm:"index;type:varchar(40)" json:"-"`
	Role           string    `gorm:"type:varchar(16)" json:"role"`
	Content        string    `gorm:"type:longtext" json:"content"`
	CreatedAt      time.Time `json:"-"`
}

func (Message) TableName() string { return "messages" }
