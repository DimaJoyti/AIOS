package vectordb

import (
	"context"
	"time"
)

// VectorDB defines the interface for vector database operations
type VectorDB interface {
	// Connect establishes connection to the vector database
	Connect(ctx context.Context) error

	// Disconnect closes the connection to the vector database
	Disconnect(ctx context.Context) error

	// IsConnected returns whether the database is connected
	IsConnected() bool

	// CreateCollection creates a new collection with the specified configuration
	CreateCollection(ctx context.Context, config *CollectionConfig) error

	// DeleteCollection deletes a collection
	DeleteCollection(ctx context.Context, name string) error

	// ListCollections returns all available collections
	ListCollections(ctx context.Context) ([]string, error)

	// CollectionExists checks if a collection exists
	CollectionExists(ctx context.Context, name string) (bool, error)

	// Insert inserts vectors into a collection
	Insert(ctx context.Context, collection string, vectors []*Vector) error

	// Update updates existing vectors in a collection
	Update(ctx context.Context, collection string, vectors []*Vector) error

	// Delete deletes vectors from a collection by IDs
	Delete(ctx context.Context, collection string, ids []string) error

	// Search performs similarity search
	Search(ctx context.Context, request *SearchRequest) (*SearchResult, error)

	// BatchSearch performs multiple searches in a single request
	BatchSearch(ctx context.Context, requests []*SearchRequest) ([]*SearchResult, error)

	// GetVector retrieves a vector by ID
	GetVector(ctx context.Context, collection string, id string) (*Vector, error)

	// GetVectors retrieves multiple vectors by IDs
	GetVectors(ctx context.Context, collection string, ids []string) ([]*Vector, error)

	// Count returns the number of vectors in a collection
	Count(ctx context.Context, collection string) (int64, error)

	// CreateIndex creates an index for faster search
	CreateIndex(ctx context.Context, collection string, config *IndexConfig) error

	// DeleteIndex deletes an index
	DeleteIndex(ctx context.Context, collection string, indexName string) error

	// GetCollectionInfo returns information about a collection
	GetCollectionInfo(ctx context.Context, collection string) (*CollectionInfo, error)

	// Health returns the health status of the database
	Health(ctx context.Context) (*HealthStatus, error)
}

// EmbeddingProvider defines the interface for generating embeddings
type EmbeddingProvider interface {
	// GenerateEmbedding generates an embedding for the given text
	GenerateEmbedding(ctx context.Context, text string) ([]float32, error)

	// GenerateEmbeddings generates embeddings for multiple texts
	GenerateEmbeddings(ctx context.Context, texts []string) ([][]float32, error)

	// GetDimensions returns the dimensionality of embeddings
	GetDimensions() int

	// GetModel returns the model name used for embeddings
	GetModel() string

	// GetMaxTokens returns the maximum number of tokens supported
	GetMaxTokens() int
}

// Vector represents a vector with metadata
type Vector struct {
	ID       string                 `json:"id"`
	Values   []float32              `json:"values"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// SearchRequest represents a search request
type SearchRequest struct {
	Collection string                 `json:"collection"`
	Vector     []float32              `json:"vector,omitempty"`
	Query      string                 `json:"query,omitempty"`
	TopK       int                    `json:"top_k"`
	Filter     map[string]interface{} `json:"filter,omitempty"`
	Include    []string               `json:"include,omitempty"` // metadata, values, etc.
}

// SearchResult represents search results
type SearchResult struct {
	Matches []Match `json:"matches"`
	Total   int64   `json:"total"`
}

// Match represents a single search result match
type Match struct {
	ID       string                 `json:"id"`
	Score    float32                `json:"score"`
	Values   []float32              `json:"values,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// CollectionConfig represents configuration for creating a collection
type CollectionConfig struct {
	Name        string                 `json:"name"`
	Dimension   int                    `json:"dimension"`
	Metric      string                 `json:"metric"` // cosine, euclidean, dot_product
	Description string                 `json:"description,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// IndexConfig represents configuration for creating an index
type IndexConfig struct {
	Name       string                 `json:"name"`
	Type       string                 `json:"type"` // hnsw, ivf, etc.
	Parameters map[string]interface{} `json:"parameters,omitempty"`
}

// CollectionInfo represents information about a collection
type CollectionInfo struct {
	Name        string                 `json:"name"`
	Dimension   int                    `json:"dimension"`
	Metric      string                 `json:"metric"`
	VectorCount int64                  `json:"vector_count"`
	IndexCount  int                    `json:"index_count"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// HealthStatus represents the health status of the database
type HealthStatus struct {
	Status      string                 `json:"status"` // healthy, degraded, unhealthy
	Version     string                 `json:"version"`
	Uptime      time.Duration          `json:"uptime"`
	Collections int                    `json:"collections"`
	TotalVectors int64                 `json:"total_vectors"`
	Memory      *MemoryInfo            `json:"memory,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// MemoryInfo represents memory usage information
type MemoryInfo struct {
	Used      int64 `json:"used"`
	Available int64 `json:"available"`
	Total     int64 `json:"total"`
}

// VectorStore provides high-level operations for vector storage and retrieval
type VectorStore interface {
	// AddDocuments adds documents with automatic embedding generation
	AddDocuments(ctx context.Context, collection string, documents []*Document) error

	// AddTexts adds texts with automatic embedding generation
	AddTexts(ctx context.Context, collection string, texts []string, metadatas []map[string]interface{}) error

	// SimilaritySearch performs similarity search with text query
	SimilaritySearch(ctx context.Context, collection string, query string, topK int, filter map[string]interface{}) ([]*Document, error)

	// SimilaritySearchWithScore performs similarity search with scores
	SimilaritySearchWithScore(ctx context.Context, collection string, query string, topK int, filter map[string]interface{}) ([]*DocumentWithScore, error)

	// MaxMarginalRelevanceSearch performs MMR search for diverse results
	MaxMarginalRelevanceSearch(ctx context.Context, collection string, query string, topK int, fetchK int, lambda float32, filter map[string]interface{}) ([]*Document, error)

	// GetDocuments retrieves documents by IDs
	GetDocuments(ctx context.Context, collection string, ids []string) ([]*Document, error)

	// DeleteDocuments deletes documents by IDs
	DeleteDocuments(ctx context.Context, collection string, ids []string) error

	// UpdateDocuments updates existing documents
	UpdateDocuments(ctx context.Context, collection string, documents []*Document) error
}

// Document represents a document with content and metadata
type Document struct {
	ID       string                 `json:"id"`
	Content  string                 `json:"content"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// DocumentWithScore represents a document with similarity score
type DocumentWithScore struct {
	Document *Document `json:"document"`
	Score    float32   `json:"score"`
}

// VectorDBConfig represents configuration for vector database connection
type VectorDBConfig struct {
	Provider    string                 `json:"provider"` // qdrant, weaviate, pinecone, etc.
	Host        string                 `json:"host"`
	Port        int                    `json:"port"`
	APIKey      string                 `json:"api_key,omitempty"`
	Database    string                 `json:"database,omitempty"`
	Timeout     time.Duration          `json:"timeout"`
	MaxRetries  int                    `json:"max_retries"`
	RetryDelay  time.Duration          `json:"retry_delay"`
	TLS         bool                   `json:"tls"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// EmbeddingConfig represents configuration for embedding provider
type EmbeddingConfig struct {
	Provider   string                 `json:"provider"` // openai, ollama, huggingface, etc.
	Model      string                 `json:"model"`
	APIKey     string                 `json:"api_key,omitempty"`
	BaseURL    string                 `json:"base_url,omitempty"`
	Dimensions int                    `json:"dimensions,omitempty"`
	MaxTokens  int                    `json:"max_tokens,omitempty"`
	BatchSize  int                    `json:"batch_size,omitempty"`
	Timeout    time.Duration          `json:"timeout"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// VectorStoreConfig represents configuration for vector store
type VectorStoreConfig struct {
	VectorDB  *VectorDBConfig  `json:"vector_db"`
	Embedding *EmbeddingConfig `json:"embedding"`
	ChunkSize int              `json:"chunk_size,omitempty"`
	Overlap   int              `json:"overlap,omitempty"`
}

// VectorDBManager manages multiple vector database connections
type VectorDBManager interface {
	// RegisterProvider registers a vector database provider
	RegisterProvider(name string, factory VectorDBFactory) error

	// CreateVectorDB creates a vector database instance
	CreateVectorDB(config *VectorDBConfig) (VectorDB, error)

	// CreateVectorStore creates a vector store instance
	CreateVectorStore(config *VectorStoreConfig) (VectorStore, error)

	// GetProvider returns a registered provider
	GetProvider(name string) (VectorDBFactory, error)

	// ListProviders returns all registered providers
	ListProviders() []string
}

// VectorDBFactory creates vector database instances
type VectorDBFactory interface {
	// Create creates a new vector database instance
	Create(config *VectorDBConfig) (VectorDB, error)

	// GetProviderName returns the provider name
	GetProviderName() string

	// ValidateConfig validates the configuration
	ValidateConfig(config *VectorDBConfig) error
}

// EmbeddingProviderFactory creates embedding provider instances
type EmbeddingProviderFactory interface {
	// Create creates a new embedding provider instance
	Create(config *EmbeddingConfig) (EmbeddingProvider, error)

	// GetProviderName returns the provider name
	GetProviderName() string

	// ValidateConfig validates the configuration
	ValidateConfig(config *EmbeddingConfig) error
}

// VectorDBMetrics provides metrics for vector database operations
type VectorDBMetrics interface {
	// RecordOperation records a database operation
	RecordOperation(operation string, duration time.Duration, success bool)

	// RecordVectorCount records vector count for a collection
	RecordVectorCount(collection string, count int64)

	// RecordSearchLatency records search operation latency
	RecordSearchLatency(collection string, duration time.Duration)

	// GetOperationStats returns operation statistics
	GetOperationStats() map[string]*OperationStats

	// GetCollectionStats returns collection statistics
	GetCollectionStats() map[string]*CollectionStats
}

// OperationStats represents statistics for database operations
type OperationStats struct {
	TotalOperations      int64         `json:"total_operations"`
	SuccessfulOperations int64         `json:"successful_operations"`
	FailedOperations     int64         `json:"failed_operations"`
	AverageLatency       time.Duration `json:"average_latency"`
	MinLatency           time.Duration `json:"min_latency"`
	MaxLatency           time.Duration `json:"max_latency"`
}

// CollectionStats represents statistics for a collection
type CollectionStats struct {
	Name         string        `json:"name"`
	VectorCount  int64         `json:"vector_count"`
	SearchCount  int64         `json:"search_count"`
	InsertCount  int64         `json:"insert_count"`
	UpdateCount  int64         `json:"update_count"`
	DeleteCount  int64         `json:"delete_count"`
	LastAccessed time.Time     `json:"last_accessed"`
	AverageSearchLatency time.Duration `json:"average_search_latency"`
}
