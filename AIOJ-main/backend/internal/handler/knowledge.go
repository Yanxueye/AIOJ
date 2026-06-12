package handler

import (
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

	// Count problems per knowledge point
	type ProblemCount struct {
		KnowledgePointID uint64
		Count            int
	}
	var counts []ProblemCount
	if err := h.DB.Model(&models.ProblemKnowledgePoint{}).
		Select("knowledge_point_id, COUNT(*) as count").
		Group("knowledge_point_id").Scan(&counts).Error; err != nil {
		log.Printf("[knowledge] problem count query failed: %v", err)
	}
	countMap := make(map[uint64]int)
	for _, pc := range counts {
		countMap[pc.KnowledgePointID] = pc.Count
	}

	// User mastery if logged in
	uid, _ := middleware.CurrentUserID(c)
	masteryMap := make(map[uint64]float64)
	if uid > 0 {
		var masteries []models.UserKnowledgeMastery
		if err := h.DB.Where("user_id = ?", uid).Find(&masteries).Error; err != nil {
			log.Printf("[knowledge] mastery query for uid=%d failed: %v", uid, err)
		}
		for _, m := range masteries {
			masteryMap[m.KnowledgePointID] = m.MasteryLevel
		}
	}

	utils.OK(c, gin.H{
		"nodes":   nodes,
		"edges":   edges,
		"counts":  countMap,
		"mastery": masteryMap,
	})
}

// ProblemsForKP returns problems associated with a knowledge point.
func (h *KnowledgeHandler) ProblemsForKP(c *gin.Context) {
	kpID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.BadRequest(c, "知识点ID不合法")
		return
	}

	var mappings []models.ProblemKnowledgePoint
	if err := h.DB.Where("knowledge_point_id = ?", kpID).Find(&mappings).Error; err != nil {
		log.Printf("[knowledge] mapping query for kp=%d failed: %v", kpID, err)
	}
	problemIDs := make([]uint64, len(mappings))
	for i, m := range mappings {
		problemIDs[i] = m.ProblemID
	}

	var problems []models.Problem
	if len(problemIDs) > 0 {
		if err := h.DB.Where("id IN ? AND status = ?", problemIDs, models.ProblemStatusPublished).Find(&problems).Error; err != nil {
			log.Printf("[knowledge] problems query for kp=%d failed: %v", kpID, err)
		}
	}

	utils.OK(c, gin.H{"items": problems})
}
