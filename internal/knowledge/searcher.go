package knowledge

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/aios/aios/pkg/config"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

// VectorSearcher handles vector-based search operations
type VectorSearcher struct {
	config     *config.Config
	logger     *logrus.Logger
	tracer     trace.Tracer
	repository *Repository
	embeddings *EmbeddingService
	index      *VectorIndex
}

// EmbeddingService handles text embeddings
type EmbeddingService struct {
	logger *logrus.Logger
}

// VectorIndex stores and searches vectors
type VectorIndex struct {
	vectors  map[string][]float64
	texts    map[string]string
	metadata map[string]map[string]string
}

// SearchMatch represents a search result match
type SearchMatch struct {
	ID       string            `json:"id"`
	Text     string            `json:"text"`
	Score    float64           `json:"score"`
	Metadata map[string]string `json:"metadata"`
}

// NewVectorSearcher creates a new vector searcher instance
func NewVectorSearcher(config *config.Config, repository *Repository, logger *logrus.Logger) (*VectorSearcher, error) {
	embeddings := &EmbeddingService{
		logger: logger,
	}

	index := &VectorIndex{
		vectors:  make(map[string][]float64),
		texts:    make(map[string]string),
		metadata: make(map[string]map[string]string),
	}

	return &VectorSearcher{
		config:     config,
		logger:     logger,
		tracer:     otel.Tracer("knowledge.searcher"),
		repository: repository,
		embeddings: embeddings,
		index:      index,
	}, nil
}

// Start starts the vector searcher
func (s *VectorSearcher) Start(ctx context.Context) error {
	s.logger.Info("Starting Vector Searcher...")
	return nil
}

// Stop stops the vector searcher
func (s *VectorSearcher) Stop(ctx context.Context) error {
	s.logger.Info("Stopping Vector Searcher...")
	return nil
}

// Search performs a vector-based search
func (s *VectorSearcher) Search(ctx context.Context, req *SearchRequest) (*SearchResponse, error) {
	ctx, span := s.tracer.Start(ctx, "searcher.search")
	defer span.End()

	// Set defaults
	if req.MaxResults == 0 {
		req.MaxResults = 10
	}

	var results []SearchResult

	if req.UseRAG {
		// Perform vector-based RAG search
		ragResults, err := s.performRAGSearch(ctx, req)
		if err != nil {
			return nil, fmt.Errorf("RAG search failed: %w", err)
		}
		results = ragResults
	} else {
		// Perform keyword-based search
		keywordResults, err := s.performKeywordSearch(ctx, req)
		if err != nil {
			return nil, fmt.Errorf("keyword search failed: %w", err)
		}
		results = keywordResults
	}

	// Apply filters
	if len(req.Filters) > 0 {
		results = s.applyFilters(results, req.Filters)
	}

	// Limit results
	if len(results) > req.MaxResults {
		results = results[:req.MaxResults]
	}

	response := &SearchResponse{
		Results: results,
		Total:   len(results),
		Query:   req.Query,
	}

	s.logger.WithFields(logrus.Fields{
		"query":       req.Query,
		"results":     len(results),
		"use_rag":     req.UseRAG,
		"max_results": req.MaxResults,
	}).Info("Search completed")

	return response, nil
}

// performRAGSearch performs retrieval-augmented generation search
func (s *VectorSearcher) performRAGSearch(ctx context.Context, req *SearchRequest) ([]SearchResult, error) {
	// Generate embedding for query
	queryVector, err := s.embeddings.GenerateEmbedding(req.Query)
	if err != nil {
		return nil, fmt.Errorf("failed to generate query embedding: %w", err)
	}

	// Find similar vectors
	matches := s.index.FindSimilar(queryVector, req.MaxResults*2) // Get more candidates

	// Convert to search results
	var results []SearchResult
	for _, match := range matches {
		result := SearchResult{
			ID:       match.ID,
			Title:    s.extractTitle(match.Text),
			Content:  match.Text,
			Score:    match.Score,
			Metadata: match.Metadata,
		}

		// Add URL if available in metadata
		if url, exists := match.Metadata["url"]; exists {
			result.URL = url
		}

		results = append(results, result)
	}

	return results, nil
}

// performKeywordSearch performs keyword-based search
func (s *VectorSearcher) performKeywordSearch(ctx context.Context, req *SearchRequest) ([]SearchResult, error) {
	query := strings.ToLower(req.Query)
	keywords := strings.Fields(query)

	var results []SearchResult

	// Search through all indexed texts
	for id, text := range s.index.texts {
		score := s.calculateKeywordScore(text, keywords)
		if score > 0 {
			result := SearchResult{
				ID:       id,
				Title:    s.extractTitle(text),
				Content:  text,
				Score:    score,
				Metadata: s.index.metadata[id],
			}

			// Add URL if available in metadata
			if url, exists := result.Metadata["url"]; exists {
				result.URL = url
			}

			results = append(results, result)
		}
	}

	// Sort by score (descending)
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	return results, nil
}

// calculateKeywordScore calculates a simple keyword-based score
func (s *VectorSearcher) calculateKeywordScore(text string, keywords []string) float64 {
	if len(keywords) == 0 {
		return 0
	}

	textLower := strings.ToLower(text)
	score := 0.0
	totalKeywords := float64(len(keywords))

	for _, keyword := range keywords {
		if strings.Contains(textLower, keyword) {
			// Count occurrences
			count := float64(strings.Count(textLower, keyword))
			// Weight by keyword length (longer keywords are more specific)
			weight := float64(len(keyword)) / 10.0
			if weight < 0.1 {
				weight = 0.1
			}
			score += count * weight
		}
	}

	// Normalize by number of keywords
	return score / totalKeywords
}

// applyFilters applies filters to search results
func (s *VectorSearcher) applyFilters(results []SearchResult, filters map[string]string) []SearchResult {
	var filtered []SearchResult

	for _, result := range results {
		match := true
		for key, value := range filters {
			if metaValue, exists := result.Metadata[key]; !exists || metaValue != value {
				match = false
				break
			}
		}
		if match {
			filtered = append(filtered, result)
		}
	}

	return filtered
}

// extractTitle extracts a title from text content
func (s *VectorSearcher) extractTitle(text string) string {
	lines := strings.Split(text, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if len(line) > 0 && len(line) < 100 {
			return line
		}
	}

	// If no suitable title found, use first 50 characters
	if len(text) > 50 {
		return text[:50] + "..."
	}
	return text
}

// IndexDocument adds a document to the search index
func (s *VectorSearcher) IndexDocument(ctx context.Context, doc *Document) error {
	ctx, span := s.tracer.Start(ctx, "searcher.index_document")
	defer span.End()

	// Convert metadata to string map
	docMetadata := make(map[string]string)
	for k, v := range doc.Metadata {
		if str, ok := v.(string); ok {
			docMetadata[k] = str
		}
	}

	// Index the full document
	if err := s.indexText(doc.ID.String(), doc.Content, docMetadata); err != nil {
		return fmt.Errorf("failed to index document: %w", err)
	}

	// Get and index document chunks
	chunks, err := s.repository.GetDocumentChunks(ctx, doc.ID)
	if err != nil {
		s.logger.WithError(err).Warn("Failed to get document chunks for indexing")
		return nil // Don't fail the whole operation
	}

	// Index individual chunks
	for _, chunk := range chunks {
		chunkMetadata := make(map[string]string)
		for k, v := range doc.Metadata {
			if str, ok := v.(string); ok {
				chunkMetadata[k] = str
			}
		}
		for k, v := range chunk.Metadata {
			if str, ok := v.(string); ok {
				chunkMetadata[k] = str
			}
		}

		if err := s.indexText(chunk.ID.String(), chunk.Content, chunkMetadata); err != nil {
			s.logger.WithError(err).WithField("chunk_id", chunk.ID).Warn("Failed to index chunk")
		}
	}

	s.logger.WithFields(logrus.Fields{
		"document_id": doc.ID.String(),
		"chunks":      len(chunks),
	}).Info("Document indexed successfully")

	return nil
}

// indexText adds text to the search index
func (s *VectorSearcher) indexText(id, text string, metadata map[string]string) error {
	// Generate embedding
	vector, err := s.embeddings.GenerateEmbedding(text)
	if err != nil {
		return fmt.Errorf("failed to generate embedding: %w", err)
	}

	// Store in index
	s.index.vectors[id] = vector
	s.index.texts[id] = text
	s.index.metadata[id] = metadata

	return nil
}

// GenerateEmbedding generates a simple embedding for text
func (e *EmbeddingService) GenerateEmbedding(text string) ([]float64, error) {
	// This is a simple mock embedding generator
	// In a real implementation, you'd use a proper embedding model

	words := strings.Fields(strings.ToLower(text))
	embedding := make([]float64, 384) // Common embedding dimension

	// Simple hash-based embedding
	for i, word := range words {
		if i >= len(embedding) {
			break
		}

		// Simple hash function
		hash := 0
		for _, char := range word {
			hash = hash*31 + int(char)
		}

		// Normalize to [-1, 1]
		embedding[i%len(embedding)] += float64(hash%1000)/500.0 - 1.0
	}

	// Normalize vector
	norm := 0.0
	for _, val := range embedding {
		norm += val * val
	}
	norm = math.Sqrt(norm)

	if norm > 0 {
		for i := range embedding {
			embedding[i] /= norm
		}
	}

	return embedding, nil
}

// FindSimilar finds vectors similar to the query vector
func (vi *VectorIndex) FindSimilar(queryVector []float64, maxResults int) []SearchMatch {
	var matches []SearchMatch

	for id, vector := range vi.vectors {
		similarity := cosineSimilarity(queryVector, vector)
		if similarity > 0.1 { // Minimum similarity threshold
			match := SearchMatch{
				ID:       id,
				Text:     vi.texts[id],
				Score:    similarity,
				Metadata: vi.metadata[id],
			}
			matches = append(matches, match)
		}
	}

	// Sort by similarity (descending)
	sort.Slice(matches, func(i, j int) bool {
		return matches[i].Score > matches[j].Score
	})

	// Limit results
	if len(matches) > maxResults {
		matches = matches[:maxResults]
	}

	return matches
}

// cosineSimilarity calculates cosine similarity between two vectors
func cosineSimilarity(a, b []float64) float64 {
	if len(a) != len(b) {
		return 0
	}

	dotProduct := 0.0
	normA := 0.0
	normB := 0.0

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
