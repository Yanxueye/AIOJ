package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

// --- Tool calling types ---

// ToolDefinition describes a tool available to the LLM.
type ToolDefinition struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Parameters  json.RawMessage `json:"parameters"` // JSON Schema
}

type toolDefinition struct {
	Type     string         `json:"type"`
	Function toolFunction   `json:"function"`
}

type toolFunction struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Parameters  json.RawMessage `json:"parameters"`
}

// ParsedToolCall represents a tool call requested by the LLM, with Arguments already unmarshalled.
type ParsedToolCall struct {
	ID        string                 `json:"id"`
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}

type toolCall struct {
	ID       string           `json:"id"`
	Type     string           `json:"type"`
	Function toolCallFunction `json:"function"`
}

type toolCallFunction struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"` // raw JSON string
}

// ChatWithToolsResult holds the response from a tool-capable LLM call.
type ChatWithToolsResult struct {
	Content   string
	ToolCalls []ParsedToolCall
}

// --- Request / Response types ---

// OpenAIClient is an OpenAI-compatible API client (works with DeepSeek, MIMO, OpenAI, etc.)
type OpenAIClient struct {
	baseURL         string
	apiKey          string
	model           string
	embeddingModel  string
	thinkingEnabled bool
	httpClient      *http.Client
}

type openaiChatRequest struct {
	Model      string            `json:"model"`
	Messages   []openaiMessage   `json:"messages"`
	Tools      []toolDefinition  `json:"tools,omitempty"`
	ToolChoice interface{}       `json:"tool_choice,omitempty"`
	Thinking   *thinkingOption   `json:"thinking,omitempty"`
}

type thinkingOption struct {
	Type string `json:"type"`
}

type openaiMessage struct {
	Role       string          `json:"role"`
	Content    *string         `json:"content"`               // pointer to distinguish empty vs absent
	ToolCalls  json.RawMessage `json:"tool_calls,omitempty"`  // raw JSON array for tool calls
	ToolCallID string          `json:"tool_call_id,omitempty"` // for "tool" role
}

type openaiChatResponse struct {
	Choices []struct {
		Message openaiMessage `json:"message"`
		FinishReason string   `json:"finish_reason"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

type openaiEmbeddingRequest struct {
	Model string `json:"model"`
	Input string `json:"input"`
}

type openaiEmbeddingResponse struct {
	Data []struct {
		Embedding []float64 `json:"embedding"`
	} `json:"data"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func NewOpenAIClient(baseURL, apiKey, model, embeddingModel string, thinkingEnabled bool) *OpenAIClient {
	if embeddingModel == "" {
		embeddingModel = "text-embedding-3-small"
	}
	return &OpenAIClient{
		baseURL:         baseURL,
		apiKey:          apiKey,
		model:           model,
		embeddingModel:  embeddingModel,
		thinkingEnabled: thinkingEnabled,
		httpClient: &http.Client{
			Timeout: 180 * time.Second,
		},
	}
}

// Chat sends a plain chat completion request (no tools).
func (c *OpenAIClient) Chat(messages []Message) (string, error) {
	msgs := make([]openaiMessage, len(messages))
	for i, m := range messages {
		content := m.Content
		msgs[i] = openaiMessage{Role: m.Role, Content: &content}
	}

	req := openaiChatRequest{
		Model:    c.model,
		Messages: msgs,
	}
	if c.thinkingEnabled {
		req.Thinking = &thinkingOption{Type: "enabled"}
	} else {
		req.Thinking = &thinkingOption{Type: "disabled"}
	}

	body, err := json.Marshal(req)
	if err != nil {
		return "", err
	}

	httpReq, err := http.NewRequest("POST", c.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("openai request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		body := string(respBody)
		if len(body) > 200 {
			body = body[:200] + "...(truncated)"
		}
		return "", fmt.Errorf("openai returned %d: %s", resp.StatusCode, body)
	}

	var chatResp openaiChatResponse
	if err := json.Unmarshal(respBody, &chatResp); err != nil {
		return "", err
	}
	if chatResp.Error != nil {
		return "", fmt.Errorf("openai error: %s", chatResp.Error.Message)
	}
	if len(chatResp.Choices) == 0 {
		return "", fmt.Errorf("openai returned no choices")
	}
	msg := chatResp.Choices[0].Message
	if msg.Content == nil {
		return "", nil
	}
	return *msg.Content, nil
}

// ChatWithTools sends a chat completion request with tool definitions and returns
// either text content, tool calls, or both.
func (c *OpenAIClient) ChatWithTools(messages []Message, tools []ToolDefinition, toolChoice string) (*ChatWithToolsResult, error) {
	msgs := make([]openaiMessage, len(messages))
	for i, m := range messages {
		omsg := openaiMessage{Role: m.Role}
		switch m.Role {
		case "tool":
			content := m.Content
			omsg.ToolCallID = m.ToolCallID
			omsg.Content = &content
		case "assistant":
			if len(m.ToolCallsJSON) > 0 {
				// Assistant message with tool_calls: content=null, tool_calls from ToolCallsJSON
				omsg.Content = nil
				omsg.ToolCalls = m.ToolCallsJSON
			} else {
				content := m.Content
				omsg.Content = &content
			}
		default:
			content := m.Content
			omsg.Content = &content
		}
		msgs[i] = omsg
	}

	req := openaiChatRequest{
		Model:    c.model,
		Messages: msgs,
	}

	// Attach tool definitions
	if len(tools) > 0 {
		defs := make([]toolDefinition, len(tools))
		for i, t := range tools {
			defs[i] = toolDefinition{
				Type: "function",
				Function: toolFunction{
					Name:        t.Name,
					Description: t.Description,
					Parameters:  t.Parameters,
				},
			}
		}
		req.Tools = defs
			if toolChoice == "" {
				toolChoice = "auto"
			}
			req.ToolChoice = toolChoice
	}

	// Disable thinking by default (faster, cheaper). Set AI_THINKING=true to enable.
	if c.thinkingEnabled {
		req.Thinking = &thinkingOption{Type: "enabled"}
	} else {
		req.Thinking = &thinkingOption{Type: "disabled"}
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	if len(tools) > 0 {
		log.Printf("[ai] >> ChatWithTools request (tools=%d): %s", len(tools), string(body))
	}

	httpReq, err := http.NewRequest("POST", c.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("new request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("do: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	log.Printf("[ai] << ChatWithTools response: %s", string(respBody[:min(len(respBody), 800)]))
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("openai returned %d: %s", resp.StatusCode, string(respBody[:200]))
	}

	var chatResp openaiChatResponse
	if err := json.Unmarshal(respBody, &chatResp); err != nil {
		return nil, err
	}
	if chatResp.Error != nil {
		return nil, fmt.Errorf("openai error: %s", chatResp.Error.Message)
	}
	if len(chatResp.Choices) == 0 {
		return nil, fmt.Errorf("openai returned no choices")
	}

	msg := chatResp.Choices[0].Message
	result := &ChatWithToolsResult{}

	if msg.Content != nil {
		result.Content = *msg.Content
	}

	// Parse tool_calls from raw JSON if present
	if len(msg.ToolCalls) > 0 {
		var rawCalls []struct {
			ID       string `json:"id"`
			Type     string `json:"type"`
			Function struct {
				Name      string `json:"name"`
				Arguments string `json:"arguments"`
			} `json:"function"`
		}
		if err := json.Unmarshal(msg.ToolCalls, &rawCalls); err == nil {
			result.ToolCalls = make([]ParsedToolCall, len(rawCalls))
			for i, tc := range rawCalls {
				var args map[string]interface{}
				if err := json.Unmarshal([]byte(tc.Function.Arguments), &args); err != nil {
					args = map[string]interface{}{"raw": tc.Function.Arguments}
				}
				result.ToolCalls[i] = ParsedToolCall{
					ID:        tc.ID,
					Name:      tc.Function.Name,
					Arguments: args,
				}
			}
		}
	}

	return result, nil
}

// Embedding generates an embedding vector
func (c *OpenAIClient) Embedding(text string) ([]float64, error) {
	req := openaiEmbeddingRequest{
		Model: c.embeddingModel,
		Input: text,
	}
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequest("POST", c.baseURL+"/embeddings", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("openai embedding request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		// Truncate body to avoid logging full HTML error pages
		body := string(respBody)
		if len(body) > 200 {
			body = body[:200] + "...(truncated)"
		}
		return nil, fmt.Errorf("openai returned %d: %s", resp.StatusCode, body)
	}

	var embResp openaiEmbeddingResponse
	if err := json.Unmarshal(respBody, &embResp); err != nil {
		return nil, err
	}
	if embResp.Error != nil {
		return nil, fmt.Errorf("openai error: %s", embResp.Error.Message)
	}
	if len(embResp.Data) == 0 {
		return nil, fmt.Errorf("openai returned no embeddings")
	}
	return embResp.Data[0].Embedding, nil
}

// Health checks if the API is reachable
func (c *OpenAIClient) Health() error {
	httpReq, err := http.NewRequest("GET", c.baseURL+"/models", nil)
	if err != nil {
		return err
	}
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("openai returned %d", resp.StatusCode)
	}
	return nil
}
