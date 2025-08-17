package providers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

// OllamaClient provides integration with Ollama API
type OllamaClient struct {
	baseURL    string
	httpClient *http.Client
	logger     *logrus.Logger
	tracer     trace.Tracer
}

// OllamaRequest represents a request to Ollama API
type OllamaRequest struct {
	Model    string                 `json:"model"`
	Prompt   string                 `json:"prompt,omitempty"`
	Messages []OllamaMessage        `json:"messages,omitempty"`
	Stream   bool                   `json:"stream"`
	Options  map[string]interface{} `json:"options,omitempty"`
	Format   string                 `json:"format,omitempty"`
	System   string                 `json:"system,omitempty"`
}

// OllamaMessage represents a message in Ollama chat format
type OllamaMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// OllamaResponse represents a response from Ollama API
type OllamaResponse struct {
	Model              string    `json:"model"`
	CreatedAt          time.Time `json:"created_at"`
	Response           string    `json:"response"`
	Done               bool      `json:"done"`
	Context            []int     `json:"context,omitempty"`
	TotalDuration      int64     `json:"total_duration,omitempty"`
	LoadDuration       int64     `json:"load_duration,omitempty"`
	PromptEvalCount    int       `json:"prompt_eval_count,omitempty"`
	PromptEvalDuration int64     `json:"prompt_eval_duration,omitempty"`
	EvalCount          int       `json:"eval_count,omitempty"`
	EvalDuration       int64     `json:"eval_duration,omitempty"`
}

// OllamaStreamResponse represents a streaming response chunk
type OllamaStreamResponse struct {
	Model     string    `json:"model"`
	CreatedAt time.Time `json:"created_at"`
	Response  string    `json:"response"`
	Done      bool      `json:"done"`
	Context   []int     `json:"context,omitempty"`
}

// OllamaModelInfo represents model information from Ollama
type OllamaModelInfo struct {
	Name       string            `json:"name"`
	Size       int64             `json:"size"`
	Digest     string            `json:"digest"`
	Details    OllamaModelDetail `json:"details"`
	ModifiedAt time.Time         `json:"modified_at"`
}

// OllamaModelDetail represents detailed model information
type OllamaModelDetail struct {
	Format            string   `json:"format"`
	Family            string   `json:"family"`
	Families          []string `json:"families"`
	ParameterSize     string   `json:"parameter_size"`
	QuantizationLevel string   `json:"quantization_level"`
}

// NewOllamaClient creates a new Ollama client
func NewOllamaClient(baseURL string, logger *logrus.Logger) *OllamaClient {
	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}

	return &OllamaClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 5 * time.Minute, // Long timeout for model inference
		},
		logger: logger,
		tracer: otel.Tracer("ai.ollama_client"),
	}
}

// Generate generates text using Ollama
func (c *OllamaClient) Generate(ctx context.Context, request *OllamaRequest) (*OllamaResponse, error) {
	ctx, span := c.tracer.Start(ctx, "ollama.Generate")
	defer span.End()

	start := time.Now()
	c.logger.WithFields(logrus.Fields{
		"model":  request.Model,
		"prompt": request.Prompt[:min(100, len(request.Prompt))],
	}).Info("Generating text with Ollama")

	// Prepare request
	request.Stream = false
	reqBody, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Make HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/api/generate", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("ollama API error: %d - %s", resp.StatusCode, string(body))
	}

	// Parse response
	var ollamaResp OllamaResponse
	if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	c.logger.WithFields(logrus.Fields{
		"model":           request.Model,
		"response_length": len(ollamaResp.Response),
		"processing_time": time.Since(start),
		"eval_count":      ollamaResp.EvalCount,
	}).Info("Ollama generation completed")

	return &ollamaResp, nil
}

// GenerateStream generates text with streaming using Ollama
func (c *OllamaClient) GenerateStream(ctx context.Context, request *OllamaRequest, callback func(*OllamaStreamResponse) error) error {
	ctx, span := c.tracer.Start(ctx, "ollama.GenerateStream")
	defer span.End()

	start := time.Now()
	c.logger.WithFields(logrus.Fields{
		"model":  request.Model,
		"prompt": request.Prompt[:min(100, len(request.Prompt))],
	}).Info("Starting streaming generation with Ollama")

	// Prepare request
	request.Stream = true
	reqBody, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	// Make HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/api/generate", bytes.NewBuffer(reqBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("ollama API error: %d - %s", resp.StatusCode, string(body))
	}

	// Process streaming response
	decoder := json.NewDecoder(resp.Body)
	chunkCount := 0

	for {
		var chunk OllamaStreamResponse
		if err := decoder.Decode(&chunk); err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("failed to decode stream chunk: %w", err)
		}

		chunkCount++
		if err := callback(&chunk); err != nil {
			return fmt.Errorf("callback error: %w", err)
		}

		if chunk.Done {
			break
		}
	}

	c.logger.WithFields(logrus.Fields{
		"model":           request.Model,
		"chunks_received": chunkCount,
		"processing_time": time.Since(start),
	}).Info("Ollama streaming completed")

	return nil
}

// Chat performs chat completion using Ollama
func (c *OllamaClient) Chat(ctx context.Context, model string, messages []OllamaMessage, options map[string]interface{}) (*OllamaResponse, error) {
	ctx, span := c.tracer.Start(ctx, "ollama.Chat")
	defer span.End()

	start := time.Now()
	c.logger.WithFields(logrus.Fields{
		"model":         model,
		"message_count": len(messages),
	}).Info("Starting chat with Ollama")

	request := &OllamaRequest{
		Model:    model,
		Messages: messages,
		Stream:   false,
		Options:  options,
	}

	reqBody, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/api/chat", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("ollama API error: %d - %s", resp.StatusCode, string(body))
	}

	var ollamaResp OllamaResponse
	if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	c.logger.WithFields(logrus.Fields{
		"model":           model,
		"response_length": len(ollamaResp.Response),
		"processing_time": time.Since(start),
	}).Info("Ollama chat completed")

	return &ollamaResp, nil
}

// ListModels lists available models from Ollama
func (c *OllamaClient) ListModels(ctx context.Context) ([]OllamaModelInfo, error) {
	ctx, span := c.tracer.Start(ctx, "ollama.ListModels")
	defer span.End()

	c.logger.Info("Listing models from Ollama")

	httpReq, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/api/tags", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("ollama API error: %d - %s", resp.StatusCode, string(body))
	}

	var response struct {
		Models []OllamaModelInfo `json:"models"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	c.logger.WithField("model_count", len(response.Models)).Info("Listed models from Ollama")

	return response.Models, nil
}

// PullModel pulls a model from Ollama registry
func (c *OllamaClient) PullModel(ctx context.Context, modelName string) error {
	ctx, span := c.tracer.Start(ctx, "ollama.PullModel")
	defer span.End()

	c.logger.WithField("model", modelName).Info("Pulling model from Ollama")

	request := map[string]string{
		"name": modelName,
	}

	reqBody, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/api/pull", bytes.NewBuffer(reqBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("ollama API error: %d - %s", resp.StatusCode, string(body))
	}

	c.logger.WithField("model", modelName).Info("Model pulled successfully")

	return nil
}

// IsHealthy checks if Ollama service is healthy
func (c *OllamaClient) IsHealthy(ctx context.Context) bool {
	ctx, span := c.tracer.Start(ctx, "ollama.IsHealthy")
	defer span.End()

	httpReq, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/api/tags", nil)
	if err != nil {
		return false
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}

// Helper function
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
