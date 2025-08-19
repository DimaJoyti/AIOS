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

// AIOrchestrator orchestrates AI services and provides a unified interface
type AIOrchestrator struct {
	modelManager  *ModelManager
	promptManager *PromptManager
	safetyFilter  *SafetyFilter
	analytics     *AIAnalytics
	config        *OrchestratorConfig
	logger        *logrus.Logger
	tracer        trace.Tracer
	mutex         sync.RWMutex
}

// OrchestratorConfig holds configuration for the AI orchestrator
type OrchestratorConfig struct {
	DefaultModel        string        `json:"default_model"`
	MaxConcurrentTasks  int           `json:"max_concurrent_tasks"`
	DefaultTimeout      time.Duration `json:"default_timeout"`
	EnableSafetyFilter  bool          `json:"enable_safety_filter"`
	EnableAnalytics     bool          `json:"enable_analytics"`
	CacheEnabled        bool          `json:"cache_enabled"`
	CacheTTL            time.Duration `json:"cache_ttl"`
	RateLimitPerMinute  int           `json:"rate_limit_per_minute"`
	CostLimitPerHour    float64       `json:"cost_limit_per_hour"`
	EnableLoadBalancing bool          `json:"enable_load_balancing"`
}

// AIRequest represents a unified AI request
type AIRequest struct {
	Type         string                 `json:"type"` // text, image, audio, video, multimodal
	ModelID      string                 `json:"model_id,omitempty"`
	Prompt       string                 `json:"prompt,omitempty"`
	SystemPrompt string                 `json:"system_prompt,omitempty"`
	Messages     []ChatMessage          `json:"messages,omitempty"`
	Config       *ModelConfig           `json:"config,omitempty"`
	TemplateID   string                 `json:"template_id,omitempty"`
	Variables    map[string]interface{} `json:"variables,omitempty"`
	ChainID      string                 `json:"chain_id,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	UserID       string                 `json:"user_id,omitempty"`
	SessionID    string                 `json:"session_id,omitempty"`
	Priority     int                    `json:"priority,omitempty"`
	Timeout      time.Duration          `json:"timeout,omitempty"`
}

// AIResponse represents a unified AI response
type AIResponse struct {
	ID           string                 `json:"id"`
	Type         string                 `json:"type"`
	Text         string                 `json:"text,omitempty"`
	Images       []GeneratedImage       `json:"images,omitempty"`
	AudioURL     string                 `json:"audio_url,omitempty"`
	VideoURL     string                 `json:"video_url,omitempty"`
	Usage        TokenUsage             `json:"usage"`
	Cost         float64                `json:"cost"`
	Latency      time.Duration          `json:"latency"`
	ModelID      string                 `json:"model_id"`
	Provider     string                 `json:"provider"`
	FinishReason string                 `json:"finish_reason,omitempty"`
	SafetyFlags  []SafetyFlag           `json:"safety_flags,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt    time.Time              `json:"created_at"`
}

// SafetyFilter provides content safety filtering
type SafetyFilter struct {
	enabled bool
	rules   []SafetyRule
	logger  *logrus.Logger
}

// SafetyRule defines a content safety rule
type SafetyRule struct {
	ID       string   `json:"id"`
	Name     string   `json:"name"`
	Type     string   `json:"type"`     // content, toxicity, bias, privacy
	Severity string   `json:"severity"` // low, medium, high, critical
	Keywords []string `json:"keywords,omitempty"`
	Patterns []string `json:"patterns,omitempty"`
	Action   string   `json:"action"` // warn, block, filter
	Enabled  bool     `json:"enabled"`
}

// SafetyFlag represents a safety concern
type SafetyFlag struct {
	RuleID      string  `json:"rule_id"`
	Type        string  `json:"type"`
	Severity    string  `json:"severity"`
	Confidence  float64 `json:"confidence"`
	Description string  `json:"description"`
	Action      string  `json:"action"`
}

// AIAnalytics provides analytics for AI operations
type AIAnalytics struct {
	enabled bool
	metrics map[string]*AnalyticsMetrics
	mutex   sync.RWMutex
	logger  *logrus.Logger
}

// AnalyticsMetrics represents analytics metrics
type AnalyticsMetrics struct {
	TotalRequests      int64              `json:"total_requests"`
	SuccessfulRequests int64              `json:"successful_requests"`
	FailedRequests     int64              `json:"failed_requests"`
	TotalCost          float64            `json:"total_cost"`
	TotalTokens        int64              `json:"total_tokens"`
	AverageLatency     time.Duration      `json:"average_latency"`
	TopModels          map[string]int64   `json:"top_models"`
	TopUsers           map[string]int64   `json:"top_users"`
	CostByModel        map[string]float64 `json:"cost_by_model"`
	LastUpdate         time.Time          `json:"last_update"`
}

// NewAIOrchestrator creates a new AI orchestrator
func NewAIOrchestrator(config *OrchestratorConfig, logger *logrus.Logger) *AIOrchestrator {
	orchestrator := &AIOrchestrator{
		modelManager:  NewModelManager(logger),
		promptManager: NewPromptManager(logger),
		safetyFilter:  NewSafetyFilter(config.EnableSafetyFilter, logger),
		analytics:     NewAIAnalytics(config.EnableAnalytics, logger),
		config:        config,
		logger:        logger,
		tracer:        otel.Tracer("ai.services.orchestrator"),
	}

	return orchestrator
}

// ProcessRequest processes a unified AI request
func (ao *AIOrchestrator) ProcessRequest(ctx context.Context, request *AIRequest) (*AIResponse, error) {
	ctx, span := ao.tracer.Start(ctx, "ai_orchestrator.process_request")
	defer span.End()

	startTime := time.Now()
	responseID := generateResponseID()

	// Apply timeout
	if request.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, request.Timeout)
		defer cancel()
	} else if ao.config.DefaultTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, ao.config.DefaultTimeout)
		defer cancel()
	}

	// Safety filtering on input
	if ao.config.EnableSafetyFilter {
		if flags := ao.safetyFilter.CheckInput(request); len(flags) > 0 {
			for _, flag := range flags {
				if flag.Action == "block" {
					return &AIResponse{
						ID:          responseID,
						Type:        request.Type,
						SafetyFlags: flags,
						CreatedAt:   time.Now(),
					}, fmt.Errorf("request blocked by safety filter: %s", flag.Description)
				}
			}
		}
	}

	var response *AIResponse
	var err error

	// Route request based on type and method
	switch {
	case request.ChainID != "":
		response, err = ao.processChainRequest(ctx, request)
	case request.TemplateID != "":
		response, err = ao.processTemplateRequest(ctx, request)
	default:
		response, err = ao.processDirectRequest(ctx, request)
	}

	if err != nil {
		// Record analytics for failed request
		if ao.config.EnableAnalytics {
			ao.analytics.RecordRequest(request, nil, err)
		}
		return nil, err
	}

	// Safety filtering on output
	if ao.config.EnableSafetyFilter {
		if flags := ao.safetyFilter.CheckOutput(response); len(flags) > 0 {
			response.SafetyFlags = append(response.SafetyFlags, flags...)
			for _, flag := range flags {
				if flag.Action == "block" {
					response.Text = "[Content blocked by safety filter]"
					break
				} else if flag.Action == "filter" {
					response.Text = ao.safetyFilter.FilterContent(response.Text, flag)
				}
			}
		}
	}

	// Set response metadata
	response.ID = responseID
	response.Latency = time.Since(startTime)
	response.CreatedAt = startTime

	// Record analytics
	if ao.config.EnableAnalytics {
		ao.analytics.RecordRequest(request, response, nil)
	}

	ao.logger.WithFields(logrus.Fields{
		"response_id": responseID,
		"type":        request.Type,
		"model_id":    response.ModelID,
		"latency":     response.Latency,
		"cost":        response.Cost,
		"user_id":     request.UserID,
	}).Info("AI request processed")

	return response, nil
}

// processDirectRequest processes a direct AI request
func (ao *AIOrchestrator) processDirectRequest(ctx context.Context, request *AIRequest) (*AIResponse, error) {
	switch request.Type {
	case "text":
		return ao.processTextRequest(ctx, request)
	case "image":
		return ao.processImageRequest(ctx, request)
	case "audio":
		return ao.processAudioRequest(ctx, request)
	case "video":
		return ao.processVideoRequest(ctx, request)
	default:
		return nil, fmt.Errorf("unsupported request type: %s", request.Type)
	}
}

// processTextRequest processes a text generation request
func (ao *AIOrchestrator) processTextRequest(ctx context.Context, request *AIRequest) (*AIResponse, error) {
	// Select model if not specified
	modelID := request.ModelID
	if modelID == "" {
		if ao.config.DefaultModel != "" {
			modelID = ao.config.DefaultModel
		} else {
			model, err := ao.modelManager.SelectBestModel("text", map[string]interface{}{
				"user_id": request.UserID,
			})
			if err != nil {
				return nil, err
			}
			modelID = model.ID
		}
	}

	// Create text generation request
	textReq := &TextGenerationRequest{
		ModelID:      modelID,
		Prompt:       request.Prompt,
		SystemPrompt: request.SystemPrompt,
		Messages:     request.Messages,
		Metadata:     request.Metadata,
	}

	// Apply configuration
	if request.Config != nil {
		textReq.Config = *request.Config
	} else {
		// Use default configuration
		textReq.Config = ModelConfig{
			MaxTokens:   1000,
			Temperature: 0.7,
		}
	}

	// Generate text
	textResp, err := ao.modelManager.GenerateText(ctx, textReq)
	if err != nil {
		return nil, err
	}

	return &AIResponse{
		Type:         "text",
		Text:         textResp.Text,
		Usage:        textResp.Usage,
		Cost:         textResp.Cost,
		ModelID:      textResp.ModelID,
		Provider:     textResp.Provider,
		FinishReason: textResp.FinishReason,
		Metadata:     textResp.Metadata,
	}, nil
}

// processTemplateRequest processes a template-based request
func (ao *AIOrchestrator) processTemplateRequest(ctx context.Context, request *AIRequest) (*AIResponse, error) {
	execution, err := ao.promptManager.ExecuteTemplate(ctx, ao.modelManager, request.TemplateID, request.Variables)
	if err != nil {
		return nil, err
	}

	if execution.Result == nil {
		return nil, fmt.Errorf("template execution failed")
	}

	return &AIResponse{
		Type:         "text",
		Text:         execution.Result.Text,
		Usage:        execution.Result.Usage,
		Cost:         execution.Result.Cost,
		ModelID:      execution.Result.ModelID,
		Provider:     execution.Result.Provider,
		FinishReason: execution.Result.FinishReason,
		Metadata:     execution.Result.Metadata,
	}, nil
}

// processChainRequest processes a chain-based request
func (ao *AIOrchestrator) processChainRequest(ctx context.Context, request *AIRequest) (*AIResponse, error) {
	executions, err := ao.promptManager.ExecuteChain(ctx, ao.modelManager, request.ChainID, request.Variables)
	if err != nil {
		return nil, err
	}

	if len(executions) == 0 {
		return nil, fmt.Errorf("chain execution produced no results")
	}

	// Combine results from all executions
	var combinedText string
	var totalCost float64
	var totalUsage TokenUsage
	var lastExecution *PromptExecution

	for _, execution := range executions {
		if execution.Result != nil {
			combinedText += execution.Result.Text + "\n"
			totalCost += execution.Result.Cost
			totalUsage.PromptTokens += execution.Result.Usage.PromptTokens
			totalUsage.CompletionTokens += execution.Result.Usage.CompletionTokens
			totalUsage.TotalTokens += execution.Result.Usage.TotalTokens
			lastExecution = execution
		}
	}

	if lastExecution == nil {
		return nil, fmt.Errorf("no successful executions in chain")
	}

	return &AIResponse{
		Type:         "text",
		Text:         combinedText,
		Usage:        totalUsage,
		Cost:         totalCost,
		ModelID:      lastExecution.Result.ModelID,
		Provider:     lastExecution.Result.Provider,
		FinishReason: lastExecution.Result.FinishReason,
		Metadata:     request.Metadata,
	}, nil
}

// processImageRequest processes an image generation request
func (ao *AIOrchestrator) processImageRequest(ctx context.Context, request *AIRequest) (*AIResponse, error) {
	// Implementation for image generation
	// This would use the model manager to generate images
	return nil, fmt.Errorf("image generation not implemented")
}

// processAudioRequest processes an audio processing request
func (ao *AIOrchestrator) processAudioRequest(ctx context.Context, request *AIRequest) (*AIResponse, error) {
	// Implementation for audio processing
	// This would use the model manager to process audio
	return nil, fmt.Errorf("audio processing not implemented")
}

// processVideoRequest processes a video processing request
func (ao *AIOrchestrator) processVideoRequest(ctx context.Context, request *AIRequest) (*AIResponse, error) {
	// Implementation for video processing
	// This would use the model manager to process video
	return nil, fmt.Errorf("video processing not implemented")
}

// GetModelManager returns the model manager
func (ao *AIOrchestrator) GetModelManager() *ModelManager {
	return ao.modelManager
}

// GetPromptManager returns the prompt manager
func (ao *AIOrchestrator) GetPromptManager() *PromptManager {
	return ao.promptManager
}

// GetAnalytics returns analytics data
func (ao *AIOrchestrator) GetAnalytics() *AIAnalytics {
	return ao.analytics
}

// GetHealth returns health status of all components
func (ao *AIOrchestrator) GetHealth() map[string]interface{} {
	health := map[string]interface{}{
		"status": "healthy",
		"components": map[string]interface{}{
			"model_manager":  ao.modelManager.GetProviderHealth(),
			"prompt_manager": "healthy",
			"safety_filter":  ao.safetyFilter.GetStatus(),
			"analytics":      ao.analytics.GetStatus(),
		},
		"timestamp": time.Now(),
	}

	return health
}

// NewSafetyFilter creates a new safety filter
func NewSafetyFilter(enabled bool, logger *logrus.Logger) *SafetyFilter {
	return &SafetyFilter{
		enabled: enabled,
		rules:   make([]SafetyRule, 0),
		logger:  logger,
	}
}

// CheckInput checks input for safety violations
func (sf *SafetyFilter) CheckInput(request *AIRequest) []SafetyFlag {
	if !sf.enabled {
		return nil
	}

	var flags []SafetyFlag
	// Implementation for input safety checking
	return flags
}

// CheckOutput checks output for safety violations
func (sf *SafetyFilter) CheckOutput(response *AIResponse) []SafetyFlag {
	if !sf.enabled {
		return nil
	}

	var flags []SafetyFlag
	// Implementation for output safety checking
	return flags
}

// FilterContent filters content based on safety rules
func (sf *SafetyFilter) FilterContent(content string, flag SafetyFlag) string {
	// Implementation for content filtering
	return content
}

// GetStatus returns safety filter status
func (sf *SafetyFilter) GetStatus() string {
	if sf.enabled {
		return "enabled"
	}
	return "disabled"
}

// NewAIAnalytics creates a new AI analytics instance
func NewAIAnalytics(enabled bool, logger *logrus.Logger) *AIAnalytics {
	return &AIAnalytics{
		enabled: enabled,
		metrics: make(map[string]*AnalyticsMetrics),
		logger:  logger,
	}
}

// RecordRequest records analytics for a request
func (aa *AIAnalytics) RecordRequest(request *AIRequest, response *AIResponse, err error) {
	if !aa.enabled {
		return
	}

	aa.mutex.Lock()
	defer aa.mutex.Unlock()

	// Implementation for analytics recording
}

// GetStatus returns analytics status
func (aa *AIAnalytics) GetStatus() string {
	if aa.enabled {
		return "enabled"
	}
	return "disabled"
}

// generateResponseID generates a unique response ID
func generateResponseID() string {
	return fmt.Sprintf("resp_%d", time.Now().UnixNano())
}
