package collaboration

import (
	"time"
)

// Team collaboration types and interfaces for AIOS

// CollaborationEngine defines the interface for team collaboration operations
type CollaborationEngine interface {
	// Team Management
	CreateTeam(team *Team) (*Team, error)
	GetTeam(teamID string) (*Team, error)
	UpdateTeam(team *Team) (*Team, error)
	DeleteTeam(teamID string) error
	ListTeams(filter *TeamFilter) ([]*Team, error)
	
	// Member Management
	AddTeamMember(teamID string, member *TeamMember) error
	RemoveTeamMember(teamID, userID string) error
	UpdateTeamMember(teamID string, member *TeamMember) error
	GetTeamMembers(teamID string) ([]*TeamMember, error)
	
	// Role Management
	CreateRole(role *Role) (*Role, error)
	GetRole(roleID string) (*Role, error)
	UpdateRole(role *Role) (*Role, error)
	DeleteRole(roleID string) error
	ListRoles(filter *RoleFilter) ([]*Role, error)
	AssignRole(userID, roleID string) error
	RevokeRole(userID, roleID string) error
	
	// Communication
	SendMessage(message *Message) (*Message, error)
	GetMessages(channelID string, filter *MessageFilter) ([]*Message, error)
	CreateChannel(channel *Channel) (*Channel, error)
	GetChannel(channelID string) (*Channel, error)
	ListChannels(teamID string, filter *ChannelFilter) ([]*Channel, error)
	
	// Notifications
	SendNotification(notification *Notification) error
	GetNotifications(userID string, filter *NotificationFilter) ([]*Notification, error)
	MarkNotificationRead(notificationID string) error
	GetNotificationSettings(userID string) (*NotificationSettings, error)
	UpdateNotificationSettings(userID string, settings *NotificationSettings) error
	
	// Collaborative Workflows
	CreateCollaborativeWorkflow(workflow *CollaborativeWorkflow) (*CollaborativeWorkflow, error)
	GetCollaborativeWorkflow(workflowID string) (*CollaborativeWorkflow, error)
	UpdateCollaborativeWorkflow(workflow *CollaborativeWorkflow) (*CollaborativeWorkflow, error)
	ListCollaborativeWorkflows(teamID string, filter *WorkflowFilter) ([]*CollaborativeWorkflow, error)
	
	// Document Collaboration
	CreateDocument(document *Document) (*Document, error)
	GetDocument(documentID string) (*Document, error)
	UpdateDocument(document *Document) (*Document, error)
	ShareDocument(documentID string, sharing *DocumentSharing) error
	GetDocumentVersions(documentID string) ([]*DocumentVersion, error)
	
	// Activity Tracking
	LogActivity(activity *Activity) error
	GetActivities(filter *ActivityFilter) ([]*Activity, error)
	GetTeamActivity(teamID string, filter *ActivityFilter) ([]*Activity, error)
	GetUserActivity(userID string, filter *ActivityFilter) ([]*Activity, error)
}

// Core Types

// Team represents a collaborative team
type Team struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Type        TeamType               `json:"type"`
	Status      TeamStatus             `json:"status"`
	Settings    *TeamSettings          `json:"settings"`
	Members     []*TeamMember          `json:"members,omitempty"`
	Channels    []*Channel             `json:"channels,omitempty"`
	Projects    []string               `json:"projects,omitempty"`
	Tags        []string               `json:"tags,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	CreatedBy   string                 `json:"created_by"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// TeamType defines the type of team
type TeamType string

const (
	TeamTypeDevelopment TeamType = "development"
	TeamTypeDesign      TeamType = "design"
	TeamTypeQA          TeamType = "qa"
	TeamTypeDevOps      TeamType = "devops"
	TeamTypeProduct     TeamType = "product"
	TeamTypeMarketing   TeamType = "marketing"
	TeamTypeSales       TeamType = "sales"
	TeamTypeSupport     TeamType = "support"
	TeamTypeCrossFunctional TeamType = "cross_functional"
)

// TeamStatus defines the status of a team
type TeamStatus string

const (
	TeamStatusActive   TeamStatus = "active"
	TeamStatusInactive TeamStatus = "inactive"
	TeamStatusArchived TeamStatus = "archived"
)

// TeamSettings contains team configuration
type TeamSettings struct {
	IsPublic              bool                   `json:"is_public"`
	AllowExternalMembers  bool                   `json:"allow_external_members"`
	RequireApproval       bool                   `json:"require_approval"`
	DefaultRole           string                 `json:"default_role"`
	NotificationSettings  *TeamNotificationSettings `json:"notification_settings"`
	WorkflowSettings      *TeamWorkflowSettings  `json:"workflow_settings"`
	IntegrationSettings   map[string]interface{} `json:"integration_settings,omitempty"`
}

// TeamNotificationSettings contains team notification preferences
type TeamNotificationSettings struct {
	EnableEmailNotifications bool     `json:"enable_email_notifications"`
	EnableSlackIntegration    bool     `json:"enable_slack_integration"`
	EnableWebhooks           bool     `json:"enable_webhooks"`
	NotificationChannels     []string `json:"notification_channels"`
	QuietHours              *QuietHours `json:"quiet_hours,omitempty"`
}

// TeamWorkflowSettings contains team workflow configuration
type TeamWorkflowSettings struct {
	DefaultWorkflow       string                 `json:"default_workflow"`
	RequireCodeReview     bool                   `json:"require_code_review"`
	MinimumReviewers      int                    `json:"minimum_reviewers"`
	AutoAssignReviewers   bool                   `json:"auto_assign_reviewers"`
	EnableAutomation      bool                   `json:"enable_automation"`
	WorkflowTemplates     []string               `json:"workflow_templates"`
	CustomFields          map[string]interface{} `json:"custom_fields,omitempty"`
}

// QuietHours defines when notifications should be suppressed
type QuietHours struct {
	Enabled   bool   `json:"enabled"`
	StartTime string `json:"start_time"` // HH:MM format
	EndTime   string `json:"end_time"`   // HH:MM format
	Timezone  string `json:"timezone"`
	Weekdays  []int  `json:"weekdays"` // 0=Sunday, 1=Monday, etc.
}

// TeamMember represents a team member
type TeamMember struct {
	UserID      string                 `json:"user_id"`
	Username    string                 `json:"username"`
	Email       string                 `json:"email"`
	DisplayName string                 `json:"display_name"`
	Role        string                 `json:"role"`
	Status      MemberStatus           `json:"status"`
	Permissions []string               `json:"permissions"`
	JoinedAt    time.Time              `json:"joined_at"`
	LastActive  *time.Time             `json:"last_active,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// MemberStatus defines the status of a team member
type MemberStatus string

const (
	MemberStatusActive    MemberStatus = "active"
	MemberStatusInactive  MemberStatus = "inactive"
	MemberStatusPending   MemberStatus = "pending"
	MemberStatusSuspended MemberStatus = "suspended"
)

// Role represents a role with permissions
type Role struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Type        RoleType               `json:"type"`
	Permissions []*Permission          `json:"permissions"`
	IsSystem    bool                   `json:"is_system"`
	TeamID      string                 `json:"team_id,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	CreatedBy   string                 `json:"created_by"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// RoleType defines the type of role
type RoleType string

const (
	RoleTypeSystem RoleType = "system"
	RoleTypeTeam   RoleType = "team"
	RoleTypeCustom RoleType = "custom"
)

// Permission represents a specific permission
type Permission struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Resource    string                 `json:"resource"`
	Action      string                 `json:"action"`
	Scope       PermissionScope        `json:"scope"`
	Conditions  []*PermissionCondition `json:"conditions,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// PermissionScope defines the scope of a permission
type PermissionScope string

const (
	PermissionScopeGlobal  PermissionScope = "global"
	PermissionScopeTeam    PermissionScope = "team"
	PermissionScopeProject PermissionScope = "project"
	PermissionScopeUser    PermissionScope = "user"
)

// PermissionCondition defines conditions for permission application
type PermissionCondition struct {
	Field    string      `json:"field"`
	Operator string      `json:"operator"`
	Value    interface{} `json:"value"`
}

// Communication Types

// Channel represents a communication channel
type Channel struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Type        ChannelType            `json:"type"`
	TeamID      string                 `json:"team_id"`
	Members     []string               `json:"members"`
	Settings    *ChannelSettings       `json:"settings"`
	LastMessage *Message               `json:"last_message,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	CreatedBy   string                 `json:"created_by"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// ChannelType defines the type of channel
type ChannelType string

const (
	ChannelTypePublic    ChannelType = "public"
	ChannelTypePrivate   ChannelType = "private"
	ChannelTypeDirect    ChannelType = "direct"
	ChannelTypeAnnouncement ChannelType = "announcement"
)

// ChannelSettings contains channel configuration
type ChannelSettings struct {
	IsArchived        bool     `json:"is_archived"`
	AllowThreads      bool     `json:"allow_threads"`
	AllowFileUploads  bool     `json:"allow_file_uploads"`
	RetentionDays     int      `json:"retention_days"`
	MutedMembers      []string `json:"muted_members,omitempty"`
	PinnedMessages    []string `json:"pinned_messages,omitempty"`
}

// Message represents a chat message
type Message struct {
	ID          string                 `json:"id"`
	ChannelID   string                 `json:"channel_id"`
	UserID      string                 `json:"user_id"`
	Username    string                 `json:"username"`
	Content     string                 `json:"content"`
	Type        MessageType            `json:"type"`
	ThreadID    string                 `json:"thread_id,omitempty"`
	ReplyToID   string                 `json:"reply_to_id,omitempty"`
	Attachments []*MessageAttachment   `json:"attachments,omitempty"`
	Reactions   []*MessageReaction     `json:"reactions,omitempty"`
	Mentions    []string               `json:"mentions,omitempty"`
	IsEdited    bool                   `json:"is_edited"`
	EditedAt    *time.Time             `json:"edited_at,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
}

// MessageType defines the type of message
type MessageType string

const (
	MessageTypeText        MessageType = "text"
	MessageTypeFile        MessageType = "file"
	MessageTypeImage       MessageType = "image"
	MessageTypeCode        MessageType = "code"
	MessageTypeSystem      MessageType = "system"
	MessageTypeNotification MessageType = "notification"
)

// MessageAttachment represents a file attachment
type MessageAttachment struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Type     string `json:"type"`
	Size     int64  `json:"size"`
	URL      string `json:"url"`
	MimeType string `json:"mime_type"`
}

// MessageReaction represents a reaction to a message
type MessageReaction struct {
	Emoji   string   `json:"emoji"`
	UserIDs []string `json:"user_ids"`
	Count   int      `json:"count"`
}

// Notification Types

// Notification represents a user notification
type Notification struct {
	ID          string                 `json:"id"`
	UserID      string                 `json:"user_id"`
	Type        NotificationType       `json:"type"`
	Title       string                 `json:"title"`
	Message     string                 `json:"message"`
	Priority    NotificationPriority   `json:"priority"`
	Category    string                 `json:"category"`
	Source      *NotificationSource    `json:"source"`
	Actions     []*NotificationAction  `json:"actions,omitempty"`
	IsRead      bool                   `json:"is_read"`
	ReadAt      *time.Time             `json:"read_at,omitempty"`
	ExpiresAt   *time.Time             `json:"expires_at,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
}

// NotificationType defines the type of notification
type NotificationType string

const (
	NotificationTypeInfo     NotificationType = "info"
	NotificationTypeWarning  NotificationType = "warning"
	NotificationTypeError    NotificationType = "error"
	NotificationTypeSuccess  NotificationType = "success"
	NotificationTypeMention  NotificationType = "mention"
	NotificationTypeAssignment NotificationType = "assignment"
	NotificationTypeDeadline NotificationType = "deadline"
	NotificationTypeApproval NotificationType = "approval"
)

// NotificationPriority defines the priority of a notification
type NotificationPriority string

const (
	NotificationPriorityLow    NotificationPriority = "low"
	NotificationPriorityNormal NotificationPriority = "normal"
	NotificationPriorityHigh   NotificationPriority = "high"
	NotificationPriorityUrgent NotificationPriority = "urgent"
)

// NotificationSource contains information about the notification source
type NotificationSource struct {
	Type     string `json:"type"`
	ID       string `json:"id"`
	Name     string `json:"name"`
	URL      string `json:"url,omitempty"`
	IconURL  string `json:"icon_url,omitempty"`
}

// NotificationAction represents an action that can be taken on a notification
type NotificationAction struct {
	ID    string `json:"id"`
	Label string `json:"label"`
	URL   string `json:"url,omitempty"`
	Type  string `json:"type"` // button, link, api_call
}

// NotificationSettings contains user notification preferences
type NotificationSettings struct {
	UserID              string                           `json:"user_id"`
	EmailNotifications  *EmailNotificationSettings      `json:"email_notifications"`
	PushNotifications   *PushNotificationSettings       `json:"push_notifications"`
	InAppNotifications  *InAppNotificationSettings      `json:"in_app_notifications"`
	CategorySettings    map[string]*CategorySettings    `json:"category_settings"`
	QuietHours          *QuietHours                     `json:"quiet_hours,omitempty"`
	UpdatedAt           time.Time                       `json:"updated_at"`
}

// EmailNotificationSettings contains email notification preferences
type EmailNotificationSettings struct {
	Enabled     bool     `json:"enabled"`
	Frequency   string   `json:"frequency"` // immediate, hourly, daily, weekly
	Categories  []string `json:"categories"`
	DigestTime  string   `json:"digest_time,omitempty"` // HH:MM format
}

// PushNotificationSettings contains push notification preferences
type PushNotificationSettings struct {
	Enabled    bool     `json:"enabled"`
	Categories []string `json:"categories"`
	Sound      bool     `json:"sound"`
	Vibration  bool     `json:"vibration"`
}

// InAppNotificationSettings contains in-app notification preferences
type InAppNotificationSettings struct {
	Enabled    bool     `json:"enabled"`
	Categories []string `json:"categories"`
	ShowBadge  bool     `json:"show_badge"`
	AutoRead   bool     `json:"auto_read"`
}

// CategorySettings contains settings for a specific notification category
type CategorySettings struct {
	Enabled   bool                 `json:"enabled"`
	Priority  NotificationPriority `json:"priority"`
	Channels  []string             `json:"channels"` // email, push, in_app
}

// Collaborative Workflow Types

// CollaborativeWorkflow represents a workflow that involves multiple team members
type CollaborativeWorkflow struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	TeamID      string                 `json:"team_id"`
	Type        WorkflowType           `json:"type"`
	Status      WorkflowStatus         `json:"status"`
	Steps       []*WorkflowStep        `json:"steps"`
	Participants []*WorkflowParticipant `json:"participants"`
	Settings    *WorkflowSettings      `json:"settings"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	CreatedBy   string                 `json:"created_by"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// WorkflowType defines the type of collaborative workflow
type WorkflowType string

const (
	WorkflowTypeCodeReview   WorkflowType = "code_review"
	WorkflowTypeApproval     WorkflowType = "approval"
	WorkflowTypeDocumentReview WorkflowType = "document_review"
	WorkflowTypeDecisionMaking WorkflowType = "decision_making"
	WorkflowTypeOnboarding   WorkflowType = "onboarding"
	WorkflowTypeCustom       WorkflowType = "custom"
)

// WorkflowStatus defines the status of a workflow
type WorkflowStatus string

const (
	WorkflowStatusDraft      WorkflowStatus = "draft"
	WorkflowStatusActive     WorkflowStatus = "active"
	WorkflowStatusInProgress WorkflowStatus = "in_progress"
	WorkflowStatusCompleted  WorkflowStatus = "completed"
	WorkflowStatusCancelled  WorkflowStatus = "cancelled"
	WorkflowStatusArchived   WorkflowStatus = "archived"
)

// WorkflowStep represents a step in a collaborative workflow
type WorkflowStep struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Type        StepType               `json:"type"`
	Order       int                    `json:"order"`
	Assignees   []string               `json:"assignees"`
	Dependencies []string              `json:"dependencies,omitempty"`
	DueDate     *time.Time             `json:"due_date,omitempty"`
	Status      StepStatus             `json:"status"`
	Result      *StepResult            `json:"result,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// StepType defines the type of workflow step
type StepType string

const (
	StepTypeReview    StepType = "review"
	StepTypeApproval  StepType = "approval"
	StepTypeTask      StepType = "task"
	StepTypeDecision  StepType = "decision"
	StepTypeNotification StepType = "notification"
	StepTypeAutomation StepType = "automation"
)

// StepStatus defines the status of a workflow step
type StepStatus string

const (
	StepStatusPending    StepStatus = "pending"
	StepStatusInProgress StepStatus = "in_progress"
	StepStatusCompleted  StepStatus = "completed"
	StepStatusSkipped    StepStatus = "skipped"
	StepStatusFailed     StepStatus = "failed"
)

// StepResult contains the result of a workflow step
type StepResult struct {
	Status    string                 `json:"status"`
	Decision  string                 `json:"decision,omitempty"`
	Comments  string                 `json:"comments,omitempty"`
	Data      map[string]interface{} `json:"data,omitempty"`
	CompletedBy string               `json:"completed_by"`
	CompletedAt time.Time            `json:"completed_at"`
}

// WorkflowParticipant represents a participant in a workflow
type WorkflowParticipant struct {
	UserID      string                 `json:"user_id"`
	Role        string                 `json:"role"`
	Permissions []string               `json:"permissions"`
	Status      ParticipantStatus      `json:"status"`
	JoinedAt    time.Time              `json:"joined_at"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// ParticipantStatus defines the status of a workflow participant
type ParticipantStatus string

const (
	ParticipantStatusActive   ParticipantStatus = "active"
	ParticipantStatusInactive ParticipantStatus = "inactive"
	ParticipantStatusCompleted ParticipantStatus = "completed"
)

// WorkflowSettings contains workflow configuration
type WorkflowSettings struct {
	AutoStart           bool                   `json:"auto_start"`
	RequireAllApprovals bool                   `json:"require_all_approvals"`
	AllowParallelSteps  bool                   `json:"allow_parallel_steps"`
	TimeoutDuration     *time.Duration         `json:"timeout_duration,omitempty"`
	NotificationSettings *WorkflowNotificationSettings `json:"notification_settings"`
	EscalationRules     []*EscalationRule      `json:"escalation_rules,omitempty"`
}

// WorkflowNotificationSettings contains workflow notification configuration
type WorkflowNotificationSettings struct {
	NotifyOnStart      bool     `json:"notify_on_start"`
	NotifyOnComplete   bool     `json:"notify_on_complete"`
	NotifyOnStepChange bool     `json:"notify_on_step_change"`
	NotifyOnDeadline   bool     `json:"notify_on_deadline"`
	Recipients         []string `json:"recipients"`
}

// EscalationRule defines when and how to escalate workflow issues
type EscalationRule struct {
	Condition   string        `json:"condition"`
	Delay       time.Duration `json:"delay"`
	Action      string        `json:"action"`
	Recipients  []string      `json:"recipients"`
	Message     string        `json:"message,omitempty"`
}

// Document Collaboration Types

// Document represents a collaborative document
type Document struct {
	ID          string                 `json:"id"`
	Title       string                 `json:"title"`
	Content     string                 `json:"content"`
	Type        DocumentType           `json:"type"`
	Format      string                 `json:"format"`
	TeamID      string                 `json:"team_id,omitempty"`
	ProjectID   string                 `json:"project_id,omitempty"`
	Status      DocumentStatus         `json:"status"`
	Version     int                    `json:"version"`
	Sharing     *DocumentSharing       `json:"sharing"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	CreatedBy   string                 `json:"created_by"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedBy   string                 `json:"updated_by"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// DocumentType defines the type of document
type DocumentType string

const (
	DocumentTypeMarkdown     DocumentType = "markdown"
	DocumentTypeText         DocumentType = "text"
	DocumentTypeCode         DocumentType = "code"
	DocumentTypeSpecification DocumentType = "specification"
	DocumentTypeDesign       DocumentType = "design"
	DocumentTypePresentation DocumentType = "presentation"
	DocumentTypeSpreadsheet  DocumentType = "spreadsheet"
)

// DocumentStatus defines the status of a document
type DocumentStatus string

const (
	DocumentStatusDraft     DocumentStatus = "draft"
	DocumentStatusReview    DocumentStatus = "review"
	DocumentStatusApproved  DocumentStatus = "approved"
	DocumentStatusPublished DocumentStatus = "published"
	DocumentStatusArchived  DocumentStatus = "archived"
)

// DocumentSharing contains document sharing settings
type DocumentSharing struct {
	IsPublic      bool                    `json:"is_public"`
	SharedWith    []*DocumentPermission   `json:"shared_with"`
	LinkSharing   *LinkSharingSettings    `json:"link_sharing,omitempty"`
	ExpiresAt     *time.Time              `json:"expires_at,omitempty"`
}

// DocumentPermission represents permission for a user or team on a document
type DocumentPermission struct {
	Type        string                 `json:"type"` // user, team, role
	ID          string                 `json:"id"`
	Permission  DocumentPermissionType `json:"permission"`
	GrantedBy   string                 `json:"granted_by"`
	GrantedAt   time.Time              `json:"granted_at"`
}

// DocumentPermissionType defines the level of access to a document
type DocumentPermissionType string

const (
	DocumentPermissionRead    DocumentPermissionType = "read"
	DocumentPermissionComment DocumentPermissionType = "comment"
	DocumentPermissionEdit    DocumentPermissionType = "edit"
	DocumentPermissionAdmin   DocumentPermissionType = "admin"
)

// LinkSharingSettings contains settings for link-based sharing
type LinkSharingSettings struct {
	Enabled    bool                   `json:"enabled"`
	Permission DocumentPermissionType `json:"permission"`
	RequireAuth bool                  `json:"require_auth"`
	ExpiresAt  *time.Time             `json:"expires_at,omitempty"`
}

// DocumentVersion represents a version of a document
type DocumentVersion struct {
	ID          string                 `json:"id"`
	DocumentID  string                 `json:"document_id"`
	Version     int                    `json:"version"`
	Content     string                 `json:"content"`
	Changes     string                 `json:"changes,omitempty"`
	CreatedBy   string                 `json:"created_by"`
	CreatedAt   time.Time              `json:"created_at"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// Activity Tracking Types

// Activity represents a user or team activity
type Activity struct {
	ID          string                 `json:"id"`
	Type        ActivityType           `json:"type"`
	UserID      string                 `json:"user_id"`
	Username    string                 `json:"username"`
	TeamID      string                 `json:"team_id,omitempty"`
	ProjectID   string                 `json:"project_id,omitempty"`
	Action      string                 `json:"action"`
	Resource    *ActivityResource      `json:"resource"`
	Description string                 `json:"description"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	Timestamp   time.Time              `json:"timestamp"`
}

// ActivityType defines the type of activity
type ActivityType string

const (
	ActivityTypeTeam        ActivityType = "team"
	ActivityTypeProject     ActivityType = "project"
	ActivityTypeTask        ActivityType = "task"
	ActivityTypeDocument    ActivityType = "document"
	ActivityTypeWorkflow    ActivityType = "workflow"
	ActivityTypeMessage     ActivityType = "message"
	ActivityTypeNotification ActivityType = "notification"
)

// ActivityResource contains information about the resource involved in the activity
type ActivityResource struct {
	Type string `json:"type"`
	ID   string `json:"id"`
	Name string `json:"name"`
	URL  string `json:"url,omitempty"`
}

// Filter Types

// TeamFilter contains filters for team queries
type TeamFilter struct {
	Type      TeamType   `json:"type,omitempty"`
	Status    TeamStatus `json:"status,omitempty"`
	CreatedBy string     `json:"created_by,omitempty"`
	Tags      []string   `json:"tags,omitempty"`
	Search    string     `json:"search,omitempty"`
	Limit     int        `json:"limit,omitempty"`
	Offset    int        `json:"offset,omitempty"`
}

// RoleFilter contains filters for role queries
type RoleFilter struct {
	Type   RoleType `json:"type,omitempty"`
	TeamID string   `json:"team_id,omitempty"`
	Search string   `json:"search,omitempty"`
	Limit  int      `json:"limit,omitempty"`
	Offset int      `json:"offset,omitempty"`
}

// MessageFilter contains filters for message queries
type MessageFilter struct {
	UserID    string      `json:"user_id,omitempty"`
	Type      MessageType `json:"type,omitempty"`
	ThreadID  string      `json:"thread_id,omitempty"`
	Since     *time.Time  `json:"since,omitempty"`
	Until     *time.Time  `json:"until,omitempty"`
	Search    string      `json:"search,omitempty"`
	Limit     int         `json:"limit,omitempty"`
	Offset    int         `json:"offset,omitempty"`
}

// ChannelFilter contains filters for channel queries
type ChannelFilter struct {
	Type     ChannelType `json:"type,omitempty"`
	MemberID string      `json:"member_id,omitempty"`
	Search   string      `json:"search,omitempty"`
	Limit    int         `json:"limit,omitempty"`
	Offset   int         `json:"offset,omitempty"`
}

// NotificationFilter contains filters for notification queries
type NotificationFilter struct {
	Type     NotificationType     `json:"type,omitempty"`
	Priority NotificationPriority `json:"priority,omitempty"`
	Category string               `json:"category,omitempty"`
	IsRead   *bool                `json:"is_read,omitempty"`
	Since    *time.Time           `json:"since,omitempty"`
	Until    *time.Time           `json:"until,omitempty"`
	Limit    int                  `json:"limit,omitempty"`
	Offset   int                  `json:"offset,omitempty"`
}

// WorkflowFilter contains filters for workflow queries
type WorkflowFilter struct {
	Type      WorkflowType   `json:"type,omitempty"`
	Status    WorkflowStatus `json:"status,omitempty"`
	CreatedBy string         `json:"created_by,omitempty"`
	Search    string         `json:"search,omitempty"`
	Limit     int            `json:"limit,omitempty"`
	Offset    int            `json:"offset,omitempty"`
}

// ActivityFilter contains filters for activity queries
type ActivityFilter struct {
	Type      ActivityType `json:"type,omitempty"`
	UserID    string       `json:"user_id,omitempty"`
	TeamID    string       `json:"team_id,omitempty"`
	ProjectID string       `json:"project_id,omitempty"`
	Action    string       `json:"action,omitempty"`
	Since     *time.Time   `json:"since,omitempty"`
	Until     *time.Time   `json:"until,omitempty"`
	Limit     int          `json:"limit,omitempty"`
	Offset    int          `json:"offset,omitempty"`
}
