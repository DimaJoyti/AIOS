package knowledge

import (
	"time"
)

// Document represents a knowledge document
type Document struct {
	ID              string                 `json:"id"`
	Title           string                 `json:"title"`
	Content         string                 `json:"content"`
	ContentType     string                 `json:"content_type"` // text, pdf, image, audio, video
	Language        string                 `json:"language"`
	Source          string                 `json:"source"`
	URL             string                 `json:"url,omitempty"`
	Author          string                 `json:"author,omitempty"`
	Tags            []string               `json:"tags,omitempty"`
	Categories      []string               `json:"categories,omitempty"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
	Embedding       []float32              `json:"embedding,omitempty"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
	Version         int                    `json:"version"`
	Status          DocumentStatus         `json:"status"`
	KnowledgeBaseID string                 `json:"knowledge_base_id"`
}

// DocumentChunk represents a chunk of a document
type DocumentChunk struct {
	ID          string                 `json:"id"`
	DocumentID  string                 `json:"document_id"`
	Content     string                 `json:"content"`
	ChunkIndex  int                    `json:"chunk_index"`
	StartOffset int                    `json:"start_offset"`
	EndOffset   int                    `json:"end_offset"`
	Embedding   []float32              `json:"embedding,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
}

// ProcessedDocument represents a processed document
type ProcessedDocument struct {
	Document    *Document        `json:"document"`
	Chunks      []*DocumentChunk `json:"chunks"`
	Entities    []*Entity        `json:"entities"`
	Keywords    []string         `json:"keywords"`
	Summary     string           `json:"summary"`
	Language    string           `json:"language"`
	Sentiment   *Sentiment       `json:"sentiment,omitempty"`
	Topics      []*Topic         `json:"topics,omitempty"`
	ProcessedAt time.Time        `json:"processed_at"`
}

// KnowledgeBase represents a knowledge base
type KnowledgeBase struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	OwnerID     string                 `json:"owner_id"`
	Config      *KnowledgeBaseConfig   `json:"config"`
	Stats       *KnowledgeBaseStats    `json:"stats"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	Status      KnowledgeBaseStatus    `json:"status"`
}

// KnowledgeBaseConfig represents configuration for a knowledge base
type KnowledgeBaseConfig struct {
	EmbeddingModel    string                 `json:"embedding_model"`
	ChunkingStrategy  ChunkingStrategy       `json:"chunking_strategy"`
	ChunkSize         int                    `json:"chunk_size"`
	ChunkOverlap      int                    `json:"chunk_overlap"`
	IndexingEnabled   bool                   `json:"indexing_enabled"`
	GraphEnabled      bool                   `json:"graph_enabled"`
	MultiModalEnabled bool                   `json:"multimodal_enabled"`
	VersioningEnabled bool                   `json:"versioning_enabled"`
	SecurityLevel     SecurityLevel          `json:"security_level"`
	RetentionPolicy   *RetentionPolicy       `json:"retention_policy,omitempty"`
	Settings          map[string]interface{} `json:"settings,omitempty"`
}

// Entity represents a knowledge entity
type Entity struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Type        string                 `json:"type"`
	Description string                 `json:"description,omitempty"`
	Properties  map[string]interface{} `json:"properties,omitempty"`
	Aliases     []string               `json:"aliases,omitempty"`
	Embedding   []float32              `json:"embedding,omitempty"`
	Confidence  float32                `json:"confidence"`
	Source      string                 `json:"source"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// Relationship represents a relationship between entities
type Relationship struct {
	ID         string                 `json:"id"`
	FromEntity string                 `json:"from_entity"`
	ToEntity   string                 `json:"to_entity"`
	Type       string                 `json:"type"`
	Properties map[string]interface{} `json:"properties,omitempty"`
	Confidence float32                `json:"confidence"`
	Source     string                 `json:"source"`
	CreatedAt  time.Time              `json:"created_at"`
	UpdatedAt  time.Time              `json:"updated_at"`
}

// SearchQuery represents a search query
type SearchQuery struct {
	Query     string                 `json:"query"`
	Filters   map[string]interface{} `json:"filters,omitempty"`
	Options   *SearchOptions         `json:"options,omitempty"`
	UserID    string                 `json:"user_id,omitempty"`
	SessionID string                 `json:"session_id,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
}

// SearchOptions represents search options
type SearchOptions struct {
	TopK             int                    `json:"top_k"`
	Threshold        float32                `json:"threshold"`
	IncludeMetadata  bool                   `json:"include_metadata"`
	IncludeEmbedding bool                   `json:"include_embedding"`
	SearchType       SearchType             `json:"search_type"`
	RerankingEnabled bool                   `json:"reranking_enabled"`
	Filters          map[string]interface{} `json:"filters,omitempty"`
	KnowledgeBaseIDs []string               `json:"knowledge_base_ids,omitempty"`
}

// SearchResult represents search results
type SearchResult struct {
	Documents      []*Document            `json:"documents"`
	TotalCount     int                    `json:"total_count"`
	Query          string                 `json:"query"`
	ProcessingTime time.Duration          `json:"processing_time"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
	Suggestions    []string               `json:"suggestions,omitempty"`
}

// RAGOptions represents RAG options
type RAGOptions struct {
	RetrievalOptions  *RetrievalOptions  `json:"retrieval_options,omitempty"`
	GenerationOptions *GenerationOptions `json:"generation_options,omitempty"`
	ContextLength     int                `json:"context_length"`
	IncludeSources    bool               `json:"include_sources"`
	StreamResponse    bool               `json:"stream_response"`
}

// RetrievalOptions represents retrieval options
type RetrievalOptions struct {
	TopK             int                    `json:"top_k"`
	Threshold        float32                `json:"threshold"`
	RerankingEnabled bool                   `json:"reranking_enabled"`
	HybridSearch     bool                   `json:"hybrid_search"`
	GraphSearch      bool                   `json:"graph_search"`
	Filters          map[string]interface{} `json:"filters,omitempty"`
	KnowledgeBaseIDs []string               `json:"knowledge_base_ids,omitempty"`
}

// GenerationOptions represents generation options
type GenerationOptions struct {
	Model        string                 `json:"model"`
	Temperature  float32                `json:"temperature"`
	MaxTokens    int                    `json:"max_tokens"`
	SystemPrompt string                 `json:"system_prompt,omitempty"`
	Instructions string                 `json:"instructions,omitempty"`
	Format       string                 `json:"format,omitempty"`
	Parameters   map[string]interface{} `json:"parameters,omitempty"`
}

// RAGResponse represents a RAG response
type RAGResponse struct {
	Response       string                 `json:"response"`
	Sources        []*Document            `json:"sources"`
	Query          string                 `json:"query"`
	Context        string                 `json:"context,omitempty"`
	Citations      []*Citation            `json:"citations,omitempty"`
	Confidence     float32                `json:"confidence"`
	ProcessingTime time.Duration          `json:"processing_time"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
	Timestamp      time.Time              `json:"timestamp"`
}

// RetrievalResult represents retrieval results
type RetrievalResult struct {
	Documents      []*Document            `json:"documents"`
	Context        string                 `json:"context"`
	Query          string                 `json:"query"`
	ProcessingTime time.Duration          `json:"processing_time"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

// GenerationResult represents generation results
type GenerationResult struct {
	Response       string                 `json:"response"`
	TokensUsed     int                    `json:"tokens_used"`
	Model          string                 `json:"model"`
	ProcessingTime time.Duration          `json:"processing_time"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

// Citation represents a citation
type Citation struct {
	DocumentID string  `json:"document_id"`
	ChunkID    string  `json:"chunk_id,omitempty"`
	StartPos   int     `json:"start_pos"`
	EndPos     int     `json:"end_pos"`
	Text       string  `json:"text"`
	Confidence float32 `json:"confidence"`
}

// ProcessedQuery represents a processed query
type ProcessedQuery struct {
	OriginalQuery string         `json:"original_query"`
	CleanedQuery  string         `json:"cleaned_query"`
	Intent        *QueryIntent   `json:"intent,omitempty"`
	Entities      []*QueryEntity `json:"entities,omitempty"`
	Keywords      []string       `json:"keywords"`
	Language      string         `json:"language"`
	Embedding     []float32      `json:"embedding,omitempty"`
	ProcessedAt   time.Time      `json:"processed_at"`
}

// QueryIntent represents query intent
type QueryIntent struct {
	Type       string  `json:"type"`
	Confidence float32 `json:"confidence"`
	Category   string  `json:"category"`
	Action     string  `json:"action,omitempty"`
}

// QueryEntity represents an entity in a query
type QueryEntity struct {
	Text       string  `json:"text"`
	Type       string  `json:"type"`
	StartPos   int     `json:"start_pos"`
	EndPos     int     `json:"end_pos"`
	Confidence float32 `json:"confidence"`
}

// ExpandedQuery represents an expanded query
type ExpandedQuery struct {
	OriginalQuery string   `json:"original_query"`
	ExpandedTerms []string `json:"expanded_terms"`
	Synonyms      []string `json:"synonyms"`
	RelatedTerms  []string `json:"related_terms"`
	Suggestions   []string `json:"suggestions"`
}

// Topic represents a topic
type Topic struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	Keywords    []string  `json:"keywords"`
	Confidence  float32   `json:"confidence"`
	CreatedAt   time.Time `json:"created_at"`
}

// Sentiment represents sentiment analysis
type Sentiment struct {
	Score      float32 `json:"score"`     // -1.0 to 1.0
	Magnitude  float32 `json:"magnitude"` // 0.0 to 1.0
	Label      string  `json:"label"`     // positive, negative, neutral
	Confidence float32 `json:"confidence"`
}

// Enums and constants
type DocumentStatus string

const (
	DocumentStatusDraft      DocumentStatus = "draft"
	DocumentStatusProcessing DocumentStatus = "processing"
	DocumentStatusActive     DocumentStatus = "active"
	DocumentStatusArchived   DocumentStatus = "archived"
	DocumentStatusDeleted    DocumentStatus = "deleted"
)

type KnowledgeBaseStatus string

const (
	KnowledgeBaseStatusActive      KnowledgeBaseStatus = "active"
	KnowledgeBaseStatusIndexing    KnowledgeBaseStatus = "indexing"
	KnowledgeBaseStatusMaintenance KnowledgeBaseStatus = "maintenance"
	KnowledgeBaseStatusArchived    KnowledgeBaseStatus = "archived"
)

type ChunkingStrategy string

const (
	ChunkingStrategyFixed     ChunkingStrategy = "fixed"
	ChunkingStrategySentence  ChunkingStrategy = "sentence"
	ChunkingStrategyParagraph ChunkingStrategy = "paragraph"
	ChunkingStrategySemantic  ChunkingStrategy = "semantic"
	ChunkingStrategyRecursive ChunkingStrategy = "recursive"
)

type SearchType string

const (
	SearchTypeSemantic SearchType = "semantic"
	SearchTypeKeyword  SearchType = "keyword"
	SearchTypeHybrid   SearchType = "hybrid"
	SearchTypeGraph    SearchType = "graph"
)

type SecurityLevel string

const (
	SecurityLevelPublic       SecurityLevel = "public"
	SecurityLevelInternal     SecurityLevel = "internal"
	SecurityLevelRestricted   SecurityLevel = "restricted"
	SecurityLevelConfidential SecurityLevel = "confidential"
)

type ExportFormat string

const (
	ExportFormatJSON     ExportFormat = "json"
	ExportFormatXML      ExportFormat = "xml"
	ExportFormatCSV      ExportFormat = "csv"
	ExportFormatMarkdown ExportFormat = "markdown"
	ExportFormatPDF      ExportFormat = "pdf"
)

type GraphExportFormat string

const (
	GraphExportFormatGraphML GraphExportFormat = "graphml"
	GraphExportFormatGEXF    GraphExportFormat = "gexf"
	GraphExportFormatJSON    GraphExportFormat = "json"
	GraphExportFormatCypher  GraphExportFormat = "cypher"
)

type AlertStatus string

const (
	AlertStatusActive     AlertStatus = "active"
	AlertStatusResolved   AlertStatus = "resolved"
	AlertStatusSuppressed AlertStatus = "suppressed"
)

// Additional types for comprehensive knowledge management

// KnowledgeBaseStats represents statistics for a knowledge base
type KnowledgeBaseStats struct {
	DocumentCount     int       `json:"document_count"`
	ChunkCount        int       `json:"chunk_count"`
	EntityCount       int       `json:"entity_count"`
	RelationshipCount int       `json:"relationship_count"`
	TotalSize         int64     `json:"total_size"`
	LastIndexed       time.Time `json:"last_indexed"`
	IndexHealth       float32   `json:"index_health"`
}

// RetentionPolicy represents data retention policy
type RetentionPolicy struct {
	MaxAge             time.Duration `json:"max_age"`
	MaxDocuments       int           `json:"max_documents"`
	ArchiveAfter       time.Duration `json:"archive_after"`
	DeleteAfter        time.Duration `json:"delete_after"`
	CompressionEnabled bool          `json:"compression_enabled"`
}

// ProcessedContent represents processed multi-modal content
type ProcessedContent struct {
	ContentType string                 `json:"content_type"`
	Text        string                 `json:"text"`
	Metadata    map[string]interface{} `json:"metadata"`
	Confidence  float32                `json:"confidence"`
	ProcessedAt time.Time              `json:"processed_at"`
}

// ValidationResult represents validation results
type ValidationResult struct {
	Valid    bool     `json:"valid"`
	Errors   []string `json:"errors,omitempty"`
	Warnings []string `json:"warnings,omitempty"`
	Score    float32  `json:"score"`
}

// ConsistencyReport represents consistency check results
type ConsistencyReport struct {
	Consistent      bool                `json:"consistent"`
	Issues          []*ConsistencyIssue `json:"issues,omitempty"`
	Score           float32             `json:"score"`
	CheckedAt       time.Time           `json:"checked_at"`
	Recommendations []string            `json:"recommendations,omitempty"`
}

// ConsistencyIssue represents a consistency issue
type ConsistencyIssue struct {
	Type        string `json:"type"`
	Severity    string `json:"severity"`
	Description string `json:"description"`
	DocumentID  string `json:"document_id,omitempty"`
	EntityID    string `json:"entity_id,omitempty"`
}

// DuplicateGroup represents a group of duplicate documents
type DuplicateGroup struct {
	Documents  []string `json:"documents"`
	Similarity float32  `json:"similarity"`
	Confidence float32  `json:"confidence"`
	Reason     string   `json:"reason"`
}

// TimeRange represents a time range
type TimeRange struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// KnowledgeMetrics represents knowledge metrics
type KnowledgeMetrics struct {
	TotalDocuments      int                    `json:"total_documents"`
	TotalKnowledgeBases int                    `json:"total_knowledge_bases"`
	TotalQueries        int64                  `json:"total_queries"`
	AverageResponseTime time.Duration          `json:"average_response_time"`
	IndexHealth         float32                `json:"index_health"`
	StorageUsed         int64                  `json:"storage_used"`
	Metadata            map[string]interface{} `json:"metadata,omitempty"`
	CollectedAt         time.Time              `json:"collected_at"`
}

// SearchAnalytics represents search analytics
type SearchAnalytics struct {
	TotalQueries        int64                  `json:"total_queries"`
	UniqueQueries       int64                  `json:"unique_queries"`
	AverageResponseTime time.Duration          `json:"average_response_time"`
	TopQueries          []string               `json:"top_queries"`
	FailedQueries       int64                  `json:"failed_queries"`
	TimeRange           *TimeRange             `json:"time_range"`
	Metadata            map[string]interface{} `json:"metadata,omitempty"`
}

// CachedQuery represents a cached query
type CachedQuery struct {
	Query      string      `json:"query"`
	Result     interface{} `json:"result"`
	Similarity float32     `json:"similarity"`
	CachedAt   time.Time   `json:"cached_at"`
	ExpiresAt  time.Time   `json:"expires_at"`
}

// CacheStats represents cache statistics
type CacheStats struct {
	HitRate     float32   `json:"hit_rate"`
	TotalHits   int64     `json:"total_hits"`
	TotalMisses int64     `json:"total_misses"`
	Size        int       `json:"size"`
	MaxSize     int       `json:"max_size"`
	LastCleanup time.Time `json:"last_cleanup"`
}

// IndexStats represents index statistics
type IndexStats struct {
	DocumentCount      int       `json:"document_count"`
	ChunkCount         int       `json:"chunk_count"`
	IndexSize          int64     `json:"index_size"`
	LastUpdated        time.Time `json:"last_updated"`
	Health             float32   `json:"health"`
	FragmentationLevel float32   `json:"fragmentation_level"`
}

// QueryContext represents context for query processing
type QueryContext struct {
	UserID      string                 `json:"user_id,omitempty"`
	SessionID   string                 `json:"session_id,omitempty"`
	History     []string               `json:"history,omitempty"`
	Preferences map[string]interface{} `json:"preferences,omitempty"`
	Timestamp   time.Time              `json:"timestamp"`
}

// Permissions represents access permissions
type Permissions struct {
	Read   []string `json:"read"`
	Write  []string `json:"write"`
	Delete []string `json:"delete"`
	Admin  []string `json:"admin"`
}

// SharedKnowledgeBase represents a shared knowledge base
type SharedKnowledgeBase struct {
	KnowledgeBaseID string       `json:"knowledge_base_id"`
	Name            string       `json:"name"`
	SharedBy        string       `json:"shared_by"`
	SharedWith      string       `json:"shared_with"`
	Permissions     *Permissions `json:"permissions"`
	SharedAt        time.Time    `json:"shared_at"`
}

// Annotation represents a document annotation
type Annotation struct {
	ID         string                 `json:"id"`
	DocumentID string                 `json:"document_id"`
	UserID     string                 `json:"user_id"`
	Text       string                 `json:"text"`
	StartPos   int                    `json:"start_pos"`
	EndPos     int                    `json:"end_pos"`
	Type       string                 `json:"type"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt  time.Time              `json:"created_at"`
}

// Comment represents a document comment
type Comment struct {
	ID         string                 `json:"id"`
	DocumentID string                 `json:"document_id"`
	UserID     string                 `json:"user_id"`
	Text       string                 `json:"text"`
	ParentID   string                 `json:"parent_id,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt  time.Time              `json:"created_at"`
	UpdatedAt  time.Time              `json:"updated_at"`
}

// RecommendationOptions represents options for recommendations
type RecommendationOptions struct {
	MaxResults int                    `json:"max_results"`
	Categories []string               `json:"categories,omitempty"`
	Tags       []string               `json:"tags,omitempty"`
	Filters    map[string]interface{} `json:"filters,omitempty"`
}

// QueryRecommendationOptions represents options for query recommendations
type QueryRecommendationOptions struct {
	MaxResults int      `json:"max_results"`
	Categories []string `json:"categories,omitempty"`
	Language   string   `json:"language,omitempty"`
}

// TopicRecommendationOptions represents options for topic recommendations
type TopicRecommendationOptions struct {
	MaxResults int      `json:"max_results"`
	Categories []string `json:"categories,omitempty"`
	Depth      int      `json:"depth,omitempty"`
}

// TrendingTopic represents a trending topic
type TrendingTopic struct {
	Topic     *Topic     `json:"topic"`
	Score     float32    `json:"score"`
	Mentions  int        `json:"mentions"`
	Growth    float32    `json:"growth"`
	TimeRange *TimeRange `json:"time_range"`
}

// QueryEvent represents a query event for monitoring
type QueryEvent struct {
	ID           string                 `json:"id"`
	Query        string                 `json:"query"`
	UserID       string                 `json:"user_id,omitempty"`
	SessionID    string                 `json:"session_id,omitempty"`
	ResultCount  int                    `json:"result_count"`
	ResponseTime time.Duration          `json:"response_time"`
	Success      bool                   `json:"success"`
	ErrorMessage string                 `json:"error_message,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	Timestamp    time.Time              `json:"timestamp"`
}

// RetrievalEvent represents a retrieval event
type RetrievalEvent struct {
	ID             string                 `json:"id"`
	Query          string                 `json:"query"`
	DocumentsFound int                    `json:"documents_found"`
	ResponseTime   time.Duration          `json:"response_time"`
	Method         string                 `json:"method"` // semantic, keyword, hybrid
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
	Timestamp      time.Time              `json:"timestamp"`
}

// GenerationEvent represents a generation event
type GenerationEvent struct {
	ID           string                 `json:"id"`
	Query        string                 `json:"query"`
	Model        string                 `json:"model"`
	TokensUsed   int                    `json:"tokens_used"`
	ResponseTime time.Duration          `json:"response_time"`
	Success      bool                   `json:"success"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	Timestamp    time.Time              `json:"timestamp"`
}

// QueryMetrics represents query metrics
type QueryMetrics struct {
	TotalQueries        int64         `json:"total_queries"`
	SuccessfulQueries   int64         `json:"successful_queries"`
	FailedQueries       int64         `json:"failed_queries"`
	AverageResponseTime time.Duration `json:"average_response_time"`
	TopQueries          []string      `json:"top_queries"`
	TimeRange           *TimeRange    `json:"time_range"`
}

// PerformanceMetrics represents performance metrics
type PerformanceMetrics struct {
	AverageQueryTime      time.Duration `json:"average_query_time"`
	AverageRetrievalTime  time.Duration `json:"average_retrieval_time"`
	AverageGenerationTime time.Duration `json:"average_generation_time"`
	CacheHitRate          float32       `json:"cache_hit_rate"`
	IndexHealth           float32       `json:"index_health"`
	TimeRange             *TimeRange    `json:"time_range"`
}

// UsageMetrics represents usage metrics
type UsageMetrics struct {
	ActiveUsers        int        `json:"active_users"`
	TotalSessions      int        `json:"total_sessions"`
	DocumentsAccessed  int        `json:"documents_accessed"`
	KnowledgeBasesUsed int        `json:"knowledge_bases_used"`
	TimeRange          *TimeRange `json:"time_range"`
}

// Alert represents a system alert
type Alert struct {
	ID         string                 `json:"id"`
	Type       string                 `json:"type"`
	Severity   string                 `json:"severity"`
	Message    string                 `json:"message"`
	Source     string                 `json:"source"`
	Status     AlertStatus            `json:"status"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt  time.Time              `json:"created_at"`
	ResolvedAt *time.Time             `json:"resolved_at,omitempty"`
}

// Version represents a knowledge base version
type Version struct {
	ID          string    `json:"id"`
	Number      int       `json:"number"`
	Description string    `json:"description"`
	CreatedBy   string    `json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
	Size        int64     `json:"size"`
	Changes     int       `json:"changes"`
}

// VersionComparison represents a comparison between versions
type VersionComparison struct {
	FromVersion string                 `json:"from_version"`
	ToVersion   string                 `json:"to_version"`
	Changes     []*VersionChange       `json:"changes"`
	Summary     map[string]interface{} `json:"summary"`
	CreatedAt   time.Time              `json:"created_at"`
}

// VersionChange represents a change between versions
type VersionChange struct {
	Type       string                 `json:"type"` // added, modified, deleted
	DocumentID string                 `json:"document_id"`
	Field      string                 `json:"field,omitempty"`
	OldValue   interface{}            `json:"old_value,omitempty"`
	NewValue   interface{}            `json:"new_value,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// DocumentVersion represents a document version
type DocumentVersion struct {
	ID         string    `json:"id"`
	DocumentID string    `json:"document_id"`
	Version    int       `json:"version"`
	Content    string    `json:"content"`
	Changes    string    `json:"changes"`
	CreatedBy  string    `json:"created_by"`
	CreatedAt  time.Time `json:"created_at"`
}

// EncryptedDocument represents an encrypted document
type EncryptedDocument struct {
	ID            string                 `json:"id"`
	EncryptedData []byte                 `json:"encrypted_data"`
	KeyID         string                 `json:"key_id"`
	Algorithm     string                 `json:"algorithm"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt     time.Time              `json:"created_at"`
}

// AuditEntry represents an audit log entry
type AuditEntry struct {
	ID        string                 `json:"id"`
	UserID    string                 `json:"user_id"`
	Action    string                 `json:"action"`
	Resource  string                 `json:"resource"`
	Success   bool                   `json:"success"`
	Details   map[string]interface{} `json:"details,omitempty"`
	IPAddress string                 `json:"ip_address,omitempty"`
	UserAgent string                 `json:"user_agent,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
}

// AuditFilters represents filters for audit queries
type AuditFilters struct {
	UserID    string     `json:"user_id,omitempty"`
	Action    string     `json:"action,omitempty"`
	Resource  string     `json:"resource,omitempty"`
	Success   *bool      `json:"success,omitempty"`
	StartTime *time.Time `json:"start_time,omitempty"`
	EndTime   *time.Time `json:"end_time,omitempty"`
	Limit     int        `json:"limit,omitempty"`
}

// Workflow represents an automated workflow
type Workflow struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Trigger     *WorkflowTrigger       `json:"trigger"`
	Steps       []*WorkflowStep        `json:"steps"`
	Enabled     bool                   `json:"enabled"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// WorkflowTrigger represents a workflow trigger
type WorkflowTrigger struct {
	Type       string                 `json:"type"` // schedule, event, manual
	Schedule   string                 `json:"schedule,omitempty"`
	Event      string                 `json:"event,omitempty"`
	Conditions map[string]interface{} `json:"conditions,omitempty"`
}

// WorkflowStep represents a workflow step
type WorkflowStep struct {
	ID         string                 `json:"id"`
	Name       string                 `json:"name"`
	Type       string                 `json:"type"`
	Action     string                 `json:"action"`
	Parameters map[string]interface{} `json:"parameters,omitempty"`
	Conditions map[string]interface{} `json:"conditions,omitempty"`
	OnSuccess  string                 `json:"on_success,omitempty"`
	OnFailure  string                 `json:"on_failure,omitempty"`
}

// WorkflowInput represents workflow input
type WorkflowInput struct {
	Parameters map[string]interface{} `json:"parameters,omitempty"`
	Context    map[string]interface{} `json:"context,omitempty"`
}

// WorkflowResult represents workflow execution result
type WorkflowResult struct {
	ID          string                 `json:"id"`
	WorkflowID  string                 `json:"workflow_id"`
	Status      string                 `json:"status"` // running, completed, failed
	Steps       []*WorkflowStepResult  `json:"steps"`
	Output      map[string]interface{} `json:"output,omitempty"`
	Error       string                 `json:"error,omitempty"`
	StartedAt   time.Time              `json:"started_at"`
	CompletedAt *time.Time             `json:"completed_at,omitempty"`
}

// WorkflowStepResult represents a workflow step result
type WorkflowStepResult struct {
	StepID      string                 `json:"step_id"`
	Status      string                 `json:"status"`
	Output      map[string]interface{} `json:"output,omitempty"`
	Error       string                 `json:"error,omitempty"`
	StartedAt   time.Time              `json:"started_at"`
	CompletedAt *time.Time             `json:"completed_at,omitempty"`
}

// Schedule represents a workflow schedule
type Schedule struct {
	Type       string     `json:"type"` // cron, interval
	Expression string     `json:"expression"`
	StartTime  time.Time  `json:"start_time"`
	EndTime    *time.Time `json:"end_time,omitempty"`
	Timezone   string     `json:"timezone,omitempty"`
}

// BackupOptions represents backup options
type BackupOptions struct {
	IncludeDocuments bool                   `json:"include_documents"`
	IncludeIndex     bool                   `json:"include_index"`
	IncludeGraph     bool                   `json:"include_graph"`
	Compression      bool                   `json:"compression"`
	Encryption       bool                   `json:"encryption"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
}

// BackupResult represents backup result
type BackupResult struct {
	BackupID      string     `json:"backup_id"`
	Size          int64      `json:"size"`
	DocumentCount int        `json:"document_count"`
	Status        string     `json:"status"`
	CreatedAt     time.Time  `json:"created_at"`
	CompletedAt   *time.Time `json:"completed_at,omitempty"`
	Error         string     `json:"error,omitempty"`
}
