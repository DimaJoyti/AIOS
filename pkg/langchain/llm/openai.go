package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// OpenAILLM implements the LLM interface for OpenAI
type OpenAILLM struct {
	config     *LLMConfig
	httpClient *http.Client
	logger     *logrus.Logger
	tracer     trace.Tracer
}

// OpenAIMessage represents an OpenAI API message
type OpenAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// OpenAICompletionRequest represents an OpenAI completion request
type OpenAICompletionRequest struct {
	Model       string          `json:"model"`
	Messages    []OpenAIMessage `json:"messages"`
	MaxTokens   int             `json:"max_tokens,omitempty"`
	Temperature float64         `json:"temperature,omitempty"`
	TopP        float64         `json:"top_p,omitempty"`
	Stop        []string        `json:"stop,omitempty"`
	Stream      bool            `json:"stream,omitempty"`
}

// OpenAICompletionResponse represents an OpenAI completion response
type OpenAICompletionResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index   int `json:"index"`
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

// OpenAIEmbeddingRequest represents an OpenAI embedding request
type OpenAIEmbeddingRequest struct {
	Input []string `json:"input"`
	Model string   `json:"model"`
}

// OpenAIEmbeddingResponse represents an OpenAI embedding response
type OpenAIEmbeddingResponse struct {
	Object string `json:"object"`
	Data   []struct {
		Object    string    `json:"object"`
		Embedding []float64 `json:"embedding"`
		Index     int       `json:"index"`
	} `json:"data"`
	Model string `json:"model"`
	Usage struct {
		PromptTokens int `json:"prompt_tokens"`
		TotalTokens  int `json:"total_tokens"`
	} `json:"usage"`
}

// NewOpenAILLM creates a new OpenAI LLM instance
func NewOpenAILLM(config *LLMConfig, logger *logrus.Logger) (LLM, error) {
	if config.APIKey == "" {
		return nil, fmt.Errorf("OpenAI API key is required")
	}

	if config.BaseURL == "" {
		config.BaseURL = "https://api.openai.com/v1"
	}

	if config.Model == "" {
		config.Model = "gpt-3.5-turbo"
	}

	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}

	httpClient := &http.Client{
		Timeout: config.Timeout,
	}

	return &OpenAILLM{
		config:     config,
		httpClient: httpClient,
		logger:     logger,
		tracer:     otel.Tracer("langchain.llm.openai"),
	}, nil
}

// Complete generates a completion for the given request
func (o *OpenAILLM) Complete(ctx context.Context, req *CompletionRequest) (*CompletionResponse, error) {
	ctx, span := o.tracer.Start(ctx, "openai.complete")
	defer span.End()

	span.SetAttributes(
		attribute.String("llm.provider", string(ProviderOpenAI)),
		attribute.String("llm.model", req.Model),
		attribute.Int("llm.max_tokens", req.MaxTokens),
	)

	// Convert to OpenAI format
	openaiReq := o.convertToOpenAIRequest(req)

	// Make API request
	respBody, err := o.makeRequest(ctx, "/chat/completions", openaiReq)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to make OpenAI request: %w", err)
	}

	// Parse response
	var openaiResp OpenAICompletionResponse
	if err := json.Unmarshal(respBody, &openaiResp); err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to parse OpenAI response: %w", err)
	}

	// Convert to standard format
	response := o.convertFromOpenAIResponse(&openaiResp)
	
	span.SetAttributes(
		attribute.Int("llm.usage.prompt_tokens", response.Usage.PromptTokens),
		attribute.Int("llm.usage.completion_tokens", response.Usage.CompletionTokens),
		attribute.Int("llm.usage.total_tokens", response.Usage.TotalTokens),
	)

	return response, nil
}

// Stream generates a streaming completion for the given request
func (o *OpenAILLM) Stream(ctx context.Context, req *CompletionRequest) (<-chan StreamResponse, error) {
	ctx, span := o.tracer.Start(ctx, "openai.stream")
	defer span.End()

	responseCh := make(chan StreamResponse, 10)

	go func() {
		defer close(responseCh)

		// Convert to OpenAI format with streaming enabled
		openaiReq := o.convertToOpenAIRequest(req)
		openaiReq.Stream = true

		// Make streaming request
		resp, err := o.makeStreamingRequest(ctx, "/chat/completions", openaiReq)
		if err != nil {
			responseCh <- StreamResponse{Error: err}
			return
		}
		defer resp.Body.Close()

		// Process streaming response
		o.processStreamingResponse(resp.Body, responseCh)
	}()

	return responseCh, nil
}

// GetEmbeddings generates embeddings for the given text
func (o *OpenAILLM) GetEmbeddings(ctx context.Context, req *EmbeddingRequest) (*EmbeddingResponse, error) {
	ctx, span := o.tracer.Start(ctx, "openai.embeddings")
	defer span.End()

	model := req.Model
	if model == "" {
		model = "text-embedding-ada-002"
	}

	openaiReq := OpenAIEmbeddingRequest{
		Input: req.Input,
		Model: model,
	}

	// Make API request
	respBody, err := o.makeRequest(ctx, "/embeddings", openaiReq)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to make OpenAI embeddings request: %w", err)
	}

	// Parse response
	var openaiResp OpenAIEmbeddingResponse
	if err := json.Unmarshal(respBody, &openaiResp); err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to parse OpenAI embeddings response: %w", err)
	}

	// Convert to standard format
	embeddings := make([][]float64, len(openaiResp.Data))
	for i, data := range openaiResp.Data {
		embeddings[i] = data.Embedding
	}

	return &EmbeddingResponse{
		Embeddings: embeddings,
		Model:      openaiResp.Model,
		Usage: TokenUsage{
			PromptTokens: openaiResp.Usage.PromptTokens,
			TotalTokens:  openaiResp.Usage.TotalTokens,
		},
	}, nil
}

// GetProvider returns the provider type
func (o *OpenAILLM) GetProvider() LLMProvider {
	return ProviderOpenAI
}

// GetModels returns available models for this provider
func (o *OpenAILLM) GetModels(ctx context.Context) ([]string, error) {
	// For OpenAI, return common models
	return []string{
		"gpt-4",
		"gpt-4-turbo",
		"gpt-3.5-turbo",
		"text-embedding-ada-002",
		"text-embedding-3-small",
		"text-embedding-3-large",
	}, nil
}

// ValidateModel checks if a model is available
func (o *OpenAILLM) ValidateModel(ctx context.Context, model string) error {
	models, err := o.GetModels(ctx)
	if err != nil {
		return err
	}

	for _, m := range models {
		if m == model {
			return nil
		}
	}

	return fmt.Errorf("model %s is not available", model)
}

// Close closes the LLM client and cleans up resources
func (o *OpenAILLM) Close() error {
	// Nothing to close for HTTP client
	return nil
}

// Helper methods

func (o *OpenAILLM) convertToOpenAIRequest(req *CompletionRequest) *OpenAICompletionRequest {
	messages := make([]OpenAIMessage, len(req.Messages))
	for i, msg := range req.Messages {
		messages[i] = OpenAIMessage{
			Role:    msg.Role,
			Content: msg.Content,
		}
	}

	model := req.Model
	if model == "" {
		model = o.config.Model
	}

	return &OpenAICompletionRequest{
		Model:       model,
		Messages:    messages,
		MaxTokens:   req.MaxTokens,
		Temperature: req.Temperature,
		TopP:        req.TopP,
		Stop:        req.Stop,
		Stream:      req.Stream,
	}
}

func (o *OpenAILLM) convertFromOpenAIResponse(resp *OpenAICompletionResponse) *CompletionResponse {
	content := ""
	if len(resp.Choices) > 0 {
		content = resp.Choices[0].Message.Content
	}

	return &CompletionResponse{
		ID:      resp.ID,
		Content: content,
		Model:   resp.Model,
		Usage: TokenUsage{
			PromptTokens:     resp.Usage.PromptTokens,
			CompletionTokens: resp.Usage.CompletionTokens,
			TotalTokens:      resp.Usage.TotalTokens,
		},
	}
}

func (o *OpenAILLM) makeRequest(ctx context.Context, endpoint string, payload interface{}) ([]byte, error) {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", o.config.BaseURL+endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+o.config.APIKey)

	resp, err := o.httpClient.Do(req)
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

func (o *OpenAILLM) makeStreamingRequest(ctx context.Context, endpoint string, payload interface{}) (*http.Response, error) {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", o.config.BaseURL+endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+o.config.APIKey)
	req.Header.Set("Accept", "text/event-stream")

	return o.httpClient.Do(req)
}

func (o *OpenAILLM) processStreamingResponse(body io.Reader, responseCh chan<- StreamResponse) {
	// Simple streaming response processing
	// In a real implementation, you would parse Server-Sent Events properly
	buf := make([]byte, 4096)
	for {
		n, err := body.Read(buf)
		if err != nil {
			if err != io.EOF {
				responseCh <- StreamResponse{Error: err}
			}
			break
		}

		content := string(buf[:n])
		if strings.Contains(content, "data: [DONE]") {
			responseCh <- StreamResponse{Done: true}
			break
		}

		// Parse and send chunk
		responseCh <- StreamResponse{
			Content: content,
			Done:    false,
		}
	}
}
