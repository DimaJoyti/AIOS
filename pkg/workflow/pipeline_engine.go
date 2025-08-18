package workflow

import (
	"context"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// DefaultPipelineEngine implements the PipelineEngine interface
type DefaultPipelineEngine struct {
	logger         *logrus.Logger
	tracer         trace.Tracer
	workflowEngine WorkflowEngine

	// In-memory storage (in production, this would be a database)
	pipelines  map[string]*Pipeline
	executions map[string]*PipelineExecution
	artifacts  map[string]*Artifact

	// Execution management
	executionQueue chan *PipelineExecution
	workers        int
	stopChan       chan struct{}
	wg             sync.WaitGroup
	mu             sync.RWMutex
}

// PipelineEngineConfig represents configuration for the pipeline engine
type PipelineEngineConfig struct {
	Workers        int           `json:"workers"`
	QueueSize      int           `json:"queue_size"`
	DefaultTimeout time.Duration `json:"default_timeout"`
	MaxRetries     int           `json:"max_retries"`
	ArtifactStore  string        `json:"artifact_store"`
}

// NewDefaultPipelineEngine creates a new pipeline engine
func NewDefaultPipelineEngine(config *PipelineEngineConfig, workflowEngine WorkflowEngine, logger *logrus.Logger) (PipelineEngine, error) {
	if config == nil {
		config = &PipelineEngineConfig{
			Workers:        3,
			QueueSize:      500,
			DefaultTimeout: 60 * time.Minute,
			MaxRetries:     2,
			ArtifactStore:  "/tmp/artifacts",
		}
	}

	engine := &DefaultPipelineEngine{
		logger:         logger,
		tracer:         otel.Tracer("pipeline.engine"),
		workflowEngine: workflowEngine,
		pipelines:      make(map[string]*Pipeline),
		executions:     make(map[string]*PipelineExecution),
		artifacts:      make(map[string]*Artifact),
		executionQueue: make(chan *PipelineExecution, config.QueueSize),
		workers:        config.Workers,
		stopChan:       make(chan struct{}),
	}

	// Start worker goroutines
	engine.startWorkers()

	return engine, nil
}

// Pipeline management

// CreatePipeline creates a new pipeline
func (pe *DefaultPipelineEngine) CreatePipeline(ctx context.Context, pipeline *Pipeline) (*Pipeline, error) {
	ctx, span := pe.tracer.Start(ctx, "pipeline_engine.create_pipeline")
	defer span.End()

	span.SetAttributes(
		attribute.String("pipeline.name", pipeline.Name),
		attribute.String("pipeline.created_by", pipeline.CreatedBy),
	)

	// Generate ID if not provided
	if pipeline.ID == "" {
		pipeline.ID = uuid.New().String()
	}

	// Set timestamps
	now := time.Now()
	pipeline.CreatedAt = now
	pipeline.UpdatedAt = now

	// Set default values
	if pipeline.Status == "" {
		pipeline.Status = PipelineStatusActive
	}
	if pipeline.Timeout == 0 {
		pipeline.Timeout = 60 * time.Minute
	}

	// Validate pipeline
	if err := pe.validatePipeline(pipeline); err != nil {
		return nil, fmt.Errorf("pipeline validation failed: %w", err)
	}

	// Store pipeline
	pe.mu.Lock()
	pe.pipelines[pipeline.ID] = pipeline
	pe.mu.Unlock()

	pe.logger.WithFields(logrus.Fields{
		"pipeline_id":   pipeline.ID,
		"pipeline_name": pipeline.Name,
		"created_by":    pipeline.CreatedBy,
		"stages":        len(pipeline.Stages),
	}).Info("Pipeline created successfully")

	return pipeline, nil
}

// GetPipeline retrieves a pipeline by ID
func (pe *DefaultPipelineEngine) GetPipeline(ctx context.Context, pipelineID string) (*Pipeline, error) {
	ctx, span := pe.tracer.Start(ctx, "pipeline_engine.get_pipeline")
	defer span.End()

	span.SetAttributes(attribute.String("pipeline.id", pipelineID))

	pe.mu.RLock()
	pipeline, exists := pe.pipelines[pipelineID]
	pe.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("pipeline not found: %s", pipelineID)
	}

	return pipeline, nil
}

// UpdatePipeline updates an existing pipeline
func (pe *DefaultPipelineEngine) UpdatePipeline(ctx context.Context, pipeline *Pipeline) (*Pipeline, error) {
	ctx, span := pe.tracer.Start(ctx, "pipeline_engine.update_pipeline")
	defer span.End()

	span.SetAttributes(attribute.String("pipeline.id", pipeline.ID))

	pe.mu.Lock()
	existing, exists := pe.pipelines[pipeline.ID]
	if !exists {
		pe.mu.Unlock()
		return nil, fmt.Errorf("pipeline not found: %s", pipeline.ID)
	}

	// Update timestamp and preserve creation info
	pipeline.UpdatedAt = time.Now()
	pipeline.CreatedAt = existing.CreatedAt
	pipeline.CreatedBy = existing.CreatedBy

	// Validate updated pipeline
	if err := pe.validatePipeline(pipeline); err != nil {
		pe.mu.Unlock()
		return nil, fmt.Errorf("pipeline validation failed: %w", err)
	}

	// Store updated pipeline
	pe.pipelines[pipeline.ID] = pipeline
	pe.mu.Unlock()

	pe.logger.WithFields(logrus.Fields{
		"pipeline_id":   pipeline.ID,
		"pipeline_name": pipeline.Name,
		"updated_by":    pipeline.UpdatedBy,
	}).Info("Pipeline updated successfully")

	return pipeline, nil
}

// DeletePipeline deletes a pipeline
func (pe *DefaultPipelineEngine) DeletePipeline(ctx context.Context, pipelineID string) error {
	ctx, span := pe.tracer.Start(ctx, "pipeline_engine.delete_pipeline")
	defer span.End()

	span.SetAttributes(attribute.String("pipeline.id", pipelineID))

	pe.mu.Lock()
	pipeline, exists := pe.pipelines[pipelineID]
	if !exists {
		pe.mu.Unlock()
		return fmt.Errorf("pipeline not found: %s", pipelineID)
	}

	// Delete associated executions and artifacts
	for executionID, execution := range pe.executions {
		if execution.PipelineID == pipelineID {
			delete(pe.executions, executionID)
		}
	}

	// Delete pipeline
	delete(pe.pipelines, pipelineID)
	pe.mu.Unlock()

	pe.logger.WithFields(logrus.Fields{
		"pipeline_id":   pipelineID,
		"pipeline_name": pipeline.Name,
	}).Info("Pipeline deleted successfully")

	return nil
}

// ListPipelines lists pipelines with filtering
func (pe *DefaultPipelineEngine) ListPipelines(ctx context.Context, filter *PipelineFilter) ([]*Pipeline, error) {
	ctx, span := pe.tracer.Start(ctx, "pipeline_engine.list_pipelines")
	defer span.End()

	pe.mu.RLock()
	var pipelines []*Pipeline
	for _, pipeline := range pe.pipelines {
		if pe.matchesPipelineFilter(pipeline, filter) {
			pipelines = append(pipelines, pipeline)
		}
	}
	pe.mu.RUnlock()

	// Apply pagination
	if filter != nil {
		if filter.Offset > 0 && filter.Offset < len(pipelines) {
			pipelines = pipelines[filter.Offset:]
		}
		if filter.Limit > 0 && filter.Limit < len(pipelines) {
			pipelines = pipelines[:filter.Limit]
		}
	}

	span.SetAttributes(attribute.Int("pipelines.count", len(pipelines)))

	return pipelines, nil
}

// Pipeline execution

// ExecutePipeline executes a pipeline with given parameters
func (pe *DefaultPipelineEngine) ExecutePipeline(ctx context.Context, pipelineID string, params *ExecutionParams) (*PipelineExecution, error) {
	ctx, span := pe.tracer.Start(ctx, "pipeline_engine.execute_pipeline")
	defer span.End()

	span.SetAttributes(attribute.String("pipeline.id", pipelineID))

	// Get pipeline
	pipeline, err := pe.GetPipeline(ctx, pipelineID)
	if err != nil {
		return nil, err
	}

	// Check if pipeline is active
	if pipeline.Status != PipelineStatusActive {
		return nil, fmt.Errorf("pipeline is not active: %s", pipeline.Status)
	}

	// Create execution
	execution := &PipelineExecution{
		ID:          uuid.New().String(),
		PipelineID:  pipelineID,
		Status:      ExecutionStatusPending,
		Branch:      params.Branch,
		Commit:      params.Commit,
		Author:      "system",
		TriggerType: params.TriggerType,
		StartedAt:   time.Now(),
		Stages:      []*StageExecution{},
		Variables:   params.Variables,
		Artifacts:   []*Artifact{},
		Logs:        []string{},
		Metadata:    make(map[string]interface{}),
	}

	if params.TriggerType == "" {
		execution.TriggerType = "manual"
	}

	// Store execution
	pe.mu.Lock()
	pe.executions[execution.ID] = execution
	pe.mu.Unlock()

	// Queue execution for processing
	select {
	case pe.executionQueue <- execution:
		pe.logger.WithFields(logrus.Fields{
			"execution_id": execution.ID,
			"pipeline_id":  pipelineID,
			"branch":       execution.Branch,
			"trigger_type": execution.TriggerType,
		}).Info("Pipeline execution queued")
	default:
		execution.Status = ExecutionStatusFailure
		execution.Error = "execution queue is full"
		return execution, fmt.Errorf("execution queue is full")
	}

	return execution, nil
}

// GetPipelineExecution retrieves an execution by ID
func (pe *DefaultPipelineEngine) GetPipelineExecution(ctx context.Context, executionID string) (*PipelineExecution, error) {
	pe.mu.RLock()
	execution, exists := pe.executions[executionID]
	pe.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("execution not found: %s", executionID)
	}

	return execution, nil
}

// ListPipelineExecutions lists executions with filtering
func (pe *DefaultPipelineEngine) ListPipelineExecutions(ctx context.Context, filter *PipelineExecutionFilter) ([]*PipelineExecution, error) {
	pe.mu.RLock()
	var executions []*PipelineExecution
	for _, execution := range pe.executions {
		if pe.matchesPipelineExecutionFilter(execution, filter) {
			executions = append(executions, execution)
		}
	}
	pe.mu.RUnlock()

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

// CancelPipelineExecution cancels a running execution
func (pe *DefaultPipelineEngine) CancelPipelineExecution(ctx context.Context, executionID string) error {
	pe.mu.Lock()
	execution, exists := pe.executions[executionID]
	if !exists {
		pe.mu.Unlock()
		return fmt.Errorf("execution not found: %s", executionID)
	}

	if execution.Status == ExecutionStatusRunning {
		execution.Status = ExecutionStatusCancelled
		now := time.Now()
		execution.CompletedAt = &now
		execution.Duration = now.Sub(execution.StartedAt)
		execution.Error = "execution cancelled by user"
	}
	pe.mu.Unlock()

	pe.logger.WithField("execution_id", executionID).Info("Pipeline execution cancelled")
	return nil
}

// RetryPipelineExecution retries a failed execution
func (pe *DefaultPipelineEngine) RetryPipelineExecution(ctx context.Context, executionID string) (*PipelineExecution, error) {
	pe.mu.RLock()
	originalExecution, exists := pe.executions[executionID]
	pe.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("execution not found: %s", executionID)
	}

	// Create new execution based on original
	params := &ExecutionParams{
		Branch:      originalExecution.Branch,
		Commit:      originalExecution.Commit,
		Variables:   originalExecution.Variables,
		TriggerType: "retry",
	}

	return pe.ExecutePipeline(ctx, originalExecution.PipelineID, params)
}

// Worker management

// startWorkers starts the pipeline execution workers
func (pe *DefaultPipelineEngine) startWorkers() {
	for i := 0; i < pe.workers; i++ {
		pe.wg.Add(1)
		go pe.worker(i)
	}

	pe.logger.WithField("workers", pe.workers).Info("Pipeline workers started")
}

// worker processes pipeline executions
func (pe *DefaultPipelineEngine) worker(id int) {
	defer pe.wg.Done()

	pe.logger.WithField("worker_id", id).Info("Pipeline worker started")

	for {
		select {
		case execution := <-pe.executionQueue:
			pe.processExecution(execution)
		case <-pe.stopChan:
			pe.logger.WithField("worker_id", id).Info("Pipeline worker stopping")
			return
		}
	}
}

// processExecution processes a single pipeline execution
func (pe *DefaultPipelineEngine) processExecution(execution *PipelineExecution) {
	ctx := context.Background()
	ctx, span := pe.tracer.Start(ctx, "pipeline_engine.process_execution")
	defer span.End()

	span.SetAttributes(
		attribute.String("execution.id", execution.ID),
		attribute.String("pipeline.id", execution.PipelineID),
	)

	// Update execution status
	execution.Status = ExecutionStatusRunning
	execution.Logs = append(execution.Logs, fmt.Sprintf("Pipeline execution started at %s", time.Now().Format(time.RFC3339)))

	// Get pipeline
	pipeline, err := pe.GetPipeline(ctx, execution.PipelineID)
	if err != nil {
		pe.failExecution(execution, fmt.Sprintf("Failed to get pipeline: %v", err))
		return
	}

	// Execute pipeline stages
	for i, stage := range pipeline.Stages {
		if err := pe.executeStage(ctx, execution, stage, i); err != nil {
			pe.failExecution(execution, fmt.Sprintf("Stage '%s' failed: %v", stage.Name, err))
			return
		}
	}

	// Complete execution
	pe.completeExecution(execution)
}

// executeStage executes a single stage
func (pe *DefaultPipelineEngine) executeStage(ctx context.Context, execution *PipelineExecution, stage *Stage, stageIndex int) error {
	// Check if stage is enabled
	if !stage.Enabled {
		execution.Logs = append(execution.Logs, fmt.Sprintf("Stage '%s' skipped (disabled)", stage.Name))
		return nil
	}

	// Check stage dependencies
	if !pe.checkStageDependencies(execution, stage) {
		execution.Logs = append(execution.Logs, fmt.Sprintf("Stage '%s' skipped (dependencies not met)", stage.Name))
		return nil
	}

	// Create stage execution
	stageExecution := &StageExecution{
		ID:        uuid.New().String(),
		StageID:   stage.ID,
		Name:      stage.Name,
		Status:    ExecutionStatusRunning,
		StartedAt: time.Now(),
		Jobs:      []*JobExecution{},
		Variables: execution.Variables,
		Logs:      []string{},
		Metadata:  make(map[string]interface{}),
	}

	execution.Stages = append(execution.Stages, stageExecution)
	execution.Logs = append(execution.Logs, fmt.Sprintf("Executing stage '%s'", stage.Name))

	// Execute stage jobs
	for _, job := range stage.Jobs {
		if err := pe.executeJob(ctx, execution, stageExecution, job); err != nil {
			stageExecution.Status = ExecutionStatusFailure
			stageExecution.Error = err.Error()
			now := time.Now()
			stageExecution.CompletedAt = &now
			stageExecution.Duration = now.Sub(stageExecution.StartedAt)
			return err
		}
	}

	// Complete stage
	now := time.Now()
	stageExecution.Status = ExecutionStatusSuccess
	stageExecution.CompletedAt = &now
	stageExecution.Duration = now.Sub(stageExecution.StartedAt)

	execution.Logs = append(execution.Logs, fmt.Sprintf("Stage '%s' completed successfully", stage.Name))
	return nil
}

// executeJob executes a single job
func (pe *DefaultPipelineEngine) executeJob(ctx context.Context, execution *PipelineExecution, stageExecution *StageExecution, job *Job) error {
	// Check if job is enabled
	if !job.Enabled {
		stageExecution.Logs = append(stageExecution.Logs, fmt.Sprintf("Job '%s' skipped (disabled)", job.Name))
		return nil
	}

	// Create job execution
	jobExecution := &JobExecution{
		ID:        uuid.New().String(),
		JobID:     job.ID,
		Name:      job.Name,
		Status:    ExecutionStatusRunning,
		StartedAt: time.Now(),
		Steps:     []*StepExecution{},
		Variables: stageExecution.Variables,
		Artifacts: []*Artifact{},
		Logs:      []string{},
		Metadata:  make(map[string]interface{}),
	}

	stageExecution.Jobs = append(stageExecution.Jobs, jobExecution)
	stageExecution.Logs = append(stageExecution.Logs, fmt.Sprintf("Executing job '%s'", job.Name))

	// Execute job steps
	for _, step := range job.Steps {
		if err := pe.executeStep(ctx, execution, jobExecution, step); err != nil {
			jobExecution.Status = ExecutionStatusFailure
			jobExecution.Error = err.Error()
			now := time.Now()
			jobExecution.CompletedAt = &now
			jobExecution.Duration = now.Sub(jobExecution.StartedAt)
			return err
		}
	}

	// Complete job
	now := time.Now()
	jobExecution.Status = ExecutionStatusSuccess
	jobExecution.CompletedAt = &now
	jobExecution.Duration = now.Sub(jobExecution.StartedAt)

	stageExecution.Logs = append(stageExecution.Logs, fmt.Sprintf("Job '%s' completed successfully", job.Name))
	return nil
}

// executeStep executes a single step
func (pe *DefaultPipelineEngine) executeStep(ctx context.Context, execution *PipelineExecution, jobExecution *JobExecution, step *Step) error {
	// Check if step is enabled
	if !step.Enabled {
		jobExecution.Logs = append(jobExecution.Logs, fmt.Sprintf("Step '%s' skipped (disabled)", step.Name))
		return nil
	}

	// Create step execution
	stepExecution := &StepExecution{
		ID:        uuid.New().String(),
		StepID:    step.ID,
		Name:      step.Name,
		Status:    ExecutionStatusRunning,
		StartedAt: time.Now(),
		Command:   step.Command,
		Logs:      []string{},
		Metadata:  make(map[string]interface{}),
	}

	jobExecution.Steps = append(jobExecution.Steps, stepExecution)
	jobExecution.Logs = append(jobExecution.Logs, fmt.Sprintf("Executing step '%s'", step.Name))

	// Execute step based on type
	var err error
	switch step.Type {
	case StepTypeCommand:
		err = pe.executeCommandStep(ctx, stepExecution, step)
	case StepTypeScript:
		err = pe.executeScriptStep(ctx, stepExecution, step)
	case StepTypeAction:
		err = pe.executeActionStep(ctx, stepExecution, step)
	default:
		err = fmt.Errorf("unsupported step type: %s", step.Type)
	}

	// Update step execution
	now := time.Now()
	stepExecution.CompletedAt = &now
	stepExecution.Duration = now.Sub(stepExecution.StartedAt)

	if err != nil {
		stepExecution.Status = ExecutionStatusFailure
		stepExecution.Error = err.Error()
		stepExecution.ExitCode = 1

		if !step.ContinueOnError {
			return err
		}
	} else {
		stepExecution.Status = ExecutionStatusSuccess
		stepExecution.ExitCode = 0
	}

	jobExecution.Logs = append(jobExecution.Logs, fmt.Sprintf("Step '%s' completed with status %s", step.Name, stepExecution.Status))
	return nil
}

// Artifact management

// UploadArtifact uploads an artifact for an execution
func (pe *DefaultPipelineEngine) UploadArtifact(ctx context.Context, executionID string, artifact *Artifact, reader io.Reader) error {
	ctx, span := pe.tracer.Start(ctx, "pipeline_engine.upload_artifact")
	defer span.End()

	span.SetAttributes(
		attribute.String("execution.id", executionID),
		attribute.String("artifact.name", artifact.Name),
	)

	// Check if execution exists
	pe.mu.RLock()
	execution, exists := pe.executions[executionID]
	pe.mu.RUnlock()

	if !exists {
		return fmt.Errorf("execution not found: %s", executionID)
	}

	// Generate artifact ID if not provided
	if artifact.ID == "" {
		artifact.ID = uuid.New().String()
	}

	// Set artifact metadata
	artifact.CreatedAt = time.Now()
	if artifact.Path == "" {
		artifact.Path = fmt.Sprintf("/artifacts/%s/%s", executionID, artifact.Name)
	}

	// In a real implementation, this would save to actual storage
	// For now, we'll simulate by reading the content and storing metadata
	content, err := io.ReadAll(reader)
	if err != nil {
		return fmt.Errorf("failed to read artifact content: %w", err)
	}

	artifact.Size = int64(len(content))
	artifact.Checksum = fmt.Sprintf("sha256:%x", content) // Simplified checksum

	// Store artifact
	pe.mu.Lock()
	pe.artifacts[artifact.ID] = artifact
	execution.Artifacts = append(execution.Artifacts, artifact)
	pe.mu.Unlock()

	pe.logger.WithFields(map[string]interface{}{
		"artifact_id":   artifact.ID,
		"artifact_name": artifact.Name,
		"execution_id":  executionID,
		"size":          artifact.Size,
	}).Info("Artifact uploaded successfully")

	return nil
}

// DownloadArtifact downloads an artifact
func (pe *DefaultPipelineEngine) DownloadArtifact(ctx context.Context, artifactID string, writer io.Writer) error {
	ctx, span := pe.tracer.Start(ctx, "pipeline_engine.download_artifact")
	defer span.End()

	span.SetAttributes(attribute.String("artifact.id", artifactID))

	pe.mu.RLock()
	artifact, exists := pe.artifacts[artifactID]
	pe.mu.RUnlock()

	if !exists {
		return fmt.Errorf("artifact not found: %s", artifactID)
	}

	// In a real implementation, this would read from actual storage
	// For now, we'll simulate by writing some content
	content := fmt.Sprintf("Artifact content for %s (size: %d bytes)", artifact.Name, artifact.Size)
	_, err := writer.Write([]byte(content))
	if err != nil {
		return fmt.Errorf("failed to write artifact content: %w", err)
	}

	pe.logger.WithFields(map[string]interface{}{
		"artifact_id":   artifactID,
		"artifact_name": artifact.Name,
	}).Info("Artifact downloaded successfully")

	return nil
}

// ListArtifacts lists artifacts for an execution
func (pe *DefaultPipelineEngine) ListArtifacts(ctx context.Context, executionID string) ([]*Artifact, error) {
	pe.mu.RLock()
	execution, exists := pe.executions[executionID]
	pe.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("execution not found: %s", executionID)
	}

	return execution.Artifacts, nil
}

// DeleteArtifact deletes an artifact
func (pe *DefaultPipelineEngine) DeleteArtifact(ctx context.Context, artifactID string) error {
	ctx, span := pe.tracer.Start(ctx, "pipeline_engine.delete_artifact")
	defer span.End()

	span.SetAttributes(attribute.String("artifact.id", artifactID))

	pe.mu.Lock()
	artifact, exists := pe.artifacts[artifactID]
	if !exists {
		pe.mu.Unlock()
		return fmt.Errorf("artifact not found: %s", artifactID)
	}

	// Remove from artifacts map
	delete(pe.artifacts, artifactID)

	// Remove from execution artifacts list
	for _, execution := range pe.executions {
		for i, execArtifact := range execution.Artifacts {
			if execArtifact.ID == artifactID {
				execution.Artifacts = append(execution.Artifacts[:i], execution.Artifacts[i+1:]...)
				break
			}
		}
	}
	pe.mu.Unlock()

	pe.logger.WithFields(map[string]interface{}{
		"artifact_id":   artifactID,
		"artifact_name": artifact.Name,
	}).Info("Artifact deleted successfully")

	return nil
}

// Pipeline triggers

// RegisterPipelineTrigger registers a trigger for a pipeline
func (pe *DefaultPipelineEngine) RegisterPipelineTrigger(ctx context.Context, pipelineID string, trigger *PipelineTrigger) error {
	ctx, span := pe.tracer.Start(ctx, "pipeline_engine.register_trigger")
	defer span.End()

	// Get pipeline
	pipeline, err := pe.GetPipeline(ctx, pipelineID)
	if err != nil {
		return err
	}

	// Generate trigger ID if not provided
	if trigger.ID == "" {
		trigger.ID = uuid.New().String()
	}

	// Add trigger to pipeline
	pe.mu.Lock()
	pipeline.Triggers = append(pipeline.Triggers, trigger)
	pe.mu.Unlock()

	pe.logger.WithFields(map[string]interface{}{
		"trigger_id":   trigger.ID,
		"pipeline_id":  pipelineID,
		"trigger_type": trigger.Type,
	}).Info("Pipeline trigger registered")

	return nil
}

// UnregisterPipelineTrigger unregisters a trigger
func (pe *DefaultPipelineEngine) UnregisterPipelineTrigger(ctx context.Context, pipelineID string, triggerID string) error {
	// Get pipeline
	pipeline, err := pe.GetPipeline(ctx, pipelineID)
	if err != nil {
		return err
	}

	// Remove trigger from pipeline
	pe.mu.Lock()
	for i, trigger := range pipeline.Triggers {
		if trigger.ID == triggerID {
			pipeline.Triggers = append(pipeline.Triggers[:i], pipeline.Triggers[i+1:]...)
			break
		}
	}
	pe.mu.Unlock()

	pe.logger.WithFields(map[string]interface{}{
		"trigger_id":  triggerID,
		"pipeline_id": pipelineID,
	}).Info("Pipeline trigger unregistered")

	return nil
}

// TriggerPipeline triggers pipelines based on an event
func (pe *DefaultPipelineEngine) TriggerPipeline(ctx context.Context, event *Event) ([]*PipelineExecution, error) {
	ctx, span := pe.tracer.Start(ctx, "pipeline_engine.trigger_pipeline")
	defer span.End()

	span.SetAttributes(
		attribute.String("event.type", event.Type),
		attribute.String("event.source", event.Source),
	)

	var executions []*PipelineExecution

	// Find pipelines that match this event
	pe.mu.RLock()
	for _, pipeline := range pe.pipelines {
		if pe.eventMatchesPipeline(event, pipeline) {
			// Create execution parameters from event
			params := &ExecutionParams{
				TriggerType: "event",
				Variables:   event.Data,
			}

			// Extract common fields from event data
			if branch, ok := event.Data["branch"].(string); ok {
				params.Branch = branch
			}
			if commit, ok := event.Data["commit"].(string); ok {
				params.Commit = commit
			}

			pe.mu.RUnlock()
			execution, err := pe.ExecutePipeline(ctx, pipeline.ID, params)
			pe.mu.RLock()

			if err != nil {
				pe.logger.WithError(err).WithField("pipeline_id", pipeline.ID).Warn("Failed to execute triggered pipeline")
				continue
			}
			executions = append(executions, execution)
		}
	}
	pe.mu.RUnlock()

	pe.logger.WithFields(map[string]interface{}{
		"event_type":          event.Type,
		"triggered_pipelines": len(executions),
	}).Info("Event processed and pipelines triggered")

	return executions, nil
}

// eventMatchesPipeline checks if an event matches a pipeline's triggers
func (pe *DefaultPipelineEngine) eventMatchesPipeline(event *Event, pipeline *Pipeline) bool {
	for _, trigger := range pipeline.Triggers {
		if !trigger.Enabled {
			continue
		}

		// Check trigger type
		switch trigger.Type {
		case TriggerTypePush:
			if event.Type == "push" {
				return pe.checkTriggerConditions(trigger, event.Data)
			}
		case TriggerTypePR:
			if event.Type == "pull_request" {
				return pe.checkTriggerConditions(trigger, event.Data)
			}
		case TriggerTypeEvent:
			if len(trigger.Events) > 0 {
				for _, triggerEvent := range trigger.Events {
					if event.Type == triggerEvent {
						return pe.checkTriggerConditions(trigger, event.Data)
					}
				}
			}
		}
	}

	return false
}

// checkTriggerConditions checks if trigger conditions are met
func (pe *DefaultPipelineEngine) checkTriggerConditions(trigger *PipelineTrigger, eventData map[string]interface{}) bool {
	// Check branch filter
	if len(trigger.Branches) > 0 {
		if branch, ok := eventData["branch"].(string); ok {
			found := false
			for _, triggerBranch := range trigger.Branches {
				if branch == triggerBranch {
					found = true
					break
				}
			}
			if !found {
				return false
			}
		}
	}

	// Check path filter
	if len(trigger.Paths) > 0 {
		if paths, ok := eventData["paths"].([]string); ok {
			found := false
			for _, triggerPath := range trigger.Paths {
				for _, eventPath := range paths {
					if strings.HasPrefix(eventPath, triggerPath) {
						found = true
						break
					}
				}
				if found {
					break
				}
			}
			if !found {
				return false
			}
		}
	}

	return true
}
