package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"agent-service/internal/ai"
	"agent-service/internal/rag"
	"agent-service/internal/tools"

	"github.com/gin-gonic/gin"
)

// CandidateTagDict is the unified algorithm tag dictionary.
const CandidateTagDict = "枚举、模拟、排序、二分、双指针、前缀和、差分、分治、贪心、递归、离散化、数组、链表、栈、单调栈、队列、单调队列、堆、哈希表、并查集、字典树、线段树、树状数组、平衡树、分块、动态规划、背包DP、区间DP、树形DP、数位DP、状态压缩DP、DP优化、计数DP、概率DP、博弈论DP、图论、最短路、最小生成树、网络流、二分图、拓扑排序、强连通分量、桥和割点、树上问题、LCA、数学、质数、GCD/LCM、快速幂、模运算、组合数学、容斥原理、概率期望、矩阵、高斯消元、莫比乌斯反演、博弈论、字符串、字符串处理、KMP、Trie、后缀数组、后缀自动机、AC自动机、Manacher、哈希、搜索、BFS、DFS、迭代加深、IDA*、双向BFS、启发式搜索、折半搜索、回溯、贪心算法、区间贪心、排序贪心、反悔贪心、计算几何、向量、凸包、半平面交、最近点对、旋转卡壳、位运算、位操作、状态压缩、集合运算"

type Handler struct {
	ai       *ai.Client
	rag      *rag.Service
	toolExec *tools.Executor
}

func New(aiClient *ai.Client, ragService *rag.Service, ojBaseURL string) *Handler {
	return &Handler{
		ai:       aiClient,
		rag:      ragService,
		toolExec: tools.NewExecutor(ojBaseURL, ragService),
	}
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
	ProblemID     uint64          `json:"problemId"`
	ProblemTitle  string          `json:"problemTitle"`
	ProblemDesc   string          `json:"problemContent"`
	Editorial     string          `json:"problemEditorial,omitempty"`
	AlgorithmTags []string        `json:"algorithmTags,omitempty"`
	Language      string          `json:"language"`
	Code          string          `json:"code"`
	Problem       *ProblemPayload `json:"problem,omitempty"`
}

// GenerateSolution generates a solution draft
func (h *Handler) GenerateSolution(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": "failed to read body"})
		return
	}

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

	// 合并 Problem 字段
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

	// 验证必要字段
	if req.Code == "" || req.Language == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": "code and language required"})
		return
	}

	// 初始化 Agent 适配器
	agentAdapter, err := NewAgentAdapter(
		"knowledge.json",
		"knowledge_with_embedding.json",
		"records.json",
	)
	if err != nil {
		log.Printf("[agent] 初始化失败: %v", err)
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"data": gin.H{
				"solution": "AI 服务初始化失败",
				"provider": "unavailable",
			},
		})
		return
	}
	defer agentAdapter.Close()

	// 转换参数
	cfg := ConvertToStartConfig(&req)

	// 调用 Agent 生成辅导
	tutorial, err := agentAdapter.GenerateTutorial(cfg)
	if err != nil {
		log.Printf("[agent] 生成失败: %v", err)
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"data": gin.H{
				"solution": "AI 生成失败: " + err.Error(),
				"provider": "unavailable",
			},
		})
		return
	}

	// 转换为题解格式
	rawJSON, _ := json.Marshal(tutorial)
	solutionData := ConvertTutorialToSolutionResponse(tutorial, string(rawJSON))

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": solutionData,
	})
}

// ChatRequest is a general chat request
type ChatRequest struct {
	Messages []ai.Message `json:"messages"`
	Context  string       `json:"context"`
	Tags     []string     `json:"-"`
}

type ChatPayload struct {
	UserID         uint64           `json:"userId"`
	ConversationID string           `json:"conversationId"`
	Message        string           `json:"message"`
	History        []ai.Message     `json:"history"`
	Problem        *ProblemPayload  `json:"problem,omitempty"`
	ExtraProblems  []ProblemPayload `json:"extraProblems,omitempty"`
	CodeLanguage   string           `json:"codeLanguage,omitempty"`
	Code           string           `json:"code,omitempty"`
}

type PipelineEnvelope struct {
	Task    string          `json:"task"`
	Model   string          `json:"model,omitempty"`
	Payload json.RawMessage `json:"payload"`
}

type FailedCaseData struct {
	Input    string `json:"input"`
	Expected string `json:"expected"`
	Actual   string `json:"actual"`
}

type CodeDiagnosisPayload struct {
	UserID        uint64           `json:"userId"`
	ProblemID     uint64           `json:"problemId"`
	SubmissionID  uint64           `json:"submissionId,omitempty"`
	Language      string           `json:"language"`
	Code          string           `json:"code"`
	JudgeStatus   string           `json:"judgeStatus,omitempty"`
	ErrorMessage  string           `json:"errorMessage,omitempty"`
	RuntimeMs     int              `json:"runtimeMs,omitempty"`
	MemoryKb      int              `json:"memoryKb,omitempty"`
	ProblemTitle  string           `json:"problemTitle,omitempty"`
	ProblemDesc   string           `json:"problemContent,omitempty"`
	Editorial     string           `json:"editorial,omitempty"`
	Samples       []SampleData     `json:"samples,omitempty"`
	AlgorithmTags []string         `json:"algorithmTags,omitempty"`
	RecentSubs    []SubmissionData `json:"recentSubmissions,omitempty"`
	FailedCase    *FailedCaseData  `json:"failedCase,omitempty"`
	Problem       *ProblemPayload  `json:"problem,omitempty"`
}

type ProblemPayload struct {
	ID              uint64       `json:"id"`
	Title           string       `json:"title"`
	Content         string       `json:"content,omitempty"`
	Editorial       string       `json:"editorial,omitempty"`
	Tags            []string     `json:"tags,omitempty"`
	Samples         []SampleData `json:"samples,omitempty"`
	Difficulty      string       `json:"difficulty,omitempty"`
	DifficultyScore int          `json:"difficultyScore,omitempty"`
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

type SolvePayload struct {
	UserID        uint64          `json:"userId"`
	ProblemID     uint64          `json:"problemId"`
	Question      string          `json:"question,omitempty"`
	Level         string          `json:"level"`
	ProblemTitle  string          `json:"problemTitle,omitempty"`
	ProblemDesc   string          `json:"problemContent,omitempty"`
	Editorial     string          `json:"editorial,omitempty"`
	Samples       []SampleData    `json:"samples,omitempty"`
	AlgorithmTags []string        `json:"algorithmTags,omitempty"`
	Language      string          `json:"language,omitempty"`
	EditorCode    string          `json:"editorCode,omitempty"`
	Code          string          `json:"code,omitempty"`
	JudgeError    string          `json:"judgeError,omitempty"`
	Problem       *ProblemPayload `json:"problem,omitempty"`
}

// CodeDiagnosis handles code diagnosis (AIOJ backend compatible)
func (h *Handler) CodeDiagnosis(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": "failed to read body"})
		return
	}
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
	var req CodeDiagnosisPayload
	if err := json.Unmarshal(body, &req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": "invalid request"})
		return
	}
	h.handleDiagnosis(c, &req)
}

func (h *Handler) handleDiagnosis(c *gin.Context, req *CodeDiagnosisPayload) {
	mergeProblemFields(req)
	if req.Code == "" || req.Language == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": "code and language required"})
		return
	}
	prompt := fmt.Sprintf("题目: %s\n题目描述: %s", req.ProblemTitle, req.ProblemDesc)
	if len(req.Samples) > 0 {
		prompt += "\n\n样例:\n"
		for i, s := range req.Samples {
			prompt += fmt.Sprintf("样例%d:\n  输入: %s\n  期望输出: %s\n", i+1, s.Input, s.Expected)
		}
	}
	if req.Editorial != "" {
		prompt += fmt.Sprintf("\n\n官方题解:\n%s", req.Editorial)
	}
	if len(req.RecentSubs) > 0 {
		prompt += "\n\n用户最近的提交记录:\n"
		for _, sub := range req.RecentSubs {
			prompt += fmt.Sprintf("- %s (%s): %s\n", sub.Status, sub.Language, sub.CreatedAt)
		}
	}
	prompt += fmt.Sprintf("\n用户语言: %s\n评测状态: %s\n错误信息: %s\n当前用户代码:\n%s", req.Language, req.JudgeStatus, req.ErrorMessage, req.Code)
	if req.FailedCase != nil && req.JudgeStatus != "Accepted" {
		prompt += fmt.Sprintf("\n未通过的测试点:\n  输入: %s\n  预期输出: %s\n  实际输出: %s", req.FailedCase.Input, req.FailedCase.Expected, req.FailedCase.Actual)
	}
	prompt += fmt.Sprintf("\n\n候选算法标签（algorithmTags 必须从中选取）：\n%s", CandidateTagDict)
	prompt += "\n\n输出JSON：{timeComplexity,spaceComplexity,algorithmTags,suggestions}。time/space用Markdown大O加粗。"
	systemMsg := "你是算法竞赛教练。分析代码质量和问题。用JSON格式回复。"
	if req.JudgeStatus == "Accepted" {
		systemMsg = "你是算法竞赛教练。代码已通过全部测试，分析算法思路、复杂度和优化空间。用JSON格式回复。"
	}
	t0 := time.Now()
	resp, err := h.ai.Chat([]ai.Message{{Role: "system", Content: systemMsg}, {Role: "user", Content: prompt}})
	log.Printf("[ai] code-diagnosis LLM call took %v", time.Since(t0))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "data": gin.H{"summary": "AI 服务暂时不可用", "rawMarkdown": "AI 服务暂时不可用", "provider": "unavailable"}})
		return
	}
	var structured map[string]interface{}
	if json.Unmarshal([]byte(resp), &structured) == nil {
		structured["rawMarkdown"] = resp
		structured["provider"] = "agent-service"
		c.JSON(http.StatusOK, gin.H{"code": 0, "data": structured})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": gin.H{"summary": "代码诊断完成", "rawMarkdown": resp, "provider": "agent-service"}})
}

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

type KnowledgeGraphPayload struct {
	Scope    string                  `json:"scope"`
	Problems []ProblemData           `json:"problems,omitempty"`
	TagStats map[string]TagStatsData `json:"tagStats,omitempty"`
}
type ProblemData struct {
	ID       uint64   `json:"id"`
	Title    string   `json:"title"`
	Tags     []string `json:"tags"`
	Status   string   `json:"status"`
	Attempts int      `json:"attempts"`
}
type TagStatsData struct {
	Solved    int     `json:"solved"`
	Attempted int     `json:"attempted"`
	ACRate    float64 `json:"acRate"`
}

func (h *Handler) KnowledgeGraph(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": "failed to read body"})
		return
	}
	var envelope PipelineEnvelope
	var req KnowledgeGraphPayload
	if err := json.Unmarshal(body, &envelope); err == nil && envelope.Task != "" {
		json.Unmarshal(envelope.Payload, &req)
	} else {
		json.Unmarshal(body, &req)
	}
	prompt := "根据用户的做题记录分析知识掌握情况，生成知识图谱。\n\n"
	if len(req.Problems) > 0 {
		for _, p := range req.Problems {
			prompt += fmt.Sprintf("- %s (标签: %s, 状态: %s, 尝试: %d)\n", p.Title, strings.Join(p.Tags, ", "), p.Status, p.Attempts)
		}
	}
	if len(req.TagStats) > 0 {
		for tag, stats := range req.TagStats {
			prompt += fmt.Sprintf("- %s: 解决 %d/%d 题, 通过率 %.1f%%\n", tag, stats.Solved, stats.Attempted, stats.ACRate)
		}
	}
	prompt += fmt.Sprintf("\n\n候选标签：\n%s\n\n返回JSON：{nodes:[{id,label,mastery,category}],edges:[{source,target,type}],suggestions:[]}", CandidateTagDict)
	resp, err := h.ai.Chat([]ai.Message{{Role: "system", Content: "你是算法竞赛教练。分析知识掌握情况生成知识图谱。用JSON格式回复。"}, {Role: "user", Content: prompt}})
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "data": gin.H{"nodes": []gin.H{}, "edges": []gin.H{}, "rawMarkdown": "AI 暂不可用", "provider": "unavailable"}})
		return
	}
	var structured map[string]interface{}
	if json.Unmarshal([]byte(resp), &structured) == nil {
		structured["rawMarkdown"] = resp
		structured["provider"] = "agent-service"
		c.JSON(http.StatusOK, gin.H{"code": 0, "data": structured})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": gin.H{"nodes": []gin.H{}, "edges": []gin.H{}, "rawMarkdown": resp, "provider": "agent-service"}})
}

type CreateStudyPlanPayload struct {
	UserID     uint64                   `json:"userId"`
	Problems   []ProblemData            `json:"problems"`
	TagStats   map[string]TagStatsData  `json:"tagStats"`
	Candidates map[string][]ProblemData `json:"candidates"`
}

func (h *Handler) CreateStudyPlan(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	var envelope PipelineEnvelope
	var req CreateStudyPlanPayload
	if json.Unmarshal(body, &envelope); err == nil && envelope.Task != "" {
		json.Unmarshal(envelope.Payload, &req)
	} else {
		json.Unmarshal(body, &req)
	}
	prompt := "根据用户的做题记录和薄弱知识点创建个性化学习题单。\n\n"
	if len(req.Problems) > 0 {
		for _, p := range req.Problems {
			prompt += fmt.Sprintf("- %s (标签: %s, 状态: %s)\n", p.Title, strings.Join(p.Tags, ", "), p.Status)
		}
	}
	if len(req.TagStats) > 0 {
		for tag, stats := range req.TagStats {
			prompt += fmt.Sprintf("- %s: 通过率 %.1f%%, 解决 %d/%d 题\n", tag, stats.ACRate, stats.Solved, stats.Attempted)
		}
	}
	if len(req.Candidates) > 0 {
		for tag, probs := range req.Candidates {
			prompt += fmt.Sprintf("\n【%s】:\n", tag)
			for _, p := range probs {
				prompt += fmt.Sprintf("  - #%d %s\n", p.ID, p.Title)
			}
		}
	}
	prompt += "\n\n返回JSON：{title,description,problemIDs}"
	resp, err := h.ai.Chat([]ai.Message{{Role: "system", Content: "你是算法教练。根据学生学习情况制定个性化题单。用JSON格式回复。"}, {Role: "user", Content: prompt}})
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "data": gin.H{"title": "默认题单", "description": "AI不可用", "problemIDs": []uint64{}, "rawMarkdown": "", "provider": "unavailable"}})
		return
	}
	var structured map[string]interface{}
	if json.Unmarshal([]byte(resp), &structured) == nil {
		structured["rawMarkdown"] = resp
		structured["provider"] = "agent-service"
		c.JSON(http.StatusOK, gin.H{"code": 0, "data": structured})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": gin.H{"title": "AI推荐题单", "description": resp, "problemIDs": []uint64{}, "rawMarkdown": resp, "provider": "agent-service"}})
}

func (h *Handler) Solve(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": "failed to read body"})
		return
	}
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
	var req SolvePayload
	if err := json.Unmarshal(body, &req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": "invalid request"})
		return
	}
	h.handleSolve(c, &req)
}

func (h *Handler) handleSolve(c *gin.Context, req *SolvePayload) {
	mergeSolveProblemFields(req)
	if req.EditorCode == "" && req.Code != "" {
		req.EditorCode = req.Code
	}
	level := req.Level
	if level == "" {
		level = "hint"
	}
	contextInfo := fmt.Sprintf("题目: %s\n题目描述: %s", req.ProblemTitle, req.ProblemDesc)
	if len(req.Samples) > 0 {
		for i, s := range req.Samples {
			contextInfo += fmt.Sprintf("\n样例%d:\n  输入: %s\n  期望输出: %s\n", i+1, s.Input, s.Expected)
		}
	}
	if req.Editorial != "" {
		contextInfo += fmt.Sprintf("\n\n官方题解:\n%s", req.Editorial)
	}
	if req.EditorCode != "" {
		contextInfo += fmt.Sprintf("\n\n用户当前代码:\n%s", req.EditorCode)
	}
	if req.JudgeError != "" {
		contextInfo += fmt.Sprintf("\n\n之前的判题结果: %s", req.JudgeError)
	}
	var systemPrompt, userPrompt string
	switch level {
	case "hint":
		systemPrompt = "你是启发式算法教练。用旁敲侧击引导思考，不要给答案。用JSON格式回复。"
		if len(req.AlgorithmTags) > 0 {
			contextInfo += fmt.Sprintf("\n\n算法标签: %s", strings.Join(req.AlgorithmTags, ", "))
		}
		userPrompt = contextInfo + "\n\n请给出启发式提示。返回JSON：{answer}"
	case "explain":
		systemPrompt = "你是算法教练。分析用户代码指出最大问题，不给解决方案。用JSON格式回复。"
		userPrompt = contextInfo + "\n\n分析当前代码最大问题。返回JSON：{answer}"
	case "full":
		systemPrompt = "你是算法教练。生成完整可运行代码。用JSON格式回复。"
		if len(req.AlgorithmTags) > 0 {
			contextInfo += fmt.Sprintf("\n\n算法标签: %s", strings.Join(req.AlgorithmTags, ", "))
		}
		userPrompt = contextInfo + "\n\n生成完整代码。返回JSON：{answer,code,language,timeComplexity,spaceComplexity}"
		if req.JudgeError != "" {
			userPrompt += fmt.Sprintf("\n之前的判题结果: %s\n请据此修正。", req.JudgeError)
		}
	default:
		systemPrompt = "你是算法教练。用JSON格式回复。"
		userPrompt = contextInfo
	}
	resp, err := h.ai.Chat([]ai.Message{{Role: "system", Content: systemPrompt}, {Role: "user", Content: userPrompt}})
	log.Printf("[ai] solve LLM call took %v (level=%s)", time.Since(time.Now()), level)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "data": gin.H{"answer": "AI 暂不可用", "provider": "unavailable"}})
		return
	}
	var structured map[string]interface{}
	if json.Unmarshal([]byte(resp), &structured) == nil {
		structured["rawMarkdown"] = resp
		structured["provider"] = "agent-service"
		c.JSON(http.StatusOK, gin.H{"code": 0, "data": structured})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": gin.H{"answer": resp, "rawMarkdown": resp, "provider": "agent-service"}})
}

// =========================================================================
// NEW: Unified Chat with Two-Phase Agent Loop + Tool Calling
// =========================================================================

type UnifiedChatRequest struct {
	Mode           string           `json:"mode"`
	UserID         uint64           `json:"user_id"`
	ConversationID string           `json:"conversation_id,omitempty"`
	Messages       []ai.Message     `json:"messages"`
	Problem        *ProblemPayload  `json:"problem,omitempty"`
	ExtraProblems  []ProblemPayload `json:"extraProblems,omitempty"`
	Code           string           `json:"code,omitempty"`
	Language       string           `json:"language,omitempty"`
}

func (h *Handler) UnifiedChat(c *gin.Context) {
	var req UnifiedChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": "invalid request"})
		return
	}
	mode := req.Mode
	if mode == "" {
		mode = "chat"
	}
	if !tools.IsValidMode(mode) {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": "unknown mode: " + mode})
		return
	}
	maxRounds := tools.MaxRoundsForMode(mode)
	toolDefs := tools.ForMode(mode)
	messages := h.buildModeMessages(req, mode)
	round := 0
	if len(toolDefs) > 0 {
		for round < maxRounds {
			aiTools := make([]ai.ToolDefinition, len(toolDefs))
			for i, td := range toolDefs {
				aiTools[i] = ai.ToolDefinition{Name: td.Name, Description: td.Description, Parameters: td.Schema}
			}
			result, err := h.ai.ChatWithTools(messages, aiTools, "auto")
			if err != nil {
				c.JSON(http.StatusOK, gin.H{"code": -1, "data": gin.H{"reply": "AI unavailable", "provider": "unavailable"}})
				return
			}
			if len(result.ToolCalls) == 0 {
				break
			}
			for _, tc := range result.ToolCalls {
				tcJSON, _ := json.Marshal([]map[string]interface{}{{"id": tc.ID, "type": "function", "function": map[string]string{"name": tc.Name, "arguments": mustMarshal(tc.Arguments)}}})
				messages = append(messages, ai.Message{Role: "assistant", ToolCallsJSON: tcJSON})
				toolResult := h.toolExec.Execute(tc.ID, tc.Name, tc.Arguments, req.UserID)
				messages = append(messages, ai.Message{Role: "tool", Content: toolResult.Content, ToolCallID: tc.ID})
			}
			round++
		}
	}
	// Phase 2 transition: tell LLM to deliver final answer to user
	cleanMsgs := []ai.Message{{Role: "system", Content: "以上是内部工具调用和数据收集过程。现在你需要基于收集到的数据，直接向用户给出最终答案。不要点评之前的数据获取过程是否相关、是否杂乱，不要说'刚才查询到'之类的元描述。就像你一开始就知道这些信息一样，自然地回答用户的问题。"}}
	for _, m := range messages {
		if m.Role == "tool" {
			cleanMsgs = append(cleanMsgs, ai.Message{Role: "user", Content: m.Content})
		} else {
			cleanMsgs = append(cleanMsgs, m)
		}
	}
	finalResp, err := h.ai.Chat(cleanMsgs)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": -1, "data": gin.H{"reply": "AI unavailable", "provider": "unavailable"}})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": gin.H{"reply": finalResp, "rawMarkdown": finalResp, "provider": h.ai.ProviderName(), "roundsUsed": round, "mode": mode}})
}

func (h *Handler) buildModeMessages(req UnifiedChatRequest, mode string) []ai.Message {
	var systemMsg string
	switch mode {
	case "chat":
		systemMsg = "你是竞赛AI助手，用中文回答时结合用户实际情况做个性化讲解。工具：search_problems搜题目、query_user_problems查做题记录、retrieve_knowledge查知识、get_user_code取代码。"
	case "code-diagnosis":
		systemMsg = "你是代码审查专家。分析用户代码和判题结果。返回JSON格式：timeComplexity用Markdown大O加粗，spaceComplexity同理，algorithmTags用已知标签名，suggestions是改进建议列表。必须返回纯JSON不要markdown代码块。"
	case "generate-solution":
		systemMsg = "你是题解撰写专家。根据用户通过的代码生成高质量题解。必须严格按以下JSON格式返回：{\"title\":\"题解标题\",\"content\":\"题解正文（Markdown，按一、问题分析 二、推理流程 三、算法设计与分析 四、代码实现 五、总结 的结构编写）\",\"algorithmTags\":[\"标签1\"],\"complexity\":{\"time\":\"**O(n)**\",\"space\":\"**O(1)**\"}}。content必须用Markdown格式分五个部分编写。time/space用纯大O加粗不要附加说明文字。algorithmTags必须从候选标签中选取。必须返回纯JSON不要markdown代码块。"
	case "knowledge-graph":
		systemMsg = "你是竞赛AI教练。先用 query_user_problems 查用户所有做题记录，分析薄弱点后返回JSON。JSON格式：nodes数组每项含id/label/mastery/category，edges数组含source/target/type，suggestions数组。节点id和label用标签名。mastery根据AC率和做题量判定为mastered/proficient/familiar/learning/unattempted。必须返回纯JSON不要其他文本。"
	case "study-plan":
		systemMsg = "你是竞赛AI教练。先用 query_user_problems 查已做题找薄弱标签，再用 query_user_problems 查未做题，制定题单后返回JSON。JSON格式：title,description,problemIDs数组。必须返回纯JSON不要其他文本。"
	case "solve":
		systemMsg = "你是算法教练。你拥有 query_user_problems、submit_code、retrieve_knowledge 工具可自主调用。学生需要解题帮助时自主决定何时用哪个工具获取信息或验证代码。"
	default:
		systemMsg = "你是竞赛AI助手，用中文回答。"
	}
	if req.Problem != nil {
		systemMsg += fmt.Sprintf("\n\n当前题目: %s\n题面:\n%s", req.Problem.Title, req.Problem.Content)
		if len(req.Problem.Tags) > 0 {
			systemMsg += fmt.Sprintf("\n算法标签: %s", strings.Join(req.Problem.Tags, ", "))
		}
		if req.Problem.Editorial != "" {
			systemMsg += fmt.Sprintf("\n官方题解: %s", req.Problem.Editorial)
		}
	}
	if len(req.ExtraProblems) > 0 {
		systemMsg += "\n\n关联题目："
		for _, ep := range req.ExtraProblems {
			systemMsg += fmt.Sprintf("\n- #%d %s (标签: %s)", ep.ID, ep.Title, strings.Join(ep.Tags, ", "))
		}
	}
	if req.Code != "" {
		lang := req.Language
		if lang == "" {
			lang = "代码"
		}
		systemMsg += fmt.Sprintf("\n\n用户代码（%s）:\n%s", lang, req.Code)
	}
	msgs := []ai.Message{{Role: "system", Content: systemMsg}}
	msgs = append(msgs, req.Messages...)
	return msgs
}

func mustMarshal(v interface{}) string { b, _ := json.Marshal(v); return string(b) }

// =========================================================================
// Chat handler (legacy + unified dispatch)
// =========================================================================

func (h *Handler) Chat(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read body"})
		return
	}

	var req ChatRequest

	// 1. Try envelope format (AIOJ backend)
	var envelope PipelineEnvelope
	if err := json.Unmarshal(body, &envelope); err == nil && envelope.Task != "" {
		var payload ChatPayload
		if err := json.Unmarshal(envelope.Payload, &payload); err == nil {
			req = h.chatPayloadToRequest(&payload)
		}
		// If legacy parse failed, try UnifiedChatRequest from payload
		if len(req.Messages) == 0 {
			var unified UnifiedChatRequest
			if err := json.Unmarshal(envelope.Payload, &unified); err == nil && unified.Mode != "" {
				c.Request.Body = io.NopCloser(bytes.NewReader(envelope.Payload))
				h.UnifiedChat(c)
				return
			}
		}
	}

	// 2. Try direct UnifiedChat format
	if len(req.Messages) == 0 {
		var unified UnifiedChatRequest
		if err := json.Unmarshal(body, &unified); err == nil && unified.Mode != "" {
			c.Request.Body = io.NopCloser(bytes.NewReader(body))
			h.UnifiedChat(c)
			return
		}
	}

	// 3. Legacy direct ChatRequest
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

	systemContent := "你是一个算法竞赛AI助手，擅长算法、数据结构、编程竞赛相关问题。请用中文回答，回答简洁准确。"
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
	req.Messages = append([]ai.Message{{Role: "system", Content: systemContent}}, req.Messages...)
	resp, err := h.ai.Chat(req.Messages)
	log.Printf("[ai] chat LLM call took %v (messages=%d)", time.Since(time.Now()), len(req.Messages))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "AI 服务暂时不可用，请稍后重试"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"reply": resp, "response": resp})
}

func (h *Handler) chatPayloadToRequest(payload *ChatPayload) ChatRequest {
	var messages []ai.Message
	for _, m := range payload.History {
		if m.Content != "" {
			messages = append(messages, ai.Message{Role: m.Role, Content: m.Content})
		}
	}
	if payload.Message != "" {
		messages = append(messages, ai.Message{Role: "user", Content: payload.Message})
	}
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
	for i, ep := range payload.ExtraProblems {
		contextParts = append(contextParts, fmt.Sprintf("关联题目%d: #%d %s (标签: %s)", i+1, ep.ID, ep.Title, strings.Join(ep.Tags, ", ")))
	}
	return ChatRequest{Messages: messages, Context: strings.Join(contextParts, "\n\n"), Tags: tags}
}
