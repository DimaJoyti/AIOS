package dataintegration

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Helper methods for data integration engine

// Validation methods

// validateDataSource validates a data source configuration
func (die *DefaultDataIntegrationEngine) validateDataSource(source *DataSource) error {
	if source.Name == "" {
		return fmt.Errorf("data source name is required")
	}

	if len(source.Name) > 100 {
		return fmt.Errorf("data source name must be 100 characters or less")
	}

	if source.Type == "" {
		return fmt.Errorf("data source type is required")
	}

	if source.Provider == "" {
		return fmt.Errorf("data source provider is required")
	}

	// Validate credentials if provided
	if source.Credentials != nil {
		if err := die.validateCredentials(source.Credentials); err != nil {
			return fmt.Errorf("credentials validation failed: %w", err)
		}
	}

	return nil
}

// validateCredentials validates data source credentials
func (die *DefaultDataIntegrationEngine) validateCredentials(credentials *DataSourceCredentials) error {
	if credentials.Type == "" {
		return fmt.Errorf("credential type is required")
	}

	switch credentials.Type {
	case CredentialTypeAPIKey:
		if credentials.APIKey == "" {
			return fmt.Errorf("API key is required for API key authentication")
		}
	case CredentialTypeBearer:
		if credentials.Token == "" {
			return fmt.Errorf("token is required for bearer authentication")
		}
	case CredentialTypeBasic:
		if credentials.Username == "" || credentials.Password == "" {
			return fmt.Errorf("username and password are required for basic authentication")
		}
	case CredentialTypeOAuth2:
		if credentials.OAuth == nil {
			return fmt.Errorf("OAuth credentials are required for OAuth2 authentication")
		}
		if credentials.OAuth.ClientID == "" || credentials.OAuth.ClientSecret == "" {
			return fmt.Errorf("client ID and client secret are required for OAuth2")
		}
	}

	return nil
}

// validatePipeline validates a pipeline configuration
func (die *DefaultDataIntegrationEngine) validatePipeline(pipeline *DataPipeline) error {
	if pipeline.Name == "" {
		return fmt.Errorf("pipeline name is required")
	}

	if len(pipeline.Name) > 100 {
		return fmt.Errorf("pipeline name must be 100 characters or less")
	}

	if pipeline.DataSourceID == "" {
		return fmt.Errorf("data source ID is required")
	}

	// Validate that data source exists
	if _, exists := die.dataSources[pipeline.DataSourceID]; !exists {
		return fmt.Errorf("data source not found: %s", pipeline.DataSourceID)
	}

	// Validate transformations
	for i, transformation := range pipeline.Transformations {
		if transformation.Name == "" {
			return fmt.Errorf("transformation %d name is required", i)
		}
		if transformation.Type == "" {
			return fmt.Errorf("transformation %d type is required", i)
		}
	}

	return nil
}

// Filter matching methods

// matchesDataSourceFilter checks if a data source matches the given filter
func (die *DefaultDataIntegrationEngine) matchesDataSourceFilter(source *DataSource, filter *DataSourceFilter) bool {
	if filter == nil {
		return true
	}

	// Type filter
	if filter.Type != "" && source.Type != filter.Type {
		return false
	}

	// Provider filter
	if filter.Provider != "" && source.Provider != filter.Provider {
		return false
	}

	// Status filter
	if filter.Status != "" && source.Status != filter.Status {
		return false
	}

	// Created by filter
	if filter.CreatedBy != "" && source.CreatedBy != filter.CreatedBy {
		return false
	}

	// Search filter
	if filter.Search != "" {
		searchLower := strings.ToLower(filter.Search)
		if !(strings.Contains(strings.ToLower(source.Name), searchLower) ||
			strings.Contains(strings.ToLower(source.Description), searchLower) ||
			strings.Contains(strings.ToLower(source.Type), searchLower) ||
			strings.Contains(strings.ToLower(source.Provider), searchLower)) {
			return false
		}
	}

	return true
}

// matchesPipelineFilter checks if a pipeline matches the given filter
func (die *DefaultDataIntegrationEngine) matchesPipelineFilter(pipeline *DataPipeline, filter *PipelineFilter) bool {
	if filter == nil {
		return true
	}

	// Data source ID filter
	if filter.DataSourceID != "" && pipeline.DataSourceID != filter.DataSourceID {
		return false
	}

	// Status filter
	if filter.Status != "" && pipeline.Status != filter.Status {
		return false
	}

	// Created by filter
	if filter.CreatedBy != "" && pipeline.CreatedBy != filter.CreatedBy {
		return false
	}

	// Search filter
	if filter.Search != "" {
		searchLower := strings.ToLower(filter.Search)
		if !(strings.Contains(strings.ToLower(pipeline.Name), searchLower) ||
			strings.Contains(strings.ToLower(pipeline.Description), searchLower)) {
			return false
		}
	}

	return true
}

// matchesLogFilter checks if a log entry matches the given filter
func (die *DefaultDataIntegrationEngine) matchesLogFilter(log *DataSourceLog, filter *LogFilter) bool {
	if filter == nil {
		return true
	}

	// Level filter
	if filter.Level != "" && log.Level != filter.Level {
		return false
	}

	// Operation filter
	if filter.Operation != "" && log.Operation != filter.Operation {
		return false
	}

	// Time range filters
	if filter.Since != nil && log.Timestamp.Before(*filter.Since) {
		return false
	}
	if filter.Until != nil && log.Timestamp.After(*filter.Until) {
		return false
	}

	// Search filter
	if filter.Search != "" {
		searchLower := strings.ToLower(filter.Search)
		if !(strings.Contains(strings.ToLower(log.Message), searchLower) ||
			strings.Contains(strings.ToLower(log.Operation), searchLower) ||
			strings.Contains(strings.ToLower(log.Error), searchLower)) {
			return false
		}
	}

	return true
}

// Utility methods

// removeFromIndex removes an item from a string slice
func (die *DefaultDataIntegrationEngine) removeFromIndex(slice []string, item string) []string {
	for i, v := range slice {
		if v == item {
			return append(slice[:i], slice[i+1:]...)
		}
	}
	return slice
}

// logDataSourceEvent logs a data source event
func (die *DefaultDataIntegrationEngine) logDataSourceEvent(sourceID string, level LogLevel, operation, message string, duration *time.Duration, err error) {
	logEntry := &DataSourceLog{
		ID:        uuid.New().String(),
		SourceID:  sourceID,
		Level:     level,
		Message:   message,
		Operation: operation,
		Timestamp: time.Now(),
	}

	if duration != nil {
		logEntry.Duration = *duration
	}

	if err != nil {
		logEntry.Error = err.Error()
	}

	die.mu.Lock()
	die.logs = append(die.logs, logEntry)
	
	// Keep only recent logs (simple cleanup)
	if len(die.logs) > 10000 {
		die.logs = die.logs[1000:]
	}
	die.mu.Unlock()
}

// startHealthCheckLoop starts the health check background loop
func (die *DefaultDataIntegrationEngine) startHealthCheckLoop(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		die.performHealthChecks()
	}
}

// performHealthChecks performs health checks on all active data sources
func (die *DefaultDataIntegrationEngine) performHealthChecks() {
	die.mu.RLock()
	var activeSources []*DataSource
	for _, source := range die.dataSources {
		if source.Status == DataSourceStatusActive {
			activeSources = append(activeSources, source)
		}
	}
	die.mu.RUnlock()

	for _, source := range activeSources {
		die.performDataSourceHealthCheck(source)
	}
}

// performDataSourceHealthCheck performs a health check on a specific data source
func (die *DefaultDataIntegrationEngine) performDataSourceHealthCheck(source *DataSource) {
	start := time.Now()
	
	connector, err := die.GetConnector(source.Type)
	if err != nil {
		die.updateHealthStatus(source.ID, HealthStatusUnhealthy, 
			fmt.Sprintf("Connector not found: %s", source.Type), []*HealthCheck{})
		return
	}

	// Get connector health
	connectorHealth := connector.GetHealth()
	
	// Perform connection test
	var checks []*HealthCheck
	
	// Connection check
	connectionCheck := &HealthCheck{
		Name:      "connection",
		Status:    HealthStatusHealthy,
		Message:   "Connection is healthy",
		Duration:  time.Since(start),
		Timestamp: time.Now(),
	}

	if !connector.IsConnected() {
		connectionCheck.Status = HealthStatusUnhealthy
		connectionCheck.Message = "Not connected to external service"
	}

	checks = append(checks, connectionCheck)

	// Connector-specific health check
	connectorCheck := &HealthCheck{
		Name:      "connector",
		Status:    connectorHealth.Status,
		Message:   connectorHealth.Message,
		Duration:  time.Since(start),
		Timestamp: time.Now(),
	}

	checks = append(checks, connectorCheck)

	// Determine overall status
	overallStatus := HealthStatusHealthy
	message := "All checks passed"

	for _, check := range checks {
		if check.Status == HealthStatusUnhealthy {
			overallStatus = HealthStatusUnhealthy
			message = "One or more health checks failed"
			break
		} else if check.Status == HealthStatusDegraded && overallStatus == HealthStatusHealthy {
			overallStatus = HealthStatusDegraded
			message = "Service is degraded"
		}
	}

	die.updateHealthStatus(source.ID, overallStatus, message, checks)
}

// updateHealthStatus updates the health status of a data source
func (die *DefaultDataIntegrationEngine) updateHealthStatus(sourceID string, status HealthStatus, message string, checks []*HealthCheck) {
	die.mu.Lock()
	defer die.mu.Unlock()

	health, exists := die.sourceHealth[sourceID]
	if !exists {
		health = &DataSourceHealth{
			SourceID: sourceID,
		}
		die.sourceHealth[sourceID] = health
	}

	now := time.Now()
	
	// Calculate uptime
	if health.LastCheck.IsZero() {
		health.Uptime = 0
	} else {
		if status == HealthStatusHealthy {
			health.Uptime += now.Sub(health.LastCheck)
		}
	}

	health.Status = status
	health.Message = message
	health.LastCheck = now
	health.Checks = checks
}
