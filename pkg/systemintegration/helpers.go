package systemintegration

import (
	"fmt"
	"strings"
	"time"
)

// Helper methods for system integration hub

// Validation methods

// validateService validates a service configuration
func (hub *DefaultSystemIntegrationHub) validateService(service *Service) error {
	if service.Name == "" {
		return fmt.Errorf("service name is required")
	}

	if len(service.Name) > 100 {
		return fmt.Errorf("service name must be 100 characters or less")
	}

	if service.Type == "" {
		return fmt.Errorf("service type is required")
	}

	// Validate endpoints
	for i, endpoint := range service.Endpoints {
		if endpoint.Name == "" {
			return fmt.Errorf("endpoint %d name is required", i)
		}
		if endpoint.URL == "" {
			return fmt.Errorf("endpoint %d URL is required", i)
		}
	}

	return nil
}

// validateRoute validates a route configuration
func (hub *DefaultSystemIntegrationHub) validateRoute(route *APIRoute) error {
	if route.Name == "" {
		return fmt.Errorf("route name is required")
	}

	if route.Path == "" {
		return fmt.Errorf("route path is required")
	}

	if route.Method == "" {
		return fmt.Errorf("route method is required")
	}

	if route.ServiceID == "" {
		return fmt.Errorf("service ID is required")
	}

	// Validate that service exists
	if _, exists := hub.services[route.ServiceID]; !exists {
		return fmt.Errorf("service not found: %s", route.ServiceID)
	}

	return nil
}

// Filter matching methods

// matchesServiceFilter checks if a service matches the given filter
func (hub *DefaultSystemIntegrationHub) matchesServiceFilter(service *Service, filter *ServiceFilter) bool {
	if filter == nil {
		return true
	}

	// Type filter
	if filter.Type != "" && service.Type != filter.Type {
		return false
	}

	// Status filter
	if filter.Status != "" && service.Status != filter.Status {
		return false
	}

	// Tags filter
	if len(filter.Tags) > 0 {
		hasTag := false
		for _, filterTag := range filter.Tags {
			for _, serviceTag := range service.Tags {
				if serviceTag == filterTag {
					hasTag = true
					break
				}
			}
			if hasTag {
				break
			}
		}
		if !hasTag {
			return false
		}
	}

	// Search filter
	if filter.Search != "" {
		searchLower := strings.ToLower(filter.Search)
		if !(strings.Contains(strings.ToLower(service.Name), searchLower) ||
			strings.Contains(strings.ToLower(service.Description), searchLower) ||
			strings.Contains(strings.ToLower(string(service.Type)), searchLower)) {
			return false
		}
	}

	return true
}

// matchesRouteFilter checks if a route matches the given filter
func (hub *DefaultSystemIntegrationHub) matchesRouteFilter(route *APIRoute, filter *RouteFilter) bool {
	if filter == nil {
		return true
	}

	// Service ID filter
	if filter.ServiceID != "" && route.ServiceID != filter.ServiceID {
		return false
	}

	// Method filter
	if filter.Method != "" && route.Method != filter.Method {
		return false
	}

	// Status filter
	if filter.Status != "" && route.Status != filter.Status {
		return false
	}

	// Search filter
	if filter.Search != "" {
		searchLower := strings.ToLower(filter.Search)
		if !(strings.Contains(strings.ToLower(route.Name), searchLower) ||
			strings.Contains(strings.ToLower(route.Path), searchLower)) {
			return false
		}
	}

	return true
}

// matchesEventFilter checks if an event matches the given filter
func (hub *DefaultSystemIntegrationHub) matchesEventFilter(event *SystemEvent, filter *EventFilter) bool {
	if filter == nil {
		return true
	}

	// Type filter
	if filter.Type != "" && event.Type != filter.Type {
		return false
	}

	// Source filter
	if filter.Source != "" && event.Source != filter.Source {
		return false
	}

	// Target filter
	if filter.Target != "" && event.Target != filter.Target {
		return false
	}

	// Priority filter
	if filter.Priority != "" && event.Priority != filter.Priority {
		return false
	}

	// Time range filters
	if filter.Since != nil && event.Timestamp.Before(*filter.Since) {
		return false
	}
	if filter.Until != nil && event.Timestamp.After(*filter.Until) {
		return false
	}

	// Search filter
	if filter.Search != "" {
		searchLower := strings.ToLower(filter.Search)
		dataStr := fmt.Sprintf("%v", event.Data)
		if !(strings.Contains(strings.ToLower(event.Type), searchLower) ||
			strings.Contains(strings.ToLower(event.Source), searchLower) ||
			strings.Contains(strings.ToLower(dataStr), searchLower)) {
			return false
		}
	}

	return true
}

// matchesDiscoveryCriteria checks if a service matches discovery criteria
func (hub *DefaultSystemIntegrationHub) matchesDiscoveryCriteria(service *Service, criteria *DiscoveryCriteria) bool {
	if criteria == nil {
		return true
	}

	// Type filter
	if criteria.Type != "" && service.Type != criteria.Type {
		return false
	}

	// Health status filter
	if criteria.HealthStatus != "" {
		if health, exists := hub.serviceHealth[service.ID]; exists {
			if health.Status != criteria.HealthStatus {
				return false
			}
		} else if criteria.HealthStatus != HealthStatusUnknown {
			return false
		}
	}

	// Version filter
	if criteria.Version != "" && service.Version != criteria.Version {
		return false
	}

	// Tags filter
	if len(criteria.Tags) > 0 {
		hasTag := false
		for _, criteriaTag := range criteria.Tags {
			for _, serviceTag := range service.Tags {
				if serviceTag == criteriaTag {
					hasTag = true
					break
				}
			}
			if hasTag {
				break
			}
		}
		if !hasTag {
			return false
		}
	}

	// Metadata filter
	if len(criteria.Metadata) > 0 {
		for key, value := range criteria.Metadata {
			if serviceValue, exists := service.Metadata[key]; !exists || serviceValue != value {
				return false
			}
		}
	}

	return true
}

// Utility methods

// removeFromServiceIndex removes an item from a service index slice
func (hub *DefaultSystemIntegrationHub) removeFromServiceIndex(slice []string, item string) []string {
	for i, v := range slice {
		if v == item {
			return append(slice[:i], slice[i+1:]...)
		}
	}
	return slice
}

// removeFromRouteIndex removes an item from a route index slice
func (hub *DefaultSystemIntegrationHub) removeFromRouteIndex(slice []string, item string) []string {
	for i, v := range slice {
		if v == item {
			return append(slice[:i], slice[i+1:]...)
		}
	}
	return slice
}

// startHealthCheckLoop starts the health check background loop
func (hub *DefaultSystemIntegrationHub) startHealthCheckLoop(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		hub.performHealthChecks()
	}
}

// performHealthChecks performs health checks on all registered services
func (hub *DefaultSystemIntegrationHub) performHealthChecks() {
	hub.mu.RLock()
	var services []*Service
	for _, service := range hub.services {
		if service.Status == ServiceStatusHealthy || service.Status == ServiceStatusDegraded {
			services = append(services, service)
		}
	}
	hub.mu.RUnlock()

	for _, service := range services {
		hub.performServiceHealthCheck(service)
	}

	// Update system health
	hub.updateSystemHealth()
}

// performServiceHealthCheck performs a health check on a specific service
func (hub *DefaultSystemIntegrationHub) performServiceHealthCheck(service *Service) {
	start := time.Now()

	var checks []*HealthCheck
	overallStatus := HealthStatusHealthy
	message := "All checks passed"

	// Check each endpoint
	for _, endpoint := range service.Endpoints {
		check := &HealthCheck{
			Name:      fmt.Sprintf("endpoint_%s", endpoint.Name),
			Status:    HealthStatusHealthy,
			Message:   "Endpoint is healthy",
			Duration:  time.Since(start),
			Timestamp: time.Now(),
		}

		// Simple health check simulation
		// In a real implementation, this would make actual HTTP requests
		if endpoint.Health != nil && endpoint.Health.Status != HealthStatusHealthy {
			check.Status = endpoint.Health.Status
			check.Message = endpoint.Health.Message
			if check.Status == HealthStatusUnhealthy {
				overallStatus = HealthStatusUnhealthy
				message = "One or more endpoints are unhealthy"
			} else if check.Status == HealthStatusDegraded && overallStatus == HealthStatusHealthy {
				overallStatus = HealthStatusDegraded
				message = "One or more endpoints are degraded"
			}
		}

		checks = append(checks, check)
	}

	// Update service health
	health := &ServiceHealth{
		ServiceID: service.ID,
		Status:    overallStatus,
		Message:   message,
		LastCheck: time.Now(),
		Checks:    checks,
	}

	// Calculate uptime
	if existingHealth, exists := hub.serviceHealth[service.ID]; exists {
		if overallStatus == HealthStatusHealthy {
			health.Uptime = existingHealth.Uptime + time.Since(existingHealth.LastCheck)
		} else {
			health.Uptime = existingHealth.Uptime
		}
	}

	hub.UpdateServiceHealth(service.ID, health)
}

// updateSystemHealth updates the overall system health
func (hub *DefaultSystemIntegrationHub) updateSystemHealth() {
	hub.mu.Lock()
	defer hub.mu.Unlock()

	var healthyServices, degradedServices, unhealthyServices int
	var serviceHealthList []*ServiceHealth

	for _, health := range hub.serviceHealth {
		serviceHealthList = append(serviceHealthList, health)
		switch health.Status {
		case HealthStatusHealthy:
			healthyServices++
		case HealthStatusDegraded:
			degradedServices++
		case HealthStatusUnhealthy:
			unhealthyServices++
		}
	}

	// Determine overall system status
	var systemStatus HealthStatus
	var systemMessage string

	if unhealthyServices > 0 {
		systemStatus = HealthStatusUnhealthy
		systemMessage = fmt.Sprintf("%d services unhealthy, %d degraded, %d healthy",
			unhealthyServices, degradedServices, healthyServices)
	} else if degradedServices > 0 {
		systemStatus = HealthStatusDegraded
		systemMessage = fmt.Sprintf("%d services degraded, %d healthy",
			degradedServices, healthyServices)
	} else {
		systemStatus = HealthStatusHealthy
		systemMessage = fmt.Sprintf("All %d services healthy", healthyServices)
	}

	// Calculate system uptime
	now := time.Now()
	if hub.systemHealth.LastCheck.IsZero() {
		hub.systemHealth.Uptime = 0
	} else {
		if systemStatus == HealthStatusHealthy {
			hub.systemHealth.Uptime += now.Sub(hub.systemHealth.LastCheck)
		}
	}

	hub.systemHealth.Status = systemStatus
	hub.systemHealth.Message = systemMessage
	hub.systemHealth.Services = serviceHealthList
	hub.systemHealth.LastCheck = now
}

// startMetricsCollectionLoop starts the metrics collection background loop
func (hub *DefaultSystemIntegrationHub) startMetricsCollectionLoop(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		hub.collectMetrics()
	}
}

// collectMetrics collects metrics from all services
func (hub *DefaultSystemIntegrationHub) collectMetrics() {
	// Implementation would collect actual metrics from services
	// For now, this is a placeholder
	hub.logger.Debug("Collecting system metrics")
}

// matchesAlertFilter checks if an alert matches the given filter
func (hub *DefaultSystemIntegrationHub) matchesAlertFilter(alert *Alert, filter *AlertFilter) bool {
	if filter == nil {
		return true
	}

	// Type filter
	if filter.Type != "" && alert.Type != filter.Type {
		return false
	}

	// Severity filter
	if filter.Severity != "" && alert.Severity != filter.Severity {
		return false
	}

	// Status filter
	if filter.Status != "" && alert.Status != filter.Status {
		return false
	}

	// Source filter
	if filter.Source != "" && alert.Source != filter.Source {
		return false
	}

	// Time range filters
	if filter.Since != nil && alert.CreatedAt.Before(*filter.Since) {
		return false
	}
	if filter.Until != nil && alert.CreatedAt.After(*filter.Until) {
		return false
	}

	// Search filter
	if filter.Search != "" {
		searchLower := strings.ToLower(filter.Search)
		if !(strings.Contains(strings.ToLower(alert.Name), searchLower) ||
			strings.Contains(strings.ToLower(alert.Description), searchLower) ||
			strings.Contains(strings.ToLower(string(alert.Type)), searchLower)) {
			return false
		}
	}

	return true
}
