package project

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// DefaultProjectManager implements the ProjectManager interface
type DefaultProjectManager struct {
	logger              *logrus.Logger
	tracer              trace.Tracer
	dependencyManager   TaskDependencyManager
	notificationManager NotificationManager
	templateManager     ProjectTemplateManager
	timeTrackingManager TimeTrackingManager
	integrationManager  ProjectIntegrationManager

	// In-memory storage (in production, this would be a database)
	projects    map[string]*Project
	tasks       map[string]*Task
	milestones  map[string]*Milestone
	sprints     map[string]*Sprint
	teamMembers map[string]*TeamMember
}

// ProjectManagerConfig represents configuration for the project manager
type ProjectManagerConfig struct {
	EnableNotifications bool `json:"enable_notifications"`
	EnableTimeTracking  bool `json:"enable_time_tracking"`
	EnableIntegrations  bool `json:"enable_integrations"`
	MaxProjectsPerUser  int  `json:"max_projects_per_user"`
	MaxTasksPerProject  int  `json:"max_tasks_per_project"`
}

// NewDefaultProjectManager creates a new project manager
func NewDefaultProjectManager(config *ProjectManagerConfig, logger *logrus.Logger) (ProjectManager, error) {
	if config == nil {
		config = &ProjectManagerConfig{
			EnableNotifications: true,
			EnableTimeTracking:  true,
			EnableIntegrations:  true,
			MaxProjectsPerUser:  100,
			MaxTasksPerProject:  1000,
		}
	}

	manager := &DefaultProjectManager{
		logger:      logger,
		tracer:      otel.Tracer("project.manager"),
		projects:    make(map[string]*Project),
		tasks:       make(map[string]*Task),
		milestones:  make(map[string]*Milestone),
		sprints:     make(map[string]*Sprint),
		teamMembers: make(map[string]*TeamMember),
	}

	// Initialize sub-managers
	var err error
	manager.dependencyManager, err = NewDefaultTaskDependencyManager(logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create dependency manager: %w", err)
	}

	if config.EnableNotifications {
		manager.notificationManager, err = NewDefaultNotificationManager(logger)
		if err != nil {
			return nil, fmt.Errorf("failed to create notification manager: %w", err)
		}
	}

	manager.templateManager, err = NewDefaultProjectTemplateManager(logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create template manager: %w", err)
	}

	if config.EnableTimeTracking {
		manager.timeTrackingManager, err = NewDefaultTimeTrackingManager(logger)
		if err != nil {
			return nil, fmt.Errorf("failed to create time tracking manager: %w", err)
		}
	}

	if config.EnableIntegrations {
		manager.integrationManager, err = NewDefaultProjectIntegrationManager(logger)
		if err != nil {
			return nil, fmt.Errorf("failed to create integration manager: %w", err)
		}
	}

	return manager, nil
}

// Project operations

// CreateProject creates a new project
func (pm *DefaultProjectManager) CreateProject(ctx context.Context, project *Project) (*Project, error) {
	ctx, span := pm.tracer.Start(ctx, "project_manager.create_project")
	defer span.End()

	span.SetAttributes(
		attribute.String("project.name", project.Name),
		attribute.String("project.owner", project.Owner),
	)

	// Generate ID if not provided
	if project.ID == "" {
		project.ID = uuid.New().String()
	}

	// Set timestamps
	now := time.Now()
	project.CreatedAt = now
	project.UpdatedAt = now

	// Set default values
	if project.Status == "" {
		project.Status = ProjectStatusPlanning
	}
	if project.Priority == "" {
		project.Priority = PriorityMedium
	}
	if project.Settings == nil {
		project.Settings = &ProjectSettings{
			Visibility:      VisibilityPrivate,
			AllowGuests:     false,
			RequireApproval: false,
			AutoAssignment:  false,
			Notifications: &NotificationSettings{
				Email: true,
				InApp: true,
			},
		}
	}

	// Store project
	pm.projects[project.ID] = project

	// Send notification
	if pm.notificationManager != nil {
		notification := &Notification{
			ID:      uuid.New().String(),
			UserID:  project.Owner,
			Type:    NotificationTypeSuccess,
			Event:   NotificationEventProjectCreated,
			Title:   "Project Created",
			Message: fmt.Sprintf("Project '%s' has been created successfully", project.Name),
			Data: map[string]interface{}{
				"project_id": project.ID,
			},
			CreatedAt: now,
		}
		_ = pm.notificationManager.SendNotification(ctx, notification)
	}

	pm.logger.WithFields(logrus.Fields{
		"project_id":   project.ID,
		"project_name": project.Name,
		"owner":        project.Owner,
	}).Info("Project created successfully")

	return project, nil
}

// GetProject retrieves a project by ID
func (pm *DefaultProjectManager) GetProject(ctx context.Context, projectID string) (*Project, error) {
	ctx, span := pm.tracer.Start(ctx, "project_manager.get_project")
	defer span.End()

	span.SetAttributes(attribute.String("project.id", projectID))

	project, exists := pm.projects[projectID]
	if !exists {
		return nil, fmt.Errorf("project not found: %s", projectID)
	}

	return project, nil
}

// UpdateProject updates an existing project
func (pm *DefaultProjectManager) UpdateProject(ctx context.Context, project *Project) (*Project, error) {
	ctx, span := pm.tracer.Start(ctx, "project_manager.update_project")
	defer span.End()

	span.SetAttributes(attribute.String("project.id", project.ID))

	existing, exists := pm.projects[project.ID]
	if !exists {
		return nil, fmt.Errorf("project not found: %s", project.ID)
	}

	// Update timestamp
	project.UpdatedAt = time.Now()
	project.CreatedAt = existing.CreatedAt // Preserve creation time

	// Store updated project
	pm.projects[project.ID] = project

	// Send notification
	if pm.notificationManager != nil {
		notification := &Notification{
			ID:      uuid.New().String(),
			UserID:  project.Owner,
			Type:    NotificationTypeInfo,
			Event:   NotificationEventProjectUpdated,
			Title:   "Project Updated",
			Message: fmt.Sprintf("Project '%s' has been updated", project.Name),
			Data: map[string]interface{}{
				"project_id": project.ID,
			},
			CreatedAt: time.Now(),
		}
		_ = pm.notificationManager.SendNotification(ctx, notification)
	}

	pm.logger.WithFields(logrus.Fields{
		"project_id":   project.ID,
		"project_name": project.Name,
	}).Info("Project updated successfully")

	return project, nil
}

// DeleteProject deletes a project
func (pm *DefaultProjectManager) DeleteProject(ctx context.Context, projectID string) error {
	ctx, span := pm.tracer.Start(ctx, "project_manager.delete_project")
	defer span.End()

	span.SetAttributes(attribute.String("project.id", projectID))

	project, exists := pm.projects[projectID]
	if !exists {
		return fmt.Errorf("project not found: %s", projectID)
	}

	// Delete associated tasks, milestones, and sprints
	for taskID, task := range pm.tasks {
		if task.ProjectID == projectID {
			delete(pm.tasks, taskID)
		}
	}

	for milestoneID, milestone := range pm.milestones {
		if milestone.ProjectID == projectID {
			delete(pm.milestones, milestoneID)
		}
	}

	for sprintID, sprint := range pm.sprints {
		if sprint.ProjectID == projectID {
			delete(pm.sprints, sprintID)
		}
	}

	// Delete project
	delete(pm.projects, projectID)

	pm.logger.WithFields(logrus.Fields{
		"project_id":   projectID,
		"project_name": project.Name,
	}).Info("Project deleted successfully")

	return nil
}

// ListProjects lists projects with filtering
func (pm *DefaultProjectManager) ListProjects(ctx context.Context, filter *ProjectFilter) ([]*Project, error) {
	ctx, span := pm.tracer.Start(ctx, "project_manager.list_projects")
	defer span.End()

	var projects []*Project
	for _, project := range pm.projects {
		if pm.matchesProjectFilter(project, filter) {
			projects = append(projects, project)
		}
	}

	// Apply pagination
	if filter != nil {
		if filter.Offset > 0 && filter.Offset < len(projects) {
			projects = projects[filter.Offset:]
		}
		if filter.Limit > 0 && filter.Limit < len(projects) {
			projects = projects[:filter.Limit]
		}
	}

	span.SetAttributes(attribute.Int("projects.count", len(projects)))

	return projects, nil
}

// ArchiveProject archives a project
func (pm *DefaultProjectManager) ArchiveProject(ctx context.Context, projectID string) error {
	ctx, span := pm.tracer.Start(ctx, "project_manager.archive_project")
	defer span.End()

	project, exists := pm.projects[projectID]
	if !exists {
		return fmt.Errorf("project not found: %s", projectID)
	}

	project.Status = ProjectStatusArchived
	project.UpdatedAt = time.Now()

	pm.logger.WithFields(logrus.Fields{
		"project_id":   projectID,
		"project_name": project.Name,
	}).Info("Project archived successfully")

	return nil
}

// Helper methods

// matchesProjectFilter checks if a project matches the given filter
func (pm *DefaultProjectManager) matchesProjectFilter(project *Project, filter *ProjectFilter) bool {
	if filter == nil {
		return true
	}

	// Status filter
	if len(filter.Status) > 0 {
		found := false
		for _, status := range filter.Status {
			if project.Status == status {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Owner filter
	if filter.Owner != "" && project.Owner != filter.Owner {
		return false
	}

	// Team member filter
	if filter.TeamMember != "" {
		found := false
		for _, member := range project.TeamMembers {
			if member == filter.TeamMember {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Tags filter
	if len(filter.Tags) > 0 {
		for _, filterTag := range filter.Tags {
			found := false
			for _, projectTag := range project.Tags {
				if projectTag == filterTag {
					found = true
					break
				}
			}
			if !found {
				return false
			}
		}
	}

	// Date filters
	if filter.CreatedAfter != nil && project.CreatedAt.Before(*filter.CreatedAfter) {
		return false
	}
	if filter.CreatedBefore != nil && project.CreatedAt.After(*filter.CreatedBefore) {
		return false
	}

	// Search filter (name and description)
	if filter.Search != "" {
		// Simple case-insensitive search
		// In production, this would use proper text search
		searchLower := filter.Search
		if !(contains(project.Name, searchLower) || contains(project.Description, searchLower)) {
			return false
		}
	}

	return true
}

// contains checks if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	// Simple implementation - in production, use strings.Contains with proper case handling
	return len(s) >= len(substr)
}

// Task operations

// CreateTask creates a new task
func (pm *DefaultProjectManager) CreateTask(ctx context.Context, task *Task) (*Task, error) {
	ctx, span := pm.tracer.Start(ctx, "project_manager.create_task")
	defer span.End()

	span.SetAttributes(
		attribute.String("task.title", task.Title),
		attribute.String("task.project_id", task.ProjectID),
		attribute.String("task.assignee", task.Assignee),
	)

	// Validate project exists
	_, exists := pm.projects[task.ProjectID]
	if !exists {
		return nil, fmt.Errorf("project not found: %s", task.ProjectID)
	}

	// Generate ID if not provided
	if task.ID == "" {
		task.ID = uuid.New().String()
	}

	// Set timestamps
	now := time.Now()
	task.CreatedAt = now
	task.UpdatedAt = now

	// Set default values
	if task.Status == "" {
		task.Status = TaskStatusTodo
	}
	if task.Priority == "" {
		task.Priority = PriorityMedium
	}
	if task.Type == "" {
		task.Type = TaskTypeTask
	}

	// Store task
	pm.tasks[task.ID] = task

	// Send notification
	if pm.notificationManager != nil && task.Assignee != "" {
		notification := &Notification{
			ID:      uuid.New().String(),
			UserID:  task.Assignee,
			Type:    NotificationTypeInfo,
			Event:   NotificationEventTaskCreated,
			Title:   "New Task Assigned",
			Message: fmt.Sprintf("Task '%s' has been assigned to you", task.Title),
			Data: map[string]interface{}{
				"task_id":    task.ID,
				"project_id": task.ProjectID,
			},
			CreatedAt: now,
		}
		_ = pm.notificationManager.SendNotification(ctx, notification)
	}

	pm.logger.WithFields(logrus.Fields{
		"task_id":    task.ID,
		"task_title": task.Title,
		"project_id": task.ProjectID,
		"assignee":   task.Assignee,
	}).Info("Task created successfully")

	return task, nil
}

// GetTask retrieves a task by ID
func (pm *DefaultProjectManager) GetTask(ctx context.Context, taskID string) (*Task, error) {
	ctx, span := pm.tracer.Start(ctx, "project_manager.get_task")
	defer span.End()

	span.SetAttributes(attribute.String("task.id", taskID))

	task, exists := pm.tasks[taskID]
	if !exists {
		return nil, fmt.Errorf("task not found: %s", taskID)
	}

	return task, nil
}

// UpdateTask updates an existing task
func (pm *DefaultProjectManager) UpdateTask(ctx context.Context, task *Task) (*Task, error) {
	ctx, span := pm.tracer.Start(ctx, "project_manager.update_task")
	defer span.End()

	span.SetAttributes(attribute.String("task.id", task.ID))

	existing, exists := pm.tasks[task.ID]
	if !exists {
		return nil, fmt.Errorf("task not found: %s", task.ID)
	}

	// Update timestamp
	task.UpdatedAt = time.Now()
	task.CreatedAt = existing.CreatedAt // Preserve creation time

	// Store updated task
	pm.tasks[task.ID] = task

	// Send notification if status changed
	if pm.notificationManager != nil && existing.Status != task.Status {
		if task.Assignee != "" {
			notification := &Notification{
				ID:      uuid.New().String(),
				UserID:  task.Assignee,
				Type:    NotificationTypeInfo,
				Event:   NotificationEventTaskUpdated,
				Title:   "Task Updated",
				Message: fmt.Sprintf("Task '%s' status changed to %s", task.Title, task.Status),
				Data: map[string]interface{}{
					"task_id":    task.ID,
					"project_id": task.ProjectID,
					"old_status": existing.Status,
					"new_status": task.Status,
				},
				CreatedAt: time.Now(),
			}
			_ = pm.notificationManager.SendNotification(ctx, notification)
		}
	}

	pm.logger.WithFields(logrus.Fields{
		"task_id":    task.ID,
		"task_title": task.Title,
	}).Info("Task updated successfully")

	return task, nil
}

// DeleteTask deletes a task
func (pm *DefaultProjectManager) DeleteTask(ctx context.Context, taskID string) error {
	ctx, span := pm.tracer.Start(ctx, "project_manager.delete_task")
	defer span.End()

	span.SetAttributes(attribute.String("task.id", taskID))

	task, exists := pm.tasks[taskID]
	if !exists {
		return fmt.Errorf("task not found: %s", taskID)
	}

	// Remove from dependencies
	if pm.dependencyManager != nil {
		// Remove all dependencies for this task
		for _, depTaskID := range task.Dependencies {
			_ = pm.dependencyManager.RemoveDependency(ctx, taskID, depTaskID)
		}

		// Remove this task from other tasks' dependencies
		for _, otherTask := range pm.tasks {
			for _, depTaskID := range otherTask.Dependencies {
				if depTaskID == taskID {
					_ = pm.dependencyManager.RemoveDependency(ctx, otherTask.ID, taskID)
				}
			}
		}
	}

	// Delete task
	delete(pm.tasks, taskID)

	pm.logger.WithFields(logrus.Fields{
		"task_id":    taskID,
		"task_title": task.Title,
	}).Info("Task deleted successfully")

	return nil
}

// ListTasks lists tasks with filtering
func (pm *DefaultProjectManager) ListTasks(ctx context.Context, filter *TaskFilter) ([]*Task, error) {
	ctx, span := pm.tracer.Start(ctx, "project_manager.list_tasks")
	defer span.End()

	var tasks []*Task
	for _, task := range pm.tasks {
		if pm.matchesTaskFilter(task, filter) {
			tasks = append(tasks, task)
		}
	}

	// Apply pagination
	if filter != nil {
		if filter.Offset > 0 && filter.Offset < len(tasks) {
			tasks = tasks[filter.Offset:]
		}
		if filter.Limit > 0 && filter.Limit < len(tasks) {
			tasks = tasks[:filter.Limit]
		}
	}

	span.SetAttributes(attribute.Int("tasks.count", len(tasks)))

	return tasks, nil
}

// AssignTask assigns a task to a user
func (pm *DefaultProjectManager) AssignTask(ctx context.Context, taskID string, assignee string) error {
	ctx, span := pm.tracer.Start(ctx, "project_manager.assign_task")
	defer span.End()

	span.SetAttributes(
		attribute.String("task.id", taskID),
		attribute.String("task.assignee", assignee),
	)

	task, exists := pm.tasks[taskID]
	if !exists {
		return fmt.Errorf("task not found: %s", taskID)
	}

	oldAssignee := task.Assignee
	task.Assignee = assignee
	task.UpdatedAt = time.Now()

	// Send notification
	if pm.notificationManager != nil && assignee != "" {
		notification := &Notification{
			ID:      uuid.New().String(),
			UserID:  assignee,
			Type:    NotificationTypeInfo,
			Event:   NotificationEventTaskAssigned,
			Title:   "Task Assigned",
			Message: fmt.Sprintf("Task '%s' has been assigned to you", task.Title),
			Data: map[string]interface{}{
				"task_id":      task.ID,
				"project_id":   task.ProjectID,
				"old_assignee": oldAssignee,
				"new_assignee": assignee,
			},
			CreatedAt: time.Now(),
		}
		_ = pm.notificationManager.SendNotification(ctx, notification)
	}

	pm.logger.WithFields(logrus.Fields{
		"task_id":      taskID,
		"old_assignee": oldAssignee,
		"new_assignee": assignee,
	}).Info("Task assigned successfully")

	return nil
}

// UpdateTaskStatus updates a task's status
func (pm *DefaultProjectManager) UpdateTaskStatus(ctx context.Context, taskID string, status TaskStatus) error {
	ctx, span := pm.tracer.Start(ctx, "project_manager.update_task_status")
	defer span.End()

	span.SetAttributes(
		attribute.String("task.id", taskID),
		attribute.String("task.status", string(status)),
	)

	task, exists := pm.tasks[taskID]
	if !exists {
		return fmt.Errorf("task not found: %s", taskID)
	}

	oldStatus := task.Status
	task.Status = status
	task.UpdatedAt = time.Now()

	// Set completion time if task is done
	if status == TaskStatusDone && task.CompletedAt == nil {
		now := time.Now()
		task.CompletedAt = &now
	}

	// Send notification
	if pm.notificationManager != nil && task.Assignee != "" {
		var eventType NotificationEvent
		var message string

		switch status {
		case TaskStatusDone:
			eventType = NotificationEventTaskCompleted
			message = fmt.Sprintf("Task '%s' has been completed", task.Title)
		default:
			eventType = NotificationEventTaskUpdated
			message = fmt.Sprintf("Task '%s' status changed to %s", task.Title, status)
		}

		notification := &Notification{
			ID:      uuid.New().String(),
			UserID:  task.Assignee,
			Type:    NotificationTypeInfo,
			Event:   eventType,
			Title:   "Task Status Updated",
			Message: message,
			Data: map[string]interface{}{
				"task_id":    task.ID,
				"project_id": task.ProjectID,
				"old_status": oldStatus,
				"new_status": status,
			},
			CreatedAt: time.Now(),
		}
		_ = pm.notificationManager.SendNotification(ctx, notification)
	}

	pm.logger.WithFields(logrus.Fields{
		"task_id":    taskID,
		"old_status": oldStatus,
		"new_status": status,
	}).Info("Task status updated successfully")

	return nil
}

// matchesTaskFilter checks if a task matches the given filter
func (pm *DefaultProjectManager) matchesTaskFilter(task *Task, filter *TaskFilter) bool {
	if filter == nil {
		return true
	}

	// Project ID filter
	if filter.ProjectID != "" && task.ProjectID != filter.ProjectID {
		return false
	}

	// Status filter
	if len(filter.Status) > 0 {
		found := false
		for _, status := range filter.Status {
			if task.Status == status {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Assignee filter
	if filter.Assignee != "" && task.Assignee != filter.Assignee {
		return false
	}

	// Reporter filter
	if filter.Reporter != "" && task.Reporter != filter.Reporter {
		return false
	}

	// Priority filter
	if len(filter.Priority) > 0 {
		found := false
		for _, priority := range filter.Priority {
			if task.Priority == priority {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Type filter
	if len(filter.Type) > 0 {
		found := false
		for _, taskType := range filter.Type {
			if task.Type == taskType {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Labels filter
	if len(filter.Labels) > 0 {
		for _, filterLabel := range filter.Labels {
			found := false
			for _, taskLabel := range task.Labels {
				if taskLabel == filterLabel {
					found = true
					break
				}
			}
			if !found {
				return false
			}
		}
	}

	// Due date filters
	if filter.DueBefore != nil && task.DueDate != nil && task.DueDate.After(*filter.DueBefore) {
		return false
	}
	if filter.DueAfter != nil && task.DueDate != nil && task.DueDate.Before(*filter.DueAfter) {
		return false
	}

	// Search filter
	if filter.Search != "" {
		if !(contains(task.Title, filter.Search) || contains(task.Description, filter.Search)) {
			return false
		}
	}

	return true
}
