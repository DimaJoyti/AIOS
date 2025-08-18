# AIOS Team Collaboration Features

## Overview

The AIOS Team Collaboration Features provide a comprehensive platform for team communication, workflow management, document collaboration, and activity tracking. Built with modern collaboration patterns, it offers real-time messaging, role-based permissions, collaborative workflows, document sharing, and advanced notification systems to enhance team productivity and coordination.

## 🏗️ Architecture

### Core Components

```
Team Collaboration Features
├── Team Management (team creation, member management, roles)
├── Communication Engine (channels, messaging, real-time chat)
├── Role & Permission System (RBAC, fine-grained permissions)
├── Workflow Orchestration (collaborative workflows, approvals)
├── Document Collaboration (shared docs, version control)
├── Notification System (multi-channel notifications, preferences)
├── Activity Tracking (audit logs, team analytics)
└── Integration Layer (external tools, webhooks, APIs)
```

### Key Features

- **👥 Team Management**: Complete team lifecycle with member management and role assignment
- **💬 Real-Time Communication**: Multi-channel messaging with threads and file sharing
- **🔐 Role-Based Access Control**: Granular permissions and security policies
- **🔄 Collaborative Workflows**: Multi-step approval processes and task orchestration
- **📄 Document Collaboration**: Shared documents with version control and real-time editing
- **🔔 Smart Notifications**: Multi-channel notifications with user preferences
- **📊 Activity Tracking**: Comprehensive audit logs and team analytics
- **🔗 External Integrations**: Slack, email, webhooks, and third-party tools

## 🚀 Quick Start

### Basic Team Setup

```go
package main

import (
    "time"
    "github.com/aios/aios/pkg/collaboration"
    "github.com/sirupsen/logrus"
)

func main() {
    logger := logrus.New()
    
    // Create collaboration engine
    config := &collaboration.CollaborationEngineConfig{
        MaxTeams:            100,
        MaxChannelsPerTeam:  50,
        MaxMembersPerTeam:   100,
        MessageRetention:    90 * 24 * time.Hour,
        ActivityRetention:   30 * 24 * time.Hour,
        EnableRealTime:      true,
        EnableNotifications: true,
    }
    
    collaborationEngine := collaboration.NewDefaultCollaborationEngine(config, logger)
    
    // Create a development team
    team := &collaboration.Team{
        Name:        "Development Team",
        Description: "Main development team",
        Type:        collaboration.TeamTypeDevelopment,
        Status:      collaboration.TeamStatusActive,
        Settings: &collaboration.TeamSettings{
            IsPublic:             false,
            AllowExternalMembers: false,
            RequireApproval:      true,
            DefaultRole:          "member",
        },
        CreatedBy: "team-lead@company.com",
    }
    
    createdTeam, err := collaborationEngine.CreateTeam(team)
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("Team created: %s (ID: %s)\n", createdTeam.Name, createdTeam.ID)
}
```

## 👥 Team Management

### Creating and Managing Teams

```go
// Create a new team
team := &collaboration.Team{
    Name:        "Full-Stack Development Team",
    Description: "Team responsible for full-stack application development",
    Type:        collaboration.TeamTypeDevelopment,
    Status:      collaboration.TeamStatusActive,
    Settings: &collaboration.TeamSettings{
        IsPublic:             false,
        AllowExternalMembers: false,
        RequireApproval:      true,
        DefaultRole:          "member",
        NotificationSettings: &collaboration.TeamNotificationSettings{
            EnableEmailNotifications: true,
            EnableSlackIntegration:   true,
            EnableWebhooks:          false,
            NotificationChannels:    []string{"general", "development"},
        },
        WorkflowSettings: &collaboration.TeamWorkflowSettings{
            RequireCodeReview:   true,
            MinimumReviewers:    2,
            AutoAssignReviewers: true,
            EnableAutomation:    true,
            WorkflowTemplates:   []string{"code-review", "feature-development"},
        },
    },
    Tags:      []string{"development", "full-stack", "agile"},
    CreatedBy: "team-lead@company.com",
}

createdTeam, err := collaborationEngine.CreateTeam(team)

// Update team settings
team.Description = "Updated team description"
team.Settings.WorkflowSettings.MinimumReviewers = 3
updatedTeam, err := collaborationEngine.UpdateTeam(team)

// List teams with filtering
teams, err := collaborationEngine.ListTeams(&collaboration.TeamFilter{
    Type:   collaboration.TeamTypeDevelopment,
    Status: collaboration.TeamStatusActive,
    Search: "development",
    Limit:  10,
})

// Get team details
team, err := collaborationEngine.GetTeam(teamID)
fmt.Printf("Team: %s with %d members\n", team.Name, len(team.Members))
```

### Member Management

```go
// Add team members
members := []*collaboration.TeamMember{
    {
        UserID:      "alice@company.com",
        Username:    "alice",
        Email:       "alice@company.com",
        DisplayName: "Alice Johnson",
        Role:        "admin",
        Status:      collaboration.MemberStatusActive,
        Permissions: []string{"manage_team", "manage_projects", "code_review"},
    },
    {
        UserID:      "bob@company.com",
        Username:    "bob",
        Email:       "bob@company.com",
        DisplayName: "Bob Smith",
        Role:        "member",
        Status:      collaboration.MemberStatusActive,
        Permissions: []string{"participate_projects", "code_review"},
    },
}

for _, member := range members {
    err := collaborationEngine.AddTeamMember(teamID, member)
    if err != nil {
        fmt.Printf("Failed to add member %s: %v\n", member.DisplayName, err)
        continue
    }
    fmt.Printf("Added member: %s (%s)\n", member.DisplayName, member.Role)
}

// Update member role
member.Role = "senior_member"
member.Permissions = append(member.Permissions, "mentor_junior")
err = collaborationEngine.UpdateTeamMember(teamID, member)

// Remove team member
err = collaborationEngine.RemoveTeamMember(teamID, "bob@company.com")

// Get team members
members, err := collaborationEngine.GetTeamMembers(teamID)
for _, member := range members {
    fmt.Printf("Member: %s (%s) - %s\n", 
        member.DisplayName, member.Role, member.Status)
}
```

## 🔐 Role and Permission Management

### Creating Custom Roles

```go
// Create a custom role
role := &collaboration.Role{
    Name:        "Senior Developer",
    Description: "Senior developer with code review and mentoring responsibilities",
    Type:        collaboration.RoleTypeCustom,
    Permissions: []*collaboration.Permission{
        {
            Name:        "code_review",
            Description: "Review and approve code changes",
            Resource:    "code",
            Action:      "review,approve",
            Scope:       collaboration.PermissionScopeProject,
        },
        {
            Name:        "mentor_junior",
            Description: "Mentor junior developers",
            Resource:    "team",
            Action:      "mentor",
            Scope:       collaboration.PermissionScopeTeam,
        },
        {
            Name:        "deploy_staging",
            Description: "Deploy to staging environment",
            Resource:    "deployment",
            Action:      "deploy",
            Scope:       collaboration.PermissionScopeProject,
            Conditions: []*collaboration.PermissionCondition{
                {
                    Field:    "environment",
                    Operator: "equals",
                    Value:    "staging",
                },
            },
        },
    },
    TeamID:    teamID,
    CreatedBy: "admin@company.com",
}

createdRole, err := collaborationEngine.CreateRole(role)

// Assign role to user
err = collaborationEngine.AssignRole("alice@company.com", createdRole.ID)

// List available roles
roles, err := collaborationEngine.ListRoles(&collaboration.RoleFilter{
    Type:   collaboration.RoleTypeCustom,
    TeamID: teamID,
    Search: "developer",
})

// Revoke role from user
err = collaborationEngine.RevokeRole("alice@company.com", roleID)
```

## 💬 Communication and Messaging

### Channel Management

```go
// Create communication channels
channels := []*collaboration.Channel{
    {
        Name:        "development",
        Description: "Development discussions and updates",
        Type:        collaboration.ChannelTypePublic,
        TeamID:      teamID,
        Members:     []string{"alice@company.com", "bob@company.com", "charlie@company.com"},
        Settings: &collaboration.ChannelSettings{
            IsArchived:       false,
            AllowThreads:     true,
            AllowFileUploads: true,
            RetentionDays:    90,
        },
        CreatedBy: "alice@company.com",
    },
    {
        Name:        "code-reviews",
        Description: "Code review discussions and feedback",
        Type:        collaboration.ChannelTypePublic,
        TeamID:      teamID,
        Members:     []string{"alice@company.com", "bob@company.com"},
        Settings: &collaboration.ChannelSettings{
            IsArchived:       false,
            AllowThreads:     true,
            AllowFileUploads: true,
            RetentionDays:    180,
        },
        CreatedBy: "alice@company.com",
    },
}

for _, channel := range channels {
    createdChannel, err := collaborationEngine.CreateChannel(channel)
    if err != nil {
        continue
    }
    fmt.Printf("Created channel: #%s (%d members)\n", 
        createdChannel.Name, len(createdChannel.Members))
}

// List team channels
channels, err := collaborationEngine.ListChannels(teamID, &collaboration.ChannelFilter{
    Type: collaboration.ChannelTypePublic,
})

// Get specific channel
channel, err := collaborationEngine.GetChannel(channelID)
```

### Real-Time Messaging

```go
// Send messages to channels
messages := []*collaboration.Message{
    {
        ChannelID: channelID,
        UserID:    "alice@company.com",
        Username:  "alice",
        Content:   "Good morning team! Let's discuss today's sprint goals.",
        Type:      collaboration.MessageTypeText,
    },
    {
        ChannelID: channelID,
        UserID:    "bob@company.com",
        Username:  "bob",
        Content:   "I've completed the user authentication feature. Ready for code review!",
        Type:      collaboration.MessageTypeText,
        Mentions:  []string{"alice@company.com", "diana@company.com"},
    },
    {
        ChannelID: channelID,
        UserID:    "charlie@company.com",
        Username:  "charlie",
        Content:   "```go\nfunc authenticate(token string) bool {\n    return validateToken(token)\n}\n```",
        Type:      collaboration.MessageTypeCode,
    },
}

for _, message := range messages {
    sentMessage, err := collaborationEngine.SendMessage(message)
    if err != nil {
        continue
    }
    fmt.Printf("%s: %s\n", sentMessage.Username, sentMessage.Content)
}

// Get channel messages
messages, err := collaborationEngine.GetMessages(channelID, &collaboration.MessageFilter{
    Since: &[]time.Time{time.Now().Add(-24 * time.Hour)}[0],
    Limit: 50,
})

// Get messages with threading
threadMessages, err := collaborationEngine.GetMessages(channelID, &collaboration.MessageFilter{
    ThreadID: parentMessageID,
    Limit:    20,
})

// Search messages
searchResults, err := collaborationEngine.GetMessages(channelID, &collaboration.MessageFilter{
    Search: "authentication",
    Limit:  10,
})
```

## 🔄 Collaborative Workflows

### Creating Approval Workflows

```go
// Create a code review workflow
codeReviewWorkflow := &collaboration.CollaborativeWorkflow{
    Name:        "Feature Code Review Process",
    Description: "Standard code review workflow for new features",
    TeamID:      teamID,
    Type:        collaboration.WorkflowTypeCodeReview,
    Status:      collaboration.WorkflowStatusActive,
    Steps: []*collaboration.WorkflowStep{
        {
            Name:        "Initial Review",
            Description: "First pass review by senior developer",
            Type:        collaboration.StepTypeReview,
            Assignees:   []string{"alice@company.com"},
            DueDate:     &[]time.Time{time.Now().Add(24 * time.Hour)}[0],
        },
        {
            Name:        "Security Review",
            Description: "Security-focused code review",
            Type:        collaboration.StepTypeReview,
            Assignees:   []string{"diana@company.com"},
            Dependencies: []string{"Initial Review"},
            DueDate:     &[]time.Time{time.Now().Add(48 * time.Hour)}[0],
        },
        {
            Name:        "Final Approval",
            Description: "Final approval and merge authorization",
            Type:        collaboration.StepTypeApproval,
            Assignees:   []string{"alice@company.com"},
            Dependencies: []string{"Security Review"},
            DueDate:     &[]time.Time{time.Now().Add(72 * time.Hour)}[0],
        },
    },
    Participants: []*collaboration.WorkflowParticipant{
        {
            UserID:      "bob@company.com",
            Role:        "author",
            Permissions: []string{"view", "comment"},
            Status:      collaboration.ParticipantStatusActive,
        },
        {
            UserID:      "alice@company.com",
            Role:        "reviewer",
            Permissions: []string{"view", "comment", "approve"},
            Status:      collaboration.ParticipantStatusActive,
        },
    },
    Settings: &collaboration.WorkflowSettings{
        AutoStart:           true,
        RequireAllApprovals: true,
        AllowParallelSteps:  false,
        TimeoutDuration:     &[]time.Duration{7 * 24 * time.Hour}[0],
        NotificationSettings: &collaboration.WorkflowNotificationSettings{
            NotifyOnStart:      true,
            NotifyOnComplete:   true,
            NotifyOnStepChange: true,
            NotifyOnDeadline:   true,
            Recipients:         []string{"alice@company.com", "bob@company.com"},
        },
        EscalationRules: []*collaboration.EscalationRule{
            {
                Condition:  "step_overdue",
                Delay:      24 * time.Hour,
                Action:     "notify_manager",
                Recipients: []string{"manager@company.com"},
                Message:    "Code review is overdue and requires attention",
            },
        },
    },
    CreatedBy: "alice@company.com",
}

createdWorkflow, err := collaborationEngine.CreateCollaborativeWorkflow(codeReviewWorkflow)

// List team workflows
workflows, err := collaborationEngine.ListCollaborativeWorkflows(teamID, &collaboration.WorkflowFilter{
    Type:   collaboration.WorkflowTypeCodeReview,
    Status: collaboration.WorkflowStatusActive,
})

// Get workflow details
workflow, err := collaborationEngine.GetCollaborativeWorkflow(workflowID)
fmt.Printf("Workflow: %s (%s) - %d steps\n", 
    workflow.Name, workflow.Status, len(workflow.Steps))

// Update workflow status
workflow.Status = collaboration.WorkflowStatusCompleted
updatedWorkflow, err := collaborationEngine.UpdateCollaborativeWorkflow(workflow)
```

### Decision-Making Workflows

```go
// Create a decision-making workflow
decisionWorkflow := &collaboration.CollaborativeWorkflow{
    Name:        "Architecture Decision Process",
    Description: "Process for making architectural decisions",
    TeamID:      teamID,
    Type:        collaboration.WorkflowTypeDecisionMaking,
    Steps: []*collaboration.WorkflowStep{
        {
            Name:        "Proposal Review",
            Description: "Review the architectural proposal",
            Type:        collaboration.StepTypeReview,
            Assignees:   []string{"alice@company.com", "bob@company.com", "charlie@company.com"},
        },
        {
            Name:        "Technical Discussion",
            Description: "Discuss technical implications",
            Type:        collaboration.StepTypeTask,
            Assignees:   []string{"alice@company.com"},
            Dependencies: []string{"Proposal Review"},
        },
        {
            Name:        "Final Decision",
            Description: "Make the final architectural decision",
            Type:        collaboration.StepTypeDecision,
            Assignees:   []string{"alice@company.com"},
            Dependencies: []string{"Technical Discussion"},
        },
    ],
    Settings: &collaboration.WorkflowSettings{
        RequireAllApprovals: false, // Majority approval
        AllowParallelSteps:  true,
    },
    CreatedBy: "alice@company.com",
}

createdDecisionWorkflow, err := collaborationEngine.CreateCollaborativeWorkflow(decisionWorkflow)
```

## 📄 Document Collaboration

### Creating and Managing Documents

```go
// Create a collaborative document
document := &collaboration.Document{
    Title:     "API Design Guidelines",
    Content:   `# API Design Guidelines

## RESTful API Principles
1. Use HTTP methods appropriately (GET, POST, PUT, DELETE)
2. Use meaningful resource names
3. Implement proper status codes
4. Version your APIs

## Authentication
- Use JWT tokens for authentication
- Implement proper token refresh mechanisms
- Use HTTPS for all API endpoints

## Error Handling
- Return consistent error response format
- Include error codes and descriptive messages
- Log errors for debugging purposes
`,
    Type:      collaboration.DocumentTypeMarkdown,
    Format:    "markdown",
    TeamID:    teamID,
    ProjectID: "project-api-redesign",
    Status:    collaboration.DocumentStatusReview,
    Sharing: &collaboration.DocumentSharing{
        IsPublic: false,
        SharedWith: []*collaboration.DocumentPermission{
            {
                Type:       "team",
                ID:         teamID,
                Permission: collaboration.DocumentPermissionEdit,
                GrantedBy:  "alice@company.com",
                GrantedAt:  time.Now(),
            },
            {
                Type:       "user",
                ID:         "external-consultant@company.com",
                Permission: collaboration.DocumentPermissionComment,
                GrantedBy:  "alice@company.com",
                GrantedAt:  time.Now(),
            },
        },
        LinkSharing: &collaboration.LinkSharingSettings{
            Enabled:     true,
            Permission:  collaboration.DocumentPermissionRead,
            RequireAuth: true,
            ExpiresAt:   &[]time.Time{time.Now().Add(30 * 24 * time.Hour)}[0],
        },
    },
    CreatedBy: "alice@company.com",
}

createdDoc, err := collaborationEngine.CreateDocument(document)

// Update document content
document.Content += "\n\n## Rate Limiting\n- Implement rate limiting to prevent abuse\n- Use sliding window or token bucket algorithms"
document.UpdatedBy = "bob@company.com"
updatedDoc, err := collaborationEngine.UpdateDocument(document)

// Share document with additional users
newSharing := &collaboration.DocumentSharing{
    IsPublic: false,
    SharedWith: []*collaboration.DocumentPermission{
        {
            Type:       "user",
            ID:         "product-manager@company.com",
            Permission: collaboration.DocumentPermissionRead,
            GrantedBy:  "alice@company.com",
            GrantedAt:  time.Now(),
        },
    },
}

err = collaborationEngine.ShareDocument(document.ID, newSharing)

// Get document versions
versions, err := collaborationEngine.GetDocumentVersions(document.ID)
for _, version := range versions {
    fmt.Printf("Version %d: %s (by %s)\n", 
        version.Version, version.Changes, version.CreatedBy)
}

// Get document
doc, err := collaborationEngine.GetDocument(documentID)
fmt.Printf("Document: %s (v%d) - %s\n", 
    doc.Title, doc.Version, doc.Status)
```

## 🔔 Notification System

### Sending Notifications

```go
// Send different types of notifications
notifications := []*collaboration.Notification{
    {
        UserID:   "bob@company.com",
        Type:     collaboration.NotificationTypeMention,
        Title:    "Code Review Request",
        Message:  "Alice mentioned you in #development regarding your authentication feature.",
        Priority: collaboration.NotificationPriorityHigh,
        Category: "code_review",
        Source: &collaboration.NotificationSource{
            Type:    "channel",
            ID:      channelID,
            Name:    "development",
            URL:     fmt.Sprintf("/channels/%s", channelID),
        },
        Actions: []*collaboration.NotificationAction{
            {
                ID:    "view_message",
                Label: "View Message",
                URL:   fmt.Sprintf("/channels/%s", channelID),
                Type:  "link",
            },
            {
                ID:    "mark_reviewed",
                Label: "Mark as Reviewed",
                Type:  "api_call",
            },
        },
        ExpiresAt: &[]time.Time{time.Now().Add(7 * 24 * time.Hour)}[0],
    },
    {
        UserID:   "diana@company.com",
        Type:     collaboration.NotificationTypeAssignment,
        Title:    "Security Review Assignment",
        Message:  "You have been assigned to review the security aspects of the authentication feature.",
        Priority: collaboration.NotificationPriorityNormal,
        Category: "workflow",
        Source: &collaboration.NotificationSource{
            Type:    "workflow",
            ID:      workflowID,
            Name:    "Feature Code Review Process",
            URL:     fmt.Sprintf("/workflows/%s", workflowID),
        },
    },
    {
        UserID:   "charlie@company.com",
        Type:     collaboration.NotificationTypeDeadline,
        Title:    "Sprint Deadline Approaching",
        Message:  "The current sprint ends in 2 days. Please complete your assigned tasks.",
        Priority: collaboration.NotificationPriorityUrgent,
        Category: "deadline",
    },
}

for _, notification := range notifications {
    err := collaborationEngine.SendNotification(notification)
    if err != nil {
        continue
    }
    fmt.Printf("Sent notification: %s to %s\n", 
        notification.Title, notification.UserID)
}

// Get user notifications
userNotifications, err := collaborationEngine.GetNotifications("bob@company.com", &collaboration.NotificationFilter{
    IsRead:   &[]bool{false}[0], // Unread notifications only
    Priority: collaboration.NotificationPriorityHigh,
    Since:    &[]time.Time{time.Now().Add(-24 * time.Hour)}[0],
    Limit:    20,
})

// Mark notification as read
err = collaborationEngine.MarkNotificationRead(notificationID)
```

### Notification Preferences

```go
// Get user notification settings
settings, err := collaborationEngine.GetNotificationSettings("alice@company.com")

// Update notification preferences
newSettings := &collaboration.NotificationSettings{
    UserID: "alice@company.com",
    EmailNotifications: &collaboration.EmailNotificationSettings{
        Enabled:    true,
        Frequency:  "daily", // immediate, hourly, daily, weekly
        Categories: []string{"mention", "assignment", "deadline"},
        DigestTime: "09:00",
    },
    PushNotifications: &collaboration.PushNotificationSettings{
        Enabled:    true,
        Categories: []string{"mention", "assignment"},
        Sound:      true,
        Vibration:  false,
    },
    InAppNotifications: &collaboration.InAppNotificationSettings{
        Enabled:    true,
        Categories: []string{"mention", "assignment", "deadline", "approval"},
        ShowBadge:  true,
        AutoRead:   false,
    },
    CategorySettings: map[string]*collaboration.CategorySettings{
        "mention": {
            Enabled:  true,
            Priority: collaboration.NotificationPriorityHigh,
            Channels: []string{"email", "push", "in_app"},
        },
        "assignment": {
            Enabled:  true,
            Priority: collaboration.NotificationPriorityNormal,
            Channels: []string{"email", "in_app"},
        },
        "deadline": {
            Enabled:  true,
            Priority: collaboration.NotificationPriorityHigh,
            Channels: []string{"email", "push", "in_app"},
        },
    },
    QuietHours: &collaboration.QuietHours{
        Enabled:   true,
        StartTime: "22:00",
        EndTime:   "08:00",
        Timezone:  "UTC",
        Weekdays:  []int{1, 2, 3, 4, 5}, // Monday to Friday
    },
}

err = collaborationEngine.UpdateNotificationSettings("alice@company.com", newSettings)
```

## 📊 Activity Tracking and Analytics

### Activity Monitoring

```go
// Log custom activities
activity := &collaboration.Activity{
    Type:        collaboration.ActivityTypeProject,
    UserID:      "alice@company.com",
    Username:    "alice",
    TeamID:      teamID,
    ProjectID:   "project-api-redesign",
    Action:      "milestone_completed",
    Resource: &collaboration.ActivityResource{
        Type: "milestone",
        ID:   "milestone-mvp",
        Name: "MVP Milestone",
        URL:  "/projects/project-api-redesign/milestones/milestone-mvp",
    },
    Description: "Completed MVP milestone for API redesign project",
}

err := collaborationEngine.LogActivity(activity)

// Get team activities
teamActivities, err := collaborationEngine.GetTeamActivity(teamID, &collaboration.ActivityFilter{
    Type:  collaboration.ActivityTypeTeam,
    Since: &[]time.Time{time.Now().Add(-7 * 24 * time.Hour)}[0],
    Limit: 50,
})

// Get user activities
userActivities, err := collaborationEngine.GetUserActivity("alice@company.com", &collaboration.ActivityFilter{
    Action: "created",
    Since:  &[]time.Time{time.Now().Add(-30 * 24 * time.Hour)}[0],
    Limit:  100,
})

// Get all activities with filtering
allActivities, err := collaborationEngine.GetActivities(&collaboration.ActivityFilter{
    Type:      collaboration.ActivityTypeDocument,
    ProjectID: "project-api-redesign",
    Since:     &[]time.Time{time.Now().Add(-24 * time.Hour)}[0],
    Limit:     25,
})

// Analyze team activity patterns
fmt.Printf("Team Activity Summary:\n")
for _, activity := range teamActivities {
    fmt.Printf("  %s: %s (%s)\n", 
        activity.Username, activity.Description, 
        activity.Timestamp.Format("2006-01-02 15:04"))
}
```

### Team Analytics

```go
// Get comprehensive team analytics
team, err := collaborationEngine.GetTeam(teamID)
if err != nil {
    return err
}

fmt.Printf("Team Analytics for: %s\n", team.Name)
fmt.Printf("==========================================\n")

// Team overview
fmt.Printf("Team Overview:\n")
fmt.Printf("  Members: %d\n", len(team.Members))
fmt.Printf("  Channels: %d\n", len(team.Channels))
fmt.Printf("  Type: %s\n", team.Type)
fmt.Printf("  Status: %s\n", team.Status)

// Member activity levels
fmt.Printf("\nMember Activity (Last 7 days):\n")
for _, member := range team.Members {
    activities, _ := collaborationEngine.GetUserActivity(member.UserID, &collaboration.ActivityFilter{
        TeamID: teamID,
        Since:  &[]time.Time{time.Now().Add(-7 * 24 * time.Hour)}[0],
    })
    fmt.Printf("  %s: %d activities\n", member.DisplayName, len(activities))
}

// Channel statistics
fmt.Printf("\nChannel Activity:\n")
for _, channel := range team.Channels {
    messages, _ := collaborationEngine.GetMessages(channel.ID, &collaboration.MessageFilter{
        Since: &[]time.Time{time.Now().Add(-7 * 24 * time.Hour)}[0],
    })
    fmt.Printf("  #%s: %d messages, %d members\n", 
        channel.Name, len(messages), len(channel.Members))
}

// Workflow statistics
workflows, _ := collaborationEngine.ListCollaborativeWorkflows(teamID, &collaboration.WorkflowFilter{})
fmt.Printf("\nWorkflow Summary:\n")
fmt.Printf("  Total Workflows: %d\n", len(workflows))

statusCounts := make(map[collaboration.WorkflowStatus]int)
for _, workflow := range workflows {
    statusCounts[workflow.Status]++
}

for status, count := range statusCounts {
    fmt.Printf("  %s: %d\n", status, count)
}
```

## 🧪 Testing

Run the comprehensive test suite:

```bash
# Run all collaboration tests
go test ./pkg/collaboration/...

# Run with race detection
go test -race ./pkg/collaboration/...

# Run integration tests
go test -tags=integration ./pkg/collaboration/...

# Run team collaboration example
go run examples/team_collaboration_example.go
```

## 📖 Examples

See the complete example in `examples/team_collaboration_example.go` for a comprehensive demonstration including:

- Development team creation with role-based permissions
- Multi-channel communication setup
- Real-time messaging with mentions and threads
- Code review workflow orchestration
- Document collaboration with version control
- Multi-channel notification system
- Activity tracking and team analytics
- Member management and role assignment

## 🤝 Contributing

1. Follow established patterns and interfaces
2. Add comprehensive tests for new collaboration features
3. Update documentation for API changes
4. Ensure proper error handling and logging
5. Use OpenTelemetry for observability
6. Implement proper security and permission checks

## 📄 License

This Team Collaboration Features system is part of the AIOS project and follows the same licensing terms.
