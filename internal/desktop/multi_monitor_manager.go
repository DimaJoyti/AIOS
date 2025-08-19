package desktop

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/aios/aios/internal/ai"
	"github.com/aios/aios/pkg/models"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

// MultiMonitorManager handles multiple display management
type MultiMonitorManager struct {
	logger *logrus.Logger
	tracer trace.Tracer
	config MultiMonitorConfig
	mu     sync.RWMutex

	// Monitor management
	monitors       map[string]*Monitor
	primaryMonitor string
	monitorLayout  MonitorLayout

	// AI integration
	aiOrchestrator *ai.Orchestrator

	// State tracking
	windowDistribution map[string][]string // monitor_id -> window_ids
	monitorHistory     []MonitorEvent
	lastUpdate         time.Time

	// Callbacks
	onMonitorAdded   func(*Monitor)
	onMonitorRemoved func(string)
	onLayoutChanged  func(MonitorLayout)
}

// MultiMonitorConfig defines multi-monitor configuration
type MultiMonitorConfig struct {
	AutoDetect         bool          `json:"auto_detect"`
	PrimaryMonitorID   string        `json:"primary_monitor_id"`
	DefaultLayout      string        `json:"default_layout"`
	WindowDistribution string        `json:"window_distribution"` // "balanced", "primary_focused", "ai_optimized"
	CrossMonitorSnap   bool          `json:"cross_monitor_snap"`
	UnifiedWorkspaces  bool          `json:"unified_workspaces"`
	MonitorSwitchDelay time.Duration `json:"monitor_switch_delay"`
	AIOptimization     bool          `json:"ai_optimization"`
}

// Monitor represents a physical display
type Monitor struct {
	ID           string              `json:"id"`
	Name         string              `json:"name"`
	Resolution   models.Size         `json:"resolution"`
	Position     models.Position     `json:"position"`
	Scale        float64             `json:"scale"`
	Rotation     int                 `json:"rotation"` // 0, 90, 180, 270
	IsPrimary    bool                `json:"is_primary"`
	IsConnected  bool                `json:"is_connected"`
	RefreshRate  int                 `json:"refresh_rate"`
	ColorProfile string              `json:"color_profile"`
	Workspaces   []int               `json:"workspaces"`
	Properties   map[string]string   `json:"properties"`
	Capabilities MonitorCapabilities `json:"capabilities"`
	LastSeen     time.Time           `json:"last_seen"`
}

// MonitorCapabilities defines what a monitor supports
type MonitorCapabilities struct {
	MaxResolution   models.Size `json:"max_resolution"`
	SupportedScales []float64   `json:"supported_scales"`
	SupportsHDR     bool        `json:"supports_hdr"`
	SupportsVRR     bool        `json:"supports_vrr"` // Variable Refresh Rate
	ColorDepth      int         `json:"color_depth"`
	Brightness      int         `json:"brightness"`
	Contrast        int         `json:"contrast"`
}

// MonitorLayout defines the arrangement of monitors
type MonitorLayout struct {
	Type        string                     `json:"type"` // "horizontal", "vertical", "custom"
	Arrangement map[string]models.Position `json:"arrangement"`
	TotalArea   models.Rectangle           `json:"total_area"`
	CreatedAt   time.Time                  `json:"created_at"`
}

// MonitorEvent represents a monitor-related event
type MonitorEvent struct {
	Type      string                 `json:"type"` // "added", "removed", "changed", "layout_updated"
	MonitorID string                 `json:"monitor_id"`
	Timestamp time.Time              `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
}

// WindowDistributionStrategy defines how windows are distributed across monitors
type WindowDistributionStrategy interface {
	DistributeWindows(windows []*models.Window, monitors map[string]*Monitor) map[string][]*models.Window
	GetName() string
	GetDescription() string
}

// NewMultiMonitorManager creates a new multi-monitor manager
func NewMultiMonitorManager(logger *logrus.Logger, config MultiMonitorConfig, aiOrchestrator *ai.Orchestrator) *MultiMonitorManager {
	tracer := otel.Tracer("multi-monitor-manager")

	manager := &MultiMonitorManager{
		logger:             logger,
		tracer:             tracer,
		config:             config,
		monitors:           make(map[string]*Monitor),
		windowDistribution: make(map[string][]string),
		monitorHistory:     make([]MonitorEvent, 0),
		aiOrchestrator:     aiOrchestrator,
		lastUpdate:         time.Now(),
	}

	// Initialize with auto-detection if enabled
	if config.AutoDetect {
		go manager.startMonitorDetection()
	}

	return manager
}

// DetectMonitors detects connected monitors
func (mmm *MultiMonitorManager) DetectMonitors(ctx context.Context) error {
	ctx, span := mmm.tracer.Start(ctx, "multiMonitorManager.DetectMonitors")
	defer span.End()

	mmm.mu.Lock()
	defer mmm.mu.Unlock()

	// Platform-specific monitor detection would go here
	// For now, simulate detection
	detectedMonitors := mmm.simulateMonitorDetection()

	// Update monitor list
	for _, monitor := range detectedMonitors {
		if existing, exists := mmm.monitors[monitor.ID]; exists {
			// Update existing monitor
			mmm.updateMonitor(existing, monitor)
		} else {
			// Add new monitor
			mmm.addMonitor(monitor)
		}
	}

	// Remove disconnected monitors
	mmm.removeDisconnectedMonitors(detectedMonitors)

	// Update layout
	mmm.updateMonitorLayout()

	mmm.lastUpdate = time.Now()

	return nil
}

// simulateMonitorDetection simulates monitor detection for development
func (mmm *MultiMonitorManager) simulateMonitorDetection() []*Monitor {
	monitors := []*Monitor{
		{
			ID:          "monitor-1",
			Name:        "Primary Display",
			Resolution:  models.Size{Width: 1920, Height: 1080},
			Position:    models.Position{X: 0, Y: 0},
			Scale:       1.0,
			IsPrimary:   true,
			IsConnected: true,
			RefreshRate: 60,
			Workspaces:  []int{1, 2, 3, 4},
			Capabilities: MonitorCapabilities{
				MaxResolution:   models.Size{Width: 1920, Height: 1080},
				SupportedScales: []float64{1.0, 1.25, 1.5},
				SupportsHDR:     false,
				ColorDepth:      24,
			},
			LastSeen: time.Now(),
		},
	}

	// Add second monitor if configured
	if len(mmm.monitors) > 1 || mmm.config.DefaultLayout != "single" {
		monitors = append(monitors, &Monitor{
			ID:          "monitor-2",
			Name:        "Secondary Display",
			Resolution:  models.Size{Width: 1920, Height: 1080},
			Position:    models.Position{X: 1920, Y: 0},
			Scale:       1.0,
			IsPrimary:   false,
			IsConnected: true,
			RefreshRate: 60,
			Workspaces:  []int{5, 6, 7, 8},
			Capabilities: MonitorCapabilities{
				MaxResolution:   models.Size{Width: 1920, Height: 1080},
				SupportedScales: []float64{1.0, 1.25, 1.5},
				SupportsHDR:     false,
				ColorDepth:      24,
			},
			LastSeen: time.Now(),
		})
	}

	return monitors
}

// addMonitor adds a new monitor
func (mmm *MultiMonitorManager) addMonitor(monitor *Monitor) {
	mmm.monitors[monitor.ID] = monitor
	mmm.windowDistribution[monitor.ID] = make([]string, 0)

	// Set as primary if it's the first monitor or explicitly marked
	if len(mmm.monitors) == 1 || monitor.IsPrimary {
		mmm.primaryMonitor = monitor.ID
	}

	// Record event
	mmm.recordMonitorEvent(MonitorEvent{
		Type:      "added",
		MonitorID: monitor.ID,
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"name":       monitor.Name,
			"resolution": monitor.Resolution,
			"position":   monitor.Position,
		},
	})

	// Notify callback
	if mmm.onMonitorAdded != nil {
		mmm.onMonitorAdded(monitor)
	}

	mmm.logger.WithFields(logrus.Fields{
		"monitor_id":   monitor.ID,
		"monitor_name": monitor.Name,
		"resolution":   fmt.Sprintf("%dx%d", monitor.Resolution.Width, monitor.Resolution.Height),
	}).Info("Monitor added")
}

// updateMonitor updates an existing monitor
func (mmm *MultiMonitorManager) updateMonitor(existing, updated *Monitor) {
	// Update properties
	existing.Name = updated.Name
	existing.Resolution = updated.Resolution
	existing.Position = updated.Position
	existing.Scale = updated.Scale
	existing.Rotation = updated.Rotation
	existing.IsConnected = updated.IsConnected
	existing.RefreshRate = updated.RefreshRate
	existing.LastSeen = time.Now()

	// Record event if significant changes
	if existing.Resolution != updated.Resolution || existing.Position != updated.Position {
		mmm.recordMonitorEvent(MonitorEvent{
			Type:      "changed",
			MonitorID: existing.ID,
			Timestamp: time.Now(),
			Data: map[string]interface{}{
				"old_resolution": existing.Resolution,
				"new_resolution": updated.Resolution,
				"old_position":   existing.Position,
				"new_position":   updated.Position,
			},
		})
	}
}

// removeDisconnectedMonitors removes monitors that are no longer connected
func (mmm *MultiMonitorManager) removeDisconnectedMonitors(detectedMonitors []*Monitor) {
	detectedIDs := make(map[string]bool)
	for _, monitor := range detectedMonitors {
		detectedIDs[monitor.ID] = true
	}

	for id := range mmm.monitors {
		if !detectedIDs[id] {
			mmm.removeMonitor(id)
		}
	}
}

// removeMonitor removes a monitor
func (mmm *MultiMonitorManager) removeMonitor(monitorID string) {
	monitor, exists := mmm.monitors[monitorID]
	if !exists {
		return
	}

	// Move windows from removed monitor to primary
	if windows, hasWindows := mmm.windowDistribution[monitorID]; hasWindows && len(windows) > 0 {
		if mmm.primaryMonitor != "" && mmm.primaryMonitor != monitorID {
			mmm.windowDistribution[mmm.primaryMonitor] = append(mmm.windowDistribution[mmm.primaryMonitor], windows...)
		}
	}

	delete(mmm.monitors, monitorID)
	delete(mmm.windowDistribution, monitorID)

	// Update primary monitor if needed
	if mmm.primaryMonitor == monitorID {
		mmm.selectNewPrimaryMonitor()
	}

	// Record event
	mmm.recordMonitorEvent(MonitorEvent{
		Type:      "removed",
		MonitorID: monitorID,
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"name": monitor.Name,
		},
	})

	// Notify callback
	if mmm.onMonitorRemoved != nil {
		mmm.onMonitorRemoved(monitorID)
	}

	mmm.logger.WithField("monitor_id", monitorID).Info("Monitor removed")
}

// selectNewPrimaryMonitor selects a new primary monitor
func (mmm *MultiMonitorManager) selectNewPrimaryMonitor() {
	for id, monitor := range mmm.monitors {
		if monitor.IsConnected {
			mmm.primaryMonitor = id
			monitor.IsPrimary = true
			break
		}
	}
}

// updateMonitorLayout updates the monitor layout
func (mmm *MultiMonitorManager) updateMonitorLayout() {
	if len(mmm.monitors) == 0 {
		return
	}

	arrangement := make(map[string]models.Position)
	var minX, minY, maxX, maxY int

	for id, monitor := range mmm.monitors {
		if !monitor.IsConnected {
			continue
		}

		arrangement[id] = monitor.Position

		// Calculate total area
		if monitor.Position.X < minX {
			minX = monitor.Position.X
		}
		if monitor.Position.Y < minY {
			minY = monitor.Position.Y
		}

		maxX = max(maxX, monitor.Position.X+monitor.Resolution.Width)
		maxY = max(maxY, monitor.Position.Y+monitor.Resolution.Height)
	}

	mmm.monitorLayout = MonitorLayout{
		Type:        mmm.detectLayoutType(),
		Arrangement: arrangement,
		TotalArea: models.Rectangle{
			X:      minX,
			Y:      minY,
			Width:  maxX - minX,
			Height: maxY - minY,
		},
		CreatedAt: time.Now(),
	}

	// Record layout change event
	mmm.recordMonitorEvent(MonitorEvent{
		Type:      "layout_updated",
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"layout_type": mmm.monitorLayout.Type,
			"total_area":  mmm.monitorLayout.TotalArea,
		},
	})

	// Notify callback
	if mmm.onLayoutChanged != nil {
		mmm.onLayoutChanged(mmm.monitorLayout)
	}
}

// detectLayoutType detects the type of monitor layout
func (mmm *MultiMonitorManager) detectLayoutType() string {
	if len(mmm.monitors) <= 1 {
		return "single"
	}

	// Simple detection logic
	// In a full implementation, this would be more sophisticated
	return "horizontal"
}

// DistributeWindows distributes windows across monitors
func (mmm *MultiMonitorManager) DistributeWindows(ctx context.Context, windows []*models.Window) error {
	ctx, span := mmm.tracer.Start(ctx, "multiMonitorManager.DistributeWindows")
	defer span.End()

	mmm.mu.Lock()
	defer mmm.mu.Unlock()

	// Clear current distribution
	for monitorID := range mmm.windowDistribution {
		mmm.windowDistribution[monitorID] = make([]string, 0)
	}

	// Get distribution strategy
	strategy := mmm.getDistributionStrategy()

	// Distribute windows
	distribution := strategy.DistributeWindows(windows, mmm.monitors)

	// Update internal state
	for monitorID, monitorWindows := range distribution {
		windowIDs := make([]string, len(monitorWindows))
		for i, window := range monitorWindows {
			windowIDs[i] = window.ID
		}
		mmm.windowDistribution[monitorID] = windowIDs
	}

	return nil
}

// getDistributionStrategy returns the appropriate distribution strategy
func (mmm *MultiMonitorManager) getDistributionStrategy() WindowDistributionStrategy {
	switch mmm.config.WindowDistribution {
	case "balanced":
		return &BalancedDistributionStrategy{}
	case "primary_focused":
		return &PrimaryFocusedDistributionStrategy{primaryMonitor: mmm.primaryMonitor}
	case "ai_optimized":
		return &AIOptimizedDistributionStrategy{aiOrchestrator: mmm.aiOrchestrator}
	default:
		return &BalancedDistributionStrategy{}
	}
}

// startMonitorDetection starts automatic monitor detection
func (mmm *MultiMonitorManager) startMonitorDetection() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			if err := mmm.DetectMonitors(ctx); err != nil {
				mmm.logger.WithError(err).Error("Failed to detect monitors")
			}
			cancel()
		}
	}
}

// Helper methods

func (mmm *MultiMonitorManager) recordMonitorEvent(event MonitorEvent) {
	mmm.monitorHistory = append(mmm.monitorHistory, event)

	// Keep only recent history
	if len(mmm.monitorHistory) > 1000 {
		mmm.monitorHistory = mmm.monitorHistory[100:]
	}
}

func (mmm *MultiMonitorManager) GetMonitors() map[string]*Monitor {
	mmm.mu.RLock()
	defer mmm.mu.RUnlock()

	result := make(map[string]*Monitor)
	for id, monitor := range mmm.monitors {
		result[id] = monitor
	}
	return result
}

func (mmm *MultiMonitorManager) GetPrimaryMonitor() *Monitor {
	mmm.mu.RLock()
	defer mmm.mu.RUnlock()

	if mmm.primaryMonitor != "" {
		return mmm.monitors[mmm.primaryMonitor]
	}
	return nil
}

func (mmm *MultiMonitorManager) GetMonitorLayout() MonitorLayout {
	mmm.mu.RLock()
	defer mmm.mu.RUnlock()
	return mmm.monitorLayout
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
