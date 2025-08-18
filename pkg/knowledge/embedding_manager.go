package knowledge

import (
	"context"
	"fmt"
	"math"
	"strings"
	"sync"

	"github.com/aios/aios/pkg/vectordb"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// DefaultEmbeddingManager implements the EmbeddingManager interface
type DefaultEmbeddingManager struct {
	provider vectordb.EmbeddingProvider
	logger   *logrus.Logger
	tracer   trace.Tracer
	config   *EmbeddingManagerConfig
	cache    map[string][]float32
	mu       sync.RWMutex
}

// EmbeddingManagerConfig represents configuration for the embedding manager
type EmbeddingManagerConfig struct {
	Provider    string `json:"provider"`   // "openai", "ollama", etc.
	Model       string `json:"model"`      // embedding model name
	Dimensions  int    `json:"dimensions"` // embedding dimensions
	BatchSize   int    `json:"batch_size"` // batch size for processing
	CacheSize   int    `json:"cache_size"` // max cache entries
	EnableCache bool   `json:"enable_cache"`
}

// NewDefaultEmbeddingManager creates a new default embedding manager
func NewDefaultEmbeddingManager(model string, logger *logrus.Logger) (EmbeddingManager, error) {
	config := &EmbeddingManagerConfig{
		Provider:    "openai",
		Model:       model,
		Dimensions:  1536, // Default for OpenAI ada-002
		BatchSize:   100,
		CacheSize:   10000,
		EnableCache: true,
	}

	manager := &DefaultEmbeddingManager{
		logger: logger,
		tracer: otel.Tracer("knowledge.embedding_manager"),
		config: config,
		cache:  make(map[string][]float32),
	}

	// Initialize embedding provider
	if err := manager.initializeProvider(); err != nil {
		return nil, fmt.Errorf("failed to initialize embedding provider: %w", err)
	}

	return manager, nil
}

// initializeProvider initializes the embedding provider
func (em *DefaultEmbeddingManager) initializeProvider() error {
	embeddingManager := vectordb.NewEmbeddingManager(em.logger)
	
	var config *vectordb.EmbeddingConfig
	switch em.config.Provider {
	case "openai":
		config = &vectordb.EmbeddingConfig{
			Provider:   "openai",
			Model:      em.config.Model,
			Dimensions: em.config.Dimensions,
			BatchSize:  em.config.BatchSize,
		}
	case "ollama":
		config = &vectordb.EmbeddingConfig{
			Provider:   "ollama",
			Model:      em.config.Model,
			Dimensions: em.config.Dimensions,
			BaseURL:    "http://localhost:11434",
		}
	default:
		return fmt.Errorf("unsupported embedding provider: %s", em.config.Provider)
	}

	provider, err := embeddingManager.CreateProvider(config)
	if err != nil {
		return fmt.Errorf("failed to create embedding provider: %w", err)
	}
	
	em.provider = provider
	return nil
}

// GenerateEmbedding generates an embedding for a single text
func (em *DefaultEmbeddingManager) GenerateEmbedding(ctx context.Context, text string) ([]float32, error) {
	ctx, span := em.tracer.Start(ctx, "embedding_manager.generate_embedding")
	defer span.End()

	span.SetAttributes(
		attribute.String("embedding.model", em.config.Model),
		attribute.Int("text.length", len(text)),
	)

	// Check cache first
	if em.config.EnableCache {
		if cached := em.getCachedEmbedding(text); cached != nil {
			span.SetAttributes(attribute.Bool("cache.hit", true))
			return cached, nil
		}
	}

	// Generate embedding
	embedding, err := em.provider.GenerateEmbedding(ctx, text)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to generate embedding: %w", err)
	}

	// Cache the result
	if em.config.EnableCache {
		em.cacheEmbedding(text, embedding)
	}

	span.SetAttributes(
		attribute.Bool("cache.hit", false),
		attribute.Int("embedding.dimensions", len(embedding)),
	)

	return embedding, nil
}

// GenerateEmbeddings generates embeddings for multiple texts
func (em *DefaultEmbeddingManager) GenerateEmbeddings(ctx context.Context, texts []string) ([][]float32, error) {
	ctx, span := em.tracer.Start(ctx, "embedding_manager.generate_embeddings")
	defer span.End()

	span.SetAttributes(
		attribute.String("embedding.model", em.config.Model),
		attribute.Int("texts.count", len(texts)),
	)

	if len(texts) == 0 {
		return [][]float32{}, nil
	}

	var embeddings [][]float32
	var uncachedTexts []string
	var uncachedIndices []int

	// Check cache for each text
	if em.config.EnableCache {
		embeddings = make([][]float32, len(texts))
		for i, text := range texts {
			if cached := em.getCachedEmbedding(text); cached != nil {
				embeddings[i] = cached
			} else {
				uncachedTexts = append(uncachedTexts, text)
				uncachedIndices = append(uncachedIndices, i)
			}
		}
	} else {
		uncachedTexts = texts
		for i := range texts {
			uncachedIndices = append(uncachedIndices, i)
		}
		embeddings = make([][]float32, len(texts))
	}

	// Generate embeddings for uncached texts
	if len(uncachedTexts) > 0 {
		newEmbeddings, err := em.provider.GenerateEmbeddings(ctx, uncachedTexts)
		if err != nil {
			span.RecordError(err)
			return nil, fmt.Errorf("failed to generate embeddings: %w", err)
		}

		// Place new embeddings in correct positions and cache them
		for i, embedding := range newEmbeddings {
			originalIndex := uncachedIndices[i]
			embeddings[originalIndex] = embedding

			// Cache the result
			if em.config.EnableCache {
				em.cacheEmbedding(uncachedTexts[i], embedding)
			}
		}
	}

	span.SetAttributes(
		attribute.Int("cache.hits", len(texts)-len(uncachedTexts)),
		attribute.Int("cache.misses", len(uncachedTexts)),
	)

	return embeddings, nil
}

// GetEmbeddingDimensions returns the embedding dimensions
func (em *DefaultEmbeddingManager) GetEmbeddingDimensions() int {
	return em.config.Dimensions
}

// GetEmbeddingModel returns the embedding model name
func (em *DefaultEmbeddingManager) GetEmbeddingModel() string {
	return em.config.Model
}

// CompareEmbeddings compares two embeddings using cosine similarity
func (em *DefaultEmbeddingManager) CompareEmbeddings(embedding1, embedding2 []float32) float32 {
	if len(embedding1) != len(embedding2) {
		return 0.0
	}

	return cosineSimilarity(embedding1, embedding2)
}

// getCachedEmbedding retrieves an embedding from cache
func (em *DefaultEmbeddingManager) getCachedEmbedding(text string) []float32 {
	em.mu.RLock()
	defer em.mu.RUnlock()

	if embedding, exists := em.cache[text]; exists {
		// Return a copy to avoid modification
		result := make([]float32, len(embedding))
		copy(result, embedding)
		return result
	}

	return nil
}

// cacheEmbedding stores an embedding in cache
func (em *DefaultEmbeddingManager) cacheEmbedding(text string, embedding []float32) {
	em.mu.Lock()
	defer em.mu.Unlock()

	// Check cache size limit
	if len(em.cache) >= em.config.CacheSize {
		// Simple eviction: remove first entry (FIFO)
		for key := range em.cache {
			delete(em.cache, key)
			break
		}
	}

	// Store a copy to avoid modification
	cached := make([]float32, len(embedding))
	copy(cached, embedding)
	em.cache[text] = cached
}

// cosineSimilarity calculates cosine similarity between two vectors
func cosineSimilarity(a, b []float32) float32 {
	if len(a) != len(b) {
		return 0.0
	}

	var dotProduct, normA, normB float64

	for i := 0; i < len(a); i++ {
		dotProduct += float64(a[i]) * float64(b[i])
		normA += float64(a[i]) * float64(a[i])
		normB += float64(b[i]) * float64(b[i])
	}

	if normA == 0.0 || normB == 0.0 {
		return 0.0
	}

	return float32(dotProduct / (math.Sqrt(normA) * math.Sqrt(normB)))
}

// DefaultDocumentRetriever implements the DocumentRetriever interface
type DefaultDocumentRetriever struct {
	embeddingManager EmbeddingManager
	vectorStore      vectordb.VectorStore
	documents        map[string]*Document
	logger           *logrus.Logger
	tracer           trace.Tracer
	mu               sync.RWMutex
}

// NewDefaultDocumentRetriever creates a new default document retriever
func NewDefaultDocumentRetriever(embeddingManager EmbeddingManager, logger *logrus.Logger) (DocumentRetriever, error) {
	retriever := &DefaultDocumentRetriever{
		embeddingManager: embeddingManager,
		documents:        make(map[string]*Document),
		logger:           logger,
		tracer:           otel.Tracer("knowledge.document_retriever"),
	}

	return retriever, nil
}

// Retrieve retrieves documents based on a query
func (dr *DefaultDocumentRetriever) Retrieve(ctx context.Context, query string, options *RetrievalOptions) ([]*Document, error) {
	ctx, span := dr.tracer.Start(ctx, "document_retriever.retrieve")
	defer span.End()

	span.SetAttributes(
		attribute.String("query", query),
		attribute.Int("top_k", options.TopK),
	)

	// Generate query embedding
	queryEmbedding, err := dr.embeddingManager.GenerateEmbedding(ctx, query)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to generate query embedding: %w", err)
	}

	// Retrieve by embedding
	return dr.RetrieveByEmbedding(ctx, queryEmbedding, options)
}

// RetrieveByEmbedding retrieves documents based on an embedding
func (dr *DefaultDocumentRetriever) RetrieveByEmbedding(ctx context.Context, embedding []float32, options *RetrievalOptions) ([]*Document, error) {
	ctx, span := dr.tracer.Start(ctx, "document_retriever.retrieve_by_embedding")
	defer span.End()

	dr.mu.RLock()
	defer dr.mu.RUnlock()

	// Calculate similarities with all documents
	type docSimilarity struct {
		document   *Document
		similarity float32
	}

	var similarities []docSimilarity

	for _, doc := range dr.documents {
		if len(doc.Embedding) == 0 {
			continue
		}

		similarity := dr.embeddingManager.CompareEmbeddings(embedding, doc.Embedding)
		if similarity >= options.Threshold {
			similarities = append(similarities, docSimilarity{
				document:   doc,
				similarity: similarity,
			})
		}
	}

	// Sort by similarity (descending)
	for i := 0; i < len(similarities)-1; i++ {
		for j := i + 1; j < len(similarities); j++ {
			if similarities[i].similarity < similarities[j].similarity {
				similarities[i], similarities[j] = similarities[j], similarities[i]
			}
		}
	}

	// Return top K documents
	var results []*Document
	limit := options.TopK
	if limit > len(similarities) {
		limit = len(similarities)
	}

	for i := 0; i < limit; i++ {
		results = append(results, similarities[i].document)
	}

	span.SetAttributes(
		attribute.Int("candidates", len(similarities)),
		attribute.Int("results", len(results)),
	)

	return results, nil
}

// HybridRetrieve performs hybrid retrieval (semantic + keyword)
func (dr *DefaultDocumentRetriever) HybridRetrieve(ctx context.Context, query string, options *RetrievalOptions) ([]*Document, error) {
	ctx, span := dr.tracer.Start(ctx, "document_retriever.hybrid_retrieve")
	defer span.End()

	// For now, fall back to semantic retrieval
	// In a real implementation, this would combine semantic and keyword search
	return dr.Retrieve(ctx, query, options)
}

// DefaultDocumentReranker implements the DocumentReranker interface
type DefaultDocumentReranker struct {
	embeddingManager EmbeddingManager
	logger           *logrus.Logger
	tracer           trace.Tracer
}

// NewDefaultDocumentReranker creates a new default document reranker
func NewDefaultDocumentReranker(embeddingManager EmbeddingManager, logger *logrus.Logger) (DocumentReranker, error) {
	return &DefaultDocumentReranker{
		embeddingManager: embeddingManager,
		logger:           logger,
		tracer:           otel.Tracer("knowledge.document_reranker"),
	}, nil
}

// Rerank reranks documents based on relevance to the query
func (dr *DefaultDocumentReranker) Rerank(ctx context.Context, query string, documents []*Document) ([]*Document, error) {
	ctx, span := dr.tracer.Start(ctx, "document_reranker.rerank")
	defer span.End()

	if len(documents) <= 1 {
		return documents, nil
	}

	// Generate query embedding
	queryEmbedding, err := dr.embeddingManager.GenerateEmbedding(ctx, query)
	if err != nil {
		span.RecordError(err)
		return documents, err // Return original order on error
	}

	// Calculate relevance scores
	type docScore struct {
		document *Document
		score    float32
	}

	var scores []docScore
	for _, doc := range documents {
		score, err := dr.CalculateRelevanceScore(ctx, query, doc)
		if err != nil {
			// Use similarity as fallback
			if len(doc.Embedding) > 0 {
				score = dr.embeddingManager.CompareEmbeddings(queryEmbedding, doc.Embedding)
			} else {
				score = 0.0
			}
		}

		scores = append(scores, docScore{
			document: doc,
			score:    score,
		})
	}

	// Sort by score (descending)
	for i := 0; i < len(scores)-1; i++ {
		for j := i + 1; j < len(scores); j++ {
			if scores[i].score < scores[j].score {
				scores[i], scores[j] = scores[j], scores[i]
			}
		}
	}

	// Extract reranked documents
	var reranked []*Document
	for _, score := range scores {
		reranked = append(reranked, score.document)
	}

	return reranked, nil
}

// CalculateRelevanceScore calculates relevance score for a document
func (dr *DefaultDocumentReranker) CalculateRelevanceScore(ctx context.Context, query string, document *Document) (float32, error) {
	// Simple relevance scoring based on multiple factors
	var score float32

	// Factor 1: Semantic similarity (if embeddings available)
	if len(document.Embedding) > 0 {
		queryEmbedding, err := dr.embeddingManager.GenerateEmbedding(ctx, query)
		if err == nil {
			semanticScore := dr.embeddingManager.CompareEmbeddings(queryEmbedding, document.Embedding)
			score += semanticScore * 0.7 // 70% weight for semantic similarity
		}
	}

	// Factor 2: Keyword matching
	keywordScore := calculateKeywordScore(query, document.Content)
	score += keywordScore * 0.2 // 20% weight for keyword matching

	// Factor 3: Title relevance
	titleScore := calculateKeywordScore(query, document.Title)
	score += titleScore * 0.1 // 10% weight for title relevance

	return score, nil
}

// calculateKeywordScore calculates keyword matching score
func calculateKeywordScore(query, text string) float32 {
	queryWords := strings.Fields(strings.ToLower(query))
	textLower := strings.ToLower(text)

	matches := 0
	for _, word := range queryWords {
		if strings.Contains(textLower, word) {
			matches++
		}
	}

	if len(queryWords) == 0 {
		return 0.0
	}

	return float32(matches) / float32(len(queryWords))
}
