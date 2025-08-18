package collaboration

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Helper methods for collaboration engine

// Validation methods

// validateTeam validates a team configuration
func (ce *DefaultCollaborationEngine) validateTeam(team *Team) error {
	if team.Name == "" {
		return fmt.Errorf("team name is required")
	}

	if len(team.Name) > 100 {
		return fmt.Errorf("team name must be 100 characters or less")
	}

	if team.Type == "" {
		team.Type = TeamTypeDevelopment
	}

	if team.Status == "" {
		team.Status = TeamStatusActive
	}

	return nil
}

// validateRole validates a role configuration
func (ce *DefaultCollaborationEngine) validateRole(role *Role) error {
	if role.Name == "" {
		return fmt.Errorf("role name is required")
	}

	if len(role.Name) > 50 {
		return fmt.Errorf("role name must be 50 characters or less")
	}

	if role.Type == "" {
		role.Type = RoleTypeCustom
	}

	// Validate permissions
	for _, permission := range role.Permissions {
		if permission.Name == "" {
			return fmt.Errorf("permission name is required")
		}
		if permission.Resource == "" {
			return fmt.Errorf("permission resource is required")
		}
		if permission.Action == "" {
			return fmt.Errorf("permission action is required")
		}
	}

	return nil
}

// validateChannel validates a channel configuration
func (ce *DefaultCollaborationEngine) validateChannel(channel *Channel) error {
	if channel.Name == "" {
		return fmt.Errorf("channel name is required")
	}

	if len(channel.Name) > 50 {
		return fmt.Errorf("channel name must be 50 characters or less")
	}

	if channel.Type == "" {
		channel.Type = ChannelTypePublic
	}

	if channel.TeamID == "" {
		return fmt.Errorf("team ID is required")
	}

	return nil
}

// validateMessage validates a message
func (ce *DefaultCollaborationEngine) validateMessage(message *Message) error {
	if message.Content == "" {
		return fmt.Errorf("message content is required")
	}

	if len(message.Content) > 4000 {
		return fmt.Errorf("message content must be 4000 characters or less")
	}

	if message.ChannelID == "" {
		return fmt.Errorf("channel ID is required")
	}

	if message.UserID == "" {
		return fmt.Errorf("user ID is required")
	}

	if message.Type == "" {
		message.Type = MessageTypeText
	}

	return nil
}

// validateNotification validates a notification
func (ce *DefaultCollaborationEngine) validateNotification(notification *Notification) error {
	if notification.UserID == "" {
		return fmt.Errorf("user ID is required")
	}

	if notification.Title == "" {
		return fmt.Errorf("notification title is required")
	}

	if notification.Type == "" {
		notification.Type = NotificationTypeInfo
	}

	if notification.Priority == "" {
		notification.Priority = NotificationPriorityNormal
	}

	return nil
}

// validateWorkflow validates a collaborative workflow
func (ce *DefaultCollaborationEngine) validateWorkflow(workflow *CollaborativeWorkflow) error {
	if workflow.Name == "" {
		return fmt.Errorf("workflow name is required")
	}

	if workflow.TeamID == "" {
		return fmt.Errorf("team ID is required")
	}

	if workflow.Type == "" {
		workflow.Type = WorkflowTypeCustom
	}

	if workflow.Status == "" {
		workflow.Status = WorkflowStatusDraft
	}

	// Validate steps
	for i, step := range workflow.Steps {
		if step.Name == "" {
			return fmt.Errorf("step %d name is required", i)
		}
		if step.Type == "" {
			return fmt.Errorf("step %d type is required", i)
		}
		if len(step.Assignees) == 0 {
			return fmt.Errorf("step %d must have at least one assignee", i)
		}
	}

	return nil
}

// validateDocument validates a document
func (ce *DefaultCollaborationEngine) validateDocument(document *Document) error {
	if document.Title == "" {
		return fmt.Errorf("document title is required")
	}

	if document.Type == "" {
		document.Type = DocumentTypeMarkdown
	}

	if document.Status == "" {
		document.Status = DocumentStatusDraft
	}

	if document.Format == "" {
		document.Format = "markdown"
	}

	return nil
}

// Filter matching methods

// matchesTeamFilter checks if a team matches the given filter
func (ce *DefaultCollaborationEngine) matchesTeamFilter(team *Team, filter *TeamFilter) bool {
	if filter == nil {
		return true
	}

	// Type filter
	if filter.Type != "" && team.Type != filter.Type {
		return false
	}

	// Status filter
	if filter.Status != "" && team.Status != filter.Status {
		return false
	}

	// Created by filter
	if filter.CreatedBy != "" && team.CreatedBy != filter.CreatedBy {
		return false
	}

	// Tags filter
	if len(filter.Tags) > 0 {
		for _, filterTag := range filter.Tags {
			found := false
			for _, teamTag := range team.Tags {
				if teamTag == filterTag {
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
		if !(strings.Contains(strings.ToLower(team.Name), searchLower) ||
			strings.Contains(strings.ToLower(team.Description), searchLower)) {
			return false
		}
	}

	return true
}

// matchesRoleFilter checks if a role matches the given filter
func (ce *DefaultCollaborationEngine) matchesRoleFilter(role *Role, filter *RoleFilter) bool {
	if filter == nil {
		return true
	}

	// Type filter
	if filter.Type != "" && role.Type != filter.Type {
		return false
	}

	// Team ID filter
	if filter.TeamID != "" && role.TeamID != filter.TeamID {
		return false
	}

	// Search filter
	if filter.Search != "" {
		searchLower := strings.ToLower(filter.Search)
		if !(strings.Contains(strings.ToLower(role.Name), searchLower) ||
			strings.Contains(strings.ToLower(role.Description), searchLower)) {
			return false
		}
	}

	return true
}

// matchesChannelFilter checks if a channel matches the given filter
func (ce *DefaultCollaborationEngine) matchesChannelFilter(channel *Channel, filter *ChannelFilter) bool {
	if filter == nil {
		return true
	}

	// Type filter
	if filter.Type != "" && channel.Type != filter.Type {
		return false
	}

	// Member ID filter
	if filter.MemberID != "" {
		found := false
		for _, memberID := range channel.Members {
			if memberID == filter.MemberID {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Search filter
	if filter.Search != "" {
		searchLower := strings.ToLower(filter.Search)
		if !(strings.Contains(strings.ToLower(channel.Name), searchLower) ||
			strings.Contains(strings.ToLower(channel.Description), searchLower)) {
			return false
		}
	}

	return true
}

// matchesMessageFilter checks if a message matches the given filter
func (ce *DefaultCollaborationEngine) matchesMessageFilter(message *Message, filter *MessageFilter) bool {
	if filter == nil {
		return true
	}

	// User ID filter
	if filter.UserID != "" && message.UserID != filter.UserID {
		return false
	}

	// Type filter
	if filter.Type != "" && message.Type != filter.Type {
		return false
	}

	// Thread ID filter
	if filter.ThreadID != "" && message.ThreadID != filter.ThreadID {
		return false
	}

	// Time range filters
	if filter.Since != nil && message.CreatedAt.Before(*filter.Since) {
		return false
	}
	if filter.Until != nil && message.CreatedAt.After(*filter.Until) {
		return false
	}

	// Search filter
	if filter.Search != "" {
		searchLower := strings.ToLower(filter.Search)
		if !strings.Contains(strings.ToLower(message.Content), searchLower) {
			return false
		}
	}

	return true
}

// matchesNotificationFilter checks if a notification matches the given filter
func (ce *DefaultCollaborationEngine) matchesNotificationFilter(notification *Notification, filter *NotificationFilter) bool {
	if filter == nil {
		return true
	}

	// Type filter
	if filter.Type != "" && notification.Type != filter.Type {
		return false
	}

	// Priority filter
	if filter.Priority != "" && notification.Priority != filter.Priority {
		return false
	}

	// Category filter
	if filter.Category != "" && notification.Category != filter.Category {
		return false
	}

	// Read status filter
	if filter.IsRead != nil && notification.IsRead != *filter.IsRead {
		return false
	}

	// Time range filters
	if filter.Since != nil && notification.CreatedAt.Before(*filter.Since) {
		return false
	}
	if filter.Until != nil && notification.CreatedAt.After(*filter.Until) {
		return false
	}

	return true
}

// matchesWorkflowFilter checks if a workflow matches the given filter
func (ce *DefaultCollaborationEngine) matchesWorkflowFilter(workflow *CollaborativeWorkflow, filter *WorkflowFilter) bool {
	if filter == nil {
		return true
	}

	// Type filter
	if filter.Type != "" && workflow.Type != filter.Type {
		return false
	}

	// Status filter
	if filter.Status != "" && workflow.Status != filter.Status {
		return false
	}

	// Created by filter
	if filter.CreatedBy != "" && workflow.CreatedBy != filter.CreatedBy {
		return false
	}

	// Search filter
	if filter.Search != "" {
		searchLower := strings.ToLower(filter.Search)
		if !(strings.Contains(strings.ToLower(workflow.Name), searchLower) ||
			strings.Contains(strings.ToLower(workflow.Description), searchLower)) {
			return false
		}
	}

	return true
}

// matchesActivityFilter checks if an activity matches the given filter
func (ce *DefaultCollaborationEngine) matchesActivityFilter(activity *Activity, filter *ActivityFilter) bool {
	if filter == nil {
		return true
	}

	// Type filter
	if filter.Type != "" && activity.Type != filter.Type {
		return false
	}

	// User ID filter
	if filter.UserID != "" && activity.UserID != filter.UserID {
		return false
	}

	// Team ID filter
	if filter.TeamID != "" && activity.TeamID != filter.TeamID {
		return false
	}

	// Project ID filter
	if filter.ProjectID != "" && activity.ProjectID != filter.ProjectID {
		return false
	}

	// Action filter
	if filter.Action != "" && activity.Action != filter.Action {
		return false
	}

	// Time range filters
	if filter.Since != nil && activity.Timestamp.Before(*filter.Since) {
		return false
	}
	if filter.Until != nil && activity.Timestamp.After(*filter.Until) {
		return false
	}

	return true
}

// Utility methods

// getTeamMembersInternal retrieves team members without locking (internal use)
func (ce *DefaultCollaborationEngine) getTeamMembersInternal(teamID string) []*TeamMember {
	if team, exists := ce.teams[teamID]; exists {
		return team.Members
	}
	return []*TeamMember{}
}

// logActivity logs an activity (internal method)
func (ce *DefaultCollaborationEngine) logActivity(activity *Activity) {
	if activity.ID == "" {
		activity.ID = uuid.New().String()
	}

	ce.mu.Lock()
	ce.activities = append(ce.activities, activity)
	
	// Keep only recent activities (simple cleanup)
	if len(ce.activities) > 10000 {
		ce.activities = ce.activities[1000:]
	}
	ce.mu.Unlock()
}

// createDefaultRoles creates system default roles
func (ce *DefaultCollaborationEngine) createDefaultRoles() {
	now := time.Now()

	// Admin role
	adminRole := &Role{
		ID:          uuid.New().String(),
		Name:        "Admin",
		Description: "Full administrative access",
		Type:        RoleTypeSystem,
		Permissions: []*Permission{
			{
				ID:          uuid.New().String(),
				Name:        "manage_team",
				Description: "Manage team settings and members",
				Resource:    "team",
				Action:      "*",
				Scope:       PermissionScopeTeam,
			},
			{
				ID:          uuid.New().String(),
				Name:        "manage_projects",
				Description: "Manage all projects",
				Resource:    "project",
				Action:      "*",
				Scope:       PermissionScopeTeam,
			},
		},
		IsSystem:  true,
		CreatedBy: "system",
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Member role
	memberRole := &Role{
		ID:          uuid.New().String(),
		Name:        "Member",
		Description: "Standard team member access",
		Type:        RoleTypeSystem,
		Permissions: []*Permission{
			{
				ID:          uuid.New().String(),
				Name:        "view_team",
				Description: "View team information",
				Resource:    "team",
				Action:      "read",
				Scope:       PermissionScopeTeam,
			},
			{
				ID:          uuid.New().String(),
				Name:        "participate_projects",
				Description: "Participate in team projects",
				Resource:    "project",
				Action:      "read,write",
				Scope:       PermissionScopeProject,
			},
		},
		IsSystem:  true,
		CreatedBy: "system",
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Viewer role
	viewerRole := &Role{
		ID:          uuid.New().String(),
		Name:        "Viewer",
		Description: "Read-only access",
		Type:        RoleTypeSystem,
		Permissions: []*Permission{
			{
				ID:          uuid.New().String(),
				Name:        "view_team",
				Description: "View team information",
				Resource:    "team",
				Action:      "read",
				Scope:       PermissionScopeTeam,
			},
		},
		IsSystem:  true,
		CreatedBy: "system",
		CreatedAt: now,
		UpdatedAt: now,
	}

	ce.roles[adminRole.ID] = adminRole
	ce.roles[memberRole.ID] = memberRole
	ce.roles[viewerRole.ID] = viewerRole

	ce.logger.WithField("roles", 3).Info("Default roles created")
}
