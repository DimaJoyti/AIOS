package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aios/aios/pkg/project"
	"github.com/sirupsen/logrus"
)

func main() {
	// Initialize logger with text formatter for cleaner demo output
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)
	logger.SetFormatter(&logrus.TextFormatter{
		DisableTimestamp: true,
		ForceColors:      true,
	})

	fmt.Println("🚀 AIOS Project Management & Workflow Integration Demo")
	fmt.Println("=====================================================")

	// Run the comprehensive demo
	if err := runProjectWorkflowDemo(logger); err != nil {
		log.Fatalf("Demo failed: %v", err)
	}

	fmt.Println("\n✅ Project Management & Workflow Integration Demo completed successfully!")
}

func runProjectWorkflowDemo(logger *logrus.Logger) error {
	ctx := context.Background()

	// Step 1: Create Project Manager
	fmt.Println("\n1. Creating Project Manager...")
	config := &project.ProjectManagerConfig{
		EnableNotifications: true,
		EnableTimeTracking:  true,
		EnableIntegrations:  true,
		MaxProjectsPerUser:  100,
		MaxTasksPerProject:  1000,
	}

	projectManager, err := project.NewDefaultProjectManager(config, logger)
	if err != nil {
		return fmt.Errorf("failed to create project manager: %w", err)
	}
	fmt.Println("✓ Project Manager created successfully")

	// Step 2: Create a Software Development Project
	fmt.Println("\n2. Creating Software Development Project...")
	newProject := &project.Project{
		Name:        "AIOS Web Dashboard",
		Description: "A modern web dashboard for AIOS project management",
		Owner:       "john.doe@example.com",
		TeamMembers: []string{"jane.smith@example.com", "bob.wilson@example.com"},
		Priority:    project.PriorityHigh,
		Tags:        []string{"web", "dashboard", "react", "typescript"},
		Repository: &project.Repository{
			URL:      "https://github.com/aios/dashboard",
			Branch:   "main",
			Provider: "github",
		},
	}

	createdProject, err := projectManager.CreateProject(ctx, newProject)
	if err != nil {
		return fmt.Errorf("failed to create project: %w", err)
	}

	fmt.Printf("✓ Project created: %s (ID: %s)\n", createdProject.Name, createdProject.ID)

	// Step 3: Create Project Tasks
	fmt.Println("\n3. Creating Project Tasks...")
	tasks := []*project.Task{
		{
			ProjectID:   createdProject.ID,
			Title:       "Set up project repository",
			Description: "Initialize Git repository and set up basic project structure",
			Type:        project.TaskTypeTask,
			Priority:    project.PriorityHigh,
			Assignee:    "john.doe@example.com",
			Reporter:    "john.doe@example.com",
			Labels:      []string{"setup", "infrastructure"},
		},
		{
			ProjectID:   createdProject.ID,
			Title:       "Design system architecture",
			Description: "Create system architecture diagrams and technical specifications",
			Type:        project.TaskTypeTask,
			Priority:    project.PriorityHigh,
			Assignee:    "jane.smith@example.com",
			Reporter:    "john.doe@example.com",
			Labels:      []string{"design", "architecture"},
		},
		{
			ProjectID:   createdProject.ID,
			Title:       "Implement user authentication",
			Description: "Build secure user authentication system with JWT tokens",
			Type:        project.TaskTypeFeature,
			Priority:    project.PriorityMedium,
			Assignee:    "bob.wilson@example.com",
			Reporter:    "john.doe@example.com",
			Labels:      []string{"auth", "security", "backend"},
		},
		{
			ProjectID:   createdProject.ID,
			Title:       "Create responsive UI components",
			Description: "Build reusable React components with responsive design",
			Type:        project.TaskTypeFeature,
			Priority:    project.PriorityMedium,
			Assignee:    "jane.smith@example.com",
			Reporter:    "john.doe@example.com",
			Labels:      []string{"ui", "react", "frontend"},
		},
	}

	var createdTasks []*project.Task
	for _, task := range tasks {
		createdTask, err := projectManager.CreateTask(ctx, task)
		if err != nil {
			fmt.Printf("   ⚠️  Failed to create task '%s': %v\n", task.Title, err)
			continue
		}
		createdTasks = append(createdTasks, createdTask)
		fmt.Printf("   ✓ Task created: %s (ID: %s)\n", createdTask.Title, createdTask.ID)
	}

	// Step 4: Create Project Milestones
	fmt.Println("\n4. Creating Project Milestones...")
	milestones := []*project.Milestone{
		{
			ProjectID:   createdProject.ID,
			Name:        "Project Setup Complete",
			Description: "Repository setup and initial architecture design completed",
			DueDate:     timePtr(time.Now().AddDate(0, 0, 14)), // 2 weeks from now
		},
		{
			ProjectID:   createdProject.ID,
			Name:        "MVP Release",
			Description: "Minimum viable product with core features",
			DueDate:     timePtr(time.Now().AddDate(0, 2, 0)), // 2 months from now
		},
		{
			ProjectID:   createdProject.ID,
			Name:        "Production Release",
			Description: "Full production release with all features",
			DueDate:     timePtr(time.Now().AddDate(0, 4, 0)), // 4 months from now
		},
	}

	for _, milestone := range milestones {
		createdMilestone, err := projectManager.CreateMilestone(ctx, milestone)
		if err != nil {
			fmt.Printf("   ⚠️  Failed to create milestone '%s': %v\n", milestone.Name, err)
			continue
		}
		fmt.Printf("   ✓ Milestone created: %s (ID: %s)\n", createdMilestone.Name, createdMilestone.ID)
	}

	// Step 5: Create Sprint
	fmt.Println("\n5. Creating Sprint...")
	sprint := &project.Sprint{
		ProjectID: createdProject.ID,
		Name:      "Sprint 1 - Foundation",
		Goal:      "Set up project foundation and basic architecture",
		StartDate: time.Now(),
		EndDate:   time.Now().AddDate(0, 0, 14),                     // 2 weeks
		Capacity:  40,                                               // 40 story points
		Tasks:     []string{createdTasks[0].ID, createdTasks[1].ID}, // First two tasks
	}

	createdSprint, err := projectManager.CreateSprint(ctx, sprint)
	if err != nil {
		fmt.Printf("   ⚠️  Failed to create sprint: %v\n", err)
	} else {
		fmt.Printf("   ✓ Sprint created: %s (ID: %s)\n", createdSprint.Name, createdSprint.ID)
	}

	// Step 6: Add Team Members
	fmt.Println("\n6. Adding Team Members...")
	teamMembers := []*project.TeamMember{
		{
			Username: "john.doe",
			Email:    "john.doe@example.com",
			FullName: "John Doe",
			Role:     project.TeamRoleOwner,
			Skills:   []string{"Go", "React", "DevOps", "Architecture"},
			Timezone: "UTC-5",
			Status:   project.MemberStatusActive,
		},
		{
			Username: "jane.smith",
			Email:    "jane.smith@example.com",
			FullName: "Jane Smith",
			Role:     project.TeamRoleDeveloper,
			Skills:   []string{"React", "TypeScript", "UI/UX", "CSS"},
			Timezone: "UTC-8",
			Status:   project.MemberStatusActive,
		},
		{
			Username: "bob.wilson",
			Email:    "bob.wilson@example.com",
			FullName: "Bob Wilson",
			Role:     project.TeamRoleDeveloper,
			Skills:   []string{"Go", "PostgreSQL", "Docker", "Security"},
			Timezone: "UTC+1",
			Status:   project.MemberStatusActive,
		},
	}

	for _, member := range teamMembers {
		err := projectManager.AddTeamMember(ctx, createdProject.ID, member)
		if err != nil {
			fmt.Printf("   ⚠️  Failed to add team member '%s': %v\n", member.FullName, err)
			continue
		}
		fmt.Printf("   ✓ Team member added: %s (%s)\n", member.FullName, member.Role)
	}

	// Step 7: Simulate Task Progress
	fmt.Println("\n7. Simulating Task Progress...")
	if len(createdTasks) > 0 {
		// Start first task
		err := projectManager.UpdateTaskStatus(ctx, createdTasks[0].ID, project.TaskStatusInProgress)
		if err == nil {
			fmt.Printf("   ✓ Task '%s' started\n", createdTasks[0].Title)
		}

		// Complete first task
		time.Sleep(100 * time.Millisecond) // Simulate some work
		err = projectManager.UpdateTaskStatus(ctx, createdTasks[0].ID, project.TaskStatusDone)
		if err == nil {
			fmt.Printf("   ✓ Task '%s' completed\n", createdTasks[0].Title)
		}

		// Start second task
		if len(createdTasks) > 1 {
			err = projectManager.UpdateTaskStatus(ctx, createdTasks[1].ID, project.TaskStatusInProgress)
			if err == nil {
				fmt.Printf("   ✓ Task '%s' started\n", createdTasks[1].Title)
			}
		}
	}

	// Step 8: Generate Project Analytics
	fmt.Println("\n8. Generating Project Analytics...")
	timeRange := &project.TimeRange{
		Start: time.Now().AddDate(0, 0, -30), // Last 30 days
		End:   time.Now(),
	}

	analytics, err := projectManager.GetProjectAnalytics(ctx, createdProject.ID, timeRange)
	if err != nil {
		fmt.Printf("   ⚠️  Failed to generate analytics: %v\n", err)
	} else {
		fmt.Printf("   ✓ Project Analytics Generated:\n")
		fmt.Printf("     - Tasks Created: %d\n", analytics.TasksCreated)
		fmt.Printf("     - Tasks Completed: %d\n", analytics.TasksCompleted)
		fmt.Printf("     - Tasks In Progress: %d\n", analytics.TasksInProgress)
		fmt.Printf("     - Velocity: %.1f\n", analytics.Velocity)
	}

	// Step 9: List Projects and Tasks
	fmt.Println("\n9. Listing Projects and Tasks...")

	// List all projects
	projects, err := projectManager.ListProjects(ctx, &project.ProjectFilter{
		Limit: 10,
	})
	if err != nil {
		fmt.Printf("   ⚠️  Failed to list projects: %v\n", err)
	} else {
		fmt.Printf("   ✓ Found %d projects\n", len(projects))
		for _, proj := range projects {
			fmt.Printf("     - %s (%s) - %s\n", proj.Name, proj.Status, proj.Owner)
		}
	}

	// List project tasks
	projectTasks, err := projectManager.ListTasks(ctx, &project.TaskFilter{
		ProjectID: createdProject.ID,
		Limit:     20,
	})
	if err != nil {
		fmt.Printf("   ⚠️  Failed to list tasks: %v\n", err)
	} else {
		fmt.Printf("   ✓ Found %d tasks in project\n", len(projectTasks))
		for _, task := range projectTasks {
			fmt.Printf("     - %s (%s) - %s [%s]\n", task.Title, task.Status, task.Assignee, task.Priority)
		}
	}

	// Step 10: Generate Project Report
	fmt.Println("\n10. Generating Project Report...")
	reportRequest := &project.ReportRequest{
		Type:      project.ReportTypeProject,
		ProjectID: createdProject.ID,
		TimeRange: timeRange,
		Format:    project.ReportFormatJSON,
	}

	report, err := projectManager.GenerateReport(ctx, reportRequest)
	if err != nil {
		fmt.Printf("   ⚠️  Failed to generate report: %v\n", err)
	} else {
		fmt.Printf("   ✓ Project Report Generated:\n")
		fmt.Printf("     - Report ID: %s\n", report.ID)
		fmt.Printf("     - Type: %s\n", report.Type)
		fmt.Printf("     - Format: %s\n", report.Format)
		fmt.Printf("     - Generated At: %s\n", report.GeneratedAt.Format(time.RFC3339))
		fmt.Printf("     - Content Length: %d characters\n", len(report.Content))
	}

	return nil
}

// Helper function to create time pointer
func timePtr(t time.Time) *time.Time {
	return &t
}
