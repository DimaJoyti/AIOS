package analytics

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Helper methods for analytics engine

// Validation methods

// validateDashboard validates a dashboard configuration
func (ae *DefaultAnalyticsEngine) validateDashboard(dashboard *Dashboard) error {
	if dashboard.Name == "" {
		return fmt.Errorf("dashboard name is required")
	}

	if dashboard.Layout == nil {
		dashboard.Layout = &DashboardLayout{
			Type:    "grid",
			Columns: 12,
			Rows:    8,
			Gap:     16,
		}
	}

	// Validate widgets
	for i, widget := range dashboard.Widgets {
		if err := ae.validateWidget(widget); err != nil {
			return fmt.Errorf("widget %d validation failed: %w", i, err)
		}
	}

	return nil
}

// validateWidget validates a widget configuration
func (ae *DefaultAnalyticsEngine) validateWidget(widget *Widget) error {
	if widget.ID == "" {
		widget.ID = uuid.New().String()
	}

	if widget.Title == "" {
		return fmt.Errorf("widget title is required")
	}

	if widget.Type == "" {
		return fmt.Errorf("widget type is required")
	}

	// Validate widget type
	validTypes := []WidgetType{
		WidgetTypeChart, WidgetTypeTable, WidgetTypeMetric,
		WidgetTypeProgress, WidgetTypeList, WidgetTypeCalendar,
		WidgetTypeKanban, WidgetTypeCustom,
	}

	found := false
	for _, validType := range validTypes {
		if widget.Type == validType {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("invalid widget type: %s", widget.Type)
	}

	// Set default position and size if not provided
	if widget.Position == nil {
		widget.Position = &WidgetPosition{X: 0, Y: 0}
	}
	if widget.Size == nil {
		widget.Size = &WidgetSize{Width: 4, Height: 3}
	}

	return nil
}

// Filter matching methods

// matchesDashboardFilter checks if a dashboard matches the given filter
func (ae *DefaultAnalyticsEngine) matchesDashboardFilter(dashboard *Dashboard, filter *DashboardFilter) bool {
	if filter == nil {
		return true
	}

	// Created by filter
	if filter.CreatedBy != "" && dashboard.CreatedBy != filter.CreatedBy {
		return false
	}

	// Public filter
	if filter.IsPublic != nil && dashboard.IsPublic != *filter.IsPublic {
		return false
	}

	// Tags filter
	if len(filter.Tags) > 0 {
		for _, filterTag := range filter.Tags {
			found := false
			for _, dashboardTag := range dashboard.Tags {
				if dashboardTag == filterTag {
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
		if !(strings.Contains(strings.ToLower(dashboard.Name), searchLower) ||
			strings.Contains(strings.ToLower(dashboard.Description), searchLower)) {
			return false
		}
	}

	return true
}

// Data generation helper methods

// generateVelocityTrend creates sample velocity trend data
func (ae *DefaultAnalyticsEngine) generateVelocityTrend(timeRange *TimeRange) []*TrendPoint {
	var points []*TrendPoint
	duration := timeRange.End.Sub(timeRange.Start)
	weeks := int(duration.Hours() / (7 * 24))

	baseVelocity := 10.0
	for i := 0; i < weeks; i++ {
		timestamp := timeRange.Start.AddDate(0, 0, i*7)
		variation := math.Sin(float64(i)*0.5) * 2
		trend := float64(i) * 0.2 // Slight upward trend
		velocity := baseVelocity + variation + trend

		points = append(points, &TrendPoint{
			Timestamp: timestamp,
			Value:     velocity,
			Metadata: map[string]interface{}{
				"sprint": fmt.Sprintf("Sprint %d", i+1),
				"points": int(velocity),
			},
		})
	}

	return points
}

// generateExecutionTrends creates sample execution trend data
func (ae *DefaultAnalyticsEngine) generateExecutionTrends(timeRange *TimeRange) []*TrendPoint {
	var points []*TrendPoint
	duration := timeRange.End.Sub(timeRange.Start)
	days := int(duration.Hours() / 24)

	for i := 0; i < days; i++ {
		timestamp := timeRange.Start.AddDate(0, 0, i)
		baseExecutions := 8.0
		variation := math.Sin(float64(i)*0.3) * 3
		executions := baseExecutions + variation

		if executions < 0 {
			executions = 0
		}

		points = append(points, &TrendPoint{
			Timestamp: timestamp,
			Value:     executions,
			Metadata: map[string]interface{}{
				"successful": int(executions * 0.9),
				"failed":     int(executions * 0.1),
			},
		})
	}

	return points
}

// generateBuildTrends creates sample build trend data
func (ae *DefaultAnalyticsEngine) generateBuildTrends(timeRange *TimeRange) []*TrendPoint {
	var points []*TrendPoint
	duration := timeRange.End.Sub(timeRange.Start)
	days := int(duration.Hours() / 24)

	for i := 0; i < days; i++ {
		timestamp := timeRange.Start.AddDate(0, 0, i)
		baseBuilds := 5.0
		variation := math.Cos(float64(i)*0.2) * 2
		builds := baseBuilds + variation

		if builds < 0 {
			builds = 0
		}

		points = append(points, &TrendPoint{
			Timestamp: timestamp,
			Value:     builds,
			Metadata: map[string]interface{}{
				"successful": int(builds * 0.91),
				"failed":     int(builds * 0.09),
			},
		})
	}

	return points
}

// generateStageTrend creates sample stage performance trend data
func (ae *DefaultAnalyticsEngine) generateStageTrend(timeRange *TimeRange) []*TrendPoint {
	var points []*TrendPoint
	duration := timeRange.End.Sub(timeRange.Start)
	days := int(duration.Hours() / 24)

	for i := 0; i < days; i += 3 { // Every 3 days
		timestamp := timeRange.Start.AddDate(0, 0, i)
		baseTime := 180.0 // 3 minutes in seconds
		variation := math.Sin(float64(i)*0.1) * 30
		stageTime := baseTime + variation

		points = append(points, &TrendPoint{
			Timestamp: timestamp,
			Value:     stageTime,
			Metadata: map[string]interface{}{
				"builds":           5 + i/3,
				"avg_time_minutes": stageTime / 60,
			},
		})
	}

	return points
}

// Default data creation methods

// createDefaultDashboards creates default analytics dashboards
func (ae *DefaultAnalyticsEngine) createDefaultDashboards() {
	// Project Overview Dashboard
	projectDashboard := &Dashboard{
		ID:          uuid.New().String(),
		Name:        "Project Overview",
		Description: "Comprehensive project analytics dashboard",
		Layout: &DashboardLayout{
			Type:    "grid",
			Columns: 12,
			Rows:    8,
			Gap:     16,
		},
		Widgets: []*Widget{
			{
				ID:          uuid.New().String(),
				Type:        WidgetTypeMetric,
				Title:       "Completion Rate",
				Description: "Overall project completion percentage",
				Position:    &WidgetPosition{X: 0, Y: 0},
				Size:        &WidgetSize{Width: 3, Height: 2},
				Config: map[string]interface{}{
					"metric_type": "percentage",
					"color":       "green",
				},
				DataSource: &DataSource{
					Type:  "project_metrics",
					Query: "completion_rate",
				},
				RefreshRate: 300, // 5 minutes
			},
			{
				ID:          uuid.New().String(),
				Type:        WidgetTypeChart,
				Title:       "Velocity Trend",
				Description: "Sprint velocity over time",
				Position:    &WidgetPosition{X: 3, Y: 0},
				Size:        &WidgetSize{Width: 6, Height: 4},
				Config: map[string]interface{}{
					"chart_type": "line",
					"x_axis":     "time",
					"y_axis":     "velocity",
				},
				DataSource: &DataSource{
					Type:  "project_trends",
					Query: "velocity",
				},
				RefreshRate: 600, // 10 minutes
			},
			{
				ID:          uuid.New().String(),
				Type:        WidgetTypeProgress,
				Title:       "Sprint Progress",
				Description: "Current sprint completion status",
				Position:    &WidgetPosition{X: 9, Y: 0},
				Size:        &WidgetSize{Width: 3, Height: 2},
				Config: map[string]interface{}{
					"progress_type": "circular",
					"show_percentage": true,
				},
				DataSource: &DataSource{
					Type:  "sprint_metrics",
					Query: "current_sprint_progress",
				},
				RefreshRate: 300,
			},
		},
		IsPublic:  true,
		CreatedBy: "system",
		Tags:      []string{"project", "overview", "default"},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Team Performance Dashboard
	teamDashboard := &Dashboard{
		ID:          uuid.New().String(),
		Name:        "Team Performance",
		Description: "Team productivity and collaboration metrics",
		Layout: &DashboardLayout{
			Type:    "grid",
			Columns: 12,
			Rows:    8,
			Gap:     16,
		},
		Widgets: []*Widget{
			{
				ID:          uuid.New().String(),
				Type:        WidgetTypeChart,
				Title:       "Team Productivity",
				Description: "Tasks completed per team member",
				Position:    &WidgetPosition{X: 0, Y: 0},
				Size:        &WidgetSize{Width: 6, Height: 4},
				Config: map[string]interface{}{
					"chart_type": "bar",
					"orientation": "horizontal",
				},
				DataSource: &DataSource{
					Type:  "team_metrics",
					Query: "productivity_by_member",
				},
				RefreshRate: 600,
			},
			{
				ID:          uuid.New().String(),
				Type:        WidgetTypeMetric,
				Title:       "Utilization Rate",
				Description: "Team capacity utilization",
				Position:    &WidgetPosition{X: 6, Y: 0},
				Size:        &WidgetSize{Width: 3, Height: 2},
				Config: map[string]interface{}{
					"metric_type": "percentage",
					"color":       "blue",
				},
				DataSource: &DataSource{
					Type:  "team_workload",
					Query: "utilization_rate",
				},
				RefreshRate: 300,
			},
		},
		IsPublic:  true,
		CreatedBy: "system",
		Tags:      []string{"team", "performance", "default"},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Store dashboards
	ae.dashboards[projectDashboard.ID] = projectDashboard
	ae.dashboards[teamDashboard.ID] = teamDashboard

	ae.logger.WithField("dashboards", len(ae.dashboards)).Info("Default dashboards created")
}

// createDefaultReportTemplates creates default report templates
func (ae *DefaultAnalyticsEngine) createDefaultReportTemplates() {
	// Project Summary Template
	projectTemplate := &ReportTemplate{
		ID:          uuid.New().String(),
		Name:        "Project Summary Report",
		Description: "Comprehensive project performance summary",
		Type:        ReportTypeProject,
		Template: `
# {{.Title}}

## Project Overview
- **Project**: {{.ProjectName}}
- **Period**: {{.TimeRange.Start}} to {{.TimeRange.End}}
- **Completion Rate**: {{.CompletionRate}}%

## Key Metrics
- **Total Tasks**: {{.TasksTotal}}
- **Completed**: {{.TasksCompleted}}
- **Velocity**: {{.Velocity}} points/week

## Quality Indicators
- **Bug Count**: {{.BugCount}}
- **Test Coverage**: {{.TestCoverage}}%
- **Code Review Rate**: {{.CodeReviewRate}}%
`,
		Parameters: map[string]interface{}{
			"include_charts":     true,
			"include_milestones": true,
			"format":             "pdf",
		},
		IsPublic:  true,
		CreatedBy: "system",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Executive Summary Template
	executiveTemplate := &ReportTemplate{
		ID:          uuid.New().String(),
		Name:        "Executive Summary",
		Description: "High-level business impact summary",
		Type:        ReportTypeExecutive,
		Template: `
# {{.Title}}

## Executive Summary
{{.Summary}}

## Key Performance Indicators
- **ROI**: {{.ROI}}x
- **Efficiency Gain**: {{.EfficiencyGain}}%
- **Cost Savings**: {{.CostSavings}}

## Strategic Recommendations
{{.Recommendations}}
`,
		Parameters: map[string]interface{}{
			"include_financials": true,
			"include_trends":     true,
			"format":             "pdf",
		},
		IsPublic:  true,
		CreatedBy: "system",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Store templates
	ae.reportTemplates[projectTemplate.ID] = projectTemplate
	ae.reportTemplates[executiveTemplate.ID] = executiveTemplate

	ae.logger.WithField("templates", len(ae.reportTemplates)).Info("Default report templates created")
}
