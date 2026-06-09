package handler

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/terminaloj/backend/internal/judger"
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

var (
	idCounter uint64 = 100000
	idOnce    sync.Once
)

func (h *SubmissionHandler) nextSubmissionID() (uint64, error) {
	var initErr error
	idOnce.Do(func() {
		var maxID uint64
		if err := h.DB.Model(&models.Submission{}).Select("COALESCE(MAX(id), 100000)").Scan(&maxID).Error; err != nil {
			initErr = err
			return
		}
		if maxID > atomic.LoadUint64(&idCounter) {
			atomic.StoreUint64(&idCounter, maxID)
		}
	})
	if initErr != nil {
		return 0, initErr
	}
	return atomic.AddUint64(&idCounter, 1), nil
}

func (h *SubmissionHandler) Submit(c *gin.Context) {
	uid, _ := middleware.CurrentUserID(c)

	var req submitReq
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "参数不合法")
		return
	}

	normalizedLang, ok := judger.NormalizeLanguage(req.Language)
	if !ok {
		utils.BadRequest(c, "暂不支持该语言")
		return
	}
	req.Language = normalizedLang

	if n := len(req.Code); n == 0 || n > 65536 {
		utils.BadRequest(c, "代码长度不合法")
		return
	}

	var problem models.Problem
	if err := h.DB.First(&problem, req.ProblemID).Error; err != nil {
		utils.NotFound(c, "题目不存在")
		return
	}

	submissionID, err := h.nextSubmissionID()
	if err != nil {
		utils.Server(c, "生成提交编号失败: "+err.Error())
		return
	}
	now := time.Now().UTC()
	traceID := fmt.Sprintf("judge-%d-%d", problem.ID, submissionID)

	task := &mq.SubmitTask{
		SubmissionID: submissionID,
		UserID:       uid,
		ProblemID:    problem.ID,
		ProblemTitle: problem.Title,
		TraceID:      traceID,
		Language:     req.Language,
		Code:         req.Code,
		EnqueuedAt:   now,
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()
	if err := h.Broker.Publish(ctx, task); err != nil {
		utils.Server(c, "队列发布失败: "+err.Error())
		return
	}

	utils.OK(c, gin.H{
		"id":            submissionID,
		"problemId":     problem.ID,
		"traceId":       traceID,
		"status":        models.StatusPending,
		"language":      req.Language,
		"runtime":       0,
		"runtimeMs":     0,
		"memory":        "0.0",
		"memoryKb":      0,
		"compileOutput": "",
		"errorMessage":  "",
		"caseResults":   []gin.H{},
		"createdAt":     now.Format("2006-01-02T15:04:05.000Z"),
		"updatedAt":     now.Format("2006-01-02T15:04:05.000Z"),
		"codeLength":    len(req.Code),
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
	for i := range rows {
		list = append(list, submissionView(&rows[i], false))
	}
	utils.OK(c, gin.H{"list": list, "total": total})
}

func (h *SubmissionHandler) Detail(c *gin.Context) {
	uid, _ := middleware.CurrentUserID(c)
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.BadRequest(c, "提交编号不合法")
		return
	}

	var s models.Submission
	if err := h.DB.Preload("CaseResults", func(db *gorm.DB) *gorm.DB {
		return db.Order("case_no ASC")
	}).Where("id = ? AND user_id = ?", id, uid).First(&s).Error; err != nil {
		utils.NotFound(c, "提交不存在")
		return
	}
	utils.OK(c, submissionView(&s, true))
}

func submissionView(s *models.Submission, withCases bool) gin.H {
	view := gin.H{
		"id":            s.ID,
		"problemId":     s.ProblemID,
		"problemTitle":  s.ProblemTitle,
		"traceId":       s.TraceID,
		"status":        s.Status,
		"language":      s.Language,
		"runtime":       s.Runtime,
		"runtimeMs":     s.RuntimeMS,
		"memory":        s.Memory,
		"memoryKb":      s.MemoryKB,
		"compileOutput": s.CompileOutput,
		"errorMessage":  s.ErrorMessage,
		"createdAt":     s.CreatedAt.UTC().Format("2006-01-02T15:04:05.000Z"),
		"updatedAt":     s.UpdatedAt.UTC().Format("2006-01-02T15:04:05.000Z"),
		"codeLength":    s.CodeLength,
	}
	if s.QueueStartedAt != nil {
		view["queueStartedAt"] = s.QueueStartedAt.UTC().Format("2006-01-02T15:04:05.000Z")
	}
	if s.JudgeStartedAt != nil {
		view["judgeStartedAt"] = s.JudgeStartedAt.UTC().Format("2006-01-02T15:04:05.000Z")
	}
	if s.FinishedAt != nil {
		view["finishedAt"] = s.FinishedAt.UTC().Format("2006-01-02T15:04:05.000Z")
	}
	if withCases {
		cases := make([]gin.H, 0, len(s.CaseResults))
		for _, item := range s.CaseResults {
			cases = append(cases, gin.H{
				"submissionId":  item.SubmissionID,
				"caseNo":        item.CaseNo,
				"status":        item.Status,
				"runtimeMs":     item.RuntimeMS,
				"memoryKb":      item.MemoryKB,
				"stdoutBytes":   item.StdoutBytes,
				"stderrBytes":   item.StderrBytes,
				"signal":        item.Signal,
				"stdoutPreview": item.StdoutPreview,
				"stderrPreview": item.StderrPreview,
			})
		}
		view["caseResults"] = cases
	}
	return view
}
