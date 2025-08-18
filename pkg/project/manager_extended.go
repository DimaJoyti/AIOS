package project

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/attribute"
)

// Extended methods for DefaultProjectManager

// Milestone operations

// CreateMilestone creates a new milestone
func (pm *DefaultProjectManager) CreateMilestone(ctx context.Context, milestone *Milestone) (*Milestone, error) {
	ctx, span := pm.tracer.Start(ctx, "project_manager.create_milestone")
	defer span.End()

	// Validate project exists
	_, exists := pm.projects[milestone.ProjectID]
	if !exists {
		return nil, fmt.Errorf("project not found: %s", milestone.ProjectID)
	}

	// Generate ID if not provided
	if milestone.ID == "" {
		milestone.ID = uuid.New().String()
	}

	// Set timestamps
	now := time.Now()
	milestone.CreatedAt = now
	milestone.UpdatedAt = now

	// Set default values
	if milestone.Status == "" {
		milestone.Status = MilestoneStatusOpen
	}

	// Store milestone
	pm.milestones[milestone.ID] = milestone

	pm.logger.WithFields(logrus.Fields{
		"milestone_id":   milestone.ID,
		"milestone_name": milestone.Name,
		"project_id":     milestone.ProjectID,
	}).Info("Milestone created successfully")

	return milestone, nil
}

// GetMilestone retrieves a milestone by ID
func (pm *DefaultProjectManager) GetMilestone(ctx context.Context, milestoneID string) (*Milestone, error) {
	milestone, exists := pm.milestones[milestoneID]
	if !exists {
		return nil, fmt.Errorf("milestone not found: %s", milestoneID)
	}
	return milestone, nil
}

// UpdateMilestone updates an existing milestone
func (pm *DefaultProjectManager) UpdateMilestone(ctx context.Context, milestone *Milestone) (*Milestone, error) {
	existing, exists := pm.milestones[milestone.ID]
	if !exists {
		return nil, fmt.Errorf("milestone not found: %s", milestone.ID)
	}

	milestone.UpdatedAt = time.Now()
	milestone.CreatedAt = existing.CreatedAt // Preserve creation time

	pm.milestones[milestone.ID] = milestone
	return milestone, nil
}

// DeleteMilestone deletes a milestone
func (pm *DefaultProjectManager) DeleteMilestone(ctx context.Context, milestoneID string) error {
	_, exists := pm.milestones[milestoneID]
	if !exists {
		return fmt.Errorf("milestone not found: %s", milestoneID)
	}

	delete(pm.milestones, milestoneID)
	return nil
}

// ListMilestones lists milestones for a project
func (pm *DefaultProjectManager) ListMilestones(ctx context.Context, projectID string) ([]*Milestone, error) {
	var milestones []*Milestone
	for _, milestone := range pm.milestones {
		if milestone.ProjectID == projectID {
			milestones = append(milestones, milestone)
		}
	}
	return milestones, nil
}

// Sprint operations

// CreateSprint creates a new sprint
func (pm *DefaultProjectManager) CreateSprint(ctx context.Context, sprint *Sprint) (*Sprint, error) {
	ctx, span := pm.tracer.Start(ctx, "project_manager.create_sprint")
	defer span.End()

	// Validate project exists
	_, exists := pm.projects[sprint.ProjectID]
	if !exists {
		return nil, fmt.Errorf("project not found: %s", sprint.ProjectID)
	}

	// Generate ID if not provided
	if sprint.ID == "" {
		sprint.ID = uuid.New().String()
	}

	// Set timestamps
	now := time.Now()
	sprint.CreatedAt = now
	sprint.UpdatedAt = now

	// Set default values
	if sprint.Status == "" {
		sprint.Status = SprintStatusPlanning
	}

	// Store sprint
	pm.sprints[sprint.ID] = sprint

	pm.logger.WithFields(logrus.Fields{
		"sprint_id":   sprint.ID,
		"sprint_name": sprint.Name,
		"project_id":  sprint.ProjectID,
	}).Info("Sprint created successfully")

	return sprint, nil
}

// GetSprint retrieves a sprint by ID
func (pm *DefaultProjectManager) GetSprint(ctx context.Context, sprintID string) (*Sprint, error) {
	sprint, exists := pm.sprints[sprintID]
	if !exists {
		return nil, fmt.Errorf("sprint not found: %s", sprintID)
	}
	return sprint, nil
}

// UpdateSprint updates an existing sprint
func (pm *DefaultProjectManager) UpdateSprint(ctx context.Context, sprint *Sprint) (*Sprint, error) {
	existing, exists := pm.sprints[sprint.ID]
	if !exists {
		return nil, fmt.Errorf("sprint not found: %s", sprint.ID)
	}

	sprint.UpdatedAt = time.Now()
	sprint.CreatedAt = existing.CreatedAt // Preserve creation time

	pm.sprints[sprint.ID] = sprint
	return sprint, nil
}

// DeleteSprint deletes a sprint
func (pm *DefaultProjectManager) DeleteSprint(ctx context.Context, sprintID string) error {
	_, exists := pm.sprints[sprintID]
	if !exists {
		return fmt.Errorf("sprint not found: %s", sprintID)
	}

	delete(pm.sprints, sprintID)
	return nil
}

// ListSprints lists sprints for a project
func (pm *DefaultProjectManager) ListSprints(ctx context.Context, projectID string) ([]*Sprint, error) {
	var sprints []*Sprint
	for _, sprint := range pm.sprints {
		if sprint.ProjectID == projectID {
			sprints = append(sprints, sprint)
		}
	}
	return sprints, nil
}

// StartSprint starts a sprint
func (pm *DefaultProjectManager) StartSprint(ctx context.Context, sprintID string) error {
	sprint, exists := pm.sprints[sprintID]
	if !exists {
		return fmt.Errorf("sprint not found: %s", sprintID)
	}

	sprint.Status = SprintStatusActive
	sprint.UpdatedAt = time.Now()

	pm.logger.WithField("sprint_id", sprintID).Info("Sprint started")
	return nil
}

// CompleteSprint completes a sprint
func (pm *DefaultProjectManager) CompleteSprint(ctx context.Context, sprintID string) error {
	sprint, exists := pm.sprints[sprintID]
	if !exists {
		return fmt.Errorf("sprint not found: %s", sprintID)
	}

	sprint.Status = SprintStatusCompleted
	sprint.UpdatedAt = time.Now()

	pm.logger.WithField("sprint_id", sprintID).Info("Sprint completed")
	return nil
}

// Team operations

// AddTeamMember adds a team member to a project
func (pm *DefaultProjectManager) AddTeamMember(ctx context.Context, projectID string, member *TeamMember) error {
	ctx, span := pm.tracer.Start(ctx, "project_manager.add_team_member")
	defer span.End()

	project, exists := pm.projects[projectID]
	if !exists {
		return fmt.Errorf("project not found: %s", projectID)
	}

	// Generate ID if not provided
	if member.ID == "" {
		member.ID = uuid.New().String()
	}

	// Set timestamps
	now := time.Now()
	member.JoinedAt = now
	member.UpdatedAt = now

	// Store team member
	pm.teamMembers[member.ID] = member

	// Add to project team members list
	project.TeamMembers = append(project.TeamMembers, member.ID)
	project.UpdatedAt = now

	pm.logger.WithFields(logrus.Fields{
		"member_id":   member.ID,
		"member_name": member.FullName,
		"project_id":  projectID,
		"role":        member.Role,
	}).Info("Team member added successfully")

	return nil
}

// RemoveTeamMember removes a team member from a project
func (pm *DefaultProjectManager) RemoveTeamMember(ctx context.Context, projectID string, memberID string) error {
	project, exists := pm.projects[projectID]
	if !exists {
		return fmt.Errorf("project not found: %s", projectID)
	}

	// Remove from project team members list
	for i, id := range project.TeamMembers {
		if id == memberID {
			project.TeamMembers = append(project.TeamMembers[:i], project.TeamMembers[i+1:]...)
			break
		}
	}

	// Remove from team members storage
	delete(pm.teamMembers, memberID)

	project.UpdatedAt = time.Now()

	pm.logger.WithFields(logrus.Fields{
		"member_id":  memberID,
		"project_id": projectID,
	}).Info("Team member removed successfully")

	return nil
}

// UpdateTeamMember updates a team member
func (pm *DefaultProjectManager) UpdateTeamMember(ctx context.Context, member *TeamMember) error {
	existing, exists := pm.teamMembers[member.ID]
	if !exists {
		return fmt.Errorf("team member not found: %s", member.ID)
	}

	member.UpdatedAt = time.Now()
	member.JoinedAt = existing.JoinedAt // Preserve join time

	pm.teamMembers[member.ID] = member
	return nil
}

// GetTeamMember retrieves a team member by ID
func (pm *DefaultProjectManager) GetTeamMember(ctx context.Context, memberID string) (*TeamMember, error) {
	member, exists := pm.teamMembers[memberID]
	if !exists {
		return nil, fmt.Errorf("team member not found: %s", memberID)
	}
	return member, nil
}

// ListTeamMembers lists team members for a project
func (pm *DefaultProjectManager) ListTeamMembers(ctx context.Context, projectID string) ([]*TeamMember, error) {
	project, exists := pm.projects[projectID]
	if !exists {
		return nil, fmt.Errorf("project not found: %s", projectID)
	}

	var members []*TeamMember
	for _, memberID := range project.TeamMembers {
		if member, exists := pm.teamMembers[memberID]; exists {
			members = append(members, member)
		}
	}

	return members, nil
}

// Analytics and reporting

// GetProjectAnalytics generates project analytics
func (pm *DefaultProjectManager) GetProjectAnalytics(ctx context.Context, projectID string, timeRange *TimeRange) (*ProjectAnalytics, error) {
	ctx, span := pm.tracer.Start(ctx, "project_manager.get_project_analytics")
	defer span.End()

	span.SetAttributes(attribute.String("project.id", projectID))

	// Count tasks by status
	var tasksCreated, tasksCompleted, tasksInProgress int
	statusDistribution := make(map[TaskStatus]int)
	priorityDistribution := make(map[Priority]int)

	for _, task := range pm.tasks {
		if task.ProjectID != projectID {
			continue
		}

		// Check if task is within time range
		if timeRange != nil {
			if task.CreatedAt.Before(timeRange.Start) || task.CreatedAt.After(timeRange.End) {
				continue
			}
		}

		tasksCreated++
		statusDistribution[task.Status]++
		priorityDistribution[task.Priority]++

		switch task.Status {
		case TaskStatusDone:
			tasksCompleted++
		case TaskStatusInProgress:
			tasksInProgress++
		}
	}

	// Calculate velocity (simplified)
	velocity := float32(tasksCompleted) / 7.0 // tasks per week

	analytics := &ProjectAnalytics{
		ProjectID:            projectID,
		TimeRange:            timeRange,
		TasksCreated:         tasksCreated,
		TasksCompleted:       tasksCompleted,
		TasksInProgress:      tasksInProgress,
		AverageLeadTime:      24 * time.Hour, // Simplified
		AverageCycleTime:     8 * time.Hour,  // Simplified
		Velocity:             velocity,
		BurndownData:         []*BurndownPoint{},
		StatusDistribution:   statusDistribution,
		PriorityDistribution: priorityDistribution,
		Metadata:             make(map[string]interface{}),
	}

	return analytics, nil
}

// GetTaskAnalytics generates task analytics
func (pm *DefaultProjectManager) GetTaskAnalytics(ctx context.Context, projectID string, timeRange *TimeRange) (*TaskAnalytics, error) {
	// Simplified implementation
	return &TaskAnalytics{
		ProjectID:             projectID,
		TimeRange:             timeRange,
		CompletionRate:        0.75,
		AverageTimeToComplete: 3 * 24 * time.Hour,
		TasksByAssignee:       make(map[string]int),
		TasksByType:           make(map[TaskType]int),
		TasksByPriority:       make(map[Priority]int),
		OverdueTasks:          0,
		Metadata:              make(map[string]interface{}),
	}, nil
}

// GetTeamAnalytics generates team analytics
func (pm *DefaultProjectManager) GetTeamAnalytics(ctx context.Context, projectID string, timeRange *TimeRange) (*TeamAnalytics, error) {
	project, exists := pm.projects[projectID]
	if !exists {
		return nil, fmt.Errorf("project not found: %s", projectID)
	}

	return &TeamAnalytics{
		ProjectID:            projectID,
		TimeRange:            timeRange,
		TeamSize:             len(project.TeamMembers),
		ActiveMembers:        len(project.TeamMembers),
		WorkloadDistribution: make(map[string]float32),
		ProductivityMetrics: &ProductivityMetrics{
			TasksPerMember:      5.0,
			AverageTaskDuration: 2 * 24 * time.Hour,
			CodeCommits:         50,
			CodeReviews:         20,
			BugsFixed:           10,
		},
		CollaborationMetrics: &CollaborationMetrics{
			CommentsPerTask:     3.5,
			ReviewParticipation: 0.8,
			KnowledgeSharing:    0.7,
			CrossTeamWork:       0.6,
		},
		Metadata: make(map[string]interface{}),
	}, nil
}

// GenerateReport generates a project report
func (pm *DefaultProjectManager) GenerateReport(ctx context.Context, request *ReportRequest) (*Report, error) {
	ctx, span := pm.tracer.Start(ctx, "project_manager.generate_report")
	defer span.End()

	report := &Report{
		ID:          uuid.New().String(),
		Type:        request.Type,
		Title:       fmt.Sprintf("%s Report", request.Type),
		Format:      request.Format,
		GeneratedAt: time.Now(),
		GeneratedBy: "system",
		Metadata:    make(map[string]interface{}),
	}

	// Generate content based on report type
	switch request.Type {
	case ReportTypeProject:
		analytics, err := pm.GetProjectAnalytics(ctx, request.ProjectID, request.TimeRange)
		if err != nil {
			return nil, fmt.Errorf("failed to get project analytics: %w", err)
		}
		report.Content = fmt.Sprintf(`Project Report
=============

Project ID: %s
Time Range: %s to %s

Tasks Created: %d
Tasks Completed: %d
Tasks In Progress: %d
Velocity: %.1f tasks/week

Status Distribution:
- Todo: %d
- In Progress: %d
- Done: %d
`,
			analytics.ProjectID,
			analytics.TimeRange.Start.Format("2006-01-02"),
			analytics.TimeRange.End.Format("2006-01-02"),
			analytics.TasksCreated,
			analytics.TasksCompleted,
			analytics.TasksInProgress,
			analytics.Velocity,
			analytics.StatusDistribution[TaskStatusTodo],
			analytics.StatusDistribution[TaskStatusInProgress],
			analytics.StatusDistribution[TaskStatusDone],
		)

	default:
		report.Content = "Report content not implemented for this type"
	}

	pm.logger.WithFields(logrus.Fields{
		"report_id":   report.ID,
		"report_type": report.Type,
		"project_id":  request.ProjectID,
	}).Info("Report generated successfully")

	return report, nil
}
