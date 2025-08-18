package workflow

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// DefaultWorkflowEngine implements the WorkflowEngine interface
type DefaultWorkflowEngine struct {
	logger             *logrus.Logger
	tracer             trace.Tracer
	actionExecutor     ActionExecutor
	triggerManager     TriggerManager
	scheduleManager    ScheduleManager
	integrationManager IntegrationManager

	// In-memory storage (in production, this would be a database)
	workflows  map[string]*Workflow
	executions map[string]*WorkflowExecution
	templates  map[string]*WorkflowTemplate

	// Execution management
	executionQueue chan *WorkflowExecution
	workers        int
	stopChan       chan struct{}
	wg             sync.WaitGroup
	mu             sync.RWMutex
}

// WorkflowEngineConfig represents configuration for the workflow engine
type WorkflowEngineConfig struct {
	Workers            int           `json:"workers"`
	QueueSize          int           `json:"queue_size"`
	DefaultTimeout     time.Duration `json:"default_timeout"`
	MaxRetries         int           `json:"max_retries"`
	EnableScheduling   bool          `json:"enable_scheduling"`
	EnableIntegrations bool          `json:"enable_integrations"`
}

// NewDefaultWorkflowEngine creates a new workflow engine
func NewDefaultWorkflowEngine(config *WorkflowEngineConfig, logger *logrus.Logger) (WorkflowEngine, error) {
	if config == nil {
		config = &WorkflowEngineConfig{
			Workers:            5,
			QueueSize:          1000,
			DefaultTimeout:     30 * time.Minute,
			MaxRetries:         3,
			EnableScheduling:   true,
			EnableIntegrations: true,
		}
	}

	engine := &DefaultWorkflowEngine{
		logger:         logger,
		tracer:         otel.Tracer("workflow.engine"),
		workflows:      make(map[string]*Workflow),
		executions:     make(map[string]*WorkflowExecution),
		templates:      make(map[string]*WorkflowTemplate),
		executionQueue: make(chan *WorkflowExecution, config.QueueSize),
		workers:        config.Workers,
		stopChan:       make(chan struct{}),
	}

	// Initialize sub-components
	var err error
	engine.actionExecutor, err = NewDefaultActionExecutor(logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create action executor: %w", err)
	}

	engine.triggerManager, err = NewDefaultTriggerManager(logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create trigger manager: %w", err)
	}

	if config.EnableScheduling {
		engine.scheduleManager, err = NewDefaultScheduleManager(logger)
		if err != nil {
			return nil, fmt.Errorf("failed to create schedule manager: %w", err)
		}
	}

	if config.EnableIntegrations {
		engine.integrationManager, err = NewDefaultIntegrationManager(logger)
		if err != nil {
			return nil, fmt.Errorf("failed to create integration manager: %w", err)
		}
	}

	// Start worker goroutines
	engine.startWorkers()

	// Create default templates
	engine.createDefaultTemplates()

	return engine, nil
}

// Workflow management

// CreateWorkflow creates a new workflow
func (we *DefaultWorkflowEngine) CreateWorkflow(ctx context.Context, workflow *Workflow) (*Workflow, error) {
	ctx, span := we.tracer.Start(ctx, "workflow_engine.create_workflow")
	defer span.End()

	span.SetAttributes(
		attribute.String("workflow.name", workflow.Name),
		attribute.String("workflow.created_by", workflow.CreatedBy),
	)

	// Generate ID if not provided
	if workflow.ID == "" {
		workflow.ID = uuid.New().String()
	}

	// Set timestamps
	now := time.Now()
	workflow.CreatedAt = now
	workflow.UpdatedAt = now

	// Set default values
	if workflow.Status == "" {
		workflow.Status = WorkflowStatusDraft
	}
	if workflow.Version == "" {
		workflow.Version = "1.0.0"
	}
	if workflow.Timeout == 0 {
		workflow.Timeout = 30 * time.Minute
	}

	// Validate workflow
	if err := we.validateWorkflow(workflow); err != nil {
		return nil, fmt.Errorf("workflow validation failed: %w", err)
	}

	// Store workflow
	we.mu.Lock()
	we.workflows[workflow.ID] = workflow
	we.mu.Unlock()

	// Register triggers
	for _, trigger := range workflow.Triggers {
		if err := we.triggerManager.RegisterTrigger(ctx, workflow.ID, trigger); err != nil {
			we.logger.WithError(err).Warn("Failed to register trigger")
		}
	}

	// Schedule workflow if needed
	if workflow.Schedule != nil && we.scheduleManager != nil {
		if err := we.scheduleManager.ScheduleWorkflow(ctx, workflow.ID, workflow.Schedule); err != nil {
			we.logger.WithError(err).Warn("Failed to schedule workflow")
		}
	}

	we.logger.WithFields(logrus.Fields{
		"workflow_id":   workflow.ID,
		"workflow_name": workflow.Name,
		"created_by":    workflow.CreatedBy,
		"triggers":      len(workflow.Triggers),
		"actions":       len(workflow.Actions),
	}).Info("Workflow created successfully")

	return workflow, nil
}

// GetWorkflow retrieves a workflow by ID
func (we *DefaultWorkflowEngine) GetWorkflow(ctx context.Context, workflowID string) (*Workflow, error) {
	ctx, span := we.tracer.Start(ctx, "workflow_engine.get_workflow")
	defer span.End()

	span.SetAttributes(attribute.String("workflow.id", workflowID))

	we.mu.RLock()
	workflow, exists := we.workflows[workflowID]
	we.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("workflow not found: %s", workflowID)
	}

	return workflow, nil
}

// UpdateWorkflow updates an existing workflow
func (we *DefaultWorkflowEngine) UpdateWorkflow(ctx context.Context, workflow *Workflow) (*Workflow, error) {
	ctx, span := we.tracer.Start(ctx, "workflow_engine.update_workflow")
	defer span.End()

	span.SetAttributes(attribute.String("workflow.id", workflow.ID))

	we.mu.Lock()
	existing, exists := we.workflows[workflow.ID]
	if !exists {
		we.mu.Unlock()
		return nil, fmt.Errorf("workflow not found: %s", workflow.ID)
	}

	// Update timestamp and preserve creation info
	workflow.UpdatedAt = time.Now()
	workflow.CreatedAt = existing.CreatedAt
	workflow.CreatedBy = existing.CreatedBy

	// Validate updated workflow
	if err := we.validateWorkflow(workflow); err != nil {
		we.mu.Unlock()
		return nil, fmt.Errorf("workflow validation failed: %w", err)
	}

	// Store updated workflow
	we.workflows[workflow.ID] = workflow
	we.mu.Unlock()

	we.logger.WithFields(logrus.Fields{
		"workflow_id":   workflow.ID,
		"workflow_name": workflow.Name,
		"updated_by":    workflow.UpdatedBy,
	}).Info("Workflow updated successfully")

	return workflow, nil
}

// DeleteWorkflow deletes a workflow
func (we *DefaultWorkflowEngine) DeleteWorkflow(ctx context.Context, workflowID string) error {
	ctx, span := we.tracer.Start(ctx, "workflow_engine.delete_workflow")
	defer span.End()

	span.SetAttributes(attribute.String("workflow.id", workflowID))

	we.mu.Lock()
	workflow, exists := we.workflows[workflowID]
	if !exists {
		we.mu.Unlock()
		return fmt.Errorf("workflow not found: %s", workflowID)
	}

	// Remove triggers
	for _, trigger := range workflow.Triggers {
		_ = we.triggerManager.UnregisterTrigger(ctx, trigger.ID)
	}

	// Unschedule workflow
	if workflow.Schedule != nil && we.scheduleManager != nil {
		_ = we.scheduleManager.UnscheduleWorkflow(ctx, workflowID)
	}

	// Delete workflow
	delete(we.workflows, workflowID)
	we.mu.Unlock()

	we.logger.WithFields(logrus.Fields{
		"workflow_id":   workflowID,
		"workflow_name": workflow.Name,
	}).Info("Workflow deleted successfully")

	return nil
}

// ListWorkflows lists workflows with filtering
func (we *DefaultWorkflowEngine) ListWorkflows(ctx context.Context, filter *WorkflowFilter) ([]*Workflow, error) {
	ctx, span := we.tracer.Start(ctx, "workflow_engine.list_workflows")
	defer span.End()

	we.mu.RLock()
	var workflows []*Workflow
	for _, workflow := range we.workflows {
		if we.matchesWorkflowFilter(workflow, filter) {
			workflows = append(workflows, workflow)
		}
	}
	we.mu.RUnlock()

	// Apply pagination
	if filter != nil {
		if filter.Offset > 0 && filter.Offset < len(workflows) {
			workflows = workflows[filter.Offset:]
		}
		if filter.Limit > 0 && filter.Limit < len(workflows) {
			workflows = workflows[:filter.Limit]
		}
	}

	span.SetAttributes(attribute.Int("workflows.count", len(workflows)))

	return workflows, nil
}

// Workflow execution

// ExecuteWorkflow executes a workflow with given input
func (we *DefaultWorkflowEngine) ExecuteWorkflow(ctx context.Context, workflowID string, input map[string]interface{}) (*WorkflowExecution, error) {
	ctx, span := we.tracer.Start(ctx, "workflow_engine.execute_workflow")
	defer span.End()

	span.SetAttributes(attribute.String("workflow.id", workflowID))

	// Get workflow
	workflow, err := we.GetWorkflow(ctx, workflowID)
	if err != nil {
		return nil, err
	}

	// Check if workflow is active
	if workflow.Status != WorkflowStatusActive {
		return nil, fmt.Errorf("workflow is not active: %s", workflow.Status)
	}

	// Create execution
	execution := &WorkflowExecution{
		ID:         uuid.New().String(),
		WorkflowID: workflowID,
		Status:     ExecutionStatusPending,
		StartedAt:  time.Now(),
		Steps:      []*ExecutionStep{},
		Variables:  input,
		Logs:       []string{},
		Metadata:   make(map[string]interface{}),
	}

	// Store execution
	we.mu.Lock()
	we.executions[execution.ID] = execution
	we.mu.Unlock()

	// Queue execution for processing
	select {
	case we.executionQueue <- execution:
		we.logger.WithFields(logrus.Fields{
			"execution_id": execution.ID,
			"workflow_id":  workflowID,
		}).Info("Workflow execution queued")
	default:
		execution.Status = ExecutionStatusFailure
		execution.Error = "execution queue is full"
		return execution, fmt.Errorf("execution queue is full")
	}

	return execution, nil
}

// GetExecution retrieves an execution by ID
func (we *DefaultWorkflowEngine) GetExecution(ctx context.Context, executionID string) (*WorkflowExecution, error) {
	we.mu.RLock()
	execution, exists := we.executions[executionID]
	we.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("execution not found: %s", executionID)
	}

	return execution, nil
}

// ListExecutions lists executions with filtering
func (we *DefaultWorkflowEngine) ListExecutions(ctx context.Context, filter *ExecutionFilter) ([]*WorkflowExecution, error) {
	we.mu.RLock()
	var executions []*WorkflowExecution
	for _, execution := range we.executions {
		if we.matchesExecutionFilter(execution, filter) {
			executions = append(executions, execution)
		}
	}
	we.mu.RUnlock()

	// Apply pagination
	if filter != nil {
		if filter.Offset > 0 && filter.Offset < len(executions) {
			executions = executions[filter.Offset:]
		}
		if filter.Limit > 0 && filter.Limit < len(executions) {
			executions = executions[:filter.Limit]
		}
	}

	return executions, nil
}

// CancelExecution cancels a running execution
func (we *DefaultWorkflowEngine) CancelExecution(ctx context.Context, executionID string) error {
	we.mu.Lock()
	execution, exists := we.executions[executionID]
	if !exists {
		we.mu.Unlock()
		return fmt.Errorf("execution not found: %s", executionID)
	}

	if execution.Status == ExecutionStatusRunning {
		execution.Status = ExecutionStatusCancelled
		now := time.Now()
		execution.CompletedAt = &now
		execution.Duration = now.Sub(execution.StartedAt)
		execution.Error = "execution cancelled by user"
	}
	we.mu.Unlock()

	we.logger.WithField("execution_id", executionID).Info("Execution cancelled")
	return nil
}

// RetryExecution retries a failed execution
func (we *DefaultWorkflowEngine) RetryExecution(ctx context.Context, executionID string) (*WorkflowExecution, error) {
	we.mu.RLock()
	originalExecution, exists := we.executions[executionID]
	we.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("execution not found: %s", executionID)
	}

	// Create new execution based on original
	return we.ExecuteWorkflow(ctx, originalExecution.WorkflowID, originalExecution.Variables)
}

// Event handling

// TriggerWorkflow triggers workflows based on an event
func (we *DefaultWorkflowEngine) TriggerWorkflow(ctx context.Context, event *Event) ([]*WorkflowExecution, error) {
	ctx, span := we.tracer.Start(ctx, "workflow_engine.trigger_workflow")
	defer span.End()

	span.SetAttributes(
		attribute.String("event.type", event.Type),
		attribute.String("event.source", event.Source),
	)

	// Find workflows that match this event
	workflows, err := we.triggerManager.ProcessEvent(ctx, event)
	if err != nil {
		return nil, fmt.Errorf("failed to process event: %w", err)
	}

	var executions []*WorkflowExecution
	for _, workflow := range workflows {
		// Create execution with event data
		execution, err := we.ExecuteWorkflow(ctx, workflow.ID, map[string]interface{}{
			"event": event,
		})
		if err != nil {
			we.logger.WithError(err).WithField("workflow_id", workflow.ID).Warn("Failed to execute triggered workflow")
			continue
		}
		executions = append(executions, execution)
	}

	we.logger.WithFields(logrus.Fields{
		"event_type":          event.Type,
		"triggered_workflows": len(workflows),
		"executions_started":  len(executions),
	}).Info("Event processed and workflows triggered")

	return executions, nil
}

// RegisterTrigger registers a trigger for a workflow
func (we *DefaultWorkflowEngine) RegisterTrigger(ctx context.Context, workflowID string, trigger *Trigger) error {
	return we.triggerManager.RegisterTrigger(ctx, workflowID, trigger)
}

// UnregisterTrigger unregisters a trigger
func (we *DefaultWorkflowEngine) UnregisterTrigger(ctx context.Context, workflowID string, triggerID string) error {
	return we.triggerManager.UnregisterTrigger(ctx, triggerID)
}

// Template management

// CreateTemplate creates a new workflow template
func (we *DefaultWorkflowEngine) CreateTemplate(ctx context.Context, template *WorkflowTemplate) (*WorkflowTemplate, error) {
	ctx, span := we.tracer.Start(ctx, "workflow_engine.create_template")
	defer span.End()

	// Generate ID if not provided
	if template.ID == "" {
		template.ID = uuid.New().String()
	}

	// Set timestamps
	now := time.Now()
	template.CreatedAt = now
	template.UpdatedAt = now

	// Store template
	we.mu.Lock()
	we.templates[template.ID] = template
	we.mu.Unlock()

	we.logger.WithFields(logrus.Fields{
		"template_id":   template.ID,
		"template_name": template.Name,
		"created_by":    template.CreatedBy,
	}).Info("Workflow template created")

	return template, nil
}

// GetTemplate retrieves a template by ID
func (we *DefaultWorkflowEngine) GetTemplate(ctx context.Context, templateID string) (*WorkflowTemplate, error) {
	we.mu.RLock()
	template, exists := we.templates[templateID]
	we.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("template not found: %s", templateID)
	}

	return template, nil
}

// ListTemplates lists templates with filtering
func (we *DefaultWorkflowEngine) ListTemplates(ctx context.Context, filter *TemplateFilter) ([]*WorkflowTemplate, error) {
	we.mu.RLock()
	var templates []*WorkflowTemplate
	for _, template := range we.templates {
		if we.matchesTemplateFilter(template, filter) {
			templates = append(templates, template)
		}
	}
	we.mu.RUnlock()

	// Apply pagination
	if filter != nil {
		if filter.Offset > 0 && filter.Offset < len(templates) {
			templates = templates[filter.Offset:]
		}
		if filter.Limit > 0 && filter.Limit < len(templates) {
			templates = templates[:filter.Limit]
		}
	}

	return templates, nil
}

// CreateWorkflowFromTemplate creates a workflow from a template
func (we *DefaultWorkflowEngine) CreateWorkflowFromTemplate(ctx context.Context, templateID string, workflowData *Workflow) (*Workflow, error) {
	template, err := we.GetTemplate(ctx, templateID)
	if err != nil {
		return nil, err
	}

	// Create workflow from template
	workflow := &Workflow{
		ID:          uuid.New().String(),
		Name:        workflowData.Name,
		Description: workflowData.Description,
		Version:     template.Workflow.Version,
		Status:      WorkflowStatusDraft,
		Triggers:    template.Workflow.Triggers,
		Actions:     template.Workflow.Actions,
		Conditions:  template.Workflow.Conditions,
		Variables:   template.Workflow.Variables,
		Schedule:    template.Workflow.Schedule,
		Timeout:     template.Workflow.Timeout,
		RetryPolicy: template.Workflow.RetryPolicy,
		Tags:        append(template.Tags, workflowData.Tags...),
		CreatedBy:   workflowData.CreatedBy,
		UpdatedBy:   workflowData.CreatedBy,
		Metadata:    workflowData.Metadata,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Increment template usage count
	template.UsageCount++

	return we.CreateWorkflow(ctx, workflow)
}

// Worker management

// startWorkers starts the workflow execution workers
func (we *DefaultWorkflowEngine) startWorkers() {
	for i := 0; i < we.workers; i++ {
		we.wg.Add(1)
		go we.worker(i)
	}

	we.logger.WithField("workers", we.workers).Info("Workflow workers started")
}

// worker processes workflow executions
func (we *DefaultWorkflowEngine) worker(id int) {
	defer we.wg.Done()

	we.logger.WithField("worker_id", id).Info("Workflow worker started")

	for {
		select {
		case execution := <-we.executionQueue:
			we.processExecution(execution)
		case <-we.stopChan:
			we.logger.WithField("worker_id", id).Info("Workflow worker stopping")
			return
		}
	}
}

// processExecution processes a single workflow execution
func (we *DefaultWorkflowEngine) processExecution(execution *WorkflowExecution) {
	ctx := context.Background()
	ctx, span := we.tracer.Start(ctx, "workflow_engine.process_execution")
	defer span.End()

	span.SetAttributes(
		attribute.String("execution.id", execution.ID),
		attribute.String("workflow.id", execution.WorkflowID),
	)

	// Update execution status
	execution.Status = ExecutionStatusRunning
	execution.Logs = append(execution.Logs, fmt.Sprintf("Execution started at %s", time.Now().Format(time.RFC3339)))

	// Get workflow
	workflow, err := we.GetWorkflow(ctx, execution.WorkflowID)
	if err != nil {
		we.failExecution(execution, fmt.Sprintf("Failed to get workflow: %v", err))
		return
	}

	// Execute workflow actions
	for i, action := range workflow.Actions {
		if err := we.executeAction(ctx, execution, action, i); err != nil {
			we.failExecution(execution, fmt.Sprintf("Action '%s' failed: %v", action.Name, err))
			return
		}
	}

	// Complete execution
	we.completeExecution(execution)
}

// executeAction executes a single action
func (we *DefaultWorkflowEngine) executeAction(ctx context.Context, execution *WorkflowExecution, action *Action, stepIndex int) error {
	// Check if action is enabled
	if !action.Enabled {
		execution.Logs = append(execution.Logs, fmt.Sprintf("Action '%s' skipped (disabled)", action.Name))
		return nil
	}

	// Check action conditions
	if !we.evaluateConditions(action.Conditions, execution.Variables) {
		execution.Logs = append(execution.Logs, fmt.Sprintf("Action '%s' skipped (conditions not met)", action.Name))
		return nil
	}

	// Create execution step
	step := &ExecutionStep{
		ID:         uuid.New().String(),
		ActionID:   action.ID,
		ActionType: string(action.Type),
		Status:     StepStatusRunning,
		StartedAt:  time.Now(),
		Input:      execution.Variables,
		Output:     make(map[string]interface{}),
		Metadata:   make(map[string]interface{}),
	}

	execution.Steps = append(execution.Steps, step)
	execution.Logs = append(execution.Logs, fmt.Sprintf("Executing action '%s' (%s)", action.Name, action.Type))

	// Execute action with timeout
	actionCtx := ctx
	if action.Timeout > 0 {
		var cancel context.CancelFunc
		actionCtx, cancel = context.WithTimeout(ctx, action.Timeout)
		defer cancel()
	}

	// Execute action with retry policy
	var output map[string]interface{}
	var err error
	maxRetries := 1
	if action.RetryPolicy != nil {
		maxRetries = action.RetryPolicy.MaxRetries + 1
	}

	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			step.RetryCount = attempt
			delay := time.Duration(attempt) * time.Second
			if action.RetryPolicy != nil {
				delay = action.RetryPolicy.RetryDelay
				if action.RetryPolicy.BackoffFactor > 1 {
					delay = time.Duration(float64(delay) * action.RetryPolicy.BackoffFactor)
				}
				if action.RetryPolicy.MaxDelay > 0 && delay > action.RetryPolicy.MaxDelay {
					delay = action.RetryPolicy.MaxDelay
				}
			}
			time.Sleep(delay)
			execution.Logs = append(execution.Logs, fmt.Sprintf("Retrying action '%s' (attempt %d)", action.Name, attempt+1))
		}

		output, err = we.actionExecutor.ExecuteAction(actionCtx, action, execution.Variables)
		if err == nil {
			break
		}

		execution.Logs = append(execution.Logs, fmt.Sprintf("Action '%s' attempt %d failed: %v", action.Name, attempt+1, err))
	}

	// Update step
	now := time.Now()
	step.CompletedAt = &now
	step.Duration = now.Sub(step.StartedAt)
	step.Output = output

	if err != nil {
		step.Status = StepStatusFailure
		step.Error = err.Error()
		return err
	}

	step.Status = StepStatusSuccess

	// Merge output into execution variables
	if output != nil {
		for key, value := range output {
			execution.Variables[key] = value
		}
	}

	execution.Logs = append(execution.Logs, fmt.Sprintf("Action '%s' completed successfully", action.Name))
	return nil
}

// completeExecution marks an execution as completed
func (we *DefaultWorkflowEngine) completeExecution(execution *WorkflowExecution) {
	now := time.Now()
	execution.Status = ExecutionStatusSuccess
	execution.CompletedAt = &now
	execution.Duration = now.Sub(execution.StartedAt)
	execution.Logs = append(execution.Logs, fmt.Sprintf("Execution completed at %s", now.Format(time.RFC3339)))

	we.logger.WithFields(logrus.Fields{
		"execution_id": execution.ID,
		"workflow_id":  execution.WorkflowID,
		"duration":     execution.Duration,
		"steps":        len(execution.Steps),
	}).Info("Workflow execution completed successfully")
}

// failExecution marks an execution as failed
func (we *DefaultWorkflowEngine) failExecution(execution *WorkflowExecution, errorMsg string) {
	now := time.Now()
	execution.Status = ExecutionStatusFailure
	execution.CompletedAt = &now
	execution.Duration = now.Sub(execution.StartedAt)
	execution.Error = errorMsg
	execution.Logs = append(execution.Logs, fmt.Sprintf("Execution failed at %s: %s", now.Format(time.RFC3339), errorMsg))

	we.logger.WithFields(logrus.Fields{
		"execution_id": execution.ID,
		"workflow_id":  execution.WorkflowID,
		"error":        errorMsg,
		"duration":     execution.Duration,
	}).Error("Workflow execution failed")
}
