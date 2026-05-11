package models

import "time"

type User struct {
	ID           uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	Username     string    `gorm:"type:varchar(32);uniqueIndex;not null" json:"username"`
	Email        string    `gorm:"type:varchar(128);uniqueIndex;not null" json:"email"`
	PasswordHash string    `gorm:"type:varchar(128);not null" json:"-"`
	Avatar       string    `gorm:"type:varchar(256)" json:"avatar"`
	Bio          string    `gorm:"type:varchar(256)" json:"bio"`
	Rating       int       `gorm:"default:1200" json:"rating"`
	CreatedAt    time.Time `json:"registeredAt"`
	UpdatedAt    time.Time `json:"-"`
}

func (User) TableName() string { return "users" }

// Profile is the DTO returned to the frontend, matching API.md 1.1 / 2.1.
type Profile struct {
	ID                 uint64         `json:"id"`
	Username           string         `json:"username"`
	Email              string         `json:"email"`
	Avatar             string         `json:"avatar"`
	Bio                string         `json:"bio"`
	Rating             int            `json:"rating"`
	Rank               int            `json:"rank"`
	SolvedCount        int            `json:"solvedCount"`
	TotalSubmissions  int            `json:"totalSubmissions"`
	AcceptRate         string         `json:"acceptRate"`
	RegisteredAt       string         `json:"registeredAt"`
	SolvedByDifficulty map[string]int `json:"solvedByDifficulty,omitempty"`
	SolvedByAlgorithm  map[string]int `json:"solvedByAlgorithm,omitempty"`
	RecentActivity     []DailyCount   `json:"recentActivity,omitempty"`
}

type DailyCount struct {
	Date  string `json:"date"`
	Count int    `json:"count"`
}
