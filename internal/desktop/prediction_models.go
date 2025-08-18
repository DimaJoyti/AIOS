package desktop

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/aios/aios/internal/ai"
	"github.com/aios/aios/pkg/models"
)

// FrequencyBasedModel predicts based on historical frequency
type FrequencyBasedModel struct {
	accuracy          float64
	applicationCounts map[string]int
	totalEvents       int
}

func NewFrequencyBasedModel() *FrequencyBasedModel {
	return &FrequencyBasedModel{
		accuracy:          0.5, // Initial accuracy
		applicationCounts: make(map[string]int),
		totalEvents:       0,
	}
}

func (fbm *FrequencyBasedModel) GetName() string {
	return "frequency_based"
}

func (fbm *FrequencyBasedModel) GetAccuracy() float64 {
	return fbm.accuracy
}

func (fbm *FrequencyBasedModel) Train(history []FocusEvent) error {
	fbm.applicationCounts = make(map[string]int)
	fbm.totalEvents = len(history)
	
	for _, event := range history {
		fbm.applicationCounts[event.Application]++
	}
	
	// Calculate accuracy based on how well frequency predicts actual usage
	if fbm.totalEvents > 0 {
		fbm.accuracy = fbm.calculateAccuracy(history)
	}
	
	return nil
}

func (fbm *FrequencyBasedModel) Predict(contextFeatures *ContextFeatures, history []FocusEvent) []WindowPrediction {
	predictions := make([]WindowPrediction, 0)
	
	// Get unique applications from recent history
	recentApps := make(map[string]bool)
	for _, app := range contextFeatures.RecentActivity {
		recentApps[app] = true
	}
	
	// Create predictions based on frequency
	for app, count := range fbm.applicationCounts {
		if recentApps[app] { // Only predict for currently available applications
			confidence := float64(count) / float64(fbm.totalEvents)
			
			predictions = append(predictions, WindowPrediction{
				WindowID:    fmt.Sprintf("%s_window", app),
				Application: app,
				Confidence:  confidence,
				Reasoning:   fmt.Sprintf("Frequency: %d/%d (%.2f%%)", count, fbm.totalEvents, confidence*100),
				Timestamp:   time.Now(),
			})
		}
	}
	
	// Sort by confidence
	sort.Slice(predictions, func(i, j int) bool {
		return predictions[i].Confidence > predictions[j].Confidence
	})
	
	// Return top 5 predictions
	if len(predictions) > 5 {
		predictions = predictions[:5]
	}
	
	return predictions
}

func (fbm *FrequencyBasedModel) calculateAccuracy(history []FocusEvent) float64 {
	if len(history) < 2 {
		return 0.5
	}
	
	correct := 0
	total := 0
	
	// Test predictions against actual history
	for i := 1; i < len(history); i++ {
		// Predict based on frequency up to this point
		tempCounts := make(map[string]int)
		for j := 0; j < i; j++ {
			tempCounts[history[j].Application]++
		}
		
		// Find most frequent application
		maxCount := 0
		mostFrequent := ""
		for app, count := range tempCounts {
			if count > maxCount {
				maxCount = count
				mostFrequent = app
			}
		}
		
		// Check if prediction matches actual
		if mostFrequent == history[i].Application {
			correct++
		}
		total++
	}
	
	if total > 0 {
		return float64(correct) / float64(total)
	}
	return 0.5
}

// PatternBasedModel predicts based on learned patterns
type PatternBasedModel struct {
	accuracy float64
	patterns map[string]float64 // pattern -> confidence
}

func NewPatternBasedModel() *PatternBasedModel {
	return &PatternBasedModel{
		accuracy: 0.6,
		patterns: make(map[string]float64),
	}
}

func (pbm *PatternBasedModel) GetName() string {
	return "pattern_based"
}

func (pbm *PatternBasedModel) GetAccuracy() float64 {
	return pbm.accuracy
}

func (pbm *PatternBasedModel) Train(history []FocusEvent) error {
	pbm.patterns = make(map[string]float64)
	
	// Learn 2-gram patterns
	for i := 1; i < len(history); i++ {
		pattern := fmt.Sprintf("%s->%s", history[i-1].Application, history[i].Application)
		pbm.patterns[pattern]++
	}
	
	// Normalize patterns to probabilities
	total := 0.0
	for _, count := range pbm.patterns {
		total += count
	}
	
	if total > 0 {
		for pattern := range pbm.patterns {
			pbm.patterns[pattern] /= total
		}
	}
	
	pbm.accuracy = pbm.calculateAccuracy(history)
	return nil
}

func (pbm *PatternBasedModel) Predict(contextFeatures *ContextFeatures, history []FocusEvent) []WindowPrediction {
	predictions := make([]WindowPrediction, 0)
	
	if len(history) == 0 {
		return predictions
	}
	
	// Get last application
	lastApp := history[len(history)-1].Application
	
	// Find patterns starting with last application
	for pattern, confidence := range pbm.patterns {
		if len(pattern) > len(lastApp)+2 && pattern[:len(lastApp)] == lastApp && pattern[len(lastApp):len(lastApp)+2] == "->" {
			nextApp := pattern[len(lastApp)+2:]
			
			predictions = append(predictions, WindowPrediction{
				WindowID:    fmt.Sprintf("%s_window", nextApp),
				Application: nextApp,
				Confidence:  confidence,
				Reasoning:   fmt.Sprintf("Pattern: %s (%.2f%%)", pattern, confidence*100),
				Timestamp:   time.Now(),
			})
		}
	}
	
	// Sort by confidence
	sort.Slice(predictions, func(i, j int) bool {
		return predictions[i].Confidence > predictions[j].Confidence
	})
	
	return predictions
}

func (pbm *PatternBasedModel) calculateAccuracy(history []FocusEvent) float64 {
	if len(history) < 3 {
		return 0.6
	}
	
	correct := 0
	total := 0
	
	// Test pattern predictions
	for i := 2; i < len(history); i++ {
		pattern := fmt.Sprintf("%s->%s", history[i-2].Application, history[i-1].Application)
		if confidence, exists := pbm.patterns[pattern]; exists && confidence > 0.1 {
			// This is a simplified accuracy calculation
			if history[i-1].Application == history[i].Application {
				correct++
			}
			total++
		}
	}
	
	if total > 0 {
		return float64(correct) / float64(total)
	}
	return 0.6
}

// TemporalModel predicts based on time patterns
type TemporalModel struct {
	accuracy      float64
	timePatterns  map[string]map[string]float64 // hour -> app -> probability
	dayPatterns   map[string]map[string]float64 // day -> app -> probability
}

func NewTemporalModel() *TemporalModel {
	return &TemporalModel{
		accuracy:     0.55,
		timePatterns: make(map[string]map[string]float64),
		dayPatterns:  make(map[string]map[string]float64),
	}
}

func (tm *TemporalModel) GetName() string {
	return "temporal"
}

func (tm *TemporalModel) GetAccuracy() float64 {
	return tm.accuracy
}

func (tm *TemporalModel) Train(history []FocusEvent) error {
	tm.timePatterns = make(map[string]map[string]float64)
	tm.dayPatterns = make(map[string]map[string]float64)
	
	// Learn time-based patterns
	for _, event := range history {
		hour := fmt.Sprintf("hour_%d", event.Timestamp.Hour())
		day := event.Timestamp.Weekday().String()
		
		// Hour patterns
		if _, exists := tm.timePatterns[hour]; !exists {
			tm.timePatterns[hour] = make(map[string]float64)
		}
		tm.timePatterns[hour][event.Application]++
		
		// Day patterns
		if _, exists := tm.dayPatterns[day]; !exists {
			tm.dayPatterns[day] = make(map[string]float64)
		}
		tm.dayPatterns[day][event.Application]++
	}
	
	// Normalize to probabilities
	tm.normalizePatterns(tm.timePatterns)
	tm.normalizePatterns(tm.dayPatterns)
	
	tm.accuracy = tm.calculateAccuracy(history)
	return nil
}

func (tm *TemporalModel) normalizePatterns(patterns map[string]map[string]float64) {
	for timeKey, appCounts := range patterns {
		total := 0.0
		for _, count := range appCounts {
			total += count
		}
		
		if total > 0 {
			for app := range appCounts {
				patterns[timeKey][app] /= total
			}
		}
	}
}

func (tm *TemporalModel) Predict(contextFeatures *ContextFeatures, history []FocusEvent) []WindowPrediction {
	predictions := make([]WindowPrediction, 0)
	
	hour := fmt.Sprintf("hour_%d", contextFeatures.TimeOfDay)
	day := time.Now().Weekday().String()
	
	// Combine hour and day predictions
	combinedScores := make(map[string]float64)
	
	// Hour-based predictions
	if hourPatterns, exists := tm.timePatterns[hour]; exists {
		for app, prob := range hourPatterns {
			combinedScores[app] += prob * 0.7 // Weight hour patterns more heavily
		}
	}
	
	// Day-based predictions
	if dayPatterns, exists := tm.dayPatterns[day]; exists {
		for app, prob := range dayPatterns {
			combinedScores[app] += prob * 0.3
		}
	}
	
	// Create predictions
	for app, score := range combinedScores {
		predictions = append(predictions, WindowPrediction{
			WindowID:    fmt.Sprintf("%s_window", app),
			Application: app,
			Confidence:  score,
			Reasoning:   fmt.Sprintf("Temporal: %s at %s (%.2f%%)", app, hour, score*100),
			Timestamp:   time.Now(),
		})
	}
	
	// Sort by confidence
	sort.Slice(predictions, func(i, j int) bool {
		return predictions[i].Confidence > predictions[j].Confidence
	})
	
	return predictions
}

func (tm *TemporalModel) calculateAccuracy(history []FocusEvent) float64 {
	// Simplified accuracy calculation for temporal patterns
	return 0.55
}

// ContextBasedModel predicts based on current contextFeatures
type ContextBasedModel struct {
	accuracy        float64
	contextFeaturesPatterns map[string]map[string]float64 // contextFeatures -> app -> probability
}

func NewContextBasedModel() *ContextBasedModel {
	return &ContextBasedModel{
		accuracy:        0.65,
		contextFeaturesPatterns: make(map[string]map[string]float64),
	}
}

func (cbm *ContextBasedModel) GetName() string {
	return "contextFeatures_based"
}

func (cbm *ContextBasedModel) GetAccuracy() float64 {
	return cbm.accuracy
}

func (cbm *ContextBasedModel) Train(history []FocusEvent) error {
	cbm.contextFeaturesPatterns = make(map[string]map[string]float64)
	
	// Learn contextFeatures-based patterns
	for _, event := range history {
		// Create contextFeatures key from available contextFeatures information
		contextFeaturesKey := cbm.createContextKey(event.Context)
		
		if _, exists := cbm.contextFeaturesPatterns[contextFeaturesKey]; !exists {
			cbm.contextFeaturesPatterns[contextFeaturesKey] = make(map[string]float64)
		}
		cbm.contextFeaturesPatterns[contextFeaturesKey][event.Application]++
	}
	
	// Normalize to probabilities
	for contextFeaturesKey, appCounts := range cbm.contextFeaturesPatterns {
		total := 0.0
		for _, count := range appCounts {
			total += count
		}
		
		if total > 0 {
			for app := range appCounts {
				cbm.contextFeaturesPatterns[contextFeaturesKey][app] /= total
			}
		}
	}
	
	return nil
}

func (cbm *ContextBasedModel) createContextKey(contextFeatures map[string]interface{}) string {
	// Create a simple contextFeatures key from available contextFeatures
	// In a real implementation, this would be more sophisticated
	return "default_contextFeatures"
}

func (cbm *ContextBasedModel) Predict(contextFeatures *ContextFeatures, history []FocusEvent) []WindowPrediction {
	predictions := make([]WindowPrediction, 0)
	
	// Create contextFeatures key for current contextFeatures
	contextFeaturesKey := "default_contextFeatures" // Simplified
	
	if patterns, exists := cbm.contextFeaturesPatterns[contextFeaturesKey]; exists {
		for app, prob := range patterns {
			predictions = append(predictions, WindowPrediction{
				WindowID:    fmt.Sprintf("%s_window", app),
				Application: app,
				Confidence:  prob,
				Reasoning:   fmt.Sprintf("Context: %s (%.2f%%)", contextFeaturesKey, prob*100),
				Timestamp:   time.Now(),
			})
		}
	}
	
	// Sort by confidence
	sort.Slice(predictions, func(i, j int) bool {
		return predictions[i].Confidence > predictions[j].Confidence
	})
	
	return predictions
}

// AIAssistedModel uses AI for predictions
type AIAssistedModel struct {
	accuracy       float64
	aiOrchestrator *ai.Orchestrator
}

func NewAIAssistedModel(aiOrchestrator *ai.Orchestrator) *AIAssistedModel {
	return &AIAssistedModel{
		accuracy:       0.75,
		aiOrchestrator: aiOrchestrator,
	}
}

func (aam *AIAssistedModel) GetName() string {
	return "ai_assisted"
}

func (aam *AIAssistedModel) GetAccuracy() float64 {
	return aam.accuracy
}

func (aam *AIAssistedModel) Train(history []FocusEvent) error {
	// AI model training would be more complex
	// For now, just return success
	return nil
}

func (aam *AIAssistedModel) Predict(contextFeatures *ContextFeatures, history []FocusEvent) []WindowPrediction {
	if aam.aiOrchestrator == nil {
		return []WindowPrediction{}
	}
	
	// Create AI request for focus prediction
	aiRequest := &models.AIRequest{
		ID:   fmt.Sprintf("focus-prediction-%d", time.Now().Unix()),
		Type: "prediction",
		Input: fmt.Sprintf("Predict next window focus based on context: %+v and history: %+v", contextFeatures, history),
		Parameters: map[string]interface{}{
			"task":    "focus_prediction",
			"context": contextFeatures,
			"history": aam.summarizeHistory(history),
		},
		Timeout:   2 * time.Second,
		Timestamp: time.Now(),
	}
	
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	
	response, err := aam.aiOrchestrator.ProcessRequest(ctx, aiRequest)
	if err != nil {
		return []WindowPrediction{}
	}
	
	// Parse AI response
	return aam.parseAIResponse(response)
}

func (aam *AIAssistedModel) summarizeHistory(history []FocusEvent) []map[string]interface{} {
	// Summarize recent history for AI
	recentCount := min(10, len(history))
	summary := make([]map[string]interface{}, recentCount)
	
	for i := 0; i < recentCount; i++ {
		event := history[len(history)-1-i]
		summary[i] = map[string]interface{}{
			"application": event.Application,
			"duration":    event.Duration.Seconds(),
			"time_ago":    time.Since(event.Timestamp).Seconds(),
		}
	}
	
	return summary
}

func (aam *AIAssistedModel) parseAIResponse(response *models.AIResponse) []WindowPrediction {
	// Parse AI response into predictions
	// This would be more sophisticated in a real implementation
	predictions := make([]WindowPrediction, 0)
	
	if result, ok := response.Result.(map[string]interface{}); ok {
		if predList, ok := result["predictions"].([]interface{}); ok {
			for _, pred := range predList {
				if predMap, ok := pred.(map[string]interface{}); ok {
					prediction := WindowPrediction{
						WindowID:    fmt.Sprintf("%v_window", predMap["application"]),
						Application: fmt.Sprintf("%v", predMap["application"]),
						Confidence:  0.7, // Default confidence
						Reasoning:   "AI prediction",
						Timestamp:   time.Now(),
					}
					
					if conf, ok := predMap["confidence"].(float64); ok {
						prediction.Confidence = conf
					}
					
					predictions = append(predictions, prediction)
				}
			}
		}
	}
	
	return predictions
}

// Helper function
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
