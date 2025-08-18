# AIOS Integration APIs

## Overview

The AIOS Integration APIs provide a comprehensive platform for connecting with external services, tools, and platforms. Built with enterprise-grade capabilities, it offers pluggable adapters, secure authentication, webhook handling, rate limiting, and comprehensive monitoring to enable seamless integration with third-party services.

## 🏗️ Architecture

### Core Components

```
Integration APIs
├── Integration Engine (core orchestration, lifecycle management)
├── Adapter Framework (pluggable integration adapters)
├── Authentication System (OAuth, API keys, tokens)
├── Webhook Manager (incoming/outgoing webhooks)
├── Rate Limiting (request throttling, circuit breakers)
├── Event System (event publishing, subscription)
├── Configuration Manager (secure config management)
├── Health Monitor (integration health checks)
├── Metrics Collector (performance analytics)
└── Security Layer (credential management, encryption)
```

### Key Features

- **🔌 Pluggable Adapters**: Extensible adapter framework for any external service
- **🔐 Secure Authentication**: OAuth 2.0, API keys, tokens with secure credential storage
- **🪝 Webhook Management**: Bidirectional webhook support with signature verification
- **⚡ Rate Limiting**: Intelligent rate limiting and circuit breaker patterns
- **📊 Real-Time Monitoring**: Health checks, metrics, and performance analytics
- **🔄 Event-Driven Architecture**: Pub/sub event system for integration events
- **⚙️ Configuration Management**: Secure, validated configuration with hot reloading
- **🛡️ Enterprise Security**: Encryption, audit logging, and compliance features

## 🚀 Quick Start

### Basic Integration Setup

```go
package main

import (
    "time"
    "github.com/aios/aios/pkg/integrations"
    "github.com/aios/aios/pkg/integrations/adapters"
    "github.com/sirupsen/logrus"
)

func main() {
    logger := logrus.New()
    
    // Create integration engine
    config := &integrations.IntegrationEngineConfig{
        MaxIntegrations:     1000,
        MaxWebhooks:         5000,
        DefaultTimeout:      30 * time.Second,
        LogRetention:        7 * 24 * time.Hour,
        MetricsRetention:    30 * 24 * time.Hour,
        HealthCheckInterval: 5 * time.Minute,
        EnableMetrics:       true,
        EnableHealthChecks:  true,
    }
    
    integrationEngine := integrations.NewDefaultIntegrationEngine(config, logger)
    
    // Register adapters
    githubAdapter := adapters.NewGitHubAdapter()
    integrationEngine.RegisterAdapter(githubAdapter)
    
    slackAdapter := adapters.NewSlackAdapter()
    integrationEngine.RegisterAdapter(slackAdapter)
    
    // Create GitHub integration
    githubIntegration := &integrations.Integration{
        Name:        "GitHub Integration",
        Description: "Integration with GitHub for repository management",
        Type:        "github",
        Provider:    "GitHub",
        Config: &integrations.IntegrationConfig{
            BaseURL: "https://api.github.com",
            Custom: map[string]interface{}{
                "token":        "your-github-token",
                "organization": "your-org",
            },
        },
        Credentials: &integrations.IntegrationCredentials{
            Type:  integrations.CredentialTypeBearer,
            Token: "your-github-token",
        },
        Settings: &integrations.IntegrationSettings{
            AutoSync:       true,
            SyncInterval:   15 * time.Minute,
            EnableWebhooks: true,
            EnableEvents:   true,
        },
        CreatedBy: "admin@company.com",
    }
    
    createdIntegration, err := integrationEngine.CreateIntegration(githubIntegration)
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("Integration created: %s\n", createdIntegration.ID)
}
```

## 🔌 Integration Adapters

### Built-in Adapters

#### GitHub Adapter

Complete GitHub integration with repository management, issues, pull requests, and webhooks:

```go
// Register GitHub adapter
githubAdapter := adapters.NewGitHubAdapter()
err := integrationEngine.RegisterAdapter(githubAdapter)

// Create GitHub integration
githubIntegration := &integrations.Integration{
    Name:     "GitHub Repository Integration",
    Type:     "github",
    Provider: "GitHub",
    Config: &integrations.IntegrationConfig{
        BaseURL: "https://api.github.com",
        Custom: map[string]interface{}{
            "token":        "ghp_your_token_here",
            "organization": "your-org",
            "repository":   "your-repo",
        },
    },
    Credentials: &integrations.IntegrationCredentials{
        Type:  integrations.CredentialTypeBearer,
        Token: "ghp_your_token_here",
    },
}

createdIntegration, err := integrationEngine.CreateIntegration(githubIntegration)

// Test the integration
testResult, err := integrationEngine.TestIntegration(createdIntegration.ID)
fmt.Printf("GitHub test: %t - %s\n", testResult.Success, testResult.Message)

// Available operations:
// - list_repositories, get_repository, create_repository
// - list_issues, get_issue, create_issue, update_issue
// - list_pull_requests, get_pull_request, create_pull_request
// - list_commits, get_commit
// - create_webhook, list_webhooks, delete_webhook
```

#### Slack Adapter

Comprehensive Slack integration for messaging, channels, and team communication:

```go
// Register Slack adapter
slackAdapter := adapters.NewSlackAdapter()
err := integrationEngine.RegisterAdapter(slackAdapter)

// Create Slack integration
slackIntegration := &integrations.Integration{
    Name:     "Slack Team Communication",
    Type:     "slack",
    Provider: "Slack",
    Config: &integrations.IntegrationConfig{
        BaseURL: "https://slack.com/api",
        Custom: map[string]interface{}{
            "token":           "xoxb-your-bot-token",
            "workspace":       "your-workspace",
            "default_channel": "#general",
        },
    },
    Credentials: &integrations.IntegrationCredentials{
        Type:  integrations.CredentialTypeBearer,
        Token: "xoxb-your-bot-token",
    },
}

createdIntegration, err := integrationEngine.CreateIntegration(slackIntegration)

// Available operations:
// - send_message, list_channels, get_channel, create_channel
// - list_users, get_user, upload_file, get_team_info
// - post_ephemeral, update_message, delete_message
```

### Creating Custom Adapters

Implement the `IntegrationAdapter` interface to create custom adapters:

```go
type CustomAdapter struct {
    client    *http.Client
    baseURL   string
    connected bool
}

// Implement required methods
func (ca *CustomAdapter) GetType() string {
    return "custom_service"
}

func (ca *CustomAdapter) GetName() string {
    return "Custom Service Integration"
}

func (ca *CustomAdapter) GetDescription() string {
    return "Integration with custom external service"
}

func (ca *CustomAdapter) GetVersion() string {
    return "1.0.0"
}

func (ca *CustomAdapter) GetSupportedOperations() []string {
    return []string{"custom_operation", "another_operation"}
}

func (ca *CustomAdapter) GetConfigSchema() *integrations.ConfigSchema {
    return &integrations.ConfigSchema{
        Properties: map[string]*integrations.ConfigProperty{
            "api_key": {
                Type:        "string",
                Description: "API key for authentication",
                Sensitive:   true,
            },
            "base_url": {
                Type:        "string",
                Description: "Service base URL",
                Default:     "https://api.service.com",
            },
        },
        Required: []string{"api_key"},
    }
}

func (ca *CustomAdapter) ValidateConfig(config map[string]interface{}) error {
    if apiKey, exists := config["api_key"]; !exists || apiKey == "" {
        return fmt.Errorf("API key is required")
    }
    return nil
}

func (ca *CustomAdapter) Connect(ctx context.Context, config map[string]interface{}) error {
    // Implement connection logic
    ca.connected = true
    return nil
}

func (ca *CustomAdapter) Disconnect(ctx context.Context) error {
    ca.connected = false
    return nil
}

func (ca *CustomAdapter) IsConnected() bool {
    return ca.connected
}

func (ca *CustomAdapter) TestConnection(ctx context.Context) error {
    // Implement connection test
    return nil
}

func (ca *CustomAdapter) Execute(ctx context.Context, operation string, params map[string]interface{}) (*integrations.OperationResult, error) {
    // Implement operation execution
    switch operation {
    case "custom_operation":
        return ca.performCustomOperation(ctx, params)
    default:
        return &integrations.OperationResult{
            Success: false,
            Error:   fmt.Sprintf("unsupported operation: %s", operation),
        }, nil
    }
}

func (ca *CustomAdapter) SupportsWebhooks() bool {
    return true
}

func (ca *CustomAdapter) GetWebhookConfig() *integrations.WebhookConfig {
    return &integrations.WebhookConfig{
        Timeout:         30 * time.Second,
        ContentType:     "application/json",
        SignatureHeader: "X-Custom-Signature",
    }
}

func (ca *CustomAdapter) ProcessWebhookPayload(payload []byte, headers map[string]string) (*integrations.IntegrationEvent, error) {
    // Process webhook payload and return event
    var data map[string]interface{}
    if err := json.Unmarshal(payload, &data); err != nil {
        return nil, err
    }
    
    return &integrations.IntegrationEvent{
        Type:      "custom.event",
        Source:    "custom_service",
        Data:      data,
        Timestamp: time.Now(),
    }, nil
}

func (ca *CustomAdapter) GetHealth() *integrations.AdapterHealth {
    return &integrations.AdapterHealth{
        Status:    integrations.HealthStatusHealthy,
        Message:   "Custom adapter is healthy",
        Connected: ca.connected,
    }
}

func (ca *CustomAdapter) GetMetrics() *integrations.AdapterMetrics {
    return &integrations.AdapterMetrics{
        OperationCount:   make(map[string]int64),
        AverageLatency:   make(map[string]time.Duration),
        ErrorCount:       make(map[string]int64),
        LastOperation:    time.Now(),
        ConnectionUptime: time.Hour,
    }
}

// Register the custom adapter
customAdapter := &CustomAdapter{}
err := integrationEngine.RegisterAdapter(customAdapter)
```

## 🔐 Authentication and Security

### Credential Types

Support for multiple authentication methods:

```go
// API Key Authentication
credentials := &integrations.IntegrationCredentials{
    Type:   integrations.CredentialTypeAPIKey,
    APIKey: "your-api-key",
}

// Bearer Token Authentication
credentials := &integrations.IntegrationCredentials{
    Type:  integrations.CredentialTypeBearer,
    Token: "your-bearer-token",
}

// Basic Authentication
credentials := &integrations.IntegrationCredentials{
    Type:     integrations.CredentialTypeBasic,
    Username: "your-username",
    Password: "your-password",
}

// OAuth 2.0 Authentication
credentials := &integrations.IntegrationCredentials{
    Type: integrations.CredentialTypeOAuth2,
    OAuth: &integrations.OAuthCredentials{
        ClientID:     "your-client-id",
        ClientSecret: "your-client-secret",
        AccessToken:  "access-token",
        RefreshToken: "refresh-token",
        TokenType:    "Bearer",
        Scope:        []string{"read", "write"},
        ExpiresAt:    time.Now().Add(time.Hour),
    },
}

// Custom Authentication
credentials := &integrations.IntegrationCredentials{
    Type: integrations.CredentialTypeCustom,
    Custom: map[string]interface{}{
        "custom_header": "custom-value",
        "signature":     "hmac-signature",
    },
}
```

### Secure Configuration

```go
// Integration configuration with security settings
config := &integrations.IntegrationConfig{
    BaseURL:    "https://api.service.com",
    APIVersion: "v1",
    Timeout:    30 * time.Second,
    RetryPolicy: &integrations.RetryPolicy{
        MaxRetries:    3,
        InitialDelay:  1 * time.Second,
        MaxDelay:      30 * time.Second,
        BackoffFactor: 2.0,
        RetryableErrors: []string{"timeout", "rate_limit"},
    },
    RateLimit: &integrations.RateLimit{
        RequestsPerSecond: 100,
        BurstSize:         200,
        WindowSize:        time.Minute,
    },
    Custom: map[string]interface{}{
        "encryption_key": "encrypted-value",
        "webhook_secret": "webhook-signing-secret",
    },
}

// Validate configuration
err := integrationEngine.ValidateConfiguration(config)
if err != nil {
    fmt.Printf("Configuration validation failed: %v\n", err)
}

// Update integration configuration
err = integrationEngine.UpdateConfiguration(integrationID, config)
```

## 🪝 Webhook Management

### Creating Webhooks

```go
// Create incoming webhook for GitHub events
githubWebhook := &integrations.Webhook{
    IntegrationID: githubIntegrationID,
    Name:          "GitHub Repository Events",
    URL:           "https://your-app.com/webhooks/github",
    Method:        "POST",
    Events:        []string{"push", "pull_request", "issues", "release"},
    Status:        integrations.WebhookStatusActive,
    Headers: map[string]string{
        "User-Agent":    "AIOS-Integration/1.0",
        "Content-Type":  "application/json",
    },
    Secret: "your-webhook-secret",
    Config: &integrations.WebhookConfig{
        Timeout:         30 * time.Second,
        ContentType:     "application/json",
        SignatureHeader: "X-Hub-Signature-256",
        RetryPolicy: &integrations.RetryPolicy{
            MaxRetries:    3,
            InitialDelay:  2 * time.Second,
            MaxDelay:      60 * time.Second,
            BackoffFactor: 2.0,
        },
    },
}

createdWebhook, err := integrationEngine.CreateWebhook(githubWebhook)

// List webhooks for an integration
webhooks, err := integrationEngine.ListWebhooks(&integrations.WebhookFilter{
    IntegrationID: githubIntegrationID,
    Status:        integrations.WebhookStatusActive,
})

// Process incoming webhook
err = integrationEngine.ProcessWebhook(webhookID, payload, headers)
```

### Webhook Security

```go
// Webhook with signature verification
webhook := &integrations.Webhook{
    Secret: "your-signing-secret",
    Config: &integrations.WebhookConfig{
        SignatureHeader: "X-Hub-Signature-256",
        // Signature verification is handled automatically
    },
}

// Custom webhook validation
func validateWebhookSignature(payload []byte, signature string, secret string) bool {
    mac := hmac.New(sha256.New, []byte(secret))
    mac.Write(payload)
    expectedSignature := hex.EncodeToString(mac.Sum(nil))
    return hmac.Equal([]byte(signature), []byte(expectedSignature))
}
```

## 📊 Monitoring and Analytics

### Health Monitoring

```go
// Get integration health status
health, err := integrationEngine.GetIntegrationHealth(integrationID)
fmt.Printf("Health Status: %s\n", health.Status)
fmt.Printf("Message: %s\n", health.Message)
fmt.Printf("Last Check: %s\n", health.LastCheck)
fmt.Printf("Uptime: %s\n", health.Uptime)

// Health check details
for _, check := range health.Checks {
    fmt.Printf("Check %s: %s - %s\n", check.Name, check.Status, check.Message)
}

// Enable/disable integrations based on health
if health.Status == integrations.HealthStatusUnhealthy {
    err := integrationEngine.DisableIntegration(integrationID)
    if err != nil {
        fmt.Printf("Failed to disable unhealthy integration: %v\n", err)
    }
}
```

### Performance Metrics

```go
// Get integration metrics
timeRange := &integrations.TimeRange{
    Start: time.Now().Add(-24 * time.Hour),
    End:   time.Now(),
}

metrics, err := integrationEngine.GetIntegrationMetrics(integrationID, timeRange)
fmt.Printf("Metrics for last 24 hours:\n")
fmt.Printf("  Total Requests: %d\n", metrics.RequestCount)
fmt.Printf("  Success Rate: %.1f%%\n", metrics.SuccessRate*100)
fmt.Printf("  Average Latency: %s\n", metrics.AverageLatency)
fmt.Printf("  P95 Latency: %s\n", metrics.P95Latency)
fmt.Printf("  P99 Latency: %s\n", metrics.P99Latency)
fmt.Printf("  Throughput: %.1f req/sec\n", metrics.ThroughputPerSec)

// Error analysis
fmt.Printf("Errors by type:\n")
for errorType, count := range metrics.ErrorsByType {
    fmt.Printf("  %s: %d\n", errorType, count)
}

// Set up alerts based on metrics
if metrics.SuccessRate < 0.95 {
    // Send alert for low success rate
    fmt.Printf("ALERT: Success rate below 95%% for integration %s\n", integrationID)
}

if metrics.P95Latency > 5*time.Second {
    // Send alert for high latency
    fmt.Printf("ALERT: High latency detected for integration %s\n", integrationID)
}
```

### Logging and Audit

```go
// Get integration logs
logs, err := integrationEngine.GetIntegrationLogs(integrationID, &integrations.LogFilter{
    Level:  integrations.LogLevelError,
    Since:  &[]time.Time{time.Now().Add(-24 * time.Hour)}[0],
    Limit:  100,
})

fmt.Printf("Error logs from last 24 hours:\n")
for _, log := range logs {
    fmt.Printf("[%s] %s: %s\n", log.Level, log.Operation, log.Message)
    if log.Error != "" {
        fmt.Printf("  Error: %s\n", log.Error)
    }
    if log.Duration > 0 {
        fmt.Printf("  Duration: %s\n", log.Duration)
    }
}

// Search logs
searchLogs, err := integrationEngine.GetIntegrationLogs(integrationID, &integrations.LogFilter{
    Search: "authentication",
    Limit:  50,
})

// Filter by operation
operationLogs, err := integrationEngine.GetIntegrationLogs(integrationID, &integrations.LogFilter{
    Operation: "webhook_processing",
    Level:     integrations.LogLevelInfo,
    Limit:     25,
})
```

## 🔄 Event System

### Event Publishing and Subscription

```go
// Subscribe to integration events
type MyEventHandler struct{}

func (h *MyEventHandler) HandleEvent(ctx context.Context, event *integrations.IntegrationEvent) error {
    fmt.Printf("Received event: %s from %s\n", event.Type, event.Source)
    
    switch event.Type {
    case "github.push":
        return h.handleGitHubPush(event)
    case "slack.message":
        return h.handleSlackMessage(event)
    default:
        fmt.Printf("Unhandled event type: %s\n", event.Type)
    }
    
    return nil
}

func (h *MyEventHandler) GetEventTypes() []string {
    return []string{"github.push", "slack.message"}
}

// Register event handler
handler := &MyEventHandler{}
err := integrationEngine.SubscribeToEvents("github.push", handler)
err = integrationEngine.SubscribeToEvents("slack.message", handler)

// Publish custom events
customEvent := &integrations.IntegrationEvent{
    Type:   "custom.deployment",
    Source: "ci_cd_system",
    Data: map[string]interface{}{
        "environment": "production",
        "version":     "v1.2.3",
        "status":      "success",
    },
    Metadata: map[string]interface{}{
        "deployment_id": "deploy-123",
        "triggered_by":  "user@company.com",
    },
}

err = integrationEngine.PublishEvent(customEvent)
```

## 🧪 Testing

Run the comprehensive test suite:

```bash
# Run all integration tests
go test ./pkg/integrations/...

# Run with race detection
go test -race ./pkg/integrations/...

# Run integration tests with external services
go test -tags=integration ./pkg/integrations/...

# Run integration APIs example
go run examples/integration_apis_example.go
```

## 📖 Examples

See the complete example in `examples/integration_apis_example.go` for a comprehensive demonstration including:

- Integration engine setup and configuration
- GitHub and Slack adapter registration
- Integration creation with authentication
- Webhook management and processing
- Health monitoring and metrics collection
- Event handling and custom adapters
- Security and error handling patterns

## 🤝 Contributing

1. Follow established patterns and interfaces
2. Add comprehensive tests for new adapters and features
3. Update documentation for API changes
4. Ensure proper error handling and logging
5. Use OpenTelemetry for observability
6. Implement proper security and authentication

## 📄 License

This Integration APIs system is part of the AIOS project and follows the same licensing terms.
