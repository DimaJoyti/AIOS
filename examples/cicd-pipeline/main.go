package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aios/aios/pkg/workflow"
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

	fmt.Println("🚀 AIOS CI/CD Pipeline Integration Demo")
	fmt.Println("======================================")

	// Run the comprehensive demo
	if err := runCICDPipelineDemo(logger); err != nil {
		log.Fatalf("Demo failed: %v", err)
	}

	fmt.Println("\n✅ CI/CD Pipeline Integration Demo completed successfully!")
}

func runCICDPipelineDemo(logger *logrus.Logger) error {
	ctx := context.Background()

	// Step 1: Create Workflow Engine (required for pipeline engine)
	fmt.Println("\n1. Creating Workflow Engine...")
	workflowConfig := &workflow.WorkflowEngineConfig{
		Workers:            2,
		QueueSize:          100,
		DefaultTimeout:     30 * time.Minute,
		MaxRetries:         3,
		EnableScheduling:   true,
		EnableIntegrations: true,
	}

	workflowEngine, err := workflow.NewDefaultWorkflowEngine(workflowConfig, logger)
	if err != nil {
		return fmt.Errorf("failed to create workflow engine: %w", err)
	}
	fmt.Println("✓ Workflow Engine created successfully")

	// Step 2: Create Pipeline Engine
	fmt.Println("\n2. Creating Pipeline Engine...")
	pipelineConfig := &workflow.PipelineEngineConfig{
		Workers:        3,
		QueueSize:      200,
		DefaultTimeout: 60 * time.Minute,
		MaxRetries:     2,
		ArtifactStore:  "/tmp/artifacts",
	}

	pipelineEngine, err := workflow.NewDefaultPipelineEngine(pipelineConfig, workflowEngine, logger)
	if err != nil {
		return fmt.Errorf("failed to create pipeline engine: %w", err)
	}
	fmt.Println("✓ Pipeline Engine created successfully")

	// Step 3: Create Full-Stack Application Pipeline
	fmt.Println("\n3. Creating Full-Stack Application Pipeline...")
	fullStackPipeline := &workflow.Pipeline{
		Name:        "Full-Stack Application CI/CD",
		Description: "Complete CI/CD pipeline for a full-stack web application",
		Repository: &workflow.Repository{
			URL:      "https://github.com/company/fullstack-app",
			Branch:   "main",
			Provider: "github",
		},
		Stages: []*workflow.Stage{
			{
				Name:        "Build",
				Description: "Build frontend and backend components",
				Jobs: []*workflow.Job{
					{
						Name:        "Frontend Build",
						Description: "Build React frontend application",
						Environment: "node:18",
						Steps: []*workflow.Step{
							{
								Name:    "Checkout Code",
								Type:    workflow.StepTypeAction,
								Action:  "actions/checkout@v3",
								Enabled: true,
							},
							{
								Name:   "Setup Node.js",
								Type:   workflow.StepTypeAction,
								Action: "actions/setup-node@v3",
								With: map[string]interface{}{
									"node-version": "18",
									"cache":        "npm",
								},
								Enabled: true,
							},
							{
								Name:    "Install Dependencies",
								Type:    workflow.StepTypeCommand,
								Command: "cd frontend && npm ci",
								Timeout: 5 * time.Minute,
								Enabled: true,
							},
							{
								Name:    "Run Tests",
								Type:    workflow.StepTypeCommand,
								Command: "cd frontend && npm test -- --coverage --watchAll=false",
								Timeout: 10 * time.Minute,
								Enabled: true,
							},
							{
								Name:    "Build Application",
								Type:    workflow.StepTypeCommand,
								Command: "cd frontend && npm run build",
								Timeout: 10 * time.Minute,
								Enabled: true,
							},
						},
						Enabled: true,
					},
					{
						Name:        "Backend Build",
						Description: "Build Go backend application",
						Environment: "golang:1.21",
						Steps: []*workflow.Step{
							{
								Name:    "Checkout Code",
								Type:    workflow.StepTypeAction,
								Action:  "actions/checkout@v3",
								Enabled: true,
							},
							{
								Name:   "Setup Go",
								Type:   workflow.StepTypeAction,
								Action: "actions/setup-go@v4",
								With: map[string]interface{}{
									"go-version": "1.21",
								},
								Enabled: true,
							},
							{
								Name:    "Download Dependencies",
								Type:    workflow.StepTypeCommand,
								Command: "cd backend && go mod download",
								Timeout: 5 * time.Minute,
								Enabled: true,
							},
							{
								Name:    "Run Tests",
								Type:    workflow.StepTypeCommand,
								Command: "cd backend && go test -v -race -coverprofile=coverage.out ./...",
								Timeout: 15 * time.Minute,
								Enabled: true,
							},
							{
								Name:    "Build Binary",
								Type:    workflow.StepTypeCommand,
								Command: "cd backend && go build -o app ./cmd/server",
								Timeout: 5 * time.Minute,
								Enabled: true,
							},
						},
						Enabled: true,
					},
				},
				Enabled: true,
			},
			{
				Name:        "Test",
				Description: "Run integration and end-to-end tests",
				DependsOn:   []string{"Build"},
				Jobs: []*workflow.Job{
					{
						Name:        "Integration Tests",
						Description: "Run API integration tests",
						Steps: []*workflow.Step{
							{
								Name:    "Start Test Database",
								Type:    workflow.StepTypeCommand,
								Command: "docker run -d --name test-db -p 5432:5432 -e POSTGRES_PASSWORD=test postgres:15",
								Timeout: 2 * time.Minute,
								Enabled: true,
							},
							{
								Name:    "Wait for Database",
								Type:    workflow.StepTypeCommand,
								Command: "sleep 10",
								Enabled: true,
							},
							{
								Name:    "Run Integration Tests",
								Type:    workflow.StepTypeCommand,
								Command: "cd backend && go test -tags=integration ./tests/integration/...",
								Environment: map[string]string{
									"DATABASE_URL": "postgres://postgres:test@localhost:5432/test?sslmode=disable",
								},
								Timeout: 20 * time.Minute,
								Enabled: true,
							},
							{
								Name:            "Cleanup",
								Type:            workflow.StepTypeCommand,
								Command:         "docker stop test-db && docker rm test-db",
								ContinueOnError: true,
								Enabled:         true,
							},
						},
						Enabled: true,
					},
					{
						Name:        "E2E Tests",
						Description: "Run end-to-end tests with Playwright",
						Steps: []*workflow.Step{
							{
								Name: "Start Application",
								Type: workflow.StepTypeScript,
								Script: `
									cd backend && ./app &
									APP_PID=$!
									cd ../frontend && npm start &
									FRONTEND_PID=$!
									echo "APP_PID=$APP_PID" >> $GITHUB_ENV
									echo "FRONTEND_PID=$FRONTEND_PID" >> $GITHUB_ENV
									sleep 30
								`,
								Timeout: 5 * time.Minute,
								Enabled: true,
							},
							{
								Name:    "Run E2E Tests",
								Type:    workflow.StepTypeCommand,
								Command: "cd e2e && npx playwright test",
								Timeout: 15 * time.Minute,
								Enabled: true,
							},
							{
								Name: "Stop Application",
								Type: workflow.StepTypeScript,
								Script: `
									kill $APP_PID || true
									kill $FRONTEND_PID || true
								`,
								ContinueOnError: true,
								Enabled:         true,
							},
						},
						Enabled: true,
					},
				},
				Enabled: true,
			},
			{
				Name:        "Security",
				Description: "Security scanning and vulnerability assessment",
				DependsOn:   []string{"Build"},
				Jobs: []*workflow.Job{
					{
						Name:        "Security Scan",
						Description: "Run security vulnerability scans",
						Steps: []*workflow.Step{
							{
								Name:            "Frontend Security Scan",
								Type:            workflow.StepTypeCommand,
								Command:         "cd frontend && npm audit --audit-level=high",
								ContinueOnError: true,
								Enabled:         true,
							},
							{
								Name:            "Backend Security Scan",
								Type:            workflow.StepTypeCommand,
								Command:         "cd backend && go list -json -deps ./... | nancy sleuth",
								ContinueOnError: true,
								Enabled:         true,
							},
							{
								Name:            "Container Security Scan",
								Type:            workflow.StepTypeCommand,
								Command:         "trivy image --severity HIGH,CRITICAL myapp:latest",
								ContinueOnError: true,
								Enabled:         true,
							},
						},
						Enabled: true,
					},
				},
				Enabled: true,
			},
			{
				Name:        "Deploy",
				Description: "Deploy to staging and production environments",
				DependsOn:   []string{"Test", "Security"},
				Jobs: []*workflow.Job{
					{
						Name:        "Deploy to Staging",
						Description: "Deploy application to staging environment",
						Steps: []*workflow.Step{
							{
								Name:    "Build Docker Image",
								Type:    workflow.StepTypeCommand,
								Command: "docker build -t myapp:staging .",
								Timeout: 10 * time.Minute,
								Enabled: true,
							},
							{
								Name:    "Push to Registry",
								Type:    workflow.StepTypeCommand,
								Command: "docker push myapp:staging",
								Timeout: 5 * time.Minute,
								Enabled: true,
							},
							{
								Name:    "Deploy to Kubernetes",
								Type:    workflow.StepTypeCommand,
								Command: "kubectl apply -f k8s/staging/ --namespace=staging",
								Timeout: 5 * time.Minute,
								Enabled: true,
							},
							{
								Name:    "Wait for Deployment",
								Type:    workflow.StepTypeCommand,
								Command: "kubectl rollout status deployment/myapp --namespace=staging --timeout=300s",
								Timeout: 6 * time.Minute,
								Enabled: true,
							},
						},
						Conditions: []*workflow.Condition{
							{
								Field:    "branch",
								Operator: workflow.OperatorIn,
								Value:    []interface{}{"main", "develop"},
								Type:     workflow.ValueTypeArray,
							},
						},
						Enabled: true,
					},
					{
						Name:        "Deploy to Production",
						Description: "Deploy application to production environment",
						Steps: []*workflow.Step{
							{
								Name:    "Build Production Image",
								Type:    workflow.StepTypeCommand,
								Command: "docker build -t myapp:production .",
								Timeout: 10 * time.Minute,
								Enabled: true,
							},
							{
								Name:    "Push to Registry",
								Type:    workflow.StepTypeCommand,
								Command: "docker push myapp:production",
								Timeout: 5 * time.Minute,
								Enabled: true,
							},
							{
								Name:    "Deploy to Production",
								Type:    workflow.StepTypeCommand,
								Command: "kubectl apply -f k8s/production/ --namespace=production",
								Timeout: 5 * time.Minute,
								Enabled: true,
							},
							{
								Name:    "Wait for Deployment",
								Type:    workflow.StepTypeCommand,
								Command: "kubectl rollout status deployment/myapp --namespace=production --timeout=600s",
								Timeout: 11 * time.Minute,
								Enabled: true,
							},
							{
								Name:    "Run Smoke Tests",
								Type:    workflow.StepTypeCommand,
								Command: "curl -f https://myapp.com/health || exit 1",
								Timeout: 2 * time.Minute,
								Enabled: true,
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
				Enabled: true,
			},
		},
		Triggers: []*workflow.PipelineTrigger{
			{
				Type:     workflow.TriggerTypePush,
				Branches: []string{"main", "develop"},
				Enabled:  true,
			},
			{
				Type:    workflow.TriggerTypePR,
				Events:  []string{"opened", "synchronize"},
				Enabled: true,
			},
		},
		Variables: map[string]interface{}{
			"NODE_VERSION":    "18",
			"GO_VERSION":      "1.21",
			"DOCKER_REGISTRY": "registry.company.com",
			"ENVIRONMENT":     "staging",
		},
		Timeout:   120 * time.Minute,
		CreatedBy: "devops-team@company.com",
		Tags:      []string{"fullstack", "ci", "cd", "production"},
	}

	createdPipeline, err := pipelineEngine.CreatePipeline(ctx, fullStackPipeline)
	if err != nil {
		return fmt.Errorf("failed to create pipeline: %w", err)
	}
	fmt.Printf("✓ Full-Stack Pipeline created: %s (ID: %s)\n", createdPipeline.Name, createdPipeline.ID)

	// Step 4: Create Microservice Pipeline
	fmt.Println("\n4. Creating Microservice Pipeline...")
	microservicePipeline := &workflow.Pipeline{
		Name:        "Microservice CI/CD",
		Description: "Lightweight CI/CD pipeline for microservices",
		Repository: &workflow.Repository{
			URL:      "https://github.com/company/user-service",
			Branch:   "main",
			Provider: "github",
		},
		Stages: []*workflow.Stage{
			{
				Name: "Build & Test",
				Jobs: []*workflow.Job{
					{
						Name: "Go Service",
						Steps: []*workflow.Step{
							{
								Name:    "Test",
								Type:    workflow.StepTypeCommand,
								Command: "go test -v ./...",
								Enabled: true,
							},
							{
								Name:    "Build",
								Type:    workflow.StepTypeCommand,
								Command: "go build -o service ./cmd/server",
								Enabled: true,
							},
						},
						Enabled: true,
					},
				},
				Enabled: true,
			},
			{
				Name:      "Deploy",
				DependsOn: []string{"Build & Test"},
				Jobs: []*workflow.Job{
					{
						Name: "Container Deploy",
						Steps: []*workflow.Step{
							{
								Name:    "Build Image",
								Type:    workflow.StepTypeCommand,
								Command: "docker build -t user-service:latest .",
								Enabled: true,
							},
							{
								Name:    "Deploy",
								Type:    workflow.StepTypeCommand,
								Command: "kubectl apply -f k8s/",
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
				Branches: []string{"main"},
				Enabled:  true,
			},
		},
		Timeout:   30 * time.Minute,
		CreatedBy: "microservices-team@company.com",
		Tags:      []string{"microservice", "go", "kubernetes"},
	}

	createdMicroservice, err := pipelineEngine.CreatePipeline(ctx, microservicePipeline)
	if err != nil {
		return fmt.Errorf("failed to create microservice pipeline: %w", err)
	}
	fmt.Printf("✓ Microservice Pipeline created: %s (ID: %s)\n", createdMicroservice.Name, createdMicroservice.ID)

	// Step 5: Execute Pipeline Manually
	fmt.Println("\n5. Executing Pipeline Manually...")

	executionParams := &workflow.ExecutionParams{
		Branch:      "main",
		Commit:      "abc123def456",
		TriggerType: "manual",
		Variables: map[string]interface{}{
			"ENVIRONMENT": "staging",
			"VERSION":     "v1.2.3",
			"DEPLOY_ENV":  "staging",
		},
	}

	execution, err := pipelineEngine.ExecutePipeline(ctx, createdMicroservice.ID, executionParams)
	if err != nil {
		fmt.Printf("   ⚠️  Failed to execute pipeline: %v\n", err)
	} else {
		fmt.Printf("   ✓ Pipeline execution started: %s\n", execution.ID)
	}

	// Step 6: Simulate Event-Driven Pipeline Trigger
	fmt.Println("\n6. Simulating Event-Driven Pipeline Triggers...")

	// Simulate push event
	pushEvent := &workflow.Event{
		Type:    "push",
		Source:  "github",
		Subject: "repository/main",
		Data: map[string]interface{}{
			"branch":     "main",
			"commit":     "def456ghi789",
			"author":     "developer@company.com",
			"message":    "Add new feature for user authentication",
			"repository": "company/fullstack-app",
			"paths":      []string{"backend/auth/", "frontend/src/auth/"},
		},
		Timestamp: time.Now(),
		UserID:    "developer@company.com",
	}

	pushExecutions, err := pipelineEngine.TriggerPipeline(ctx, pushEvent)
	if err != nil {
		fmt.Printf("   ⚠️  Failed to trigger pipelines for push event: %v\n", err)
	} else {
		fmt.Printf("   ✓ Push event triggered %d pipeline executions\n", len(pushExecutions))
	}

	// Simulate pull request event
	prEvent := &workflow.Event{
		Type:    "pull_request",
		Source:  "github",
		Subject: "repository/pr/42",
		Data: map[string]interface{}{
			"action":     "opened",
			"branch":     "feature/new-dashboard",
			"base":       "main",
			"commit":     "ghi789jkl012",
			"author":     "frontend-dev@company.com",
			"title":      "Add new dashboard component",
			"repository": "company/fullstack-app",
		},
		Timestamp: time.Now(),
		UserID:    "frontend-dev@company.com",
	}

	prExecutions, err := pipelineEngine.TriggerPipeline(ctx, prEvent)
	if err != nil {
		fmt.Printf("   ⚠️  Failed to trigger pipelines for PR event: %v\n", err)
	} else {
		fmt.Printf("   ✓ PR event triggered %d pipeline executions\n", len(prExecutions))
	}

	// Step 7: Monitor Pipeline Executions
	fmt.Println("\n7. Monitoring Pipeline Executions...")
	time.Sleep(3 * time.Second) // Give some time for processing

	allExecutions := append(pushExecutions, prExecutions...)
	if execution != nil {
		allExecutions = append(allExecutions, execution)
	}

	for i, exec := range allExecutions {
		updatedExecution, err := pipelineEngine.GetPipelineExecution(ctx, exec.ID)
		if err != nil {
			fmt.Printf("   ⚠️  Failed to get execution %d status: %v\n", i+1, err)
			continue
		}

		fmt.Printf("   ✓ Execution %d (%s): %s\n", i+1, updatedExecution.ID[:8], updatedExecution.Status)
		fmt.Printf("     Branch: %s, Trigger: %s\n", updatedExecution.Branch, updatedExecution.TriggerType)
		if updatedExecution.Error != "" {
			fmt.Printf("     Error: %s\n", updatedExecution.Error)
		}
		if len(updatedExecution.Stages) > 0 {
			fmt.Printf("     Stages completed: %d\n", len(updatedExecution.Stages))
			for _, stage := range updatedExecution.Stages {
				fmt.Printf("       - %s: %s (%d jobs)\n", stage.Name, stage.Status, len(stage.Jobs))
			}
		}
	}

	// Step 8: Artifact Management Demo
	fmt.Println("\n8. Demonstrating Artifact Management...")

	if len(allExecutions) > 0 {
		testExecution := allExecutions[0]

		// Upload test artifact
		artifact := &workflow.Artifact{
			Name: "test-results.xml",
			Type: "test-report",
			Path: "/artifacts/test-results.xml",
		}

		artifactContent := `<?xml version="1.0" encoding="UTF-8"?>
<testsuites tests="25" failures="0" errors="0" time="12.345">
  <testsuite name="UserService" tests="15" failures="0" errors="0" time="8.123">
    <testcase name="TestCreateUser" time="0.123"/>
    <testcase name="TestUpdateUser" time="0.234"/>
  </testsuite>
  <testsuite name="AuthService" tests="10" failures="0" errors="0" time="4.222">
    <testcase name="TestLogin" time="0.456"/>
    <testcase name="TestLogout" time="0.123"/>
  </testsuite>
</testsuites>`

		err = pipelineEngine.UploadArtifact(ctx, testExecution.ID, artifact, strings.NewReader(artifactContent))
		if err != nil {
			fmt.Printf("   ⚠️  Failed to upload artifact: %v\n", err)
		} else {
			fmt.Printf("   ✓ Artifact uploaded: %s (ID: %s)\n", artifact.Name, artifact.ID)
		}

		// List artifacts
		artifacts, err := pipelineEngine.ListArtifacts(ctx, testExecution.ID)
		if err != nil {
			fmt.Printf("   ⚠️  Failed to list artifacts: %v\n", err)
		} else {
			fmt.Printf("   ✓ Found %d artifacts for execution:\n", len(artifacts))
			for _, art := range artifacts {
				fmt.Printf("     - %s (%s) - %d bytes\n", art.Name, art.Type, art.Size)
			}
		}

		// Download artifact
		if len(artifacts) > 0 {
			var downloadBuffer bytes.Buffer
			err = pipelineEngine.DownloadArtifact(ctx, artifacts[0].ID, &downloadBuffer)
			if err != nil {
				fmt.Printf("   ⚠️  Failed to download artifact: %v\n", err)
			} else {
				fmt.Printf("   ✓ Artifact downloaded: %d bytes\n", downloadBuffer.Len())
			}
		}
	}

	// Step 9: Pipeline Analytics
	fmt.Println("\n9. Pipeline Analytics and Reporting...")

	// List all pipelines
	pipelines, err := pipelineEngine.ListPipelines(ctx, &workflow.PipelineFilter{
		Limit: 10,
	})
	if err != nil {
		fmt.Printf("   ⚠️  Failed to list pipelines: %v\n", err)
	} else {
		fmt.Printf("   ✓ Found %d pipelines:\n", len(pipelines))
		for _, pipeline := range pipelines {
			fmt.Printf("     - %s (%s) - %s [%d stages]\n",
				pipeline.Name, pipeline.Status, pipeline.CreatedBy, len(pipeline.Stages))
		}
	}

	// List executions with filtering
	executions, err := pipelineEngine.ListPipelineExecutions(ctx, &workflow.PipelineExecutionFilter{
		Status: []workflow.ExecutionStatus{
			workflow.ExecutionStatusSuccess,
			workflow.ExecutionStatusFailure,
			workflow.ExecutionStatusRunning,
		},
		Limit: 20,
	})
	if err != nil {
		fmt.Printf("   ⚠️  Failed to list executions: %v\n", err)
	} else {
		fmt.Printf("   ✓ Found %d recent executions:\n", len(executions))

		var successful, failed, running int
		var totalDuration time.Duration

		for _, exec := range executions {
			switch exec.Status {
			case workflow.ExecutionStatusSuccess:
				successful++
				totalDuration += exec.Duration
			case workflow.ExecutionStatusFailure:
				failed++
				totalDuration += exec.Duration
			case workflow.ExecutionStatusRunning:
				running++
			}
		}

		if len(executions) > 0 {
			successRate := float64(successful) / float64(len(executions)) * 100
			avgDuration := totalDuration / time.Duration(successful+failed)

			fmt.Printf("     Success Rate: %.1f%% (%d/%d)\n", successRate, successful, len(executions))
			fmt.Printf("     Average Duration: %s\n", avgDuration.Round(time.Second))
			fmt.Printf("     Currently Running: %d\n", running)
		}
	}

	// Step 10: Pipeline Management Operations
	fmt.Println("\n10. Pipeline Management Operations...")

	// Register additional trigger
	newTrigger := &workflow.PipelineTrigger{
		Type:    workflow.TriggerTypeTag,
		Tags:    []string{"v*"},
		Events:  []string{"tag.created"},
		Enabled: true,
	}

	err = pipelineEngine.RegisterPipelineTrigger(ctx, createdPipeline.ID, newTrigger)
	if err != nil {
		fmt.Printf("   ⚠️  Failed to register trigger: %v\n", err)
	} else {
		fmt.Printf("   ✓ Tag trigger registered for pipeline: %s\n", createdPipeline.Name)
	}

	// Update pipeline status
	createdPipeline.Status = workflow.PipelineStatusActive
	createdPipeline.UpdatedBy = "admin@company.com"

	_, err = pipelineEngine.UpdatePipeline(ctx, createdPipeline)
	if err != nil {
		fmt.Printf("   ⚠️  Failed to update pipeline: %v\n", err)
	} else {
		fmt.Printf("   ✓ Pipeline status updated to: %s\n", createdPipeline.Status)
	}

	return nil
}
