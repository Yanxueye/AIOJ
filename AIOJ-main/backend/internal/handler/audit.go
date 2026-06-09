package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/terminaloj/backend/internal/models"
	"github.com/terminaloj/backend/internal/utils"
	"gorm.io/gorm"
)

type AuditHandler struct {
	DB *gorm.DB
}

func (h *AuditHandler) List(c *gin.Context) {
	q := h.DB.Model(&models.AuditLog{})
	if action := c.Query("action"); action != "" {
		q = q.Where("action = ?", action)
	}
	if resourceType := c.Query("resourceType"); resourceType != "" {
		q = q.Where("resource_type = ?", resourceType)
	}
	if username := c.Query("username"); username != "" {
		q = q.Where("username = ?", username)
	}

	var rows []models.AuditLog
	if err := q.Order("id DESC").Limit(200).Find(&rows).Error; err != nil {
		utils.Server(c, err.Error())
		return
	}
	utils.OK(c, gin.H{"items": rows})
}
