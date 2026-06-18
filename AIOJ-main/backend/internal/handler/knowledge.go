package handler

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/terminaloj/backend/internal/data"
	"github.com/terminaloj/backend/internal/middleware"
	"github.com/terminaloj/backend/internal/models"
	"github.com/terminaloj/backend/internal/utils"
	"gorm.io/gorm"
)

type KnowledgeHandler struct {
	DB *gorm.DB
}

// List returns all knowledge points, optionally filtered by category.
// Data is served from the hardcoded knowledge tree, not the database.
func (h *KnowledgeHandler) List(c *gin.Context) {
	category := c.Query("category")
	points := data.KnowledgeTree()
	filtered := make([]data.KPNode, 0, len(points))
	for _, p := range points {
		if category != "" && p.Category != category {
			continue
		}
		filtered = append(filtered, p)
	}
	utils.OK(c, gin.H{"items": filtered})
}

// Graph returns knowledge points with their relationships for visualization.
func (h *KnowledgeHandler) Graph(c *gin.Context) {
	points := data.KnowledgeTree()

	// Assign stable IDs based on index
	type Node struct {
		ID       int     `json:"id"`
		Name     string  `json:"name"`
		Category string  `json:"category"`
		ParentID *int    `json:"parentId,omitempty"`
		Color    string  `json:"color,omitempty"`
		Icon     string  `json:"icon,omitempty"`
	}
	nameToID := make(map[string]int, len(points))
	nodes := make([]Node, len(points))
	for i, p := range points {
		nameToID[p.Name] = i+1
		nodes[i] = Node{
			ID:       i+1,
			Name:     p.Name,
			Category: p.Category,
			Color:    p.Color,
			Icon:     p.Icon,
		}
		if p.ParentName != "" {
			if pid, ok := nameToID[p.ParentName]; ok {
				pidCopy := pid
				nodes[i].ParentID = &pidCopy
			}
		}
	}

	// Build edges from parent-child relationships
	type Edge struct {
		Source int `json:"source"`
		Target int `json:"target"`
	}
	var edges []Edge
	for _, n := range nodes {
		if n.ParentID != nil {
			edges = append(edges, Edge{Source: *n.ParentID, Target: n.ID})
		}
	}

	// Count problems per knowledge point by tag name
	countMap := make(map[int]int)
	for _, p := range points {
		var cnt int64
		h.DB.Model(&models.Problem{}).
			Where("status = ? AND JSON_CONTAINS(tags, ?)", models.ProblemStatusPublished,
				fmt.Sprintf(`"%s"`, p.Name)).Count(&cnt)
		countMap[nameToID[p.Name]] = int(cnt)
	}

	// User mastery computed in real-time
	uid, _ := middleware.CurrentUserID(c)
	masteryMap := make(map[int]float64)
	if uid > 0 {
		var triedPIDs []uint64
		h.DB.Model(&models.Submission{}).
			Where("user_id = ?", uid).
			Distinct("problem_id").
			Pluck("problem_id", &triedPIDs)
		triedSet := make(map[uint64]bool, len(triedPIDs))
		for _, id := range triedPIDs {
			triedSet[id] = true
		}

		for _, p := range points {
			total := countMap[nameToID[p.Name]]
			if total == 0 {
				continue
			}
			var tagProblems []models.Problem
			h.DB.Where("status = ? AND JSON_CONTAINS(tags, ?)", models.ProblemStatusPublished,
				fmt.Sprintf(`"%s"`, p.Name)).Find(&tagProblems)
			tried := 0
			for _, prob := range tagProblems {
				if triedSet[prob.ID] {
					tried++
				}
			}
			if tried > 0 {
				masteryMap[nameToID[p.Name]] = float64(tried) / float64(total) * 100
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
// split into untried and tried sections. Matching by tag name.
func (h *KnowledgeHandler) ProblemsForKP(c *gin.Context) {
	idStr := c.Param("id")
	kpID, err := strconv.Atoi(idStr)
	if err != nil {
		utils.BadRequest(c, "知识点ID不合法")
		return
	}

	points := data.KnowledgeTree()
	if kpID < 1 || kpID > len(points) {
		utils.NotFound(c, "知识点不存在")
		return
	}
	kpID--
	kp := points[kpID]

	// Find published problems tagged with this knowledge point name
	var problems []models.Problem
	h.DB.Where("status = ? AND JSON_CONTAINS(tags, ?)", models.ProblemStatusPublished,
		fmt.Sprintf(`"%s"`, kp.Name)).Find(&problems)

	uid, _ := middleware.CurrentUserID(c)
	untried := make([]models.Problem, 0)
	tried := make([]models.Problem, 0)

	if uid > 0 && len(problems) > 0 {
		pids := make([]uint64, len(problems))
		for i, p := range problems {
			pids[i] = p.ID
		}
		var triedPIDs []uint64
		h.DB.Model(&models.Submission{}).
			Where("user_id = ? AND problem_id IN ?", uid, pids).
			Distinct("problem_id").
			Pluck("problem_id", &triedPIDs)
		triedIDs := make(map[uint64]bool, len(triedPIDs))
		for _, pid := range triedPIDs {
			triedIDs[pid] = true
		}
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

// ProblemsByTags returns problems matching multiple knowledge point names.
type problemsByTagsReq struct {
	Tags        []string `json:"tags" binding:"required"`
	OnlyUntried bool     `json:"onlyUntried"`
}

// ProblemsByTags returns problems matching multiple tags, split into untried/tried.
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
		for i, p := range allProblems {
			pids[i] = p.ID
		}
		var triedPIDs []uint64
		h.DB.Model(&models.Submission{}).
			Where("user_id = ? AND problem_id IN ?", uid, pids).
			Distinct("problem_id").Pluck("problem_id", &triedPIDs)
		triedSet := make(map[uint64]bool, len(triedPIDs))
		for _, pid := range triedPIDs {
			triedSet[pid] = true
		}
		for _, p := range allProblems {
			if triedSet[p.ID] {
				tried = append(tried, p)
			} else {
				untried = append(untried, p)
			}
		}
	} else {
		untried = allProblems
	}
	utils.OK(c, gin.H{"untried": untried, "tried": tried})
}

// TagHandler serves algorithm tags from the hardcoded knowledge tree.
type TagHandler struct{}

// List returns all algorithm tags grouped by category.
func (h *TagHandler) List(c *gin.Context) {
	tags := data.Tags()
	grouped := make(map[string][]gin.H)
	for _, t := range tags {
		grouped[t.Category] = append(grouped[t.Category], gin.H{
			"name":     t.Name,
			"category": t.Category,
			"parent":   t.Parent,
		})
	}
	categories := make([]gin.H, 0, len(grouped))
	for cat, items := range grouped {
		categories = append(categories, gin.H{
			"category": cat,
			"tags":     items,
		})
	}
	utils.OK(c, gin.H{"categories": categories, "total": len(tags)})
}

// Names returns just the tag names as a flat list.
func (h *TagHandler) Names(c *gin.Context) {
	names := data.TagNames()
	// Filter out category-only tags (those ending with "（分类）")
	leafNames := make([]string, 0, len(names))
	for _, n := range names {
		if !strings.HasSuffix(n, "（分类）") {
			leafNames = append(leafNames, n)
		}
	}
	utils.OK(c, gin.H{"names": leafNames, "total": len(leafNames)})
}
