package vectordb

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// DefaultVectorStore implements VectorStore interface
type DefaultVectorStore struct {
	vectorDB  VectorDB
	embedding EmbeddingProvider
	config    *VectorStoreConfig
	logger    *logrus.Logger
	tracer    trace.Tracer
}

// NewVectorStore creates a new vector store
func NewVectorStore(vectorDB VectorDB, embedding EmbeddingProvider, config *VectorStoreConfig, logger *logrus.Logger) VectorStore {
	if config == nil {
		config = &VectorStoreConfig{
			ChunkSize: 1000,
			Overlap:   200,
		}
	}

	return &DefaultVectorStore{
		vectorDB:  vectorDB,
		embedding: embedding,
		config:    config,
		logger:    logger,
		tracer:    otel.Tracer("vectordb.store"),
	}
}

// AddDocuments adds documents with automatic embedding generation
func (s *DefaultVectorStore) AddDocuments(ctx context.Context, collection string, documents []*Document) error {
	ctx, span := s.tracer.Start(ctx, "vector_store.add_documents")
	defer span.End()

	span.SetAttributes(
		attribute.String("collection.name", collection),
		attribute.Int("documents.count", len(documents)),
	)

	if len(documents) == 0 {
		return nil
	}

	// Extract texts for embedding generation
	texts := make([]string, len(documents))
	for i, doc := range documents {
		texts[i] = doc.Content
	}

	// Generate embeddings
	embeddings, err := s.embedding.GenerateEmbeddings(ctx, texts)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to generate embeddings: %w", err)
	}

	// Create vectors
	vectors := make([]*Vector, len(documents))
	for i, doc := range documents {
		id := doc.ID
		if id == "" {
			id = uuid.New().String()
		}

		metadata := make(map[string]interface{})
		if doc.Metadata != nil {
			for k, v := range doc.Metadata {
				metadata[k] = v
			}
		}
		metadata["content"] = doc.Content
		metadata["added_at"] = time.Now().Unix()

		vectors[i] = &Vector{
			ID:       id,
			Values:   embeddings[i],
			Metadata: metadata,
		}
	}

	// Insert vectors
	if err := s.vectorDB.Insert(ctx, collection, vectors); err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to insert vectors: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"collection": collection,
		"count":      len(documents),
	}).Info("Added documents to vector store")

	return nil
}

// AddTexts adds texts with automatic embedding generation
func (s *DefaultVectorStore) AddTexts(ctx context.Context, collection string, texts []string, metadatas []map[string]interface{}) error {
	ctx, span := s.tracer.Start(ctx, "vector_store.add_texts")
	defer span.End()

	span.SetAttributes(
		attribute.String("collection.name", collection),
		attribute.Int("texts.count", len(texts)),
	)

	if len(texts) == 0 {
		return nil
	}

	// Ensure metadatas slice has the same length as texts
	if metadatas == nil {
		metadatas = make([]map[string]interface{}, len(texts))
		for i := range metadatas {
			metadatas[i] = make(map[string]interface{})
		}
	} else if len(metadatas) < len(texts) {
		// Extend metadatas slice
		for i := len(metadatas); i < len(texts); i++ {
			metadatas = append(metadatas, make(map[string]interface{}))
		}
	}

	// Create documents
	documents := make([]*Document, len(texts))
	for i, text := range texts {
		documents[i] = &Document{
			ID:       uuid.New().String(),
			Content:  text,
			Metadata: metadatas[i],
		}
	}

	return s.AddDocuments(ctx, collection, documents)
}

// SimilaritySearch performs similarity search with text query
func (s *DefaultVectorStore) SimilaritySearch(ctx context.Context, collection string, query string, topK int, filter map[string]interface{}) ([]*Document, error) {
	results, err := s.SimilaritySearchWithScore(ctx, collection, query, topK, filter)
	if err != nil {
		return nil, err
	}

	documents := make([]*Document, len(results))
	for i, result := range results {
		documents[i] = result.Document
	}

	return documents, nil
}

// SimilaritySearchWithScore performs similarity search with scores
func (s *DefaultVectorStore) SimilaritySearchWithScore(ctx context.Context, collection string, query string, topK int, filter map[string]interface{}) ([]*DocumentWithScore, error) {
	ctx, span := s.tracer.Start(ctx, "vector_store.similarity_search_with_score")
	defer span.End()

	span.SetAttributes(
		attribute.String("collection.name", collection),
		attribute.String("query", query),
		attribute.Int("top_k", topK),
	)

	// Generate embedding for query
	queryEmbedding, err := s.embedding.GenerateEmbedding(ctx, query)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to generate query embedding: %w", err)
	}

	// Perform search
	searchRequest := &SearchRequest{
		Collection: collection,
		Vector:     queryEmbedding,
		TopK:       topK,
		Filter:     filter,
		Include:    []string{"metadata"},
	}

	searchResult, err := s.vectorDB.Search(ctx, searchRequest)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to search vectors: %w", err)
	}

	// Convert matches to documents with scores
	results := make([]*DocumentWithScore, len(searchResult.Matches))
	for i, match := range searchResult.Matches {
		content := ""
		if contentValue, exists := match.Metadata["content"]; exists {
			if contentStr, ok := contentValue.(string); ok {
				content = contentStr
			}
		}

		// Remove content from metadata to avoid duplication
		metadata := make(map[string]interface{})
		for k, v := range match.Metadata {
			if k != "content" {
				metadata[k] = v
			}
		}

		document := &Document{
			ID:       match.ID,
			Content:  content,
			Metadata: metadata,
		}

		results[i] = &DocumentWithScore{
			Document: document,
			Score:    match.Score,
		}
	}

	s.logger.WithFields(logrus.Fields{
		"collection": collection,
		"query":      query,
		"results":    len(results),
	}).Debug("Performed similarity search")

	return results, nil
}

// MaxMarginalRelevanceSearch performs MMR search for diverse results
func (s *DefaultVectorStore) MaxMarginalRelevanceSearch(ctx context.Context, collection string, query string, topK int, fetchK int, lambda float32, filter map[string]interface{}) ([]*Document, error) {
	ctx, span := s.tracer.Start(ctx, "vector_store.mmr_search")
	defer span.End()

	span.SetAttributes(
		attribute.String("collection.name", collection),
		attribute.String("query", query),
		attribute.Int("top_k", topK),
		attribute.Int("fetch_k", fetchK),
		attribute.Float64("lambda", float64(lambda)),
	)

	if fetchK <= 0 {
		fetchK = topK * 2 // Default fetch more than needed
	}

	// Get initial candidates
	candidates, err := s.SimilaritySearchWithScore(ctx, collection, query, fetchK, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get candidates: %w", err)
	}

	if len(candidates) == 0 {
		return []*Document{}, nil
	}

	// Generate query embedding
	queryEmbedding, err := s.embedding.GenerateEmbedding(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to generate query embedding: %w", err)
	}

	// Get embeddings for all candidates
	candidateTexts := make([]string, len(candidates))
	for i, candidate := range candidates {
		candidateTexts[i] = candidate.Document.Content
	}

	candidateEmbeddings, err := s.embedding.GenerateEmbeddings(ctx, candidateTexts)
	if err != nil {
		return nil, fmt.Errorf("failed to generate candidate embeddings: %w", err)
	}

	// Perform MMR selection
	selected := make([]*Document, 0, topK)
	selectedIndices := make(map[int]bool)

	for len(selected) < topK && len(selected) < len(candidates) {
		bestScore := float32(-1)
		bestIndex := -1

		for i := range candidates {
			if selectedIndices[i] {
				continue
			}

			// Calculate relevance score (similarity to query)
			relevanceScore := cosineSimilarity(queryEmbedding, candidateEmbeddings[i])

			// Calculate diversity score (maximum similarity to already selected)
			diversityScore := float32(0)
			for selectedIdx := range selectedIndices {
				similarity := cosineSimilarity(candidateEmbeddings[i], candidateEmbeddings[selectedIdx])
				if similarity > diversityScore {
					diversityScore = similarity
				}
			}

			// MMR score: lambda * relevance - (1 - lambda) * diversity
			mmrScore := lambda*relevanceScore - (1-lambda)*diversityScore

			if mmrScore > bestScore {
				bestScore = mmrScore
				bestIndex = i
			}
		}

		if bestIndex >= 0 {
			selected = append(selected, candidates[bestIndex].Document)
			selectedIndices[bestIndex] = true
		} else {
			break
		}
	}

	s.logger.WithFields(logrus.Fields{
		"collection": collection,
		"query":      query,
		"selected":   len(selected),
		"candidates": len(candidates),
	}).Debug("Performed MMR search")

	return selected, nil
}

// GetDocuments retrieves documents by IDs
func (s *DefaultVectorStore) GetDocuments(ctx context.Context, collection string, ids []string) ([]*Document, error) {
	ctx, span := s.tracer.Start(ctx, "vector_store.get_documents")
	defer span.End()

	span.SetAttributes(
		attribute.String("collection.name", collection),
		attribute.Int("ids.count", len(ids)),
	)

	vectors, err := s.vectorDB.GetVectors(ctx, collection, ids)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to get vectors: %w", err)
	}

	documents := make([]*Document, len(vectors))
	for i, vector := range vectors {
		content := ""
		if contentValue, exists := vector.Metadata["content"]; exists {
			if contentStr, ok := contentValue.(string); ok {
				content = contentStr
			}
		}

		// Remove content from metadata
		metadata := make(map[string]interface{})
		for k, v := range vector.Metadata {
			if k != "content" {
				metadata[k] = v
			}
		}

		documents[i] = &Document{
			ID:       vector.ID,
			Content:  content,
			Metadata: metadata,
		}
	}

	return documents, nil
}

// DeleteDocuments deletes documents by IDs
func (s *DefaultVectorStore) DeleteDocuments(ctx context.Context, collection string, ids []string) error {
	ctx, span := s.tracer.Start(ctx, "vector_store.delete_documents")
	defer span.End()

	span.SetAttributes(
		attribute.String("collection.name", collection),
		attribute.Int("ids.count", len(ids)),
	)

	if err := s.vectorDB.Delete(ctx, collection, ids); err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to delete vectors: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"collection": collection,
		"count":      len(ids),
	}).Info("Deleted documents from vector store")

	return nil
}

// UpdateDocuments updates existing documents
func (s *DefaultVectorStore) UpdateDocuments(ctx context.Context, collection string, documents []*Document) error {
	ctx, span := s.tracer.Start(ctx, "vector_store.update_documents")
	defer span.End()

	span.SetAttributes(
		attribute.String("collection.name", collection),
		attribute.Int("documents.count", len(documents)),
	)

	if len(documents) == 0 {
		return nil
	}

	// Extract texts for embedding generation
	texts := make([]string, len(documents))
	for i, doc := range documents {
		texts[i] = doc.Content
	}

	// Generate embeddings
	embeddings, err := s.embedding.GenerateEmbeddings(ctx, texts)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to generate embeddings: %w", err)
	}

	// Create vectors
	vectors := make([]*Vector, len(documents))
	for i, doc := range documents {
		metadata := make(map[string]interface{})
		if doc.Metadata != nil {
			for k, v := range doc.Metadata {
				metadata[k] = v
			}
		}
		metadata["content"] = doc.Content
		metadata["updated_at"] = time.Now().Unix()

		vectors[i] = &Vector{
			ID:       doc.ID,
			Values:   embeddings[i],
			Metadata: metadata,
		}
	}

	// Update vectors
	if err := s.vectorDB.Update(ctx, collection, vectors); err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to update vectors: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"collection": collection,
		"count":      len(documents),
	}).Info("Updated documents in vector store")

	return nil
}

// Helper functions

// cosineSimilarity calculates cosine similarity between two vectors
func cosineSimilarity(a, b []float32) float32 {
	if len(a) != len(b) {
		return 0
	}

	var dotProduct, normA, normB float32
	for i := 0; i < len(a); i++ {
		dotProduct += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}

	if normA == 0 || normB == 0 {
		return 0
	}

	return dotProduct / (float32(math.Sqrt(float64(normA))) * float32(math.Sqrt(float64(normB))))
}

// TextSplitter splits text into chunks
type TextSplitter struct {
	ChunkSize int
	Overlap   int
}

// NewTextSplitter creates a new text splitter
func NewTextSplitter(chunkSize, overlap int) *TextSplitter {
	return &TextSplitter{
		ChunkSize: chunkSize,
		Overlap:   overlap,
	}
}

// SplitText splits text into chunks
func (ts *TextSplitter) SplitText(text string) []string {
	if len(text) <= ts.ChunkSize {
		return []string{text}
	}

	var chunks []string
	start := 0

	for start < len(text) {
		end := start + ts.ChunkSize
		if end > len(text) {
			end = len(text)
		}

		// Try to break at word boundary
		if end < len(text) {
			for i := end; i > start && i > end-100; i-- {
				if text[i] == ' ' || text[i] == '\n' || text[i] == '\t' {
					end = i
					break
				}
			}
		}

		chunk := strings.TrimSpace(text[start:end])
		if chunk != "" {
			chunks = append(chunks, chunk)
		}

		start = end - ts.Overlap
		if start < 0 {
			start = 0
		}
		if start >= end {
			start = end
		}
	}

	return chunks
}

// SplitDocuments splits documents into smaller chunks
func (ts *TextSplitter) SplitDocuments(documents []*Document) []*Document {
	var result []*Document

	for _, doc := range documents {
		chunks := ts.SplitText(doc.Content)

		for i, chunk := range chunks {
			metadata := make(map[string]interface{})
			if doc.Metadata != nil {
				for k, v := range doc.Metadata {
					metadata[k] = v
				}
			}

			// Add chunk metadata
			metadata["chunk_index"] = i
			metadata["total_chunks"] = len(chunks)
			metadata["original_id"] = doc.ID

			chunkDoc := &Document{
				ID:       fmt.Sprintf("%s_chunk_%d", doc.ID, i),
				Content:  chunk,
				Metadata: metadata,
			}

			result = append(result, chunkDoc)
		}
	}

	return result
}
