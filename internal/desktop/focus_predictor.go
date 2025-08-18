package desktop

import (
	"context"
	"fmt"
	"math"
	"sort"
	"sync"
	"time"

	"github.com/aios/aios/internal/ai"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

// FocusPredictor predicts which window the user is likely to focus next
type FocusPredictor struct {
	logger  *logrus.Logger
	tracer  trace.Tracer
	config  FocusPredictorConfig
	mu      sync.RWMutex
	
	// AI integration
	aiOrchestrator *ai.Orchestrator
	
	// Learning data
	focusHistory    []FocusEvent
	patterns        []FocusPattern
	userBehavior    *UserBehaviorProfile
	contextFeatures *ContextFeatures
	
	// Prediction models
	models map[string]PredictionModel
	
	// Performance metrics
	predictions     []PredictionResult
	accuracy        float64
	lastTraining    time.Time
	trainingEnabled bool
}

// FocusPredictorConfig defines focus predictor configuration
type FocusPredictorConfig struct {
	HistorySize        int           `json:"history_size"`
	PredictionWindow   time.Duration `json:"prediction_window"`
	MinConfidence      float64       `json:"min_confidence"`
	LearningEnabled    bool          `json:"learning_enabled"`
	RetrainingInterval time.Duration `json:"retraining_interval"`
	FeatureWeights     map[string]float64 `json:"feature_weights"`
	AIAssisted         bool          `json:"ai_assisted"`
}

// FocusEvent represents a window focus event
type FocusEvent struct {
	WindowID      string                 `json:"window_id"`
	Application   string                 `json:"application"`
	Timestamp     time.Time              `json:"timestamp"`
	Duration      time.Duration          `json:"duration"`
	PreviousWindow string                `json:"previous_window"`
	Context       map[string]interface{} `json:"context"`
	UserAction    string                 `json:"user_action"` // "click", "keyboard", "alt_tab", "ai_suggestion"
	Workspace     int                    `json:"workspace"`
	Monitor       string                 `json:"monitor"`
}

// FocusPattern represents a learned focus pattern
type FocusPattern struct {
	ID          string                 `json:"id"`
	Sequence    []string               `json:"sequence"`    // Window IDs or application types
	Frequency   int                    `json:"frequency"`
	Confidence  float64                `json:"confidence"`
	Context     map[string]interface{} `json:"context"`
	LastSeen    time.Time              `json:"last_seen"`
	Weight      float64                `json:"weight"`
}

// UserBehaviorProfile represents learned user behavior
type UserBehaviorProfile struct {
	PreferredApplications []string               `json:"preferred_applications"`
	FocusPatterns        map[string]float64     `json:"focus_patterns"`
	TimePreferences      map[string]float64     `json:"time_preferences"`
	ContextPreferences   map[string]interface{} `json:"context_preferences"`
	AverageSessionTime   time.Duration          `json:"average_session_time"`
	MultitaskingLevel    float64                `json:"multitasking_level"`
	LastUpdated          time.Time              `json:"last_updated"`
}

// ContextFeatures represents current context features for prediction
type ContextFeatures struct {
	TimeOfDay        int                    `json:"time_of_day"`
	DayOfWeek        int                    `json:"day_of_week"`
	ActiveWindows    []string               `json:"active_windows"`
	RecentActivity   []string               `json:"recent_activity"`
	KeyboardActivity bool                   `json:"keyboard_activity"`
	MouseActivity    bool                   `json:"mouse_activity"`
	SystemLoad       float64                `json:"system_load"`
	CustomContext    map[string]interface{} `json:"custom_context"`
}

// PredictionModel interface for different prediction algorithms
type PredictionModel interface {
	Predict(context *ContextFeatures, history []FocusEvent) []WindowPrediction
	Train(history []FocusEvent) error
	GetAccuracy() float64
	GetName() string
}

// WindowPrediction represents a window focus prediction
type WindowPrediction struct {
	WindowID    string    `json:"window_id"`
	Application string    `json:"application"`
	Confidence  float64   `json:"confidence"`
	Reasoning   string    `json:"reasoning"`
	Timestamp   time.Time `json:"timestamp"`
}

// PredictionResult represents the result of a prediction
type PredictionResult struct {
	Prediction  WindowPrediction `json:"prediction"`
	Actual      string           `json:"actual"`
	Correct     bool             `json:"correct"`
	Timestamp   time.Time        `json:"timestamp"`
	ContextHash string           `json:"context_hash"`
}

// NewFocusPredictor creates a new focus predictor
func NewFocusPredictor(logger *logrus.Logger, config FocusPredictorConfig, aiOrchestrator *ai.Orchestrator) *FocusPredictor {
	tracer := otel.Tracer("focus-predictor")
	
	predictor := &FocusPredictor{
		logger:         logger,
		tracer:         tracer,
		config:         config,
		aiOrchestrator: aiOrchestrator,
		focusHistory:   make([]FocusEvent, 0),
		patterns:       make([]FocusPattern, 0),
		predictions:    make([]PredictionResult, 0),
		models:         make(map[string]PredictionModel),
		userBehavior: &UserBehaviorProfile{
			PreferredApplications: make([]string, 0),
			FocusPatterns:        make(map[string]float64),
			TimePreferences:      make(map[string]float64),
			ContextPreferences:   make(map[string]interface{}),
			LastUpdated:          time.Now(),
		},
		contextFeatures: &ContextFeatures{
			ActiveWindows:  make([]string, 0),
			RecentActivity: make([]string, 0),
			CustomContext:  make(map[string]interface{}),
		},
		trainingEnabled: config.LearningEnabled,
	}
	
	// Initialize prediction models
	predictor.initializePredictionModels()
	
	return predictor
}

// initializePredictionModels initializes the prediction models
func (fp *FocusPredictor) initializePredictionModels() {
	fp.models["frequency"] = NewFrequencyBasedModel()
	fp.models["pattern"] = NewPatternBasedModel()
	fp.models["temporal"] = NewTemporalModel()
	fp.models["context"] = NewContextBasedModel()
	
	if fp.config.AIAssisted && fp.aiOrchestrator != nil {
		fp.models["ai"] = NewAIAssistedModel(fp.aiOrchestrator)
	}
}

// RecordFocusEvent records a window focus event
func (fp *FocusPredictor) RecordFocusEvent(ctx context.Context, event FocusEvent) error {
	ctx, span := fp.tracer.Start(ctx, "focusPredictor.RecordFocusEvent")
	defer span.End()
	
	fp.mu.Lock()
	defer fp.mu.Unlock()
	
	// Add to history
	fp.focusHistory = append(fp.focusHistory, event)
	
	// Maintain history size
	if len(fp.focusHistory) > fp.config.HistorySize {
		fp.focusHistory = fp.focusHistory[len(fp.focusHistory)-fp.config.HistorySize:]
	}
	
	// Update user behavior profile
	fp.updateUserBehavior(event)
	
	// Learn patterns if enabled
	if fp.trainingEnabled {
		fp.learnPatterns()
	}
	
	// Retrain models periodically
	if time.Since(fp.lastTraining) > fp.config.RetrainingInterval {
		go fp.retrainModels()
	}
	
	fp.logger.WithFields(logrus.Fields{
		"window_id":   event.WindowID,
		"application": event.Application,
		"duration":    event.Duration,
	}).Debug("Focus event recorded")
	
	return nil
}

// PredictNextFocus predicts the next window to be focused
func (fp *FocusPredictor) PredictNextFocus(ctx context.Context) ([]WindowPrediction, error) {
	ctx, span := fp.tracer.Start(ctx, "focusPredictor.PredictNextFocus")
	defer span.End()
	
	fp.mu.RLock()
	defer fp.mu.RUnlock()
	
	// Update context features
	fp.updateContextFeatures()
	
	// Get predictions from all models
	allPredictions := make([]WindowPrediction, 0)
	
	for modelName, model := range fp.models {
		predictions := model.Predict(fp.contextFeatures, fp.focusHistory)
		
		// Weight predictions based on model accuracy
		weight := fp.getModelWeight(modelName)
		for i := range predictions {
			predictions[i].Confidence *= weight
			predictions[i].Reasoning = fmt.Sprintf("%s (%s)", predictions[i].Reasoning, modelName)
		}
		
		allPredictions = append(allPredictions, predictions...)
	}
	
	// Combine and rank predictions
	combinedPredictions := fp.combinePredictions(allPredictions)
	
	// Filter by minimum confidence
	filteredPredictions := make([]WindowPrediction, 0)
	for _, prediction := range combinedPredictions {
		if prediction.Confidence >= fp.config.MinConfidence {
			filteredPredictions = append(filteredPredictions, prediction)
		}
	}
	
	fp.logger.WithField("prediction_count", len(filteredPredictions)).Debug("Focus predictions generated")
	
	return filteredPredictions, nil
}

// updateUserBehavior updates the user behavior profile
func (fp *FocusPredictor) updateUserBehavior(event FocusEvent) {
	// Update preferred applications
	fp.updatePreferredApplications(event.Application)
	
	// Update time preferences
	timeKey := fmt.Sprintf("hour_%d", event.Timestamp.Hour())
	if _, exists := fp.userBehavior.TimePreferences[timeKey]; !exists {
		fp.userBehavior.TimePreferences[timeKey] = 0
	}
	fp.userBehavior.TimePreferences[timeKey]++
	
	// Update focus patterns
	if len(fp.focusHistory) > 1 {
		prevEvent := fp.focusHistory[len(fp.focusHistory)-2]
		patternKey := fmt.Sprintf("%s->%s", prevEvent.Application, event.Application)
		if _, exists := fp.userBehavior.FocusPatterns[patternKey]; !exists {
			fp.userBehavior.FocusPatterns[patternKey] = 0
		}
		fp.userBehavior.FocusPatterns[patternKey]++
	}
	
	fp.userBehavior.LastUpdated = time.Now()
}

// updatePreferredApplications updates the list of preferred applications
func (fp *FocusPredictor) updatePreferredApplications(application string) {
	// Check if application is already in the list
	for i, app := range fp.userBehavior.PreferredApplications {
		if app == application {
			// Move to front (most recently used)
			fp.userBehavior.PreferredApplications = append(
				[]string{application},
				append(fp.userBehavior.PreferredApplications[:i], fp.userBehavior.PreferredApplications[i+1:]...)...,
			)
			return
		}
	}
	
	// Add new application to front
	fp.userBehavior.PreferredApplications = append([]string{application}, fp.userBehavior.PreferredApplications...)
	
	// Limit list size
	if len(fp.userBehavior.PreferredApplications) > 20 {
		fp.userBehavior.PreferredApplications = fp.userBehavior.PreferredApplications[:20]
	}
}

// learnPatterns learns focus patterns from history
func (fp *FocusPredictor) learnPatterns() {
	if len(fp.focusHistory) < 3 {
		return
	}
	
	// Look for sequences of 2-4 windows
	for seqLen := 2; seqLen <= 4 && seqLen <= len(fp.focusHistory); seqLen++ {
		for i := 0; i <= len(fp.focusHistory)-seqLen; i++ {
			sequence := make([]string, seqLen)
			for j := 0; j < seqLen; j++ {
				sequence[j] = fp.focusHistory[i+j].Application
			}
			
			fp.updatePattern(sequence)
		}
	}
}

// updatePattern updates or creates a focus pattern
func (fp *FocusPredictor) updatePattern(sequence []string) {
	sequenceKey := fmt.Sprintf("%v", sequence)
	
	// Find existing pattern
	for i := range fp.patterns {
		if fmt.Sprintf("%v", fp.patterns[i].Sequence) == sequenceKey {
			fp.patterns[i].Frequency++
			fp.patterns[i].LastSeen = time.Now()
			fp.patterns[i].Confidence = fp.calculatePatternConfidence(&fp.patterns[i])
			return
		}
	}
	
	// Create new pattern
	pattern := FocusPattern{
		ID:        fmt.Sprintf("pattern_%d", time.Now().Unix()),
		Sequence:  sequence,
		Frequency: 1,
		LastSeen:  time.Now(),
		Weight:    1.0,
	}
	pattern.Confidence = fp.calculatePatternConfidence(&pattern)
	
	fp.patterns = append(fp.patterns, pattern)
}

// calculatePatternConfidence calculates confidence for a pattern
func (fp *FocusPredictor) calculatePatternConfidence(pattern *FocusPattern) float64 {
	// Simple confidence calculation based on frequency and recency
	recencyFactor := 1.0 - (time.Since(pattern.LastSeen).Hours() / 168.0) // Week-based decay
	if recencyFactor < 0 {
		recencyFactor = 0
	}
	
	frequencyFactor := math.Min(float64(pattern.Frequency)/10.0, 1.0)
	
	return (frequencyFactor + recencyFactor) / 2.0
}

// updateContextFeatures updates current context features
func (fp *FocusPredictor) updateContextFeatures() {
	now := time.Now()
	fp.contextFeatures.TimeOfDay = now.Hour()
	fp.contextFeatures.DayOfWeek = int(now.Weekday())
	
	// Update recent activity from focus history
	if len(fp.focusHistory) > 0 {
		recentCount := minInt(5, len(fp.focusHistory))
		fp.contextFeatures.RecentActivity = make([]string, recentCount)
		for i := 0; i < recentCount; i++ {
			fp.contextFeatures.RecentActivity[i] = fp.focusHistory[len(fp.focusHistory)-1-i].Application
		}
	}
}

// combinePredictions combines predictions from multiple models
func (fp *FocusPredictor) combinePredictions(predictions []WindowPrediction) []WindowPrediction {
	// Group predictions by window ID
	predictionMap := make(map[string][]WindowPrediction)
	for _, pred := range predictions {
		predictionMap[pred.WindowID] = append(predictionMap[pred.WindowID], pred)
	}
	
	// Combine predictions for each window
	combined := make([]WindowPrediction, 0)
	for windowID, preds := range predictionMap {
		if len(preds) == 1 {
			combined = append(combined, preds[0])
		} else {
			// Average confidence and combine reasoning
			totalConfidence := 0.0
			reasons := make([]string, 0)
			
			for _, pred := range preds {
				totalConfidence += pred.Confidence
				reasons = append(reasons, pred.Reasoning)
			}
			
			combinedPred := WindowPrediction{
				WindowID:    windowID,
				Application: preds[0].Application,
				Confidence:  totalConfidence / float64(len(preds)),
				Reasoning:   fmt.Sprintf("Combined: %s", joinStrings(reasons, ", ")),
				Timestamp:   time.Now(),
			}
			
			combined = append(combined, combinedPred)
		}
	}
	
	// Sort by confidence
	sort.Slice(combined, func(i, j int) bool {
		return combined[i].Confidence > combined[j].Confidence
	})
	
	return combined
}

// getModelWeight returns the weight for a prediction model
func (fp *FocusPredictor) getModelWeight(modelName string) float64 {
	if weight, exists := fp.config.FeatureWeights[modelName]; exists {
		return weight
	}
	
	// Default weight based on model accuracy
	if model, exists := fp.models[modelName]; exists {
		return model.GetAccuracy()
	}
	
	return 1.0
}

// retrainModels retrains all prediction models
func (fp *FocusPredictor) retrainModels() {
	fp.mu.Lock()
	defer fp.mu.Unlock()
	
	for _, model := range fp.models {
		if err := model.Train(fp.focusHistory); err != nil {
			fp.logger.WithError(err).WithField("model", model.GetName()).Error("Failed to retrain model")
		}
	}
	
	fp.lastTraining = time.Now()
	fp.logger.Debug("Prediction models retrained")
}

// GetAccuracy returns the current prediction accuracy
func (fp *FocusPredictor) GetAccuracy() float64 {
	fp.mu.RLock()
	defer fp.mu.RUnlock()
	return fp.accuracy
}

// GetUserBehaviorProfile returns the current user behavior profile
func (fp *FocusPredictor) GetUserBehaviorProfile() *UserBehaviorProfile {
	fp.mu.RLock()
	defer fp.mu.RUnlock()
	return fp.userBehavior
}

// Helper functions

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func joinStrings(elems []string, sep string) string {
	// This would normally be from the strings package
	if len(elems) == 0 {
		return ""
	}
	if len(elems) == 1 {
		return elems[0]
	}
	
	result := elems[0]
	for i := 1; i < len(elems); i++ {
		result += sep + elems[i]
	}
	return result
}
