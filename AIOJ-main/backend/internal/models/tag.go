package models

// AlgorithmTag defines all valid algorithm tags in the system.
// Problems reference tags by name (stored in Problem.Tags as StringSlice).
// This table is populated by syncTagsFromKnowledgePoints in seed.go,
// which mirrors KnowledgePoint names — the single source of truth.
type AlgorithmTag struct {
	ID       uint64 `gorm:"primaryKey;autoIncrement" json:"id"`
	Name     string `gorm:"type:varchar(64);uniqueIndex;not null" json:"name"`
	Category string `gorm:"type:varchar(32);index;not null" json:"category"`
	Parent   string `gorm:"type:varchar(64);index" json:"parent,omitempty"`
	OrderNo  int    `gorm:"default:0" json:"orderNo"`
}

func (AlgorithmTag) TableName() string { return "algorithm_tags" }
