package knowledge

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aios/aios/pkg/langchain/llm"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// DefaultRAGPipeline implements the RAGPipeline interface
type DefaultRAGPipeline struct {
	embeddingManager EmbeddingManager
	llmManager       llm.LLMManager
	retriever        DocumentRetriever
	reranker         DocumentReranker
	generator        ResponseGenerator
	logger           *logrus.Logger
	tracer           trace.Tracer
	config           *RAGPipelineConfig
}

// RAGPipelineConfig represents configuration for the RAG pipeline
type RAGPipelineConfig struct {
	DefaultTopK           int     `json:"default_top_k"`
	DefaultThreshold      float32 `json:"default_threshold"`
	MaxContextLength      int     `json:"max_context_length"`
	RerankingEnabled      bool    `json:"reranking_enabled"`
	CitationEnabled       bool    `json:"citation_enabled"`
	StreamingEnabled      bool    `json:"streaming_enabled"`
	DefaultModel          string  `json:"default_model"`
	DefaultTemperature    float32 `json:"default_temperature"`
	DefaultMaxTokens      int     `json:"default_max_tokens"`
}

// DocumentRetriever handles document retrieval
type DocumentRetriever interface {
	Retrieve(ctx context.Context, query string, options *RetrievalOptions) ([]*Document, error)
	RetrieveByEmbedding(ctx context.Context, embedding []float32, options *RetrievalOptions) ([]*Document, error)
	HybridRetrieve(ctx context.Context, query string, options *RetrievalOptions) ([]*Document, error)
}

// DocumentReranker handles document reranking
type DocumentReranker interface {
	Rerank(ctx context.Context, query string, documents []*Document) ([]*Document, error)
	CalculateRelevanceScore(ctx context.Context, query string, document *Document) (float32, error)
}

// ResponseGenerator handles response generation
type ResponseGenerator interface {
	Generate(ctx context.Context, query string, context []*Document, options *GenerationOptions) (*GenerationResult, error)
	GenerateWithCitations(ctx context.Context, query string, context []*Document, options *GenerationOptions) (*GenerationResult, error)
	StreamGenerate(ctx context.Context, query string, context []*Document, options *GenerationOptions) (<-chan string, error)
}

// NewDefaultRAGPipeline creates a new default RAG pipeline
func NewDefaultRAGPipeline(embeddingManager EmbeddingManager, logger *logrus.Logger) (RAGPipeline, error) {
	config := &RAGPipelineConfig{
		DefaultTopK:        10,
		DefaultThreshold:   0.7,
		MaxContextLength:   8000,
		RerankingEnabled:   true,
		CitationEnabled:    true,
		StreamingEnabled:   false,
		DefaultModel:       "gpt-4",
		DefaultTemperature: 0.7,
		DefaultMaxTokens:   2000,
	}

	pipeline := &DefaultRAGPipeline{
		embeddingManager: embeddingManager,
		config:           config,
		logger:           logger,
		tracer:           otel.Tracer("knowledge.rag_pipeline"),
	}

	// Initialize components
	if err := pipeline.initializeComponents(); err != nil {
		return nil, fmt.Errorf("failed to initialize RAG pipeline components: %w", err)
	}

	return pipeline, nil
}

// initializeComponents initializes RAG pipeline components
func (rp *DefaultRAGPipeline) initializeComponents() error {
	// Initialize retriever
	retriever, err := NewDefaultDocumentRetriever(rp.embeddingManager, rp.logger)
	if err != nil {
		return fmt.Errorf("failed to initialize document retriever: %w", err)
	}
	rp.retriever = retriever

	// Initialize reranker
	reranker, err := NewDefaultDocumentReranker(rp.embeddingManager, rp.logger)
	if err != nil {
		return fmt.Errorf("failed to initialize document reranker: %w", err)
	}
	rp.reranker = reranker

	// Initialize generator
	generator, err := NewDefaultResponseGenerator(rp.logger)
	if err != nil {
		return fmt.Errorf("failed to initialize response generator: %w", err)
	}
	rp.generator = generator

	return nil
}

// Retrieve retrieves relevant documents for a query
func (rp *DefaultRAGPipeline) Retrieve(ctx context.Context, query string, options *RetrievalOptions) (*RetrievalResult, error) {
	ctx, span := rp.tracer.Start(ctx, "rag_pipeline.retrieve")
	defer span.End()

	startTime := time.Now()
	span.SetAttributes(
		attribute.String("query", query),
		attribute.Int("top_k", options.TopK),
		attribute.Float64("threshold", float64(options.Threshold)),
	)

	// Set default options
	if options == nil {
		options = &RetrievalOptions{
			TopK:      rp.config.DefaultTopK,
			Threshold: rp.config.DefaultThreshold,
		}
	}

	// Retrieve documents
	var documents []*Document
	var err error

	if options.HybridSearch {
		documents, err = rp.retriever.HybridRetrieve(ctx, query, options)
	} else {
		documents, err = rp.retriever.Retrieve(ctx, query, options)
	}

	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to retrieve documents: %w", err)
	}

	// Rerank if enabled
	if rp.config.RerankingEnabled && options.RerankingEnabled {
		documents, err = rp.reranker.Rerank(ctx, query, documents)
		if err != nil {
			rp.logger.WithError(err).Warn("Failed to rerank documents, using original order")
		}
	}

	// Build context
	context := rp.buildContext(documents, rp.config.MaxContextLength)

	result := &RetrievalResult{
		Documents:      documents,
		Context:        context,
		Query:          query,
		ProcessingTime: time.Since(startTime),
		Metadata: map[string]interface{}{
			"retrieved_count": len(documents),
			"context_length":  len(context),
			"reranked":        rp.config.RerankingEnabled && options.RerankingEnabled,
		},
	}

	rp.logger.WithFields(logrus.Fields{
		"query":           query,
		"retrieved_count": len(documents),
		"context_length":  len(context),
		"processing_time": time.Since(startTime),
	}).Debug("Documents retrieved successfully")

	return result, nil
}

// Rerank reranks documents based on relevance to the query
func (rp *DefaultRAGPipeline) Rerank(ctx context.Context, query string, documents []*Document) ([]*Document, error) {
	ctx, span := rp.tracer.Start(ctx, "rag_pipeline.rerank")
	defer span.End()

	span.SetAttributes(
		attribute.String("query", query),
		attribute.Int("document_count", len(documents)),
	)

	return rp.reranker.Rerank(ctx, query, documents)
}

// Generate generates a response using retrieved context
func (rp *DefaultRAGPipeline) Generate(ctx context.Context, query string, context []*Document, options *GenerationOptions) (*GenerationResult, error) {
	ctx, span := rp.tracer.Start(ctx, "rag_pipeline.generate")
	defer span.End()

	span.SetAttributes(
		attribute.String("query", query),
		attribute.Int("context_count", len(context)),
	)

	// Set default options
	if options == nil {
		options = &GenerationOptions{
			Model:       rp.config.DefaultModel,
			Temperature: rp.config.DefaultTemperature,
			MaxTokens:   rp.config.DefaultMaxTokens,
		}
	}

	// Generate response
	if rp.config.CitationEnabled {
		return rp.generator.GenerateWithCitations(ctx, query, context, options)
	}

	return rp.generator.Generate(ctx, query, context, options)
}

// Pipeline executes the complete RAG pipeline
func (rp *DefaultRAGPipeline) Pipeline(ctx context.Context, query string, options *RAGOptions) (*RAGResponse, error) {
	ctx, span := rp.tracer.Start(ctx, "rag_pipeline.pipeline")
	defer span.End()

	startTime := time.Now()
	span.SetAttributes(attribute.String("query", query))

	// Set default options
	if options == nil {
		options = &RAGOptions{
			RetrievalOptions: &RetrievalOptions{
				TopK:      rp.config.DefaultTopK,
				Threshold: rp.config.DefaultThreshold,
			},
			GenerationOptions: &GenerationOptions{
				Model:       rp.config.DefaultModel,
				Temperature: rp.config.DefaultTemperature,
				MaxTokens:   rp.config.DefaultMaxTokens,
			},
			ContextLength:  rp.config.MaxContextLength,
			IncludeSources: true,
		}
	}

	// Step 1: Retrieve relevant documents
	retrievalResult, err := rp.Retrieve(ctx, query, options.RetrievalOptions)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("retrieval failed: %w", err)
	}

	// Step 2: Generate response
	generationResult, err := rp.Generate(ctx, query, retrievalResult.Documents, options.GenerationOptions)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("generation failed: %w", err)
	}

	// Step 3: Build RAG response
	response := &RAGResponse{
		Response:       generationResult.Response,
		Query:          query,
		Context:        retrievalResult.Context,
		ProcessingTime: time.Since(startTime),
		Timestamp:      time.Now(),
		Metadata: map[string]interface{}{
			"retrieval_time":  retrievalResult.ProcessingTime,
			"generation_time": generationResult.ProcessingTime,
			"total_time":      time.Since(startTime),
			"model":           generationResult.Model,
			"tokens_used":     generationResult.TokensUsed,
		},
	}

	// Add sources if requested
	if options.IncludeSources {
		response.Sources = retrievalResult.Documents
	}

	// Add citations if available
	if rp.config.CitationEnabled {
		response.Citations = rp.extractCitations(generationResult.Response, retrievalResult.Documents)
	}

	// Calculate confidence score
	response.Confidence = rp.calculateConfidence(retrievalResult.Documents, generationResult)

	rp.logger.WithFields(logrus.Fields{
		"query":           query,
		"sources_count":   len(response.Sources),
		"citations_count": len(response.Citations),
		"confidence":      response.Confidence,
		"processing_time": response.ProcessingTime,
	}).Info("RAG pipeline completed successfully")

	return response, nil
}

// buildContext builds context string from documents
func (rp *DefaultRAGPipeline) buildContext(documents []*Document, maxLength int) string {
	var contextParts []string
	currentLength := 0

	for i, doc := range documents {
		docText := fmt.Sprintf("[Document %d: %s]\n%s\n", i+1, doc.Title, doc.Content)
		
		if currentLength+len(docText) > maxLength {
			// Try to fit partial content
			remaining := maxLength - currentLength
			if remaining > 100 { // Only add if we have reasonable space
				truncated := docText[:remaining-3] + "..."
				contextParts = append(contextParts, truncated)
			}
			break
		}

		contextParts = append(contextParts, docText)
		currentLength += len(docText)
	}

	return strings.Join(contextParts, "\n")
}

// extractCitations extracts citations from the response
func (rp *DefaultRAGPipeline) extractCitations(response string, documents []*Document) []*Citation {
	var citations []*Citation

	// Simple citation extraction based on document references
	for i, doc := range documents {
		docRef := fmt.Sprintf("[Document %d", i+1)
		if strings.Contains(response, docRef) {
			citation := &Citation{
				DocumentID: doc.ID,
				Text:       doc.Title,
				Confidence: 0.8, // Default confidence
			}
			citations = append(citations, citation)
		}
	}

	return citations
}

// calculateConfidence calculates confidence score for the response
func (rp *DefaultRAGPipeline) calculateConfidence(documents []*Document, result *GenerationResult) float32 {
	if len(documents) == 0 {
		return 0.0
	}

	// Simple confidence calculation based on number and quality of sources
	baseConfidence := float32(0.5)
	
	// Boost confidence based on number of relevant documents
	documentBoost := float32(len(documents)) * 0.1
	if documentBoost > 0.4 {
		documentBoost = 0.4
	}

	// Consider response length as a factor
	responseLength := len(result.Response)
	lengthBoost := float32(0.0)
	if responseLength > 100 {
		lengthBoost = 0.1
	}

	confidence := baseConfidence + documentBoost + lengthBoost
	if confidence > 1.0 {
		confidence = 1.0
	}

	return confidence
}
