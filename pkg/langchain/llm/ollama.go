package llm

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
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// OllamaLLM implements the LLM interface for Ollama
type OllamaLLM struct {
	config     *LLMConfig
	httpClient *http.Client
	logger     *logrus.Logger
	tracer     trace.Tracer
}

// OllamaMessage represents an Ollama API message
type OllamaMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// OllamaChatRequest represents an Ollama chat request
type OllamaChatRequest struct {
	Model    string                 `json:"model"`
	Messages []OllamaMessage        `json:"messages"`
	Stream   bool                   `json:"stream,omitempty"`
	Options  map[string]interface{} `json:"options,omitempty"`
}

// OllamaChatResponse represents an Ollama chat response
type OllamaChatResponse struct {
	Model              string        `json:"model"`
	CreatedAt          time.Time     `json:"created_at"`
	Message            OllamaMessage `json:"message"`
	Done               bool          `json:"done"`
	TotalDuration      int64         `json:"total_duration,omitempty"`
	LoadDuration       int64         `json:"load_duration,omitempty"`
	PromptEvalCount    int           `json:"prompt_eval_count,omitempty"`
	PromptEvalDuration int64         `json:"prompt_eval_duration,omitempty"`
	EvalCount          int           `json:"eval_count,omitempty"`
	EvalDuration       int64         `json:"eval_duration,omitempty"`
}

// OllamaEmbeddingRequest represents an Ollama embedding request
type OllamaEmbeddingRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
}

// OllamaEmbeddingResponse represents an Ollama embedding response
type OllamaEmbeddingResponse struct {
	Embedding []float64 `json:"embedding"`
}

// OllamaModelsResponse represents the response from the models endpoint
type OllamaModelsResponse struct {
	Models []struct {
		Name       string    `json:"name"`
		ModifiedAt time.Time `json:"modified_at"`
		Size       int64     `json:"size"`
		Digest     string    `json:"digest"`
		Details    struct {
			Format            string   `json:"format"`
			Family            string   `json:"family"`
			Families          []string `json:"families"`
			ParameterSize     string   `json:"parameter_size"`
			QuantizationLevel string   `json:"quantization_level"`
		} `json:"details"`
	} `json:"models"`
}

// NewOllamaLLM creates a new Ollama LLM instance
func NewOllamaLLM(config *LLMConfig, logger *logrus.Logger) (LLM, error) {
	if config.BaseURL == "" {
		config.BaseURL = "http://localhost:11434"
	}

	if config.Model == "" {
		config.Model = "llama2"
	}

	if config.Timeout == 0 {
		config.Timeout = 60 * time.Second // Ollama can be slower
	}

	httpClient := &http.Client{
		Timeout: config.Timeout,
	}

	return &OllamaLLM{
		config:     config,
		httpClient: httpClient,
		logger:     logger,
		tracer:     otel.Tracer("langchain.llm.ollama"),
	}, nil
}

// Complete generates a completion for the given request
func (o *OllamaLLM) Complete(ctx context.Context, req *CompletionRequest) (*CompletionResponse, error) {
	ctx, span := o.tracer.Start(ctx, "ollama.complete")
	defer span.End()

	span.SetAttributes(
		attribute.String("llm.provider", string(ProviderOllama)),
		attribute.String("llm.model", req.Model),
		attribute.Int("llm.max_tokens", req.MaxTokens),
	)

	// Convert to Ollama format
	ollamaReq := o.convertToOllamaRequest(req)

	// Make API request
	respBody, err := o.makeRequest(ctx, "/api/chat", ollamaReq)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to make Ollama request: %w", err)
	}

	// Parse response
	var ollamaResp OllamaChatResponse
	if err := json.Unmarshal(respBody, &ollamaResp); err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to parse Ollama response: %w", err)
	}

	// Convert to standard format
	response := o.convertFromOllamaResponse(&ollamaResp)

	span.SetAttributes(
		attribute.Int("llm.usage.prompt_tokens", response.Usage.PromptTokens),
		attribute.Int("llm.usage.completion_tokens", response.Usage.CompletionTokens),
		attribute.Int("llm.usage.total_tokens", response.Usage.TotalTokens),
	)

	return response, nil
}

// Stream generates a streaming completion for the given request
func (o *OllamaLLM) Stream(ctx context.Context, req *CompletionRequest) (<-chan StreamResponse, error) {
	ctx, span := o.tracer.Start(ctx, "ollama.stream")
	defer span.End()

	responseCh := make(chan StreamResponse, 10)

	go func() {
		defer close(responseCh)

		// Convert to Ollama format with streaming enabled
		ollamaReq := o.convertToOllamaRequest(req)
		ollamaReq.Stream = true

		// Make streaming request
		resp, err := o.makeStreamingRequest(ctx, "/api/chat", ollamaReq)
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
func (o *OllamaLLM) GetEmbeddings(ctx context.Context, req *EmbeddingRequest) (*EmbeddingResponse, error) {
	ctx, span := o.tracer.Start(ctx, "ollama.embeddings")
	defer span.End()

	model := req.Model
	if model == "" {
		model = "nomic-embed-text" // Default embedding model for Ollama
	}

	embeddings := make([][]float64, len(req.Input))

	// Ollama processes embeddings one at a time
	for i, text := range req.Input {
		ollamaReq := OllamaEmbeddingRequest{
			Model:  model,
			Prompt: text,
		}

		// Make API request
		respBody, err := o.makeRequest(ctx, "/api/embeddings", ollamaReq)
		if err != nil {
			span.RecordError(err)
			return nil, fmt.Errorf("failed to make Ollama embeddings request: %w", err)
		}

		// Parse response
		var ollamaResp OllamaEmbeddingResponse
		if err := json.Unmarshal(respBody, &ollamaResp); err != nil {
			span.RecordError(err)
			return nil, fmt.Errorf("failed to parse Ollama embeddings response: %w", err)
		}

		embeddings[i] = ollamaResp.Embedding
	}

	return &EmbeddingResponse{
		Embeddings: embeddings,
		Model:      model,
		Usage: TokenUsage{
			// Ollama doesn't provide token usage for embeddings
			PromptTokens: len(req.Input),
			TotalTokens:  len(req.Input),
		},
	}, nil
}

// GetProvider returns the provider type
func (o *OllamaLLM) GetProvider() LLMProvider {
	return ProviderOllama
}

// GetModels returns available models for this provider
func (o *OllamaLLM) GetModels(ctx context.Context) ([]string, error) {
	ctx, span := o.tracer.Start(ctx, "ollama.get_models")
	defer span.End()

	req, err := http.NewRequestWithContext(ctx, "GET", o.config.BaseURL+"/api/tags", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

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

	var modelsResp OllamaModelsResponse
	if err := json.Unmarshal(body, &modelsResp); err != nil {
		return nil, fmt.Errorf("failed to parse models response: %w", err)
	}

	models := make([]string, len(modelsResp.Models))
	for i, model := range modelsResp.Models {
		models[i] = model.Name
	}

	return models, nil
}

// ValidateModel checks if a model is available
func (o *OllamaLLM) ValidateModel(ctx context.Context, model string) error {
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
func (o *OllamaLLM) Close() error {
	// Nothing to close for HTTP client
	return nil
}

// Helper methods

func (o *OllamaLLM) convertToOllamaRequest(req *CompletionRequest) *OllamaChatRequest {
	messages := make([]OllamaMessage, len(req.Messages))
	for i, msg := range req.Messages {
		messages[i] = OllamaMessage{
			Role:    msg.Role,
			Content: msg.Content,
		}
	}

	model := req.Model
	if model == "" {
		model = o.config.Model
	}

	options := make(map[string]interface{})
	if req.MaxTokens > 0 {
		options["num_predict"] = req.MaxTokens
	}
	if req.Temperature > 0 {
		options["temperature"] = req.Temperature
	}
	if req.TopP > 0 {
		options["top_p"] = req.TopP
	}

	return &OllamaChatRequest{
		Model:    model,
		Messages: messages,
		Stream:   req.Stream,
		Options:  options,
	}
}

func (o *OllamaLLM) convertFromOllamaResponse(resp *OllamaChatResponse) *CompletionResponse {
	return &CompletionResponse{
		ID:      fmt.Sprintf("ollama-%d", time.Now().Unix()),
		Content: resp.Message.Content,
		Model:   resp.Model,
		Usage: TokenUsage{
			PromptTokens:     resp.PromptEvalCount,
			CompletionTokens: resp.EvalCount,
			TotalTokens:      resp.PromptEvalCount + resp.EvalCount,
		},
	}
}

func (o *OllamaLLM) makeRequest(ctx context.Context, endpoint string, payload interface{}) ([]byte, error) {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", o.config.BaseURL+endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

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

func (o *OllamaLLM) makeStreamingRequest(ctx context.Context, endpoint string, payload interface{}) (*http.Response, error) {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", o.config.BaseURL+endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	return o.httpClient.Do(req)
}

func (o *OllamaLLM) processStreamingResponse(body io.Reader, responseCh chan<- StreamResponse) {
	decoder := json.NewDecoder(body)

	for {
		var resp OllamaChatResponse
		if err := decoder.Decode(&resp); err != nil {
			if err != io.EOF {
				responseCh <- StreamResponse{Error: err}
			}
			break
		}

		responseCh <- StreamResponse{
			ID:      fmt.Sprintf("ollama-%d", time.Now().Unix()),
			Content: resp.Message.Content,
			Done:    resp.Done,
		}

		if resp.Done {
			break
		}
	}
}
