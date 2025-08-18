package desktop

import (
	"math"

	"github.com/aios/aios/internal/ai"
	"github.com/aios/aios/pkg/models"
)

// BinarySpacePartitioning implements BSP tiling algorithm
type BinarySpacePartitioning struct {
	splitRatio    float64
	minWindowSize models.Size
}

// NewBinarySpacePartitioning creates a new BSP algorithm
func NewBinarySpacePartitioning() *BinarySpacePartitioning {
	return &BinarySpacePartitioning{
		splitRatio:    0.5,
		minWindowSize: models.Size{Width: 200, Height: 150},
	}
}

func (bsp *BinarySpacePartitioning) Name() string { return "bsp" }
func (bsp *BinarySpacePartitioning) Description() string { return "Binary Space Partitioning" }
func (bsp *BinarySpacePartitioning) SupportsResize() bool { return true }

func (bsp *BinarySpacePartitioning) GetParameters() map[string]interface{} {
	return map[string]interface{}{
		"split_ratio":     bsp.splitRatio,
		"min_window_size": bsp.minWindowSize,
	}
}

func (bsp *BinarySpacePartitioning) SetParameters(params map[string]interface{}) error {
	if ratio, ok := params["split_ratio"].(float64); ok {
		bsp.splitRatio = ratio
	}
	return nil
}

func (bsp *BinarySpacePartitioning) Tile(windows []*models.Window, workspace models.Rectangle) ([]*WindowPlacement, error) {
	if len(windows) == 0 {
		return []*WindowPlacement{}, nil
	}

	placements := make([]*WindowPlacement, len(windows))
	bsp.tileRecursive(windows, workspace, 0, placements, true)
	return placements, nil
}

func (bsp *BinarySpacePartitioning) tileRecursive(windows []*models.Window, rect models.Rectangle, index int, placements []*WindowPlacement, horizontal bool) {
	if index >= len(windows) {
		return
	}

	if index == len(windows)-1 {
		// Last window gets the remaining space
		placements[index] = &WindowPlacement{
			WindowID: windows[index].ID,
			Position: models.Position{X: rect.X, Y: rect.Y},
			Size:     models.Size{Width: rect.Width, Height: rect.Height},
		}
		return
	}

	// Split the rectangle
	var rect1, rect2 models.Rectangle
	if horizontal {
		splitX := rect.X + int(float64(rect.Width)*bsp.splitRatio)
		rect1 = models.Rectangle{X: rect.X, Y: rect.Y, Width: splitX - rect.X, Height: rect.Height}
		rect2 = models.Rectangle{X: splitX, Y: rect.Y, Width: rect.X + rect.Width - splitX, Height: rect.Height}
	} else {
		splitY := rect.Y + int(float64(rect.Height)*bsp.splitRatio)
		rect1 = models.Rectangle{X: rect.X, Y: rect.Y, Width: rect.Width, Height: splitY - rect.Y}
		rect2 = models.Rectangle{X: rect.X, Y: splitY, Width: rect.Width, Height: rect.Y + rect.Height - splitY}
	}

	// Place current window in first rectangle
	placements[index] = &WindowPlacement{
		WindowID: windows[index].ID,
		Position: models.Position{X: rect1.X, Y: rect1.Y},
		Size:     models.Size{Width: rect1.Width, Height: rect1.Height},
	}

	// Recursively tile remaining windows in second rectangle
	bsp.tileRecursive(windows[index+1:], rect2, index+1, placements, !horizontal)
}

// MasterStackAlgorithm implements master-stack tiling
type MasterStackAlgorithm struct {
	masterRatio   float64
	stackVertical bool
}

func NewMasterStackAlgorithm() *MasterStackAlgorithm {
	return &MasterStackAlgorithm{
		masterRatio:   0.6,
		stackVertical: true,
	}
}

func (ms *MasterStackAlgorithm) Name() string { return "master_stack" }
func (ms *MasterStackAlgorithm) Description() string { return "Master-Stack Layout" }
func (ms *MasterStackAlgorithm) SupportsResize() bool { return true }

func (ms *MasterStackAlgorithm) GetParameters() map[string]interface{} {
	return map[string]interface{}{
		"master_ratio":   ms.masterRatio,
		"stack_vertical": ms.stackVertical,
	}
}

func (ms *MasterStackAlgorithm) SetParameters(params map[string]interface{}) error {
	if ratio, ok := params["master_ratio"].(float64); ok {
		ms.masterRatio = ratio
	}
	if vertical, ok := params["stack_vertical"].(bool); ok {
		ms.stackVertical = vertical
	}
	return nil
}

func (ms *MasterStackAlgorithm) Tile(windows []*models.Window, workspace models.Rectangle) ([]*WindowPlacement, error) {
	if len(windows) == 0 {
		return []*WindowPlacement{}, nil
	}

	placements := make([]*WindowPlacement, len(windows))

	if len(windows) == 1 {
		// Single window takes full space
		placements[0] = &WindowPlacement{
			WindowID: windows[0].ID,
			Position: models.Position{X: workspace.X, Y: workspace.Y},
			Size:     models.Size{Width: workspace.Width, Height: workspace.Height},
		}
		return placements, nil
	}

	// Calculate master area
	masterWidth := int(float64(workspace.Width) * ms.masterRatio)
	stackWidth := workspace.Width - masterWidth

	// Place master window
	placements[0] = &WindowPlacement{
		WindowID: windows[0].ID,
		Position: models.Position{X: workspace.X, Y: workspace.Y},
		Size:     models.Size{Width: masterWidth, Height: workspace.Height},
	}

	// Place stack windows
	stackWindows := windows[1:]
	if ms.stackVertical {
		stackHeight := workspace.Height / len(stackWindows)
		for i, window := range stackWindows {
			placements[i+1] = &WindowPlacement{
				WindowID: window.ID,
				Position: models.Position{
					X: workspace.X + masterWidth,
					Y: workspace.Y + i*stackHeight,
				},
				Size: models.Size{Width: stackWidth, Height: stackHeight},
			}
		}
	} else {
		stackWidth := stackWidth / len(stackWindows)
		for i, window := range stackWindows {
			placements[i+1] = &WindowPlacement{
				WindowID: window.ID,
				Position: models.Position{
					X: workspace.X + masterWidth + i*stackWidth,
					Y: workspace.Y,
				},
				Size: models.Size{Width: stackWidth, Height: workspace.Height},
			}
		}
	}

	return placements, nil
}

// GridAlgorithm implements grid-based tiling
type GridAlgorithm struct {
	columns int
	rows    int
}

func NewGridAlgorithm() *GridAlgorithm {
	return &GridAlgorithm{
		columns: 0, // Auto-calculate
		rows:    0, // Auto-calculate
	}
}

func (g *GridAlgorithm) Name() string { return "grid" }
func (g *GridAlgorithm) Description() string { return "Grid Layout" }
func (g *GridAlgorithm) SupportsResize() bool { return true }

func (g *GridAlgorithm) GetParameters() map[string]interface{} {
	return map[string]interface{}{
		"columns": g.columns,
		"rows":    g.rows,
	}
}

func (g *GridAlgorithm) SetParameters(params map[string]interface{}) error {
	if cols, ok := params["columns"].(int); ok {
		g.columns = cols
	}
	if rows, ok := params["rows"].(int); ok {
		g.rows = rows
	}
	return nil
}

func (g *GridAlgorithm) Tile(windows []*models.Window, workspace models.Rectangle) ([]*WindowPlacement, error) {
	if len(windows) == 0 {
		return []*WindowPlacement{}, nil
	}

	// Calculate optimal grid dimensions
	cols, rows := g.calculateGridDimensions(len(windows))
	
	cellWidth := workspace.Width / cols
	cellHeight := workspace.Height / rows

	placements := make([]*WindowPlacement, len(windows))

	for i, window := range windows {
		col := i % cols
		row := i / cols

		placements[i] = &WindowPlacement{
			WindowID: window.ID,
			Position: models.Position{
				X: workspace.X + col*cellWidth,
				Y: workspace.Y + row*cellHeight,
			},
			Size: models.Size{Width: cellWidth, Height: cellHeight},
		}
	}

	return placements, nil
}

func (g *GridAlgorithm) calculateGridDimensions(windowCount int) (int, int) {
	if g.columns > 0 && g.rows > 0 {
		return g.columns, g.rows
	}

	// Calculate optimal grid based on window count
	sqrt := math.Sqrt(float64(windowCount))
	cols := int(math.Ceil(sqrt))
	rows := int(math.Ceil(float64(windowCount) / float64(cols)))

	return cols, rows
}

// SpiralAlgorithm implements spiral tiling
type SpiralAlgorithm struct {
	clockwise bool
}

func NewSpiralAlgorithm() *SpiralAlgorithm {
	return &SpiralAlgorithm{clockwise: true}
}

func (s *SpiralAlgorithm) Name() string { return "spiral" }
func (s *SpiralAlgorithm) Description() string { return "Spiral Layout" }
func (s *SpiralAlgorithm) SupportsResize() bool { return false }

func (s *SpiralAlgorithm) GetParameters() map[string]interface{} {
	return map[string]interface{}{"clockwise": s.clockwise}
}

func (s *SpiralAlgorithm) SetParameters(params map[string]interface{}) error {
	if cw, ok := params["clockwise"].(bool); ok {
		s.clockwise = cw
	}
	return nil
}

func (s *SpiralAlgorithm) Tile(windows []*models.Window, workspace models.Rectangle) ([]*WindowPlacement, error) {
	// Simplified spiral implementation - place windows in a spiral pattern
	// For now, fall back to grid layout
	// A full spiral implementation would be more complex
	gridAlgo := NewGridAlgorithm()
	return gridAlgo.Tile(windows, workspace)
}

// FloatingAlgorithm implements floating window management
type FloatingAlgorithm struct {
	cascade bool
	offset  int
}

func NewFloatingAlgorithm() *FloatingAlgorithm {
	return &FloatingAlgorithm{
		cascade: true,
		offset:  30,
	}
}

func (f *FloatingAlgorithm) Name() string { return "floating" }
func (f *FloatingAlgorithm) Description() string { return "Floating Windows" }
func (f *FloatingAlgorithm) SupportsResize() bool { return true }

func (f *FloatingAlgorithm) GetParameters() map[string]interface{} {
	return map[string]interface{}{
		"cascade": f.cascade,
		"offset":  f.offset,
	}
}

func (f *FloatingAlgorithm) SetParameters(params map[string]interface{}) error {
	if cascade, ok := params["cascade"].(bool); ok {
		f.cascade = cascade
	}
	if offset, ok := params["offset"].(int); ok {
		f.offset = offset
	}
	return nil
}

func (f *FloatingAlgorithm) Tile(windows []*models.Window, workspace models.Rectangle) ([]*WindowPlacement, error) {
	placements := make([]*WindowPlacement, len(windows))

	for i, window := range windows {
		if f.cascade {
			// Cascade windows with offset
			placements[i] = &WindowPlacement{
				WindowID: window.ID,
				Position: models.Position{
					X: workspace.X + i*f.offset,
					Y: workspace.Y + i*f.offset,
				},
				Size: models.Size{
					Width:  workspace.Width - i*f.offset*2,
					Height: workspace.Height - i*f.offset*2,
				},
			}
		} else {
			// Keep original positions and sizes
			placements[i] = &WindowPlacement{
				WindowID: window.ID,
				Position: window.Position,
				Size:     window.Size,
			}
		}
	}

	return placements, nil
}

// TabletAlgorithm optimized for tablet/touch interfaces
type TabletAlgorithm struct {
	maxColumns int
	minSize    models.Size
}

func NewTabletAlgorithm() *TabletAlgorithm {
	return &TabletAlgorithm{
		maxColumns: 2,
		minSize:    models.Size{Width: 400, Height: 300},
	}
}

func (t *TabletAlgorithm) Name() string { return "tablet" }
func (t *TabletAlgorithm) Description() string { return "Tablet-Optimized Layout" }
func (t *TabletAlgorithm) SupportsResize() bool { return true }

func (t *TabletAlgorithm) GetParameters() map[string]interface{} {
	return map[string]interface{}{
		"max_columns": t.maxColumns,
		"min_size":    t.minSize,
	}
}

func (t *TabletAlgorithm) SetParameters(params map[string]interface{}) error {
	if cols, ok := params["max_columns"].(int); ok {
		t.maxColumns = cols
	}
	return nil
}

func (t *TabletAlgorithm) Tile(windows []*models.Window, workspace models.Rectangle) ([]*WindowPlacement, error) {
	// Use grid algorithm with tablet-specific constraints
	gridAlgo := NewGridAlgorithm()
	gridAlgo.columns = minInt(t.maxColumns, len(windows))
	return gridAlgo.Tile(windows, workspace)
}

// AIOptimizedAlgorithm uses AI to determine optimal layout
type AIOptimizedAlgorithm struct {
	aiOrchestrator *ai.Orchestrator
	fallbackAlgo   TilingAlgorithm
}

func NewAIOptimizedAlgorithm(aiOrchestrator *ai.Orchestrator) *AIOptimizedAlgorithm {
	return &AIOptimizedAlgorithm{
		aiOrchestrator: aiOrchestrator,
		fallbackAlgo:   NewGridAlgorithm(),
	}
}

func (ai *AIOptimizedAlgorithm) Name() string { return "ai_optimized" }
func (ai *AIOptimizedAlgorithm) Description() string { return "AI-Optimized Layout" }
func (ai *AIOptimizedAlgorithm) SupportsResize() bool { return true }

func (ai *AIOptimizedAlgorithm) GetParameters() map[string]interface{} {
	return map[string]interface{}{}
}

func (ai *AIOptimizedAlgorithm) SetParameters(params map[string]interface{}) error {
	return nil
}

func (ai *AIOptimizedAlgorithm) Tile(windows []*models.Window, workspace models.Rectangle) ([]*WindowPlacement, error) {
	// For now, use fallback algorithm
	// In a full implementation, this would use AI to generate optimal layouts
	return ai.fallbackAlgo.Tile(windows, workspace)
}

// Helper functions

// min function removed - use minInt from focus_predictor.go
