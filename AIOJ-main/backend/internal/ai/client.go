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
)

type Client struct {
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
	Editorial       string   `json:"editorial,omitempty"`
	Samples         []Sample `json:"samples,omitempty"`
	TimeLimit       int      `json:"timeLimit"`
	MemoryLimit     int      `json:"memoryLimit"`
}

type Sample struct {
	Input    string `json:"input"`
	Expected string `json:"expected"`
}

type SubmissionDigest struct {
	ID           uint64 `json:"id"`
	ProblemID    uint64 `json:"problemId"`
	ProblemTitle string `json:"problemTitle"`
	Language     string `json:"language"`
	Status       string `json:"status"`
	Code         string `json:"code,omitempty"`
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
	CodeLanguage   string          `json:"codeLanguage,omitempty"`
	Code           string          `json:"code,omitempty"`
}

type ChatResponse struct {
	Reply    string         `json:"reply"`
	Provider string         `json:"provider,omitempty"`
	Metadata map[string]any `json:"metadata,omitempty"`
}

type FailedCase struct {
	Input    string `json:"input"`
	Expected string `json:"expected"`
	Actual   string `json:"actual"`
}

type CodeDiagnosisRequest struct {
	UserID       uint64              `json:"userId"`
	Problem      *ProblemContext     `json:"problem,omitempty"`
	SubmissionID uint64              `json:"submissionId,omitempty"`
	Language     string              `json:"language"`
	Code         string              `json:"code"`
	JudgeStatus  string              `json:"judgeStatus,omitempty"`
	ErrorMessage string              `json:"errorMessage,omitempty"`
	RuntimeMs    int                 `json:"runtimeMs,omitempty"`
	MemoryKb     int                 `json:"memoryKb,omitempty"`
	RecentSubs   []SubmissionDigest  `json:"recentSubmissions,omitempty"`
	FailedCase   *FailedCase         `json:"failedCase,omitempty"`
}

type CodeIssue struct {
	Line     int    `json:"line,omitempty"`
	Severity string `json:"severity"`
	Message  string `json:"message"`
	Hint     string `json:"hint,omitempty"`
}

type CodeDiagnosisResponse struct {
	Summary         string      `json:"summary"`
	Issues          []CodeIssue `json:"issues"`
	Suggestions     []string    `json:"suggestions"`
	FixedCode       string      `json:"fixedCode,omitempty"`
	TimeComplexity  string      `json:"timeComplexity,omitempty"`
	SpaceComplexity string      `json:"spaceComplexity,omitempty"`
	AlgorithmTags   []string    `json:"algorithmTags,omitempty"`
	RawMarkdown     string      `json:"rawMarkdown"`
	Provider        string      `json:"provider,omitempty"`
}

type KnowledgeGraphRequest struct {
	UserID            uint64             `json:"userId"`
	Scope             string             `json:"scope"`
	Problem           *ProblemContext    `json:"problem,omitempty"`
	RecentSubmissions []SubmissionDigest `json:"recentSubmissions,omitempty"`
	Problems          []ProblemSummary   `json:"problems,omitempty"`
	TagStats          map[string]TagStat `json:"tagStats,omitempty"`
}

type ProblemSummary struct {
	ID         uint64   `json:"id"`
	Title      string   `json:"title"`
	Tags       []string `json:"tags"`
	Status     string   `json:"status"`
	Attempts   int      `json:"attempts"`
}

type TagStat struct {
	Solved    int     `json:"solved"`
	Attempted int     `json:"attempted"`
	ACRate    float64 `json:"acRate"`
}

type GraphNode struct {
	ID       string `json:"id"`
	Label    string `json:"label"`
	Mastery  string `json:"mastery"`
	Category string `json:"category,omitempty"`
}

type GraphEdge struct {
	Source string `json:"source"`
	Target string `json:"target"`
	Type   string `json:"type"`
	Weight int    `json:"weight,omitempty"`
}

type KnowledgeGraphResponse struct {
	Nodes       []GraphNode `json:"nodes"`
	Edges       []GraphEdge `json:"edges"`
	Suggestions []string    `json:"suggestions,omitempty"`
	RawMarkdown string      `json:"rawMarkdown"`
	Provider    string      `json:"provider,omitempty"`
}

type SolveRequest struct {
	UserID     uint64          `json:"userId"`
	Problem    *ProblemContext `json:"problem"`
	Question   string          `json:"question,omitempty"`
	Level      string          `json:"level"`
	Language   string          `json:"language,omitempty"`
	EditorCode string          `json:"editorCode,omitempty"`
	JudgeError string          `json:"judgeError,omitempty"`
}

type SolveResponse struct {
	Answer         string   `json:"answer"`
	Code           string   `json:"code,omitempty"`
	Language       string   `json:"language,omitempty"`
	TimeComplexity string   `json:"timeComplexity,omitempty"`
	SpaceComplexity string  `json:"spaceComplexity,omitempty"`
	AlgorithmTags  []string `json:"algorithmTags,omitempty"`
	VerifyResult string  `json:"verifyResult,omitempty"`
	Provider    string   `json:"provider,omitempty"`
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
		endpoint: endpoint,
		apiKey:   strings.TrimSpace(cfg.APIKey),
		model:    model,
		http:     &http.Client{Timeout: time.Duration(timeout) * time.Second},
	}
}

func (c *Client) Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
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

type CreateStudyPlanRequest struct {
	UserID       uint64             `json:"userId"`
	Problems     []ProblemSummary   `json:"problems"`
	TagStats     map[string]TagStat `json:"tagStats"`
	// Candidate problems grouped by tag (unified from AIOJ backend)
	Candidates   map[string][]ProblemSummary `json:"candidates"`
}

type CreateStudyPlanResponse struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	ProblemIDs  []uint64 `json:"problemIDs"`
	RawMarkdown string   `json:"rawMarkdown"`
	Provider    string   `json:"provider,omitempty"`
}

func (c *Client) CreateStudyPlan(ctx context.Context, req CreateStudyPlanRequest) (*CreateStudyPlanResponse, error) {
	var resp CreateStudyPlanResponse
	if err := c.post(ctx, "create_study_plan", "/create-study-plan", req, &resp); err != nil {
		return nil, err
	}
	withProvider(&resp.Provider)
	return &resp, nil
}

func (c *Client) Solve(ctx context.Context, req SolveRequest) (*SolveResponse, error) {
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

type GenerateSolutionRequest struct {
	UserID  uint64          `json:"userId"`
	Problem *ProblemContext  `json:"problem,omitempty"`
	Language string          `json:"language,omitempty"`
	Code    string          `json:"code,omitempty"`
}

type GenerateSolutionResponse struct {
	Title      string            `json:"title"`
	Content    string            `json:"content"`
	Tags       []string          `json:"algorithmTags,omitempty"`
	Complexity map[string]string `json:"complexity,omitempty"`
	RawMarkdown string           `json:"rawMarkdown"`
	Provider   string            `json:"provider,omitempty"`
}

func (c *Client) GenerateSolution(ctx context.Context, req GenerateSolutionRequest) (*GenerateSolutionResponse, error) {
	var resp GenerateSolutionResponse
	if err := c.post(ctx, "generate_solution", "/generate-solution", req, &resp); err != nil {
		return nil, err
	}
	if strings.TrimSpace(resp.RawMarkdown) == "" {
		resp.RawMarkdown = resp.Content
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

func graphMarkdown(resp KnowledgeGraphResponse) string {
	var b strings.Builder
	b.WriteString("### 学习知识图谱\n\n")
	if len(resp.Suggestions) > 0 {
		b.WriteString(strings.Join(resp.Suggestions, "；") + "\n\n")
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
			b.WriteString(fmt.Sprintf("- `%s` %s（%s）\n", n.ID, n.Label, n.Mastery))
		}
	}
	return strings.TrimSpace(b.String())
}
