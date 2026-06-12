package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

const (
	ProblemStatusDraft     = "draft"
	ProblemStatusReview    = "review"
	ProblemStatusPublished = "published"
	ProblemStatusArchived  = "archived"
)

// StringSlice persists []string as a JSON text column.
type StringSlice []string

func (s StringSlice) Value() (driver.Value, error) {
	return json.Marshal(s)
}

func (s *StringSlice) Scan(v interface{}) error {
	if v == nil {
		*s = nil
		return nil
	}
	var b []byte
	switch raw := v.(type) {
	case []byte:
		b = raw
	case string:
		b = []byte(raw)
	default:
		return errors.New("StringSlice: unsupported scan type")
	}
	if len(b) == 0 {
		*s = nil
		return nil
	}
	return json.Unmarshal(b, s)
}

// TestCase holds a single I/O case definition used by legacy code paths.
type TestCase struct {
	Input    string `json:"input"`
	Expected string `json:"expected"`
}

// TestCases persists a []TestCase as JSON.
type TestCases []TestCase

func (t TestCases) Value() (driver.Value, error) { return json.Marshal(t) }

func (t *TestCases) Scan(v interface{}) error {
	if v == nil {
		*t = nil
		return nil
	}
	var b []byte
	switch raw := v.(type) {
	case []byte:
		b = raw
	case string:
		b = []byte(raw)
	default:
		return errors.New("TestCases: unsupported scan type")
	}
	if len(b) == 0 {
		*t = nil
		return nil
	}
	return json.Unmarshal(b, t)
}

type Problem struct {
	ID                uint64           `gorm:"primaryKey;autoIncrement" json:"id"`
	Title             string           `gorm:"type:varchar(128);not null;index" json:"title"`
	Difficulty        string           `gorm:"type:varchar(16);index" json:"difficulty"`
	DifficultyScore   int              `gorm:"default:800" json:"difficultyScore"`
	Rating            int              `gorm:"default:800" json:"rating"`
	Tags              StringSlice      `gorm:"type:json" json:"tags"`
	Source            string           `gorm:"type:varchar(64)" json:"source"`
	Status            string           `gorm:"type:varchar(16);index;default:'draft'" json:"status"`
	CurrentVersionID  *uint64          `gorm:"index" json:"currentVersionId,omitempty"`
	PublishedVersionID *uint64         `gorm:"index" json:"publishedVersionId,omitempty"`
	ReviewComment     string           `gorm:"type:text" json:"reviewComment,omitempty"`
	PublishedAt       *time.Time       `json:"publishedAt,omitempty"`
	PublishedBy       *uint64          `json:"publishedBy,omitempty"`
	LastEditedBy      *uint64          `json:"lastEditedBy,omitempty"`
	SubmitCount       int              `gorm:"default:0" json:"submitCount"`
	AcceptCount       int              `gorm:"default:0" json:"-"`
	CreatedAt         time.Time        `json:"-"`
	UpdatedAt         time.Time        `json:"-"`
	CurrentVersion    *ProblemVersion  `gorm:"foreignKey:CurrentVersionID" json:"currentVersion,omitempty"`
	PublishedVersion  *ProblemVersion  `gorm:"foreignKey:PublishedVersionID" json:"publishedVersion,omitempty"`
	Versions          []ProblemVersion `gorm:"foreignKey:ProblemID" json:"versions,omitempty"`
}

func (Problem) TableName() string { return "problems" }

type ProblemVersion struct {
	ID              uint64             `gorm:"primaryKey;autoIncrement" json:"id"`
	ProblemID       uint64             `gorm:"index;not null" json:"problemId"`
	VersionNo       int                `gorm:"not null" json:"versionNo"`
	Title           string             `gorm:"type:varchar(128);not null" json:"title"`
	Difficulty      string             `gorm:"type:varchar(16);index" json:"difficulty"`
	DifficultyScore int                `gorm:"default:800" json:"difficultyScore"`
	Tags            StringSlice        `gorm:"type:json" json:"tags"`
	Content         string             `gorm:"type:longtext" json:"content"`
	Constraints     string             `gorm:"type:text" json:"constraints,omitempty"`
	Source          string             `gorm:"type:varchar(64)" json:"source"`
	TimeLimit       int                `gorm:"default:1000" json:"timeLimit"`
	MemoryLimit     int                `gorm:"default:256" json:"memoryLimit"`
	OutputLimitKB   int32              `gorm:"default:1024" json:"outputLimitKb"`
	Editorial       string             `gorm:"type:longtext" json:"editorial,omitempty"`
	CreatedBy       *uint64            `json:"createdBy,omitempty"`
	PublishedAt     *time.Time         `json:"publishedAt,omitempty"`
	CreatedAt       time.Time          `json:"createdAt"`
	UpdatedAt       time.Time          `json:"updatedAt"`
	Samples         []ProblemSample    `gorm:"foreignKey:VersionID" json:"samples,omitempty"`
	TestCases       []ProblemTestCase  `gorm:"foreignKey:VersionID" json:"testCases,omitempty"`
	Templates       []ProblemTemplate  `gorm:"foreignKey:VersionID" json:"templates,omitempty"`
}

func (ProblemVersion) TableName() string { return "problem_versions" }

type ProblemSample struct {
	ID         uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	VersionID  uint64    `gorm:"index;not null" json:"versionId"`
	CaseNo     int       `gorm:"not null" json:"caseNo"`
	Input      string    `gorm:"type:text" json:"input"`
	Expected   string    `gorm:"type:text" json:"expected"`
	Explanation string   `gorm:"type:text" json:"explanation,omitempty"`
	CreatedAt  time.Time `json:"-"`
	UpdatedAt  time.Time `json:"-"`
}

func (ProblemSample) TableName() string { return "problem_samples" }

type ProblemTestCase struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	VersionID uint64    `gorm:"index;not null" json:"versionId"`
	CaseNo    int       `gorm:"not null" json:"caseNo"`
	Input     string    `gorm:"type:longtext" json:"input"`
	Expected  string    `gorm:"type:longtext" json:"expected"`
	IsHidden  bool      `gorm:"default:true" json:"isHidden"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}

func (ProblemTestCase) TableName() string { return "problem_test_cases" }

type ProblemTemplate struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	VersionID uint64    `gorm:"index;not null" json:"versionId"`
	Language  string    `gorm:"type:varchar(16);index" json:"language"`
	Code      string    `gorm:"type:longtext" json:"code"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}

func (ProblemTemplate) TableName() string { return "problem_templates" }

type AuditLog struct {
	ID           uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID       *uint64   `gorm:"index" json:"userId,omitempty"`
	Username     string    `gorm:"type:varchar(32)" json:"username"`
	UserRole     string    `gorm:"type:varchar(32);index" json:"userRole"`
	ResourceType string    `gorm:"type:varchar(32);index" json:"resourceType"`
	ResourceID   string    `gorm:"type:varchar(64);index" json:"resourceId"`
	Action       string    `gorm:"type:varchar(32);index" json:"action"`
	Detail       string    `gorm:"type:text" json:"detail,omitempty"`
	CreatedAt    time.Time `json:"createdAt"`
}

func (AuditLog) TableName() string { return "audit_logs" }

type Favorite struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID    uint64    `gorm:"uniqueIndex:idx_user_problem;not null" json:"userId"`
	ProblemID uint64    `gorm:"uniqueIndex:idx_user_problem;not null" json:"problemId"`
	CreatedAt time.Time `json:"createdAt"`
}

func (Favorite) TableName() string { return "favorites" }

type ProblemSolution struct {
	ID          uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	ProblemID   uint64    `gorm:"index;not null" json:"problemId"`
	UserID      uint64    `gorm:"index;not null" json:"userId"`
	Username    string    `gorm:"type:varchar(32)" json:"username"`
	Title       string    `gorm:"type:varchar(128)" json:"title"`
	Content     string    `gorm:"type:longtext" json:"content"`
	Language    string    `gorm:"type:varchar(16)" json:"language"`
	IsPublished bool      `gorm:"index;default:false" json:"isPublished"`
	IsOfficial  bool      `gorm:"index;default:false" json:"isOfficial"`
	LikeCount   int       `gorm:"default:0" json:"likeCount"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

func (ProblemSolution) TableName() string { return "problem_solutions" }

type SolutionLike struct {
	ID         uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	SolutionID uint64    `gorm:"uniqueIndex:idx_solution_user;not null" json:"solutionId"`
	UserID     uint64    `gorm:"uniqueIndex:idx_solution_user;not null" json:"userId"`
	CreatedAt  time.Time `json:"createdAt"`
}

func (SolutionLike) TableName() string { return "solution_likes" }

func (p Problem) OutputLimitKBOrDefault() int32 {
	if p.CurrentVersion != nil && p.CurrentVersion.OutputLimitKB > 0 {
		return p.CurrentVersion.OutputLimitKB
	}
	if p.OutputLimitFromPublished() > 0 {
		return p.OutputLimitFromPublished()
	}
	return 1024
}

func (p Problem) OutputLimitFromPublished() int32 {
	if p.PublishedVersion != nil && p.PublishedVersion.OutputLimitKB > 0 {
		return p.PublishedVersion.OutputLimitKB
	}
	return 0
}

// AcceptRate renders the percentage string expected by the frontend.
func (p Problem) AcceptRate() string {
	if p.SubmitCount <= 0 {
		return "0.0"
	}
	return sprintFloat(float64(p.AcceptCount) * 100.0 / float64(p.SubmitCount))
}

type Announcement struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	Title     string    `gorm:"type:varchar(128)" json:"title"`
	Content   string    `gorm:"type:text" json:"content"`
	Type      string    `gorm:"type:varchar(16);default:'info'" json:"type"`
	Date      string    `gorm:"type:date" json:"date"`
	CreatedAt time.Time `json:"-"`
}

func (Announcement) TableName() string { return "announcements" }
