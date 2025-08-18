package llm

import (
	"context"
	"time"
)

// LLMProvider represents different LLM providers
type LLMProvider string

const (
	ProviderOpenAI    LLMProvider = "openai"
	ProviderOllama    LLMProvider = "ollama"
	ProviderAnthropic LLMProvider = "anthropic"
	ProviderGemini    LLMProvider = "gemini"
)

// Message represents a single message in a conversation
type Message struct {
	Role      string                 `json:"role"`      // "system", "user", "assistant"
	Content   string                 `json:"content"`   // The message content
	Metadata  map[string]interface{} `json:"metadata"`  // Additional metadata
	Timestamp time.Time              `json:"timestamp"` // When the message was created
}

// CompletionRequest represents a request for LLM completion
type CompletionRequest struct {
	Messages    []Message              `json:"messages"`
	Model       string                 `json:"model"`
	MaxTokens   int                    `json:"max_tokens,omitempty"`
	Temperature float64                `json:"temperature,omitempty"`
	TopP        float64                `json:"top_p,omitempty"`
	Stop        []string               `json:"stop,omitempty"`
	Stream      bool                   `json:"stream,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// CompletionResponse represents the response from an LLM
type CompletionResponse struct {
	ID       string                 `json:"id"`
	Content  string                 `json:"content"`
	Model    string                 `json:"model"`
	Usage    TokenUsage             `json:"usage"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
	Error    error                  `json:"error,omitempty"`
}

// TokenUsage represents token usage statistics
type TokenUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// StreamResponse represents a streaming response chunk
type StreamResponse struct {
	ID      string `json:"id"`
	Content string `json:"content"`
	Done    bool   `json:"done"`
	Error   error  `json:"error,omitempty"`
}

// EmbeddingRequest represents a request for text embeddings
type EmbeddingRequest struct {
	Input []string `json:"input"`
	Model string   `json:"model"`
}

// EmbeddingResponse represents the response from an embedding request
type EmbeddingResponse struct {
	Embeddings [][]float64 `json:"embeddings"`
	Model      string      `json:"model"`
	Usage      TokenUsage  `json:"usage"`
}

// LLM defines the interface for Language Model providers
type LLM interface {
	// Complete generates a completion for the given request
	Complete(ctx context.Context, req *CompletionRequest) (*CompletionResponse, error)

	// Stream generates a streaming completion for the given request
	Stream(ctx context.Context, req *CompletionRequest) (<-chan StreamResponse, error)

	// GetEmbeddings generates embeddings for the given text
	GetEmbeddings(ctx context.Context, req *EmbeddingRequest) (*EmbeddingResponse, error)

	// GetProvider returns the provider type
	GetProvider() LLMProvider

	// GetModels returns available models for this provider
	GetModels(ctx context.Context) ([]string, error)

	// ValidateModel checks if a model is available
	ValidateModel(ctx context.Context, model string) error

	// Close closes the LLM client and cleans up resources
	Close() error
}

// LLMConfig represents configuration for an LLM provider
type LLMConfig struct {
	Provider    LLMProvider            `mapstructure:"provider"`
	APIKey      string                 `mapstructure:"api_key"`
	BaseURL     string                 `mapstructure:"base_url"`
	Model       string                 `mapstructure:"model"`
	MaxTokens   int                    `mapstructure:"max_tokens"`
	Temperature float64                `mapstructure:"temperature"`
	TopP        float64                `mapstructure:"top_p"`
	Timeout     time.Duration          `mapstructure:"timeout"`
	RetryCount  int                    `mapstructure:"retry_count"`
	Metadata    map[string]interface{} `mapstructure:"metadata"`
}

// LLMFactory creates LLM instances based on configuration
type LLMFactory interface {
	// CreateLLM creates an LLM instance for the given provider
	CreateLLM(config *LLMConfig) (LLM, error)

	// RegisterProvider registers a custom LLM provider
	RegisterProvider(provider LLMProvider, factory func(*LLMConfig) (LLM, error)) error

	// GetSupportedProviders returns list of supported providers
	GetSupportedProviders() []LLMProvider
}

// LLMManager manages multiple LLM instances and provides load balancing
type LLMManager interface {
	// AddLLM adds an LLM instance to the manager
	AddLLM(name string, llm LLM) error

	// GetLLM gets an LLM instance by name
	GetLLM(name string) (LLM, error)

	// GetDefaultLLM gets the default LLM instance
	GetDefaultLLM() (LLM, error)

	// Complete routes completion request to appropriate LLM
	Complete(ctx context.Context, req *CompletionRequest) (*CompletionResponse, error)

	// Stream routes streaming request to appropriate LLM
	Stream(ctx context.Context, req *CompletionRequest) (<-chan StreamResponse, error)

	// GetEmbeddings routes embedding request to appropriate LLM
	GetEmbeddings(ctx context.Context, req *EmbeddingRequest) (*EmbeddingResponse, error)

	// GetHealthStatus returns health status of all managed LLMs
	GetHealthStatus(ctx context.Context) map[string]bool

	// Close closes all managed LLM instances
	Close() error
}

// LLMMetrics defines metrics collection interface for LLM operations
type LLMMetrics interface {
	// RecordCompletion records completion metrics
	RecordCompletion(provider LLMProvider, model string, duration time.Duration, tokens int, success bool)

	// RecordEmbedding records embedding metrics
	RecordEmbedding(provider LLMProvider, model string, duration time.Duration, inputCount int, success bool)

	// GetCompletionStats returns completion statistics
	GetCompletionStats(provider LLMProvider) map[string]interface{}

	// GetEmbeddingStats returns embedding statistics
	GetEmbeddingStats(provider LLMProvider) map[string]interface{}
}
