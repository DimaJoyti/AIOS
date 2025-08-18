package analytics

import (
	"context"
	"fmt"
	"math"
	"sort"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// DefaultAnalyticsEngine implements the AnalyticsEngine interface
type DefaultAnalyticsEngine struct {
	logger *logrus.Logger
	tracer trace.Tracer

	// Data sources
	projectManager interface{} // Will be injected
	workflowEngine interface{} // Will be injected
	pipelineEngine interface{} // Will be injected

	// In-memory storage for analytics data
	dashboards      map[string]*Dashboard
	reportTemplates map[string]*ReportTemplate
	reportSchedules map[string]*ReportSchedule

	// Cache for computed metrics
	metricsCache map[string]*CachedMetric
	cacheTTL     time.Duration

	mu sync.RWMutex
}

// CachedMetric represents a cached analytics metric
type CachedMetric struct {
	Key       string
	Data      interface{}
	ExpiresAt time.Time
}

// AnalyticsEngineConfig represents configuration for the analytics engine
type AnalyticsEngineConfig struct {
	CacheTTL        time.Duration `json:"cache_ttl"`
	MaxCacheSize    int           `json:"max_cache_size"`
	EnableRealTime  bool          `json:"enable_real_time"`
	ComputeInterval time.Duration `json:"compute_interval"`
	RetentionPeriod time.Duration `json:"retention_period"`
}

// NewDefaultAnalyticsEngine creates a new analytics engine
func NewDefaultAnalyticsEngine(config *AnalyticsEngineConfig, logger *logrus.Logger) AnalyticsEngine {
	if config == nil {
		config = &AnalyticsEngineConfig{
			CacheTTL:        15 * time.Minute,
			MaxCacheSize:    1000,
			EnableRealTime:  true,
			ComputeInterval: 5 * time.Minute,
			RetentionPeriod: 90 * 24 * time.Hour, // 90 days
		}
	}

	engine := &DefaultAnalyticsEngine{
		logger:          logger,
		tracer:          otel.Tracer("analytics.engine"),
		dashboards:      make(map[string]*Dashboard),
		reportTemplates: make(map[string]*ReportTemplate),
		reportSchedules: make(map[string]*ReportSchedule),
		metricsCache:    make(map[string]*CachedMetric),
		cacheTTL:        config.CacheTTL,
	}

	// Create default dashboards and templates
	engine.createDefaultDashboards()
	engine.createDefaultReportTemplates()

	return engine
}

// Project Analytics

// GetProjectMetrics retrieves comprehensive project metrics
func (ae *DefaultAnalyticsEngine) GetProjectMetrics(projectID string, timeRange *TimeRange) (*ProjectMetrics, error) {
	_, span := ae.tracer.Start(context.Background(), "analytics.get_project_metrics")
	defer span.End()

	span.SetAttributes(
		attribute.String("project.id", projectID),
		attribute.String("time_range.start", timeRange.Start.Format(time.RFC3339)),
		attribute.String("time_range.end", timeRange.End.Format(time.RFC3339)),
	)

	// Check cache first
	cacheKey := fmt.Sprintf("project_metrics_%s_%d_%d", projectID, timeRange.Start.Unix(), timeRange.End.Unix())
	if cached := ae.getCachedMetric(cacheKey); cached != nil {
		if metrics, ok := cached.(*ProjectMetrics); ok {
			return metrics, nil
		}
	}

	// Simulate project metrics calculation
	// In a real implementation, this would query the actual data sources
	metrics := &ProjectMetrics{
		ProjectID:       projectID,
		TimeRange:       timeRange,
		TasksTotal:      150,
		TasksCompleted:  120,
		TasksInProgress: 25,
		TasksOverdue:    5,
		CompletionRate:  0.8,
		Velocity:        12.5,
		BurndownData:    ae.generateBurndownData(timeRange),
		SprintMetrics:   ae.generateSprintMetrics(projectID, timeRange),
		MilestoneStatus: ae.generateMilestoneStatus(projectID),
		QualityMetrics: &QualityMetrics{
			BugCount:         8,
			BugDensity:       0.05,
			DefectRate:       0.03,
			TestCoverage:     85.5,
			CodeReviewRate:   95.0,
			TechnicalDebt:    15.2,
			SecurityIssues:   2,
			PerformanceScore: 92.0,
		},
		TimeMetrics: &TimeMetrics{
			TotalTimeLogged:    480 * time.Hour,
			AverageTaskTime:    4 * time.Hour,
			EstimationAccuracy: 0.85,
			TimeToCompletion:   2 * 24 * time.Hour,
			BillableHours:      400 * time.Hour,
			NonBillableHours:   80 * time.Hour,
			UtilizationRate:    0.83,
		},
		Metadata: map[string]interface{}{
			"generated_at": time.Now(),
			"version":      "1.0",
		},
	}

	// Cache the result
	ae.setCachedMetric(cacheKey, metrics)

	ae.logger.WithFields(logrus.Fields{
		"project_id":      projectID,
		"completion_rate": metrics.CompletionRate,
		"velocity":        metrics.Velocity,
	}).Info("Project metrics calculated")

	return metrics, nil
}

// GetProjectTrends retrieves project trend data
func (ae *DefaultAnalyticsEngine) GetProjectTrends(projectID string, timeRange *TimeRange, granularity Granularity) ([]*TrendPoint, error) {
	_, span := ae.tracer.Start(context.Background(), "analytics.get_project_trends")
	defer span.End()

	// Generate trend data based on granularity
	var points []*TrendPoint
	interval := ae.getIntervalForGranularity(granularity)

	for current := timeRange.Start; current.Before(timeRange.End); current = current.Add(interval) {
		// Simulate trend data with some variation
		baseValue := 75.0
		variation := math.Sin(float64(current.Unix())/86400) * 10 // Daily variation
		value := baseValue + variation + float64(len(points))*0.5 // Slight upward trend

		points = append(points, &TrendPoint{
			Timestamp: current,
			Value:     value,
			Metadata: map[string]interface{}{
				"tasks_completed": 10 + len(points),
				"velocity":        value / 10,
			},
		})
	}

	return points, nil
}

// GetProjectComparison compares multiple projects
func (ae *DefaultAnalyticsEngine) GetProjectComparison(projectIDs []string, timeRange *TimeRange) (*ProjectComparison, error) {
	var metrics []*ProjectMetrics

	for _, projectID := range projectIDs {
		projectMetrics, err := ae.GetProjectMetrics(projectID, timeRange)
		if err != nil {
			ae.logger.WithError(err).WithField("project_id", projectID).Warn("Failed to get project metrics for comparison")
			continue
		}
		metrics = append(metrics, projectMetrics)
	}

	// Calculate comparison metrics
	comparisons := make(map[string]interface{})
	if len(metrics) > 0 {
		var totalVelocity, totalCompletion float64
		for _, m := range metrics {
			totalVelocity += m.Velocity
			totalCompletion += m.CompletionRate
		}

		comparisons["average_velocity"] = totalVelocity / float64(len(metrics))
		comparisons["average_completion_rate"] = totalCompletion / float64(len(metrics))

		// Find best and worst performing projects
		sort.Slice(metrics, func(i, j int) bool {
			return metrics[i].Velocity > metrics[j].Velocity
		})

		if len(metrics) > 0 {
			comparisons["best_velocity"] = metrics[0].ProjectID
			comparisons["worst_velocity"] = metrics[len(metrics)-1].ProjectID
		}
	}

	return &ProjectComparison{
		Projects:    projectIDs,
		TimeRange:   timeRange,
		Metrics:     metrics,
		Comparisons: comparisons,
	}, nil
}

// Team Analytics

// GetTeamMetrics retrieves team performance metrics
func (ae *DefaultAnalyticsEngine) GetTeamMetrics(teamID string, timeRange *TimeRange) (*TeamMetrics, error) {
	_, span := ae.tracer.Start(context.Background(), "analytics.get_team_metrics")
	defer span.End()

	span.SetAttributes(attribute.String("team.id", teamID))

	// Simulate team metrics
	metrics := &TeamMetrics{
		TeamID:          teamID,
		TimeRange:       timeRange,
		MemberCount:     8,
		ActiveMembers:   7,
		TasksCompleted:  95,
		AverageVelocity: 11.8,
		Productivity: &ProductivityMetrics{
			TasksPerMember:      11.9,
			AverageTaskDuration: 6 * time.Hour,
			CodeCommits:         156,
			CodeReviews:         89,
			BugsFixed:           23,
			FeaturesDelivered:   12,
			ProductivityScore:   87.5,
			TrendDirection:      "up",
		},
		Collaboration: &CollaborationMetrics{
			CommentsPerTask:     4.2,
			ReviewParticipation: 0.92,
			KnowledgeSharing:    0.78,
			CrossTeamWork:       0.65,
			MeetingEfficiency:   0.82,
			CommunicationScore:  85.0,
		},
		Workload: &WorkloadMetrics{
			TotalCapacity:    320.0, // hours
			UtilizedCapacity: 285.0,
			UtilizationRate:  0.89,
			WorkloadDistribution: map[string]float64{
				"member1": 0.95,
				"member2": 0.88,
				"member3": 0.92,
				"member4": 0.85,
				"member5": 0.90,
				"member6": 0.87,
				"member7": 0.83,
			},
			OverloadedMembers:    []string{"member1"},
			UnderutilizedMembers: []string{"member7"},
			BalanceScore:         0.78,
		},
		Metadata: map[string]interface{}{
			"team_type": "development",
			"location":  "distributed",
		},
	}

	return metrics, nil
}

// GetTeamProductivity retrieves team productivity metrics
func (ae *DefaultAnalyticsEngine) GetTeamProductivity(teamID string, timeRange *TimeRange) (*ProductivityMetrics, error) {
	teamMetrics, err := ae.GetTeamMetrics(teamID, timeRange)
	if err != nil {
		return nil, err
	}
	return teamMetrics.Productivity, nil
}

// GetTeamWorkload retrieves team workload metrics
func (ae *DefaultAnalyticsEngine) GetTeamWorkload(teamID string, timeRange *TimeRange) (*WorkloadMetrics, error) {
	teamMetrics, err := ae.GetTeamMetrics(teamID, timeRange)
	if err != nil {
		return nil, err
	}
	return teamMetrics.Workload, nil
}

// Task Analytics

// GetTaskMetrics retrieves task-related metrics
func (ae *DefaultAnalyticsEngine) GetTaskMetrics(projectID string, timeRange *TimeRange) (*TaskMetrics, error) {
	_, span := ae.tracer.Start(context.Background(), "analytics.get_task_metrics")
	defer span.End()

	metrics := &TaskMetrics{
		ProjectID:        projectID,
		TimeRange:        timeRange,
		TotalTasks:       150,
		CompletedTasks:   120,
		CompletionRate:   0.8,
		AverageLeadTime:  5 * 24 * time.Hour,
		AverageCycleTime: 3 * 24 * time.Hour,
		TaskDistribution: &TaskDistribution{
			ByType: map[string]int{
				"feature": 65,
				"bug":     25,
				"task":    35,
				"story":   20,
				"epic":    5,
			},
			ByPriority: map[string]int{
				"high":   30,
				"medium": 85,
				"low":    35,
			},
			ByStatus: map[string]int{
				"todo":        30,
				"in_progress": 25,
				"review":      15,
				"done":        120,
			},
			ByAssignee: map[string]int{
				"developer1": 25,
				"developer2": 22,
				"developer3": 28,
				"developer4": 20,
				"developer5": 25,
			},
			ByLabel: map[string]int{
				"frontend": 45,
				"backend":  55,
				"api":      30,
				"ui":       25,
				"testing":  20,
			},
		},
		VelocityMetrics: &VelocityMetrics{
			CurrentVelocity:   12.5,
			AverageVelocity:   11.8,
			VelocityTrend:     ae.generateVelocityTrend(timeRange),
			Throughput:        8.5,
			LeadTimeP50:       4 * 24 * time.Hour,
			LeadTimeP95:       8 * 24 * time.Hour,
			CycleTimeP50:      2 * 24 * time.Hour,
			CycleTimeP95:      5 * 24 * time.Hour,
			PredictedVelocity: 13.2,
		},
		Metadata: map[string]interface{}{
			"methodology":   "scrum",
			"sprint_length": 14,
		},
	}

	return metrics, nil
}

// GetTaskDistribution retrieves task distribution metrics
func (ae *DefaultAnalyticsEngine) GetTaskDistribution(projectID string, timeRange *TimeRange) (*TaskDistribution, error) {
	taskMetrics, err := ae.GetTaskMetrics(projectID, timeRange)
	if err != nil {
		return nil, err
	}
	return taskMetrics.TaskDistribution, nil
}

// GetTaskVelocity retrieves task velocity metrics
func (ae *DefaultAnalyticsEngine) GetTaskVelocity(projectID string, timeRange *TimeRange) (*VelocityMetrics, error) {
	taskMetrics, err := ae.GetTaskMetrics(projectID, timeRange)
	if err != nil {
		return nil, err
	}
	return taskMetrics.VelocityMetrics, nil
}

// Workflow Analytics

// GetWorkflowMetrics retrieves workflow execution metrics
func (ae *DefaultAnalyticsEngine) GetWorkflowMetrics(workflowID string, timeRange *TimeRange) (*WorkflowMetrics, error) {
	_, span := ae.tracer.Start(context.Background(), "analytics.get_workflow_metrics")
	defer span.End()

	span.SetAttributes(attribute.String("workflow.id", workflowID))

	metrics := &WorkflowMetrics{
		WorkflowID:      workflowID,
		TimeRange:       timeRange,
		TotalExecutions: 245,
		SuccessfulRuns:  220,
		FailedRuns:      25,
		AverageRunTime:  8 * time.Minute,
		SuccessRate:     0.898,
		ExecutionTrends: ae.generateExecutionTrends(timeRange),
		ErrorDistribution: map[string]int{
			"timeout":          8,
			"network_error":    5,
			"validation_error": 7,
			"resource_limit":   3,
			"unknown":          2,
		},
		PerformanceMetrics: &PerformanceMetrics{
			AverageResponseTime: 2 * time.Second,
			P50ResponseTime:     1500 * time.Millisecond,
			P95ResponseTime:     5 * time.Second,
			P99ResponseTime:     12 * time.Second,
			ThroughputPerSecond: 15.5,
			ErrorRate:           0.102,
			AvailabilityRate:    0.995,
		},
		Metadata: map[string]interface{}{
			"workflow_type": "ci_cd",
			"complexity":    "medium",
		},
	}

	return metrics, nil
}

// GetPipelineMetrics retrieves CI/CD pipeline metrics
func (ae *DefaultAnalyticsEngine) GetPipelineMetrics(pipelineID string, timeRange *TimeRange) (*PipelineMetrics, error) {
	_, span := ae.tracer.Start(context.Background(), "analytics.get_pipeline_metrics")
	defer span.End()

	span.SetAttributes(attribute.String("pipeline.id", pipelineID))

	metrics := &PipelineMetrics{
		PipelineID:       pipelineID,
		TimeRange:        timeRange,
		TotalBuilds:      156,
		SuccessfulBuilds: 142,
		FailedBuilds:     14,
		AverageBuildTime: 12 * time.Minute,
		SuccessRate:      0.910,
		BuildTrends:      ae.generateBuildTrends(timeRange),
		FailureReasons: map[string]int{
			"test_failure":   6,
			"build_error":    4,
			"timeout":        2,
			"infrastructure": 2,
		},
		StageMetrics: []*StageMetrics{
			{
				StageName:        "Build",
				AverageTime:      3 * time.Minute,
				SuccessRate:      0.98,
				FailureReasons:   map[string]int{"compile_error": 2, "dependency": 1},
				PerformanceTrend: ae.generateStageTrend(timeRange),
			},
			{
				StageName:        "Test",
				AverageTime:      6 * time.Minute,
				SuccessRate:      0.92,
				FailureReasons:   map[string]int{"unit_test": 4, "integration_test": 2},
				PerformanceTrend: ae.generateStageTrend(timeRange),
			},
			{
				StageName:        "Deploy",
				AverageTime:      3 * time.Minute,
				SuccessRate:      0.95,
				FailureReasons:   map[string]int{"deployment_error": 2, "rollback": 1},
				PerformanceTrend: ae.generateStageTrend(timeRange),
			},
		},
		Metadata: map[string]interface{}{
			"pipeline_type": "full_stack",
			"environment":   "production",
		},
	}

	return metrics, nil
}

// GetAutomationMetrics retrieves overall automation metrics
func (ae *DefaultAnalyticsEngine) GetAutomationMetrics(timeRange *TimeRange) (*AutomationMetrics, error) {
	_, span := ae.tracer.Start(context.Background(), "analytics.get_automation_metrics")
	defer span.End()

	metrics := &AutomationMetrics{
		TimeRange:         timeRange,
		TotalWorkflows:    45,
		ActiveWorkflows:   38,
		TotalExecutions:   1250,
		AutomationSavings: 320 * time.Hour,
		EfficiencyGain:    0.35,
		ROI:               4.2,
		AdoptionRate:      0.78,
		Metadata: map[string]interface{}{
			"cost_savings":     "$45,000",
			"time_saved_days":  40,
			"manual_processes": 12,
		},
	}

	return metrics, nil
}

// Helper methods for data generation and caching

// getCachedMetric retrieves a metric from cache if not expired
func (ae *DefaultAnalyticsEngine) getCachedMetric(key string) interface{} {
	ae.mu.RLock()
	defer ae.mu.RUnlock()

	if cached, exists := ae.metricsCache[key]; exists {
		if time.Now().Before(cached.ExpiresAt) {
			return cached.Data
		}
		// Remove expired entry
		delete(ae.metricsCache, key)
	}
	return nil
}

// setCachedMetric stores a metric in cache
func (ae *DefaultAnalyticsEngine) setCachedMetric(key string, data interface{}) {
	ae.mu.Lock()
	defer ae.mu.Unlock()

	ae.metricsCache[key] = &CachedMetric{
		Key:       key,
		Data:      data,
		ExpiresAt: time.Now().Add(ae.cacheTTL),
	}
}

// getIntervalForGranularity returns the time interval for a given granularity
func (ae *DefaultAnalyticsEngine) getIntervalForGranularity(granularity Granularity) time.Duration {
	switch granularity {
	case GranularityHour:
		return time.Hour
	case GranularityDay:
		return 24 * time.Hour
	case GranularityWeek:
		return 7 * 24 * time.Hour
	case GranularityMonth:
		return 30 * 24 * time.Hour
	case GranularityYear:
		return 365 * 24 * time.Hour
	default:
		return 24 * time.Hour
	}
}

// generateBurndownData creates sample burndown chart data
func (ae *DefaultAnalyticsEngine) generateBurndownData(timeRange *TimeRange) []*BurndownPoint {
	var points []*BurndownPoint
	totalTasks := 150
	duration := timeRange.End.Sub(timeRange.Start)
	days := int(duration.Hours() / 24)

	for i := 0; i <= days; i++ {
		date := timeRange.Start.AddDate(0, 0, i)
		remaining := totalTasks - (i * totalTasks / days)
		ideal := totalTasks - (i * totalTasks / days)
		actual := remaining + int(math.Sin(float64(i))*5) // Add some variation

		if actual < 0 {
			actual = 0
		}

		points = append(points, &BurndownPoint{
			Date:      date,
			Remaining: remaining,
			Ideal:     ideal,
			Actual:    actual,
		})
	}

	return points
}

// generateSprintMetrics creates sample sprint metrics
func (ae *DefaultAnalyticsEngine) generateSprintMetrics(projectID string, timeRange *TimeRange) []*SprintMetrics {
	var sprints []*SprintMetrics
	sprintDuration := 14 * 24 * time.Hour // 2 weeks

	for i := 0; i < 3; i++ {
		startDate := timeRange.Start.AddDate(0, 0, i*14)
		endDate := startDate.Add(sprintDuration)

		if endDate.After(timeRange.End) {
			break
		}

		planned := 40 + i*5
		completed := planned - int(math.Abs(math.Sin(float64(i)))*10)

		sprints = append(sprints, &SprintMetrics{
			SprintID:         fmt.Sprintf("sprint-%d", i+1),
			SprintName:       fmt.Sprintf("Sprint %d", i+1),
			StartDate:        startDate,
			EndDate:          endDate,
			PlannedPoints:    planned,
			CompletedPoints:  completed,
			Velocity:         float64(completed) / 2.0, // points per week
			CompletionRate:   float64(completed) / float64(planned),
			TasksCompleted:   completed / 2,
			TasksCarriedOver: (planned - completed) / 3,
		})
	}

	return sprints
}

// generateMilestoneStatus creates sample milestone status
func (ae *DefaultAnalyticsEngine) generateMilestoneStatus(projectID string) []*MilestoneStatus {
	now := time.Now()
	return []*MilestoneStatus{
		{
			MilestoneID:    "milestone-1",
			MilestoneName:  "MVP Release",
			DueDate:        &[]time.Time{now.AddDate(0, 0, 30)}[0],
			CompletionRate: 0.75,
			TasksTotal:     40,
			TasksCompleted: 30,
			IsOverdue:      false,
			DaysRemaining:  30,
		},
		{
			MilestoneID:    "milestone-2",
			MilestoneName:  "Beta Release",
			DueDate:        &[]time.Time{now.AddDate(0, 0, 60)}[0],
			CompletionRate: 0.45,
			TasksTotal:     60,
			TasksCompleted: 27,
			IsOverdue:      false,
			DaysRemaining:  60,
		},
		{
			MilestoneID:    "milestone-3",
			MilestoneName:  "Production Release",
			DueDate:        &[]time.Time{now.AddDate(0, 0, 90)}[0],
			CompletionRate: 0.15,
			TasksTotal:     80,
			TasksCompleted: 12,
			IsOverdue:      false,
			DaysRemaining:  90,
		},
	}
}
