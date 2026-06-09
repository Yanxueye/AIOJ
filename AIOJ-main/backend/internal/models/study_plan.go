package models

import "time"

type StudyPlan struct {
	ID          uint64          `gorm:"primaryKey;autoIncrement" json:"id"`
	Title       string          `gorm:"type:varchar(128);index" json:"title"`
	Description string          `gorm:"type:text" json:"description"`
	Difficulty  string          `gorm:"type:varchar(16);index" json:"difficulty"`
	Tags        StringSlice     `gorm:"type:json" json:"tags"`
	CreatedAt   time.Time       `json:"createdAt"`
	UpdatedAt   time.Time       `json:"updatedAt"`
	Items       []StudyPlanItem `gorm:"foreignKey:PlanID" json:"items,omitempty"`
}

func (StudyPlan) TableName() string { return "study_plans" }

type StudyPlanItem struct {
	ID         uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	PlanID     uint64    `gorm:"index;not null" json:"planId"`
	ProblemID  uint64    `gorm:"index;not null" json:"problemId"`
	OrderNo    int       `gorm:"index" json:"orderNo"`
	Title      string    `gorm:"type:varchar(128)" json:"title"`
	Difficulty string    `gorm:"type:varchar(16)" json:"difficulty"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

func (StudyPlanItem) TableName() string { return "study_plan_items" }

type UserPlanProgress struct {
	ID               uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID           uint64    `gorm:"index;not null" json:"userId"`
	PlanID           uint64    `gorm:"index;not null" json:"planId"`
	CompletedCount   int       `json:"completedCount"`
	LastCompletedAt  *time.Time `json:"lastCompletedAt,omitempty"`
	CreatedAt        time.Time `json:"createdAt"`
	UpdatedAt        time.Time `json:"updatedAt"`
}

func (UserPlanProgress) TableName() string { return "user_plan_progress" }

type UserPlanProgressItem struct {
	ID          uint64     `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID      uint64     `gorm:"index;not null" json:"userId"`
	PlanID      uint64     `gorm:"index;not null" json:"planId"`
	ProblemID   uint64     `gorm:"index;not null" json:"problemId"`
	Completed   bool       `gorm:"index;default:false" json:"completed"`
	CompletedAt *time.Time `json:"completedAt,omitempty"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
}

func (UserPlanProgressItem) TableName() string { return "user_plan_progress_items" }

type DailyChallenge struct {
	ID         uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	ProblemID  uint64    `gorm:"index;not null" json:"problemId"`
	Title      string    `gorm:"type:varchar(128)" json:"title"`
	Difficulty string    `gorm:"type:varchar(16)" json:"difficulty"`
	Date       string    `gorm:"type:varchar(16);uniqueIndex" json:"date"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

func (DailyChallenge) TableName() string { return "daily_challenges" }

type StudyCheckin struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID    uint64    `gorm:"index;not null" json:"userId"`
	Date      string    `gorm:"type:varchar(16);index" json:"date"`
	Count     int       `json:"count"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func (StudyCheckin) TableName() string { return "study_checkins" }
