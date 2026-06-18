package handler

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/terminaloj/backend/internal/data"
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
		var problems []models.Problem
		h.DB.Where("status = ?", models.ProblemStatusPublished).Order("RAND()").Limit(5).Find(&problems)
		utils.OK(c, gin.H{"items": formatRecommendations(problems, nil)})
		return
	}

	var user models.User
	if err := h.DB.First(&user, uid).Error; err != nil {
		log.Printf("[recommendation] user query for uid=%d failed: %v", uid, err)
		utils.OK(c, gin.H{"items": []gin.H{}})
		return
	}

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
	_ = solvedSet

	var weakKPs []models.UserKnowledgeMastery
	if err := h.DB.Where("user_id = ? AND mastery_level < 50", uid).
		Order("mastery_level ASC").Limit(5).Find(&weakKPs).Error; err != nil {
		log.Printf("[recommendation] weak KP query for uid=%d failed: %v", uid, err)
	}

	// Find problems related to weak knowledge points by tag name
	var candidateIDs []uint64
	if len(weakKPs) > 0 {
		kpData := data.KnowledgeTree()
		for _, wk := range weakKPs {
			if wk.KnowledgePointID >= 0 && wk.KnowledgePointID < len(kpData) {
				kpName := kpData[wk.KnowledgePointID].Name
				var batch []models.Problem
				h.DB.Where("status = ? AND JSON_CONTAINS(tags, ?) AND id NOT IN ?",
					models.ProblemStatusPublished,
					fmt.Sprintf(`"%s"`, kpName), solvedIDs).
					Limit(3).Find(&batch)
				for _, p := range batch {
					candidateIDs = append(candidateIDs, p.ID)
				}
			}
		}
	}

	var problems []models.Problem
	if len(candidateIDs) > 0 {
		h.DB.Where("id IN ? AND status = ? AND rating >= ? AND rating <= ?",
			candidateIDs, models.ProblemStatusPublished, user.Rating-200, user.Rating+400).
			Order("RAND()").Limit(5).Find(&problems)
	}

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
			"difficulty": p.Difficulty,
			"rating":     p.Rating,
			"tags":       p.Tags,
			"ratingDiff": diff,
		}
	}
	return items
}

// countProblemsByTag counts published problems matching a tag name via JSON_CONTAINS.
func (h *RecommendationHandler) countProblemsByTag(tagName string) int64 {
	var cnt int64
	h.DB.Model(&models.Problem{}).
		Where("status = ? AND JSON_CONTAINS(tags, ?)", models.ProblemStatusPublished,
			fmt.Sprintf(`"%s"`, tagName)).Count(&cnt)
	return cnt
}

// LearningPath returns recommended knowledge points based on user's mastery
func (h *RecommendationHandler) LearningPath(c *gin.Context) {
	uid, _ := middleware.CurrentUserID(c)
	if uid == 0 {
		utils.Unauthorized(c, "请先登录")
		return
	}

	// Get knowledge points from hardcoded data
	allKPs := data.KnowledgeTree()
	kpMap := make(map[int]data.KPNode)
	for i, kp := range allKPs {
		kpMap[i] = kp
	}

	// Get user's mastery
	var masteries []models.UserKnowledgeMastery
	if err := h.DB.Where("user_id = ?", uid).Find(&masteries).Error; err != nil {
		log.Printf("[recommendation] mastery query for uid=%d failed: %v", uid, err)
	}
	masteryMap := make(map[int]float64)
	for _, m := range masteries {
		masteryMap[m.KnowledgePointID] = m.MasteryLevel
	}

	type PathItem struct {
		KnowledgePoint data.KPNode `json:"knowledgePoint"`
		Mastery        float64     `json:"mastery"`
		Suggestion     string      `json:"suggestion"`
		ProblemCount   int         `json:"problemCount"`
	}

	var path []PathItem

	for i, kp := range allKPs {
		mastery := masteryMap[i]
		parentID := -1
		for j, other := range allKPs {
			if other.Name == kp.ParentName {
				parentID = j
				break
			}
		}

		if mastery > 60 && parentID >= 0 {
			// Mastered sub-topic, suggest siblings
			for j, sibling := range allKPs {
				if j == i {
					continue
				}
				sibParentID := -1
				for k, other := range allKPs {
					if other.Name == sibling.ParentName {
						sibParentID = k
						break
					}
				}
				if sibParentID == parentID {
					sibMastery := masteryMap[j]
					if sibMastery < 30 {
						cnt := h.countProblemsByTag(sibling.Name)
						path = append(path, PathItem{
							KnowledgePoint: sibling,
							Mastery:        sibMastery,
							Suggestion:     "已掌握" + kp.Name + "，建议拓展到" + sibling.Name,
							ProblemCount:   int(cnt),
						})
					}
				}
			}
		} else if mastery < 30 && mastery > 0 {
			cnt := h.countProblemsByTag(kp.Name)
			path = append(path, PathItem{
				KnowledgePoint: kp,
				Mastery:        mastery,
				Suggestion:     "继续练习" + kp.Name + "以提高掌握度",
				ProblemCount:   int(cnt),
			})
		}
	}

	if len(path) > 10 {
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		r.Shuffle(len(path), func(i, j int) { path[i], path[j] = path[j], path[i] })
		path = path[:10]
	}

	if len(path) == 0 {
		for i, kp := range allKPs {
			if kp.ParentName == "" {
				cnt := h.countProblemsByTag(kp.Name)
				path = append(path, PathItem{
					KnowledgePoint: kp,
					Mastery:        0,
					Suggestion:     "开始学习" + kp.Name,
					ProblemCount:   int(cnt),
				})
				if len(path) >= 5 {
					break
				}
				_ = i
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

	var masteries []models.UserKnowledgeMastery
	if err := h.DB.Where("user_id = ?", uid).Order("mastery_level ASC").Find(&masteries).Error; err != nil {
		log.Printf("[recommendation] weakness mastery query for uid=%d failed: %v", uid, err)
	}

	// Get knowledge points from hardcoded data
	allKPs := data.KnowledgeTree()
	kpNameMap := make(map[int]data.KPNode)
	for i, kp := range allKPs {
		kpNameMap[i] = kp
	}

	type Weakness struct {
		KnowledgePointID int     `json:"knowledgePointId"`
		Name             string  `json:"name"`
		Category         string  `json:"category"`
		Mastery          float64 `json:"mastery"`
		ProblemsSolved   int     `json:"problemsSolved"`
		TotalProblems    int     `json:"totalProblems"`
	}

	var weaknesses []Weakness
	for _, m := range masteries {
		if kp, ok := kpNameMap[m.KnowledgePointID]; ok {
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

	totalSolved := 0
	for _, m := range masteries {
		totalSolved += m.ProblemsSolved
	}

	utils.OK(c, gin.H{
		"weaknesses":  weaknesses,
		"rating":      user.Rating,
		"totalSolved": totalSolved,
	})
}
