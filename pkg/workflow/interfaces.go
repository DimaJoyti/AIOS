package workflow

import (
	"context"
	"io"
	"time"
)

// WorkflowEngine defines the interface for workflow automation
type WorkflowEngine interface {
	// Workflow management
	CreateWorkflow(ctx context.Context, workflow *Workflow) (*Workflow, error)
	GetWorkflow(ctx context.Context, workflowID string) (*Workflow, error)
	UpdateWorkflow(ctx context.Context, workflow *Workflow) (*Workflow, error)
	DeleteWorkflow(ctx context.Context, workflowID string) error
	ListWorkflows(ctx context.Context, filter *WorkflowFilter) ([]*Workflow, error)

	// Workflow execution
	ExecuteWorkflow(ctx context.Context, workflowID string, input map[string]interface{}) (*WorkflowExecution, error)
	GetExecution(ctx context.Context, executionID string) (*WorkflowExecution, error)
	ListExecutions(ctx context.Context, filter *ExecutionFilter) ([]*WorkflowExecution, error)
	CancelExecution(ctx context.Context, executionID string) error
	RetryExecution(ctx context.Context, executionID string) (*WorkflowExecution, error)

	// Event handling
	TriggerWorkflow(ctx context.Context, event *Event) ([]*WorkflowExecution, error)
	RegisterTrigger(ctx context.Context, workflowID string, trigger *Trigger) error
	UnregisterTrigger(ctx context.Context, workflowID string, triggerID string) error

	// Template management
	CreateTemplate(ctx context.Context, template *WorkflowTemplate) (*WorkflowTemplate, error)
	GetTemplate(ctx context.Context, templateID string) (*WorkflowTemplate, error)
	ListTemplates(ctx context.Context, filter *TemplateFilter) ([]*WorkflowTemplate, error)
	CreateWorkflowFromTemplate(ctx context.Context, templateID string, workflowData *Workflow) (*Workflow, error)
}

// PipelineEngine defines the interface for CI/CD pipeline management
type PipelineEngine interface {
	// Pipeline management
	CreatePipeline(ctx context.Context, pipeline *Pipeline) (*Pipeline, error)
	GetPipeline(ctx context.Context, pipelineID string) (*Pipeline, error)
	UpdatePipeline(ctx context.Context, pipeline *Pipeline) (*Pipeline, error)
	DeletePipeline(ctx context.Context, pipelineID string) error
	ListPipelines(ctx context.Context, filter *PipelineFilter) ([]*Pipeline, error)

	// Pipeline execution
	ExecutePipeline(ctx context.Context, pipelineID string, params *ExecutionParams) (*PipelineExecution, error)
	GetPipelineExecution(ctx context.Context, executionID string) (*PipelineExecution, error)
	ListPipelineExecutions(ctx context.Context, filter *PipelineExecutionFilter) ([]*PipelineExecution, error)
	CancelPipelineExecution(ctx context.Context, executionID string) error
	RetryPipelineExecution(ctx context.Context, executionID string) (*PipelineExecution, error)

	// Artifact management
	UploadArtifact(ctx context.Context, executionID string, artifact *Artifact, reader io.Reader) error
	DownloadArtifact(ctx context.Context, artifactID string, writer io.Writer) error
	ListArtifacts(ctx context.Context, executionID string) ([]*Artifact, error)
	DeleteArtifact(ctx context.Context, artifactID string) error

	// Pipeline triggers
	RegisterPipelineTrigger(ctx context.Context, pipelineID string, trigger *PipelineTrigger) error
	UnregisterPipelineTrigger(ctx context.Context, pipelineID string, triggerID string) error
	TriggerPipeline(ctx context.Context, event *Event) ([]*PipelineExecution, error)
}

// ActionExecutor defines the interface for executing workflow actions
type ActionExecutor interface {
	ExecuteAction(ctx context.Context, action *Action, input map[string]interface{}) (map[string]interface{}, error)
	ValidateAction(ctx context.Context, action *Action) error
	GetSupportedActionTypes() []ActionType
}

// TriggerManager defines the interface for managing workflow triggers
type TriggerManager interface {
	RegisterTrigger(ctx context.Context, workflowID string, trigger *Trigger) error
	UnregisterTrigger(ctx context.Context, triggerID string) error
	ProcessEvent(ctx context.Context, event *Event) ([]*Workflow, error)
	ListTriggers(ctx context.Context, workflowID string) ([]*Trigger, error)
	ValidateTrigger(ctx context.Context, trigger *Trigger) error
}

// ScheduleManager defines the interface for managing scheduled workflows
type ScheduleManager interface {
	ScheduleWorkflow(ctx context.Context, workflowID string, schedule *Schedule) error
	UnscheduleWorkflow(ctx context.Context, workflowID string) error
	GetScheduledWorkflows(ctx context.Context) ([]*ScheduledWorkflow, error)
	UpdateSchedule(ctx context.Context, workflowID string, schedule *Schedule) error
}

// IntegrationManager defines the interface for external integrations
type IntegrationManager interface {
	// GitHub integration
	ConnectGitHub(ctx context.Context, config *GitHubConfig) error
	SyncGitHubRepository(ctx context.Context, repoURL string) error
	CreateGitHubWebhook(ctx context.Context, repoURL string, events []string) error

	// GitLab integration
	ConnectGitLab(ctx context.Context, config *GitLabConfig) error
	SyncGitLabRepository(ctx context.Context, repoURL string) error
	CreateGitLabWebhook(ctx context.Context, repoURL string, events []string) error

	// Jenkins integration
	ConnectJenkins(ctx context.Context, config *JenkinsConfig) error
	TriggerJenkinsJob(ctx context.Context, jobName string, params map[string]interface{}) error
	GetJenkinsJobStatus(ctx context.Context, jobName string, buildNumber int) (*BuildStatus, error)

	// Slack integration
	ConnectSlack(ctx context.Context, config *SlackConfig) error
	SendSlackMessage(ctx context.Context, channel string, message string) error

	// Email integration
	ConnectEmail(ctx context.Context, config *EmailConfig) error
	SendEmail(ctx context.Context, to []string, subject string, body string) error

	// Generic webhook
	RegisterWebhook(ctx context.Context, config *WebhookConfig) error
	SendWebhook(ctx context.Context, url string, payload map[string]interface{}) error
}

// Filter and request types

// WorkflowFilter represents filters for workflow queries
type WorkflowFilter struct {
	Status    []WorkflowStatus `json:"status,omitempty"`
	CreatedBy string           `json:"created_by,omitempty"`
	Tags      []string         `json:"tags,omitempty"`
	Search    string           `json:"search,omitempty"`
	Limit     int              `json:"limit,omitempty"`
	Offset    int              `json:"offset,omitempty"`
}

// ExecutionFilter represents filters for execution queries
type ExecutionFilter struct {
	WorkflowID    string            `json:"workflow_id,omitempty"`
	Status        []ExecutionStatus `json:"status,omitempty"`
	StartedAfter  *time.Time        `json:"started_after,omitempty"`
	StartedBefore *time.Time        `json:"started_before,omitempty"`
	Limit         int               `json:"limit,omitempty"`
	Offset        int               `json:"offset,omitempty"`
}

// TemplateFilter represents filters for template queries
type TemplateFilter struct {
	Category  string   `json:"category,omitempty"`
	Tags      []string `json:"tags,omitempty"`
	CreatedBy string   `json:"created_by,omitempty"`
	IsPublic  *bool    `json:"is_public,omitempty"`
	Search    string   `json:"search,omitempty"`
	Limit     int      `json:"limit,omitempty"`
	Offset    int      `json:"offset,omitempty"`
}

// PipelineFilter represents filters for pipeline queries
type PipelineFilter struct {
	Status    []PipelineStatus `json:"status,omitempty"`
	CreatedBy string           `json:"created_by,omitempty"`
	Tags      []string         `json:"tags,omitempty"`
	Search    string           `json:"search,omitempty"`
	Limit     int              `json:"limit,omitempty"`
	Offset    int              `json:"offset,omitempty"`
}

// PipelineExecutionFilter represents filters for pipeline execution queries
type PipelineExecutionFilter struct {
	PipelineID    string            `json:"pipeline_id,omitempty"`
	Status        []ExecutionStatus `json:"status,omitempty"`
	Branch        string            `json:"branch,omitempty"`
	Author        string            `json:"author,omitempty"`
	StartedAfter  *time.Time        `json:"started_after,omitempty"`
	StartedBefore *time.Time        `json:"started_before,omitempty"`
	Limit         int               `json:"limit,omitempty"`
	Offset        int               `json:"offset,omitempty"`
}

// ExecutionParams represents parameters for pipeline execution
type ExecutionParams struct {
	Branch      string                 `json:"branch,omitempty"`
	Commit      string                 `json:"commit,omitempty"`
	Variables   map[string]interface{} `json:"variables,omitempty"`
	TriggerType string                 `json:"trigger_type,omitempty"`
}

// ScheduledWorkflow represents a scheduled workflow
type ScheduledWorkflow struct {
	WorkflowID string          `json:"workflow_id"`
	Schedule   *Schedule       `json:"schedule"`
	NextRun    time.Time       `json:"next_run"`
	LastRun    *time.Time      `json:"last_run,omitempty"`
	LastStatus ExecutionStatus `json:"last_status,omitempty"`
	Enabled    bool            `json:"enabled"`
}

// BuildStatus represents build status information
type BuildStatus struct {
	JobName     string          `json:"job_name"`
	BuildNumber int             `json:"build_number"`
	Status      ExecutionStatus `json:"status"`
	StartedAt   time.Time       `json:"started_at"`
	Duration    time.Duration   `json:"duration"`
	Result      string          `json:"result,omitempty"`
	URL         string          `json:"url,omitempty"`
}

// Integration configuration types

// GitHubConfig represents GitHub integration configuration
type GitHubConfig struct {
	Token        string `json:"token"`
	Organization string `json:"organization,omitempty"`
	BaseURL      string `json:"base_url,omitempty"` // For GitHub Enterprise
}

// GitLabConfig represents GitLab integration configuration
type GitLabConfig struct {
	Token   string `json:"token"`
	BaseURL string `json:"base_url"`
	GroupID string `json:"group_id,omitempty"`
}

// JenkinsConfig represents Jenkins integration configuration
type JenkinsConfig struct {
	URL      string `json:"url"`
	Username string `json:"username"`
	Token    string `json:"token"`
}

// SlackConfig represents Slack integration configuration
type SlackConfig struct {
	Token     string `json:"token"`
	Channel   string `json:"default_channel,omitempty"`
	Username  string `json:"username,omitempty"`
	IconEmoji string `json:"icon_emoji,omitempty"`
}

// EmailConfig represents email integration configuration
type EmailConfig struct {
	SMTPHost    string `json:"smtp_host"`
	SMTPPort    int    `json:"smtp_port"`
	Username    string `json:"username"`
	Password    string `json:"password"`
	FromAddress string `json:"from_address"`
	FromName    string `json:"from_name,omitempty"`
	UseTLS      bool   `json:"use_tls"`
}

// WebhookConfig represents webhook configuration
type WebhookConfig struct {
	URL     string            `json:"url"`
	Method  string            `json:"method"`
	Headers map[string]string `json:"headers,omitempty"`
	Secret  string            `json:"secret,omitempty"`
	Timeout time.Duration     `json:"timeout,omitempty"`
}

// Analytics and reporting types

// WorkflowAnalytics represents workflow analytics data
type WorkflowAnalytics struct {
	WorkflowID        string                 `json:"workflow_id"`
	TimeRange         *TimeRange             `json:"time_range"`
	TotalExecutions   int                    `json:"total_executions"`
	SuccessfulRuns    int                    `json:"successful_runs"`
	FailedRuns        int                    `json:"failed_runs"`
	AverageRunTime    time.Duration          `json:"average_run_time"`
	SuccessRate       float32                `json:"success_rate"`
	ExecutionTrends   []*ExecutionTrend      `json:"execution_trends"`
	ErrorDistribution map[string]int         `json:"error_distribution"`
	Metadata          map[string]interface{} `json:"metadata,omitempty"`
}

// ExecutionTrend represents execution trend data
type ExecutionTrend struct {
	Date       time.Time     `json:"date"`
	Executions int           `json:"executions"`
	Successes  int           `json:"successes"`
	Failures   int           `json:"failures"`
	AvgRunTime time.Duration `json:"avg_run_time"`
}

// TimeRange represents a time range
type TimeRange struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// PipelineAnalytics represents pipeline analytics data
type PipelineAnalytics struct {
	PipelineID       string                 `json:"pipeline_id"`
	TimeRange        *TimeRange             `json:"time_range"`
	TotalBuilds      int                    `json:"total_builds"`
	SuccessfulBuilds int                    `json:"successful_builds"`
	FailedBuilds     int                    `json:"failed_builds"`
	AverageBuildTime time.Duration          `json:"average_build_time"`
	SuccessRate      float32                `json:"success_rate"`
	BuildTrends      []*BuildTrend          `json:"build_trends"`
	FailureReasons   map[string]int         `json:"failure_reasons"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
}

// BuildTrend represents build trend data
type BuildTrend struct {
	Date         time.Time     `json:"date"`
	Builds       int           `json:"builds"`
	Successes    int           `json:"successes"`
	Failures     int           `json:"failures"`
	AvgBuildTime time.Duration `json:"avg_build_time"`
}
