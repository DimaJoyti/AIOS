package workflow

import (
	"time"
)

// Workflow automation types

// Workflow represents an automated workflow
type Workflow struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Version     string                 `json:"version"`
	Status      WorkflowStatus         `json:"status"`
	Triggers    []*Trigger             `json:"triggers"`
	Actions     []*Action              `json:"actions"`
	Conditions  []*Condition           `json:"conditions"`
	Variables   map[string]interface{} `json:"variables,omitempty"`
	Schedule    *Schedule              `json:"schedule,omitempty"`
	Timeout     time.Duration          `json:"timeout"`
	RetryPolicy *RetryPolicy           `json:"retry_policy,omitempty"`
	Tags        []string               `json:"tags,omitempty"`
	CreatedBy   string                 `json:"created_by"`
	UpdatedBy   string                 `json:"updated_by"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// WorkflowExecution represents a workflow execution instance
type WorkflowExecution struct {
	ID           string                 `json:"id"`
	WorkflowID   string                 `json:"workflow_id"`
	Status       ExecutionStatus        `json:"status"`
	TriggerEvent *Event                 `json:"trigger_event,omitempty"`
	StartedAt    time.Time              `json:"started_at"`
	CompletedAt  *time.Time             `json:"completed_at,omitempty"`
	Duration     time.Duration          `json:"duration"`
	Steps        []*ExecutionStep       `json:"steps"`
	Variables    map[string]interface{} `json:"variables,omitempty"`
	Error        string                 `json:"error,omitempty"`
	Logs         []string               `json:"logs,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// ExecutionStep represents a single step in workflow execution
type ExecutionStep struct {
	ID          string                 `json:"id"`
	ActionID    string                 `json:"action_id"`
	ActionType  string                 `json:"action_type"`
	Status      StepStatus             `json:"status"`
	StartedAt   time.Time              `json:"started_at"`
	CompletedAt *time.Time             `json:"completed_at,omitempty"`
	Duration    time.Duration          `json:"duration"`
	Input       map[string]interface{} `json:"input,omitempty"`
	Output      map[string]interface{} `json:"output,omitempty"`
	Error       string                 `json:"error,omitempty"`
	RetryCount  int                    `json:"retry_count"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// Trigger represents a workflow trigger
type Trigger struct {
	ID         string                 `json:"id"`
	Type       TriggerType            `json:"type"`
	Event      string                 `json:"event"`
	Conditions []*Condition           `json:"conditions,omitempty"`
	Config     map[string]interface{} `json:"config,omitempty"`
	Enabled    bool                   `json:"enabled"`
}

// Action represents a workflow action
type Action struct {
	ID          string                 `json:"id"`
	Type        ActionType             `json:"type"`
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	Config      map[string]interface{} `json:"config"`
	Timeout     time.Duration          `json:"timeout,omitempty"`
	RetryPolicy *RetryPolicy           `json:"retry_policy,omitempty"`
	DependsOn   []string               `json:"depends_on,omitempty"`
	Conditions  []*Condition           `json:"conditions,omitempty"`
	Enabled     bool                   `json:"enabled"`
}

// Condition represents a conditional expression
type Condition struct {
	ID       string      `json:"id"`
	Field    string      `json:"field"`
	Operator Operator    `json:"operator"`
	Value    interface{} `json:"value"`
	Type     ValueType   `json:"type"`
}

// Schedule represents a workflow schedule
type Schedule struct {
	Type      ScheduleType  `json:"type"`
	CronExpr  string        `json:"cron_expr,omitempty"`
	Interval  time.Duration `json:"interval,omitempty"`
	StartTime *time.Time    `json:"start_time,omitempty"`
	EndTime   *time.Time    `json:"end_time,omitempty"`
	Timezone  string        `json:"timezone,omitempty"`
	Enabled   bool          `json:"enabled"`
}

// RetryPolicy represents retry configuration
type RetryPolicy struct {
	MaxRetries    int           `json:"max_retries"`
	RetryDelay    time.Duration `json:"retry_delay"`
	BackoffFactor float64       `json:"backoff_factor"`
	MaxDelay      time.Duration `json:"max_delay"`
	RetryOn       []string      `json:"retry_on,omitempty"` // Error types to retry on
}

// Event represents a workflow event
type Event struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	Source    string                 `json:"source"`
	Subject   string                 `json:"subject,omitempty"`
	Data      map[string]interface{} `json:"data"`
	Timestamp time.Time              `json:"timestamp"`
	UserID    string                 `json:"user_id,omitempty"`
	SessionID string                 `json:"session_id,omitempty"`
}

// WorkflowTemplate represents a workflow template
type WorkflowTemplate struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Category    string                 `json:"category"`
	Tags        []string               `json:"tags"`
	Workflow    *Workflow              `json:"workflow"`
	IsPublic    bool                   `json:"is_public"`
	CreatedBy   string                 `json:"created_by"`
	UsageCount  int                    `json:"usage_count"`
	Rating      float32                `json:"rating"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// CI/CD Pipeline types

// Pipeline represents a CI/CD pipeline
type Pipeline struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Repository  *Repository            `json:"repository"`
	Stages      []*Stage               `json:"stages"`
	Variables   map[string]interface{} `json:"variables,omitempty"`
	Triggers    []*PipelineTrigger     `json:"triggers"`
	Schedule    *Schedule              `json:"schedule,omitempty"`
	Timeout     time.Duration          `json:"timeout"`
	Status      PipelineStatus         `json:"status"`
	Tags        []string               `json:"tags,omitempty"`
	CreatedBy   string                 `json:"created_by"`
	UpdatedBy   string                 `json:"updated_by"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// PipelineExecution represents a pipeline execution
type PipelineExecution struct {
	ID          string                 `json:"id"`
	PipelineID  string                 `json:"pipeline_id"`
	Status      ExecutionStatus        `json:"status"`
	Branch      string                 `json:"branch"`
	Commit      string                 `json:"commit"`
	CommitMsg   string                 `json:"commit_message,omitempty"`
	Author      string                 `json:"author,omitempty"`
	TriggerType string                 `json:"trigger_type"`
	StartedAt   time.Time              `json:"started_at"`
	CompletedAt *time.Time             `json:"completed_at,omitempty"`
	Duration    time.Duration          `json:"duration"`
	Stages      []*StageExecution      `json:"stages"`
	Variables   map[string]interface{} `json:"variables,omitempty"`
	Artifacts   []*Artifact            `json:"artifacts,omitempty"`
	Error       string                 `json:"error,omitempty"`
	Logs        []string               `json:"logs,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// Stage represents a pipeline stage
type Stage struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	Jobs        []*Job                 `json:"jobs"`
	DependsOn   []string               `json:"depends_on,omitempty"`
	Conditions  []*Condition           `json:"conditions,omitempty"`
	Variables   map[string]interface{} `json:"variables,omitempty"`
	Timeout     time.Duration          `json:"timeout,omitempty"`
	Enabled     bool                   `json:"enabled"`
}

// StageExecution represents stage execution
type StageExecution struct {
	ID          string                 `json:"id"`
	StageID     string                 `json:"stage_id"`
	Name        string                 `json:"name"`
	Status      ExecutionStatus        `json:"status"`
	StartedAt   time.Time              `json:"started_at"`
	CompletedAt *time.Time             `json:"completed_at,omitempty"`
	Duration    time.Duration          `json:"duration"`
	Jobs        []*JobExecution        `json:"jobs"`
	Variables   map[string]interface{} `json:"variables,omitempty"`
	Error       string                 `json:"error,omitempty"`
	Logs        []string               `json:"logs,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// Job represents a pipeline job
type Job struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	Steps       []*Step                `json:"steps"`
	Environment string                 `json:"environment,omitempty"`
	Variables   map[string]interface{} `json:"variables,omitempty"`
	Timeout     time.Duration          `json:"timeout,omitempty"`
	RetryPolicy *RetryPolicy           `json:"retry_policy,omitempty"`
	Conditions  []*Condition           `json:"conditions,omitempty"`
	Enabled     bool                   `json:"enabled"`
}

// JobExecution represents job execution
type JobExecution struct {
	ID          string                 `json:"id"`
	JobID       string                 `json:"job_id"`
	Name        string                 `json:"name"`
	Status      ExecutionStatus        `json:"status"`
	StartedAt   time.Time              `json:"started_at"`
	CompletedAt *time.Time             `json:"completed_at,omitempty"`
	Duration    time.Duration          `json:"duration"`
	Steps       []*StepExecution       `json:"steps"`
	Variables   map[string]interface{} `json:"variables,omitempty"`
	Artifacts   []*Artifact            `json:"artifacts,omitempty"`
	Error       string                 `json:"error,omitempty"`
	Logs        []string               `json:"logs,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// Step represents a pipeline step
type Step struct {
	ID              string                 `json:"id"`
	Name            string                 `json:"name"`
	Type            StepType               `json:"type"`
	Command         string                 `json:"command,omitempty"`
	Script          string                 `json:"script,omitempty"`
	Action          string                 `json:"action,omitempty"`
	With            map[string]interface{} `json:"with,omitempty"`
	Environment     map[string]string      `json:"environment,omitempty"`
	WorkingDir      string                 `json:"working_dir,omitempty"`
	Timeout         time.Duration          `json:"timeout,omitempty"`
	ContinueOnError bool                   `json:"continue_on_error"`
	Conditions      []*Condition           `json:"conditions,omitempty"`
	Enabled         bool                   `json:"enabled"`
}

// StepExecution represents step execution
type StepExecution struct {
	ID          string                 `json:"id"`
	StepID      string                 `json:"step_id"`
	Name        string                 `json:"name"`
	Status      ExecutionStatus        `json:"status"`
	StartedAt   time.Time              `json:"started_at"`
	CompletedAt *time.Time             `json:"completed_at,omitempty"`
	Duration    time.Duration          `json:"duration"`
	Command     string                 `json:"command,omitempty"`
	ExitCode    int                    `json:"exit_code"`
	Output      string                 `json:"output,omitempty"`
	Error       string                 `json:"error,omitempty"`
	Logs        []string               `json:"logs,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// PipelineTrigger represents a pipeline trigger
type PipelineTrigger struct {
	ID         string                 `json:"id"`
	Type       TriggerType            `json:"type"`
	Branches   []string               `json:"branches,omitempty"`
	Tags       []string               `json:"tags,omitempty"`
	Paths      []string               `json:"paths,omitempty"`
	Events     []string               `json:"events,omitempty"`
	Conditions []*Condition           `json:"conditions,omitempty"`
	Config     map[string]interface{} `json:"config,omitempty"`
	Enabled    bool                   `json:"enabled"`
}

// Repository represents a code repository
type Repository struct {
	URL      string `json:"url"`
	Branch   string `json:"branch"`
	Provider string `json:"provider"` // github, gitlab, bitbucket
	Token    string `json:"token,omitempty"`
	SSHKey   string `json:"ssh_key,omitempty"`
}

// Artifact represents a build artifact
type Artifact struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Type        string     `json:"type"`
	Path        string     `json:"path"`
	Size        int64      `json:"size"`
	Checksum    string     `json:"checksum,omitempty"`
	DownloadURL string     `json:"download_url,omitempty"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
}

// Enums

type WorkflowStatus string

const (
	WorkflowStatusDraft    WorkflowStatus = "draft"
	WorkflowStatusActive   WorkflowStatus = "active"
	WorkflowStatusInactive WorkflowStatus = "inactive"
	WorkflowStatusArchived WorkflowStatus = "archived"
)

type ExecutionStatus string

const (
	ExecutionStatusPending   ExecutionStatus = "pending"
	ExecutionStatusRunning   ExecutionStatus = "running"
	ExecutionStatusSuccess   ExecutionStatus = "success"
	ExecutionStatusFailure   ExecutionStatus = "failure"
	ExecutionStatusCancelled ExecutionStatus = "cancelled"
	ExecutionStatusTimeout   ExecutionStatus = "timeout"
)

type StepStatus string

const (
	StepStatusPending   StepStatus = "pending"
	StepStatusRunning   StepStatus = "running"
	StepStatusSuccess   StepStatus = "success"
	StepStatusFailure   StepStatus = "failure"
	StepStatusSkipped   StepStatus = "skipped"
	StepStatusCancelled StepStatus = "cancelled"
)

type TriggerType string

const (
	TriggerTypeManual   TriggerType = "manual"
	TriggerTypeSchedule TriggerType = "schedule"
	TriggerTypeWebhook  TriggerType = "webhook"
	TriggerTypeEvent    TriggerType = "event"
	TriggerTypePush     TriggerType = "push"
	TriggerTypePR       TriggerType = "pull_request"
	TriggerTypeTag      TriggerType = "tag"
	TriggerTypeRelease  TriggerType = "release"
)

type ActionType string

const (
	ActionTypeHTTP         ActionType = "http"
	ActionTypeEmail        ActionType = "email"
	ActionTypeSlack        ActionType = "slack"
	ActionTypeWebhook      ActionType = "webhook"
	ActionTypeScript       ActionType = "script"
	ActionTypeDatabase     ActionType = "database"
	ActionTypeFileSystem   ActionType = "filesystem"
	ActionTypeNotification ActionType = "notification"
	ActionTypeIntegration  ActionType = "integration"
)

type Operator string

const (
	OperatorEquals      Operator = "equals"
	OperatorNotEquals   Operator = "not_equals"
	OperatorGreaterThan Operator = "greater_than"
	OperatorLessThan    Operator = "less_than"
	OperatorContains    Operator = "contains"
	OperatorStartsWith  Operator = "starts_with"
	OperatorEndsWith    Operator = "ends_with"
	OperatorRegex       Operator = "regex"
	OperatorIn          Operator = "in"
	OperatorNotIn       Operator = "not_in"
)

type ValueType string

const (
	ValueTypeString  ValueType = "string"
	ValueTypeNumber  ValueType = "number"
	ValueTypeBoolean ValueType = "boolean"
	ValueTypeArray   ValueType = "array"
	ValueTypeObject  ValueType = "object"
)

type ScheduleType string

const (
	ScheduleTypeCron     ScheduleType = "cron"
	ScheduleTypeInterval ScheduleType = "interval"
	ScheduleTypeOnce     ScheduleType = "once"
)

type PipelineStatus string

const (
	PipelineStatusActive   PipelineStatus = "active"
	PipelineStatusInactive PipelineStatus = "inactive"
	PipelineStatusArchived PipelineStatus = "archived"
)

type StepType string

const (
	StepTypeCommand StepType = "command"
	StepTypeScript  StepType = "script"
	StepTypeAction  StepType = "action"
	StepTypeBuild   StepType = "build"
	StepTypeTest    StepType = "test"
	StepTypeDeploy  StepType = "deploy"
)
