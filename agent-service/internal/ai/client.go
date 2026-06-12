package ai

import (
	"log"
)

// Message represents a chat message
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Client is a unified AI client that tries OpenAI-compatible API first, falls back to Ollama
type Client struct {
	primary  *OpenAIClient
	fallback *OllamaClient
	provider string
}

// NewClient creates a new AI client with primary (OpenAI-compatible) and fallback (Ollama)
func NewClient(openaiKey, openaiBaseURL, openaiModel, ollamaURL, ollamaModel, provider string) *Client {
	c := &Client{
		provider: provider,
	}

	if openaiKey != "" {
		c.primary = NewOpenAIClient(openaiBaseURL, openaiKey, openaiModel)
		log.Printf("[ai] primary: OpenAI-compatible (%s, model=%s)", openaiBaseURL, openaiModel)
	}

	if ollamaURL != "" {
		c.fallback = NewOllamaClient(ollamaURL, ollamaModel)
		log.Printf("[ai] fallback: Ollama (%s, model=%s)", ollamaURL, ollamaModel)
	}

	return c
}

// Chat sends a chat completion request, trying primary first then fallback
func (c *Client) Chat(messages []Message) (string, error) {
	// Try primary (OpenAI-compatible API) first
	if c.primary != nil && c.provider == "openai" {
		resp, err := c.primary.Chat(messages)
		if err == nil {
			return resp, nil
		}
		log.Printf("[ai] primary API failed: %v, falling back to Ollama", err)
	}

	// Fallback to Ollama
	if c.fallback != nil {
		resp, err := c.fallback.Chat(messages)
		if err == nil {
			return resp, nil
		}
		return "", err
	}

	// If provider is "ollama", try Ollama first
	if c.fallback != nil && c.provider == "ollama" {
		resp, err := c.fallback.Chat(messages)
		if err == nil {
			return resp, nil
		}
		log.Printf("[ai] Ollama failed: %v, trying primary API", err)
		if c.primary != nil {
			return c.primary.Chat(messages)
		}
		return "", err
	}

	// Last resort: try whatever is available
	if c.primary != nil {
		return c.primary.Chat(messages)
	}
	if c.fallback != nil {
		return c.fallback.Chat(messages)
	}

	return "", ErrNoProvider
}

// Embedding generates an embedding vector, trying primary first then fallback
func (c *Client) Embedding(text string) ([]float64, error) {
	if c.primary != nil && c.provider == "openai" {
		emb, err := c.primary.Embedding(text)
		if err == nil {
			return emb, nil
		}
		log.Printf("[ai] primary embedding failed: %v, falling back to Ollama", err)
	}

	if c.fallback != nil {
		return c.fallback.Embedding(text)
	}

	if c.primary != nil {
		return c.primary.Embedding(text)
	}

	return nil, ErrNoProvider
}

// Health checks if any AI provider is reachable
func (c *Client) Health() error {
	if c.primary != nil {
		if err := c.primary.Health(); err == nil {
			return nil
		}
	}
	if c.fallback != nil {
		return c.fallback.Health()
	}
	return ErrNoProvider
}
