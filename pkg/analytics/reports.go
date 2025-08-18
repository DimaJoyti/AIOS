package analytics

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Report generation and management methods

// GenerateReport generates a report based on the request
func (ae *DefaultAnalyticsEngine) GenerateReport(request *ReportRequest) (*Report, error) {
	_, span := ae.tracer.Start(context.Background(), "analytics.generate_report")
	defer span.End()

	// Generate report content based on type
	var content string
	var data map[string]interface{}

	switch request.Type {
	case ReportTypeProject:
		content, data = ae.generateProjectReport(request)
	case ReportTypeTeam:
		content, data = ae.generateTeamReport(request)
	case ReportTypeTask:
		content, data = ae.generateTaskReport(request)
	case ReportTypeWorkflow:
		content, data = ae.generateWorkflowReport(request)
	case ReportTypePipeline:
		content, data = ae.generatePipelineReport(request)
	case ReportTypeExecutive:
		content, data = ae.generateExecutiveReport(request)
	default:
		return nil, fmt.Errorf("unsupported report type: %s", request.Type)
	}

	report := &Report{
		ID:          uuid.New().String(),
		Type:        request.Type,
		Title:       request.Title,
		Description: request.Description,
		Format:      request.Format,
		Content:     content,
		Data:        data,
		GeneratedAt: time.Now(),
		GeneratedBy: "system", // In production, this would be the actual user
		Size:        int64(len(content)),
		URL:         fmt.Sprintf("/reports/%s", uuid.New().String()),
		Metadata:    request.Metadata,
	}

	ae.logger.WithFields(map[string]interface{}{
		"report_id":   report.ID,
		"report_type": report.Type,
		"format":      report.Format,
		"size":        report.Size,
	}).Info("Report generated successfully")

	return report, nil
}

// GetReportTemplates retrieves available report templates
func (ae *DefaultAnalyticsEngine) GetReportTemplates() ([]*ReportTemplate, error) {
	ae.mu.RLock()
	defer ae.mu.RUnlock()

	var templates []*ReportTemplate
	for _, template := range ae.reportTemplates {
		templates = append(templates, template)
	}

	return templates, nil
}

// ScheduleReport schedules a recurring report
func (ae *DefaultAnalyticsEngine) ScheduleReport(schedule *ReportSchedule) error {
	if schedule.ID == "" {
		schedule.ID = uuid.New().String()
	}

	// Set timestamps
	now := time.Now()
	schedule.CreatedAt = now
	schedule.UpdatedAt = now

	// Calculate next run time based on cron schedule
	// In a real implementation, this would use a proper cron parser
	schedule.NextRun = now.Add(24 * time.Hour) // Default to daily

	ae.mu.Lock()
	ae.reportSchedules[schedule.ID] = schedule
	ae.mu.Unlock()

	ae.logger.WithFields(map[string]interface{}{
		"schedule_id":   schedule.ID,
		"schedule_name": schedule.Name,
		"next_run":      schedule.NextRun,
	}).Info("Report scheduled successfully")

	return nil
}

// Dashboard management methods

// GetDashboard retrieves a dashboard by ID
func (ae *DefaultAnalyticsEngine) GetDashboard(dashboardID string) (*Dashboard, error) {
	ae.mu.RLock()
	defer ae.mu.RUnlock()

	dashboard, exists := ae.dashboards[dashboardID]
	if !exists {
		return nil, fmt.Errorf("dashboard not found: %s", dashboardID)
	}

	return dashboard, nil
}

// CreateDashboard creates a new dashboard
func (ae *DefaultAnalyticsEngine) CreateDashboard(dashboard *Dashboard) (*Dashboard, error) {
	if dashboard.ID == "" {
		dashboard.ID = uuid.New().String()
	}

	// Set timestamps
	now := time.Now()
	dashboard.CreatedAt = now
	dashboard.UpdatedAt = now

	// Validate dashboard
	if err := ae.validateDashboard(dashboard); err != nil {
		return nil, fmt.Errorf("dashboard validation failed: %w", err)
	}

	ae.mu.Lock()
	ae.dashboards[dashboard.ID] = dashboard
	ae.mu.Unlock()

	ae.logger.WithFields(map[string]interface{}{
		"dashboard_id":   dashboard.ID,
		"dashboard_name": dashboard.Name,
		"widgets":        len(dashboard.Widgets),
	}).Info("Dashboard created successfully")

	return dashboard, nil
}

// UpdateDashboard updates an existing dashboard
func (ae *DefaultAnalyticsEngine) UpdateDashboard(dashboard *Dashboard) (*Dashboard, error) {
	ae.mu.Lock()
	existing, exists := ae.dashboards[dashboard.ID]
	if !exists {
		ae.mu.Unlock()
		return nil, fmt.Errorf("dashboard not found: %s", dashboard.ID)
	}

	// Preserve creation info
	dashboard.CreatedAt = existing.CreatedAt
	dashboard.CreatedBy = existing.CreatedBy
	dashboard.UpdatedAt = time.Now()

	// Validate dashboard
	if err := ae.validateDashboard(dashboard); err != nil {
		ae.mu.Unlock()
		return nil, fmt.Errorf("dashboard validation failed: %w", err)
	}

	ae.dashboards[dashboard.ID] = dashboard
	ae.mu.Unlock()

	ae.logger.WithFields(map[string]interface{}{
		"dashboard_id":   dashboard.ID,
		"dashboard_name": dashboard.Name,
	}).Info("Dashboard updated successfully")

	return dashboard, nil
}

// ListDashboards lists dashboards with filtering
func (ae *DefaultAnalyticsEngine) ListDashboards(filter *DashboardFilter) ([]*Dashboard, error) {
	ae.mu.RLock()
	var dashboards []*Dashboard
	for _, dashboard := range ae.dashboards {
		if ae.matchesDashboardFilter(dashboard, filter) {
			dashboards = append(dashboards, dashboard)
		}
	}
	ae.mu.RUnlock()

	// Apply pagination
	if filter != nil {
		if filter.Offset > 0 && filter.Offset < len(dashboards) {
			dashboards = dashboards[filter.Offset:]
		}
		if filter.Limit > 0 && filter.Limit < len(dashboards) {
			dashboards = dashboards[:filter.Limit]
		}
	}

	return dashboards, nil
}

// Helper methods for report generation

// generateProjectReport generates a project report
func (ae *DefaultAnalyticsEngine) generateProjectReport(request *ReportRequest) (string, map[string]interface{}) {
	projectID := ""
	if id, exists := request.Filters["project_id"]; exists {
		projectID = fmt.Sprintf("%v", id)
	}

	metrics, _ := ae.GetProjectMetrics(projectID, request.TimeRange)

	content := fmt.Sprintf(`
# Project Report: %s

## Summary
- **Project ID**: %s
- **Time Range**: %s to %s
- **Completion Rate**: %.1f%%
- **Velocity**: %.1f points/week

## Task Metrics
- **Total Tasks**: %d
- **Completed Tasks**: %d
- **In Progress**: %d
- **Overdue**: %d

## Quality Metrics
- **Bug Count**: %d
- **Test Coverage**: %.1f%%
- **Code Review Rate**: %.1f%%

## Time Metrics
- **Total Time Logged**: %s
- **Average Task Time**: %s
- **Utilization Rate**: %.1f%%
`,
		request.Title,
		metrics.ProjectID,
		request.TimeRange.Start.Format("2006-01-02"),
		request.TimeRange.End.Format("2006-01-02"),
		metrics.CompletionRate*100,
		metrics.Velocity,
		metrics.TasksTotal,
		metrics.TasksCompleted,
		metrics.TasksInProgress,
		metrics.TasksOverdue,
		metrics.QualityMetrics.BugCount,
		metrics.QualityMetrics.TestCoverage,
		metrics.QualityMetrics.CodeReviewRate,
		metrics.TimeMetrics.TotalTimeLogged,
		metrics.TimeMetrics.AverageTaskTime,
		metrics.TimeMetrics.UtilizationRate*100,
	)

	data := map[string]interface{}{
		"metrics": metrics,
		"charts": map[string]interface{}{
			"burndown":   metrics.BurndownData,
			"sprints":    metrics.SprintMetrics,
			"milestones": metrics.MilestoneStatus,
		},
	}

	return content, data
}

// generateTeamReport generates a team report
func (ae *DefaultAnalyticsEngine) generateTeamReport(request *ReportRequest) (string, map[string]interface{}) {
	teamID := ""
	if id, exists := request.Filters["team_id"]; exists {
		teamID = fmt.Sprintf("%v", id)
	}

	metrics, _ := ae.GetTeamMetrics(teamID, request.TimeRange)

	content := fmt.Sprintf(`
# Team Report: %s

## Team Overview
- **Team ID**: %s
- **Time Range**: %s to %s
- **Team Members**: %d
- **Active Members**: %d

## Productivity Metrics
- **Tasks Completed**: %d
- **Average Velocity**: %.1f
- **Tasks per Member**: %.1f
- **Productivity Score**: %.1f

## Collaboration Metrics
- **Comments per Task**: %.1f
- **Review Participation**: %.1f%%
- **Communication Score**: %.1f

## Workload Metrics
- **Utilization Rate**: %.1f%%
- **Balance Score**: %.1f
- **Overloaded Members**: %d
- **Underutilized Members**: %d
`,
		request.Title,
		metrics.TeamID,
		request.TimeRange.Start.Format("2006-01-02"),
		request.TimeRange.End.Format("2006-01-02"),
		metrics.MemberCount,
		metrics.ActiveMembers,
		metrics.TasksCompleted,
		metrics.AverageVelocity,
		metrics.Productivity.TasksPerMember,
		metrics.Productivity.ProductivityScore,
		metrics.Collaboration.CommentsPerTask,
		metrics.Collaboration.ReviewParticipation*100,
		metrics.Collaboration.CommunicationScore,
		metrics.Workload.UtilizationRate*100,
		metrics.Workload.BalanceScore,
		len(metrics.Workload.OverloadedMembers),
		len(metrics.Workload.UnderutilizedMembers),
	)

	data := map[string]interface{}{
		"metrics": metrics,
		"charts": map[string]interface{}{
			"workload_distribution": metrics.Workload.WorkloadDistribution,
			"productivity_trend":    metrics.Productivity.TrendDirection,
		},
	}

	return content, data
}

// generateTaskReport generates a task report
func (ae *DefaultAnalyticsEngine) generateTaskReport(request *ReportRequest) (string, map[string]interface{}) {
	projectID := ""
	if id, exists := request.Filters["project_id"]; exists {
		projectID = fmt.Sprintf("%v", id)
	}

	metrics, _ := ae.GetTaskMetrics(projectID, request.TimeRange)

	content := fmt.Sprintf(`
# Task Report: %s

## Task Overview
- **Project ID**: %s
- **Time Range**: %s to %s
- **Total Tasks**: %d
- **Completed Tasks**: %d
- **Completion Rate**: %.1f%%

## Velocity Metrics
- **Current Velocity**: %.1f
- **Average Velocity**: %.1f
- **Predicted Velocity**: %.1f
- **Throughput**: %.1f

## Lead Time Metrics
- **Average Lead Time**: %s
- **Lead Time P50**: %s
- **Lead Time P95**: %s

## Task Distribution
- **By Type**: Features: %d, Bugs: %d, Tasks: %d
- **By Priority**: High: %d, Medium: %d, Low: %d
`,
		request.Title,
		metrics.ProjectID,
		request.TimeRange.Start.Format("2006-01-02"),
		request.TimeRange.End.Format("2006-01-02"),
		metrics.TotalTasks,
		metrics.CompletedTasks,
		metrics.CompletionRate*100,
		metrics.VelocityMetrics.CurrentVelocity,
		metrics.VelocityMetrics.AverageVelocity,
		metrics.VelocityMetrics.PredictedVelocity,
		metrics.VelocityMetrics.Throughput,
		metrics.AverageLeadTime,
		metrics.VelocityMetrics.LeadTimeP50,
		metrics.VelocityMetrics.LeadTimeP95,
		metrics.TaskDistribution.ByType["feature"],
		metrics.TaskDistribution.ByType["bug"],
		metrics.TaskDistribution.ByType["task"],
		metrics.TaskDistribution.ByPriority["high"],
		metrics.TaskDistribution.ByPriority["medium"],
		metrics.TaskDistribution.ByPriority["low"],
	)

	data := map[string]interface{}{
		"metrics": metrics,
		"charts": map[string]interface{}{
			"velocity_trend":    metrics.VelocityMetrics.VelocityTrend,
			"task_distribution": metrics.TaskDistribution,
		},
	}

	return content, data
}

// generateWorkflowReport generates a workflow report
func (ae *DefaultAnalyticsEngine) generateWorkflowReport(request *ReportRequest) (string, map[string]interface{}) {
	workflowID := ""
	if id, exists := request.Filters["workflow_id"]; exists {
		workflowID = fmt.Sprintf("%v", id)
	}

	metrics, _ := ae.GetWorkflowMetrics(workflowID, request.TimeRange)

	content := fmt.Sprintf(`
# Workflow Report: %s

## Workflow Overview
- **Workflow ID**: %s
- **Time Range**: %s to %s
- **Total Executions**: %d
- **Success Rate**: %.1f%%

## Performance Metrics
- **Average Run Time**: %s
- **Successful Runs**: %d
- **Failed Runs**: %d
- **Average Response Time**: %s
- **Throughput**: %.1f/sec

## Error Analysis
- **Error Rate**: %.1f%%
- **Top Error Types**:
  - Timeout: %d
  - Network Error: %d
  - Validation Error: %d
`,
		request.Title,
		metrics.WorkflowID,
		request.TimeRange.Start.Format("2006-01-02"),
		request.TimeRange.End.Format("2006-01-02"),
		metrics.TotalExecutions,
		metrics.SuccessRate*100,
		metrics.AverageRunTime,
		metrics.SuccessfulRuns,
		metrics.FailedRuns,
		metrics.PerformanceMetrics.AverageResponseTime,
		metrics.PerformanceMetrics.ThroughputPerSecond,
		metrics.PerformanceMetrics.ErrorRate*100,
		metrics.ErrorDistribution["timeout"],
		metrics.ErrorDistribution["network_error"],
		metrics.ErrorDistribution["validation_error"],
	)

	data := map[string]interface{}{
		"metrics": metrics,
		"charts": map[string]interface{}{
			"execution_trends":   metrics.ExecutionTrends,
			"error_distribution": metrics.ErrorDistribution,
		},
	}

	return content, data
}

// generatePipelineReport generates a pipeline report
func (ae *DefaultAnalyticsEngine) generatePipelineReport(request *ReportRequest) (string, map[string]interface{}) {
	pipelineID := ""
	if id, exists := request.Filters["pipeline_id"]; exists {
		pipelineID = fmt.Sprintf("%v", id)
	}

	metrics, _ := ae.GetPipelineMetrics(pipelineID, request.TimeRange)

	content := fmt.Sprintf(`
# Pipeline Report: %s

## Pipeline Overview
- **Pipeline ID**: %s
- **Time Range**: %s to %s
- **Total Builds**: %d
- **Success Rate**: %.1f%%

## Build Metrics
- **Successful Builds**: %d
- **Failed Builds**: %d
- **Average Build Time**: %s

## Stage Performance
- **Build Stage**: %.1f%% success, %s avg time
- **Test Stage**: %.1f%% success, %s avg time
- **Deploy Stage**: %.1f%% success, %s avg time

## Failure Analysis
- **Test Failures**: %d
- **Build Errors**: %d
- **Timeouts**: %d
`,
		request.Title,
		metrics.PipelineID,
		request.TimeRange.Start.Format("2006-01-02"),
		request.TimeRange.End.Format("2006-01-02"),
		metrics.TotalBuilds,
		metrics.SuccessRate*100,
		metrics.SuccessfulBuilds,
		metrics.FailedBuilds,
		metrics.AverageBuildTime,
		metrics.StageMetrics[0].SuccessRate*100,
		metrics.StageMetrics[0].AverageTime,
		metrics.StageMetrics[1].SuccessRate*100,
		metrics.StageMetrics[1].AverageTime,
		metrics.StageMetrics[2].SuccessRate*100,
		metrics.StageMetrics[2].AverageTime,
		metrics.FailureReasons["test_failure"],
		metrics.FailureReasons["build_error"],
		metrics.FailureReasons["timeout"],
	)

	data := map[string]interface{}{
		"metrics": metrics,
		"charts": map[string]interface{}{
			"build_trends":    metrics.BuildTrends,
			"stage_metrics":   metrics.StageMetrics,
			"failure_reasons": metrics.FailureReasons,
		},
	}

	return content, data
}

// generateExecutiveReport generates an executive summary report
func (ae *DefaultAnalyticsEngine) generateExecutiveReport(request *ReportRequest) (string, map[string]interface{}) {
	automationMetrics, _ := ae.GetAutomationMetrics(request.TimeRange)

	content := fmt.Sprintf(`
# Executive Summary Report: %s

## Key Performance Indicators
- **Time Range**: %s to %s
- **Total Workflows**: %d
- **Active Workflows**: %d
- **Total Executions**: %d

## Business Impact
- **Automation Savings**: %s
- **Efficiency Gain**: %.1f%%
- **ROI**: %.1fx
- **Adoption Rate**: %.1f%%

## Cost Savings
- **Estimated Cost Savings**: %s
- **Time Saved**: %d days
- **Manual Processes Eliminated**: %d

## Recommendations
1. Continue expanding automation coverage
2. Focus on high-impact workflows
3. Improve adoption in underutilized teams
4. Invest in advanced analytics capabilities
`,
		request.Title,
		request.TimeRange.Start.Format("2006-01-02"),
		request.TimeRange.End.Format("2006-01-02"),
		automationMetrics.TotalWorkflows,
		automationMetrics.ActiveWorkflows,
		automationMetrics.TotalExecutions,
		automationMetrics.AutomationSavings,
		automationMetrics.EfficiencyGain*100,
		automationMetrics.ROI,
		automationMetrics.AdoptionRate*100,
		automationMetrics.Metadata["cost_savings"],
		automationMetrics.Metadata["time_saved_days"],
		automationMetrics.Metadata["manual_processes"],
	)

	data := map[string]interface{}{
		"metrics": automationMetrics,
		"charts": map[string]interface{}{
			"roi_trend":         []interface{}{},
			"adoption_trend":    []interface{}{},
			"savings_breakdown": automationMetrics.Metadata,
		},
	}

	return content, data
}
