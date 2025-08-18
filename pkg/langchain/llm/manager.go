package llm

import (
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
)

// Manager defines the interface for LLM management
type Manager interface {
	// Model management
	RegisterModel(model Model) error
	UnregisterModel(modelID string) error
	GetModel(modelID string) (Model, error)
	ListModels() []Model
	GetAvailableModels() []ModelInfo

	// Generation methods
	Generate(ctx context.Context, request *GenerationRequest) (*GenerationResponse, error)
	GenerateStream(ctx context.Context, request *GenerationRequest) (<-chan *StreamChunk, error)

	// Configuration
	SetDefaultModel(modelID string) error
	GetDefaultModel() (Model, error)

	// Monitoring
	GetMetrics() *ManagerMetrics
	GetStatus() *ManagerStatus
}

// Model defines the interface for LLM models
type Model interface {
	GetID() string
	GetName() string
	GetProvider() string
	GetCapabilities() []string
	Generate(ctx context.Context, request *GenerationRequest) (*GenerationResponse, error)
	GenerateStream(ctx context.Context, request *GenerationRequest) (<-chan *StreamChunk, error)
	IsAvailable() bool
	GetConfig() *ModelConfig
}

// ModelInfo represents model information
type ModelInfo struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Provider     string   `json:"provider"`
	Description  string   `json:"description"`
	Capabilities []string `json:"capabilities"`
	Available    bool     `json:"available"`
	MaxTokens    int      `json:"max_tokens"`
	CostPerToken float64  `json:"cost_per_token"`
}

// ModelConfig represents model configuration
type ModelConfig struct {
	MaxTokens        int                    `json:"max_tokens"`
	Temperature      float64                `json:"temperature"`
	TopP             float64                `json:"top_p"`
	FrequencyPenalty float64                `json:"frequency_penalty"`
	PresencePenalty  float64                `json:"presence_penalty"`
	StopSequences    []string               `json:"stop_sequences"`
	Parameters       map[string]interface{} `json:"parameters"`
}

// GenerationRequest represents a generation request
type GenerationRequest struct {
	ModelID  string                 `json:"model_id"`
	Messages []*Message             `json:"messages"`
	Config   *ModelConfig           `json:"config,omitempty"`
	Stream   bool                   `json:"stream"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// ChatMessage represents a chat message for the manager
type ChatMessage struct {
	Role    string `json:"role"` // "system", "user", "assistant"
	Content string `json:"content"`
}

// GenerationResponse represents a generation response
type GenerationResponse struct {
	ID           string                 `json:"id"`
	ModelID      string                 `json:"model_id"`
	Content      string                 `json:"content"`
	FinishReason string                 `json:"finish_reason"`
	Usage        *TokenUsage            `json:"usage"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt    time.Time              `json:"created_at"`
}

// StreamChunk represents a streaming response chunk
type StreamChunk struct {
	ID           string                 `json:"id"`
	ModelID      string                 `json:"model_id"`
	Content      string                 `json:"content"`
	Delta        string                 `json:"delta"`
	FinishReason string                 `json:"finish_reason,omitempty"`
	Usage        *TokenUsage            `json:"usage,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt    time.Time              `json:"created_at"`
}

// ManagerTokenUsage represents token usage information for the manager
type ManagerTokenUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// ManagerMetrics represents LLM manager metrics
type ManagerMetrics struct {
	TotalRequests      int64         `json:"total_requests"`
	SuccessfulRequests int64         `json:"successful_requests"`
	FailedRequests     int64         `json:"failed_requests"`
	AverageLatency     time.Duration `json:"average_latency"`
	TotalTokens        int64         `json:"total_tokens"`
	TotalCost          float64       `json:"total_cost"`
	RequestsPerSecond  float64       `json:"requests_per_second"`
	ErrorRate          float64       `json:"error_rate"`
	StartTime          time.Time     `json:"start_time"`
	LastUpdate         time.Time     `json:"last_update"`
}

// ManagerStatus represents LLM manager status
type ManagerStatus struct {
	Running         bool            `json:"running"`
	ModelCount      int             `json:"model_count"`
	AvailableModels int             `json:"available_models"`
	ActiveRequests  int             `json:"active_requests"`
	DefaultModelID  string          `json:"default_model_id"`
	ModelStatuses   map[string]bool `json:"model_statuses"`
	LastActivity    time.Time       `json:"last_activity"`
}

// DefaultManager implements the Manager interface
type DefaultManager struct {
	models       map[string]Model
	defaultModel string
	logger       *logrus.Logger
	metrics      *ManagerMetrics
}

// NewDefaultManager creates a new default LLM manager
func NewDefaultManager(logger *logrus.Logger) Manager {
	return &DefaultManager{
		models: make(map[string]Model),
		logger: logger,
		metrics: &ManagerMetrics{
			StartTime:  time.Now(),
			LastUpdate: time.Now(),
		},
	}
}

// RegisterModel registers a model
func (m *DefaultManager) RegisterModel(model Model) error {
	m.models[model.GetID()] = model
	m.logger.WithField("model_id", model.GetID()).Info("LLM model registered")
	return nil
}

// UnregisterModel unregisters a model
func (m *DefaultManager) UnregisterModel(modelID string) error {
	delete(m.models, modelID)
	m.logger.WithField("model_id", modelID).Info("LLM model unregistered")
	return nil
}

// GetModel gets a model by ID
func (m *DefaultManager) GetModel(modelID string) (Model, error) {
	model, exists := m.models[modelID]
	if !exists {
		return nil, fmt.Errorf("model not found: %s", modelID)
	}
	return model, nil
}

// ListModels lists all models
func (m *DefaultManager) ListModels() []Model {
	models := make([]Model, 0, len(m.models))
	for _, model := range m.models {
		models = append(models, model)
	}
	return models
}

// GetAvailableModels returns available model information
func (m *DefaultManager) GetAvailableModels() []ModelInfo {
	var models []ModelInfo
	for _, model := range m.models {
		config := model.GetConfig()
		models = append(models, ModelInfo{
			ID:           model.GetID(),
			Name:         model.GetName(),
			Provider:     model.GetProvider(),
			Capabilities: model.GetCapabilities(),
			Available:    model.IsAvailable(),
			MaxTokens:    config.MaxTokens,
		})
	}
	return models
}

// Generate generates text using the specified model
func (m *DefaultManager) Generate(ctx context.Context, request *GenerationRequest) (*GenerationResponse, error) {
	startTime := time.Now()

	// Get model
	modelID := request.ModelID
	if modelID == "" {
		modelID = m.defaultModel
	}

	model, err := m.GetModel(modelID)
	if err != nil {
		m.updateMetrics(false, time.Since(startTime), 0)
		return nil, err
	}

	// Generate response
	response, err := model.Generate(ctx, request)
	if err != nil {
		m.updateMetrics(false, time.Since(startTime), 0)
		return nil, err
	}

	// Update metrics
	tokens := 0
	if response.Usage != nil {
		tokens = response.Usage.TotalTokens
	}
	m.updateMetrics(true, time.Since(startTime), int64(tokens))

	return response, nil
}

// GenerateStream generates streaming text
func (m *DefaultManager) GenerateStream(ctx context.Context, request *GenerationRequest) (<-chan *StreamChunk, error) {
	// Get model
	modelID := request.ModelID
	if modelID == "" {
		modelID = m.defaultModel
	}

	model, err := m.GetModel(modelID)
	if err != nil {
		return nil, err
	}

	return model.GenerateStream(ctx, request)
}

// SetDefaultModel sets the default model
func (m *DefaultManager) SetDefaultModel(modelID string) error {
	if _, exists := m.models[modelID]; !exists {
		return fmt.Errorf("model not found: %s", modelID)
	}
	m.defaultModel = modelID
	return nil
}

// GetDefaultModel gets the default model
func (m *DefaultManager) GetDefaultModel() (Model, error) {
	if m.defaultModel == "" {
		return nil, fmt.Errorf("no default model set")
	}
	return m.GetModel(m.defaultModel)
}

// GetMetrics returns manager metrics
func (m *DefaultManager) GetMetrics() *ManagerMetrics {
	m.metrics.LastUpdate = time.Now()
	return m.metrics
}

// GetStatus returns manager status
func (m *DefaultManager) GetStatus() *ManagerStatus {
	modelStatuses := make(map[string]bool)
	availableCount := 0

	for id, model := range m.models {
		available := model.IsAvailable()
		modelStatuses[id] = available
		if available {
			availableCount++
		}
	}

	return &ManagerStatus{
		Running:         true,
		ModelCount:      len(m.models),
		AvailableModels: availableCount,
		DefaultModelID:  m.defaultModel,
		ModelStatuses:   modelStatuses,
		LastActivity:    time.Now(),
	}
}

// updateMetrics updates manager metrics
func (m *DefaultManager) updateMetrics(success bool, latency time.Duration, tokens int64) {
	m.metrics.TotalRequests++
	if success {
		m.metrics.SuccessfulRequests++
	} else {
		m.metrics.FailedRequests++
	}

	m.metrics.TotalTokens += tokens

	if m.metrics.TotalRequests > 0 {
		m.metrics.ErrorRate = float64(m.metrics.FailedRequests) / float64(m.metrics.TotalRequests)
	}

	// Update average latency (simple moving average)
	if m.metrics.AverageLatency == 0 {
		m.metrics.AverageLatency = latency
	} else {
		m.metrics.AverageLatency = (m.metrics.AverageLatency + latency) / 2
	}
}
