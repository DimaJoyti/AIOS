package memory

import (
	"context"
	"time"

	"github.com/aios/aios/pkg/langchain/llm"
)

// Memory defines the interface for memory systems
type Memory interface {
	// LoadMemoryVariables loads memory variables for the given input
	LoadMemoryVariables(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error)

	// SaveContext saves the context from input and output
	SaveContext(ctx context.Context, input map[string]interface{}, output map[string]interface{}) error

	// Clear clears all memory
	Clear(ctx context.Context) error

	// GetMemoryKeys returns the keys that this memory system provides
	GetMemoryKeys() []string

	// GetMemoryType returns the type of this memory system
	GetMemoryType() string
}

// ConversationMemory manages conversation history
type ConversationMemory interface {
	Memory

	// AddMessage adds a message to the conversation
	AddMessage(ctx context.Context, message llm.Message) error

	// GetMessages returns all messages in the conversation
	GetMessages(ctx context.Context) ([]llm.Message, error)

	// GetRecentMessages returns the most recent N messages
	GetRecentMessages(ctx context.Context, count int) ([]llm.Message, error)

	// GetMessagesByTimeRange returns messages within a time range
	GetMessagesByTimeRange(ctx context.Context, start, end time.Time) ([]llm.Message, error)

	// SetMaxTokens sets the maximum number of tokens to keep in memory
	SetMaxTokens(maxTokens int)

	// SetMaxMessages sets the maximum number of messages to keep in memory
	SetMaxMessages(maxMessages int)
}

// VectorMemory manages vector-based memory for semantic search
type VectorMemory interface {
	Memory

	// AddDocument adds a document to vector memory
	AddDocument(ctx context.Context, doc *Document) error

	// AddDocuments adds multiple documents to vector memory
	AddDocuments(ctx context.Context, docs []*Document) error

	// SearchSimilar searches for similar documents
	SearchSimilar(ctx context.Context, query string, limit int) ([]*Document, error)

	// SearchSimilarWithScores searches for similar documents with similarity scores
	SearchSimilarWithScores(ctx context.Context, query string, limit int) ([]*DocumentWithScore, error)

	// DeleteDocument deletes a document by ID
	DeleteDocument(ctx context.Context, id string) error

	// GetDocument retrieves a document by ID
	GetDocument(ctx context.Context, id string) (*Document, error)

	// GetDocumentCount returns the total number of documents
	GetDocumentCount(ctx context.Context) (int, error)
}

// EntityMemory manages entity-based memory
type EntityMemory interface {
	Memory

	// AddEntity adds or updates an entity
	AddEntity(ctx context.Context, entity *Entity) error

	// GetEntity retrieves an entity by name
	GetEntity(ctx context.Context, name string) (*Entity, error)

	// GetEntities retrieves all entities
	GetEntities(ctx context.Context) ([]*Entity, error)

	// UpdateEntity updates an entity's information
	UpdateEntity(ctx context.Context, name string, info string) error

	// DeleteEntity deletes an entity
	DeleteEntity(ctx context.Context, name string) error

	// ExtractEntities extracts entities from text
	ExtractEntities(ctx context.Context, text string) ([]*Entity, error)
}

// SummaryMemory manages summarized conversation history
type SummaryMemory interface {
	Memory

	// AddToSummary adds new information to the summary
	AddToSummary(ctx context.Context, text string) error

	// GetSummary returns the current summary
	GetSummary(ctx context.Context) (string, error)

	// UpdateSummary updates the summary with new information
	UpdateSummary(ctx context.Context, newInfo string) error

	// SetSummaryLLM sets the LLM used for summarization
	SetSummaryLLM(llm llm.LLM)

	// SetMaxSummaryTokens sets the maximum tokens for the summary
	SetMaxSummaryTokens(maxTokens int)
}

// Document represents a document in vector memory
type Document struct {
	ID       string                 `json:"id"`
	Content  string                 `json:"content"`
	Metadata map[string]interface{} `json:"metadata"`
	Vector   []float64              `json:"vector,omitempty"`
}

// DocumentWithScore represents a document with similarity score
type DocumentWithScore struct {
	Document *Document `json:"document"`
	Score    float64   `json:"score"`
}

// Entity represents an entity in entity memory
type Entity struct {
	Name        string                 `json:"name"`
	Type        string                 `json:"type"`
	Description string                 `json:"description"`
	Properties  map[string]interface{} `json:"properties"`
	LastUpdated time.Time              `json:"last_updated"`
}

// MemoryConfig represents configuration for memory systems
type MemoryConfig struct {
	Type                string                 `json:"type"`
	MaxTokens           int                    `json:"max_tokens,omitempty"`
	MaxMessages         int                    `json:"max_messages,omitempty"`
	MaxDocuments        int                    `json:"max_documents,omitempty"`
	MaxEntities         int                    `json:"max_entities,omitempty"`
	RetentionDays       int                    `json:"retention_days,omitempty"`
	VectorDimension     int                    `json:"vector_dimension,omitempty"`
	SimilarityThreshold float64                `json:"similarity_threshold,omitempty"`
	Metadata            map[string]interface{} `json:"metadata,omitempty"`
}

// MemoryManager manages multiple memory systems
type MemoryManager interface {
	// AddMemory adds a memory system
	AddMemory(name string, memory Memory) error

	// GetMemory retrieves a memory system by name
	GetMemory(name string) (Memory, error)

	// RemoveMemory removes a memory system
	RemoveMemory(name string) error

	// LoadAllMemoryVariables loads variables from all memory systems
	LoadAllMemoryVariables(ctx context.Context, input map[string]interface{}) (map[string]interface{}, error)

	// SaveAllContext saves context to all applicable memory systems
	SaveAllContext(ctx context.Context, input map[string]interface{}, output map[string]interface{}) error

	// ClearAll clears all memory systems
	ClearAll(ctx context.Context) error

	// GetMemoryTypes returns all memory types
	GetMemoryTypes() []string
}

// MemoryStore defines the interface for memory storage backends
type MemoryStore interface {
	// Store stores data with a key
	Store(ctx context.Context, key string, data interface{}) error

	// Retrieve retrieves data by key
	Retrieve(ctx context.Context, key string) (interface{}, error)

	// Delete deletes data by key
	Delete(ctx context.Context, key string) error

	// List lists all keys with optional prefix
	List(ctx context.Context, prefix string) ([]string, error)

	// Exists checks if a key exists
	Exists(ctx context.Context, key string) (bool, error)

	// Close closes the store
	Close() error
}

// VectorStore defines the interface for vector storage backends
type VectorStore interface {
	// AddVectors adds vectors with IDs and metadata
	AddVectors(ctx context.Context, vectors []Vector) error

	// SearchVectors searches for similar vectors
	SearchVectors(ctx context.Context, query []float64, limit int) ([]VectorSearchResult, error)

	// DeleteVector deletes a vector by ID
	DeleteVector(ctx context.Context, id string) error

	// GetVector retrieves a vector by ID
	GetVector(ctx context.Context, id string) (*Vector, error)

	// GetVectorCount returns the total number of vectors
	GetVectorCount(ctx context.Context) (int, error)

	// Close closes the vector store
	Close() error
}

// Vector represents a vector with metadata
type Vector struct {
	ID       string                 `json:"id"`
	Vector   []float64              `json:"vector"`
	Metadata map[string]interface{} `json:"metadata"`
}

// VectorSearchResult represents a vector search result
type VectorSearchResult struct {
	Vector *Vector `json:"vector"`
	Score  float64 `json:"score"`
}

// MemoryOptimizer optimizes memory usage and performance
type MemoryOptimizer interface {
	// OptimizeMemory optimizes a memory system
	OptimizeMemory(ctx context.Context, memory Memory) error

	// CompactMemory compacts memory by removing old or irrelevant data
	CompactMemory(ctx context.Context, memory Memory) error

	// AnalyzeMemory analyzes memory usage and provides recommendations
	AnalyzeMemory(ctx context.Context, memory Memory) (*MemoryAnalysis, error)

	// SetRetentionPolicy sets data retention policies
	SetRetentionPolicy(policy *RetentionPolicy) error
}

// MemoryAnalysis represents memory analysis results
type MemoryAnalysis struct {
	MemoryType      string                 `json:"memory_type"`
	TotalSize       int64                  `json:"total_size"`
	ItemCount       int                    `json:"item_count"`
	OldestItem      time.Time              `json:"oldest_item"`
	NewestItem      time.Time              `json:"newest_item"`
	Recommendations []string               `json:"recommendations"`
	Metrics         map[string]interface{} `json:"metrics"`
}

// RetentionPolicy defines data retention policies
type RetentionPolicy struct {
	MaxAge          time.Duration `json:"max_age"`
	MaxItems        int           `json:"max_items"`
	MaxSize         int64         `json:"max_size"`
	CompactionRules []string      `json:"compaction_rules"`
}

// MemoryMetrics collects metrics about memory usage
type MemoryMetrics interface {
	// RecordMemoryOperation records a memory operation
	RecordMemoryOperation(memoryType string, operation string, duration time.Duration, success bool)

	// RecordMemorySize records memory size
	RecordMemorySize(memoryType string, size int64)

	// RecordMemoryAccess records memory access patterns
	RecordMemoryAccess(memoryType string, accessType string, count int)

	// GetMemoryStats returns memory statistics
	GetMemoryStats(memoryType string) map[string]interface{}

	// GetOverallStats returns overall memory statistics
	GetOverallStats() map[string]interface{}
}

// MemoryBuilder provides a fluent interface for building memory systems
type MemoryBuilder interface {
	// WithType sets the memory type
	WithType(memoryType string) MemoryBuilder

	// WithMaxTokens sets the maximum tokens
	WithMaxTokens(maxTokens int) MemoryBuilder

	// WithMaxMessages sets the maximum messages
	WithMaxMessages(maxMessages int) MemoryBuilder

	// WithMaxDocuments sets the maximum documents
	WithMaxDocuments(maxDocuments int) MemoryBuilder

	// WithRetention sets retention policy
	WithRetention(days int) MemoryBuilder

	// WithVectorDimension sets vector dimension
	WithVectorDimension(dimension int) MemoryBuilder

	// WithStore sets the storage backend
	WithStore(store MemoryStore) MemoryBuilder

	// WithVectorStore sets the vector storage backend
	WithVectorStore(store VectorStore) MemoryBuilder

	// WithLLM sets the LLM for operations like summarization
	WithLLM(llm llm.LLM) MemoryBuilder

	// WithMetadata adds metadata
	WithMetadata(key string, value interface{}) MemoryBuilder

	// Build builds the memory system
	Build() (Memory, error)

	// BuildConversationMemory builds conversation memory
	BuildConversationMemory() (ConversationMemory, error)

	// BuildVectorMemory builds vector memory
	BuildVectorMemory() (VectorMemory, error)

	// BuildEntityMemory builds entity memory
	BuildEntityMemory() (EntityMemory, error)

	// BuildSummaryMemory builds summary memory
	BuildSummaryMemory() (SummaryMemory, error)
}
