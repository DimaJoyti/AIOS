package workflow

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Helper methods for DefaultPipelineEngine

// validatePipeline validates a pipeline definition
func (pe *DefaultPipelineEngine) validatePipeline(pipeline *Pipeline) error {
	if pipeline.Name == "" {
		return fmt.Errorf("pipeline name is required")
	}

	if len(pipeline.Stages) == 0 {
		return fmt.Errorf("pipeline must have at least one stage")
	}

	// Validate stages
	for i, stage := range pipeline.Stages {
		if err := pe.validateStage(stage); err != nil {
			return fmt.Errorf("stage %d validation failed: %w", i, err)
		}
	}

	return nil
}

// validateStage validates a stage definition
func (pe *DefaultPipelineEngine) validateStage(stage *Stage) error {
	if stage.ID == "" {
		stage.ID = uuid.New().String()
	}

	if stage.Name == "" {
		return fmt.Errorf("stage name is required")
	}

	if len(stage.Jobs) == 0 {
		return fmt.Errorf("stage must have at least one job")
	}

	// Validate jobs
	for i, job := range stage.Jobs {
		if err := pe.validateJob(job); err != nil {
			return fmt.Errorf("job %d validation failed: %w", i, err)
		}
	}

	return nil
}

// validateJob validates a job definition
func (pe *DefaultPipelineEngine) validateJob(job *Job) error {
	if job.ID == "" {
		job.ID = uuid.New().String()
	}

	if job.Name == "" {
		return fmt.Errorf("job name is required")
	}

	if len(job.Steps) == 0 {
		return fmt.Errorf("job must have at least one step")
	}

	// Validate steps
	for i, step := range job.Steps {
		if err := pe.validateStep(step); err != nil {
			return fmt.Errorf("step %d validation failed: %w", i, err)
		}
	}

	return nil
}

// validateStep validates a step definition
func (pe *DefaultPipelineEngine) validateStep(step *Step) error {
	if step.ID == "" {
		step.ID = uuid.New().String()
	}

	if step.Name == "" {
		return fmt.Errorf("step name is required")
	}

	if step.Type == "" {
		return fmt.Errorf("step type is required")
	}

	// Validate step type
	validTypes := []StepType{
		StepTypeCommand, StepTypeScript, StepTypeAction,
		StepTypeBuild, StepTypeTest, StepTypeDeploy,
	}

	found := false
	for _, validType := range validTypes {
		if step.Type == validType {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("invalid step type: %s", step.Type)
	}

	return nil
}

// Filter matching methods

// matchesPipelineFilter checks if a pipeline matches the given filter
func (pe *DefaultPipelineEngine) matchesPipelineFilter(pipeline *Pipeline, filter *PipelineFilter) bool {
	if filter == nil {
		return true
	}

	// Status filter
	if len(filter.Status) > 0 {
		found := false
		for _, status := range filter.Status {
			if pipeline.Status == status {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Created by filter
	if filter.CreatedBy != "" && pipeline.CreatedBy != filter.CreatedBy {
		return false
	}

	// Tags filter
	if len(filter.Tags) > 0 {
		for _, filterTag := range filter.Tags {
			found := false
			for _, pipelineTag := range pipeline.Tags {
				if pipelineTag == filterTag {
					found = true
					break
				}
			}
			if !found {
				return false
			}
		}
	}

	// Search filter
	if filter.Search != "" {
		searchLower := strings.ToLower(filter.Search)
		if !(strings.Contains(strings.ToLower(pipeline.Name), searchLower) ||
			strings.Contains(strings.ToLower(pipeline.Description), searchLower)) {
			return false
		}
	}

	return true
}

// matchesPipelineExecutionFilter checks if an execution matches the given filter
func (pe *DefaultPipelineEngine) matchesPipelineExecutionFilter(execution *PipelineExecution, filter *PipelineExecutionFilter) bool {
	if filter == nil {
		return true
	}

	// Pipeline ID filter
	if filter.PipelineID != "" && execution.PipelineID != filter.PipelineID {
		return false
	}

	// Status filter
	if len(filter.Status) > 0 {
		found := false
		for _, status := range filter.Status {
			if execution.Status == status {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Branch filter
	if filter.Branch != "" && execution.Branch != filter.Branch {
		return false
	}

	// Author filter
	if filter.Author != "" && execution.Author != filter.Author {
		return false
	}

	// Date filters
	if filter.StartedAfter != nil && execution.StartedAt.Before(*filter.StartedAfter) {
		return false
	}
	if filter.StartedBefore != nil && execution.StartedAt.After(*filter.StartedBefore) {
		return false
	}

	return true
}

// Dependency checking

// checkStageDependencies checks if stage dependencies are satisfied
func (pe *DefaultPipelineEngine) checkStageDependencies(execution *PipelineExecution, stage *Stage) bool {
	if len(stage.DependsOn) == 0 {
		return true // No dependencies
	}

	// Check if all dependent stages have completed successfully
	for _, dependentStageID := range stage.DependsOn {
		found := false
		for _, stageExecution := range execution.Stages {
			if stageExecution.StageID == dependentStageID {
				if stageExecution.Status != ExecutionStatusSuccess {
					return false
				}
				found = true
				break
			}
		}
		if !found {
			return false // Dependent stage not found or not executed
		}
	}

	return true
}

// Step execution methods

// executeCommandStep executes a command step
func (pe *DefaultPipelineEngine) executeCommandStep(ctx context.Context, stepExecution *StepExecution, step *Step) error {
	if step.Command == "" {
		return fmt.Errorf("command is required for command step")
	}

	cmd := exec.CommandContext(ctx, "sh", "-c", step.Command)

	// Set working directory
	if step.WorkingDir != "" {
		cmd.Dir = step.WorkingDir
	}

	// Set environment variables
	if len(step.Environment) > 0 {
		env := os.Environ()
		for key, value := range step.Environment {
			env = append(env, fmt.Sprintf("%s=%s", key, value))
		}
		cmd.Env = env
	}

	// Execute command
	output, err := cmd.CombinedOutput()
	stepExecution.Output = string(output)

	if err != nil {
		return fmt.Errorf("command failed: %w", err)
	}

	pe.logger.WithField("command", step.Command).Info("Command step executed successfully")
	return nil
}

// executeScriptStep executes a script step
func (pe *DefaultPipelineEngine) executeScriptStep(ctx context.Context, stepExecution *StepExecution, step *Step) error {
	if step.Script == "" {
		return fmt.Errorf("script is required for script step")
	}

	// Create temporary script file
	tmpFile, err := os.CreateTemp("", "pipeline-script-*.sh")
	if err != nil {
		return fmt.Errorf("failed to create temp script file: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	// Write script content
	if _, err := tmpFile.WriteString(step.Script); err != nil {
		return fmt.Errorf("failed to write script content: %w", err)
	}
	tmpFile.Close()

	// Make script executable
	if err := os.Chmod(tmpFile.Name(), 0755); err != nil {
		return fmt.Errorf("failed to make script executable: %w", err)
	}

	// Execute script
	cmd := exec.CommandContext(ctx, tmpFile.Name())

	// Set working directory
	if step.WorkingDir != "" {
		cmd.Dir = step.WorkingDir
	}

	// Set environment variables
	if len(step.Environment) > 0 {
		env := os.Environ()
		for key, value := range step.Environment {
			env = append(env, fmt.Sprintf("%s=%s", key, value))
		}
		cmd.Env = env
	}

	// Execute script
	output, err := cmd.CombinedOutput()
	stepExecution.Output = string(output)

	if err != nil {
		return fmt.Errorf("script failed: %w", err)
	}

	pe.logger.WithField("script_file", tmpFile.Name()).Info("Script step executed successfully")
	return nil
}

// executeActionStep executes an action step
func (pe *DefaultPipelineEngine) executeActionStep(ctx context.Context, stepExecution *StepExecution, step *Step) error {
	if step.Action == "" {
		return fmt.Errorf("action is required for action step")
	}

	// Simulate action execution (in production, this would integrate with actual actions)
	pe.logger.WithField("action", step.Action).Info("Action step executed (simulated)")

	stepExecution.Output = fmt.Sprintf("Action '%s' executed successfully", step.Action)
	return nil
}

// Execution completion methods

// completeExecution marks an execution as completed
func (pe *DefaultPipelineEngine) completeExecution(execution *PipelineExecution) {
	now := time.Now()
	execution.Status = ExecutionStatusSuccess
	execution.CompletedAt = &now
	execution.Duration = now.Sub(execution.StartedAt)
	execution.Logs = append(execution.Logs, fmt.Sprintf("Pipeline execution completed at %s", now.Format(time.RFC3339)))

	pe.logger.WithFields(map[string]interface{}{
		"execution_id": execution.ID,
		"pipeline_id":  execution.PipelineID,
		"duration":     execution.Duration,
		"stages":       len(execution.Stages),
	}).Info("Pipeline execution completed successfully")
}

// failExecution marks an execution as failed
func (pe *DefaultPipelineEngine) failExecution(execution *PipelineExecution, errorMsg string) {
	now := time.Now()
	execution.Status = ExecutionStatusFailure
	execution.CompletedAt = &now
	execution.Duration = now.Sub(execution.StartedAt)
	execution.Error = errorMsg
	execution.Logs = append(execution.Logs, fmt.Sprintf("Pipeline execution failed at %s: %s", now.Format(time.RFC3339), errorMsg))

	pe.logger.WithFields(map[string]interface{}{
		"execution_id": execution.ID,
		"pipeline_id":  execution.PipelineID,
		"error":        errorMsg,
		"duration":     execution.Duration,
	}).Error("Pipeline execution failed")
}
