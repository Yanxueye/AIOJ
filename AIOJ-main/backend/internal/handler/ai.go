package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	aisvc "github.com/terminaloj/backend/internal/ai"
	"github.com/terminaloj/backend/internal/config"
	"github.com/terminaloj/backend/internal/judger"
	"github.com/terminaloj/backend/internal/middleware"
	"github.com/terminaloj/backend/internal/models"
	"github.com/terminaloj/backend/internal/utils"
	"gorm.io/gorm"
)

type AIHandler struct {
	DB         *gorm.DB
	Client     *aisvc.Client
	Judger     judger.JudgerClient
}

type chatReq struct {
	Message        string          `json:"message" binding:"required"`
	History        []aisvc.Message `json:"history"`
	ProblemID      *uint64         `json:"problem_id"`
	ConversationID string          `json:"conversation_id"`
	CodeLanguage   string          `json:"code_language,omitempty"`
	Code           string          `json:"code,omitempty"`
}

type codeDiagnosisReq struct {
	ProblemID    uint64 `json:"problemId"`
	SubmissionID uint64 `json:"submissionId"`
	Language     string `json:"language"`
	Code         string `json:"code"`
	JudgeStatus  string `json:"judgeStatus"`
	ErrorMessage string `json:"errorMessage"`
	RuntimeMs    int    `json:"runtimeMs"`
	MemoryKb     int    `json:"memoryKb"`
}

type knowledgeGraphReq struct {
	ProblemID *uint64 `json:"problemId"`
	Scope     string  `json:"scope"`
}

type solveReq struct {
	ProblemID uint64 `json:"problemId" binding:"required"`
	Question  string `json:"question"`
	Level     string `json:"level"`
	Language  string `json:"language"`
	Code      string `json:"code"`
}

func (h *AIHandler) Chat(c *gin.Context) {
	uid, _ := middleware.CurrentUserID(c)
	var req chatReq
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数不合法")
		return
	}
	req.Message = strings.TrimSpace(req.Message)
	if req.Message == "" {
		utils.BadRequest(c, "消息不能为空")
		return
	}

	problemCtx, err := h.problemContext(req.ProblemID)
	if err != nil {
		utils.BadRequest(c, "题目不存在")
		return
	}

	conv, err := h.ensureConversation(uid, strings.TrimSpace(req.ConversationID), req.ProblemID, req.Message)
	if err != nil {
		utils.Server(c, err.Error())
		return
	}
	if err := h.DB.Create(&models.Message{ConversationID: conv.ID, Role: "user", Content: req.Message}).Error; err != nil {
		utils.Server(c, err.Error())
		return
	}

	resp, err := h.aiClient().Chat(c.Request.Context(), aisvc.ChatRequest{
		UserID:         uid,
		ConversationID: conv.ID,
		Message:        req.Message,
		History:        sanitizeHistory(req.History),
		Problem:        problemCtx,
		CodeLanguage:   strings.TrimSpace(req.CodeLanguage),
		Code:           strings.TrimSpace(req.Code),
	})
	if err != nil {
		utils.Server(c, err.Error())
		return
	}
	reply := strings.TrimSpace(resp.Reply)
	if reply == "" {
		utils.Server(c, "AI 服务返回空回复")
		return
	}
	if err := h.DB.Create(&models.Message{ConversationID: conv.ID, Role: "assistant", Content: reply}).Error; err != nil {
		utils.Server(c, err.Error())
		return
	}

	utils.OK(c, gin.H{
		"reply":          reply,
		"conversationId": conv.ID,
		"provider":       resp.Provider,
		"metadata":       resp.Metadata,
	})
}

func (h *AIHandler) History(c *gin.Context) {
	uid, _ := middleware.CurrentUserID(c)
	type row struct {
		ID           string
		Title        string
		ProblemID    *uint64
		CreatedAt    time.Time
		MessageCount int
	}
	var rows []row
	h.DB.Raw(`
		SELECT c.id, c.title, c.problem_id, c.created_at, COALESCE(COUNT(m.id),0) AS message_count
		FROM conversations c
		LEFT JOIN messages m ON m.conversation_id = c.id
		WHERE c.user_id = ?
		GROUP BY c.id, c.title, c.problem_id, c.created_at
		ORDER BY c.created_at DESC LIMIT 50`, uid).Scan(&rows)
	list := make([]gin.H, 0, len(rows))
	for _, r := range rows {
		list = append(list, gin.H{
			"id":           r.ID,
			"title":        r.Title,
			"problemId":    r.ProblemID,
			"createdAt":    formatTime(r.CreatedAt),
			"messageCount": r.MessageCount,
		})
	}
	utils.OK(c, gin.H{"conversations": list})
}

func (h *AIHandler) Messages(c *gin.Context) {
	uid, _ := middleware.CurrentUserID(c)
	convID := strings.TrimSpace(c.Param("id"))
	if convID == "" {
		utils.BadRequest(c, "会话 ID 不能为空")
		return
	}
	var conv models.Conversation
	if err := h.DB.Where("id = ? AND user_id = ?", convID, uid).First(&conv).Error; err != nil {
		utils.NotFound(c, "会话不存在")
		return
	}
	var rows []models.Message
	if err := h.DB.Where("conversation_id = ?", conv.ID).Order("id ASC").Find(&rows).Error; err != nil {
		utils.Server(c, err.Error())
		return
	}
	messages := make([]gin.H, 0, len(rows))
	for _, m := range rows {
		messages = append(messages, gin.H{
			"id":        m.ID,
			"role":      m.Role,
			"content":   m.Content,
			"createdAt": formatTime(m.CreatedAt),
		})
	}
	utils.OK(c, gin.H{
		"conversation": gin.H{
			"id":        conv.ID,
			"title":     conv.Title,
			"problemId": conv.ProblemID,
			"createdAt": formatTime(conv.CreatedAt),
		},
		"messages": messages,
	})
}

func (h *AIHandler) DeleteConversation(c *gin.Context) {
	uid, _ := middleware.CurrentUserID(c)
	convID := c.Param("id")
	if convID == "" {
		utils.BadRequest(c, "会话 ID 不能为空")
		return
	}
	h.DB.Where("conversation_id = ?", convID).Delete(&models.Message{})
	h.DB.Where("id = ? AND user_id = ?", convID, uid).Delete(&models.Conversation{})
	utils.OK(c, nil)
}

func (h *AIHandler) CodeDiagnosis(c *gin.Context) {
	uid, _ := middleware.CurrentUserID(c)
	var req codeDiagnosisReq
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数不合法")
		return
	}

	if req.SubmissionID > 0 {
		var sub models.Submission
		if err := h.DB.Where("id = ? AND user_id = ?", req.SubmissionID, uid).First(&sub).Error; err != nil {
			utils.NotFound(c, "提交记录不存在")
			return
		}
		if req.ProblemID == 0 {
			req.ProblemID = sub.ProblemID
		}
		if strings.TrimSpace(req.Language) == "" {
			req.Language = sub.Language
		}
		if strings.TrimSpace(req.Code) == "" {
			req.Code = sub.Code
		}
		if strings.TrimSpace(req.JudgeStatus) == "" {
			req.JudgeStatus = sub.Status
		}
		if strings.TrimSpace(req.ErrorMessage) == "" {
			req.ErrorMessage = sub.ErrorMessage
		}
	}

	if req.ProblemID == 0 {
		utils.BadRequest(c, "problemId 不能为空")
		return
	}
	if strings.TrimSpace(req.Language) == "" {
		utils.BadRequest(c, "language 不能为空")
		return
	}
	if strings.TrimSpace(req.Code) == "" {
		utils.BadRequest(c, "code 不能为空")
		return
	}
	problemCtx, err := h.problemContext(&req.ProblemID)
	if err != nil {
		utils.BadRequest(c, "题目不存在")
		return
	}

	// Get the most recent submission for this problem
	recentSubs, _ := h.recentSubmissions(uid, &req.ProblemID, 1)

	// Look up the first failing test case for non-AC submissions
	var failedCase *aisvc.FailedCase
	if req.SubmissionID > 0 && req.JudgeStatus != "Accepted" {
		failedCase = h.findFailedCase(req.SubmissionID, req.ProblemID, uid)
	}

	resp, err := h.aiClient().DiagnoseCode(c.Request.Context(), aisvc.CodeDiagnosisRequest{
		UserID:       uid,
		Problem:      problemCtx,
		SubmissionID: req.SubmissionID,
		Language:     strings.TrimSpace(req.Language),
		Code:         req.Code,
		JudgeStatus:  strings.TrimSpace(req.JudgeStatus),
		ErrorMessage: strings.TrimSpace(req.ErrorMessage),
		RuntimeMs:    req.RuntimeMs,
		MemoryKb:     req.MemoryKb,
		RecentSubs:   recentSubs,
		FailedCase:   failedCase,
	})
	if err != nil {
		utils.Server(c, err.Error())
		return
	}
	utils.OK(c, resp)
}

type generateSolutionReq struct {
	ProblemID uint64 `json:"problemId" binding:"required"`
	Language  string `json:"language"`
	Code      string `json:"code"`
}

func (h *AIHandler) GenerateSolution(c *gin.Context) {
	uid, _ := middleware.CurrentUserID(c)
	var req generateSolutionReq
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数不合法")
		return
	}

	problemCtx, err := h.problemContext(&req.ProblemID)
	if err != nil {
		utils.BadRequest(c, "题目不存在")
		return
	}

	// If no code provided, get the user's latest AC for this problem
	code := strings.TrimSpace(req.Code)
	language := strings.TrimSpace(req.Language)
	if code == "" {
		var sub models.Submission
		if err := h.DB.Where("user_id = ? AND problem_id = ? AND status = ?", uid, req.ProblemID, models.StatusAccepted).
			Order("id DESC").First(&sub).Error; err != nil {
			utils.BadRequest(c, "未找到通过的提交记录")
			return
		}
		code = sub.Code
		if language == "" {
			language = sub.Language
		}
	}

	resp, err := h.aiClient().GenerateSolution(c.Request.Context(), aisvc.GenerateSolutionRequest{
		UserID:  uid,
		Problem: problemCtx,
		Language: language,
		Code:     code,
	})
	if err != nil {
		utils.Server(c, err.Error())
		return
	}
	utils.OK(c, resp)
}

func (h *AIHandler) KnowledgeGraph(c *gin.Context) {
	uid, _ := middleware.CurrentUserID(c)
	var req knowledgeGraphReq
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数不合法")
		return
	}
	scope := strings.TrimSpace(req.Scope)
	if scope == "" {
		scope = "recent"
	}
	problemCtx, err := h.problemContext(req.ProblemID)
	if err != nil {
		utils.BadRequest(c, "题目不存在")
		return
	}
	recent, err := h.recentSubmissions(uid, req.ProblemID, 50)
	if err != nil {
		utils.Server(c, err.Error())
		return
	}

	// Aggregate submissions into problems and tag stats for agent-service
	problems := h.aggregateProblems(recent)
	tagStats := h.aggregateTagStats(recent)

	resp, err := h.aiClient().BuildKnowledgeGraph(c.Request.Context(), aisvc.KnowledgeGraphRequest{
		UserID:            uid,
		Scope:             scope,
		Problem:           problemCtx,
		RecentSubmissions: recent,
		Problems:          problems,
		TagStats:          tagStats,
	})
	if err != nil {
		utils.Server(c, err.Error())
		return
	}

	// Persist knowledge graph to database
	nodesJSON, _ := json.Marshal(resp.Nodes)
	edgesJSON, _ := json.Marshal(resp.Edges)
	graph := models.UserKnowledgeGraph{
		UserID:  uid,
		Scope:   scope,
		Nodes:   string(nodesJSON),
		Edges:   string(edgesJSON),
		Summary: strings.Join(resp.Suggestions, "；"),
	}
	// Upsert: update if exists for same scope, create if not
	var existing models.UserKnowledgeGraph
	if err := h.DB.Where("user_id = ? AND scope = ?", uid, scope).First(&existing).Error; err == nil {
		existing.Nodes = string(nodesJSON)
		existing.Edges = string(edgesJSON)
		existing.Summary = strings.Join(resp.Suggestions, "；")
		if saveErr := h.DB.Save(&existing).Error; saveErr != nil {
			log.Printf("[ai] knowledge graph save failed: %v", saveErr)
		}
	} else {
		if createErr := h.DB.Create(&graph).Error; createErr != nil {
			log.Printf("[ai] knowledge graph create failed: %v", createErr)
		}
	}

	utils.OK(c, resp)
}

func (h *AIHandler) CreateStudyPlan(c *gin.Context) {
	uid, _ := middleware.CurrentUserID(c)
	if uid == 0 {
		utils.Unauthorized(c, "请先登录")
		return
	}
	// 1. Gather user's submission data (same as KnowledgeGraph)
	recent, err := h.recentSubmissions(uid, nil, 50)
	if err != nil {
		utils.Server(c, err.Error())
		return
	}
	problems := h.aggregateProblems(recent)
	tagStats := h.aggregateTagStats(recent)

	// 2. Find weak tags and gather candidate problems for each
	candidates := make(map[string][]aisvc.ProblemSummary)
	for tag, stat := range tagStats {
		if stat.ACRate < 50 || stat.Solved < 3 {
			// Weak tag: find untried published problems with this tag
			var tagProblems []models.Problem
			h.DB.Where("status = ? AND JSON_CONTAINS(tags, ?)", models.ProblemStatusPublished,
				fmt.Sprintf(`"%s"`, tag)).Find(&tagProblems)
			triedSet := make(map[uint64]bool)
			for _, p := range problems {
				triedSet[p.ID] = true
			}
			for _, tp := range tagProblems {
				if !triedSet[tp.ID] {
					candidates[tag] = append(candidates[tag], aisvc.ProblemSummary{
						ID:     tp.ID,
						Title:  tp.Title,
						Tags:   tp.Tags,
						Status: "unattempted",
					})
				}
			}
		}
	}

	// 3. Call agent service to create study plan
	resp, err := h.aiClient().CreateStudyPlan(c.Request.Context(), aisvc.CreateStudyPlanRequest{
		UserID:     uid,
		Problems:   problems,
		TagStats:   tagStats,
		Candidates: candidates,
	})
	if err != nil {
		utils.Server(c, err.Error())
		return
	}

	// 4. Validate problem IDs exist in DB
	var count int64
	h.DB.Model(&models.Problem{}).Where("id IN ?", resp.ProblemIDs).Count(&count)

	// 5. Create the study plan
	plan := models.StudyPlan{
		UserID:      uid,
		Title:       resp.Title,
		Description: resp.Description,
		Difficulty:  "中等",
	}
	if err := h.DB.Create(&plan).Error; err != nil {
		utils.Server(c, err.Error())
		return
	}
	for i, pid := range resp.ProblemIDs {
		var p models.Problem
		if h.DB.First(&p, pid).Error != nil { continue }
		h.DB.Create(&models.StudyPlanItem{
			PlanID: plan.ID, ProblemID: pid, OrderNo: i + 1,
			Title: p.Title, Difficulty: p.Difficulty,
		})
	}
	utils.OK(c, gin.H{"id": plan.ID, "title": plan.Title, "problemCount": len(resp.ProblemIDs)})
}

func (h *AIHandler) Solve(c *gin.Context) {
	uid, _ := middleware.CurrentUserID(c)
	var req solveReq
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, "请求参数不合法")
		return
	}
	level := strings.TrimSpace(req.Level)
	if level == "" {
		level = "hint"
	}
	if level != "hint" && level != "explain" && level != "full" {
		utils.BadRequest(c, "level 只能是 hint/explain/full")
		return
	}
	problemCtx, err := h.problemContext(&req.ProblemID)
	if err != nil {
		utils.BadRequest(c, "题目不存在")
		return
	}

	solveReq := aisvc.SolveRequest{
		UserID:     uid,
		Problem:    problemCtx,
		Question:   strings.TrimSpace(req.Question),
		Level:      level,
		Language:   strings.TrimSpace(req.Language),
		EditorCode: req.Code,
	}

	// For full level, implement state machine with judge retry
	if level == "full" {
		h.solveWithRetry(c, solveReq)
		return
	}

	resp, err := h.aiClient().Solve(c.Request.Context(), solveReq)
	if err != nil {
		utils.Server(c, err.Error())
		return
	}
	utils.OK(c, resp)
}

// solveWithRetry implements the state machine for full-level solve:
// 1. AI generates code
// 2. OJ backend judges (without saving record)
// 3. If not AC, send result back to AI for modification
// 4. Retry up to 3 times
// 5. Return code or "sorry" message
func (h *AIHandler) solveWithRetry(c *gin.Context, req aisvc.SolveRequest) {
	maxRetries := 3

	for attempt := 0; attempt < maxRetries; attempt++ {
		resp, err := h.aiClient().Solve(c.Request.Context(), req)
		if err != nil {
			utils.Server(c, err.Error())
			return
		}

		// If no code generated, return as-is (hint/explain mode)
		if resp.Code == "" {
			utils.OK(c, resp)
			return
		}

		// Try to judge the generated code
		lang := resp.Language
		if lang == "" {
			lang = "cpp"
		}

		judgeResult, err := h.judgeCode(c, req.Problem.ID, lang, resp.Code)
		if err != nil {
			// Judge failed, return AI response without verification
			resp.VerifyResult = "判题服务暂时不可用，无法验证代码"
			utils.OK(c, resp)
			return
		}

		if judgeResult == "Accepted" {
			resp.VerifyResult = "✅ 代码已通过验证"
			utils.OK(c, resp)
			return
		}

		// Not AC, prepare for retry
		req.JudgeError = fmt.Sprintf("判题结果: %s", judgeResult)
		req.EditorCode = resp.Code

		if attempt == maxRetries-1 {
			// Last attempt, return with failure message
			resp.VerifyResult = fmt.Sprintf("❌ 经过 %d 次尝试仍无法通过（%s）", maxRetries, judgeResult)
			resp.Answer = "抱歉，我暂时无法生成通过此题的代码。建议你先理解算法思路，再自己实现。"
			utils.OK(c, resp)
			return
		}
	}
}

// judgeCode runs code through the judge service without saving a submission record
func (h *AIHandler) judgeCode(c *gin.Context, problemID uint64, language, code string) (string, error) {
	// Load problem for test cases
	var problem models.Problem
	if err := h.DB.Preload("PublishedVersion").Preload("PublishedVersion.TestCases").First(&problem, problemID).Error; err != nil {
		return "", fmt.Errorf("load problem: %w", err)
	}
	if problem.PublishedVersion == nil {
		return "", fmt.Errorf("problem has no published version")
	}

	// Use the run endpoint logic (synchronous judge)
	normalizedLang, ok := judger.NormalizeLanguage(language)
	if !ok {
		return "", fmt.Errorf("unsupported language: %s", language)
	}

	testCases := make([]judger.TestCase, 0, len(problem.PublishedVersion.TestCases))
	for _, tc := range problem.PublishedVersion.TestCases {
		testCases = append(testCases, judger.TestCase{
			CaseNo:   int32(tc.CaseNo),
			Input:    tc.Input,
			Expected: tc.Expected,
		})
	}

	traceID := fmt.Sprintf("ai-verify-%d-%d", problemID, time.Now().UnixNano())
	resp, err := h.Judger.Judge(c.Request.Context(), &judger.JudgeRequest{
		SubmissionID:  0, // Don't save submission record
		ProblemID:     problem.ID,
		TraceID:       traceID,
		Language:      normalizedLang,
		Code:          code,
		TimeLimitMS:   int32(problem.PublishedVersion.TimeLimit),
		MemoryLimitMB: int32(problem.PublishedVersion.MemoryLimit),
		OutputLimitKB: problem.PublishedVersion.OutputLimitKB,
		RunMode:       "submit", // Use submit mode to test all cases
		TestCases:     testCases,
	})
	if err != nil {
		return "", fmt.Errorf("judge failed: %w", err)
	}

	return resp.Status, nil
}

func (h *AIHandler) aiClient() *aisvc.Client {
	if h.Client != nil {
		return h.Client
	}
	return aisvc.NewClient(config.AIConfig{})
}

func (h *AIHandler) ensureConversation(uid uint64, convID string, problemID *uint64, firstMsg string) (*models.Conversation, error) {
	if convID != "" {
		var c models.Conversation
		if err := h.DB.Where("id = ? AND user_id = ?", convID, uid).First(&c).Error; err == nil {
			return &c, nil
		}
	}
	title := firstMsg
	if len([]rune(title)) > 24 {
		title = string([]rune(title)[:24]) + "..."
	}
	c := &models.Conversation{
		ID:        uuid.NewString(),
		UserID:    uid,
		ProblemID: problemID,
		Title:     title,
		CreatedAt: time.Now(),
	}
	return c, h.DB.Create(c).Error
}

func (h *AIHandler) problemContext(problemID *uint64) (*aisvc.ProblemContext, error) {
	if problemID == nil {
		return nil, nil
	}
	if *problemID == 0 {
		return nil, errors.New("empty problem id")
	}
	var p models.Problem
	if err := h.DB.Preload("PublishedVersion").Preload("PublishedVersion.Samples").
		Preload("CurrentVersion").Preload("CurrentVersion.Samples").First(&p, *problemID).Error; err != nil {
		return nil, err
	}
	version := p.PublishedVersion
	if version == nil {
		version = p.CurrentVersion
	}
	content := ""
	editorial := ""
	timeLimit := 0
	memoryLimit := 0
	tags := append([]string(nil), p.Tags...)
	var samples []aisvc.Sample
	if version != nil {
		content = version.Content
		editorial = version.Editorial
		timeLimit = version.TimeLimit
		memoryLimit = version.MemoryLimit
		tags = append([]string(nil), version.Tags...)
		for _, s := range version.Samples {
			samples = append(samples, aisvc.Sample{
				Input:    s.Input,
				Expected: s.Expected,
			})
		}
	}
	return &aisvc.ProblemContext{
		ID:              p.ID,
		Title:           p.Title,
		Difficulty:      p.Difficulty,
		DifficultyScore: p.DifficultyScore,
		Tags:            tags,
		Content:         content,
		Editorial:       editorial,
		Samples:         samples,
		TimeLimit:       timeLimit,
		MemoryLimit:     memoryLimit,
	}, nil
}

func (h *AIHandler) recentSubmissions(uid uint64, problemID *uint64, limit int) ([]aisvc.SubmissionDigest, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	q := h.DB.Where("user_id = ?", uid).Order("created_at DESC").Limit(limit)
	if problemID != nil && *problemID > 0 {
		q = q.Where("problem_id = ?", *problemID)
	}
	var rows []models.Submission
	if err := q.Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]aisvc.SubmissionDigest, 0, len(rows))
	for _, s := range rows {
		out = append(out, aisvc.SubmissionDigest{
			ID:           s.ID,
			ProblemID:    s.ProblemID,
			ProblemTitle: s.ProblemTitle,
			Language:     s.Language,
			Status:       s.Status,
			Code:         s.Code,
			Runtime:      s.Runtime,
			Memory:       s.Memory,
			CodeLength:   s.CodeLength,
			ErrorMessage: s.ErrorMessage,
			CreatedAt:    formatTime(s.CreatedAt),
		})
	}
	return out, nil
}

func sanitizeHistory(history []aisvc.Message) []aisvc.Message {
	if len(history) > 20 {
		history = history[len(history)-20:]
	}
	out := make([]aisvc.Message, 0, len(history))
	for _, m := range history {
		role := strings.TrimSpace(m.Role)
		content := strings.TrimSpace(m.Content)
		if content == "" {
			continue
		}
		if role != "system" && role != "user" && role != "assistant" {
			role = "user"
		}
		out = append(out, aisvc.Message{Role: role, Content: content})
	}
	return out
}

func formatTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.UTC().Format("2006-01-02T15:04:05.000Z")
}

// aggregateProblems groups submissions by problem and returns a summary list.
func (h *AIHandler) aggregateProblems(subs []aisvc.SubmissionDigest) []aisvc.ProblemSummary {
	type agg struct {
		title    string
		tags     []string
		statuses map[string]bool
		attempts int
	}
	problemMap := make(map[uint64]*agg)

	// Get problem info from DB for tags
	problemIDs := make([]uint64, 0)
	seen := make(map[uint64]bool)
	for _, s := range subs {
		if !seen[s.ProblemID] {
			problemIDs = append(problemIDs, s.ProblemID)
			seen[s.ProblemID] = true
		}
	}

	problemTags := make(map[uint64][]string)
	problemTitles := make(map[uint64]string)
	if len(problemIDs) > 0 {
		var problems []models.Problem
		h.DB.Where("id IN ?", problemIDs).Find(&problems)
		for _, p := range problems {
			problemTags[p.ID] = p.Tags
			problemTitles[p.ID] = p.Title
		}
	}

	for _, s := range subs {
		a, ok := problemMap[s.ProblemID]
		if !ok {
			a = &agg{
				title:    problemTitles[s.ProblemID],
				tags:     problemTags[s.ProblemID],
				statuses: make(map[string]bool),
			}
			problemMap[s.ProblemID] = a
		}
		a.attempts++
		a.statuses[s.Status] = true
	}

	out := make([]aisvc.ProblemSummary, 0, len(problemMap))
	for pid, a := range problemMap {
		status := "attempted"
		if a.statuses["Accepted"] {
			status = "solved"
		}
		out = append(out, aisvc.ProblemSummary{
			ID:       pid,
			Title:    a.title,
			Tags:     a.tags,
			Status:   status,
			Attempts: a.attempts,
		})
	}
	return out
}

// aggregateTagStats computes per-tag solve/attempt statistics.
func (h *AIHandler) aggregateTagStats(subs []aisvc.SubmissionDigest) map[string]aisvc.TagStat {
	type tagAgg struct {
		solved    map[uint64]bool
		attempted map[uint64]bool
	}

	// Get problem tags
	problemIDs := make([]uint64, 0)
	seen := make(map[uint64]bool)
	for _, s := range subs {
		if !seen[s.ProblemID] {
			problemIDs = append(problemIDs, s.ProblemID)
			seen[s.ProblemID] = true
		}
	}

	problemTags := make(map[uint64][]string)
	if len(problemIDs) > 0 {
		var problems []models.Problem
		h.DB.Where("id IN ?", problemIDs).Find(&problems)
		for _, p := range problems {
			problemTags[p.ID] = p.Tags
		}
	}

	tagMap := make(map[string]*tagAgg)
	for _, s := range subs {
		tags := problemTags[s.ProblemID]
		for _, tag := range tags {
			a, ok := tagMap[tag]
			if !ok {
				a = &tagAgg{
					solved:    make(map[uint64]bool),
					attempted: make(map[uint64]bool),
				}
				tagMap[tag] = a
			}
			a.attempted[s.ProblemID] = true
			if s.Status == "Accepted" {
				a.solved[s.ProblemID] = true
			}
		}
	}

	out := make(map[string]aisvc.TagStat, len(tagMap))
	for tag, a := range tagMap {
		attempted := len(a.attempted)
		solved := len(a.solved)
		acRate := 0.0
		if attempted > 0 {
			acRate = float64(solved) * 100.0 / float64(attempted)
		}
		out[tag] = aisvc.TagStat{
			Solved:    solved,
			Attempted: attempted,
			ACRate:    acRate,
		}
	}
	return out
}

// findFailedCase looks up the first non-AC test case result for a submission.
func (h *AIHandler) findFailedCase(submissionID, problemID, uid uint64) *aisvc.FailedCase {
	// Get submission case results
	var caseResults []models.SubmissionCaseResult
	h.DB.Where("submission_id = ?", submissionID).Order("case_no ASC").Find(&caseResults)

	// Find the first non-Accepted case
	var failedCaseNo int
	for _, cr := range caseResults {
		if cr.Status != models.StatusAccepted {
			failedCaseNo = cr.CaseNo
			break
		}
	}
	if failedCaseNo == 0 {
		return nil
	}

	// Get the actual output from the case result
	var actual string
	for _, cr := range caseResults {
		if cr.CaseNo == failedCaseNo {
			actual = cr.StdoutPreview
			break
		}
	}

	// Look up the test case input/expected from the problem
	var problem models.Problem
	if err := h.DB.Preload("PublishedVersion").First(&problem, problemID).Error; err != nil {
		return nil
	}
	if problem.PublishedVersion == nil {
		return nil
	}

	var tc models.ProblemTestCase
	if err := h.DB.Where("version_id = ? AND case_no = ?", problem.PublishedVersion.ID, failedCaseNo).First(&tc).Error; err != nil {
		return nil
	}

	return &aisvc.FailedCase{
		Input:    tc.Input,
		Expected: tc.Expected,
		Actual:   actual,
	}
}
