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

// DefaultResponseGenerator implements the ResponseGenerator interface
type DefaultResponseGenerator struct {
	llmManager llm.LLMManager
	logger     *logrus.Logger
	tracer     trace.Tracer
	config     *ResponseGeneratorConfig
}

// ResponseGeneratorConfig represents configuration for the response generator
type ResponseGeneratorConfig struct {
	DefaultModel       string  `json:"default_model"`
	DefaultTemperature float32 `json:"default_temperature"`
	DefaultMaxTokens   int     `json:"default_max_tokens"`
	SystemPrompt       string  `json:"system_prompt"`
	CitationTemplate   string  `json:"citation_template"`
	MaxContextLength   int     `json:"max_context_length"`
}

// NewDefaultResponseGenerator creates a new default response generator
func NewDefaultResponseGenerator(logger *logrus.Logger) (ResponseGenerator, error) {
	config := &ResponseGeneratorConfig{
		DefaultModel:       "gpt-4",
		DefaultTemperature: 0.7,
		DefaultMaxTokens:   2000,
		MaxContextLength:   8000,
		SystemPrompt: `You are a helpful AI assistant that answers questions based on the provided context. 
Use only the information from the context to answer questions. If the context doesn't contain enough information to answer the question, say so clearly.
Be accurate, concise, and helpful in your responses.`,
		CitationTemplate: "[Source: %s]",
	}

	generator := &DefaultResponseGenerator{
		logger: logger,
		tracer: otel.Tracer("knowledge.response_generator"),
		config: config,
	}

	// Initialize LLM manager
	if err := generator.initializeLLMManager(); err != nil {
		return nil, fmt.Errorf("failed to initialize LLM manager: %w", err)
	}

	return generator, nil
}

// initializeLLMManager initializes the LLM manager
func (rg *DefaultResponseGenerator) initializeLLMManager() error {
	// Create LLM factory
	factory := llm.NewLLMFactory(rg.logger)
	
	// Create LLM manager
	manager := llm.NewLLMManager(factory, rg.logger)

	rg.llmManager = manager
	return nil
}

// Generate generates a response using the provided context
func (rg *DefaultResponseGenerator) Generate(ctx context.Context, query string, context []*Document, options *GenerationOptions) (*GenerationResult, error) {
	ctx, span := rg.tracer.Start(ctx, "response_generator.generate")
	defer span.End()

	startTime := time.Now()
	span.SetAttributes(
		attribute.String("query", query),
		attribute.Int("context_count", len(context)),
		attribute.String("model", options.Model),
	)

	// Set default options
	if options == nil {
		options = &GenerationOptions{
			Model:       rg.config.DefaultModel,
			Temperature: rg.config.DefaultTemperature,
			MaxTokens:   rg.config.DefaultMaxTokens,
		}
	}

	// Build context string
	contextStr := rg.buildContextString(context)

	// Build prompt
	prompt := rg.buildPrompt(query, contextStr, options)

	// Generate response
	response, err := rg.generateResponse(ctx, prompt, options)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to generate response: %w", err)
	}

	result := &GenerationResult{
		Response:       response.Content,
		TokensUsed:     response.Usage.TotalTokens,
		Model:          options.Model,
		ProcessingTime: time.Since(startTime),
		Metadata: map[string]interface{}{
			"prompt_tokens":     response.Usage.PromptTokens,
			"completion_tokens": response.Usage.CompletionTokens,
			"context_length":    len(contextStr),
		},
	}

	rg.logger.WithFields(logrus.Fields{
		"query":           query,
		"model":           options.Model,
		"tokens_used":     result.TokensUsed,
		"processing_time": result.ProcessingTime,
	}).Debug("Response generated successfully")

	return result, nil
}

// GenerateWithCitations generates a response with citations
func (rg *DefaultResponseGenerator) GenerateWithCitations(ctx context.Context, query string, context []*Document, options *GenerationOptions) (*GenerationResult, error) {
	ctx, span := rg.tracer.Start(ctx, "response_generator.generate_with_citations")
	defer span.End()

	// Generate regular response first
	result, err := rg.Generate(ctx, query, context, options)
	if err != nil {
		return nil, err
	}

	// Add citations to the response
	citedResponse := rg.addCitations(result.Response, context)
	result.Response = citedResponse

	return result, nil
}

// StreamGenerate generates a streaming response
func (rg *DefaultResponseGenerator) StreamGenerate(ctx context.Context, query string, context []*Document, options *GenerationOptions) (<-chan string, error) {
	ctx, span := rg.tracer.Start(ctx, "response_generator.stream_generate")
	defer span.End()

	// For now, return a simple channel with the complete response
	// In a real implementation, this would stream tokens as they're generated
	responseChan := make(chan string, 1)

	go func() {
		defer close(responseChan)

		result, err := rg.Generate(ctx, query, context, options)
		if err != nil {
			rg.logger.WithError(err).Error("Failed to generate streaming response")
			return
		}

		responseChan <- result.Response
	}()

	return responseChan, nil
}

// buildContextString builds a context string from documents
func (rg *DefaultResponseGenerator) buildContextString(documents []*Document) string {
	var contextParts []string

	for i, doc := range documents {
		// Limit context length
		content := doc.Content
		if len(content) > 1000 {
			content = content[:1000] + "..."
		}

		contextPart := fmt.Sprintf("Document %d (Title: %s):\n%s", i+1, doc.Title, content)
		contextParts = append(contextParts, contextPart)

		// Check total length
		totalLength := len(strings.Join(contextParts, "\n\n"))
		if totalLength > rg.config.MaxContextLength {
			break
		}
	}

	return strings.Join(contextParts, "\n\n")
}

// buildPrompt builds the complete prompt for the LLM
func (rg *DefaultResponseGenerator) buildPrompt(query, context string, options *GenerationOptions) string {
	systemPrompt := rg.config.SystemPrompt
	if options.SystemPrompt != "" {
		systemPrompt = options.SystemPrompt
	}

	instructions := ""
	if options.Instructions != "" {
		instructions = "\nAdditional Instructions: " + options.Instructions
	}

	prompt := fmt.Sprintf(`%s%s

Context:
%s

Question: %s

Answer:`, systemPrompt, instructions, context, query)

	return prompt
}

// generateResponse generates a response using the LLM
func (rg *DefaultResponseGenerator) generateResponse(ctx context.Context, prompt string, options *GenerationOptions) (*llm.CompletionResponse, error) {
	request := &llm.CompletionRequest{
		Model: options.Model,
		Messages: []llm.Message{
			{
				Role:    "user",
				Content: prompt,
			},
		},
		Temperature: float64(options.Temperature),
		MaxTokens:   options.MaxTokens,
	}

	// Add any additional parameters
	if options.Parameters != nil {
		// Handle additional parameters as needed
		if topP, exists := options.Parameters["top_p"]; exists {
			if topPFloat, ok := topP.(float64); ok {
				request.TopP = topPFloat
			}
		}
	}

	return rg.llmManager.Complete(ctx, request)
}

// addCitations adds citations to the response
func (rg *DefaultResponseGenerator) addCitations(response string, documents []*Document) string {
	// Simple citation addition - in a real implementation, this would be more sophisticated
	var citations []string

	for i, doc := range documents {
		citation := fmt.Sprintf(rg.config.CitationTemplate, doc.Title)
		citations = append(citations, fmt.Sprintf("[%d] %s", i+1, citation))
	}

	if len(citations) > 0 {
		citationText := "\n\nSources:\n" + strings.Join(citations, "\n")
		response += citationText
	}

	return response
}

// DefaultSemanticCache implements the SemanticCache interface
type DefaultSemanticCache struct {
	cache   map[string]*CachedQuery
	logger  *logrus.Logger
	tracer  trace.Tracer
	config  *SemanticCacheConfig
}

// SemanticCacheConfig represents configuration for the semantic cache
type SemanticCacheConfig struct {
	MaxSize           int           `json:"max_size"`
	DefaultTTL        time.Duration `json:"default_ttl"`
	SimilarityThreshold float32     `json:"similarity_threshold"`
	EnableCompression bool          `json:"enable_compression"`
}

// NewDefaultSemanticCache creates a new default semantic cache
func NewDefaultSemanticCache(defaultTTL time.Duration, logger *logrus.Logger) (SemanticCache, error) {
	config := &SemanticCacheConfig{
		MaxSize:             10000,
		DefaultTTL:          defaultTTL,
		SimilarityThreshold: 0.9,
		EnableCompression:   false,
	}

	cache := &DefaultSemanticCache{
		cache:  make(map[string]*CachedQuery),
		logger: logger,
		tracer: otel.Tracer("knowledge.semantic_cache"),
		config: config,
	}

	return cache, nil
}

// GetSimilarQueries finds similar queries in the cache
func (sc *DefaultSemanticCache) GetSimilarQueries(ctx context.Context, query string, threshold float32) ([]*CachedQuery, error) {
	ctx, span := sc.tracer.Start(ctx, "semantic_cache.get_similar_queries")
	defer span.End()

	var similar []*CachedQuery

	// Simple string similarity for now
	// In a real implementation, this would use embedding similarity
	for _, cached := range sc.cache {
		if time.Now().After(cached.ExpiresAt) {
			continue
		}

		similarity := calculateStringSimilarity(query, cached.Query)
		if similarity >= threshold {
			cached.Similarity = similarity
			similar = append(similar, cached)
		}
	}

	return similar, nil
}

// CacheQuery caches a query and its result
func (sc *DefaultSemanticCache) CacheQuery(ctx context.Context, query string, result interface{}, ttl time.Duration) error {
	ctx, span := sc.tracer.Start(ctx, "semantic_cache.cache_query")
	defer span.End()

	if ttl == 0 {
		ttl = sc.config.DefaultTTL
	}

	cached := &CachedQuery{
		Query:     query,
		Result:    result,
		CachedAt:  time.Now(),
		ExpiresAt: time.Now().Add(ttl),
	}

	sc.cache[query] = cached

	// Simple eviction if cache is full
	if len(sc.cache) > sc.config.MaxSize {
		sc.evictOldest()
	}

	return nil
}

// CacheEmbedding caches an embedding
func (sc *DefaultSemanticCache) CacheEmbedding(ctx context.Context, text string, embedding []float32, ttl time.Duration) error {
	// For now, store as a regular cache entry
	return sc.CacheQuery(ctx, "embedding:"+text, embedding, ttl)
}

// GetCachedEmbedding retrieves a cached embedding
func (sc *DefaultSemanticCache) GetCachedEmbedding(ctx context.Context, text string) ([]float32, bool, error) {
	key := "embedding:" + text
	if cached, exists := sc.cache[key]; exists && time.Now().Before(cached.ExpiresAt) {
		if embedding, ok := cached.Result.([]float32); ok {
			return embedding, true, nil
		}
	}
	return nil, false, nil
}

// InvalidateCache invalidates cache entries matching a pattern
func (sc *DefaultSemanticCache) InvalidateCache(ctx context.Context, pattern string) error {
	for key := range sc.cache {
		if strings.Contains(key, pattern) {
			delete(sc.cache, key)
		}
	}
	return nil
}

// GetCacheStats returns cache statistics
func (sc *DefaultSemanticCache) GetCacheStats(ctx context.Context) (*CacheStats, error) {
	totalEntries := len(sc.cache)
	validEntries := 0

	for _, cached := range sc.cache {
		if time.Now().Before(cached.ExpiresAt) {
			validEntries++
		}
	}

	_ = totalEntries // Mark as used for potential future use

	stats := &CacheStats{
		Size:    validEntries,
		MaxSize: sc.config.MaxSize,
		// Note: Hit/miss rates would need to be tracked separately
	}

	return stats, nil
}

// evictOldest removes the oldest cache entry
func (sc *DefaultSemanticCache) evictOldest() {
	var oldestKey string
	var oldestTime time.Time

	for key, cached := range sc.cache {
		if oldestKey == "" || cached.CachedAt.Before(oldestTime) {
			oldestKey = key
			oldestTime = cached.CachedAt
		}
	}

	if oldestKey != "" {
		delete(sc.cache, oldestKey)
	}
}

// calculateStringSimilarity calculates simple string similarity
func calculateStringSimilarity(s1, s2 string) float32 {
	if s1 == s2 {
		return 1.0
	}

	// Simple Jaccard similarity based on words
	words1 := strings.Fields(strings.ToLower(s1))
	words2 := strings.Fields(strings.ToLower(s2))

	set1 := make(map[string]bool)
	set2 := make(map[string]bool)

	for _, word := range words1 {
		set1[word] = true
	}
	for _, word := range words2 {
		set2[word] = true
	}

	intersection := 0
	for word := range set1 {
		if set2[word] {
			intersection++
		}
	}

	union := len(set1) + len(set2) - intersection
	if union == 0 {
		return 0.0
	}

	return float32(intersection) / float32(union)
}
