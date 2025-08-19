package main

import (
	"fmt"
	"log"
	"time"

	"github.com/aios/aios/pkg/integrations"
	"github.com/aios/aios/pkg/integrations/adapters"
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

	fmt.Println("🔗 AIOS Integration APIs Demo")
	fmt.Println("=============================")

	// Run the comprehensive demo
	if err := runIntegrationAPIsDemo(logger); err != nil {
		log.Fatalf("Demo failed: %v", err)
	}

	fmt.Println("\n✅ Integration APIs Demo completed successfully!")
}

func runIntegrationAPIsDemo(logger *logrus.Logger) error {
	// Step 1: Create Integration Engine
	fmt.Println("\n1. Creating Integration Engine...")
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
	fmt.Println("✓ Integration Engine created successfully")

	// Step 2: Register Integration Adapters
	fmt.Println("\n2. Registering Integration Adapters...")

	// Register GitHub adapter
	githubAdapter := adapters.NewGitHubAdapter()
	err := integrationEngine.RegisterAdapter(githubAdapter)
	if err != nil {
		return fmt.Errorf("failed to register GitHub adapter: %w", err)
	}
	fmt.Printf("   ✓ GitHub Adapter registered: %s v%s\n",
		githubAdapter.GetName(), githubAdapter.GetVersion())

	// Register Slack adapter
	slackAdapter := adapters.NewSlackAdapter()
	err = integrationEngine.RegisterAdapter(slackAdapter)
	if err != nil {
		return fmt.Errorf("failed to register Slack adapter: %w", err)
	}
	fmt.Printf("   ✓ Slack Adapter registered: %s v%s\n",
		slackAdapter.GetName(), slackAdapter.GetVersion())

	// List registered adapters
	adapterTypes := integrationEngine.ListAdapters()
	fmt.Printf("   ✓ Total adapters registered: %d\n", len(adapterTypes))
	for _, adapterType := range adapterTypes {
		adapter, _ := integrationEngine.GetAdapter(adapterType)
		fmt.Printf("     - %s: %s\n", adapterType, adapter.GetDescription())
	}

	// Step 3: Create GitHub Integration
	fmt.Println("\n3. Creating GitHub Integration...")

	githubIntegration := &integrations.Integration{
		Name:        "AIOS GitHub Integration",
		Description: "Integration with GitHub for repository management and CI/CD",
		Type:        "github",
		Provider:    "GitHub",
		Status:      integrations.IntegrationStatusConfiguring,
		Config: &integrations.IntegrationConfig{
			BaseURL:    "https://api.github.com",
			APIVersion: "v3",
			Timeout:    30 * time.Second,
			RetryPolicy: &integrations.RetryPolicy{
				MaxRetries:    3,
				InitialDelay:  1 * time.Second,
				MaxDelay:      30 * time.Second,
				BackoffFactor: 2.0,
			},
			RateLimit: &integrations.RateLimit{
				RequestsPerSecond: 5000, // GitHub's rate limit
				BurstSize:         100,
				WindowSize:        time.Hour,
			},
			Custom: map[string]interface{}{
				"token":        "ghp_example_token_12345", // Demo token
				"organization": "aios-org",
				"repository":   "aios",
			},
		},
		Credentials: &integrations.IntegrationCredentials{
			Type:  integrations.CredentialTypeBearer,
			Token: "ghp_example_token_12345", // Demo token
		},
		Settings: &integrations.IntegrationSettings{
			AutoSync:        true,
			SyncInterval:    15 * time.Minute,
			EnableWebhooks:  true,
			EnableEvents:    true,
			LogLevel:        "info",
			NotifyOnError:   true,
			ErrorRecipients: []string{"devops@company.com"},
		},
		CreatedBy: "admin@company.com",
	}

	createdGitHubIntegration, err := integrationEngine.CreateIntegration(githubIntegration)
	if err != nil {
		return fmt.Errorf("failed to create GitHub integration: %w", err)
	}

	fmt.Printf("   ✓ GitHub Integration Created: %s (ID: %s)\n",
		createdGitHubIntegration.Name, createdGitHubIntegration.ID)
	fmt.Printf("     - Type: %s\n", createdGitHubIntegration.Type)
	fmt.Printf("     - Provider: %s\n", createdGitHubIntegration.Provider)
	fmt.Printf("     - Status: %s\n", createdGitHubIntegration.Status)
	fmt.Printf("     - Auto Sync: %t\n", createdGitHubIntegration.Settings.AutoSync)
	fmt.Printf("     - Webhooks Enabled: %t\n", createdGitHubIntegration.Settings.EnableWebhooks)

	// Step 4: Create Slack Integration
	fmt.Println("\n4. Creating Slack Integration...")

	slackIntegration := &integrations.Integration{
		Name:        "AIOS Slack Integration",
		Description: "Integration with Slack for team notifications and communication",
		Type:        "slack",
		Provider:    "Slack",
		Status:      integrations.IntegrationStatusConfiguring,
		Config: &integrations.IntegrationConfig{
			BaseURL: "https://slack.com/api",
			Timeout: 30 * time.Second,
			RetryPolicy: &integrations.RetryPolicy{
				MaxRetries:    3,
				InitialDelay:  1 * time.Second,
				MaxDelay:      30 * time.Second,
				BackoffFactor: 2.0,
			},
			Custom: map[string]interface{}{
				"token":           "xoxb-example-slack-token", // Demo token
				"workspace":       "aios-workspace",
				"default_channel": "#general",
			},
		},
		Credentials: &integrations.IntegrationCredentials{
			Type:  integrations.CredentialTypeBearer,
			Token: "xoxb-example-slack-token", // Demo token
		},
		Settings: &integrations.IntegrationSettings{
			AutoSync:        false,
			EnableWebhooks:  true,
			EnableEvents:    true,
			LogLevel:        "info",
			NotifyOnError:   true,
			ErrorRecipients: []string{"team@company.com"},
		},
		CreatedBy: "admin@company.com",
	}

	createdSlackIntegration, err := integrationEngine.CreateIntegration(slackIntegration)
	if err != nil {
		return fmt.Errorf("failed to create Slack integration: %w", err)
	}

	fmt.Printf("   ✓ Slack Integration Created: %s (ID: %s)\n",
		createdSlackIntegration.Name, createdSlackIntegration.ID)
	fmt.Printf("     - Type: %s\n", createdSlackIntegration.Type)
	fmt.Printf("     - Provider: %s\n", createdSlackIntegration.Provider)
	fmt.Printf("     - Status: %s\n", createdSlackIntegration.Status)
	fmt.Printf("     - Webhooks Enabled: %t\n", createdSlackIntegration.Settings.EnableWebhooks)

	// Step 5: Create Webhooks
	fmt.Println("\n5. Creating Integration Webhooks...")

	// GitHub webhook
	githubWebhook := &integrations.Webhook{
		IntegrationID: createdGitHubIntegration.ID,
		Name:          "GitHub Repository Events",
		URL:           "https://aios.company.com/webhooks/github",
		Method:        "POST",
		Events:        []string{"push", "pull_request", "issues", "release"},
		Status:        integrations.WebhookStatusActive,
		Headers: map[string]string{
			"User-Agent": "AIOS-Integration/1.0",
		},
		Secret: "github_webhook_secret_123",
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

	createdGitHubWebhook, err := integrationEngine.CreateWebhook(githubWebhook)
	if err != nil {
		return fmt.Errorf("failed to create GitHub webhook: %w", err)
	}

	fmt.Printf("   ✓ GitHub Webhook Created: %s (ID: %s)\n",
		createdGitHubWebhook.Name, createdGitHubWebhook.ID)
	fmt.Printf("     - URL: %s\n", createdGitHubWebhook.URL)
	fmt.Printf("     - Events: %v\n", createdGitHubWebhook.Events)
	fmt.Printf("     - Status: %s\n", createdGitHubWebhook.Status)

	// Slack webhook
	slackWebhook := &integrations.Webhook{
		IntegrationID: createdSlackIntegration.ID,
		Name:          "Slack Events API",
		URL:           "https://aios.company.com/webhooks/slack",
		Method:        "POST",
		Events:        []string{"message", "app_mention", "team_join"},
		Status:        integrations.WebhookStatusActive,
		Headers: map[string]string{
			"User-Agent": "AIOS-Integration/1.0",
		},
		Config: &integrations.WebhookConfig{
			Timeout:         30 * time.Second,
			ContentType:     "application/json",
			SignatureHeader: "X-Slack-Signature",
		},
	}

	createdSlackWebhook, err := integrationEngine.CreateWebhook(slackWebhook)
	if err != nil {
		return fmt.Errorf("failed to create Slack webhook: %w", err)
	}

	fmt.Printf("   ✓ Slack Webhook Created: %s (ID: %s)\n",
		createdSlackWebhook.Name, createdSlackWebhook.ID)
	fmt.Printf("     - URL: %s\n", createdSlackWebhook.URL)
	fmt.Printf("     - Events: %v\n", createdSlackWebhook.Events)

	// Step 6: Test Integrations
	fmt.Println("\n6. Testing Integration Connections...")

	// Test GitHub integration
	githubTestResult, err := integrationEngine.TestIntegration(createdGitHubIntegration.ID)
	if err != nil {
		return fmt.Errorf("failed to test GitHub integration: %w", err)
	}

	fmt.Printf("   ✓ GitHub Integration Test:\n")
	fmt.Printf("     - Success: %t\n", githubTestResult.Success)
	fmt.Printf("     - Message: %s\n", githubTestResult.Message)
	fmt.Printf("     - Duration: %s\n", githubTestResult.Duration)
	fmt.Printf("     - Capabilities: %v\n", githubTestResult.Capabilities)

	// Test Slack integration
	slackTestResult, err := integrationEngine.TestIntegration(createdSlackIntegration.ID)
	if err != nil {
		return fmt.Errorf("failed to test Slack integration: %w", err)
	}

	fmt.Printf("   ✓ Slack Integration Test:\n")
	fmt.Printf("     - Success: %t\n", slackTestResult.Success)
	fmt.Printf("     - Message: %s\n", slackTestResult.Message)
	fmt.Printf("     - Duration: %s\n", slackTestResult.Duration)
	fmt.Printf("     - Capabilities: %v\n", slackTestResult.Capabilities)

	// Step 7: Integration Operations
	fmt.Println("\n7. Performing Integration Operations...")

	// Simulate GitHub operations (these would normally connect to real APIs)
	fmt.Printf("   ✓ GitHub Operations (simulated):\n")
	githubOps := []string{"list_repositories", "list_issues", "list_pull_requests"}
	for _, op := range githubOps {
		fmt.Printf("     - %s: Available\n", op)
	}

	// Simulate Slack operations
	fmt.Printf("   ✓ Slack Operations (simulated):\n")
	slackOps := []string{"send_message", "list_channels", "get_team_info"}
	for _, op := range slackOps {
		fmt.Printf("     - %s: Available\n", op)
	}

	// Step 8: Integration Management
	fmt.Println("\n8. Managing Integrations...")

	// List all integrations
	allIntegrations, err := integrationEngine.ListIntegrations(&integrations.IntegrationFilter{
		Limit: 10,
	})
	if err != nil {
		return fmt.Errorf("failed to list integrations: %w", err)
	}

	fmt.Printf("   ✓ Active Integrations (%d total):\n", len(allIntegrations))
	for _, integration := range allIntegrations {
		fmt.Printf("     - %s (%s): %s\n",
			integration.Name, integration.Type, integration.Status)
	}

	// List all webhooks
	allWebhooks, err := integrationEngine.ListWebhooks(&integrations.WebhookFilter{
		Limit: 10,
	})
	if err != nil {
		return fmt.Errorf("failed to list webhooks: %w", err)
	}

	fmt.Printf("   ✓ Active Webhooks (%d total):\n", len(allWebhooks))
	for _, webhook := range allWebhooks {
		fmt.Printf("     - %s: %s (%s)\n",
			webhook.Name, webhook.URL, webhook.Status)
	}

	// Step 9: Health Monitoring
	fmt.Println("\n9. Monitoring Integration Health...")

	// Get GitHub integration health
	githubHealth, err := integrationEngine.GetIntegrationHealth(createdGitHubIntegration.ID)
	if err != nil {
		return fmt.Errorf("failed to get GitHub integration health: %w", err)
	}

	fmt.Printf("   ✓ GitHub Integration Health:\n")
	fmt.Printf("     - Status: %s\n", githubHealth.Status)
	fmt.Printf("     - Message: %s\n", githubHealth.Message)
	fmt.Printf("     - Last Check: %s\n", githubHealth.LastCheck.Format("15:04:05"))
	fmt.Printf("     - Uptime: %s\n", githubHealth.Uptime)
	fmt.Printf("     - Health Checks: %d\n", len(githubHealth.Checks))

	// Get Slack integration health
	slackHealth, err := integrationEngine.GetIntegrationHealth(createdSlackIntegration.ID)
	if err != nil {
		return fmt.Errorf("failed to get Slack integration health: %w", err)
	}

	fmt.Printf("   ✓ Slack Integration Health:\n")
	fmt.Printf("     - Status: %s\n", slackHealth.Status)
	fmt.Printf("     - Message: %s\n", slackHealth.Message)
	fmt.Printf("     - Last Check: %s\n", slackHealth.LastCheck.Format("15:04:05"))
	fmt.Printf("     - Uptime: %s\n", slackHealth.Uptime)

	// Step 10: Integration Analytics
	fmt.Println("\n10. Integration Analytics and Metrics...")

	// Get integration metrics
	timeRange := &integrations.TimeRange{
		Start: time.Now().Add(-24 * time.Hour),
		End:   time.Now(),
	}

	githubMetrics, err := integrationEngine.GetIntegrationMetrics(createdGitHubIntegration.ID, timeRange)
	if err != nil {
		return fmt.Errorf("failed to get GitHub metrics: %w", err)
	}

	fmt.Printf("   ✓ GitHub Integration Metrics (24h):\n")
	fmt.Printf("     - Total Requests: %d\n", githubMetrics.RequestCount)
	fmt.Printf("     - Success Rate: %.1f%%\n", githubMetrics.SuccessRate*100)
	fmt.Printf("     - Average Latency: %s\n", githubMetrics.AverageLatency)
	fmt.Printf("     - Throughput: %.1f req/sec\n", githubMetrics.ThroughputPerSec)

	slackMetrics, err := integrationEngine.GetIntegrationMetrics(createdSlackIntegration.ID, timeRange)
	if err != nil {
		return fmt.Errorf("failed to get Slack metrics: %w", err)
	}

	fmt.Printf("   ✓ Slack Integration Metrics (24h):\n")
	fmt.Printf("     - Total Requests: %d\n", slackMetrics.RequestCount)
	fmt.Printf("     - Success Rate: %.1f%%\n", slackMetrics.SuccessRate*100)
	fmt.Printf("     - Average Latency: %s\n", slackMetrics.AverageLatency)
	fmt.Printf("     - Throughput: %.1f req/sec\n", slackMetrics.ThroughputPerSec)

	// Get integration logs
	githubLogs, err := integrationEngine.GetIntegrationLogs(createdGitHubIntegration.ID, &integrations.LogFilter{
		Level: integrations.LogLevelInfo,
		Limit: 5,
	})
	if err != nil {
		return fmt.Errorf("failed to get GitHub logs: %w", err)
	}

	fmt.Printf("   ✓ Recent GitHub Integration Logs (%d entries):\n", len(githubLogs))
	for i, logEntry := range githubLogs {
		if i >= 3 { // Show only first 3
			break
		}
		fmt.Printf("     %d. [%s] %s: %s\n",
			i+1, logEntry.Level, logEntry.Operation, logEntry.Message)
	}

	return nil
}
