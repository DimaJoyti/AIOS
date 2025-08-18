package vectordb

import (
	"fmt"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

// DefaultVectorDBManager implements VectorDBManager interface
type DefaultVectorDBManager struct {
	vectorDBFactories  map[string]VectorDBFactory
	embeddingFactories map[string]EmbeddingProviderFactory
	logger             *logrus.Logger
	tracer             trace.Tracer
	mu                 sync.RWMutex
}

// NewVectorDBManager creates a new vector database manager
func NewVectorDBManager(logger *logrus.Logger) VectorDBManager {
	manager := &DefaultVectorDBManager{
		vectorDBFactories:  make(map[string]VectorDBFactory),
		embeddingFactories: make(map[string]EmbeddingProviderFactory),
		logger:             logger,
		tracer:             otel.Tracer("vectordb.manager"),
	}

	// Register default providers
	manager.RegisterProvider("qdrant", NewQdrantFactory())

	// Register default embedding providers
	manager.registerEmbeddingFactory("openai", NewOpenAIEmbeddingFactory())
	manager.registerEmbeddingFactory("ollama", NewOllamaEmbeddingFactory())

	return manager
}

// RegisterProvider registers a vector database provider
func (m *DefaultVectorDBManager) RegisterProvider(name string, factory VectorDBFactory) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if factory == nil {
		return fmt.Errorf("factory cannot be nil")
	}

	m.vectorDBFactories[name] = factory
	m.logger.WithField("provider", name).Info("Registered vector database provider")

	return nil
}

// CreateVectorDB creates a vector database instance
func (m *DefaultVectorDBManager) CreateVectorDB(config *VectorDBConfig) (VectorDB, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	factory, exists := m.vectorDBFactories[config.Provider]
	if !exists {
		return nil, fmt.Errorf("unknown vector database provider: %s", config.Provider)
	}

	vectorDB, err := factory.Create(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create vector database: %w", err)
	}

	m.logger.WithFields(logrus.Fields{
		"provider": config.Provider,
		"host":     config.Host,
		"port":     config.Port,
	}).Info("Created vector database instance")

	return vectorDB, nil
}

// CreateVectorStore creates a vector store instance
func (m *DefaultVectorDBManager) CreateVectorStore(config *VectorStoreConfig) (VectorStore, error) {
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	// Create vector database
	vectorDB, err := m.CreateVectorDB(config.VectorDB)
	if err != nil {
		return nil, fmt.Errorf("failed to create vector database: %w", err)
	}

	// Create embedding provider
	embeddingProvider, err := m.createEmbeddingProvider(config.Embedding)
	if err != nil {
		return nil, fmt.Errorf("failed to create embedding provider: %w", err)
	}

	// Create vector store
	vectorStore := NewVectorStore(vectorDB, embeddingProvider, config, m.logger)

	m.logger.WithFields(logrus.Fields{
		"vector_db_provider": config.VectorDB.Provider,
		"embedding_provider": config.Embedding.Provider,
	}).Info("Created vector store instance")

	return vectorStore, nil
}

// GetProvider returns a registered provider
func (m *DefaultVectorDBManager) GetProvider(name string) (VectorDBFactory, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	factory, exists := m.vectorDBFactories[name]
	if !exists {
		return nil, fmt.Errorf("unknown vector database provider: %s", name)
	}

	return factory, nil
}

// ListProviders returns all registered providers
func (m *DefaultVectorDBManager) ListProviders() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	providers := make([]string, 0, len(m.vectorDBFactories))
	for name := range m.vectorDBFactories {
		providers = append(providers, name)
	}

	return providers
}

// registerEmbeddingFactory registers an embedding provider factory
func (m *DefaultVectorDBManager) registerEmbeddingFactory(name string, factory EmbeddingProviderFactory) error {
	if factory == nil {
		return fmt.Errorf("factory cannot be nil")
	}

	m.embeddingFactories[name] = factory
	m.logger.WithField("provider", name).Info("Registered embedding provider")

	return nil
}

// createEmbeddingProvider creates an embedding provider
func (m *DefaultVectorDBManager) createEmbeddingProvider(config *EmbeddingConfig) (EmbeddingProvider, error) {
	if config == nil {
		return nil, fmt.Errorf("embedding config cannot be nil")
	}

	factory, exists := m.embeddingFactories[config.Provider]
	if !exists {
		return nil, fmt.Errorf("unknown embedding provider: %s", config.Provider)
	}

	return factory.Create(config)
}

// DefaultVectorDBMetrics implements VectorDBMetrics interface
type DefaultVectorDBMetrics struct {
	operationStats  map[string]*OperationStats
	collectionStats map[string]*CollectionStats
	logger          *logrus.Logger
	mu              sync.RWMutex
}

// NewVectorDBMetrics creates a new vector database metrics instance
func NewVectorDBMetrics(logger *logrus.Logger) VectorDBMetrics {
	return &DefaultVectorDBMetrics{
		operationStats:  make(map[string]*OperationStats),
		collectionStats: make(map[string]*CollectionStats),
		logger:          logger,
	}
}

// RecordOperation records a database operation
func (m *DefaultVectorDBMetrics) RecordOperation(operation string, duration time.Duration, success bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	stats, exists := m.operationStats[operation]
	if !exists {
		stats = &OperationStats{
			MinLatency: duration,
			MaxLatency: duration,
		}
		m.operationStats[operation] = stats
	}

	stats.TotalOperations++
	if success {
		stats.SuccessfulOperations++
	} else {
		stats.FailedOperations++
	}

	// Update latency statistics
	if duration < stats.MinLatency {
		stats.MinLatency = duration
	}
	if duration > stats.MaxLatency {
		stats.MaxLatency = duration
	}

	// Calculate average latency
	totalLatency := time.Duration(stats.AverageLatency.Nanoseconds()*int64(stats.TotalOperations-1)) + duration
	stats.AverageLatency = time.Duration(totalLatency.Nanoseconds() / int64(stats.TotalOperations))
}

// RecordVectorCount records vector count for a collection
func (m *DefaultVectorDBMetrics) RecordVectorCount(collection string, count int64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	stats, exists := m.collectionStats[collection]
	if !exists {
		stats = &CollectionStats{
			Name:         collection,
			LastAccessed: time.Now(),
		}
		m.collectionStats[collection] = stats
	}

	stats.VectorCount = count
	stats.LastAccessed = time.Now()
}

// RecordSearchLatency records search operation latency
func (m *DefaultVectorDBMetrics) RecordSearchLatency(collection string, duration time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	stats, exists := m.collectionStats[collection]
	if !exists {
		stats = &CollectionStats{
			Name:         collection,
			LastAccessed: time.Now(),
		}
		m.collectionStats[collection] = stats
	}

	stats.SearchCount++
	stats.LastAccessed = time.Now()

	// Calculate average search latency
	if stats.SearchCount == 1 {
		stats.AverageSearchLatency = duration
	} else {
		totalLatency := time.Duration(stats.AverageSearchLatency.Nanoseconds()*int64(stats.SearchCount-1)) + duration
		stats.AverageSearchLatency = time.Duration(totalLatency.Nanoseconds() / int64(stats.SearchCount))
	}
}

// GetOperationStats returns operation statistics
func (m *DefaultVectorDBMetrics) GetOperationStats() map[string]*OperationStats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Return a copy to avoid race conditions
	result := make(map[string]*OperationStats)
	for k, v := range m.operationStats {
		statsCopy := *v
		result[k] = &statsCopy
	}

	return result
}

// GetCollectionStats returns collection statistics
func (m *DefaultVectorDBMetrics) GetCollectionStats() map[string]*CollectionStats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Return a copy to avoid race conditions
	result := make(map[string]*CollectionStats)
	for k, v := range m.collectionStats {
		statsCopy := *v
		result[k] = &statsCopy
	}

	return result
}

// VectorDBBuilder provides a fluent interface for building vector database configurations
type VectorDBBuilder struct {
	config *VectorDBConfig
}

// NewVectorDBBuilder creates a new vector database builder
func NewVectorDBBuilder() *VectorDBBuilder {
	return &VectorDBBuilder{
		config: &VectorDBConfig{
			Timeout:    30 * time.Second,
			MaxRetries: 3,
			RetryDelay: 1 * time.Second,
			Metadata:   make(map[string]interface{}),
		},
	}
}

// WithProvider sets the provider
func (b *VectorDBBuilder) WithProvider(provider string) *VectorDBBuilder {
	b.config.Provider = provider
	return b
}

// WithHost sets the host
func (b *VectorDBBuilder) WithHost(host string) *VectorDBBuilder {
	b.config.Host = host
	return b
}

// WithPort sets the port
func (b *VectorDBBuilder) WithPort(port int) *VectorDBBuilder {
	b.config.Port = port
	return b
}

// WithAPIKey sets the API key
func (b *VectorDBBuilder) WithAPIKey(apiKey string) *VectorDBBuilder {
	b.config.APIKey = apiKey
	return b
}

// WithDatabase sets the database name
func (b *VectorDBBuilder) WithDatabase(database string) *VectorDBBuilder {
	b.config.Database = database
	return b
}

// WithTimeout sets the timeout
func (b *VectorDBBuilder) WithTimeout(timeout time.Duration) *VectorDBBuilder {
	b.config.Timeout = timeout
	return b
}

// WithTLS enables/disables TLS
func (b *VectorDBBuilder) WithTLS(tls bool) *VectorDBBuilder {
	b.config.TLS = tls
	return b
}

// WithMetadata adds metadata
func (b *VectorDBBuilder) WithMetadata(key string, value interface{}) *VectorDBBuilder {
	b.config.Metadata[key] = value
	return b
}

// Build builds the configuration
func (b *VectorDBBuilder) Build() *VectorDBConfig {
	return b.config
}

// EmbeddingBuilder provides a fluent interface for building embedding configurations
type EmbeddingBuilder struct {
	config *EmbeddingConfig
}

// NewEmbeddingBuilder creates a new embedding builder
func NewEmbeddingBuilder() *EmbeddingBuilder {
	return &EmbeddingBuilder{
		config: &EmbeddingConfig{
			Timeout:   30 * time.Second,
			BatchSize: 100,
			Metadata:  make(map[string]interface{}),
		},
	}
}

// WithProvider sets the provider
func (b *EmbeddingBuilder) WithProvider(provider string) *EmbeddingBuilder {
	b.config.Provider = provider
	return b
}

// WithModel sets the model
func (b *EmbeddingBuilder) WithModel(model string) *EmbeddingBuilder {
	b.config.Model = model
	return b
}

// WithAPIKey sets the API key
func (b *EmbeddingBuilder) WithAPIKey(apiKey string) *EmbeddingBuilder {
	b.config.APIKey = apiKey
	return b
}

// WithBaseURL sets the base URL
func (b *EmbeddingBuilder) WithBaseURL(baseURL string) *EmbeddingBuilder {
	b.config.BaseURL = baseURL
	return b
}

// WithDimensions sets the dimensions
func (b *EmbeddingBuilder) WithDimensions(dimensions int) *EmbeddingBuilder {
	b.config.Dimensions = dimensions
	return b
}

// WithTimeout sets the timeout
func (b *EmbeddingBuilder) WithTimeout(timeout time.Duration) *EmbeddingBuilder {
	b.config.Timeout = timeout
	return b
}

// WithBatchSize sets the batch size
func (b *EmbeddingBuilder) WithBatchSize(batchSize int) *EmbeddingBuilder {
	b.config.BatchSize = batchSize
	return b
}

// WithMetadata adds metadata
func (b *EmbeddingBuilder) WithMetadata(key string, value interface{}) *EmbeddingBuilder {
	b.config.Metadata[key] = value
	return b
}

// Build builds the configuration
func (b *EmbeddingBuilder) Build() *EmbeddingConfig {
	return b.config
}

// VectorStoreBuilder provides a fluent interface for building vector store configurations
type VectorStoreBuilder struct {
	config *VectorStoreConfig
}

// NewVectorStoreBuilder creates a new vector store builder
func NewVectorStoreBuilder() *VectorStoreBuilder {
	return &VectorStoreBuilder{
		config: &VectorStoreConfig{
			ChunkSize: 1000,
			Overlap:   200,
		},
	}
}

// WithVectorDB sets the vector database configuration
func (b *VectorStoreBuilder) WithVectorDB(vectorDB *VectorDBConfig) *VectorStoreBuilder {
	b.config.VectorDB = vectorDB
	return b
}

// WithEmbedding sets the embedding configuration
func (b *VectorStoreBuilder) WithEmbedding(embedding *EmbeddingConfig) *VectorStoreBuilder {
	b.config.Embedding = embedding
	return b
}

// WithChunkSize sets the chunk size
func (b *VectorStoreBuilder) WithChunkSize(chunkSize int) *VectorStoreBuilder {
	b.config.ChunkSize = chunkSize
	return b
}

// WithOverlap sets the overlap
func (b *VectorStoreBuilder) WithOverlap(overlap int) *VectorStoreBuilder {
	b.config.Overlap = overlap
	return b
}

// Build builds the configuration
func (b *VectorStoreBuilder) Build() *VectorStoreConfig {
	return b.config
}
