package collaboration

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
)

// Workflow and document collaboration methods

// CreateCollaborativeWorkflow creates a new collaborative workflow
func (ce *DefaultCollaborationEngine) CreateCollaborativeWorkflow(workflow *CollaborativeWorkflow) (*CollaborativeWorkflow, error) {
	_, span := ce.tracer.Start(context.Background(), "collaboration.create_workflow")
	defer span.End()

	if workflow.ID == "" {
		workflow.ID = uuid.New().String()
	}

	// Set timestamps
	now := time.Now()
	workflow.CreatedAt = now
	workflow.UpdatedAt = now

	// Set default settings if not provided
	if workflow.Settings == nil {
		workflow.Settings = &WorkflowSettings{
			AutoStart:           false,
			RequireAllApprovals: true,
			AllowParallelSteps:  false,
			NotificationSettings: &WorkflowNotificationSettings{
				NotifyOnStart:      true,
				NotifyOnComplete:   true,
				NotifyOnStepChange: true,
				NotifyOnDeadline:   true,
				Recipients:         []string{},
			},
		}
	}

	// Initialize step IDs and set default values
	for i, step := range workflow.Steps {
		if step.ID == "" {
			step.ID = uuid.New().String()
		}
		step.Order = i + 1
		step.Status = StepStatusPending
		step.CreatedAt = now
		step.UpdatedAt = now
	}

	// Validate workflow
	if err := ce.validateWorkflow(workflow); err != nil {
		return nil, fmt.Errorf("workflow validation failed: %w", err)
	}

	// Verify team exists
	ce.mu.RLock()
	team, exists := ce.teams[workflow.TeamID]
	ce.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("team not found: %s", workflow.TeamID)
	}

	ce.mu.Lock()
	ce.workflows[workflow.ID] = workflow
	ce.mu.Unlock()

	// Log activity
	ce.logActivity(&Activity{
		Type:        ActivityTypeWorkflow,
		UserID:      workflow.CreatedBy,
		TeamID:      workflow.TeamID,
		Action:      "created",
		Resource:    &ActivityResource{Type: "workflow", ID: workflow.ID, Name: workflow.Name},
		Description: fmt.Sprintf("Created workflow '%s' in team '%s'", workflow.Name, team.Name),
		Timestamp:   now,
	})

	span.SetAttributes(
		attribute.String("workflow.id", workflow.ID),
		attribute.String("workflow.name", workflow.Name),
		attribute.String("workflow.type", string(workflow.Type)),
		attribute.String("team.id", workflow.TeamID),
		attribute.Int("workflow.steps", len(workflow.Steps)),
	)

	ce.logger.WithFields(map[string]interface{}{
		"workflow_id":   workflow.ID,
		"workflow_name": workflow.Name,
		"workflow_type": workflow.Type,
		"team_id":       workflow.TeamID,
		"steps":         len(workflow.Steps),
	}).Info("Collaborative workflow created successfully")

	return workflow, nil
}

// GetCollaborativeWorkflow retrieves a workflow by ID
func (ce *DefaultCollaborationEngine) GetCollaborativeWorkflow(workflowID string) (*CollaborativeWorkflow, error) {
	_, span := ce.tracer.Start(context.Background(), "collaboration.get_workflow")
	defer span.End()

	ce.mu.RLock()
	workflow, exists := ce.workflows[workflowID]
	ce.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("workflow not found: %s", workflowID)
	}

	span.SetAttributes(attribute.String("workflow.id", workflowID))

	return workflow, nil
}

// UpdateCollaborativeWorkflow updates an existing workflow
func (ce *DefaultCollaborationEngine) UpdateCollaborativeWorkflow(workflow *CollaborativeWorkflow) (*CollaborativeWorkflow, error) {
	_, span := ce.tracer.Start(context.Background(), "collaboration.update_workflow")
	defer span.End()

	ce.mu.Lock()
	existing, exists := ce.workflows[workflow.ID]
	if !exists {
		ce.mu.Unlock()
		return nil, fmt.Errorf("workflow not found: %s", workflow.ID)
	}

	// Preserve creation info
	workflow.CreatedBy = existing.CreatedBy
	workflow.CreatedAt = existing.CreatedAt
	workflow.UpdatedAt = time.Now()

	// Validate workflow
	if err := ce.validateWorkflow(workflow); err != nil {
		ce.mu.Unlock()
		return nil, fmt.Errorf("workflow validation failed: %w", err)
	}

	ce.workflows[workflow.ID] = workflow
	ce.mu.Unlock()

	// Log activity
	ce.logActivity(&Activity{
		Type:        ActivityTypeWorkflow,
		UserID:      "system", // In production, this would be the actual user
		TeamID:      workflow.TeamID,
		Action:      "updated",
		Resource:    &ActivityResource{Type: "workflow", ID: workflow.ID, Name: workflow.Name},
		Description: fmt.Sprintf("Updated workflow '%s'", workflow.Name),
		Timestamp:   time.Now(),
	})

	span.SetAttributes(attribute.String("workflow.id", workflow.ID))

	ce.logger.WithField("workflow_id", workflow.ID).Info("Workflow updated successfully")

	return workflow, nil
}

// ListCollaborativeWorkflows lists workflows for a team
func (ce *DefaultCollaborationEngine) ListCollaborativeWorkflows(teamID string, filter *WorkflowFilter) ([]*CollaborativeWorkflow, error) {
	_, span := ce.tracer.Start(context.Background(), "collaboration.list_workflows")
	defer span.End()

	// Verify team exists
	ce.mu.RLock()
	_, exists := ce.teams[teamID]
	if !exists {
		ce.mu.RUnlock()
		return nil, fmt.Errorf("team not found: %s", teamID)
	}

	var workflows []*CollaborativeWorkflow
	for _, workflow := range ce.workflows {
		if workflow.TeamID == teamID && ce.matchesWorkflowFilter(workflow, filter) {
			workflows = append(workflows, workflow)
		}
	}
	ce.mu.RUnlock()

	// Sort workflows by creation time (newest first)
	sort.Slice(workflows, func(i, j int) bool {
		return workflows[i].CreatedAt.After(workflows[j].CreatedAt)
	})

	// Apply pagination
	if filter != nil {
		if filter.Offset > 0 && filter.Offset < len(workflows) {
			workflows = workflows[filter.Offset:]
		}
		if filter.Limit > 0 && filter.Limit < len(workflows) {
			workflows = workflows[:filter.Limit]
		}
	}

	span.SetAttributes(
		attribute.String("team.id", teamID),
		attribute.Int("workflows.count", len(workflows)),
	)

	return workflows, nil
}

// CreateDocument creates a new collaborative document
func (ce *DefaultCollaborationEngine) CreateDocument(document *Document) (*Document, error) {
	_, span := ce.tracer.Start(context.Background(), "collaboration.create_document")
	defer span.End()

	if document.ID == "" {
		document.ID = uuid.New().String()
	}

	// Set timestamps
	now := time.Now()
	document.CreatedAt = now
	document.UpdatedAt = now
	document.UpdatedBy = document.CreatedBy
	document.Version = 1

	// Set default sharing if not provided
	if document.Sharing == nil {
		document.Sharing = &DocumentSharing{
			IsPublic:   false,
			SharedWith: []*DocumentPermission{},
		}
	}

	// Validate document
	if err := ce.validateDocument(document); err != nil {
		return nil, fmt.Errorf("document validation failed: %w", err)
	}

	ce.mu.Lock()
	ce.documents[document.ID] = document
	
	// Create initial version
	initialVersion := &DocumentVersion{
		ID:         uuid.New().String(),
		DocumentID: document.ID,
		Version:    1,
		Content:    document.Content,
		Changes:    "Initial version",
		CreatedBy:  document.CreatedBy,
		CreatedAt:  now,
	}
	ce.documentVersions[document.ID] = []*DocumentVersion{initialVersion}
	ce.mu.Unlock()

	// Log activity
	activityDescription := fmt.Sprintf("Created document '%s'", document.Title)
	if document.TeamID != "" {
		if team, exists := ce.teams[document.TeamID]; exists {
			activityDescription = fmt.Sprintf("Created document '%s' in team '%s'", document.Title, team.Name)
		}
	}

	ce.logActivity(&Activity{
		Type:        ActivityTypeDocument,
		UserID:      document.CreatedBy,
		TeamID:      document.TeamID,
		ProjectID:   document.ProjectID,
		Action:      "created",
		Resource:    &ActivityResource{Type: "document", ID: document.ID, Name: document.Title},
		Description: activityDescription,
		Timestamp:   now,
	})

	span.SetAttributes(
		attribute.String("document.id", document.ID),
		attribute.String("document.title", document.Title),
		attribute.String("document.type", string(document.Type)),
		attribute.String("document.status", string(document.Status)),
	)

	ce.logger.WithFields(map[string]interface{}{
		"document_id":    document.ID,
		"document_title": document.Title,
		"document_type":  document.Type,
		"team_id":        document.TeamID,
		"project_id":     document.ProjectID,
	}).Info("Document created successfully")

	return document, nil
}

// GetDocument retrieves a document by ID
func (ce *DefaultCollaborationEngine) GetDocument(documentID string) (*Document, error) {
	_, span := ce.tracer.Start(context.Background(), "collaboration.get_document")
	defer span.End()

	ce.mu.RLock()
	document, exists := ce.documents[documentID]
	ce.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("document not found: %s", documentID)
	}

	span.SetAttributes(attribute.String("document.id", documentID))

	return document, nil
}

// UpdateDocument updates an existing document
func (ce *DefaultCollaborationEngine) UpdateDocument(document *Document) (*Document, error) {
	_, span := ce.tracer.Start(context.Background(), "collaboration.update_document")
	defer span.End()

	ce.mu.Lock()
	existing, exists := ce.documents[document.ID]
	if !exists {
		ce.mu.Unlock()
		return nil, fmt.Errorf("document not found: %s", document.ID)
	}

	// Preserve creation info and increment version
	document.CreatedBy = existing.CreatedBy
	document.CreatedAt = existing.CreatedAt
	document.Version = existing.Version + 1
	document.UpdatedAt = time.Now()

	// Validate document
	if err := ce.validateDocument(document); err != nil {
		ce.mu.Unlock()
		return nil, fmt.Errorf("document validation failed: %w", err)
	}

	ce.documents[document.ID] = document

	// Create new version if content changed
	if existing.Content != document.Content {
		newVersion := &DocumentVersion{
			ID:         uuid.New().String(),
			DocumentID: document.ID,
			Version:    document.Version,
			Content:    document.Content,
			Changes:    "Content updated",
			CreatedBy:  document.UpdatedBy,
			CreatedAt:  document.UpdatedAt,
		}
		ce.documentVersions[document.ID] = append(ce.documentVersions[document.ID], newVersion)
	}
	ce.mu.Unlock()

	// Log activity
	ce.logActivity(&Activity{
		Type:        ActivityTypeDocument,
		UserID:      document.UpdatedBy,
		TeamID:      document.TeamID,
		ProjectID:   document.ProjectID,
		Action:      "updated",
		Resource:    &ActivityResource{Type: "document", ID: document.ID, Name: document.Title},
		Description: fmt.Sprintf("Updated document '%s' (v%d)", document.Title, document.Version),
		Timestamp:   time.Now(),
	})

	span.SetAttributes(
		attribute.String("document.id", document.ID),
		attribute.Int("document.version", document.Version),
	)

	ce.logger.WithFields(map[string]interface{}{
		"document_id": document.ID,
		"version":     document.Version,
		"updated_by":  document.UpdatedBy,
	}).Info("Document updated successfully")

	return document, nil
}

// ShareDocument shares a document with users or teams
func (ce *DefaultCollaborationEngine) ShareDocument(documentID string, sharing *DocumentSharing) error {
	_, span := ce.tracer.Start(context.Background(), "collaboration.share_document")
	defer span.End()

	ce.mu.Lock()
	document, exists := ce.documents[documentID]
	if !exists {
		ce.mu.Unlock()
		return fmt.Errorf("document not found: %s", documentID)
	}

	document.Sharing = sharing
	document.UpdatedAt = time.Now()
	ce.mu.Unlock()

	// Log activity
	ce.logActivity(&Activity{
		Type:        ActivityTypeDocument,
		UserID:      "system", // In production, this would be the actual user
		TeamID:      document.TeamID,
		ProjectID:   document.ProjectID,
		Action:      "shared",
		Resource:    &ActivityResource{Type: "document", ID: document.ID, Name: document.Title},
		Description: fmt.Sprintf("Updated sharing settings for document '%s'", document.Title),
		Timestamp:   time.Now(),
	})

	span.SetAttributes(
		attribute.String("document.id", documentID),
		attribute.Bool("document.public", sharing.IsPublic),
		attribute.Int("document.shared_with", len(sharing.SharedWith)),
	)

	ce.logger.WithFields(map[string]interface{}{
		"document_id":  documentID,
		"is_public":    sharing.IsPublic,
		"shared_count": len(sharing.SharedWith),
	}).Info("Document sharing updated successfully")

	return nil
}

// GetDocumentVersions retrieves all versions of a document
func (ce *DefaultCollaborationEngine) GetDocumentVersions(documentID string) ([]*DocumentVersion, error) {
	_, span := ce.tracer.Start(context.Background(), "collaboration.get_document_versions")
	defer span.End()

	// Verify document exists
	ce.mu.RLock()
	_, exists := ce.documents[documentID]
	if !exists {
		ce.mu.RUnlock()
		return nil, fmt.Errorf("document not found: %s", documentID)
	}

	versions := ce.documentVersions[documentID]
	ce.mu.RUnlock()

	// Sort versions by version number (newest first)
	sort.Slice(versions, func(i, j int) bool {
		return versions[i].Version > versions[j].Version
	})

	span.SetAttributes(
		attribute.String("document.id", documentID),
		attribute.Int("versions.count", len(versions)),
	)

	return versions, nil
}

// LogActivity logs an activity
func (ce *DefaultCollaborationEngine) LogActivity(activity *Activity) error {
	_, span := ce.tracer.Start(context.Background(), "collaboration.log_activity")
	defer span.End()

	if activity.ID == "" {
		activity.ID = uuid.New().String()
	}

	if activity.Timestamp.IsZero() {
		activity.Timestamp = time.Now()
	}

	ce.logActivity(activity)

	span.SetAttributes(
		attribute.String("activity.id", activity.ID),
		attribute.String("activity.type", string(activity.Type)),
		attribute.String("activity.action", activity.Action),
		attribute.String("user.id", activity.UserID),
	)

	return nil
}

// GetActivities retrieves activities with filtering
func (ce *DefaultCollaborationEngine) GetActivities(filter *ActivityFilter) ([]*Activity, error) {
	_, span := ce.tracer.Start(context.Background(), "collaboration.get_activities")
	defer span.End()

	ce.mu.RLock()
	var activities []*Activity
	for _, activity := range ce.activities {
		if ce.matchesActivityFilter(activity, filter) {
			activities = append(activities, activity)
		}
	}
	ce.mu.RUnlock()

	// Sort activities by timestamp (newest first)
	sort.Slice(activities, func(i, j int) bool {
		return activities[i].Timestamp.After(activities[j].Timestamp)
	})

	// Apply pagination
	if filter != nil {
		if filter.Offset > 0 && filter.Offset < len(activities) {
			activities = activities[filter.Offset:]
		}
		if filter.Limit > 0 && filter.Limit < len(activities) {
			activities = activities[:filter.Limit]
		}
	}

	span.SetAttributes(attribute.Int("activities.count", len(activities)))

	return activities, nil
}

// GetTeamActivity retrieves activities for a specific team
func (ce *DefaultCollaborationEngine) GetTeamActivity(teamID string, filter *ActivityFilter) ([]*Activity, error) {
	_, span := ce.tracer.Start(context.Background(), "collaboration.get_team_activity")
	defer span.End()

	// Set team filter
	if filter == nil {
		filter = &ActivityFilter{}
	}
	filter.TeamID = teamID

	activities, err := ce.GetActivities(filter)
	if err != nil {
		return nil, err
	}

	span.SetAttributes(
		attribute.String("team.id", teamID),
		attribute.Int("activities.count", len(activities)),
	)

	return activities, nil
}

// GetUserActivity retrieves activities for a specific user
func (ce *DefaultCollaborationEngine) GetUserActivity(userID string, filter *ActivityFilter) ([]*Activity, error) {
	_, span := ce.tracer.Start(context.Background(), "collaboration.get_user_activity")
	defer span.End()

	// Set user filter
	if filter == nil {
		filter = &ActivityFilter{}
	}
	filter.UserID = userID

	activities, err := ce.GetActivities(filter)
	if err != nil {
		return nil, err
	}

	span.SetAttributes(
		attribute.String("user.id", userID),
		attribute.Int("activities.count", len(activities)),
	)

	return activities, nil
}
