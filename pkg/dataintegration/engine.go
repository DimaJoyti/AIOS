package dataintegration

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

// DefaultDataIntegrationEngine implements the DataIntegrationEngine interface
type DefaultDataIntegrationEngine struct {
	logger *logrus.Logger
	tracer trace.Tracer

	// Storage
	dataSources map[string]*DataSource
	pipelines   map[string]*DataPipeline
	logs        []*DataSourceLog

	// Connectors
	connectors map[string]DataConnector

	// Event handling
	eventHandlers map[string][]DataEventHandler

	// Metrics and monitoring
	sourceMetrics map[string]*DataSourceMetrics
	sourceHealth  map[string]*DataSourceHealth

	// Pipeline execution
	pipelineStatuses map[string]*PipelineStatus

	// Indexes
	sourcesByType     map[string][]string
	sourcesByProvider map[string][]string
	pipelinesBySource map[string][]string

	mu sync.RWMutex
}

// DataIntegrationEngineConfig represents configuration for the data integration engine
type DataIntegrationEngineConfig struct {
	MaxDataSources      int           `json:"max_data_sources"`
	MaxPipelines        int           `json:"max_pipelines"`
	DefaultTimeout      time.Duration `json:"default_timeout"`
	LogRetention        time.Duration `json:"log_retention"`
	MetricsRetention    time.Duration `json:"metrics_retention"`
	HealthCheckInterval time.Duration `json:"health_check_interval"`
	EnableMetrics       bool          `json:"enable_metrics"`
	EnableHealthChecks  bool          `json:"enable_health_checks"`
	MaxConcurrentJobs   int           `json:"max_concurrent_jobs"`
}

// NewDefaultDataIntegrationEngine creates a new data integration engine
func NewDefaultDataIntegrationEngine(config *DataIntegrationEngineConfig, logger *logrus.Logger) DataIntegrationEngine {
	if config == nil {
		config = &DataIntegrationEngineConfig{
			MaxDataSources:      1000,
			MaxPipelines:        5000,
			DefaultTimeout:      30 * time.Second,
			LogRetention:        7 * 24 * time.Hour,
			MetricsRetention:    30 * 24 * time.Hour,
			HealthCheckInterval: 5 * time.Minute,
			EnableMetrics:       true,
			EnableHealthChecks:  true,
			MaxConcurrentJobs:   10,
		}
	}

	engine := &DefaultDataIntegrationEngine{
		logger:            logger,
		tracer:            otel.Tracer("dataintegration.engine"),
		dataSources:       make(map[string]*DataSource),
		pipelines:         make(map[string]*DataPipeline),
		logs:              make([]*DataSourceLog, 0),
		connectors:        make(map[string]DataConnector),
		eventHandlers:     make(map[string][]DataEventHandler),
		sourceMetrics:     make(map[string]*DataSourceMetrics),
		sourceHealth:      make(map[string]*DataSourceHealth),
		pipelineStatuses:  make(map[string]*PipelineStatus),
		sourcesByType:     make(map[string][]string),
		sourcesByProvider: make(map[string][]string),
		pipelinesBySource: make(map[string][]string),
	}

	// Start background tasks
	if config.EnableHealthChecks {
		go engine.startHealthCheckLoop(config.HealthCheckInterval)
	}

	return engine
}

// Data Source Management

// CreateDataSource creates a new data source
func (die *DefaultDataIntegrationEngine) CreateDataSource(source *DataSource) (*DataSource, error) {
	_, span := die.tracer.Start(context.Background(), "dataintegration.create_data_source")
	defer span.End()

	if source.ID == "" {
		source.ID = uuid.New().String()
	}

	// Set timestamps
	now := time.Now()
	source.CreatedAt = now
	source.UpdatedAt = now

	// Set default status
	if source.Status == "" {
		source.Status = DataSourceStatusConfiguring
	}

	// Set default settings
	if source.Settings == nil {
		source.Settings = &DataSourceSettings{
			AutoSync:        true,
			SyncInterval:    15 * time.Minute,
			EnableStreaming: false,
			EnableEvents:    true,
			DataRetention:   30 * 24 * time.Hour,
			MaxRecords:      1000000,
			LogLevel:        "info",
			NotifyOnError:   true,
		}
	}

	// Set default config
	if source.Config == nil {
		source.Config = &DataSourceConfig{
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

	// Validate data source
	if err := die.validateDataSource(source); err != nil {
		return nil, fmt.Errorf("data source validation failed: %w", err)
	}

	die.mu.Lock()
	die.dataSources[source.ID] = source
	die.sourcesByType[source.Type] = append(die.sourcesByType[source.Type], source.ID)
	die.sourcesByProvider[source.Provider] = append(die.sourcesByProvider[source.Provider], source.ID)

	// Initialize health status
	die.sourceHealth[source.ID] = &DataSourceHealth{
		SourceID:  source.ID,
		Status:    HealthStatusUnknown,
		Message:   "Data source created",
		LastCheck: now,
		Uptime:    0,
		Checks:    []*HealthCheck{},
	}
	die.mu.Unlock()

	// Log creation
	die.logDataSourceEvent(source.ID, LogLevelInfo, "data_source_created",
		fmt.Sprintf("Data source '%s' created", source.Name), nil, nil)

	span.SetAttributes(
		attribute.String("data_source.id", source.ID),
		attribute.String("data_source.name", source.Name),
		attribute.String("data_source.type", source.Type),
		attribute.String("data_source.provider", source.Provider),
	)

	die.logger.WithFields(logrus.Fields{
		"data_source_id":   source.ID,
		"data_source_name": source.Name,
		"data_source_type": source.Type,
		"provider":         source.Provider,
	}).Info("Data source created successfully")

	return source, nil
}

// GetDataSource retrieves a data source by ID
func (die *DefaultDataIntegrationEngine) GetDataSource(sourceID string) (*DataSource, error) {
	_, span := die.tracer.Start(context.Background(), "dataintegration.get_data_source")
	defer span.End()

	die.mu.RLock()
	source, exists := die.dataSources[sourceID]
	die.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("data source not found: %s", sourceID)
	}

	span.SetAttributes(attribute.String("data_source.id", sourceID))

	return source, nil
}

// UpdateDataSource updates an existing data source
func (die *DefaultDataIntegrationEngine) UpdateDataSource(source *DataSource) (*DataSource, error) {
	_, span := die.tracer.Start(context.Background(), "dataintegration.update_data_source")
	defer span.End()

	die.mu.Lock()
	existing, exists := die.dataSources[source.ID]
	if !exists {
		die.mu.Unlock()
		return nil, fmt.Errorf("data source not found: %s", source.ID)
	}

	// Preserve creation info
	source.CreatedBy = existing.CreatedBy
	source.CreatedAt = existing.CreatedAt
	source.UpdatedAt = time.Now()

	// Validate data source
	if err := die.validateDataSource(source); err != nil {
		die.mu.Unlock()
		return nil, fmt.Errorf("data source validation failed: %w", err)
	}

	die.dataSources[source.ID] = source
	die.mu.Unlock()

	// Log update
	die.logDataSourceEvent(source.ID, LogLevelInfo, "data_source_updated",
		fmt.Sprintf("Data source '%s' updated", source.Name), nil, nil)

	span.SetAttributes(attribute.String("data_source.id", source.ID))

	die.logger.WithField("data_source_id", source.ID).Info("Data source updated successfully")

	return source, nil
}

// DeleteDataSource deletes a data source
func (die *DefaultDataIntegrationEngine) DeleteDataSource(sourceID string) error {
	_, span := die.tracer.Start(context.Background(), "dataintegration.delete_data_source")
	defer span.End()

	die.mu.Lock()
	source, exists := die.dataSources[sourceID]
	if !exists {
		die.mu.Unlock()
		return fmt.Errorf("data source not found: %s", sourceID)
	}

	// Remove from indexes
	die.removeFromIndex(die.sourcesByType[source.Type], sourceID)
	die.removeFromIndex(die.sourcesByProvider[source.Provider], sourceID)

	// Delete associated pipelines
	if pipelineIDs, exists := die.pipelinesBySource[sourceID]; exists {
		for _, pipelineID := range pipelineIDs {
			delete(die.pipelines, pipelineID)
			delete(die.pipelineStatuses, pipelineID)
		}
		delete(die.pipelinesBySource, sourceID)
	}

	// Clean up
	delete(die.dataSources, sourceID)
	delete(die.sourceHealth, sourceID)
	delete(die.sourceMetrics, sourceID)
	die.mu.Unlock()

	// Disconnect connector if connected
	if connector, err := die.GetConnector(source.Type); err == nil {
		if connector.IsConnected() {
			connector.Disconnect(context.Background())
		}
	}

	// Log deletion
	die.logDataSourceEvent(sourceID, LogLevelInfo, "data_source_deleted",
		fmt.Sprintf("Data source '%s' deleted", source.Name), nil, nil)

	span.SetAttributes(attribute.String("data_source.id", sourceID))

	die.logger.WithField("data_source_id", sourceID).Info("Data source deleted successfully")

	return nil
}

// ListDataSources lists data sources with filtering
func (die *DefaultDataIntegrationEngine) ListDataSources(filter *DataSourceFilter) ([]*DataSource, error) {
	_, span := die.tracer.Start(context.Background(), "dataintegration.list_data_sources")
	defer span.End()

	die.mu.RLock()
	var sources []*DataSource
	for _, source := range die.dataSources {
		if die.matchesDataSourceFilter(source, filter) {
			sources = append(sources, source)
		}
	}
	die.mu.RUnlock()

	// Sort sources by creation date (newest first)
	sort.Slice(sources, func(i, j int) bool {
		return sources[i].CreatedAt.After(sources[j].CreatedAt)
	})

	// Apply pagination
	if filter != nil {
		if filter.Offset > 0 && filter.Offset < len(sources) {
			sources = sources[filter.Offset:]
		}
		if filter.Limit > 0 && filter.Limit < len(sources) {
			sources = sources[:filter.Limit]
		}
	}

	span.SetAttributes(attribute.Int("data_sources.count", len(sources)))

	return sources, nil
}

// Data Source Operations

// EnableDataSource enables a data source
func (die *DefaultDataIntegrationEngine) EnableDataSource(sourceID string) error {
	_, span := die.tracer.Start(context.Background(), "dataintegration.enable_data_source")
	defer span.End()

	die.mu.Lock()
	source, exists := die.dataSources[sourceID]
	if !exists {
		die.mu.Unlock()
		return fmt.Errorf("data source not found: %s", sourceID)
	}

	source.Status = DataSourceStatusActive
	source.UpdatedAt = time.Now()
	die.mu.Unlock()

	// Connect connector
	if connector, err := die.GetConnector(source.Type); err == nil {
		ctx, cancel := context.WithTimeout(context.Background(), source.Config.Timeout)
		defer cancel()

		if err := connector.Connect(ctx, source.Config.Custom); err != nil {
			source.Status = DataSourceStatusError
			return fmt.Errorf("failed to connect connector: %w", err)
		}
	}

	// Log enablement
	die.logDataSourceEvent(sourceID, LogLevelInfo, "data_source_enabled",
		fmt.Sprintf("Data source '%s' enabled", source.Name), nil, nil)

	span.SetAttributes(attribute.String("data_source.id", sourceID))

	die.logger.WithField("data_source_id", sourceID).Info("Data source enabled successfully")

	return nil
}

// DisableDataSource disables a data source
func (die *DefaultDataIntegrationEngine) DisableDataSource(sourceID string) error {
	_, span := die.tracer.Start(context.Background(), "dataintegration.disable_data_source")
	defer span.End()

	die.mu.Lock()
	source, exists := die.dataSources[sourceID]
	if !exists {
		die.mu.Unlock()
		return fmt.Errorf("data source not found: %s", sourceID)
	}

	source.Status = DataSourceStatusInactive
	source.UpdatedAt = time.Now()
	die.mu.Unlock()

	// Disconnect connector
	if connector, err := die.GetConnector(source.Type); err == nil {
		if connector.IsConnected() {
			connector.Disconnect(context.Background())
		}
	}

	// Log disablement
	die.logDataSourceEvent(sourceID, LogLevelInfo, "data_source_disabled",
		fmt.Sprintf("Data source '%s' disabled", source.Name), nil, nil)

	span.SetAttributes(attribute.String("data_source.id", sourceID))

	die.logger.WithField("data_source_id", sourceID).Info("Data source disabled successfully")

	return nil
}

// TestDataSource tests a data source connection
func (die *DefaultDataIntegrationEngine) TestDataSource(sourceID string) (*DataSourceTestResult, error) {
	_, span := die.tracer.Start(context.Background(), "dataintegration.test_data_source")
	defer span.End()

	start := time.Now()

	source, err := die.GetDataSource(sourceID)
	if err != nil {
		return nil, err
	}

	connector, err := die.GetConnector(source.Type)
	if err != nil {
		return &DataSourceTestResult{
			Success:  false,
			Message:  fmt.Sprintf("Connector not found: %s", source.Type),
			Duration: time.Since(start),
			TestedAt: time.Now(),
		}, nil
	}

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), source.Config.Timeout)
	defer cancel()

	err = connector.TestConnection(ctx)
	result := &DataSourceTestResult{
		Success:      err == nil,
		Duration:     time.Since(start),
		TestedAt:     time.Now(),
		Capabilities: connector.GetSupportedOperations(),
	}

	if err != nil {
		result.Message = fmt.Sprintf("Connection test failed: %v", err)
		die.logDataSourceEvent(sourceID, LogLevelError, "test_failed", result.Message, nil, err)
	} else {
		result.Message = "Connection test successful"
		die.logDataSourceEvent(sourceID, LogLevelInfo, "test_successful", result.Message, nil, nil)
	}

	span.SetAttributes(
		attribute.String("data_source.id", sourceID),
		attribute.Bool("test.success", result.Success),
		attribute.Int64("test.duration_ms", result.Duration.Milliseconds()),
	)

	return result, nil
}

// RefreshDataSource refreshes a data source's credentials or configuration
func (die *DefaultDataIntegrationEngine) RefreshDataSource(sourceID string) error {
	_, span := die.tracer.Start(context.Background(), "dataintegration.refresh_data_source")
	defer span.End()

	source, err := die.GetDataSource(sourceID)
	if err != nil {
		return err
	}

	// Update last sync time
	now := time.Now()
	source.LastSyncAt = &now
	source.UpdatedAt = now

	die.mu.Lock()
	die.dataSources[sourceID] = source
	die.mu.Unlock()

	// Log refresh
	die.logDataSourceEvent(sourceID, LogLevelInfo, "data_source_refreshed",
		fmt.Sprintf("Data source '%s' refreshed", source.Name), nil, nil)

	span.SetAttributes(attribute.String("data_source.id", sourceID))

	die.logger.WithField("data_source_id", sourceID).Info("Data source refreshed successfully")

	return nil
}

// Connector Management

// RegisterConnector registers a data connector
func (die *DefaultDataIntegrationEngine) RegisterConnector(connector DataConnector) error {
	_, span := die.tracer.Start(context.Background(), "dataintegration.register_connector")
	defer span.End()

	connectorType := connector.GetType()

	die.mu.Lock()
	die.connectors[connectorType] = connector
	die.mu.Unlock()

	span.SetAttributes(
		attribute.String("connector.type", connectorType),
		attribute.String("connector.name", connector.GetName()),
		attribute.String("connector.version", connector.GetVersion()),
	)

	die.logger.WithFields(logrus.Fields{
		"connector_type":    connectorType,
		"connector_name":    connector.GetName(),
		"connector_version": connector.GetVersion(),
	}).Info("Connector registered successfully")

	return nil
}

// GetConnector retrieves a connector by type
func (die *DefaultDataIntegrationEngine) GetConnector(connectorType string) (DataConnector, error) {
	_, span := die.tracer.Start(context.Background(), "dataintegration.get_connector")
	defer span.End()

	die.mu.RLock()
	connector, exists := die.connectors[connectorType]
	die.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("connector not found: %s", connectorType)
	}

	span.SetAttributes(attribute.String("connector.type", connectorType))

	return connector, nil
}

// ListConnectors lists all registered connectors
func (die *DefaultDataIntegrationEngine) ListConnectors() []string {
	_, span := die.tracer.Start(context.Background(), "dataintegration.list_connectors")
	defer span.End()

	die.mu.RLock()
	var connectorTypes []string
	for connectorType := range die.connectors {
		connectorTypes = append(connectorTypes, connectorType)
	}
	die.mu.RUnlock()

	sort.Strings(connectorTypes)

	span.SetAttributes(attribute.Int("connectors.count", len(connectorTypes)))

	return connectorTypes
}

// Pipeline Management (placeholder implementations)

// CreatePipeline creates a new data pipeline
func (die *DefaultDataIntegrationEngine) CreatePipeline(pipeline *DataPipeline) (*DataPipeline, error) {
	// Implementation would be added here
	return pipeline, nil
}

// GetPipeline retrieves a pipeline by ID
func (die *DefaultDataIntegrationEngine) GetPipeline(pipelineID string) (*DataPipeline, error) {
	// Implementation would be added here
	return nil, fmt.Errorf("pipeline not found: %s", pipelineID)
}

// UpdatePipeline updates an existing pipeline
func (die *DefaultDataIntegrationEngine) UpdatePipeline(pipeline *DataPipeline) (*DataPipeline, error) {
	// Implementation would be added here
	return pipeline, nil
}

// DeletePipeline deletes a pipeline
func (die *DefaultDataIntegrationEngine) DeletePipeline(pipelineID string) error {
	// Implementation would be added here
	return nil
}

// ListPipelines lists pipelines with filtering
func (die *DefaultDataIntegrationEngine) ListPipelines(filter *PipelineFilter) ([]*DataPipeline, error) {
	// Implementation would be added here
	return []*DataPipeline{}, nil
}

// StartPipeline starts a pipeline
func (die *DefaultDataIntegrationEngine) StartPipeline(pipelineID string) error {
	// Implementation would be added here
	return nil
}

// StopPipeline stops a pipeline
func (die *DefaultDataIntegrationEngine) StopPipeline(pipelineID string) error {
	// Implementation would be added here
	return nil
}

// GetPipelineStatus gets pipeline status
func (die *DefaultDataIntegrationEngine) GetPipelineStatus(pipelineID string) (*PipelineStatus, error) {
	// Implementation would be added here
	status := PipelineStatusStopped
	return &status, nil
}

// GetPipelineMetrics gets pipeline metrics
func (die *DefaultDataIntegrationEngine) GetPipelineMetrics(pipelineID string, timeRange *TimeRange) (*PipelineMetrics, error) {
	// Return mock metrics for demo
	return &PipelineMetrics{
		PipelineID:       pipelineID,
		TimeRange:        timeRange,
		RunCount:         24,
		SuccessfulRuns:   22,
		FailedRuns:       2,
		RecordsProcessed: 15000,
		AverageRunTime:   5 * time.Minute,
		LastRunTime:      4 * time.Minute,
		TotalDataSize:    1024 * 1024 * 50, // 50MB
		ErrorsByStage: map[string]int64{
			"extraction":     1,
			"transformation": 1,
			"validation":     0,
			"storage":        0,
		},
	}, nil
}

// Data Processing (placeholder implementations)

// ProcessData processes data through a pipeline
func (die *DefaultDataIntegrationEngine) ProcessData(ctx context.Context, data *DataRecord, pipeline *DataPipeline) (*ProcessedData, error) {
	// Implementation would be added here
	return &ProcessedData{
		Records:      []*DataRecord{data},
		ProcessedAt:  time.Now(),
		PipelineID:   pipeline.ID,
		SourceID:     data.SourceID,
		TotalRecords: 1,
		ValidRecords: 1,
		ErrorRecords: 0,
	}, nil
}

// ValidateData validates data against a schema
func (die *DefaultDataIntegrationEngine) ValidateData(data *DataRecord, schema *DataSchema) (*ValidationResult, error) {
	// Implementation would be added here
	return &ValidationResult{
		Valid:        true,
		ValidRecords: 1,
		TotalRecords: 1,
		ValidatedAt:  time.Now(),
	}, nil
}

// TransformData applies transformations to data
func (die *DefaultDataIntegrationEngine) TransformData(data *DataRecord, transformations []*DataTransformation) (*DataRecord, error) {
	// Implementation would be added here
	return data, nil
}

// Storage Management (placeholder implementations)

// StoreData stores processed data
func (die *DefaultDataIntegrationEngine) StoreData(ctx context.Context, data *ProcessedData, storage *StorageConfig) error {
	// Implementation would be added here
	return nil
}

// RetrieveData retrieves data based on query
func (die *DefaultDataIntegrationEngine) RetrieveData(ctx context.Context, query *DataQuery) (*DataResult, error) {
	// Implementation would be added here
	return &DataResult{
		Records:    []*DataRecord{},
		TotalCount: 0,
		QueryTime:  10 * time.Millisecond,
		ExecutedAt: time.Now(),
	}, nil
}

// IndexData indexes data for fast retrieval
func (die *DefaultDataIntegrationEngine) IndexData(ctx context.Context, data *ProcessedData, indexConfig *IndexConfig) error {
	// Implementation would be added here
	return nil
}

// Monitoring and Analytics

// GetDataSourceHealth retrieves data source health status
func (die *DefaultDataIntegrationEngine) GetDataSourceHealth(sourceID string) (*DataSourceHealth, error) {
	die.mu.RLock()
	health, exists := die.sourceHealth[sourceID]
	die.mu.RUnlock()

	if !exists {
		return &DataSourceHealth{
			SourceID:  sourceID,
			Status:    HealthStatusUnknown,
			Message:   "Health status not available",
			LastCheck: time.Now(),
			Uptime:    0,
			Checks:    []*HealthCheck{},
		}, nil
	}

	return health, nil
}

// GetDataSourceMetrics retrieves data source metrics
func (die *DefaultDataIntegrationEngine) GetDataSourceMetrics(sourceID string, timeRange *TimeRange) (*DataSourceMetrics, error) {
	// Return mock metrics for demo
	return &DataSourceMetrics{
		SourceID:         sourceID,
		TimeRange:        timeRange,
		RecordsExtracted: 5000,
		RecordsProcessed: 4800,
		RecordsStored:    4750,
		ErrorCount:       50,
		SuccessRate:      0.96,
		AverageLatency:   200 * time.Millisecond,
		P95Latency:       500 * time.Millisecond,
		P99Latency:       1 * time.Second,
		ThroughputPerSec: 5.2,
		ErrorsByType: map[string]int64{
			"connection_timeout": 20,
			"parse_error":        15,
			"validation_error":   10,
			"storage_error":      5,
		},
	}, nil
}

// GetDataSourceLogs retrieves data source logs
func (die *DefaultDataIntegrationEngine) GetDataSourceLogs(sourceID string, filter *LogFilter) ([]*DataSourceLog, error) {
	die.mu.RLock()
	var logs []*DataSourceLog
	for _, log := range die.logs {
		if log.SourceID == sourceID && die.matchesLogFilter(log, filter) {
			logs = append(logs, log)
		}
	}
	die.mu.RUnlock()

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

// Event Management (placeholder implementations)

// PublishDataEvent publishes a data event
func (die *DefaultDataIntegrationEngine) PublishDataEvent(event *DataEvent) error {
	// Implementation would be added here
	return nil
}

// SubscribeToDataEvents subscribes to data events
func (die *DefaultDataIntegrationEngine) SubscribeToDataEvents(eventType string, handler DataEventHandler) error {
	// Implementation would be added here
	return nil
}

// UnsubscribeFromDataEvents unsubscribes from data events
func (die *DefaultDataIntegrationEngine) UnsubscribeFromDataEvents(eventType string, handler DataEventHandler) error {
	// Implementation would be added here
	return nil
}
