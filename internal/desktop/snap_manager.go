package desktop

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/aios/aios/internal/ai"
	"github.com/aios/aios/pkg/models"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

// SnapManager handles intelligent window snapping
type SnapManager struct {
	logger *logrus.Logger
	tracer trace.Tracer
	config SnapConfig
	mu     sync.RWMutex

	// AI integration
	aiOrchestrator *ai.Orchestrator

	// Snap zones and targets
	snapZones   []SnapZone
	snapTargets []SnapTarget
	customZones []CustomSnapZone

	// State tracking
	activeSnaps map[string]*ActiveSnap
	snapHistory []SnapEvent
	preferences *SnapPreferences

	// Performance
	lastUpdate  time.Time
	snapCount   int
	successRate float64
}

// SnapConfig defines snap manager configuration
type SnapConfig struct {
	SnapThreshold      int           `json:"snap_threshold"` // pixels
	SnapDelay          time.Duration `json:"snap_delay"`
	ShowSnapPreview    bool          `json:"show_snap_preview"`
	EnableMagneticSnap bool          `json:"enable_magnetic_snap"`
	CrossMonitorSnap   bool          `json:"cross_monitor_snap"`
	AISnapSuggestions  bool          `json:"ai_snap_suggestions"`
	SnapAnimation      bool          `json:"snap_animation"`
	SnapFeedback       bool          `json:"snap_feedback"`
	CustomZonesEnabled bool          `json:"custom_zones_enabled"`
	SmartSnapZones     bool          `json:"smart_snap_zones"`
}

// SnapZone represents a snapping zone
type SnapZone struct {
	ID         string                 `json:"id"`
	Name       string                 `json:"name"`
	Type       SnapZoneType           `json:"type"`
	Area       models.Rectangle       `json:"area"`
	Monitor    string                 `json:"monitor"`
	Priority   int                    `json:"priority"`
	Enabled    bool                   `json:"enabled"`
	Magnetic   bool                   `json:"magnetic"`
	Conditions []SnapCondition        `json:"conditions"`
	Actions    []SnapAction           `json:"actions"`
	Metadata   map[string]interface{} `json:"metadata"`
}

// SnapZoneType defines the type of snap zone
type SnapZoneType string

const (
	SnapZoneEdge    SnapZoneType = "edge"
	SnapZoneCorner  SnapZoneType = "corner"
	SnapZoneCenter  SnapZoneType = "center"
	SnapZoneQuarter SnapZoneType = "quarter"
	SnapZoneThird   SnapZoneType = "third"
	SnapZoneCustom  SnapZoneType = "custom"
	SnapZoneAI      SnapZoneType = "ai_suggested"
)

// SnapTarget represents a potential snap target
type SnapTarget struct {
	Zone       SnapZone        `json:"zone"`
	Position   models.Position `json:"position"`
	Size       models.Size     `json:"size"`
	Confidence float64         `json:"confidence"`
	Distance   float64         `json:"distance"`
	Reasoning  string          `json:"reasoning"`
	WindowID   string          `json:"window_id"`
	Timestamp  time.Time       `json:"timestamp"`
}

// CustomSnapZone represents a user-defined snap zone
type CustomSnapZone struct {
	ID          string           `json:"id"`
	Name        string           `json:"name"`
	Description string           `json:"description"`
	Area        models.Rectangle `json:"area"`
	Monitor     string           `json:"monitor"`
	Hotkey      string           `json:"hotkey"`
	Conditions  []SnapCondition  `json:"conditions"`
	CreatedAt   time.Time        `json:"created_at"`
	UsageCount  int              `json:"usage_count"`
}

// SnapCondition defines when a snap zone should be active
type SnapCondition struct {
	Type     string      `json:"type"`     // "application", "window_count", "time", "workspace"
	Operator string      `json:"operator"` // "equals", "contains", "greater_than"
	Value    interface{} `json:"value"`
}

// SnapAction defines what happens when snapping to a zone
type SnapAction struct {
	Type       string                 `json:"type"` // "resize", "move", "maximize", "tile"
	Parameters map[string]interface{} `json:"parameters"`
	Animation  bool                   `json:"animation"`
}

// ActiveSnap represents an active snapping operation
type ActiveSnap struct {
	WindowID   string    `json:"window_id"`
	TargetZone SnapZone  `json:"target_zone"`
	StartTime  time.Time `json:"start_time"`
	Progress   float64   `json:"progress"`
	Status     string    `json:"status"` // "detecting", "previewing", "snapping", "completed"
}

// SnapEvent represents a snapping event
type SnapEvent struct {
	WindowID  string                 `json:"window_id"`
	ZoneID    string                 `json:"zone_id"`
	ZoneType  SnapZoneType           `json:"zone_type"`
	Timestamp time.Time              `json:"timestamp"`
	Duration  time.Duration          `json:"duration"`
	Success   bool                   `json:"success"`
	Method    string                 `json:"method"` // "drag", "hotkey", "ai_suggestion"
	Context   map[string]interface{} `json:"context"`
}

// SnapPreferences stores user snapping preferences
type SnapPreferences struct {
	PreferredZones    []string           `json:"preferred_zones"`
	ZoneWeights       map[string]float64 `json:"zone_weights"`
	ApplicationRules  map[string]string  `json:"application_rules"`
	DisabledZones     []string           `json:"disabled_zones"`
	CustomHotkeys     map[string]string  `json:"custom_hotkeys"`
	AdaptationEnabled bool               `json:"adaptation_enabled"`
	LastUpdated       time.Time          `json:"last_updated"`
}

// NewSnapManager creates a new snap manager
func NewSnapManager(logger *logrus.Logger, config SnapConfig, aiOrchestrator *ai.Orchestrator) *SnapManager {
	tracer := otel.Tracer("snap-manager")

	manager := &SnapManager{
		logger:         logger,
		tracer:         tracer,
		config:         config,
		aiOrchestrator: aiOrchestrator,
		snapZones:      make([]SnapZone, 0),
		snapTargets:    make([]SnapTarget, 0),
		customZones:    make([]CustomSnapZone, 0),
		activeSnaps:    make(map[string]*ActiveSnap),
		snapHistory:    make([]SnapEvent, 0),
		preferences: &SnapPreferences{
			PreferredZones:    make([]string, 0),
			ZoneWeights:       make(map[string]float64),
			ApplicationRules:  make(map[string]string),
			DisabledZones:     make([]string, 0),
			CustomHotkeys:     make(map[string]string),
			AdaptationEnabled: true,
			LastUpdated:       time.Now(),
		},
		lastUpdate: time.Now(),
	}

	// Initialize default snap zones
	manager.initializeDefaultZones()

	return manager
}

// initializeDefaultZones creates default snap zones
func (sm *SnapManager) initializeDefaultZones() {
	// Standard edge zones
	zones := []SnapZone{
		{
			ID:       "left-half",
			Name:     "Left Half",
			Type:     SnapZoneEdge,
			Area:     models.Rectangle{X: 0, Y: 0, Width: 50, Height: 100}, // Percentage
			Priority: 1,
			Enabled:  true,
			Magnetic: true,
		},
		{
			ID:       "right-half",
			Name:     "Right Half",
			Type:     SnapZoneEdge,
			Area:     models.Rectangle{X: 50, Y: 0, Width: 50, Height: 100},
			Priority: 1,
			Enabled:  true,
			Magnetic: true,
		},
		{
			ID:       "top-half",
			Name:     "Top Half",
			Type:     SnapZoneEdge,
			Area:     models.Rectangle{X: 0, Y: 0, Width: 100, Height: 50},
			Priority: 2,
			Enabled:  true,
			Magnetic: true,
		},
		{
			ID:       "bottom-half",
			Name:     "Bottom Half",
			Type:     SnapZoneEdge,
			Area:     models.Rectangle{X: 0, Y: 50, Width: 100, Height: 50},
			Priority: 2,
			Enabled:  true,
			Magnetic: true,
		},
		{
			ID:       "maximize",
			Name:     "Maximize",
			Type:     SnapZoneCenter,
			Area:     models.Rectangle{X: 0, Y: 0, Width: 100, Height: 100},
			Priority: 3,
			Enabled:  true,
			Magnetic: false,
		},
	}

	// Corner zones
	corners := []SnapZone{
		{
			ID:       "top-left",
			Name:     "Top Left Quarter",
			Type:     SnapZoneCorner,
			Area:     models.Rectangle{X: 0, Y: 0, Width: 50, Height: 50},
			Priority: 1,
			Enabled:  true,
			Magnetic: true,
		},
		{
			ID:       "top-right",
			Name:     "Top Right Quarter",
			Type:     SnapZoneCorner,
			Area:     models.Rectangle{X: 50, Y: 0, Width: 50, Height: 50},
			Priority: 1,
			Enabled:  true,
			Magnetic: true,
		},
		{
			ID:       "bottom-left",
			Name:     "Bottom Left Quarter",
			Type:     SnapZoneCorner,
			Area:     models.Rectangle{X: 0, Y: 50, Width: 50, Height: 50},
			Priority: 1,
			Enabled:  true,
			Magnetic: true,
		},
		{
			ID:       "bottom-right",
			Name:     "Bottom Right Quarter",
			Type:     SnapZoneCorner,
			Area:     models.Rectangle{X: 50, Y: 50, Width: 50, Height: 50},
			Priority: 1,
			Enabled:  true,
			Magnetic: true,
		},
	}

	sm.snapZones = append(sm.snapZones, zones...)
	sm.snapZones = append(sm.snapZones, corners...)

	// Initialize zone weights
	for _, zone := range sm.snapZones {
		sm.preferences.ZoneWeights[zone.ID] = 1.0
	}
}

// DetectSnapTargets detects potential snap targets for a window
func (sm *SnapManager) DetectSnapTargets(ctx context.Context, window *models.Window, mousePos models.Position) ([]SnapTarget, error) {
	ctx, span := sm.tracer.Start(ctx, "snapManager.DetectSnapTargets")
	defer span.End()

	sm.mu.RLock()
	defer sm.mu.RUnlock()

	targets := make([]SnapTarget, 0)

	// Check each snap zone
	for _, zone := range sm.snapZones {
		if !zone.Enabled {
			continue
		}

		// Check if zone conditions are met
		if !sm.evaluateZoneConditions(&zone, window) {
			continue
		}

		// Calculate distance to zone
		distance := sm.calculateDistanceToZone(mousePos, zone)

		// Check if within snap threshold
		if distance <= float64(sm.config.SnapThreshold) {
			target := SnapTarget{
				Zone:       zone,
				Position:   sm.calculateSnapPosition(window, zone),
				Size:       sm.calculateSnapSize(window, zone),
				Confidence: sm.calculateSnapConfidence(window, zone, distance),
				Distance:   distance,
				Reasoning:  fmt.Sprintf("Within %dpx of %s zone", sm.config.SnapThreshold, zone.Name),
				WindowID:   window.ID,
				Timestamp:  time.Now(),
			}

			targets = append(targets, target)
		}
	}

	// Add AI-suggested targets if enabled
	if sm.config.AISnapSuggestions && sm.aiOrchestrator != nil {
		aiTargets := sm.getAISnapSuggestions(ctx, window, mousePos)
		targets = append(targets, aiTargets...)
	}

	// Sort by confidence and distance
	sort.Slice(targets, func(i, j int) bool {
		if targets[i].Confidence != targets[j].Confidence {
			return targets[i].Confidence > targets[j].Confidence
		}
		return targets[i].Distance < targets[j].Distance
	})

	sm.logger.WithFields(logrus.Fields{
		"window_id":    window.ID,
		"target_count": len(targets),
		"mouse_pos":    mousePos,
	}).Debug("Snap targets detected")

	return targets, nil
}

// SnapToTarget snaps a window to a specific target
func (sm *SnapManager) SnapToTarget(ctx context.Context, window *models.Window, target SnapTarget) error {
	ctx, span := sm.tracer.Start(ctx, "snapManager.SnapToTarget")
	defer span.End()

	start := time.Now()

	// Create active snap
	activeSnap := &ActiveSnap{
		WindowID:   window.ID,
		TargetZone: target.Zone,
		StartTime:  start,
		Progress:   0.0,
		Status:     "snapping",
	}

	sm.mu.Lock()
	sm.activeSnaps[window.ID] = activeSnap
	sm.mu.Unlock()

	// Execute snap actions
	success := true
	var snapError error

	for _, action := range target.Zone.Actions {
		if err := sm.executeSnapAction(&action, window, target); err != nil {
			success = false
			snapError = err
			break
		}
	}

	// Update active snap
	sm.mu.Lock()
	activeSnap.Progress = 1.0
	activeSnap.Status = "completed"
	delete(sm.activeSnaps, window.ID)
	sm.mu.Unlock()

	// Record snap event
	event := SnapEvent{
		WindowID:  window.ID,
		ZoneID:    target.Zone.ID,
		ZoneType:  target.Zone.Type,
		Timestamp: start,
		Duration:  time.Since(start),
		Success:   success,
		Method:    "drag", // This would be determined by the calling context
		Context: map[string]interface{}{
			"confidence": target.Confidence,
			"distance":   target.Distance,
		},
	}

	sm.recordSnapEvent(event)

	// Update preferences based on success
	if success {
		sm.updateSnapPreferences(target.Zone.ID, true)
	}

	sm.logger.WithFields(logrus.Fields{
		"window_id": window.ID,
		"zone_id":   target.Zone.ID,
		"success":   success,
		"duration":  event.Duration,
	}).Debug("Window snapped")

	return snapError
}

// calculateDistanceToZone calculates distance from a point to a snap zone
func (sm *SnapManager) calculateDistanceToZone(point models.Position, zone SnapZone) float64 {
	// Convert percentage-based zone to absolute coordinates
	// This would need screen dimensions in a real implementation
	screenWidth, screenHeight := 1920, 1080 // Placeholder

	zoneX := (zone.Area.X * screenWidth) / 100
	zoneY := (zone.Area.Y * screenHeight) / 100
	zoneW := (zone.Area.Width * screenWidth) / 100
	zoneH := (zone.Area.Height * screenHeight) / 100

	// Calculate distance to zone edges
	dx := math.Max(0, math.Max(float64(zoneX-point.X), float64(point.X-(zoneX+zoneW))))
	dy := math.Max(0, math.Max(float64(zoneY-point.Y), float64(point.Y-(zoneY+zoneH))))

	return math.Sqrt(dx*dx + dy*dy)
}

// calculateSnapPosition calculates the position for snapping
func (sm *SnapManager) calculateSnapPosition(window *models.Window, zone SnapZone) models.Position {
	// Convert percentage-based zone to absolute coordinates
	screenWidth, screenHeight := 1920, 1080 // Placeholder

	x := (zone.Area.X * screenWidth) / 100
	y := (zone.Area.Y * screenHeight) / 100

	return models.Position{X: x, Y: y}
}

// calculateSnapSize calculates the size for snapping
func (sm *SnapManager) calculateSnapSize(window *models.Window, zone SnapZone) models.Size {
	// Convert percentage-based zone to absolute coordinates
	screenWidth, screenHeight := 1920, 1080 // Placeholder

	width := (zone.Area.Width * screenWidth) / 100
	height := (zone.Area.Height * screenHeight) / 100

	return models.Size{Width: width, Height: height}
}

// calculateSnapConfidence calculates confidence for a snap target
func (sm *SnapManager) calculateSnapConfidence(window *models.Window, zone SnapZone, distance float64) float64 {
	// Base confidence from distance
	distanceConfidence := 1.0 - (distance / float64(sm.config.SnapThreshold))

	// Zone priority factor
	priorityFactor := 1.0 / float64(zone.Priority)

	// User preference factor
	preferenceWeight := sm.preferences.ZoneWeights[zone.ID]

	// Combine factors
	confidence := distanceConfidence * priorityFactor * preferenceWeight

	// Clamp to [0, 1]
	if confidence < 0 {
		confidence = 0
	}
	if confidence > 1 {
		confidence = 1
	}

	return confidence
}

// evaluateZoneConditions evaluates if zone conditions are met
func (sm *SnapManager) evaluateZoneConditions(zone *SnapZone, window *models.Window) bool {
	if len(zone.Conditions) == 0 {
		return true
	}

	for _, condition := range zone.Conditions {
		if !sm.evaluateSnapCondition(&condition, window) {
			return false
		}
	}

	return true
}

// evaluateSnapCondition evaluates a single snap condition
func (sm *SnapManager) evaluateSnapCondition(condition *SnapCondition, window *models.Window) bool {
	switch condition.Type {
	case "application":
		return sm.evaluateApplicationCondition(condition, window.Application)
	case "window_count":
		// Would need access to window count
		return true
	case "time":
		return sm.evaluateTimeCondition(condition)
	case "workspace":
		return sm.evaluateWorkspaceCondition(condition, window.Workspace)
	default:
		return true
	}
}

// evaluateApplicationCondition evaluates application-based conditions
func (sm *SnapManager) evaluateApplicationCondition(condition *SnapCondition, application string) bool {
	conditionValue, ok := condition.Value.(string)
	if !ok {
		return false
	}

	switch condition.Operator {
	case "equals":
		return application == conditionValue
	case "contains":
		return strings.Contains(strings.ToLower(application), strings.ToLower(conditionValue))
	default:
		return false
	}
}

// evaluateTimeCondition evaluates time-based conditions
func (sm *SnapManager) evaluateTimeCondition(condition *SnapCondition) bool {
	// Implementation would check time-based conditions
	return true
}

// evaluateWorkspaceCondition evaluates workspace-based conditions
func (sm *SnapManager) evaluateWorkspaceCondition(condition *SnapCondition, workspace int) bool {
	conditionValue, ok := condition.Value.(float64)
	if !ok {
		return false
	}

	switch condition.Operator {
	case "equals":
		return workspace == int(conditionValue)
	default:
		return false
	}
}

// executeSnapAction executes a snap action
func (sm *SnapManager) executeSnapAction(action *SnapAction, window *models.Window, target SnapTarget) error {
	switch action.Type {
	case "move":
		// Implementation would move the window
		return nil
	case "resize":
		// Implementation would resize the window
		return nil
	case "maximize":
		// Implementation would maximize the window
		return nil
	default:
		return fmt.Errorf("unknown snap action: %s", action.Type)
	}
}

// getAISnapSuggestions gets AI-powered snap suggestions
func (sm *SnapManager) getAISnapSuggestions(ctx context.Context, window *models.Window, mousePos models.Position) []SnapTarget {
	// Create AI request for snap suggestions
	aiRequest := &models.AIRequest{
		ID:    fmt.Sprintf("snap-suggestion-%s-%d", window.ID, time.Now().Unix()),
		Type:  "suggestion",
		Input: fmt.Sprintf("Suggest optimal snap zones for window %s at position %+v", window.Application, mousePos),
		Parameters: map[string]interface{}{
			"task":      "snap_suggestion",
			"window":    window,
			"mouse_pos": mousePos,
			"history":   sm.getRecentSnapHistory(10),
		},
		Timeout:   1 * time.Second,
		Timestamp: time.Now(),
	}

	response, err := sm.aiOrchestrator.ProcessRequest(ctx, aiRequest)
	if err != nil {
		return []SnapTarget{}
	}

	// Parse AI response into snap targets
	return sm.parseAISnapResponse(response, window)
}

// parseAISnapResponse parses AI response into snap targets
func (sm *SnapManager) parseAISnapResponse(response *models.AIResponse, window *models.Window) []SnapTarget {
	// Implementation would parse AI response
	return []SnapTarget{}
}

// Helper methods

func (sm *SnapManager) recordSnapEvent(event SnapEvent) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.snapHistory = append(sm.snapHistory, event)
	sm.snapCount++

	// Maintain history size
	if len(sm.snapHistory) > 1000 {
		sm.snapHistory = sm.snapHistory[100:]
	}

	// Update success rate
	successCount := 0
	for _, e := range sm.snapHistory {
		if e.Success {
			successCount++
		}
	}
	sm.successRate = float64(successCount) / float64(len(sm.snapHistory))
}

func (sm *SnapManager) updateSnapPreferences(zoneID string, success bool) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if success {
		sm.preferences.ZoneWeights[zoneID] *= 1.1
	} else {
		sm.preferences.ZoneWeights[zoneID] *= 0.9
	}

	// Normalize weights
	total := 0.0
	for _, weight := range sm.preferences.ZoneWeights {
		total += weight
	}

	if total > 0 {
		for zoneID := range sm.preferences.ZoneWeights {
			sm.preferences.ZoneWeights[zoneID] /= total
		}
	}

	sm.preferences.LastUpdated = time.Now()
}

func (sm *SnapManager) getRecentSnapHistory(count int) []SnapEvent {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	if len(sm.snapHistory) <= count {
		return sm.snapHistory
	}
	return sm.snapHistory[len(sm.snapHistory)-count:]
}

// Public API methods

func (sm *SnapManager) GetSnapZones() []SnapZone {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.snapZones
}

func (sm *SnapManager) GetSnapPreferences() *SnapPreferences {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.preferences
}

func (sm *SnapManager) GetSnapMetrics() map[string]interface{} {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	return map[string]interface{}{
		"total_snaps":  sm.snapCount,
		"success_rate": sm.successRate,
		"active_snaps": len(sm.activeSnaps),
		"zone_count":   len(sm.snapZones),
		"custom_zones": len(sm.customZones),
	}
}
