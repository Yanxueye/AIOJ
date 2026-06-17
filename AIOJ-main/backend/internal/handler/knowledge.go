package handler

import (
	"fmt"
	"log"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/terminaloj/backend/internal/middleware"
	"github.com/terminaloj/backend/internal/models"
	"github.com/terminaloj/backend/internal/utils"
	"gorm.io/gorm"
)

type KnowledgeHandler struct {
	DB *gorm.DB
}

// List returns all knowledge points, optionally filtered by category.
func (h *KnowledgeHandler) List(c *gin.Context) {
	category := c.Query("category")
	var points []models.KnowledgePoint
	q := h.DB
	if category != "" {
		q = q.Where("category = ?", category)
	}
	if err := q.Order("category, name").Find(&points).Error; err != nil {
		log.Printf("[knowledge] list query failed: %v", err)
	}
	utils.OK(c, gin.H{"items": points})
}

// Graph returns knowledge points with their relationships for visualization.
func (h *KnowledgeHandler) Graph(c *gin.Context) {
	var points []models.KnowledgePoint
	if err := h.DB.Find(&points).Error; err != nil {
		log.Printf("[knowledge] graph query failed: %v", err)
	}

	// Build nodes
	type Node struct {
		ID       uint64  `json:"id"`
		Name     string  `json:"name"`
		Category string  `json:"category"`
		ParentID *uint64 `json:"parentId,omitempty"`
		Color    string  `json:"color,omitempty"`
		Icon     string  `json:"icon,omitempty"`
	}
	nodes := make([]Node, len(points))
	for i, p := range points {
		nodes[i] = Node{
			ID:       p.ID,
			Name:     p.Name,
			Category: p.Category,
			ParentID: p.ParentID,
			Color:    p.Color,
			Icon:     p.Icon,
		}
	}

	// Build edges (parent-child relationships)
	type Edge struct {
		Source uint64 `json:"source"`
		Target uint64 `json:"target"`
	}
	var edges []Edge
	for _, p := range points {
		if p.ParentID != nil {
			edges = append(edges, Edge{Source: *p.ParentID, Target: p.ID})
		}
	}

	// Count problems per knowledge point — match by tag name
	countMap := make(map[uint64]int)
	for i := range points {
		kp := &points[i]
		var cnt int64
		h.DB.Model(&models.Problem{}).
			Where("status = ? AND JSON_CONTAINS(tags, ?)", models.ProblemStatusPublished,
				fmt.Sprintf(`"%s"`, kp.Name)).Count(&cnt)
		countMap[kp.ID] = int(cnt)
	}

	// User mastery computed in real-time: for each KP, tried/total
	uid, _ := middleware.CurrentUserID(c)
	masteryMap := make(map[uint64]float64)
	if uid > 0 {
		var triedPIDs []uint64
		h.DB.Model(&models.Submission{}).
			Where("user_id = ?", uid).
			Distinct("problem_id").
			Pluck("problem_id", &triedPIDs)
		triedSet := make(map[uint64]bool, len(triedPIDs))
		for _, id := range triedPIDs { triedSet[id] = true }

		for i := range points {
			kp := &points[i]
			total := countMap[kp.ID]
			if total == 0 { continue }

			var tagProblems []models.Problem
			h.DB.Where("status = ? AND JSON_CONTAINS(tags, ?)", models.ProblemStatusPublished,
				fmt.Sprintf(`"%s"`, kp.Name)).Find(&tagProblems)
			tried := 0
			for _, p := range tagProblems {
				if triedSet[p.ID] { tried++ }
			}
			if tried > 0 {
				masteryMap[kp.ID] = float64(tried) / float64(total) * 100
			}
		}
	}

	utils.OK(c, gin.H{
		"nodes":   nodes,
		"edges":   edges,
		"counts":  countMap,
		"mastery": masteryMap,
	})
}

// ProblemsForKP returns problems associated with a knowledge point,
// split into untried (recommended) and tried sections.
// Matching is done by tag name: the knowledge point name is used to
// search problems whose Tags field contains that name (JSON_CONTAINS).
func (h *KnowledgeHandler) ProblemsForKP(c *gin.Context) {
	kpID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.BadRequest(c, "知识点ID不合法")
		return
	}

	// Get knowledge point name
	var kp models.KnowledgePoint
	if err := h.DB.First(&kp, kpID).Error; err != nil {
		utils.NotFound(c, "知识点不存在")
		return
	}

	// Find published problems tagged with this knowledge point name
	var problems []models.Problem
	h.DB.Where("status = ? AND JSON_CONTAINS(tags, ?)", models.ProblemStatusPublished,
		fmt.Sprintf(`"%s"`, kp.Name)).Find(&problems)

	// Separate untried (recommended) vs tried
	uid, _ := middleware.CurrentUserID(c)
	untried := make([]models.Problem, 0)
	tried := make([]models.Problem, 0)

	if uid > 0 && len(problems) > 0 {
		pids := make([]uint64, len(problems))
		for i, p := range problems { pids[i] = p.ID }
		var triedPIDs []uint64
		h.DB.Model(&models.Submission{}).
			Where("user_id = ? AND problem_id IN ?", uid, pids).
			Distinct("problem_id").
			Pluck("problem_id", &triedPIDs)
		triedIDs := make(map[uint64]bool, len(triedPIDs))
		for _, pid := range triedPIDs { triedIDs[pid] = true }
		for _, p := range problems {
			if triedIDs[p.ID] {
				tried = append(tried, p)
			} else {
				untried = append(untried, p)
			}
		}
	} else {
		untried = problems
	}

	utils.OK(c, gin.H{"untried": untried, "tried": tried})
}

// ProblemsByTags returns problems matching multiple knowledge point names,
// split into untried and tried. Used by agent-service for AI 题单创建.
type problemsByTagsReq struct {
	Tags        []string `json:"tags" binding:"required"`
	OnlyUntried bool     `json:"onlyUntried"`
}

func (h *KnowledgeHandler) ProblemsByTags(c *gin.Context) {
	var req problemsByTagsReq
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "参数不合法")
		return
	}
	var allProblems []models.Problem
	seen := make(map[uint64]bool)
	for _, tag := range req.Tags {
		var batch []models.Problem
		h.DB.Where("status = ? AND JSON_CONTAINS(tags, ?)", models.ProblemStatusPublished,
			fmt.Sprintf(`"%s"`, tag)).Find(&batch)
		for _, p := range batch {
			if !seen[p.ID] {
				seen[p.ID] = true
				allProblems = append(allProblems, p)
			}
		}
	}
	untried := make([]models.Problem, 0)
	tried := make([]models.Problem, 0)
	uid, _ := middleware.CurrentUserID(c)
	if uid > 0 && len(allProblems) > 0 {
		pids := make([]uint64, len(allProblems))
		for i, p := range allProblems { pids[i] = p.ID }
		var triedPIDs []uint64
		h.DB.Model(&models.Submission{}).
			Where("user_id = ? AND problem_id IN ?", uid, pids).
			Distinct("problem_id").Pluck("problem_id", &triedPIDs)
		triedSet := make(map[uint64]bool, len(triedPIDs))
		for _, pid := range triedPIDs { triedSet[pid] = true }
		for _, p := range allProblems {
			if triedSet[p.ID] { tried = append(tried, p) } else { untried = append(untried, p) }
		}
	} else {
		untried = allProblems
	}
	utils.OK(c, gin.H{"untried": untried, "tried": tried})
}
