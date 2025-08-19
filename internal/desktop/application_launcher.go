package desktop

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/aios/aios/internal/ai"
	"github.com/aios/aios/pkg/models"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

// ApplicationLauncher handles intelligent application launching
type ApplicationLauncher struct {
	logger         *logrus.Logger
	tracer         trace.Tracer
	aiOrchestrator *ai.Orchestrator
	applications   map[string]*models.Application
	recentApps     []*models.Application
	favoriteApps   []*models.Application
	mu             sync.RWMutex
	running        bool
	stopCh         chan struct{}
}

// NewApplicationLauncher creates a new application launcher
func NewApplicationLauncher(logger *logrus.Logger, aiOrchestrator *ai.Orchestrator) (*ApplicationLauncher, error) {
	tracer := otel.Tracer("application-launcher")

	return &ApplicationLauncher{
		logger:         logger,
		tracer:         tracer,
		aiOrchestrator: aiOrchestrator,
		applications:   make(map[string]*models.Application),
		recentApps:     []*models.Application{},
		favoriteApps:   []*models.Application{},
		stopCh:         make(chan struct{}),
	}, nil
}

// Start initializes the application launcher
func (al *ApplicationLauncher) Start(ctx context.Context) error {
	ctx, span := al.tracer.Start(ctx, "applicationLauncher.Start")
	defer span.End()

	al.mu.Lock()
	defer al.mu.Unlock()

	if al.running {
		return fmt.Errorf("application launcher is already running")
	}

	al.logger.Info("Starting application launcher")

	// Discover applications
	if err := al.discoverApplications(); err != nil {
		return fmt.Errorf("failed to discover applications: %w", err)
	}

	// Start monitoring
	go al.monitorApplications()

	al.running = true
	al.logger.Info("Application launcher started successfully")

	return nil
}

// Stop shuts down the application launcher
func (al *ApplicationLauncher) Stop(ctx context.Context) error {
	ctx, span := al.tracer.Start(ctx, "applicationLauncher.Stop")
	defer span.End()

	al.mu.Lock()
	defer al.mu.Unlock()

	if !al.running {
		return nil
	}

	al.logger.Info("Stopping application launcher")

	close(al.stopCh)
	al.running = false
	al.logger.Info("Application launcher stopped")

	return nil
}

// GetStatus returns the current application launcher status
func (al *ApplicationLauncher) GetStatus(ctx context.Context) (*models.ApplicationStatus, error) {
	ctx, span := al.tracer.Start(ctx, "applicationLauncher.GetStatus")
	defer span.End()

	al.mu.RLock()
	defer al.mu.RUnlock()

	applications := make([]*models.Application, 0, len(al.applications))
	for _, app := range al.applications {
		applications = append(applications, app)
	}

	return &models.ApplicationStatus{
		Running:          al.running,
		ApplicationCount: len(al.applications),
		Applications:     applications,
		RecentApps:       al.recentApps,
		FavoriteApps:     al.favoriteApps,
		Timestamp:        time.Now(),
	}, nil
}

// ListApplications returns all available applications
func (al *ApplicationLauncher) ListApplications(ctx context.Context) ([]*models.Application, error) {
	ctx, span := al.tracer.Start(ctx, "applicationLauncher.ListApplications")
	defer span.End()

	al.mu.RLock()
	defer al.mu.RUnlock()

	applications := make([]*models.Application, 0, len(al.applications))
	for _, app := range al.applications {
		applications = append(applications, app)
	}

	return applications, nil
}

// GetApplication returns a specific application
func (al *ApplicationLauncher) GetApplication(ctx context.Context, id string) (*models.Application, error) {
	ctx, span := al.tracer.Start(ctx, "applicationLauncher.GetApplication")
	defer span.End()

	al.mu.RLock()
	defer al.mu.RUnlock()

	app, exists := al.applications[id]
	if !exists {
		return nil, fmt.Errorf("application %s not found", id)
	}

	return app, nil
}

// LaunchApplication launches an application
func (al *ApplicationLauncher) LaunchApplication(ctx context.Context, id string) error {
	ctx, span := al.tracer.Start(ctx, "applicationLauncher.LaunchApplication")
	defer span.End()

	al.logger.WithField("app_id", id).Info("Launching application")

	al.mu.Lock()
	defer al.mu.Unlock()

	app, exists := al.applications[id]
	if !exists {
		return fmt.Errorf("application %s not found", id)
	}

	// Launch the application
	cmd := exec.CommandContext(ctx, app.Executable)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to launch application %s: %w", app.Name, err)
	}

	// Update application state
	app.Running = true
	app.LaunchCount++
	app.LastLaunched = time.Now()

	// Add to recent apps
	al.addToRecentApps(app)

	al.logger.WithFields(logrus.Fields{
		"app_id":       id,
		"app_name":     app.Name,
		"launch_count": app.LaunchCount,
	}).Info("Application launched")

	return nil
}

// SearchApplications searches for applications using AI
func (al *ApplicationLauncher) SearchApplications(ctx context.Context, query string) ([]*models.Application, error) {
	ctx, span := al.tracer.Start(ctx, "applicationLauncher.SearchApplications")
	defer span.End()

	al.logger.WithField("query", query).Info("Searching applications")

	al.mu.RLock()
	defer al.mu.RUnlock()

	// Simple text-based search for now
	results := []*models.Application{}
	queryLower := strings.ToLower(query)

	for _, app := range al.applications {
		if al.matchesQuery(app, queryLower) {
			results = append(results, app)
		}
	}

	// TODO: Use AI for semantic search and ranking
	if al.aiOrchestrator != nil {
		results = al.rankSearchResults(ctx, query, results)
	}

	al.logger.WithFields(logrus.Fields{
		"query":        query,
		"result_count": len(results),
	}).Info("Application search completed")

	return results, nil
}

// GetRecentApplications returns recently used applications
func (al *ApplicationLauncher) GetRecentApplications(ctx context.Context) ([]*models.Application, error) {
	ctx, span := al.tracer.Start(ctx, "applicationLauncher.GetRecentApplications")
	defer span.End()

	al.mu.RLock()
	defer al.mu.RUnlock()

	return al.recentApps, nil
}

// GetFavoriteApplications returns favorite applications
func (al *ApplicationLauncher) GetFavoriteApplications(ctx context.Context) ([]*models.Application, error) {
	ctx, span := al.tracer.Start(ctx, "applicationLauncher.GetFavoriteApplications")
	defer span.End()

	al.mu.RLock()
	defer al.mu.RUnlock()

	return al.favoriteApps, nil
}

// AddToFavorites adds an application to favorites
func (al *ApplicationLauncher) AddToFavorites(ctx context.Context, id string) error {
	ctx, span := al.tracer.Start(ctx, "applicationLauncher.AddToFavorites")
	defer span.End()

	al.mu.Lock()
	defer al.mu.Unlock()

	app, exists := al.applications[id]
	if !exists {
		return fmt.Errorf("application %s not found", id)
	}

	// Check if already in favorites
	for _, fav := range al.favoriteApps {
		if fav.ID == id {
			return nil // Already in favorites
		}
	}

	al.favoriteApps = append(al.favoriteApps, app)

	al.logger.WithFields(logrus.Fields{
		"app_id":   id,
		"app_name": app.Name,
	}).Info("Application added to favorites")

	return nil
}

// RemoveFromFavorites removes an application from favorites
func (al *ApplicationLauncher) RemoveFromFavorites(ctx context.Context, id string) error {
	ctx, span := al.tracer.Start(ctx, "applicationLauncher.RemoveFromFavorites")
	defer span.End()

	al.mu.Lock()
	defer al.mu.Unlock()

	for i, fav := range al.favoriteApps {
		if fav.ID == id {
			al.favoriteApps = append(al.favoriteApps[:i], al.favoriteApps[i+1:]...)

			al.logger.WithFields(logrus.Fields{
				"app_id":   id,
				"app_name": fav.Name,
			}).Info("Application removed from favorites")

			return nil
		}
	}

	return fmt.Errorf("application %s not found in favorites", id)
}

// Helper methods

func (al *ApplicationLauncher) discoverApplications() error {
	// Create mock applications for demonstration
	al.applications = map[string]*models.Application{
		"firefox": {
			ID:           "firefox",
			Name:         "firefox",
			DisplayName:  "Firefox",
			Description:  "Web browser",
			Icon:         "/usr/share/icons/firefox.png",
			Category:     "Network",
			Executable:   "firefox",
			Keywords:     []string{"web", "browser", "internet"},
			MimeTypes:    []string{"text/html", "application/xhtml+xml"},
			Running:      false,
			Windows:      []string{},
			LaunchCount:  0,
			LastLaunched: time.Time{},
		},
		"terminal": {
			ID:           "terminal",
			Name:         "gnome-terminal",
			DisplayName:  "Terminal",
			Description:  "Terminal emulator",
			Icon:         "/usr/share/icons/terminal.png",
			Category:     "System",
			Executable:   "gnome-terminal",
			Keywords:     []string{"terminal", "shell", "command"},
			MimeTypes:    []string{},
			Running:      false,
			Windows:      []string{},
			LaunchCount:  0,
			LastLaunched: time.Time{},
		},
		"code": {
			ID:           "code",
			Name:         "code",
			DisplayName:  "Visual Studio Code",
			Description:  "Code editor",
			Icon:         "/usr/share/icons/vscode.png",
			Category:     "Development",
			Executable:   "code",
			Keywords:     []string{"editor", "code", "development", "programming"},
			MimeTypes:    []string{"text/plain", "application/json"},
			Running:      false,
			Windows:      []string{},
			LaunchCount:  0,
			LastLaunched: time.Time{},
		},
	}

	// Set default favorites
	al.favoriteApps = []*models.Application{
		al.applications["firefox"],
		al.applications["terminal"],
		al.applications["code"],
	}

	al.logger.WithField("app_count", len(al.applications)).Info("Applications discovered")
	return nil
}

func (al *ApplicationLauncher) matchesQuery(app *models.Application, query string) bool {
	// Check name, display name, description, and keywords
	if strings.Contains(strings.ToLower(app.Name), query) ||
		strings.Contains(strings.ToLower(app.DisplayName), query) ||
		strings.Contains(strings.ToLower(app.Description), query) {
		return true
	}

	// Check keywords
	for _, keyword := range app.Keywords {
		if strings.Contains(strings.ToLower(keyword), query) {
			return true
		}
	}

	return false
}

func (al *ApplicationLauncher) rankSearchResults(ctx context.Context, query string, results []*models.Application) []*models.Application {
	// TODO: Use AI to rank search results based on:
	// - User preferences and usage patterns
	// - Semantic similarity to query
	// - Context awareness (time of day, current workspace, etc.)

	// For now, sort by launch count
	for i := 0; i < len(results)-1; i++ {
		for j := i + 1; j < len(results); j++ {
			if results[i].LaunchCount < results[j].LaunchCount {
				results[i], results[j] = results[j], results[i]
			}
		}
	}

	return results
}

func (al *ApplicationLauncher) addToRecentApps(app *models.Application) {
	// Remove if already in recent apps
	for i, recent := range al.recentApps {
		if recent.ID == app.ID {
			al.recentApps = append(al.recentApps[:i], al.recentApps[i+1:]...)
			break
		}
	}

	// Add to front
	al.recentApps = append([]*models.Application{app}, al.recentApps...)

	// Keep only last 10 recent apps
	if len(al.recentApps) > 10 {
		al.recentApps = al.recentApps[:10]
	}
}

func (al *ApplicationLauncher) monitorApplications() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			al.updateApplicationStatus()

		case <-al.stopCh:
			al.logger.Debug("Application monitoring stopped")
			return
		}
	}
}

func (al *ApplicationLauncher) updateApplicationStatus() {
	al.mu.Lock()
	defer al.mu.Unlock()

	// TODO: Check actual application status via process monitoring
	// For now, keep existing status
}
