package workflows

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/aios/aios/internal/ai"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

// WorkflowEngine manages and executes complex AI workflows
type WorkflowEngine struct {
	logger       *logrus.Logger
	tracer       trace.Tracer
	orchestrator *ai.Orchestrator

	// Workflow management
	workflows  map[string]*Workflow
	executions map[string]*WorkflowExecution
	templates  map[string]*WorkflowTemplate
	mu         sync.RWMutex

	// Execution control
	executionQueue chan *WorkflowExecution
	workers        []*WorkflowWorker
	maxConcurrent  int

	// Configuration
	config WorkflowEngineConfig
}

// WorkflowEngineConfig represents workflow engine configuration
type WorkflowEngineConfig struct {
	MaxConcurrentWorkflows int           `json:"max_concurrent_workflows"`
	WorkerCount            int           `json:"worker_count"`
	ExecutionTimeout       time.Duration `json:"execution_timeout"`
	RetryAttempts          int           `json:"retry_attempts"`
	EnablePersistence      bool          `json:"enable_persistence"`
	EnableMetrics          bool          `json:"enable_metrics"`
}

// Workflow represents a complete AI workflow definition
type Workflow struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Version     string                 `json:"version"`
	Steps       []WorkflowStep         `json:"steps"`
	Triggers    []WorkflowTrigger      `json:"triggers"`
	Variables   map[string]interface{} `json:"variables"`
	Metadata    map[string]interface{} `json:"metadata"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	CreatedBy   string                 `json:"created_by"`
	IsActive    bool                   `json:"is_active"`
}

// WorkflowStep represents a single step in a workflow
type WorkflowStep struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Type        StepType               `json:"type"`
	Service     string                 `json:"service"`
	Action      string                 `json:"action"`
	Parameters  map[string]interface{} `json:"parameters"`
	Conditions  []StepCondition        `json:"conditions"`
	OnSuccess   []string               `json:"on_success"`
	OnFailure   []string               `json:"on_failure"`
	Timeout     time.Duration          `json:"timeout"`
	RetryPolicy *RetryPolicy           `json:"retry_policy,omitempty"`
	Parallel    bool                   `json:"parallel"`
}

// StepType represents different types of workflow steps
type StepType string

const (
	StepTypeLLM        StepType = "llm"
	StepTypeVision     StepType = "vision"
	StepTypeVoice      StepType = "voice"
	StepTypeNLP        StepType = "nlp"
	StepTypeRAG        StepType = "rag"
	StepTypeMultiModal StepType = "multimodal"
	StepTypeSystem     StepType = "system"
	StepTypeCondition  StepType = "condition"
	StepTypeLoop       StepType = "loop"
	StepTypeParallel   StepType = "parallel"
	StepTypeCustom     StepType = "custom"
)

// WorkflowTrigger represents workflow trigger conditions
type WorkflowTrigger struct {
	ID         string                 `json:"id"`
	Type       TriggerType            `json:"type"`
	Conditions map[string]interface{} `json:"conditions"`
	Schedule   string                 `json:"schedule,omitempty"`
	IsActive   bool                   `json:"is_active"`
}

// TriggerType represents different trigger types
type TriggerType string

const (
	TriggerTypeManual    TriggerType = "manual"
	TriggerTypeScheduled TriggerType = "scheduled"
	TriggerTypeEvent     TriggerType = "event"
	TriggerTypeWebhook   TriggerType = "webhook"
	TriggerTypeFile      TriggerType = "file"
	TriggerTypeAPI       TriggerType = "api"
)

// WorkflowExecution represents an active workflow execution
type WorkflowExecution struct {
	ID          string                 `json:"id"`
	WorkflowID  string                 `json:"workflow_id"`
	Status      ExecutionStatus        `json:"status"`
	StartTime   time.Time              `json:"start_time"`
	EndTime     *time.Time             `json:"end_time,omitempty"`
	CurrentStep string                 `json:"current_step"`
	StepResults map[string]StepResult  `json:"step_results"`
	Variables   map[string]interface{} `json:"variables"`
	Error       string                 `json:"error,omitempty"`
	Progress    float64                `json:"progress"`
	Metadata    map[string]interface{} `json:"metadata"`
	TriggerData map[string]interface{} `json:"trigger_data,omitempty"`
}

// ExecutionStatus represents workflow execution status
type ExecutionStatus string

const (
	StatusPending   ExecutionStatus = "pending"
	StatusRunning   ExecutionStatus = "running"
	StatusCompleted ExecutionStatus = "completed"
	StatusFailed    ExecutionStatus = "failed"
	StatusCancelled ExecutionStatus = "cancelled"
	StatusPaused    ExecutionStatus = "paused"
)

// StepResult represents the result of a workflow step
type StepResult struct {
	StepID    string                 `json:"step_id"`
	Status    ExecutionStatus        `json:"status"`
	StartTime time.Time              `json:"start_time"`
	EndTime   *time.Time             `json:"end_time,omitempty"`
	Output    interface{}            `json:"output,omitempty"`
	Error     string                 `json:"error,omitempty"`
	Duration  time.Duration          `json:"duration"`
	Attempts  int                    `json:"attempts"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// StepCondition represents conditions for step execution
type StepCondition struct {
	Variable string      `json:"variable"`
	Operator string      `json:"operator"` // eq, ne, gt, lt, contains, exists
	Value    interface{} `json:"value"`
}

// RetryPolicy defines retry behavior for failed steps
type RetryPolicy struct {
	MaxAttempts int           `json:"max_attempts"`
	Delay       time.Duration `json:"delay"`
	BackoffType string        `json:"backoff_type"` // fixed, exponential, linear
	MaxDelay    time.Duration `json:"max_delay"`
}

// WorkflowTemplate represents a reusable workflow template
type WorkflowTemplate struct {
	ID          string              `json:"id"`
	Name        string              `json:"name"`
	Description string              `json:"description"`
	Category    string              `json:"category"`
	Template    Workflow            `json:"template"`
	Parameters  []TemplateParameter `json:"parameters"`
	Tags        []string            `json:"tags"`
	CreatedAt   time.Time           `json:"created_at"`
	UsageCount  int                 `json:"usage_count"`
}

// TemplateParameter represents a configurable template parameter
type TemplateParameter struct {
	Name         string      `json:"name"`
	Type         string      `json:"type"`
	Description  string      `json:"description"`
	DefaultValue interface{} `json:"default_value,omitempty"`
	Required     bool        `json:"required"`
	Options      []string    `json:"options,omitempty"`
}

// WorkflowWorker processes workflow executions
type WorkflowWorker struct {
	ID     string
	engine *WorkflowEngine
	stopCh chan struct{}
	logger *logrus.Entry
}

// NewWorkflowEngine creates a new workflow engine
func NewWorkflowEngine(
	logger *logrus.Logger,
	orchestrator *ai.Orchestrator,
	config WorkflowEngineConfig,
) *WorkflowEngine {
	engine := &WorkflowEngine{
		logger:         logger,
		tracer:         otel.Tracer("ai.workflow_engine"),
		orchestrator:   orchestrator,
		workflows:      make(map[string]*Workflow),
		executions:     make(map[string]*WorkflowExecution),
		templates:      make(map[string]*WorkflowTemplate),
		executionQueue: make(chan *WorkflowExecution, 1000),
		maxConcurrent:  config.MaxConcurrentWorkflows,
		config:         config,
	}

	// Initialize workers
	engine.workers = make([]*WorkflowWorker, config.WorkerCount)
	for i := 0; i < config.WorkerCount; i++ {
		engine.workers[i] = &WorkflowWorker{
			ID:     fmt.Sprintf("worker-%d", i),
			engine: engine,
			stopCh: make(chan struct{}),
			logger: logger.WithField("worker_id", fmt.Sprintf("worker-%d", i)),
		}
	}

	// Load default templates
	engine.loadDefaultTemplates()

	return engine
}

// Start initializes the workflow engine
func (we *WorkflowEngine) Start(ctx context.Context) error {
	ctx, span := we.tracer.Start(ctx, "workflow_engine.Start")
	defer span.End()

	we.logger.Info("Starting workflow engine")

	// Start workers
	for _, worker := range we.workers {
		go worker.start()
	}

	we.logger.WithFields(logrus.Fields{
		"worker_count":      len(we.workers),
		"max_concurrent":    we.maxConcurrent,
		"execution_timeout": we.config.ExecutionTimeout,
	}).Info("Workflow engine started successfully")

	return nil
}

// Stop shuts down the workflow engine
func (we *WorkflowEngine) Stop(ctx context.Context) error {
	ctx, span := we.tracer.Start(ctx, "workflow_engine.Stop")
	defer span.End()

	we.logger.Info("Stopping workflow engine")

	// Stop workers
	for _, worker := range we.workers {
		close(worker.stopCh)
	}

	// Cancel running executions
	we.mu.Lock()
	for _, execution := range we.executions {
		if execution.Status == StatusRunning {
			execution.Status = StatusCancelled
			now := time.Now()
			execution.EndTime = &now
		}
	}
	we.mu.Unlock()

	we.logger.Info("Workflow engine stopped")
	return nil
}

// CreateWorkflow creates a new workflow
func (we *WorkflowEngine) CreateWorkflow(ctx context.Context, workflow *Workflow) error {
	ctx, span := we.tracer.Start(ctx, "workflow_engine.CreateWorkflow")
	defer span.End()

	// Validate workflow
	if err := we.validateWorkflow(workflow); err != nil {
		return fmt.Errorf("workflow validation failed: %w", err)
	}

	// Set metadata
	workflow.ID = generateWorkflowID()
	workflow.CreatedAt = time.Now()
	workflow.UpdatedAt = time.Now()
	workflow.IsActive = true

	we.mu.Lock()
	we.workflows[workflow.ID] = workflow
	we.mu.Unlock()

	we.logger.WithFields(logrus.Fields{
		"workflow_id":   workflow.ID,
		"workflow_name": workflow.Name,
		"step_count":    len(workflow.Steps),
	}).Info("Workflow created")

	return nil
}

// ExecuteWorkflow starts a workflow execution
func (we *WorkflowEngine) ExecuteWorkflow(ctx context.Context, workflowID string, triggerData map[string]interface{}) (*WorkflowExecution, error) {
	ctx, span := we.tracer.Start(ctx, "workflow_engine.ExecuteWorkflow")
	defer span.End()

	we.mu.RLock()
	workflow, exists := we.workflows[workflowID]
	we.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("workflow not found: %s", workflowID)
	}

	if !workflow.IsActive {
		return nil, fmt.Errorf("workflow is not active: %s", workflowID)
	}

	// Create execution
	execution := &WorkflowExecution{
		ID:          generateExecutionID(),
		WorkflowID:  workflowID,
		Status:      StatusPending,
		StartTime:   time.Now(),
		StepResults: make(map[string]StepResult),
		Variables:   make(map[string]interface{}),
		Progress:    0.0,
		Metadata:    make(map[string]interface{}),
		TriggerData: triggerData,
	}

	// Copy workflow variables
	for k, v := range workflow.Variables {
		execution.Variables[k] = v
	}

	// Add trigger data to variables
	for k, v := range triggerData {
		execution.Variables[k] = v
	}

	we.mu.Lock()
	we.executions[execution.ID] = execution
	we.mu.Unlock()

	// Queue for execution
	select {
	case we.executionQueue <- execution:
		we.logger.WithFields(logrus.Fields{
			"execution_id": execution.ID,
			"workflow_id":  workflowID,
		}).Info("Workflow execution queued")
	default:
		return nil, fmt.Errorf("execution queue is full")
	}

	return execution, nil
}

// GetExecution retrieves a workflow execution
func (we *WorkflowEngine) GetExecution(ctx context.Context, executionID string) (*WorkflowExecution, error) {
	ctx, span := we.tracer.Start(ctx, "workflow_engine.GetExecution")
	defer span.End()

	we.mu.RLock()
	execution, exists := we.executions[executionID]
	we.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("execution not found: %s", executionID)
	}

	// Return a copy
	executionCopy := *execution
	return &executionCopy, nil
}

// ListExecutions lists workflow executions with filtering
func (we *WorkflowEngine) ListExecutions(ctx context.Context, filter ExecutionFilter) ([]*WorkflowExecution, error) {
	ctx, span := we.tracer.Start(ctx, "workflow_engine.ListExecutions")
	defer span.End()

	we.mu.RLock()
	defer we.mu.RUnlock()

	var executions []*WorkflowExecution
	for _, execution := range we.executions {
		if we.matchesFilter(execution, filter) {
			executionCopy := *execution
			executions = append(executions, &executionCopy)
		}
	}

	return executions, nil
}

// ExecutionFilter represents filtering criteria for executions
type ExecutionFilter struct {
	WorkflowID string          `json:"workflow_id,omitempty"`
	Status     ExecutionStatus `json:"status,omitempty"`
	StartTime  *time.Time      `json:"start_time,omitempty"`
	EndTime    *time.Time      `json:"end_time,omitempty"`
	Limit      int             `json:"limit,omitempty"`
}

// Helper methods

func (we *WorkflowEngine) validateWorkflow(workflow *Workflow) error {
	if workflow.Name == "" {
		return fmt.Errorf("workflow name is required")
	}

	if len(workflow.Steps) == 0 {
		return fmt.Errorf("workflow must have at least one step")
	}

	// Validate steps
	stepIDs := make(map[string]bool)
	for _, step := range workflow.Steps {
		if step.ID == "" {
			return fmt.Errorf("step ID is required")
		}

		if stepIDs[step.ID] {
			return fmt.Errorf("duplicate step ID: %s", step.ID)
		}
		stepIDs[step.ID] = true

		if step.Type == "" {
			return fmt.Errorf("step type is required for step: %s", step.ID)
		}
	}

	return nil
}

func (we *WorkflowEngine) matchesFilter(execution *WorkflowExecution, filter ExecutionFilter) bool {
	if filter.WorkflowID != "" && execution.WorkflowID != filter.WorkflowID {
		return false
	}

	if filter.Status != "" && execution.Status != filter.Status {
		return false
	}

	if filter.StartTime != nil && execution.StartTime.Before(*filter.StartTime) {
		return false
	}

	if filter.EndTime != nil && execution.EndTime != nil && execution.EndTime.After(*filter.EndTime) {
		return false
	}

	return true
}

func (we *WorkflowEngine) loadDefaultTemplates() {
	// Load built-in workflow templates
	templates := []*WorkflowTemplate{
		{
			ID:          "content-analysis",
			Name:        "Content Analysis Pipeline",
			Description: "Analyze text content using multiple AI services",
			Category:    "analysis",
			Template: Workflow{
				Name:        "Content Analysis",
				Description: "Multi-step content analysis workflow",
				Steps: []WorkflowStep{
					{
						ID:      "nlp-analysis",
						Name:    "NLP Analysis",
						Type:    StepTypeNLP,
						Service: "nlp",
						Action:  "analyze_text",
						Parameters: map[string]interface{}{
							"text": "{{input.text}}",
						},
					},
					{
						ID:      "sentiment-analysis",
						Name:    "Sentiment Analysis",
						Type:    StepTypeLLM,
						Service: "llm",
						Action:  "analyze_sentiment",
						Parameters: map[string]interface{}{
							"text": "{{input.text}}",
						},
					},
				},
			},
			Tags:      []string{"analysis", "nlp", "content"},
			CreatedAt: time.Now(),
		},
	}

	we.mu.Lock()
	for _, template := range templates {
		we.templates[template.ID] = template
	}
	we.mu.Unlock()

	we.logger.WithField("template_count", len(templates)).Info("Default workflow templates loaded")
}

func generateWorkflowID() string {
	return fmt.Sprintf("workflow_%d", time.Now().UnixNano())
}

func generateExecutionID() string {
	return fmt.Sprintf("exec_%d", time.Now().UnixNano())
}

// WorkflowWorker methods

func (w *WorkflowWorker) start() {
	w.logger.Info("Workflow worker started")

	for {
		select {
		case execution := <-w.engine.executionQueue:
			w.executeWorkflow(execution)
		case <-w.stopCh:
			w.logger.Info("Workflow worker stopped")
			return
		}
	}
}

func (w *WorkflowWorker) executeWorkflow(execution *WorkflowExecution) {
	ctx := context.Background()
	ctx, span := w.engine.tracer.Start(ctx, "workflow_worker.executeWorkflow")
	defer span.End()

	w.logger.WithFields(logrus.Fields{
		"execution_id": execution.ID,
		"workflow_id":  execution.WorkflowID,
	}).Info("Starting workflow execution")

	// Update execution status
	w.engine.mu.Lock()
	execution.Status = StatusRunning
	w.engine.mu.Unlock()

	// Get workflow
	w.engine.mu.RLock()
	workflow, exists := w.engine.workflows[execution.WorkflowID]
	w.engine.mu.RUnlock()

	if !exists {
		w.failExecution(execution, "workflow not found")
		return
	}

	// Execute steps
	for i, step := range workflow.Steps {
		if execution.Status != StatusRunning {
			break
		}

		// Check conditions
		if !w.evaluateConditions(step.Conditions, execution.Variables) {
			w.logger.WithField("step_id", step.ID).Info("Step conditions not met, skipping")
			continue
		}

		// Execute step
		result := w.executeStep(ctx, step, execution)

		// Update execution
		w.engine.mu.Lock()
		execution.StepResults[step.ID] = result
		execution.CurrentStep = step.ID
		execution.Progress = float64(i+1) / float64(len(workflow.Steps))
		w.engine.mu.Unlock()

		// Handle step result
		if result.Status == StatusFailed {
			if len(step.OnFailure) > 0 {
				// Handle failure path
				w.logger.WithField("step_id", step.ID).Info("Step failed, executing failure path")
				// TODO: Implement failure path execution
			} else {
				w.failExecution(execution, fmt.Sprintf("step %s failed: %s", step.ID, result.Error))
				return
			}
		}

		// Update variables with step output
		if result.Output != nil {
			if outputMap, ok := result.Output.(map[string]interface{}); ok {
				for k, v := range outputMap {
					execution.Variables[fmt.Sprintf("%s.%s", step.ID, k)] = v
				}
			}
			execution.Variables[step.ID] = result.Output
		}
	}

	// Complete execution
	w.completeExecution(execution)
}

func (w *WorkflowWorker) executeStep(ctx context.Context, step WorkflowStep, execution *WorkflowExecution) StepResult {
	startTime := time.Now()
	result := StepResult{
		StepID:    step.ID,
		Status:    StatusRunning,
		StartTime: startTime,
		Attempts:  1,
		Metadata:  make(map[string]interface{}),
	}

	w.logger.WithField("step_id", step.ID).Info("Executing workflow step")

	// Resolve parameters
	resolvedParams := w.resolveParameters(step.Parameters, execution.Variables)

	// Execute based on step type
	var output interface{}
	var err error

	switch step.Type {
	case StepTypeLLM:
		output, err = w.executeLLMStep(ctx, step, resolvedParams)
	case StepTypeVision:
		output, err = w.executeVisionStep(ctx, step, resolvedParams)
	case StepTypeVoice:
		output, err = w.executeVoiceStep(ctx, step, resolvedParams)
	case StepTypeNLP:
		output, err = w.executeNLPStep(ctx, step, resolvedParams)
	case StepTypeRAG:
		output, err = w.executeRAGStep(ctx, step, resolvedParams)
	case StepTypeSystem:
		output, err = w.executeSystemStep(ctx, step, resolvedParams)
	default:
		err = fmt.Errorf("unsupported step type: %s", step.Type)
	}

	// Update result
	endTime := time.Now()
	result.EndTime = &endTime
	result.Duration = endTime.Sub(startTime)

	if err != nil {
		result.Status = StatusFailed
		result.Error = err.Error()
		w.logger.WithError(err).WithField("step_id", step.ID).Error("Step execution failed")
	} else {
		result.Status = StatusCompleted
		result.Output = output
		w.logger.WithField("step_id", step.ID).Info("Step execution completed")
	}

	return result
}

func (w *WorkflowWorker) executeLLMStep(ctx context.Context, step WorkflowStep, params map[string]interface{}) (interface{}, error) {
	llmService := w.engine.orchestrator.GetLLMService()

	query, ok := params["query"].(string)
	if !ok {
		return nil, fmt.Errorf("query parameter is required for LLM step")
	}

	response, err := llmService.ProcessQuery(ctx, query)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"text":        response.Text,
		"confidence":  response.Confidence,
		"tokens_used": response.TokensUsed,
		"model":       response.Model,
	}, nil
}

func (w *WorkflowWorker) executeVisionStep(ctx context.Context, step WorkflowStep, params map[string]interface{}) (interface{}, error) {
	cvService := w.engine.orchestrator.GetCVService()

	imageData, ok := params["image"].([]byte)
	if !ok {
		return nil, fmt.Errorf("image parameter is required for vision step")
	}

	analysis, err := cvService.AnalyzeScreen(ctx, imageData)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"elements":      analysis.Elements,
		"text":          analysis.Text,
		"actions":       analysis.Actions,
		"accessibility": analysis.Accessibility,
	}, nil
}

func (w *WorkflowWorker) executeVoiceStep(ctx context.Context, step WorkflowStep, params map[string]interface{}) (interface{}, error) {
	voiceService := w.engine.orchestrator.GetVoiceService()

	if audioData, ok := params["audio"].([]byte); ok {
		// Speech to text
		recognition, err := voiceService.SpeechToText(ctx, audioData)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{
			"text":       recognition.Text,
			"confidence": recognition.Confidence,
		}, nil
	} else if text, ok := params["text"].(string); ok {
		// Text to speech
		synthesis, err := voiceService.TextToSpeech(ctx, text)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{
			"audio":    synthesis.Audio,
			"duration": synthesis.Duration,
		}, nil
	}

	return nil, fmt.Errorf("either audio or text parameter is required for voice step")
}

func (w *WorkflowWorker) executeNLPStep(ctx context.Context, step WorkflowStep, params map[string]interface{}) (interface{}, error) {
	nlpService := w.engine.orchestrator.GetNLPService()

	text, ok := params["text"].(string)
	if !ok {
		return nil, fmt.Errorf("text parameter is required for NLP step")
	}

	intent, err := nlpService.ParseIntent(ctx, text)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"intent":     intent.Intent,
		"confidence": intent.Confidence,
		"entities":   intent.Entities,
	}, nil
}

func (w *WorkflowWorker) executeRAGStep(ctx context.Context, step WorkflowStep, params map[string]interface{}) (interface{}, error) {
	// TODO: Implement RAG step execution
	return map[string]interface{}{
		"result": "RAG step executed",
	}, nil
}

func (w *WorkflowWorker) executeSystemStep(ctx context.Context, step WorkflowStep, params map[string]interface{}) (interface{}, error) {
	action, ok := params["action"].(string)
	if !ok {
		return nil, fmt.Errorf("action parameter is required for system step")
	}

	switch action {
	case "delay":
		if duration, ok := params["duration"].(string); ok {
			if d, err := time.ParseDuration(duration); err == nil {
				time.Sleep(d)
				return map[string]interface{}{"delayed": duration}, nil
			}
		}
		return nil, fmt.Errorf("invalid duration for delay action")
	case "log":
		if message, ok := params["message"].(string); ok {
			w.logger.Info(message)
			return map[string]interface{}{"logged": message}, nil
		}
		return nil, fmt.Errorf("message parameter required for log action")
	default:
		return nil, fmt.Errorf("unsupported system action: %s", action)
	}
}

func (w *WorkflowWorker) evaluateConditions(conditions []StepCondition, variables map[string]interface{}) bool {
	if len(conditions) == 0 {
		return true
	}

	for _, condition := range conditions {
		if !w.evaluateCondition(condition, variables) {
			return false
		}
	}

	return true
}

func (w *WorkflowWorker) evaluateCondition(condition StepCondition, variables map[string]interface{}) bool {
	value, exists := variables[condition.Variable]
	if !exists && condition.Operator != "exists" {
		return false
	}

	switch condition.Operator {
	case "exists":
		return exists
	case "eq":
		return value == condition.Value
	case "ne":
		return value != condition.Value
	case "gt":
		if v1, ok := value.(float64); ok {
			if v2, ok := condition.Value.(float64); ok {
				return v1 > v2
			}
		}
	case "lt":
		if v1, ok := value.(float64); ok {
			if v2, ok := condition.Value.(float64); ok {
				return v1 < v2
			}
		}
	case "contains":
		if v1, ok := value.(string); ok {
			if v2, ok := condition.Value.(string); ok {
				return fmt.Sprintf("%v", v1) == v2
			}
		}
	}

	return false
}

func (w *WorkflowWorker) resolveParameters(params map[string]interface{}, variables map[string]interface{}) map[string]interface{} {
	resolved := make(map[string]interface{})

	for key, value := range params {
		if strValue, ok := value.(string); ok {
			// Simple template resolution - replace {{variable}} with actual values
			for varName, varValue := range variables {
				placeholder := fmt.Sprintf("{{%s}}", varName)
				if strValue == placeholder {
					resolved[key] = varValue
					break
				}
			}
			if resolved[key] == nil {
				resolved[key] = strValue
			}
		} else {
			resolved[key] = value
		}
	}

	return resolved
}

func (w *WorkflowWorker) failExecution(execution *WorkflowExecution, errorMsg string) {
	w.engine.mu.Lock()
	execution.Status = StatusFailed
	execution.Error = errorMsg
	now := time.Now()
	execution.EndTime = &now
	w.engine.mu.Unlock()

	w.logger.WithFields(logrus.Fields{
		"execution_id": execution.ID,
		"error":        errorMsg,
	}).Error("Workflow execution failed")
}

func (w *WorkflowWorker) completeExecution(execution *WorkflowExecution) {
	w.engine.mu.Lock()
	execution.Status = StatusCompleted
	execution.Progress = 1.0
	now := time.Now()
	execution.EndTime = &now
	w.engine.mu.Unlock()

	w.logger.WithField("execution_id", execution.ID).Info("Workflow execution completed")
}
