package workflow

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Helper methods for DefaultWorkflowEngine

// validateWorkflow validates a workflow definition
func (we *DefaultWorkflowEngine) validateWorkflow(workflow *Workflow) error {
	if workflow.Name == "" {
		return fmt.Errorf("workflow name is required")
	}

	if len(workflow.Actions) == 0 {
		return fmt.Errorf("workflow must have at least one action")
	}

	// Validate actions
	for i, action := range workflow.Actions {
		if err := we.validateAction(action); err != nil {
			return fmt.Errorf("action %d validation failed: %w", i, err)
		}
	}

	// Validate triggers
	for i, trigger := range workflow.Triggers {
		if err := we.validateTrigger(trigger); err != nil {
			return fmt.Errorf("trigger %d validation failed: %w", i, err)
		}
	}

	// Validate conditions
	for i, condition := range workflow.Conditions {
		if err := we.validateCondition(condition); err != nil {
			return fmt.Errorf("condition %d validation failed: %w", i, err)
		}
	}

	return nil
}

// validateAction validates an action definition
func (we *DefaultWorkflowEngine) validateAction(action *Action) error {
	if action.ID == "" {
		action.ID = uuid.New().String()
	}

	if action.Name == "" {
		return fmt.Errorf("action name is required")
	}

	if action.Type == "" {
		return fmt.Errorf("action type is required")
	}

	// Validate action type is supported
	supportedTypes := we.actionExecutor.GetSupportedActionTypes()
	found := false
	for _, supportedType := range supportedTypes {
		if action.Type == supportedType {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("unsupported action type: %s", action.Type)
	}

	// Validate action configuration
	if err := we.actionExecutor.ValidateAction(nil, action); err != nil {
		return fmt.Errorf("action configuration validation failed: %w", err)
	}

	return nil
}

// validateTrigger validates a trigger definition
func (we *DefaultWorkflowEngine) validateTrigger(trigger *Trigger) error {
	if trigger.ID == "" {
		trigger.ID = uuid.New().String()
	}

	if trigger.Type == "" {
		return fmt.Errorf("trigger type is required")
	}

	// Validate trigger type
	validTypes := []TriggerType{
		TriggerTypeManual, TriggerTypeSchedule, TriggerTypeWebhook,
		TriggerTypeEvent, TriggerTypePush, TriggerTypePR,
		TriggerTypeTag, TriggerTypeRelease,
	}

	found := false
	for _, validType := range validTypes {
		if trigger.Type == validType {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("invalid trigger type: %s", trigger.Type)
	}

	// Validate conditions
	for i, condition := range trigger.Conditions {
		if err := we.validateCondition(condition); err != nil {
			return fmt.Errorf("trigger condition %d validation failed: %w", i, err)
		}
	}

	return nil
}

// validateCondition validates a condition definition
func (we *DefaultWorkflowEngine) validateCondition(condition *Condition) error {
	if condition.ID == "" {
		condition.ID = uuid.New().String()
	}

	if condition.Field == "" {
		return fmt.Errorf("condition field is required")
	}

	if condition.Operator == "" {
		return fmt.Errorf("condition operator is required")
	}

	// Validate operator
	validOperators := []Operator{
		OperatorEquals, OperatorNotEquals, OperatorGreaterThan, OperatorLessThan,
		OperatorContains, OperatorStartsWith, OperatorEndsWith, OperatorRegex,
		OperatorIn, OperatorNotIn,
	}

	found := false
	for _, validOp := range validOperators {
		if condition.Operator == validOp {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("invalid condition operator: %s", condition.Operator)
	}

	return nil
}

// evaluateConditions evaluates a list of conditions against variables
func (we *DefaultWorkflowEngine) evaluateConditions(conditions []*Condition, variables map[string]interface{}) bool {
	if len(conditions) == 0 {
		return true // No conditions means always true
	}

	// All conditions must be true (AND logic)
	for _, condition := range conditions {
		if !we.evaluateCondition(condition, variables) {
			return false
		}
	}

	return true
}

// evaluateCondition evaluates a single condition
func (we *DefaultWorkflowEngine) evaluateCondition(condition *Condition, variables map[string]interface{}) bool {
	// Get field value from variables
	fieldValue, exists := variables[condition.Field]
	if !exists {
		// Field doesn't exist, condition fails
		return false
	}

	// Convert values to strings for comparison
	fieldStr := fmt.Sprintf("%v", fieldValue)
	valueStr := fmt.Sprintf("%v", condition.Value)

	switch condition.Operator {
	case OperatorEquals:
		return fieldStr == valueStr

	case OperatorNotEquals:
		return fieldStr != valueStr

	case OperatorGreaterThan:
		return we.compareNumeric(fieldValue, condition.Value, ">")

	case OperatorLessThan:
		return we.compareNumeric(fieldValue, condition.Value, "<")

	case OperatorContains:
		return strings.Contains(fieldStr, valueStr)

	case OperatorStartsWith:
		return strings.HasPrefix(fieldStr, valueStr)

	case OperatorEndsWith:
		return strings.HasSuffix(fieldStr, valueStr)

	case OperatorRegex:
		regex, err := regexp.Compile(valueStr)
		if err != nil {
			return false
		}
		return regex.MatchString(fieldStr)

	case OperatorIn:
		// Check if field value is in the array
		if valueArray, ok := condition.Value.([]interface{}); ok {
			for _, item := range valueArray {
				if fmt.Sprintf("%v", item) == fieldStr {
					return true
				}
			}
		}
		return false

	case OperatorNotIn:
		// Check if field value is NOT in the array
		if valueArray, ok := condition.Value.([]interface{}); ok {
			for _, item := range valueArray {
				if fmt.Sprintf("%v", item) == fieldStr {
					return false
				}
			}
			return true
		}
		return true

	default:
		return false
	}
}

// compareNumeric compares two values numerically
func (we *DefaultWorkflowEngine) compareNumeric(a, b interface{}, operator string) bool {
	// Simple numeric comparison - in production, this would be more robust
	aFloat, aOk := we.toFloat64(a)
	bFloat, bOk := we.toFloat64(b)

	if !aOk || !bOk {
		return false
	}

	switch operator {
	case ">":
		return aFloat > bFloat
	case "<":
		return aFloat < bFloat
	case ">=":
		return aFloat >= bFloat
	case "<=":
		return aFloat <= bFloat
	default:
		return false
	}
}

// toFloat64 converts a value to float64
func (we *DefaultWorkflowEngine) toFloat64(value interface{}) (float64, bool) {
	switch v := value.(type) {
	case float64:
		return v, true
	case float32:
		return float64(v), true
	case int:
		return float64(v), true
	case int32:
		return float64(v), true
	case int64:
		return float64(v), true
	default:
		return 0, false
	}
}

// Filter matching methods

// matchesWorkflowFilter checks if a workflow matches the given filter
func (we *DefaultWorkflowEngine) matchesWorkflowFilter(workflow *Workflow, filter *WorkflowFilter) bool {
	if filter == nil {
		return true
	}

	// Status filter
	if len(filter.Status) > 0 {
		found := false
		for _, status := range filter.Status {
			if workflow.Status == status {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Created by filter
	if filter.CreatedBy != "" && workflow.CreatedBy != filter.CreatedBy {
		return false
	}

	// Tags filter
	if len(filter.Tags) > 0 {
		for _, filterTag := range filter.Tags {
			found := false
			for _, workflowTag := range workflow.Tags {
				if workflowTag == filterTag {
					found = true
					break
				}
			}
			if !found {
				return false
			}
		}
	}

	// Search filter
	if filter.Search != "" {
		searchLower := strings.ToLower(filter.Search)
		if !(strings.Contains(strings.ToLower(workflow.Name), searchLower) ||
			strings.Contains(strings.ToLower(workflow.Description), searchLower)) {
			return false
		}
	}

	return true
}

// matchesExecutionFilter checks if an execution matches the given filter
func (we *DefaultWorkflowEngine) matchesExecutionFilter(execution *WorkflowExecution, filter *ExecutionFilter) bool {
	if filter == nil {
		return true
	}

	// Workflow ID filter
	if filter.WorkflowID != "" && execution.WorkflowID != filter.WorkflowID {
		return false
	}

	// Status filter
	if len(filter.Status) > 0 {
		found := false
		for _, status := range filter.Status {
			if execution.Status == status {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Date filters
	if filter.StartedAfter != nil && execution.StartedAt.Before(*filter.StartedAfter) {
		return false
	}
	if filter.StartedBefore != nil && execution.StartedAt.After(*filter.StartedBefore) {
		return false
	}

	return true
}

// matchesTemplateFilter checks if a template matches the given filter
func (we *DefaultWorkflowEngine) matchesTemplateFilter(template *WorkflowTemplate, filter *TemplateFilter) bool {
	if filter == nil {
		return true
	}

	// Category filter
	if filter.Category != "" && template.Category != filter.Category {
		return false
	}

	// Created by filter
	if filter.CreatedBy != "" && template.CreatedBy != filter.CreatedBy {
		return false
	}

	// Public filter
	if filter.IsPublic != nil && template.IsPublic != *filter.IsPublic {
		return false
	}

	// Tags filter
	if len(filter.Tags) > 0 {
		for _, filterTag := range filter.Tags {
			found := false
			for _, templateTag := range template.Tags {
				if templateTag == filterTag {
					found = true
					break
				}
			}
			if !found {
				return false
			}
		}
	}

	// Search filter
	if filter.Search != "" {
		searchLower := strings.ToLower(filter.Search)
		if !(strings.Contains(strings.ToLower(template.Name), searchLower) ||
			strings.Contains(strings.ToLower(template.Description), searchLower)) {
			return false
		}
	}

	return true
}

// createDefaultTemplates creates default workflow templates
func (we *DefaultWorkflowEngine) createDefaultTemplates() {
	// CI/CD Pipeline Template
	cicdTemplate := &WorkflowTemplate{
		ID:          uuid.New().String(),
		Name:        "CI/CD Pipeline",
		Description: "Continuous integration and deployment workflow",
		Category:    "DevOps",
		Tags:        []string{"ci", "cd", "deployment", "testing"},
		Workflow: &Workflow{
			Name:        "CI/CD Pipeline",
			Description: "Automated build, test, and deployment workflow",
			Version:     "1.0.0",
			Triggers: []*Trigger{
				{
					ID:   uuid.New().String(),
					Type: TriggerTypePush,
					Config: map[string]interface{}{
						"branches": []string{"main", "develop"},
					},
					Enabled: true,
				},
			},
			Actions: []*Action{
				{
					ID:          uuid.New().String(),
					Type:        ActionTypeScript,
					Name:        "Run Tests",
					Description: "Execute unit and integration tests",
					Config: map[string]interface{}{
						"script": "npm test",
					},
					Enabled: true,
				},
				{
					ID:          uuid.New().String(),
					Type:        ActionTypeScript,
					Name:        "Build Application",
					Description: "Build the application for deployment",
					Config: map[string]interface{}{
						"script": "npm run build",
					},
					Enabled: true,
				},
				{
					ID:          uuid.New().String(),
					Type:        ActionTypeWebhook,
					Name:        "Deploy to Production",
					Description: "Deploy application to production environment",
					Config: map[string]interface{}{
						"url":    "https://deploy.example.com/webhook",
						"method": "POST",
					},
					Enabled: true,
				},
			},
			Timeout: 30 * time.Minute,
		},
		IsPublic:  true,
		CreatedBy: "system",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	we.templates[cicdTemplate.ID] = cicdTemplate

	// Notification Template
	notificationTemplate := &WorkflowTemplate{
		ID:          uuid.New().String(),
		Name:        "Issue Notification",
		Description: "Notify team when issues are created or updated",
		Category:    "Notifications",
		Tags:        []string{"notification", "slack", "email", "issues"},
		Workflow: &Workflow{
			Name:        "Issue Notification",
			Description: "Send notifications for issue events",
			Version:     "1.0.0",
			Triggers: []*Trigger{
				{
					ID:   uuid.New().String(),
					Type: TriggerTypeEvent,
					Config: map[string]interface{}{
						"events": []string{"issue.created", "issue.updated"},
					},
					Enabled: true,
				},
			},
			Actions: []*Action{
				{
					ID:          uuid.New().String(),
					Type:        ActionTypeSlack,
					Name:        "Notify Slack",
					Description: "Send notification to Slack channel",
					Config: map[string]interface{}{
						"channel": "#development",
						"message": "Issue {{.issue.title}} has been {{.action}}",
					},
					Enabled: true,
				},
				{
					ID:          uuid.New().String(),
					Type:        ActionTypeEmail,
					Name:        "Send Email",
					Description: "Send email notification to assignee",
					Config: map[string]interface{}{
						"to":      "{{.issue.assignee.email}}",
						"subject": "Issue {{.issue.title}} - {{.action}}",
						"body":    "Issue details: {{.issue.description}}",
					},
					Enabled: true,
				},
			},
			Timeout: 5 * time.Minute,
		},
		IsPublic:  true,
		CreatedBy: "system",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	we.templates[notificationTemplate.ID] = notificationTemplate

	we.logger.WithField("templates", len(we.templates)).Info("Default workflow templates created")
}
