package rag

import (
	"math"
	"sort"
	"sync"
)

// Document represents a document stored in the vector store
type Document struct {
	ID       string            `json:"id"`
	Content  string            `json:"content"`
	Metadata map[string]string `json:"metadata"`
	Vector   []float64         `vector:"-"` // embedding vector
}

// SearchResult represents a search result with similarity score
type SearchResult struct {
	Document   Document `json:"document"`
	Similarity float64  `json:"similarity"`
}

// VectorStore is an in-memory vector store for RAG
type VectorStore struct {
	mu        sync.RWMutex
	documents map[string]*Document
	dimension int
}

// NewVectorStore creates a new in-memory vector store
func NewVectorStore(dimension int) *VectorStore {
	return &VectorStore{
		documents: make(map[string]*Document),
		dimension: dimension,
	}
}

// Add adds a document to the store
func (vs *VectorStore) Add(doc *Document) {
	vs.mu.Lock()
	defer vs.mu.Unlock()
	vs.documents[doc.ID] = doc
}

// AddBatch adds multiple documents to the store
func (vs *VectorStore) AddBatch(docs []*Document) {
	vs.mu.Lock()
	defer vs.mu.Unlock()
	for _, doc := range docs {
		vs.documents[doc.ID] = doc
	}
}

// Get retrieves a document by ID
func (vs *VectorStore) Get(id string) (*Document, bool) {
	vs.mu.RLock()
	defer vs.mu.RUnlock()
	doc, ok := vs.documents[id]
	return doc, ok
}

// Delete removes a document by ID
func (vs *VectorStore) Delete(id string) {
	vs.mu.Lock()
	defer vs.mu.Unlock()
	delete(vs.documents, id)
}

// Count returns the number of documents in the store
func (vs *VectorStore) Count() int {
	vs.mu.RLock()
	defer vs.mu.RUnlock()
	return len(vs.documents)
}

// Search finds the top-k most similar documents to the query vector
func (vs *VectorStore) Search(queryVector []float64, topK int) []SearchResult {
	vs.mu.RLock()
	defer vs.mu.RUnlock()

	if len(queryVector) != vs.dimension || topK <= 0 {
		return nil
	}

	type scored struct {
		doc        *Document
		similarity float64
	}

	results := make([]scored, 0, len(vs.documents))
	for _, doc := range vs.documents {
		if len(doc.Vector) != vs.dimension {
			continue
		}
		sim := cosineSimilarity(queryVector, doc.Vector)
		results = append(results, scored{doc: doc, similarity: sim})
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].similarity > results[j].similarity
	})

	if topK > len(results) {
		topK = len(results)
	}

	searchResults := make([]SearchResult, topK)
	for i := 0; i < topK; i++ {
		searchResults[i] = SearchResult{
			Document:   *results[i].doc,
			Similarity: results[i].similarity,
		}
	}
	return searchResults
}

// SearchByMetadata finds documents matching metadata key-value pairs
func (vs *VectorStore) SearchByMetadata(key, value string, limit int) []Document {
	vs.mu.RLock()
	defer vs.mu.RUnlock()

	results := make([]Document, 0)
	for _, doc := range vs.documents {
		if v, ok := doc.Metadata[key]; ok && v == value {
			results = append(results, *doc)
			if limit > 0 && len(results) >= limit {
				break
			}
		}
	}
	return results
}

// cosineSimilarity computes cosine similarity between two vectors
func cosineSimilarity(a, b []float64) float64 {
	if len(a) != len(b) || len(a) == 0 {
		return 0
	}

	var dotProduct, normA, normB float64
	for i := range a {
		dotProduct += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}

	if normA == 0 || normB == 0 {
		return 0
	}
	return dotProduct / (math.Sqrt(normA) * math.Sqrt(normB))
}
