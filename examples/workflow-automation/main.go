package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aios/aios/pkg/workflow"
	"github.com/sirupsen/logrus"
)

func main() {
	// Initialize logger
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)
	logger.SetFormatter(&logrus.JSONFormatter{})

	fmt.Println("🔄 AIOS Workflow Automation Engine Demo")
	fmt.Println("======================================")

	// Run the comprehensive demo
	if err := runWorkflowAutomationDemo(logger); err != nil {
		log.Fatalf("Demo failed: %v", err)
	}

	fmt.Println("\n✅ Workflow Automation Engine Demo completed successfully!")
}

func runWorkflowAutomationDemo(logger *logrus.Logger) error {
	ctx := context.Background()

	// Step 1: Create Workflow Engine
	fmt.Println("\n1. Creating Workflow Engine...")
	config := &workflow.WorkflowEngineConfig{
		Workers:            3,
		QueueSize:          100,
		DefaultTimeout:     30 * time.Minute,
		MaxRetries:         3,
		EnableScheduling:   true,
		EnableIntegrations: true,
	}

	workflowEngine, err := workflow.NewDefaultWorkflowEngine(config, logger)
	if err != nil {
		return fmt.Errorf("failed to create workflow engine: %w", err)
	}
	fmt.Println("✓ Workflow Engine created successfully")

	// Step 2: Create CI/CD Workflow
	fmt.Println("\n2. Creating CI/CD Workflow...")
	cicdWorkflow := &workflow.Workflow{
		Name:        "CI/CD Pipeline",
		Description: "Automated continuous integration and deployment workflow",
		CreatedBy:   "devops-team@company.com",
		Tags:        []string{"ci", "cd", "automation", "deployment"},
		Triggers: []*workflow.Trigger{
			{
				Type:  workflow.TriggerTypePush,
				Event: "push",
				Conditions: []*workflow.Condition{
					{
						Field:    "branch",
						Operator: workflow.OperatorEquals,
						Value:    "main",
						Type:     workflow.ValueTypeString,
					},
				},
				Enabled: true,
			},
			{
				Type:  workflow.TriggerTypePR,
				Event: "pull_request",
				Conditions: []*workflow.Condition{
					{
						Field:    "action",
						Operator: workflow.OperatorIn,
						Value:    []interface{}{"opened", "synchronize"},
						Type:     workflow.ValueTypeArray,
					},
				},
				Enabled: true,
			},
		},
		Actions: []*workflow.Action{
			{
				Type:        workflow.ActionTypeScript,
				Name:        "Run Unit Tests",
				Description: "Execute unit test suite",
				Config: map[string]interface{}{
					"script": "npm test",
					"shell":  "/bin/bash",
				},
				Timeout: 10 * time.Minute,
				RetryPolicy: &workflow.RetryPolicy{
					MaxRetries:    2,
					RetryDelay:    30 * time.Second,
					BackoffFactor: 2.0,
					MaxDelay:      5 * time.Minute,
				},
				Enabled: true,
			},
			{
				Type:        workflow.ActionTypeScript,
				Name:        "Build Application",
				Description: "Build the application for deployment",
				Config: map[string]interface{}{
					"script": "npm run build",
					"shell":  "/bin/bash",
				},
				Timeout: 15 * time.Minute,
				Enabled: true,
			},
			{
				Type:        workflow.ActionTypeSlack,
				Name:        "Notify Team",
				Description: "Send build notification to Slack",
				Config: map[string]interface{}{
					"channel": "#deployments",
					"message": "🚀 Build completed for {{.branch}} - {{.commit_sha}}",
				},
				Enabled: true,
			},
			{
				Type:        workflow.ActionTypeWebhook,
				Name:        "Deploy to Staging",
				Description: "Deploy application to staging environment",
				Config: map[string]interface{}{
					"url":    "https://deploy.company.com/staging",
					"method": "POST",
					"headers": map[string]interface{}{
						"Authorization": "Bearer staging-token",
						"Content-Type":  "application/json",
					},
					"payload": map[string]interface{}{
						"branch":      "{{.branch}}",
						"commit_sha":  "{{.commit_sha}}",
						"environment": "staging",
					},
				},
				Conditions: []*workflow.Condition{
					{
						Field:    "branch",
						Operator: workflow.OperatorEquals,
						Value:    "main",
						Type:     workflow.ValueTypeString,
					},
				},
				Enabled: true,
			},
		},
		Variables: map[string]interface{}{
			"environment": "staging",
			"notify_team": true,
		},
		Timeout: 45 * time.Minute,
		RetryPolicy: &workflow.RetryPolicy{
			MaxRetries:    1,
			RetryDelay:    5 * time.Minute,
			BackoffFactor: 1.0,
		},
	}

	createdCICD, err := workflowEngine.CreateWorkflow(ctx, cicdWorkflow)
	if err != nil {
		return fmt.Errorf("failed to create CI/CD workflow: %w", err)
	}
	fmt.Printf("✓ CI/CD Workflow created: %s (ID: %s)\n", createdCICD.Name, createdCICD.ID)

	// Step 3: Create Issue Notification Workflow
	fmt.Println("\n3. Creating Issue Notification Workflow...")
	issueWorkflow := &workflow.Workflow{
		Name:        "Issue Notification System",
		Description: "Automated notifications for issue management",
		CreatedBy:   "project-manager@company.com",
		Tags:        []string{"notifications", "issues", "team-communication"},
		Triggers: []*workflow.Trigger{
			{
				Type:    workflow.TriggerTypeEvent,
				Event:   "issue.created",
				Enabled: true,
			},
			{
				Type:  workflow.TriggerTypeEvent,
				Event: "issue.updated",
				Conditions: []*workflow.Condition{
					{
						Field:    "priority",
						Operator: workflow.OperatorIn,
						Value:    []interface{}{"high", "critical"},
						Type:     workflow.ValueTypeArray,
					},
				},
				Enabled: true,
			},
		},
		Actions: []*workflow.Action{
			{
				Type:        workflow.ActionTypeSlack,
				Name:        "Notify Development Team",
				Description: "Send issue notification to development channel",
				Config: map[string]interface{}{
					"channel": "#development",
					"message": "📋 New issue: {{.issue.title}} (Priority: {{.issue.priority}})\nAssigned to: {{.issue.assignee}}\nURL: {{.issue.url}}",
				},
				Enabled: true,
			},
			{
				Type:        workflow.ActionTypeEmail,
				Name:        "Email Assignee",
				Description: "Send email notification to issue assignee",
				Config: map[string]interface{}{
					"to":      "{{.issue.assignee_email}}",
					"subject": "Issue Assigned: {{.issue.title}}",
					"body":    "You have been assigned a new issue:\n\nTitle: {{.issue.title}}\nDescription: {{.issue.description}}\nPriority: {{.issue.priority}}\nDue Date: {{.issue.due_date}}\n\nPlease review and update the status accordingly.",
				},
				Conditions: []*workflow.Condition{
					{
						Field:    "issue.assignee_email",
						Operator: workflow.OperatorNotEquals,
						Value:    "",
						Type:     workflow.ValueTypeString,
					},
				},
				Enabled: true,
			},
			{
				Type:        workflow.ActionTypeWebhook,
				Name:        "Update Project Dashboard",
				Description: "Update external project dashboard",
				Config: map[string]interface{}{
					"url":    "https://dashboard.company.com/api/issues",
					"method": "POST",
					"payload": map[string]interface{}{
						"issue_id":   "{{.issue.id}}",
						"title":      "{{.issue.title}}",
						"priority":   "{{.issue.priority}}",
						"assignee":   "{{.issue.assignee}}",
						"created_at": "{{.issue.created_at}}",
					},
				},
				Enabled: true,
			},
		},
		Timeout: 5 * time.Minute,
	}

	createdIssue, err := workflowEngine.CreateWorkflow(ctx, issueWorkflow)
	if err != nil {
		return fmt.Errorf("failed to create issue workflow: %w", err)
	}
	fmt.Printf("✓ Issue Notification Workflow created: %s (ID: %s)\n", createdIssue.Name, createdIssue.ID)

	// Step 4: Activate Workflows
	fmt.Println("\n4. Activating Workflows...")

	// Activate CI/CD workflow
	createdCICD.Status = workflow.WorkflowStatusActive
	_, err = workflowEngine.UpdateWorkflow(ctx, createdCICD)
	if err != nil {
		fmt.Printf("   ⚠️  Failed to activate CI/CD workflow: %v\n", err)
	} else {
		fmt.Printf("   ✓ CI/CD Workflow activated\n")
	}

	// Activate issue workflow
	createdIssue.Status = workflow.WorkflowStatusActive
	_, err = workflowEngine.UpdateWorkflow(ctx, createdIssue)
	if err != nil {
		fmt.Printf("   ⚠️  Failed to activate issue workflow: %v\n", err)
	} else {
		fmt.Printf("   ✓ Issue Notification Workflow activated\n")
	}

	// Step 5: Simulate Workflow Triggers
	fmt.Println("\n5. Simulating Workflow Triggers...")

	// Simulate push event
	pushEvent := &workflow.Event{
		Type:    "push",
		Source:  "github",
		Subject: "repository/main",
		Data: map[string]interface{}{
			"branch":     "main",
			"commit_sha": "abc123def456",
			"author":     "developer@company.com",
			"message":    "Fix critical bug in authentication",
		},
		Timestamp: time.Now(),
		UserID:    "developer@company.com",
	}

	executions, err := workflowEngine.TriggerWorkflow(ctx, pushEvent)
	if err != nil {
		fmt.Printf("   ⚠️  Failed to trigger workflows for push event: %v\n", err)
	} else {
		fmt.Printf("   ✓ Push event triggered %d workflow executions\n", len(executions))
	}

	// Simulate issue creation event
	issueEvent := &workflow.Event{
		Type:    "issue.created",
		Source:  "project-management",
		Subject: "issue/123",
		Data: map[string]interface{}{
			"issue": map[string]interface{}{
				"id":             "123",
				"title":          "Critical security vulnerability",
				"description":    "SQL injection vulnerability found in user authentication",
				"priority":       "critical",
				"assignee":       "security-team@company.com",
				"assignee_email": "security-team@company.com",
				"created_at":     time.Now().Format(time.RFC3339),
				"due_date":       time.Now().AddDate(0, 0, 1).Format(time.RFC3339),
				"url":            "https://issues.company.com/123",
			},
		},
		Timestamp: time.Now(),
		UserID:    "project-manager@company.com",
	}

	issueExecutions, err := workflowEngine.TriggerWorkflow(ctx, issueEvent)
	if err != nil {
		fmt.Printf("   ⚠️  Failed to trigger workflows for issue event: %v\n", err)
	} else {
		fmt.Printf("   ✓ Issue event triggered %d workflow executions\n", len(issueExecutions))
	}

	// Step 6: Execute Workflow Manually
	fmt.Println("\n6. Executing Workflow Manually...")

	manualExecution, err := workflowEngine.ExecuteWorkflow(ctx, createdCICD.ID, map[string]interface{}{
		"branch":     "feature/new-feature",
		"commit_sha": "xyz789abc123",
		"author":     "developer2@company.com",
		"trigger":    "manual",
	})
	if err != nil {
		fmt.Printf("   ⚠️  Failed to execute workflow manually: %v\n", err)
	} else {
		fmt.Printf("   ✓ Manual execution started: %s\n", manualExecution.ID)
	}

	// Step 7: Wait for executions to complete and check status
	fmt.Println("\n7. Checking Execution Status...")
	time.Sleep(2 * time.Second) // Give some time for processing

	allExecutions := append(executions, issueExecutions...)
	if manualExecution != nil {
		allExecutions = append(allExecutions, manualExecution)
	}

	for i, execution := range allExecutions {
		updatedExecution, err := workflowEngine.GetExecution(ctx, execution.ID)
		if err != nil {
			fmt.Printf("   ⚠️  Failed to get execution %d status: %v\n", i+1, err)
			continue
		}

		fmt.Printf("   ✓ Execution %d (%s): %s\n", i+1, updatedExecution.ID[:8], updatedExecution.Status)
		if updatedExecution.Error != "" {
			fmt.Printf("     Error: %s\n", updatedExecution.Error)
		}
		if len(updatedExecution.Steps) > 0 {
			fmt.Printf("     Steps completed: %d\n", len(updatedExecution.Steps))
		}
	}

	// Step 8: List All Workflows
	fmt.Println("\n8. Listing All Workflows...")
	workflows, err := workflowEngine.ListWorkflows(ctx, &workflow.WorkflowFilter{
		Limit: 10,
	})
	if err != nil {
		fmt.Printf("   ⚠️  Failed to list workflows: %v\n", err)
	} else {
		fmt.Printf("   ✓ Found %d workflows:\n", len(workflows))
		for _, wf := range workflows {
			fmt.Printf("     - %s (%s) - %s [%d triggers, %d actions]\n",
				wf.Name, wf.Status, wf.CreatedBy, len(wf.Triggers), len(wf.Actions))
		}
	}

	// Step 9: List Workflow Templates
	fmt.Println("\n9. Listing Workflow Templates...")
	templates, err := workflowEngine.ListTemplates(ctx, &workflow.TemplateFilter{
		Limit: 10,
	})
	if err != nil {
		fmt.Printf("   ⚠️  Failed to list templates: %v\n", err)
	} else {
		fmt.Printf("   ✓ Found %d templates:\n", len(templates))
		for _, template := range templates {
			fmt.Printf("     - %s (%s) - %s [Usage: %d]\n",
				template.Name, template.Category, template.CreatedBy, template.UsageCount)
		}
	}

	// Step 10: Create Workflow from Template
	fmt.Println("\n10. Creating Workflow from Template...")
	if len(templates) > 0 {
		templateWorkflow, err := workflowEngine.CreateWorkflowFromTemplate(ctx, templates[0].ID, &workflow.Workflow{
			Name:        "Custom CI/CD Pipeline",
			Description: "Customized CI/CD pipeline based on template",
			CreatedBy:   "devops-lead@company.com",
			Tags:        []string{"custom", "ci", "cd"},
		})
		if err != nil {
			fmt.Printf("   ⚠️  Failed to create workflow from template: %v\n", err)
		} else {
			fmt.Printf("   ✓ Workflow created from template: %s (ID: %s)\n",
				templateWorkflow.Name, templateWorkflow.ID)
		}
	}

	return nil
}
