package handler

import (
	"context"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/terminaloj/backend/internal/middleware"
	"github.com/terminaloj/backend/internal/models"
	"github.com/terminaloj/backend/internal/mq"
	"github.com/terminaloj/backend/internal/utils"
	"gorm.io/gorm"
)

type SubmissionHandler struct {
	DB     *gorm.DB
	Broker *mq.Broker
}

type submitReq struct {
	ProblemID uint64 `json:"problemId" binding:"required"`
	Language  string `json:"language" binding:"required"`
	Code      string `json:"code" binding:"required"`
}

var allowedLanguages = map[string]bool{"cpp": true, "java": true, "python": true, "go": true}

// idSequence generates monotonically increasing submission ids in the
// 100000+ range to match the frontend mock look-and-feel. A MySQL auto
// increment would also work, but we want the id before enqueueing so the
// client gets an immediate reference. The sequence is thread-safe via
// sync/atomic so concurrent submits do not collide.
var idCounter uint64 = 100000

func nextSubmissionID() uint64 {
	return atomic.AddUint64(&idCounter, 1)
}

func (h *SubmissionHandler) Submit(c *gin.Context) {
	uid, _ := middleware.CurrentUserID(c)
	var req submitReq
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "参数不合法")
		return
	}
	if !allowedLanguages[req.Language] {
		utils.BadRequest(c, "不支持的语言")
		return
	}
	if n := len(req.Code); n == 0 || n > 65536 {
		utils.BadRequest(c, "代码长度不合法")
		return
	}

	var problem models.Problem
	if err := h.DB.First(&problem, req.ProblemID).Error; err != nil {
		utils.NotFound(c, "题目不存在")
		return
	}

	// Initialise the id locally; the full record is persisted by the worker
	// so MySQL writes stay off the request path.
	submissionID := nextSubmissionID()

	task := &mq.SubmitTask{
		SubmissionID: submissionID,
		UserID:       uid,
		ProblemID:    problem.ID,
		ProblemTitle: problem.Title,
		Language:     req.Language,
		Code:         req.Code,
		EnqueuedAt:   time.Now(),
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()
	if err := h.Broker.Publish(ctx, task); err != nil {
		utils.Server(c, "队列发布失败: "+err.Error())
		return
	}
	utils.OK(c, gin.H{
		"id":         submissionID,
		"problemId":  problem.ID,
		"status":     models.StatusPending,
		"language":   req.Language,
		"runtime":    0,
		"memory":     "0.0",
		"createdAt":  time.Now().UTC().Format("2006-01-02T15:04:05.000Z"),
		"codeLength": len(req.Code),
	})
}

func (h *SubmissionHandler) List(c *gin.Context) {
	uid, _ := middleware.CurrentUserID(c)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 20
	}

	q := h.DB.Model(&models.Submission{}).Where("user_id = ?", uid)
	if pid := c.Query("problemId"); pid != "" {
		q = q.Where("problem_id = ?", pid)
	}
	if status := c.Query("status"); status != "" {
		q = q.Where("status = ?", status)
	}
	switch c.DefaultQuery("sortBy", "time") {
	case "problemId":
		q = q.Order("problem_id ASC, id DESC")
	default:
		q = q.Order("id DESC")
	}

	var total int64
	q.Count(&total)
	var rows []models.Submission
	if err := q.Offset((page - 1) * size).Limit(size).Find(&rows).Error; err != nil {
		utils.Server(c, err.Error())
		return
	}
	list := make([]gin.H, 0, len(rows))
	for _, r := range rows {
		list = append(list, submissionView(&r))
	}
	utils.OK(c, gin.H{"list": list, "total": total})
}

func (h *SubmissionHandler) Detail(c *gin.Context) {
	uid, _ := middleware.CurrentUserID(c)
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.BadRequest(c, "提交号不合法")
		return
	}
	var s models.Submission
	if err := h.DB.Where("id = ? AND user_id = ?", id, uid).First(&s).Error; err != nil {
		utils.NotFound(c, "提交不存在")
		return
	}
	utils.OK(c, submissionView(&s))
}

func submissionView(s *models.Submission) gin.H {
	return gin.H{
		"id":           s.ID,
		"problemId":    s.ProblemID,
		"problemTitle": s.ProblemTitle,
		"status":       s.Status,
		"language":     s.Language,
		"runtime":      s.Runtime,
		"memory":       s.Memory,
		"createdAt":    s.CreatedAt.UTC().Format("2006-01-02T15:04:05.000Z"),
		"codeLength":   s.CodeLength,
	}
}
