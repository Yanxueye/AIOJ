package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
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

// TestCase holds a single I/O sample used by the judger service.
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
	ID              uint64      `gorm:"primaryKey" json:"id"`
	Title           string      `gorm:"type:varchar(128);not null;index" json:"title"`
	Difficulty      string      `gorm:"type:varchar(16);index" json:"difficulty"`
	DifficultyScore int         `gorm:"default:800" json:"difficultyScore"`
	Tags            StringSlice `gorm:"type:json" json:"tags"`
	Content         string      `gorm:"type:longtext" json:"content,omitempty"`
	TimeLimit       int         `gorm:"default:1000" json:"timeLimit"`
	MemoryLimit     int         `gorm:"default:256" json:"memoryLimit"`
	Source          string      `gorm:"type:varchar(64)" json:"source"`
	TestCases       TestCases   `gorm:"type:json" json:"-"`
	SubmitCount     int         `gorm:"default:0" json:"submitCount"`
	AcceptCount     int         `gorm:"default:0" json:"-"`
	CreatedAt       time.Time   `json:"-"`
	UpdatedAt       time.Time   `json:"-"`
}

func (Problem) TableName() string { return "problems" }

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
	Date      string    `gorm:"type:varchar(16)" json:"date"`
	CreatedAt time.Time `json:"-"`
}

func (Announcement) TableName() string { return "announcements" }
