# AIOS Workflow Automation Engine

## Overview

The AIOS Workflow Automation Engine is a powerful, enterprise-grade automation platform that enables organizations to create, manage, and execute complex workflows with ease. Built with scalability, reliability, and extensibility in mind, it supports event-driven automation, CI/CD pipelines, and seamless integration with popular development tools.

## 🏗️ Architecture

### Core Components

```
Workflow Automation Engine
├── Workflow Engine (workflow management, execution orchestration)
├── Action Executor (HTTP, email, Slack, scripts, webhooks)
├── Trigger Manager (event processing, condition evaluation)
├── Schedule Manager (cron scheduling, interval-based execution)
├── Integration Manager (GitHub, GitLab, Jenkins, Slack)
├── Template System (reusable workflow templates)
└── Execution Workers (parallel processing, retry logic)
```

### Key Features

- **🔄 Event-Driven Automation**: React to GitHub pushes, issue updates, and custom events
- **⚡ Parallel Execution**: Multi-worker architecture for high-throughput processing
- **🔁 Retry Logic**: Configurable retry policies with exponential backoff
- **📋 Template System**: Reusable workflow templates for common patterns
- **🎯 Conditional Logic**: Advanced condition evaluation and branching
- **🔗 Rich Integrations**: GitHub, GitLab, Jenkins, Slack, email, webhooks
- **📊 Observability**: OpenTelemetry tracing and structured logging
- **⏰ Scheduling**: Cron-based and interval scheduling support

## 🚀 Quick Start

### Basic Workflow Creation

```go
package main

import (
    "context"
    "github.com/aios/aios/pkg/workflow"
    "github.com/sirupsen/logrus"
)

func main() {
    logger := logrus.New()
    
    // Create workflow engine
    config := &workflow.WorkflowEngineConfig{
        Workers:            5,
        QueueSize:          1000,
        DefaultTimeout:     30 * time.Minute,
        MaxRetries:         3,
        EnableScheduling:   true,
        EnableIntegrations: true,
    }
    
    engine, err := workflow.NewDefaultWorkflowEngine(config, logger)
    if err != nil {
        panic(err)
    }
    
    // Create a simple notification workflow
    notificationWorkflow := &workflow.Workflow{
        Name:        "Issue Notification",
        Description: "Send notifications when issues are created",
        Triggers: []*workflow.Trigger{
            {
                Type:  workflow.TriggerTypeEvent,
                Event: "issue.created",
                Enabled: true,
            },
        },
        Actions: []*workflow.Action{
            {
                Type: workflow.ActionTypeSlack,
                Name: "Notify Team",
                Config: map[string]interface{}{
                    "channel": "#development",
                    "message": "New issue: {{.issue.title}}",
                },
                Enabled: true,
            },
        },
    }
    
    // Create and activate workflow
    createdWorkflow, err := engine.CreateWorkflow(context.Background(), notificationWorkflow)
    if err != nil {
        panic(err)
    }
    
    createdWorkflow.Status = workflow.WorkflowStatusActive
    engine.UpdateWorkflow(context.Background(), createdWorkflow)
}
```

## 🔄 Workflow Components

### Workflows

A workflow defines the complete automation logic:

```go
workflow := &workflow.Workflow{
    Name:        "CI/CD Pipeline",
    Description: "Automated build and deployment",
    Version:     "1.0.0",
    Status:      workflow.WorkflowStatusActive,
    
    // Event triggers
    Triggers: []*workflow.Trigger{
        {
            Type:  workflow.TriggerTypePush,
            Event: "push",
            Conditions: []*workflow.Condition{
                {
                    Field:    "branch",
                    Operator: workflow.OperatorEquals,
                    Value:    "main",
                },
            },
            Enabled: true,
        },
    },
    
    // Actions to execute
    Actions: []*workflow.Action{
        {
            Type: workflow.ActionTypeScript,
            Name: "Run Tests",
            Config: map[string]interface{}{
                "script": "npm test",
            },
            Timeout: 10 * time.Minute,
            RetryPolicy: &workflow.RetryPolicy{
                MaxRetries:    2,
                RetryDelay:    30 * time.Second,
                BackoffFactor: 2.0,
            },
            Enabled: true,
        },
        {
            Type: workflow.ActionTypeSlack,
            Name: "Notify Success",
            Config: map[string]interface{}{
                "channel": "#deployments",
                "message": "✅ Build completed for {{.branch}}",
            },
            Enabled: true,
        },
    ],
    
    // Global settings
    Variables: map[string]interface{}{
        "environment": "production",
        "notify_team": true,
    },
    Timeout: 45 * time.Minute,
    
    // Metadata
    Tags:      []string{"ci", "cd", "automation"},
    CreatedBy: "devops-team@company.com",
}
```

### Triggers

Triggers define when workflows should execute:

**Event Triggers:**
```go
trigger := &workflow.Trigger{
    Type:  workflow.TriggerTypeEvent,
    Event: "issue.updated",
    Conditions: []*workflow.Condition{
        {
            Field:    "priority",
            Operator: workflow.OperatorIn,
            Value:    []interface{}{"high", "critical"},
        },
    },
    Enabled: true,
}
```

**Git Triggers:**
```go
pushTrigger := &workflow.Trigger{
    Type:  workflow.TriggerTypePush,
    Event: "push",
    Conditions: []*workflow.Condition{
        {
            Field:    "branch",
            Operator: workflow.OperatorEquals,
            Value:    "main",
        },
    },
    Enabled: true,
}

prTrigger := &workflow.Trigger{
    Type:  workflow.TriggerTypePR,
    Event: "pull_request",
    Conditions: []*workflow.Condition{
        {
            Field:    "action",
            Operator: workflow.OperatorIn,
            Value:    []interface{}{"opened", "synchronize"},
        },
    },
    Enabled: true,
}
```

**Schedule Triggers:**
```go
scheduleTrigger := &workflow.Trigger{
    Type: workflow.TriggerTypeSchedule,
    Config: map[string]interface{}{
        "cron": "0 9 * * MON-FRI", // 9 AM weekdays
    },
    Enabled: true,
}
```

### Actions

Actions define what the workflow should do:

**HTTP/Webhook Actions:**
```go
httpAction := &workflow.Action{
    Type: workflow.ActionTypeHTTP,
    Name: "Deploy to Production",
    Config: map[string]interface{}{
        "url":    "https://deploy.company.com/api/deploy",
        "method": "POST",
        "headers": map[string]interface{}{
            "Authorization": "Bearer {{.deploy_token}}",
            "Content-Type":  "application/json",
        },
        "payload": map[string]interface{}{
            "branch":      "{{.branch}}",
            "commit_sha":  "{{.commit_sha}}",
            "environment": "production",
        },
    },
    Timeout: 15 * time.Minute,
    Enabled: true,
}
```

**Script Actions:**
```go
scriptAction := &workflow.Action{
    Type: workflow.ActionTypeScript,
    Name: "Run Integration Tests",
    Config: map[string]interface{}{
        "script": `
            echo "Running integration tests..."
            npm run test:integration
            echo "Tests completed with exit code: $?"
        `,
        "shell": "/bin/bash",
    },
    Timeout: 20 * time.Minute,
    RetryPolicy: &workflow.RetryPolicy{
        MaxRetries:    2,
        RetryDelay:    1 * time.Minute,
        BackoffFactor: 2.0,
    },
    Enabled: true,
}
```

**Notification Actions:**
```go
slackAction := &workflow.Action{
    Type: workflow.ActionTypeSlack,
    Name: "Notify Team",
    Config: map[string]interface{}{
        "channel": "#development",
        "message": `
🚀 Deployment Status: {{.status}}
📦 Version: {{.version}}
🌿 Branch: {{.branch}}
👤 Author: {{.author}}
⏰ Duration: {{.duration}}
        `,
    },
    Enabled: true,
}

emailAction := &workflow.Action{
    Type: workflow.ActionTypeEmail,
    Name: "Email Stakeholders",
    Config: map[string]interface{}{
        "to":      []string{"stakeholders@company.com"},
        "subject": "Deployment Completed - {{.version}}",
        "body":    "The deployment of version {{.version}} has been completed successfully.",
    },
    Conditions: []*workflow.Condition{
        {
            Field:    "environment",
            Operator: workflow.OperatorEquals,
            Value:    "production",
        },
    },
    Enabled: true,
}
```

### Conditions

Conditions provide powerful logic for controlling workflow execution:

**Comparison Operators:**
```go
conditions := []*workflow.Condition{
    {
        Field:    "priority",
        Operator: workflow.OperatorEquals,
        Value:    "high",
    },
    {
        Field:    "assignee_count",
        Operator: workflow.OperatorGreaterThan,
        Value:    0,
    },
    {
        Field:    "title",
        Operator: workflow.OperatorContains,
        Value:    "urgent",
    },
    {
        Field:    "labels",
        Operator: workflow.OperatorIn,
        Value:    []interface{}{"bug", "security"},
    },
    {
        Field:    "description",
        Operator: workflow.OperatorRegex,
        Value:    "^(fix|bug|hotfix):",
    },
}
```

## 🎯 Advanced Features

### Template System

Create reusable workflow templates:

```go
// Create template
template := &workflow.WorkflowTemplate{
    Name:        "Standard CI/CD",
    Description: "Standard CI/CD pipeline template",
    Category:    "DevOps",
    Tags:        []string{"ci", "cd", "template"},
    Workflow: &workflow.Workflow{
        Name: "CI/CD Pipeline",
        Triggers: []*workflow.Trigger{
            {
                Type:  workflow.TriggerTypePush,
                Event: "push",
                Enabled: true,
            },
        },
        Actions: []*workflow.Action{
            {
                Type: workflow.ActionTypeScript,
                Name: "Build and Test",
                Config: map[string]interface{}{
                    "script": "npm ci && npm test && npm run build",
                },
                Enabled: true,
            },
        },
    },
    IsPublic:  true,
    CreatedBy: "platform-team",
}

createdTemplate, err := engine.CreateTemplate(ctx, template)

// Use template to create workflow
newWorkflow, err := engine.CreateWorkflowFromTemplate(ctx, createdTemplate.ID, &workflow.Workflow{
    Name:        "My Custom Pipeline",
    Description: "Customized CI/CD for my project",
    CreatedBy:   "developer@company.com",
})
```

### Event Processing

Trigger workflows with events:

```go
// Create event
event := &workflow.Event{
    Type:      "issue.created",
    Source:    "github",
    Subject:   "repository/issues/123",
    Data: map[string]interface{}{
        "issue": map[string]interface{}{
            "id":       123,
            "title":    "Critical bug in authentication",
            "priority": "high",
            "assignee": "security-team@company.com",
        },
        "repository": "my-app",
        "author":     "reporter@company.com",
    },
    Timestamp: time.Now(),
    UserID:    "reporter@company.com",
}

// Trigger workflows
executions, err := engine.TriggerWorkflow(ctx, event)
```

### Manual Execution

Execute workflows manually with custom input:

```go
execution, err := engine.ExecuteWorkflow(ctx, workflowID, map[string]interface{}{
    "branch":      "feature/new-feature",
    "commit_sha":  "abc123def456",
    "environment": "staging",
    "manual":      true,
    "triggered_by": "developer@company.com",
})

// Monitor execution
for {
    updatedExecution, err := engine.GetExecution(ctx, execution.ID)
    if err != nil {
        break
    }
    
    if updatedExecution.Status == workflow.ExecutionStatusSuccess ||
       updatedExecution.Status == workflow.ExecutionStatusFailure {
        break
    }
    
    time.Sleep(5 * time.Second)
}
```

## 🔗 Integrations

### GitHub Integration

```go
// Connect GitHub
githubConfig := &workflow.GitHubConfig{
    Token:        "ghp_xxxxxxxxxxxx",
    Organization: "my-company",
}

err := integrationManager.ConnectGitHub(ctx, githubConfig)

// Create webhook
err = integrationManager.CreateGitHubWebhook(ctx, 
    "https://github.com/my-company/my-repo", 
    []string{"push", "pull_request", "issues"})
```

### Slack Integration

```go
// Connect Slack
slackConfig := &workflow.SlackConfig{
    Token:     "xoxb-xxxxxxxxxxxx",
    Channel:   "#general",
    Username:  "AIOS Bot",
    IconEmoji: ":robot_face:",
}

err := integrationManager.ConnectSlack(ctx, slackConfig)

// Send message
err = integrationManager.SendSlackMessage(ctx, "#deployments", 
    "🚀 Deployment completed successfully!")
```

### Jenkins Integration

```go
// Connect Jenkins
jenkinsConfig := &workflow.JenkinsConfig{
    URL:      "https://jenkins.company.com",
    Username: "api-user",
    Token:    "api-token",
}

err := integrationManager.ConnectJenkins(ctx, jenkinsConfig)

// Trigger job
err = integrationManager.TriggerJenkinsJob(ctx, "deploy-production", map[string]interface{}{
    "BRANCH":      "main",
    "ENVIRONMENT": "production",
})

// Check status
status, err := integrationManager.GetJenkinsJobStatus(ctx, "deploy-production", 42)
```

## 📊 Monitoring and Observability

### Execution Monitoring

```go
// List executions
executions, err := engine.ListExecutions(ctx, &workflow.ExecutionFilter{
    WorkflowID: workflowID,
    Status:     []workflow.ExecutionStatus{workflow.ExecutionStatusRunning},
    Limit:      10,
})

// Get execution details
execution, err := engine.GetExecution(ctx, executionID)

fmt.Printf("Execution Status: %s\n", execution.Status)
fmt.Printf("Duration: %s\n", execution.Duration)
fmt.Printf("Steps: %d\n", len(execution.Steps))

for _, step := range execution.Steps {
    fmt.Printf("  - %s: %s (%s)\n", step.ActionType, step.Status, step.Duration)
}
```

### Workflow Analytics

```go
// List workflows with filters
workflows, err := engine.ListWorkflows(ctx, &workflow.WorkflowFilter{
    Status:    []workflow.WorkflowStatus{workflow.WorkflowStatusActive},
    CreatedBy: "devops-team@company.com",
    Tags:      []string{"ci", "cd"},
    Search:    "deployment",
    Limit:     20,
})

// Get workflow metrics
for _, wf := range workflows {
    executions, _ := engine.ListExecutions(ctx, &workflow.ExecutionFilter{
        WorkflowID: wf.ID,
        Limit:      100,
    })
    
    var successful, failed int
    for _, exec := range executions {
        if exec.Status == workflow.ExecutionStatusSuccess {
            successful++
        } else if exec.Status == workflow.ExecutionStatusFailure {
            failed++
        }
    }
    
    successRate := float64(successful) / float64(len(executions)) * 100
    fmt.Printf("Workflow: %s - Success Rate: %.1f%%\n", wf.Name, successRate)
}
```

## 🧪 Testing

Run the comprehensive test suite:

```bash
# Run all workflow tests
go test ./pkg/workflow/...

# Run with race detection
go test -race ./pkg/workflow/...

# Run integration tests
go test -tags=integration ./pkg/workflow/...

# Run example
go run examples/workflow_automation_example.go
```

## 📖 Examples

See the complete example in `examples/workflow_automation_example.go` for a comprehensive demonstration including:

- CI/CD pipeline automation
- Issue notification workflows
- Event-driven triggers
- Manual workflow execution
- Template usage
- Integration examples

## 🤝 Contributing

1. Follow established patterns and interfaces
2. Add comprehensive tests for new features
3. Update documentation for API changes
4. Ensure proper error handling and logging
5. Use OpenTelemetry for observability

## 📄 License

This workflow automation engine is part of the AIOS project and follows the same licensing terms.
