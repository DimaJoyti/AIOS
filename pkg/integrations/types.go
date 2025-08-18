package integrations

import (
	"context"
	"time"
)

// Integration APIs types and interfaces for AIOS

// IntegrationEngine defines the interface for managing integrations
type IntegrationEngine interface {
	// Integration Management
	CreateIntegration(integration *Integration) (*Integration, error)
	GetIntegration(integrationID string) (*Integration, error)
	UpdateIntegration(integration *Integration) (*Integration, error)
	DeleteIntegration(integrationID string) error
	ListIntegrations(filter *IntegrationFilter) ([]*Integration, error)
	
	// Integration Operations
	EnableIntegration(integrationID string) error
	DisableIntegration(integrationID string) error
	TestIntegration(integrationID string) (*IntegrationTestResult, error)
	RefreshIntegration(integrationID string) error
	
	// Adapter Management
	RegisterAdapter(adapter IntegrationAdapter) error
	GetAdapter(adapterType string) (IntegrationAdapter, error)
	ListAdapters() []string
	
	// Webhook Management
	CreateWebhook(webhook *Webhook) (*Webhook, error)
	GetWebhook(webhookID string) (*Webhook, error)
	UpdateWebhook(webhook *Webhook) (*Webhook, error)
	DeleteWebhook(webhookID string) error
	ListWebhooks(filter *WebhookFilter) ([]*Webhook, error)
	ProcessWebhook(webhookID string, payload []byte, headers map[string]string) error
	
	// Event Management
	PublishEvent(event *IntegrationEvent) error
	SubscribeToEvents(eventType string, handler EventHandler) error
	UnsubscribeFromEvents(eventType string, handler EventHandler) error
	
	// Configuration Management
	GetConfiguration(integrationID string) (*IntegrationConfig, error)
	UpdateConfiguration(integrationID string, config *IntegrationConfig) error
	ValidateConfiguration(config *IntegrationConfig) error
	
	// Monitoring and Analytics
	GetIntegrationMetrics(integrationID string, timeRange *TimeRange) (*IntegrationMetrics, error)
	GetIntegrationHealth(integrationID string) (*IntegrationHealth, error)
	GetIntegrationLogs(integrationID string, filter *LogFilter) ([]*IntegrationLog, error)
}

// IntegrationAdapter defines the interface for integration adapters
type IntegrationAdapter interface {
	// Adapter Information
	GetType() string
	GetName() string
	GetDescription() string
	GetVersion() string
	GetSupportedOperations() []string
	
	// Configuration
	GetConfigSchema() *ConfigSchema
	ValidateConfig(config map[string]interface{}) error
	
	// Connection Management
	Connect(ctx context.Context, config map[string]interface{}) error
	Disconnect(ctx context.Context) error
	IsConnected() bool
	TestConnection(ctx context.Context) error
	
	// Operations
	Execute(ctx context.Context, operation string, params map[string]interface{}) (*OperationResult, error)
	
	// Event Handling
	SupportsWebhooks() bool
	GetWebhookConfig() *WebhookConfig
	ProcessWebhookPayload(payload []byte, headers map[string]string) (*IntegrationEvent, error)
	
	// Health and Monitoring
	GetHealth() *AdapterHealth
	GetMetrics() *AdapterMetrics
}

// EventHandler defines the interface for event handlers
type EventHandler interface {
	HandleEvent(ctx context.Context, event *IntegrationEvent) error
	GetEventTypes() []string
}

// Core Types

// Integration represents an integration configuration
type Integration struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Type        string                 `json:"type"`
	Provider    string                 `json:"provider"`
	Status      IntegrationStatus      `json:"status"`
	Config      *IntegrationConfig     `json:"config"`
	Credentials *IntegrationCredentials `json:"credentials,omitempty"`
	Settings    *IntegrationSettings   `json:"settings"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	CreatedBy   string                 `json:"created_by"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	LastSyncAt  *time.Time             `json:"last_sync_at,omitempty"`
}

// IntegrationStatus defines the status of an integration
type IntegrationStatus string

const (
	IntegrationStatusActive     IntegrationStatus = "active"
	IntegrationStatusInactive   IntegrationStatus = "inactive"
	IntegrationStatusError      IntegrationStatus = "error"
	IntegrationStatusConfiguring IntegrationStatus = "configuring"
	IntegrationStatusTesting    IntegrationStatus = "testing"
)

// IntegrationConfig contains integration configuration
type IntegrationConfig struct {
	BaseURL     string                 `json:"base_url,omitempty"`
	APIVersion  string                 `json:"api_version,omitempty"`
	Timeout     time.Duration          `json:"timeout,omitempty"`
	RetryPolicy *RetryPolicy           `json:"retry_policy,omitempty"`
	RateLimit   *RateLimit             `json:"rate_limit,omitempty"`
	Custom      map[string]interface{} `json:"custom,omitempty"`
}

// IntegrationCredentials contains authentication credentials
type IntegrationCredentials struct {
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
	CredentialTypeAPIKey    CredentialType = "api_key"
	CredentialTypeBearer    CredentialType = "bearer"
	CredentialTypeBasic     CredentialType = "basic"
	CredentialTypeOAuth2    CredentialType = "oauth2"
	CredentialTypeCustom    CredentialType = "custom"
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

// IntegrationSettings contains integration behavior settings
type IntegrationSettings struct {
	AutoSync         bool                   `json:"auto_sync"`
	SyncInterval     time.Duration          `json:"sync_interval"`
	EnableWebhooks   bool                   `json:"enable_webhooks"`
	EnableEvents     bool                   `json:"enable_events"`
	LogLevel         string                 `json:"log_level"`
	NotifyOnError    bool                   `json:"notify_on_error"`
	ErrorRecipients  []string               `json:"error_recipients,omitempty"`
	CustomSettings   map[string]interface{} `json:"custom_settings,omitempty"`
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

// Webhook Types

// Webhook represents a webhook configuration
type Webhook struct {
	ID            string                 `json:"id"`
	IntegrationID string                 `json:"integration_id"`
	Name          string                 `json:"name"`
	URL           string                 `json:"url"`
	Method        string                 `json:"method"`
	Headers       map[string]string      `json:"headers,omitempty"`
	Secret        string                 `json:"secret,omitempty"`
	Events        []string               `json:"events"`
	Status        WebhookStatus          `json:"status"`
	Config        *WebhookConfig         `json:"config"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
	LastTriggered *time.Time             `json:"last_triggered,omitempty"`
}

// WebhookStatus defines the status of a webhook
type WebhookStatus string

const (
	WebhookStatusActive   WebhookStatus = "active"
	WebhookStatusInactive WebhookStatus = "inactive"
	WebhookStatusError    WebhookStatus = "error"
)

// WebhookConfig contains webhook configuration
type WebhookConfig struct {
	Timeout         time.Duration          `json:"timeout"`
	RetryPolicy     *RetryPolicy           `json:"retry_policy,omitempty"`
	SignatureHeader string                 `json:"signature_header,omitempty"`
	ContentType     string                 `json:"content_type"`
	Custom          map[string]interface{} `json:"custom,omitempty"`
}

// Event Types

// IntegrationEvent represents an integration event
type IntegrationEvent struct {
	ID            string                 `json:"id"`
	Type          string                 `json:"type"`
	Source        string                 `json:"source"`
	IntegrationID string                 `json:"integration_id"`
	Data          map[string]interface{} `json:"data"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
	Timestamp     time.Time              `json:"timestamp"`
	ProcessedAt   *time.Time             `json:"processed_at,omitempty"`
}

// Configuration Types

// ConfigSchema defines the configuration schema for an adapter
type ConfigSchema struct {
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

// Operation Types

// OperationResult represents the result of an operation
type OperationResult struct {
	Success   bool                   `json:"success"`
	Data      map[string]interface{} `json:"data,omitempty"`
	Error     string                 `json:"error,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	Duration  time.Duration          `json:"duration"`
	Timestamp time.Time              `json:"timestamp"`
}

// Testing Types

// IntegrationTestResult represents the result of an integration test
type IntegrationTestResult struct {
	Success      bool                   `json:"success"`
	Message      string                 `json:"message"`
	Details      map[string]interface{} `json:"details,omitempty"`
	Duration     time.Duration          `json:"duration"`
	TestedAt     time.Time              `json:"tested_at"`
	Capabilities []string               `json:"capabilities,omitempty"`
}

// Monitoring Types

// IntegrationMetrics contains integration performance metrics
type IntegrationMetrics struct {
	IntegrationID    string                 `json:"integration_id"`
	TimeRange        *TimeRange             `json:"time_range"`
	RequestCount     int64                  `json:"request_count"`
	SuccessCount     int64                  `json:"success_count"`
	ErrorCount       int64                  `json:"error_count"`
	SuccessRate      float64                `json:"success_rate"`
	AverageLatency   time.Duration          `json:"average_latency"`
	P95Latency       time.Duration          `json:"p95_latency"`
	P99Latency       time.Duration          `json:"p99_latency"`
	ThroughputPerSec float64                `json:"throughput_per_sec"`
	ErrorsByType     map[string]int64       `json:"errors_by_type"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
}

// IntegrationHealth represents the health status of an integration
type IntegrationHealth struct {
	IntegrationID string                 `json:"integration_id"`
	Status        HealthStatus           `json:"status"`
	Message       string                 `json:"message"`
	LastCheck     time.Time              `json:"last_check"`
	Uptime        time.Duration          `json:"uptime"`
	Checks        []*HealthCheck         `json:"checks"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
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

// AdapterHealth represents adapter-specific health information
type AdapterHealth struct {
	Status       HealthStatus           `json:"status"`
	Message      string                 `json:"message"`
	Connected    bool                   `json:"connected"`
	LastActivity time.Time              `json:"last_activity"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// AdapterMetrics contains adapter performance metrics
type AdapterMetrics struct {
	OperationCount   map[string]int64       `json:"operation_count"`
	AverageLatency   map[string]time.Duration `json:"average_latency"`
	ErrorCount       map[string]int64       `json:"error_count"`
	LastOperation    time.Time              `json:"last_operation"`
	ConnectionUptime time.Duration          `json:"connection_uptime"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
}

// Logging Types

// IntegrationLog represents an integration log entry
type IntegrationLog struct {
	ID            string                 `json:"id"`
	IntegrationID string                 `json:"integration_id"`
	Level         LogLevel               `json:"level"`
	Message       string                 `json:"message"`
	Operation     string                 `json:"operation,omitempty"`
	Duration      time.Duration          `json:"duration,omitempty"`
	Error         string                 `json:"error,omitempty"`
	Data          map[string]interface{} `json:"data,omitempty"`
	Timestamp     time.Time              `json:"timestamp"`
}

// LogLevel defines the log level
type LogLevel string

const (
	LogLevelDebug LogLevel = "debug"
	LogLevelInfo  LogLevel = "info"
	LogLevelWarn  LogLevel = "warn"
	LogLevelError LogLevel = "error"
)

// Filter Types

// IntegrationFilter contains filters for integration queries
type IntegrationFilter struct {
	Type      string            `json:"type,omitempty"`
	Provider  string            `json:"provider,omitempty"`
	Status    IntegrationStatus `json:"status,omitempty"`
	CreatedBy string            `json:"created_by,omitempty"`
	Search    string            `json:"search,omitempty"`
	Limit     int               `json:"limit,omitempty"`
	Offset    int               `json:"offset,omitempty"`
}

// WebhookFilter contains filters for webhook queries
type WebhookFilter struct {
	IntegrationID string        `json:"integration_id,omitempty"`
	Status        WebhookStatus `json:"status,omitempty"`
	Event         string        `json:"event,omitempty"`
	Search        string        `json:"search,omitempty"`
	Limit         int           `json:"limit,omitempty"`
	Offset        int           `json:"offset,omitempty"`
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

// TimeRange represents a time range for queries
type TimeRange struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}
