package ai

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/aios/aios/internal/ai/cache"
	"github.com/aios/aios/pkg/models"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

// LLMService implements the LanguageModelService interface using Ollama
type LLMService struct {
	config        AIServiceConfig
	logger        *logrus.Logger
	tracer        trace.Tracer
	httpClient    *http.Client
	ollamaBaseURL string
	conversations map[string]*Conversation
	cache         *cache.MemoryCacheManager
	semanticCache *cache.SemanticCacheManager
	mu            sync.RWMutex
}

// Conversation represents a conversation context
type Conversation struct {
	ID       string                 `json:"id"`
	Messages []ConversationMessage  `json:"messages"`
	Context  map[string]interface{} `json:"context"`
	Created  time.Time              `json:"created"`
	Updated  time.Time              `json:"updated"`
}

// ConversationMessage represents a message in a conversation
type ConversationMessage struct {
	Role      string    `json:"role"` // user, assistant, system
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

// OllamaRequest represents a request to Ollama API
type OllamaRequest struct {
	Model   string                 `json:"model"`
	Prompt  string                 `json:"prompt"`
	Stream  bool                   `json:"stream"`
	Options map[string]interface{} `json:"options,omitempty"`
	Context []int                  `json:"context,omitempty"`
}

// OllamaResponse represents a response from Ollama API
type OllamaResponse struct {
	Model           string `json:"model"`
	Response        string `json:"response"`
	Done            bool   `json:"done"`
	Context         []int  `json:"context,omitempty"`
	TotalDuration   int64  `json:"total_duration,omitempty"`
	LoadDuration    int64  `json:"load_duration,omitempty"`
	PromptEvalCount int    `json:"prompt_eval_count,omitempty"`
	EvalCount       int    `json:"eval_count,omitempty"`
	EvalDuration    int64  `json:"eval_duration,omitempty"`
}

// NewLLMService creates a new language model service
func NewLLMService(config AIServiceConfig, logger *logrus.Logger) *LLMService {
	tracer := otel.Tracer("llm-service")

	// Initialize caching if enabled
	var memCache *cache.MemoryCacheManager
	var semCache *cache.SemanticCacheManager

	if config.CacheEnabled {
		memCache = cache.NewMemoryCacheManager(
			config.CacheMaxSize,
			config.CacheTTL,
			logger,
		)

		if config.SemanticCache {
			semCache = cache.NewSemanticCacheManager(
				int(config.CacheMaxSize/2), // Use half for semantic cache
				config.CacheTTL,
				0.8, // 80% similarity threshold
				logger,
			)
		}
	}

	return &LLMService{
		config:        config,
		logger:        logger,
		tracer:        tracer,
		httpClient:    &http.Client{Timeout: config.OllamaTimeout},
		ollamaBaseURL: fmt.Sprintf("http://%s:%d", config.OllamaHost, config.OllamaPort),
		conversations: make(map[string]*Conversation),
		cache:         memCache,
		semanticCache: semCache,
	}
}

// ProcessQuery processes a natural language query and returns a response
func (s *LLMService) ProcessQuery(ctx context.Context, query string) (*models.LLMResponse, error) {
	ctx, span := s.tracer.Start(ctx, "llm.ProcessQuery")
	defer span.End()

	start := time.Now()

	s.logger.WithField("query", query).Info("Processing LLM query")

	// Prepare Ollama request
	request := OllamaRequest{
		Model:  s.config.DefaultModel,
		Prompt: query,
		Stream: false,
		Options: map[string]interface{}{
			"temperature": s.config.Temperature,
			"num_predict": s.config.MaxTokens,
		},
	}

	// Make request to Ollama
	response, err := s.makeOllamaRequest(ctx, "/api/generate", request)
	if err != nil {
		return nil, fmt.Errorf("failed to make Ollama request: %w", err)
	}

	processingTime := time.Since(start)

	llmResponse := &models.LLMResponse{
		Text:           response.Response,
		Confidence:     0.8, // TODO: Calculate actual confidence
		TokensUsed:     response.EvalCount + response.PromptEvalCount,
		Model:          response.Model,
		ProcessingTime: processingTime,
		Metadata: map[string]interface{}{
			"total_duration":    response.TotalDuration,
			"load_duration":     response.LoadDuration,
			"prompt_eval_count": response.PromptEvalCount,
			"eval_count":        response.EvalCount,
			"eval_duration":     response.EvalDuration,
		},
		Timestamp: time.Now(),
	}

	s.logger.WithFields(logrus.Fields{
		"tokens_used":     llmResponse.TokensUsed,
		"processing_time": processingTime,
		"model":           response.Model,
	}).Info("LLM query processed successfully")

	return llmResponse, nil
}

// GenerateCode generates code based on a prompt
func (s *LLMService) GenerateCode(ctx context.Context, prompt string) (*models.CodeResponse, error) {
	ctx, span := s.tracer.Start(ctx, "llm.GenerateCode")
	defer span.End()

	// Enhance prompt for code generation
	codePrompt := fmt.Sprintf("Generate code for the following request. Provide clean, well-commented code:\n\n%s", prompt)

	response, err := s.ProcessQuery(ctx, codePrompt)
	if err != nil {
		return nil, fmt.Errorf("failed to generate code: %w", err)
	}

	// TODO: Parse language from response or prompt
	language := "unknown"

	codeResponse := &models.CodeResponse{
		Code:        response.Text,
		Language:    language,
		Explanation: "Generated code based on the provided prompt",
		Confidence:  response.Confidence,
		Suggestions: []string{
			"Review the generated code for correctness",
			"Test the code before using in production",
			"Add appropriate error handling",
		},
		Timestamp: time.Now(),
	}

	return codeResponse, nil
}

// AnalyzeText analyzes text for various insights
func (s *LLMService) AnalyzeText(ctx context.Context, text string) (*models.TextAnalysis, error) {
	ctx, span := s.tracer.Start(ctx, "llm.AnalyzeText")
	defer span.End()

	analysisPrompt := fmt.Sprintf(`Analyze the following text and provide:
1. A brief summary
2. Key keywords (comma-separated)
3. Main topics (comma-separated)
4. Sentiment (positive/negative/neutral)
5. Complexity level (1-10)

Text to analyze:
%s

Please format your response as JSON with fields: summary, keywords, topics, sentiment, complexity`, text)

	_, err := s.ProcessQuery(ctx, analysisPrompt)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze text: %w", err)
	}

	// TODO: Parse structured response from LLM
	analysis := &models.TextAnalysis{
		Summary:    "Text analysis summary",          // TODO: Extract from LLM response
		Keywords:   []string{"keyword1", "keyword2"}, // TODO: Extract from LLM response
		Entities:   []models.NamedEntity{},           // TODO: Implement entity extraction
		Sentiment:  models.SentimentScore{Score: 0.0, Label: "neutral", Confidence: 0.8},
		Language:   "en",                         // TODO: Detect language
		Complexity: 5.0,                          // TODO: Extract from LLM response
		Topics:     []string{"topic1", "topic2"}, // TODO: Extract from LLM response
		Metadata: map[string]interface{}{
			"original_text_length": len(text),
			"analysis_model":       s.config.DefaultModel,
		},
		Timestamp: time.Now(),
	}

	return analysis, nil
}

// Chat maintains a conversation context
func (s *LLMService) Chat(ctx context.Context, message string, conversationID string) (*models.ChatResponse, error) {
	ctx, span := s.tracer.Start(ctx, "llm.Chat")
	defer span.End()

	// Get or create conversation
	conversation := s.getOrCreateConversation(conversationID)

	// Add user message to conversation
	conversation.Messages = append(conversation.Messages, ConversationMessage{
		Role:      "user",
		Content:   message,
		Timestamp: time.Now(),
	})

	// Build context from conversation history
	contextPrompt := s.buildConversationContext(conversation)

	// Process with LLM
	llmResponse, err := s.ProcessQuery(ctx, contextPrompt)
	if err != nil {
		return nil, fmt.Errorf("failed to process chat message: %w", err)
	}

	// Add assistant response to conversation
	conversation.Messages = append(conversation.Messages, ConversationMessage{
		Role:      "assistant",
		Content:   llmResponse.Text,
		Timestamp: time.Now(),
	})
	conversation.Updated = time.Now()

	chatResponse := &models.ChatResponse{
		Message:        llmResponse.Text,
		ConversationID: conversationID,
		Context:        conversation.Context,
		Suggestions:    []string{},                  // TODO: Generate suggestions
		Actions:        []models.ActionSuggestion{}, // TODO: Generate action suggestions
		Confidence:     llmResponse.Confidence,
		Timestamp:      time.Now(),
	}

	return chatResponse, nil
}

// Summarize creates a summary of the given text
func (s *LLMService) Summarize(ctx context.Context, text string) (*models.SummaryResponse, error) {
	ctx, span := s.tracer.Start(ctx, "llm.Summarize")
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
func (s *LLMService) Translate(ctx context.Context, text, fromLang, toLang string) (*models.TranslationResponse, error) {
	ctx, span := s.tracer.Start(ctx, "llm.Translate")
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

// GetModels returns available language models
func (s *LLMService) GetModels(ctx context.Context) ([]models.AIModel, error) {
	ctx, span := s.tracer.Start(ctx, "llm.GetModels")
	defer span.End()

	// TODO: Implement actual model listing from Ollama
	models := []models.AIModel{
		{
			ID:           "llama2",
			Name:         "Llama 2",
			Version:      "7b",
			Type:         "llm",
			Size:         3800000000, // ~3.8GB
			Description:  "Llama 2 7B parameter model",
			Capabilities: []string{"text-generation", "chat", "code"},
			Status:       "available",
			CreatedAt:    time.Now().Add(-24 * time.Hour),
			UpdatedAt:    time.Now(),
		},
	}

	return models, nil
}

// LoadModel loads a specific model
func (s *LLMService) LoadModel(ctx context.Context, modelName string) error {
	ctx, span := s.tracer.Start(ctx, "llm.LoadModel")
	defer span.End()

	s.logger.WithField("model", modelName).Info("Loading model")

	// TODO: Implement actual model loading via Ollama API
	// For now, just update the default model
	s.config.DefaultModel = modelName

	return nil
}

// UnloadModel unloads a specific model
func (s *LLMService) UnloadModel(ctx context.Context, modelName string) error {
	ctx, span := s.tracer.Start(ctx, "llm.UnloadModel")
	defer span.End()

	s.logger.WithField("model", modelName).Info("Unloading model")

	// TODO: Implement actual model unloading via Ollama API
	return nil
}

// Helper methods

func (s *LLMService) makeOllamaRequest(ctx context.Context, endpoint string, request OllamaRequest) (*OllamaResponse, error) {
	requestBody, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := s.ollamaBaseURL + endpoint
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to make HTTP request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Ollama API returned status %d", resp.StatusCode)
	}

	var ollamaResponse OllamaResponse
	if err := json.NewDecoder(resp.Body).Decode(&ollamaResponse); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &ollamaResponse, nil
}

func (s *LLMService) getOrCreateConversation(conversationID string) *Conversation {
	if conversation, exists := s.conversations[conversationID]; exists {
		return conversation
	}

	conversation := &Conversation{
		ID:       conversationID,
		Messages: []ConversationMessage{},
		Context:  make(map[string]interface{}),
		Created:  time.Now(),
		Updated:  time.Now(),
	}

	s.conversations[conversationID] = conversation
	return conversation
}

func (s *LLMService) buildConversationContext(conversation *Conversation) string {
	var contextBuilder bytes.Buffer

	contextBuilder.WriteString("You are a helpful AI assistant. Here is the conversation history:\n\n")

	for _, message := range conversation.Messages {
		contextBuilder.WriteString(fmt.Sprintf("%s: %s\n", message.Role, message.Content))
	}

	contextBuilder.WriteString("\nPlease respond to the latest user message:")

	return contextBuilder.String()
}

// ProcessQueryStream processes a query with streaming response
func (s *LLMService) ProcessQueryStream(ctx context.Context, query string) (<-chan *models.LLMStreamChunk, error) {
	ctx, span := s.tracer.Start(ctx, "llm.ProcessQueryStream")
	defer span.End()

	s.logger.WithField("query", query).Info("Processing streaming query")

	// Create response channel
	responseChan := make(chan *models.LLMStreamChunk, 10)

	// Start streaming in a goroutine
	go func() {
		defer close(responseChan)

		// Check cache first
		if s.cache != nil {
			cacheKey := cache.GenerateKey("llm_query", query, s.config.DefaultModel)
			if cached, found, err := s.cache.Get(ctx, cacheKey); err == nil && found {
				if cachedResponse, ok := cached.(*models.LLMResponse); ok {
					// Send cached response as a single chunk
					chunk := &models.LLMStreamChunk{
						ID:        fmt.Sprintf("cached_%d", time.Now().UnixNano()),
						Content:   cachedResponse.Text,
						Delta:     cachedResponse.Text,
						Finished:  true,
						Metadata:  map[string]interface{}{"cached": true},
						Timestamp: time.Now(),
					}
					responseChan <- chunk
					return
				}
			}
		}

		// Create streaming request
		ollamaReq := OllamaRequest{
			Model:  s.config.DefaultModel,
			Prompt: query,
			Stream: true,
			Options: map[string]interface{}{
				"temperature": s.config.Temperature,
				"num_predict": s.config.MaxTokens,
			},
		}

		reqBody, err := json.Marshal(ollamaReq)
		if err != nil {
			s.logger.WithError(err).Error("Failed to marshal streaming request")
			return
		}

		// Make streaming request
		req, err := http.NewRequestWithContext(ctx, "POST", s.ollamaBaseURL+"/api/generate", bytes.NewBuffer(reqBody))
		if err != nil {
			s.logger.WithError(err).Error("Failed to create streaming request")
			return
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := s.httpClient.Do(req)
		if err != nil {
			s.logger.WithError(err).Error("Failed to make streaming request")
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			s.logger.WithField("status", resp.StatusCode).Error("Streaming request failed")
			return
		}

		// Process streaming response
		scanner := bufio.NewScanner(resp.Body)
		var fullResponse strings.Builder
		chunkID := fmt.Sprintf("stream_%d", time.Now().UnixNano())

		for scanner.Scan() {
			line := scanner.Text()
			if line == "" {
				continue
			}

			var ollamaChunk OllamaResponse
			if err := json.Unmarshal([]byte(line), &ollamaChunk); err != nil {
				s.logger.WithError(err).Error("Failed to parse streaming chunk")
				continue
			}

			// Create stream chunk
			chunk := &models.LLMStreamChunk{
				ID:       chunkID,
				Content:  fullResponse.String() + ollamaChunk.Response,
				Delta:    ollamaChunk.Response,
				Finished: ollamaChunk.Done,
				Metadata: map[string]interface{}{
					"model": s.config.DefaultModel,
				},
				Timestamp: time.Now(),
			}

			fullResponse.WriteString(ollamaChunk.Response)
			responseChan <- chunk

			if ollamaChunk.Done {
				// Cache the complete response
				if s.cache != nil {
					cacheKey := cache.GenerateKey("llm_query", query, s.config.DefaultModel)
					llmResponse := &models.LLMResponse{
						Text:      fullResponse.String(),
						Model:     s.config.DefaultModel,
						Timestamp: time.Now(),
					}
					s.cache.Set(ctx, cacheKey, llmResponse, s.config.CacheTTL)
				}
				break
			}
		}

		if err := scanner.Err(); err != nil {
			s.logger.WithError(err).Error("Error reading streaming response")
		}
	}()

	return responseChan, nil
}

// GenerateCodeStream generates code with streaming response
func (s *LLMService) GenerateCodeStream(ctx context.Context, prompt string) (<-chan *models.CodeStreamChunk, error) {
	ctx, span := s.tracer.Start(ctx, "llm.GenerateCodeStream")
	defer span.End()

	s.logger.WithField("prompt", prompt).Info("Generating code with streaming")

	// Create response channel
	responseChan := make(chan *models.CodeStreamChunk, 10)

	// Start streaming in a goroutine
	go func() {
		defer close(responseChan)

		// Enhance prompt for code generation
		codePrompt := fmt.Sprintf("Generate code for the following request. Provide clean, well-commented code:\n\n%s", prompt)

		// Use the streaming query method
		llmStream, err := s.ProcessQueryStream(ctx, codePrompt)
		if err != nil {
			s.logger.WithError(err).Error("Failed to start code generation stream")
			return
		}

		chunkID := fmt.Sprintf("code_%d", time.Now().UnixNano())

		for chunk := range llmStream {
			// Detect programming language (simple heuristic)
			language := "text"
			if strings.Contains(chunk.Content, "func ") || strings.Contains(chunk.Content, "package ") {
				language = "go"
			} else if strings.Contains(chunk.Content, "def ") || strings.Contains(chunk.Content, "import ") {
				language = "python"
			} else if strings.Contains(chunk.Content, "function ") || strings.Contains(chunk.Content, "const ") {
				language = "javascript"
			}

			codeChunk := &models.CodeStreamChunk{
				ID:          chunkID,
				Code:        chunk.Content,
				Delta:       chunk.Delta,
				Language:    language,
				Finished:    chunk.Finished,
				Explanation: "Generated code based on the provided prompt",
				Metadata:    chunk.Metadata,
				Timestamp:   chunk.Timestamp,
			}

			responseChan <- codeChunk
		}
	}()

	return responseChan, nil
}

// ChatStream maintains a conversation context with streaming
func (s *LLMService) ChatStream(ctx context.Context, message string, conversationID string) (<-chan *models.ChatStreamChunk, error) {
	ctx, span := s.tracer.Start(ctx, "llm.ChatStream")
	defer span.End()

	s.logger.WithFields(logrus.Fields{
		"message":         message,
		"conversation_id": conversationID,
	}).Info("Processing streaming chat")

	// Create response channel
	responseChan := make(chan *models.ChatStreamChunk, 10)

	// Start streaming in a goroutine
	go func() {
		defer close(responseChan)

		s.mu.Lock()
		conversation := s.getOrCreateConversation(conversationID)
		s.mu.Unlock()

		// Add user message to conversation
		conversation.Messages = append(conversation.Messages, ConversationMessage{
			Role:      "user",
			Content:   message,
			Timestamp: time.Now(),
		})

		// Build context with conversation history
		contextPrompt := s.buildConversationContext(conversation)

		// Use the streaming query method
		llmStream, err := s.ProcessQueryStream(ctx, contextPrompt)
		if err != nil {
			s.logger.WithError(err).Error("Failed to start chat stream")
			return
		}

		chunkID := fmt.Sprintf("chat_%d", time.Now().UnixNano())
		var fullResponse strings.Builder

		for chunk := range llmStream {
			chatChunk := &models.ChatStreamChunk{
				ID:             chunkID,
				ConversationID: conversationID,
				Content:        chunk.Content,
				Delta:          chunk.Delta,
				Role:           "assistant",
				Finished:       chunk.Finished,
				Metadata:       chunk.Metadata,
				Timestamp:      chunk.Timestamp,
			}

			fullResponse.WriteString(chunk.Delta)
			responseChan <- chatChunk

			if chunk.Finished {
				// Add assistant response to conversation
				s.mu.Lock()
				conversation.Messages = append(conversation.Messages, ConversationMessage{
					Role:      "assistant",
					Content:   fullResponse.String(),
					Timestamp: time.Now(),
				})
				conversation.Updated = time.Now()
				s.mu.Unlock()
			}
		}
	}()

	return responseChan, nil
}

// FunctionCall executes a function call based on the model's response
func (s *LLMService) FunctionCall(ctx context.Context, functionName string, parameters map[string]any) (*models.FunctionCallResponse, error) {
	ctx, span := s.tracer.Start(ctx, "llm.FunctionCall")
	defer span.End()

	start := time.Now()
	s.logger.WithFields(logrus.Fields{
		"function": functionName,
		"params":   parameters,
	}).Info("Executing function call")

	// TODO: Implement actual function calling
	// This would involve:
	// 1. Function registry lookup
	// 2. Parameter validation
	// 3. Function execution
	// 4. Result formatting

	// Mock implementation
	response := &models.FunctionCallResponse{
		ID:       fmt.Sprintf("func_%d", time.Now().UnixNano()),
		Name:     functionName,
		Result:   fmt.Sprintf("Function %s executed with parameters: %v", functionName, parameters),
		Success:  true,
		Duration: time.Since(start),
		Metadata: map[string]interface{}{
			"model": s.config.DefaultModel,
		},
		Timestamp: time.Now(),
	}

	s.logger.WithFields(logrus.Fields{
		"function":       functionName,
		"success":        response.Success,
		"execution_time": response.Duration,
	}).Info("Function call completed")

	return response, nil
}

// EmbedText generates embeddings for text
func (s *LLMService) EmbedText(ctx context.Context, text string) (*models.EmbeddingResponse, error) {
	ctx, span := s.tracer.Start(ctx, "llm.EmbedText")
	defer span.End()

	start := time.Now()
	s.logger.WithField("text_length", len(text)).Info("Generating text embedding")

	// TODO: Implement actual text embedding using embedding models
	// This would involve:
	// 1. Loading embedding model
	// 2. Text preprocessing
	// 3. Generating embeddings
	// 4. Normalizing vectors

	// Mock implementation - generate random embedding
	dimension := 768 // Common embedding dimension
	embedding := make([]float64, dimension)
	for i := range embedding {
		embedding[i] = float64(i%100) / 100.0 // Simple pattern for demo
	}

	response := &models.EmbeddingResponse{
		Text:      text,
		Embedding: embedding,
		Model:     "text-embedding-model",
		Dimension: dimension,
		Timestamp: time.Now(),
	}

	s.logger.WithFields(logrus.Fields{
		"text_length":     len(text),
		"embedding_dim":   dimension,
		"processing_time": time.Since(start),
	}).Info("Text embedding completed")

	return response, nil
}

// BatchEmbed generates embeddings for multiple texts
func (s *LLMService) BatchEmbed(ctx context.Context, texts []string) (*models.BatchEmbeddingResponse, error) {
	ctx, span := s.tracer.Start(ctx, "llm.BatchEmbed")
	defer span.End()

	start := time.Now()
	s.logger.WithField("batch_size", len(texts)).Info("Generating batch embeddings")

	// TODO: Implement actual batch embedding
	// This would involve:
	// 1. Batch processing optimization
	// 2. Parallel embedding generation
	// 3. Memory management for large batches

	// Mock implementation
	dimension := 768
	embeddings := make([][]float64, len(texts))

	for i := range texts {
		embedding := make([]float64, dimension)
		for j := range embedding {
			embedding[j] = float64((i+j)%100) / 100.0 // Simple pattern for demo
		}
		embeddings[i] = embedding
	}

	response := &models.BatchEmbeddingResponse{
		Texts:      texts,
		Embeddings: embeddings,
		Model:      "text-embedding-model",
		Dimension:  dimension,
		Timestamp:  time.Now(),
	}

	s.logger.WithFields(logrus.Fields{
		"batch_size":      len(texts),
		"embedding_dim":   dimension,
		"processing_time": time.Since(start),
	}).Info("Batch embedding completed")

	return response, nil
}

// GetModelInfo gets detailed information about a model
func (s *LLMService) GetModelInfo(ctx context.Context, modelName string) (*models.ModelInfo, error) {
	ctx, span := s.tracer.Start(ctx, "llm.GetModelInfo")
	defer span.End()

	s.logger.WithField("model", modelName).Info("Getting model information")

	// TODO: Implement actual model info retrieval from Ollama
	// This would involve:
	// 1. Querying Ollama API for model details
	// 2. Parsing model metadata
	// 3. Calculating model statistics

	// Mock implementation
	info := &models.ModelInfo{
		ID:           modelName,
		Name:         modelName,
		Version:      "latest",
		Type:         "llm",
		Provider:     "ollama",
		Size:         3800000000, // 3.8GB
		Parameters:   7000000000, // 7B parameters
		Description:  fmt.Sprintf("Language model: %s", modelName),
		Capabilities: []string{"text-generation", "chat", "code", "summarization"},
		Languages:    []string{"en", "es", "fr", "de", "it"},
		MaxTokens:    s.config.MaxTokens,
		ContextSize:  4096,
		Precision:    "fp16",
		Hardware:     []string{"cpu", "gpu"},
		License:      "custom",
		Status:       "available",
		MemoryUsage:  3800000000,
		CreatedAt:    time.Now().Add(-30 * 24 * time.Hour),
		UpdatedAt:    time.Now(),
	}

	s.logger.WithFields(logrus.Fields{
		"model":      modelName,
		"size":       info.Size,
		"parameters": info.Parameters,
	}).Info("Model information retrieved")

	return info, nil
}

// ChatWithHistory maintains a conversation with full message history
func (s *LLMService) ChatWithHistory(ctx context.Context, messages []models.ChatMessage) (*models.ChatResponse, error) {
	ctx, span := s.tracer.Start(ctx, "llm_service.ChatWithHistory")
	defer span.End()

	if len(messages) == 0 {
		return nil, fmt.Errorf("no messages provided")
	}

	// Convert messages to a single prompt
	var promptBuilder strings.Builder
	for _, msg := range messages {
		switch msg.Role {
		case "system":
			promptBuilder.WriteString(fmt.Sprintf("System: %s\n", msg.Content))
		case "user":
			promptBuilder.WriteString(fmt.Sprintf("User: %s\n", msg.Content))
		case "assistant":
			promptBuilder.WriteString(fmt.Sprintf("Assistant: %s\n", msg.Content))
		}
	}
	promptBuilder.WriteString("Assistant: ")

	prompt := promptBuilder.String()

	// Use the existing ProcessQuery method
	queryResponse, err := s.ProcessQuery(ctx, prompt)
	if err != nil {
		return nil, err
	}

	// Convert to ChatResponse
	response := &models.ChatResponse{
		Message:        queryResponse.Text,
		ConversationID: "history-based",
		Confidence:     queryResponse.Confidence,
		Timestamp:      time.Now(),
	}

	return response, nil
}
