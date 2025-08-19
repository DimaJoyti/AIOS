package providers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/aios/aios/internal/ai/services"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

// OpenAIProvider implements the ModelProvider interface for OpenAI
type OpenAIProvider struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
	logger     *logrus.Logger
	tracer     trace.Tracer
	usage      *services.ProviderUsage
	health     *services.ProviderHealth
}

// OpenAIRequest represents a request to OpenAI API
type OpenAIRequest struct {
	Model       string          `json:"model"`
	Messages    []OpenAIMessage `json:"messages,omitempty"`
	Prompt      string          `json:"prompt,omitempty"`
	MaxTokens   int             `json:"max_tokens,omitempty"`
	Temperature float64         `json:"temperature,omitempty"`
	TopP        float64         `json:"top_p,omitempty"`
	N           int             `json:"n,omitempty"`
	Stream      bool            `json:"stream,omitempty"`
	Stop        []string        `json:"stop,omitempty"`
	User        string          `json:"user,omitempty"`
}

// OpenAIMessage represents a message in OpenAI format
type OpenAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// OpenAIResponse represents a response from OpenAI API
type OpenAIResponse struct {
	ID      string         `json:"id"`
	Object  string         `json:"object"`
	Created int64          `json:"created"`
	Model   string         `json:"model"`
	Choices []OpenAIChoice `json:"choices"`
	Usage   OpenAIUsage    `json:"usage"`
	Error   *OpenAIError   `json:"error,omitempty"`
}

// OpenAIChoice represents a choice in OpenAI response
type OpenAIChoice struct {
	Index        int            `json:"index"`
	Message      *OpenAIMessage `json:"message,omitempty"`
	Text         string         `json:"text,omitempty"`
	FinishReason string         `json:"finish_reason"`
}

// OpenAIUsage represents usage information from OpenAI
type OpenAIUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// OpenAIError represents an error from OpenAI API
type OpenAIError struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Code    string `json:"code"`
}

// OpenAIImageRequest represents an image generation request
type OpenAIImageRequest struct {
	Prompt         string `json:"prompt"`
	N              int    `json:"n,omitempty"`
	Size           string `json:"size,omitempty"`
	ResponseFormat string `json:"response_format,omitempty"`
	User           string `json:"user,omitempty"`
}

// OpenAIImageResponse represents an image generation response
type OpenAIImageResponse struct {
	Created int64             `json:"created"`
	Data    []OpenAIImageData `json:"data"`
	Error   *OpenAIError      `json:"error,omitempty"`
}

// OpenAIImageData represents image data from OpenAI
type OpenAIImageData struct {
	URL     string `json:"url,omitempty"`
	B64JSON string `json:"b64_json,omitempty"`
}

// NewOpenAIProvider creates a new OpenAI provider
func NewOpenAIProvider(apiKey string, logger *logrus.Logger) *OpenAIProvider {
	return &OpenAIProvider{
		apiKey:  apiKey,
		baseURL: "https://api.openai.com/v1",
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
		logger: logger,
		tracer: otel.Tracer("ai.providers.openai"),
		usage: &services.ProviderUsage{
			LastRequestTime: time.Now(),
		},
		health: &services.ProviderHealth{
			Status:    "healthy",
			LastCheck: time.Now(),
		},
	}
}

// GetName returns the provider name
func (p *OpenAIProvider) GetName() string {
	return "openai"
}

// GetModels returns available models
func (p *OpenAIProvider) GetModels() []string {
	return []string{
		"gpt-4",
		"gpt-4-turbo",
		"gpt-3.5-turbo",
		"text-davinci-003",
		"text-curie-001",
		"text-babbage-001",
		"text-ada-001",
		"dall-e-3",
		"dall-e-2",
		"whisper-1",
	}
}

// IsAvailable checks if the provider is available
func (p *OpenAIProvider) IsAvailable() bool {
	return p.health.Status == "healthy"
}

// GenerateText generates text using OpenAI models
func (p *OpenAIProvider) GenerateText(ctx context.Context, request *services.TextGenerationRequest) (*services.TextGenerationResponse, error) {
	ctx, span := p.tracer.Start(ctx, "openai_provider.generate_text")
	defer span.End()

	startTime := time.Now()

	// Convert request to OpenAI format
	openaiReq := &OpenAIRequest{
		Model:       request.ModelID,
		MaxTokens:   request.Config.MaxTokens,
		Temperature: request.Config.Temperature,
		TopP:        request.Config.TopP,
		Stop:        request.Config.StopSequences,
	}

	// Handle different input formats
	if len(request.Messages) > 0 {
		// Chat completion format
		openaiReq.Messages = make([]OpenAIMessage, len(request.Messages))
		for i, msg := range request.Messages {
			openaiReq.Messages[i] = OpenAIMessage{
				Role:    msg.Role,
				Content: msg.Content,
			}
		}

		// Add system prompt if provided
		if request.SystemPrompt != "" {
			systemMsg := OpenAIMessage{
				Role:    "system",
				Content: request.SystemPrompt,
			}
			openaiReq.Messages = append([]OpenAIMessage{systemMsg}, openaiReq.Messages...)
		}
	} else {
		// Completion format
		openaiReq.Prompt = request.Prompt
	}

	// Make API request
	var endpoint string
	if len(request.Messages) > 0 || request.SystemPrompt != "" {
		endpoint = "/chat/completions"
	} else {
		endpoint = "/completions"
	}

	response, err := p.makeRequest(ctx, endpoint, openaiReq)
	if err != nil {
		p.updateHealth("unhealthy", err.Error())
		return nil, err
	}

	var openaiResp OpenAIResponse
	if err := json.Unmarshal(response, &openaiResp); err != nil {
		return nil, fmt.Errorf("failed to parse OpenAI response: %w", err)
	}

	if openaiResp.Error != nil {
		return nil, fmt.Errorf("OpenAI API error: %s", openaiResp.Error.Message)
	}

	if len(openaiResp.Choices) == 0 {
		return nil, fmt.Errorf("no choices returned from OpenAI")
	}

	// Extract text from response
	var text string
	var finishReason string
	choice := openaiResp.Choices[0]

	if choice.Message != nil {
		text = choice.Message.Content
	} else {
		text = choice.Text
	}
	finishReason = choice.FinishReason

	latency := time.Since(startTime)

	// Calculate cost (simplified pricing)
	cost := p.calculateCost(request.ModelID, openaiResp.Usage.PromptTokens, openaiResp.Usage.CompletionTokens)

	// Update usage statistics
	p.updateUsage(latency, cost)
	p.updateHealth("healthy", "")

	return &services.TextGenerationResponse{
		Text:         text,
		FinishReason: finishReason,
		Usage: services.TokenUsage{
			PromptTokens:     openaiResp.Usage.PromptTokens,
			CompletionTokens: openaiResp.Usage.CompletionTokens,
			TotalTokens:      openaiResp.Usage.TotalTokens,
		},
		ModelID:  request.ModelID,
		Provider: "openai",
		Latency:  latency,
		Cost:     cost,
		Metadata: request.Metadata,
	}, nil
}

// GenerateImage generates images using OpenAI DALL-E
func (p *OpenAIProvider) GenerateImage(ctx context.Context, request *services.ImageGenerationRequest) (*services.ImageGenerationResponse, error) {
	ctx, span := p.tracer.Start(ctx, "openai_provider.generate_image")
	defer span.End()

	startTime := time.Now()

	// Convert request to OpenAI format
	openaiReq := &OpenAIImageRequest{
		Prompt: request.Prompt,
		N:      request.Count,
		Size:   fmt.Sprintf("%dx%d", request.Width, request.Height),
	}

	if request.Format == "base64" {
		openaiReq.ResponseFormat = "b64_json"
	} else {
		openaiReq.ResponseFormat = "url"
	}

	// Make API request
	response, err := p.makeRequest(ctx, "/images/generations", openaiReq)
	if err != nil {
		p.updateHealth("unhealthy", err.Error())
		return nil, err
	}

	var openaiResp OpenAIImageResponse
	if err := json.Unmarshal(response, &openaiResp); err != nil {
		return nil, fmt.Errorf("failed to parse OpenAI image response: %w", err)
	}

	if openaiResp.Error != nil {
		return nil, fmt.Errorf("OpenAI API error: %s", openaiResp.Error.Message)
	}

	// Convert response
	images := make([]services.GeneratedImage, len(openaiResp.Data))
	for i, data := range openaiResp.Data {
		images[i] = services.GeneratedImage{
			URL:    data.URL,
			Base64: data.B64JSON,
			Width:  request.Width,
			Height: request.Height,
			Format: request.Format,
		}
	}

	latency := time.Since(startTime)
	cost := p.calculateImageCost(request.ModelID, request.Width, request.Height, request.Count)

	p.updateUsage(latency, cost)
	p.updateHealth("healthy", "")

	return &services.ImageGenerationResponse{
		Images:   images,
		ModelID:  request.ModelID,
		Provider: "openai",
		Latency:  latency,
		Cost:     cost,
		Metadata: request.Metadata,
	}, nil
}

// ProcessAudio processes audio using OpenAI Whisper
func (p *OpenAIProvider) ProcessAudio(ctx context.Context, request *services.AudioProcessingRequest) (*services.AudioProcessingResponse, error) {
	ctx, span := p.tracer.Start(ctx, "openai_provider.process_audio")
	defer span.End()

	startTime := time.Now()

	// For now, return a placeholder implementation
	// In a real implementation, you would use OpenAI's Whisper API

	latency := time.Since(startTime)
	cost := 0.006 // Whisper pricing per minute

	p.updateUsage(latency, cost)
	p.updateHealth("healthy", "")

	return &services.AudioProcessingResponse{
		Text:     "Audio transcription placeholder",
		Duration: time.Minute,
		ModelID:  request.ModelID,
		Provider: "openai",
		Latency:  latency,
		Cost:     cost,
		Metadata: request.Metadata,
	}, nil
}

// ProcessVideo processes video (not supported by OpenAI)
func (p *OpenAIProvider) ProcessVideo(ctx context.Context, request *services.VideoProcessingRequest) (*services.VideoProcessingResponse, error) {
	return nil, fmt.Errorf("video processing not supported by OpenAI provider")
}

// GetUsage returns usage statistics
func (p *OpenAIProvider) GetUsage() *services.ProviderUsage {
	return p.usage
}

// GetHealth returns health status
func (p *OpenAIProvider) GetHealth() *services.ProviderHealth {
	return p.health
}

// makeRequest makes an HTTP request to OpenAI API
func (p *OpenAIProvider) makeRequest(ctx context.Context, endpoint string, payload interface{}) ([]byte, error) {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.apiKey)

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	return body, nil
}

// calculateCost calculates the cost for text generation
func (p *OpenAIProvider) calculateCost(modelID string, promptTokens, completionTokens int) float64 {
	// Simplified pricing calculation
	var inputCost, outputCost float64

	switch modelID {
	case "gpt-4":
		inputCost = 0.03 / 1000  // $0.03 per 1K tokens
		outputCost = 0.06 / 1000 // $0.06 per 1K tokens
	case "gpt-4-turbo":
		inputCost = 0.01 / 1000  // $0.01 per 1K tokens
		outputCost = 0.03 / 1000 // $0.03 per 1K tokens
	case "gpt-3.5-turbo":
		inputCost = 0.0015 / 1000 // $0.0015 per 1K tokens
		outputCost = 0.002 / 1000 // $0.002 per 1K tokens
	default:
		inputCost = 0.002 / 1000 // Default pricing
		outputCost = 0.002 / 1000
	}

	return float64(promptTokens)*inputCost + float64(completionTokens)*outputCost
}

// calculateImageCost calculates the cost for image generation
func (p *OpenAIProvider) calculateImageCost(modelID string, width, height, count int) float64 {
	// Simplified DALL-E pricing
	switch modelID {
	case "dall-e-3":
		if width <= 1024 && height <= 1024 {
			return 0.04 * float64(count) // $0.04 per image
		}
		return 0.08 * float64(count) // $0.08 per image for larger sizes
	case "dall-e-2":
		if width <= 512 && height <= 512 {
			return 0.016 * float64(count) // $0.016 per image
		} else if width <= 1024 && height <= 1024 {
			return 0.018 * float64(count) // $0.018 per image
		}
		return 0.02 * float64(count) // $0.02 per image for largest size
	default:
		return 0.02 * float64(count)
	}
}

// updateUsage updates usage statistics
func (p *OpenAIProvider) updateUsage(latency time.Duration, cost float64) {
	p.usage.RequestsToday++
	p.usage.CostToday += cost
	p.usage.LastRequestTime = time.Now()

	// Update average latency (simplified)
	if p.usage.AverageLatency == 0 {
		p.usage.AverageLatency = latency
	} else {
		p.usage.AverageLatency = (p.usage.AverageLatency + latency) / 2
	}
}

// updateHealth updates health status
func (p *OpenAIProvider) updateHealth(status, issue string) {
	p.health.Status = status
	p.health.LastCheck = time.Now()

	if issue != "" {
		p.health.Issues = []string{issue}
	} else {
		p.health.Issues = nil
	}
}
