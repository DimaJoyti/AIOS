package ai

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/aios/aios/pkg/models"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

// RAGServiceImpl implements the RAGService interface
type RAGServiceImpl struct {
	config       AIServiceConfig
	logger       *logrus.Logger
	tracer       trace.Tracer
	documents    map[string]*models.Document
	embeddings   map[string][]float64
	textIndex    map[string][]string // Simple text-based index
	mu           sync.RWMutex
}

// NewRAGService creates a new RAG service instance
func NewRAGService(config AIServiceConfig, logger *logrus.Logger) RAGService {
	return &RAGServiceImpl{
		config:     config,
		logger:     logger,
		tracer:     otel.Tracer("ai.rag_service"),
		documents:  make(map[string]*models.Document),
		embeddings: make(map[string][]float64),
		textIndex:  make(map[string][]string),
	}
}

// IndexDocuments indexes documents for retrieval
func (r *RAGServiceImpl) IndexDocuments(ctx context.Context, documents []models.Document) error {
	ctx, span := r.tracer.Start(ctx, "rag.IndexDocuments")
	defer span.End()

	start := time.Now()
	r.logger.WithField("document_count", len(documents)).Info("Indexing documents for RAG")

	if !r.config.RAGEnabled {
		return fmt.Errorf("RAG service is disabled")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	for _, doc := range documents {
		// Store document
		r.documents[doc.ID] = &doc

		// Generate embeddings (mock implementation)
		embedding := r.generateEmbedding(doc.Content)
		r.embeddings[doc.ID] = embedding

		// Build text index
		r.addToTextIndex(doc.ID, doc.Content)

		r.logger.WithFields(logrus.Fields{
			"doc_id":       doc.ID,
			"content_size": len(doc.Content),
		}).Debug("Document indexed")
	}

	r.logger.WithFields(logrus.Fields{
		"indexed_count":   len(documents),
		"total_docs":      len(r.documents),
		"processing_time": time.Since(start),
	}).Info("Document indexing completed")

	return nil
}

// SearchDocuments searches for relevant documents
func (r *RAGServiceImpl) SearchDocuments(ctx context.Context, query string, limit int) (*models.DocumentSearchResponse, error) {
	ctx, span := r.tracer.Start(ctx, "rag.SearchDocuments")
	defer span.End()

	start := time.Now()
	r.logger.WithFields(logrus.Fields{
		"query": query,
		"limit": limit,
	}).Info("Searching documents")

	if !r.config.RAGEnabled {
		return nil, fmt.Errorf("RAG service is disabled")
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	// Generate query embedding
	queryEmbedding := r.generateEmbedding(query)

	// Find candidate documents using text index
	candidates := r.findCandidates(query)

	// Calculate similarities and rank results
	var results []models.DocumentResult
	for _, docID := range candidates {
		doc, exists := r.documents[docID]
		if !exists {
			continue
		}

		docEmbedding, exists := r.embeddings[docID]
		if !exists {
			continue
		}

		// Calculate similarity
		similarity := r.cosineSimilarity(queryEmbedding, docEmbedding)

		// Generate snippet
		snippet := r.generateSnippet(doc.Content, query)

		result := models.DocumentResult{
			Document: *doc,
			Score:    similarity,
			Snippet:  snippet,
		}

		results = append(results, result)
	}

	// Sort by score (descending)
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	// Apply limit
	if limit > 0 && len(results) > limit {
		results = results[:limit]
	}

	response := &models.DocumentSearchResponse{
		Documents: results,
		Query:     query,
		Total:     len(results),
		Metadata: map[string]interface{}{
			"processing_time": time.Since(start).Milliseconds(),
			"candidates":      len(candidates),
			"embedding_model": r.config.EmbeddingModel,
		},
		Timestamp: time.Now(),
	}

	r.logger.WithFields(logrus.Fields{
		"query":           query,
		"results_count":   len(results),
		"processing_time": time.Since(start),
	}).Info("Document search completed")

	return response, nil
}

// GenerateWithContext generates responses using retrieved context
func (r *RAGServiceImpl) GenerateWithContext(ctx context.Context, query string, context []models.Document) (*models.RAGResponse, error) {
	ctx, span := r.tracer.Start(ctx, "rag.GenerateWithContext")
	defer span.End()

	start := time.Now()
	r.logger.WithFields(logrus.Fields{
		"query":         query,
		"context_count": len(context),
	}).Info("Generating response with context")

	if !r.config.RAGEnabled {
		return nil, fmt.Errorf("RAG service is disabled")
	}

	// TODO: Implement actual RAG generation using LLM
	// This would involve:
	// 1. Context preparation and formatting
	// 2. Prompt engineering with retrieved documents
	// 3. LLM generation with context injection
	// 4. Response post-processing

	// Mock implementation
	contextText := r.buildContextText(context)
	response := r.generateResponse(query, contextText)

	// Create document results for sources
	var sources []models.DocumentResult
	for i, doc := range context {
		sources = append(sources, models.DocumentResult{
			Document: doc,
			Score:    0.9 - float64(i)*0.1, // Decreasing relevance
			Snippet:  r.generateSnippet(doc.Content, query),
		})
	}

	ragResponse := &models.RAGResponse{
		Response: response,
		Sources:  sources,
		Query:    query,
		Model:    r.config.DefaultModel,
		Metadata: map[string]interface{}{
			"processing_time": time.Since(start).Milliseconds(),
			"context_length":  len(contextText),
			"sources_count":   len(sources),
		},
		Timestamp: time.Now(),
	}

	r.logger.WithFields(logrus.Fields{
		"query":           query,
		"response_length": len(response),
		"sources_count":   len(sources),
		"processing_time": time.Since(start),
	}).Info("RAG generation completed")

	return ragResponse, nil
}

// UpdateIndex updates the document index
func (r *RAGServiceImpl) UpdateIndex(ctx context.Context, documentID string, document models.Document) error {
	ctx, span := r.tracer.Start(ctx, "rag.UpdateIndex")
	defer span.End()

	r.logger.WithField("doc_id", documentID).Info("Updating document index")

	if !r.config.RAGEnabled {
		return fmt.Errorf("RAG service is disabled")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	// Remove old document from text index if it exists
	if oldDoc, exists := r.documents[documentID]; exists {
		r.removeFromTextIndex(documentID, oldDoc.Content)
	}

	// Update document
	r.documents[documentID] = &document

	// Generate new embeddings
	embedding := r.generateEmbedding(document.Content)
	r.embeddings[documentID] = embedding

	// Update text index
	r.addToTextIndex(documentID, document.Content)

	r.logger.WithField("doc_id", documentID).Info("Document index updated")

	return nil
}

// DeleteFromIndex removes documents from the index
func (r *RAGServiceImpl) DeleteFromIndex(ctx context.Context, documentIDs []string) error {
	ctx, span := r.tracer.Start(ctx, "rag.DeleteFromIndex")
	defer span.End()

	r.logger.WithField("doc_count", len(documentIDs)).Info("Deleting documents from index")

	if !r.config.RAGEnabled {
		return fmt.Errorf("RAG service is disabled")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	for _, docID := range documentIDs {
		if doc, exists := r.documents[docID]; exists {
			// Remove from text index
			r.removeFromTextIndex(docID, doc.Content)

			// Remove from storage
			delete(r.documents, docID)
			delete(r.embeddings, docID)

			r.logger.WithField("doc_id", docID).Debug("Document deleted from index")
		}
	}

	r.logger.WithField("deleted_count", len(documentIDs)).Info("Documents deleted from index")

	return nil
}

// Helper methods

func (r *RAGServiceImpl) generateEmbedding(text string) []float64 {
	// TODO: Implement actual embedding generation using embedding models
	// This would use models like sentence-transformers, OpenAI embeddings, etc.

	// Mock implementation - simple hash-based embedding
	dimension := 768
	embedding := make([]float64, dimension)

	// Simple hash-based approach for demo
	words := strings.Fields(strings.ToLower(text))
	for i, word := range words {
		if i >= dimension {
			break
		}
		// Simple hash function
		hash := 0
		for _, char := range word {
			hash = hash*31 + int(char)
		}
		embedding[i%dimension] = float64(hash%1000) / 1000.0
	}

	// Normalize the embedding
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

	return embedding
}

func (r *RAGServiceImpl) cosineSimilarity(a, b []float64) float64 {
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

func (r *RAGServiceImpl) findCandidates(query string) []string {
	words := strings.Fields(strings.ToLower(query))
	candidateSet := make(map[string]bool)

	for _, word := range words {
		if len(word) > 3 { // Skip short words
			if docIDs, exists := r.textIndex[word]; exists {
				for _, docID := range docIDs {
					candidateSet[docID] = true
				}
			}
		}
	}

	var candidates []string
	for docID := range candidateSet {
		candidates = append(candidates, docID)
	}

	return candidates
}

func (r *RAGServiceImpl) addToTextIndex(docID, content string) {
	words := strings.Fields(strings.ToLower(content))
	for _, word := range words {
		if len(word) > 3 { // Skip short words
			r.textIndex[word] = append(r.textIndex[word], docID)
		}
	}
}

func (r *RAGServiceImpl) removeFromTextIndex(docID, content string) {
	words := strings.Fields(strings.ToLower(content))
	for _, word := range words {
		if docIDs, exists := r.textIndex[word]; exists {
			for i, id := range docIDs {
				if id == docID {
					r.textIndex[word] = append(docIDs[:i], docIDs[i+1:]...)
					break
				}
			}
			// Clean up empty entries
			if len(r.textIndex[word]) == 0 {
				delete(r.textIndex, word)
			}
		}
	}
}

func (r *RAGServiceImpl) generateSnippet(content, query string) string {
	words := strings.Fields(strings.ToLower(query))
	contentLower := strings.ToLower(content)

	// Find the best position for snippet
	bestPos := 0
	maxMatches := 0

	for i := 0; i < len(content)-200; i += 50 {
		matches := 0
		snippet := contentLower[i:min(i+200, len(contentLower))]
		for _, word := range words {
			if strings.Contains(snippet, word) {
				matches++
			}
		}
		if matches > maxMatches {
			maxMatches = matches
			bestPos = i
		}
	}

	// Extract snippet
	start := bestPos
	end := min(start+200, len(content))
	snippet := content[start:end]

	// Clean up snippet boundaries
	if start > 0 {
		snippet = "..." + snippet
	}
	if end < len(content) {
		snippet = snippet + "..."
	}

	return snippet
}

func (r *RAGServiceImpl) buildContextText(documents []models.Document) string {
	var contextBuilder strings.Builder
	contextBuilder.WriteString("Relevant context:\n\n")

	for i, doc := range documents {
		contextBuilder.WriteString(fmt.Sprintf("Document %d (%s):\n", i+1, doc.Title))
		contextBuilder.WriteString(doc.Content)
		contextBuilder.WriteString("\n\n")
	}

	return contextBuilder.String()
}

func (r *RAGServiceImpl) generateResponse(query, context string) string {
	// TODO: Implement actual LLM generation with context
	// This would involve calling the LLM service with the context

	// Mock implementation
	contextLength := len(context)
	switch {
	case contextLength > 5000:
		return fmt.Sprintf("Based on the extensive context provided, I can give you a comprehensive answer to '%s'. The documents contain detailed information that directly addresses your question with multiple perspectives and supporting evidence.", query)
	case contextLength > 1000:
		return fmt.Sprintf("According to the relevant documents, regarding '%s': The context provides good coverage of this topic with specific details and examples that help answer your question.", query)
	case contextLength > 100:
		return fmt.Sprintf("Based on the available context about '%s': I can provide a basic answer using the information from the retrieved documents, though the context is somewhat limited.", query)
	default:
		return fmt.Sprintf("I have limited context to answer '%s', but based on what's available, I can provide some general information.", query)
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
