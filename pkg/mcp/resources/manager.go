package resources

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/aios/aios/pkg/mcp/protocol"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// DefaultResourceManager implements ResourceManager interface
type DefaultResourceManager struct {
	resources   map[string]MCPResource
	subscribers map[string][]ResourceCallback
	cache       ResourceCache
	validator   ResourceValidator
	metrics     ResourceMetrics
	watcher     ResourceWatcher
	logger      *logrus.Logger
	tracer      trace.Tracer
	mu          sync.RWMutex
}

// ResourceManagerConfig represents resource manager configuration
type ResourceManagerConfig struct {
	EnableCache    bool                   `json:"enable_cache"`
	EnableWatcher  bool                   `json:"enable_watcher"`
	EnableMetrics  bool                   `json:"enable_metrics"`
	CacheTTL       string                 `json:"cache_ttl"`
	MaxResources   int                    `json:"max_resources"`
	AllowedSchemes []string               `json:"allowed_schemes"`
	Metadata       map[string]interface{} `json:"metadata"`
}

// NewResourceManager creates a new resource manager
func NewResourceManager(config *ResourceManagerConfig, logger *logrus.Logger) (ResourceManager, error) {
	if config.MaxResources <= 0 {
		config.MaxResources = 10000 // Default max resources
	}

	manager := &DefaultResourceManager{
		resources:   make(map[string]MCPResource),
		subscribers: make(map[string][]ResourceCallback),
		logger:      logger,
		tracer:      otel.Tracer("mcp.resources.manager"),
	}

	// Initialize optional components
	if config.EnableCache {
		cache, err := NewMemoryResourceCache(logger)
		if err != nil {
			return nil, fmt.Errorf("failed to create resource cache: %w", err)
		}
		manager.cache = cache
	}

	if config.EnableWatcher {
		watcher, err := NewResourceWatcher(logger)
		if err != nil {
			return nil, fmt.Errorf("failed to create resource watcher: %w", err)
		}
		manager.watcher = watcher
	}

	if config.EnableMetrics {
		metrics, err := NewResourceMetrics(logger)
		if err != nil {
			return nil, fmt.Errorf("failed to create resource metrics: %w", err)
		}
		manager.metrics = metrics
	}

	// Create validator
	validator, err := NewResourceValidator(config.AllowedSchemes, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource validator: %w", err)
	}
	manager.validator = validator

	return manager, nil
}

// RegisterResource registers a resource
func (rm *DefaultResourceManager) RegisterResource(resource MCPResource) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	uri := resource.GetURI()
	if uri == "" {
		return fmt.Errorf("resource URI cannot be empty")
	}

	// Validate resource
	if err := rm.validator.ValidateResource(resource); err != nil {
		return fmt.Errorf("resource validation failed: %w", err)
	}

	// Check if resource already exists
	if _, exists := rm.resources[uri]; exists {
		return fmt.Errorf("resource with URI %s already exists", uri)
	}

	rm.resources[uri] = resource

	rm.logger.WithFields(logrus.Fields{
		"uri":      uri,
		"name":     resource.GetName(),
		"category": resource.GetCategory(),
	}).Info("Resource registered")

	// Start watching if resource is watchable
	if resource.IsWatchable() && rm.watcher != nil {
		if err := rm.watcher.Watch(uri); err != nil {
			rm.logger.WithError(err).WithField("uri", uri).Error("Failed to start watching resource")
		}
	}

	return nil
}

// UnregisterResource unregisters a resource
func (rm *DefaultResourceManager) UnregisterResource(uri string) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	resource, exists := rm.resources[uri]
	if !exists {
		return fmt.Errorf("resource not found: %s", uri)
	}

	// Stop watching
	if resource.IsWatchable() && rm.watcher != nil {
		if err := rm.watcher.Unwatch(uri); err != nil {
			rm.logger.WithError(err).WithField("uri", uri).Error("Failed to stop watching resource")
		}
	}

	// Remove from cache
	if rm.cache != nil {
		rm.cache.Delete(uri)
	}

	// Remove subscribers
	delete(rm.subscribers, uri)

	// Remove resource
	delete(rm.resources, uri)

	rm.logger.WithField("uri", uri).Info("Resource unregistered")

	return nil
}

// GetResource retrieves a resource by URI
func (rm *DefaultResourceManager) GetResource(uri string) (MCPResource, error) {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	resource, exists := rm.resources[uri]
	if !exists {
		return nil, fmt.Errorf("resource not found: %s", uri)
	}

	return resource, nil
}

// ListResources returns all registered resources
func (rm *DefaultResourceManager) ListResources(cursor string, limit int) (*protocol.ListResourcesResult, error) {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	// Convert resources to protocol format
	var resources []protocol.Resource
	var uris []string

	for uri := range rm.resources {
		uris = append(uris, uri)
	}

	// Sort for consistent ordering
	sort.Strings(uris)

	// Apply cursor-based pagination
	startIndex := 0
	if cursor != "" {
		for i, uri := range uris {
			if uri == cursor {
				startIndex = i + 1
				break
			}
		}
	}

	// Apply limit
	if limit <= 0 {
		limit = 100 // Default limit
	}

	endIndex := startIndex + limit
	if endIndex > len(uris) {
		endIndex = len(uris)
	}

	// Build resource list
	for i := startIndex; i < endIndex; i++ {
		uri := uris[i]
		resource := rm.resources[uri]

		protocolResource := protocol.Resource{
			URI:         resource.GetURI(),
			Name:        resource.GetName(),
			Description: resource.GetDescription(),
			MimeType:    resource.GetMimeType(),
			Annotations: resource.GetAnnotations(),
		}

		resources = append(resources, protocolResource)
	}

	// Determine next cursor
	var nextCursor string
	if endIndex < len(uris) {
		nextCursor = uris[endIndex-1]
	}

	return &protocol.ListResourcesResult{
		Resources:  resources,
		NextCursor: nextCursor,
	}, nil
}

// ReadResource reads a resource's content
func (rm *DefaultResourceManager) ReadResource(ctx context.Context, uri string) (*protocol.ReadResourceResult, error) {
	ctx, span := rm.tracer.Start(ctx, "resource_manager.read_resource")
	defer span.End()

	span.SetAttributes(
		attribute.String("resource.uri", uri),
	)

	start := time.Now()

	// Check cache first
	if rm.cache != nil {
		if content, found := rm.cache.Get(uri); found {
			if rm.metrics != nil {
				rm.metrics.RecordCacheHit(uri, true)
				rm.metrics.RecordResourceAccess(uri, "read", time.Since(start), true)
			}

			return &protocol.ReadResourceResult{
				Contents: content,
			}, nil
		}
		if rm.metrics != nil {
			rm.metrics.RecordCacheHit(uri, false)
		}
	}

	// Get resource
	resource, err := rm.GetResource(uri)
	if err != nil {
		if rm.metrics != nil {
			rm.metrics.RecordResourceAccess(uri, "read", time.Since(start), false)
		}
		return nil, err
	}

	// Read content
	content, err := resource.ReadContent(ctx)
	if err != nil {
		if rm.metrics != nil {
			rm.metrics.RecordResourceAccess(uri, "read", time.Since(start), false)
		}
		span.RecordError(err)
		return nil, fmt.Errorf("failed to read resource content: %w", err)
	}

	// Cache content
	if rm.cache != nil {
		if err := rm.cache.Set(uri, content, 0); err != nil {
			rm.logger.WithError(err).WithField("uri", uri).Error("Failed to cache resource content")
		}
	}

	// Record metrics
	if rm.metrics != nil {
		var totalSize int64
		for _, c := range content {
			totalSize += int64(len(c.Text) + len(c.Blob))
		}
		rm.metrics.RecordResourceSize(uri, totalSize)
		rm.metrics.RecordResourceAccess(uri, "read", time.Since(start), true)
	}

	rm.logger.WithFields(logrus.Fields{
		"uri":           uri,
		"content_count": len(content),
		"duration":      time.Since(start),
	}).Debug("Resource content read")

	return &protocol.ReadResourceResult{
		Contents: content,
	}, nil
}

// SubscribeResource subscribes to resource changes
func (rm *DefaultResourceManager) SubscribeResource(ctx context.Context, uri string, callback ResourceCallback) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	// Check if resource exists
	resource, exists := rm.resources[uri]
	if !exists {
		return fmt.Errorf("resource not found: %s", uri)
	}

	// Check if resource is watchable
	if !resource.IsWatchable() {
		return fmt.Errorf("resource is not watchable: %s", uri)
	}

	// Add subscriber
	rm.subscribers[uri] = append(rm.subscribers[uri], callback)

	rm.logger.WithField("uri", uri).Debug("Subscribed to resource changes")

	return nil
}

// UnsubscribeResource unsubscribes from resource changes
func (rm *DefaultResourceManager) UnsubscribeResource(ctx context.Context, uri string) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	// Remove all subscribers for this URI
	delete(rm.subscribers, uri)

	rm.logger.WithField("uri", uri).Debug("Unsubscribed from resource changes")

	return nil
}

// ValidateResource validates a resource configuration
func (rm *DefaultResourceManager) ValidateResource(resource MCPResource) error {
	return rm.validator.ValidateResource(resource)
}

// RefreshResource refreshes a resource's content
func (rm *DefaultResourceManager) RefreshResource(ctx context.Context, uri string) error {
	// Remove from cache to force refresh
	if rm.cache != nil {
		rm.cache.Delete(uri)
	}

	// Trigger resource change notification
	rm.notifyResourceChanged(ctx, uri)

	rm.logger.WithField("uri", uri).Debug("Resource refreshed")

	return nil
}

// Helper methods

func (rm *DefaultResourceManager) notifyResourceChanged(ctx context.Context, uri string) {
	rm.mu.RLock()
	callbacks := rm.subscribers[uri]
	resource := rm.resources[uri]
	rm.mu.RUnlock()

	if len(callbacks) == 0 || resource == nil {
		return
	}

	// Read updated content
	content, err := resource.ReadContent(ctx)
	if err != nil {
		rm.logger.WithError(err).WithField("uri", uri).Error("Failed to read resource content for notification")
		return
	}

	// Notify all subscribers
	for _, callback := range callbacks {
		go func(cb ResourceCallback) {
			if err := cb.OnResourceChanged(ctx, uri, content); err != nil {
				rm.logger.WithError(err).WithField("uri", uri).Error("Resource change callback failed")
			}
		}(callback)
	}
}

// defaultResourceCallback implements ResourceCallback for internal use
type defaultResourceCallback struct {
	manager *DefaultResourceManager
}

func (cb *defaultResourceCallback) OnResourceChanged(ctx context.Context, uri string, content []protocol.ResourceContent) error {
	// Update cache
	if cb.manager.cache != nil {
		if err := cb.manager.cache.Set(uri, content, 0); err != nil {
			cb.manager.logger.WithError(err).WithField("uri", uri).Error("Failed to update cache on resource change")
		}
	}

	// Notify subscribers
	cb.manager.notifyResourceChanged(ctx, uri)

	return nil
}

func (cb *defaultResourceCallback) OnResourceDeleted(ctx context.Context, uri string) error {
	// Remove from cache
	if cb.manager.cache != nil {
		cb.manager.cache.Delete(uri)
	}

	// Notify subscribers
	cb.manager.mu.RLock()
	callbacks := cb.manager.subscribers[uri]
	cb.manager.mu.RUnlock()

	for _, callback := range callbacks {
		go func(callback ResourceCallback) {
			if err := callback.OnResourceDeleted(ctx, uri); err != nil {
				cb.manager.logger.WithError(err).WithField("uri", uri).Error("Resource deletion callback failed")
			}
		}(callback)
	}

	return nil
}

func (cb *defaultResourceCallback) OnResourceError(ctx context.Context, uri string, err error) error {
	cb.manager.logger.WithError(err).WithField("uri", uri).Error("Resource error occurred")

	// Notify subscribers
	cb.manager.mu.RLock()
	callbacks := cb.manager.subscribers[uri]
	cb.manager.mu.RUnlock()

	for _, callback := range callbacks {
		go func(callback ResourceCallback) {
			if err := callback.OnResourceError(ctx, uri, err); err != nil {
				cb.manager.logger.WithError(err).WithField("uri", uri).Error("Resource error callback failed")
			}
		}(callback)
	}

	return nil
}
