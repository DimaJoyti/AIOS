package services

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

// ModelManager manages AI models and their configurations
type ModelManager struct {
	models       map[string]*AIModel
	providers    map[string]ModelProvider
	cache        *ModelCache
	loadBalancer *LoadBalancer
	monitor      *ModelMonitor
	logger       *logrus.Logger
	tracer       trace.Tracer
	mutex        sync.RWMutex
}

// AIModel represents an AI model configuration
type AIModel struct {
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	Provider     string                 `json:"provider"`
	Type         string                 `json:"type"` // text, image, audio, video, multimodal
	Version      string                 `json:"version"`
	Capabilities []string               `json:"capabilities"`
	Config       ModelConfig            `json:"config"`
	Limits       ModelLimits            `json:"limits"`
	Pricing      ModelPricing           `json:"pricing"`
	Status       string                 `json:"status"` // active, inactive, deprecated
	Metadata     map[string]interface{} `json:"metadata"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
}

// ModelConfig holds model-specific configuration
type ModelConfig struct {
	MaxTokens        int                    `json:"max_tokens"`
	Temperature      float64                `json:"temperature"`
	TopP             float64                `json:"top_p"`
	FrequencyPenalty float64                `json:"frequency_penalty"`
	PresencePenalty  float64                `json:"presence_penalty"`
	StopSequences    []string               `json:"stop_sequences"`
	SystemPrompt     string                 `json:"system_prompt"`
	Parameters       map[string]interface{} `json:"parameters"`
}

// ModelLimits defines usage limits for a model
type ModelLimits struct {
	RequestsPerMinute int     `json:"requests_per_minute"`
	TokensPerMinute   int     `json:"tokens_per_minute"`
	RequestsPerDay    int     `json:"requests_per_day"`
	MaxConcurrency    int     `json:"max_concurrency"`
	TimeoutSeconds    int     `json:"timeout_seconds"`
	MaxRetries        int     `json:"max_retries"`
	CostLimit         float64 `json:"cost_limit"`
}

// ModelPricing defines pricing information for a model
type ModelPricing struct {
	InputTokenCost  float64 `json:"input_token_cost"`
	OutputTokenCost float64 `json:"output_token_cost"`
	RequestCost     float64 `json:"request_cost"`
	Currency        string  `json:"currency"`
}

// ModelProvider defines the interface for AI model providers
type ModelProvider interface {
	GetName() string
	GetModels() []string
	IsAvailable() bool
	GenerateText(ctx context.Context, request *TextGenerationRequest) (*TextGenerationResponse, error)
	GenerateImage(ctx context.Context, request *ImageGenerationRequest) (*ImageGenerationResponse, error)
	ProcessAudio(ctx context.Context, request *AudioProcessingRequest) (*AudioProcessingResponse, error)
	ProcessVideo(ctx context.Context, request *VideoProcessingRequest) (*VideoProcessingResponse, error)
	GetUsage() *ProviderUsage
	GetHealth() *ProviderHealth
}

// TextGenerationRequest represents a text generation request
type TextGenerationRequest struct {
	ModelID      string                 `json:"model_id"`
	Prompt       string                 `json:"prompt"`
	SystemPrompt string                 `json:"system_prompt,omitempty"`
	Messages     []ChatMessage          `json:"messages,omitempty"`
	Config       ModelConfig            `json:"config"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// TextGenerationResponse represents a text generation response
type TextGenerationResponse struct {
	Text         string                 `json:"text"`
	FinishReason string                 `json:"finish_reason"`
	Usage        TokenUsage             `json:"usage"`
	ModelID      string                 `json:"model_id"`
	Provider     string                 `json:"provider"`
	Latency      time.Duration          `json:"latency"`
	Cost         float64                `json:"cost"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// ChatMessage represents a chat message
type ChatMessage struct {
	Role    string `json:"role"` // system, user, assistant
	Content string `json:"content"`
}

// TokenUsage represents token usage information
type TokenUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// ImageGenerationRequest represents an image generation request
type ImageGenerationRequest struct {
	ModelID        string                 `json:"model_id"`
	Prompt         string                 `json:"prompt"`
	NegativePrompt string                 `json:"negative_prompt,omitempty"`
	Width          int                    `json:"width"`
	Height         int                    `json:"height"`
	Steps          int                    `json:"steps"`
	Guidance       float64                `json:"guidance"`
	Seed           int64                  `json:"seed,omitempty"`
	Count          int                    `json:"count"`
	Format         string                 `json:"format"` // png, jpg, webp
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

// ImageGenerationResponse represents an image generation response
type ImageGenerationResponse struct {
	Images   []GeneratedImage       `json:"images"`
	ModelID  string                 `json:"model_id"`
	Provider string                 `json:"provider"`
	Latency  time.Duration          `json:"latency"`
	Cost     float64                `json:"cost"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// GeneratedImage represents a generated image
type GeneratedImage struct {
	URL      string                 `json:"url,omitempty"`
	Base64   string                 `json:"base64,omitempty"`
	Width    int                    `json:"width"`
	Height   int                    `json:"height"`
	Format   string                 `json:"format"`
	Seed     int64                  `json:"seed,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// AudioProcessingRequest represents an audio processing request
type AudioProcessingRequest struct {
	ModelID   string                 `json:"model_id"`
	AudioData []byte                 `json:"audio_data"`
	Format    string                 `json:"format"` // wav, mp3, flac
	Task      string                 `json:"task"`   // transcribe, translate, generate
	Language  string                 `json:"language,omitempty"`
	Prompt    string                 `json:"prompt,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// AudioProcessingResponse represents an audio processing response
type AudioProcessingResponse struct {
	Text     string                 `json:"text,omitempty"`
	AudioURL string                 `json:"audio_url,omitempty"`
	Duration time.Duration          `json:"duration"`
	ModelID  string                 `json:"model_id"`
	Provider string                 `json:"provider"`
	Latency  time.Duration          `json:"latency"`
	Cost     float64                `json:"cost"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// VideoProcessingRequest represents a video processing request
type VideoProcessingRequest struct {
	ModelID   string                 `json:"model_id"`
	VideoData []byte                 `json:"video_data"`
	Format    string                 `json:"format"` // mp4, avi, mov
	Task      string                 `json:"task"`   // analyze, caption, generate
	Prompt    string                 `json:"prompt,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// VideoProcessingResponse represents a video processing response
type VideoProcessingResponse struct {
	Text     string                 `json:"text,omitempty"`
	VideoURL string                 `json:"video_url,omitempty"`
	Duration time.Duration          `json:"duration"`
	ModelID  string                 `json:"model_id"`
	Provider string                 `json:"provider"`
	Latency  time.Duration          `json:"latency"`
	Cost     float64                `json:"cost"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// ProviderUsage represents provider usage statistics
type ProviderUsage struct {
	RequestsToday   int64         `json:"requests_today"`
	TokensToday     int64         `json:"tokens_today"`
	CostToday       float64       `json:"cost_today"`
	RequestsPerMin  int           `json:"requests_per_min"`
	TokensPerMin    int           `json:"tokens_per_min"`
	AverageLatency  time.Duration `json:"average_latency"`
	ErrorRate       float64       `json:"error_rate"`
	LastRequestTime time.Time     `json:"last_request_time"`
}

// ProviderHealth represents provider health status
type ProviderHealth struct {
	Status    string        `json:"status"` // healthy, degraded, unhealthy
	Latency   time.Duration `json:"latency"`
	ErrorRate float64       `json:"error_rate"`
	LastCheck time.Time     `json:"last_check"`
	Issues    []string      `json:"issues,omitempty"`
}

// NewModelManager creates a new model manager
func NewModelManager(logger *logrus.Logger) *ModelManager {
	return &ModelManager{
		models:       make(map[string]*AIModel),
		providers:    make(map[string]ModelProvider),
		cache:        NewModelCache(logger),
		loadBalancer: NewLoadBalancer(logger),
		monitor:      NewModelMonitor(logger),
		logger:       logger,
		tracer:       otel.Tracer("ai.services.model_manager"),
	}
}

// RegisterModel registers a new AI model
func (mm *ModelManager) RegisterModel(model *AIModel) error {
	mm.mutex.Lock()
	defer mm.mutex.Unlock()

	model.CreatedAt = time.Now()
	model.UpdatedAt = time.Now()
	mm.models[model.ID] = model

	mm.logger.WithFields(logrus.Fields{
		"model_id": model.ID,
		"provider": model.Provider,
		"type":     model.Type,
	}).Info("AI model registered")

	return nil
}

// RegisterProvider registers a new model provider
func (mm *ModelManager) RegisterProvider(provider ModelProvider) error {
	mm.mutex.Lock()
	defer mm.mutex.Unlock()

	mm.providers[provider.GetName()] = provider
	mm.logger.WithField("provider", provider.GetName()).Info("Model provider registered")

	return nil
}

// GetModel retrieves a model by ID
func (mm *ModelManager) GetModel(modelID string) (*AIModel, error) {
	mm.mutex.RLock()
	defer mm.mutex.RUnlock()

	model, exists := mm.models[modelID]
	if !exists {
		return nil, fmt.Errorf("model not found: %s", modelID)
	}

	return model, nil
}

// ListModels returns all registered models
func (mm *ModelManager) ListModels() []*AIModel {
	mm.mutex.RLock()
	defer mm.mutex.RUnlock()

	models := make([]*AIModel, 0, len(mm.models))
	for _, model := range mm.models {
		models = append(models, model)
	}

	return models
}

// GetModelsByType returns models filtered by type
func (mm *ModelManager) GetModelsByType(modelType string) []*AIModel {
	mm.mutex.RLock()
	defer mm.mutex.RUnlock()

	var models []*AIModel
	for _, model := range mm.models {
		if model.Type == modelType {
			models = append(models, model)
		}
	}

	return models
}

// GetModelsByProvider returns models filtered by provider
func (mm *ModelManager) GetModelsByProvider(provider string) []*AIModel {
	mm.mutex.RLock()
	defer mm.mutex.RUnlock()

	var models []*AIModel
	for _, model := range mm.models {
		if model.Provider == provider {
			models = append(models, model)
		}
	}

	return models
}

// SelectBestModel selects the best model for a given task
func (mm *ModelManager) SelectBestModel(taskType string, requirements map[string]interface{}) (*AIModel, error) {
	models := mm.GetModelsByType(taskType)
	if len(models) == 0 {
		return nil, fmt.Errorf("no models available for task type: %s", taskType)
	}

	// Use load balancer to select best model
	return mm.loadBalancer.SelectModel(models, requirements)
}

// GenerateText generates text using the specified model
func (mm *ModelManager) GenerateText(ctx context.Context, request *TextGenerationRequest) (*TextGenerationResponse, error) {
	ctx, span := mm.tracer.Start(ctx, "model_manager.generate_text")
	defer span.End()

	// Get model
	model, err := mm.GetModel(request.ModelID)
	if err != nil {
		return nil, err
	}

	// Get provider
	provider, exists := mm.providers[model.Provider]
	if !exists {
		return nil, fmt.Errorf("provider not found: %s", model.Provider)
	}

	// Check cache first
	if cached := mm.cache.GetTextResponse(request); cached != nil {
		mm.logger.Debug("Returning cached text response")
		return cached, nil
	}

	// Generate text
	response, err := provider.GenerateText(ctx, request)
	if err != nil {
		mm.monitor.RecordError(model.ID, err)
		return nil, err
	}

	// Cache response
	mm.cache.SetTextResponse(request, response)

	// Record metrics
	mm.monitor.RecordRequest(model.ID, response.Latency, response.Cost)

	return response, nil
}

// GetProviderHealth returns health status for all providers
func (mm *ModelManager) GetProviderHealth() map[string]*ProviderHealth {
	mm.mutex.RLock()
	defer mm.mutex.RUnlock()

	health := make(map[string]*ProviderHealth)
	for name, provider := range mm.providers {
		health[name] = provider.GetHealth()
	}

	return health
}

// GetUsageStats returns usage statistics for all providers
func (mm *ModelManager) GetUsageStats() map[string]*ProviderUsage {
	mm.mutex.RLock()
	defer mm.mutex.RUnlock()

	usage := make(map[string]*ProviderUsage)
	for name, provider := range mm.providers {
		usage[name] = provider.GetUsage()
	}

	return usage
}

// UpdateModelConfig updates a model's configuration
func (mm *ModelManager) UpdateModelConfig(modelID string, config ModelConfig) error {
	mm.mutex.Lock()
	defer mm.mutex.Unlock()

	model, exists := mm.models[modelID]
	if !exists {
		return fmt.Errorf("model not found: %s", modelID)
	}

	model.Config = config
	model.UpdatedAt = time.Now()

	mm.logger.WithField("model_id", modelID).Info("Model configuration updated")
	return nil
}

// DeactivateModel deactivates a model
func (mm *ModelManager) DeactivateModel(modelID string) error {
	mm.mutex.Lock()
	defer mm.mutex.Unlock()

	model, exists := mm.models[modelID]
	if !exists {
		return fmt.Errorf("model not found: %s", modelID)
	}

	model.Status = "inactive"
	model.UpdatedAt = time.Now()

	mm.logger.WithField("model_id", modelID).Info("Model deactivated")
	return nil
}
