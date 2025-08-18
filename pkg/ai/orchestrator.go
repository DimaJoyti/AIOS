package ai

import (
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
)

// Orchestrator defines the interface for AI orchestration
type Orchestrator interface {
	// Execution methods
	ExecuteTask(ctx context.Context, task *Task) (*TaskResult, error)
	ExecuteWorkflow(ctx context.Context, workflow *Workflow) (*WorkflowResult, error)

	// Management methods
	RegisterAgent(agent Agent) error
	UnregisterAgent(agentID string) error
	GetAgent(agentID string) (Agent, error)
	ListAgents() []Agent

	// Monitoring methods
	GetStatus() *OrchestratorStatus
	GetMetrics() *OrchestratorMetrics

	// Lifecycle methods
	Start(ctx context.Context) error
	Stop() error
}

// Task represents an AI task
type Task struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Description string                 `json:"description"`
	Input       map[string]interface{} `json:"input"`
	Config      *TaskConfig            `json:"config,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
}

// TaskConfig represents task configuration
type TaskConfig struct {
	Timeout    time.Duration          `json:"timeout"`
	MaxRetries int                    `json:"max_retries"`
	AgentID    string                 `json:"agent_id,omitempty"`
	Parameters map[string]interface{} `json:"parameters,omitempty"`
}

// TaskResult represents the result of a task execution
type TaskResult struct {
	TaskID      string                 `json:"task_id"`
	Success     bool                   `json:"success"`
	Output      map[string]interface{} `json:"output,omitempty"`
	Error       string                 `json:"error,omitempty"`
	Duration    time.Duration          `json:"duration"`
	AgentID     string                 `json:"agent_id"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	CompletedAt time.Time              `json:"completed_at"`
}

// Workflow represents a workflow of AI tasks
type Workflow struct {
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	Description  string                 `json:"description"`
	Tasks        []*Task                `json:"tasks"`
	Dependencies map[string][]string    `json:"dependencies"`
	Config       *WorkflowConfig        `json:"config,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt    time.Time              `json:"created_at"`
}

// WorkflowConfig represents workflow configuration
type WorkflowConfig struct {
	Timeout         time.Duration          `json:"timeout"`
	MaxConcurrency  int                    `json:"max_concurrency"`
	FailureStrategy string                 `json:"failure_strategy"` // "stop", "continue", "retry"
	Parameters      map[string]interface{} `json:"parameters,omitempty"`
}

// WorkflowResult represents the result of a workflow execution
type WorkflowResult struct {
	WorkflowID  string                 `json:"workflow_id"`
	Success     bool                   `json:"success"`
	TaskResults map[string]*TaskResult `json:"task_results"`
	Error       string                 `json:"error,omitempty"`
	Duration    time.Duration          `json:"duration"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	CompletedAt time.Time              `json:"completed_at"`
}

// Agent defines the interface for AI agents
type Agent interface {
	GetID() string
	GetName() string
	GetDescription() string
	GetCapabilities() []string
	ExecuteTask(ctx context.Context, task *Task) (*TaskResult, error)
	IsAvailable() bool
	GetStatus() *AgentStatus
}

// AgentStatus represents the status of an agent
type AgentStatus struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Available   bool      `json:"available"`
	Busy        bool      `json:"busy"`
	LastActive  time.Time `json:"last_active"`
	TaskCount   int64     `json:"task_count"`
	ErrorCount  int64     `json:"error_count"`
	SuccessRate float64   `json:"success_rate"`
}

// OrchestratorStatus represents the status of the orchestrator
type OrchestratorStatus struct {
	Running        bool                    `json:"running"`
	AgentCount     int                     `json:"agent_count"`
	ActiveTasks    int                     `json:"active_tasks"`
	QueuedTasks    int                     `json:"queued_tasks"`
	CompletedTasks int64                   `json:"completed_tasks"`
	FailedTasks    int64                   `json:"failed_tasks"`
	Uptime         time.Duration           `json:"uptime"`
	LastActivity   time.Time               `json:"last_activity"`
	AgentStatuses  map[string]*AgentStatus `json:"agent_statuses"`
}

// OrchestratorMetrics represents orchestrator metrics
type OrchestratorMetrics struct {
	TotalTasks       int64         `json:"total_tasks"`
	SuccessfulTasks  int64         `json:"successful_tasks"`
	FailedTasks      int64         `json:"failed_tasks"`
	AverageLatency   time.Duration `json:"average_latency"`
	ThroughputPerSec float64       `json:"throughput_per_sec"`
	ErrorRate        float64       `json:"error_rate"`
	StartTime        time.Time     `json:"start_time"`
	LastUpdate       time.Time     `json:"last_update"`
}

// DefaultOrchestrator implements the Orchestrator interface
type DefaultOrchestrator struct {
	agents    map[string]Agent
	logger    *logrus.Logger
	running   bool
	startTime time.Time
	metrics   *OrchestratorMetrics
}

// NewDefaultOrchestrator creates a new default orchestrator
func NewDefaultOrchestrator(logger *logrus.Logger) Orchestrator {
	return &DefaultOrchestrator{
		agents:    make(map[string]Agent),
		logger:    logger,
		startTime: time.Now(),
		metrics: &OrchestratorMetrics{
			StartTime:  time.Now(),
			LastUpdate: time.Now(),
		},
	}
}

// ExecuteTask executes a single task
func (o *DefaultOrchestrator) ExecuteTask(ctx context.Context, task *Task) (*TaskResult, error) {
	startTime := time.Now()

	// Find appropriate agent
	var agent Agent
	if task.Config != nil && task.Config.AgentID != "" {
		var exists bool
		agent, exists = o.agents[task.Config.AgentID]
		if !exists {
			return &TaskResult{
				TaskID:      task.ID,
				Success:     false,
				Error:       "specified agent not found",
				Duration:    time.Since(startTime),
				CompletedAt: time.Now(),
			}, nil
		}
	} else {
		// Find first available agent
		for _, a := range o.agents {
			if a.IsAvailable() {
				agent = a
				break
			}
		}
	}

	if agent == nil {
		return &TaskResult{
			TaskID:      task.ID,
			Success:     false,
			Error:       "no available agents",
			Duration:    time.Since(startTime),
			CompletedAt: time.Now(),
		}, nil
	}

	// Execute task
	result, err := agent.ExecuteTask(ctx, task)
	if err != nil {
		return &TaskResult{
			TaskID:      task.ID,
			Success:     false,
			Error:       err.Error(),
			Duration:    time.Since(startTime),
			AgentID:     agent.GetID(),
			CompletedAt: time.Now(),
		}, err
	}

	// Update metrics
	o.updateMetrics(result)

	return result, nil
}

// ExecuteWorkflow executes a workflow
func (o *DefaultOrchestrator) ExecuteWorkflow(ctx context.Context, workflow *Workflow) (*WorkflowResult, error) {
	startTime := time.Now()

	result := &WorkflowResult{
		WorkflowID:  workflow.ID,
		TaskResults: make(map[string]*TaskResult),
		CompletedAt: time.Now(),
	}

	// Simple sequential execution for now
	for _, task := range workflow.Tasks {
		taskResult, err := o.ExecuteTask(ctx, task)
		if err != nil {
			result.Success = false
			result.Error = err.Error()
			result.Duration = time.Since(startTime)
			return result, err
		}

		result.TaskResults[task.ID] = taskResult

		if !taskResult.Success && workflow.Config != nil && workflow.Config.FailureStrategy == "stop" {
			result.Success = false
			result.Error = "workflow stopped due to task failure"
			result.Duration = time.Since(startTime)
			return result, nil
		}
	}

	result.Success = true
	result.Duration = time.Since(startTime)
	return result, nil
}

// RegisterAgent registers an agent
func (o *DefaultOrchestrator) RegisterAgent(agent Agent) error {
	o.agents[agent.GetID()] = agent
	o.logger.WithField("agent_id", agent.GetID()).Info("Agent registered")
	return nil
}

// UnregisterAgent unregisters an agent
func (o *DefaultOrchestrator) UnregisterAgent(agentID string) error {
	delete(o.agents, agentID)
	o.logger.WithField("agent_id", agentID).Info("Agent unregistered")
	return nil
}

// GetAgent gets an agent by ID
func (o *DefaultOrchestrator) GetAgent(agentID string) (Agent, error) {
	agent, exists := o.agents[agentID]
	if !exists {
		return nil, fmt.Errorf("agent not found: %s", agentID)
	}
	return agent, nil
}

// ListAgents lists all agents
func (o *DefaultOrchestrator) ListAgents() []Agent {
	agents := make([]Agent, 0, len(o.agents))
	for _, agent := range o.agents {
		agents = append(agents, agent)
	}
	return agents
}

// GetStatus returns orchestrator status
func (o *DefaultOrchestrator) GetStatus() *OrchestratorStatus {
	agentStatuses := make(map[string]*AgentStatus)
	for id, agent := range o.agents {
		agentStatuses[id] = agent.GetStatus()
	}

	return &OrchestratorStatus{
		Running:       o.running,
		AgentCount:    len(o.agents),
		Uptime:        time.Since(o.startTime),
		LastActivity:  time.Now(),
		AgentStatuses: agentStatuses,
	}
}

// GetMetrics returns orchestrator metrics
func (o *DefaultOrchestrator) GetMetrics() *OrchestratorMetrics {
	o.metrics.LastUpdate = time.Now()
	return o.metrics
}

// Start starts the orchestrator
func (o *DefaultOrchestrator) Start(ctx context.Context) error {
	o.running = true
	o.logger.Info("AI Orchestrator started")
	return nil
}

// Stop stops the orchestrator
func (o *DefaultOrchestrator) Stop() error {
	o.running = false
	o.logger.Info("AI Orchestrator stopped")
	return nil
}

// updateMetrics updates orchestrator metrics
func (o *DefaultOrchestrator) updateMetrics(result *TaskResult) {
	o.metrics.TotalTasks++
	if result.Success {
		o.metrics.SuccessfulTasks++
	} else {
		o.metrics.FailedTasks++
	}

	if o.metrics.TotalTasks > 0 {
		o.metrics.ErrorRate = float64(o.metrics.FailedTasks) / float64(o.metrics.TotalTasks)
	}
}
