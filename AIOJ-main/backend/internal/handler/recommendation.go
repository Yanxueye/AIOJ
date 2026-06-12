package handler

import (
	"log"
	"math/rand"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/terminaloj/backend/internal/middleware"
	"github.com/terminaloj/backend/internal/models"
	"github.com/terminaloj/backend/internal/utils"
	"gorm.io/gorm"
)

type RecommendationHandler struct {
	DB *gorm.DB
}

// DailyRecommendation returns 5 recommended problems based on user's weak knowledge points and rating
func (h *RecommendationHandler) DailyRecommendation(c *gin.Context) {
	uid, _ := middleware.CurrentUserID(c)
	if uid == 0 {
		// Not logged in, return random problems
		var problems []models.Problem
		h.DB.Where("status = ?", models.ProblemStatusPublished).Order("RAND()").Limit(5).Find(&problems)
		utils.OK(c, gin.H{"items": formatRecommendations(problems, nil)})
		return
	}

	// Get user rating
	var user models.User
	if err := h.DB.First(&user, uid).Error; err != nil {
		log.Printf("[recommendation] user query for uid=%d failed: %v", uid, err)
		utils.OK(c, gin.H{"items": []gin.H{}})
		return
	}

	// Get user's solved problem IDs
	var solvedIDs []uint64
	if err := h.DB.Model(&models.Submission{}).
		Where("user_id = ? AND status = ?", uid, models.StatusAccepted).
		Distinct("problem_id").Pluck("problem_id", &solvedIDs).Error; err != nil {
		log.Printf("[recommendation] solved IDs query for uid=%d failed: %v", uid, err)
	}
	solvedSet := make(map[uint64]bool)
	for _, id := range solvedIDs {
		solvedSet[id] = true
	}

	// Find weak knowledge points (low mastery)
	var weakKPs []models.UserKnowledgeMastery
	if err := h.DB.Where("user_id = ? AND mastery_level < 50", uid).
		Order("mastery_level ASC").Limit(5).Find(&weakKPs).Error; err != nil {
		log.Printf("[recommendation] weak KP query for uid=%d failed: %v", uid, err)
	}

	// Get problems related to weak knowledge points (single query, no N+1)
	var candidateIDs []uint64
	if len(weakKPs) > 0 {
		kpIDs := make([]uint64, len(weakKPs))
		for i, kp := range weakKPs {
			kpIDs[i] = kp.KnowledgePointID
		}
		var mappings []models.ProblemKnowledgePoint
		h.DB.Where("knowledge_point_id IN ?", kpIDs).Find(&mappings)
		for _, m := range mappings {
			if !solvedSet[m.ProblemID] {
				candidateIDs = append(candidateIDs, m.ProblemID)
			}
		}
	}

	// If not enough candidates from weak KPs, add random unsolved problems
	var problems []models.Problem
	if len(candidateIDs) > 0 {
		h.DB.Where("id IN ? AND status = ? AND rating >= ? AND rating <= ?",
			candidateIDs, models.ProblemStatusPublished, user.Rating-200, user.Rating+400).
			Order("RAND()").Limit(5).Find(&problems)
	}

	// Fill remaining slots with rating-matched random problems
	if len(problems) < 5 {
		remaining := 5 - len(problems)
		excludeIDs := make([]uint64, len(problems))
		for i, p := range problems {
			excludeIDs[i] = p.ID
		}
		for _, id := range solvedIDs {
			excludeIDs = append(excludeIDs, id)
		}

		var extra []models.Problem
		q := h.DB.Where("status = ?", models.ProblemStatusPublished)
		if len(excludeIDs) > 0 {
			q = q.Where("id NOT IN ?", excludeIDs)
		}
		q.Where("rating >= ? AND rating <= ?", user.Rating-200, user.Rating+400).
			Order("RAND()").Limit(remaining).Find(&extra)
		problems = append(problems, extra...)
	}

	// If still not enough, relax rating constraint
	if len(problems) < 5 {
		remaining := 5 - len(problems)
		excludeIDs := make([]uint64, len(problems))
		for i, p := range problems {
			excludeIDs[i] = p.ID
		}
		for _, id := range solvedIDs {
			excludeIDs = append(excludeIDs, id)
		}

		var extra []models.Problem
		q := h.DB.Where("status = ?", models.ProblemStatusPublished)
		if len(excludeIDs) > 0 {
			q = q.Where("id NOT IN ?", excludeIDs)
		}
		q.Order("RAND()").Limit(remaining).Find(&extra)
		problems = append(problems, extra...)
	}

	utils.OK(c, gin.H{"items": formatRecommendations(problems, &user)})
}

func formatRecommendations(problems []models.Problem, user *models.User) []gin.H {
	items := make([]gin.H, len(problems))
	for i, p := range problems {
		diff := 0
		if user != nil {
			diff = p.Rating - user.Rating
		}
		items[i] = gin.H{
			"id":         p.ID,
			"title":      p.Title,
			"difficulty":  p.Difficulty,
			"rating":     p.Rating,
			"tags":       p.Tags,
			"ratingDiff": diff,
		}
	}
	return items
}

// LearningPath returns recommended knowledge points based on user's mastery
func (h *RecommendationHandler) LearningPath(c *gin.Context) {
	uid, _ := middleware.CurrentUserID(c)
	if uid == 0 {
		utils.Unauthorized(c, "请先登录")
		return
	}

	// Get all knowledge points
	var allKPs []models.KnowledgePoint
	if err := h.DB.Find(&allKPs).Error; err != nil {
		log.Printf("[recommendation] all KPs query failed: %v", err)
	}
	kpMap := make(map[uint64]*models.KnowledgePoint)
	for i := range allKPs {
		kpMap[allKPs[i].ID] = &allKPs[i]
	}

	// Get user's mastery
	var masteries []models.UserKnowledgeMastery
	if err := h.DB.Where("user_id = ?", uid).Find(&masteries).Error; err != nil {
		log.Printf("[recommendation] mastery query for uid=%d failed: %v", uid, err)
	}
	masteryMap := make(map[uint64]float64)
	for _, m := range masteries {
		masteryMap[m.KnowledgePointID] = m.MasteryLevel
	}

	// Build learning path: mastered topics → suggest related/next topics
	type PathItem struct {
		KnowledgePoint models.KnowledgePoint `json:"knowledgePoint"`
		Mastery        float64              `json:"mastery"`
		Suggestion     string               `json:"suggestion"`
		ProblemCount   int                  `json:"problemCount"`
	}

	var path []PathItem

	// Find mastered topics (mastery > 60) and suggest their children or siblings
	for _, kp := range allKPs {
		mastery := masteryMap[kp.ID]
		if mastery > 60 && kp.ParentID != nil {
			// Mastered sub-topic, suggest siblings
			parentID := *kp.ParentID
			for _, sibling := range allKPs {
				if sibling.ParentID != nil && *sibling.ParentID == parentID && sibling.ID != kp.ID {
					sibMastery := masteryMap[sibling.ID]
					if sibMastery < 30 {
						var count int64
						h.DB.Model(&models.ProblemKnowledgePoint{}).Where("knowledge_point_id = ?", sibling.ID).Count(&count)
						path = append(path, PathItem{
							KnowledgePoint: sibling,
							Mastery:        sibMastery,
							Suggestion:     "已掌握" + kp.Name + "，建议拓展到" + sibling.Name,
							ProblemCount:   int(count),
						})
					}
				}
			}
		} else if mastery < 30 && mastery > 0 {
			// Has some exposure but not mastered
			var count int64
			h.DB.Model(&models.ProblemKnowledgePoint{}).Where("knowledge_point_id = ?", kp.ID).Count(&count)
			path = append(path, PathItem{
				KnowledgePoint: kp,
				Mastery:        mastery,
				Suggestion:     "继续练习" + kp.Name + "以提高掌握度",
				ProblemCount:   int(count),
			})
		}
	}

	// Limit to 10 suggestions
	if len(path) > 10 {
		// Shuffle and pick 10
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		r.Shuffle(len(path), func(i, j int) { path[i], path[j] = path[j], path[i] })
		path = path[:10]
	}

	// If no path items, suggest starting with basics
	if len(path) == 0 {
		for _, kp := range allKPs {
			if kp.ParentID == nil {
				// Top-level category
				var count int64
				h.DB.Model(&models.ProblemKnowledgePoint{}).Where("knowledge_point_id = ?", kp.ID).Count(&count)
				path = append(path, PathItem{
					KnowledgePoint: kp,
					Mastery:        0,
					Suggestion:     "开始学习" + kp.Name,
					ProblemCount:   int(count),
				})
				if len(path) >= 5 {
					break
				}
			}
		}
	}

	utils.OK(c, gin.H{"items": path})
}

// WeaknessAnalysis analyzes user's weak areas
func (h *RecommendationHandler) WeaknessAnalysis(c *gin.Context) {
	uid, _ := middleware.CurrentUserID(c)
	if uid == 0 {
		utils.Unauthorized(c, "请先登录")
		return
	}

	var user models.User
	h.DB.First(&user, uid)

	// Get mastery data
	var masteries []models.UserKnowledgeMastery
	if err := h.DB.Where("user_id = ?", uid).Order("mastery_level ASC").Find(&masteries).Error; err != nil {
		log.Printf("[recommendation] weakness mastery query for uid=%d failed: %v", uid, err)
	}

	// Get all knowledge points
	var allKPs []models.KnowledgePoint
	if err := h.DB.Find(&allKPs).Error; err != nil {
		log.Printf("[recommendation] weakness all KPs query failed: %v", err)
	}
	kpMap := make(map[uint64]models.KnowledgePoint)
	for _, kp := range allKPs {
		kpMap[kp.ID] = kp
	}

	type Weakness struct {
		KnowledgePointID uint64  `json:"knowledgePointId"`
		Name             string  `json:"name"`
		Category         string  `json:"category"`
		Mastery          float64 `json:"mastery"`
		ProblemsSolved   int     `json:"problemsSolved"`
		TotalProblems    int     `json:"totalProblems"`
	}

	var weaknesses []Weakness
	for _, m := range masteries {
		if kp, ok := kpMap[m.KnowledgePointID]; ok {
			weaknesses = append(weaknesses, Weakness{
				KnowledgePointID: m.KnowledgePointID,
				Name:             kp.Name,
				Category:         kp.Category,
				Mastery:          m.MasteryLevel,
				ProblemsSolved:   m.ProblemsSolved,
				TotalProblems:    m.TotalProblems,
			})
		}
	}

	// Calculate overall stats
	totalSolved := 0
	for _, m := range masteries {
		totalSolved += m.ProblemsSolved
	}

	utils.OK(c, gin.H{
		"weaknesses": weaknesses,
		"rating":     user.Rating,
		"totalSolved": totalSolved,
	})
}

