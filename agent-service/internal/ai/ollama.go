package ai

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

var ErrNoProvider = errors.New("no AI provider configured")

// OllamaClient is a client for the local Ollama API
type OllamaClient struct {
	baseURL         string
	model           string
	embeddingModel  string // separate model for embeddings
	thinkingEnabled bool
	httpClient      *http.Client
}

type ollamaChatRequest struct {
	Model    string          `json:"model"`
	Messages []ollamaMessage `json:"messages"`
	Stream   bool            `json:"stream"`
}

type ollamaMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ollamaChatResponse struct {
	Message ollamaMessage `json:"message"`
	Done    bool          `json:"done"`
}

type ollamaEmbeddingRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
}

type ollamaEmbeddingResponse struct {
	Embedding []float64 `json:"embedding"`
}

func NewOllamaClient(baseURL, model, embeddingModel string, thinkingEnabled bool) *OllamaClient {
	if embeddingModel == "" {
		embeddingModel = model
	}
	return &OllamaClient{
		baseURL:         baseURL,
		model:           model,
		embeddingModel:  embeddingModel,
		thinkingEnabled: thinkingEnabled,
		httpClient: &http.Client{
			Timeout: 180 * time.Second,
		},
	}
}

// Chat sends a chat completion request to Ollama
func (c *OllamaClient) Chat(messages []Message) (string, error) {
	msgs := make([]ollamaMessage, len(messages))
	for i, m := range messages {
		content := m.Content
		// Prepend /no_think to system message to disable thinking mode for speed
		if !c.thinkingEnabled && i == 0 && m.Role == "system" {
			content = "/no_think\n" + content
		}
		msgs[i] = ollamaMessage{Role: m.Role, Content: content}
	}

	req := ollamaChatRequest{
		Model:    c.model,
		Messages: msgs,
		Stream:   false,
	}
	body, err := json.Marshal(req)
	if err != nil {
		return "", err
	}

	resp, err := c.httpClient.Post(c.baseURL+"/api/chat", "application/json", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("ollama request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("ollama returned %d: %s", resp.StatusCode, string(respBody))
	}

	var chatResp ollamaChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		return "", err
	}
	return chatResp.Message.Content, nil
}

// Embedding generates an embedding vector for the given text
func (c *OllamaClient) Embedding(text string) ([]float64, error) {
	req := ollamaEmbeddingRequest{
		Model:  c.embeddingModel,
		Prompt: text,
	}
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Post(c.baseURL+"/api/embeddings", "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("ollama embedding request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("ollama returned %d: %s", resp.StatusCode, string(respBody))
	}

	var embResp ollamaEmbeddingResponse
	if err := json.NewDecoder(resp.Body).Decode(&embResp); err != nil {
		return nil, err
	}
	return embResp.Embedding, nil
}

// Health checks if Ollama is reachable
func (c *OllamaClient) Health() error {
	resp, err := c.httpClient.Get(c.baseURL + "/api/tags")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("ollama returned %d", resp.StatusCode)
	}
	return nil
}
