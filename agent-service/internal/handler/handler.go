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
	"agent-service/internal/rag"

	"github.com/gin-gonic/gin"
)

// CandidateTagDict is the unified algorithm tag dictionary.
// LLM outputs for algorithmTags MUST be selected from this list.
const CandidateTagDict = "枚举、模拟、排序、二分、双指针、前缀和、差分、分治、贪心、递归、离散化、数组、链表、栈、单调栈、队列、单调队列、堆、哈希表、并查集、字典树、线段树、树状数组、平衡树、分块、动态规划、背包DP、区间DP、树形DP、数位DP、状态压缩DP、DP优化、计数DP、概率DP、博弈论DP、图论、最短路、最小生成树、网络流、二分图、拓扑排序、强连通分量、桥和割点、树上问题、LCA、数学、质数、GCD/LCM、快速幂、模运算、组合数学、容斥原理、概率期望、矩阵、高斯消元、莫比乌斯反演、博弈论、字符串、字符串处理、KMP、Trie、后缀数组、后缀自动机、AC自动机、Manacher、哈希、搜索、BFS、DFS、迭代加深、IDA*、双向BFS、启发式搜索、折半搜索、回溯、贪心算法、区间贪心、排序贪心、反悔贪心、计算几何、向量、凸包、半平面交、最近点对、旋转卡壳、位运算、位操作、状态压缩、集合运算"

type Handler struct {
	ai  *ai.Client
	rag *rag.Service
}

func New(aiClient *ai.Client, ragService *rag.Service) *Handler {
	return &Handler{ai: aiClient, rag: ragService}
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

// GenerateSolutionPayload is the payload for generating a solution draft
type GenerateSolutionPayload struct {
	ProblemID     uint64       `json:"problemId"`
	ProblemTitle  string       `json:"problemTitle"`
	ProblemDesc   string       `json:"problemContent"`
	Editorial     string       `json:"problemEditorial,omitempty"`
	AlgorithmTags []string     `json:"algorithmTags,omitempty"`
	Language      string       `json:"language"`
	Code          string       `json:"code"`
	// Nested problem object (sent by AIOJ backend)
	Problem       *ProblemPayload `json:"problem,omitempty"`
}

// GenerateSolution generates a solution draft
func (h *Handler) GenerateSolution(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": "failed to read body"})
		return
	}

	// Try envelope format first
	var envelope PipelineEnvelope
	var req GenerateSolutionPayload
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

	// Merge nested problem object
	if req.Problem != nil {
		if req.ProblemID == 0 {
			req.ProblemID = req.Problem.ID
		}
		if req.ProblemTitle == "" {
			req.ProblemTitle = req.Problem.Title
		}
		if req.ProblemDesc == "" {
			req.ProblemDesc = req.Problem.Content
		}
		if req.Editorial == "" {
			req.Editorial = req.Problem.Editorial
		}
		if len(req.AlgorithmTags) == 0 && len(req.Problem.Tags) > 0 {
			req.AlgorithmTags = req.Problem.Tags
		}
	}

	// RAG retrieval using algorithm tags
	ragContext := ""
	if h.rag != nil && h.rag.IsInitialized() && len(req.AlgorithmTags) > 0 {
		ragContext = h.rag.BuildContext(strings.Join(req.AlgorithmTags, " "), 2000)
	}

	// Build prompt
	systemPrompt := "你是一个算法竞赛题解撰写专家。你的任务是帮助用户基于其通过的代码编写一篇高质量的题解。要求：文风严谨、逻辑清晰、语言简洁，不要使用emoji，不要有AI套话（如'首先'、'其次'、'总之'等过渡词），直接切入要点，用科学论文般的简洁风格。请始终以JSON格式回复。"

	userPrompt := fmt.Sprintf("题目: %s\n题目描述: %s", req.ProblemTitle, req.ProblemDesc)

	if req.Editorial != "" {
		userPrompt += fmt.Sprintf("\n\n官方题解:\n%s", req.Editorial)
	}

	if len(req.AlgorithmTags) > 0 {
		userPrompt += fmt.Sprintf("\n\n算法标签: %s", strings.Join(req.AlgorithmTags, ", "))
	}

	if ragContext != "" {
		userPrompt += fmt.Sprintf("\n\n相关知识（从 OI-Wiki 检索）：\n---\n%s\n---", ragContext)
	}

	userPrompt += fmt.Sprintf(`

用户通过的代码（%s）:
%s`, req.Language, req.Code)

	userPrompt += fmt.Sprintf(`

请基于以上信息生成一篇题解草稿，包含：
1. 解题思路概述
2. 时间空间复杂度

输出要求：
- title：题解标题，简洁概括解法
- content：题解正文，用 Markdown 格式，直接切入要点，不要有AI套话
- algorithmTags：题解涉及的算法标签，必须从候选标签中选取
- complexity.time / complexity.space：纯大O表示法，用Markdown加粗（如 **O(n)**），不要附加说明文字

候选算法标签（algorithmTags 必须从中选取）：
%s

请严格按以下JSON格式返回（不要包含其他文本）：
{
  "title": "题解标题",
  "content": "题解内容（Markdown）",
  "algorithmTags": ["哈希表"],
  "complexity": {"time": "**O(n)**", "space": "**O(n)**"}
}`, CandidateTagDict)

	t0 := time.Now()
	resp, err := h.ai.Chat([]ai.Message{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: userPrompt},
	})
	log.Printf("[ai] generate-solution LLM call took %v", time.Since(t0))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "data": gin.H{
			"solution":    "AI 服务暂时不可用，请稍后重试。",
			"provider":    "unavailable",
		}})
		return
	}

	// Try to parse structured JSON response
	var structured map[string]interface{}
	if json.Unmarshal([]byte(resp), &structured) == nil {
		structured["rawMarkdown"] = resp
		structured["provider"] = "agent-service"
		c.JSON(http.StatusOK, gin.H{"code": 0, "data": structured})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "data": gin.H{
		"solution":    resp,
		"rawMarkdown": resp,
		"provider":    "agent-service",
	}})
}

// ChatRequest is a general chat request
type ChatRequest struct {
	Messages []ai.Message `json:"messages"`
	Context  string       `json:"context"`
	Tags     []string     `json:"-"` // For RAG query, not from JSON
}

// ChatPayload is the AIOJ backend's chat request format (inside envelope)
type ChatPayload struct {
	UserID         uint64          `json:"userId"`
	ConversationID string          `json:"conversationId"`
	Message        string          `json:"message"`
	History        []ai.Message    `json:"history"`
	Problem        *ProblemPayload `json:"problem,omitempty"`
	CodeLanguage   string          `json:"codeLanguage,omitempty"`
	Code           string          `json:"code,omitempty"`
}

// PipelineEnvelope is the wrapper format used by AIOJ backend's AI client
type PipelineEnvelope struct {
	Task    string          `json:"task"`
	Model   string          `json:"model,omitempty"`
	Payload json.RawMessage `json:"payload"`
}

// FailedCaseData represents the first failing test case
type FailedCaseData struct {
	Input    string `json:"input"`
	Expected string `json:"expected"`
	Actual   string `json:"actual"`
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
	FailedCase     *FailedCaseData     `json:"failedCase,omitempty"`
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

	// Add failed test case for non-AC submissions
	if req.FailedCase != nil && req.JudgeStatus != "Accepted" {
		prompt += fmt.Sprintf(`

未通过的测试点:
  输入: %s
  预期输出: %s
  实际输出: %s`, req.FailedCase.Input, req.FailedCase.Expected, req.FailedCase.Actual)
	}

	prompt += fmt.Sprintf(`

候选算法标签（algorithmTags 必须从中选取）：
%s`, CandidateTagDict)

	if req.JudgeStatus == "Accepted" {
		prompt += `

重要提示：该代码已通过全部测试用例。请不要质疑代码的正确性，不要假设存在格式问题。请专注于分析算法思路、复杂度和优化空间。

输出要求：
- timeComplexity：用户代码的时间复杂度，纯大O表示法，用Markdown加粗（如 **O(n)**），不要附加说明文字
- spaceComplexity：用户代码的空间复杂度，纯大O表示法，用Markdown加粗，不要附加说明文字
- algorithmTags：用户代码实际实现的算法标签，必须从候选标签中选取
- suggestions：一句话简要建议

请严格按以下JSON格式返回（不要包含其他文本）：
{
  "timeComplexity": "**O(n)**",
  "spaceComplexity": "**O(1)**",
  "algorithmTags": ["哈希表"],
  "suggestions": ["建议内容"]
}`
	} else {
		prompt += `

输出要求：
- timeComplexity：用户代码的时间复杂度，纯大O表示法，用Markdown加粗（如 **O(n)**），不要附加说明文字
- spaceComplexity：用户代码的空间复杂度，纯大O表示法，用Markdown加粗，不要附加说明文字
- algorithmTags：用户代码实际实现的算法标签，必须从候选标签中选取
- suggestions：一句话简要建议

请严格按以下JSON格式返回（不要包含其他文本）：
{
  "timeComplexity": "**O(n)**",
  "spaceComplexity": "**O(1)**",
  "algorithmTags": ["哈希表"],
  "suggestions": ["建议内容"]
}`
	}

	systemMsg := "你是一个算法竞赛教练。用户提交的代码未通过评测，请分析代码中的问题并给出改进建议。回答简洁，直接指出核心问题，不要泛泛而谈。请始终以JSON格式回复。"
	if req.JudgeStatus == "Accepted" {
		systemMsg = "你是一个算法竞赛教练和代码审查专家。用户提交的代码已通过全部测试用例，请分析代码的算法思路、时间空间复杂度和可优化空间。代码是正确的，不要质疑其正确性，不要假设存在格式问题。回答简洁，直接给出结论。请始终以JSON格式回复。"
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
	ID       uint64   `json:"id"`
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
	prompt += fmt.Sprintf(`

输出要求：
- nodes：知识图谱节点，每个节点包含 id、label（算法名称）、mastery（掌握等级）、category（所属分类）
  - mastery 为枚举类型，根据做题数量、AC率、题目难度综合判断：
    - unattempted：未接触 — 没做过相关题目
    - learning：初学 — 做过少量题目，还在摸索
    - familiar：了解 — 做过一些题目，基本理解思路
    - proficient：熟练 — 做过较多题目，能独立解题
    - mastered：精通 — 做过大量题目包括困难题，解题稳定
  - category 必须是以下分类之一：基础算法、数据结构、动态规划、图论、数学、字符串、搜索、贪心、计算几何、位运算
- edges：节点之间的关系边，包含 source、target、type（related/contains/prerequisite）
- suggestions：一句话建议列表，针对薄弱知识点

标签约束：节点的 id 和 label 必须来自统一标签字典：
%s

请严格按以下JSON格式返回（不要包含其他文本）：
{
  "nodes": [
    {"id": "哈希表", "label": "哈希表", "mastery": "proficient", "category": "基础算法"}
  ],
  "edges": [
    {"source": "数组", "target": "哈希表", "type": "related"}
  ],
  "suggestions": ["建议内容"]
}`, CandidateTagDict)

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

// CreateStudyPlanPayload is the payload for AI 题单创建
type CreateStudyPlanPayload struct {
	UserID     uint64              `json:"userId"`
	Problems   []ProblemData       `json:"problems"`
	TagStats   map[string]TagStatsData `json:"tagStats"`
	Candidates map[string][]ProblemData `json:"candidates"`
}

// CreateStudyPlan generates a study plan based on user's weak areas
func (h *Handler) CreateStudyPlan(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": "failed to read body"})
		return
	}
	var envelope PipelineEnvelope
	var req CreateStudyPlanPayload
	if json.Unmarshal(body, &envelope); err == nil && envelope.Task != "" {
		json.Unmarshal(envelope.Payload, &req)
	} else {
		json.Unmarshal(body, &req)
	}

	// Build prompt: analyze user's weak areas and select appropriate problems
	prompt := "根据用户的做题记录和薄弱知识点，创建一个针对性的学习题单。\n\n"
	if len(req.Problems) > 0 {
		prompt += "用户做过的题目：\n"
		for _, p := range req.Problems {
			prompt += fmt.Sprintf("- %s (标签: %s, 状态: %s)\n", p.Title, strings.Join(p.Tags, ", "), p.Status)
		}
	}
	if len(req.TagStats) > 0 {
		prompt += "\n知识点掌握情况：\n"
		for tag, stats := range req.TagStats {
			prompt += fmt.Sprintf("- %s: 通过率 %.1f%%, 解决 %d/%d 题\n", tag, stats.ACRate, stats.Solved, stats.Attempted)
		}
	}
	if len(req.Candidates) > 0 {
		prompt += "\n可选的推荐题目（按知识点分组）：\n"
		for tag, probs := range req.Candidates {
			prompt += fmt.Sprintf("\n【%s】:\n", tag)
			for _, p := range probs {
				prompt += fmt.Sprintf("  - #%d %s\n", p.ID, p.Title)
			}
		}
	}

	prompt += fmt.Sprintf(`
输出要求：
- title: 题单标题（简洁有吸引力）
- description: 题单描述（说明学习目标和涵盖的知识点）
- problemIDs: 从候选题目中选取的题目ID列表（按难度递进排列，5~15道）

请严格按以下JSON格式返回（不要包含其他文本）：
{
  "title": "题单标题",
  "description": "题单描述",
  "problemIDs": [1001, 1003, 1005]
}`)

	resp, err := h.ai.Chat([]ai.Message{
		{Role: "system", Content: "你是一个算法竞赛教练，擅长根据学生的学习情况制定个性化题单。请始终以JSON格式回复。分析学生的薄弱知识点，从候选题目中选择合适的题目组成题单。题目按难度递进排列，题单应覆盖2~4个薄弱知识点，包含5~15道题。"},
		{Role: "user", Content: prompt},
	})
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "data": gin.H{
			"title": "默认题单", "description": "AI服务不可用", "problemIDs": []uint64{}, "rawMarkdown": "", "provider": "unavailable",
		}})
		return
	}

	var structured map[string]interface{}
	if json.Unmarshal([]byte(resp), &structured) == nil {
		structured["rawMarkdown"] = resp
		structured["provider"] = "agent-service"
		c.JSON(http.StatusOK, gin.H{"code": 0, "data": structured})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "data": gin.H{
		"title": "AI 推荐题单", "description": resp, "problemIDs": []uint64{}, "rawMarkdown": resp, "provider": "agent-service",
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

	// Build context
	contextInfo := fmt.Sprintf("题目: %s\n题目描述: %s", req.ProblemTitle, req.ProblemDesc)

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
		contextInfo += fmt.Sprintf("\n\n之前的判题结果: %s", req.JudgeError)
	}

	// Level-specific system and user prompts
	var systemPrompt, userPrompt string

	switch level {
	case "hint":
		systemPrompt = "你是一个善于启发式教学的算法教练。用旁敲侧击的方式引导学生思考，不要直接给出答案或代码。一句话点到即止。请始终以JSON格式回复。"
		if len(req.AlgorithmTags) > 0 {
			contextInfo += fmt.Sprintf("\n\n算法标签: %s", strings.Join(req.AlgorithmTags, ", "))
		}
		userPrompt = contextInfo + "\n\n请用旁敲侧击的方式给出一句话提示，不要直接指出问题所在，而是引导用户自己发现关键点。如果用户没有代码，提示相关的算法方向即可。"
		userPrompt += "\n\n请严格按以下JSON格式返回（不要包含其他文本）：\n{\"answer\": \"一句话启发式提示\"}"

	case "explain":
		systemPrompt = "你是一个算法竞赛教练。请分析用户代码，直接指出最大的一个问题所在，可以引用具体代码位置，但不要给出解决方案，让用户自己思考。回答简洁。请始终以JSON格式回复。"
		userPrompt = contextInfo + "\n\n请分析用户当前代码，直接指出最大的一个问题所在（可以引用具体代码位置），但不要给出解决方案，让用户自己思考如何修改。回答简洁。"
		userPrompt += "\n\n请严格按以下JSON格式返回（不要包含其他文本）：\n{\"answer\": \"当前代码最大的问题（Markdown）\"}"

	case "full":
		systemPrompt = "你是一个算法竞赛教练和程序员。请根据题目要求生成完整可运行的代码。如果用户代码思路基本正确，尽量沿用用户的思路进行修正；如果用户思路有根本性错误，则生成其他正确的实现。回答简洁，代码完整。请始终以JSON格式回复。"
		if len(req.AlgorithmTags) > 0 {
			contextInfo += fmt.Sprintf("\n\n算法标签: %s", strings.Join(req.AlgorithmTags, ", "))
		}
		if req.JudgeError != "" {
			userPrompt = contextInfo + "\n\n之前的判题结果: " + req.JudgeError + "\n请根据判题结果修正代码。"
		} else {
			userPrompt = contextInfo
		}
		userPrompt += "\n\n请生成完整的可运行代码来解决这道题目。要求：如果用户代码的思路基本正确，尽量沿着用户的思路进行修正和完善；如果用户思路有根本性错误，则生成其他正确的实现。"
		userPrompt += "\n\n输出要求：\n- answer：简要说明解题思路\n- code：完整可运行的代码\n- language：编程语言\n- timeComplexity / spaceComplexity：纯大O表示法，用Markdown加粗，不要附加说明文字"
		userPrompt += "\n\n请严格按以下JSON格式返回（不要包含其他文本）：\n{\"answer\": \"解题思路简述\", \"code\": \"完整代码\", \"language\": \"cpp\", \"timeComplexity\": \"**O(n)**\", \"spaceComplexity\": \"**O(1)**\"}"

	default:
		systemPrompt = "你是一个算法竞赛教练。请始终以JSON格式回复。"
		userPrompt = contextInfo
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
			"provider": "unavailable",
		}})
		return
	}

	// Try to parse structured JSON response
	var structured map[string]interface{}
	if json.Unmarshal([]byte(resp), &structured) == nil {
		structured["rawMarkdown"] = resp
		structured["provider"] = "agent-service"
		c.JSON(http.StatusOK, gin.H{"code": 0, "data": structured})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "data": gin.H{
		"answer":      resp,
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
	systemContent := "你是一个算法竞赛AI助手，擅长算法、数据结构、编程竞赛相关问题。请用中文回答，回答简洁准确。"

	// Add RAG context using algorithm tags (not user message)
	if h.rag != nil && h.rag.IsInitialized() {
		ragQuery := ""
		if len(req.Tags) > 0 {
			ragQuery = strings.Join(req.Tags, " ")
		}
		if ragQuery != "" {
			ragContext := h.rag.BuildContext(ragQuery, 2000)
			if ragContext != "" {
				systemContent += "\n\n相关知识（从 OI-Wiki 检索）：\n---\n" + ragContext + "\n---"
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

	// Add code context if available
	if payload.Code != "" {
		lang := payload.CodeLanguage
		if lang == "" {
			lang = "代码"
		}
		contextParts = append(contextParts, fmt.Sprintf("用户当前代码（%s）：\n%s", lang, payload.Code))
	}

	var tags []string
	if payload.Problem != nil {
		tags = payload.Problem.Tags
	}

	return ChatRequest{
		Messages: messages,
		Context:  strings.Join(contextParts, "\n\n"),
		Tags:     tags,
	}
}
