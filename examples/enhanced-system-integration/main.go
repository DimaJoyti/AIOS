package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aios/aios/pkg/systemintegration"
	"github.com/sirupsen/logrus"
)

func main() {
	// Initialize logger
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)
	logger.SetFormatter(&logrus.TextFormatter{
		DisableTimestamp: true,
		ForceColors:      true,
	})

	fmt.Println("🔗 AIOS Enhanced System Integration Demo")
	fmt.Println("========================================")

	// Run the comprehensive demo
	if err := runEnhancedSystemIntegrationDemo(logger); err != nil {
		log.Fatalf("Demo failed: %v", err)
	}

	fmt.Println("\n✅ Enhanced System Integration Demo completed successfully!")
}

func runEnhancedSystemIntegrationDemo(logger *logrus.Logger) error {
	// Step 1: Create System Integration Hub
	fmt.Println("\n1. Creating System Integration Hub...")
	config := &systemintegration.SystemIntegrationHubConfig{
		MaxServices:         1000,
		MaxRoutes:           5000,
		MaxEvents:           100000,
		MaxDataFlows:        1000,
		DefaultTimeout:      30 * time.Second,
		EventRetention:      24 * time.Hour,
		MetricsRetention:    7 * 24 * time.Hour,
		HealthCheckInterval: 30 * time.Second,
		EnableMetrics:       true,
		EnableHealthChecks:  true,
		EnableSecurity:      true,
	}

	hub := systemintegration.NewDefaultSystemIntegrationHub(config, logger)
	fmt.Println("✓ System Integration Hub created successfully")

	// Step 2: Register Core AIOS Services
	fmt.Println("\n2. Registering Core AIOS Services...")

	// Register Agent Service
	agentService := &systemintegration.Service{
		Name:        "AIOS Agent Service",
		Description: "Core AI agent orchestration and management service",
		Type:        systemintegration.ServiceTypeAgent,
		Version:     "1.0.0",
		Status:      systemintegration.ServiceStatusHealthy,
		Endpoints: []*systemintegration.ServiceEndpoint{
			{
				ID:       "agent-api",
				Name:     "Agent API",
				URL:      "http://localhost:8080",
				Protocol: "http",
				Port:     8080,
				Path:     "/api/v1/agents",
				Method:   "GET",
				Auth: &systemintegration.EndpointAuth{
					Type:  systemintegration.AuthTypeBearer,
					Token: "agent-service-token",
				},
			},
			{
				ID:       "agent-websocket",
				Name:     "Agent WebSocket",
				URL:      "ws://localhost:8080/ws",
				Protocol: "websocket",
				Port:     8080,
				Path:     "/ws",
			},
		},
		Tags: []string{"core", "ai", "agent"},
		Metadata: map[string]interface{}{
			"max_agents":       100,
			"supported_models": []string{"gpt-4", "claude-3", "llama-2"},
			"capabilities":     []string{"reasoning", "planning", "execution"},
		},
	}

	createdAgentService, err := hub.RegisterService(agentService)
	if err != nil {
		return fmt.Errorf("failed to register agent service: %w", err)
	}

	fmt.Printf("   ✓ Agent Service registered: %s (ID: %s)\n",
		createdAgentService.Name, createdAgentService.ID)

	// Register Collaboration Service
	collaborationService := &systemintegration.Service{
		Name:        "AIOS Collaboration Service",
		Description: "Real-time collaboration and communication service",
		Type:        systemintegration.ServiceTypeCollaboration,
		Version:     "1.0.0",
		Status:      systemintegration.ServiceStatusHealthy,
		Endpoints: []*systemintegration.ServiceEndpoint{
			{
				ID:       "collab-api",
				Name:     "Collaboration API",
				URL:      "http://localhost:8081",
				Protocol: "http",
				Port:     8081,
				Path:     "/api/v1/collaboration",
				Method:   "GET",
			},
			{
				ID:       "collab-realtime",
				Name:     "Real-time Events",
				URL:      "ws://localhost:8081/realtime",
				Protocol: "websocket",
				Port:     8081,
				Path:     "/realtime",
			},
		},
		Dependencies: []string{createdAgentService.ID},
		Tags:         []string{"core", "collaboration", "realtime"},
		Metadata: map[string]interface{}{
			"max_rooms":        1000,
			"max_participants": 50,
			"features":         []string{"chat", "video", "screen_share", "whiteboard"},
		},
	}

	createdCollabService, err := hub.RegisterService(collaborationService)
	if err != nil {
		return fmt.Errorf("failed to register collaboration service: %w", err)
	}

	fmt.Printf("   ✓ Collaboration Service registered: %s (ID: %s)\n",
		createdCollabService.Name, createdCollabService.ID)

	// Register Integration Service
	integrationService := &systemintegration.Service{
		Name:        "AIOS Integration Service",
		Description: "External integrations and API management service",
		Type:        systemintegration.ServiceTypeIntegration,
		Version:     "1.0.0",
		Status:      systemintegration.ServiceStatusHealthy,
		Endpoints: []*systemintegration.ServiceEndpoint{
			{
				ID:       "integration-api",
				Name:     "Integration API",
				URL:      "http://localhost:8082",
				Protocol: "http",
				Port:     8082,
				Path:     "/api/v1/integrations",
				Method:   "GET",
			},
			{
				ID:       "webhook-handler",
				Name:     "Webhook Handler",
				URL:      "http://localhost:8082/webhooks",
				Protocol: "http",
				Port:     8082,
				Path:     "/webhooks",
				Method:   "POST",
			},
		},
		Tags: []string{"core", "integration", "external"},
		Metadata: map[string]interface{}{
			"supported_integrations": []string{"github", "slack", "jira", "salesforce"},
			"webhook_support":        true,
			"rate_limiting":          true,
		},
	}

	createdIntegrationService, err := hub.RegisterService(integrationService)
	if err != nil {
		return fmt.Errorf("failed to register integration service: %w", err)
	}

	fmt.Printf("   ✓ Integration Service registered: %s (ID: %s)\n",
		createdIntegrationService.Name, createdIntegrationService.ID)

	// Register Data Integration Service
	dataIntegrationService := &systemintegration.Service{
		Name:        "AIOS Data Integration Service",
		Description: "External data integration and processing service",
		Type:        systemintegration.ServiceTypeDataIntegration,
		Version:     "1.0.0",
		Status:      systemintegration.ServiceStatusHealthy,
		Endpoints: []*systemintegration.ServiceEndpoint{
			{
				ID:       "data-api",
				Name:     "Data Integration API",
				URL:      "http://localhost:8083",
				Protocol: "http",
				Port:     8083,
				Path:     "/api/v1/data",
				Method:   "GET",
			},
			{
				ID:       "pipeline-api",
				Name:     "Pipeline Management API",
				URL:      "http://localhost:8083/pipelines",
				Protocol: "http",
				Port:     8083,
				Path:     "/pipelines",
				Method:   "GET",
			},
		},
		Tags: []string{"core", "data", "etl"},
		Metadata: map[string]interface{}{
			"supported_sources": []string{"databases", "apis", "files", "streams"},
			"connectors":        []string{"postgresql", "mongodb", "elasticsearch", "kafka"},
			"transformations":   []string{"map", "filter", "aggregate", "enrich"},
		},
	}

	createdDataIntegrationService, err := hub.RegisterService(dataIntegrationService)
	if err != nil {
		return fmt.Errorf("failed to register data integration service: %w", err)
	}

	fmt.Printf("   ✓ Data Integration Service registered: %s (ID: %s)\n",
		createdDataIntegrationService.Name, createdDataIntegrationService.ID)

	// Step 3: Service Discovery
	fmt.Println("\n3. Service Discovery and Management...")

	// List all registered services
	allServices, err := hub.ListServices(&systemintegration.ServiceFilter{
		Limit: 10,
	})
	if err != nil {
		return fmt.Errorf("failed to list services: %w", err)
	}

	fmt.Printf("   ✓ Total Registered Services: %d\n", len(allServices))
	for _, service := range allServices {
		fmt.Printf("     - %s (%s): %s - %d endpoints\n",
			service.Name, service.Type, service.Status, len(service.Endpoints))
	}

	// Discover core services
	coreServices, err := hub.DiscoverServices(&systemintegration.DiscoveryCriteria{
		Tags: []string{"core"},
	})
	if err != nil {
		return fmt.Errorf("failed to discover core services: %w", err)
	}

	fmt.Printf("   ✓ Core Services Discovered: %d\n", len(coreServices))
	for _, service := range coreServices {
		fmt.Printf("     - %s: %s\n", service.Name, service.Type)
	}

	// Step 4: API Gateway Configuration
	fmt.Println("\n4. Configuring API Gateway Routes...")

	// Create route for agent service
	agentRoute := &systemintegration.APIRoute{
		Name:      "Agent Service Route",
		Path:      "/api/agents/*",
		Method:    "GET",
		ServiceID: createdAgentService.ID,
		Endpoint:  "/api/v1/agents",
		Config: &systemintegration.RouteConfig{
			Timeout: 30 * time.Second,
			RetryPolicy: &systemintegration.RetryPolicy{
				MaxRetries:    3,
				InitialDelay:  1 * time.Second,
				MaxDelay:      10 * time.Second,
				BackoffFactor: 2.0,
			},
			CircuitBreaker: &systemintegration.CircuitBreakerConfig{
				Enabled:          true,
				FailureThreshold: 5,
				RecoveryTimeout:  30 * time.Second,
			},
		},
		RateLimit: &systemintegration.RateLimit{
			RequestsPerSecond: 100,
			BurstSize:         200,
			WindowSize:        time.Minute,
		},
		Auth: &systemintegration.RouteAuth{
			Required: true,
			Types:    []systemintegration.AuthType{systemintegration.AuthTypeBearer},
			Scopes:   []string{"agents:read"},
		},
		Status: systemintegration.RouteStatusActive,
	}

	createdAgentRoute, err := hub.CreateRoute(agentRoute)
	if err != nil {
		return fmt.Errorf("failed to create agent route: %w", err)
	}

	fmt.Printf("   ✓ Agent Route created: %s -> %s\n",
		createdAgentRoute.Path, createdAgentRoute.Endpoint)

	// Create route for collaboration service
	collabRoute := &systemintegration.APIRoute{
		Name:      "Collaboration Service Route",
		Path:      "/api/collaboration/*",
		Method:    "GET",
		ServiceID: createdCollabService.ID,
		Endpoint:  "/api/v1/collaboration",
		Config: &systemintegration.RouteConfig{
			Timeout: 15 * time.Second,
			LoadBalancing: &systemintegration.LoadBalancingConfig{
				Strategy:    systemintegration.LoadBalancingRoundRobin,
				HealthCheck: true,
			},
		},
		RateLimit: &systemintegration.RateLimit{
			RequestsPerSecond: 200,
			BurstSize:         400,
			WindowSize:        time.Minute,
		},
		Status: systemintegration.RouteStatusActive,
	}

	createdCollabRoute, err := hub.CreateRoute(collabRoute)
	if err != nil {
		return fmt.Errorf("failed to create collaboration route: %w", err)
	}

	fmt.Printf("   ✓ Collaboration Route created: %s -> %s\n",
		createdCollabRoute.Path, createdCollabRoute.Endpoint)

	// List all routes
	allRoutes, err := hub.ListRoutes(&systemintegration.RouteFilter{
		Status: systemintegration.RouteStatusActive,
		Limit:  10,
	})
	if err != nil {
		return fmt.Errorf("failed to list routes: %w", err)
	}

	fmt.Printf("   ✓ Active Routes: %d\n", len(allRoutes))
	for _, route := range allRoutes {
		fmt.Printf("     - %s %s -> %s\n", route.Method, route.Path, route.Endpoint)
	}

	// Step 5: Event Bus and System Events
	fmt.Println("\n5. Event Bus and System Communication...")

	// Create a custom event handler
	eventHandler := &CustomEventHandler{
		ID:         "demo-handler",
		EventTypes: []string{"service.registered", "service.updated", "alert.created"},
	}

	// Subscribe to events
	subscription := &systemintegration.EventSubscription{
		EventTypes: []string{"service.registered", "service.updated", "alert.created"},
		Handler:    eventHandler,
		Config: &systemintegration.SubscriptionConfig{
			DeliveryMode: systemintegration.DeliveryModeAsync,
			RetryPolicy: &systemintegration.RetryPolicy{
				MaxRetries:    3,
				InitialDelay:  1 * time.Second,
				MaxDelay:      10 * time.Second,
				BackoffFactor: 2.0,
			},
		},
	}

	err = hub.SubscribeToEvents(subscription)
	if err != nil {
		return fmt.Errorf("failed to subscribe to events: %w", err)
	}

	fmt.Printf("   ✓ Subscribed to events: %v\n", subscription.EventTypes)

	// Publish a custom system event
	customEvent := &systemintegration.SystemEvent{
		Type:   "system.demo",
		Source: "demo_application",
		Data: map[string]interface{}{
			"message":   "Enhanced System Integration Demo",
			"timestamp": time.Now(),
			"services":  len(allServices),
			"routes":    len(allRoutes),
		},
		Priority: systemintegration.EventPriorityNormal,
	}

	err = hub.PublishEvent(customEvent)
	if err != nil {
		return fmt.Errorf("failed to publish custom event: %w", err)
	}

	fmt.Printf("   ✓ Published custom event: %s\n", customEvent.Type)

	// Get event history
	eventHistory, err := hub.GetEventHistory(&systemintegration.EventFilter{
		Type:  "service.registered",
		Limit: 5,
	})
	if err != nil {
		return fmt.Errorf("failed to get event history: %w", err)
	}

	fmt.Printf("   ✓ Recent Service Registration Events: %d\n", len(eventHistory))
	for i, event := range eventHistory {
		if i >= 3 { // Show only first 3
			break
		}
		fmt.Printf("     %d. %s from %s at %s\n",
			i+1, event.Type, event.Source, event.Timestamp.Format("15:04:05"))
	}

	// Step 6: System Health Monitoring
	fmt.Println("\n6. System Health and Monitoring...")

	// Get overall system health
	systemHealth, err := hub.GetSystemHealth()
	if err != nil {
		return fmt.Errorf("failed to get system health: %w", err)
	}

	fmt.Printf("   ✓ System Health Status: %s\n", systemHealth.Status)
	fmt.Printf("     - Message: %s\n", systemHealth.Message)
	fmt.Printf("     - Last Check: %s\n", systemHealth.LastCheck.Format("15:04:05"))
	fmt.Printf("     - Uptime: %s\n", systemHealth.Uptime)
	fmt.Printf("     - Services Monitored: %d\n", len(systemHealth.Services))

	// Get system metrics
	timeRange := &systemintegration.TimeRange{
		Start: time.Now().Add(-1 * time.Hour),
		End:   time.Now(),
	}

	systemMetrics, err := hub.GetSystemMetrics(timeRange)
	if err != nil {
		return fmt.Errorf("failed to get system metrics: %w", err)
	}

	fmt.Printf("   ✓ System Metrics (1 hour):\n")
	fmt.Printf("     - Total Requests: %d\n", systemMetrics.RequestCount)
	fmt.Printf("     - Success Rate: %.1f%%\n", systemMetrics.SuccessRate*100)
	fmt.Printf("     - Average Latency: %s\n", systemMetrics.AverageLatency)
	fmt.Printf("     - Throughput: %.1f req/sec\n", systemMetrics.ThroughputPerSec)
	fmt.Printf("     - CPU Usage: %.1f%%\n", systemMetrics.ResourceUsage.CPUUsage)
	fmt.Printf("     - Memory Usage: %.1f%%\n", systemMetrics.ResourceUsage.MemoryUsage)

	// Get service-specific metrics
	agentMetrics, err := hub.GetServiceMetrics(createdAgentService.ID, timeRange)
	if err != nil {
		return fmt.Errorf("failed to get agent service metrics: %w", err)
	}

	fmt.Printf("   ✓ Agent Service Metrics:\n")
	fmt.Printf("     - Requests: %d\n", agentMetrics.RequestCount)
	fmt.Printf("     - Success Rate: %.1f%%\n", agentMetrics.SuccessRate*100)
	fmt.Printf("     - Average Latency: %s\n", agentMetrics.AverageLatency)

	// Step 7: Security and Authentication
	fmt.Println("\n7. Security and Authentication...")

	// Test authentication
	authRequest := &systemintegration.AuthRequest{
		Token:      "demo-bearer-token",
		Method:     "GET",
		Path:       "/api/agents",
		RemoteAddr: "127.0.0.1",
		UserAgent:  "AIOS-Demo/1.0",
	}

	authResult, err := hub.AuthenticateRequest(authRequest)
	if err != nil {
		return fmt.Errorf("failed to authenticate request: %w", err)
	}

	fmt.Printf("   ✓ Authentication Result:\n")
	fmt.Printf("     - Success: %t\n", authResult.Success)
	fmt.Printf("     - User ID: %s\n", authResult.UserID)
	fmt.Printf("     - Username: %s\n", authResult.Username)
	fmt.Printf("     - Roles: %v\n", authResult.Roles)

	// Test authorization
	authzRequest := &systemintegration.AuthzRequest{
		UserID:   authResult.UserID,
		Resource: "agents",
		Action:   "read",
		Context: map[string]interface{}{
			"service": "agent_service",
			"method":  "GET",
		},
	}

	authzResult, err := hub.AuthorizeRequest(authzRequest)
	if err != nil {
		return fmt.Errorf("failed to authorize request: %w", err)
	}

	fmt.Printf("   ✓ Authorization Result:\n")
	fmt.Printf("     - Allowed: %t\n", authzResult.Allowed)
	fmt.Printf("     - Reason: %s\n", authzResult.Reason)

	// Step 8: Alerts and Notifications
	fmt.Println("\n8. Alerts and Notifications...")

	// Create a performance alert
	performanceAlert := &systemintegration.Alert{
		Name:        "High Latency Alert",
		Description: "Alert triggered when system latency exceeds threshold",
		Type:        systemintegration.AlertTypePerformance,
		Severity:    systemintegration.AlertSeverityMedium,
		Source:      "system_monitor",
		Target:      "system",
		Condition:   "avg_latency > 100ms",
		Threshold: &systemintegration.AlertThreshold{
			Metric:   "avg_latency",
			Operator: ">",
			Value:    100.0,
			Duration: 5 * time.Minute,
		},
		Actions: []*systemintegration.AlertAction{
			{
				Type:    systemintegration.AlertActionTypeEmail,
				Target:  "ops-team@company.com",
				Enabled: true,
			},
			{
				Type:    systemintegration.AlertActionTypeSlack,
				Target:  "#alerts",
				Enabled: true,
			},
		},
	}

	createdAlert, err := hub.CreateAlert(performanceAlert)
	if err != nil {
		return fmt.Errorf("failed to create alert: %w", err)
	}

	fmt.Printf("   ✓ Performance Alert created: %s (ID: %s)\n",
		createdAlert.Name, createdAlert.ID)
	fmt.Printf("     - Type: %s\n", createdAlert.Type)
	fmt.Printf("     - Severity: %s\n", createdAlert.Severity)
	fmt.Printf("     - Condition: %s\n", createdAlert.Condition)
	fmt.Printf("     - Actions: %d\n", len(createdAlert.Actions))

	// Get all alerts
	allAlerts, err := hub.GetAlerts(&systemintegration.AlertFilter{
		Status: systemintegration.AlertStatusActive,
		Limit:  10,
	})
	if err != nil {
		return fmt.Errorf("failed to get alerts: %w", err)
	}

	fmt.Printf("   ✓ Active Alerts: %d\n", len(allAlerts))
	for _, alert := range allAlerts {
		fmt.Printf("     - %s (%s): %s\n", alert.Name, alert.Severity, alert.Type)
	}

	// Step 9: Configuration Management
	fmt.Println("\n9. Configuration Management...")

	// Set system configuration
	systemConfig := &systemintegration.Configuration{
		Key:         "system.max_concurrent_requests",
		Value:       1000,
		Type:        systemintegration.ConfigurationTypeInteger,
		Description: "Maximum number of concurrent requests the system can handle",
		Sensitive:   false,
	}

	err = hub.SetConfiguration(systemConfig)
	if err != nil {
		return fmt.Errorf("failed to set configuration: %w", err)
	}

	fmt.Printf("   ✓ Configuration set: %s = %v\n", systemConfig.Key, systemConfig.Value)

	// Get configuration
	retrievedConfig, err := hub.GetConfiguration("system.max_concurrent_requests")
	if err != nil {
		return fmt.Errorf("failed to get configuration: %w", err)
	}

	fmt.Printf("   ✓ Configuration retrieved: %s = %v (%s)\n",
		retrievedConfig.Key, retrievedConfig.Value, retrievedConfig.Type)

	// Get service configuration
	agentServiceConfig, err := hub.GetServiceConfiguration(createdAgentService.ID)
	if err != nil {
		return fmt.Errorf("failed to get service configuration: %w", err)
	}

	fmt.Printf("   ✓ Agent Service Configuration:\n")
	fmt.Printf("     - Service ID: %s\n", agentServiceConfig.ServiceID)
	fmt.Printf("     - Version: %s\n", agentServiceConfig.Version)
	fmt.Printf("     - Config Entries: %d\n", len(agentServiceConfig.Config))

	return nil
}

// CustomEventHandler implements the EventHandler interface
type CustomEventHandler struct {
	ID         string
	EventTypes []string
}

func (h *CustomEventHandler) HandleEvent(ctx context.Context, event *systemintegration.SystemEvent) error {
	fmt.Printf("     📧 Event received: %s from %s\n", event.Type, event.Source)
	return nil
}

func (h *CustomEventHandler) GetEventTypes() []string {
	return h.EventTypes
}

func (h *CustomEventHandler) GetHandlerID() string {
	return h.ID
}
