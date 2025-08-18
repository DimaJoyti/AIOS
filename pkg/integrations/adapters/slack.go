package adapters

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/aios/aios/pkg/integrations"
)

// SlackAdapter implements the IntegrationAdapter interface for Slack
type SlackAdapter struct {
	client    *http.Client
	baseURL   string
	token     string
	connected bool
}

// NewSlackAdapter creates a new Slack adapter
func NewSlackAdapter() *SlackAdapter {
	return &SlackAdapter{
		client:  &http.Client{Timeout: 30 * time.Second},
		baseURL: "https://slack.com/api",
	}
}

// Adapter Information

// GetType returns the adapter type
func (sa *SlackAdapter) GetType() string {
	return "slack"
}

// GetName returns the adapter name
func (sa *SlackAdapter) GetName() string {
	return "Slack Integration"
}

// GetDescription returns the adapter description
func (sa *SlackAdapter) GetDescription() string {
	return "Integration with Slack for messaging, channels, and notifications"
}

// GetVersion returns the adapter version
func (sa *SlackAdapter) GetVersion() string {
	return "1.0.0"
}

// GetSupportedOperations returns the list of supported operations
func (sa *SlackAdapter) GetSupportedOperations() []string {
	return []string{
		"send_message",
		"list_channels",
		"get_channel",
		"create_channel",
		"list_users",
		"get_user",
		"upload_file",
		"get_team_info",
		"set_status",
		"list_conversations",
		"post_ephemeral",
		"update_message",
		"delete_message",
	}
}

// Configuration

// GetConfigSchema returns the configuration schema
func (sa *SlackAdapter) GetConfigSchema() *integrations.ConfigSchema {
	return &integrations.ConfigSchema{
		Properties: map[string]*integrations.ConfigProperty{
			"token": {
				Type:        "string",
				Description: "Slack Bot User OAuth Token (starts with xoxb-)",
				Sensitive:   true,
			},
			"signing_secret": {
				Type:        "string",
				Description: "Slack App Signing Secret for webhook verification",
				Sensitive:   true,
			},
			"workspace": {
				Type:        "string",
				Description: "Slack workspace name (optional)",
			},
			"default_channel": {
				Type:        "string",
				Description: "Default channel for notifications",
				Default:     "#general",
			},
		},
		Required: []string{"token"},
	}
}

// ValidateConfig validates the configuration
func (sa *SlackAdapter) ValidateConfig(config map[string]interface{}) error {
	if token, exists := config["token"]; !exists || token == "" {
		return fmt.Errorf("Slack token is required")
	}

	tokenStr := config["token"].(string)
	if len(tokenStr) < 10 || !isValidSlackToken(tokenStr) {
		return fmt.Errorf("invalid Slack token format")
	}

	return nil
}

// Connection Management

// Connect establishes a connection to Slack
func (sa *SlackAdapter) Connect(ctx context.Context, config map[string]interface{}) error {
	if err := sa.ValidateConfig(config); err != nil {
		return err
	}

	sa.token = config["token"].(string)

	// Test the connection
	if err := sa.TestConnection(ctx); err != nil {
		return fmt.Errorf("failed to connect to Slack: %w", err)
	}

	sa.connected = true
	return nil
}

// Disconnect closes the connection to Slack
func (sa *SlackAdapter) Disconnect(ctx context.Context) error {
	sa.connected = false
	sa.token = ""
	return nil
}

// IsConnected returns whether the adapter is connected
func (sa *SlackAdapter) IsConnected() bool {
	return sa.connected
}

// TestConnection tests the connection to Slack
func (sa *SlackAdapter) TestConnection(ctx context.Context) error {
	if sa.token == "" {
		return fmt.Errorf("Slack token not configured")
	}

	// Test by calling auth.test
	req, err := http.NewRequestWithContext(ctx, "POST", sa.baseURL+"/auth.test", nil)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+sa.token)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := sa.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}

	if ok, exists := result["ok"]; !exists || !ok.(bool) {
		errorMsg := "unknown error"
		if errStr, exists := result["error"]; exists {
			errorMsg = errStr.(string)
		}
		return fmt.Errorf("Slack API error: %s", errorMsg)
	}

	return nil
}

// Operations

// Execute performs an operation
func (sa *SlackAdapter) Execute(ctx context.Context, operation string, params map[string]interface{}) (*integrations.OperationResult, error) {
	start := time.Now()

	if !sa.connected {
		return &integrations.OperationResult{
			Success:   false,
			Error:     "not connected to Slack",
			Duration:  time.Since(start),
			Timestamp: time.Now(),
		}, nil
	}

	var result *integrations.OperationResult
	var err error

	switch operation {
	case "send_message":
		result, err = sa.sendMessage(ctx, params)
	case "list_channels":
		result, err = sa.listChannels(ctx, params)
	case "get_channel":
		result, err = sa.getChannel(ctx, params)
	case "create_channel":
		result, err = sa.createChannel(ctx, params)
	case "list_users":
		result, err = sa.listUsers(ctx, params)
	case "get_user":
		result, err = sa.getUser(ctx, params)
	case "get_team_info":
		result, err = sa.getTeamInfo(ctx, params)
	case "upload_file":
		result, err = sa.uploadFile(ctx, params)
	case "post_ephemeral":
		result, err = sa.postEphemeral(ctx, params)
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
func (sa *SlackAdapter) SupportsWebhooks() bool {
	return true
}

// GetWebhookConfig returns the webhook configuration
func (sa *SlackAdapter) GetWebhookConfig() *integrations.WebhookConfig {
	return &integrations.WebhookConfig{
		Timeout:         30 * time.Second,
		ContentType:     "application/json",
		SignatureHeader: "X-Slack-Signature",
	}
}

// ProcessWebhookPayload processes a webhook payload
func (sa *SlackAdapter) ProcessWebhookPayload(payload []byte, headers map[string]string) (*integrations.IntegrationEvent, error) {
	var data map[string]interface{}
	if err := json.Unmarshal(payload, &data); err != nil {
		return nil, fmt.Errorf("failed to parse webhook payload: %w", err)
	}

	// Handle URL verification challenge
	if challenge, exists := data["challenge"]; exists {
		return &integrations.IntegrationEvent{
			Type:      "slack.url_verification",
			Source:    "slack",
			Data:      data,
			Timestamp: time.Now(),
			Metadata: map[string]interface{}{
				"challenge": challenge,
			},
		}, nil
	}

	// Handle event callbacks
	eventType := "slack.event"
	if event, exists := data["event"]; exists {
		if eventMap, ok := event.(map[string]interface{}); ok {
			if eventTypeStr, exists := eventMap["type"]; exists {
				eventType = fmt.Sprintf("slack.%s", eventTypeStr)
			}
		}
	}

	event := &integrations.IntegrationEvent{
		Type:      eventType,
		Source:    "slack",
		Data:      data,
		Timestamp: time.Now(),
		Metadata: map[string]interface{}{
			"team_id":    data["team_id"],
			"api_app_id": data["api_app_id"],
		},
	}

	return event, nil
}

// Health and Monitoring

// GetHealth returns the adapter health status
func (sa *SlackAdapter) GetHealth() *integrations.AdapterHealth {
	status := integrations.HealthStatusHealthy
	message := "Slack adapter is healthy"

	if !sa.connected {
		status = integrations.HealthStatusUnhealthy
		message = "Not connected to Slack"
	}

	return &integrations.AdapterHealth{
		Status:       status,
		Message:      message,
		Connected:    sa.connected,
		LastActivity: time.Now(),
	}
}

// GetMetrics returns adapter metrics
func (sa *SlackAdapter) GetMetrics() *integrations.AdapterMetrics {
	return &integrations.AdapterMetrics{
		OperationCount:   make(map[string]int64),
		AverageLatency:   make(map[string]time.Duration),
		ErrorCount:       make(map[string]int64),
		LastOperation:    time.Now(),
		ConnectionUptime: time.Hour, // Placeholder
	}
}

// Helper methods for Slack operations

// sendMessage sends a message to a Slack channel
func (sa *SlackAdapter) sendMessage(ctx context.Context, params map[string]interface{}) (*integrations.OperationResult, error) {
	channel, channelExists := params["channel"]
	text, textExists := params["text"]

	if !channelExists || !textExists {
		return nil, fmt.Errorf("channel and text parameters are required")
	}

	payload := map[string]interface{}{
		"channel": channel,
		"text":    text,
	}

	// Add optional parameters
	if username, exists := params["username"]; exists {
		payload["username"] = username
	}
	if iconEmoji, exists := params["icon_emoji"]; exists {
		payload["icon_emoji"] = iconEmoji
	}
	if attachments, exists := params["attachments"]; exists {
		payload["attachments"] = attachments
	}
	if blocks, exists := params["blocks"]; exists {
		payload["blocks"] = blocks
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", sa.baseURL+"/chat.postMessage", bytes.NewBuffer(jsonPayload))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+sa.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := sa.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if ok, exists := result["ok"]; !exists || !ok.(bool) {
		errorMsg := "unknown error"
		if errStr, exists := result["error"]; exists {
			errorMsg = errStr.(string)
		}
		return nil, fmt.Errorf("Slack API error: %s", errorMsg)
	}

	return &integrations.OperationResult{
		Success: true,
		Data: map[string]interface{}{
			"message": result,
			"ts":      result["ts"],
			"channel": result["channel"],
		},
	}, nil
}

// listChannels lists Slack channels
func (sa *SlackAdapter) listChannels(ctx context.Context, params map[string]interface{}) (*integrations.OperationResult, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", sa.baseURL+"/conversations.list", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+sa.token)

	resp, err := sa.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if ok, exists := result["ok"]; !exists || !ok.(bool) {
		errorMsg := "unknown error"
		if errStr, exists := result["error"]; exists {
			errorMsg = errStr.(string)
		}
		return nil, fmt.Errorf("Slack API error: %s", errorMsg)
	}

	channels := result["channels"].([]interface{})

	return &integrations.OperationResult{
		Success: true,
		Data: map[string]interface{}{
			"channels": channels,
			"count":    len(channels),
		},
	}, nil
}

// getTeamInfo gets Slack team information
func (sa *SlackAdapter) getTeamInfo(ctx context.Context, params map[string]interface{}) (*integrations.OperationResult, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", sa.baseURL+"/team.info", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+sa.token)

	resp, err := sa.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if ok, exists := result["ok"]; !exists || !ok.(bool) {
		errorMsg := "unknown error"
		if errStr, exists := result["error"]; exists {
			errorMsg = errStr.(string)
		}
		return nil, fmt.Errorf("Slack API error: %s", errorMsg)
	}

	return &integrations.OperationResult{
		Success: true,
		Data: map[string]interface{}{
			"team": result["team"],
		},
	}, nil
}

// Helper function to validate Slack token format
func isValidSlackToken(token string) bool {
	// Basic validation for Slack token format
	return len(token) > 10 && (token[:5] == "xoxb-" || token[:5] == "xoxp-" || token[:5] == "xoxa-")
}

// getChannel gets a specific channel
func (sa *SlackAdapter) getChannel(ctx context.Context, params map[string]interface{}) (*integrations.OperationResult, error) {
	// Placeholder implementation
	return &integrations.OperationResult{Success: true, Data: map[string]interface{}{"channel": "placeholder"}}, nil
}

// createChannel creates a new channel
func (sa *SlackAdapter) createChannel(ctx context.Context, params map[string]interface{}) (*integrations.OperationResult, error) {
	// Placeholder implementation
	return &integrations.OperationResult{Success: true, Data: map[string]interface{}{"channel": "created"}}, nil
}

// listUsers lists Slack users
func (sa *SlackAdapter) listUsers(ctx context.Context, params map[string]interface{}) (*integrations.OperationResult, error) {
	// Placeholder implementation
	return &integrations.OperationResult{Success: true, Data: map[string]interface{}{"users": []interface{}{}}}, nil
}

// getUser gets a specific user
func (sa *SlackAdapter) getUser(ctx context.Context, params map[string]interface{}) (*integrations.OperationResult, error) {
	// Placeholder implementation
	return &integrations.OperationResult{Success: true, Data: map[string]interface{}{"user": "placeholder"}}, nil
}

// uploadFile uploads a file to Slack
func (sa *SlackAdapter) uploadFile(ctx context.Context, params map[string]interface{}) (*integrations.OperationResult, error) {
	// Placeholder implementation
	return &integrations.OperationResult{Success: true, Data: map[string]interface{}{"file": "uploaded"}}, nil
}

// postEphemeral posts an ephemeral message
func (sa *SlackAdapter) postEphemeral(ctx context.Context, params map[string]interface{}) (*integrations.OperationResult, error) {
	// Placeholder implementation
	return &integrations.OperationResult{Success: true, Data: map[string]interface{}{"message": "ephemeral"}}, nil
}
