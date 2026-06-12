package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"agent-service/internal/ai"
	"agent-service/internal/judge"
	"agent-service/internal/rag"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	ai    *ai.Client
	judge *judge.Client
	rag   *rag.Service
}

func New(aiClient *ai.Client, judgeClient *judge.Client, ragService *rag.Service) *Handler {
	return &Handler{ai: aiClient, judge: judgeClient, rag: ragService}
}

func (h *Handler) Health(c *gin.Context) {
	status := gin.H{"status": "ok"}
	if err := h.ai.Health(); err != nil {
		status["ollama"] = "unreachable"
	} else {
		status["ollama"] = "ok"
	}
	c.JSON(http.StatusOK, status)
}

// HintRequest is the request for getting a hint on a wrong submission
type HintRequest struct {
	ProblemTitle string `json:"problemTitle"`
	ProblemDesc  string `json:"problemDesc"`
	Code         string `json:"code"`
	Language     string `json:"language"`
	ErrorInfo    string `json:"errorInfo"`
	Status       string `json:"status"`
}

// Hint provides a small hint for a wrong submission
func (h *Handler) Hint(c *gin.Context) {
	var req HintRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	prompt := fmt.Sprintf(`你是一个算法竞赛教练。用户正在做一道题目，提交错误了，请给出一个小提示，但不要直接给出完整解法。

题目: %s
题目描述: %s
用户语言: %s
提交状态: %s
错误信息: %s
用户代码:
%s

请严格按以下JSON格式返回：
{"hint": "简短提示（不超过3句话）", "relatedTopics": ["相关知识点"], "severity": "info/warning/error"}`,
		req.ProblemTitle, req.ProblemDesc, req.Language, req.Status, req.ErrorInfo, req.Code)

	resp, err := h.ai.Chat([]ai.Message{
		{Role: "system", Content: "你是一个友好的算法竞赛教练，善于用启发式的方式引导学生思考。请始终以JSON格式回复。"},
		{Role: "user", Content: prompt},
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "AI 服务暂时不可用，请稍后重试"})
		return
	}

	// Try to parse structured JSON response
	var structured map[string]interface{}
	if err := json.Unmarshal([]byte(resp), &structured); err == nil {
		structured["rawMarkdown"] = resp
		c.JSON(http.StatusOK, structured)
		return
	}

	c.JSON(http.StatusOK, gin.H{"hint": resp, "rawMarkdown": resp})
}

// AnalyzeRequest is the request for analyzing an accepted submission
type AnalyzeRequest struct {
	ProblemTitle   string `json:"problemTitle"`
	ProblemDesc    string `json:"problemDesc"`
	ExpectedTopics []string `json:"expectedTopics"`
	Code           string `json:"code"`
	Language       string `json:"language"`
	RuntimeMS      int    `json:"runtimeMs"`
	MemoryKB       int    `json:"memoryKb"`
}

// Analyze provides code analysis for an accepted submission
func (h *Handler) Analyze(c *gin.Context) {
	var req AnalyzeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	topics := ""
	for i, t := range req.ExpectedTopics {
		if i > 0 {
			topics += ", "
		}
		topics += t
	}

	prompt := fmt.Sprintf(`分析以下通过的代码。

题目: %s
语言: %s
运行时间: %dms
内存: %dKB
题目考察知识点: %s
代码:
%s

请严格按以下JSON格式返回：
{
  "summary": "一句话总结代码质量",
  "timeComplexity": "时间复杂度，用Markdown格式说明，如：**O(n log n)** — 使用了快速排序",
  "spaceComplexity": "空间复杂度，用Markdown格式说明，如：**O(n)** — 需要额外数组存储",
  "algorithmTags": ["使用的算法标签"],
  "codeStyle": "代码风格评价（1-2句）",
  "issues": [{"line": 行号, "severity": "warning/info", "message": "问题描述", "hint": "建议"}],
  "suggestions": ["优化建议1", "优化建议2"]
}`,
		req.ProblemTitle, req.Language, req.RuntimeMS, req.MemoryKB, topics, req.Code)

	resp, err := h.ai.Chat([]ai.Message{
		{Role: "system", Content: "你是一个资深的算法竞赛教练和代码审查专家。请始终以JSON格式回复，timeComplexity和spaceComplexity字段使用Markdown格式（如 **O(n)**）。"},
		{Role: "user", Content: prompt},
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "AI 服务暂时不可用，请稍后重试"})
		return
	}

	// Try to parse structured JSON response
	var structured map[string]interface{}
	if err := json.Unmarshal([]byte(resp), &structured); err == nil {
		structured["rawMarkdown"] = resp
		c.JSON(http.StatusOK, structured)
		return
	}

	c.JSON(http.StatusOK, gin.H{"analysis": resp, "rawMarkdown": resp})
}

// GenerateSolutionRequest is the request for generating a solution draft
type GenerateSolutionRequest struct {
	ProblemTitle string `json:"problemTitle"`
	ProblemDesc  string `json:"problemDesc"`
	Code         string `json:"code"`
	Language     string `json:"language"`
	Attempts     []AttemptInfo `json:"attempts"`
}

type AttemptInfo struct {
	Status  string `json:"status"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

// GenerateSolution generates a solution draft based on submission history
func (h *Handler) GenerateSolution(c *gin.Context) {
	var req GenerateSolutionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	attemptsDesc := ""
	for i, a := range req.Attempts {
		attemptsDesc += fmt.Sprintf("\n第%d次尝试: 状态=%s, 错误=%s", i+1, a.Status, a.Message)
	}

	prompt := fmt.Sprintf(`用户通过了一道算法题，请帮他生成一篇题解草稿。

题目: %s
题目描述: %s
最终通过的代码（%s）:
%s
用户的尝试历史:
%s

请生成一篇题解，包含：
1. 解题思路概述
2. 踩过的坑（从尝试历史中总结）
3. 实现亮点
4. 相关公式或关键算法（如状态转移方程等）
5. 时间/空间复杂度

用Markdown格式输出。`,
		req.ProblemTitle, req.ProblemDesc, req.Language, req.Code, attemptsDesc)

	resp, err := h.ai.Chat([]ai.Message{
		{Role: "system", Content: "你是一个善于写作的算法题解作者，能清晰地解释解题思路。"},
		{Role: "user", Content: prompt},
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "AI 服务暂时不可用，请稍后重试"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"solution": resp})
}

// ChatRequest is a general chat request
type ChatRequest struct {
	Messages []ai.Message `json:"messages"`
	Context  string       `json:"context"`
}

// ChatPayload is the AIOJ backend's chat request format (inside envelope)
type ChatPayload struct {
	UserID         uint64          `json:"userId"`
	ConversationID string          `json:"conversationId"`
	Message        string          `json:"message"`
	History        []ai.Message    `json:"history"`
	Problem        *ProblemPayload `json:"problem,omitempty"`
}

// PipelineEnvelope is the wrapper format used by AIOJ backend's AI client
type PipelineEnvelope struct {
	Task    string          `json:"task"`
	Model   string          `json:"model,omitempty"`
	Payload json.RawMessage `json:"payload"`
}

// CodeDiagnosisPayload is the payload for code diagnosis
type CodeDiagnosisPayload struct {
	UserID         uint64              `json:"userId"`
	ProblemID      uint64              `json:"problemId"`
	SubmissionID   uint64              `json:"submissionId,omitempty"`
	Language       string              `json:"language"`
	Code           string              `json:"code"`
	JudgeStatus    string              `json:"judgeStatus,omitempty"`
	ErrorMessage   string              `json:"errorMessage,omitempty"`
	RuntimeMs      int                 `json:"runtimeMs,omitempty"`
	MemoryKb       int                 `json:"memoryKb,omitempty"`
	ProblemTitle   string              `json:"problemTitle,omitempty"`
	ProblemDesc    string              `json:"problemContent,omitempty"`
	Editorial      string              `json:"editorial,omitempty"`
	Samples        []SampleData        `json:"samples,omitempty"`
	AlgorithmTags  []string            `json:"algorithmTags,omitempty"`
	RecentSubs     []SubmissionData    `json:"recentSubmissions,omitempty"`
	// Nested problem object (sent by AIOJ backend)
	Problem        *ProblemPayload     `json:"problem,omitempty"`
}

// ProblemPayload is the nested problem context from AIOJ backend
type ProblemPayload struct {
	ID              uint64      `json:"id"`
	Title           string      `json:"title"`
	Content         string      `json:"content,omitempty"`
	Editorial       string      `json:"editorial,omitempty"`
	Tags            []string    `json:"tags,omitempty"`
	Samples         []SampleData `json:"samples,omitempty"`
	Difficulty      string      `json:"difficulty,omitempty"`
	DifficultyScore int         `json:"difficultyScore,omitempty"`
}

type SampleData struct {
	Input    string `json:"input"`
	Expected string `json:"expected"`
}

type SubmissionData struct {
	ID           uint64 `json:"id"`
	Status       string `json:"status"`
	Code         string `json:"code,omitempty"`
	Language     string `json:"language"`
	ErrorMessage string `json:"errorMessage,omitempty"`
	CreatedAt    string `json:"createdAt"`
}

// SolvePayload is the payload for solve requests
type SolvePayload struct {
	UserID        uint64       `json:"userId"`
	ProblemID     uint64       `json:"problemId"`
	Question      string       `json:"question,omitempty"`
	Level         string       `json:"level"`
	ProblemTitle  string       `json:"problemTitle,omitempty"`
	ProblemDesc   string       `json:"problemContent,omitempty"`
	Editorial     string       `json:"editorial,omitempty"`
	Samples       []SampleData `json:"samples,omitempty"`
	AlgorithmTags []string     `json:"algorithmTags,omitempty"`
	Language      string       `json:"language,omitempty"`
	EditorCode    string       `json:"editorCode,omitempty"`
	// Frontend sends "code" instead of "editorCode"
	Code         string       `json:"code,omitempty"`
	JudgeError   string       `json:"judgeError,omitempty"`
	// Nested problem object (sent by AIOJ backend)
	Problem      *ProblemPayload `json:"problem,omitempty"`
}

// CodeDiagnosis handles code diagnosis requests (AIOJ backend compatible)
func (h *Handler) CodeDiagnosis(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": "failed to read body"})
		return
	}

	// Try envelope format first
	var envelope PipelineEnvelope
	if err := json.Unmarshal(body, &envelope); err == nil && envelope.Task != "" {
		var req CodeDiagnosisPayload
		if err := json.Unmarshal(envelope.Payload, &req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": "invalid payload"})
			return
		}
		h.handleDiagnosis(c, &req)
		return
	}

	// Try direct payload format
	var req CodeDiagnosisPayload
	if err := json.Unmarshal(body, &req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": "invalid request"})
		return
	}
	h.handleDiagnosis(c, &req)
}

func (h *Handler) handleDiagnosis(c *gin.Context, req *CodeDiagnosisPayload) {
	// Merge nested problem object into flat fields (AIOJ backend sends nested)
	mergeProblemFields(req)

	if req.Code == "" || req.Language == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": "code and language required"})
		return
	}

	// Build prompt with rich context
	prompt := ""
	if req.JudgeStatus == "Accepted" {
		prompt = fmt.Sprintf(`你是一个算法竞赛教练。用户提交了一道题目并且通过了评测，请分析代码质量。

题目: %s
题目描述: %s`, req.ProblemTitle, req.ProblemDesc)
	} else {
		prompt = fmt.Sprintf(`你是一个算法竞赛教练。用户提交了一道题目但结果不正确，请分析代码并给出诊断。

题目: %s
题目描述: %s`, req.ProblemTitle, req.ProblemDesc)
	}

	// Add samples if available
	if len(req.Samples) > 0 {
		prompt += "\n\n样例:\n"
		for i, s := range req.Samples {
			prompt += fmt.Sprintf("样例%d:\n  输入: %s\n  期望输出: %s\n", i+1, s.Input, s.Expected)
		}
	}

	// Add editorial if available
	if req.Editorial != "" {
		prompt += fmt.Sprintf("\n\n官方题解:\n%s", req.Editorial)
	}

	// Add recent submissions if available
	if len(req.RecentSubs) > 0 {
		prompt += "\n\n用户最近的提交记录:\n"
		for _, sub := range req.RecentSubs {
			prompt += fmt.Sprintf("- %s (%s): %s\n", sub.Status, sub.Language, sub.CreatedAt)
			if sub.Code != "" {
				prompt += fmt.Sprintf("  代码:\n%s\n", sub.Code)
			}
		}
	}

	prompt += fmt.Sprintf(`

用户语言: %s
评测状态: %s`, req.Language, req.JudgeStatus)

	if req.JudgeStatus == "Accepted" && req.RuntimeMs > 0 {
		prompt += fmt.Sprintf("\n执行用时: %dms", req.RuntimeMs)
	}
	if req.JudgeStatus == "Accepted" && req.MemoryKb > 0 {
		prompt += fmt.Sprintf("\n内存消耗: %d KB", req.MemoryKb)
	}

	prompt += fmt.Sprintf(`
错误信息: %s
当前用户代码:
%s`, req.ErrorMessage, req.Code)

	if req.JudgeStatus == "Accepted" {
		prompt += `

重要提示：该代码已通过全部测试用例，评测结果为 Accepted。请不要质疑代码的正确性，不要假设存在格式问题（如缺少换行符）。请专注于分析算法思路、复杂度和优化空间。

请严格按以下JSON格式返回（不要包含其他文本）：
{
  "summary": "一句话总结代码质量和算法思路",
  "timeComplexity": "时间复杂度，用Markdown格式说明，如：**O(n)** — 说明",
  "spaceComplexity": "空间复杂度，用Markdown格式说明，如：**O(n)** — 说明",
  "algorithmTags": ["使用的算法标签"],
  "issues": [],
  "suggestions": ["可选的优化建议或代码风格改进"]
}`
	} else {
		prompt += `

请严格按以下JSON格式返回（不要包含其他文本）：
{
  "summary": "一句话总结问题",
  "timeComplexity": "时间复杂度，用Markdown格式说明，如：**O(n²)** — 双重循环",
  "spaceComplexity": "空间复杂度，用Markdown格式说明，如：**O(n)** — 需要额外数组",
  "algorithmTags": ["涉及的算法标签"],
  "issues": [{"line": 行号, "severity": "error/warning/info", "message": "问题描述", "hint": "修复建议"}],
  "suggestions": ["改进建议1", "改进建议2"]
}`
	}

	systemMsg := "你是一个资深的算法竞赛教练，善于分析代码问题并给出精准的改进建议。请始终以JSON格式回复，timeComplexity和spaceComplexity字段使用Markdown格式（如 **O(n)** ）。"
	if req.JudgeStatus == "Accepted" {
		systemMsg = "你是一个资深的算法竞赛教练和代码审查专家。用户提交的代码已通过全部测试用例（Accepted），请分析代码的算法思路、时间空间复杂度和可优化空间。注意：代码是正确的，不要质疑其正确性，不要假设缺少换行符或其他格式问题。请始终以JSON格式回复，timeComplexity和spaceComplexity字段使用Markdown格式（如 **O(n)** ）。"
	}

	t0 := time.Now()
	resp, err := h.ai.Chat([]ai.Message{
		{Role: "system", Content: systemMsg},
		{Role: "user", Content: prompt},
	})
	log.Printf("[ai] code-diagnosis LLM call took %v", time.Since(t0))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "data": gin.H{
			"summary":     "AI 服务暂时不可用",
			"rawMarkdown": "AI 服务暂时不可用，请稍后重试。",
			"provider":    "unavailable",
		}})
		return
	}

	// Try to parse structured JSON response
	var structured map[string]interface{}
	if err := json.Unmarshal([]byte(resp), &structured); err == nil {
		structured["rawMarkdown"] = resp
		structured["provider"] = "agent-service"
		c.JSON(http.StatusOK, gin.H{"code": 0, "data": structured})
		return
	}

	// Fallback: return raw markdown
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": gin.H{
		"summary":     "代码诊断完成",
		"rawMarkdown": resp,
		"provider":    "agent-service",
	}})
}

// mergeProblemFields merges the nested Problem object into flat payload fields.
// AIOJ backend sends problem context as a nested object; agent-service expects flat fields.
func mergeProblemFields(req *CodeDiagnosisPayload) {
	if req.Problem == nil {
		return
	}
	p := req.Problem
	if req.ProblemID == 0 {
		req.ProblemID = p.ID
	}
	if req.ProblemTitle == "" {
		req.ProblemTitle = p.Title
	}
	if req.ProblemDesc == "" {
		req.ProblemDesc = p.Content
	}
	if req.Editorial == "" {
		req.Editorial = p.Editorial
	}
	if len(req.Samples) == 0 && len(p.Samples) > 0 {
		req.Samples = p.Samples
	}
	if len(req.AlgorithmTags) == 0 && len(p.Tags) > 0 {
		req.AlgorithmTags = p.Tags
	}
}

// mergeSolveProblemFields merges the nested Problem object into flat solve payload fields.
func mergeSolveProblemFields(req *SolvePayload) {
	if req.Problem == nil {
		return
	}
	p := req.Problem
	if req.ProblemID == 0 {
		req.ProblemID = p.ID
	}
	if req.ProblemTitle == "" {
		req.ProblemTitle = p.Title
	}
	if req.ProblemDesc == "" {
		req.ProblemDesc = p.Content
	}
	if req.Editorial == "" {
		req.Editorial = p.Editorial
	}
	if len(req.Samples) == 0 && len(p.Samples) > 0 {
		req.Samples = p.Samples
	}
	if len(req.AlgorithmTags) == 0 && len(p.Tags) > 0 {
		req.AlgorithmTags = p.Tags
	}
}

// KnowledgeGraphPayload is the payload for knowledge graph generation
type KnowledgeGraphPayload struct {
	Scope    string                   `json:"scope"`
	Problems []ProblemData            `json:"problems,omitempty"`
	TagStats map[string]TagStatsData  `json:"tagStats,omitempty"`
}

type ProblemData struct {
	Title    string   `json:"title"`
	Tags     []string `json:"tags"`
	Status   string   `json:"status"`
	Attempts int      `json:"attempts"`
}

type TagStatsData struct {
	Solved   int     `json:"solved"`
	Attempted int    `json:"attempted"`
	ACRate   float64 `json:"acRate"`
}

// KnowledgeGraph handles knowledge graph requests
func (h *Handler) KnowledgeGraph(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": "failed to read body"})
		return
	}

	// Try envelope format first
	var envelope PipelineEnvelope
	var req KnowledgeGraphPayload
	if err := json.Unmarshal(body, &envelope); err == nil && envelope.Task != "" {
		if err := json.Unmarshal(envelope.Payload, &req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": "invalid payload"})
			return
		}
	} else {
		if err := json.Unmarshal(body, &req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": "invalid request"})
			return
		}
	}

	// Build prompt for knowledge graph generation
	prompt := "根据用户的做题记录，分析用户的算法知识掌握情况，生成知识图谱。\n\n"
	if req.Scope == "recent" {
		prompt += "分析范围：最近的做题记录。\n"
	}
	if len(req.Problems) > 0 {
		prompt += "\n用户做过的题目：\n"
		for _, p := range req.Problems {
			prompt += fmt.Sprintf("- %s (标签: %s, 状态: %s, 尝试次数: %d)\n", p.Title, strings.Join(p.Tags, ", "), p.Status, p.Attempts)
		}
	}
	if len(req.TagStats) > 0 {
		prompt += "\n各知识点掌握情况：\n"
		for tag, stats := range req.TagStats {
			prompt += fmt.Sprintf("- %s: 解决 %d/%d 题, 通过率 %.1f%%\n", tag, stats.Solved, stats.Attempted, stats.ACRate)
		}
	}
	prompt += `

请严格按照以下JSON格式返回知识图谱数据：
{
  "summary": "对用户知识掌握情况的简要分析",
  "nodes": [
    {"id": "标签名", "label": "标签名", "mastery": 0-100, "category": "分类"}
  ],
  "edges": [
    {"source": "标签A", "target": "标签B", "type": "related/contains/prerequisite"}
  ],
  "suggestions": ["建议1", "建议2"]
}`

	resp, err := h.ai.Chat([]ai.Message{
		{Role: "system", Content: "你是一个算法竞赛教练，擅长分析学生的知识掌握情况并生成知识图谱。请始终以JSON格式回复。"},
		{Role: "user", Content: prompt},
	})
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "data": gin.H{
			"summary":     "AI 服务暂时不可用",
			"nodes":       []gin.H{},
			"edges":       []gin.H{},
			"rawMarkdown": "AI 服务暂时不可用，请稍后重试。",
			"provider":    "unavailable",
		}})
		return
	}

	// Try to parse structured JSON response
	var structured map[string]interface{}
	if err := json.Unmarshal([]byte(resp), &structured); err == nil {
		structured["rawMarkdown"] = resp
		structured["provider"] = "agent-service"
		c.JSON(http.StatusOK, gin.H{"code": 0, "data": structured})
		return
	}

	// Fallback: return raw markdown
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": gin.H{
		"summary":     "知识图谱分析完成",
		"nodes":       []gin.H{},
		"edges":       []gin.H{},
		"rawMarkdown": resp,
		"provider":    "agent-service",
	}})
}

// Solve handles solve requests (hint/explain/full levels)
func (h *Handler) Solve(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": "failed to read body"})
		return
	}

	// Try envelope format first
	var envelope PipelineEnvelope
	if err := json.Unmarshal(body, &envelope); err == nil && envelope.Task != "" {
		var req SolvePayload
		if err := json.Unmarshal(envelope.Payload, &req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": "invalid payload"})
			return
		}
		h.handleSolve(c, &req)
		return
	}

	// Try direct payload format
	var req SolvePayload
	if err := json.Unmarshal(body, &req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": "invalid request"})
		return
	}
	h.handleSolve(c, &req)
}

func (h *Handler) handleSolve(c *gin.Context, req *SolvePayload) {
	// Merge nested problem object into flat fields (AIOJ backend sends nested)
	mergeSolveProblemFields(req)

	// Frontend sends "code", AIOJ backend sends "editorCode"
	if req.EditorCode == "" && req.Code != "" {
		req.EditorCode = req.Code
	}

	level := req.Level
	if level == "" {
		level = "hint"
	}

	systemPrompt := "你是一个算法竞赛教练。请始终以JSON格式回复，timeComplexity和spaceComplexity字段使用Markdown格式（如 **O(n)** ）。"
	userPrompt := ""

	// Build context
	contextInfo := fmt.Sprintf("题目: %s\n题目描述: %s", req.ProblemTitle, req.ProblemDesc)

	// Add algorithm tags if available
	if len(req.AlgorithmTags) > 0 {
		contextInfo += fmt.Sprintf("\n\n算法标签: %s", strings.Join(req.AlgorithmTags, ", "))
	}

	// Add samples if available
	if len(req.Samples) > 0 {
		contextInfo += "\n\n样例:\n"
		for i, s := range req.Samples {
			contextInfo += fmt.Sprintf("样例%d:\n  输入: %s\n  期望输出: %s\n", i+1, s.Input, s.Expected)
		}
	}

	// Add editorial if available
	if req.Editorial != "" {
		contextInfo += fmt.Sprintf("\n\n官方题解:\n%s", req.Editorial)
	}

	// Add editor code if available
	if req.EditorCode != "" {
		contextInfo += fmt.Sprintf("\n\n用户当前代码:\n%s", req.EditorCode)
	}

	// Add judge error if available (for retry)
	if req.JudgeError != "" {
		contextInfo += fmt.Sprintf("\n\n之前的判题结果: %s\n请修正代码使其通过判题。", req.JudgeError)
	}

	switch level {
	case "hint":
		if req.EditorCode == "" {
			userPrompt = contextInfo + "\n\n用户还没有编写代码。请仅根据题目信息给出提示，帮助用户理解需要用到的算法知识点。"
		} else {
			userPrompt = contextInfo + "\n\n请根据用户的代码给出提示，帮助用户发现问题并改进。"
		}
		userPrompt += "\n\n请严格按以下JSON格式返回：\n{\"answer\": \"提示内容（不超过3句话）\", \"hints\": [\"提示1\", \"提示2\"], \"relatedTopics\": [\"相关知识点\"]}"
	case "explain":
		userPrompt = contextInfo + fmt.Sprintf("\n\n用户的问题: %s\n请严格按以下JSON格式返回：\n{\"answer\": \"解题思路解释，用Markdown格式\", \"hints\": [\"关键步骤1\", \"关键步骤2\"], \"relatedTopics\": [\"相关算法\"], \"timeComplexity\": \"**O(?)** — 说明\", \"spaceComplexity\": \"**O(?)** — 说明\"}", req.Question)
	case "full":
		if req.JudgeError != "" {
			userPrompt = contextInfo + "\n\n请修正代码使其通过判题。请严格按以下JSON格式返回（代码放在answer字段中，用Markdown代码块格式）：\n{\"answer\": \"修正后的代码和解释\", \"code\": \"修正后的完整代码\", \"language\": \"cpp\", \"hints\": [\"关键点1\"], \"relatedTopics\": [\"算法标签\"], \"timeComplexity\": \"**O(?)** — 说明\", \"spaceComplexity\": \"**O(?)** — 说明\"}"
		} else {
			userPrompt = contextInfo + fmt.Sprintf("\n\n用户的问题: %s\n请严格按以下JSON格式返回（代码放在answer字段中，用Markdown代码块格式，并在code字段中单独返回纯代码）：\n{\"answer\": \"完整解题思路和参考代码\", \"code\": \"完整可运行的代码\", \"language\": \"cpp\", \"hints\": [\"关键点1\"], \"relatedTopics\": [\"算法标签\"], \"timeComplexity\": \"**O(?)** — 说明\", \"spaceComplexity\": \"**O(?)** — 说明\"}", req.Question)
		}
	default:
		userPrompt = fmt.Sprintf("题目: %s\n%s", req.ProblemTitle, req.Question)
	}

	t0 := time.Now()
	resp, err := h.ai.Chat([]ai.Message{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: userPrompt},
	})
	log.Printf("[ai] solve LLM call took %v (level=%s)", time.Since(t0), level)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "data": gin.H{
			"answer":   "AI 服务暂时不可用，请稍后重试。",
			"hints":    []string{},
			"provider": "unavailable",
		}})
		return
	}

	// Try to parse structured JSON response
	var structured map[string]interface{}
	isStructured := json.Unmarshal([]byte(resp), &structured) == nil

	// Extract code from structured response if available
	if isStructured && level == "full" {
		if codeField, ok := structured["code"].(string); ok && codeField != "" {
			structured["code"] = codeField
			structured["language"] = "cpp"
		}
	}

	// If level is "full" and we have a problem ID, try to verify the generated code
	if level == "full" && req.ProblemID > 0 {
		code := ""
		if isStructured {
			if codeField, ok := structured["code"].(string); ok {
				code = codeField
			}
		}
		if code == "" {
			code = extractCodeBlock(resp, "cpp")
		}
		if code == "" {
			code = extractCodeBlock(resp, "c++")
		}
		if code != "" {
			result, err := h.judge.Submit(req.ProblemID, "cpp", code)
			if err == nil && result > 0 {
				// Poll for result up to 10 times (500ms interval, 5s total)
				var subResult *judge.SubmissionResult
				for i := 0; i < 10; i++ {
					time.Sleep(500 * time.Millisecond)
					subResult, err = h.judge.GetResult(result)
					if err != nil {
						break
					}
					if subResult.Status != "Pending" && subResult.Status != "Queueing" && subResult.Status != "Compiling" {
						break
					}
				}
				if err == nil && subResult != nil {
					verifyMsg := fmt.Sprintf("代码已自动提交验证：%s", subResult.Status)
					if subResult.Status != "Accepted" && subResult.ErrorMsg != "" {
						verifyMsg += fmt.Sprintf("，错误信息: %s", subResult.ErrorMsg)
					}
					if isStructured {
						structured["verifyResult"] = verifyMsg
					} else {
						resp += fmt.Sprintf("\n\n---\n### 🤖 AI 自检结果\n\n%s", verifyMsg)
					}
				}
			}
		}
	}

	if isStructured {
		structured["rawMarkdown"] = resp
		structured["provider"] = "agent-service"
		c.JSON(http.StatusOK, gin.H{"code": 0, "data": structured})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "data": gin.H{
		"answer":      resp,
		"hints":       []string{},
		"rawMarkdown": resp,
		"provider":    "agent-service",
	}})
}

// extractCodeBlock extracts code from a markdown code block with the given language tag
func extractCodeBlock(markdown, lang string) string {
	startTag := "```" + lang
	// Find the code block
	lines := splitLines(markdown)
	inBlock := false
	var codeLines []string
	for _, line := range lines {
		if !inBlock && len(line) >= len(startTag) && line[:len(startTag)] == startTag {
			inBlock = true
			continue
		}
		if inBlock && len(line) >= 3 && line[:3] == "```" {
			break
		}
		if inBlock {
			codeLines = append(codeLines, line)
		}
	}
	if len(codeLines) == 0 {
		return ""
	}
	result := ""
	for i, l := range codeLines {
		if i > 0 {
			result += "\n"
		}
		result += l
	}
	return result
}

func splitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}

// Chat handles general AI chat with optional context
func (h *Handler) Chat(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read body"})
		return
	}

	var req ChatRequest

	// Try envelope format first (AIOJ backend)
	var envelope PipelineEnvelope
	if err := json.Unmarshal(body, &envelope); err == nil && envelope.Task != "" {
		var payload ChatPayload
		if err := json.Unmarshal(envelope.Payload, &payload); err == nil {
			req = h.chatPayloadToRequest(&payload)
		}
	}

	// Try direct ChatRequest format
	if len(req.Messages) == 0 {
		if err := json.Unmarshal(body, &req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}
	}

	if len(req.Messages) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "messages required"})
		return
	}

	// Build system message with RAG context if available
	systemContent := "你是一个算法竞赛AI助手，擅长算法、数据结构、编程竞赛相关问题。请用中文回答。"

	// Add RAG context if available
	if h.rag != nil && h.rag.IsInitialized() {
		lastMsg := ""
		for i := len(req.Messages) - 1; i >= 0; i-- {
			if req.Messages[i].Role == "user" {
				lastMsg = req.Messages[i].Content
				break
			}
		}
		if lastMsg != "" {
			ragContext := h.rag.BuildContext(lastMsg, 2000)
			if ragContext != "" {
				systemContent += "\n\n" + ragContext
			}
		}
	}

	if req.Context != "" {
		systemContent += "\n\n以下是当前的上下文信息:\n" + req.Context
	}

	systemMsg := ai.Message{
		Role:    "system",
		Content: systemContent,
	}
	req.Messages = append([]ai.Message{systemMsg}, req.Messages...)

	t0 := time.Now()
	resp, err := h.ai.Chat(req.Messages)
	log.Printf("[ai] chat LLM call took %v (messages=%d)", time.Since(t0), len(req.Messages))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "AI 服务暂时不可用，请稍后重试"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"reply": resp, "response": resp})
}

// chatPayloadToRequest converts AIOJ backend's ChatPayload to ChatRequest format.
func (h *Handler) chatPayloadToRequest(payload *ChatPayload) ChatRequest {
	var messages []ai.Message

	// Convert history to messages
	for _, m := range payload.History {
		if m.Content != "" {
			messages = append(messages, ai.Message{Role: m.Role, Content: m.Content})
		}
	}

	// Add current message
	if payload.Message != "" {
		messages = append(messages, ai.Message{Role: "user", Content: payload.Message})
	}

	// Build context from problem info
	var contextParts []string
	if payload.Problem != nil {
		if payload.Problem.Title != "" {
			contextParts = append(contextParts, fmt.Sprintf("题目: %s", payload.Problem.Title))
		}
		if payload.Problem.Content != "" {
			contextParts = append(contextParts, fmt.Sprintf("题面:\n%s", payload.Problem.Content))
		}
		if len(payload.Problem.Samples) > 0 {
			sampleStr := "样例:\n"
			for i, s := range payload.Problem.Samples {
				sampleStr += fmt.Sprintf("样例%d:\n  输入: %s\n  期望输出: %s\n", i+1, s.Input, s.Expected)
			}
			contextParts = append(contextParts, sampleStr)
		}
		if len(payload.Problem.Tags) > 0 {
			contextParts = append(contextParts, fmt.Sprintf("标签: %s", strings.Join(payload.Problem.Tags, ", ")))
		}
	}

	return ChatRequest{
		Messages: messages,
		Context:  strings.Join(contextParts, "\n\n"),
	}
}
