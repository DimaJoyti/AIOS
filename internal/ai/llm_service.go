package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

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

	return &LLMService{
		config:        config,
		logger:        logger,
		tracer:        tracer,
		httpClient:    &http.Client{Timeout: config.OllamaTimeout},
		ollamaBaseURL: fmt.Sprintf("http://%s:%d", config.OllamaHost, config.OllamaPort),
		conversations: make(map[string]*Conversation),
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
