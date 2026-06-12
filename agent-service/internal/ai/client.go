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
func NewClient(openaiKey, openaiBaseURL, openaiModel, ollamaURL, ollamaModel, provider, embeddingModel string) *Client {
	c := &Client{
		provider: provider,
	}

	if openaiKey != "" {
		c.primary = NewOpenAIClient(openaiBaseURL, openaiKey, openaiModel, embeddingModel)
		log.Printf("[ai] primary: OpenAI-compatible (%s, model=%s, embedding=%s)", openaiBaseURL, openaiModel, embeddingModel)
	}

	if ollamaURL != "" {
		c.fallback = NewOllamaClient(ollamaURL, ollamaModel, embeddingModel)
		log.Printf("[ai] fallback: Ollama (%s, model=%s, embedding=%s)", ollamaURL, ollamaModel, embeddingModel)
	}

	return c
}

// Chat sends a chat completion request, trying providers based on configured preference
func (c *Client) Chat(messages []Message) (string, error) {
	switch c.provider {
	case "ollama":
		// Ollama-first: try Ollama, fall back to OpenAI-compatible
		if c.fallback != nil {
			resp, err := c.fallback.Chat(messages)
			if err == nil {
				return resp, nil
			}
			log.Printf("[ai] Ollama failed: %v, trying primary API", err)
		}
		if c.primary != nil {
			return c.primary.Chat(messages)
		}
	case "openai":
		// OpenAI-first: try OpenAI-compatible, fall back to Ollama
		if c.primary != nil {
			resp, err := c.primary.Chat(messages)
			if err == nil {
				return resp, nil
			}
			log.Printf("[ai] primary API failed: %v, falling back to Ollama", err)
		}
		if c.fallback != nil {
			return c.fallback.Chat(messages)
		}
	default:
		// No preference: try whatever is available
		if c.primary != nil {
			resp, err := c.primary.Chat(messages)
			if err == nil {
				return resp, nil
			}
			log.Printf("[ai] primary API failed: %v, trying fallback", err)
		}
		if c.fallback != nil {
			return c.fallback.Chat(messages)
		}
	}

	return "", ErrNoProvider
}

// Embedding generates an embedding vector.
// Unlike Chat, embeddings always prefer Ollama first (where nomic-embed-text runs),
// since OpenAI-compatible providers (e.g. MIMO) may not support /embeddings.
func (c *Client) Embedding(text string) ([]float64, error) {
	// Try Ollama first — embedding models live there
	if c.fallback != nil {
		emb, err := c.fallback.Embedding(text)
		if err == nil {
			return emb, nil
		}
		log.Printf("[ai] Ollama embedding failed: %v, trying primary API", err)
	}
	// Fall back to primary (OpenAI-compatible) if Ollama unavailable
	if c.primary != nil {
		emb, err := c.primary.Embedding(text)
		if err == nil {
			return emb, nil
		}
		// Log only the status code, not the full HTML error body
		log.Printf("[ai] primary embedding failed: %v", err)
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
