package tools

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"agent-service/internal/rag"
)

type ToolResult struct {
	ToolCallID string `json:"-"`
	Name       string `json:"-"`
	Content    string `json:"content"`
	IsError    bool   `json:"is_error"`
}

type Executor struct {
	ojBaseURL  string
	httpClient *http.Client
	rag        *rag.Service
}

func NewExecutor(ojBaseURL string, ragService *rag.Service) *Executor {
	return &Executor{
		ojBaseURL: strings.TrimRight(ojBaseURL, "/"),
		httpClient: &http.Client{Timeout: 60 * time.Second},
		rag:       ragService,
	}
}

func (e *Executor) Execute(toolCallID, name string, args map[string]interface{}, userID uint64) *ToolResult {
	switch name {
	case "query_user_problems":
		return e.queryUserProblems(toolCallID, args, userID)
	case "submit_code":
		return e.submitCode(toolCallID, args, userID)
	case "retrieve_knowledge":
		return e.retrieveKnowledge(toolCallID, args)
	case "get_user_code":
		return e.getUserCode(toolCallID, args, userID)
	case "search_problems":
		return e.searchProblems(toolCallID, args)
	default:
		return &ToolResult{ToolCallID: toolCallID, Name: name, Content: fmt.Sprintf(`{"error":"unknown tool: %s"}`, name), IsError: true}
	}
}

type queryProblemsArgs struct {
	Tags       []string `json:"tags,omitempty"`
	Status     string   `json:"status,omitempty"`
	Difficulty string   `json:"difficulty,omitempty"`
}

func (e *Executor) queryUserProblems(toolCallID string, args map[string]interface{}, userID uint64) *ToolResult {
	req := queryProblemsArgs{}
	if tags, ok := args["tags"]; ok {
		if tagList, ok := tags.([]interface{}); ok {
			for _, t := range tagList {
				if s, ok := t.(string); ok { req.Tags = append(req.Tags, s) }
			}
		}
	}
	if s, ok := args["status"].(string); ok { req.Status = s }
	if d, ok := args["difficulty"].(string); ok { req.Difficulty = d }
	body, _ := json.Marshal(req)
	httpReq, _ := http.NewRequest("POST", e.ojBaseURL+"/api/agent/problems", bytes.NewReader(body))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-User-ID", fmt.Sprintf("%d", userID))
	resp, err := e.httpClient.Do(httpReq)
	if err != nil { return e.errorResult(toolCallID, "query_user_problems", fmt.Sprintf("HTTP error: %v", err)) }
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	return &ToolResult{ToolCallID: toolCallID, Name: "query_user_problems", Content: string(respBody)}
}

type submitCodeArgs struct {
	ProblemID uint64 `json:"problem_id"`
	Code      string `json:"code"`
	Language  string `json:"language"`
}

func (e *Executor) submitCode(toolCallID string, args map[string]interface{}, userID uint64) *ToolResult {
	req := submitCodeArgs{}
	if id, ok := args["problem_id"].(float64); ok { req.ProblemID = uint64(id) }
	if code, ok := args["code"].(string); ok { req.Code = code }
	if lang, ok := args["language"].(string); ok { req.Language = lang }
	if req.ProblemID == 0 || req.Code == "" || req.Language == "" {
		return e.errorResult(toolCallID, "submit_code", "missing required fields")
	}
	body, _ := json.Marshal(req)
	httpReq, _ := http.NewRequest("POST", e.ojBaseURL+"/api/agent/judge", bytes.NewReader(body))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-User-ID", fmt.Sprintf("%d", userID))
	resp, err := e.httpClient.Do(httpReq)
	if err != nil { return e.errorResult(toolCallID, "submit_code", fmt.Sprintf("HTTP error: %v", err)) }
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	return &ToolResult{ToolCallID: toolCallID, Name: "submit_code", Content: string(respBody)}
}

func (e *Executor) retrieveKnowledge(toolCallID string, args map[string]interface{}) *ToolResult {
	if e.rag == nil || !e.rag.IsInitialized() { return e.errorResult(toolCallID, "retrieve_knowledge", "知识库尚未加载") }
	var queries []string
	var boostTags []string
	if tags, ok := args["tags"]; ok {
		if tagList, ok := tags.([]interface{}); ok {
			for _, t := range tagList {
				if s, ok := t.(string); ok {
					queries = append(queries, s)
					boostTags = append(boostTags, s)
				}
			}
		}
	}
	if q, ok := args["query"].(string); ok && q != "" { queries = append(queries, q) }
	if len(queries) == 0 { return e.errorResult(toolCallID, "retrieve_knowledge", "需要至少一个tags参数") }
	context := e.rag.BuildContextTagged(queries, boostTags, 0, 5)
	if context == "" { return &ToolResult{ToolCallID: toolCallID, Name: "retrieve_knowledge", Content: `{"result":"未找到相关知识"}`} }
	resultJSON, _ := json.Marshal(map[string]string{"result": context})
	return &ToolResult{ToolCallID: toolCallID, Name: "retrieve_knowledge", Content: string(resultJSON)}
}

func (e *Executor) getUserCode(toolCallID string, args map[string]interface{}, userID uint64) *ToolResult {
	pid := uint64(0)
	if id, ok := args["problem_id"].(float64); ok { pid = uint64(id) }
	if pid == 0 { return e.errorResult(toolCallID, "get_user_code", "missing problem_id") }
	httpReq, _ := http.NewRequest("POST", e.ojBaseURL+"/api/agent/code", bytes.NewReader([]byte(fmt.Sprintf(`{"problem_id":%d}`, pid))))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-User-ID", fmt.Sprintf("%d", userID))
	resp, err := e.httpClient.Do(httpReq)
	if err != nil { return e.errorResult(toolCallID, "get_user_code", fmt.Sprintf("HTTP error: %v", err)) }
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	return &ToolResult{ToolCallID: toolCallID, Name: "get_user_code", Content: string(respBody)}
}

func (e *Executor) searchProblems(toolCallID string, args map[string]interface{}) *ToolResult {
	query, _ := args["query"].(string)
	if query == "" { return e.errorResult(toolCallID, "search_problems", "missing query") }
	body, _ := json.Marshal(map[string]string{"query": query})
	httpReq, _ := http.NewRequest("POST", e.ojBaseURL+"/api/agent/search-problems", bytes.NewReader(body))
	httpReq.Header.Set("Content-Type", "application/json")
	resp, err := e.httpClient.Do(httpReq)
	if err != nil { return e.errorResult(toolCallID, "search_problems", fmt.Sprintf("HTTP error: %v", err)) }
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	return &ToolResult{ToolCallID: toolCallID, Name: "search_problems", Content: string(respBody)}
}

func (e *Executor) errorResult(toolCallID, name, message string) *ToolResult {
	errJSON, _ := json.Marshal(map[string]string{"error": message})
	return &ToolResult{ToolCallID: toolCallID, Name: name, Content: string(errJSON), IsError: true}
}
