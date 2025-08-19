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

// TilingEngine handles advanced window tiling algorithms
type TilingEngine struct {
	logger     *logrus.Logger
	tracer     trace.Tracer
	config     TilingConfig
	algorithms map[string]TilingAlgorithm
	mu         sync.RWMutex

	// AI integration
	aiOrchestrator *ai.Orchestrator

	// State
	currentAlgorithm string
	tilingHistory    []TilingEvent
	preferences      *UserTilingPreferences
}

// TilingConfig defines tiling engine configuration
type TilingConfig struct {
	DefaultAlgorithm  string        `json:"default_algorithm"`
	GapSize           int           `json:"gap_size"`
	BorderWidth       int           `json:"border_width"`
	MinWindowSize     models.Size   `json:"min_window_size"`
	MaxWindowSize     models.Size   `json:"max_window_size"`
	AnimationDuration time.Duration `json:"animation_duration"`
	SmartGaps         bool          `json:"smart_gaps"`
	AdaptiveLayout    bool          `json:"adaptive_layout"`
	AIOptimization    bool          `json:"ai_optimization"`
}

// TilingAlgorithm interface for different tiling strategies
type TilingAlgorithm interface {
	Name() string
	Description() string
	Tile(windows []*models.Window, workspace models.Rectangle) ([]*WindowPlacement, error)
	SupportsResize() bool
	GetParameters() map[string]interface{}
	SetParameters(params map[string]interface{}) error
}

// WindowPlacement represents a window's position and size in a tiling layout
type WindowPlacement struct {
	WindowID  string          `json:"window_id"`
	Position  models.Position `json:"position"`
	Size      models.Size     `json:"size"`
	ZIndex    int             `json:"z_index"`
	Workspace int             `json:"workspace"`
	Monitor   string          `json:"monitor"`
}

// TilingEvent represents a tiling operation
type TilingEvent struct {
	Timestamp   time.Time         `json:"timestamp"`
	Algorithm   string            `json:"algorithm"`
	WindowCount int               `json:"window_count"`
	Workspace   int               `json:"workspace"`
	Duration    time.Duration     `json:"duration"`
	Success     bool              `json:"success"`
	Error       string            `json:"error,omitempty"`
	Placements  []WindowPlacement `json:"placements"`
}

// UserTilingPreferences stores user preferences for tiling
type UserTilingPreferences struct {
	PreferredAlgorithm string             `json:"preferred_algorithm"`
	AlgorithmWeights   map[string]float64 `json:"algorithm_weights"`
	ContextPreferences map[string]string  `json:"context_preferences"`
	CustomLayouts      []CustomLayout     `json:"custom_layouts"`
	AdaptationEnabled  bool               `json:"adaptation_enabled"`
	LastUpdated        time.Time          `json:"last_updated"`
}

// CustomLayout represents a user-defined layout
type CustomLayout struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Placements  []WindowPlacement `json:"placements"`
	Conditions  []LayoutCondition `json:"conditions"`
	CreatedAt   time.Time         `json:"created_at"`
}

// LayoutCondition defines when a custom layout should be applied
type LayoutCondition struct {
	Type     string      `json:"type"`     // "window_count", "application", "time", "context"
	Operator string      `json:"operator"` // "equals", "greater_than", "contains", etc.
	Value    interface{} `json:"value"`
	Weight   float64     `json:"weight"`
}

// NewTilingEngine creates a new tiling engine
func NewTilingEngine(logger *logrus.Logger, config TilingConfig, aiOrchestrator *ai.Orchestrator) *TilingEngine {
	tracer := otel.Tracer("tiling-engine")

	engine := &TilingEngine{
		logger:         logger,
		tracer:         tracer,
		config:         config,
		algorithms:     make(map[string]TilingAlgorithm),
		aiOrchestrator: aiOrchestrator,
		tilingHistory:  make([]TilingEvent, 0),
		preferences: &UserTilingPreferences{
			PreferredAlgorithm: config.DefaultAlgorithm,
			AlgorithmWeights:   make(map[string]float64),
			ContextPreferences: make(map[string]string),
			CustomLayouts:      make([]CustomLayout, 0),
			AdaptationEnabled:  true,
			LastUpdated:        time.Now(),
		},
	}

	// Register built-in algorithms
	engine.registerBuiltinAlgorithms()

	return engine
}

// registerBuiltinAlgorithms registers the built-in tiling algorithms
func (te *TilingEngine) registerBuiltinAlgorithms() {
	algorithms := []TilingAlgorithm{
		NewBinarySpacePartitioning(),
		NewMasterStackAlgorithm(),
		NewGridAlgorithm(),
		NewSpiralAlgorithm(),
		NewFloatingAlgorithm(),
		NewTabletAlgorithm(),
		NewAIOptimizedAlgorithm(te.aiOrchestrator),
	}

	for _, algo := range algorithms {
		te.algorithms[algo.Name()] = algo
		te.preferences.AlgorithmWeights[algo.Name()] = 1.0
	}

	te.currentAlgorithm = te.config.DefaultAlgorithm
}

// TileWindows arranges windows using the current tiling algorithm
func (te *TilingEngine) TileWindows(ctx context.Context, windows []*models.Window, workspace models.Rectangle) ([]*WindowPlacement, error) {
	ctx, span := te.tracer.Start(ctx, "tilingEngine.TileWindows")
	defer span.End()

	start := time.Now()

	te.mu.Lock()
	defer te.mu.Unlock()

	// Select optimal algorithm
	algorithm := te.selectOptimalAlgorithm(ctx, windows, workspace)

	te.logger.WithFields(logrus.Fields{
		"algorithm":    algorithm,
		"window_count": len(windows),
		"workspace":    workspace,
	}).Debug("Tiling windows")

	// Get the algorithm implementation
	algo, exists := te.algorithms[algorithm]
	if !exists {
		return nil, fmt.Errorf("tiling algorithm %s not found", algorithm)
	}

	// Perform tiling
	placements, err := algo.Tile(windows, workspace)
	if err != nil {
		te.recordTilingEvent(TilingEvent{
			Timestamp:   start,
			Algorithm:   algorithm,
			WindowCount: len(windows),
			Duration:    time.Since(start),
			Success:     false,
			Error:       err.Error(),
		})
		return nil, fmt.Errorf("tiling failed: %w", err)
	}

	// Apply smart gaps if enabled
	if te.config.SmartGaps {
		placements = te.applySmartGaps(placements, workspace)
	}

	// Record successful tiling
	placementValues := make([]WindowPlacement, len(placements))
	for i, p := range placements {
		if p != nil {
			placementValues[i] = *p
		}
	}

	te.recordTilingEvent(TilingEvent{
		Timestamp:   start,
		Algorithm:   algorithm,
		WindowCount: len(windows),
		Duration:    time.Since(start),
		Success:     true,
		Placements:  placementValues,
	})

	// Update algorithm weights based on success
	te.updateAlgorithmWeights(algorithm, true, time.Since(start))

	return placements, nil
}

// selectOptimalAlgorithm chooses the best tiling algorithm for the current context
func (te *TilingEngine) selectOptimalAlgorithm(ctx context.Context, windows []*models.Window, workspace models.Rectangle) string {
	// If AI optimization is disabled, use preferred algorithm
	if !te.config.AIOptimization {
		return te.preferences.PreferredAlgorithm
	}

	// Analyze context
	context := te.analyzeContext(windows, workspace)

	// Check for custom layouts that match conditions
	if customAlgo := te.findMatchingCustomLayout(context); customAlgo != "" {
		return customAlgo
	}

	// Use AI to select optimal algorithm
	if te.aiOrchestrator != nil {
		if aiAlgo := te.getAIRecommendation(ctx, context); aiAlgo != "" {
			return aiAlgo
		}
	}

	// Fall back to weighted selection based on historical performance
	return te.selectByWeight(context)
}

// analyzeContext analyzes the current context for algorithm selection
func (te *TilingEngine) analyzeContext(windows []*models.Window, workspace models.Rectangle) map[string]interface{} {
	context := make(map[string]interface{})

	context["window_count"] = len(windows)
	context["workspace_ratio"] = float64(workspace.Width) / float64(workspace.Height)
	context["workspace_area"] = workspace.Width * workspace.Height
	context["time_of_day"] = time.Now().Hour()
	context["day_of_week"] = time.Now().Weekday().String()

	// Analyze window types
	appTypes := make(map[string]int)
	totalArea := 0
	for _, window := range windows {
		appTypes[window.Application]++
		totalArea += window.Size.Width * window.Size.Height
	}

	context["application_types"] = appTypes
	context["average_window_area"] = totalArea / len(windows)
	context["dominant_application"] = te.findDominantApplication(appTypes)

	return context
}

// findDominantApplication finds the most common application type
func (te *TilingEngine) findDominantApplication(appTypes map[string]int) string {
	maxCount := 0
	dominant := ""

	for app, count := range appTypes {
		if count > maxCount {
			maxCount = count
			dominant = app
		}
	}

	return dominant
}

// getAIRecommendation gets algorithm recommendation from AI
func (te *TilingEngine) getAIRecommendation(ctx context.Context, context map[string]interface{}) string {
	// Create AI request for algorithm recommendation
	aiRequest := &models.AIRequest{
		ID:    fmt.Sprintf("tiling-algo-%d", time.Now().Unix()),
		Type:  "recommendation",
		Input: fmt.Sprintf("Recommend optimal tiling algorithm for context: %+v", context),
		Parameters: map[string]interface{}{
			"task":       "algorithm_selection",
			"algorithms": te.getAlgorithmNames(),
			"context":    context,
			"history":    te.getRecentHistory(10),
		},
		Timeout:   5 * time.Second,
		Timestamp: time.Now(),
	}

	response, err := te.aiOrchestrator.ProcessRequest(ctx, aiRequest)
	if err != nil {
		te.logger.WithError(err).Debug("Failed to get AI algorithm recommendation")
		return ""
	}

	// Parse AI response to extract algorithm name
	if result, ok := response.Result.(string); ok {
		if _, exists := te.algorithms[result]; exists {
			return result
		}
	}

	return ""
}

// selectByWeight selects algorithm based on historical performance weights
func (te *TilingEngine) selectByWeight(context map[string]interface{}) string {
	// Calculate weighted scores for each algorithm
	scores := make(map[string]float64)

	for name, weight := range te.preferences.AlgorithmWeights {
		scores[name] = weight

		// Apply context-specific adjustments
		if contextPref, exists := te.preferences.ContextPreferences[te.contextKey(context)]; exists {
			if contextPref == name {
				scores[name] *= 1.5 // Boost preferred algorithm for this context
			}
		}
	}

	// Select algorithm with highest score
	maxScore := 0.0
	selected := te.preferences.PreferredAlgorithm

	for name, score := range scores {
		if score > maxScore {
			maxScore = score
			selected = name
		}
	}

	return selected
}

// contextKey generates a key for the current context
func (te *TilingEngine) contextKey(context map[string]interface{}) string {
	windowCount := context["window_count"].(int)
	dominantApp := context["dominant_application"].(string)
	return fmt.Sprintf("%d_%s", windowCount, dominantApp)
}

// applySmartGaps applies intelligent gap sizing based on context
func (te *TilingEngine) applySmartGaps(placements []*WindowPlacement, workspace models.Rectangle) []*WindowPlacement {
	if len(placements) <= 1 {
		return placements // No gaps needed for single window
	}

	// Calculate adaptive gap size
	gapSize := te.calculateAdaptiveGapSize(len(placements), workspace)

	// Apply gaps between windows
	for i, placement := range placements {
		// Reduce window size to accommodate gaps
		placement.Size.Width -= gapSize
		placement.Size.Height -= gapSize

		// Adjust position to center window in its allocated space
		placement.Position.X += gapSize / 2
		placement.Position.Y += gapSize / 2

		placements[i] = placement
	}

	return placements
}

// calculateAdaptiveGapSize calculates optimal gap size based on context
func (te *TilingEngine) calculateAdaptiveGapSize(windowCount int, workspace models.Rectangle) int {
	baseGap := te.config.GapSize

	// Reduce gaps for many windows
	if windowCount > 4 {
		baseGap = int(float64(baseGap) * 0.7)
	}

	// Adjust based on workspace size
	workspaceArea := workspace.Width * workspace.Height
	if workspaceArea < 1920*1080 { // Small screen
		baseGap = int(float64(baseGap) * 0.8)
	}

	return baseGap
}

// Helper methods

func (te *TilingEngine) getAlgorithmNames() []string {
	names := make([]string, 0, len(te.algorithms))
	for name := range te.algorithms {
		names = append(names, name)
	}
	return names
}

func (te *TilingEngine) getRecentHistory(count int) []TilingEvent {
	if len(te.tilingHistory) <= count {
		return te.tilingHistory
	}
	return te.tilingHistory[len(te.tilingHistory)-count:]
}

func (te *TilingEngine) recordTilingEvent(event TilingEvent) {
	te.tilingHistory = append(te.tilingHistory, event)

	// Keep only recent history
	if len(te.tilingHistory) > 1000 {
		te.tilingHistory = te.tilingHistory[100:]
	}
}

func (te *TilingEngine) updateAlgorithmWeights(algorithm string, success bool, duration time.Duration) {
	if success {
		// Increase weight for successful algorithms
		te.preferences.AlgorithmWeights[algorithm] *= 1.1

		// Bonus for fast algorithms
		if duration < 100*time.Millisecond {
			te.preferences.AlgorithmWeights[algorithm] *= 1.05
		}
	} else {
		// Decrease weight for failed algorithms
		te.preferences.AlgorithmWeights[algorithm] *= 0.9
	}

	// Normalize weights
	te.normalizeWeights()
}

func (te *TilingEngine) normalizeWeights() {
	total := 0.0
	for _, weight := range te.preferences.AlgorithmWeights {
		total += weight
	}

	if total > 0 {
		for name := range te.preferences.AlgorithmWeights {
			te.preferences.AlgorithmWeights[name] /= total
		}
	}
}

func (te *TilingEngine) findMatchingCustomLayout(context map[string]interface{}) string {
	// Implementation for finding matching custom layouts
	// This would evaluate layout conditions against the current context
	return ""
}
