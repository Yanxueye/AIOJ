package rag

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"os"
	"path/filepath"
	"strings"

	"github.com/tmc/langchaingo/documentloaders"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/textsplitter"
)

// EmbedderClient is the interface for generating embeddings
type EmbedderClient interface {
	CreateEmbedding(ctx context.Context, texts []string) ([][]float32, error)
}

// DocumentWithEmbedding stores a document with its embedding vector
type DocumentWithEmbedding struct {
	Doc       schema.Document
	Embedding []float32
}

// AIEmbedder adapts the AI client's Embedding method to the EmbedderClient interface
type AIEmbedder struct {
	EmbeddingFunc func(text string) ([]float64, error)
}

func (e *AIEmbedder) CreateEmbedding(ctx context.Context, texts []string) ([][]float32, error) {
	results := make([][]float32, len(texts))
	for i, text := range texts {
		emb, err := e.EmbeddingFunc(text)
		if err != nil {
			return nil, err
		}
		float32Emb := make([]float32, len(emb))
		for j, v := range emb {
			float32Emb[j] = float32(v)
		}
		results[i] = float32Emb
	}
	return results, nil
}

// Service provides RAG (Retrieval-Augmented Generation) capabilities
type Service struct {
	documents   []DocumentWithEmbedding
	splitter    textsplitter.TextSplitter
	embedder    EmbedderClient
	initialized bool
}

// NewService creates a new RAG service
func NewService() *Service {
	return &Service{
		splitter: textsplitter.NewRecursiveCharacter(
			textsplitter.WithChunkSize(1000),
			textsplitter.WithChunkOverlap(200),
		),
	}
}

// SetEmbedder sets the embedding client for vector-based search
func (s *Service) SetEmbedder(embedder EmbedderClient) {
	s.embedder = embedder
}

// embeddingCache is the on-disk JSON structure for cached embeddings.
type embeddingCache struct {
	Version int                      `json:"version"`
	Entries map[string][]float32     `json:"entries"` // contentHash -> embedding
}

const cacheVersion = 1

// LoadFromDirectory loads and splits markdown documents from a local directory,
// using a disk cache for embeddings to avoid re-computing on every restart.
func (s *Service) LoadFromDirectory(dir string) error {
	loader := documentloaders.NewRecursiveDirLoader(
		documentloaders.WithRoot(dir),
		documentloaders.WithMaxDepth(2),
		documentloaders.WithAllowExts("md"),
	)

	ctx := context.Background()
	docs, err := loader.LoadAndSplit(ctx, s.splitter)
	if err != nil {
		return fmt.Errorf("load documents: %w", err)
	}

	// Enrich metadata from YAML front matter
	for i := range docs {
		if docs[i].Metadata == nil {
			docs[i].Metadata = make(map[string]any)
		}
		extractFrontMatter(&docs[i])
	}

	// Load embedding cache from disk
	cachePath := filepath.Join(dir, ".embedding_cache.json")
	cache := s.loadCache(cachePath)

	// Index documents with cache support
	if err := s.indexDocumentsCached(docs, cache); err != nil {
		log.Printf("[rag] warning: failed to generate embeddings, falling back to keyword search: %v", err)
		s.documents = make([]DocumentWithEmbedding, len(docs))
		for i, doc := range docs {
			s.documents[i] = DocumentWithEmbedding{Doc: doc}
		}
	} else if s.embedder != nil {
		// Save updated cache
		s.saveCache(cachePath, cache)
	}

	s.initialized = true
	log.Printf("[rag] loaded %d chunks from %s (embeddings: %v)", len(docs), dir, s.HasEmbeddings())
	return nil
}

// contentHash computes a SHA-256 hash of the document content for cache keying.
func contentHash(content string) string {
	h := sha256.Sum256([]byte(content))
	return fmt.Sprintf("%x", h[:8]) // first 16 hex chars is enough
}

func (s *Service) loadCache(path string) embeddingCache {
	cache := embeddingCache{Version: cacheVersion, Entries: make(map[string][]float32)}
	data, err := os.ReadFile(path)
	if err != nil {
		return cache
	}
	if err := json.Unmarshal(data, &cache); err != nil || cache.Version != cacheVersion {
		return embeddingCache{Version: cacheVersion, Entries: make(map[string][]float32)}
	}
	return cache
}

func (s *Service) saveCache(path string, cache embeddingCache) {
	data, err := json.Marshal(cache)
	if err != nil {
		log.Printf("[rag] cache marshal error: %v", err)
		return
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		log.Printf("[rag] cache write error: %v", err)
	}
}

// indexDocumentsCached generates embeddings only for documents not in the cache.
func (s *Service) indexDocumentsCached(docs []schema.Document, cache embeddingCache) error {
	if s.embedder == nil || len(docs) == 0 {
		s.documents = make([]DocumentWithEmbedding, len(docs))
		for i, doc := range docs {
			s.documents[i] = DocumentWithEmbedding{Doc: doc}
		}
		return nil
	}

	// Separate cached vs uncached documents
	type docEntry struct {
		index int
		hash  string
		doc   schema.Document
	}
	var uncached []docEntry
	results := make([]DocumentWithEmbedding, len(docs))

	cached := 0
	for i, doc := range docs {
		h := contentHash(doc.PageContent)
		if emb, ok := cache.Entries[h]; ok && len(emb) > 0 {
			results[i] = DocumentWithEmbedding{Doc: doc, Embedding: emb}
			cached++
		} else {
			uncached = append(uncached, docEntry{index: i, hash: h, doc: doc})
		}
	}

	if len(uncached) == 0 {
		log.Printf("[rag] all %d chunks loaded from embedding cache", len(docs))
		s.documents = results
		return nil
	}

	log.Printf("[rag] %d cached, %d need embedding generation", cached, len(uncached))

	// Generate embeddings for uncached documents (batch in groups of 20)
	batchSize := 20
	for start := 0; start < len(uncached); start += batchSize {
		end := start + batchSize
		if end > len(uncached) {
			end = len(uncached)
		}
		batch := uncached[start:end]

		texts := make([]string, len(batch))
		for i, entry := range batch {
			texts[i] = entry.doc.PageContent
		}

		embeddings, err := s.embedder.CreateEmbedding(context.Background(), texts)
		if err != nil {
			return fmt.Errorf("batch embedding (%d-%d): %w", start, end, err)
		}

		for i, entry := range batch {
			if i < len(embeddings) {
				results[entry.index] = DocumentWithEmbedding{Doc: entry.doc, Embedding: embeddings[i]}
				cache.Entries[entry.hash] = embeddings[i]
			}
		}

		log.Printf("[rag] embedded %d/%d chunks", end, len(uncached))
	}

	s.documents = results
	return nil
}

// Search performs hybrid search: embedding similarity + keyword matching
func (s *Service) Search(query string, topK int) []SearchResult {
	if !s.initialized || len(s.documents) == 0 {
		return nil
	}

	type scored struct {
		doc   DocumentWithEmbedding
		score float64
	}

	var results []scored

	// Try embedding-based search first
	if s.embedder != nil {
		ctx := context.Background()
		queryEmb, err := s.embedder.CreateEmbedding(ctx, []string{query})
		if err == nil && len(queryEmb) > 0 {
			for _, doc := range s.documents {
				if len(doc.Embedding) > 0 {
					sim := cosineSimilarity(queryEmb[0], doc.Embedding)
					if sim > 0.1 { // threshold
						results = append(results, scored{doc: doc, score: float64(sim)})
					}
				}
			}
		}
	}

	// If no embedding results, fall back to keyword search
	if len(results) == 0 {
		queryLower := strings.ToLower(query)
		words := strings.Fields(queryLower)

		for _, doc := range s.documents {
			content := strings.ToLower(doc.Doc.PageContent)
			score := 0.0
			for _, word := range words {
				if strings.Contains(content, word) {
					score += 1.0
				}
			}
			if score > 0 {
				results = append(results, scored{doc: doc, score: score / float64(len(words))})
			}
		}
	}

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
			Document: Document{
				ID:       fmt.Sprintf("%v", results[i].doc.Doc.Metadata["id"]),
				Content:  results[i].doc.Doc.PageContent,
				Metadata: toStringMap(results[i].doc.Doc.Metadata),
			},
			Similarity: results[i].score,
		}
	}
	return searchResults
}

// BuildContext builds a context string from search results for injection into AI prompts
func (s *Service) BuildContext(query string, maxLen int) string {
	results := s.Search(query, 5)
	if len(results) == 0 {
		return ""
	}
	return BuildContext(results, maxLen)
}

// BuildCodeContext builds context for code analysis
func (s *Service) BuildCodeContext(code string, language string, maxLen int) string {
	query := fmt.Sprintf("Algorithm and data structure concepts in %s code", language)
	results := s.Search(query, 3)
	if len(results) == 0 {
		return ""
	}
	var sb strings.Builder
	sb.WriteString("以下是从 OI-Wiki 检索到的相关知识：\n\n")
	sb.WriteString(BuildContext(results, maxLen))
	return sb.String()
}

// IsInitialized returns whether the RAG service has been initialized
func (s *Service) IsInitialized() bool {
	return s.initialized
}

// DocumentCount returns the number of indexed documents
func (s *Service) DocumentCount() int {
	return len(s.documents)
}

// HasEmbeddings returns whether the service has embedding vectors
func (s *Service) HasEmbeddings() bool {
	if len(s.documents) == 0 {
		return false
	}
	return len(s.documents[0].Embedding) > 0
}

// cosineSimilarity computes cosine similarity between two vectors
func cosineSimilarity(a, b []float32) float32 {
	if len(a) != len(b) || len(a) == 0 {
		return 0
	}

	var dotProduct, normA, normB float64
	for i := range a {
		dotProduct += float64(a[i]) * float64(b[i])
		normA += float64(a[i]) * float64(a[i])
		normB += float64(b[i]) * float64(b[i])
	}

	if normA == 0 || normB == 0 {
		return 0
	}
	return float32(dotProduct / (math.Sqrt(normA) * math.Sqrt(normB)))
}

// extractFrontMatter extracts metadata from YAML front matter in document content
func extractFrontMatter(doc *schema.Document) {
	content := doc.PageContent
	if !strings.HasPrefix(content, "---") {
		return
	}

	endIdx := strings.Index(content[3:], "---")
	if endIdx < 0 {
		return
	}

	frontMatter := content[3 : endIdx+3]
	lines := strings.Split(frontMatter, "\n")
	for _, line := range lines {
		parts := strings.SplitN(line, ":", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			value = strings.Trim(value, "\"'")
			if key != "" && value != "" {
				doc.Metadata[key] = value
			}
		}
	}
}

// toStringMap converts map[string]any to map[string]string
func toStringMap(m map[string]any) map[string]string {
	result := make(map[string]string, len(m))
	for k, v := range m {
		if s, ok := v.(string); ok {
			result[k] = s
		} else {
			result[k] = fmt.Sprintf("%v", v)
		}
	}
	return result
}
