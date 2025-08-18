# AIOS Project Analytics and Reporting

## Overview

The AIOS Project Analytics and Reporting system provides comprehensive insights into project performance, team productivity, workflow efficiency, and business impact. Built with enterprise-grade capabilities, it offers real-time analytics, customizable dashboards, automated reporting, and advanced data visualization to drive data-driven decision making.

## 🏗️ Architecture

### Core Components

```
Project Analytics & Reporting
├── Analytics Engine (metrics computation, trend analysis)
├── Report Generator (automated report creation, templates)
├── Dashboard Manager (interactive dashboards, widgets)
├── Data Aggregator (multi-source data collection)
├── Metrics Calculator (KPIs, performance indicators)
├── Trend Analyzer (time-series analysis, forecasting)
├── Cache Manager (performance optimization)
└── Scheduler (automated reports, recurring analytics)
```

### Key Features

- **📊 Multi-Dimensional Analytics**: Project, team, task, workflow, and pipeline metrics
- **📈 Real-Time Dashboards**: Interactive, customizable analytics dashboards
- **📋 Automated Reporting**: Scheduled reports with multiple formats (PDF, HTML, CSV)
- **🔍 Trend Analysis**: Time-series analysis with forecasting capabilities
- **⚡ Performance Optimization**: Intelligent caching and data aggregation
- **🎯 KPI Tracking**: Comprehensive key performance indicators
- **📱 Multi-Format Export**: Support for various output formats
- **🔄 Real-Time Updates**: Live data refresh and notifications

## 🚀 Quick Start

### Basic Analytics Setup

```go
package main

import (
    "time"
    "github.com/aios/aios/pkg/analytics"
    "github.com/sirupsen/logrus"
)

func main() {
    logger := logrus.New()
    
    // Create analytics engine
    config := &analytics.AnalyticsEngineConfig{
        CacheTTL:        15 * time.Minute,
        MaxCacheSize:    1000,
        EnableRealTime:  true,
        ComputeInterval: 5 * time.Minute,
        RetentionPeriod: 90 * 24 * time.Hour,
    }
    
    analyticsEngine := analytics.NewDefaultAnalyticsEngine(config, logger)
    
    // Define analysis time range
    timeRange := &analytics.TimeRange{
        Start: time.Now().AddDate(0, 0, -30), // Last 30 days
        End:   time.Now(),
    }
    
    // Get project metrics
    projectMetrics, err := analyticsEngine.GetProjectMetrics("my-project", timeRange)
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("Project Completion Rate: %.1f%%\n", projectMetrics.CompletionRate*100)
    fmt.Printf("Project Velocity: %.1f points/week\n", projectMetrics.Velocity)
    fmt.Printf("Bug Count: %d\n", projectMetrics.QualityMetrics.BugCount)
    fmt.Printf("Test Coverage: %.1f%%\n", projectMetrics.QualityMetrics.TestCoverage)
}
```

## 📊 Analytics Capabilities

### Project Analytics

Comprehensive project performance analysis:

```go
// Get detailed project metrics
projectMetrics, err := analyticsEngine.GetProjectMetrics(projectID, timeRange)

// Key metrics available:
fmt.Printf("Tasks: %d total, %d completed (%.1f%%)\n",
    projectMetrics.TasksTotal,
    projectMetrics.TasksCompleted,
    projectMetrics.CompletionRate*100)

fmt.Printf("Velocity: %.1f points/week\n", projectMetrics.Velocity)
fmt.Printf("Quality Score: %.1f\n", projectMetrics.QualityMetrics.PerformanceScore)

// Sprint analysis
for _, sprint := range projectMetrics.SprintMetrics {
    fmt.Printf("Sprint %s: %d/%d points completed (%.1f%%)\n",
        sprint.SprintName,
        sprint.CompletedPoints,
        sprint.PlannedPoints,
        sprint.CompletionRate*100)
}

// Milestone tracking
for _, milestone := range projectMetrics.MilestoneStatus {
    fmt.Printf("Milestone %s: %.1f%% complete, %d days remaining\n",
        milestone.MilestoneName,
        milestone.CompletionRate*100,
        milestone.DaysRemaining)
}

// Get project trends
trends, err := analyticsEngine.GetProjectTrends(projectID, timeRange, analytics.GranularityWeek)
for _, point := range trends {
    fmt.Printf("Week %s: %.1f velocity\n",
        point.Timestamp.Format("2006-01-02"),
        point.Value)
}

// Compare multiple projects
comparison, err := analyticsEngine.GetProjectComparison(
    []string{"project-a", "project-b", "project-c"},
    timeRange)

fmt.Printf("Average velocity across projects: %.1f\n",
    comparison.Comparisons["average_velocity"].(float64))
```

### Team Analytics

Team performance and collaboration insights:

```go
// Get team performance metrics
teamMetrics, err := analyticsEngine.GetTeamMetrics(teamID, timeRange)

// Team overview
fmt.Printf("Team: %d members (%d active)\n",
    teamMetrics.MemberCount,
    teamMetrics.ActiveMembers)

// Productivity analysis
productivity := teamMetrics.Productivity
fmt.Printf("Productivity Score: %.1f\n", productivity.ProductivityScore)
fmt.Printf("Tasks per Member: %.1f\n", productivity.TasksPerMember)
fmt.Printf("Code Commits: %d\n", productivity.CodeCommits)
fmt.Printf("Features Delivered: %d\n", productivity.FeaturesDelivered)

// Collaboration metrics
collaboration := teamMetrics.Collaboration
fmt.Printf("Communication Score: %.1f\n", collaboration.CommunicationScore)
fmt.Printf("Review Participation: %.1f%%\n", collaboration.ReviewParticipation*100)
fmt.Printf("Knowledge Sharing: %.1f\n", collaboration.KnowledgeSharing)

// Workload distribution
workload := teamMetrics.Workload
fmt.Printf("Utilization Rate: %.1f%%\n", workload.UtilizationRate*100)
fmt.Printf("Balance Score: %.1f\n", workload.BalanceScore)

for member, utilization := range workload.WorkloadDistribution {
    fmt.Printf("  %s: %.1f%% utilized\n", member, utilization*100)
}

// Identify workload issues
if len(workload.OverloadedMembers) > 0 {
    fmt.Printf("Overloaded members: %v\n", workload.OverloadedMembers)
}
if len(workload.UnderutilizedMembers) > 0 {
    fmt.Printf("Underutilized members: %v\n", workload.UnderutilizedMembers)
}
```

### Task Analytics

Detailed task performance and velocity analysis:

```go
// Get task metrics
taskMetrics, err := analyticsEngine.GetTaskMetrics(projectID, timeRange)

// Task overview
fmt.Printf("Tasks: %d total, %d completed (%.1f%%)\n",
    taskMetrics.TotalTasks,
    taskMetrics.CompletedTasks,
    taskMetrics.CompletionRate*100)

// Lead time analysis
fmt.Printf("Average Lead Time: %s\n", taskMetrics.AverageLeadTime)
fmt.Printf("Average Cycle Time: %s\n", taskMetrics.AverageCycleTime)

// Velocity metrics
velocity := taskMetrics.VelocityMetrics
fmt.Printf("Current Velocity: %.1f\n", velocity.CurrentVelocity)
fmt.Printf("Average Velocity: %.1f\n", velocity.AverageVelocity)
fmt.Printf("Predicted Velocity: %.1f\n", velocity.PredictedVelocity)
fmt.Printf("Throughput: %.1f tasks/week\n", velocity.Throughput)

// Performance percentiles
fmt.Printf("Lead Time P50: %s\n", velocity.LeadTimeP50)
fmt.Printf("Lead Time P95: %s\n", velocity.LeadTimeP95)
fmt.Printf("Cycle Time P50: %s\n", velocity.CycleTimeP50)
fmt.Printf("Cycle Time P95: %s\n", velocity.CycleTimeP95)

// Task distribution analysis
distribution := taskMetrics.TaskDistribution
fmt.Printf("By Type: Features=%d, Bugs=%d, Tasks=%d\n",
    distribution.ByType["feature"],
    distribution.ByType["bug"],
    distribution.ByType["task"])

fmt.Printf("By Priority: High=%d, Medium=%d, Low=%d\n",
    distribution.ByPriority["high"],
    distribution.ByPriority["medium"],
    distribution.ByPriority["low"])

fmt.Printf("By Status: Todo=%d, InProgress=%d, Done=%d\n",
    distribution.ByStatus["todo"],
    distribution.ByStatus["in_progress"],
    distribution.ByStatus["done"])
```

### Workflow Analytics

Workflow execution performance and reliability:

```go
// Get workflow metrics
workflowMetrics, err := analyticsEngine.GetWorkflowMetrics(workflowID, timeRange)

// Execution overview
fmt.Printf("Workflow Executions: %d total\n", workflowMetrics.TotalExecutions)
fmt.Printf("Success Rate: %.1f%%\n", workflowMetrics.SuccessRate*100)
fmt.Printf("Average Run Time: %s\n", workflowMetrics.AverageRunTime)

// Performance metrics
perf := workflowMetrics.PerformanceMetrics
fmt.Printf("Average Response Time: %s\n", perf.AverageResponseTime)
fmt.Printf("P95 Response Time: %s\n", perf.P95ResponseTime)
fmt.Printf("Throughput: %.1f/sec\n", perf.ThroughputPerSecond)
fmt.Printf("Error Rate: %.1f%%\n", perf.ErrorRate*100)
fmt.Printf("Availability: %.1f%%\n", perf.AvailabilityRate*100)

// Error analysis
fmt.Printf("Error Distribution:\n")
for errorType, count := range workflowMetrics.ErrorDistribution {
    fmt.Printf("  %s: %d occurrences\n", errorType, count)
}

// Execution trends
for _, trend := range workflowMetrics.ExecutionTrends {
    fmt.Printf("Date %s: %.0f executions\n",
        trend.Timestamp.Format("2006-01-02"),
        trend.Value)
}
```

### Pipeline Analytics

CI/CD pipeline performance and build analysis:

```go
// Get pipeline metrics
pipelineMetrics, err := analyticsEngine.GetPipelineMetrics(pipelineID, timeRange)

// Build overview
fmt.Printf("Pipeline Builds: %d total\n", pipelineMetrics.TotalBuilds)
fmt.Printf("Success Rate: %.1f%%\n", pipelineMetrics.SuccessRate*100)
fmt.Printf("Average Build Time: %s\n", pipelineMetrics.AverageBuildTime)
fmt.Printf("Failed Builds: %d\n", pipelineMetrics.FailedBuilds)

// Stage performance analysis
fmt.Printf("Stage Performance:\n")
for _, stage := range pipelineMetrics.StageMetrics {
    fmt.Printf("  %s: %.1f%% success, %s avg time\n",
        stage.StageName,
        stage.SuccessRate*100,
        stage.AverageTime)
    
    // Stage failure analysis
    if len(stage.FailureReasons) > 0 {
        fmt.Printf("    Failures: ")
        for reason, count := range stage.FailureReasons {
            fmt.Printf("%s=%d ", reason, count)
        }
        fmt.Println()
    }
}

// Build failure analysis
fmt.Printf("Failure Reasons:\n")
for reason, count := range pipelineMetrics.FailureReasons {
    fmt.Printf("  %s: %d occurrences\n", reason, count)
}

// Build trends
for _, trend := range pipelineMetrics.BuildTrends {
    successful := trend.Metadata["successful"].(int)
    failed := trend.Metadata["failed"].(int)
    fmt.Printf("Date %s: %d successful, %d failed\n",
        trend.Timestamp.Format("2006-01-02"),
        successful, failed)
}
```

### Automation Impact Analysis

Overall automation effectiveness and ROI:

```go
// Get automation metrics
automationMetrics, err := analyticsEngine.GetAutomationMetrics(timeRange)

// Automation overview
fmt.Printf("Workflows: %d total (%d active)\n",
    automationMetrics.TotalWorkflows,
    automationMetrics.ActiveWorkflows)
fmt.Printf("Total Executions: %d\n", automationMetrics.TotalExecutions)

// Business impact
fmt.Printf("Time Saved: %s\n", automationMetrics.AutomationSavings)
fmt.Printf("Efficiency Gain: %.1f%%\n", automationMetrics.EfficiencyGain*100)
fmt.Printf("ROI: %.1fx\n", automationMetrics.ROI)
fmt.Printf("Adoption Rate: %.1f%%\n", automationMetrics.AdoptionRate*100)

// Cost analysis
fmt.Printf("Cost Savings: %s\n", automationMetrics.Metadata["cost_savings"])
fmt.Printf("Time Saved: %d days\n", automationMetrics.Metadata["time_saved_days"])
fmt.Printf("Manual Processes Eliminated: %d\n", automationMetrics.Metadata["manual_processes"])
```

## 📋 Report Generation

### Automated Report Creation

Generate comprehensive reports in multiple formats:

```go
// Project performance report
projectReportRequest := &analytics.ReportRequest{
    Type:        analytics.ReportTypeProject,
    Title:       "Monthly Project Performance Report",
    Description: "Comprehensive project analysis for the last 30 days",
    TimeRange:   timeRange,
    Filters: map[string]interface{}{
        "project_id": "my-project",
    },
    Format:     analytics.ReportFormatPDF,
    Recipients: []string{"manager@company.com", "team-lead@company.com"},
}

projectReport, err := analyticsEngine.GenerateReport(projectReportRequest)
if err != nil {
    panic(err)
}

fmt.Printf("Report Generated: %s\n", projectReport.ID)
fmt.Printf("Format: %s, Size: %d bytes\n", projectReport.Format, projectReport.Size)
fmt.Printf("Download URL: %s\n", projectReport.URL)

// Team performance report
teamReportRequest := &analytics.ReportRequest{
    Type:        analytics.ReportTypeTeam,
    Title:       "Team Productivity Analysis",
    Description: "Team performance and collaboration metrics",
    TimeRange:   timeRange,
    Filters: map[string]interface{}{
        "team_id": "development-team",
    },
    Format: analytics.ReportFormatHTML,
}

teamReport, err := analyticsEngine.GenerateReport(teamReportRequest)

// Executive summary report
executiveReportRequest := &analytics.ReportRequest{
    Type:        analytics.ReportTypeExecutive,
    Title:       "Executive Summary - Q4 2024",
    Description: "High-level business impact and ROI analysis",
    TimeRange:   timeRange,
    Format:      analytics.ReportFormatPDF,
    Recipients:  []string{"ceo@company.com", "cto@company.com"},
}

executiveReport, err := analyticsEngine.GenerateReport(executiveReportRequest)
```

### Report Scheduling

Schedule recurring reports for automated delivery:

```go
// Schedule weekly team report
weeklyTeamReport := &analytics.ReportSchedule{
    Name:       "Weekly Team Performance Report",
    ReportType: analytics.ReportTypeTeam,
    Schedule:   "0 9 * * MON", // Every Monday at 9 AM
    Recipients: []string{"team-lead@company.com", "manager@company.com"},
    Filters: map[string]interface{}{
        "team_id": "development-team",
    },
    Format:  analytics.ReportFormatPDF,
    Enabled: true,
}

err := analyticsEngine.ScheduleReport(weeklyTeamReport)

// Schedule monthly executive summary
monthlyExecutiveReport := &analytics.ReportSchedule{
    Name:       "Monthly Executive Summary",
    ReportType: analytics.ReportTypeExecutive,
    Schedule:   "0 8 1 * *", // First day of month at 8 AM
    Recipients: []string{"ceo@company.com", "board@company.com"},
    Format:     analytics.ReportFormatPDF,
    Enabled:    true,
}

err = analyticsEngine.ScheduleReport(monthlyExecutiveReport)

// Schedule daily pipeline report
dailyPipelineReport := &analytics.ReportSchedule{
    Name:       "Daily Pipeline Status",
    ReportType: analytics.ReportTypePipeline,
    Schedule:   "0 7 * * *", // Every day at 7 AM
    Recipients: []string{"devops@company.com"},
    Filters: map[string]interface{}{
        "pipeline_id": "production-pipeline",
    },
    Format:  analytics.ReportFormatHTML,
    Enabled: true,
}

err = analyticsEngine.ScheduleReport(dailyPipelineReport)
```

### Report Templates

Use and manage report templates:

```go
// Get available templates
templates, err := analyticsEngine.GetReportTemplates()

fmt.Printf("Available Templates:\n")
for _, template := range templates {
    fmt.Printf("  %s (%s): %s\n",
        template.Name,
        template.Type,
        template.Description)
}

// Create custom template
customTemplate := &analytics.ReportTemplate{
    Name:        "Custom Project Dashboard",
    Description: "Tailored project metrics for stakeholders",
    Type:        analytics.ReportTypeProject,
    Template: `
# {{.Title}}

## Project Overview
- **Project**: {{.ProjectName}}
- **Period**: {{.TimeRange.Start}} to {{.TimeRange.End}}
- **Status**: {{.Status}}

## Key Metrics
- **Completion Rate**: {{.CompletionRate}}%
- **Velocity**: {{.Velocity}} points/week
- **Quality Score**: {{.QualityScore}}

## Team Performance
- **Team Size**: {{.TeamSize}} members
- **Productivity**: {{.ProductivityScore}}
- **Collaboration**: {{.CollaborationScore}}
`,
    Parameters: map[string]interface{}{
        "include_charts":     true,
        "include_milestones": true,
        "format":             "pdf",
    },
    IsPublic:  false,
    CreatedBy: "project-manager@company.com",
}
```

## 📱 Dashboard Management

### Interactive Dashboards

Create and manage real-time analytics dashboards:

```go
// Create custom dashboard
dashboard := &analytics.Dashboard{
    Name:        "DevOps Performance Dashboard",
    Description: "Real-time DevOps metrics and KPIs",
    Layout: &analytics.DashboardLayout{
        Type:    "grid",
        Columns: 12,
        Rows:    8,
        Gap:     16,
    },
    Widgets: []*analytics.Widget{
        {
            Type:        analytics.WidgetTypeMetric,
            Title:       "Deployment Frequency",
            Description: "Daily deployment rate",
            Position:    &analytics.WidgetPosition{X: 0, Y: 0},
            Size:        &analytics.WidgetSize{Width: 3, Height: 2},
            Config: map[string]interface{}{
                "metric_type": "number",
                "unit":        "deployments/day",
                "color":       "blue",
            },
            DataSource: &analytics.DataSource{
                Type:  "pipeline_metrics",
                Query: "deployment_frequency",
            },
            RefreshRate: 300, // 5 minutes
        },
        {
            Type:        analytics.WidgetTypeChart,
            Title:       "Build Success Rate",
            Description: "Build success rate over time",
            Position:    &analytics.WidgetPosition{X: 3, Y: 0},
            Size:        &analytics.WidgetSize{Width: 6, Height: 4},
            Config: map[string]interface{}{
                "chart_type": "line",
                "y_axis":     "success_rate",
                "color":      "green",
            },
            DataSource: &analytics.DataSource{
                Type:  "pipeline_trends",
                Query: "success_rate",
            },
            RefreshRate: 600, // 10 minutes
        },
        {
            Type:        analytics.WidgetTypeProgress,
            Title:       "Sprint Progress",
            Description: "Current sprint completion",
            Position:    &analytics.WidgetPosition{X: 9, Y: 0},
            Size:        &analytics.WidgetSize{Width: 3, Height: 2},
            Config: map[string]interface{}{
                "progress_type": "circular",
                "show_percentage": true,
            },
            DataSource: &analytics.DataSource{
                Type:  "sprint_metrics",
                Query: "current_progress",
            },
            RefreshRate: 300,
        },
    },
    IsPublic:  false,
    CreatedBy: "devops-team@company.com",
    Tags:      []string{"devops", "performance", "real-time"},
}

createdDashboard, err := analyticsEngine.CreateDashboard(dashboard)
if err != nil {
    panic(err)
}

fmt.Printf("Dashboard Created: %s\n", createdDashboard.ID)
fmt.Printf("Widgets: %d\n", len(createdDashboard.Widgets))

// List all dashboards
dashboards, err := analyticsEngine.ListDashboards(&analytics.DashboardFilter{
    CreatedBy: "devops-team@company.com",
    Tags:      []string{"performance"},
    Limit:     10,
})

for _, dash := range dashboards {
    fmt.Printf("Dashboard: %s (%d widgets)\n", dash.Name, len(dash.Widgets))
}

// Get specific dashboard
dashboard, err = analyticsEngine.GetDashboard(createdDashboard.ID)

// Update dashboard
dashboard.Description = "Updated DevOps metrics dashboard"
dashboard.Widgets = append(dashboard.Widgets, &analytics.Widget{
    Type:        analytics.WidgetTypeTable,
    Title:       "Recent Deployments",
    Description: "Latest deployment history",
    Position:    &analytics.WidgetPosition{X: 0, Y: 4},
    Size:        &analytics.WidgetSize{Width: 12, Height: 3},
    DataSource: &analytics.DataSource{
        Type:  "deployment_history",
        Query: "recent_deployments",
    },
    RefreshRate: 300,
})

updatedDashboard, err := analyticsEngine.UpdateDashboard(dashboard)
```

## 🧪 Testing

Run the comprehensive test suite:

```bash
# Run all analytics tests
go test ./pkg/analytics/...

# Run with race detection
go test -race ./pkg/analytics/...

# Run integration tests
go test -tags=integration ./pkg/analytics/...

# Run analytics example
go run examples/analytics_reporting_example.go
```

## 📖 Examples

See the complete example in `examples/analytics_reporting_example.go` for a comprehensive demonstration including:

- Project performance analysis with trends and comparisons
- Team productivity and collaboration metrics
- Task velocity and distribution analysis
- Workflow execution performance monitoring
- CI/CD pipeline success rate tracking
- Automation impact and ROI calculation
- Automated report generation in multiple formats
- Interactive dashboard creation and management
- Report scheduling and template management

## 🤝 Contributing

1. Follow established patterns and interfaces
2. Add comprehensive tests for new analytics features
3. Update documentation for API changes
4. Ensure proper error handling and logging
5. Use OpenTelemetry for observability
6. Implement caching for performance optimization

## 📄 License

This Project Analytics and Reporting system is part of the AIOS project and follows the same licensing terms.
