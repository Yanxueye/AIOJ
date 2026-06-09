package handler

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/terminaloj/backend/internal/middleware"
	"github.com/terminaloj/backend/internal/models"
	"github.com/terminaloj/backend/internal/utils"
	"gorm.io/gorm"
)

type StudyPlanHandler struct {
	DB *gorm.DB
}

func (h *StudyPlanHandler) List(c *gin.Context) {
	var plans []models.StudyPlan
	if err := h.DB.Preload("Items", func(db *gorm.DB) *gorm.DB { return db.Order("order_no ASC") }).Find(&plans).Error; err != nil {
		utils.Server(c, err.Error())
		return
	}

	uid, logged := middleware.CurrentUserID(c)
	progressMap := map[uint64]models.UserPlanProgress{}
	if logged {
		var progresses []models.UserPlanProgress
		if err := h.DB.Where("user_id = ?", uid).Find(&progresses).Error; err == nil {
			for _, item := range progresses {
				progressMap[item.PlanID] = item
			}
		}
	}

	items := make([]gin.H, 0, len(plans))
	for _, item := range plans {
		progress := progressMap[item.ID]
		items = append(items, gin.H{
			"id":             item.ID,
			"title":          item.Title,
			"description":    item.Description,
			"difficulty":     item.Difficulty,
			"tags":           item.Tags,
			"problemCount":   len(item.Items),
			"completedCount": progress.CompletedCount,
		})
	}
	utils.OK(c, gin.H{"items": items})
}

func (h *StudyPlanHandler) Detail(c *gin.Context) {
	var plan models.StudyPlan
	if err := h.DB.Preload("Items", func(db *gorm.DB) *gorm.DB { return db.Order("order_no ASC") }).First(&plan, c.Param("id")).Error; err != nil {
		utils.NotFound(c, "题单不存在")
		return
	}

	var progress models.UserPlanProgress
	uid, logged := middleware.CurrentUserID(c)
	progressItems := map[uint64]models.UserPlanProgressItem{}
	if logged {
		_ = h.DB.Where("user_id = ? AND plan_id = ?", uid, plan.ID).First(&progress).Error
		var rows []models.UserPlanProgressItem
		if err := h.DB.Where("user_id = ? AND plan_id = ?", uid, plan.ID).Find(&rows).Error; err == nil {
			for _, item := range rows {
				progressItems[item.ProblemID] = item
			}
		}
	}

	items := make([]gin.H, 0, len(plan.Items))
	for _, item := range plan.Items {
		progressItem := progressItems[item.ProblemID]
		items = append(items, gin.H{
			"id":          item.ID,
			"problemId":   item.ProblemID,
			"orderNo":     item.OrderNo,
			"title":       item.Title,
			"difficulty":  item.Difficulty,
			"completed":   progressItem.Completed,
			"completedAt": formatStudyPlanTime(progressItem.CompletedAt),
		})
	}

	utils.OK(c, gin.H{
		"id":             plan.ID,
		"title":          plan.Title,
		"description":    plan.Description,
		"difficulty":     plan.Difficulty,
		"tags":           plan.Tags,
		"items":          items,
		"completedCount": progress.CompletedCount,
		"lastCompletedAt": formatStudyPlanTime(progress.LastCompletedAt),
	})
}

func (h *StudyPlanHandler) DailyChallenge(c *gin.Context) {
	var item models.DailyChallenge
	date := c.DefaultQuery("date", "2026-06-09")
	if err := h.DB.Where("date = ?", date).First(&item).Error; err != nil {
		utils.NotFound(c, "每日一题不存在")
		return
	}
	utils.OK(c, item)
}

func (h *StudyPlanHandler) Checkins(c *gin.Context) {
	uid, _ := middleware.CurrentUserID(c)
	var rows []models.StudyCheckin
	if err := h.DB.Where("user_id = ?", uid).Order("date DESC").Limit(30).Find(&rows).Error; err != nil {
		utils.Server(c, err.Error())
		return
	}
	utils.OK(c, gin.H{"items": rows})
}

func formatStudyPlanTime(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format("2006-01-02 15:04")
}
