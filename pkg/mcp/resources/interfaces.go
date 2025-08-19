package resources

import (
	"context"
	"time"

	"github.com/aios/aios/pkg/mcp/protocol"
)

// ResourceManager manages MCP resources
type ResourceManager interface {
	// RegisterResource registers a resource
	RegisterResource(resource MCPResource) error

	// UnregisterResource unregisters a resource
	UnregisterResource(uri string) error

	// GetResource retrieves a resource by URI
	GetResource(uri string) (MCPResource, error)

	// ListResources returns all registered resources
	ListResources(cursor string, limit int) (*protocol.ListResourcesResult, error)

	// ReadResource reads a resource's content
	ReadResource(ctx context.Context, uri string) (*protocol.ReadResourceResult, error)

	// SubscribeResource subscribes to resource changes
	SubscribeResource(ctx context.Context, uri string, callback ResourceCallback) error

	// UnsubscribeResource unsubscribes from resource changes
	UnsubscribeResource(ctx context.Context, uri string) error

	// ValidateResource validates a resource configuration
	ValidateResource(resource MCPResource) error

	// RefreshResource refreshes a resource's content
	RefreshResource(ctx context.Context, uri string) error
}

// MCPResource represents an MCP resource implementation
type MCPResource interface {
	// GetURI returns the resource URI
	GetURI() string

	// GetName returns the resource name
	GetName() string

	// GetDescription returns the resource description
	GetDescription() string

	// GetMimeType returns the resource MIME type
	GetMimeType() string

	// GetAnnotations returns the resource annotations
	GetAnnotations() map[string]interface{}

	// ReadContent reads the resource content
	ReadContent(ctx context.Context) ([]protocol.ResourceContent, error)

	// GetLastModified returns the last modified time
	GetLastModified() time.Time

	// GetSize returns the resource size in bytes
	GetSize() int64

	// IsWatchable returns whether the resource supports watching for changes
	IsWatchable() bool

	// Watch watches for resource changes
	Watch(ctx context.Context, callback ResourceCallback) error

	// StopWatch stops watching for changes
	StopWatch() error

	// Validate validates the resource
	Validate() error

	// GetCategory returns the resource category
	GetCategory() string

	// GetTags returns the resource tags
	GetTags() []string

	// GetMetadata returns additional metadata
	GetMetadata() map[string]interface{}
}

// ResourceCallback defines callbacks for resource events
type ResourceCallback interface {
	// OnResourceChanged is called when a resource changes
	OnResourceChanged(ctx context.Context, uri string, content []protocol.ResourceContent) error

	// OnResourceDeleted is called when a resource is deleted
	OnResourceDeleted(ctx context.Context, uri string) error

	// OnResourceError is called when a resource error occurs
	OnResourceError(ctx context.Context, uri string, err error) error
}

// FileResource provides file-based resources
type FileResource interface {
	MCPResource

	// GetFilePath returns the file path
	GetFilePath() string

	// IsDirectory returns whether this is a directory resource
	IsDirectory() bool

	// ListFiles lists files in a directory (if applicable)
	ListFiles(ctx context.Context) ([]string, error)

	// GetFileInfo gets file information
	GetFileInfo(ctx context.Context) (*FileInfo, error)
}

// DatabaseResource provides database-based resources
type DatabaseResource interface {
	MCPResource

	// GetConnectionString returns the database connection string
	GetConnectionString() string

	// GetQuery returns the SQL query
	GetQuery() string

	// GetParameters returns query parameters
	GetParameters() map[string]interface{}

	// ExecuteQuery executes the query and returns results
	ExecuteQuery(ctx context.Context) ([]map[string]interface{}, error)

	// GetSchema returns the result schema
	GetSchema() map[string]interface{}
}

// APIResource provides API-based resources
type APIResource interface {
	MCPResource

	// GetEndpoint returns the API endpoint
	GetEndpoint() string

	// GetMethod returns the HTTP method
	GetMethod() string

	// GetHeaders returns HTTP headers
	GetHeaders() map[string]string

	// GetParameters returns request parameters
	GetParameters() map[string]interface{}

	// MakeRequest makes the API request
	MakeRequest(ctx context.Context) (*APIResponse, error)

	// GetAuthConfig returns authentication configuration
	GetAuthConfig() *AuthConfig
}

// TemplateResource provides template-based resources
type TemplateResource interface {
	MCPResource

	// GetTemplate returns the template content
	GetTemplate() string

	// GetTemplateEngine returns the template engine type
	GetTemplateEngine() string

	// GetVariables returns template variables
	GetVariables() map[string]interface{}

	// RenderTemplate renders the template with variables
	RenderTemplate(ctx context.Context, variables map[string]interface{}) (string, error)

	// ValidateTemplate validates the template syntax
	ValidateTemplate() error
}

// Data structures

// FileInfo represents file information
type FileInfo struct {
	Path        string                 `json:"path"`
	Name        string                 `json:"name"`
	Size        int64                  `json:"size"`
	ModTime     time.Time              `json:"mod_time"`
	IsDirectory bool                   `json:"is_directory"`
	Permissions string                 `json:"permissions"`
	MimeType    string                 `json:"mime_type"`
	Encoding    string                 `json:"encoding"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// APIResponse represents an API response
type APIResponse struct {
	StatusCode int                    `json:"status_code"`
	Headers    map[string]string      `json:"headers"`
	Body       string                 `json:"body"`
	Data       interface{}            `json:"data"`
	Duration   time.Duration          `json:"duration"`
	Metadata   map[string]interface{} `json:"metadata"`
}

// AuthConfig represents authentication configuration
type AuthConfig struct {
	Type         string                 `json:"type"` // "bearer", "basic", "api_key", "oauth2"
	Token        string                 `json:"token,omitempty"`
	Username     string                 `json:"username,omitempty"`
	Password     string                 `json:"password,omitempty"`
	APIKey       string                 `json:"api_key,omitempty"`
	APIKeyHeader string                 `json:"api_key_header,omitempty"`
	OAuth2Config *OAuth2Config          `json:"oauth2_config,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// OAuth2Config represents OAuth2 configuration
type OAuth2Config struct {
	ClientID     string   `json:"client_id"`
	ClientSecret string   `json:"client_secret"`
	TokenURL     string   `json:"token_url"`
	Scopes       []string `json:"scopes"`
	RedirectURL  string   `json:"redirect_url,omitempty"`
}

// ResourceConfig represents resource configuration
type ResourceConfig struct {
	URI         string                 `json:"uri"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Type        string                 `json:"type"`
	MimeType    string                 `json:"mime_type"`
	Category    string                 `json:"category"`
	Tags        []string               `json:"tags"`
	Watchable   bool                   `json:"watchable"`
	Config      map[string]interface{} `json:"config"`
	Annotations map[string]interface{} `json:"annotations"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// ResourceRegistry manages resource registration and discovery
type ResourceRegistry interface {
	// RegisterResourceType registers a resource type
	RegisterResourceType(resourceType string, factory ResourceFactory) error

	// UnregisterResourceType unregisters a resource type
	UnregisterResourceType(resourceType string) error

	// CreateResource creates a resource from configuration
	CreateResource(config *ResourceConfig) (MCPResource, error)

	// GetResourceTypes returns all registered resource types
	GetResourceTypes() []string

	// ValidateResourceConfig validates a resource configuration
	ValidateResourceConfig(config *ResourceConfig) error
}

// ResourceFactory creates resources of a specific type
type ResourceFactory interface {
	// CreateResource creates a resource from configuration
	CreateResource(config *ResourceConfig) (MCPResource, error)

	// GetResourceType returns the resource type this factory creates
	GetResourceType() string

	// ValidateConfig validates the configuration for this resource type
	ValidateConfig(config map[string]interface{}) error

	// GetConfigSchema returns the configuration schema
	GetConfigSchema() map[string]interface{}
}

// ResourceWatcher watches for resource changes
type ResourceWatcher interface {
	// Watch starts watching a resource URI
	Watch(uri string) error

	// Unwatch stops watching a resource URI
	Unwatch(uri string) error

	// AddCallback adds a callback for resource changes
	AddCallback(callback WatchCallback) string

	// RemoveCallback removes a callback
	RemoveCallback(callbackID string)

	// Start starts the watcher
	Start(ctx context.Context) error

	// Stop stops the watcher
	Stop() error

	// IsWatching returns whether a resource is being watched
	IsWatching(uri string) bool

	// GetWatchedResources returns all watched resources
	GetWatchedResources() []string
}

// ResourceCache caches resource content for performance
type ResourceCache interface {
	// Get retrieves cached content
	Get(uri string) ([]protocol.ResourceContent, bool)

	// Set stores content in cache
	Set(uri string, content []protocol.ResourceContent, ttl time.Duration) error

	// Delete removes content from cache
	Delete(uri string) error

	// Clear clears all cached content
	Clear() error

	// GetStats returns cache statistics
	GetStats() map[string]interface{}

	// SetTTL sets the default TTL for cached content
	SetTTL(ttl time.Duration)
}

// ResourceValidator validates resources and their content
type ResourceValidator interface {
	// ValidateResource validates a resource
	ValidateResource(resource MCPResource) error

	// ValidateContent validates resource content
	ValidateContent(content []protocol.ResourceContent) error

	// ValidateURI validates a resource URI
	ValidateURI(uri string) error

	// GetValidationRules returns current validation rules
	GetValidationRules() map[string]interface{}

	// SetValidationRules sets custom validation rules
	SetValidationRules(rules map[string]interface{}) error
}

// ResourceMetrics collects metrics about resource usage
type ResourceMetrics interface {
	// RecordResourceAccess records resource access
	RecordResourceAccess(uri string, operation string, duration time.Duration, success bool)

	// RecordResourceSize records resource size
	RecordResourceSize(uri string, size int64)

	// RecordCacheHit records cache hit/miss
	RecordCacheHit(uri string, hit bool)

	// GetResourceStats returns statistics for a resource
	GetResourceStats(uri string) map[string]interface{}

	// GetOverallStats returns overall resource statistics
	GetOverallStats() map[string]interface{}

	// GetTopResources returns the most accessed resources
	GetTopResources(limit int) []string
}

// ResourceBuilder provides a fluent interface for building resources
type ResourceBuilder interface {
	// WithURI sets the resource URI
	WithURI(uri string) ResourceBuilder

	// WithName sets the resource name
	WithName(name string) ResourceBuilder

	// WithDescription sets the resource description
	WithDescription(description string) ResourceBuilder

	// WithType sets the resource type
	WithType(resourceType string) ResourceBuilder

	// WithMimeType sets the MIME type
	WithMimeType(mimeType string) ResourceBuilder

	// WithCategory sets the category
	WithCategory(category string) ResourceBuilder

	// WithTags adds tags
	WithTags(tags ...string) ResourceBuilder

	// WithWatchable sets whether the resource is watchable
	WithWatchable(watchable bool) ResourceBuilder

	// WithConfig sets configuration
	WithConfig(key string, value interface{}) ResourceBuilder

	// WithAnnotation adds an annotation
	WithAnnotation(key string, value interface{}) ResourceBuilder

	// WithMetadata adds metadata
	WithMetadata(key string, value interface{}) ResourceBuilder

	// Build builds the resource
	Build() (MCPResource, error)
}
