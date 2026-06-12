package rag

// Document represents a document in the RAG system
type Document struct {
	ID       string            `json:"id"`
	Content  string            `json:"content"`
	Metadata map[string]string `json:"metadata"`
}

// SearchResult represents a search result with similarity score
type SearchResult struct {
	Document   Document `json:"document"`
	Similarity float64  `json:"similarity"`
}

// BuildContext builds a context string from search results
func BuildContext(results []SearchResult, maxLen int) string {
	totalLen := 0
	var context string
	for i, result := range results {
		if totalLen >= maxLen {
			break
		}
		content := result.Document.Content
		if totalLen+len(content) > maxLen {
			content = content[:maxLen-totalLen]
		}
		if i > 0 {
			context += "\n---\n"
		}
		context += content
		totalLen += len(content)
	}
	return context
}
