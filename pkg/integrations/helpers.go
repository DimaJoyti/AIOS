package integrations

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Helper methods for integration engine

// Validation methods

// validateIntegration validates an integration configuration
func (ie *DefaultIntegrationEngine) validateIntegration(integration *Integration) error {
	if integration.Name == "" {
		return fmt.Errorf("integration name is required")
	}

	if len(integration.Name) > 100 {
		return fmt.Errorf("integration name must be 100 characters or less")
	}

	if integration.Type == "" {
		return fmt.Errorf("integration type is required")
	}

	if integration.Provider == "" {
		return fmt.Errorf("integration provider is required")
	}

	// Validate credentials if provided
	if integration.Credentials != nil {
		if err := ie.validateCredentials(integration.Credentials); err != nil {
			return fmt.Errorf("credentials validation failed: %w", err)
		}
	}

	return nil
}

// validateCredentials validates integration credentials
func (ie *DefaultIntegrationEngine) validateCredentials(credentials *IntegrationCredentials) error {
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

// validateWebhook validates a webhook configuration
func (ie *DefaultIntegrationEngine) validateWebhook(webhook *Webhook) error {
	if webhook.Name == "" {
		return fmt.Errorf("webhook name is required")
	}

	if len(webhook.Name) > 100 {
		return fmt.Errorf("webhook name must be 100 characters or less")
	}

	if webhook.URL == "" {
		return fmt.Errorf("webhook URL is required")
	}

	if webhook.IntegrationID == "" {
		return fmt.Errorf("integration ID is required")
	}

	// Validate that integration exists
	if _, exists := ie.integrations[webhook.IntegrationID]; !exists {
		return fmt.Errorf("integration not found: %s", webhook.IntegrationID)
	}

	if len(webhook.Events) == 0 {
		return fmt.Errorf("at least one event must be specified")
	}

	return nil
}

// Filter matching methods

// matchesIntegrationFilter checks if an integration matches the given filter
func (ie *DefaultIntegrationEngine) matchesIntegrationFilter(integration *Integration, filter *IntegrationFilter) bool {
	if filter == nil {
		return true
	}

	// Type filter
	if filter.Type != "" && integration.Type != filter.Type {
		return false
	}

	// Provider filter
	if filter.Provider != "" && integration.Provider != filter.Provider {
		return false
	}

	// Status filter
	if filter.Status != "" && integration.Status != filter.Status {
		return false
	}

	// Created by filter
	if filter.CreatedBy != "" && integration.CreatedBy != filter.CreatedBy {
		return false
	}

	// Search filter
	if filter.Search != "" {
		searchLower := strings.ToLower(filter.Search)
		if !(strings.Contains(strings.ToLower(integration.Name), searchLower) ||
			strings.Contains(strings.ToLower(integration.Description), searchLower) ||
			strings.Contains(strings.ToLower(integration.Type), searchLower) ||
			strings.Contains(strings.ToLower(integration.Provider), searchLower)) {
			return false
		}
	}

	return true
}

// matchesWebhookFilter checks if a webhook matches the given filter
func (ie *DefaultIntegrationEngine) matchesWebhookFilter(webhook *Webhook, filter *WebhookFilter) bool {
	if filter == nil {
		return true
	}

	// Integration ID filter
	if filter.IntegrationID != "" && webhook.IntegrationID != filter.IntegrationID {
		return false
	}

	// Status filter
	if filter.Status != "" && webhook.Status != filter.Status {
		return false
	}

	// Event filter
	if filter.Event != "" {
		found := false
		for _, event := range webhook.Events {
			if event == filter.Event {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Search filter
	if filter.Search != "" {
		searchLower := strings.ToLower(filter.Search)
		if !(strings.Contains(strings.ToLower(webhook.Name), searchLower) ||
			strings.Contains(strings.ToLower(webhook.URL), searchLower)) {
			return false
		}
	}

	return true
}

// matchesLogFilter checks if a log entry matches the given filter
func (ie *DefaultIntegrationEngine) matchesLogFilter(log *IntegrationLog, filter *LogFilter) bool {
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
func (ie *DefaultIntegrationEngine) removeFromIndex(slice []string, item string) []string {
	for i, v := range slice {
		if v == item {
			return append(slice[:i], slice[i+1:]...)
		}
	}
	return slice
}

// logIntegrationEvent logs an integration event
func (ie *DefaultIntegrationEngine) logIntegrationEvent(integrationID string, level LogLevel, operation, message string, duration *time.Duration, err error) {
	logEntry := &IntegrationLog{
		ID:            uuid.New().String(),
		IntegrationID: integrationID,
		Level:         level,
		Message:       message,
		Operation:     operation,
		Timestamp:     time.Now(),
	}

	if duration != nil {
		logEntry.Duration = *duration
	}

	if err != nil {
		logEntry.Error = err.Error()
	}

	ie.mu.Lock()
	ie.logs = append(ie.logs, logEntry)
	
	// Keep only recent logs (simple cleanup)
	if len(ie.logs) > 10000 {
		ie.logs = ie.logs[1000:]
	}
	ie.mu.Unlock()
}

// startHealthCheckLoop starts the health check background loop
func (ie *DefaultIntegrationEngine) startHealthCheckLoop(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		ie.performHealthChecks()
	}
}

// performHealthChecks performs health checks on all active integrations
func (ie *DefaultIntegrationEngine) performHealthChecks() {
	ie.mu.RLock()
	var activeIntegrations []*Integration
	for _, integration := range ie.integrations {
		if integration.Status == IntegrationStatusActive {
			activeIntegrations = append(activeIntegrations, integration)
		}
	}
	ie.mu.RUnlock()

	for _, integration := range activeIntegrations {
		ie.performIntegrationHealthCheck(integration)
	}
}

// performIntegrationHealthCheck performs a health check on a specific integration
func (ie *DefaultIntegrationEngine) performIntegrationHealthCheck(integration *Integration) {
	start := time.Now()
	
	adapter, err := ie.GetAdapter(integration.Type)
	if err != nil {
		ie.updateHealthStatus(integration.ID, HealthStatusUnhealthy, 
			fmt.Sprintf("Adapter not found: %s", integration.Type), []*HealthCheck{})
		return
	}

	// Get adapter health
	adapterHealth := adapter.GetHealth()
	
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

	if !adapter.IsConnected() {
		connectionCheck.Status = HealthStatusUnhealthy
		connectionCheck.Message = "Not connected to external service"
	}

	checks = append(checks, connectionCheck)

	// Adapter-specific health check
	adapterCheck := &HealthCheck{
		Name:      "adapter",
		Status:    adapterHealth.Status,
		Message:   adapterHealth.Message,
		Duration:  time.Since(start),
		Timestamp: time.Now(),
	}

	checks = append(checks, adapterCheck)

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

	ie.updateHealthStatus(integration.ID, overallStatus, message, checks)
}

// updateHealthStatus updates the health status of an integration
func (ie *DefaultIntegrationEngine) updateHealthStatus(integrationID string, status HealthStatus, message string, checks []*HealthCheck) {
	ie.mu.Lock()
	defer ie.mu.Unlock()

	health, exists := ie.health[integrationID]
	if !exists {
		health = &IntegrationHealth{
			IntegrationID: integrationID,
		}
		ie.health[integrationID] = health
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
