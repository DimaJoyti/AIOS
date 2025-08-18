package vectordb

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

// OpenAIEmbeddingProvider implements EmbeddingProvider for OpenAI
type OpenAIEmbeddingProvider struct {
	config     *EmbeddingConfig
	httpClient *http.Client
	logger     *logrus.Logger
	tracer     trace.Tracer
}

// OpenAIEmbeddingFactory implements EmbeddingProviderFactory for OpenAI
type OpenAIEmbeddingFactory struct{}

// NewOpenAIEmbeddingFactory creates a new OpenAI embedding factory
func NewOpenAIEmbeddingFactory() *OpenAIEmbeddingFactory {
	return &OpenAIEmbeddingFactory{}
}

// Create creates a new OpenAI embedding provider
func (f *OpenAIEmbeddingFactory) Create(config *EmbeddingConfig) (EmbeddingProvider, error) {
	if err := f.ValidateConfig(config); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	logger := logrus.New()
	if config.Metadata != nil {
		if logLevel, exists := config.Metadata["log_level"]; exists {
			if level, ok := logLevel.(string); ok {
				if parsedLevel, err := logrus.ParseLevel(level); err == nil {
					logger.SetLevel(parsedLevel)
				}
			}
		}
	}

	httpClient := &http.Client{
		Timeout: config.Timeout,
	}

	return &OpenAIEmbeddingProvider{
		config:     config,
		httpClient: httpClient,
		logger:     logger,
		tracer:     otel.Tracer("vectordb.embeddings.openai"),
	}, nil
}

// GetProviderName returns the provider name
func (f *OpenAIEmbeddingFactory) GetProviderName() string {
	return "openai"
}

// ValidateConfig validates the OpenAI embedding configuration
func (f *OpenAIEmbeddingFactory) ValidateConfig(config *EmbeddingConfig) error {
	if config.APIKey == "" {
		return fmt.Errorf("API key is required for OpenAI")
	}
	if config.Model == "" {
		config.Model = "text-embedding-ada-002"
	}
	if config.BaseURL == "" {
		config.BaseURL = "https://api.openai.com/v1"
	}
	if config.Timeout <= 0 {
		config.Timeout = 30 * time.Second
	}
	if config.BatchSize <= 0 {
		config.BatchSize = 100
	}
	if config.MaxTokens <= 0 {
		config.MaxTokens = 8191
	}

	// Set dimensions based on model
	if config.Dimensions <= 0 {
		switch config.Model {
		case "text-embedding-ada-002":
			config.Dimensions = 1536
		case "text-embedding-3-small":
			config.Dimensions = 1536
		case "text-embedding-3-large":
			config.Dimensions = 3072
		default:
			config.Dimensions = 1536 // Default
		}
	}

	return nil
}

// GenerateEmbedding generates an embedding for a single text
func (p *OpenAIEmbeddingProvider) GenerateEmbedding(ctx context.Context, text string) ([]float32, error) {
	embeddings, err := p.GenerateEmbeddings(ctx, []string{text})
	if err != nil {
		return nil, err
	}
	if len(embeddings) == 0 {
		return nil, fmt.Errorf("no embedding generated")
	}
	return embeddings[0], nil
}

// GenerateEmbeddings generates embeddings for multiple texts
func (p *OpenAIEmbeddingProvider) GenerateEmbeddings(ctx context.Context, texts []string) ([][]float32, error) {
	ctx, span := p.tracer.Start(ctx, "openai_embedding.generate_embeddings")
	defer span.End()

	span.SetAttributes(
		attribute.String("embedding.model", p.config.Model),
		attribute.Int("embedding.text_count", len(texts)),
	)

	if len(texts) == 0 {
		return [][]float32{}, nil
	}

	// Process in batches
	var allEmbeddings [][]float32
	batchSize := p.config.BatchSize

	for i := 0; i < len(texts); i += batchSize {
		end := i + batchSize
		if end > len(texts) {
			end = len(texts)
		}

		batch := texts[i:end]
		embeddings, err := p.generateBatch(ctx, batch)
		if err != nil {
			span.RecordError(err)
			return nil, fmt.Errorf("failed to generate embeddings for batch %d-%d: %w", i, end, err)
		}

		allEmbeddings = append(allEmbeddings, embeddings...)
	}

	p.logger.WithFields(logrus.Fields{
		"model":      p.config.Model,
		"text_count": len(texts),
		"batches":    (len(texts) + batchSize - 1) / batchSize,
	}).Debug("Generated embeddings")

	return allEmbeddings, nil
}

// GetDimensions returns the embedding dimensions
func (p *OpenAIEmbeddingProvider) GetDimensions() int {
	return p.config.Dimensions
}

// GetModel returns the model name
func (p *OpenAIEmbeddingProvider) GetModel() string {
	return p.config.Model
}

// GetMaxTokens returns the maximum tokens
func (p *OpenAIEmbeddingProvider) GetMaxTokens() int {
	return p.config.MaxTokens
}

func (p *OpenAIEmbeddingProvider) generateBatch(ctx context.Context, texts []string) ([][]float32, error) {
	payload := map[string]interface{}{
		"input": texts,
		"model": p.config.Model,
	}

	// Add dimensions for newer models
	if strings.Contains(p.config.Model, "text-embedding-3") {
		payload["dimensions"] = p.config.Dimensions
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := p.config.BaseURL + "/embeddings"
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.config.APIKey)

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(responseBody))
	}

	var response struct {
		Data []struct {
			Embedding []float32 `json:"embedding"`
			Index     int       `json:"index"`
		} `json:"data"`
		Usage struct {
			TotalTokens int `json:"total_tokens"`
		} `json:"usage"`
	}

	if err := json.Unmarshal(responseBody, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	embeddings := make([][]float32, len(response.Data))
	for _, item := range response.Data {
		embeddings[item.Index] = item.Embedding
	}

	return embeddings, nil
}

// OllamaEmbeddingProvider implements EmbeddingProvider for Ollama
type OllamaEmbeddingProvider struct {
	config     *EmbeddingConfig
	httpClient *http.Client
	logger     *logrus.Logger
	tracer     trace.Tracer
}

// OllamaEmbeddingFactory implements EmbeddingProviderFactory for Ollama
type OllamaEmbeddingFactory struct{}

// NewOllamaEmbeddingFactory creates a new Ollama embedding factory
func NewOllamaEmbeddingFactory() *OllamaEmbeddingFactory {
	return &OllamaEmbeddingFactory{}
}

// Create creates a new Ollama embedding provider
func (f *OllamaEmbeddingFactory) Create(config *EmbeddingConfig) (EmbeddingProvider, error) {
	if err := f.ValidateConfig(config); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	logger := logrus.New()
	httpClient := &http.Client{
		Timeout: config.Timeout,
	}

	return &OllamaEmbeddingProvider{
		config:     config,
		httpClient: httpClient,
		logger:     logger,
		tracer:     otel.Tracer("vectordb.embeddings.ollama"),
	}, nil
}

// GetProviderName returns the provider name
func (f *OllamaEmbeddingFactory) GetProviderName() string {
	return "ollama"
}

// ValidateConfig validates the Ollama embedding configuration
func (f *OllamaEmbeddingFactory) ValidateConfig(config *EmbeddingConfig) error {
	if config.Model == "" {
		config.Model = "nomic-embed-text"
	}
	if config.BaseURL == "" {
		config.BaseURL = "http://localhost:11434"
	}
	if config.Timeout <= 0 {
		config.Timeout = 60 * time.Second // Ollama can be slower
	}
	if config.BatchSize <= 0 {
		config.BatchSize = 10 // Smaller batches for local models
	}

	// Set dimensions based on model
	if config.Dimensions <= 0 {
		switch config.Model {
		case "nomic-embed-text":
			config.Dimensions = 768
		case "all-minilm":
			config.Dimensions = 384
		default:
			config.Dimensions = 768 // Default
		}
	}

	return nil
}

// GenerateEmbedding generates an embedding for a single text
func (p *OllamaEmbeddingProvider) GenerateEmbedding(ctx context.Context, text string) ([]float32, error) {
	ctx, span := p.tracer.Start(ctx, "ollama_embedding.generate_embedding")
	defer span.End()

	span.SetAttributes(
		attribute.String("embedding.model", p.config.Model),
		attribute.Int("embedding.text_length", len(text)),
	)

	payload := map[string]interface{}{
		"model":  p.config.Model,
		"prompt": text,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := p.config.BaseURL + "/api/embeddings"
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		span.RecordError(fmt.Errorf("API error: %s", string(responseBody)))
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(responseBody))
	}

	var response struct {
		Embedding []float32 `json:"embedding"`
	}

	if err := json.Unmarshal(responseBody, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return response.Embedding, nil
}

// GenerateEmbeddings generates embeddings for multiple texts
func (p *OllamaEmbeddingProvider) GenerateEmbeddings(ctx context.Context, texts []string) ([][]float32, error) {
	embeddings := make([][]float32, len(texts))

	// Process sequentially for Ollama (local model)
	for i, text := range texts {
		embedding, err := p.GenerateEmbedding(ctx, text)
		if err != nil {
			return nil, fmt.Errorf("failed to generate embedding for text %d: %w", i, err)
		}
		embeddings[i] = embedding
	}

	p.logger.WithFields(logrus.Fields{
		"model":      p.config.Model,
		"text_count": len(texts),
	}).Debug("Generated embeddings")

	return embeddings, nil
}

// GetDimensions returns the embedding dimensions
func (p *OllamaEmbeddingProvider) GetDimensions() int {
	return p.config.Dimensions
}

// GetModel returns the model name
func (p *OllamaEmbeddingProvider) GetModel() string {
	return p.config.Model
}

// GetMaxTokens returns the maximum tokens
func (p *OllamaEmbeddingProvider) GetMaxTokens() int {
	if p.config.MaxTokens > 0 {
		return p.config.MaxTokens
	}
	return 2048 // Default for local models
}

// EmbeddingManager manages embedding providers
type EmbeddingManager struct {
	factories map[string]EmbeddingProviderFactory
	logger    *logrus.Logger
}

// NewEmbeddingManager creates a new embedding manager
func NewEmbeddingManager(logger *logrus.Logger) *EmbeddingManager {
	manager := &EmbeddingManager{
		factories: make(map[string]EmbeddingProviderFactory),
		logger:    logger,
	}

	// Register default providers
	manager.RegisterFactory("openai", NewOpenAIEmbeddingFactory())
	manager.RegisterFactory("ollama", NewOllamaEmbeddingFactory())

	return manager
}

// RegisterFactory registers an embedding provider factory
func (m *EmbeddingManager) RegisterFactory(name string, factory EmbeddingProviderFactory) error {
	if factory == nil {
		return fmt.Errorf("factory cannot be nil")
	}
	m.factories[name] = factory
	m.logger.WithField("provider", name).Debug("Registered embedding provider factory")
	return nil
}

// CreateProvider creates an embedding provider
func (m *EmbeddingManager) CreateProvider(config *EmbeddingConfig) (EmbeddingProvider, error) {
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	factory, exists := m.factories[config.Provider]
	if !exists {
		return nil, fmt.Errorf("unknown embedding provider: %s", config.Provider)
	}

	return factory.Create(config)
}

// GetProviders returns all registered provider names
func (m *EmbeddingManager) GetProviders() []string {
	providers := make([]string, 0, len(m.factories))
	for name := range m.factories {
		providers = append(providers, name)
	}
	return providers
}

// ValidateConfig validates an embedding configuration
func (m *EmbeddingManager) ValidateConfig(config *EmbeddingConfig) error {
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}

	factory, exists := m.factories[config.Provider]
	if !exists {
		return fmt.Errorf("unknown embedding provider: %s", config.Provider)
	}

	return factory.ValidateConfig(config)
}
