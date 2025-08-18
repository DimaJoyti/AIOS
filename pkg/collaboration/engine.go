package collaboration

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// DefaultCollaborationEngine implements the CollaborationEngine interface
type DefaultCollaborationEngine struct {
	logger *logrus.Logger
	tracer trace.Tracer

	// In-memory storage for collaboration data
	teams                map[string]*Team
	roles                map[string]*Role
	channels             map[string]*Channel
	messages             map[string]*Message
	notifications        map[string]*Notification
	notificationSettings map[string]*NotificationSettings
	workflows            map[string]*CollaborativeWorkflow
	documents            map[string]*Document
	documentVersions     map[string][]*DocumentVersion
	activities           []*Activity

	// Indexes for efficient querying
	teamsByUser         map[string][]string
	channelsByTeam      map[string][]string
	messagesByChannel   map[string][]string
	notificationsByUser map[string][]string

	mu sync.RWMutex
}

// CollaborationEngineConfig represents configuration for the collaboration engine
type CollaborationEngineConfig struct {
	MaxTeams            int           `json:"max_teams"`
	MaxChannelsPerTeam  int           `json:"max_channels_per_team"`
	MaxMembersPerTeam   int           `json:"max_members_per_team"`
	MessageRetention    time.Duration `json:"message_retention"`
	ActivityRetention   time.Duration `json:"activity_retention"`
	EnableRealTime      bool          `json:"enable_real_time"`
	EnableNotifications bool          `json:"enable_notifications"`
}

// NewDefaultCollaborationEngine creates a new collaboration engine
func NewDefaultCollaborationEngine(config *CollaborationEngineConfig, logger *logrus.Logger) CollaborationEngine {
	if config == nil {
		config = &CollaborationEngineConfig{
			MaxTeams:            100,
			MaxChannelsPerTeam:  50,
			MaxMembersPerTeam:   100,
			MessageRetention:    90 * 24 * time.Hour, // 90 days
			ActivityRetention:   30 * 24 * time.Hour, // 30 days
			EnableRealTime:      true,
			EnableNotifications: true,
		}
	}

	engine := &DefaultCollaborationEngine{
		logger:               logger,
		tracer:               otel.Tracer("collaboration.engine"),
		teams:                make(map[string]*Team),
		roles:                make(map[string]*Role),
		channels:             make(map[string]*Channel),
		messages:             make(map[string]*Message),
		notifications:        make(map[string]*Notification),
		notificationSettings: make(map[string]*NotificationSettings),
		workflows:            make(map[string]*CollaborativeWorkflow),
		documents:            make(map[string]*Document),
		documentVersions:     make(map[string][]*DocumentVersion),
		activities:           make([]*Activity, 0),
		teamsByUser:          make(map[string][]string),
		channelsByTeam:       make(map[string][]string),
		messagesByChannel:    make(map[string][]string),
		notificationsByUser:  make(map[string][]string),
	}

	// Create default roles
	engine.createDefaultRoles()

	return engine
}

// Team Management

// CreateTeam creates a new team
func (ce *DefaultCollaborationEngine) CreateTeam(team *Team) (*Team, error) {
	_, span := ce.tracer.Start(context.Background(), "collaboration.create_team")
	defer span.End()

	if team.ID == "" {
		team.ID = uuid.New().String()
	}

	// Set timestamps
	now := time.Now()
	team.CreatedAt = now
	team.UpdatedAt = now

	// Set default settings if not provided
	if team.Settings == nil {
		team.Settings = &TeamSettings{
			IsPublic:             false,
			AllowExternalMembers: false,
			RequireApproval:      true,
			DefaultRole:          "member",
			NotificationSettings: &TeamNotificationSettings{
				EnableEmailNotifications: true,
				EnableSlackIntegration:   false,
				EnableWebhooks:           false,
				NotificationChannels:     []string{"general"},
			},
			WorkflowSettings: &TeamWorkflowSettings{
				RequireCodeReview:   true,
				MinimumReviewers:    2,
				AutoAssignReviewers: true,
				EnableAutomation:    true,
			},
		}
	}

	// Validate team
	if err := ce.validateTeam(team); err != nil {
		return nil, fmt.Errorf("team validation failed: %w", err)
	}

	ce.mu.Lock()
	ce.teams[team.ID] = team

	// Create default general channel
	generalChannel := &Channel{
		ID:          uuid.New().String(),
		Name:        "general",
		Description: "General team discussion",
		Type:        ChannelTypePublic,
		TeamID:      team.ID,
		Members:     []string{},
		Settings: &ChannelSettings{
			IsArchived:       false,
			AllowThreads:     true,
			AllowFileUploads: true,
			RetentionDays:    90,
		},
		CreatedBy: team.CreatedBy,
		CreatedAt: now,
		UpdatedAt: now,
	}

	ce.channels[generalChannel.ID] = generalChannel
	ce.channelsByTeam[team.ID] = []string{generalChannel.ID}
	ce.mu.Unlock()

	// Log activity
	ce.logActivity(&Activity{
		Type:        ActivityTypeTeam,
		UserID:      team.CreatedBy,
		TeamID:      team.ID,
		Action:      "created",
		Resource:    &ActivityResource{Type: "team", ID: team.ID, Name: team.Name},
		Description: fmt.Sprintf("Created team '%s'", team.Name),
		Timestamp:   now,
	})

	span.SetAttributes(
		attribute.String("team.id", team.ID),
		attribute.String("team.name", team.Name),
		attribute.String("team.type", string(team.Type)),
	)

	ce.logger.WithFields(logrus.Fields{
		"team_id":   team.ID,
		"team_name": team.Name,
		"team_type": team.Type,
	}).Info("Team created successfully")

	return team, nil
}

// GetTeam retrieves a team by ID
func (ce *DefaultCollaborationEngine) GetTeam(teamID string) (*Team, error) {
	_, span := ce.tracer.Start(context.Background(), "collaboration.get_team")
	defer span.End()

	span.SetAttributes(attribute.String("team.id", teamID))

	ce.mu.RLock()
	team, exists := ce.teams[teamID]
	ce.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("team not found: %s", teamID)
	}

	// Load team members
	team.Members = ce.getTeamMembersInternal(teamID)

	// Load team channels
	if channelIDs, exists := ce.channelsByTeam[teamID]; exists {
		team.Channels = make([]*Channel, 0, len(channelIDs))
		for _, channelID := range channelIDs {
			if channel, exists := ce.channels[channelID]; exists {
				team.Channels = append(team.Channels, channel)
			}
		}
	}

	return team, nil
}

// UpdateTeam updates an existing team
func (ce *DefaultCollaborationEngine) UpdateTeam(team *Team) (*Team, error) {
	_, span := ce.tracer.Start(context.Background(), "collaboration.update_team")
	defer span.End()

	ce.mu.Lock()
	existing, exists := ce.teams[team.ID]
	if !exists {
		ce.mu.Unlock()
		return nil, fmt.Errorf("team not found: %s", team.ID)
	}

	// Preserve creation info
	team.CreatedBy = existing.CreatedBy
	team.CreatedAt = existing.CreatedAt
	team.UpdatedAt = time.Now()

	// Validate team
	if err := ce.validateTeam(team); err != nil {
		ce.mu.Unlock()
		return nil, fmt.Errorf("team validation failed: %w", err)
	}

	ce.teams[team.ID] = team
	ce.mu.Unlock()

	// Log activity
	ce.logActivity(&Activity{
		Type:        ActivityTypeTeam,
		UserID:      "system", // In production, this would be the actual user
		TeamID:      team.ID,
		Action:      "updated",
		Resource:    &ActivityResource{Type: "team", ID: team.ID, Name: team.Name},
		Description: fmt.Sprintf("Updated team '%s'", team.Name),
		Timestamp:   time.Now(),
	})

	span.SetAttributes(attribute.String("team.id", team.ID))

	ce.logger.WithFields(logrus.Fields{
		"team_id":   team.ID,
		"team_name": team.Name,
	}).Info("Team updated successfully")

	return team, nil
}

// DeleteTeam deletes a team
func (ce *DefaultCollaborationEngine) DeleteTeam(teamID string) error {
	_, span := ce.tracer.Start(context.Background(), "collaboration.delete_team")
	defer span.End()

	ce.mu.Lock()
	team, exists := ce.teams[teamID]
	if !exists {
		ce.mu.Unlock()
		return fmt.Errorf("team not found: %s", teamID)
	}

	// Delete team channels
	if channelIDs, exists := ce.channelsByTeam[teamID]; exists {
		for _, channelID := range channelIDs {
			delete(ce.channels, channelID)

			// Delete channel messages
			if messageIDs, exists := ce.messagesByChannel[channelID]; exists {
				for _, messageID := range messageIDs {
					delete(ce.messages, messageID)
				}
				delete(ce.messagesByChannel, channelID)
			}
		}
		delete(ce.channelsByTeam, teamID)
	}

	// Remove team from user indexes
	for userID, teamIDs := range ce.teamsByUser {
		for i, tid := range teamIDs {
			if tid == teamID {
				ce.teamsByUser[userID] = append(teamIDs[:i], teamIDs[i+1:]...)
				break
			}
		}
	}

	delete(ce.teams, teamID)
	ce.mu.Unlock()

	// Log activity
	ce.logActivity(&Activity{
		Type:        ActivityTypeTeam,
		UserID:      "system",
		TeamID:      teamID,
		Action:      "deleted",
		Resource:    &ActivityResource{Type: "team", ID: teamID, Name: team.Name},
		Description: fmt.Sprintf("Deleted team '%s'", team.Name),
		Timestamp:   time.Now(),
	})

	span.SetAttributes(attribute.String("team.id", teamID))

	ce.logger.WithField("team_id", teamID).Info("Team deleted successfully")

	return nil
}

// ListTeams lists teams with filtering
func (ce *DefaultCollaborationEngine) ListTeams(filter *TeamFilter) ([]*Team, error) {
	_, span := ce.tracer.Start(context.Background(), "collaboration.list_teams")
	defer span.End()

	ce.mu.RLock()
	var teams []*Team
	for _, team := range ce.teams {
		if ce.matchesTeamFilter(team, filter) {
			// Create a copy to avoid modifying the original
			teamCopy := *team
			teams = append(teams, &teamCopy)
		}
	}
	ce.mu.RUnlock()

	// Sort teams by creation date (newest first)
	sort.Slice(teams, func(i, j int) bool {
		return teams[i].CreatedAt.After(teams[j].CreatedAt)
	})

	// Apply pagination
	if filter != nil {
		if filter.Offset > 0 && filter.Offset < len(teams) {
			teams = teams[filter.Offset:]
		}
		if filter.Limit > 0 && filter.Limit < len(teams) {
			teams = teams[:filter.Limit]
		}
	}

	span.SetAttributes(attribute.Int("teams.count", len(teams)))

	return teams, nil
}

// Member Management

// AddTeamMember adds a member to a team
func (ce *DefaultCollaborationEngine) AddTeamMember(teamID string, member *TeamMember) error {
	_, span := ce.tracer.Start(context.Background(), "collaboration.add_team_member")
	defer span.End()

	ce.mu.Lock()
	team, exists := ce.teams[teamID]
	if !exists {
		ce.mu.Unlock()
		return fmt.Errorf("team not found: %s", teamID)
	}

	// Check if member already exists
	for _, existingMember := range team.Members {
		if existingMember.UserID == member.UserID {
			ce.mu.Unlock()
			return fmt.Errorf("user %s is already a member of team %s", member.UserID, teamID)
		}
	}

	// Set default values
	if member.Role == "" {
		member.Role = team.Settings.DefaultRole
	}
	if member.Status == "" {
		member.Status = MemberStatusActive
	}
	member.JoinedAt = time.Now()

	// Add member to team
	team.Members = append(team.Members, member)
	team.UpdatedAt = time.Now()

	// Update user index
	ce.teamsByUser[member.UserID] = append(ce.teamsByUser[member.UserID], teamID)

	// Add member to all public channels
	for _, channelID := range ce.channelsByTeam[teamID] {
		if channel, exists := ce.channels[channelID]; exists && channel.Type == ChannelTypePublic {
			channel.Members = append(channel.Members, member.UserID)
		}
	}

	ce.mu.Unlock()

	// Log activity
	ce.logActivity(&Activity{
		Type:        ActivityTypeTeam,
		UserID:      member.UserID,
		TeamID:      teamID,
		Action:      "joined",
		Resource:    &ActivityResource{Type: "team", ID: teamID, Name: team.Name},
		Description: fmt.Sprintf("%s joined team '%s'", member.DisplayName, team.Name),
		Timestamp:   time.Now(),
	})

	span.SetAttributes(
		attribute.String("team.id", teamID),
		attribute.String("user.id", member.UserID),
		attribute.String("member.role", member.Role),
	)

	ce.logger.WithFields(logrus.Fields{
		"team_id": teamID,
		"user_id": member.UserID,
		"role":    member.Role,
	}).Info("Team member added successfully")

	return nil
}

// RemoveTeamMember removes a member from a team
func (ce *DefaultCollaborationEngine) RemoveTeamMember(teamID, userID string) error {
	_, span := ce.tracer.Start(context.Background(), "collaboration.remove_team_member")
	defer span.End()

	ce.mu.Lock()
	team, exists := ce.teams[teamID]
	if !exists {
		ce.mu.Unlock()
		return fmt.Errorf("team not found: %s", teamID)
	}

	// Find and remove member
	memberIndex := -1
	var removedMember *TeamMember
	for i, member := range team.Members {
		if member.UserID == userID {
			memberIndex = i
			removedMember = member
			break
		}
	}

	if memberIndex == -1 {
		ce.mu.Unlock()
		return fmt.Errorf("user %s is not a member of team %s", userID, teamID)
	}

	// Remove member from team
	team.Members = append(team.Members[:memberIndex], team.Members[memberIndex+1:]...)
	team.UpdatedAt = time.Now()

	// Update user index
	if teamIDs, exists := ce.teamsByUser[userID]; exists {
		for i, tid := range teamIDs {
			if tid == teamID {
				ce.teamsByUser[userID] = append(teamIDs[:i], teamIDs[i+1:]...)
				break
			}
		}
	}

	// Remove member from all team channels
	for _, channelID := range ce.channelsByTeam[teamID] {
		if channel, exists := ce.channels[channelID]; exists {
			for i, memberID := range channel.Members {
				if memberID == userID {
					channel.Members = append(channel.Members[:i], channel.Members[i+1:]...)
					break
				}
			}
		}
	}

	ce.mu.Unlock()

	// Log activity
	ce.logActivity(&Activity{
		Type:        ActivityTypeTeam,
		UserID:      userID,
		TeamID:      teamID,
		Action:      "left",
		Resource:    &ActivityResource{Type: "team", ID: teamID, Name: team.Name},
		Description: fmt.Sprintf("%s left team '%s'", removedMember.DisplayName, team.Name),
		Timestamp:   time.Now(),
	})

	span.SetAttributes(
		attribute.String("team.id", teamID),
		attribute.String("user.id", userID),
	)

	ce.logger.WithFields(logrus.Fields{
		"team_id": teamID,
		"user_id": userID,
	}).Info("Team member removed successfully")

	return nil
}

// UpdateTeamMember updates a team member
func (ce *DefaultCollaborationEngine) UpdateTeamMember(teamID string, member *TeamMember) error {
	_, span := ce.tracer.Start(context.Background(), "collaboration.update_team_member")
	defer span.End()

	ce.mu.Lock()
	team, exists := ce.teams[teamID]
	if !exists {
		ce.mu.Unlock()
		return fmt.Errorf("team not found: %s", teamID)
	}

	// Find and update member
	memberIndex := -1
	for i, existingMember := range team.Members {
		if existingMember.UserID == member.UserID {
			memberIndex = i
			break
		}
	}

	if memberIndex == -1 {
		ce.mu.Unlock()
		return fmt.Errorf("user %s is not a member of team %s", member.UserID, teamID)
	}

	// Preserve join date
	member.JoinedAt = team.Members[memberIndex].JoinedAt

	// Update member
	team.Members[memberIndex] = member
	team.UpdatedAt = time.Now()

	ce.mu.Unlock()

	span.SetAttributes(
		attribute.String("team.id", teamID),
		attribute.String("user.id", member.UserID),
		attribute.String("member.role", member.Role),
	)

	ce.logger.WithFields(logrus.Fields{
		"team_id": teamID,
		"user_id": member.UserID,
		"role":    member.Role,
	}).Info("Team member updated successfully")

	return nil
}

// GetTeamMembers retrieves all members of a team
func (ce *DefaultCollaborationEngine) GetTeamMembers(teamID string) ([]*TeamMember, error) {
	_, span := ce.tracer.Start(context.Background(), "collaboration.get_team_members")
	defer span.End()

	ce.mu.RLock()
	team, exists := ce.teams[teamID]
	ce.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("team not found: %s", teamID)
	}

	span.SetAttributes(
		attribute.String("team.id", teamID),
		attribute.Int("members.count", len(team.Members)),
	)

	return team.Members, nil
}

// Role Management

// CreateRole creates a new role
func (ce *DefaultCollaborationEngine) CreateRole(role *Role) (*Role, error) {
	_, span := ce.tracer.Start(context.Background(), "collaboration.create_role")
	defer span.End()

	if role.ID == "" {
		role.ID = uuid.New().String()
	}

	// Set timestamps
	now := time.Now()
	role.CreatedAt = now
	role.UpdatedAt = now

	// Validate role
	if err := ce.validateRole(role); err != nil {
		return nil, fmt.Errorf("role validation failed: %w", err)
	}

	ce.mu.Lock()
	ce.roles[role.ID] = role
	ce.mu.Unlock()

	span.SetAttributes(
		attribute.String("role.id", role.ID),
		attribute.String("role.name", role.Name),
		attribute.String("role.type", string(role.Type)),
	)

	ce.logger.WithFields(logrus.Fields{
		"role_id":   role.ID,
		"role_name": role.Name,
		"role_type": role.Type,
	}).Info("Role created successfully")

	return role, nil
}

// GetRole retrieves a role by ID
func (ce *DefaultCollaborationEngine) GetRole(roleID string) (*Role, error) {
	_, span := ce.tracer.Start(context.Background(), "collaboration.get_role")
	defer span.End()

	ce.mu.RLock()
	role, exists := ce.roles[roleID]
	ce.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("role not found: %s", roleID)
	}

	span.SetAttributes(attribute.String("role.id", roleID))

	return role, nil
}

// UpdateRole updates an existing role
func (ce *DefaultCollaborationEngine) UpdateRole(role *Role) (*Role, error) {
	_, span := ce.tracer.Start(context.Background(), "collaboration.update_role")
	defer span.End()

	ce.mu.Lock()
	existing, exists := ce.roles[role.ID]
	if !exists {
		ce.mu.Unlock()
		return nil, fmt.Errorf("role not found: %s", role.ID)
	}

	// Preserve creation info
	role.CreatedBy = existing.CreatedBy
	role.CreatedAt = existing.CreatedAt
	role.UpdatedAt = time.Now()

	// Validate role
	if err := ce.validateRole(role); err != nil {
		ce.mu.Unlock()
		return nil, fmt.Errorf("role validation failed: %w", err)
	}

	ce.roles[role.ID] = role
	ce.mu.Unlock()

	span.SetAttributes(attribute.String("role.id", role.ID))

	ce.logger.WithField("role_id", role.ID).Info("Role updated successfully")

	return role, nil
}

// DeleteRole deletes a role
func (ce *DefaultCollaborationEngine) DeleteRole(roleID string) error {
	_, span := ce.tracer.Start(context.Background(), "collaboration.delete_role")
	defer span.End()

	ce.mu.Lock()
	role, exists := ce.roles[roleID]
	if !exists {
		ce.mu.Unlock()
		return fmt.Errorf("role not found: %s", roleID)
	}

	// Check if role is system role
	if role.IsSystem {
		ce.mu.Unlock()
		return fmt.Errorf("cannot delete system role: %s", roleID)
	}

	delete(ce.roles, roleID)
	ce.mu.Unlock()

	span.SetAttributes(attribute.String("role.id", roleID))

	ce.logger.WithField("role_id", roleID).Info("Role deleted successfully")

	return nil
}

// ListRoles lists roles with filtering
func (ce *DefaultCollaborationEngine) ListRoles(filter *RoleFilter) ([]*Role, error) {
	_, span := ce.tracer.Start(context.Background(), "collaboration.list_roles")
	defer span.End()

	ce.mu.RLock()
	var roles []*Role
	for _, role := range ce.roles {
		if ce.matchesRoleFilter(role, filter) {
			roles = append(roles, role)
		}
	}
	ce.mu.RUnlock()

	// Sort roles by name
	sort.Slice(roles, func(i, j int) bool {
		return roles[i].Name < roles[j].Name
	})

	// Apply pagination
	if filter != nil {
		if filter.Offset > 0 && filter.Offset < len(roles) {
			roles = roles[filter.Offset:]
		}
		if filter.Limit > 0 && filter.Limit < len(roles) {
			roles = roles[:filter.Limit]
		}
	}

	span.SetAttributes(attribute.Int("roles.count", len(roles)))

	return roles, nil
}

// AssignRole assigns a role to a user
func (ce *DefaultCollaborationEngine) AssignRole(userID, roleID string) error {
	_, span := ce.tracer.Start(context.Background(), "collaboration.assign_role")
	defer span.End()

	// Verify role exists
	ce.mu.RLock()
	role, exists := ce.roles[roleID]
	ce.mu.RUnlock()

	if !exists {
		return fmt.Errorf("role not found: %s", roleID)
	}

	// Log activity
	ce.logActivity(&Activity{
		Type:        ActivityTypeTeam,
		UserID:      userID,
		Action:      "role_assigned",
		Resource:    &ActivityResource{Type: "role", ID: roleID, Name: role.Name},
		Description: fmt.Sprintf("Role '%s' assigned to user", role.Name),
		Timestamp:   time.Now(),
	})

	span.SetAttributes(
		attribute.String("user.id", userID),
		attribute.String("role.id", roleID),
	)

	ce.logger.WithFields(logrus.Fields{
		"user_id": userID,
		"role_id": roleID,
	}).Info("Role assigned successfully")

	return nil
}

// RevokeRole revokes a role from a user
func (ce *DefaultCollaborationEngine) RevokeRole(userID, roleID string) error {
	_, span := ce.tracer.Start(context.Background(), "collaboration.revoke_role")
	defer span.End()

	// Verify role exists
	ce.mu.RLock()
	role, exists := ce.roles[roleID]
	ce.mu.RUnlock()

	if !exists {
		return fmt.Errorf("role not found: %s", roleID)
	}

	// Log activity
	ce.logActivity(&Activity{
		Type:        ActivityTypeTeam,
		UserID:      userID,
		Action:      "role_revoked",
		Resource:    &ActivityResource{Type: "role", ID: roleID, Name: role.Name},
		Description: fmt.Sprintf("Role '%s' revoked from user", role.Name),
		Timestamp:   time.Now(),
	})

	span.SetAttributes(
		attribute.String("user.id", userID),
		attribute.String("role.id", roleID),
	)

	ce.logger.WithFields(logrus.Fields{
		"user_id": userID,
		"role_id": roleID,
	}).Info("Role revoked successfully")

	return nil
}
