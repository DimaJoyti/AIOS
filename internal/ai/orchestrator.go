package ai

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/aios/aios/pkg/models"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

// Orchestrator coordinates all AI services and manages complex workflows
type Orchestrator struct {
	config              AIServiceConfig
	logger              *logrus.Logger
	tracer              trace.Tracer
	llmService          LanguageModelService
	cvService           ComputerVisionService
	optimizationService SystemOptimizationService
	voiceService        VoiceService
	nlpService          NaturalLanguageService
	modelManager        ModelManager
	activeWorkflows     map[string]*WorkflowExecution
	mu                  sync.RWMutex
}

// WorkflowExecution represents an active workflow execution
type WorkflowExecution struct {
	ID        string
	Workflow  *models.AIWorkflow
	Status    string
	Results   map[string]interface{}
	Errors    []models.WorkflowError
	StartTime time.Time
	Context   context.Context
	Cancel    context.CancelFunc
}

// NewOrchestrator creates a new AI orchestrator
func NewOrchestrator(config AIServiceConfig, logger *logrus.Logger) *Orchestrator {
	tracer := otel.Tracer("ai-orchestrator")

	return &Orchestrator{
		config:          config,
		logger:          logger,
		tracer:          tracer,
		activeWorkflows: make(map[string]*WorkflowExecution),
	}
}

// Initialize sets up all AI services
func (o *Orchestrator) Initialize(ctx context.Context) error {
	o.logger.Info("Initializing AI orchestrator")

	// Initialize individual services
	o.llmService = NewLLMService(o.config, o.logger)
	o.cvService = NewCVService(o.config, o.logger)
	o.optimizationService = NewOptimizationService(o.config, o.logger)
	// TODO: Initialize voice and NLP services
	// TODO: Initialize model manager

	o.logger.Info("AI orchestrator initialized successfully")
	return nil
}

// ProcessRequest processes a complex AI request that may involve multiple services
func (o *Orchestrator) ProcessRequest(ctx context.Context, request *models.AIRequest) (*models.AIResponse, error) {
	ctx, span := o.tracer.Start(ctx, "orchestrator.ProcessRequest")
	defer span.End()

	start := time.Now()
	
	o.logger.WithFields(logrus.Fields{
		"request_id":   request.ID,
		"request_type": request.Type,
		"priority":     request.Priority,
	}).Info("Processing AI request")

	// Route request to appropriate service
	serviceID, err := o.RouteRequest(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("failed to route request: %w", err)
	}

	// Process request based on type
	var result interface{}
	var confidence float64
	var model string

	switch request.Type {
	case "chat":
		result, confidence, model, err = o.processChatRequest(ctx, request)
	case "vision":
		result, confidence, model, err = o.processVisionRequest(ctx, request)
	case "optimization":
		result, confidence, model, err = o.processOptimizationRequest(ctx, request)
	case "voice":
		result, confidence, model, err = o.processVoiceRequest(ctx, request)
	case "analysis":
		result, confidence, model, err = o.processAnalysisRequest(ctx, request)
	default:
		return nil, fmt.Errorf("unsupported request type: %s", request.Type)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to process %s request: %w", request.Type, err)
	}

	processingTime := time.Since(start)

	response := &models.AIResponse{
		RequestID:      request.ID,
		Type:           request.Type,
		Result:         result,
		Confidence:     confidence,
		ProcessingTime: processingTime,
		Model:          model,
		Metadata: map[string]interface{}{
			"service_id":      serviceID,
			"processing_time": processingTime.String(),
		},
		Timestamp: time.Now(),
	}

	o.logger.WithFields(logrus.Fields{
		"request_id":      request.ID,
		"processing_time": processingTime,
		"confidence":      confidence,
		"service_id":      serviceID,
	}).Info("AI request processed successfully")

	return response, nil
}

// GetServiceStatus gets the status of all AI services
func (o *Orchestrator) GetServiceStatus(ctx context.Context) (*models.AIServiceStatus, error) {
	ctx, span := o.tracer.Start(ctx, "orchestrator.GetServiceStatus")
	defer span.End()

	services := map[string]models.ServiceHealth{
		"llm": {
			Name:         "Language Model Service",
			Status:       "healthy",
			Uptime:       24 * time.Hour, // Mock uptime
			RequestCount: 1000,           // Mock request count
			ErrorRate:    0.01,           // Mock error rate
			Latency:      150 * time.Millisecond,
			LastCheck:    time.Now(),
		},
		"cv": {
			Name:         "Computer Vision Service",
			Status:       "healthy",
			Uptime:       24 * time.Hour,
			RequestCount: 500,
			ErrorRate:    0.02,
			Latency:      300 * time.Millisecond,
			LastCheck:    time.Now(),
		},
		"optimization": {
			Name:         "System Optimization Service",
			Status:       "healthy",
			Uptime:       24 * time.Hour,
			RequestCount: 200,
			ErrorRate:    0.005,
			Latency:      500 * time.Millisecond,
			LastCheck:    time.Now(),
		},
	}

	// TODO: Get actual model status from model manager
	modelStatuses := []models.ModelStatus{
		{
			ID:           "llama2",
			Status:       "loaded",
			LoadTime:     5 * time.Second,
			MemoryUsage:  3800000000, // 3.8GB
			RequestCount: 1000,
			ErrorCount:   10,
			LastUsed:     time.Now().Add(-5 * time.Minute),
			Timestamp:    time.Now(),
		},
	}

	// Determine overall health
	overallHealth := "healthy"
	for _, service := range services {
		if service.Status != "healthy" {
			overallHealth = "degraded"
			break
		}
	}

	status := &models.AIServiceStatus{
		Services:      services,
		Models:        modelStatuses,
		OverallHealth: overallHealth,
		Timestamp:     time.Now(),
	}

	return status, nil
}

// RouteRequest routes a request to the appropriate AI service
func (o *Orchestrator) RouteRequest(ctx context.Context, request *models.AIRequest) (string, error) {
	switch request.Type {
	case "chat", "text", "code", "translate", "summarize":
		return "llm", nil
	case "vision", "ocr", "ui_analysis", "image_classification":
		return "cv", nil
	case "optimization", "performance", "prediction":
		return "optimization", nil
	case "voice", "speech", "tts", "stt":
		return "voice", nil
	case "intent", "entity", "sentiment":
		return "nlp", nil
	default:
		return "", fmt.Errorf("unknown request type: %s", request.Type)
	}
}

// AggregateResults aggregates results from multiple AI services
func (o *Orchestrator) AggregateResults(ctx context.Context, results []models.AIResult) (*models.AggregatedResult, error) {
	ctx, span := o.tracer.Start(ctx, "orchestrator.AggregateResults")
	defer span.End()

	if len(results) == 0 {
		return nil, fmt.Errorf("no results to aggregate")
	}

	// Calculate weighted average confidence
	var totalConfidence, totalWeight float64
	for _, result := range results {
		weight := 1.0 // TODO: Implement service-specific weights
		totalConfidence += result.Confidence * weight
		totalWeight += weight
	}
	averageConfidence := totalConfidence / totalWeight

	// Determine aggregation method based on result types
	method := o.determineAggregationMethod(results)

	// Perform aggregation
	consensus, err := o.performAggregation(results, method)
	if err != nil {
		return nil, fmt.Errorf("failed to aggregate results: %w", err)
	}

	aggregated := &models.AggregatedResult{
		Results:    results,
		Consensus:  consensus,
		Confidence: averageConfidence,
		Method:     method,
		Metadata: map[string]interface{}{
			"result_count": len(results),
			"services":     o.extractServiceIDs(results),
		},
		Timestamp: time.Now(),
	}

	return aggregated, nil
}

// ManageWorkflow manages complex AI workflows
func (o *Orchestrator) ManageWorkflow(ctx context.Context, workflow *models.AIWorkflow) (*models.WorkflowResult, error) {
	ctx, span := o.tracer.Start(ctx, "orchestrator.ManageWorkflow")
	defer span.End()

	o.logger.WithFields(logrus.Fields{
		"workflow_id":   workflow.ID,
		"workflow_name": workflow.Name,
		"steps":         len(workflow.Steps),
	}).Info("Starting workflow execution")

	// Create workflow execution context
	workflowCtx, cancel := context.WithTimeout(ctx, workflow.Timeout)
	defer cancel()

	execution := &WorkflowExecution{
		ID:        workflow.ID,
		Workflow:  workflow,
		Status:    "running",
		Results:   make(map[string]interface{}),
		Errors:    []models.WorkflowError{},
		StartTime: time.Now(),
		Context:   workflowCtx,
		Cancel:    cancel,
	}

	// Store active workflow
	o.mu.Lock()
	o.activeWorkflows[workflow.ID] = execution
	o.mu.Unlock()

	// Execute workflow steps
	result := o.executeWorkflow(execution)

	// Clean up
	o.mu.Lock()
	delete(o.activeWorkflows, workflow.ID)
	o.mu.Unlock()

	o.logger.WithFields(logrus.Fields{
		"workflow_id": workflow.ID,
		"status":      result.Status,
		"duration":    result.Duration,
		"errors":      len(result.Errors),
	}).Info("Workflow execution completed")

	return result, nil
}

// Helper methods for request processing

func (o *Orchestrator) processChatRequest(ctx context.Context, request *models.AIRequest) (interface{}, float64, string, error) {
	if o.llmService == nil {
		return nil, 0, "", fmt.Errorf("LLM service not available")
	}

	// Extract message from input
	message, ok := request.Input.(string)
	if !ok {
		return nil, 0, "", fmt.Errorf("invalid input type for chat request")
	}

	// Get conversation ID from context
	conversationID := "default"
	if id, exists := request.Context["conversation_id"]; exists {
		if strID, ok := id.(string); ok {
			conversationID = strID
		}
	}

	response, err := o.llmService.Chat(ctx, message, conversationID)
	if err != nil {
		return nil, 0, "", err
	}

	return response, response.Confidence, "llama2", nil
}

func (o *Orchestrator) processVisionRequest(ctx context.Context, request *models.AIRequest) (interface{}, float64, string, error) {
	if o.cvService == nil {
		return nil, 0, "", fmt.Errorf("CV service not available")
	}

	// Extract image data from input
	imageData, ok := request.Input.([]byte)
	if !ok {
		return nil, 0, "", fmt.Errorf("invalid input type for vision request")
	}

	// Determine specific vision task
	task := "analyze_screen" // default
	if t, exists := request.Parameters["task"]; exists {
		if taskStr, ok := t.(string); ok {
			task = taskStr
		}
	}

	switch task {
	case "analyze_screen":
		result, err := o.cvService.AnalyzeScreen(ctx, imageData)
		return result, 0.8, "cv_model", err
	case "ocr":
		result, err := o.cvService.RecognizeText(ctx, imageData)
		return result, result.Confidence, "ocr_model", err
	case "classify":
		result, err := o.cvService.ClassifyImage(ctx, imageData)
		return result, result.Confidence, "classification_model", err
	default:
		return nil, 0, "", fmt.Errorf("unsupported vision task: %s", task)
	}
}

func (o *Orchestrator) processOptimizationRequest(ctx context.Context, request *models.AIRequest) (interface{}, float64, string, error) {
	if o.optimizationService == nil {
		return nil, 0, "", fmt.Errorf("optimization service not available")
	}

	// Determine optimization task
	task := "analyze" // default
	if t, exists := request.Parameters["task"]; exists {
		if taskStr, ok := t.(string); ok {
			task = taskStr
		}
	}

	switch task {
	case "analyze":
		result, err := o.optimizationService.AnalyzePerformance(ctx)
		return result, 0.9, "optimization_model", err
	case "predict":
		timeframe := 1 * time.Hour // default
		if tf, exists := request.Parameters["timeframe"]; exists {
			if duration, ok := tf.(time.Duration); ok {
				timeframe = duration
			}
		}
		result, err := o.optimizationService.PredictUsage(ctx, timeframe)
		return result, result.Confidence, "prediction_model", err
	case "recommend":
		result, err := o.optimizationService.GenerateRecommendations(ctx)
		return result, 0.85, "recommendation_model", err
	default:
		return nil, 0, "", fmt.Errorf("unsupported optimization task: %s", task)
	}
}

func (o *Orchestrator) processVoiceRequest(ctx context.Context, request *models.AIRequest) (interface{}, float64, string, error) {
	// TODO: Implement voice request processing
	return nil, 0, "", fmt.Errorf("voice service not yet implemented")
}

func (o *Orchestrator) processAnalysisRequest(ctx context.Context, request *models.AIRequest) (interface{}, float64, string, error) {
	// TODO: Implement analysis request processing
	return nil, 0, "", fmt.Errorf("analysis service not yet implemented")
}

// Helper methods for workflow execution

func (o *Orchestrator) executeWorkflow(execution *WorkflowExecution) *models.WorkflowResult {
	start := time.Now()
	
	// Build dependency graph
	dependencyGraph := o.buildDependencyGraph(execution.Workflow.Steps)
	
	// Execute steps in dependency order
	for _, step := range dependencyGraph {
		if err := o.executeWorkflowStep(execution, step); err != nil {
			execution.Status = "failed"
			execution.Errors = append(execution.Errors, models.WorkflowError{
				StepID:      step.ID,
				Error:       err.Error(),
				Recoverable: false,
				Timestamp:   time.Now(),
			})
			break
		}
	}

	if execution.Status != "failed" {
		execution.Status = "completed"
	}

	return &models.WorkflowResult{
		WorkflowID: execution.ID,
		Status:     execution.Status,
		Results:    execution.Results,
		Errors:     execution.Errors,
		Duration:   time.Since(start),
		Timestamp:  time.Now(),
	}
}

func (o *Orchestrator) executeWorkflowStep(execution *WorkflowExecution, step models.WorkflowStep) error {
	// Check if dependencies are satisfied
	for _, depID := range step.Dependencies {
		if _, exists := execution.Results[depID]; !exists {
			return fmt.Errorf("dependency %s not satisfied for step %s", depID, step.ID)
		}
	}

	// Create AI request for the step
	request := &models.AIRequest{
		ID:         uuid.New().String(),
		Type:       step.Type,
		Input:      step.Input,
		Parameters: step.Parameters,
		Timeout:    step.Timeout,
		Timestamp:  time.Now(),
	}

	// Process the request
	response, err := o.ProcessRequest(execution.Context, request)
	if err != nil {
		return fmt.Errorf("failed to execute step %s: %w", step.ID, err)
	}

	// Store result
	execution.Results[step.ID] = response.Result

	return nil
}

func (o *Orchestrator) buildDependencyGraph(steps []models.WorkflowStep) []models.WorkflowStep {
	// TODO: Implement proper topological sort for dependency resolution
	// For now, return steps in original order
	return steps
}

func (o *Orchestrator) determineAggregationMethod(results []models.AIResult) string {
	// Simple method determination based on result types
	if len(results) == 1 {
		return "single"
	}
	
	// Check if all results are from the same type of service
	firstType := results[0].Type
	allSameType := true
	for _, result := range results {
		if result.Type != firstType {
			allSameType = false
			break
		}
	}
	
	if allSameType {
		return "weighted_average"
	}
	
	return "consensus"
}

func (o *Orchestrator) performAggregation(results []models.AIResult, method string) (interface{}, error) {
	switch method {
	case "single":
		return results[0].Result, nil
	case "weighted_average":
		// TODO: Implement weighted average aggregation
		return results[0].Result, nil
	case "consensus":
		// TODO: Implement consensus aggregation
		return results[0].Result, nil
	default:
		return nil, fmt.Errorf("unsupported aggregation method: %s", method)
	}
}

func (o *Orchestrator) extractServiceIDs(results []models.AIResult) []string {
	serviceIDs := make([]string, len(results))
	for i, result := range results {
		serviceIDs[i] = result.ServiceID
	}
	return serviceIDs
}
