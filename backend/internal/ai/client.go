package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/terminaloj/backend/internal/config"
)

const (
	ProviderExternal = "external"
	ProviderMock     = "mock"
)

type Client struct {
	enabled  bool
	endpoint string
	apiKey   string
	model    string
	http     *http.Client
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ProblemContext struct {
	ID              uint64   `json:"id"`
	Title           string   `json:"title"`
	Difficulty      string   `json:"difficulty"`
	DifficultyScore int      `json:"difficultyScore"`
	Tags            []string `json:"tags"`
	Content         string   `json:"content,omitempty"`
	TimeLimit       int      `json:"timeLimit"`
	MemoryLimit     int      `json:"memoryLimit"`
}

type SubmissionDigest struct {
	ID           uint64 `json:"id"`
	ProblemID    uint64 `json:"problemId"`
	ProblemTitle string `json:"problemTitle"`
	Language     string `json:"language"`
	Status       string `json:"status"`
	Runtime      int    `json:"runtime"`
	Memory       string `json:"memory"`
	CodeLength   int    `json:"codeLength"`
	ErrorMessage string `json:"errorMessage,omitempty"`
	CreatedAt    string `json:"createdAt"`
}

type ChatRequest struct {
	UserID         uint64          `json:"userId"`
	ConversationID string          `json:"conversationId"`
	Message        string          `json:"message"`
	History        []Message       `json:"history"`
	Problem        *ProblemContext `json:"problem,omitempty"`
}

type ChatResponse struct {
	Reply    string         `json:"reply"`
	Provider string         `json:"provider,omitempty"`
	Metadata map[string]any `json:"metadata,omitempty"`
}

type CodeDiagnosisRequest struct {
	UserID       uint64          `json:"userId"`
	Problem      *ProblemContext `json:"problem,omitempty"`
	SubmissionID uint64          `json:"submissionId,omitempty"`
	Language     string          `json:"language"`
	Code         string          `json:"code"`
	JudgeStatus  string          `json:"judgeStatus,omitempty"`
	ErrorMessage string          `json:"errorMessage,omitempty"`
}

type CodeIssue struct {
	Line     int    `json:"line,omitempty"`
	Severity string `json:"severity"`
	Message  string `json:"message"`
	Hint     string `json:"hint,omitempty"`
}

type CodeDiagnosisResponse struct {
	Summary     string      `json:"summary"`
	Issues      []CodeIssue `json:"issues"`
	Suggestions []string    `json:"suggestions"`
	FixedCode   string      `json:"fixedCode,omitempty"`
	RawMarkdown string      `json:"rawMarkdown"`
	Provider    string      `json:"provider,omitempty"`
}

type KnowledgeGraphRequest struct {
	UserID            uint64             `json:"userId"`
	Scope             string             `json:"scope"`
	Problem           *ProblemContext    `json:"problem,omitempty"`
	RecentSubmissions []SubmissionDigest `json:"recentSubmissions"`
}

type GraphNode struct {
	ID       string         `json:"id"`
	Label    string         `json:"label"`
	Type     string         `json:"type"`
	Weight   int            `json:"weight,omitempty"`
	Metadata map[string]any `json:"metadata,omitempty"`
}

type GraphEdge struct {
	Source string `json:"source"`
	Target string `json:"target"`
	Type   string `json:"type"`
	Weight int    `json:"weight,omitempty"`
}

type KnowledgeGraphResponse struct {
	Summary     string      `json:"summary"`
	Nodes       []GraphNode `json:"nodes"`
	Edges       []GraphEdge `json:"edges"`
	RawMarkdown string      `json:"rawMarkdown"`
	Provider    string      `json:"provider,omitempty"`
}

type SolveRequest struct {
	UserID   uint64          `json:"userId"`
	Problem  *ProblemContext `json:"problem"`
	Question string          `json:"question,omitempty"`
	Level    string          `json:"level"`
}

type SolveResponse struct {
	Answer     string   `json:"answer"`
	Hints      []string `json:"hints"`
	Complexity string   `json:"complexity,omitempty"`
	Provider   string   `json:"provider,omitempty"`
}

type pipelineEnvelope struct {
	Task    string `json:"task"`
	Model   string `json:"model,omitempty"`
	Payload any    `json:"payload"`
}

func NewClient(cfg config.AIConfig) *Client {
	timeout := cfg.TimeoutSeconds
	if timeout <= 0 {
		timeout = 20
	}
	model := strings.TrimSpace(cfg.Model)
	if model == "" {
		model = "terminaloj-ai"
	}
	endpoint := strings.TrimRight(strings.TrimSpace(cfg.Endpoint), "/")
	return &Client{
		enabled:  cfg.Enabled && endpoint != "",
		endpoint: endpoint,
		apiKey:   strings.TrimSpace(cfg.APIKey),
		model:    model,
		http:     &http.Client{Timeout: time.Duration(timeout) * time.Second},
	}
}

func (c *Client) Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	if !c.enabled {
		resp := mockChat(req)
		return &resp, nil
	}
	var resp ChatResponse
	if err := c.post(ctx, "chat", "/chat", req, &resp); err != nil {
		return nil, err
	}
	if strings.TrimSpace(resp.Reply) == "" {
		return nil, errors.New("ai chat response missing reply")
	}
	withProvider(&resp.Provider)
	return &resp, nil
}

func (c *Client) DiagnoseCode(ctx context.Context, req CodeDiagnosisRequest) (*CodeDiagnosisResponse, error) {
	if !c.enabled {
		resp := mockCodeDiagnosis(req)
		return &resp, nil
	}
	var resp CodeDiagnosisResponse
	if err := c.post(ctx, "code_diagnosis", "/code-diagnosis", req, &resp); err != nil {
		return nil, err
	}
	if strings.TrimSpace(resp.RawMarkdown) == "" {
		resp.RawMarkdown = diagnosisMarkdown(resp)
	}
	withProvider(&resp.Provider)
	return &resp, nil
}

func (c *Client) BuildKnowledgeGraph(ctx context.Context, req KnowledgeGraphRequest) (*KnowledgeGraphResponse, error) {
	if !c.enabled {
		resp := mockKnowledgeGraph(req)
		return &resp, nil
	}
	var resp KnowledgeGraphResponse
	if err := c.post(ctx, "knowledge_graph", "/knowledge-graph", req, &resp); err != nil {
		return nil, err
	}
	if strings.TrimSpace(resp.RawMarkdown) == "" {
		resp.RawMarkdown = graphMarkdown(resp)
	}
	withProvider(&resp.Provider)
	return &resp, nil
}

func (c *Client) Solve(ctx context.Context, req SolveRequest) (*SolveResponse, error) {
	if !c.enabled {
		resp := mockSolve(req)
		return &resp, nil
	}
	var resp SolveResponse
	if err := c.post(ctx, "solve", "/solve", req, &resp); err != nil {
		return nil, err
	}
	if strings.TrimSpace(resp.Answer) == "" {
		return nil, errors.New("ai solve response missing answer")
	}
	withProvider(&resp.Provider)
	return &resp, nil
}

func (c *Client) post(ctx context.Context, task string, path string, payload any, out any) error {
	body, err := json.Marshal(pipelineEnvelope{Task: task, Model: c.model, Payload: payload})
	if err != nil {
		return fmt.Errorf("marshal ai request: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint+path, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("build ai request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-AI-Task", task)
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("call ai service: %w", err)
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if err != nil {
		return fmt.Errorf("read ai response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("ai service status %d: %s", resp.StatusCode, strings.TrimSpace(string(raw)))
	}
	if len(raw) == 0 {
		return errors.New("ai service returned empty response")
	}
	return decodeResponse(raw, out)
}

func decodeResponse(raw []byte, out any) error {
	var wrapped struct {
		Code    int             `json:"code"`
		Message string          `json:"message"`
		Data    json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(raw, &wrapped); err == nil && len(wrapped.Data) > 0 {
		if wrapped.Code != 0 {
			if wrapped.Message == "" {
				wrapped.Message = "ai service returned non-zero code"
			}
			return errors.New(wrapped.Message)
		}
		if err := json.Unmarshal(wrapped.Data, out); err != nil {
			return fmt.Errorf("decode ai data: %w", err)
		}
		return nil
	}
	if err := json.Unmarshal(raw, out); err != nil {
		return fmt.Errorf("decode ai response: %w", err)
	}
	return nil
}

func withProvider(provider *string) {
	if strings.TrimSpace(*provider) == "" {
		*provider = ProviderExternal
	}
}

func mockChat(req ChatRequest) ChatResponse {
	ctx := "独立 AI 训练模式"
	if req.Problem != nil {
		ctx = fmt.Sprintf("题目 #%d《%s》辅助模式", req.Problem.ID, req.Problem.Title)
	}
	lower := strings.ToLower(req.Message)
	switch {
	case strings.Contains(lower, "dp") || strings.Contains(req.Message, "动态规划"):
		return ChatResponse{Provider: ProviderMock, Reply: "### 动态规划思路\n\n动态规划的核心在于：\n\n1. **状态定义**：$dp[i]$ 表示子问题答案。\n2. **转移方程**：从更小规模状态推导当前状态。\n3. **边界条件**：先写出最小规模输入的答案。\n\n```cpp\nfor (int i = 1; i <= n; ++i) {\n    dp[i] = max(dp[i], dp[i - 1]);\n}\n```\n\n当前会话：**" + ctx + "**。"}
	case strings.Contains(req.Message, "二分"):
		return ChatResponse{Provider: ProviderMock, Reply: "二分查找需要满足 **单调性**。建议先写一个 `ok(x)` 判断函数，再确定查找第一个满足条件的位置或最后一个满足条件的位置。\n\n" + ctx}
	case strings.Contains(req.Message, "复杂度"):
		return ChatResponse{Provider: ProviderMock, Reply: "常见复杂度对照：$O(1) < O(\\log n) < O(n) < O(n \\log n) < O(n^2) < O(2^n)$。\n\n" + ctx}
	default:
		return ChatResponse{Provider: ProviderMock, Reply: fmt.Sprintf("你好，我收到了你的问题：\n\n> %s\n\n这是 Mock 回复。接入外部 AI 服务后，将返回基于 **%s** 的完整推理。", req.Message, ctx)}
	}
}

func mockCodeDiagnosis(req CodeDiagnosisRequest) CodeDiagnosisResponse {
	issues := []CodeIssue{}
	if strings.Contains(req.Code, "TODO") {
		issues = append(issues, CodeIssue{Severity: "warning", Message: "代码中仍包含 TODO 标记", Hint: "提交前补齐占位逻辑。"})
	}
	if strings.Count(req.Code, "{") != strings.Count(req.Code, "}") {
		issues = append(issues, CodeIssue{Severity: "error", Message: "花括号数量不匹配", Hint: "检查 if/for/function 代码块是否完整闭合。"})
	}
	if req.Language == "cpp" && strings.Contains(req.Code, "int main") && !strings.Contains(req.Code, "return 0") {
		issues = append(issues, CodeIssue{Severity: "info", Message: "main 函数没有显式 return 0", Hint: "虽然 C++ 允许省略，但补充返回值更清晰。"})
	}
	if len(issues) == 0 {
		issues = append(issues, CodeIssue{Severity: "info", Message: "Mock 检查未发现明显语法级问题", Hint: "建议结合样例和边界数据继续验证。"})
	}
	suggestions := []string{
		"先用题目样例做最小闭环测试。",
		"补充边界输入：空数组、单元素、重复元素、极值。",
		"根据题目约束重新核对时间复杂度。",
	}
	resp := CodeDiagnosisResponse{
		Summary:     "已完成代码诊断，以下结果来自 TerminalOJ AI Mock 管线。",
		Issues:      issues,
		Suggestions: suggestions,
		Provider:    ProviderMock,
	}
	resp.RawMarkdown = diagnosisMarkdown(resp)
	return resp
}

func diagnosisMarkdown(resp CodeDiagnosisResponse) string {
	var b strings.Builder
	if resp.Summary != "" {
		b.WriteString("### 代码诊断\n\n")
		b.WriteString(resp.Summary)
		b.WriteString("\n\n")
	}
	if len(resp.Issues) > 0 {
		b.WriteString("#### 发现的问题\n\n")
		for _, issue := range resp.Issues {
			line := ""
			if issue.Line > 0 {
				line = fmt.Sprintf("L%d ", issue.Line)
			}
			b.WriteString(fmt.Sprintf("- **%s%s**：%s", line, issue.Severity, issue.Message))
			if issue.Hint != "" {
				b.WriteString("。" + issue.Hint)
			}
			b.WriteString("\n")
		}
		b.WriteString("\n")
	}
	if len(resp.Suggestions) > 0 {
		b.WriteString("#### 建议\n\n")
		for _, s := range resp.Suggestions {
			b.WriteString("- " + s + "\n")
		}
	}
	if resp.FixedCode != "" {
		b.WriteString("\n#### 参考修正\n\n```text\n")
		b.WriteString(resp.FixedCode)
		b.WriteString("\n```\n")
	}
	return strings.TrimSpace(b.String())
}

func mockKnowledgeGraph(req KnowledgeGraphRequest) KnowledgeGraphResponse {
	nodes := []GraphNode{{ID: "user", Label: "当前用户", Type: "user", Weight: 1}}
	edges := []GraphEdge{}
	seen := map[string]bool{"user": true}
	addNode := func(node GraphNode) {
		if seen[node.ID] {
			return
		}
		seen[node.ID] = true
		nodes = append(nodes, node)
	}
	if req.Problem != nil {
		pid := fmt.Sprintf("problem:%d", req.Problem.ID)
		addNode(GraphNode{ID: pid, Label: req.Problem.Title, Type: "problem", Weight: 1})
		edges = append(edges, GraphEdge{Source: "user", Target: pid, Type: "practicing", Weight: 1})
		for _, tag := range req.Problem.Tags {
			tid := "tag:" + tag
			addNode(GraphNode{ID: tid, Label: tag, Type: "algorithm", Weight: 1})
			edges = append(edges, GraphEdge{Source: pid, Target: tid, Type: "requires", Weight: 1})
		}
	}
	statusWeight := map[string]int{}
	for _, sub := range req.RecentSubmissions {
		pid := fmt.Sprintf("problem:%d", sub.ProblemID)
		addNode(GraphNode{ID: pid, Label: sub.ProblemTitle, Type: "problem", Weight: 1})
		edges = append(edges, GraphEdge{Source: "user", Target: pid, Type: sub.Status, Weight: 1})
		statusWeight[sub.Status]++
	}
	for status, weight := range statusWeight {
		sid := "status:" + status
		addNode(GraphNode{ID: sid, Label: status, Type: "status", Weight: weight})
		edges = append(edges, GraphEdge{Source: "user", Target: sid, Type: "has_result", Weight: weight})
	}
	resp := KnowledgeGraphResponse{
		Summary:  fmt.Sprintf("已根据 %d 条最近提交生成学习知识图谱。", len(req.RecentSubmissions)),
		Nodes:    nodes,
		Edges:    edges,
		Provider: ProviderMock,
	}
	resp.RawMarkdown = graphMarkdown(resp)
	return resp
}

func graphMarkdown(resp KnowledgeGraphResponse) string {
	var b strings.Builder
	b.WriteString("### 学习知识图谱\n\n")
	if resp.Summary != "" {
		b.WriteString(resp.Summary + "\n\n")
	}
	b.WriteString(fmt.Sprintf("- 节点数：%d\n", len(resp.Nodes)))
	b.WriteString(fmt.Sprintf("- 关系数：%d\n", len(resp.Edges)))
	if len(resp.Nodes) > 0 {
		b.WriteString("\n#### 关键节点\n\n")
		limit := len(resp.Nodes)
		if limit > 8 {
			limit = 8
		}
		for _, n := range resp.Nodes[:limit] {
			b.WriteString(fmt.Sprintf("- `%s` %s（%s）\n", n.ID, n.Label, n.Type))
		}
	}
	return strings.TrimSpace(b.String())
}

func mockSolve(req SolveRequest) SolveResponse {
	problemTitle := "当前题目"
	if req.Problem != nil {
		problemTitle = fmt.Sprintf("#%d《%s》", req.Problem.ID, req.Problem.Title)
	}
	level := req.Level
	if level == "" {
		level = "hint"
	}
	answer := fmt.Sprintf("### %s 解题辅助\n\n当前返回级别：`%s`。建议按以下步骤推进：\n\n1. 先明确输入规模和目标复杂度。\n2. 根据标签或样例推断核心算法。\n3. 写出状态 / 数据结构定义，再写转移或遍历逻辑。\n4. 用样例和边界数据验证。", problemTitle, level)
	if req.Question != "" {
		answer += "\n\n你的问题：\n\n> " + req.Question
	}
	return SolveResponse{
		Answer: answer,
		Hints: []string{
			"从暴力解开始，确认正确性后再优化。",
			"把样例手算一遍，观察重复子问题或单调性。",
			"提交前检查边界条件和数据范围。",
		},
		Complexity: "Mock 模式无法精确判断，建议目标复杂度控制在题目约束可接受范围内。",
		Provider:   ProviderMock,
	}
}
