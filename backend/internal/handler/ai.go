package handler

import (
	"errors"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	aisvc "github.com/terminaloj/backend/internal/ai"
	"github.com/terminaloj/backend/internal/config"
	"github.com/terminaloj/backend/internal/middleware"
	"github.com/terminaloj/backend/internal/models"
	"github.com/terminaloj/backend/internal/utils"
	"gorm.io/gorm"
)

type AIHandler struct {
	DB     *gorm.DB
	Client *aisvc.Client
}

type chatReq struct {
	Message        string          `json:"message" binding:"required"`
	History        []aisvc.Message `json:"history"`
	ProblemID      *uint64         `json:"problem_id"`
	ConversationID string          `json:"conversation_id"`
}

type codeDiagnosisReq struct {
	ProblemID    uint64 `json:"problemId"`
	SubmissionID uint64 `json:"submissionId"`
	Language     string `json:"language"`
	Code         string `json:"code"`
	JudgeStatus  string `json:"judgeStatus"`
	ErrorMessage string `json:"errorMessage"`
}

type knowledgeGraphReq struct {
	ProblemID *uint64 `json:"problemId"`
	Scope     string  `json:"scope"`
}

type solveReq struct {
	ProblemID uint64 `json:"problemId" binding:"required"`
	Question  string `json:"question"`
	Level     string `json:"level"`
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

	resp, err := h.aiClient().DiagnoseCode(c.Request.Context(), aisvc.CodeDiagnosisRequest{
		UserID:       uid,
		Problem:      problemCtx,
		SubmissionID: req.SubmissionID,
		Language:     strings.TrimSpace(req.Language),
		Code:         req.Code,
		JudgeStatus:  strings.TrimSpace(req.JudgeStatus),
		ErrorMessage: strings.TrimSpace(req.ErrorMessage),
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
	resp, err := h.aiClient().BuildKnowledgeGraph(c.Request.Context(), aisvc.KnowledgeGraphRequest{
		UserID:            uid,
		Scope:             scope,
		Problem:           problemCtx,
		RecentSubmissions: recent,
	})
	if err != nil {
		utils.Server(c, err.Error())
		return
	}
	utils.OK(c, resp)
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
	resp, err := h.aiClient().Solve(c.Request.Context(), aisvc.SolveRequest{
		UserID:   uid,
		Problem:  problemCtx,
		Question: strings.TrimSpace(req.Question),
		Level:    level,
	})
	if err != nil {
		utils.Server(c, err.Error())
		return
	}
	utils.OK(c, resp)
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
	if err := h.DB.First(&p, *problemID).Error; err != nil {
		return nil, err
	}
	return &aisvc.ProblemContext{
		ID:              p.ID,
		Title:           p.Title,
		Difficulty:      p.Difficulty,
		DifficultyScore: p.DifficultyScore,
		Tags:            append([]string(nil), p.Tags...),
		Content:         p.Content,
		TimeLimit:       p.TimeLimit,
		MemoryLimit:     p.MemoryLimit,
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
