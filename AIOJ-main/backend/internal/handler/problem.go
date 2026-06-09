package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/terminaloj/backend/internal/middleware"
	"github.com/terminaloj/backend/internal/models"
	"github.com/terminaloj/backend/internal/utils"
	"gorm.io/gorm"
)

type ProblemHandler struct {
	DB *gorm.DB
}

type problemListItem struct {
	ID              uint64             `json:"id"`
	Title           string             `json:"title"`
	Difficulty      string             `json:"difficulty"`
	DifficultyScore int                `json:"difficultyScore"`
	Tags            models.StringSlice `json:"tags"`
	AcceptRate      string             `json:"acceptRate"`
	SubmitCount     int                `json:"submitCount"`
	Accepted        bool               `json:"accepted"`
}

func (h *ProblemHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 20
	}
	keyword := c.Query("keyword")
	difficulty := c.Query("difficulty")
	tag := c.Query("tag")

	q := h.DB.Model(&models.Problem{})
	if keyword != "" {
		like := "%" + keyword + "%"
		if id, err := strconv.ParseUint(keyword, 10, 64); err == nil {
			q = q.Where("id = ? OR title LIKE ?", id, like)
		} else {
			q = q.Where("title LIKE ?", like)
		}
	}
	if difficulty != "" {
		q = q.Where("difficulty = ?", difficulty)
	}
	if tag != "" {
		q = q.Where("JSON_CONTAINS(tags, JSON_QUOTE(?))", tag)
	}

	var total int64
	q.Count(&total)

	var rows []models.Problem
	if err := q.Order("id ASC").Offset((page - 1) * size).Limit(size).Find(&rows).Error; err != nil {
		utils.Server(c, err.Error())
		return
	}

	uid, logged := middleware.CurrentUserID(c)
	acceptedSet := map[uint64]bool{}
	if logged && len(rows) > 0 {
		ids := make([]uint64, 0, len(rows))
		for _, p := range rows {
			ids = append(ids, p.ID)
		}
		var solved []uint64
		h.DB.Model(&models.Submission{}).
			Where("user_id = ? AND status = ? AND problem_id IN ?", uid, models.StatusAccepted, ids).
			Distinct("problem_id").Pluck("problem_id", &solved)
		for _, id := range solved {
			acceptedSet[id] = true
		}
	}

	list := make([]problemListItem, 0, len(rows))
	for _, p := range rows {
		list = append(list, problemListItem{
			ID:              p.ID,
			Title:           p.Title,
			Difficulty:      p.Difficulty,
			DifficultyScore: p.DifficultyScore,
			Tags:            p.Tags,
			AcceptRate:      p.AcceptRate(),
			SubmitCount:     p.SubmitCount,
			Accepted:        acceptedSet[p.ID],
		})
	}
	utils.OK(c, gin.H{"list": list, "total": total})
}

func (h *ProblemHandler) Detail(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.BadRequest(c, "题号不合法")
		return
	}
	var p models.Problem
	if err := h.DB.First(&p, id).Error; err != nil {
		utils.NotFound(c, "题目不存在")
		return
	}
	accepted := false
	if uid, ok := middleware.CurrentUserID(c); ok {
		var cnt int64
		h.DB.Model(&models.Submission{}).
			Where("user_id = ? AND problem_id = ? AND status = ?", uid, p.ID, models.StatusAccepted).
			Count(&cnt)
		accepted = cnt > 0
	}
	utils.OK(c, gin.H{
		"id":              p.ID,
		"title":           p.Title,
		"difficulty":      p.Difficulty,
		"difficultyScore": p.DifficultyScore,
		"tags":            p.Tags,
		"acceptRate":      p.AcceptRate(),
		"submitCount":     p.SubmitCount,
		"accepted":        accepted,
		"content":         p.Content,
		"timeLimit":       p.TimeLimit,
		"memoryLimit":     p.MemoryLimit,
		"source":          p.Source,
	})
}

type AnnouncementHandler struct {
	DB *gorm.DB
}

func (h *AnnouncementHandler) List(c *gin.Context) {
	var rows []models.Announcement
	if err := h.DB.Order("id DESC").Limit(20).Find(&rows).Error; err != nil {
		utils.Server(c, err.Error())
		return
	}
	utils.OK(c, rows)
}
