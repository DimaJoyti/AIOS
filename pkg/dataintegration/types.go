package dataintegration

import (
	"context"
	"time"
)

// External Data Integration types and interfaces for AIOS

// DataIntegrationEngine defines the interface for managing data integration
type DataIntegrationEngine interface {
	// Data Source Management
	CreateDataSource(source *DataSource) (*DataSource, error)
	GetDataSource(sourceID string) (*DataSource, error)
	UpdateDataSource(source *DataSource) (*DataSource, error)
	DeleteDataSource(sourceID string) error
	ListDataSources(filter *DataSourceFilter) ([]*DataSource, error)

	// Data Source Operations
	EnableDataSource(sourceID string) error
	DisableDataSource(sourceID string) error
	TestDataSource(sourceID string) (*DataSourceTestResult, error)
	RefreshDataSource(sourceID string) error

	// Connector Management
	RegisterConnector(connector DataConnector) error
	GetConnector(connectorType string) (DataConnector, error)
	ListConnectors() []string

	// Pipeline Management
	CreatePipeline(pipeline *DataPipeline) (*DataPipeline, error)
	GetPipeline(pipelineID string) (*DataPipeline, error)
	UpdatePipeline(pipeline *DataPipeline) (*DataPipeline, error)
	DeletePipeline(pipelineID string) error
	ListPipelines(filter *PipelineFilter) ([]*DataPipeline, error)

	// Pipeline Operations
	StartPipeline(pipelineID string) error
	StopPipeline(pipelineID string) error
	GetPipelineStatus(pipelineID string) (*PipelineStatus, error)
	GetPipelineMetrics(pipelineID string, timeRange *TimeRange) (*PipelineMetrics, error)

	// Data Processing
	ProcessData(ctx context.Context, data *DataRecord, pipeline *DataPipeline) (*ProcessedData, error)
	ValidateData(data *DataRecord, schema *DataSchema) (*ValidationResult, error)
	TransformData(data *DataRecord, transformations []*DataTransformation) (*DataRecord, error)

	// Storage Management
	StoreData(ctx context.Context, data *ProcessedData, storage *StorageConfig) error
	RetrieveData(ctx context.Context, query *DataQuery) (*DataResult, error)
	IndexData(ctx context.Context, data *ProcessedData, indexConfig *IndexConfig) error

	// Monitoring and Analytics
	GetDataSourceHealth(sourceID string) (*DataSourceHealth, error)
	GetDataSourceMetrics(sourceID string, timeRange *TimeRange) (*DataSourceMetrics, error)
	GetDataSourceLogs(sourceID string, filter *LogFilter) ([]*DataSourceLog, error)

	// Event Management
	PublishDataEvent(event *DataEvent) error
	SubscribeToDataEvents(eventType string, handler DataEventHandler) error
	UnsubscribeFromDataEvents(eventType string, handler DataEventHandler) error
}

// DataConnector defines the interface for data connectors
type DataConnector interface {
	// Connector Information
	GetType() string
	GetName() string
	GetDescription() string
	GetVersion() string
	GetSupportedOperations() []string

	// Configuration
	GetConfigSchema() *ConnectorConfigSchema
	ValidateConfig(config map[string]interface{}) error

	// Connection Management
	Connect(ctx context.Context, config map[string]interface{}) error
	Disconnect(ctx context.Context) error
	IsConnected() bool
	TestConnection(ctx context.Context) error

	// Data Operations
	ExtractData(ctx context.Context, params *ExtractionParams) (*DataExtraction, error)
	StreamData(ctx context.Context, params *StreamParams) (<-chan *DataRecord, error)

	// Health and Monitoring
	GetHealth() *ConnectorHealth
	GetMetrics() *ConnectorMetrics
}

// DataEventHandler defines the interface for data event handlers
type DataEventHandler interface {
	HandleDataEvent(ctx context.Context, event *DataEvent) error
	GetEventTypes() []string
}

// Core Types

// DataSource represents a data source configuration
type DataSource struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Type        string                 `json:"type"`
	Provider    string                 `json:"provider"`
	Status      DataSourceStatus       `json:"status"`
	Config      *DataSourceConfig      `json:"config"`
	Credentials *DataSourceCredentials `json:"credentials,omitempty"`
	Settings    *DataSourceSettings    `json:"settings"`
	Schema      *DataSchema            `json:"schema,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	CreatedBy   string                 `json:"created_by"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	LastSyncAt  *time.Time             `json:"last_sync_at,omitempty"`
}

// DataSourceStatus defines the status of a data source
type DataSourceStatus string

const (
	DataSourceStatusActive      DataSourceStatus = "active"
	DataSourceStatusInactive    DataSourceStatus = "inactive"
	DataSourceStatusError       DataSourceStatus = "error"
	DataSourceStatusConfiguring DataSourceStatus = "configuring"
	DataSourceStatusSyncing     DataSourceStatus = "syncing"
)

// DataSourceConfig contains data source configuration
type DataSourceConfig struct {
	URL          string                 `json:"url,omitempty"`
	Endpoint     string                 `json:"endpoint,omitempty"`
	Method       string                 `json:"method,omitempty"`
	Headers      map[string]string      `json:"headers,omitempty"`
	QueryParams  map[string]string      `json:"query_params,omitempty"`
	Timeout      time.Duration          `json:"timeout,omitempty"`
	RetryPolicy  *RetryPolicy           `json:"retry_policy,omitempty"`
	RateLimit    *RateLimit             `json:"rate_limit,omitempty"`
	CrawlConfig  *CrawlConfig           `json:"crawl_config,omitempty"`
	StreamConfig *StreamConfig          `json:"stream_config,omitempty"`
	Custom       map[string]interface{} `json:"custom,omitempty"`
}

// DataSourceCredentials contains authentication credentials
type DataSourceCredentials struct {
	Type         CredentialType         `json:"type"`
	APIKey       string                 `json:"api_key,omitempty"`
	SecretKey    string                 `json:"secret_key,omitempty"`
	Token        string                 `json:"token,omitempty"`
	RefreshToken string                 `json:"refresh_token,omitempty"`
	Username     string                 `json:"username,omitempty"`
	Password     string                 `json:"password,omitempty"`
	OAuth        *OAuthCredentials      `json:"oauth,omitempty"`
	Custom       map[string]interface{} `json:"custom,omitempty"`
	ExpiresAt    *time.Time             `json:"expires_at,omitempty"`
}

// CredentialType defines the type of credentials
type CredentialType string

const (
	CredentialTypeAPIKey CredentialType = "api_key"
	CredentialTypeBearer CredentialType = "bearer"
	CredentialTypeBasic  CredentialType = "basic"
	CredentialTypeOAuth2 CredentialType = "oauth2"
	CredentialTypeCustom CredentialType = "custom"
)

// OAuthCredentials contains OAuth-specific credentials
type OAuthCredentials struct {
	ClientID     string    `json:"client_id"`
	ClientSecret string    `json:"client_secret"`
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	TokenType    string    `json:"token_type"`
	Scope        []string  `json:"scope,omitempty"`
	ExpiresAt    time.Time `json:"expires_at"`
}

// DataSourceSettings contains data source behavior settings
type DataSourceSettings struct {
	AutoSync        bool                   `json:"auto_sync"`
	SyncInterval    time.Duration          `json:"sync_interval"`
	EnableStreaming bool                   `json:"enable_streaming"`
	EnableEvents    bool                   `json:"enable_events"`
	DataRetention   time.Duration          `json:"data_retention"`
	MaxRecords      int64                  `json:"max_records"`
	LogLevel        string                 `json:"log_level"`
	NotifyOnError   bool                   `json:"notify_on_error"`
	ErrorRecipients []string               `json:"error_recipients,omitempty"`
	CustomSettings  map[string]interface{} `json:"custom_settings,omitempty"`
}

// RetryPolicy defines retry behavior
type RetryPolicy struct {
	MaxRetries      int           `json:"max_retries"`
	InitialDelay    time.Duration `json:"initial_delay"`
	MaxDelay        time.Duration `json:"max_delay"`
	BackoffFactor   float64       `json:"backoff_factor"`
	RetryableErrors []string      `json:"retryable_errors,omitempty"`
}

// RateLimit defines rate limiting configuration
type RateLimit struct {
	RequestsPerSecond int           `json:"requests_per_second"`
	BurstSize         int           `json:"burst_size"`
	WindowSize        time.Duration `json:"window_size"`
}

// CrawlConfig contains web crawling configuration
type CrawlConfig struct {
	MaxDepth        int           `json:"max_depth"`
	MaxPages        int           `json:"max_pages"`
	FollowRedirects bool          `json:"follow_redirects"`
	RespectRobots   bool          `json:"respect_robots"`
	UserAgent       string        `json:"user_agent"`
	Delay           time.Duration `json:"delay"`
	Selectors       []string      `json:"selectors,omitempty"`
	ExcludePatterns []string      `json:"exclude_patterns,omitempty"`
	IncludePatterns []string      `json:"include_patterns,omitempty"`
}

// StreamConfig contains streaming configuration
type StreamConfig struct {
	Protocol       string            `json:"protocol"`
	BufferSize     int               `json:"buffer_size"`
	ReconnectDelay time.Duration     `json:"reconnect_delay"`
	MaxReconnects  int               `json:"max_reconnects"`
	Headers        map[string]string `json:"headers,omitempty"`
	Filters        []string          `json:"filters,omitempty"`
}

// DataPipeline represents a data processing pipeline
type DataPipeline struct {
	ID              string                 `json:"id"`
	Name            string                 `json:"name"`
	Description     string                 `json:"description"`
	DataSourceID    string                 `json:"data_source_id"`
	Status          PipelineStatus         `json:"status"`
	Config          *PipelineConfig        `json:"config"`
	Transformations []*DataTransformation  `json:"transformations"`
	Storage         *StorageConfig         `json:"storage"`
	Schedule        *ScheduleConfig        `json:"schedule,omitempty"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
	CreatedBy       string                 `json:"created_by"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
	LastRunAt       *time.Time             `json:"last_run_at,omitempty"`
}

// PipelineStatus defines the status of a pipeline
type PipelineStatus string

const (
	PipelineStatusRunning   PipelineStatus = "running"
	PipelineStatusStopped   PipelineStatus = "stopped"
	PipelineStatusError     PipelineStatus = "error"
	PipelineStatusCompleted PipelineStatus = "completed"
	PipelineStatusScheduled PipelineStatus = "scheduled"
)

// PipelineConfig contains pipeline configuration
type PipelineConfig struct {
	BatchSize        int                    `json:"batch_size"`
	Parallelism      int                    `json:"parallelism"`
	Timeout          time.Duration          `json:"timeout"`
	ErrorHandling    ErrorHandlingStrategy  `json:"error_handling"`
	ValidationMode   ValidationMode         `json:"validation_mode"`
	DeduplicationKey string                 `json:"deduplication_key,omitempty"`
	Custom           map[string]interface{} `json:"custom,omitempty"`
}

// ErrorHandlingStrategy defines how to handle errors
type ErrorHandlingStrategy string

const (
	ErrorHandlingStop       ErrorHandlingStrategy = "stop"
	ErrorHandlingSkip       ErrorHandlingStrategy = "skip"
	ErrorHandlingRetry      ErrorHandlingStrategy = "retry"
	ErrorHandlingDeadLetter ErrorHandlingStrategy = "dead_letter"
)

// ValidationMode defines validation behavior
type ValidationMode string

const (
	ValidationModeStrict ValidationMode = "strict"
	ValidationModeWarn   ValidationMode = "warn"
	ValidationModeSkip   ValidationMode = "skip"
)

// DataTransformation represents a data transformation step
type DataTransformation struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Type        TransformationType     `json:"type"`
	Config      map[string]interface{} `json:"config"`
	Order       int                    `json:"order"`
	Enabled     bool                   `json:"enabled"`
	Description string                 `json:"description,omitempty"`
}

// TransformationType defines the type of transformation
type TransformationType string

const (
	TransformationTypeMap       TransformationType = "map"
	TransformationTypeFilter    TransformationType = "filter"
	TransformationTypeEnrich    TransformationType = "enrich"
	TransformationTypeValidate  TransformationType = "validate"
	TransformationTypeAggregate TransformationType = "aggregate"
	TransformationTypeCustom    TransformationType = "custom"
)

// StorageConfig contains storage configuration
type StorageConfig struct {
	Type         StorageType            `json:"type"`
	Connection   string                 `json:"connection"`
	Database     string                 `json:"database,omitempty"`
	Collection   string                 `json:"collection,omitempty"`
	Table        string                 `json:"table,omitempty"`
	IndexConfig  *IndexConfig           `json:"index_config,omitempty"`
	Partitioning *PartitionConfig       `json:"partitioning,omitempty"`
	Compression  string                 `json:"compression,omitempty"`
	Custom       map[string]interface{} `json:"custom,omitempty"`
}

// StorageType defines the type of storage
type StorageType string

const (
	StorageTypePostgreSQL StorageType = "postgresql"
	StorageTypeMongoDB    StorageType = "mongodb"
	StorageTypeElastic    StorageType = "elasticsearch"
	StorageTypeRedis      StorageType = "redis"
	StorageTypeS3         StorageType = "s3"
	StorageTypeFile       StorageType = "file"
	StorageTypeCustom     StorageType = "custom"
)

// IndexConfig contains indexing configuration
type IndexConfig struct {
	Fields     []string               `json:"fields"`
	Type       string                 `json:"type"`
	Options    map[string]interface{} `json:"options,omitempty"`
	Unique     bool                   `json:"unique"`
	Sparse     bool                   `json:"sparse"`
	Background bool                   `json:"background"`
}

// PartitionConfig contains partitioning configuration
type PartitionConfig struct {
	Strategy string                 `json:"strategy"`
	Field    string                 `json:"field"`
	Options  map[string]interface{} `json:"options,omitempty"`
}

// ScheduleConfig contains scheduling configuration
type ScheduleConfig struct {
	Type       ScheduleType `json:"type"`
	Expression string       `json:"expression"`
	Timezone   string       `json:"timezone,omitempty"`
	StartTime  *time.Time   `json:"start_time,omitempty"`
	EndTime    *time.Time   `json:"end_time,omitempty"`
	Enabled    bool         `json:"enabled"`
}

// ScheduleType defines the type of schedule
type ScheduleType string

const (
	ScheduleTypeCron     ScheduleType = "cron"
	ScheduleTypeInterval ScheduleType = "interval"
	ScheduleTypeOnce     ScheduleType = "once"
)

// Data Types

// DataRecord represents a single data record
type DataRecord struct {
	ID        string                 `json:"id"`
	SourceID  string                 `json:"source_id"`
	Data      map[string]interface{} `json:"data"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	Schema    *DataSchema            `json:"schema,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
	Hash      string                 `json:"hash,omitempty"`
}

// ProcessedData represents processed data ready for storage
type ProcessedData struct {
	Records      []*DataRecord          `json:"records"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	ProcessedAt  time.Time              `json:"processed_at"`
	PipelineID   string                 `json:"pipeline_id"`
	SourceID     string                 `json:"source_id"`
	TotalRecords int                    `json:"total_records"`
	ValidRecords int                    `json:"valid_records"`
	ErrorRecords int                    `json:"error_records"`
}

// DataSchema represents the schema of data
type DataSchema struct {
	Name        string                       `json:"name"`
	Version     string                       `json:"version"`
	Fields      map[string]*SchemaField      `json:"fields"`
	Required    []string                     `json:"required,omitempty"`
	Constraints map[string]*SchemaConstraint `json:"constraints,omitempty"`
	Metadata    map[string]interface{}       `json:"metadata,omitempty"`
}

// SchemaField represents a field in the schema
type SchemaField struct {
	Type        string      `json:"type"`
	Description string      `json:"description,omitempty"`
	Format      string      `json:"format,omitempty"`
	Default     interface{} `json:"default,omitempty"`
	Nullable    bool        `json:"nullable"`
	Enum        []string    `json:"enum,omitempty"`
	Pattern     string      `json:"pattern,omitempty"`
	MinLength   *int        `json:"min_length,omitempty"`
	MaxLength   *int        `json:"max_length,omitempty"`
	Minimum     *float64    `json:"minimum,omitempty"`
	Maximum     *float64    `json:"maximum,omitempty"`
}

// SchemaConstraint represents a constraint on the schema
type SchemaConstraint struct {
	Type       string                 `json:"type"`
	Fields     []string               `json:"fields"`
	Expression string                 `json:"expression,omitempty"`
	Message    string                 `json:"message,omitempty"`
	Options    map[string]interface{} `json:"options,omitempty"`
}

// TimeRange represents a time range for queries
type TimeRange struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// Operation Types

// ExtractionParams contains parameters for data extraction
type ExtractionParams struct {
	Query     string                 `json:"query,omitempty"`
	Filters   map[string]interface{} `json:"filters,omitempty"`
	Limit     int                    `json:"limit,omitempty"`
	Offset    int                    `json:"offset,omitempty"`
	SortBy    string                 `json:"sort_by,omitempty"`
	SortOrder string                 `json:"sort_order,omitempty"`
	Fields    []string               `json:"fields,omitempty"`
	TimeRange *TimeRange             `json:"time_range,omitempty"`
	Custom    map[string]interface{} `json:"custom,omitempty"`
}

// StreamParams contains parameters for data streaming
type StreamParams struct {
	Filters    map[string]interface{} `json:"filters,omitempty"`
	BufferSize int                    `json:"buffer_size,omitempty"`
	Timeout    time.Duration          `json:"timeout,omitempty"`
	Checkpoint string                 `json:"checkpoint,omitempty"`
	Custom     map[string]interface{} `json:"custom,omitempty"`
}

// DataExtraction represents extracted data
type DataExtraction struct {
	Records     []*DataRecord          `json:"records"`
	TotalCount  int64                  `json:"total_count"`
	HasMore     bool                   `json:"has_more"`
	NextCursor  string                 `json:"next_cursor,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	ExtractedAt time.Time              `json:"extracted_at"`
	Duration    time.Duration          `json:"duration"`
}

// DataQuery represents a data query
type DataQuery struct {
	SourceID     string                 `json:"source_id,omitempty"`
	Query        string                 `json:"query"`
	Filters      map[string]interface{} `json:"filters,omitempty"`
	Limit        int                    `json:"limit,omitempty"`
	Offset       int                    `json:"offset,omitempty"`
	SortBy       string                 `json:"sort_by,omitempty"`
	SortOrder    string                 `json:"sort_order,omitempty"`
	TimeRange    *TimeRange             `json:"time_range,omitempty"`
	Aggregations map[string]interface{} `json:"aggregations,omitempty"`
}

// DataResult represents query results
type DataResult struct {
	Records      []*DataRecord          `json:"records"`
	TotalCount   int64                  `json:"total_count"`
	Aggregations map[string]interface{} `json:"aggregations,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	QueryTime    time.Duration          `json:"query_time"`
	ExecutedAt   time.Time              `json:"executed_at"`
}

// Testing Types

// DataSourceTestResult represents the result of a data source test
type DataSourceTestResult struct {
	Success      bool                   `json:"success"`
	Message      string                 `json:"message"`
	Details      map[string]interface{} `json:"details,omitempty"`
	Duration     time.Duration          `json:"duration"`
	TestedAt     time.Time              `json:"tested_at"`
	Capabilities []string               `json:"capabilities,omitempty"`
}

// ValidationResult represents data validation results
type ValidationResult struct {
	Valid        bool                `json:"valid"`
	Errors       []ValidationError   `json:"errors,omitempty"`
	Warnings     []ValidationWarning `json:"warnings,omitempty"`
	ValidRecords int                 `json:"valid_records"`
	TotalRecords int                 `json:"total_records"`
	ValidatedAt  time.Time           `json:"validated_at"`
}

// ValidationError represents a validation error
type ValidationError struct {
	Field   string      `json:"field"`
	Message string      `json:"message"`
	Code    string      `json:"code"`
	Value   interface{} `json:"value,omitempty"`
}

// ValidationWarning represents a validation warning
type ValidationWarning struct {
	Field   string      `json:"field"`
	Message string      `json:"message"`
	Code    string      `json:"code"`
	Value   interface{} `json:"value,omitempty"`
}

// Monitoring Types

// DataSourceHealth represents the health status of a data source
type DataSourceHealth struct {
	SourceID  string                 `json:"source_id"`
	Status    HealthStatus           `json:"status"`
	Message   string                 `json:"message"`
	LastCheck time.Time              `json:"last_check"`
	Uptime    time.Duration          `json:"uptime"`
	Checks    []*HealthCheck         `json:"checks"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// HealthStatus defines the health status
type HealthStatus string

const (
	HealthStatusHealthy   HealthStatus = "healthy"
	HealthStatusDegraded  HealthStatus = "degraded"
	HealthStatusUnhealthy HealthStatus = "unhealthy"
	HealthStatusUnknown   HealthStatus = "unknown"
)

// HealthCheck represents a specific health check
type HealthCheck struct {
	Name      string                 `json:"name"`
	Status    HealthStatus           `json:"status"`
	Message   string                 `json:"message"`
	Duration  time.Duration          `json:"duration"`
	Timestamp time.Time              `json:"timestamp"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// DataSourceMetrics contains data source performance metrics
type DataSourceMetrics struct {
	SourceID         string                 `json:"source_id"`
	TimeRange        *TimeRange             `json:"time_range"`
	RecordsExtracted int64                  `json:"records_extracted"`
	RecordsProcessed int64                  `json:"records_processed"`
	RecordsStored    int64                  `json:"records_stored"`
	ErrorCount       int64                  `json:"error_count"`
	SuccessRate      float64                `json:"success_rate"`
	AverageLatency   time.Duration          `json:"average_latency"`
	P95Latency       time.Duration          `json:"p95_latency"`
	P99Latency       time.Duration          `json:"p99_latency"`
	ThroughputPerSec float64                `json:"throughput_per_sec"`
	ErrorsByType     map[string]int64       `json:"errors_by_type"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
}

// PipelineMetrics contains pipeline performance metrics
type PipelineMetrics struct {
	PipelineID       string                 `json:"pipeline_id"`
	TimeRange        *TimeRange             `json:"time_range"`
	RunCount         int64                  `json:"run_count"`
	SuccessfulRuns   int64                  `json:"successful_runs"`
	FailedRuns       int64                  `json:"failed_runs"`
	RecordsProcessed int64                  `json:"records_processed"`
	AverageRunTime   time.Duration          `json:"average_run_time"`
	LastRunTime      time.Duration          `json:"last_run_time"`
	TotalDataSize    int64                  `json:"total_data_size"`
	ErrorsByStage    map[string]int64       `json:"errors_by_stage"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
}

// ConnectorHealth represents connector-specific health information
type ConnectorHealth struct {
	Status       HealthStatus           `json:"status"`
	Message      string                 `json:"message"`
	Connected    bool                   `json:"connected"`
	LastActivity time.Time              `json:"last_activity"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// ConnectorMetrics contains connector performance metrics
type ConnectorMetrics struct {
	OperationCount   map[string]int64         `json:"operation_count"`
	AverageLatency   map[string]time.Duration `json:"average_latency"`
	ErrorCount       map[string]int64         `json:"error_count"`
	LastOperation    time.Time                `json:"last_operation"`
	ConnectionUptime time.Duration            `json:"connection_uptime"`
	Metadata         map[string]interface{}   `json:"metadata,omitempty"`
}

// Logging Types

// DataSourceLog represents a data source log entry
type DataSourceLog struct {
	ID        string                 `json:"id"`
	SourceID  string                 `json:"source_id"`
	Level     LogLevel               `json:"level"`
	Message   string                 `json:"message"`
	Operation string                 `json:"operation,omitempty"`
	Duration  time.Duration          `json:"duration,omitempty"`
	Error     string                 `json:"error,omitempty"`
	Data      map[string]interface{} `json:"data,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
}

// LogLevel defines the log level
type LogLevel string

const (
	LogLevelDebug LogLevel = "debug"
	LogLevelInfo  LogLevel = "info"
	LogLevelWarn  LogLevel = "warn"
	LogLevelError LogLevel = "error"
)

// Event Types

// DataEvent represents a data integration event
type DataEvent struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Source      string                 `json:"source"`
	SourceID    string                 `json:"source_id"`
	Data        map[string]interface{} `json:"data"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	Timestamp   time.Time              `json:"timestamp"`
	ProcessedAt *time.Time             `json:"processed_at,omitempty"`
}

// Configuration Types

// ConnectorConfigSchema defines the configuration schema for a connector
type ConnectorConfigSchema struct {
	Properties map[string]*ConfigProperty `json:"properties"`
	Required   []string                   `json:"required"`
}

// ConfigProperty defines a configuration property
type ConfigProperty struct {
	Type        string      `json:"type"`
	Description string      `json:"description"`
	Default     interface{} `json:"default,omitempty"`
	Enum        []string    `json:"enum,omitempty"`
	Format      string      `json:"format,omitempty"`
	Sensitive   bool        `json:"sensitive"`
}

// Filter Types

// DataSourceFilter contains filters for data source queries
type DataSourceFilter struct {
	Type      string           `json:"type,omitempty"`
	Provider  string           `json:"provider,omitempty"`
	Status    DataSourceStatus `json:"status,omitempty"`
	CreatedBy string           `json:"created_by,omitempty"`
	Search    string           `json:"search,omitempty"`
	Limit     int              `json:"limit,omitempty"`
	Offset    int              `json:"offset,omitempty"`
}

// PipelineFilter contains filters for pipeline queries
type PipelineFilter struct {
	DataSourceID string         `json:"data_source_id,omitempty"`
	Status       PipelineStatus `json:"status,omitempty"`
	CreatedBy    string         `json:"created_by,omitempty"`
	Search       string         `json:"search,omitempty"`
	Limit        int            `json:"limit,omitempty"`
	Offset       int            `json:"offset,omitempty"`
}

// LogFilter contains filters for log queries
type LogFilter struct {
	Level     LogLevel   `json:"level,omitempty"`
	Operation string     `json:"operation,omitempty"`
	Since     *time.Time `json:"since,omitempty"`
	Until     *time.Time `json:"until,omitempty"`
	Search    string     `json:"search,omitempty"`
	Limit     int        `json:"limit,omitempty"`
	Offset    int        `json:"offset,omitempty"`
}
