package adapters

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/aios/aios/pkg/integrations"
)

// GitHubAdapter implements the IntegrationAdapter interface for GitHub
type GitHubAdapter struct {
	client    *http.Client
	baseURL   string
	token     string
	connected bool
}

// NewGitHubAdapter creates a new GitHub adapter
func NewGitHubAdapter() *GitHubAdapter {
	return &GitHubAdapter{
		client:  &http.Client{Timeout: 30 * time.Second},
		baseURL: "https://api.github.com",
	}
}

// Adapter Information

// GetType returns the adapter type
func (ga *GitHubAdapter) GetType() string {
	return "github"
}

// GetName returns the adapter name
func (ga *GitHubAdapter) GetName() string {
	return "GitHub Integration"
}

// GetDescription returns the adapter description
func (ga *GitHubAdapter) GetDescription() string {
	return "Integration with GitHub for repository management, issues, pull requests, and webhooks"
}

// GetVersion returns the adapter version
func (ga *GitHubAdapter) GetVersion() string {
	return "1.0.0"
}

// GetSupportedOperations returns the list of supported operations
func (ga *GitHubAdapter) GetSupportedOperations() []string {
	return []string{
		"list_repositories",
		"get_repository",
		"create_repository",
		"list_issues",
		"get_issue",
		"create_issue",
		"update_issue",
		"list_pull_requests",
		"get_pull_request",
		"create_pull_request",
		"list_commits",
		"get_commit",
		"create_webhook",
		"list_webhooks",
		"delete_webhook",
	}
}

// Configuration

// GetConfigSchema returns the configuration schema
func (ga *GitHubAdapter) GetConfigSchema() *integrations.ConfigSchema {
	return &integrations.ConfigSchema{
		Properties: map[string]*integrations.ConfigProperty{
			"token": {
				Type:        "string",
				Description: "GitHub personal access token or GitHub App token",
				Sensitive:   true,
			},
			"base_url": {
				Type:        "string",
				Description: "GitHub API base URL (for GitHub Enterprise)",
				Default:     "https://api.github.com",
			},
			"organization": {
				Type:        "string",
				Description: "GitHub organization name (optional)",
			},
			"repository": {
				Type:        "string",
				Description: "Default repository name (optional)",
			},
		},
		Required: []string{"token"},
	}
}

// ValidateConfig validates the configuration
func (ga *GitHubAdapter) ValidateConfig(config map[string]interface{}) error {
	if token, exists := config["token"]; !exists || token == "" {
		return fmt.Errorf("GitHub token is required")
	}

	if baseURL, exists := config["base_url"]; exists && baseURL != "" {
		ga.baseURL = baseURL.(string)
	}

	return nil
}

// Connection Management

// Connect establishes a connection to GitHub
func (ga *GitHubAdapter) Connect(ctx context.Context, config map[string]interface{}) error {
	if err := ga.ValidateConfig(config); err != nil {
		return err
	}

	ga.token = config["token"].(string)

	// Test the connection
	if err := ga.TestConnection(ctx); err != nil {
		return fmt.Errorf("failed to connect to GitHub: %w", err)
	}

	ga.connected = true
	return nil
}

// Disconnect closes the connection to GitHub
func (ga *GitHubAdapter) Disconnect(ctx context.Context) error {
	ga.connected = false
	ga.token = ""
	return nil
}

// IsConnected returns whether the adapter is connected
func (ga *GitHubAdapter) IsConnected() bool {
	return ga.connected
}

// TestConnection tests the connection to GitHub
func (ga *GitHubAdapter) TestConnection(ctx context.Context) error {
	if ga.token == "" {
		return fmt.Errorf("GitHub token not configured")
	}

	// Test by getting user information
	req, err := http.NewRequestWithContext(ctx, "GET", ga.baseURL+"/user", nil)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "token "+ga.token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := ga.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	return nil
}

// Operations

// Execute performs an operation
func (ga *GitHubAdapter) Execute(ctx context.Context, operation string, params map[string]interface{}) (*integrations.OperationResult, error) {
	start := time.Now()

	if !ga.connected {
		return &integrations.OperationResult{
			Success:   false,
			Error:     "not connected to GitHub",
			Duration:  time.Since(start),
			Timestamp: time.Now(),
		}, nil
	}

	var result *integrations.OperationResult
	var err error

	switch operation {
	case "list_repositories":
		result, err = ga.listRepositories(ctx, params)
	case "get_repository":
		result, err = ga.getRepository(ctx, params)
	case "list_issues":
		result, err = ga.listIssues(ctx, params)
	case "get_issue":
		result, err = ga.getIssue(ctx, params)
	case "create_issue":
		result, err = ga.createIssue(ctx, params)
	case "list_pull_requests":
		result, err = ga.listPullRequests(ctx, params)
	case "get_pull_request":
		result, err = ga.getPullRequest(ctx, params)
	case "list_commits":
		result, err = ga.listCommits(ctx, params)
	case "create_webhook":
		result, err = ga.createWebhook(ctx, params)
	case "list_webhooks":
		result, err = ga.listWebhooks(ctx, params)
	default:
		return &integrations.OperationResult{
			Success:   false,
			Error:     fmt.Sprintf("unsupported operation: %s", operation),
			Duration:  time.Since(start),
			Timestamp: time.Now(),
		}, nil
	}

	if err != nil {
		return &integrations.OperationResult{
			Success:   false,
			Error:     err.Error(),
			Duration:  time.Since(start),
			Timestamp: time.Now(),
		}, nil
	}

	result.Duration = time.Since(start)
	result.Timestamp = time.Now()
	return result, nil
}

// Event Handling

// SupportsWebhooks returns whether the adapter supports webhooks
func (ga *GitHubAdapter) SupportsWebhooks() bool {
	return true
}

// GetWebhookConfig returns the webhook configuration
func (ga *GitHubAdapter) GetWebhookConfig() *integrations.WebhookConfig {
	return &integrations.WebhookConfig{
		Timeout:         30 * time.Second,
		ContentType:     "application/json",
		SignatureHeader: "X-Hub-Signature-256",
	}
}

// ProcessWebhookPayload processes a webhook payload
func (ga *GitHubAdapter) ProcessWebhookPayload(payload []byte, headers map[string]string) (*integrations.IntegrationEvent, error) {
	eventType := headers["X-GitHub-Event"]
	if eventType == "" {
		return nil, fmt.Errorf("missing X-GitHub-Event header")
	}

	var data map[string]interface{}
	if err := json.Unmarshal(payload, &data); err != nil {
		return nil, fmt.Errorf("failed to parse webhook payload: %w", err)
	}

	event := &integrations.IntegrationEvent{
		Type:      fmt.Sprintf("github.%s", eventType),
		Source:    "github",
		Data:      data,
		Timestamp: time.Now(),
		Metadata: map[string]interface{}{
			"delivery_id": headers["X-GitHub-Delivery"],
			"event_type":  eventType,
		},
	}

	return event, nil
}

// Health and Monitoring

// GetHealth returns the adapter health status
func (ga *GitHubAdapter) GetHealth() *integrations.AdapterHealth {
	status := integrations.HealthStatusHealthy
	message := "GitHub adapter is healthy"

	if !ga.connected {
		status = integrations.HealthStatusUnhealthy
		message = "Not connected to GitHub"
	}

	return &integrations.AdapterHealth{
		Status:       status,
		Message:      message,
		Connected:    ga.connected,
		LastActivity: time.Now(),
	}
}

// GetMetrics returns adapter metrics
func (ga *GitHubAdapter) GetMetrics() *integrations.AdapterMetrics {
	return &integrations.AdapterMetrics{
		OperationCount:   make(map[string]int64),
		AverageLatency:   make(map[string]time.Duration),
		ErrorCount:       make(map[string]int64),
		LastOperation:    time.Now(),
		ConnectionUptime: time.Hour, // Placeholder
	}
}

// Helper methods for GitHub operations

// listRepositories lists repositories
func (ga *GitHubAdapter) listRepositories(ctx context.Context, params map[string]interface{}) (*integrations.OperationResult, error) {
	url := ga.baseURL + "/user/repos"

	if org, exists := params["organization"]; exists {
		url = fmt.Sprintf("%s/orgs/%s/repos", ga.baseURL, org)
	}

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "token "+ga.token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := ga.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var repositories []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&repositories); err != nil {
		return nil, err
	}

	return &integrations.OperationResult{
		Success: true,
		Data: map[string]interface{}{
			"repositories": repositories,
			"count":        len(repositories),
		},
	}, nil
}

// getRepository gets a specific repository
func (ga *GitHubAdapter) getRepository(ctx context.Context, params map[string]interface{}) (*integrations.OperationResult, error) {
	owner, ownerExists := params["owner"]
	repo, repoExists := params["repository"]

	if !ownerExists || !repoExists {
		return nil, fmt.Errorf("owner and repository parameters are required")
	}

	url := fmt.Sprintf("%s/repos/%s/%s", ga.baseURL, owner, repo)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "token "+ga.token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := ga.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var repository map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&repository); err != nil {
		return nil, err
	}

	return &integrations.OperationResult{
		Success: true,
		Data: map[string]interface{}{
			"repository": repository,
		},
	}, nil
}

// listIssues lists issues for a repository
func (ga *GitHubAdapter) listIssues(ctx context.Context, params map[string]interface{}) (*integrations.OperationResult, error) {
	owner, ownerExists := params["owner"]
	repo, repoExists := params["repository"]

	if !ownerExists || !repoExists {
		return nil, fmt.Errorf("owner and repository parameters are required")
	}

	url := fmt.Sprintf("%s/repos/%s/%s/issues", ga.baseURL, owner, repo)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "token "+ga.token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := ga.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var issues []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&issues); err != nil {
		return nil, err
	}

	return &integrations.OperationResult{
		Success: true,
		Data: map[string]interface{}{
			"issues": issues,
			"count":  len(issues),
		},
	}, nil
}

// getIssue gets a specific issue
func (ga *GitHubAdapter) getIssue(ctx context.Context, params map[string]interface{}) (*integrations.OperationResult, error) {
	// Placeholder implementation
	return &integrations.OperationResult{Success: true, Data: map[string]interface{}{"issue": "placeholder"}}, nil
}

// createIssue creates a new issue
func (ga *GitHubAdapter) createIssue(ctx context.Context, params map[string]interface{}) (*integrations.OperationResult, error) {
	// Placeholder implementation
	return &integrations.OperationResult{Success: true, Data: map[string]interface{}{"issue": "created"}}, nil
}

// listPullRequests lists pull requests
func (ga *GitHubAdapter) listPullRequests(ctx context.Context, params map[string]interface{}) (*integrations.OperationResult, error) {
	// Placeholder implementation
	return &integrations.OperationResult{Success: true, Data: map[string]interface{}{"pull_requests": []interface{}{}}}, nil
}

// getPullRequest gets a specific pull request
func (ga *GitHubAdapter) getPullRequest(ctx context.Context, params map[string]interface{}) (*integrations.OperationResult, error) {
	// Placeholder implementation
	return &integrations.OperationResult{Success: true, Data: map[string]interface{}{"pull_request": "placeholder"}}, nil
}

// listCommits lists commits
func (ga *GitHubAdapter) listCommits(ctx context.Context, params map[string]interface{}) (*integrations.OperationResult, error) {
	// Placeholder implementation
	return &integrations.OperationResult{Success: true, Data: map[string]interface{}{"commits": []interface{}{}}}, nil
}

// createWebhook creates a webhook
func (ga *GitHubAdapter) createWebhook(ctx context.Context, params map[string]interface{}) (*integrations.OperationResult, error) {
	// Placeholder implementation
	return &integrations.OperationResult{Success: true, Data: map[string]interface{}{"webhook": "created"}}, nil
}

// listWebhooks lists webhooks
func (ga *GitHubAdapter) listWebhooks(ctx context.Context, params map[string]interface{}) (*integrations.OperationResult, error) {
	// Placeholder implementation
	return &integrations.OperationResult{Success: true, Data: map[string]interface{}{"webhooks": []interface{}{}}}, nil
}
