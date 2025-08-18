# AIOS Project Management & Workflow Integration

## Overview

The AIOS Project Management & Workflow Integration system provides comprehensive project management capabilities, workflow automation, and CI/CD pipeline integration. Built with enterprise-grade features, it supports agile methodologies, team collaboration, and seamless integration with development tools.

## 🏗️ Architecture

### Core Components

```
Project Management System
├── Project Manager (projects, tasks, milestones, sprints)
├── Task Dependency Manager (dependencies, critical path, scheduling)
├── Notification Manager (real-time notifications, subscriptions)
├── Template Manager (project templates, reusable workflows)
├── Time Tracking Manager (time logging, reporting, billing)
├── Integration Manager (GitHub, GitLab, Jenkins, Slack)
├── Workflow Engine (automation, triggers, actions)
├── Pipeline Engine (CI/CD, builds, deployments)
└── Analytics & Reporting (metrics, insights, dashboards)
```

### Key Features

- **📋 Project Management**: Complete project lifecycle management
- **✅ Task Management**: Advanced task tracking with dependencies
- **🏃‍♂️ Agile Support**: Sprints, milestones, velocity tracking
- **👥 Team Collaboration**: Role-based access, notifications, comments
- **⏱️ Time Tracking**: Detailed time logging and reporting
- **🔄 Workflow Automation**: Custom workflows and triggers
- **🚀 CI/CD Integration**: Pipeline management and automation
- **📊 Analytics & Reporting**: Comprehensive project insights
- **🔗 External Integrations**: GitHub, GitLab, Jenkins, Slack, Email

## 🚀 Quick Start

### Basic Project Management

```go
package main

import (
    "context"
    "github.com/aios/aios/pkg/project"
    "github.com/sirupsen/logrus"
)

func main() {
    logger := logrus.New()
    
    // Create project manager
    config := &project.ProjectManagerConfig{
        EnableNotifications: true,
        EnableTimeTracking:  true,
        EnableIntegrations:  true,
        MaxProjectsPerUser:  100,
        MaxTasksPerProject:  1000,
    }
    
    projectManager, err := project.NewDefaultProjectManager(config, logger)
    if err != nil {
        panic(err)
    }
    
    // Create a new project
    newProject := &project.Project{
        Name:        "My Awesome Project",
        Description: "A revolutionary new application",
        Owner:       "john.doe@example.com",
        TeamMembers: []string{"jane.smith@example.com"},
        Priority:    project.PriorityHigh,
        Tags:        []string{"web", "api", "microservices"},
    }
    
    project, err := projectManager.CreateProject(context.Background(), newProject)
    if err != nil {
        panic(err)
    }
    
    // Create tasks
    task := &project.Task{
        ProjectID:   project.ID,
        Title:       "Implement user authentication",
        Description: "Build secure JWT-based authentication",
        Type:        project.TaskTypeFeature,
        Priority:    project.PriorityHigh,
        Assignee:    "jane.smith@example.com",
        Labels:      []string{"auth", "security"},
    }
    
    createdTask, err := projectManager.CreateTask(context.Background(), task)
    if err != nil {
        panic(err)
    }
}
```

## 📋 Project Management Features

### Projects

Complete project lifecycle management with:

```go
// Create project with full configuration
project := &project.Project{
    Name:        "E-commerce Platform",
    Description: "Modern e-commerce solution with microservices",
    Owner:       "tech-lead@company.com",
    TeamMembers: []string{"dev1@company.com", "dev2@company.com"},
    Priority:    project.PriorityHigh,
    Status:      project.ProjectStatusActive,
    Tags:        []string{"ecommerce", "microservices", "react"},
    Repository: &project.Repository{
        URL:      "https://github.com/company/ecommerce",
        Branch:   "main",
        Provider: "github",
    },
    Settings: &project.ProjectSettings{
        Visibility:      project.VisibilityPrivate,
        RequireApproval: true,
        AutoAssignment:  false,
        Notifications: &project.NotificationSettings{
            Email: true,
            InApp: true,
            Slack: true,
        },
    },
}

createdProject, err := projectManager.CreateProject(ctx, project)
```

**Project Features:**
- **Lifecycle Management**: Planning, active, on-hold, completed, archived
- **Team Management**: Role-based access control and permissions
- **Repository Integration**: GitHub, GitLab, Bitbucket connectivity
- **Custom Settings**: Visibility, approval workflows, notifications
- **Tagging & Categorization**: Flexible organization system

### Tasks

Advanced task management with dependencies:

```go
// Create task with dependencies
task := &project.Task{
    ProjectID:      projectID,
    Title:          "API Integration Tests",
    Description:    "Comprehensive integration test suite for REST API",
    Type:           project.TaskTypeTask,
    Priority:       project.PriorityMedium,
    Assignee:       "qa-engineer@company.com",
    Reporter:       "tech-lead@company.com",
    Labels:         []string{"testing", "api", "integration"},
    Dependencies:   []string{"api-implementation-task-id"},
    EstimatedHours: 16.0,
    DueDate:        &dueDate,
}

// Add task dependencies
err := dependencyManager.AddDependency(ctx, taskID, dependsOnTaskID)

// Get critical path
criticalPath, err := dependencyManager.GetCriticalPath(ctx, projectID)
```

**Task Features:**
- **Types**: Feature, Bug, Task, Story, Epic, Spike
- **Dependencies**: Task relationships and critical path analysis
- **Time Tracking**: Estimated vs actual hours
- **Status Workflow**: Todo → In Progress → Review → Done
- **Comments & Attachments**: Rich collaboration features

### Sprints & Milestones

Agile methodology support:

```go
// Create sprint
sprint := &project.Sprint{
    ProjectID: projectID,
    Name:      "Sprint 1 - Foundation",
    Goal:      "Establish core architecture and authentication",
    StartDate: time.Now(),
    EndDate:   time.Now().AddDate(0, 0, 14), // 2 weeks
    Capacity:  40, // Story points
    Tasks:     []string{task1ID, task2ID, task3ID},
}

createdSprint, err := projectManager.CreateSprint(ctx, sprint)

// Start sprint
err = projectManager.StartSprint(ctx, sprint.ID)

// Create milestone
milestone := &project.Milestone{
    ProjectID:   projectID,
    Name:        "MVP Release",
    Description: "Minimum viable product ready for beta testing",
    DueDate:     &mvpDate,
    Tasks:       []string{task1ID, task2ID},
}
```

## ⏱️ Time Tracking

Comprehensive time tracking and reporting:

```go
// Start timer
timer, err := timeTracker.StartTimer(ctx, taskID, userID)

// Log time manually
entry := &project.TimeEntry{
    TaskID:      taskID,
    UserID:      userID,
    Description: "Implemented user authentication endpoints",
    StartTime:   startTime,
    EndTime:     &endTime,
    IsBillable:  true,
    HourlyRate:  75.0,
    Tags:        []string{"development", "backend"},
}

loggedEntry, err := timeTracker.LogTime(ctx, entry)

// Generate time report
reportRequest := &project.TimeReportRequest{
    ProjectID: projectID,
    StartDate: time.Now().AddDate(0, 0, -30), // Last 30 days
    EndDate:   time.Now(),
    GroupBy:   "user",
    Format:    project.ReportFormatPDF,
}

report, err := timeTracker.GetTimeReport(ctx, reportRequest)
```

## 🔄 Workflow Automation

Powerful workflow automation engine:

```go
// Create workflow
workflow := &workflow.Workflow{
    Name:        "Code Review Workflow",
    Description: "Automated code review process",
    Triggers: []*workflow.Trigger{
        {
            Type:  workflow.TriggerTypeEvent,
            Event: "pull_request.opened",
            Conditions: []*workflow.Condition{
                {
                    Field:    "repository",
                    Operator: workflow.OperatorEquals,
                    Value:    "main-repo",
                },
            },
        },
    },
    Actions: []*workflow.Action{
        {
            Type: workflow.ActionTypeSlack,
            Name: "Notify Team",
            Config: map[string]interface{}{
                "channel": "#code-review",
                "message": "New PR ready for review: {{.pull_request.title}}",
            },
        },
        {
            Type: workflow.ActionTypeHTTP,
            Name: "Trigger CI",
            Config: map[string]interface{}{
                "url":    "https://ci.company.com/trigger",
                "method": "POST",
                "headers": map[string]string{
                    "Authorization": "Bearer {{.ci_token}}",
                },
            },
        },
    },
}

createdWorkflow, err := workflowEngine.CreateWorkflow(ctx, workflow)
```

## 🚀 CI/CD Pipeline Integration

Comprehensive pipeline management:

```go
// Create CI/CD pipeline
pipeline := &workflow.Pipeline{
    Name:        "Production Deployment",
    Description: "Automated production deployment pipeline",
    Repository: &workflow.Repository{
        URL:      "https://github.com/company/app",
        Branch:   "main",
        Provider: "github",
    },
    Stages: []*workflow.Stage{
        {
            Name: "Build",
            Jobs: []*workflow.Job{
                {
                    Name: "Compile",
                    Steps: []*workflow.Step{
                        {
                            Name:    "Checkout Code",
                            Type:    workflow.StepTypeAction,
                            Action:  "actions/checkout@v3",
                        },
                        {
                            Name:    "Build Application",
                            Type:    workflow.StepTypeCommand,
                            Command: "go build -o app ./cmd/server",
                        },
                    },
                },
            },
        },
        {
            Name: "Test",
            DependsOn: []string{"Build"},
            Jobs: []*workflow.Job{
                {
                    Name: "Unit Tests",
                    Steps: []*workflow.Step{
                        {
                            Name:    "Run Tests",
                            Type:    workflow.StepTypeCommand,
                            Command: "go test -v ./...",
                        },
                    },
                },
            },
        },
        {
            Name: "Deploy",
            DependsOn: []string{"Test"},
            Jobs: []*workflow.Job{
                {
                    Name: "Production Deploy",
                    Steps: []*workflow.Step{
                        {
                            Name:    "Deploy to Kubernetes",
                            Type:    workflow.StepTypeScript,
                            Script:  "kubectl apply -f k8s/",
                        },
                    },
                },
            },
        },
    },
    Triggers: []*workflow.PipelineTrigger{
        {
            Type:     workflow.TriggerTypePush,
            Branches: []string{"main"},
        },
    },
}

createdPipeline, err := pipelineEngine.CreatePipeline(ctx, pipeline)

// Execute pipeline
execution, err := pipelineEngine.ExecutePipeline(ctx, pipeline.ID, &workflow.ExecutionParams{
    Branch: "main",
    Variables: map[string]interface{}{
        "ENVIRONMENT": "production",
        "VERSION":     "v1.2.3",
    },
})
```

## 🔗 External Integrations

Seamless integration with popular tools:

### GitHub Integration

```go
// Connect GitHub
githubConfig := &workflow.GitHubConfig{
    Token:        "ghp_xxxxxxxxxxxx",
    Organization: "my-company",
}

err := integrationManager.ConnectGitHub(ctx, githubConfig)

// Create webhook
err = integrationManager.CreateGitHubWebhook(ctx, repoURL, []string{
    "push", "pull_request", "issues",
})

// Sync repository
err = integrationManager.SyncGitHubRepository(ctx, repoURL)
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
err = integrationManager.SendSlackMessage(ctx, "#dev-team", 
    "🚀 Deployment to production completed successfully!")
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
```

## 📊 Analytics & Reporting

Comprehensive project insights:

```go
// Get project analytics
timeRange := &project.TimeRange{
    Start: time.Now().AddDate(0, -1, 0), // Last month
    End:   time.Now(),
}

analytics, err := projectManager.GetProjectAnalytics(ctx, projectID, timeRange)

// Access metrics
fmt.Printf("Tasks Completed: %d\n", analytics.TasksCompleted)
fmt.Printf("Velocity: %.1f tasks/week\n", analytics.Velocity)
fmt.Printf("Success Rate: %.1f%%\n", analytics.SuccessRate*100)

// Generate custom report
reportRequest := &project.ReportRequest{
    Type:      project.ReportTypeVelocity,
    ProjectID: projectID,
    TimeRange: timeRange,
    Format:    project.ReportFormatPDF,
    Filters: map[string]interface{}{
        "team_members": []string{"dev1@company.com", "dev2@company.com"},
        "task_types":   []string{"feature", "bug"},
    },
}

report, err := projectManager.GenerateReport(ctx, reportRequest)
```

**Available Analytics:**
- **Project Metrics**: Velocity, burndown, completion rates
- **Team Performance**: Workload distribution, productivity
- **Time Analysis**: Time spent by category, billable hours
- **Quality Metrics**: Bug rates, code review metrics
- **Trend Analysis**: Historical performance trends

## 🧪 Testing

Run the comprehensive test suite:

```bash
# Run all project management tests
go test ./pkg/project/...

# Run workflow tests
go test ./pkg/workflow/...

# Run integration tests
go test -tags=integration ./pkg/project/...

# Run example
go run examples/project_workflow_example.go
```

## 📖 Examples

See the complete example in `examples/project_workflow_example.go` for a comprehensive demonstration of all features including:

- Project creation and management
- Task creation with dependencies
- Sprint and milestone management
- Team member management
- Time tracking and reporting
- Analytics and insights generation

## 🤝 Contributing

1. Follow established patterns and interfaces
2. Add comprehensive tests for new features
3. Update documentation for API changes
4. Ensure proper error handling and logging
5. Use OpenTelemetry for observability

## 📄 License

This project management system is part of the AIOS project and follows the same licensing terms.
