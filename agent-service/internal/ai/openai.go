package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// OpenAIClient is an OpenAI-compatible API client (works with MIMO, OpenAI, etc.)
type OpenAIClient struct {
	baseURL        string
	apiKey         string
	model          string
	embeddingModel string
	thinkingEnabled bool
	httpClient     *http.Client
}

type openaiChatRequest struct {
	Model    string          `json:"model"`
	Messages []openaiMessage `json:"messages"`
	Thinking *thinkingOption `json:"thinking,omitempty"`
}

type thinkingOption struct {
	Enabled bool `json:"enabled"`
}

type openaiMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openaiChatResponse struct {
	Choices []struct {
		Message openaiMessage `json:"message"`
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

// Chat sends a chat completion request
func (c *OpenAIClient) Chat(messages []Message) (string, error) {
	msgs := make([]openaiMessage, len(messages))
	for i, m := range messages {
		msgs[i] = openaiMessage{Role: m.Role, Content: m.Content}
	}

	req := openaiChatRequest{
		Model:    c.model,
		Messages: msgs,
	}
	// thinking field is OpenAI-specific; DeepSeek and most compatible APIs reject it
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
	return chatResp.Choices[0].Message.Content, nil
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
