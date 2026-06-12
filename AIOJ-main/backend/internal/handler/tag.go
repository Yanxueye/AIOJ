package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/terminaloj/backend/internal/models"
	"github.com/terminaloj/backend/internal/utils"
	"gorm.io/gorm"
)

type TagHandler struct {
	DB *gorm.DB
}

// List returns all algorithm tags grouped by category.
func (h *TagHandler) List(c *gin.Context) {
	var tags []models.AlgorithmTag
	if err := h.DB.Order("category ASC, order_no ASC, name ASC").Find(&tags).Error; err != nil {
		utils.Server(c, err.Error())
		return
	}

	// Group by category
	grouped := make(map[string][]gin.H)
	for _, t := range tags {
		grouped[t.Category] = append(grouped[t.Category], gin.H{
			"id":       t.ID,
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

// Names returns just the tag names as a flat list (for AI prompt injection).
func (h *TagHandler) Names(c *gin.Context) {
	var tags []models.AlgorithmTag
	if err := h.DB.Order("category ASC, order_no ASC").Find(&tags).Error; err != nil {
		utils.Server(c, err.Error())
		return
	}

	names := make([]string, len(tags))
	for i, t := range tags {
		names[i] = t.Name
	}

	utils.OK(c, gin.H{"names": names, "total": len(names)})
}
