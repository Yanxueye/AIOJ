package models

import "time"

const (
	RoleUser  = "user"
	RoleAdmin = "admin"
)

type User struct {
	ID           uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	Username     string    `gorm:"type:varchar(32);uniqueIndex;not null" json:"username"`
	Email        string    `gorm:"type:varchar(128);uniqueIndex;not null" json:"email"`
	PasswordHash string    `gorm:"type:varchar(128);not null" json:"-"`
	Role         string    `gorm:"type:varchar(16);default:'user';index" json:"role"`
	Avatar       string    `gorm:"type:varchar(256)" json:"avatar"`
	Bio          string    `gorm:"type:varchar(256)" json:"bio"`
	Rating       int       `gorm:"default:1000" json:"rating"`
	CreatedAt    time.Time `json:"registeredAt"`
	UpdatedAt    time.Time `json:"-"`
}

func (User) TableName() string { return "users" }

// Profile is the DTO returned to the frontend, matching API.md 1.1 / 2.1.
type Profile struct {
	ID                 uint64         `json:"id"`
	Username           string         `json:"username"`
	Email              string         `json:"email"`
	Role               string         `json:"role"`
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
	Favorites          []FavoriteDigest `json:"favorites,omitempty"`
	RecentSubmissions  []SubmissionTimelineItem `json:"recentSubmissions,omitempty"`
}

type DailyCount struct {
	Date  string `json:"date"`
	Count int    `json:"count"`
}

type FavoriteDigest struct {
	ProblemID   uint64 `json:"problemId"`
	Title       string `json:"title"`
	Difficulty  string `json:"difficulty"`
	AcceptRate  string `json:"acceptRate"`
	FavoritedAt string `json:"favoritedAt"`
}

type SubmissionTimelineItem struct {
	SubmissionID uint64 `json:"submissionId"`
	ProblemID    uint64 `json:"problemId"`
	ProblemTitle string `json:"problemTitle"`
	Status       string `json:"status"`
	Language     string `json:"language"`
	CreatedAt    string `json:"createdAt"`
}

// RatingHistory records each rating change for a user.
type RatingHistory struct {
	ID         uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID     uint64    `gorm:"index;not null" json:"userId"`
	OldRating  int       `json:"oldRating"`
	NewRating  int       `json:"newRating"`
	Delta      int       `json:"delta"`
	ProblemID  uint64    `json:"problemId"`
	Reason     string    `gorm:"type:varchar(64)" json:"reason"`
	CreatedAt  time.Time `json:"createdAt"`
}

func (RatingHistory) TableName() string { return "rating_history" }
