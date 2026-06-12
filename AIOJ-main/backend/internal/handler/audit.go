package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/terminaloj/backend/internal/models"
	"github.com/terminaloj/backend/internal/utils"
	"gorm.io/gorm"
)

type AuditHandler struct {
	DB *gorm.DB
}

func (h *AuditHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 20
	}

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

	var total int64
	q.Count(&total)

	var rows []models.AuditLog
	if err := q.Order("id DESC").Offset((page - 1) * size).Limit(size).Find(&rows).Error; err != nil {
		utils.Server(c, err.Error())
		return
	}
	utils.OK(c, gin.H{"items": rows, "total": total, "page": page, "pageSize": size})
}
