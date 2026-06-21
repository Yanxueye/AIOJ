package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/terminaloj/backend/internal/config"
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
	Judger judger.JudgerClient
}

type submitReq struct {
	ProblemID uint64 `json:"problemId" binding:"required"`
	Language  string `json:"language" binding:"required"`
	Code      string `json:"code" binding:"required"`
}

type runReq struct {
	Language  string         `json:"language" binding:"required"`
	Code      string         `json:"code" binding:"required"`
	TestCases []testCaseItem `json:"testCases"`
}

type testCaseItem struct {
	Input    string `json:"input"`
	Expected string `json:"expected"`
}

func (h *SubmissionHandler) Submit(c *gin.Context) {
	uid, _ := middleware.CurrentUserID(c)

	var req submitReq
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "参数不合法")
		return
	}

	problem, normalizedLang, code, ok := h.validateSourceInput(c, req.ProblemID, req.Language, req.Code)
	if !ok {
		return
	}

	now := time.Now().UTC()

	// Create submission record directly — let DB auto-increment handle the ID
	sub := &models.Submission{
		UserID:         uid,
		ProblemID:      problem.ID,
		ProblemTitle:   problem.Title,
		Source:         "submit",
		Language:       normalizedLang,
		Code:           code,
		CodeLength:     len(code),
		Status:         models.StatusPending,
		QueueStartedAt: &now,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	if err := h.DB.Create(sub).Error; err != nil {
		utils.Server(c, "创建提交记录失败: "+err.Error())
		return
	}

	traceID := fmt.Sprintf("judge-%d-%d", problem.ID, sub.ID)

	task := &mq.SubmitTask{
		SubmissionID: sub.ID,
		UserID:       uid,
		ProblemID:    problem.ID,
		ProblemTitle: problem.Title,
		TraceID:      traceID,
		Source:       "submit",
		Language:     normalizedLang,
		Code:         code,
		EnqueuedAt:   now,
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()
	if err := h.Broker.Publish(ctx, task); err != nil {
		utils.Server(c, "队列发布失败: "+err.Error())
		return
	}

	utils.OK(c, gin.H{
		"id":            sub.ID,
		"problemId":     problem.ID,
		"problemTitle":  problem.Title,
		"traceId":       traceID,
		"source":        "submit",
		"status":        models.StatusPending,
		"language":      normalizedLang,
		"runtime":       0,
		"runtimeMs":     0,
		"memory":        "0.0",
		"memoryKb":      0,
		"compileOutput": "",
		"errorMessage":  "",
		"caseResults":   []gin.H{},
		"createdAt":     now.Format("2006-01-02T15:04:05.000Z"),
		"updatedAt":     now.Format("2006-01-02T15:04:05.000Z"),
		"codeLength":    len(code),
	})
}

func (h *SubmissionHandler) Run(c *gin.Context) {
	var req runReq
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "参数不合法")
		return
	}

	problemID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.BadRequest(c, "题号不合法")
		return
	}

	problem, normalizedLang, code, ok := h.validateSourceInput(c, problemID, req.Language, req.Code)
	if !ok {
		return
	}

	traceID := fmt.Sprintf("run-%d-%d", problem.ID, time.Now().UTC().UnixNano())

	// Build test cases from request
	testCases := make([]judger.TestCase, 0, len(req.TestCases))
	for i, tc := range req.TestCases {
		testCases = append(testCases, judger.TestCase{
			CaseNo:   int32(i + 1),
			Input:    tc.Input,
			Expected: tc.Expected,
		})
	}
	if len(testCases) == 0 {
		testCases = []judger.TestCase{{CaseNo: 1, Input: "", Expected: ""}}
	}

	resp, err := h.Judger.Judge(c.Request.Context(), &judger.JudgeRequest{
		SubmissionID:  0,
		ProblemID:     problem.ID,
		TraceID:       traceID,
		Language:      normalizedLang,
		Code:          code,
		TimeLimitMS:   int32(problem.PublishedVersion.TimeLimit),
		MemoryLimitMB: int32(problem.PublishedVersion.MemoryLimit),
		OutputLimitKB: problem.PublishedVersion.OutputLimitKB,
		RunMode:       "run",
		TestCases:     testCases,
	})
	if err != nil {
		utils.Server(c, "运行代码失败: "+err.Error())
		return
	}

	utils.OK(c, gin.H{
		"traceId":       traceID,
		"problemId":     problem.ID,
		"problemTitle":  problem.Title,
		"source":        "run",
		"status":        resp.Status,
		"language":      normalizedLang,
		"runtime":       resp.RuntimeMS,
		"runtimeMs":     resp.RuntimeMS,
		"memory":        resp.MemoryMB,
		"memoryKb":      resp.MemoryKB,
		"compileOutput": resp.CompileOut,
		"errorMessage":  resp.ErrorMessage,
		"caseResults":   buildCaseViews(resp.CaseResults),
		"stdout":        firstStdout(resp.CaseResults),
		"stderr":        firstStderr(resp.CaseResults),
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

	q := h.DB.Model(&models.Submission{}).Where("user_id = ?", uid).Where("source = ?", "submit")
	if pid := c.Query("problemId"); pid != "" {
		q = q.Where("problem_id = ?", pid)
	}
	if status := c.Query("status"); status != "" {
		q = q.Where("status = ?", status)
	}
	switch c.DefaultQuery("sortBy", "time") {
	case "problemId":
		q = q.Order("problem_id ASC, created_at DESC")
	default:
		q = q.Order("created_at DESC")
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		log.Printf("[submission] list count query failed: %v", err)
	}

	var rows []models.Submission
	if err := q.Offset((page - 1) * size).Limit(size).Find(&rows).Error; err != nil {
		utils.Server(c, err.Error())
		return
	}

	list := make([]gin.H, 0, len(rows))
	for i := range rows {
		list = append(list, submissionView(&rows[i], false, false)) // No code or cases in list
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
	utils.OK(c, submissionView(&s, true, true)) // Detail: include code + cases
}

func (h *SubmissionHandler) Cases(c *gin.Context) {
	uid, _ := middleware.CurrentUserID(c)
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.BadRequest(c, "提交编号不合法")
		return
	}

	var s models.Submission
	if err := h.DB.Where("id = ? AND user_id = ?", id, uid).First(&s).Error; err != nil {
		utils.NotFound(c, "提交不存在")
		return
	}

	var rows []models.SubmissionCaseResult
	if err := h.DB.Where("submission_id = ?", s.ID).Order("case_no ASC").Find(&rows).Error; err != nil {
		utils.Server(c, err.Error())
		return
	}

	items := make([]gin.H, 0, len(rows))
	for _, item := range rows {
		items = append(items, gin.H{
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
	utils.OK(c, gin.H{"submissionId": s.ID, "items": items})
}

func (h *SubmissionHandler) Output(c *gin.Context) {
	uid, _ := middleware.CurrentUserID(c)
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.BadRequest(c, "提交编号不合法")
		return
	}

	var s models.Submission
	if err := h.DB.Where("id = ? AND user_id = ?", id, uid).First(&s).Error; err != nil {
		utils.NotFound(c, "提交不存在")
		return
	}

	var first models.SubmissionCaseResult
	if err := h.DB.Where("submission_id = ?", s.ID).Order("case_no ASC").First(&first).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.OK(c, gin.H{"submissionId": s.ID, "stdout": "", "stderr": ""})
			return
		}
		utils.Server(c, err.Error())
		return
	}

	utils.OK(c, gin.H{
		"submissionId": s.ID,
		"stdout":       first.StdoutPreview,
		"stderr":       first.StderrPreview,
	})
}

func (h *SubmissionHandler) Stream(c *gin.Context) {
	uid, ok := middleware.CurrentUserID(c)
	if !ok {
		token := c.Query("token")
		if token == "" {
			utils.Unauthorized(c, "missing token")
			return
		}
		claims, err := utils.NewJWTManager(config.Get().JWT.Secret, config.Get().JWT.ExpireHours).Parse(token)
		if err != nil {
			utils.Unauthorized(c, "invalid token")
			return
		}
		uid = claims.UserID
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		utils.BadRequest(c, "提交编号不合法")
		return
	}

	w := c.Writer
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Flush()

	flusher, ok := w.(http.Flusher)
	if !ok {
		utils.Server(c, "stream not supported")
		return
	}

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()
	lastPayload := ""
	notFoundCount := 0

	for {
		select {
		case <-c.Request.Context().Done():
			return
		case <-ticker.C:
			var s models.Submission
			if err := h.DB.Preload("CaseResults", func(db *gorm.DB) *gorm.DB {
				return db.Order("case_no ASC")
			}).Where("id = ? AND user_id = ?", id, uid).First(&s).Error; err != nil {
				// Submission might not exist yet (worker hasn't created it) — retry
				notFoundCount++
				if notFoundCount > 30 { // ~15s timeout
					fmt.Fprintf(w, "event: error\ndata: {\"message\":\"submission not found\"}\n\n")
					flusher.Flush()
					return
				}
				continue
			}
			notFoundCount = 0
			payloadMap := submissionView(&s, false, true) // SSE: cases only, no code (save bandwidth)
			payloadBytes, _ := json.Marshal(payloadMap)
			payload := string(payloadBytes)
			if payload == lastPayload {
				continue
			}
			lastPayload = payload
			fmt.Fprintf(w, "event: submission\ndata: %s\n\n", payload)
			flusher.Flush()
			if isTerminalStatus(s.Status) {
				return
			}
		}
	}
}

func (h *SubmissionHandler) validateSourceInput(c *gin.Context, problemID uint64, language, code string) (*models.Problem, string, string, bool) {
	normalizedLang, ok := judger.NormalizeLanguage(language)
	if !ok {
		utils.BadRequest(c, "暂不支持该语言")
		return nil, "", "", false
	}
	if n := len(code); n == 0 || n > 65536 {
		utils.BadRequest(c, "代码长度不合法")
		return nil, "", "", false
	}

	var problem models.Problem
	if err := h.DB.Preload("PublishedVersion.TestCases").Preload("PublishedVersion").First(&problem, problemID).Error; err != nil {
		utils.NotFound(c, "题目不存在")
		return nil, "", "", false
	}
	if problem.Status != models.ProblemStatusPublished || problem.PublishedVersion == nil {
		utils.NotFound(c, "题目不存在")
		return nil, "", "", false
	}
	return &problem, normalizedLang, code, true
}

func submissionView(s *models.Submission, withCode bool, withCaseResults bool) gin.H {
	view := gin.H{
		"id":            s.ID,
		"problemId":     s.ProblemID,
		"problemTitle":  s.ProblemTitle,
		"traceId":       s.TraceID,
		"source":        defaultSource(s.Source),
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
	if withCode {
		view["code"] = s.Code
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
	if withCaseResults {
		view["caseResults"] = buildModelCaseViews(s.CaseResults)
	}
	return view
}

func buildCaseViews(items []judger.CaseResult) []gin.H {
	cases := make([]gin.H, 0, len(items))
	for _, item := range items {
		cases = append(cases, gin.H{
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
	return cases
}

func buildModelCaseViews(items []models.SubmissionCaseResult) []gin.H {
	cases := make([]gin.H, 0, len(items))
	for _, item := range items {
		cases = append(cases, gin.H{
			"submissionId":  item.SubmissionID,
			"caseNo":        item.CaseNo,
			"status":        item.Status,
			"runtimeMs":     item.RuntimeMS,
			"memoryKb":      item.MemoryKB,
			"stdoutBytes":   item.StdoutBytes,
			"stderrBytes":   item.StderrBytes,
			"signal":        item.Signal,
			"input":         item.Input,
			"expected":      item.Expected,
			"stdoutPreview": item.StdoutPreview,
			"stderrPreview": item.StderrPreview,
		})
	}
	return cases
}

func firstStdout(items []judger.CaseResult) string {
	if len(items) == 0 {
		return ""
	}
	return items[0].StdoutPreview
}

func firstStderr(items []judger.CaseResult) string {
	if len(items) == 0 {
		return ""
	}
	return items[0].StderrPreview
}

func defaultSource(value string) string {
	if value == "" {
		return "submit"
	}
	return value
}

func isTerminalStatus(status string) bool {
	switch status {
	case models.StatusAccepted,
		models.StatusWrong,
		models.StatusCompileErr,
		models.StatusRuntimeErr,
		models.StatusTLE,
		models.StatusMLE,
		models.StatusOLE,
		models.StatusSystemErr:
		return true
	default:
		return false
	}
}
