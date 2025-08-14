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

// ThemeManager handles desktop theme management
type ThemeManager struct {
	logger      *logrus.Logger
	tracer      trace.Tracer
	themes      map[string]*models.Theme
	activeTheme string
	mu          sync.RWMutex
	running     bool
	stopCh      chan struct{}
}

// NewThemeManager creates a new theme manager
func NewThemeManager(logger *logrus.Logger, defaultTheme string) (*ThemeManager, error) {
	tracer := otel.Tracer("theme-manager")

	return &ThemeManager{
		logger:      logger,
		tracer:      tracer,
		themes:      make(map[string]*models.Theme),
		activeTheme: defaultTheme,
		stopCh:      make(chan struct{}),
	}, nil
}

// Start initializes the theme manager
func (tm *ThemeManager) Start(ctx context.Context) error {
	ctx, span := tm.tracer.Start(ctx, "themeManager.Start")
	defer span.End()

	tm.mu.Lock()
	defer tm.mu.Unlock()

	if tm.running {
		return fmt.Errorf("theme manager is already running")
	}

	tm.logger.Info("Starting theme manager")

	// Load default themes
	tm.loadDefaultThemes()

	// Apply active theme
	if err := tm.applyTheme(tm.activeTheme); err != nil {
		tm.logger.WithError(err).Warn("Failed to apply default theme")
	}

	tm.running = true
	tm.logger.Info("Theme manager started successfully")

	return nil
}

// Stop shuts down the theme manager
func (tm *ThemeManager) Stop(ctx context.Context) error {
	ctx, span := tm.tracer.Start(ctx, "themeManager.Stop")
	defer span.End()

	tm.mu.Lock()
	defer tm.mu.Unlock()

	if !tm.running {
		return nil
	}

	tm.logger.Info("Stopping theme manager")

	close(tm.stopCh)
	tm.running = false
	tm.logger.Info("Theme manager stopped")

	return nil
}

// GetStatus returns the current theme manager status
func (tm *ThemeManager) GetStatus(ctx context.Context) (*models.ThemeStatus, error) {
	ctx, span := tm.tracer.Start(ctx, "themeManager.GetStatus")
	defer span.End()

	tm.mu.RLock()
	defer tm.mu.RUnlock()

	themes := make([]*models.Theme, 0, len(tm.themes))
	for _, theme := range tm.themes {
		themes = append(themes, theme)
	}

	return &models.ThemeStatus{
		Running:     tm.running,
		ActiveTheme: tm.activeTheme,
		ThemeCount:  len(tm.themes),
		Themes:      themes,
		Timestamp:   time.Now(),
	}, nil
}

// ListThemes returns all available themes
func (tm *ThemeManager) ListThemes(ctx context.Context) ([]*models.Theme, error) {
	ctx, span := tm.tracer.Start(ctx, "themeManager.ListThemes")
	defer span.End()

	tm.mu.RLock()
	defer tm.mu.RUnlock()

	themes := make([]*models.Theme, 0, len(tm.themes))
	for _, theme := range tm.themes {
		themes = append(themes, theme)
	}

	return themes, nil
}

// GetTheme returns a specific theme
func (tm *ThemeManager) GetTheme(ctx context.Context, id string) (*models.Theme, error) {
	ctx, span := tm.tracer.Start(ctx, "themeManager.GetTheme")
	defer span.End()

	tm.mu.RLock()
	defer tm.mu.RUnlock()

	theme, exists := tm.themes[id]
	if !exists {
		return nil, fmt.Errorf("theme %s not found", id)
	}

	return theme, nil
}

// SetTheme applies a theme
func (tm *ThemeManager) SetTheme(ctx context.Context, id string) error {
	ctx, span := tm.tracer.Start(ctx, "themeManager.SetTheme")
	defer span.End()

	tm.logger.WithField("theme_id", id).Info("Setting theme")

	tm.mu.Lock()
	defer tm.mu.Unlock()

	theme, exists := tm.themes[id]
	if !exists {
		return fmt.Errorf("theme %s not found", id)
	}

	// Deactivate current theme
	if currentTheme, exists := tm.themes[tm.activeTheme]; exists {
		currentTheme.Active = false
	}

	// Apply new theme
	if err := tm.applyTheme(id); err != nil {
		return fmt.Errorf("failed to apply theme %s: %w", id, err)
	}

	// Update state
	theme.Active = true
	tm.activeTheme = id

	tm.logger.WithFields(logrus.Fields{
		"theme_id":   id,
		"theme_name": theme.Name,
	}).Info("Theme applied")

	return nil
}

// GetActiveTheme returns the currently active theme
func (tm *ThemeManager) GetActiveTheme(ctx context.Context) (*models.Theme, error) {
	ctx, span := tm.tracer.Start(ctx, "themeManager.GetActiveTheme")
	defer span.End()

	tm.mu.RLock()
	defer tm.mu.RUnlock()

	theme, exists := tm.themes[tm.activeTheme]
	if !exists {
		return nil, fmt.Errorf("active theme %s not found", tm.activeTheme)
	}

	return theme, nil
}

// CreateTheme creates a new custom theme
func (tm *ThemeManager) CreateTheme(ctx context.Context, theme *models.Theme) error {
	ctx, span := tm.tracer.Start(ctx, "themeManager.CreateTheme")
	defer span.End()

	tm.logger.WithField("theme_name", theme.Name).Info("Creating theme")

	tm.mu.Lock()
	defer tm.mu.Unlock()

	// Check if theme already exists
	if _, exists := tm.themes[theme.ID]; exists {
		return fmt.Errorf("theme %s already exists", theme.ID)
	}

	// Set creation time
	theme.CreatedAt = time.Now()
	theme.Active = false

	tm.themes[theme.ID] = theme

	tm.logger.WithFields(logrus.Fields{
		"theme_id":   theme.ID,
		"theme_name": theme.Name,
	}).Info("Theme created")

	return nil
}

// DeleteTheme deletes a custom theme
func (tm *ThemeManager) DeleteTheme(ctx context.Context, id string) error {
	ctx, span := tm.tracer.Start(ctx, "themeManager.DeleteTheme")
	defer span.End()

	tm.logger.WithField("theme_id", id).Info("Deleting theme")

	tm.mu.Lock()
	defer tm.mu.Unlock()

	theme, exists := tm.themes[id]
	if !exists {
		return fmt.Errorf("theme %s not found", id)
	}

	// Don't delete if it's the active theme
	if theme.Active {
		return fmt.Errorf("cannot delete active theme")
	}

	delete(tm.themes, id)

	tm.logger.WithFields(logrus.Fields{
		"theme_id":   id,
		"theme_name": theme.Name,
	}).Info("Theme deleted")

	return nil
}

// Helper methods

func (tm *ThemeManager) loadDefaultThemes() {
	// Load default themes
	tm.themes = map[string]*models.Theme{
		"dark": {
			ID:          "dark",
			Name:        "Dark Theme",
			Description: "Modern dark theme with blue accents",
			Author:      "AIOS Team",
			Version:     "1.0.0",
			Colors: map[string]string{
				"primary":    "#1e293b",
				"secondary":  "#334155",
				"accent":     "#3b82f6",
				"background": "#0f172a",
				"surface":    "#1e293b",
				"text":       "#f8fafc",
				"text-muted": "#94a3b8",
			},
			Fonts: map[string]string{
				"primary":   "Inter",
				"secondary": "JetBrains Mono",
				"size":      "14px",
			},
			Icons:     "lucide",
			Wallpaper: "/usr/share/backgrounds/aios-dark.jpg",
			Active:    tm.activeTheme == "dark",
			CreatedAt: time.Now(),
		},
		"light": {
			ID:          "light",
			Name:        "Light Theme",
			Description: "Clean light theme with subtle shadows",
			Author:      "AIOS Team",
			Version:     "1.0.0",
			Colors: map[string]string{
				"primary":    "#ffffff",
				"secondary":  "#f8fafc",
				"accent":     "#3b82f6",
				"background": "#ffffff",
				"surface":    "#f8fafc",
				"text":       "#1e293b",
				"text-muted": "#64748b",
			},
			Fonts: map[string]string{
				"primary":   "Inter",
				"secondary": "JetBrains Mono",
				"size":      "14px",
			},
			Icons:     "lucide",
			Wallpaper: "/usr/share/backgrounds/aios-light.jpg",
			Active:    tm.activeTheme == "light",
			CreatedAt: time.Now(),
		},
		"aios": {
			ID:          "aios",
			Name:        "AIOS Theme",
			Description: "Official AIOS theme with gradient accents",
			Author:      "AIOS Team",
			Version:     "1.0.0",
			Colors: map[string]string{
				"primary":    "#667eea",
				"secondary":  "#764ba2",
				"accent":     "#f093fb",
				"background": "#1a1a2e",
				"surface":    "#16213e",
				"text":       "#ffffff",
				"text-muted": "#a0aec0",
			},
			Fonts: map[string]string{
				"primary":   "Inter",
				"secondary": "JetBrains Mono",
				"size":      "14px",
			},
			Icons:     "lucide",
			Wallpaper: "/usr/share/backgrounds/aios-gradient.jpg",
			Active:    tm.activeTheme == "aios",
			CreatedAt: time.Now(),
		},
	}

	tm.logger.WithField("theme_count", len(tm.themes)).Info("Default themes loaded")
}

func (tm *ThemeManager) applyTheme(id string) error {
	theme, exists := tm.themes[id]
	if !exists {
		return fmt.Errorf("theme %s not found", id)
	}

	// TODO: Apply theme to actual desktop environment
	// This would involve:
	// - Updating window manager colors and styles
	// - Changing wallpaper
	// - Updating application themes
	// - Modifying system-wide color schemes

	tm.logger.WithFields(logrus.Fields{
		"theme_id":   id,
		"theme_name": theme.Name,
	}).Debug("Theme applied")

	return nil
}
