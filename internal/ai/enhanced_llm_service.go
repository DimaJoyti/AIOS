package ai

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aios/aios/internal/ai/cache"
	"github.com/aios/aios/internal/ai/providers"
	"github.com/aios/aios/pkg/models"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

// EnhancedLLMService provides real LLM integration with Ollama
type EnhancedLLMService struct {
	config        AIServiceConfig
	logger        *logrus.Logger
	tracer        trace.Tracer
	ollamaClient  *providers.OllamaClient
	semanticCache *cache.SemanticCacheManager
	memoryCache   *cache.MemoryCacheManager
}

// NewEnhancedLLMService creates a new enhanced LLM service with real Ollama integration
func NewEnhancedLLMService(config AIServiceConfig, logger *logrus.Logger, semanticCache *cache.SemanticCacheManager, memoryCache *cache.MemoryCacheManager) *EnhancedLLMService {
	ollamaURL := fmt.Sprintf("http://%s:%d", config.OllamaHost, config.OllamaPort)
	ollamaClient := providers.NewOllamaClient(ollamaURL, logger)

	return &EnhancedLLMService{
		config:        config,
		logger:        logger,
		tracer:        otel.Tracer("ai.enhanced_llm_service"),
		ollamaClient:  ollamaClient,
		semanticCache: semanticCache,
		memoryCache:   memoryCache,
	}
}

// ProcessQuery processes a query using real LLM with caching
func (s *EnhancedLLMService) ProcessQuery(ctx context.Context, query string) (*models.LLMResponse, error) {
	ctx, span := s.tracer.Start(ctx, "enhanced_llm.ProcessQuery")
	defer span.End()

	start := time.Now()
	s.logger.WithField("query", query[:min(100, len(query))]).Info("Processing LLM query")

	// TODO: Implement caching - simplified for now
	// Check memory cache
	cacheKey := fmt.Sprintf("llm_query_%s_%s", s.config.DefaultModel, hashString(query))
	if s.memoryCache != nil {
		if cached, found, err := s.memoryCache.Get(ctx, cacheKey); err == nil && found {
			s.logger.Info("Query found in memory cache")
			if response, ok := cached.(*models.LLMResponse); ok {
				return response, nil
			}
		}
	}

	// Process with Ollama
	ollamaReq := &providers.OllamaRequest{
		Model:  s.config.DefaultModel,
		Prompt: query,
		Options: map[string]interface{}{
			"temperature": s.config.Temperature,
			"num_predict": s.config.MaxTokens,
		},
	}

	ollamaResp, err := s.ollamaClient.Generate(ctx, ollamaReq)
	if err != nil {
		s.logger.WithError(err).Error("Failed to generate response with Ollama")
		return nil, fmt.Errorf("ollama generation failed: %w", err)
	}

	// Convert to LLM response
	response := &models.LLMResponse{
		Text:           ollamaResp.Response,
		Model:          ollamaResp.Model,
		TokensUsed:     ollamaResp.EvalCount,
		Confidence:     0.85, // Default confidence
		ProcessingTime: time.Since(start),
		Timestamp:      time.Now(),
		Metadata: map[string]interface{}{
			"prompt_eval_count":    ollamaResp.PromptEvalCount,
			"prompt_eval_duration": ollamaResp.PromptEvalDuration,
			"eval_duration":        ollamaResp.EvalDuration,
			"total_duration":       ollamaResp.TotalDuration,
			"load_duration":        ollamaResp.LoadDuration,
			"cached":               false,
		},
	}

	// Cache the response
	if s.config.SemanticCache && s.semanticCache != nil {
		// TODO: Implement semantic caching with proper vector embeddings
	}
	if s.config.CacheEnabled && s.memoryCache != nil {
		s.memoryCache.Set(ctx, cacheKey, response, s.config.CacheTTL)
	}

	s.logger.WithFields(logrus.Fields{
		"query":           query[:min(100, len(query))],
		"response_length": len(response.Text),
		"tokens":          response.TokensUsed,
		"processing_time": time.Since(start),
	}).Info("LLM query processed successfully")

	return response, nil
}

// ProcessQueryStream processes a query with streaming response
func (s *EnhancedLLMService) ProcessQueryStream(ctx context.Context, query string) (<-chan *models.LLMStreamChunk, error) {
	ctx, span := s.tracer.Start(ctx, "enhanced_llm.ProcessQueryStream")
	defer span.End()

	start := time.Now()
	s.logger.WithField("query", query[:min(100, len(query))]).Info("Processing streaming LLM query")

	// Check cache first for complete response
	if s.config.SemanticCache && s.semanticCache != nil {
		// TODO: Implement semantic cache lookup for streaming
	}

	// Stream from Ollama
	ollamaReq := &providers.OllamaRequest{
		Model:  s.config.DefaultModel,
		Prompt: query,
		Options: map[string]interface{}{
			"temperature": s.config.Temperature,
			"num_predict": s.config.MaxTokens,
		},
	}

	// Create a channel for streaming chunks
	chunkChan := make(chan *models.LLMStreamChunk, 10)

	// Start streaming in a goroutine
	go func() {
		defer close(chunkChan)

		var fullResponse strings.Builder
		chunkCount := 0

		err := s.ollamaClient.GenerateStream(ctx, ollamaReq, func(chunk *providers.OllamaStreamResponse) error {
			chunkCount++

			streamChunk := &models.LLMStreamChunk{
				ID:        fmt.Sprintf("chunk_%d", chunkCount),
				Content:   chunk.Response,
				Delta:     chunk.Response,
				Finished:  chunk.Done,
				Timestamp: time.Now(),
				Metadata: map[string]interface{}{
					"chunk_number": chunkCount,
					"model":        chunk.Model,
				},
			}

			fullResponse.WriteString(chunk.Response)

			select {
			case chunkChan <- streamChunk:
				return nil
			case <-ctx.Done():
				return ctx.Err()
			}
		})

		if err != nil {
			s.logger.WithError(err).Error("Failed to stream response from Ollama")
		}

		s.logger.WithFields(logrus.Fields{
			"query":           query[:min(100, len(query))],
			"chunks_sent":     chunkCount,
			"response_length": fullResponse.Len(),
			"processing_time": time.Since(start),
		}).Info("Streaming LLM query completed")
	}()

	return chunkChan, nil
}

// ChatWithHistory performs conversational chat with full message history using real LLM
func (s *EnhancedLLMService) ChatWithHistory(ctx context.Context, messages []models.ChatMessage) (*models.ChatResponse, error) {
	ctx, span := s.tracer.Start(ctx, "enhanced_llm.Chat")
	defer span.End()

	start := time.Now()
	s.logger.WithField("message_count", len(messages)).Info("Processing chat conversation")

	// Convert to Ollama format
	ollamaMessages := make([]providers.OllamaMessage, len(messages))
	for i, msg := range messages {
		ollamaMessages[i] = providers.OllamaMessage{
			Role:    msg.Role,
			Content: msg.Content,
		}
	}

	// Generate chat response
	ollamaResp, err := s.ollamaClient.Chat(ctx, s.config.DefaultModel, ollamaMessages, map[string]interface{}{
		"temperature": s.config.Temperature,
		"num_predict": s.config.MaxTokens,
	})

	if err != nil {
		s.logger.WithError(err).Error("Failed to generate chat response with Ollama")
		return nil, fmt.Errorf("ollama chat failed: %w", err)
	}

	// Convert response
	response := &models.ChatResponse{
		Message:        ollamaResp.Response,
		ConversationID: "chat-session",
		Confidence:     0.85, // Default confidence
		Timestamp:      time.Now(),
		Context: map[string]interface{}{
			"model":                ollamaResp.Model,
			"tokens":               ollamaResp.EvalCount,
			"processing_time":      time.Since(start).Milliseconds(),
			"prompt_eval_count":    ollamaResp.PromptEvalCount,
			"prompt_eval_duration": ollamaResp.PromptEvalDuration,
			"eval_duration":        ollamaResp.EvalDuration,
			"total_duration":       ollamaResp.TotalDuration,
		},
	}

	s.logger.WithFields(logrus.Fields{
		"message_count":   len(messages),
		"response_length": len(response.Message),
		"tokens":          response.Context["tokens"],
		"processing_time": time.Since(start),
	}).Info("Chat conversation processed successfully")

	return response, nil
}

// Chat maintains a conversation context
func (s *EnhancedLLMService) Chat(ctx context.Context, message string, conversationID string) (*models.ChatResponse, error) {
	ctx, span := s.tracer.Start(ctx, "enhanced_llm.Chat")
	defer span.End()

	// Convert to message history format
	messages := []models.ChatMessage{
		{
			Role:      "user",
			Content:   message,
			Timestamp: time.Now(),
		},
	}

	// Use ChatWithHistory
	return s.ChatWithHistory(ctx, messages)
}

// ChatStream maintains a conversation context with streaming
func (s *EnhancedLLMService) ChatStream(ctx context.Context, message string, conversationID string) (<-chan *models.ChatStreamChunk, error) {
	ctx, span := s.tracer.Start(ctx, "enhanced_llm.ChatStream")
	defer span.End()

	// Create a channel for streaming chunks
	chunkChan := make(chan *models.ChatStreamChunk, 10)

	// Start streaming in a goroutine
	go func() {
		defer close(chunkChan)

		// Use the existing streaming functionality
		streamChan, err := s.ProcessQueryStream(ctx, message)
		if err != nil {
			s.logger.WithError(err).Error("Failed to start streaming chat response")
			return
		}

		// Convert LLMStreamChunk to ChatStreamChunk
		for chunk := range streamChan {
			chatChunk := &models.ChatStreamChunk{
				ID:             chunk.ID,
				Content:        chunk.Content,
				Delta:          chunk.Delta,
				Finished:       chunk.Finished,
				ConversationID: conversationID,
				Role:           "assistant",
				Timestamp:      chunk.Timestamp,
				Metadata:       chunk.Metadata,
			}

			select {
			case chunkChan <- chatChunk:
			case <-ctx.Done():
				return
			}
		}
	}()

	return chunkChan, nil
}

// GetModelInfo gets information about available models
func (s *EnhancedLLMService) GetModelInfo(ctx context.Context, modelID string) (*models.ModelInfo, error) {
	ctx, span := s.tracer.Start(ctx, "enhanced_llm.GetModelInfo")
	defer span.End()

	s.logger.WithField("model_id", modelID).Info("Getting model information")

	// List models from Ollama
	ollamaModels, err := s.ollamaClient.ListModels(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list models: %w", err)
	}

	// Find the requested model
	for _, model := range ollamaModels {
		if model.Name == modelID || strings.Contains(model.Name, modelID) {
			return &models.ModelInfo{
				ID:           model.Name,
				Name:         model.Name,
				Provider:     "ollama",
				Size:         model.Size,
				Parameters:   0, // Would need to parse from string
				Capabilities: []string{"text-generation", "chat", "completion"},
				Status:       "available",
				CreatedAt:    time.Now(),
				UpdatedAt:    model.ModifiedAt,
				Metadata: map[string]interface{}{
					"digest":             model.Digest,
					"modified_at":        model.ModifiedAt,
					"parameter_size":     model.Details.ParameterSize,
					"quantization_level": model.Details.QuantizationLevel,
					"families":           model.Details.Families,
					"format":             model.Details.Format,
					"family":             model.Details.Family,
				},
			}, nil
		}
	}

	return nil, fmt.Errorf("model not found: %s", modelID)
}

// IsHealthy checks if the LLM service is healthy
func (s *EnhancedLLMService) IsHealthy(ctx context.Context) bool {
	return s.ollamaClient.IsHealthy(ctx)
}

// Helper methods

func (s *EnhancedLLMService) convertCachedResponse(cached interface{}, query string) *models.LLMResponse {
	if response, ok := cached.(*models.LLMResponse); ok {
		// Update metadata to indicate it was cached
		if response.Metadata == nil {
			response.Metadata = make(map[string]interface{})
		}
		response.Metadata["cached"] = true
		response.Metadata["cache_hit_time"] = time.Now()
		return response
	}
	return nil
}

func (s *EnhancedLLMService) streamCachedResponse(cached interface{}, callback func(*models.LLMStreamChunk) error) error {
	if response, ok := cached.(*models.LLMResponse); ok {
		// Simulate streaming by chunking the cached response
		text := response.Text
		chunkSize := 50 // Characters per chunk

		for i := 0; i < len(text); i += chunkSize {
			end := min(i+chunkSize, len(text))
			chunk := text[i:end]

			streamChunk := &models.LLMStreamChunk{
				ID:        fmt.Sprintf("cached_chunk_%d", (i/chunkSize)+1),
				Content:   text[:end], // Full text up to this point
				Delta:     chunk,      // Just this chunk
				Finished:  end >= len(text),
				Timestamp: time.Now(),
				Metadata: map[string]interface{}{
					"cached":       true,
					"chunk_number": (i / chunkSize) + 1,
					"model":        response.Model,
				},
			}

			if err := callback(streamChunk); err != nil {
				return err
			}

			// Small delay to simulate streaming
			time.Sleep(10 * time.Millisecond)
		}
	}
	return nil
}

func hashString(s string) string {
	// Simple hash function for cache keys
	hash := uint32(0)
	for _, c := range s {
		hash = hash*31 + uint32(c)
	}
	return fmt.Sprintf("%x", hash)
}

// AnalyzeText analyzes text content using the LLM
func (s *EnhancedLLMService) AnalyzeText(ctx context.Context, text string) (*models.TextAnalysis, error) {
	ctx, span := s.tracer.Start(ctx, "enhanced_llm.AnalyzeText")
	defer span.End()

	// Create comprehensive analysis prompt
	prompt := fmt.Sprintf(`Analyze the following text and provide:
1. A brief summary
2. Key keywords (comma-separated)
3. Main topics (comma-separated)
4. Sentiment score from -1 (very negative) to 1 (very positive)
5. Language detected

Text to analyze:
%s

Please format your response as:
Summary: [summary]
Keywords: [keyword1, keyword2, ...]
Topics: [topic1, topic2, ...]
Sentiment: [score]
Language: [language]`, text)

	// Process the query
	response, err := s.ProcessQuery(ctx, prompt)
	if err != nil {
		return nil, err
	}

	// Create text analysis response (using correct TextAnalysis fields)
	analysis := &models.TextAnalysis{
		Summary:    response.Text,                                       // Simplified - would need proper parsing
		Keywords:   []string{},                                          // Would extract from response
		Entities:   []models.NamedEntity{},                              // Would extract entities
		Sentiment:  models.SentimentScore{Score: 0.0, Label: "neutral"}, // Would parse sentiment
		Language:   "en",                                                // Default to English
		Complexity: 0.5,                                                 // Default complexity
		Topics:     []string{},                                          // Would extract topics
		Timestamp:  time.Now(),
		Metadata: map[string]interface{}{
			"model":           response.Model,
			"tokens_used":     response.TokensUsed,
			"processing_time": response.ProcessingTime,
			"full_response":   response.Text,
		},
	}

	return analysis, nil
}

// BatchEmbed generates embeddings for multiple texts
func (s *EnhancedLLMService) BatchEmbed(ctx context.Context, texts []string) (*models.BatchEmbeddingResponse, error) {
	ctx, span := s.tracer.Start(ctx, "enhanced_llm.BatchEmbed")
	defer span.End()

	// TODO: Implement actual embedding generation using Ollama or another embedding model
	// For now, return mock embeddings
	embeddings := make([][]float64, len(texts))
	for i := range texts {
		// Generate mock 384-dimensional embeddings
		embedding := make([]float64, 384)
		for j := range embedding {
			embedding[j] = 0.1 // Mock value
		}
		embeddings[i] = embedding
	}

	response := &models.BatchEmbeddingResponse{
		Texts:      texts,
		Embeddings: embeddings,
		Model:      s.config.DefaultModel,
		Dimension:  384,
		Timestamp:  time.Now(),
	}

	s.logger.WithField("text_count", len(texts)).Info("Generated batch embeddings")
	return response, nil
}

// EmbedText generates embeddings for a single text
func (s *EnhancedLLMService) EmbedText(ctx context.Context, text string) (*models.EmbeddingResponse, error) {
	ctx, span := s.tracer.Start(ctx, "enhanced_llm.EmbedText")
	defer span.End()

	// TODO: Implement actual embedding generation using Ollama or another embedding model
	// For now, return mock embeddings
	embedding := make([]float64, 384)
	for i := range embedding {
		embedding[i] = 0.1 // Mock value
	}

	response := &models.EmbeddingResponse{
		Text:      text,
		Embedding: embedding,
		Model:     s.config.DefaultModel,
		Dimension: 384,
		Timestamp: time.Now(),
	}

	s.logger.WithField("text_length", len(text)).Info("Generated text embedding")
	return response, nil
}

// FunctionCall executes a function call request
func (s *EnhancedLLMService) FunctionCall(ctx context.Context, functionName string, parameters map[string]interface{}) (*models.FunctionCallResponse, error) {
	ctx, span := s.tracer.Start(ctx, "enhanced_llm.FunctionCall")
	defer span.End()

	s.logger.WithFields(logrus.Fields{
		"function": functionName,
		"params":   parameters,
	}).Info("Executing function call")

	// TODO: Implement actual function calling
	// For now, return a mock response
	response := &models.FunctionCallResponse{
		ID:        fmt.Sprintf("func_%d", time.Now().UnixNano()),
		Name:      functionName,
		Result:    map[string]interface{}{"status": "success", "message": "Function executed successfully"},
		Success:   true,
		Duration:  time.Millisecond * 100, // Mock duration
		Timestamp: time.Now(),
		Metadata: map[string]interface{}{
			"model":      s.config.DefaultModel,
			"parameters": parameters,
		},
	}

	return response, nil
}

// GenerateCode generates code based on a prompt
func (s *EnhancedLLMService) GenerateCode(ctx context.Context, prompt string) (*models.CodeResponse, error) {
	ctx, span := s.tracer.Start(ctx, "enhanced_llm.GenerateCode")
	defer span.End()

	// Create a code generation prompt
	codePrompt := fmt.Sprintf("Generate code based on the following request. Provide clean, well-commented code:\n\n%s", prompt)

	// Process the query
	response, err := s.ProcessQuery(ctx, codePrompt)
	if err != nil {
		return nil, err
	}

	// Create code response
	codeResponse := &models.CodeResponse{
		Code:        response.Text,
		Language:    "auto", // Would need to detect language
		Explanation: "Generated code based on prompt",
		Confidence:  response.Confidence,
		Suggestions: []string{}, // Would extract suggestions
		Timestamp:   time.Now(),
	}

	return codeResponse, nil
}

// GenerateCodeStream generates code with streaming response
func (s *EnhancedLLMService) GenerateCodeStream(ctx context.Context, prompt string) (<-chan *models.CodeStreamChunk, error) {
	ctx, span := s.tracer.Start(ctx, "enhanced_llm.GenerateCodeStream")
	defer span.End()

	// Create a channel for streaming chunks
	chunkChan := make(chan *models.CodeStreamChunk, 10)

	// Start streaming in a goroutine
	go func() {
		defer close(chunkChan)

		// Create a code generation prompt
		codePrompt := fmt.Sprintf("Generate code based on the following request. Provide clean, well-commented code:\n\n%s", prompt)

		// Use the existing streaming functionality
		streamChan, err := s.ProcessQueryStream(ctx, codePrompt)
		if err != nil {
			s.logger.WithError(err).Error("Failed to start streaming code generation")
			return
		}

		// Convert LLMStreamChunk to CodeStreamChunk
		for chunk := range streamChan {
			codeChunk := &models.CodeStreamChunk{
				ID:        chunk.ID,
				Code:      chunk.Content,
				Delta:     chunk.Delta,
				Language:  "auto", // Would need to detect language
				Finished:  chunk.Finished,
				Timestamp: chunk.Timestamp,
				Metadata:  chunk.Metadata,
			}

			select {
			case chunkChan <- codeChunk:
			case <-ctx.Done():
				return
			}
		}
	}()

	return chunkChan, nil
}

// GetModels returns a list of available models
func (s *EnhancedLLMService) GetModels(ctx context.Context) ([]models.AIModel, error) {
	ctx, span := s.tracer.Start(ctx, "enhanced_llm.GetModels")
	defer span.End()

	s.logger.Info("Getting list of available models")

	// List models from Ollama
	ollamaModels, err := s.ollamaClient.ListModels(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list models: %w", err)
	}

	// Convert to AIModel
	aiModels := make([]models.AIModel, 0, len(ollamaModels))
	for _, model := range ollamaModels {
		aiModel := models.AIModel{
			ID:           model.Name,
			Name:         model.Name,
			Version:      "latest", // Default version
			Type:         "llm",
			Size:         model.Size,
			Description:  fmt.Sprintf("Ollama model: %s", model.Name),
			Capabilities: []string{"text-generation", "chat", "completion"},
			Status:       "available",
			CreatedAt:    time.Now(),
			UpdatedAt:    model.ModifiedAt,
			Metadata: map[string]interface{}{
				"provider":           "ollama",
				"digest":             model.Digest,
				"modified_at":        model.ModifiedAt,
				"parameter_size":     model.Details.ParameterSize,
				"quantization_level": model.Details.QuantizationLevel,
				"families":           model.Details.Families,
				"format":             model.Details.Format,
				"family":             model.Details.Family,
			},
		}
		aiModels = append(aiModels, aiModel)
	}

	s.logger.WithField("model_count", len(aiModels)).Info("Retrieved available models")
	return aiModels, nil
}

// LoadModel loads a specific model
func (s *EnhancedLLMService) LoadModel(ctx context.Context, modelID string) error {
	ctx, span := s.tracer.Start(ctx, "enhanced_llm.LoadModel")
	defer span.End()

	s.logger.WithField("model_id", modelID).Info("Loading model")

	// TODO: Implement actual model loading with Ollama
	// For now, just verify the model exists
	models, err := s.ollamaClient.ListModels(ctx)
	if err != nil {
		return fmt.Errorf("failed to list models: %w", err)
	}

	// Check if model exists
	for _, model := range models {
		if model.Name == modelID {
			s.logger.WithField("model_id", modelID).Info("Model is available")
			return nil
		}
	}

	return fmt.Errorf("model %s not found", modelID)
}

// Summarize creates a summary of the given text
func (s *EnhancedLLMService) Summarize(ctx context.Context, text string) (*models.SummaryResponse, error) {
	ctx, span := s.tracer.Start(ctx, "enhanced_llm.Summarize")
	defer span.End()

	summaryPrompt := fmt.Sprintf("Provide a concise summary of the following text, highlighting the key points:\n\n%s", text)

	llmResponse, err := s.ProcessQuery(ctx, summaryPrompt)
	if err != nil {
		return nil, fmt.Errorf("failed to summarize text: %w", err)
	}

	originalLength := len(text)
	summaryLength := len(llmResponse.Text)
	compressionRatio := float64(summaryLength) / float64(originalLength)

	summaryResponse := &models.SummaryResponse{
		Summary:     llmResponse.Text,
		KeyPoints:   []string{}, // TODO: Extract key points
		Length:      summaryLength,
		Compression: compressionRatio,
		Timestamp:   time.Now(),
	}

	return summaryResponse, nil
}

// Translate translates text between languages
func (s *EnhancedLLMService) Translate(ctx context.Context, text, fromLang, toLang string) (*models.TranslationResponse, error) {
	ctx, span := s.tracer.Start(ctx, "enhanced_llm.Translate")
	defer span.End()

	translatePrompt := fmt.Sprintf("Translate the following text from %s to %s:\n\n%s", fromLang, toLang, text)

	llmResponse, err := s.ProcessQuery(ctx, translatePrompt)
	if err != nil {
		return nil, fmt.Errorf("failed to translate text: %w", err)
	}

	translationResponse := &models.TranslationResponse{
		TranslatedText: llmResponse.Text,
		FromLanguage:   fromLang,
		ToLanguage:     toLang,
		Confidence:     llmResponse.Confidence,
		Timestamp:      time.Now(),
	}

	return translationResponse, nil
}

// UnloadModel unloads a specific model
func (s *EnhancedLLMService) UnloadModel(ctx context.Context, modelID string) error {
	ctx, span := s.tracer.Start(ctx, "enhanced_llm.UnloadModel")
	defer span.End()

	s.logger.WithField("model_id", modelID).Info("Unloading model")

	// TODO: Implement actual model unloading with Ollama
	// For now, just log the action
	s.logger.WithField("model_id", modelID).Info("Model unloaded (placeholder implementation)")
	return nil
}
