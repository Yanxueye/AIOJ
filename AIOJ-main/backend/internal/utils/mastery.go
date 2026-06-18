package utils

import (
	"fmt"
	"math"
	"time"

	"github.com/terminaloj/backend/internal/data"
	"github.com/terminaloj/backend/internal/models"
	"gorm.io/gorm"
)

// UpdateMastery recalculates mastery for a user based on their accepted submissions.
func UpdateMastery(db *gorm.DB, userID uint64) {
	var solvedProblems []uint64
	db.Model(&models.Submission{}).
		Where("user_id = ? AND status = ?", userID, models.StatusAccepted).
		Distinct("problem_id").Pluck("problem_id", &solvedProblems)

	if len(solvedProblems) == 0 {
		return
	}

	now := time.Now().UTC()
	kps := data.KnowledgeTree()

	// Batch load existing mastery records (1 query instead of N)
	var existing []models.UserKnowledgeMastery
	db.Where("user_id = ?", userID).Find(&existing)
	existingMap := make(map[int]*models.UserKnowledgeMastery, len(existing))
	for i := range existing {
		existingMap[existing[i].KnowledgePointID] = &existing[i]
	}

	for i, kp := range kps {
		// Count total published problems with this tag
		var total int64
		db.Model(&models.Problem{}).
			Where("status = ? AND JSON_CONTAINS(tags, ?)", models.ProblemStatusPublished,
				fmt.Sprintf(`"%s"`, kp.Name)).Count(&total)
		if total == 0 {
			continue // no problems with this tag, skip
		}

		// Count solved problems with this tag
		var solvedCount int64
		if len(solvedProblems) > 0 {
			db.Model(&models.Problem{}).
				Where("id IN ? AND JSON_CONTAINS(tags, ?)", solvedProblems,
					fmt.Sprintf(`"%s"`, kp.Name)).Count(&solvedCount)
		}

		mastery := math.Min(100, float64(solvedCount)/float64(total)*100)

		if rec, ok := existingMap[i]; ok {
			rec.MasteryLevel = mastery
			rec.ProblemsSolved = int(solvedCount)
			rec.TotalProblems = int(total)
			rec.LastUpdatedAt = now
			db.Save(rec)
		} else {
			db.Create(&models.UserKnowledgeMastery{
				UserID:           userID,
				KnowledgePointID: i,
				MasteryLevel:     mastery,
				ProblemsSolved:   int(solvedCount),
				TotalProblems:    int(total),
				LastUpdatedAt:    now,
			})
		}
	}
}
