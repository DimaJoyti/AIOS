package desktop

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/aios/aios/pkg/models"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

// WorkspaceManager handles intelligent workspace management
type WorkspaceManager struct {
	logger     *logrus.Logger
	tracer     trace.Tracer
	config     WorkspaceConfig
	workspaces map[int]*models.Workspace
	active     int
	mu         sync.RWMutex
	running    bool
	stopCh     chan struct{}
}

// NewWorkspaceManager creates a new workspace manager
func NewWorkspaceManager(logger *logrus.Logger, config WorkspaceConfig) (*WorkspaceManager, error) {
	tracer := otel.Tracer("workspace-manager")

	return &WorkspaceManager{
		logger:     logger,
		tracer:     tracer,
		config:     config,
		workspaces: make(map[int]*models.Workspace),
		active:     1,
		stopCh:     make(chan struct{}),
	}, nil
}

// Start initializes the workspace manager
func (wm *WorkspaceManager) Start(ctx context.Context) error {
	ctx, span := wm.tracer.Start(ctx, "workspaceManager.Start")
	defer span.End()

	wm.mu.Lock()
	defer wm.mu.Unlock()

	if wm.running {
		return fmt.Errorf("workspace manager is already running")
	}

	wm.logger.Info("Starting workspace manager")

	// Create default workspaces
	wm.createDefaultWorkspaces()

	// Start monitoring
	go wm.monitorWorkspaces()

	wm.running = true
	wm.logger.Info("Workspace manager started successfully")

	return nil
}

// Stop shuts down the workspace manager
func (wm *WorkspaceManager) Stop(ctx context.Context) error {
	ctx, span := wm.tracer.Start(ctx, "workspaceManager.Stop")
	defer span.End()

	wm.mu.Lock()
	defer wm.mu.Unlock()

	if !wm.running {
		return nil
	}

	wm.logger.Info("Stopping workspace manager")

	close(wm.stopCh)
	wm.running = false
	wm.logger.Info("Workspace manager stopped")

	return nil
}

// GetStatus returns the current workspace manager status
func (wm *WorkspaceManager) GetStatus(ctx context.Context) (*models.WorkspaceStatus, error) {
	ctx, span := wm.tracer.Start(ctx, "workspaceManager.GetStatus")
	defer span.End()

	wm.mu.RLock()
	defer wm.mu.RUnlock()

	workspaces := make([]*models.Workspace, 0, len(wm.workspaces))
	for _, workspace := range wm.workspaces {
		workspaces = append(workspaces, workspace)
	}

	return &models.WorkspaceStatus{
		Running:         wm.running,
		ActiveWorkspace: wm.active,
		WorkspaceCount:  len(wm.workspaces),
		Workspaces:      workspaces,
		Timestamp:       time.Now(),
	}, nil
}

// ListWorkspaces returns all workspaces
func (wm *WorkspaceManager) ListWorkspaces(ctx context.Context) ([]*models.Workspace, error) {
	ctx, span := wm.tracer.Start(ctx, "workspaceManager.ListWorkspaces")
	defer span.End()

	wm.mu.RLock()
	defer wm.mu.RUnlock()

	workspaces := make([]*models.Workspace, 0, len(wm.workspaces))
	for _, workspace := range wm.workspaces {
		workspaces = append(workspaces, workspace)
	}

	return workspaces, nil
}

// GetWorkspace returns a specific workspace
func (wm *WorkspaceManager) GetWorkspace(ctx context.Context, id int) (*models.Workspace, error) {
	ctx, span := wm.tracer.Start(ctx, "workspaceManager.GetWorkspace")
	defer span.End()

	wm.mu.RLock()
	defer wm.mu.RUnlock()

	workspace, exists := wm.workspaces[id]
	if !exists {
		return nil, fmt.Errorf("workspace %d not found", id)
	}

	return workspace, nil
}

// SwitchWorkspace switches to a different workspace
func (wm *WorkspaceManager) SwitchWorkspace(ctx context.Context, id int) error {
	ctx, span := wm.tracer.Start(ctx, "workspaceManager.SwitchWorkspace")
	defer span.End()

	wm.logger.WithField("workspace_id", id).Info("Switching workspace")

	wm.mu.Lock()
	defer wm.mu.Unlock()

	workspace, exists := wm.workspaces[id]
	if !exists {
		return fmt.Errorf("workspace %d not found", id)
	}

	// Update active states
	if currentWorkspace, exists := wm.workspaces[wm.active]; exists {
		currentWorkspace.Active = false
	}

	workspace.Active = true
	workspace.LastUsed = time.Now()
	wm.active = id

	// TODO: Implement actual workspace switching via X11/Wayland
	wm.logger.WithFields(logrus.Fields{
		"workspace_id":   id,
		"workspace_name": workspace.Name,
	}).Info("Workspace switched")

	return nil
}

// CreateWorkspace creates a new workspace
func (wm *WorkspaceManager) CreateWorkspace(ctx context.Context, name string) (*models.Workspace, error) {
	ctx, span := wm.tracer.Start(ctx, "workspaceManager.CreateWorkspace")
	defer span.End()

	wm.logger.WithField("workspace_name", name).Info("Creating workspace")

	wm.mu.Lock()
	defer wm.mu.Unlock()

	// Find next available ID
	id := len(wm.workspaces) + 1
	for wm.workspaces[id] != nil {
		id++
	}

	workspace := &models.Workspace{
		ID:          id,
		Name:        name,
		Active:      false,
		WindowCount: 0,
		Windows:     []string{},
		Layout:      "default",
		CreatedAt:   time.Now(),
		LastUsed:    time.Now(),
	}

	wm.workspaces[id] = workspace

	wm.logger.WithFields(logrus.Fields{
		"workspace_id":   id,
		"workspace_name": name,
	}).Info("Workspace created")

	return workspace, nil
}

// DeleteWorkspace deletes a workspace
func (wm *WorkspaceManager) DeleteWorkspace(ctx context.Context, id int) error {
	ctx, span := wm.tracer.Start(ctx, "workspaceManager.DeleteWorkspace")
	defer span.End()

	wm.logger.WithField("workspace_id", id).Info("Deleting workspace")

	wm.mu.Lock()
	defer wm.mu.Unlock()

	workspace, exists := wm.workspaces[id]
	if !exists {
		return fmt.Errorf("workspace %d not found", id)
	}

	// Don't delete if it's the active workspace and there are windows
	if workspace.Active && workspace.WindowCount > 0 {
		return fmt.Errorf("cannot delete active workspace with windows")
	}

	// Don't delete if it's the last workspace
	if len(wm.workspaces) <= 1 {
		return fmt.Errorf("cannot delete the last workspace")
	}

	delete(wm.workspaces, id)

	// If we deleted the active workspace, switch to another one
	if workspace.Active {
		for newID := range wm.workspaces {
			wm.SwitchWorkspace(ctx, newID)
			break
		}
	}

	wm.logger.WithFields(logrus.Fields{
		"workspace_id":   id,
		"workspace_name": workspace.Name,
	}).Info("Workspace deleted")

	return nil
}

// RenameWorkspace renames a workspace
func (wm *WorkspaceManager) RenameWorkspace(ctx context.Context, id int, name string) error {
	ctx, span := wm.tracer.Start(ctx, "workspaceManager.RenameWorkspace")
	defer span.End()

	wm.logger.WithFields(logrus.Fields{
		"workspace_id": id,
		"new_name":     name,
	}).Info("Renaming workspace")

	wm.mu.Lock()
	defer wm.mu.Unlock()

	workspace, exists := wm.workspaces[id]
	if !exists {
		return fmt.Errorf("workspace %d not found", id)
	}

	oldName := workspace.Name
	workspace.Name = name

	wm.logger.WithFields(logrus.Fields{
		"workspace_id": id,
		"old_name":     oldName,
		"new_name":     name,
	}).Info("Workspace renamed")

	return nil
}

// AddWindowToWorkspace adds a window to a workspace
func (wm *WorkspaceManager) AddWindowToWorkspace(ctx context.Context, workspaceID int, windowID string) error {
	ctx, span := wm.tracer.Start(ctx, "workspaceManager.AddWindowToWorkspace")
	defer span.End()

	wm.mu.Lock()
	defer wm.mu.Unlock()

	workspace, exists := wm.workspaces[workspaceID]
	if !exists {
		return fmt.Errorf("workspace %d not found", workspaceID)
	}

	// Check if window is already in workspace
	for _, id := range workspace.Windows {
		if id == windowID {
			return nil // Already exists
		}
	}

	workspace.Windows = append(workspace.Windows, windowID)
	workspace.WindowCount = len(workspace.Windows)

	wm.logger.WithFields(logrus.Fields{
		"workspace_id": workspaceID,
		"window_id":    windowID,
		"window_count": workspace.WindowCount,
	}).Debug("Window added to workspace")

	return nil
}

// RemoveWindowFromWorkspace removes a window from a workspace
func (wm *WorkspaceManager) RemoveWindowFromWorkspace(ctx context.Context, workspaceID int, windowID string) error {
	ctx, span := wm.tracer.Start(ctx, "workspaceManager.RemoveWindowFromWorkspace")
	defer span.End()

	wm.mu.Lock()
	defer wm.mu.Unlock()

	workspace, exists := wm.workspaces[workspaceID]
	if !exists {
		return fmt.Errorf("workspace %d not found", workspaceID)
	}

	// Remove window from list
	for i, id := range workspace.Windows {
		if id == windowID {
			workspace.Windows = append(workspace.Windows[:i], workspace.Windows[i+1:]...)
			workspace.WindowCount = len(workspace.Windows)
			break
		}
	}

	wm.logger.WithFields(logrus.Fields{
		"workspace_id": workspaceID,
		"window_id":    windowID,
		"window_count": workspace.WindowCount,
	}).Debug("Window removed from workspace")

	return nil
}

// GetWorkspaceWindows returns all windows in a workspace
func (wm *WorkspaceManager) GetWorkspaceWindows(ctx context.Context, id int) ([]string, error) {
	ctx, span := wm.tracer.Start(ctx, "workspaceManager.GetWorkspaceWindows")
	defer span.End()

	wm.mu.RLock()
	defer wm.mu.RUnlock()

	workspace, exists := wm.workspaces[id]
	if !exists {
		return nil, fmt.Errorf("workspace %d not found", id)
	}

	return workspace.Windows, nil
}

// Helper methods

func (wm *WorkspaceManager) createDefaultWorkspaces() {
	// Create default workspaces based on configuration
	for i := 1; i <= wm.config.DefaultCount; i++ {
		name := fmt.Sprintf("Workspace %d", i)
		if wm.config.SmartNaming {
			name = wm.generateSmartName(i)
		}

		workspace := &models.Workspace{
			ID:          i,
			Name:        name,
			Active:      i == 1, // First workspace is active
			WindowCount: 0,
			Windows:     []string{},
			Layout:      "default",
			CreatedAt:   time.Now(),
			LastUsed:    time.Now(),
		}

		wm.workspaces[i] = workspace
	}

	wm.logger.WithField("workspace_count", wm.config.DefaultCount).Info("Default workspaces created")
}

func (wm *WorkspaceManager) generateSmartName(id int) string {
	// Generate smart workspace names based on typical usage patterns
	names := map[int]string{
		1: "Main",
		2: "Development",
		3: "Communication",
		4: "Media",
		5: "Research",
		6: "Tools",
		7: "Gaming",
		8: "Misc",
	}

	if name, exists := names[id]; exists {
		return name
	}

	return fmt.Sprintf("Workspace %d", id)
}

func (wm *WorkspaceManager) monitorWorkspaces() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			wm.updateWorkspaceStats()

		case <-wm.stopCh:
			wm.logger.Debug("Workspace monitoring stopped")
			return
		}
	}
}

func (wm *WorkspaceManager) updateWorkspaceStats() {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	// Update workspace statistics
	for range wm.workspaces {
		// TODO: Update actual window counts from window manager
		// For now, keep existing counts
	}
}
