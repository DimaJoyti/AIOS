package main

import (
	"fmt"
	"log"
	"time"

	"github.com/aios/aios/pkg/collaboration"
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

	fmt.Println("🤝 AIOS Team Collaboration Features Demo")
	fmt.Println("========================================")

	// Run the comprehensive demo
	if err := runTeamCollaborationDemo(logger); err != nil {
		log.Fatalf("Demo failed: %v", err)
	}

	fmt.Println("\n✅ Team Collaboration Demo completed successfully!")
}

func runTeamCollaborationDemo(logger *logrus.Logger) error {
	// Step 1: Create Collaboration Engine
	fmt.Println("\n1. Creating Collaboration Engine...")
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
	fmt.Println("✓ Collaboration Engine created successfully")

	// Step 2: Create Development Team
	fmt.Println("\n2. Creating Development Team...")
	devTeam := &collaboration.Team{
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

	createdTeam, err := collaborationEngine.CreateTeam(devTeam)
	if err != nil {
		return fmt.Errorf("failed to create team: %w", err)
	}

	fmt.Printf("   ✓ Team Created: %s (ID: %s)\n", createdTeam.Name, createdTeam.ID)
	fmt.Printf("     - Type: %s\n", createdTeam.Type)
	fmt.Printf("     - Status: %s\n", createdTeam.Status)
	fmt.Printf("     - Code Review Required: %t\n", createdTeam.Settings.WorkflowSettings.RequireCodeReview)
	fmt.Printf("     - Minimum Reviewers: %d\n", createdTeam.Settings.WorkflowSettings.MinimumReviewers)

	// Step 3: Add Team Members
	fmt.Println("\n3. Adding Team Members...")
	
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
		{
			UserID:      "charlie@company.com",
			Username:    "charlie",
			Email:       "charlie@company.com",
			DisplayName: "Charlie Brown",
			Role:        "member",
			Status:      collaboration.MemberStatusActive,
			Permissions: []string{"participate_projects"},
		},
		{
			UserID:      "diana@company.com",
			Username:    "diana",
			Email:       "diana@company.com",
			DisplayName: "Diana Prince",
			Role:        "member",
			Status:      collaboration.MemberStatusActive,
			Permissions: []string{"participate_projects", "code_review"},
		},
	}

	for _, member := range members {
		err := collaborationEngine.AddTeamMember(createdTeam.ID, member)
		if err != nil {
			return fmt.Errorf("failed to add team member %s: %w", member.DisplayName, err)
		}
		fmt.Printf("   ✓ Added member: %s (%s) - Role: %s\n", 
			member.DisplayName, member.Username, member.Role)
	}

	// Step 4: Create Communication Channels
	fmt.Println("\n4. Creating Communication Channels...")
	
	channels := []*collaboration.Channel{
		{
			Name:        "development",
			Description: "Development discussions and updates",
			Type:        collaboration.ChannelTypePublic,
			TeamID:      createdTeam.ID,
			Members:     []string{"alice@company.com", "bob@company.com", "charlie@company.com", "diana@company.com"},
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
			TeamID:      createdTeam.ID,
			Members:     []string{"alice@company.com", "bob@company.com", "diana@company.com"},
			Settings: &collaboration.ChannelSettings{
				IsArchived:       false,
				AllowThreads:     true,
				AllowFileUploads: true,
				RetentionDays:    180,
			},
			CreatedBy: "alice@company.com",
		},
		{
			Name:        "random",
			Description: "Casual conversations and team bonding",
			Type:        collaboration.ChannelTypePublic,
			TeamID:      createdTeam.ID,
			Members:     []string{"alice@company.com", "bob@company.com", "charlie@company.com", "diana@company.com"},
			Settings: &collaboration.ChannelSettings{
				IsArchived:       false,
				AllowThreads:     true,
				AllowFileUploads: true,
				RetentionDays:    30,
			},
			CreatedBy: "bob@company.com",
		},
	}

	for _, channel := range channels {
		createdChannel, err := collaborationEngine.CreateChannel(channel)
		if err != nil {
			return fmt.Errorf("failed to create channel %s: %w", channel.Name, err)
		}
		fmt.Printf("   ✓ Created channel: #%s (%s) - %d members\n", 
			createdChannel.Name, createdChannel.Type, len(createdChannel.Members))
	}

	// Step 5: Send Messages
	fmt.Println("\n5. Team Communication...")
	
	// Get development channel
	teamChannels, err := collaborationEngine.ListChannels(createdTeam.ID, &collaboration.ChannelFilter{
		Type: collaboration.ChannelTypePublic,
	})
	if err != nil {
		return fmt.Errorf("failed to list channels: %w", err)
	}

	var devChannelID string
	for _, channel := range teamChannels {
		if channel.Name == "development" {
			devChannelID = channel.ID
			break
		}
	}

	messages := []*collaboration.Message{
		{
			ChannelID: devChannelID,
			UserID:    "alice@company.com",
			Username:  "alice",
			Content:   "Good morning team! Let's discuss today's sprint goals.",
			Type:      collaboration.MessageTypeText,
		},
		{
			ChannelID: devChannelID,
			UserID:    "bob@company.com",
			Username:  "bob",
			Content:   "I've completed the user authentication feature. Ready for code review!",
			Type:      collaboration.MessageTypeText,
			Mentions:  []string{"alice@company.com", "diana@company.com"},
		},
		{
			ChannelID: devChannelID,
			UserID:    "charlie@company.com",
			Username:  "charlie",
			Content:   "Working on the frontend components. Should have them ready by EOD.",
			Type:      collaboration.MessageTypeText,
		},
		{
			ChannelID: devChannelID,
			UserID:    "diana@company.com",
			Username:  "diana",
			Content:   "@bob I'll review your authentication code this afternoon. Great work!",
			Type:      collaboration.MessageTypeText,
			Mentions:  []string{"bob@company.com"},
		},
	}

	for _, message := range messages {
		sentMessage, err := collaborationEngine.SendMessage(message)
		if err != nil {
			return fmt.Errorf("failed to send message: %w", err)
		}
		fmt.Printf("   ✓ %s: %s\n", sentMessage.Username, sentMessage.Content)
	}

	// Step 6: Create Collaborative Workflow
	fmt.Println("\n6. Creating Code Review Workflow...")
	
	codeReviewWorkflow := &collaboration.CollaborativeWorkflow{
		Name:        "Feature Code Review Process",
		Description: "Standard code review workflow for new features",
		TeamID:      createdTeam.ID,
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
				JoinedAt:    time.Now(),
			},
			{
				UserID:      "alice@company.com",
				Role:        "reviewer",
				Permissions: []string{"view", "comment", "approve"},
				Status:      collaboration.ParticipantStatusActive,
				JoinedAt:    time.Now(),
			},
			{
				UserID:      "diana@company.com",
				Role:        "security_reviewer",
				Permissions: []string{"view", "comment", "approve"},
				Status:      collaboration.ParticipantStatusActive,
				JoinedAt:    time.Now(),
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
				Recipients:         []string{"alice@company.com", "bob@company.com", "diana@company.com"},
			},
		},
		CreatedBy: "alice@company.com",
	}

	createdWorkflow, err := collaborationEngine.CreateCollaborativeWorkflow(codeReviewWorkflow)
	if err != nil {
		return fmt.Errorf("failed to create workflow: %w", err)
	}

	fmt.Printf("   ✓ Workflow Created: %s (ID: %s)\n", createdWorkflow.Name, createdWorkflow.ID)
	fmt.Printf("     - Type: %s\n", createdWorkflow.Type)
	fmt.Printf("     - Status: %s\n", createdWorkflow.Status)
	fmt.Printf("     - Steps: %d\n", len(createdWorkflow.Steps))
	fmt.Printf("     - Participants: %d\n", len(createdWorkflow.Participants))

	for i, step := range createdWorkflow.Steps {
		fmt.Printf("       Step %d: %s (%s) - Assignees: %d\n", 
			i+1, step.Name, step.Type, len(step.Assignees))
	}

	// Step 7: Create Collaborative Document
	fmt.Println("\n7. Creating Team Documentation...")
	
	teamDoc := &collaboration.Document{
		Title:     "Team Development Guidelines",
		Content:   `# Team Development Guidelines

## Code Review Process
1. All code must be reviewed by at least 2 team members
2. Security-sensitive code requires additional security review
3. All tests must pass before merge

## Communication Guidelines
- Use #development for technical discussions
- Use #code-reviews for review feedback
- Tag relevant team members in important messages

## Workflow Standards
- Follow the established code review workflow
- Update documentation for new features
- Maintain test coverage above 80%

## Team Practices
- Daily standups at 9:00 AM
- Sprint planning every two weeks
- Retrospectives at end of each sprint
`,
		Type:      collaboration.DocumentTypeMarkdown,
		Format:    "markdown",
		TeamID:    createdTeam.ID,
		Status:    collaboration.DocumentStatusPublished,
		Sharing: &collaboration.DocumentSharing{
			IsPublic: false,
			SharedWith: []*collaboration.DocumentPermission{
				{
					Type:       "team",
					ID:         createdTeam.ID,
					Permission: collaboration.DocumentPermissionRead,
					GrantedBy:  "alice@company.com",
					GrantedAt:  time.Now(),
				},
			},
		},
		CreatedBy: "alice@company.com",
	}

	createdDoc, err := collaborationEngine.CreateDocument(teamDoc)
	if err != nil {
		return fmt.Errorf("failed to create document: %w", err)
	}

	fmt.Printf("   ✓ Document Created: %s (ID: %s)\n", createdDoc.Title, createdDoc.ID)
	fmt.Printf("     - Type: %s\n", createdDoc.Type)
	fmt.Printf("     - Status: %s\n", createdDoc.Status)
	fmt.Printf("     - Version: %d\n", createdDoc.Version)
	fmt.Printf("     - Content Length: %d characters\n", len(createdDoc.Content))

	// Step 8: Send Notifications
	fmt.Println("\n8. Sending Team Notifications...")
	
	notifications := []*collaboration.Notification{
		{
			UserID:   "bob@company.com",
			Type:     collaboration.NotificationTypeMention,
			Title:    "Code Review Request",
			Message:  "Alice mentioned you in #development regarding your authentication feature code review.",
			Priority: collaboration.NotificationPriorityHigh,
			Category: "code_review",
			Source: &collaboration.NotificationSource{
				Type:    "channel",
				ID:      devChannelID,
				Name:    "development",
				URL:     fmt.Sprintf("/channels/%s", devChannelID),
			},
			Actions: []*collaboration.NotificationAction{
				{
					ID:    "view_message",
					Label: "View Message",
					URL:   fmt.Sprintf("/channels/%s", devChannelID),
					Type:  "link",
				},
			},
		},
		{
			UserID:   "diana@company.com",
			Type:     collaboration.NotificationTypeAssignment,
			Title:    "Security Review Assignment",
			Message:  "You have been assigned to review the security aspects of Bob's authentication feature.",
			Priority: collaboration.NotificationPriorityNormal,
			Category: "workflow",
			Source: &collaboration.NotificationSource{
				Type:    "workflow",
				ID:      createdWorkflow.ID,
				Name:    createdWorkflow.Name,
				URL:     fmt.Sprintf("/workflows/%s", createdWorkflow.ID),
			},
			Actions: []*collaboration.NotificationAction{
				{
					ID:    "view_workflow",
					Label: "View Workflow",
					URL:   fmt.Sprintf("/workflows/%s", createdWorkflow.ID),
					Type:  "link",
				},
			},
		},
		{
			UserID:   "charlie@company.com",
			Type:     collaboration.NotificationTypeInfo,
			Title:    "New Team Document",
			Message:  "Alice has created new team development guidelines. Please review when you have time.",
			Priority: collaboration.NotificationPriorityLow,
			Category: "documentation",
			Source: &collaboration.NotificationSource{
				Type:    "document",
				ID:      createdDoc.ID,
				Name:    createdDoc.Title,
				URL:     fmt.Sprintf("/documents/%s", createdDoc.ID),
			},
		},
	}

	for _, notification := range notifications {
		err := collaborationEngine.SendNotification(notification)
		if err != nil {
			return fmt.Errorf("failed to send notification: %w", err)
		}
		fmt.Printf("   ✓ Sent to %s: %s (%s)\n", 
			notification.UserID, notification.Title, notification.Priority)
	}

	// Step 9: Activity Tracking
	fmt.Println("\n9. Reviewing Team Activity...")
	
	// Get recent team activities
	teamActivities, err := collaborationEngine.GetTeamActivity(createdTeam.ID, &collaboration.ActivityFilter{
		Limit: 10,
	})
	if err != nil {
		return fmt.Errorf("failed to get team activities: %w", err)
	}

	fmt.Printf("   ✓ Recent Team Activities (%d total):\n", len(teamActivities))
	for i, activity := range teamActivities {
		if i >= 5 { // Show only first 5
			break
		}
		fmt.Printf("     %d. %s: %s (%s)\n", 
			i+1, activity.Username, activity.Description, activity.Timestamp.Format("15:04:05"))
	}

	// Step 10: Team Analytics
	fmt.Println("\n10. Team Collaboration Analytics...")
	
	// Get team information
	team, err := collaborationEngine.GetTeam(createdTeam.ID)
	if err != nil {
		return fmt.Errorf("failed to get team: %w", err)
	}

	fmt.Printf("   ✓ Team Overview:\n")
	fmt.Printf("     - Team: %s\n", team.Name)
	fmt.Printf("     - Members: %d\n", len(team.Members))
	fmt.Printf("     - Channels: %d\n", len(team.Channels))
	fmt.Printf("     - Type: %s\n", team.Type)
	fmt.Printf("     - Status: %s\n", team.Status)

	// Get channel statistics
	fmt.Printf("   ✓ Communication Channels:\n")
	for _, channel := range team.Channels {
		messages, _ := collaborationEngine.GetMessages(channel.ID, &collaboration.MessageFilter{
			Limit: 100,
		})
		fmt.Printf("     - #%s: %d messages, %d members\n", 
			channel.Name, len(messages), len(channel.Members))
	}

	// Get workflow statistics
	workflows, err := collaborationEngine.ListCollaborativeWorkflows(createdTeam.ID, &collaboration.WorkflowFilter{
		Limit: 10,
	})
	if err != nil {
		return fmt.Errorf("failed to list workflows: %w", err)
	}

	fmt.Printf("   ✓ Active Workflows: %d\n", len(workflows))
	for _, workflow := range workflows {
		fmt.Printf("     - %s (%s): %d steps, %d participants\n", 
			workflow.Name, workflow.Status, len(workflow.Steps), len(workflow.Participants))
	}

	return nil
}
