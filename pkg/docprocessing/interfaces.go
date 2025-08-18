package docprocessing

import (
	"context"
	"fmt"
	"io"
	"time"
)

// Document represents a document with content and metadata
type Document struct {
	ID          string                 `json:"id"`
	Source      string                 `json:"source"`
	Title       string                 `json:"title"`
	Content     string                 `json:"content"`
	ContentType string                 `json:"content_type"`
	Language    string                 `json:"language,omitempty"`
	Metadata    map[string]interface{} `json:"metadata"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	Size        int64                  `json:"size"`
	Checksum    string                 `json:"checksum,omitempty"`
}

// DocumentChunk represents a chunk of a document
type DocumentChunk struct {
	ID         string                 `json:"id"`
	DocumentID string                 `json:"document_id"`
	Content    string                 `json:"content"`
	ChunkIndex int                    `json:"chunk_index"`
	StartPos   int                    `json:"start_pos"`
	EndPos     int                    `json:"end_pos"`
	Metadata   map[string]interface{} `json:"metadata"`
}

// ProcessingResult represents the result of document processing
type ProcessingResult struct {
	Document *Document              `json:"document"`
	Chunks   []*DocumentChunk       `json:"chunks"`
	Success  bool                   `json:"success"`
	Error    error                  `json:"error,omitempty"`
	Duration time.Duration          `json:"duration"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// DocumentSource defines the interface for document sources
type DocumentSource interface {
	// GetDocuments retrieves documents from the source
	GetDocuments(ctx context.Context) (<-chan *Document, error)

	// GetDocument retrieves a specific document by ID
	GetDocument(ctx context.Context, id string) (*Document, error)

	// GetSourceType returns the type of the source
	GetSourceType() string

	// Configure configures the source with options
	Configure(options map[string]interface{}) error

	// Close closes the source and releases resources
	Close() error
}

// ContentExtractor defines the interface for content extraction
type ContentExtractor interface {
	// Extract extracts content from a document
	Extract(ctx context.Context, doc *Document) (*Document, error)

	// CanExtract checks if the extractor can handle the document type
	CanExtract(contentType string) bool

	// GetSupportedTypes returns supported content types
	GetSupportedTypes() []string
}

// TextProcessor defines the interface for text processing
type TextProcessor interface {
	// Process processes text content
	Process(ctx context.Context, text string, metadata map[string]interface{}) (string, map[string]interface{}, error)

	// GetProcessorType returns the processor type
	GetProcessorType() string

	// Configure configures the processor with options
	Configure(options map[string]interface{}) error
}

// MetadataExtractor defines the interface for metadata extraction
type MetadataExtractor interface {
	// ExtractMetadata extracts metadata from a document
	ExtractMetadata(ctx context.Context, doc *Document) (map[string]interface{}, error)

	// GetExtractorType returns the extractor type
	GetExtractorType() string

	// Configure configures the extractor with options
	Configure(options map[string]interface{}) error
}

// DocumentChunker defines the interface for document chunking
type DocumentChunker interface {
	// ChunkDocument splits a document into chunks
	ChunkDocument(ctx context.Context, doc *Document) ([]*DocumentChunk, error)

	// GetChunkerType returns the chunker type
	GetChunkerType() string

	// Configure configures the chunker with options
	Configure(options map[string]interface{}) error
}

// ProcessingPipeline defines the interface for document processing pipelines
type ProcessingPipeline interface {
	// ProcessDocument processes a single document through the pipeline
	ProcessDocument(ctx context.Context, doc *Document) (*ProcessingResult, error)

	// ProcessDocuments processes multiple documents
	ProcessDocuments(ctx context.Context, docs []*Document) ([]*ProcessingResult, error)

	// ProcessStream processes documents from a stream
	ProcessStream(ctx context.Context, docStream <-chan *Document) (<-chan *ProcessingResult, error)

	// AddStage adds a processing stage to the pipeline
	AddStage(stage ProcessingStage) error

	// RemoveStage removes a processing stage from the pipeline
	RemoveStage(stageName string) error

	// GetStages returns all processing stages
	GetStages() []ProcessingStage

	// Configure configures the pipeline with options
	Configure(options map[string]interface{}) error
}

// ProcessingStage defines the interface for pipeline stages
type ProcessingStage interface {
	// Process processes a document in this stage
	Process(ctx context.Context, doc *Document) (*Document, error)

	// GetStageName returns the stage name
	GetStageName() string

	// GetStageType returns the stage type
	GetStageType() string

	// Configure configures the stage with options
	Configure(options map[string]interface{}) error

	// Validate validates the stage configuration
	Validate() error
}

// DocumentProcessor defines the main interface for document processing
type DocumentProcessor interface {
	// ProcessFromSource processes documents from a source
	ProcessFromSource(ctx context.Context, source DocumentSource) (<-chan *ProcessingResult, error)

	// ProcessDocument processes a single document
	ProcessDocument(ctx context.Context, doc *Document) (*ProcessingResult, error)

	// ProcessFile processes a file
	ProcessFile(ctx context.Context, filePath string) (*ProcessingResult, error)

	// ProcessURL processes a document from URL
	ProcessURL(ctx context.Context, url string) (*ProcessingResult, error)

	// ProcessReader processes content from a reader
	ProcessReader(ctx context.Context, reader io.Reader, contentType string, metadata map[string]interface{}) (*ProcessingResult, error)

	// GetPipeline returns the processing pipeline
	GetPipeline() ProcessingPipeline

	// SetPipeline sets the processing pipeline
	SetPipeline(pipeline ProcessingPipeline) error

	// GetMetrics returns processing metrics
	GetMetrics() *ProcessingMetrics
}

// ProcessingMetrics represents processing metrics
type ProcessingMetrics struct {
	TotalDocuments     int64         `json:"total_documents"`
	ProcessedDocuments int64         `json:"processed_documents"`
	FailedDocuments    int64         `json:"failed_documents"`
	TotalChunks        int64         `json:"total_chunks"`
	AverageProcessTime time.Duration `json:"average_process_time"`
	TotalProcessTime   time.Duration `json:"total_process_time"`
	ErrorRate          float64       `json:"error_rate"`
	ThroughputPerSec   float64       `json:"throughput_per_sec"`
	LastProcessedAt    time.Time     `json:"last_processed_at"`
}

// ProcessingConfig represents configuration for document processing
type ProcessingConfig struct {
	// Pipeline configuration
	PipelineStages []StageConfig `json:"pipeline_stages"`

	// Chunking configuration
	ChunkSize    int `json:"chunk_size"`
	ChunkOverlap int `json:"chunk_overlap"`

	// Processing options
	MaxConcurrency    int           `json:"max_concurrency"`
	ProcessingTimeout time.Duration `json:"processing_timeout"`
	RetryAttempts     int           `json:"retry_attempts"`
	RetryDelay        time.Duration `json:"retry_delay"`

	// Content extraction options
	ExtractImages      bool `json:"extract_images"`
	ExtractTables      bool `json:"extract_tables"`
	ExtractMetadata    bool `json:"extract_metadata"`
	PreserveFormatting bool `json:"preserve_formatting"`

	// Text processing options
	CleanText          bool     `json:"clean_text"`
	DetectLanguage     bool     `json:"detect_language"`
	NormalizeText      bool     `json:"normalize_text"`
	RemoveStopWords    bool     `json:"remove_stop_words"`
	SupportedLanguages []string `json:"supported_languages"`

	// Output options
	OutputFormat    string `json:"output_format"`
	IncludeMetadata bool   `json:"include_metadata"`
	IncludeChunks   bool   `json:"include_chunks"`

	// Integration options
	VectorStoreConfig *VectorStoreIntegration `json:"vector_store_config,omitempty"`
	MemoryConfig      *MemoryIntegration      `json:"memory_config,omitempty"`

	// Monitoring options
	EnableMetrics bool   `json:"enable_metrics"`
	EnableTracing bool   `json:"enable_tracing"`
	LogLevel      string `json:"log_level"`

	// Custom options
	CustomOptions map[string]interface{} `json:"custom_options,omitempty"`
}

// StageConfig represents configuration for a processing stage
type StageConfig struct {
	Name    string                 `json:"name"`
	Type    string                 `json:"type"`
	Enabled bool                   `json:"enabled"`
	Order   int                    `json:"order"`
	Options map[string]interface{} `json:"options,omitempty"`
}

// VectorStoreIntegration represents vector store integration configuration
type VectorStoreIntegration struct {
	Enabled        bool   `json:"enabled"`
	CollectionName string `json:"collection_name"`
	AutoEmbed      bool   `json:"auto_embed"`
	UpdateExisting bool   `json:"update_existing"`
}

// MemoryIntegration represents memory integration configuration
type MemoryIntegration struct {
	Enabled    bool   `json:"enabled"`
	MemoryType string `json:"memory_type"`
	Namespace  string `json:"namespace"`
}

// ProcessingError represents a processing error with context
type ProcessingError struct {
	Stage     string                 `json:"stage"`
	Message   string                 `json:"message"`
	Cause     error                  `json:"cause,omitempty"`
	Document  *Document              `json:"document,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
}

func (e *ProcessingError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s (caused by: %v)", e.Stage, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Stage, e.Message)
}

// ProcessingEvent represents an event during processing
type ProcessingEvent struct {
	Type      string                 `json:"type"`
	Stage     string                 `json:"stage"`
	Document  *Document              `json:"document,omitempty"`
	Message   string                 `json:"message"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
}

// EventType constants
const (
	EventTypeStarted   = "started"
	EventTypeCompleted = "completed"
	EventTypeFailed    = "failed"
	EventTypeProgress  = "progress"
	EventTypeWarning   = "warning"
)

// ProcessingEventHandler defines the interface for handling processing events
type ProcessingEventHandler interface {
	// HandleEvent handles a processing event
	HandleEvent(ctx context.Context, event *ProcessingEvent) error

	// GetHandlerType returns the handler type
	GetHandlerType() string
}

// DocumentFilter defines the interface for filtering documents
type DocumentFilter interface {
	// ShouldProcess determines if a document should be processed
	ShouldProcess(ctx context.Context, doc *Document) (bool, error)

	// GetFilterType returns the filter type
	GetFilterType() string

	// Configure configures the filter with options
	Configure(options map[string]interface{}) error
}

// DocumentValidator defines the interface for validating documents
type DocumentValidator interface {
	// Validate validates a document
	Validate(ctx context.Context, doc *Document) error

	// GetValidatorType returns the validator type
	GetValidatorType() string

	// Configure configures the validator with options
	Configure(options map[string]interface{}) error
}

// ProcessingManager defines the interface for managing document processing
type ProcessingManager interface {
	// CreateProcessor creates a new document processor
	CreateProcessor(config *ProcessingConfig) (DocumentProcessor, error)

	// RegisterExtractor registers a content extractor
	RegisterExtractor(extractor ContentExtractor) error

	// RegisterProcessor registers a text processor
	RegisterProcessor(processor TextProcessor) error

	// RegisterChunker registers a document chunker
	RegisterChunker(chunker DocumentChunker) error

	// RegisterSource registers a document source
	RegisterSource(source DocumentSource) error

	// GetExtractor returns a content extractor for the given content type
	GetExtractor(contentType string) (ContentExtractor, error)

	// GetProcessor returns a text processor by type
	GetProcessor(processorType string) (TextProcessor, error)

	// GetChunker returns a document chunker by type
	GetChunker(chunkerType string) (DocumentChunker, error)

	// GetSource returns a document source by type
	GetSource(sourceType string) (DocumentSource, error)

	// ListExtractors returns all registered extractors
	ListExtractors() []ContentExtractor

	// ListProcessors returns all registered processors
	ListProcessors() []TextProcessor

	// ListChunkers returns all registered chunkers
	ListChunkers() []DocumentChunker

	// ListSources returns all registered sources
	ListSources() []DocumentSource
}
