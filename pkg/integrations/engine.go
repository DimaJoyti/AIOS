package integrations

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// DefaultIntegrationEngine implements the IntegrationEngine interface
type DefaultIntegrationEngine struct {
	logger *logrus.Logger
	tracer trace.Tracer

	// Storage
	integrations map[string]*Integration
	webhooks     map[string]*Webhook
	logs         []*IntegrationLog

	// Adapters
	adapters map[string]IntegrationAdapter

	// Event handling
	eventHandlers map[string][]EventHandler

	// Metrics and monitoring
	metrics map[string]*IntegrationMetrics
	health  map[string]*IntegrationHealth

	// Indexes
	integrationsByType     map[string][]string
	integrationsByProvider map[string][]string
	webhooksByIntegration  map[string][]string

	mu sync.RWMutex
}

// IntegrationEngineConfig represents configuration for the integration engine
type IntegrationEngineConfig struct {
	MaxIntegrations     int           `json:"max_integrations"`
	MaxWebhooks         int           `json:"max_webhooks"`
	DefaultTimeout      time.Duration `json:"default_timeout"`
	LogRetention        time.Duration `json:"log_retention"`
	MetricsRetention    time.Duration `json:"metrics_retention"`
	HealthCheckInterval time.Duration `json:"health_check_interval"`
	EnableMetrics       bool          `json:"enable_metrics"`
	EnableHealthChecks  bool          `json:"enable_health_checks"`
}

// NewDefaultIntegrationEngine creates a new integration engine
func NewDefaultIntegrationEngine(config *IntegrationEngineConfig, logger *logrus.Logger) IntegrationEngine {
	if config == nil {
		config = &IntegrationEngineConfig{
			MaxIntegrations:     1000,
			MaxWebhooks:         5000,
			DefaultTimeout:      30 * time.Second,
			LogRetention:        7 * 24 * time.Hour,
			MetricsRetention:    30 * 24 * time.Hour,
			HealthCheckInterval: 5 * time.Minute,
			EnableMetrics:       true,
			EnableHealthChecks:  true,
		}
	}

	engine := &DefaultIntegrationEngine{
		logger:                 logger,
		tracer:                 otel.Tracer("integrations.engine"),
		integrations:           make(map[string]*Integration),
		webhooks:               make(map[string]*Webhook),
		logs:                   make([]*IntegrationLog, 0),
		adapters:               make(map[string]IntegrationAdapter),
		eventHandlers:          make(map[string][]EventHandler),
		metrics:                make(map[string]*IntegrationMetrics),
		health:                 make(map[string]*IntegrationHealth),
		integrationsByType:     make(map[string][]string),
		integrationsByProvider: make(map[string][]string),
		webhooksByIntegration:  make(map[string][]string),
	}

	// Start background tasks
	if config.EnableHealthChecks {
		go engine.startHealthCheckLoop(config.HealthCheckInterval)
	}

	return engine
}

// Integration Management

// CreateIntegration creates a new integration
func (ie *DefaultIntegrationEngine) CreateIntegration(integration *Integration) (*Integration, error) {
	_, span := ie.tracer.Start(context.Background(), "integrations.create_integration")
	defer span.End()

	if integration.ID == "" {
		integration.ID = uuid.New().String()
	}

	// Set timestamps
	now := time.Now()
	integration.CreatedAt = now
	integration.UpdatedAt = now

	// Set default status
	if integration.Status == "" {
		integration.Status = IntegrationStatusConfiguring
	}

	// Set default settings
	if integration.Settings == nil {
		integration.Settings = &IntegrationSettings{
			AutoSync:       true,
			SyncInterval:   15 * time.Minute,
			EnableWebhooks: true,
			EnableEvents:   true,
			LogLevel:       "info",
			NotifyOnError:  true,
		}
	}

	// Set default config
	if integration.Config == nil {
		integration.Config = &IntegrationConfig{
			Timeout: 30 * time.Second,
			RetryPolicy: &RetryPolicy{
				MaxRetries:    3,
				InitialDelay:  1 * time.Second,
				MaxDelay:      30 * time.Second,
				BackoffFactor: 2.0,
			},
			RateLimit: &RateLimit{
				RequestsPerSecond: 10,
				BurstSize:         20,
				WindowSize:        time.Minute,
			},
		}
	}

	// Validate integration
	if err := ie.validateIntegration(integration); err != nil {
		return nil, fmt.Errorf("integration validation failed: %w", err)
	}

	ie.mu.Lock()
	ie.integrations[integration.ID] = integration
	ie.integrationsByType[integration.Type] = append(ie.integrationsByType[integration.Type], integration.ID)
	ie.integrationsByProvider[integration.Provider] = append(ie.integrationsByProvider[integration.Provider], integration.ID)

	// Initialize health status
	ie.health[integration.ID] = &IntegrationHealth{
		IntegrationID: integration.ID,
		Status:        HealthStatusUnknown,
		Message:       "Integration created",
		LastCheck:     now,
		Uptime:        0,
		Checks:        []*HealthCheck{},
	}
	ie.mu.Unlock()

	// Log creation
	ie.logIntegrationEvent(integration.ID, LogLevelInfo, "integration_created",
		fmt.Sprintf("Integration '%s' created", integration.Name), nil, nil)

	span.SetAttributes(
		attribute.String("integration.id", integration.ID),
		attribute.String("integration.name", integration.Name),
		attribute.String("integration.type", integration.Type),
		attribute.String("integration.provider", integration.Provider),
	)

	ie.logger.WithFields(logrus.Fields{
		"integration_id":   integration.ID,
		"integration_name": integration.Name,
		"integration_type": integration.Type,
		"provider":         integration.Provider,
	}).Info("Integration created successfully")

	return integration, nil
}

// GetIntegration retrieves an integration by ID
func (ie *DefaultIntegrationEngine) GetIntegration(integrationID string) (*Integration, error) {
	_, span := ie.tracer.Start(context.Background(), "integrations.get_integration")
	defer span.End()

	ie.mu.RLock()
	integration, exists := ie.integrations[integrationID]
	ie.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("integration not found: %s", integrationID)
	}

	span.SetAttributes(attribute.String("integration.id", integrationID))

	return integration, nil
}

// UpdateIntegration updates an existing integration
func (ie *DefaultIntegrationEngine) UpdateIntegration(integration *Integration) (*Integration, error) {
	_, span := ie.tracer.Start(context.Background(), "integrations.update_integration")
	defer span.End()

	ie.mu.Lock()
	existing, exists := ie.integrations[integration.ID]
	if !exists {
		ie.mu.Unlock()
		return nil, fmt.Errorf("integration not found: %s", integration.ID)
	}

	// Preserve creation info
	integration.CreatedBy = existing.CreatedBy
	integration.CreatedAt = existing.CreatedAt
	integration.UpdatedAt = time.Now()

	// Validate integration
	if err := ie.validateIntegration(integration); err != nil {
		ie.mu.Unlock()
		return nil, fmt.Errorf("integration validation failed: %w", err)
	}

	ie.integrations[integration.ID] = integration
	ie.mu.Unlock()

	// Log update
	ie.logIntegrationEvent(integration.ID, LogLevelInfo, "integration_updated",
		fmt.Sprintf("Integration '%s' updated", integration.Name), nil, nil)

	span.SetAttributes(attribute.String("integration.id", integration.ID))

	ie.logger.WithField("integration_id", integration.ID).Info("Integration updated successfully")

	return integration, nil
}

// DeleteIntegration deletes an integration
func (ie *DefaultIntegrationEngine) DeleteIntegration(integrationID string) error {
	_, span := ie.tracer.Start(context.Background(), "integrations.delete_integration")
	defer span.End()

	ie.mu.Lock()
	integration, exists := ie.integrations[integrationID]
	if !exists {
		ie.mu.Unlock()
		return fmt.Errorf("integration not found: %s", integrationID)
	}

	// Remove from indexes
	ie.removeFromIndex(ie.integrationsByType[integration.Type], integrationID)
	ie.removeFromIndex(ie.integrationsByProvider[integration.Provider], integrationID)

	// Delete associated webhooks
	if webhookIDs, exists := ie.webhooksByIntegration[integrationID]; exists {
		for _, webhookID := range webhookIDs {
			delete(ie.webhooks, webhookID)
		}
		delete(ie.webhooksByIntegration, integrationID)
	}

	// Clean up
	delete(ie.integrations, integrationID)
	delete(ie.health, integrationID)
	delete(ie.metrics, integrationID)
	ie.mu.Unlock()

	// Disconnect adapter if connected
	if adapter, err := ie.GetAdapter(integration.Type); err == nil {
		if adapter.IsConnected() {
			adapter.Disconnect(context.Background())
		}
	}

	// Log deletion
	ie.logIntegrationEvent(integrationID, LogLevelInfo, "integration_deleted",
		fmt.Sprintf("Integration '%s' deleted", integration.Name), nil, nil)

	span.SetAttributes(attribute.String("integration.id", integrationID))

	ie.logger.WithField("integration_id", integrationID).Info("Integration deleted successfully")

	return nil
}

// ListIntegrations lists integrations with filtering
func (ie *DefaultIntegrationEngine) ListIntegrations(filter *IntegrationFilter) ([]*Integration, error) {
	_, span := ie.tracer.Start(context.Background(), "integrations.list_integrations")
	defer span.End()

	ie.mu.RLock()
	var integrations []*Integration
	for _, integration := range ie.integrations {
		if ie.matchesIntegrationFilter(integration, filter) {
			integrations = append(integrations, integration)
		}
	}
	ie.mu.RUnlock()

	// Sort integrations by creation date (newest first)
	sort.Slice(integrations, func(i, j int) bool {
		return integrations[i].CreatedAt.After(integrations[j].CreatedAt)
	})

	// Apply pagination
	if filter != nil {
		if filter.Offset > 0 && filter.Offset < len(integrations) {
			integrations = integrations[filter.Offset:]
		}
		if filter.Limit > 0 && filter.Limit < len(integrations) {
			integrations = integrations[:filter.Limit]
		}
	}

	span.SetAttributes(attribute.Int("integrations.count", len(integrations)))

	return integrations, nil
}

// Integration Operations

// EnableIntegration enables an integration
func (ie *DefaultIntegrationEngine) EnableIntegration(integrationID string) error {
	_, span := ie.tracer.Start(context.Background(), "integrations.enable_integration")
	defer span.End()

	ie.mu.Lock()
	integration, exists := ie.integrations[integrationID]
	if !exists {
		ie.mu.Unlock()
		return fmt.Errorf("integration not found: %s", integrationID)
	}

	integration.Status = IntegrationStatusActive
	integration.UpdatedAt = time.Now()
	ie.mu.Unlock()

	// Connect adapter
	if adapter, err := ie.GetAdapter(integration.Type); err == nil {
		ctx, cancel := context.WithTimeout(context.Background(), integration.Config.Timeout)
		defer cancel()

		if err := adapter.Connect(ctx, integration.Config.Custom); err != nil {
			integration.Status = IntegrationStatusError
			return fmt.Errorf("failed to connect adapter: %w", err)
		}
	}

	// Log enablement
	ie.logIntegrationEvent(integrationID, LogLevelInfo, "integration_enabled",
		fmt.Sprintf("Integration '%s' enabled", integration.Name), nil, nil)

	span.SetAttributes(attribute.String("integration.id", integrationID))

	ie.logger.WithField("integration_id", integrationID).Info("Integration enabled successfully")

	return nil
}

// DisableIntegration disables an integration
func (ie *DefaultIntegrationEngine) DisableIntegration(integrationID string) error {
	_, span := ie.tracer.Start(context.Background(), "integrations.disable_integration")
	defer span.End()

	ie.mu.Lock()
	integration, exists := ie.integrations[integrationID]
	if !exists {
		ie.mu.Unlock()
		return fmt.Errorf("integration not found: %s", integrationID)
	}

	integration.Status = IntegrationStatusInactive
	integration.UpdatedAt = time.Now()
	ie.mu.Unlock()

	// Disconnect adapter
	if adapter, err := ie.GetAdapter(integration.Type); err == nil {
		if adapter.IsConnected() {
			adapter.Disconnect(context.Background())
		}
	}

	// Log disablement
	ie.logIntegrationEvent(integrationID, LogLevelInfo, "integration_disabled",
		fmt.Sprintf("Integration '%s' disabled", integration.Name), nil, nil)

	span.SetAttributes(attribute.String("integration.id", integrationID))

	ie.logger.WithField("integration_id", integrationID).Info("Integration disabled successfully")

	return nil
}

// TestIntegration tests an integration connection
func (ie *DefaultIntegrationEngine) TestIntegration(integrationID string) (*IntegrationTestResult, error) {
	_, span := ie.tracer.Start(context.Background(), "integrations.test_integration")
	defer span.End()

	start := time.Now()

	integration, err := ie.GetIntegration(integrationID)
	if err != nil {
		return nil, err
	}

	adapter, err := ie.GetAdapter(integration.Type)
	if err != nil {
		return &IntegrationTestResult{
			Success:  false,
			Message:  fmt.Sprintf("Adapter not found: %s", integration.Type),
			Duration: time.Since(start),
			TestedAt: time.Now(),
		}, nil
	}

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), integration.Config.Timeout)
	defer cancel()

	err = adapter.TestConnection(ctx)
	result := &IntegrationTestResult{
		Success:      err == nil,
		Duration:     time.Since(start),
		TestedAt:     time.Now(),
		Capabilities: adapter.GetSupportedOperations(),
	}

	if err != nil {
		result.Message = fmt.Sprintf("Connection test failed: %v", err)
		ie.logIntegrationEvent(integrationID, LogLevelError, "test_failed", result.Message, nil, err)
	} else {
		result.Message = "Connection test successful"
		ie.logIntegrationEvent(integrationID, LogLevelInfo, "test_successful", result.Message, nil, nil)
	}

	span.SetAttributes(
		attribute.String("integration.id", integrationID),
		attribute.Bool("test.success", result.Success),
		attribute.Int64("test.duration_ms", result.Duration.Milliseconds()),
	)

	return result, nil
}

// RefreshIntegration refreshes an integration's credentials or configuration
func (ie *DefaultIntegrationEngine) RefreshIntegration(integrationID string) error {
	_, span := ie.tracer.Start(context.Background(), "integrations.refresh_integration")
	defer span.End()

	integration, err := ie.GetIntegration(integrationID)
	if err != nil {
		return err
	}

	// Update last sync time
	now := time.Now()
	integration.LastSyncAt = &now
	integration.UpdatedAt = now

	ie.mu.Lock()
	ie.integrations[integrationID] = integration
	ie.mu.Unlock()

	// Log refresh
	ie.logIntegrationEvent(integrationID, LogLevelInfo, "integration_refreshed",
		fmt.Sprintf("Integration '%s' refreshed", integration.Name), nil, nil)

	span.SetAttributes(attribute.String("integration.id", integrationID))

	ie.logger.WithField("integration_id", integrationID).Info("Integration refreshed successfully")

	return nil
}

// Adapter Management

// RegisterAdapter registers an integration adapter
func (ie *DefaultIntegrationEngine) RegisterAdapter(adapter IntegrationAdapter) error {
	_, span := ie.tracer.Start(context.Background(), "integrations.register_adapter")
	defer span.End()

	adapterType := adapter.GetType()

	ie.mu.Lock()
	ie.adapters[adapterType] = adapter
	ie.mu.Unlock()

	span.SetAttributes(
		attribute.String("adapter.type", adapterType),
		attribute.String("adapter.name", adapter.GetName()),
		attribute.String("adapter.version", adapter.GetVersion()),
	)

	ie.logger.WithFields(logrus.Fields{
		"adapter_type":    adapterType,
		"adapter_name":    adapter.GetName(),
		"adapter_version": adapter.GetVersion(),
	}).Info("Adapter registered successfully")

	return nil
}

// GetAdapter retrieves an adapter by type
func (ie *DefaultIntegrationEngine) GetAdapter(adapterType string) (IntegrationAdapter, error) {
	_, span := ie.tracer.Start(context.Background(), "integrations.get_adapter")
	defer span.End()

	ie.mu.RLock()
	adapter, exists := ie.adapters[adapterType]
	ie.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("adapter not found: %s", adapterType)
	}

	span.SetAttributes(attribute.String("adapter.type", adapterType))

	return adapter, nil
}

// ListAdapters lists all registered adapters
func (ie *DefaultIntegrationEngine) ListAdapters() []string {
	_, span := ie.tracer.Start(context.Background(), "integrations.list_adapters")
	defer span.End()

	ie.mu.RLock()
	var adapterTypes []string
	for adapterType := range ie.adapters {
		adapterTypes = append(adapterTypes, adapterType)
	}
	ie.mu.RUnlock()

	sort.Strings(adapterTypes)

	span.SetAttributes(attribute.Int("adapters.count", len(adapterTypes)))

	return adapterTypes
}

// Event Management

// PublishEvent publishes an integration event
func (ie *DefaultIntegrationEngine) PublishEvent(event *IntegrationEvent) error {
	_, span := ie.tracer.Start(context.Background(), "integrations.publish_event")
	defer span.End()

	if event.ID == "" {
		event.ID = uuid.New().String()
	}

	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	// Notify event handlers
	ie.mu.RLock()
	handlers, exists := ie.eventHandlers[event.Type]
	ie.mu.RUnlock()

	if exists {
		for _, handler := range handlers {
			go func(h EventHandler) {
				ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
				defer cancel()

				if err := h.HandleEvent(ctx, event); err != nil {
					ie.logger.WithError(err).WithField("event_type", event.Type).Warn("Event handler failed")
				}
			}(handler)
		}
	}

	span.SetAttributes(
		attribute.String("event.id", event.ID),
		attribute.String("event.type", event.Type),
		attribute.String("event.source", event.Source),
	)

	ie.logger.WithFields(logrus.Fields{
		"event_id":   event.ID,
		"event_type": event.Type,
		"source":     event.Source,
	}).Info("Event published successfully")

	return nil
}

// SubscribeToEvents subscribes to events of a specific type
func (ie *DefaultIntegrationEngine) SubscribeToEvents(eventType string, handler EventHandler) error {
	_, span := ie.tracer.Start(context.Background(), "integrations.subscribe_to_events")
	defer span.End()

	ie.mu.Lock()
	ie.eventHandlers[eventType] = append(ie.eventHandlers[eventType], handler)
	ie.mu.Unlock()

	span.SetAttributes(attribute.String("event.type", eventType))

	ie.logger.WithField("event_type", eventType).Info("Subscribed to events")

	return nil
}

// UnsubscribeFromEvents unsubscribes from events of a specific type
func (ie *DefaultIntegrationEngine) UnsubscribeFromEvents(eventType string, handler EventHandler) error {
	_, span := ie.tracer.Start(context.Background(), "integrations.unsubscribe_from_events")
	defer span.End()

	ie.mu.Lock()
	handlers := ie.eventHandlers[eventType]
	for i, h := range handlers {
		if h == handler {
			ie.eventHandlers[eventType] = append(handlers[:i], handlers[i+1:]...)
			break
		}
	}
	ie.mu.Unlock()

	span.SetAttributes(attribute.String("event.type", eventType))

	ie.logger.WithField("event_type", eventType).Info("Unsubscribed from events")

	return nil
}

// Configuration Management

// GetConfiguration retrieves integration configuration
func (ie *DefaultIntegrationEngine) GetConfiguration(integrationID string) (*IntegrationConfig, error) {
	_, span := ie.tracer.Start(context.Background(), "integrations.get_configuration")
	defer span.End()

	integration, err := ie.GetIntegration(integrationID)
	if err != nil {
		return nil, err
	}

	span.SetAttributes(attribute.String("integration.id", integrationID))

	return integration.Config, nil
}

// UpdateConfiguration updates integration configuration
func (ie *DefaultIntegrationEngine) UpdateConfiguration(integrationID string, config *IntegrationConfig) error {
	_, span := ie.tracer.Start(context.Background(), "integrations.update_configuration")
	defer span.End()

	ie.mu.Lock()
	integration, exists := ie.integrations[integrationID]
	if !exists {
		ie.mu.Unlock()
		return fmt.Errorf("integration not found: %s", integrationID)
	}

	// Validate configuration
	if err := ie.ValidateConfiguration(config); err != nil {
		ie.mu.Unlock()
		return fmt.Errorf("configuration validation failed: %w", err)
	}

	integration.Config = config
	integration.UpdatedAt = time.Now()
	ie.mu.Unlock()

	// Log configuration update
	ie.logIntegrationEvent(integrationID, LogLevelInfo, "configuration_updated",
		"Integration configuration updated", nil, nil)

	span.SetAttributes(attribute.String("integration.id", integrationID))

	ie.logger.WithField("integration_id", integrationID).Info("Configuration updated successfully")

	return nil
}

// ValidateConfiguration validates integration configuration
func (ie *DefaultIntegrationEngine) ValidateConfiguration(config *IntegrationConfig) error {
	if config.Timeout <= 0 {
		config.Timeout = 30 * time.Second
	}

	if config.RetryPolicy != nil {
		if config.RetryPolicy.MaxRetries < 0 {
			return fmt.Errorf("max retries cannot be negative")
		}
		if config.RetryPolicy.BackoffFactor <= 0 {
			config.RetryPolicy.BackoffFactor = 2.0
		}
	}

	if config.RateLimit != nil {
		if config.RateLimit.RequestsPerSecond <= 0 {
			return fmt.Errorf("requests per second must be positive")
		}
		if config.RateLimit.BurstSize <= 0 {
			config.RateLimit.BurstSize = config.RateLimit.RequestsPerSecond
		}
	}

	return nil
}

// Webhook Management (placeholder implementations)

// CreateWebhook creates a new webhook
func (ie *DefaultIntegrationEngine) CreateWebhook(webhook *Webhook) (*Webhook, error) {
	// Implementation would be added here
	return webhook, nil
}

// GetWebhook retrieves a webhook by ID
func (ie *DefaultIntegrationEngine) GetWebhook(webhookID string) (*Webhook, error) {
	// Implementation would be added here
	return nil, fmt.Errorf("webhook not found: %s", webhookID)
}

// UpdateWebhook updates an existing webhook
func (ie *DefaultIntegrationEngine) UpdateWebhook(webhook *Webhook) (*Webhook, error) {
	// Implementation would be added here
	return webhook, nil
}

// DeleteWebhook deletes a webhook
func (ie *DefaultIntegrationEngine) DeleteWebhook(webhookID string) error {
	// Implementation would be added here
	return nil
}

// ListWebhooks lists webhooks with filtering
func (ie *DefaultIntegrationEngine) ListWebhooks(filter *WebhookFilter) ([]*Webhook, error) {
	// Implementation would be added here
	return []*Webhook{}, nil
}

// ProcessWebhook processes an incoming webhook payload
func (ie *DefaultIntegrationEngine) ProcessWebhook(webhookID string, payload []byte, headers map[string]string) error {
	// Implementation would be added here
	return nil
}

// Monitoring and Analytics

// GetIntegrationMetrics retrieves integration metrics
func (ie *DefaultIntegrationEngine) GetIntegrationMetrics(integrationID string, timeRange *TimeRange) (*IntegrationMetrics, error) {
	// Return mock metrics for demo
	return &IntegrationMetrics{
		IntegrationID:    integrationID,
		TimeRange:        timeRange,
		RequestCount:     1250,
		SuccessCount:     1200,
		ErrorCount:       50,
		SuccessRate:      0.96,
		AverageLatency:   150 * time.Millisecond,
		P95Latency:       300 * time.Millisecond,
		P99Latency:       500 * time.Millisecond,
		ThroughputPerSec: 2.5,
		ErrorsByType: map[string]int64{
			"timeout":      20,
			"rate_limit":   15,
			"auth_error":   10,
			"server_error": 5,
		},
	}, nil
}

// GetIntegrationHealth retrieves integration health status
func (ie *DefaultIntegrationEngine) GetIntegrationHealth(integrationID string) (*IntegrationHealth, error) {
	ie.mu.RLock()
	health, exists := ie.health[integrationID]
	ie.mu.RUnlock()

	if !exists {
		return &IntegrationHealth{
			IntegrationID: integrationID,
			Status:        HealthStatusUnknown,
			Message:       "Health status not available",
			LastCheck:     time.Now(),
			Uptime:        0,
			Checks:        []*HealthCheck{},
		}, nil
	}

	return health, nil
}

// GetIntegrationLogs retrieves integration logs
func (ie *DefaultIntegrationEngine) GetIntegrationLogs(integrationID string, filter *LogFilter) ([]*IntegrationLog, error) {
	ie.mu.RLock()
	var logs []*IntegrationLog
	for _, log := range ie.logs {
		if log.IntegrationID == integrationID && ie.matchesLogFilter(log, filter) {
			logs = append(logs, log)
		}
	}
	ie.mu.RUnlock()

	// Sort logs by timestamp (newest first)
	sort.Slice(logs, func(i, j int) bool {
		return logs[i].Timestamp.After(logs[j].Timestamp)
	})

	// Apply pagination
	if filter != nil {
		if filter.Offset > 0 && filter.Offset < len(logs) {
			logs = logs[filter.Offset:]
		}
		if filter.Limit > 0 && filter.Limit < len(logs) {
			logs = logs[:filter.Limit]
		}
	}

	return logs, nil
}
