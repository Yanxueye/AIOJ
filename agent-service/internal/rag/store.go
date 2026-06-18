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

// BuildContext builds a context string from search results, truncating at maxLen if > 0.
func BuildContext(results []SearchResult, maxLen int) string {
	var context string
	for i, result := range results {
		content := result.Document.Content
		if maxLen > 0 && len(context)+len(content) > maxLen {
			if maxLen > len(context) {
				context += content[:maxLen-len(context)]
			}
			break
		}
		if i > 0 {
			context += "\n---\n"
		}
		context += content
	}
	return context
}
