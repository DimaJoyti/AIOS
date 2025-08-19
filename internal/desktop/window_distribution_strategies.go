package desktop

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/aios/aios/internal/ai"
	"github.com/aios/aios/pkg/models"
)

// BalancedDistributionStrategy distributes windows evenly across monitors
type BalancedDistributionStrategy struct{}

func (bds *BalancedDistributionStrategy) GetName() string {
	return "balanced"
}

func (bds *BalancedDistributionStrategy) GetDescription() string {
	return "Distributes windows evenly across all available monitors"
}

func (bds *BalancedDistributionStrategy) DistributeWindows(windows []*models.Window, monitors map[string]*Monitor) map[string][]*models.Window {
	distribution := make(map[string][]*models.Window)

	// Initialize distribution map
	connectedMonitors := make([]*Monitor, 0)
	for _, monitor := range monitors {
		if monitor.IsConnected {
			connectedMonitors = append(connectedMonitors, monitor)
			distribution[monitor.ID] = make([]*models.Window, 0)
		}
	}

	if len(connectedMonitors) == 0 {
		return distribution
	}

	// Sort monitors by ID for consistent distribution
	sort.Slice(connectedMonitors, func(i, j int) bool {
		return connectedMonitors[i].ID < connectedMonitors[j].ID
	})

	// Distribute windows round-robin
	for i, window := range windows {
		monitorIndex := i % len(connectedMonitors)
		monitorID := connectedMonitors[monitorIndex].ID
		distribution[monitorID] = append(distribution[monitorID], window)
	}

	return distribution
}

// PrimaryFocusedDistributionStrategy focuses windows on the primary monitor
type PrimaryFocusedDistributionStrategy struct {
	primaryMonitor string
}

func (pfds *PrimaryFocusedDistributionStrategy) GetName() string {
	return "primary_focused"
}

func (pfds *PrimaryFocusedDistributionStrategy) GetDescription() string {
	return "Places most windows on the primary monitor, overflow to secondary monitors"
}

func (pfds *PrimaryFocusedDistributionStrategy) DistributeWindows(windows []*models.Window, monitors map[string]*Monitor) map[string][]*models.Window {
	distribution := make(map[string][]*models.Window)

	// Initialize distribution map
	var primaryMon *Monitor
	secondaryMonitors := make([]*Monitor, 0)

	for _, monitor := range monitors {
		if monitor.IsConnected {
			distribution[monitor.ID] = make([]*models.Window, 0)
			if monitor.ID == pfds.primaryMonitor {
				primaryMon = monitor
			} else {
				secondaryMonitors = append(secondaryMonitors, monitor)
			}
		}
	}

	if primaryMon == nil {
		// Fall back to balanced distribution if no primary monitor
		balanced := &BalancedDistributionStrategy{}
		return balanced.DistributeWindows(windows, monitors)
	}

	// Calculate primary monitor capacity (e.g., 70% of windows)
	primaryCapacity := int(float64(len(windows)) * 0.7)
	if primaryCapacity < 1 {
		primaryCapacity = len(windows)
	}

	// Place windows on primary monitor
	for i := 0; i < primaryCapacity && i < len(windows); i++ {
		distribution[primaryMon.ID] = append(distribution[primaryMon.ID], windows[i])
	}

	// Distribute remaining windows to secondary monitors
	remainingWindows := windows[primaryCapacity:]
	if len(remainingWindows) > 0 && len(secondaryMonitors) > 0 {
		for i, window := range remainingWindows {
			monitorIndex := i % len(secondaryMonitors)
			monitorID := secondaryMonitors[monitorIndex].ID
			distribution[monitorID] = append(distribution[monitorID], window)
		}
	}

	return distribution
}

// AIOptimizedDistributionStrategy uses AI to optimize window distribution
type AIOptimizedDistributionStrategy struct {
	aiOrchestrator *ai.Orchestrator
}

func (aods *AIOptimizedDistributionStrategy) GetName() string {
	return "ai_optimized"
}

func (aods *AIOptimizedDistributionStrategy) GetDescription() string {
	return "Uses AI to optimize window distribution based on context and user behavior"
}

func (aods *AIOptimizedDistributionStrategy) DistributeWindows(windows []*models.Window, monitors map[string]*Monitor) map[string][]*models.Window {
	// If AI orchestrator is not available, fall back to balanced distribution
	if aods.aiOrchestrator == nil {
		balanced := &BalancedDistributionStrategy{}
		return balanced.DistributeWindows(windows, monitors)
	}

	// Analyze context for AI optimization
	distributionContext := aods.analyzeDistributionContext(windows, monitors)

	// Get AI recommendation
	recommendation := aods.getAIRecommendation(distributionContext)

	// Apply AI recommendation or fall back to balanced
	if recommendation != nil {
		return aods.applyAIRecommendation(windows, monitors, recommendation)
	}

	// Fall back to balanced distribution
	balanced := &BalancedDistributionStrategy{}
	return balanced.DistributeWindows(windows, monitors)
}

func (aods *AIOptimizedDistributionStrategy) analyzeDistributionContext(windows []*models.Window, monitors map[string]*Monitor) map[string]interface{} {
	analysisContext := make(map[string]interface{})

	// Window analysis
	analysisContext["window_count"] = len(windows)
	analysisContext["window_types"] = aods.analyzeWindowTypes(windows)
	analysisContext["window_sizes"] = aods.analyzeWindowSizes(windows)

	// Monitor analysis
	analysisContext["monitor_count"] = len(monitors)
	analysisContext["monitor_resolutions"] = aods.analyzeMonitorResolutions(monitors)
	analysisContext["monitor_layout"] = aods.analyzeMonitorLayout(monitors)

	// Temporal context
	analysisContext["time_of_day"] = time.Now().Hour()
	analysisContext["day_of_week"] = time.Now().Weekday().String()

	return analysisContext
}

func (aods *AIOptimizedDistributionStrategy) analyzeWindowTypes(windows []*models.Window) map[string]int {
	types := make(map[string]int)
	for _, window := range windows {
		types[window.Application]++
	}
	return types
}

func (aods *AIOptimizedDistributionStrategy) analyzeWindowSizes(windows []*models.Window) map[string]interface{} {
	if len(windows) == 0 {
		return map[string]interface{}{}
	}

	totalArea := 0
	minArea := windows[0].Size.Width * windows[0].Size.Height
	maxArea := minArea

	for _, window := range windows {
		area := window.Size.Width * window.Size.Height
		totalArea += area
		if area < minArea {
			minArea = area
		}
		if area > maxArea {
			maxArea = area
		}
	}

	return map[string]interface{}{
		"average_area": totalArea / len(windows),
		"min_area":     minArea,
		"max_area":     maxArea,
	}
}

func (aods *AIOptimizedDistributionStrategy) analyzeMonitorResolutions(monitors map[string]*Monitor) []models.Size {
	resolutions := make([]models.Size, 0)
	for _, monitor := range monitors {
		if monitor.IsConnected {
			resolutions = append(resolutions, monitor.Resolution)
		}
	}
	return resolutions
}

func (aods *AIOptimizedDistributionStrategy) analyzeMonitorLayout(monitors map[string]*Monitor) string {
	connectedCount := 0
	for _, monitor := range monitors {
		if monitor.IsConnected {
			connectedCount++
		}
	}

	switch connectedCount {
	case 1:
		return "single"
	case 2:
		return "dual"
	case 3:
		return "triple"
	default:
		return "multi"
	}
}

func (aods *AIOptimizedDistributionStrategy) getAIRecommendation(analysisContext map[string]interface{}) map[string]interface{} {
	// Create AI request for distribution recommendation
	aiRequest := &models.AIRequest{
		ID:    fmt.Sprintf("window-distribution-%d", time.Now().Unix()),
		Type:  "recommendation",
		Input: fmt.Sprintf("Optimize window distribution for context: %+v", analysisContext),
		Parameters: map[string]interface{}{
			"task":    "window_distribution",
			"context": analysisContext,
		},
		Timeout:   3 * time.Second,
		Timestamp: time.Now(),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	response, err := aods.aiOrchestrator.ProcessRequest(ctx, aiRequest)
	if err != nil {
		return nil
	}

	// Parse AI response
	if result, ok := response.Result.(map[string]interface{}); ok {
		return result
	}

	return nil
}

func (aods *AIOptimizedDistributionStrategy) applyAIRecommendation(windows []*models.Window, monitors map[string]*Monitor, recommendation map[string]interface{}) map[string][]*models.Window {
	distribution := make(map[string][]*models.Window)

	// Initialize distribution map
	for _, monitor := range monitors {
		if monitor.IsConnected {
			distribution[monitor.ID] = make([]*models.Window, 0)
		}
	}

	// Try to parse AI recommendation
	if monitorAssignments, ok := recommendation["monitor_assignments"].(map[string]interface{}); ok {
		// Apply AI-recommended assignments
		for windowID, monitorID := range monitorAssignments {
			if monitorIDStr, ok := monitorID.(string); ok {
				// Find window by ID
				for _, window := range windows {
					if window.ID == windowID {
						if _, exists := distribution[monitorIDStr]; exists {
							distribution[monitorIDStr] = append(distribution[monitorIDStr], window)
						}
						break
					}
				}
			}
		}

		// Check if all windows were assigned
		totalAssigned := 0
		for _, windowList := range distribution {
			totalAssigned += len(windowList)
		}

		if totalAssigned == len(windows) {
			return distribution
		}
	}

	// If AI recommendation couldn't be applied, fall back to balanced
	balanced := &BalancedDistributionStrategy{}
	return balanced.DistributeWindows(windows, monitors)
}

// ApplicationAwareDistributionStrategy distributes based on application types
type ApplicationAwareDistributionStrategy struct {
	applicationRules map[string]string // application -> preferred_monitor_type
}

func NewApplicationAwareDistributionStrategy() *ApplicationAwareDistributionStrategy {
	return &ApplicationAwareDistributionStrategy{
		applicationRules: map[string]string{
			"code_editor":   "primary",
			"web_browser":   "secondary",
			"terminal":      "primary",
			"media_player":  "secondary",
			"communication": "secondary",
			"design_tool":   "primary",
			"game":          "primary",
		},
	}
}

func (aads *ApplicationAwareDistributionStrategy) GetName() string {
	return "application_aware"
}

func (aads *ApplicationAwareDistributionStrategy) GetDescription() string {
	return "Distributes windows based on application types and predefined rules"
}

func (aads *ApplicationAwareDistributionStrategy) DistributeWindows(windows []*models.Window, monitors map[string]*Monitor) map[string][]*models.Window {
	distribution := make(map[string][]*models.Window)

	// Find primary and secondary monitors
	var primaryMonitor, secondaryMonitor *Monitor
	for _, monitor := range monitors {
		if monitor.IsConnected {
			distribution[monitor.ID] = make([]*models.Window, 0)
			if monitor.IsPrimary {
				primaryMonitor = monitor
			} else if secondaryMonitor == nil {
				secondaryMonitor = monitor
			}
		}
	}

	// Distribute windows based on application rules
	for _, window := range windows {
		targetMonitor := aads.selectMonitorForApplication(window.Application, primaryMonitor, secondaryMonitor)
		if targetMonitor != nil {
			distribution[targetMonitor.ID] = append(distribution[targetMonitor.ID], window)
		} else if primaryMonitor != nil {
			// Fall back to primary monitor
			distribution[primaryMonitor.ID] = append(distribution[primaryMonitor.ID], window)
		}
	}

	return distribution
}

func (aads *ApplicationAwareDistributionStrategy) selectMonitorForApplication(application string, primary, secondary *Monitor) *Monitor {
	if rule, exists := aads.applicationRules[application]; exists {
		switch rule {
		case "primary":
			return primary
		case "secondary":
			if secondary != nil {
				return secondary
			}
			return primary // Fall back to primary if no secondary
		}
	}

	// Default to primary monitor
	return primary
}

// ContextAwareDistributionStrategy distributes based on current context
type ContextAwareDistributionStrategy struct {
	timeBasedRules map[string]string // time_range -> strategy
}

func NewContextAwareDistributionStrategy() *ContextAwareDistributionStrategy {
	return &ContextAwareDistributionStrategy{
		timeBasedRules: map[string]string{
			"morning":   "primary_focused", // 6-12
			"afternoon": "balanced",        // 12-18
			"evening":   "primary_focused", // 18-22
			"night":     "balanced",        // 22-6
		},
	}
}

func (cads *ContextAwareDistributionStrategy) GetName() string {
	return "context_aware"
}

func (cads *ContextAwareDistributionStrategy) GetDescription() string {
	return "Adapts distribution strategy based on time of day and other contextual factors"
}

func (cads *ContextAwareDistributionStrategy) DistributeWindows(windows []*models.Window, monitors map[string]*Monitor) map[string][]*models.Window {
	// Determine current context
	strategy := cads.selectStrategyForContext()

	// Apply the selected strategy
	switch strategy {
	case "balanced":
		balanced := &BalancedDistributionStrategy{}
		return balanced.DistributeWindows(windows, monitors)
	case "primary_focused":
		// Find primary monitor
		primaryID := ""
		for _, monitor := range monitors {
			if monitor.IsPrimary {
				primaryID = monitor.ID
				break
			}
		}
		primaryFocused := &PrimaryFocusedDistributionStrategy{primaryMonitor: primaryID}
		return primaryFocused.DistributeWindows(windows, monitors)
	default:
		balanced := &BalancedDistributionStrategy{}
		return balanced.DistributeWindows(windows, monitors)
	}
}

func (cads *ContextAwareDistributionStrategy) selectStrategyForContext() string {
	hour := time.Now().Hour()

	switch {
	case hour >= 6 && hour < 12:
		return cads.timeBasedRules["morning"]
	case hour >= 12 && hour < 18:
		return cads.timeBasedRules["afternoon"]
	case hour >= 18 && hour < 22:
		return cads.timeBasedRules["evening"]
	default:
		return cads.timeBasedRules["night"]
	}
}
