package handler

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/terminaloj/backend/internal/middleware"
	"github.com/terminaloj/backend/internal/models"
	"github.com/terminaloj/backend/internal/utils"
	aisvc "github.com/terminaloj/backend/internal/ai"
)

// QueryUserProblems handles POST /api/agent/problems (called by agent-service tool executor)
func (h *AIHandler) QueryUserProblems(c *gin.Context) {
	uid, _ := middleware.CurrentUserID(c)
	if uid == 0 {
		utils.BadRequest(c, "user id required")
		return
	}

	var req struct {
		Tags       []string `json:"tags,omitempty"`
		Status     string   `json:"status,omitempty"`
		Difficulty string   `json:"difficulty,omitempty"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "invalid request")
		return
	}

	subs, err := h.recentSubmissions(uid, nil, 500)
	if err != nil {
		utils.Server(c, err.Error())
		return
	}

	problems := h.aggregateProblems(subs)
	tagStats := h.aggregateTagStats(subs)

	filtered := make([]aisvc.ProblemSummary, 0)
	for _, p := range problems {
		if req.Status != "" && p.Status != req.Status {
			continue
		}
		filtered = append(filtered, p)
	}

	// For untried or all, query DB for problems user hasn't attempted
	if req.Status == "untried" || req.Status == "" {
		attemptedIDs := make([]uint64, 0)
		for _, p := range problems {
			attemptedIDs = append(attemptedIDs, p.ID)
		}

		var dbProblems []models.Problem
		query := h.DB.Where("status = ?", models.ProblemStatusPublished)
		if len(req.Tags) > 0 {
			for _, tag := range req.Tags {
				query = query.Where("JSON_CONTAINS(tags, ?)", fmt.Sprintf("\"%s\"", tag))
			}
		}
		if req.Status == "untried" && len(attemptedIDs) > 0 {
			query = query.Where("id NOT IN ?", attemptedIDs)
		}
		if err := query.Limit(50).Find(&dbProblems).Error; err == nil {
			for _, p := range dbProblems {
				filtered = append(filtered, aisvc.ProblemSummary{
					ID:     p.ID,
					Title:  p.Title,
					Tags:   p.Tags,
					Status: "untried",
				})
			}
		}
	}

	utils.OK(c, gin.H{
		"problems":  filtered,
		"tag_stats": tagStats,
	})
}

// SubmitAndJudge handles POST /api/agent/judge (called by agent-service tool executor)
func (h *AIHandler) SubmitAndJudge(c *gin.Context) {
	uid, _ := middleware.CurrentUserID(c)
	if uid == 0 {
		utils.BadRequest(c, "user id required")
		return
	}

	var req struct {
		ProblemID uint64 `json:"problem_id" binding:"required"`
		Code      string `json:"code" binding:"required"`
		Language  string `json:"language" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "invalid request")
		return
	}

	status, err := h.judgeCode(c, req.ProblemID, req.Language, req.Code)
	if err != nil {
		utils.OK(c, gin.H{"error": fmt.Sprintf("judge failed: %v", err), "status": "System Error"})
		return
	}

	utils.OK(c, gin.H{
		"submission_id":   0,
		"status":          status,
		"testcase_passed": 0,
		"testcase_total":  0,
	})
}

// GetUserCode handles POST /api/agent/code (returns latest submission code for a problem)
func (h *AIHandler) GetUserCode(c *gin.Context) {
	uid, _ := middleware.CurrentUserID(c)
	if uid == 0 { utils.BadRequest(c, "user id required"); return }
	var req struct { ProblemID uint64 `json:"problem_id"` }
	if err := c.ShouldBindJSON(&req); err != nil { utils.BadRequest(c, "invalid request"); return }
	var sub models.Submission
	if err := h.DB.Where("user_id = ? AND problem_id = ?", uid, req.ProblemID).Order("id DESC").First(&sub).Error; err != nil {
		utils.OK(c, gin.H{"code": "", "language": "", "found": false})
		return
	}
	utils.OK(c, gin.H{"code": sub.Code, "language": sub.Language, "status": sub.Status, "found": true, "submission_id": sub.ID})
}

// SearchProblems handles POST /api/agent/search-problems (fuzzy search by title)
func (h *AIHandler) SearchProblems(c *gin.Context) {
	var req struct { Query string `json:"query"` }
	if err := c.ShouldBindJSON(&req); err != nil || req.Query == "" { utils.BadRequest(c, "invalid request"); return }
	var problems []models.Problem
	h.DB.Where("LOWER(title) LIKE ? AND status = ?", "%"+strings.ToLower(req.Query)+"%", models.ProblemStatusPublished).Limit(10).Find(&problems)
	result := make([]gin.H, len(problems))
	for i, p := range problems { result[i] = gin.H{"id": p.ID, "title": p.Title, "tags": p.Tags, "difficulty": p.Difficulty} }
	utils.OK(c, gin.H{"problems": result})
}
