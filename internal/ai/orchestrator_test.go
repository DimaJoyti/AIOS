package ai

import (
	"context"
	"testing"
	"time"

	"github.com/aios/aios/pkg/models"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewOrchestrator(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	config := AIServiceConfig{
		OllamaHost:    "localhost",
		OllamaPort:    11434,
		OllamaTimeout: 30 * time.Second,
		DefaultModel:  "llama2",
		MaxTokens:     2048,
		Temperature:   0.7,
	}

	orchestrator := NewOrchestrator(config, logger)
	assert.NotNil(t, orchestrator)
	assert.NotNil(t, orchestrator.logger)
	assert.NotNil(t, orchestrator.tracer)
	assert.Equal(t, config, orchestrator.config)
}

func TestOrchestratorInitialize(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	config := AIServiceConfig{
		OllamaHost:    "localhost",
		OllamaPort:    11434,
		OllamaTimeout: 30 * time.Second,
		DefaultModel:  "llama2",
		MaxTokens:     2048,
		Temperature:   0.7,
	}

	orchestrator := NewOrchestrator(config, logger)
	ctx := context.Background()

	err := orchestrator.Initialize(ctx)
	require.NoError(t, err)

	// Verify services are initialized
	assert.NotNil(t, orchestrator.llmService)
	assert.NotNil(t, orchestrator.cvService)
	assert.NotNil(t, orchestrator.optimizationService)
}

func TestOrchestratorRouteRequest(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	config := AIServiceConfig{}
	orchestrator := NewOrchestrator(config, logger)
	ctx := context.Background()

	tests := []struct {
		name           string
		requestType    string
		expectedService string
		expectError    bool
	}{
		{
			name:           "Chat request",
			requestType:    "chat",
			expectedService: "llm",
			expectError:    false,
		},
		{
			name:           "Vision request",
			requestType:    "vision",
			expectedService: "cv",
			expectError:    false,
		},
		{
			name:           "Optimization request",
			requestType:    "optimization",
			expectedService: "optimization",
			expectError:    false,
		},
		{
			name:           "Unknown request",
			requestType:    "unknown",
			expectedService: "",
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := &models.AIRequest{
				ID:   uuid.New().String(),
				Type: tt.requestType,
			}

			serviceID, err := orchestrator.RouteRequest(ctx, request)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedService, serviceID)
			}
		})
	}
}

func TestOrchestratorGetServiceStatus(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	config := AIServiceConfig{}
	orchestrator := NewOrchestrator(config, logger)
	ctx := context.Background()

	status, err := orchestrator.GetServiceStatus(ctx)
	require.NoError(t, err)
	assert.NotNil(t, status)
	assert.NotEmpty(t, status.Services)
	assert.NotEmpty(t, status.Models)
	assert.Equal(t, "healthy", status.OverallHealth)
}

func TestOrchestratorProcessOptimizationRequest(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	config := AIServiceConfig{}
	orchestrator := NewOrchestrator(config, logger)
	ctx := context.Background()

	// Initialize the orchestrator
	err := orchestrator.Initialize(ctx)
	require.NoError(t, err)

	// Create an optimization request
	request := &models.AIRequest{
		ID:   uuid.New().String(),
		Type: "optimization",
		Parameters: map[string]interface{}{
			"task": "analyze",
		},
		Timeout:   30 * time.Second,
		Timestamp: time.Now(),
	}

	// Process the request
	response, err := orchestrator.ProcessRequest(ctx, request)
	require.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, request.ID, response.RequestID)
	assert.Equal(t, "optimization", response.Type)
	assert.NotNil(t, response.Result)
}

func TestOrchestratorAggregateResults(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	config := AIServiceConfig{}
	orchestrator := NewOrchestrator(config, logger)
	ctx := context.Background()

	// Create mock results
	results := []models.AIResult{
		{
			ServiceID:      "llm",
			Type:           "chat",
			Result:         "Hello, world!",
			Confidence:     0.9,
			ProcessingTime: 100 * time.Millisecond,
			Timestamp:      time.Now(),
		},
		{
			ServiceID:      "llm",
			Type:           "chat",
			Result:         "Hi there!",
			Confidence:     0.8,
			ProcessingTime: 120 * time.Millisecond,
			Timestamp:      time.Now(),
		},
	}

	aggregated, err := orchestrator.AggregateResults(ctx, results)
	require.NoError(t, err)
	assert.NotNil(t, aggregated)
	assert.Equal(t, len(results), len(aggregated.Results))
	assert.Greater(t, aggregated.Confidence, 0.0)
	assert.NotEmpty(t, aggregated.Method)
}

func TestOrchestratorManageWorkflow(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	config := AIServiceConfig{}
	orchestrator := NewOrchestrator(config, logger)
	ctx := context.Background()

	// Initialize the orchestrator
	err := orchestrator.Initialize(ctx)
	require.NoError(t, err)

	// Create a simple workflow
	workflow := &models.AIWorkflow{
		ID:   uuid.New().String(),
		Name: "Test Workflow",
		Steps: []models.WorkflowStep{
			{
				ID:      "step1",
				Type:    "optimization",
				Service: "optimization",
				Parameters: map[string]interface{}{
					"task": "analyze",
				},
				Timeout: 30 * time.Second,
			},
		},
		Input:     nil,
		Timeout:   60 * time.Second,
		Timestamp: time.Now(),
	}

	// Execute the workflow
	result, err := orchestrator.ManageWorkflow(ctx, workflow)
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, workflow.ID, result.WorkflowID)
	assert.Contains(t, []string{"completed", "failed"}, result.Status)
}

func BenchmarkOrchestratorProcessRequest(b *testing.B) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel) // Reduce logging for benchmark

	config := AIServiceConfig{}
	orchestrator := NewOrchestrator(config, logger)
	ctx := context.Background()

	// Initialize the orchestrator
	err := orchestrator.Initialize(ctx)
	require.NoError(b, err)

	request := &models.AIRequest{
		ID:   uuid.New().String(),
		Type: "optimization",
		Parameters: map[string]interface{}{
			"task": "analyze",
		},
		Timeout:   30 * time.Second,
		Timestamp: time.Now(),
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := orchestrator.ProcessRequest(ctx, request)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}
