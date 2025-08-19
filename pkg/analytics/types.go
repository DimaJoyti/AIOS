package analytics

import (
	"time"
)

// Analytics and reporting types for AIOS

// AnalyticsEngine defines the interface for analytics operations
type AnalyticsEngine interface {
	// Project Analytics
	GetProjectMetrics(projectID string, timeRange *TimeRange) (*ProjectMetrics, error)
	GetProjectTrends(projectID string, timeRange *TimeRange, granularity Granularity) ([]*TrendPoint, error)
	GetProjectComparison(projectIDs []string, timeRange *TimeRange) (*ProjectComparison, error)

	// Team Analytics
	GetTeamMetrics(teamID string, timeRange *TimeRange) (*TeamMetrics, error)
	GetTeamProductivity(teamID string, timeRange *TimeRange) (*ProductivityMetrics, error)
	GetTeamWorkload(teamID string, timeRange *TimeRange) (*WorkloadMetrics, error)

	// Task Analytics
	GetTaskMetrics(projectID string, timeRange *TimeRange) (*TaskMetrics, error)
	GetTaskDistribution(projectID string, timeRange *TimeRange) (*TaskDistribution, error)
	GetTaskVelocity(projectID string, timeRange *TimeRange) (*VelocityMetrics, error)

	// Workflow Analytics
	GetWorkflowMetrics(workflowID string, timeRange *TimeRange) (*WorkflowMetrics, error)
	GetPipelineMetrics(pipelineID string, timeRange *TimeRange) (*PipelineMetrics, error)
	GetAutomationMetrics(timeRange *TimeRange) (*AutomationMetrics, error)

	// Reports
	GenerateReport(request *ReportRequest) (*Report, error)
	GetReportTemplates() ([]*ReportTemplate, error)
	ScheduleReport(schedule *ReportSchedule) error

	// Dashboards
	GetDashboard(dashboardID string) (*Dashboard, error)
	CreateDashboard(dashboard *Dashboard) (*Dashboard, error)
	UpdateDashboard(dashboard *Dashboard) (*Dashboard, error)
	ListDashboards(filter *DashboardFilter) ([]*Dashboard, error)
}

// Core Analytics Types

// TimeRange represents a time period for analytics
type TimeRange struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// Granularity defines the time granularity for trend analysis
type Granularity string

const (
	GranularityHour  Granularity = "hour"
	GranularityDay   Granularity = "day"
	GranularityWeek  Granularity = "week"
	GranularityMonth Granularity = "month"
	GranularityYear  Granularity = "year"
)

// TrendPoint represents a single point in a trend analysis
type TrendPoint struct {
	Timestamp time.Time              `json:"timestamp"`
	Value     float64                `json:"value"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// Project Analytics Types

// ProjectMetrics contains comprehensive project metrics
type ProjectMetrics struct {
	ProjectID       string                 `json:"project_id"`
	TimeRange       *TimeRange             `json:"time_range"`
	TasksTotal      int                    `json:"tasks_total"`
	TasksCompleted  int                    `json:"tasks_completed"`
	TasksInProgress int                    `json:"tasks_in_progress"`
	TasksOverdue    int                    `json:"tasks_overdue"`
	CompletionRate  float64                `json:"completion_rate"`
	Velocity        float64                `json:"velocity"`
	BurndownData    []*BurndownPoint       `json:"burndown_data"`
	SprintMetrics   []*SprintMetrics       `json:"sprint_metrics"`
	MilestoneStatus []*MilestoneStatus     `json:"milestone_status"`
	QualityMetrics  *QualityMetrics        `json:"quality_metrics"`
	TimeMetrics     *TimeMetrics           `json:"time_metrics"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

// BurndownPoint represents a point in the burndown chart
type BurndownPoint struct {
	Date      time.Time `json:"date"`
	Remaining int       `json:"remaining"`
	Ideal     int       `json:"ideal"`
	Actual    int       `json:"actual"`
}

// SprintMetrics contains sprint-specific metrics
type SprintMetrics struct {
	SprintID         string    `json:"sprint_id"`
	SprintName       string    `json:"sprint_name"`
	StartDate        time.Time `json:"start_date"`
	EndDate          time.Time `json:"end_date"`
	PlannedPoints    int       `json:"planned_points"`
	CompletedPoints  int       `json:"completed_points"`
	Velocity         float64   `json:"velocity"`
	CompletionRate   float64   `json:"completion_rate"`
	TasksCompleted   int       `json:"tasks_completed"`
	TasksCarriedOver int       `json:"tasks_carried_over"`
}

// MilestoneStatus represents milestone progress
type MilestoneStatus struct {
	MilestoneID    string     `json:"milestone_id"`
	MilestoneName  string     `json:"milestone_name"`
	DueDate        *time.Time `json:"due_date"`
	CompletionRate float64    `json:"completion_rate"`
	TasksTotal     int        `json:"tasks_total"`
	TasksCompleted int        `json:"tasks_completed"`
	IsOverdue      bool       `json:"is_overdue"`
	DaysRemaining  int        `json:"days_remaining"`
}

// QualityMetrics contains code and project quality metrics
type QualityMetrics struct {
	BugCount         int     `json:"bug_count"`
	BugDensity       float64 `json:"bug_density"`
	DefectRate       float64 `json:"defect_rate"`
	TestCoverage     float64 `json:"test_coverage"`
	CodeReviewRate   float64 `json:"code_review_rate"`
	TechnicalDebt    float64 `json:"technical_debt"`
	SecurityIssues   int     `json:"security_issues"`
	PerformanceScore float64 `json:"performance_score"`
}

// TimeMetrics contains time-related metrics
type TimeMetrics struct {
	TotalTimeLogged    time.Duration `json:"total_time_logged"`
	AverageTaskTime    time.Duration `json:"average_task_time"`
	EstimationAccuracy float64       `json:"estimation_accuracy"`
	TimeToCompletion   time.Duration `json:"time_to_completion"`
	BillableHours      time.Duration `json:"billable_hours"`
	NonBillableHours   time.Duration `json:"non_billable_hours"`
	UtilizationRate    float64       `json:"utilization_rate"`
}

// ProjectComparison contains comparison metrics between projects
type ProjectComparison struct {
	Projects    []string               `json:"projects"`
	TimeRange   *TimeRange             `json:"time_range"`
	Metrics     []*ProjectMetrics      `json:"metrics"`
	Comparisons map[string]interface{} `json:"comparisons"`
}

// Team Analytics Types

// TeamMetrics contains team performance metrics
type TeamMetrics struct {
	TeamID          string                 `json:"team_id"`
	TimeRange       *TimeRange             `json:"time_range"`
	MemberCount     int                    `json:"member_count"`
	ActiveMembers   int                    `json:"active_members"`
	TasksCompleted  int                    `json:"tasks_completed"`
	AverageVelocity float64                `json:"average_velocity"`
	Productivity    *ProductivityMetrics   `json:"productivity"`
	Collaboration   *CollaborationMetrics  `json:"collaboration"`
	Workload        *WorkloadMetrics       `json:"workload"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

// ProductivityMetrics contains productivity measurements
type ProductivityMetrics struct {
	TasksPerMember      float64       `json:"tasks_per_member"`
	AverageTaskDuration time.Duration `json:"average_task_duration"`
	CodeCommits         int           `json:"code_commits"`
	CodeReviews         int           `json:"code_reviews"`
	BugsFixed           int           `json:"bugs_fixed"`
	FeaturesDelivered   int           `json:"features_delivered"`
	ProductivityScore   float64       `json:"productivity_score"`
	TrendDirection      string        `json:"trend_direction"`
}

// CollaborationMetrics contains team collaboration measurements
type CollaborationMetrics struct {
	CommentsPerTask     float64 `json:"comments_per_task"`
	ReviewParticipation float64 `json:"review_participation"`
	KnowledgeSharing    float64 `json:"knowledge_sharing"`
	CrossTeamWork       float64 `json:"cross_team_work"`
	MeetingEfficiency   float64 `json:"meeting_efficiency"`
	CommunicationScore  float64 `json:"communication_score"`
}

// WorkloadMetrics contains workload distribution metrics
type WorkloadMetrics struct {
	TotalCapacity        float64            `json:"total_capacity"`
	UtilizedCapacity     float64            `json:"utilized_capacity"`
	UtilizationRate      float64            `json:"utilization_rate"`
	WorkloadDistribution map[string]float64 `json:"workload_distribution"`
	OverloadedMembers    []string           `json:"overloaded_members"`
	UnderutilizedMembers []string           `json:"underutilized_members"`
	BalanceScore         float64            `json:"balance_score"`
}

// Task Analytics Types

// TaskMetrics contains task-related metrics
type TaskMetrics struct {
	ProjectID        string                 `json:"project_id"`
	TimeRange        *TimeRange             `json:"time_range"`
	TotalTasks       int                    `json:"total_tasks"`
	CompletedTasks   int                    `json:"completed_tasks"`
	CompletionRate   float64                `json:"completion_rate"`
	AverageLeadTime  time.Duration          `json:"average_lead_time"`
	AverageCycleTime time.Duration          `json:"average_cycle_time"`
	TaskDistribution *TaskDistribution      `json:"task_distribution"`
	VelocityMetrics  *VelocityMetrics       `json:"velocity_metrics"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
}

// TaskDistribution contains task distribution by various dimensions
type TaskDistribution struct {
	ByType     map[string]int `json:"by_type"`
	ByPriority map[string]int `json:"by_priority"`
	ByStatus   map[string]int `json:"by_status"`
	ByAssignee map[string]int `json:"by_assignee"`
	ByLabel    map[string]int `json:"by_label"`
}

// VelocityMetrics contains velocity and throughput metrics
type VelocityMetrics struct {
	CurrentVelocity   float64       `json:"current_velocity"`
	AverageVelocity   float64       `json:"average_velocity"`
	VelocityTrend     []*TrendPoint `json:"velocity_trend"`
	Throughput        float64       `json:"throughput"`
	LeadTimeP50       time.Duration `json:"lead_time_p50"`
	LeadTimeP95       time.Duration `json:"lead_time_p95"`
	CycleTimeP50      time.Duration `json:"cycle_time_p50"`
	CycleTimeP95      time.Duration `json:"cycle_time_p95"`
	PredictedVelocity float64       `json:"predicted_velocity"`
}

// Workflow Analytics Types

// WorkflowMetrics contains workflow execution metrics
type WorkflowMetrics struct {
	WorkflowID         string                 `json:"workflow_id"`
	TimeRange          *TimeRange             `json:"time_range"`
	TotalExecutions    int                    `json:"total_executions"`
	SuccessfulRuns     int                    `json:"successful_runs"`
	FailedRuns         int                    `json:"failed_runs"`
	AverageRunTime     time.Duration          `json:"average_run_time"`
	SuccessRate        float64                `json:"success_rate"`
	ExecutionTrends    []*TrendPoint          `json:"execution_trends"`
	ErrorDistribution  map[string]int         `json:"error_distribution"`
	PerformanceMetrics *PerformanceMetrics    `json:"performance_metrics"`
	Metadata           map[string]interface{} `json:"metadata,omitempty"`
}

// PipelineMetrics contains CI/CD pipeline metrics
type PipelineMetrics struct {
	PipelineID       string                 `json:"pipeline_id"`
	TimeRange        *TimeRange             `json:"time_range"`
	TotalBuilds      int                    `json:"total_builds"`
	SuccessfulBuilds int                    `json:"successful_builds"`
	FailedBuilds     int                    `json:"failed_builds"`
	AverageBuildTime time.Duration          `json:"average_build_time"`
	SuccessRate      float64                `json:"success_rate"`
	BuildTrends      []*TrendPoint          `json:"build_trends"`
	FailureReasons   map[string]int         `json:"failure_reasons"`
	StageMetrics     []*StageMetrics        `json:"stage_metrics"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
}

// StageMetrics contains pipeline stage metrics
type StageMetrics struct {
	StageName        string         `json:"stage_name"`
	AverageTime      time.Duration  `json:"average_time"`
	SuccessRate      float64        `json:"success_rate"`
	FailureReasons   map[string]int `json:"failure_reasons"`
	PerformanceTrend []*TrendPoint  `json:"performance_trend"`
}

// AutomationMetrics contains overall automation metrics
type AutomationMetrics struct {
	TimeRange         *TimeRange             `json:"time_range"`
	TotalWorkflows    int                    `json:"total_workflows"`
	ActiveWorkflows   int                    `json:"active_workflows"`
	TotalExecutions   int                    `json:"total_executions"`
	AutomationSavings time.Duration          `json:"automation_savings"`
	EfficiencyGain    float64                `json:"efficiency_gain"`
	ROI               float64                `json:"roi"`
	AdoptionRate      float64                `json:"adoption_rate"`
	Metadata          map[string]interface{} `json:"metadata,omitempty"`
}

// PerformanceMetrics contains performance-related metrics
type PerformanceMetrics struct {
	AverageResponseTime time.Duration `json:"average_response_time"`
	P50ResponseTime     time.Duration `json:"p50_response_time"`
	P95ResponseTime     time.Duration `json:"p95_response_time"`
	P99ResponseTime     time.Duration `json:"p99_response_time"`
	ThroughputPerSecond float64       `json:"throughput_per_second"`
	ErrorRate           float64       `json:"error_rate"`
	AvailabilityRate    float64       `json:"availability_rate"`
}

// Report Types

// ReportRequest represents a request for generating a report
type ReportRequest struct {
	Type        ReportType             `json:"type"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	TimeRange   *TimeRange             `json:"time_range"`
	Filters     map[string]interface{} `json:"filters,omitempty"`
	Format      ReportFormat           `json:"format"`
	Recipients  []string               `json:"recipients,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// ReportType defines the type of report
type ReportType string

const (
	ReportTypeProject     ReportType = "project"
	ReportTypeTeam        ReportType = "team"
	ReportTypeTask        ReportType = "task"
	ReportTypeWorkflow    ReportType = "workflow"
	ReportTypePipeline    ReportType = "pipeline"
	ReportTypeCustom      ReportType = "custom"
	ReportTypeExecutive   ReportType = "executive"
	ReportTypeOperational ReportType = "operational"
)

// ReportFormat defines the output format of reports
type ReportFormat string

const (
	ReportFormatPDF   ReportFormat = "pdf"
	ReportFormatHTML  ReportFormat = "html"
	ReportFormatJSON  ReportFormat = "json"
	ReportFormatCSV   ReportFormat = "csv"
	ReportFormatExcel ReportFormat = "excel"
)

// Report represents a generated report
type Report struct {
	ID          string                 `json:"id"`
	Type        ReportType             `json:"type"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Format      ReportFormat           `json:"format"`
	Content     string                 `json:"content"`
	Data        map[string]interface{} `json:"data,omitempty"`
	GeneratedAt time.Time              `json:"generated_at"`
	GeneratedBy string                 `json:"generated_by"`
	Size        int64                  `json:"size"`
	URL         string                 `json:"url,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// ReportTemplate represents a reusable report template
type ReportTemplate struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Type        ReportType             `json:"type"`
	Template    string                 `json:"template"`
	Parameters  map[string]interface{} `json:"parameters,omitempty"`
	IsPublic    bool                   `json:"is_public"`
	CreatedBy   string                 `json:"created_by"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// ReportSchedule represents a scheduled report
type ReportSchedule struct {
	ID         string                 `json:"id"`
	Name       string                 `json:"name"`
	ReportType ReportType             `json:"report_type"`
	Schedule   string                 `json:"schedule"` // Cron expression
	Recipients []string               `json:"recipients"`
	Filters    map[string]interface{} `json:"filters,omitempty"`
	Format     ReportFormat           `json:"format"`
	Enabled    bool                   `json:"enabled"`
	LastRun    *time.Time             `json:"last_run,omitempty"`
	NextRun    time.Time              `json:"next_run"`
	CreatedBy  string                 `json:"created_by"`
	CreatedAt  time.Time              `json:"created_at"`
	UpdatedAt  time.Time              `json:"updated_at"`
}

// Dashboard Types

// Dashboard represents an analytics dashboard
type Dashboard struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Layout      *DashboardLayout       `json:"layout"`
	Widgets     []*Widget              `json:"widgets"`
	Filters     []*Filter              `json:"filters,omitempty"`
	IsPublic    bool                   `json:"is_public"`
	CreatedBy   string                 `json:"created_by"`
	SharedWith  []string               `json:"shared_with,omitempty"`
	Tags        []string               `json:"tags,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// DashboardLayout defines the layout configuration
type DashboardLayout struct {
	Type    string `json:"type"` // grid, flex, custom
	Columns int    `json:"columns"`
	Rows    int    `json:"rows"`
	Gap     int    `json:"gap"`
}

// Widget represents a dashboard widget
type Widget struct {
	ID          string                 `json:"id"`
	Type        WidgetType             `json:"type"`
	Title       string                 `json:"title"`
	Description string                 `json:"description,omitempty"`
	Position    *WidgetPosition        `json:"position"`
	Size        *WidgetSize            `json:"size"`
	Config      map[string]interface{} `json:"config"`
	DataSource  *DataSource            `json:"data_source"`
	RefreshRate int                    `json:"refresh_rate"` // seconds
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// WidgetType defines the type of widget
type WidgetType string

const (
	WidgetTypeChart    WidgetType = "chart"
	WidgetTypeTable    WidgetType = "table"
	WidgetTypeMetric   WidgetType = "metric"
	WidgetTypeProgress WidgetType = "progress"
	WidgetTypeList     WidgetType = "list"
	WidgetTypeCalendar WidgetType = "calendar"
	WidgetTypeKanban   WidgetType = "kanban"
	WidgetTypeCustom   WidgetType = "custom"
)

// WidgetPosition defines widget position in the dashboard
type WidgetPosition struct {
	X int `json:"x"`
	Y int `json:"y"`
}

// WidgetSize defines widget dimensions
type WidgetSize struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

// DataSource defines the data source for a widget
type DataSource struct {
	Type       string                 `json:"type"`
	Query      string                 `json:"query"`
	Parameters map[string]interface{} `json:"parameters,omitempty"`
	CacheTime  int                    `json:"cache_time"` // seconds
}

// Filter represents a dashboard filter
type Filter struct {
	ID           string      `json:"id"`
	Name         string      `json:"name"`
	Type         FilterType  `json:"type"`
	Field        string      `json:"field"`
	Options      []string    `json:"options,omitempty"`
	DefaultValue interface{} `json:"default_value,omitempty"`
	Required     bool        `json:"required"`
}

// FilterType defines the type of filter
type FilterType string

const (
	FilterTypeSelect      FilterType = "select"
	FilterTypeMultiSelect FilterType = "multi_select"
	FilterTypeDate        FilterType = "date"
	FilterTypeDateRange   FilterType = "date_range"
	FilterTypeText        FilterType = "text"
	FilterTypeNumber      FilterType = "number"
	FilterTypeBoolean     FilterType = "boolean"
)

// DashboardFilter represents filters for dashboard queries
type DashboardFilter struct {
	CreatedBy string   `json:"created_by,omitempty"`
	IsPublic  *bool    `json:"is_public,omitempty"`
	Tags      []string `json:"tags,omitempty"`
	Search    string   `json:"search,omitempty"`
	Limit     int      `json:"limit,omitempty"`
	Offset    int      `json:"offset,omitempty"`
}
