# AIOS CI/CD Pipeline Integration

## Overview

The AIOS CI/CD Pipeline Integration provides a comprehensive, enterprise-grade continuous integration and deployment platform. Built on top of the workflow automation engine, it offers advanced pipeline orchestration, multi-stage builds, artifact management, and seamless integration with popular CI/CD tools and platforms.

## 🏗️ Architecture

### Core Components

```
CI/CD Pipeline Integration
├── Pipeline Engine (pipeline orchestration, execution management)
├── Stage Orchestrator (dependency management, parallel execution)
├── Job Executor (containerized builds, environment management)
├── Step Processor (command execution, script running, action handling)
├── Artifact Manager (build outputs, test results, deployment packages)
├── Trigger System (Git events, webhooks, scheduled builds)
├── Integration Layer (GitHub Actions, GitLab CI, Jenkins)
└── Analytics Engine (build metrics, performance insights)
```

### Key Features

- **🔄 Multi-Stage Pipelines**: Complex build workflows with stage dependencies
- **⚡ Parallel Execution**: Concurrent job execution for faster builds
- **📦 Artifact Management**: Build outputs, test results, and deployment packages
- **🎯 Event-Driven Triggers**: Git push, pull requests, tags, and custom events
- **🔗 Platform Integration**: GitHub Actions, GitLab CI, Jenkins compatibility
- **📊 Build Analytics**: Performance metrics and success rate tracking
- **🛡️ Security Scanning**: Integrated vulnerability assessment
- **🚀 Multi-Environment Deployment**: Staging, production, and custom environments

## 🚀 Quick Start

### Basic Pipeline Creation

```go
package main

import (
    "context"
    "github.com/aios/aios/pkg/workflow"
    "github.com/sirupsen/logrus"
)

func main() {
    logger := logrus.New()
    
    // Create workflow engine (required for pipeline engine)
    workflowEngine, err := workflow.NewDefaultWorkflowEngine(nil, logger)
    if err != nil {
        panic(err)
    }
    
    // Create pipeline engine
    pipelineEngine, err := workflow.NewDefaultPipelineEngine(nil, workflowEngine, logger)
    if err != nil {
        panic(err)
    }
    
    // Create a simple CI pipeline
    pipeline := &workflow.Pipeline{
        Name:        "Node.js CI Pipeline",
        Description: "Build and test Node.js application",
        Repository: &workflow.Repository{
            URL:      "https://github.com/company/my-app",
            Branch:   "main",
            Provider: "github",
        },
        Stages: []*workflow.Stage{
            {
                Name: "Build & Test",
                Jobs: []*workflow.Job{
                    {
                        Name: "Node.js Build",
                        Steps: []*workflow.Step{
                            {
                                Name:    "Install Dependencies",
                                Type:    workflow.StepTypeCommand,
                                Command: "npm ci",
                                Enabled: true,
                            },
                            {
                                Name:    "Run Tests",
                                Type:    workflow.StepTypeCommand,
                                Command: "npm test",
                                Enabled: true,
                            },
                            {
                                Name:    "Build Application",
                                Type:    workflow.StepTypeCommand,
                                Command: "npm run build",
                                Enabled: true,
                            },
                        },
                        Enabled: true,
                    },
                },
                Enabled: true,
            },
        },
        Triggers: []*workflow.PipelineTrigger{
            {
                Type:     workflow.TriggerTypePush,
                Branches: []string{"main", "develop"},
                Enabled:  true,
            },
        },
    }
    
    // Create pipeline
    createdPipeline, err := pipelineEngine.CreatePipeline(context.Background(), pipeline)
    if err != nil {
        panic(err)
    }
    
    // Execute pipeline
    execution, err := pipelineEngine.ExecutePipeline(context.Background(), createdPipeline.ID, &workflow.ExecutionParams{
        Branch:      "main",
        TriggerType: "manual",
    })
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("Pipeline execution started: %s\n", execution.ID)
}
```

## 🔄 Pipeline Components

### Pipelines

A pipeline defines the complete CI/CD workflow:

```go
pipeline := &workflow.Pipeline{
    Name:        "Full-Stack Application CI/CD",
    Description: "Complete CI/CD pipeline for full-stack application",
    Repository: &workflow.Repository{
        URL:      "https://github.com/company/fullstack-app",
        Branch:   "main",
        Provider: "github",
        Token:    "ghp_xxxxxxxxxxxx", // Optional for private repos
    },
    
    // Multi-stage pipeline
    Stages: []*workflow.Stage{
        {
            Name:        "Build",
            Description: "Build frontend and backend components",
            Jobs: []*workflow.Job{
                {
                    Name:        "Frontend Build",
                    Environment: "node:18",
                    Steps: []*workflow.Step{
                        {
                            Name:    "Setup Node.js",
                            Type:    workflow.StepTypeAction,
                            Action:  "actions/setup-node@v3",
                            With: map[string]interface{}{
                                "node-version": "18",
                                "cache":        "npm",
                            },
                        },
                        {
                            Name:    "Install Dependencies",
                            Type:    workflow.StepTypeCommand,
                            Command: "npm ci",
                            Timeout: 5 * time.Minute,
                        },
                        {
                            Name:    "Build Application",
                            Type:    workflow.StepTypeCommand,
                            Command: "npm run build",
                            Timeout: 10 * time.Minute,
                        },
                    ],
                    Enabled: true,
                },
                {
                    Name:        "Backend Build",
                    Environment: "golang:1.21",
                    Steps: []*workflow.Step{
                        {
                            Name:    "Build Binary",
                            Type:    workflow.StepTypeCommand,
                            Command: "go build -o app ./cmd/server",
                            Timeout: 5 * time.Minute,
                        },
                    ],
                    Enabled: true,
                },
            },
            Enabled: true,
        },
        {
            Name:      "Test",
            DependsOn: []string{"Build"}, // Stage dependency
            Jobs: []*workflow.Job{
                {
                    Name: "Integration Tests",
                    Steps: []*workflow.Step{
                        {
                            Name:    "Run Integration Tests",
                            Type:    workflow.StepTypeScript,
                            Script: `
                                docker run -d --name test-db postgres:15
                                sleep 10
                                go test -tags=integration ./...
                                docker stop test-db && docker rm test-db
                            `,
                            Timeout: 20 * time.Minute,
                        },
                    },
                    Enabled: true,
                },
            },
            Enabled: true,
        },
        {
            Name:      "Deploy",
            DependsOn: []string{"Test"},
            Jobs: []*workflow.Job{
                {
                    Name: "Deploy to Staging",
                    Steps: []*workflow.Step{
                        {
                            Name:    "Build Docker Image",
                            Type:    workflow.StepTypeCommand,
                            Command: "docker build -t myapp:staging .",
                        },
                        {
                            Name:    "Deploy to Kubernetes",
                            Type:    workflow.StepTypeCommand,
                            Command: "kubectl apply -f k8s/staging/",
                        },
                    ],
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
            Enabled: true,
        },
    },
    
    // Pipeline triggers
    Triggers: []*workflow.PipelineTrigger{
        {
            Type:     workflow.TriggerTypePush,
            Branches: []string{"main", "develop"},
            Enabled:  true,
        },
        {
            Type:   workflow.TriggerTypePR,
            Events: []string{"opened", "synchronize"},
            Enabled: true,
        },
        {
            Type:    workflow.TriggerTypeTag,
            Tags:    []string{"v*"},
            Enabled: true,
        },
    ],
    
    // Global variables
    Variables: map[string]interface{}{
        "NODE_VERSION":    "18",
        "GO_VERSION":      "1.21",
        "DOCKER_REGISTRY": "registry.company.com",
    },
    
    Timeout:   120 * time.Minute,
    CreatedBy: "devops-team@company.com",
    Tags:      []string{"fullstack", "production"},
}
```

### Stages

Stages organize jobs into logical groups with dependencies:

```go
stages := []*workflow.Stage{
    {
        Name:        "Build",
        Description: "Compile and package application",
        Jobs: []*workflow.Job{
            // Build jobs
        },
        Enabled: true,
    },
    {
        Name:        "Test",
        Description: "Run all test suites",
        DependsOn:   []string{"Build"}, // Wait for Build stage
        Jobs: []*workflow.Job{
            // Test jobs
        },
        Conditions: []*workflow.Condition{
            {
                Field:    "skip_tests",
                Operator: workflow.OperatorNotEquals,
                Value:    "true",
            },
        },
        Enabled: true,
    },
    {
        Name:        "Security",
        Description: "Security scanning and vulnerability assessment",
        DependsOn:   []string{"Build"}, // Parallel with Test stage
        Jobs: []*workflow.Job{
            // Security jobs
        },
        Enabled: true,
    },
    {
        Name:        "Deploy",
        Description: "Deploy to target environments",
        DependsOn:   []string{"Test", "Security"}, // Wait for both stages
        Jobs: []*workflow.Job{
            // Deployment jobs
        },
        Enabled: true,
    },
}
```

### Jobs

Jobs define the execution environment and steps:

```go
jobs := []*workflow.Job{
    {
        Name:        "Frontend Tests",
        Description: "Run frontend unit and integration tests",
        Environment: "node:18-alpine",
        Steps: []*workflow.Step{
            {
                Name:    "Install Dependencies",
                Type:    workflow.StepTypeCommand,
                Command: "npm ci --prefer-offline",
                Timeout: 5 * time.Minute,
            },
            {
                Name:    "Run Unit Tests",
                Type:    workflow.StepTypeCommand,
                Command: "npm run test:unit -- --coverage",
                Timeout: 10 * time.Minute,
            },
            {
                Name:    "Run E2E Tests",
                Type:    workflow.StepTypeCommand,
                Command: "npm run test:e2e",
                Timeout: 15 * time.Minute,
                ContinueOnError: true, // Don't fail pipeline if E2E tests fail
            },
        },
        Variables: map[string]interface{}{
            "NODE_ENV": "test",
            "CI":       "true",
        },
        Timeout: 30 * time.Minute,
        RetryPolicy: &workflow.RetryPolicy{
            MaxRetries:    2,
            RetryDelay:    1 * time.Minute,
            BackoffFactor: 2.0,
        },
        Enabled: true,
    },
    {
        Name:        "Backend Tests",
        Description: "Run backend unit and integration tests",
        Environment: "golang:1.21",
        Steps: []*workflow.Step{
            {
                Name:    "Download Dependencies",
                Type:    workflow.StepTypeCommand,
                Command: "go mod download",
            },
            {
                Name:    "Run Tests",
                Type:    workflow.StepTypeCommand,
                Command: "go test -v -race -coverprofile=coverage.out ./...",
                Environment: map[string]string{
                    "CGO_ENABLED": "1",
                    "GOOS":        "linux",
                },
            },
            {
                Name:    "Upload Coverage",
                Type:    workflow.StepTypeAction,
                Action:  "codecov/codecov-action@v3",
                With: map[string]interface{}{
                    "file": "coverage.out",
                },
            },
        },
        Enabled: true,
    },
}
```

### Steps

Steps define individual actions within jobs:

**Command Steps:**
```go
commandStep := &workflow.Step{
    Name:    "Build Application",
    Type:    workflow.StepTypeCommand,
    Command: "go build -ldflags='-s -w' -o app ./cmd/server",
    Environment: map[string]string{
        "CGO_ENABLED": "0",
        "GOOS":        "linux",
        "GOARCH":      "amd64",
    },
    WorkingDir: "./backend",
    Timeout:    5 * time.Minute,
    Enabled:    true,
}
```

**Script Steps:**
```go
scriptStep := &workflow.Step{
    Name: "Setup Test Environment",
    Type: workflow.StepTypeScript,
    Script: `
        #!/bin/bash
        set -e
        
        echo "Starting test database..."
        docker run -d --name test-db \
            -e POSTGRES_PASSWORD=test \
            -p 5432:5432 \
            postgres:15
        
        echo "Waiting for database to be ready..."
        until docker exec test-db pg_isready; do
            sleep 1
        done
        
        echo "Database is ready!"
    `,
    Timeout: 2 * time.Minute,
    Enabled: true,
}
```

**Action Steps:**
```go
actionStep := &workflow.Step{
    Name:   "Deploy to AWS",
    Type:   workflow.StepTypeAction,
    Action: "aws-actions/configure-aws-credentials@v2",
    With: map[string]interface{}{
        "aws-access-key-id":     "${{ secrets.AWS_ACCESS_KEY_ID }}",
        "aws-secret-access-key": "${{ secrets.AWS_SECRET_ACCESS_KEY }}",
        "aws-region":            "us-west-2",
    },
    Enabled: true,
}
```

## 🎯 Advanced Features

### Event-Driven Triggers

Trigger pipelines automatically based on Git events:

```go
// Push triggers
pushTrigger := &workflow.PipelineTrigger{
    Type:     workflow.TriggerTypePush,
    Branches: []string{"main", "develop", "release/*"},
    Paths:    []string{"src/", "package.json"}, // Only trigger on specific paths
    Enabled:  true,
}

// Pull request triggers
prTrigger := &workflow.PipelineTrigger{
    Type:   workflow.TriggerTypePR,
    Events: []string{"opened", "synchronize", "reopened"},
    Branches: []string{"main"}, // Target branches
    Enabled: true,
}

// Tag triggers
tagTrigger := &workflow.PipelineTrigger{
    Type:    workflow.TriggerTypeTag,
    Tags:    []string{"v*", "release-*"},
    Events:  []string{"tag.created"},
    Enabled: true,
}

// Register triggers
err := pipelineEngine.RegisterPipelineTrigger(ctx, pipelineID, pushTrigger)
err = pipelineEngine.RegisterPipelineTrigger(ctx, pipelineID, prTrigger)
err = pipelineEngine.RegisterPipelineTrigger(ctx, pipelineID, tagTrigger)
```

### Artifact Management

Handle build outputs, test results, and deployment packages:

```go
// Upload build artifact
artifact := &workflow.Artifact{
    Name: "application-v1.2.3.tar.gz",
    Type: "application",
    Path: "/artifacts/builds/application-v1.2.3.tar.gz",
}

err := pipelineEngine.UploadArtifact(ctx, executionID, artifact, artifactReader)

// Upload test results
testArtifact := &workflow.Artifact{
    Name: "test-results.xml",
    Type: "test-report",
}

err = pipelineEngine.UploadArtifact(ctx, executionID, testArtifact, testResultsReader)

// List execution artifacts
artifacts, err := pipelineEngine.ListArtifacts(ctx, executionID)

// Download artifact
var downloadBuffer bytes.Buffer
err = pipelineEngine.DownloadArtifact(ctx, artifactID, &downloadBuffer)
```

### Pipeline Execution

Execute pipelines manually or via events:

```go
// Manual execution
params := &workflow.ExecutionParams{
    Branch:      "feature/new-feature",
    Commit:      "abc123def456",
    TriggerType: "manual",
    Variables: map[string]interface{}{
        "ENVIRONMENT": "staging",
        "VERSION":     "v1.2.3-beta",
        "DEPLOY_ENV":  "staging",
    },
}

execution, err := pipelineEngine.ExecutePipeline(ctx, pipelineID, params)

// Event-driven execution
event := &workflow.Event{
    Type:      "push",
    Source:    "github",
    Subject:   "repository/main",
    Data: map[string]interface{}{
        "branch":     "main",
        "commit":     "def456ghi789",
        "author":     "developer@company.com",
        "repository": "company/my-app",
    },
    Timestamp: time.Now(),
}

executions, err := pipelineEngine.TriggerPipeline(ctx, event)
```

### Pipeline Monitoring

Monitor execution status and performance:

```go
// Get execution details
execution, err := pipelineEngine.GetPipelineExecution(ctx, executionID)

fmt.Printf("Status: %s\n", execution.Status)
fmt.Printf("Duration: %s\n", execution.Duration)
fmt.Printf("Branch: %s\n", execution.Branch)

// Monitor stages and jobs
for _, stage := range execution.Stages {
    fmt.Printf("Stage: %s (%s)\n", stage.Name, stage.Status)
    for _, job := range stage.Jobs {
        fmt.Printf("  Job: %s (%s)\n", job.Name, job.Status)
        for _, step := range job.Steps {
            fmt.Printf("    Step: %s (%s) - %s\n", step.Name, step.Status, step.Duration)
        }
    }
}

// List executions with filtering
executions, err := pipelineEngine.ListPipelineExecutions(ctx, &workflow.PipelineExecutionFilter{
    PipelineID: pipelineID,
    Status:     []workflow.ExecutionStatus{workflow.ExecutionStatusRunning},
    Branch:     "main",
    Limit:      10,
})

// Cancel running execution
err = pipelineEngine.CancelPipelineExecution(ctx, executionID)

// Retry failed execution
newExecution, err := pipelineEngine.RetryPipelineExecution(ctx, executionID)
```

## 📊 Analytics and Reporting

### Pipeline Metrics

```go
// List pipelines with analytics
pipelines, err := pipelineEngine.ListPipelines(ctx, &workflow.PipelineFilter{
    Status: []workflow.PipelineStatus{workflow.PipelineStatusActive},
    Tags:   []string{"production"},
    Limit:  20,
})

// Calculate success rates
for _, pipeline := range pipelines {
    executions, _ := pipelineEngine.ListPipelineExecutions(ctx, &workflow.PipelineExecutionFilter{
        PipelineID: pipeline.ID,
        Limit:      100,
    })
    
    var successful, failed int
    var totalDuration time.Duration
    
    for _, exec := range executions {
        switch exec.Status {
        case workflow.ExecutionStatusSuccess:
            successful++
            totalDuration += exec.Duration
        case workflow.ExecutionStatusFailure:
            failed++
        }
    }
    
    if len(executions) > 0 {
        successRate := float64(successful) / float64(len(executions)) * 100
        avgDuration := totalDuration / time.Duration(successful)
        
        fmt.Printf("Pipeline: %s\n", pipeline.Name)
        fmt.Printf("  Success Rate: %.1f%%\n", successRate)
        fmt.Printf("  Average Duration: %s\n", avgDuration)
        fmt.Printf("  Total Executions: %d\n", len(executions))
    }
}
```

## 🧪 Testing

Run the comprehensive test suite:

```bash
# Run all pipeline tests
go test ./pkg/workflow/...

# Run with race detection
go test -race ./pkg/workflow/...

# Run integration tests
go test -tags=integration ./pkg/workflow/...

# Run CI/CD pipeline example
go run examples/cicd_pipeline_example.go
```

## 📖 Examples

See the complete example in `examples/cicd_pipeline_example.go` for a comprehensive demonstration including:

- Full-stack application pipeline with multiple stages
- Microservice pipeline for lightweight deployments
- Event-driven pipeline triggers
- Artifact management and storage
- Pipeline monitoring and analytics
- Multi-environment deployment strategies

## 🤝 Contributing

1. Follow established patterns and interfaces
2. Add comprehensive tests for new features
3. Update documentation for API changes
4. Ensure proper error handling and logging
5. Use OpenTelemetry for observability

## 📄 License

This CI/CD pipeline integration is part of the AIOS project and follows the same licensing terms.
