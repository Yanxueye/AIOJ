package handler

import (
	"fmt"
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
	q := h.DB.Preload("Items", func(db *gorm.DB) *gorm.DB { return db.Order("order_no ASC") })
	if kw := c.Query("q"); kw != "" {
		q = q.Where("title LIKE ?", "%"+kw+"%")
	}
	if err := q.Order("created_at DESC").Find(&plans).Error; err != nil {
		utils.Server(c, err.Error())
		return
	}

	uid, logged := middleware.CurrentUserID(c)
	progressMap := map[uint64]models.UserPlanProgress{}
	favSet := map[uint64]bool{}
	userNames := map[uint64]string{}
	if logged {
		var progresses []models.UserPlanProgress
		h.DB.Where("user_id = ?", uid).Find(&progresses)
		for _, p := range progresses { progressMap[p.PlanID] = p }
		var favs []models.StudyPlanFavorite
		h.DB.Where("user_id = ?", uid).Find(&favs)
		for _, f := range favs { favSet[f.PlanID] = true }
	}
	// Collect user IDs for display names
	for _, plan := range plans {
		if plan.UserID > 0 { userNames[plan.UserID] = "" }
	}
	for uid := range userNames {
		var u models.User
		if h.DB.Select("username").First(&u, uid).Error == nil {
			userNames[uid] = u.Username
		}
	}

	items := make([]gin.H, 0, len(plans))
	for _, plan := range plans {
		progress := progressMap[plan.ID]
		items = append(items, gin.H{
			"id":             plan.ID,
			"title":          plan.Title,
			"description":    plan.Description,
			"difficulty":     plan.Difficulty,
			"tags":           plan.Tags,
			"userId":         plan.UserID,
			"username":       userNames[plan.UserID],
			"isOwner":        logged && plan.UserID == uid,
			"isFavorited":    favSet[plan.ID],
			"problemCount":   len(plan.Items),
			"completedCount": progress.CompletedCount,
			"createdAt":      plan.CreatedAt.Format("2006-01-02"),
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
	uid, logged := middleware.CurrentUserID(c)
	var username string
	if plan.UserID > 0 {
		var u models.User
		if h.DB.Select("username").First(&u, plan.UserID).Error == nil { username = u.Username }
	}
	var favCount int64
	h.DB.Model(&models.StudyPlanFavorite{}).Where("plan_id = ?", plan.ID).Count(&favCount)
	favorited := false
	if logged {
		var f models.StudyPlanFavorite
		if h.DB.Where("user_id = ? AND plan_id = ?", uid, plan.ID).First(&f).Error == nil { favorited = true }
	}

	var progress models.UserPlanProgress
	progressItems := map[uint64]models.UserPlanProgressItem{}
	if logged {
		h.DB.Where("user_id = ? AND plan_id = ?", uid, plan.ID).First(&progress)
		var rows []models.UserPlanProgressItem
		if h.DB.Where("user_id = ? AND plan_id = ?", uid, plan.ID).Find(&rows).Error == nil {
			for _, row := range rows { progressItems[row.ProblemID] = row }
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
		"id": plan.ID, "title": plan.Title, "description": plan.Description,
		"difficulty": plan.Difficulty, "tags": plan.Tags,
		"userId": plan.UserID, "username": username,
		"isOwner": logged && plan.UserID == uid,
		"isFavorited": favorited, "favoriteCount": favCount,
		"items": items, "completedCount": progress.CompletedCount,
		"lastCompletedAt": formatStudyPlanTime(progress.LastCompletedAt),
	})
}

func (h *StudyPlanHandler) DailyChallenge(c *gin.Context) {
	var item models.DailyChallenge
	date := c.DefaultQuery("date", time.Now().Format("2006-01-02"))
	if err := h.DB.Where("date = ?", date).First(&item).Error; err != nil {
		if err := h.DB.Order("date DESC").First(&item).Error; err != nil {
			utils.NotFound(c, "每日一题不存在")
			return
		}
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

type createPlanReq struct {
	Title       string             `json:"title" binding:"required"`
	Description string             `json:"description"`
	Difficulty  string             `json:"difficulty"`
	Tags        models.StringSlice `json:"tags"`
	ProblemIDs  []uint64           `json:"problemIDs" binding:"required"`
}

func (h *StudyPlanHandler) Create(c *gin.Context) {
	uid, _ := middleware.CurrentUserID(c)
	if uid == 0 { utils.Unauthorized(c, "请先登录"); return }
	var req createPlanReq
	if err := c.ShouldBindJSON(&req); err != nil { utils.BadRequest(c, "参数不合法"); return }
	var count int64
	h.DB.Model(&models.Problem{}).Where("id IN ?", req.ProblemIDs).Count(&count)
	if int(count) != len(req.ProblemIDs) { utils.BadRequest(c, "部分题目ID不存在"); return }
	plan := models.StudyPlan{
		UserID: uid, Title: req.Title, Description: req.Description,
		Difficulty: req.Difficulty, Tags: req.Tags,
	}
	if err := h.DB.Create(&plan).Error; err != nil { utils.Server(c, err.Error()); return }
	for i, pid := range req.ProblemIDs {
		var p models.Problem
		if err := h.DB.First(&p, pid).Error; err != nil { continue }
		h.DB.Create(&models.StudyPlanItem{
			PlanID: plan.ID, ProblemID: pid, OrderNo: i + 1, Title: p.Title, Difficulty: p.Difficulty,
		})
	}
	utils.OK(c, gin.H{"id": plan.ID})
}

func (h *StudyPlanHandler) Update(c *gin.Context) {
	uid, _ := middleware.CurrentUserID(c)
	if uid == 0 { utils.Unauthorized(c, "请先登录"); return }
	var plan models.StudyPlan
	if err := h.DB.First(&plan, c.Param("id")).Error; err != nil { utils.NotFound(c, "题单不存在"); return }
	if plan.UserID != uid { utils.Forbidden(c, "只能修改自己的题单"); return }
	var req createPlanReq
	if err := c.ShouldBindJSON(&req); err != nil { utils.BadRequest(c, "参数不合法"); return }
	plan.Title = req.Title
	plan.Description = req.Description
	plan.Difficulty = req.Difficulty
	plan.Tags = req.Tags
	h.DB.Save(&plan)
	// Replace items
	h.DB.Where("plan_id = ?", plan.ID).Delete(&models.StudyPlanItem{})
	for i, pid := range req.ProblemIDs {
		var p models.Problem
		if err := h.DB.First(&p, pid).Error; err != nil { continue }
		h.DB.Create(&models.StudyPlanItem{
			PlanID: plan.ID, ProblemID: pid, OrderNo: i + 1, Title: p.Title, Difficulty: p.Difficulty,
		})
	}
	utils.OK(c, gin.H{"id": plan.ID})
}

func (h *StudyPlanHandler) Delete(c *gin.Context) {
	uid, _ := middleware.CurrentUserID(c)
	var plan models.StudyPlan
	if err := h.DB.First(&plan, c.Param("id")).Error; err != nil { utils.NotFound(c, "题单不存在"); return }
	if plan.UserID != uid { utils.Forbidden(c, "只能删除自己的题单"); return }
	h.DB.Where("plan_id = ?", plan.ID).Delete(&models.StudyPlanItem{})
	h.DB.Where("plan_id = ?", plan.ID).Delete(&models.UserPlanProgressItem{})
	h.DB.Where("plan_id = ?", plan.ID).Delete(&models.UserPlanProgress{})
	h.DB.Where("plan_id = ?", plan.ID).Delete(&models.StudyPlanFavorite{})
	h.DB.Delete(&plan)
	utils.OK(c, nil)
}

func (h *StudyPlanHandler) Favorite(c *gin.Context) {
	uid, _ := middleware.CurrentUserID(c)
	if uid == 0 { utils.Unauthorized(c, "请先登录"); return }
	planID, err := parseUintParam(c, "id")
	if err != nil { utils.BadRequest(c, "参数不合法"); return }
	var plan models.StudyPlan
	if h.DB.First(&plan, planID).Error != nil { utils.NotFound(c, "题单不存在"); return }
	// Toggle: if exists, unfavorite; else favorite
	var fav models.StudyPlanFavorite
	if h.DB.Where("user_id = ? AND plan_id = ?", uid, planID).First(&fav).Error == nil {
		h.DB.Delete(&fav)
		utils.OK(c, gin.H{"favorited": false})
	} else {
		h.DB.Create(&models.StudyPlanFavorite{UserID: uid, PlanID: planID})
		utils.OK(c, gin.H{"favorited": true})
	}
}

func formatStudyPlanTime(t *time.Time) string {
	if t == nil { return "" }
	return t.Format("2006-01-02 15:04")
}

func parseUintParam(c *gin.Context, name string) (uint64, error) {
	id := c.Param(name)
	var v uint64
	_, err := fmt.Sscanf(id, "%d", &v)
	return v, err
}
