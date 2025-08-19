package workflow

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

// Component implementations for workflow engine

// DefaultActionExecutor implements ActionExecutor interface
type DefaultActionExecutor struct {
	logger *logrus.Logger
	tracer trace.Tracer
	client *http.Client
}

func NewDefaultActionExecutor(logger *logrus.Logger) (ActionExecutor, error) {
	return &DefaultActionExecutor{
		logger: logger,
		tracer: otel.Tracer("workflow.action_executor"),
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

func (ae *DefaultActionExecutor) ExecuteAction(ctx context.Context, action *Action, input map[string]interface{}) (map[string]interface{}, error) {
	ctx, span := ae.tracer.Start(ctx, "action_executor.execute_action")
	defer span.End()

	ae.logger.WithFields(logrus.Fields{
		"action_id":   action.ID,
		"action_type": action.Type,
		"action_name": action.Name,
	}).Info("Executing action")

	switch action.Type {
	case ActionTypeHTTP:
		return ae.executeHTTPAction(ctx, action, input)
	case ActionTypeEmail:
		return ae.executeEmailAction(ctx, action, input)
	case ActionTypeSlack:
		return ae.executeSlackAction(ctx, action, input)
	case ActionTypeWebhook:
		return ae.executeWebhookAction(ctx, action, input)
	case ActionTypeScript:
		return ae.executeScriptAction(ctx, action, input)
	case ActionTypeDatabase:
		return ae.executeDatabaseAction(ctx, action, input)
	case ActionTypeFileSystem:
		return ae.executeFileSystemAction(ctx, action, input)
	case ActionTypeNotification:
		return ae.executeNotificationAction(ctx, action, input)
	case ActionTypeIntegration:
		return ae.executeIntegrationAction(ctx, action, input)
	default:
		return nil, fmt.Errorf("unsupported action type: %s", action.Type)
	}
}

func (ae *DefaultActionExecutor) ValidateAction(ctx context.Context, action *Action) error {
	switch action.Type {
	case ActionTypeHTTP, ActionTypeWebhook:
		if _, exists := action.Config["url"]; !exists {
			return fmt.Errorf("HTTP/Webhook action requires 'url' in config")
		}
	case ActionTypeEmail:
		if _, exists := action.Config["to"]; !exists {
			return fmt.Errorf("Email action requires 'to' in config")
		}
		if _, exists := action.Config["subject"]; !exists {
			return fmt.Errorf("Email action requires 'subject' in config")
		}
	case ActionTypeSlack:
		if _, exists := action.Config["channel"]; !exists {
			return fmt.Errorf("Slack action requires 'channel' in config")
		}
		if _, exists := action.Config["message"]; !exists {
			return fmt.Errorf("Slack action requires 'message' in config")
		}
	case ActionTypeScript:
		if _, exists := action.Config["script"]; !exists {
			return fmt.Errorf("Script action requires 'script' in config")
		}
	}
	return nil
}

func (ae *DefaultActionExecutor) GetSupportedActionTypes() []ActionType {
	return []ActionType{
		ActionTypeHTTP,
		ActionTypeEmail,
		ActionTypeSlack,
		ActionTypeWebhook,
		ActionTypeScript,
		ActionTypeDatabase,
		ActionTypeFileSystem,
		ActionTypeNotification,
		ActionTypeIntegration,
	}
}

// Action execution methods

func (ae *DefaultActionExecutor) executeHTTPAction(ctx context.Context, action *Action, input map[string]interface{}) (map[string]interface{}, error) {
	url, _ := action.Config["url"].(string)
	method, _ := action.Config["method"].(string)
	if method == "" {
		method = "GET"
	}

	var body io.Reader
	if payload, exists := action.Config["payload"]; exists {
		jsonData, _ := json.Marshal(payload)
		body = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Add headers
	if headers, exists := action.Config["headers"].(map[string]interface{}); exists {
		for key, value := range headers {
			req.Header.Set(key, fmt.Sprintf("%v", value))
		}
	}

	resp, err := ae.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	result := map[string]interface{}{
		"status_code": resp.StatusCode,
		"response":    string(responseBody),
		"headers":     resp.Header,
	}

	if resp.StatusCode >= 400 {
		return result, fmt.Errorf("HTTP request failed with status %d", resp.StatusCode)
	}

	ae.logger.WithFields(logrus.Fields{
		"url":         url,
		"method":      method,
		"status_code": resp.StatusCode,
	}).Info("HTTP action executed successfully")

	return result, nil
}

func (ae *DefaultActionExecutor) executeEmailAction(ctx context.Context, action *Action, input map[string]interface{}) (map[string]interface{}, error) {
	to, _ := action.Config["to"].(string)
	subject, _ := action.Config["subject"].(string)
	body, _ := action.Config["body"].(string)

	// Template substitution (simplified)
	subject = ae.substituteVariables(subject, input)
	body = ae.substituteVariables(body, input)

	// In a real implementation, this would send an actual email
	ae.logger.WithFields(logrus.Fields{
		"to":      to,
		"subject": subject,
	}).Info("Email action executed (simulated)")

	return map[string]interface{}{
		"sent":    true,
		"to":      to,
		"subject": subject,
	}, nil
}

func (ae *DefaultActionExecutor) executeSlackAction(ctx context.Context, action *Action, input map[string]interface{}) (map[string]interface{}, error) {
	channel, _ := action.Config["channel"].(string)
	message, _ := action.Config["message"].(string)

	// Template substitution
	message = ae.substituteVariables(message, input)

	// In a real implementation, this would send to Slack API
	ae.logger.WithFields(logrus.Fields{
		"channel": channel,
		"message": message,
	}).Info("Slack action executed (simulated)")

	return map[string]interface{}{
		"sent":    true,
		"channel": channel,
		"message": message,
	}, nil
}

func (ae *DefaultActionExecutor) executeWebhookAction(ctx context.Context, action *Action, input map[string]interface{}) (map[string]interface{}, error) {
	return ae.executeHTTPAction(ctx, action, input)
}

func (ae *DefaultActionExecutor) executeScriptAction(ctx context.Context, action *Action, input map[string]interface{}) (map[string]interface{}, error) {
	script, _ := action.Config["script"].(string)
	shell, _ := action.Config["shell"].(string)
	if shell == "" {
		shell = "/bin/sh"
	}

	cmd := exec.CommandContext(ctx, shell, "-c", script)

	// Set environment variables from input
	env := []string{}
	for key, value := range input {
		env = append(env, fmt.Sprintf("%s=%v", key, value))
	}
	cmd.Env = env

	output, err := cmd.CombinedOutput()
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"output":  string(output),
			"error":   err.Error(),
		}, fmt.Errorf("script execution failed: %w", err)
	}

	ae.logger.WithField("script", script).Info("Script action executed successfully")

	return map[string]interface{}{
		"success": true,
		"output":  string(output),
	}, nil
}

func (ae *DefaultActionExecutor) executeDatabaseAction(ctx context.Context, action *Action, input map[string]interface{}) (map[string]interface{}, error) {
	// Simplified database action
	query, _ := action.Config["query"].(string)

	ae.logger.WithField("query", query).Info("Database action executed (simulated)")

	return map[string]interface{}{
		"executed": true,
		"query":    query,
		"rows":     0,
	}, nil
}

func (ae *DefaultActionExecutor) executeFileSystemAction(ctx context.Context, action *Action, input map[string]interface{}) (map[string]interface{}, error) {
	operation, _ := action.Config["operation"].(string)
	path, _ := action.Config["path"].(string)

	ae.logger.WithFields(logrus.Fields{
		"operation": operation,
		"path":      path,
	}).Info("FileSystem action executed (simulated)")

	return map[string]interface{}{
		"executed":  true,
		"operation": operation,
		"path":      path,
	}, nil
}

func (ae *DefaultActionExecutor) executeNotificationAction(ctx context.Context, action *Action, input map[string]interface{}) (map[string]interface{}, error) {
	message, _ := action.Config["message"].(string)
	recipients, _ := action.Config["recipients"].([]interface{})

	message = ae.substituteVariables(message, input)

	ae.logger.WithFields(logrus.Fields{
		"message":    message,
		"recipients": len(recipients),
	}).Info("Notification action executed (simulated)")

	return map[string]interface{}{
		"sent":       true,
		"message":    message,
		"recipients": len(recipients),
	}, nil
}

func (ae *DefaultActionExecutor) executeIntegrationAction(ctx context.Context, action *Action, input map[string]interface{}) (map[string]interface{}, error) {
	integration, _ := action.Config["integration"].(string)
	operation, _ := action.Config["operation"].(string)

	ae.logger.WithFields(logrus.Fields{
		"integration": integration,
		"operation":   operation,
	}).Info("Integration action executed (simulated)")

	return map[string]interface{}{
		"executed":    true,
		"integration": integration,
		"operation":   operation,
	}, nil
}

// substituteVariables performs simple variable substitution
func (ae *DefaultActionExecutor) substituteVariables(template string, variables map[string]interface{}) string {
	result := template
	for key, value := range variables {
		placeholder := fmt.Sprintf("{{.%s}}", key)
		result = strings.ReplaceAll(result, placeholder, fmt.Sprintf("%v", value))
	}
	return result
}

// DefaultTriggerManager implements TriggerManager interface
type DefaultTriggerManager struct {
	logger   *logrus.Logger
	tracer   trace.Tracer
	triggers map[string]*TriggerRegistration // triggerID -> registration
}

type TriggerRegistration struct {
	WorkflowID string
	Trigger    *Trigger
}

func NewDefaultTriggerManager(logger *logrus.Logger) (TriggerManager, error) {
	return &DefaultTriggerManager{
		logger:   logger,
		tracer:   otel.Tracer("workflow.trigger_manager"),
		triggers: make(map[string]*TriggerRegistration),
	}, nil
}

func (tm *DefaultTriggerManager) RegisterTrigger(ctx context.Context, workflowID string, trigger *Trigger) error {
	ctx, span := tm.tracer.Start(ctx, "trigger_manager.register_trigger")
	defer span.End()

	if trigger.ID == "" {
		trigger.ID = uuid.New().String()
	}

	registration := &TriggerRegistration{
		WorkflowID: workflowID,
		Trigger:    trigger,
	}

	tm.triggers[trigger.ID] = registration

	tm.logger.WithFields(logrus.Fields{
		"trigger_id":   trigger.ID,
		"workflow_id":  workflowID,
		"trigger_type": trigger.Type,
	}).Info("Trigger registered")

	return nil
}

func (tm *DefaultTriggerManager) UnregisterTrigger(ctx context.Context, triggerID string) error {
	delete(tm.triggers, triggerID)

	tm.logger.WithField("trigger_id", triggerID).Info("Trigger unregistered")
	return nil
}

func (tm *DefaultTriggerManager) ProcessEvent(ctx context.Context, event *Event) ([]*Workflow, error) {
	ctx, span := tm.tracer.Start(ctx, "trigger_manager.process_event")
	defer span.End()

	var matchedWorkflows []*Workflow

	for _, registration := range tm.triggers {
		if tm.eventMatchesTrigger(event, registration.Trigger) {
			// In a real implementation, this would fetch the actual workflow
			workflow := &Workflow{
				ID:   registration.WorkflowID,
				Name: fmt.Sprintf("Workflow-%s", registration.WorkflowID),
			}
			matchedWorkflows = append(matchedWorkflows, workflow)
		}
	}

	tm.logger.WithFields(logrus.Fields{
		"event_type":        event.Type,
		"matched_workflows": len(matchedWorkflows),
	}).Info("Event processed")

	return matchedWorkflows, nil
}

func (tm *DefaultTriggerManager) ListTriggers(ctx context.Context, workflowID string) ([]*Trigger, error) {
	var triggers []*Trigger
	for _, registration := range tm.triggers {
		if registration.WorkflowID == workflowID {
			triggers = append(triggers, registration.Trigger)
		}
	}
	return triggers, nil
}

func (tm *DefaultTriggerManager) ValidateTrigger(ctx context.Context, trigger *Trigger) error {
	if trigger.Type == "" {
		return fmt.Errorf("trigger type is required")
	}
	return nil
}

func (tm *DefaultTriggerManager) eventMatchesTrigger(event *Event, trigger *Trigger) bool {
	if !trigger.Enabled {
		return false
	}

	// Simple event matching - in production this would be more sophisticated
	switch trigger.Type {
	case TriggerTypeEvent:
		if triggerEvent, exists := trigger.Config["event"].(string); exists {
			return event.Type == triggerEvent
		}
		if triggerEvents, exists := trigger.Config["events"].([]interface{}); exists {
			for _, te := range triggerEvents {
				if event.Type == fmt.Sprintf("%v", te) {
					return true
				}
			}
		}
	case TriggerTypeWebhook:
		return event.Type == "webhook"
	case TriggerTypePush:
		return event.Type == "push"
	case TriggerTypePR:
		return event.Type == "pull_request"
	}

	return false
}

// DefaultScheduleManager implements ScheduleManager interface
type DefaultScheduleManager struct {
	logger             *logrus.Logger
	tracer             trace.Tracer
	scheduledWorkflows map[string]*ScheduledWorkflow
}

func NewDefaultScheduleManager(logger *logrus.Logger) (ScheduleManager, error) {
	return &DefaultScheduleManager{
		logger:             logger,
		tracer:             otel.Tracer("workflow.schedule_manager"),
		scheduledWorkflows: make(map[string]*ScheduledWorkflow),
	}, nil
}

func (sm *DefaultScheduleManager) ScheduleWorkflow(ctx context.Context, workflowID string, schedule *Schedule) error {
	ctx, span := sm.tracer.Start(ctx, "schedule_manager.schedule_workflow")
	defer span.End()

	scheduledWorkflow := &ScheduledWorkflow{
		WorkflowID: workflowID,
		Schedule:   schedule,
		NextRun:    sm.calculateNextRun(schedule),
		Enabled:    schedule.Enabled,
	}

	sm.scheduledWorkflows[workflowID] = scheduledWorkflow

	sm.logger.WithFields(logrus.Fields{
		"workflow_id":   workflowID,
		"schedule_type": schedule.Type,
		"next_run":      scheduledWorkflow.NextRun,
	}).Info("Workflow scheduled")

	return nil
}

func (sm *DefaultScheduleManager) UnscheduleWorkflow(ctx context.Context, workflowID string) error {
	delete(sm.scheduledWorkflows, workflowID)

	sm.logger.WithField("workflow_id", workflowID).Info("Workflow unscheduled")
	return nil
}

func (sm *DefaultScheduleManager) GetScheduledWorkflows(ctx context.Context) ([]*ScheduledWorkflow, error) {
	var workflows []*ScheduledWorkflow
	for _, workflow := range sm.scheduledWorkflows {
		workflows = append(workflows, workflow)
	}
	return workflows, nil
}

func (sm *DefaultScheduleManager) UpdateSchedule(ctx context.Context, workflowID string, schedule *Schedule) error {
	if scheduledWorkflow, exists := sm.scheduledWorkflows[workflowID]; exists {
		scheduledWorkflow.Schedule = schedule
		scheduledWorkflow.NextRun = sm.calculateNextRun(schedule)
		scheduledWorkflow.Enabled = schedule.Enabled

		sm.logger.WithField("workflow_id", workflowID).Info("Schedule updated")
	}
	return nil
}

func (sm *DefaultScheduleManager) calculateNextRun(schedule *Schedule) time.Time {
	now := time.Now()

	switch schedule.Type {
	case ScheduleTypeInterval:
		if schedule.Interval > 0 {
			return now.Add(schedule.Interval)
		}
	case ScheduleTypeOnce:
		if schedule.StartTime != nil {
			return *schedule.StartTime
		}
	case ScheduleTypeCron:
		// Simplified cron calculation - in production use a proper cron library
		return now.Add(time.Hour) // Default to 1 hour
	}

	return now.Add(time.Hour) // Default fallback
}

// DefaultIntegrationManager implements IntegrationManager interface
type DefaultIntegrationManager struct {
	logger *logrus.Logger
	tracer trace.Tracer
	client *http.Client
}

func NewDefaultIntegrationManager(logger *logrus.Logger) (IntegrationManager, error) {
	return &DefaultIntegrationManager{
		logger: logger,
		tracer: otel.Tracer("workflow.integration_manager"),
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

// GitHub integration
func (im *DefaultIntegrationManager) ConnectGitHub(ctx context.Context, config *GitHubConfig) error {
	im.logger.WithField("organization", config.Organization).Info("GitHub integration connected (simulated)")
	return nil
}

func (im *DefaultIntegrationManager) SyncGitHubRepository(ctx context.Context, repoURL string) error {
	im.logger.WithField("repo_url", repoURL).Info("GitHub repository synced (simulated)")
	return nil
}

func (im *DefaultIntegrationManager) CreateGitHubWebhook(ctx context.Context, repoURL string, events []string) error {
	im.logger.WithFields(logrus.Fields{
		"repo_url": repoURL,
		"events":   events,
	}).Info("GitHub webhook created (simulated)")
	return nil
}

// GitLab integration
func (im *DefaultIntegrationManager) ConnectGitLab(ctx context.Context, config *GitLabConfig) error {
	im.logger.WithField("base_url", config.BaseURL).Info("GitLab integration connected (simulated)")
	return nil
}

func (im *DefaultIntegrationManager) SyncGitLabRepository(ctx context.Context, repoURL string) error {
	im.logger.WithField("repo_url", repoURL).Info("GitLab repository synced (simulated)")
	return nil
}

func (im *DefaultIntegrationManager) CreateGitLabWebhook(ctx context.Context, repoURL string, events []string) error {
	im.logger.WithFields(logrus.Fields{
		"repo_url": repoURL,
		"events":   events,
	}).Info("GitLab webhook created (simulated)")
	return nil
}

// Jenkins integration
func (im *DefaultIntegrationManager) ConnectJenkins(ctx context.Context, config *JenkinsConfig) error {
	im.logger.WithField("jenkins_url", config.URL).Info("Jenkins integration connected (simulated)")
	return nil
}

func (im *DefaultIntegrationManager) TriggerJenkinsJob(ctx context.Context, jobName string, params map[string]interface{}) error {
	im.logger.WithFields(logrus.Fields{
		"job_name": jobName,
		"params":   params,
	}).Info("Jenkins job triggered (simulated)")
	return nil
}

func (im *DefaultIntegrationManager) GetJenkinsJobStatus(ctx context.Context, jobName string, buildNumber int) (*BuildStatus, error) {
	status := &BuildStatus{
		JobName:     jobName,
		BuildNumber: buildNumber,
		Status:      ExecutionStatusSuccess,
		StartedAt:   time.Now().Add(-10 * time.Minute),
		Duration:    10 * time.Minute,
		Result:      "SUCCESS",
		URL:         fmt.Sprintf("https://jenkins.example.com/job/%s/%d", jobName, buildNumber),
	}

	im.logger.WithFields(logrus.Fields{
		"job_name":     jobName,
		"build_number": buildNumber,
		"status":       status.Status,
	}).Info("Jenkins job status retrieved (simulated)")

	return status, nil
}

// Slack integration
func (im *DefaultIntegrationManager) ConnectSlack(ctx context.Context, config *SlackConfig) error {
	im.logger.WithField("channel", config.Channel).Info("Slack integration connected (simulated)")
	return nil
}

func (im *DefaultIntegrationManager) SendSlackMessage(ctx context.Context, channel string, message string) error {
	im.logger.WithFields(logrus.Fields{
		"channel": channel,
		"message": message,
	}).Info("Slack message sent (simulated)")
	return nil
}

// Email integration
func (im *DefaultIntegrationManager) ConnectEmail(ctx context.Context, config *EmailConfig) error {
	im.logger.WithFields(logrus.Fields{
		"smtp_host": config.SMTPHost,
		"smtp_port": config.SMTPPort,
	}).Info("Email integration connected (simulated)")
	return nil
}

func (im *DefaultIntegrationManager) SendEmail(ctx context.Context, to []string, subject string, body string) error {
	im.logger.WithFields(logrus.Fields{
		"to":      to,
		"subject": subject,
	}).Info("Email sent (simulated)")
	return nil
}

// Generic webhook
func (im *DefaultIntegrationManager) RegisterWebhook(ctx context.Context, config *WebhookConfig) error {
	im.logger.WithFields(logrus.Fields{
		"url":    config.URL,
		"method": config.Method,
	}).Info("Webhook registered (simulated)")
	return nil
}

func (im *DefaultIntegrationManager) SendWebhook(ctx context.Context, url string, payload map[string]interface{}) error {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create webhook request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := im.client.Do(req)
	if err != nil {
		return fmt.Errorf("webhook request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("webhook failed with status %d", resp.StatusCode)
	}

	im.logger.WithFields(logrus.Fields{
		"url":         url,
		"status_code": resp.StatusCode,
	}).Info("Webhook sent successfully")

	return nil
}
