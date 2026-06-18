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
			textsplitter.WithChunkSize(1500),
			textsplitter.WithChunkOverlap(300),
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
	rawDocs, err := loader.Load(ctx)
	if err != nil {
		return fmt.Errorf("load documents: %w", err)
	}
	// Split by ## headings (keep section+content together)
	docs := splitByHeadings(rawDocs, 2000)
	log.Printf("[rag] split %d raw docs into %d chunks by headings", len(rawDocs), len(docs))

	// Enrich metadata from YAML front matter
	for i := range docs {
		if docs[i].Metadata == nil {
			docs[i].Metadata = make(map[string]any)
		}
		extractFrontMatter(&docs[i])
		// Derive search tags from front matter (title + category)
		tags := []string{}
		title := ""
		category := ""
		if t, ok := docs[i].Metadata["title"].(string); ok && t != "" {
			tags = append(tags, t)
			title = t
		}
		if c, ok := docs[i].Metadata["category"].(string); ok && c != "" {
			tags = append(tags, c)
			category = c
		}
		if len(tags) > 0 {
			docs[i].Metadata["tags"] = tags
		}
		// Prefix chunk content with document context so LLM understands what topic it belongs to
		prefix := ""
		if title != "" { prefix += "[" + title + "]" }
		if category != "" { prefix += " [" + category + "]" }
		if prefix != "" {
			docs[i].PageContent = prefix + "\n" + docs[i].PageContent
		}
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

// Search performs hybrid search: embedding similarity + keyword matching + tag boost.
// Results are deduplicated by content hash to avoid duplicate chunks.
func (s *Service) Search(query string, topK int) []SearchResult {
	return s.searchWithBoost(query, topK, nil)
}

func (s *Service) searchWithBoost(query string, topK int, boostTags []string) []SearchResult {
	if !s.initialized || len(s.documents) == 0 {
		return nil
	}

	type scored struct {
		doc   DocumentWithEmbedding
		score float64
	}

	var results []scored
	seen := make(map[string]bool) // dedup by content hash

	// Try embedding-based search first
	if s.embedder != nil {
		ctx := context.Background()
		queryEmb, err := s.embedder.CreateEmbedding(ctx, []string{query})
		if err == nil && len(queryEmb) > 0 {
			for _, doc := range s.documents {
				if len(doc.Embedding) > 0 {
					sim := cosineSimilarity(queryEmb[0], doc.Embedding)
					if sim > 0.1 { // threshold
						h := contentHash(doc.Doc.PageContent)
						if seen[h] {
							continue
						}
						seen[h] = true
						score := float64(sim)
						if tagMatchBoost(boostTags, doc.Doc.Metadata) {
							score *= 3.0
						}
						results = append(results, scored{doc: doc, score: score})
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
				h := contentHash(doc.Doc.PageContent)
				if seen[h] {
					continue
				}
				seen[h] = true
				if tagMatchBoost(boostTags, doc.Doc.Metadata) {
					score *= 3.0
				}
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

// SearchMulti performs multiple searches and returns deduplicated merged results.
func (s *Service) SearchMulti(queries []string, topK int) []SearchResult {
	if len(queries) == 0 {
		return nil
	}
	if len(queries) == 1 {
		return s.Search(queries[0], topK)
	}

	type scored struct {
		doc   DocumentWithEmbedding
		score float64
	}
	merged := make(map[string]*scored) // contentHash -> result

	for _, q := range queries {
		q = strings.TrimSpace(q)
		if q == "" {
			continue
		}
		results := s.Search(q, topK*2)
		for _, r := range results {
			h := contentHash(r.Document.Content)
			if existing, ok := merged[h]; ok {
				if r.Similarity > existing.score {
					existing.score = r.Similarity
				}
			} else {
				for _, doc := range s.documents {
					if contentHash(doc.Doc.PageContent) == h {
						merged[h] = &scored{doc: doc, score: r.Similarity}
						break
					}
				}
			}
		}
	}

	var results []scored
	for _, v := range merged {
		results = append(results, *v)
	}
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

// BuildContext searches and builds context (maxLen=0 means no limit).
func (s *Service) BuildContext(query string, maxLen int) string {
	results := s.Search(query, 5)
	if len(results) == 0 {
		return ""
	}
	return BuildContext(results, maxLen)
}

// BuildContextMulti searches using multiple queries, merges deduplicated results.
func (s *Service) BuildContextMulti(queries []string, maxLen, topK int) string {
	if len(queries) == 0 {
		return ""
	}
	results := s.SearchMulti(queries, topK)
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

// splitByHeadings splits documents by ## headings, keeping sections intact.
// Sections larger than maxLen are further split at paragraph boundaries.
func splitByHeadings(docs []schema.Document, maxLen int) []schema.Document {
	var result []schema.Document
	for _, doc := range docs {
		sections := splitByMarker(doc.PageContent, "\n## ", maxLen)
		for _, sec := range sections {
			newDoc := schema.Document{
				PageContent: sec,
				Metadata:    make(map[string]any),
			}
			for k, v := range doc.Metadata {
				newDoc.Metadata[k] = v
			}
			result = append(result, newDoc)
		}
	}
	return result
}

func splitByMarker(content, marker string, maxLen int) []string {
	// Split on marker boundaries
	parts := strings.Split(content, marker)
	if len(parts) <= 1 {
		// No marker found, try to split long content at paragraphs
		if len(content) <= maxLen {
			return []string{content}
		}
		return splitPara(content, maxLen)
	}

	var result []string
	var current strings.Builder
	for i, part := range parts {
		prefix := ""
		if i > 0 {
			prefix = marker
		}
		chunk := prefix + part
		// If this section is small enough, accumulate
		if current.Len()+len(chunk) <= maxLen {
			current.WriteString(chunk)
		} else {
			// Flush previous accumulator
			if current.Len() > 0 {
				result = append(result, current.String())
				current.Reset()
			}
			// If this section alone fits, start new accumulator
			if len(chunk) <= maxLen {
				current.WriteString(chunk)
			} else {
				// Section too big, split at paragraphs
				subs := splitPara(chunk, maxLen)
				result = append(result, subs[:len(subs)-1]...)
				current.WriteString(subs[len(subs)-1])
			}
		}
	}
	if current.Len() > 0 {
		result = append(result, current.String())
	}
	return result
}

func splitPara(content string, maxLen int) []string {
	if len(content) <= maxLen {
		return []string{content}
	}
	// Split on double newline (paragraph boundary)
	paras := strings.Split(content, "\n\n")
	var result []string
	var current strings.Builder
	for _, para := range paras {
		if current.Len()+len(para) <= maxLen {
			if current.Len() > 0 { current.WriteString("\n\n") }
			current.WriteString(para)
		} else {
			if current.Len() > 0 {
				result = append(result, current.String())
				current.Reset()
			}
			if len(para) <= maxLen {
				current.WriteString(para)
			} else {
				// Para still too big, hard split
				for i := 0; i < len(para); i += maxLen {
					end := i + maxLen
					if end > len(para) { end = len(para) }
					result = append(result, para[i:end])
				}
			}
		}
	}
	if current.Len() > 0 {
		result = append(result, current.String())
	}
	return result
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

// tagMatchBoost returns true if any boostTag matches a tag in the document metadata.
func tagMatchBoost(boostTags []string, metadata map[string]any) bool {
	if len(boostTags) == 0 { return false }
	docTags, _ := metadata["tags"].([]string)
	for _, bt := range boostTags {
		for _, dt := range docTags {
			if strings.Contains(strings.ToLower(dt), strings.ToLower(bt)) {
				return true
			}
		}
	}
	return false
}

// BuildContextTagged searches with tag boosting for better relevance.
func (s *Service) BuildContextTagged(queries []string, tags []string, maxLen, topK int) string {
	if len(queries) == 0 { return "" }
	var allResults []SearchResult
	seen := make(map[string]bool)
	for _, q := range queries {
		q = strings.TrimSpace(q)
		if q == "" { continue }
		results := s.searchWithBoost(q, topK*2, tags)
		for _, r := range results {
			h := contentHash(r.Document.Content)
			if seen[h] { continue }
			seen[h] = true
			allResults = append(allResults, r)
		}
	}
	// Sort by similarity descending
	for i := 0; i < len(allResults); i++ {
		for j := i + 1; j < len(allResults); j++ {
			if allResults[j].Similarity > allResults[i].Similarity {
				allResults[i], allResults[j] = allResults[j], allResults[i]
			}
		}
	}
	if topK < len(allResults) { allResults = allResults[:topK] }
	if len(allResults) == 0 { return "" }
	return BuildContext(allResults, maxLen)
}
