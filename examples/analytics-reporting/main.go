package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aios/aios/pkg/analytics"
	"github.com/sirupsen/logrus"
)

func main() {
	// Initialize logger
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)
	logger.SetFormatter(&logrus.TextFormatter{
		DisableTimestamp: true,
		ForceColors:      true,
	})

	fmt.Println("📊 AIOS Project Analytics and Reporting Demo")
	fmt.Println("============================================")

	// Run the comprehensive demo
	if err := runAnalyticsReportingDemo(logger); err != nil {
		log.Fatalf("Demo failed: %v", err)
	}

	fmt.Println("\n✅ Analytics and Reporting Demo completed successfully!")
}

func runAnalyticsReportingDemo(logger *logrus.Logger) error {
	// Step 1: Create Analytics Engine
	fmt.Println("\n1. Creating Analytics Engine...")
	config := &analytics.AnalyticsEngineConfig{
		CacheTTL:        15 * time.Minute,
		MaxCacheSize:    1000,
		EnableRealTime:  true,
		ComputeInterval: 5 * time.Minute,
		RetentionPeriod: 90 * 24 * time.Hour,
	}

	analyticsEngine := analytics.NewDefaultAnalyticsEngine(config, logger)
	fmt.Println("✓ Analytics Engine created successfully")

	// Step 2: Define Time Range for Analysis
	fmt.Println("\n2. Setting Up Analysis Time Range...")
	timeRange := &analytics.TimeRange{
		Start: time.Now().AddDate(0, 0, -30), // Last 30 days
		End:   time.Now(),
	}
	fmt.Printf("✓ Analysis period: %s to %s\n",
		timeRange.Start.Format("2006-01-02"),
		timeRange.End.Format("2006-01-02"))

	// Step 3: Project Analytics
	fmt.Println("\n3. Analyzing Project Metrics...")
	projectID := "project-fullstack-app"

	projectMetrics, err := analyticsEngine.GetProjectMetrics(projectID, timeRange)
	if err != nil {
		return fmt.Errorf("failed to get project metrics: %w", err)
	}

	fmt.Printf("   ✓ Project: %s\n", projectID)
	fmt.Printf("     - Total Tasks: %d\n", projectMetrics.TasksTotal)
	fmt.Printf("     - Completed: %d (%.1f%%)\n",
		projectMetrics.TasksCompleted,
		projectMetrics.CompletionRate*100)
	fmt.Printf("     - Velocity: %.1f points/week\n", projectMetrics.Velocity)
	fmt.Printf("     - Bug Count: %d\n", projectMetrics.QualityMetrics.BugCount)
	fmt.Printf("     - Test Coverage: %.1f%%\n", projectMetrics.QualityMetrics.TestCoverage)

	// Get project trends
	projectTrends, err := analyticsEngine.GetProjectTrends(projectID, timeRange, analytics.GranularityWeek)
	if err != nil {
		return fmt.Errorf("failed to get project trends: %w", err)
	}
	fmt.Printf("     - Trend Points: %d weekly data points\n", len(projectTrends))

	// Step 4: Team Analytics
	fmt.Println("\n4. Analyzing Team Performance...")
	teamID := "team-development"

	teamMetrics, err := analyticsEngine.GetTeamMetrics(teamID, timeRange)
	if err != nil {
		return fmt.Errorf("failed to get team metrics: %w", err)
	}

	fmt.Printf("   ✓ Team: %s\n", teamID)
	fmt.Printf("     - Team Size: %d members (%d active)\n",
		teamMetrics.MemberCount, teamMetrics.ActiveMembers)
	fmt.Printf("     - Tasks Completed: %d\n", teamMetrics.TasksCompleted)
	fmt.Printf("     - Average Velocity: %.1f\n", teamMetrics.AverageVelocity)
	fmt.Printf("     - Productivity Score: %.1f\n", teamMetrics.Productivity.ProductivityScore)
	fmt.Printf("     - Utilization Rate: %.1f%%\n", teamMetrics.Workload.UtilizationRate*100)
	fmt.Printf("     - Communication Score: %.1f\n", teamMetrics.Collaboration.CommunicationScore)

	// Step 5: Task Analytics
	fmt.Println("\n5. Analyzing Task Metrics...")

	taskMetrics, err := analyticsEngine.GetTaskMetrics(projectID, timeRange)
	if err != nil {
		return fmt.Errorf("failed to get task metrics: %w", err)
	}

	fmt.Printf("   ✓ Task Analysis:\n")
	fmt.Printf("     - Total Tasks: %d\n", taskMetrics.TotalTasks)
	fmt.Printf("     - Completion Rate: %.1f%%\n", taskMetrics.CompletionRate*100)
	fmt.Printf("     - Average Lead Time: %s\n", taskMetrics.AverageLeadTime)
	fmt.Printf("     - Current Velocity: %.1f\n", taskMetrics.VelocityMetrics.CurrentVelocity)
	fmt.Printf("     - Predicted Velocity: %.1f\n", taskMetrics.VelocityMetrics.PredictedVelocity)

	// Task distribution analysis
	fmt.Printf("   ✓ Task Distribution:\n")
	fmt.Printf("     - Features: %d, Bugs: %d, Tasks: %d\n",
		taskMetrics.TaskDistribution.ByType["feature"],
		taskMetrics.TaskDistribution.ByType["bug"],
		taskMetrics.TaskDistribution.ByType["task"])
	fmt.Printf("     - High Priority: %d, Medium: %d, Low: %d\n",
		taskMetrics.TaskDistribution.ByPriority["high"],
		taskMetrics.TaskDistribution.ByPriority["medium"],
		taskMetrics.TaskDistribution.ByPriority["low"])

	// Step 6: Workflow Analytics
	fmt.Println("\n6. Analyzing Workflow Performance...")
	workflowID := "workflow-ci-cd-pipeline"

	workflowMetrics, err := analyticsEngine.GetWorkflowMetrics(workflowID, timeRange)
	if err != nil {
		return fmt.Errorf("failed to get workflow metrics: %w", err)
	}

	fmt.Printf("   ✓ Workflow: %s\n", workflowID)
	fmt.Printf("     - Total Executions: %d\n", workflowMetrics.TotalExecutions)
	fmt.Printf("     - Success Rate: %.1f%%\n", workflowMetrics.SuccessRate*100)
	fmt.Printf("     - Average Run Time: %s\n", workflowMetrics.AverageRunTime)
	fmt.Printf("     - Throughput: %.1f executions/sec\n",
		workflowMetrics.PerformanceMetrics.ThroughputPerSecond)

	// Step 7: Pipeline Analytics
	fmt.Println("\n7. Analyzing CI/CD Pipeline Performance...")
	pipelineID := "pipeline-fullstack-deployment"

	pipelineMetrics, err := analyticsEngine.GetPipelineMetrics(pipelineID, timeRange)
	if err != nil {
		return fmt.Errorf("failed to get pipeline metrics: %w", err)
	}

	fmt.Printf("   ✓ Pipeline: %s\n", pipelineID)
	fmt.Printf("     - Total Builds: %d\n", pipelineMetrics.TotalBuilds)
	fmt.Printf("     - Success Rate: %.1f%%\n", pipelineMetrics.SuccessRate*100)
	fmt.Printf("     - Average Build Time: %s\n", pipelineMetrics.AverageBuildTime)
	fmt.Printf("     - Failed Builds: %d\n", pipelineMetrics.FailedBuilds)

	// Stage performance analysis
	fmt.Printf("   ✓ Stage Performance:\n")
	for _, stage := range pipelineMetrics.StageMetrics {
		fmt.Printf("     - %s: %.1f%% success, %s avg time\n",
			stage.StageName, stage.SuccessRate*100, stage.AverageTime)
	}

	// Step 8: Automation Impact Analysis
	fmt.Println("\n8. Analyzing Automation Impact...")

	automationMetrics, err := analyticsEngine.GetAutomationMetrics(timeRange)
	if err != nil {
		return fmt.Errorf("failed to get automation metrics: %w", err)
	}

	fmt.Printf("   ✓ Automation Overview:\n")
	fmt.Printf("     - Total Workflows: %d (%d active)\n",
		automationMetrics.TotalWorkflows, automationMetrics.ActiveWorkflows)
	fmt.Printf("     - Total Executions: %d\n", automationMetrics.TotalExecutions)
	fmt.Printf("     - Time Saved: %s\n", automationMetrics.AutomationSavings)
	fmt.Printf("     - Efficiency Gain: %.1f%%\n", automationMetrics.EfficiencyGain*100)
	fmt.Printf("     - ROI: %.1fx\n", automationMetrics.ROI)
	fmt.Printf("     - Adoption Rate: %.1f%%\n", automationMetrics.AdoptionRate*100)

	// Step 9: Project Comparison
	fmt.Println("\n9. Comparing Multiple Projects...")

	projectIDs := []string{"project-fullstack-app", "project-mobile-app", "project-api-service"}
	comparison, err := analyticsEngine.GetProjectComparison(projectIDs, timeRange)
	if err != nil {
		return fmt.Errorf("failed to get project comparison: %w", err)
	}

	fmt.Printf("   ✓ Project Comparison (%d projects):\n", len(comparison.Projects))
	fmt.Printf("     - Average Velocity: %.1f\n",
		comparison.Comparisons["average_velocity"].(float64))
	fmt.Printf("     - Average Completion Rate: %.1f%%\n",
		comparison.Comparisons["average_completion_rate"].(float64)*100)
	fmt.Printf("     - Best Velocity: %s\n",
		comparison.Comparisons["best_velocity"].(string))
	fmt.Printf("     - Worst Velocity: %s\n",
		comparison.Comparisons["worst_velocity"].(string))

	// Step 10: Report Generation
	fmt.Println("\n10. Generating Analytics Reports...")

	// Generate project report
	projectReportRequest := &analytics.ReportRequest{
		Type:        analytics.ReportTypeProject,
		Title:       "Monthly Project Performance Report",
		Description: "Comprehensive analysis of project performance for the last 30 days",
		TimeRange:   timeRange,
		Filters: map[string]interface{}{
			"project_id": projectID,
		},
		Format: analytics.ReportFormatHTML,
	}

	projectReport, err := analyticsEngine.GenerateReport(projectReportRequest)
	if err != nil {
		return fmt.Errorf("failed to generate project report: %w", err)
	}

	fmt.Printf("   ✓ Project Report Generated:\n")
	fmt.Printf("     - ID: %s\n", projectReport.ID)
	fmt.Printf("     - Type: %s\n", projectReport.Type)
	fmt.Printf("     - Format: %s\n", projectReport.Format)
	fmt.Printf("     - Size: %d bytes\n", projectReport.Size)
	fmt.Printf("     - URL: %s\n", projectReport.URL)

	// Generate team report
	teamReportRequest := &analytics.ReportRequest{
		Type:        analytics.ReportTypeTeam,
		Title:       "Team Performance Analysis",
		Description: "Team productivity and collaboration metrics",
		TimeRange:   timeRange,
		Filters: map[string]interface{}{
			"team_id": teamID,
		},
		Format: analytics.ReportFormatPDF,
	}

	teamReport, err := analyticsEngine.GenerateReport(teamReportRequest)
	if err != nil {
		return fmt.Errorf("failed to generate team report: %w", err)
	}

	fmt.Printf("   ✓ Team Report Generated:\n")
	fmt.Printf("     - ID: %s\n", teamReport.ID)
	fmt.Printf("     - Type: %s\n", teamReport.Type)
	fmt.Printf("     - Format: %s\n", teamReport.Format)
	fmt.Printf("     - Size: %d bytes\n", teamReport.Size)

	// Generate executive summary
	executiveReportRequest := &analytics.ReportRequest{
		Type:        analytics.ReportTypeExecutive,
		Title:       "Executive Summary - Q4 2024",
		Description: "High-level business impact and ROI analysis",
		TimeRange:   timeRange,
		Format:      analytics.ReportFormatPDF,
		Recipients:  []string{"ceo@company.com", "cto@company.com"},
	}

	executiveReport, err := analyticsEngine.GenerateReport(executiveReportRequest)
	if err != nil {
		return fmt.Errorf("failed to generate executive report: %w", err)
	}

	fmt.Printf("   ✓ Executive Report Generated:\n")
	fmt.Printf("     - ID: %s\n", executiveReport.ID)
	fmt.Printf("     - Type: %s\n", executiveReport.Type)
	fmt.Printf("     - Recipients: %v\n", executiveReportRequest.Recipients)

	// Step 11: Dashboard Management
	fmt.Println("\n11. Managing Analytics Dashboards...")

	// Create custom dashboard
	customDashboard := &analytics.Dashboard{
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
				RefreshRate: 300,
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
				RefreshRate: 600,
			},
		},
		IsPublic:  false,
		CreatedBy: "devops-team@company.com",
		Tags:      []string{"devops", "performance", "custom"},
	}

	createdDashboard, err := analyticsEngine.CreateDashboard(customDashboard)
	if err != nil {
		return fmt.Errorf("failed to create dashboard: %w", err)
	}

	fmt.Printf("   ✓ Custom Dashboard Created:\n")
	fmt.Printf("     - ID: %s\n", createdDashboard.ID)
	fmt.Printf("     - Name: %s\n", createdDashboard.Name)
	fmt.Printf("     - Widgets: %d\n", len(createdDashboard.Widgets))

	// List all dashboards
	dashboards, err := analyticsEngine.ListDashboards(&analytics.DashboardFilter{
		Limit: 10,
	})
	if err != nil {
		return fmt.Errorf("failed to list dashboards: %w", err)
	}

	fmt.Printf("   ✓ Available Dashboards (%d total):\n", len(dashboards))
	for _, dashboard := range dashboards {
		fmt.Printf("     - %s (%s) - %d widgets [%s]\n",
			dashboard.Name, dashboard.CreatedBy, len(dashboard.Widgets),
			strings.Join(dashboard.Tags, ", "))
	}

	// Step 12: Report Templates and Scheduling
	fmt.Println("\n12. Managing Report Templates and Scheduling...")

	// Get available templates
	templates, err := analyticsEngine.GetReportTemplates()
	if err != nil {
		return fmt.Errorf("failed to get report templates: %w", err)
	}

	fmt.Printf("   ✓ Available Report Templates (%d total):\n", len(templates))
	for _, template := range templates {
		fmt.Printf("     - %s (%s) - %s\n",
			template.Name, template.Type, template.Description)
	}

	// Schedule recurring report
	reportSchedule := &analytics.ReportSchedule{
		Name:       "Weekly Team Performance Report",
		ReportType: analytics.ReportTypeTeam,
		Schedule:   "0 9 * * MON", // Every Monday at 9 AM
		Recipients: []string{"team-lead@company.com", "manager@company.com"},
		Filters: map[string]interface{}{
			"team_id": teamID,
		},
		Format:  analytics.ReportFormatPDF,
		Enabled: true,
	}

	err = analyticsEngine.ScheduleReport(reportSchedule)
	if err != nil {
		return fmt.Errorf("failed to schedule report: %w", err)
	}

	fmt.Printf("   ✓ Report Scheduled:\n")
	fmt.Printf("     - Name: %s\n", reportSchedule.Name)
	fmt.Printf("     - Schedule: %s\n", reportSchedule.Schedule)
	fmt.Printf("     - Recipients: %v\n", reportSchedule.Recipients)
	fmt.Printf("     - Next Run: %s\n", reportSchedule.NextRun.Format("2006-01-02 15:04:05"))

	return nil
}
