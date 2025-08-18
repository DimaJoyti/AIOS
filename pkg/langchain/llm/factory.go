package llm

import (
	"context"
	"fmt"
	"sync"

	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

// DefaultLLMFactory is the default implementation of LLMFactory
type DefaultLLMFactory struct {
	providers map[LLMProvider]func(*LLMConfig) (LLM, error)
	logger    *logrus.Logger
	tracer    trace.Tracer
	mu        sync.RWMutex
}

// NewLLMFactory creates a new LLM factory with default providers
func NewLLMFactory(logger *logrus.Logger) LLMFactory {
	factory := &DefaultLLMFactory{
		providers: make(map[LLMProvider]func(*LLMConfig) (LLM, error)),
		logger:    logger,
		tracer:    otel.Tracer("langchain.llm.factory"),
	}

	// Register default providers
	factory.registerDefaultProviders()

	return factory
}

// CreateLLM creates an LLM instance for the given provider
func (f *DefaultLLMFactory) CreateLLM(config *LLMConfig) (LLM, error) {
	f.mu.RLock()
	factoryFunc, exists := f.providers[config.Provider]
	f.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("unsupported LLM provider: %s", config.Provider)
	}

	f.logger.WithFields(logrus.Fields{
		"provider": config.Provider,
		"model":    config.Model,
	}).Info("Creating LLM instance")

	llm, err := factoryFunc(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create LLM for provider %s: %w", config.Provider, err)
	}

	return llm, nil
}

// RegisterProvider registers a custom LLM provider
func (f *DefaultLLMFactory) RegisterProvider(provider LLMProvider, factory func(*LLMConfig) (LLM, error)) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if factory == nil {
		return fmt.Errorf("factory function cannot be nil")
	}

	f.providers[provider] = factory
	f.logger.WithField("provider", provider).Info("Registered LLM provider")

	return nil
}

// GetSupportedProviders returns list of supported providers
func (f *DefaultLLMFactory) GetSupportedProviders() []LLMProvider {
	f.mu.RLock()
	defer f.mu.RUnlock()

	providers := make([]LLMProvider, 0, len(f.providers))
	for provider := range f.providers {
		providers = append(providers, provider)
	}

	return providers
}

// registerDefaultProviders registers the default LLM providers
func (f *DefaultLLMFactory) registerDefaultProviders() {
	// Register OpenAI provider
	f.providers[ProviderOpenAI] = func(config *LLMConfig) (LLM, error) {
		return NewOpenAILLM(config, f.logger)
	}

	// Register Ollama provider
	f.providers[ProviderOllama] = func(config *LLMConfig) (LLM, error) {
		return NewOllamaLLM(config, f.logger)
	}

	// Register Anthropic provider
	f.providers[ProviderAnthropic] = func(config *LLMConfig) (LLM, error) {
		return NewAnthropicLLM(config, f.logger)
	}

	// Register Gemini provider
	f.providers[ProviderGemini] = func(config *LLMConfig) (LLM, error) {
		return NewGeminiLLM(config, f.logger)
	}
}

// DefaultLLMManager is the default implementation of LLMManager
type DefaultLLMManager struct {
	llms       map[string]LLM
	defaultLLM string
	factory    LLMFactory
	logger     *logrus.Logger
	tracer     trace.Tracer
	mu         sync.RWMutex
}

// NewLLMManager creates a new LLM manager
func NewLLMManager(factory LLMFactory, logger *logrus.Logger) LLMManager {
	return &DefaultLLMManager{
		llms:    make(map[string]LLM),
		factory: factory,
		logger:  logger,
		tracer:  otel.Tracer("langchain.llm.manager"),
	}
}

// AddLLM adds an LLM instance to the manager
func (m *DefaultLLMManager) AddLLM(name string, llm LLM) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if llm == nil {
		return fmt.Errorf("LLM instance cannot be nil")
	}

	m.llms[name] = llm

	// Set as default if it's the first LLM
	if m.defaultLLM == "" {
		m.defaultLLM = name
	}

	m.logger.WithFields(logrus.Fields{
		"name":     name,
		"provider": llm.GetProvider(),
	}).Info("Added LLM to manager")

	return nil
}

// GetLLM gets an LLM instance by name
func (m *DefaultLLMManager) GetLLM(name string) (LLM, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	llm, exists := m.llms[name]
	if !exists {
		return nil, fmt.Errorf("LLM not found: %s", name)
	}

	return llm, nil
}

// GetDefaultLLM gets the default LLM instance
func (m *DefaultLLMManager) GetDefaultLLM() (LLM, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.defaultLLM == "" {
		return nil, fmt.Errorf("no default LLM configured")
	}

	llm, exists := m.llms[m.defaultLLM]
	if !exists {
		return nil, fmt.Errorf("default LLM not found: %s", m.defaultLLM)
	}

	return llm, nil
}

// Complete routes completion request to appropriate LLM
func (m *DefaultLLMManager) Complete(ctx context.Context, req *CompletionRequest) (*CompletionResponse, error) {
	ctx, span := m.tracer.Start(ctx, "llm_manager.complete")
	defer span.End()

	llm, err := m.GetDefaultLLM()
	if err != nil {
		return nil, fmt.Errorf("failed to get default LLM: %w", err)
	}

	return llm.Complete(ctx, req)
}

// Stream routes streaming request to appropriate LLM
func (m *DefaultLLMManager) Stream(ctx context.Context, req *CompletionRequest) (<-chan StreamResponse, error) {
	ctx, span := m.tracer.Start(ctx, "llm_manager.stream")
	defer span.End()

	llm, err := m.GetDefaultLLM()
	if err != nil {
		return nil, fmt.Errorf("failed to get default LLM: %w", err)
	}

	return llm.Stream(ctx, req)
}

// GetEmbeddings routes embedding request to appropriate LLM
func (m *DefaultLLMManager) GetEmbeddings(ctx context.Context, req *EmbeddingRequest) (*EmbeddingResponse, error) {
	ctx, span := m.tracer.Start(ctx, "llm_manager.embeddings")
	defer span.End()

	llm, err := m.GetDefaultLLM()
	if err != nil {
		return nil, fmt.Errorf("failed to get default LLM: %w", err)
	}

	return llm.GetEmbeddings(ctx, req)
}

// GetHealthStatus returns health status of all managed LLMs
func (m *DefaultLLMManager) GetHealthStatus(ctx context.Context) map[string]bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	status := make(map[string]bool)
	for name, llm := range m.llms {
		// Simple health check by trying to get models
		_, err := llm.GetModels(ctx)
		status[name] = err == nil
	}

	return status
}

// Close closes all managed LLM instances
func (m *DefaultLLMManager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var lastErr error
	for name, llm := range m.llms {
		if err := llm.Close(); err != nil {
			m.logger.WithError(err).WithField("name", name).Error("Failed to close LLM")
			lastErr = err
		}
	}

	m.llms = make(map[string]LLM)
	m.defaultLLM = ""

	return lastErr
}
