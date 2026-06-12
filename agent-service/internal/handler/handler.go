package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"agent-service/internal/ai"
	"agent-service/internal/judge"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	ai    *ai.Client
	judge *judge.Client
}

func New(aiClient *ai.Client, judgeClient *judge.Client) *Handler {
	return &Handler{ai: aiClient, judge: judgeClient}
}

func (h *Handler) Health(c *gin.Context) {
	status := gin.H{"status": "ok"}
	if err := h.ai.Health(); err != nil {
		status["ollama"] = "unreachable: " + err.Error()
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

请给出一个简短的提示（不超过3句话），帮助用户思考正确的方向。不要包含完整代码。`,
		req.ProblemTitle, req.ProblemDesc, req.Language, req.Status, req.ErrorInfo, req.Code)

	resp, err := h.ai.Chat([]ai.Message{
		{Role: "system", Content: "你是一个友好的算法竞赛教练，善于用启发式的方式引导学生思考。"},
		{Role: "user", Content: prompt},
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "AI service error: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"hint": resp})
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

	prompt := fmt.Sprintf(`分析以下通过的代码，给出：
1. 代码风格评价（1-2句）
2. 时间复杂度分析
3. 空间复杂度分析
4. 使用的知识点 vs 题目考察的知识点（题目考察: %s）
5. 优化方向建议

题目: %s
语言: %s
运行时间: %dms
内存: %dKB
代码:
%s

请用简洁的中文回答，每点不超过2句话。`,
		topics, req.ProblemTitle, req.Language, req.RuntimeMS, req.MemoryKB, req.Code)

	resp, err := h.ai.Chat([]ai.Message{
		{Role: "system", Content: "你是一个资深的算法竞赛教练和代码审查专家。"},
		{Role: "user", Content: prompt},
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "AI service error: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"analysis": resp})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "AI service error: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"solution": resp})
}

// ChatRequest is a general chat request
type ChatRequest struct {
	Messages []ai.Message `json:"messages"`
	Context  string       `json:"context"`
}

// PipelineEnvelope is the wrapper format used by AIOJ backend's AI client
type PipelineEnvelope struct {
	Task    string          `json:"task"`
	Model   string          `json:"model,omitempty"`
	Payload json.RawMessage `json:"payload"`
}

// CodeDiagnosisPayload is the payload for code diagnosis
type CodeDiagnosisPayload struct {
	UserID       uint64 `json:"userId"`
	ProblemID    uint64 `json:"problemId"`
	SubmissionID uint64 `json:"submissionId,omitempty"`
	Language     string `json:"language"`
	Code         string `json:"code"`
	JudgeStatus  string `json:"judgeStatus,omitempty"`
	ErrorMessage string `json:"errorMessage,omitempty"`
	ProblemTitle string `json:"problemTitle,omitempty"`
	ProblemDesc  string `json:"problemContent,omitempty"`
}

// SolvePayload is the payload for solve requests
type SolvePayload struct {
	UserID   uint64 `json:"userId"`
	ProblemID uint64 `json:"problemId"`
	Question string `json:"question,omitempty"`
	Level    string `json:"level"`
	ProblemTitle string `json:"problemTitle,omitempty"`
	ProblemDesc  string `json:"problemContent,omitempty"`
}

// CodeDiagnosis handles code diagnosis requests (AIOJ backend compatible)
func (h *Handler) CodeDiagnosis(c *gin.Context) {
	var envelope PipelineEnvelope
	if err := c.ShouldBindJSON(&envelope); err != nil {
		// Try direct binding (non-envelope format)
		var req CodeDiagnosisPayload
		if err2 := c.ShouldBindJSON(&req); err2 != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": "invalid request"})
			return
		}
		h.handleDiagnosis(c, &req)
		return
	}
	var req CodeDiagnosisPayload
	if err := json.Unmarshal(envelope.Payload, &req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": "invalid payload"})
		return
	}
	h.handleDiagnosis(c, &req)
}

func (h *Handler) handleDiagnosis(c *gin.Context, req *CodeDiagnosisPayload) {
	if req.Code == "" || req.Language == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": "code and language required"})
		return
	}

	prompt := fmt.Sprintf(`你是一个算法竞赛教练。用户提交了一道题目但结果不正确，请分析代码并给出诊断。

题目: %s
题目描述: %s
用户语言: %s
评测状态: %s
错误信息: %s
用户代码:
%s

请给出：
1. 问题诊断（代码哪里有问题）
2. 改进建议
3. 相关知识点提示

用中文回答，格式用Markdown。`,
		req.ProblemTitle, req.ProblemDesc, req.Language, req.JudgeStatus, req.ErrorMessage, req.Code)

	resp, err := h.ai.Chat([]ai.Message{
		{Role: "system", Content: "你是一个资深的算法竞赛教练，善于分析代码问题并给出精准的改进建议。"},
		{Role: "user", Content: prompt},
	})
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 0, "data": gin.H{
			"summary":     "AI 服务暂时不可用",
			"rawMarkdown": "AI 服务暂时不可用，请稍后重试。",
			"provider":    "unavailable",
		}})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "data": gin.H{
		"summary":     "代码诊断完成",
		"rawMarkdown": resp,
		"provider":    "agent-service",
	}})
}

// KnowledgeGraph handles knowledge graph requests
func (h *Handler) KnowledgeGraph(c *gin.Context) {
	var envelope PipelineEnvelope
	if err := c.ShouldBindJSON(&envelope); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": "invalid request"})
		return
	}
	// For now, return a basic response - full RAG integration would go here
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": gin.H{
		"summary":     "知识图谱功能正在完善中",
		"nodes":       []gin.H{},
		"edges":       []gin.H{},
		"rawMarkdown": "知识图谱功能正在完善中，请使用独立的知识图谱页面查看。",
		"provider":    "agent-service",
	}})
}

// Solve handles solve requests (hint/explain/full levels)
func (h *Handler) Solve(c *gin.Context) {
	var envelope PipelineEnvelope
	if err := c.ShouldBindJSON(&envelope); err != nil {
		// Try direct binding
		var req SolvePayload
		if err2 := c.ShouldBindJSON(&req); err2 != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": "invalid request"})
			return
		}
		h.handleSolve(c, &req)
		return
	}
	var req SolvePayload
	if err := json.Unmarshal(envelope.Payload, &req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": "invalid payload"})
		return
	}
	h.handleSolve(c, &req)
}

func (h *Handler) handleSolve(c *gin.Context, req *SolvePayload) {
	level := req.Level
	if level == "" {
		level = "hint"
	}

	systemPrompt := "你是一个算法竞赛教练。"
	userPrompt := ""

	switch level {
	case "hint":
		userPrompt = fmt.Sprintf("题目: %s\n用户请求提示。请给出一个简短的提示（不超过3句话），帮助用户思考正确的方向，但不要给出完整解法。", req.ProblemTitle)
	case "explain":
		userPrompt = fmt.Sprintf("题目: %s\n用户的问题: %s\n请解释解题思路，包含关键算法和步骤，但不给出完整代码。", req.ProblemTitle, req.Question)
	case "full":
		userPrompt = fmt.Sprintf("题目: %s\n题目描述: %s\n用户的问题: %s\n请给出完整的解题思路和参考代码（C++）。代码用 ```cpp 包裹。", req.ProblemTitle, req.ProblemDesc, req.Question)
	default:
		userPrompt = fmt.Sprintf("题目: %s\n%s", req.ProblemTitle, req.Question)
	}

	resp, err := h.ai.Chat([]ai.Message{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: userPrompt},
	})
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 0, "data": gin.H{
			"answer":   "AI 服务暂时不可用，请稍后重试。",
			"hints":    []string{},
			"provider": "unavailable",
		}})
		return
	}

	// If level is "full" and we have a problem ID, try to verify the generated code
	if level == "full" && req.ProblemID > 0 {
		code := extractCodeBlock(resp, "cpp")
		if code == "" {
			code = extractCodeBlock(resp, "c++")
		}
		if code != "" {
			result, err := h.judge.Submit(req.ProblemID, "cpp", code)
			if err == nil && result > 0 {
				// Poll for result
				time.Sleep(2 * time.Second)
				subResult, err := h.judge.GetResult(result)
				if err == nil {
					resp += fmt.Sprintf("\n\n---\n### 🤖 AI 自检结果\n\n代码已自动提交验证：%s", subResult.Status)
					if subResult.Status != "Accepted" && subResult.ErrorMsg != "" {
						resp += fmt.Sprintf("\n错误信息: %s", subResult.ErrorMsg)
					}
				}
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "data": gin.H{
		"answer":   resp,
		"hints":    []string{},
		"provider": "agent-service",
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
	var req ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	if req.Context != "" {
		systemMsg := ai.Message{
			Role:    "system",
			Content: "你是一个算法竞赛AI助手。以下是当前的上下文信息:\n" + req.Context,
		}
		req.Messages = append([]ai.Message{systemMsg}, req.Messages...)
	}

	resp, err := h.ai.Chat(req.Messages)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "AI service error: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"response": resp})
}
