package project

import (
	"time"
)

// Core project management types

// Project represents a project with tasks, milestones, and team members
type Project struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Status      ProjectStatus          `json:"status"`
	Priority    Priority               `json:"priority"`
	StartDate   *time.Time             `json:"start_date,omitempty"`
	EndDate     *time.Time             `json:"end_date,omitempty"`
	DueDate     *time.Time             `json:"due_date,omitempty"`
	Owner       string                 `json:"owner"`
	TeamMembers []string               `json:"team_members"`
	Tags        []string               `json:"tags"`
	Progress    float32                `json:"progress"`
	Budget      *Budget                `json:"budget,omitempty"`
	Repository  *Repository            `json:"repository,omitempty"`
	Settings    *ProjectSettings       `json:"settings"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// Task represents a task within a project
type Task struct {
	ID             string                 `json:"id"`
	ProjectID      string                 `json:"project_id"`
	Title          string                 `json:"title"`
	Description    string                 `json:"description"`
	Status         TaskStatus             `json:"status"`
	Priority       Priority               `json:"priority"`
	Type           TaskType               `json:"type"`
	Assignee       string                 `json:"assignee,omitempty"`
	Reporter       string                 `json:"reporter"`
	Labels         []string               `json:"labels"`
	Dependencies   []string               `json:"dependencies"` // Task IDs this task depends on
	Subtasks       []string               `json:"subtasks"`     // Child task IDs
	ParentTask     string                 `json:"parent_task,omitempty"`
	EstimatedHours float32                `json:"estimated_hours,omitempty"`
	ActualHours    float32                `json:"actual_hours,omitempty"`
	StartDate      *time.Time             `json:"start_date,omitempty"`
	DueDate        *time.Time             `json:"due_date,omitempty"`
	CompletedAt    *time.Time             `json:"completed_at,omitempty"`
	Attachments    []*Attachment          `json:"attachments,omitempty"`
	Comments       []*Comment             `json:"comments,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt      time.Time              `json:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at"`
}

// Milestone represents a project milestone
type Milestone struct {
	ID          string                 `json:"id"`
	ProjectID   string                 `json:"project_id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Status      MilestoneStatus        `json:"status"`
	DueDate     *time.Time             `json:"due_date,omitempty"`
	CompletedAt *time.Time             `json:"completed_at,omitempty"`
	Tasks       []string               `json:"tasks"` // Task IDs associated with this milestone
	Progress    float32                `json:"progress"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// Sprint represents an agile sprint
type Sprint struct {
	ID        string                 `json:"id"`
	ProjectID string                 `json:"project_id"`
	Name      string                 `json:"name"`
	Goal      string                 `json:"goal"`
	Status    SprintStatus           `json:"status"`
	StartDate time.Time              `json:"start_date"`
	EndDate   time.Time              `json:"end_date"`
	Tasks     []string               `json:"tasks"`
	Capacity  int                    `json:"capacity"` // Story points or hours
	Velocity  float32                `json:"velocity"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
}

// TeamMember represents a team member
type TeamMember struct {
	ID           string                 `json:"id"`
	Username     string                 `json:"username"`
	Email        string                 `json:"email"`
	FullName     string                 `json:"full_name"`
	Role         TeamRole               `json:"role"`
	Permissions  []Permission           `json:"permissions"`
	Skills       []string               `json:"skills"`
	Availability *Availability          `json:"availability,omitempty"`
	Timezone     string                 `json:"timezone"`
	Avatar       string                 `json:"avatar,omitempty"`
	Status       MemberStatus           `json:"status"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	JoinedAt     time.Time              `json:"joined_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
}

// Comment represents a comment on a task or project
type Comment struct {
	ID        string                 `json:"id"`
	Author    string                 `json:"author"`
	Content   string                 `json:"content"`
	Type      CommentType            `json:"type"`
	Mentions  []string               `json:"mentions,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
}

// Attachment represents a file attachment
type Attachment struct {
	ID         string                 `json:"id"`
	Name       string                 `json:"name"`
	Type       string                 `json:"type"`
	Size       int64                  `json:"size"`
	URL        string                 `json:"url"`
	UploadedBy string                 `json:"uploaded_by"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
	UploadedAt time.Time              `json:"uploaded_at"`
}

// Budget represents project budget information
type Budget struct {
	Total     float64   `json:"total"`
	Spent     float64   `json:"spent"`
	Currency  string    `json:"currency"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Repository represents a code repository
type Repository struct {
	URL      string `json:"url"`
	Branch   string `json:"branch"`
	Provider string `json:"provider"` // github, gitlab, bitbucket
	Token    string `json:"token,omitempty"`
}

// ProjectSettings represents project configuration
type ProjectSettings struct {
	Visibility      Visibility             `json:"visibility"`
	AllowGuests     bool                   `json:"allow_guests"`
	RequireApproval bool                   `json:"require_approval"`
	AutoAssignment  bool                   `json:"auto_assignment"`
	Notifications   *NotificationSettings  `json:"notifications"`
	Integrations    map[string]interface{} `json:"integrations,omitempty"`
	CustomFields    []*CustomField         `json:"custom_fields,omitempty"`
}

// NotificationSettings represents notification preferences
type NotificationSettings struct {
	Email   bool `json:"email"`
	InApp   bool `json:"in_app"`
	Slack   bool `json:"slack"`
	Webhook bool `json:"webhook"`
}

// CustomField represents a custom field definition
type CustomField struct {
	ID       string      `json:"id"`
	Name     string      `json:"name"`
	Type     FieldType   `json:"type"`
	Required bool        `json:"required"`
	Options  []string    `json:"options,omitempty"`
	Default  interface{} `json:"default,omitempty"`
}

// Availability represents team member availability
type Availability struct {
	WorkingHours *WorkingHours `json:"working_hours"`
	TimeOff      []*TimeOff    `json:"time_off,omitempty"`
	Capacity     float32       `json:"capacity"` // 0.0 to 1.0
}

// WorkingHours represents working hours
type WorkingHours struct {
	Monday    *DayHours `json:"monday"`
	Tuesday   *DayHours `json:"tuesday"`
	Wednesday *DayHours `json:"wednesday"`
	Thursday  *DayHours `json:"thursday"`
	Friday    *DayHours `json:"friday"`
	Saturday  *DayHours `json:"saturday"`
	Sunday    *DayHours `json:"sunday"`
}

// DayHours represents hours for a specific day
type DayHours struct {
	Start string `json:"start"` // HH:MM format
	End   string `json:"end"`   // HH:MM format
}

// TimeOff represents time off periods
type TimeOff struct {
	StartDate time.Time `json:"start_date"`
	EndDate   time.Time `json:"end_date"`
	Type      string    `json:"type"` // vacation, sick, holiday
	Reason    string    `json:"reason,omitempty"`
}

// Enums and constants

type ProjectStatus string

const (
	ProjectStatusPlanning  ProjectStatus = "planning"
	ProjectStatusActive    ProjectStatus = "active"
	ProjectStatusOnHold    ProjectStatus = "on_hold"
	ProjectStatusCompleted ProjectStatus = "completed"
	ProjectStatusCancelled ProjectStatus = "cancelled"
	ProjectStatusArchived  ProjectStatus = "archived"
)

type TaskStatus string

const (
	TaskStatusTodo       TaskStatus = "todo"
	TaskStatusInProgress TaskStatus = "in_progress"
	TaskStatusInReview   TaskStatus = "in_review"
	TaskStatusBlocked    TaskStatus = "blocked"
	TaskStatusDone       TaskStatus = "done"
	TaskStatusCancelled  TaskStatus = "cancelled"
)

type TaskType string

const (
	TaskTypeFeature TaskType = "feature"
	TaskTypeBug     TaskType = "bug"
	TaskTypeTask    TaskType = "task"
	TaskTypeStory   TaskType = "story"
	TaskTypeEpic    TaskType = "epic"
	TaskTypeSpike   TaskType = "spike"
)

type Priority string

const (
	PriorityLow      Priority = "low"
	PriorityMedium   Priority = "medium"
	PriorityHigh     Priority = "high"
	PriorityCritical Priority = "critical"
)

type MilestoneStatus string

const (
	MilestoneStatusOpen      MilestoneStatus = "open"
	MilestoneStatusCompleted MilestoneStatus = "completed"
	MilestoneStatusOverdue   MilestoneStatus = "overdue"
)

type SprintStatus string

const (
	SprintStatusPlanning  SprintStatus = "planning"
	SprintStatusActive    SprintStatus = "active"
	SprintStatusCompleted SprintStatus = "completed"
)

type TeamRole string

const (
	TeamRoleOwner      TeamRole = "owner"
	TeamRoleAdmin      TeamRole = "admin"
	TeamRoleMaintainer TeamRole = "maintainer"
	TeamRoleDeveloper  TeamRole = "developer"
	TeamRoleReporter   TeamRole = "reporter"
	TeamRoleGuest      TeamRole = "guest"
)

type Permission string

const (
	PermissionRead        Permission = "read"
	PermissionWrite       Permission = "write"
	PermissionDelete      Permission = "delete"
	PermissionAdmin       Permission = "admin"
	PermissionManageTeam  Permission = "manage_team"
	PermissionManageRoles Permission = "manage_roles"
)

type MemberStatus string

const (
	MemberStatusActive   MemberStatus = "active"
	MemberStatusInactive MemberStatus = "inactive"
	MemberStatusInvited  MemberStatus = "invited"
)

type CommentType string

const (
	CommentTypeGeneral CommentType = "general"
	CommentTypeUpdate  CommentType = "update"
	CommentTypeReview  CommentType = "review"
	CommentTypeSystem  CommentType = "system"
)

type Visibility string

const (
	VisibilityPublic   Visibility = "public"
	VisibilityPrivate  Visibility = "private"
	VisibilityInternal Visibility = "internal"
)

type FieldType string

const (
	FieldTypeText        FieldType = "text"
	FieldTypeNumber      FieldType = "number"
	FieldTypeDate        FieldType = "date"
	FieldTypeSelect      FieldType = "select"
	FieldTypeMultiSelect FieldType = "multi_select"
	FieldTypeBoolean     FieldType = "boolean"
	FieldTypeURL         FieldType = "url"
	FieldTypeEmail       FieldType = "email"
)

// Additional types for project management

// ProjectSchedule represents a project schedule
type ProjectSchedule struct {
	ProjectID    string                   `json:"project_id"`
	StartDate    time.Time                `json:"start_date"`
	EndDate      time.Time                `json:"end_date"`
	CriticalPath []*Task                  `json:"critical_path"`
	TaskSchedule map[string]*TaskSchedule `json:"task_schedule"`
	Milestones   []*MilestoneSchedule     `json:"milestones"`
	GeneratedAt  time.Time                `json:"generated_at"`
}

// TaskSchedule represents task scheduling information
type TaskSchedule struct {
	TaskID         string        `json:"task_id"`
	EarliestStart  time.Time     `json:"earliest_start"`
	LatestStart    time.Time     `json:"latest_start"`
	EarliestFinish time.Time     `json:"earliest_finish"`
	LatestFinish   time.Time     `json:"latest_finish"`
	Slack          time.Duration `json:"slack"`
	IsCritical     bool          `json:"is_critical"`
}

// MilestoneSchedule represents milestone scheduling
type MilestoneSchedule struct {
	MilestoneID   string    `json:"milestone_id"`
	ScheduledDate time.Time `json:"scheduled_date"`
	IsCritical    bool      `json:"is_critical"`
}

// Notification represents a notification
type Notification struct {
	ID        string                 `json:"id"`
	UserID    string                 `json:"user_id"`
	Type      NotificationType       `json:"type"`
	Event     NotificationEvent      `json:"event"`
	Title     string                 `json:"title"`
	Message   string                 `json:"message"`
	Data      map[string]interface{} `json:"data,omitempty"`
	Read      bool                   `json:"read"`
	CreatedAt time.Time              `json:"created_at"`
	ReadAt    *time.Time             `json:"read_at,omitempty"`
}

// NotificationFilter represents notification filters
type NotificationFilter struct {
	Types     []NotificationType  `json:"types,omitempty"`
	Events    []NotificationEvent `json:"events,omitempty"`
	Read      *bool               `json:"read,omitempty"`
	ProjectID string              `json:"project_id,omitempty"`
	Limit     int                 `json:"limit,omitempty"`
	Offset    int                 `json:"offset,omitempty"`
}

// ProjectTemplate represents a project template
type ProjectTemplate struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Category    string                 `json:"category"`
	Tags        []string               `json:"tags"`
	Project     *Project               `json:"project"`
	Tasks       []*Task                `json:"tasks"`
	Milestones  []*Milestone           `json:"milestones"`
	IsPublic    bool                   `json:"is_public"`
	CreatedBy   string                 `json:"created_by"`
	UsageCount  int                    `json:"usage_count"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// TemplateFilter represents template filters
type TemplateFilter struct {
	Category  string   `json:"category,omitempty"`
	Tags      []string `json:"tags,omitempty"`
	CreatedBy string   `json:"created_by,omitempty"`
	IsPublic  *bool    `json:"is_public,omitempty"`
	Search    string   `json:"search,omitempty"`
	Limit     int      `json:"limit,omitempty"`
	Offset    int      `json:"offset,omitempty"`
}

// TimeEntry represents a time tracking entry
type TimeEntry struct {
	ID          string                 `json:"id"`
	TaskID      string                 `json:"task_id"`
	UserID      string                 `json:"user_id"`
	Description string                 `json:"description,omitempty"`
	StartTime   time.Time              `json:"start_time"`
	EndTime     *time.Time             `json:"end_time,omitempty"`
	Duration    time.Duration          `json:"duration"`
	IsBillable  bool                   `json:"is_billable"`
	HourlyRate  float64                `json:"hourly_rate,omitempty"`
	Tags        []string               `json:"tags,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// TimeEntryFilter represents time entry filters
type TimeEntryFilter struct {
	TaskID    string     `json:"task_id,omitempty"`
	UserID    string     `json:"user_id,omitempty"`
	ProjectID string     `json:"project_id,omitempty"`
	StartDate *time.Time `json:"start_date,omitempty"`
	EndDate   *time.Time `json:"end_date,omitempty"`
	Tags      []string   `json:"tags,omitempty"`
	Billable  *bool      `json:"billable,omitempty"`
	Limit     int        `json:"limit,omitempty"`
	Offset    int        `json:"offset,omitempty"`
}

// TimeReportRequest represents a time report request
type TimeReportRequest struct {
	ProjectID string       `json:"project_id,omitempty"`
	UserID    string       `json:"user_id,omitempty"`
	StartDate time.Time    `json:"start_date"`
	EndDate   time.Time    `json:"end_date"`
	GroupBy   string       `json:"group_by"` // user, task, project, day, week, month
	Format    ReportFormat `json:"format"`
}

// TimeReport represents a time report
type TimeReport struct {
	Request       *TimeReportRequest     `json:"request"`
	TotalHours    float64                `json:"total_hours"`
	BillableHours float64                `json:"billable_hours"`
	TotalCost     float64                `json:"total_cost"`
	Entries       []*TimeReportEntry     `json:"entries"`
	Summary       map[string]interface{} `json:"summary"`
	GeneratedAt   time.Time              `json:"generated_at"`
}

// TimeReportEntry represents an entry in a time report
type TimeReportEntry struct {
	Label         string       `json:"label"`
	Hours         float64      `json:"hours"`
	BillableHours float64      `json:"billable_hours"`
	Cost          float64      `json:"cost"`
	Entries       []*TimeEntry `json:"entries,omitempty"`
}

// Integration types

// CIConfig represents CI/CD configuration
type CIConfig struct {
	Provider string                 `json:"provider"` // github_actions, gitlab_ci, jenkins
	URL      string                 `json:"url"`
	Token    string                 `json:"token"`
	Config   map[string]interface{} `json:"config"`
}

// BuildResult represents a build result
type BuildResult struct {
	ID         string                 `json:"id"`
	Status     BuildStatusType        `json:"status"`
	Branch     string                 `json:"branch"`
	Commit     string                 `json:"commit"`
	StartedAt  time.Time              `json:"started_at"`
	FinishedAt *time.Time             `json:"finished_at,omitempty"`
	Duration   time.Duration          `json:"duration"`
	Logs       string                 `json:"logs,omitempty"`
	Artifacts  []string               `json:"artifacts,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// BuildStatus represents build status information
type BuildStatus struct {
	BuildID   string          `json:"build_id"`
	Status    BuildStatusType `json:"status"`
	Progress  float32         `json:"progress"`
	Message   string          `json:"message,omitempty"`
	UpdatedAt time.Time       `json:"updated_at"`
}

// IssueTrackerConfig represents issue tracker configuration
type IssueTrackerConfig struct {
	Provider string                 `json:"provider"` // jira, github, gitlab
	URL      string                 `json:"url"`
	Token    string                 `json:"token"`
	Project  string                 `json:"project"`
	Config   map[string]interface{} `json:"config"`
}

// Enums for additional types

type NotificationType string

const (
	NotificationTypeInfo    NotificationType = "info"
	NotificationTypeWarning NotificationType = "warning"
	NotificationTypeError   NotificationType = "error"
	NotificationTypeSuccess NotificationType = "success"
)

type NotificationEvent string

const (
	NotificationEventTaskCreated         NotificationEvent = "task_created"
	NotificationEventTaskUpdated         NotificationEvent = "task_updated"
	NotificationEventTaskAssigned        NotificationEvent = "task_assigned"
	NotificationEventTaskCompleted       NotificationEvent = "task_completed"
	NotificationEventProjectCreated      NotificationEvent = "project_created"
	NotificationEventProjectUpdated      NotificationEvent = "project_updated"
	NotificationEventMilestoneReached    NotificationEvent = "milestone_reached"
	NotificationEventSprintStarted       NotificationEvent = "sprint_started"
	NotificationEventSprintCompleted     NotificationEvent = "sprint_completed"
	NotificationEventDeadlineApproaching NotificationEvent = "deadline_approaching"
	NotificationEventMemberAdded         NotificationEvent = "member_added"
	NotificationEventMemberRemoved       NotificationEvent = "member_removed"
)

type BuildStatusType string

const (
	BuildStatusPending   BuildStatusType = "pending"
	BuildStatusRunning   BuildStatusType = "running"
	BuildStatusSuccess   BuildStatusType = "success"
	BuildStatusFailure   BuildStatusType = "failure"
	BuildStatusCancelled BuildStatusType = "cancelled"
)
