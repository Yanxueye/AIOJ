package utils

import (
	"math"
	"time"

	"github.com/terminaloj/backend/internal/models"
	"gorm.io/gorm"
)

// UpdateMastery recalculates mastery for a user based on their accepted submissions
func UpdateMastery(db *gorm.DB, userID uint64) {
	var solvedProblems []uint64
	db.Model(&models.Submission{}).
		Where("user_id = ? AND status = ?", userID, models.StatusAccepted).
		Distinct("problem_id").Pluck("problem_id", &solvedProblems)

	if len(solvedProblems) == 0 {
		return
	}

	var mappings []models.ProblemKnowledgePoint
	db.Where("problem_id IN ?", solvedProblems).Find(&mappings)

	kpSolved := make(map[uint64]int)
	for _, m := range mappings {
		kpSolved[m.KnowledgePointID]++
	}

	// Use SQL GROUP BY instead of loading all rows into memory
	type kpCount struct {
		KnowledgePointID uint64
		Count            int
	}
	var counts []kpCount
	db.Model(&models.ProblemKnowledgePoint{}).
		Select("knowledge_point_id, COUNT(*) as count").
		Group("knowledge_point_id").Scan(&counts)
	kpTotal := make(map[uint64]int, len(counts))
	for _, c := range counts {
		kpTotal[c.KnowledgePointID] = c.Count
	}

	now := time.Now().UTC()
	for kpID, solved := range kpSolved {
		total := kpTotal[kpID]
		if total == 0 {
			total = 1
		}
		mastery := math.Min(100, float64(solved)/float64(total)*100)

		var record models.UserKnowledgeMastery
		err := db.Where("user_id = ? AND knowledge_point_id = ?", userID, kpID).First(&record).Error
		if err == nil {
			record.MasteryLevel = mastery
			record.ProblemsSolved = solved
			record.TotalProblems = total
			record.LastUpdatedAt = now
			db.Save(&record)
		} else {
			record = models.UserKnowledgeMastery{
				UserID:           userID,
				KnowledgePointID: kpID,
				MasteryLevel:     mastery,
				ProblemsSolved:   solved,
				TotalProblems:    total,
				LastUpdatedAt:    now,
			}
			db.Create(&record)
		}
	}
}
