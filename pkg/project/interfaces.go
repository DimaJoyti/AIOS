package project

import (
	"context"
	"time"
)

// ProjectManager defines the interface for project management operations
type ProjectManager interface {
	// Project operations
	CreateProject(ctx context.Context, project *Project) (*Project, error)
	GetProject(ctx context.Context, projectID string) (*Project, error)
	UpdateProject(ctx context.Context, project *Project) (*Project, error)
	DeleteProject(ctx context.Context, projectID string) error
	ListProjects(ctx context.Context, filter *ProjectFilter) ([]*Project, error)
	ArchiveProject(ctx context.Context, projectID string) error

	// Task operations
	CreateTask(ctx context.Context, task *Task) (*Task, error)
	GetTask(ctx context.Context, taskID string) (*Task, error)
	UpdateTask(ctx context.Context, task *Task) (*Task, error)
	DeleteTask(ctx context.Context, taskID string) error
	ListTasks(ctx context.Context, filter *TaskFilter) ([]*Task, error)
	AssignTask(ctx context.Context, taskID string, assignee string) error
	UpdateTaskStatus(ctx context.Context, taskID string, status TaskStatus) error

	// Milestone operations
	CreateMilestone(ctx context.Context, milestone *Milestone) (*Milestone, error)
	GetMilestone(ctx context.Context, milestoneID string) (*Milestone, error)
	UpdateMilestone(ctx context.Context, milestone *Milestone) (*Milestone, error)
	DeleteMilestone(ctx context.Context, milestoneID string) error
	ListMilestones(ctx context.Context, projectID string) ([]*Milestone, error)

	// Sprint operations
	CreateSprint(ctx context.Context, sprint *Sprint) (*Sprint, error)
	GetSprint(ctx context.Context, sprintID string) (*Sprint, error)
	UpdateSprint(ctx context.Context, sprint *Sprint) (*Sprint, error)
	DeleteSprint(ctx context.Context, sprintID string) error
	ListSprints(ctx context.Context, projectID string) ([]*Sprint, error)
	StartSprint(ctx context.Context, sprintID string) error
	CompleteSprint(ctx context.Context, sprintID string) error

	// Team operations
	AddTeamMember(ctx context.Context, projectID string, member *TeamMember) error
	RemoveTeamMember(ctx context.Context, projectID string, memberID string) error
	UpdateTeamMember(ctx context.Context, member *TeamMember) error
	GetTeamMember(ctx context.Context, memberID string) (*TeamMember, error)
	ListTeamMembers(ctx context.Context, projectID string) ([]*TeamMember, error)

	// Analytics and reporting
	GetProjectAnalytics(ctx context.Context, projectID string, timeRange *TimeRange) (*ProjectAnalytics, error)
	GetTaskAnalytics(ctx context.Context, projectID string, timeRange *TimeRange) (*TaskAnalytics, error)
	GetTeamAnalytics(ctx context.Context, projectID string, timeRange *TimeRange) (*TeamAnalytics, error)
	GenerateReport(ctx context.Context, request *ReportRequest) (*Report, error)
}

// TaskDependencyManager handles task dependencies and scheduling
type TaskDependencyManager interface {
	AddDependency(ctx context.Context, taskID string, dependsOnTaskID string) error
	RemoveDependency(ctx context.Context, taskID string, dependsOnTaskID string) error
	GetDependencies(ctx context.Context, taskID string) ([]*Task, error)
	GetDependents(ctx context.Context, taskID string) ([]*Task, error)
	ValidateDependencies(ctx context.Context, taskID string) error
	GetCriticalPath(ctx context.Context, projectID string) ([]*Task, error)
	CalculateSchedule(ctx context.Context, projectID string) (*ProjectSchedule, error)
}

// NotificationManager handles project notifications
type NotificationManager interface {
	SendNotification(ctx context.Context, notification *Notification) error
	SubscribeToProject(ctx context.Context, projectID string, userID string, events []NotificationEvent) error
	UnsubscribeFromProject(ctx context.Context, projectID string, userID string) error
	GetNotifications(ctx context.Context, userID string, filter *NotificationFilter) ([]*Notification, error)
	MarkAsRead(ctx context.Context, notificationID string) error
	GetNotificationSettings(ctx context.Context, userID string) (*NotificationSettings, error)
	UpdateNotificationSettings(ctx context.Context, userID string, settings *NotificationSettings) error
}

// ProjectTemplateManager handles project templates
type ProjectTemplateManager interface {
	CreateTemplate(ctx context.Context, template *ProjectTemplate) (*ProjectTemplate, error)
	GetTemplate(ctx context.Context, templateID string) (*ProjectTemplate, error)
	UpdateTemplate(ctx context.Context, template *ProjectTemplate) (*ProjectTemplate, error)
	DeleteTemplate(ctx context.Context, templateID string) error
	ListTemplates(ctx context.Context, filter *TemplateFilter) ([]*ProjectTemplate, error)
	CreateProjectFromTemplate(ctx context.Context, templateID string, projectData *Project) (*Project, error)
}

// TimeTrackingManager handles time tracking for tasks
type TimeTrackingManager interface {
	StartTimer(ctx context.Context, taskID string, userID string) (*TimeEntry, error)
	StopTimer(ctx context.Context, entryID string) (*TimeEntry, error)
	LogTime(ctx context.Context, entry *TimeEntry) (*TimeEntry, error)
	GetTimeEntries(ctx context.Context, filter *TimeEntryFilter) ([]*TimeEntry, error)
	GetTimeReport(ctx context.Context, request *TimeReportRequest) (*TimeReport, error)
	UpdateTimeEntry(ctx context.Context, entry *TimeEntry) (*TimeEntry, error)
	DeleteTimeEntry(ctx context.Context, entryID string) error
}

// ProjectIntegrationManager handles external integrations
type ProjectIntegrationManager interface {
	ConnectRepository(ctx context.Context, projectID string, repo *Repository) error
	SyncWithRepository(ctx context.Context, projectID string) error
	ConnectCI(ctx context.Context, projectID string, config *CIConfig) error
	TriggerBuild(ctx context.Context, projectID string, branch string) (*BuildResult, error)
	GetBuildStatus(ctx context.Context, projectID string, buildID string) (*BuildStatus, error)
	ConnectIssueTracker(ctx context.Context, projectID string, config *IssueTrackerConfig) error
	SyncIssues(ctx context.Context, projectID string) error
}

// Filter and request types

// ProjectFilter represents filters for project queries
type ProjectFilter struct {
	Status        []ProjectStatus `json:"status,omitempty"`
	Owner         string          `json:"owner,omitempty"`
	TeamMember    string          `json:"team_member,omitempty"`
	Tags          []string        `json:"tags,omitempty"`
	CreatedAfter  *time.Time      `json:"created_after,omitempty"`
	CreatedBefore *time.Time      `json:"created_before,omitempty"`
	Search        string          `json:"search,omitempty"`
	Limit         int             `json:"limit,omitempty"`
	Offset        int             `json:"offset,omitempty"`
}

// TaskFilter represents filters for task queries
type TaskFilter struct {
	ProjectID   string       `json:"project_id,omitempty"`
	Status      []TaskStatus `json:"status,omitempty"`
	Assignee    string       `json:"assignee,omitempty"`
	Reporter    string       `json:"reporter,omitempty"`
	Priority    []Priority   `json:"priority,omitempty"`
	Type        []TaskType   `json:"type,omitempty"`
	Labels      []string     `json:"labels,omitempty"`
	MilestoneID string       `json:"milestone_id,omitempty"`
	SprintID    string       `json:"sprint_id,omitempty"`
	DueBefore   *time.Time   `json:"due_before,omitempty"`
	DueAfter    *time.Time   `json:"due_after,omitempty"`
	Search      string       `json:"search,omitempty"`
	Limit       int          `json:"limit,omitempty"`
	Offset      int          `json:"offset,omitempty"`
}

// TimeRange represents a time range for analytics
type TimeRange struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// Analytics types

// ProjectAnalytics represents project analytics data
type ProjectAnalytics struct {
	ProjectID            string                 `json:"project_id"`
	TimeRange            *TimeRange             `json:"time_range"`
	TasksCreated         int                    `json:"tasks_created"`
	TasksCompleted       int                    `json:"tasks_completed"`
	TasksInProgress      int                    `json:"tasks_in_progress"`
	AverageLeadTime      time.Duration          `json:"average_lead_time"`
	AverageCycleTime     time.Duration          `json:"average_cycle_time"`
	Velocity             float32                `json:"velocity"`
	BurndownData         []*BurndownPoint       `json:"burndown_data"`
	StatusDistribution   map[TaskStatus]int     `json:"status_distribution"`
	PriorityDistribution map[Priority]int       `json:"priority_distribution"`
	Metadata             map[string]interface{} `json:"metadata,omitempty"`
}

// TaskAnalytics represents task analytics data
type TaskAnalytics struct {
	ProjectID             string                 `json:"project_id"`
	TimeRange             *TimeRange             `json:"time_range"`
	CompletionRate        float32                `json:"completion_rate"`
	AverageTimeToComplete time.Duration          `json:"average_time_to_complete"`
	TasksByAssignee       map[string]int         `json:"tasks_by_assignee"`
	TasksByType           map[TaskType]int       `json:"tasks_by_type"`
	TasksByPriority       map[Priority]int       `json:"tasks_by_priority"`
	OverdueTasks          int                    `json:"overdue_tasks"`
	Metadata              map[string]interface{} `json:"metadata,omitempty"`
}

// TeamAnalytics represents team analytics data
type TeamAnalytics struct {
	ProjectID            string                 `json:"project_id"`
	TimeRange            *TimeRange             `json:"time_range"`
	TeamSize             int                    `json:"team_size"`
	ActiveMembers        int                    `json:"active_members"`
	WorkloadDistribution map[string]float32     `json:"workload_distribution"`
	ProductivityMetrics  *ProductivityMetrics   `json:"productivity_metrics"`
	CollaborationMetrics *CollaborationMetrics  `json:"collaboration_metrics"`
	Metadata             map[string]interface{} `json:"metadata,omitempty"`
}

// BurndownPoint represents a point in burndown chart
type BurndownPoint struct {
	Date      time.Time `json:"date"`
	Remaining int       `json:"remaining"`
	Ideal     int       `json:"ideal"`
}

// ProductivityMetrics represents team productivity metrics
type ProductivityMetrics struct {
	TasksPerMember      float32       `json:"tasks_per_member"`
	AverageTaskDuration time.Duration `json:"average_task_duration"`
	CodeCommits         int           `json:"code_commits"`
	CodeReviews         int           `json:"code_reviews"`
	BugsFixed           int           `json:"bugs_fixed"`
}

// CollaborationMetrics represents team collaboration metrics
type CollaborationMetrics struct {
	CommentsPerTask     float32 `json:"comments_per_task"`
	ReviewParticipation float32 `json:"review_participation"`
	KnowledgeSharing    float32 `json:"knowledge_sharing"`
	CrossTeamWork       float32 `json:"cross_team_work"`
}

// Report types

// ReportRequest represents a report generation request
type ReportRequest struct {
	Type      ReportType             `json:"type"`
	ProjectID string                 `json:"project_id"`
	TimeRange *TimeRange             `json:"time_range"`
	Format    ReportFormat           `json:"format"`
	Filters   map[string]interface{} `json:"filters,omitempty"`
	Options   map[string]interface{} `json:"options,omitempty"`
}

// Report represents a generated report
type Report struct {
	ID          string                 `json:"id"`
	Type        ReportType             `json:"type"`
	Title       string                 `json:"title"`
	Content     string                 `json:"content"`
	Format      ReportFormat           `json:"format"`
	GeneratedAt time.Time              `json:"generated_at"`
	GeneratedBy string                 `json:"generated_by"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

type ReportType string

const (
	ReportTypeProject  ReportType = "project"
	ReportTypeTask     ReportType = "task"
	ReportTypeTeam     ReportType = "team"
	ReportTypeTime     ReportType = "time"
	ReportTypeVelocity ReportType = "velocity"
	ReportTypeBurndown ReportType = "burndown"
	ReportTypeCustom   ReportType = "custom"
)

type ReportFormat string

const (
	ReportFormatHTML ReportFormat = "html"
	ReportFormatPDF  ReportFormat = "pdf"
	ReportFormatCSV  ReportFormat = "csv"
	ReportFormatJSON ReportFormat = "json"
)
