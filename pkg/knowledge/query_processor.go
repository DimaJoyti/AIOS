package knowledge

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// DefaultQueryProcessor implements the QueryProcessor interface
type DefaultQueryProcessor struct {
	embeddingManager EmbeddingManager
	logger           *logrus.Logger
	tracer           trace.Tracer
}

// NewDefaultQueryProcessor creates a new default query processor
func NewDefaultQueryProcessor(embeddingManager EmbeddingManager, logger *logrus.Logger) (QueryProcessor, error) {
	return &DefaultQueryProcessor{
		embeddingManager: embeddingManager,
		logger:           logger,
		tracer:           otel.Tracer("knowledge.query_processor"),
	}, nil
}

// ProcessQuery processes a query and returns a processed query
func (qp *DefaultQueryProcessor) ProcessQuery(ctx context.Context, query string) (*ProcessedQuery, error) {
	ctx, span := qp.tracer.Start(ctx, "query_processor.process_query")
	defer span.End()

	span.SetAttributes(
		attribute.String("query", query),
		attribute.Int("query.length", len(query)),
	)

	// Clean and normalize the query
	cleanedQuery := strings.TrimSpace(query)
	normalizedQuery := strings.ToLower(cleanedQuery)

	// Extract keywords (simple word splitting)
	keywords := strings.Fields(normalizedQuery)
	
	// Remove common stop words
	stopWords := map[string]bool{
		"the": true, "a": true, "an": true, "and": true, "or": true, "but": true,
		"in": true, "on": true, "at": true, "to": true, "for": true, "of": true,
		"with": true, "by": true, "is": true, "are": true, "was": true, "were": true,
		"be": true, "been": true, "being": true, "have": true, "has": true, "had": true,
		"do": true, "does": true, "did": true, "will": true, "would": true, "could": true,
		"should": true, "what": true, "where": true, "when": true, "why": true, "how": true,
	}

	var filteredKeywords []string
	for _, keyword := range keywords {
		if !stopWords[keyword] && len(keyword) > 2 {
			filteredKeywords = append(filteredKeywords, keyword)
		}
	}

	// Generate embedding for the query
	embedding, err := qp.embeddingManager.GenerateEmbedding(ctx, cleanedQuery)
	if err != nil {
		span.RecordError(err)
		qp.logger.WithError(err).Warn("Failed to generate query embedding")
		// Continue without embedding
		embedding = nil
	}

	processed := &ProcessedQuery{
		OriginalQuery: query,
		CleanedQuery:  cleanedQuery,
		Keywords:      filteredKeywords,
		Embedding:     embedding,
		Language:      "en", // Simple assumption
		ProcessedAt:   time.Now(),
	}

	span.SetAttributes(
		attribute.Int("keywords.count", len(filteredKeywords)),
		attribute.Bool("embedding.available", embedding != nil),
	)

	return processed, nil
}

// ExpandQuery expands a query with related terms
func (qp *DefaultQueryProcessor) ExpandQuery(ctx context.Context, query string) (*ExpandedQuery, error) {
	ctx, span := qp.tracer.Start(ctx, "query_processor.expand_query")
	defer span.End()

	// First process the query
	processed, err := qp.ProcessQuery(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to process query: %w", err)
	}

	// Simple query expansion (in a real system, this would use a thesaurus or word embeddings)
	synonyms := make(map[string][]string)
	relatedTerms := make(map[string][]string)

	// Add some basic synonyms and related terms
	for _, keyword := range processed.Keywords {
		switch keyword {
		case "ai", "artificial":
			synonyms[keyword] = []string{"artificial intelligence", "machine intelligence"}
			relatedTerms[keyword] = []string{"machine learning", "deep learning", "neural networks"}
		case "machine", "learning":
			synonyms[keyword] = []string{"ml", "automated learning"}
			relatedTerms[keyword] = []string{"artificial intelligence", "data science", "algorithms"}
		case "neural", "networks":
			synonyms[keyword] = []string{"nn", "artificial neural networks"}
			relatedTerms[keyword] = []string{"deep learning", "perceptron", "backpropagation"}
		case "language", "processing":
			synonyms[keyword] = []string{"nlp", "natural language processing"}
			relatedTerms[keyword] = []string{"text processing", "linguistics", "computational linguistics"}
		}
	}

	// Build expanded terms
	var expandedTermsList []string
	var synonymsList []string
	var relatedTermsList []string
	
	expandedTermsList = append(expandedTermsList, processed.CleanedQuery)
	
	for _, synonymGroup := range synonyms {
		synonymsList = append(synonymsList, synonymGroup...)
		expandedTermsList = append(expandedTermsList, synonymGroup...)
	}
	
	for _, relatedGroup := range relatedTerms {
		relatedTermsList = append(relatedTermsList, relatedGroup...)
		expandedTermsList = append(expandedTermsList, relatedGroup...)
	}

	expanded := &ExpandedQuery{
		OriginalQuery: query,
		ExpandedTerms: expandedTermsList,
		Synonyms:      synonymsList,
		RelatedTerms:  relatedTermsList,
	}

	span.SetAttributes(
		attribute.Int("synonyms.count", len(synonyms)),
		attribute.Int("related_terms.count", len(relatedTerms)),
	)

	return expanded, nil
}

// ExtractIntent extracts intent from a query
func (qp *DefaultQueryProcessor) ExtractIntent(ctx context.Context, query string) (*QueryIntent, error) {
	ctx, span := qp.tracer.Start(ctx, "query_processor.extract_intent")
	defer span.End()

	normalizedQuery := strings.ToLower(strings.TrimSpace(query))

	// Simple intent classification based on keywords and patterns
	var intentType string
	var confidence float32 = 0.5 // Default confidence
	var category string
	var action string

	// Question patterns
	if strings.HasPrefix(normalizedQuery, "what is") || 
	   strings.HasPrefix(normalizedQuery, "what are") ||
	   strings.Contains(normalizedQuery, "explain") ||
	   strings.Contains(normalizedQuery, "describe") {
		intentType = "informational"
		category = "question"
		action = "explain"
		confidence = 0.8
	} else if strings.HasPrefix(normalizedQuery, "how to") ||
		     strings.HasPrefix(normalizedQuery, "how do") ||
		     strings.Contains(normalizedQuery, "tutorial") ||
		     strings.Contains(normalizedQuery, "guide") {
		intentType = "instructional"
		category = "how_to"
		action = "guide"
		confidence = 0.8
	} else if strings.Contains(normalizedQuery, "compare") ||
		     strings.Contains(normalizedQuery, "vs") ||
		     strings.Contains(normalizedQuery, "versus") ||
		     strings.Contains(normalizedQuery, "difference") {
		intentType = "comparative"
		category = "comparison"
		action = "compare"
		confidence = 0.8
	} else if strings.Contains(normalizedQuery, "find") ||
		     strings.Contains(normalizedQuery, "search") ||
		     strings.Contains(normalizedQuery, "look for") {
		intentType = "search"
		category = "retrieval"
		action = "find"
		confidence = 0.7
	} else {
		intentType = "informational" // Default
		category = "general"
		action = "retrieve"
		confidence = 0.5
	}

	intent := &QueryIntent{
		Type:       intentType,
		Confidence: confidence,
		Category:   category,
		Action:     action,
	}

	span.SetAttributes(
		attribute.String("intent.type", intentType),
		attribute.Float64("intent.confidence", float64(confidence)),
		attribute.String("intent.category", category),
	)

	return intent, nil
}

// ExtractEntities extracts entities from a query
func (qp *DefaultQueryProcessor) ExtractEntities(ctx context.Context, query string) ([]*QueryEntity, error) {
	ctx, span := qp.tracer.Start(ctx, "query_processor.extract_entities")
	defer span.End()

	normalizedQuery := strings.ToLower(strings.TrimSpace(query))
	var entities []*QueryEntity

	// Extract simple entities (topics)
	knownTopics := []string{
		"artificial intelligence", "ai", "machine learning", "ml", "deep learning",
		"neural networks", "natural language processing", "nlp", "computer vision",
		"robotics", "data science", "algorithms", "programming", "technology",
	}

	for _, topic := range knownTopics {
		if strings.Contains(normalizedQuery, topic) {
			startPos := strings.Index(normalizedQuery, topic)
			entity := &QueryEntity{
				Type:       "topic",
				Text:       topic,
				StartPos:   startPos,
				EndPos:     startPos + len(topic),
				Confidence: 0.8,
			}
			entities = append(entities, entity)
		}
	}

	span.SetAttributes(
		attribute.Int("entities.count", len(entities)),
	)

	return entities, nil
}

// RewriteQuery rewrites a query based on context
func (qp *DefaultQueryProcessor) RewriteQuery(ctx context.Context, query string, context *QueryContext) (string, error) {
	ctx, span := qp.tracer.Start(ctx, "query_processor.rewrite_query")
	defer span.End()

	// Simple query rewriting - in a real system this would be more sophisticated
	rewrittenQuery := strings.TrimSpace(query)
	
	// If context has history, we could use it for context
	if context != nil && len(context.History) > 0 {
		// For now, just return the original query
		// In a real system, this would analyze conversation context
		qp.logger.WithFields(logrus.Fields{
			"original_query":   query,
			"history_entries":  len(context.History),
		}).Debug("Rewriting query with context")
	}

	span.SetAttributes(
		attribute.String("original_query", query),
		attribute.String("rewritten_query", rewrittenQuery),
		attribute.Bool("context_available", context != nil),
	)

	return rewrittenQuery, nil
}