package rag

import (
	"context"
	"fmt"
	"strings"

	"agent-service/internal/ai"
)

// Ensure context is used (it's part of the interface signature)
var _ = context.Background

// Retriever combines vector search with knowledge graph retrieval
type Retriever struct {
	store    *VectorStore
	aiClient *ai.Client
}

// NewRetriever creates a new RAG retriever
func NewRetriever(store *VectorStore, aiClient *ai.Client) *Retriever {
	return &Retriever{
		store:    store,
		aiClient: aiClient,
	}
}

// IndexDocument indexes a document by generating its embedding and storing it
func (r *Retriever) IndexDocument(ctx context.Context, doc *Document) error {
	if r.aiClient == nil {
		return fmt.Errorf("AI client not configured")
	}

	embedding, err := r.aiClient.Embedding(doc.Content)
	if err != nil {
		return fmt.Errorf("generate embedding: %w", err)
	}

	doc.Vector = embedding
	r.store.Add(doc)
	return nil
}

// IndexBatch indexes multiple documents
func (r *Retriever) IndexBatch(ctx context.Context, docs []*Document) error {
	for _, doc := range docs {
		if err := r.IndexDocument(ctx, doc); err != nil {
			return err
		}
	}
	return nil
}

// Retrieve finds relevant documents for a query
func (r *Retriever) Retrieve(ctx context.Context, query string, topK int) ([]SearchResult, error) {
	if r.aiClient == nil {
		return r.keywordSearch(query, topK), nil
	}

	// Generate query embedding
	queryVector, err := r.aiClient.Embedding(query)
	if err != nil {
		// Fall back to keyword search
		return r.keywordSearch(query, topK), nil
	}

	// Vector similarity search
	results := r.store.Search(queryVector, topK)
	if len(results) == 0 {
		// Fall back to keyword search
		return r.keywordSearch(query, topK), nil
	}

	return results, nil
}

// RetrieveForCode finds relevant context for code analysis
func (r *Retriever) RetrieveForCode(ctx context.Context, code string, language string, topK int) ([]SearchResult, error) {
	query := fmt.Sprintf("Algorithm and data structure concepts in %s code:\n%s", language, truncate(code, 500))
	return r.Retrieve(ctx, query, topK)
}

// keywordSearch performs a simple keyword-based search as fallback
func (r *Retriever) keywordSearch(query string, topK int) []SearchResult {
	query = strings.ToLower(query)
	words := strings.Fields(query)

	type scored struct {
		doc   Document
		score float64
	}

	var results []scored
	r.store.mu.RLock()
	for _, doc := range r.store.documents {
		content := strings.ToLower(doc.Content)
		score := 0.0
		for _, word := range words {
			if strings.Contains(content, word) {
				score += 1.0
			}
		}
		if score > 0 {
			results = append(results, scored{doc: *doc, score: score})
		}
	}
	r.store.mu.RUnlock()

	// Sort by score
	for i := 0; i < len(results); i++ {
		for j := i + 1; j < len(results); j++ {
			if results[j].score > results[i].score {
				results[i], results[j] = results[j], results[i]
			}
		}
	}

	if topK > len(results) {
		topK = len(results)
	}

	searchResults := make([]SearchResult, topK)
	for i := 0; i < topK; i++ {
		searchResults[i] = SearchResult{
			Document:   results[i].doc,
			Similarity: results[i].score / float64(len(words)),
		}
	}
	return searchResults
}

// BuildContext builds a context string from search results
func BuildContext(results []SearchResult, maxLen int) string {
	var sb strings.Builder
	totalLen := 0

	for i, result := range results {
		if totalLen >= maxLen {
			break
		}
		content := result.Document.Content
		if totalLen+len(content) > maxLen {
			content = content[:maxLen-totalLen]
		}
		if i > 0 {
			sb.WriteString("\n---\n")
		}
		sb.WriteString(content)
		totalLen += len(content)
	}

	return sb.String()
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
