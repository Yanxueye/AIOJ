package ai

import (
	"encoding/json"
	"log"
)

// Message represents a chat message.
type Message struct {
	Role          string          `json:"role"`
	Content       string          `json:"content,omitempty"`
	ToolCallID    string          `json:"tool_call_id,omitempty"`
	ToolCallsJSON json.RawMessage `json:"tool_calls,omitempty"`
}

// Client is a unified AI client that tries OpenAI-compatible API first, falls back to Ollama
type Client struct {
	primary  *OpenAIClient
	fallback *OllamaClient
	provider string
}

// NewClient creates a new AI client with primary (OpenAI-compatible) and fallback (Ollama)
func NewClient(openaiKey, openaiBaseURL, openaiModel, ollamaURL, ollamaModel, provider, embeddingModel string, thinkingEnabled bool) *Client {
	c := &Client{
		provider: provider,
	}

	if openaiKey != "" {
		c.primary = NewOpenAIClient(openaiBaseURL, openaiKey, openaiModel, embeddingModel, thinkingEnabled)
		log.Printf("[ai] primary: OpenAI-compatible (%s, model=%s, embedding=%s, thinking=%v)", openaiBaseURL, openaiModel, embeddingModel, thinkingEnabled)
	}

	if ollamaURL != "" {
		c.fallback = NewOllamaClient(ollamaURL, ollamaModel, embeddingModel, thinkingEnabled)
		log.Printf("[ai] fallback: Ollama (%s, model=%s, embedding=%s, thinking=%v)", ollamaURL, ollamaModel, embeddingModel, thinkingEnabled)
	}

	return c
}

// Chat sends a chat completion request, trying providers based on configured preference
func (c *Client) Chat(messages []Message) (string, error) {
	switch c.provider {
	case "ollama":
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
		if c.primary != nil {
			resp, err := c.primary.Chat(messages)
			if err == nil {
				return resp, nil
			}
			log.Printf("[ai] primary API error: %v", err)
		} else {
			log.Printf("[ai] primary (OpenAI) is nil — API key missing or empty?")
		}
		if c.fallback != nil {
			resp, err := c.fallback.Chat(messages)
			if err == nil { return resp, nil }
			log.Printf("[ai] fallback Ollama also failed: %v", err)
		}
	default:
		if c.primary != nil {
			resp, err := c.primary.Chat(messages)
			if err == nil { return resp, nil }
			log.Printf("[ai] primary API error: %v", err)
		}
		if c.fallback != nil {
			resp, err := c.fallback.Chat(messages)
			if err == nil { return resp, nil }
			log.Printf("[ai] fallback also failed: %v", err)
		}
	}
	log.Printf("[ai] Chat failed — no provider available (primary=%v, fallback=%v)", c.primary != nil, c.fallback != nil)
	return "", ErrNoProvider
}

// ChatWithTools sends a chat completion request with tool definitions and returns
// either text content, tool calls, or both.
// Ollama fallback does not support tools — if Ollama is the only provider, tools are dropped.
func (c *Client) ChatWithTools(messages []Message, tools []ToolDefinition, toolChoice string) (*ChatWithToolsResult, error) {
	// If primary (OpenAI-compatible) is available, use it
	if c.primary != nil {
		resp, err := c.primary.ChatWithTools(messages, tools, toolChoice)
		if err == nil {
			return resp, nil
		}
		log.Printf("[ai] primary ChatWithTools failed: %v", err)
		// Fall through to fallback
	}

	// Fallback to Ollama without tools
	if c.fallback != nil {
		text, err := c.fallback.Chat(messages)
		if err != nil {
			return nil, err
		}
		return &ChatWithToolsResult{Content: text}, nil
	}

	return nil, ErrNoProvider
}

// ProviderName returns the name of the currently active provider.
func (c *Client) ProviderName() string {
	if c.primary != nil {
		return "agent-service"
	}
	if c.fallback != nil {
		return "ollama"
	}
	return "unavailable"
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
