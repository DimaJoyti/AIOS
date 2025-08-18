package memory

import (
	"context"
	"fmt"
	"math"
	"sort"
	"sync"

	"github.com/aios/aios/pkg/langchain/llm"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// DefaultVectorMemory implements VectorMemory interface
type DefaultVectorMemory struct {
	documents       map[string]*Document
	vectors         map[string][]float64
	maxDocuments    int
	vectorDimension int
	threshold       float64
	embeddingLLM    llm.LLM
	vectorStore     VectorStore
	memoryKeys      []string
	logger          *logrus.Logger
	tracer          trace.Tracer
	mu              sync.RWMutex
}

// VectorMemoryConfig represents configuration for vector memory
type VectorMemoryConfig struct {
	MaxDocuments    int         `json:"max_documents,omitempty"`
	VectorDimension int         `json:"vector_dimension,omitempty"`
	Threshold       float64     `json:"threshold,omitempty"`
	EmbeddingLLM    llm.LLM     `json:"-"`
	VectorStore     VectorStore `json:"-"`
	MemoryKeys      []string    `json:"memory_keys,omitempty"`
}

// NewVectorMemory creates a new vector memory
func NewVectorMemory(config *VectorMemoryConfig, logger *logrus.Logger) (VectorMemory, error) {
	if config.MaxDocuments <= 0 {
		config.MaxDocuments = 1000 // Default max documents
	}

	if config.VectorDimension <= 0 {
		config.VectorDimension = 1536 // Default OpenAI embedding dimension
	}

	if config.Threshold <= 0 {
		config.Threshold = 0.7 // Default similarity threshold
	}

	memoryKeys := config.MemoryKeys
	if len(memoryKeys) == 0 {
		memoryKeys = []string{"relevant_documents", "context"}
	}

	memory := &DefaultVectorMemory{
		documents:       make(map[string]*Document),
		vectors:         make(map[string][]float64),
		maxDocuments:    config.MaxDocuments,
		vectorDimension: config.VectorDimension,
		threshold:       config.Threshold,
		embeddingLLM:    config.EmbeddingLLM,
		vectorStore:     config.VectorStore,
		memoryKeys:      memoryKeys,
		logger:          logger,
		tracer:          otel.Tracer("langchain.memory.vector"),
	}

	return memory, nil
}

// LoadMemoryVariables loads memory variables for the given input
func (m *DefaultVectorMemory) LoadMemoryVariables(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error) {
	ctx, span := m.tracer.Start(ctx, "vector_memory.load_variables")
	defer span.End()

	variables := make(map[string]interface{})

	// Extract query from input
	query := ""
	if q, exists := input["input"]; exists {
		query = fmt.Sprintf("%v", q)
	} else if q, exists := input["query"]; exists {
		query = fmt.Sprintf("%v", q)
	}

	if query != "" {
		// Search for relevant documents
		docs, err := m.SearchSimilar(ctx, query, 5) // Default to top 5 documents
		if err != nil {
			m.logger.WithError(err).Error("Failed to search similar documents")
			return variables, nil // Return empty variables instead of error
		}

		// Add relevant documents to variables
		for _, key := range m.memoryKeys {
			switch key {
			case "relevant_documents":
				variables[key] = docs
			case "context":
				context := ""
				for _, doc := range docs {
					context += doc.Content + "\n\n"
				}
				variables[key] = context
			case "document_ids":
				ids := make([]string, len(docs))
				for i, doc := range docs {
					ids[i] = doc.ID
				}
				variables[key] = ids
			}
		}

		span.SetAttributes(
			attribute.String("query", query),
			attribute.Int("relevant_docs_count", len(docs)),
		)
	}

	return variables, nil
}

// SaveContext saves the context from input and output
func (m *DefaultVectorMemory) SaveContext(ctx context.Context, input map[string]interface{}, output map[string]interface{}) error {
	ctx, span := m.tracer.Start(ctx, "vector_memory.save_context")
	defer span.End()

	// Extract content to save as document
	var content string
	if userInput, exists := input["input"]; exists {
		content += fmt.Sprintf("Input: %v\n", userInput)
	}
	if assistantOutput, exists := output["text"]; exists {
		content += fmt.Sprintf("Output: %v", assistantOutput)
	}

	if content != "" {
		doc := &Document{
			ID:      uuid.New().String(),
			Content: content,
			Metadata: map[string]interface{}{
				"timestamp": ctx.Value("timestamp"),
				"source":    "conversation",
			},
		}

		if err := m.AddDocument(ctx, doc); err != nil {
			return fmt.Errorf("failed to add document to vector memory: %w", err)
		}

		span.SetAttributes(
			attribute.String("document.id", doc.ID),
			attribute.Int("document.content_length", len(doc.Content)),
		)
	}

	return nil
}

// Clear clears all memory
func (m *DefaultVectorMemory) Clear(ctx context.Context) error {
	ctx, span := m.tracer.Start(ctx, "vector_memory.clear")
	defer span.End()

	m.mu.Lock()
	defer m.mu.Unlock()

	m.documents = make(map[string]*Document)
	m.vectors = make(map[string][]float64)

	// Clear from vector store if available
	if m.vectorStore != nil {
		// Note: This would require a ClearAll method on VectorStore interface
		// For now, we'll just clear the local cache
	}

	return nil
}

// GetMemoryKeys returns the keys that this memory system provides
func (m *DefaultVectorMemory) GetMemoryKeys() []string {
	return m.memoryKeys
}

// GetMemoryType returns the type of this memory system
func (m *DefaultVectorMemory) GetMemoryType() string {
	return "vector"
}

// AddDocument adds a document to vector memory
func (m *DefaultVectorMemory) AddDocument(ctx context.Context, doc *Document) error {
	ctx, span := m.tracer.Start(ctx, "vector_memory.add_document")
	defer span.End()

	// Generate embedding if not provided
	if len(doc.Vector) == 0 {
		if m.embeddingLLM == nil {
			return fmt.Errorf("embedding LLM not configured and no vector provided")
		}

		embedding, err := m.generateEmbedding(ctx, doc.Content)
		if err != nil {
			return fmt.Errorf("failed to generate embedding: %w", err)
		}
		doc.Vector = embedding
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Add to local storage
	m.documents[doc.ID] = doc
	m.vectors[doc.ID] = doc.Vector

	// Trim if necessary
	m.trimDocuments()

	// Add to vector store if available
	if m.vectorStore != nil {
		vector := Vector{
			ID:       doc.ID,
			Vector:   doc.Vector,
			Metadata: doc.Metadata,
		}
		if err := m.vectorStore.AddVectors(ctx, []Vector{vector}); err != nil {
			return fmt.Errorf("failed to add vector to store: %w", err)
		}
	}

	span.SetAttributes(
		attribute.String("document.id", doc.ID),
		attribute.Int("vector.dimension", len(doc.Vector)),
	)

	return nil
}

// AddDocuments adds multiple documents to vector memory
func (m *DefaultVectorMemory) AddDocuments(ctx context.Context, docs []*Document) error {
	for _, doc := range docs {
		if err := m.AddDocument(ctx, doc); err != nil {
			return err
		}
	}
	return nil
}

// SearchSimilar searches for similar documents
func (m *DefaultVectorMemory) SearchSimilar(ctx context.Context, query string, limit int) ([]*Document, error) {
	ctx, span := m.tracer.Start(ctx, "vector_memory.search_similar")
	defer span.End()

	// Generate query embedding
	queryVector, err := m.generateEmbedding(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to generate query embedding: %w", err)
	}

	// Search in vector store if available
	if m.vectorStore != nil {
		results, err := m.vectorStore.SearchVectors(ctx, queryVector, limit)
		if err != nil {
			return nil, fmt.Errorf("failed to search vector store: %w", err)
		}

		docs := make([]*Document, len(results))
		for i, result := range results {
			doc, exists := m.documents[result.Vector.ID]
			if !exists {
				// Create document from vector store result
				doc = &Document{
					ID:       result.Vector.ID,
					Content:  "", // Content might not be stored in vector store
					Metadata: result.Vector.Metadata,
					Vector:   result.Vector.Vector,
				}
			}
			docs[i] = doc
		}

		span.SetAttributes(
			attribute.Int("search.results_count", len(docs)),
			attribute.Int("search.limit", limit),
		)

		return docs, nil
	}

	// Fallback to local search
	return m.searchLocal(ctx, queryVector, limit)
}

// SearchSimilarWithScores searches for similar documents with similarity scores
func (m *DefaultVectorMemory) SearchSimilarWithScores(ctx context.Context, query string, limit int) ([]*DocumentWithScore, error) {
	ctx, span := m.tracer.Start(ctx, "vector_memory.search_similar_with_scores")
	defer span.End()

	// Generate query embedding
	queryVector, err := m.generateEmbedding(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to generate query embedding: %w", err)
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	type docScore struct {
		doc   *Document
		score float64
	}

	var docScores []docScore
	for id, vector := range m.vectors {
		score := m.cosineSimilarity(queryVector, vector)
		if score >= m.threshold {
			doc := m.documents[id]
			docScores = append(docScores, docScore{doc: doc, score: score})
		}
	}

	// Sort by score descending
	sort.Slice(docScores, func(i, j int) bool {
		return docScores[i].score > docScores[j].score
	})

	// Limit results
	if limit > 0 && len(docScores) > limit {
		docScores = docScores[:limit]
	}

	// Convert to DocumentWithScore
	results := make([]*DocumentWithScore, len(docScores))
	for i, ds := range docScores {
		results[i] = &DocumentWithScore{
			Document: ds.doc,
			Score:    ds.score,
		}
	}

	span.SetAttributes(
		attribute.Int("search.results_count", len(results)),
		attribute.Float64("search.min_score", m.threshold),
	)

	return results, nil
}

// DeleteDocument deletes a document by ID
func (m *DefaultVectorMemory) DeleteDocument(ctx context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.documents, id)
	delete(m.vectors, id)

	// Delete from vector store if available
	if m.vectorStore != nil {
		if err := m.vectorStore.DeleteVector(ctx, id); err != nil {
			return fmt.Errorf("failed to delete vector from store: %w", err)
		}
	}

	return nil
}

// GetDocument retrieves a document by ID
func (m *DefaultVectorMemory) GetDocument(ctx context.Context, id string) (*Document, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	doc, exists := m.documents[id]
	if !exists {
		return nil, fmt.Errorf("document not found: %s", id)
	}

	return doc, nil
}

// GetDocumentCount returns the total number of documents
func (m *DefaultVectorMemory) GetDocumentCount(ctx context.Context) (int, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return len(m.documents), nil
}

// Helper methods

func (m *DefaultVectorMemory) generateEmbedding(ctx context.Context, text string) ([]float64, error) {
	if m.embeddingLLM == nil {
		return nil, fmt.Errorf("embedding LLM not configured")
	}

	req := &llm.EmbeddingRequest{
		Input: []string{text},
	}

	response, err := m.embeddingLLM.GetEmbeddings(ctx, req)
	if err != nil {
		return nil, err
	}

	if len(response.Embeddings) == 0 {
		return nil, fmt.Errorf("no embeddings returned")
	}

	return response.Embeddings[0], nil
}

func (m *DefaultVectorMemory) searchLocal(ctx context.Context, queryVector []float64, limit int) ([]*Document, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	type docScore struct {
		doc   *Document
		score float64
	}

	var docScores []docScore
	for id, vector := range m.vectors {
		score := m.cosineSimilarity(queryVector, vector)
		if score >= m.threshold {
			doc := m.documents[id]
			docScores = append(docScores, docScore{doc: doc, score: score})
		}
	}

	// Sort by score descending
	sort.Slice(docScores, func(i, j int) bool {
		return docScores[i].score > docScores[j].score
	})

	// Limit results
	if limit > 0 && len(docScores) > limit {
		docScores = docScores[:limit]
	}

	// Convert to documents
	docs := make([]*Document, len(docScores))
	for i, ds := range docScores {
		docs[i] = ds.doc
	}

	return docs, nil
}

func (m *DefaultVectorMemory) cosineSimilarity(a, b []float64) float64 {
	if len(a) != len(b) {
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

func (m *DefaultVectorMemory) trimDocuments() {
	if len(m.documents) <= m.maxDocuments {
		return
	}

	// Remove oldest documents (simple FIFO strategy)
	// In a real implementation, you might want more sophisticated strategies
	count := len(m.documents) - m.maxDocuments
	for id := range m.documents {
		if count <= 0 {
			break
		}
		delete(m.documents, id)
		delete(m.vectors, id)
		count--
	}
}
