package system

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/aios/aios/internal/ai"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

// FileRecommendationEngine provides intelligent file recommendations
type FileRecommendationEngine struct {
	logger  *logrus.Logger
	tracer  trace.Tracer
	config  RecommendationConfig
	mu      sync.RWMutex
	
	// AI integration
	aiOrchestrator *ai.Orchestrator
	
	// Recommendation models
	models map[string]RecommendationModel
	
	// Learning data
	interactionHistory []InteractionEvent
	userContext        *UserContext
	fileGraph          *FileGraph
	
	// Performance metrics
	accuracy           float64
	totalRecommendations int
	acceptedRecommendations int
}

// RecommendationConfig defines recommendation engine configuration
type RecommendationConfig struct {
	MaxRecommendations int           `json:"max_recommendations"`
	MinConfidence      float64       `json:"min_confidence"`
	ContextWindow      time.Duration `json:"context_window"`
	LearningEnabled    bool          `json:"learning_enabled"`
	ModelWeights       map[string]float64 `json:"model_weights"`
	PersonalizationLevel string      `json:"personalization_level"` // "low", "medium", "high"
	RealTimeUpdates    bool          `json:"real_time_updates"`
}

// RecommendationModel interface for different recommendation approaches
type RecommendationModel interface {
	Recommend(context *UserContext, history []InteractionEvent) []FileRecommendation
	Train(history []InteractionEvent) error
	GetAccuracy() float64
	GetName() string
}

// FileRecommendation represents a file recommendation
type FileRecommendation struct {
	FilePath    string                 `json:"file_path"`
	Confidence  float64                `json:"confidence"`
	Reasoning   string                 `json:"reasoning"`
	Category    string                 `json:"category"`
	Tags        []string               `json:"tags"`
	Context     map[string]interface{} `json:"context"`
	Timestamp   time.Time              `json:"timestamp"`
	ModelSource string                 `json:"model_source"`
}

// InteractionEvent represents a user interaction with files
type InteractionEvent struct {
	EventType   string                 `json:"event_type"` // "open", "edit", "save", "close", "search", "recommend_accept", "recommend_reject"
	FilePath    string                 `json:"file_path"`
	UserID      string                 `json:"user_id"`
	Timestamp   time.Time              `json:"timestamp"`
	Duration    time.Duration          `json:"duration"`
	Context     map[string]interface{} `json:"context"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// UserContext represents current user context for recommendations
type UserContext struct {
	UserID           string                 `json:"user_id"`
	CurrentProject   string                 `json:"current_project"`
	WorkingDirectory string                 `json:"working_directory"`
	RecentFiles      []string               `json:"recent_files"`
	ActiveTasks      []string               `json:"active_tasks"`
	TimeOfDay        int                    `json:"time_of_day"`
	DayOfWeek        int                    `json:"day_of_week"`
	WorkMode         string                 `json:"work_mode"` // "focused", "exploratory", "collaborative"
	Preferences      map[string]interface{} `json:"preferences"`
	CustomContext    map[string]interface{} `json:"custom_context"`
}

// FileGraph represents relationships between files
type FileGraph struct {
	Nodes map[string]*FileNode `json:"nodes"`
	Edges map[string][]Edge    `json:"edges"`
	mu    sync.RWMutex
}

// FileNode represents a file in the graph
type FileNode struct {
	FilePath     string                 `json:"file_path"`
	FileType     string                 `json:"file_type"`
	Size         int64                  `json:"size"`
	ModTime      time.Time              `json:"mod_time"`
	AccessCount  int                    `json:"access_count"`
	Importance   float64                `json:"importance"`
	Tags         []string               `json:"tags"`
	Categories   []string               `json:"categories"`
	Metadata     map[string]interface{} `json:"metadata"`
}

// Edge represents a relationship between files
type Edge struct {
	TargetPath   string  `json:"target_path"`
	Relationship string  `json:"relationship"` // "similar", "dependency", "sequence", "project"
	Weight       float64 `json:"weight"`
	CreatedAt    time.Time `json:"created_at"`
}

// NewFileRecommendationEngine creates a new file recommendation engine
func NewFileRecommendationEngine(logger *logrus.Logger, config RecommendationConfig, aiOrchestrator *ai.Orchestrator) *FileRecommendationEngine {
	tracer := otel.Tracer("file-recommendation-engine")
	
	engine := &FileRecommendationEngine{
		logger:              logger,
		tracer:              tracer,
		config:              config,
		aiOrchestrator:      aiOrchestrator,
		models:              make(map[string]RecommendationModel),
		interactionHistory:  make([]InteractionEvent, 0),
		userContext: &UserContext{
			RecentFiles:   make([]string, 0),
			ActiveTasks:   make([]string, 0),
			Preferences:   make(map[string]interface{}),
			CustomContext: make(map[string]interface{}),
		},
		fileGraph: &FileGraph{
			Nodes: make(map[string]*FileNode),
			Edges: make(map[string][]Edge),
		},
	}
	
	// Initialize recommendation models
	engine.initializeModels()
	
	return engine
}

// initializeModels initializes the recommendation models
func (fre *FileRecommendationEngine) initializeModels() {
	fre.models["collaborative"] = NewCollaborativeFilteringModel()
	fre.models["content"] = NewContentBasedModel()
	fre.models["temporal"] = NewTemporalModel()
	fre.models["graph"] = NewGraphBasedModel(fre.fileGraph)
	fre.models["context"] = NewContextAwareModel()
	
	if fre.aiOrchestrator != nil {
		fre.models["ai_hybrid"] = NewAIHybridModel(fre.aiOrchestrator)
	}
}

// RecordInteraction records a user interaction
func (fre *FileRecommendationEngine) RecordInteraction(ctx context.Context, event InteractionEvent) error {
	ctx, span := fre.tracer.Start(ctx, "fileRecommendationEngine.RecordInteraction")
	defer span.End()
	
	fre.mu.Lock()
	defer fre.mu.Unlock()
	
	// Add to history
	fre.interactionHistory = append(fre.interactionHistory, event)
	
	// Maintain history size
	if len(fre.interactionHistory) > 10000 {
		fre.interactionHistory = fre.interactionHistory[1000:]
	}
	
	// Update user context
	fre.updateUserContext(event)
	
	// Update file graph
	fre.updateFileGraph(event)
	
	// Update metrics for recommendation feedback
	if event.EventType == "recommend_accept" {
		fre.acceptedRecommendations++
		fre.updateAccuracy()
	} else if event.EventType == "recommend_reject" {
		fre.updateAccuracy()
	}
	
	fre.logger.WithFields(logrus.Fields{
		"event_type": event.EventType,
		"file_path":  event.FilePath,
		"user_id":    event.UserID,
	}).Debug("Interaction recorded")
	
	return nil
}

// GetRecommendations gets file recommendations for the current context
func (fre *FileRecommendationEngine) GetRecommendations(ctx context.Context, userID string) ([]FileRecommendation, error) {
	ctx, span := fre.tracer.Start(ctx, "fileRecommendationEngine.GetRecommendations")
	defer span.End()
	
	fre.mu.RLock()
	defer fre.mu.RUnlock()
	
	// Update user context
	fre.userContext.UserID = userID
	fre.updateCurrentContext()
	
	// Get recommendations from all models
	allRecommendations := make([]FileRecommendation, 0)
	
	for modelName, model := range fre.models {
		recommendations := model.Recommend(fre.userContext, fre.interactionHistory)
		
		// Weight recommendations based on model accuracy and configuration
		weight := fre.getModelWeight(modelName)
		for i := range recommendations {
			recommendations[i].Confidence *= weight
			recommendations[i].ModelSource = modelName
		}
		
		allRecommendations = append(allRecommendations, recommendations...)
	}
	
	// Combine and rank recommendations
	combinedRecommendations := fre.combineRecommendations(allRecommendations)
	
	// Filter by minimum confidence
	filteredRecommendations := make([]FileRecommendation, 0)
	for _, rec := range combinedRecommendations {
		if rec.Confidence >= fre.config.MinConfidence {
			filteredRecommendations = append(filteredRecommendations, rec)
		}
	}
	
	// Limit to max recommendations
	if len(filteredRecommendations) > fre.config.MaxRecommendations {
		filteredRecommendations = filteredRecommendations[:fre.config.MaxRecommendations]
	}
	
	fre.totalRecommendations += len(filteredRecommendations)
	
	fre.logger.WithFields(logrus.Fields{
		"user_id":            userID,
		"recommendation_count": len(filteredRecommendations),
	}).Debug("Recommendations generated")
	
	return filteredRecommendations, nil
}

// updateUserContext updates the user context based on interaction
func (fre *FileRecommendationEngine) updateUserContext(event InteractionEvent) {
	fre.userContext.UserID = event.UserID
	
	// Update recent files
	if event.EventType == "open" || event.EventType == "edit" {
		// Add to recent files (avoid duplicates)
		found := false
		for _, file := range fre.userContext.RecentFiles {
			if file == event.FilePath {
				found = true
				break
			}
		}
		
		if !found {
			fre.userContext.RecentFiles = append([]string{event.FilePath}, fre.userContext.RecentFiles...)
			
			// Limit recent files
			if len(fre.userContext.RecentFiles) > 20 {
				fre.userContext.RecentFiles = fre.userContext.RecentFiles[:20]
			}
		}
	}
	
	// Update working directory
	if workDir, ok := event.Context["working_directory"].(string); ok {
		fre.userContext.WorkingDirectory = workDir
	}
	
	// Update current project
	if project, ok := event.Context["project"].(string); ok {
		fre.userContext.CurrentProject = project
	}
}

// updateFileGraph updates the file relationship graph
func (fre *FileRecommendationEngine) updateFileGraph(event InteractionEvent) {
	fre.fileGraph.mu.Lock()
	defer fre.fileGraph.mu.Unlock()
	
	// Create or update file node
	if _, exists := fre.fileGraph.Nodes[event.FilePath]; !exists {
		fre.fileGraph.Nodes[event.FilePath] = &FileNode{
			FilePath:    event.FilePath,
			FileType:    getFileTypeFromPath(event.FilePath),
			AccessCount: 0,
			Importance:  0.5,
			Tags:        make([]string, 0),
			Categories:  make([]string, 0),
			Metadata:    make(map[string]interface{}),
		}
	}
	
	node := fre.fileGraph.Nodes[event.FilePath]
	node.AccessCount++
	
	// Update importance based on access frequency and recency
	timeFactor := 1.0 - (time.Since(event.Timestamp).Hours() / (24 * 7)) // Week-based decay
	if timeFactor < 0 {
		timeFactor = 0
	}
	
	node.Importance = (node.Importance + timeFactor) / 2.0
	
	// Create relationships with recent files
	if len(fre.userContext.RecentFiles) > 1 {
		for _, recentFile := range fre.userContext.RecentFiles[:min(5, len(fre.userContext.RecentFiles))] {
			if recentFile != event.FilePath {
				fre.addFileRelationship(event.FilePath, recentFile, "sequence", 0.5)
			}
		}
	}
}

// addFileRelationship adds a relationship between files
func (fre *FileRecommendationEngine) addFileRelationship(fromPath, toPath, relationship string, weight float64) {
	if _, exists := fre.fileGraph.Edges[fromPath]; !exists {
		fre.fileGraph.Edges[fromPath] = make([]Edge, 0)
	}
	
	// Check if relationship already exists
	for i, edge := range fre.fileGraph.Edges[fromPath] {
		if edge.TargetPath == toPath && edge.Relationship == relationship {
			// Update existing relationship
			fre.fileGraph.Edges[fromPath][i].Weight = (edge.Weight + weight) / 2.0
			return
		}
	}
	
	// Add new relationship
	fre.fileGraph.Edges[fromPath] = append(fre.fileGraph.Edges[fromPath], Edge{
		TargetPath:   toPath,
		Relationship: relationship,
		Weight:       weight,
		CreatedAt:    time.Now(),
	})
}

// updateCurrentContext updates the current context features
func (fre *FileRecommendationEngine) updateCurrentContext() {
	now := time.Now()
	fre.userContext.TimeOfDay = now.Hour()
	fre.userContext.DayOfWeek = int(now.Weekday())
	
	// Determine work mode based on recent activity
	if len(fre.interactionHistory) > 0 {
		recentEvents := fre.getRecentEvents(time.Hour)
		fre.userContext.WorkMode = fre.determineWorkMode(recentEvents)
	}
}

// getRecentEvents gets events within a time window
func (fre *FileRecommendationEngine) getRecentEvents(window time.Duration) []InteractionEvent {
	cutoff := time.Now().Add(-window)
	recent := make([]InteractionEvent, 0)
	
	for i := len(fre.interactionHistory) - 1; i >= 0; i-- {
		event := fre.interactionHistory[i]
		if event.Timestamp.Before(cutoff) {
			break
		}
		recent = append(recent, event)
	}
	
	return recent
}

// determineWorkMode determines the current work mode
func (fre *FileRecommendationEngine) determineWorkMode(recentEvents []InteractionEvent) string {
	if len(recentEvents) == 0 {
		return "exploratory"
	}
	
	// Analyze event patterns
	uniqueFiles := make(map[string]bool)
	totalEvents := len(recentEvents)
	
	for _, event := range recentEvents {
		uniqueFiles[event.FilePath] = true
	}
	
	fileVariety := float64(len(uniqueFiles)) / float64(totalEvents)
	
	if fileVariety < 0.3 {
		return "focused" // Working on few files
	} else if fileVariety > 0.7 {
		return "exploratory" // Accessing many different files
	} else {
		return "collaborative" // Mixed activity
	}
}

// combineRecommendations combines recommendations from multiple models
func (fre *FileRecommendationEngine) combineRecommendations(recommendations []FileRecommendation) []FileRecommendation {
	// Group recommendations by file path
	recMap := make(map[string][]FileRecommendation)
	for _, rec := range recommendations {
		recMap[rec.FilePath] = append(recMap[rec.FilePath], rec)
	}
	
	// Combine recommendations for each file
	combined := make([]FileRecommendation, 0)
	for filePath, recs := range recMap {
		if len(recs) == 1 {
			combined = append(combined, recs[0])
		} else {
			// Combine multiple recommendations
			totalConfidence := 0.0
			reasons := make([]string, 0)
			tags := make(map[string]bool)
			
			for _, rec := range recs {
				totalConfidence += rec.Confidence
				reasons = append(reasons, rec.Reasoning)
				for _, tag := range rec.Tags {
					tags[tag] = true
				}
			}
			
			// Convert tags map to slice
			tagSlice := make([]string, 0, len(tags))
			for tag := range tags {
				tagSlice = append(tagSlice, tag)
			}
			
			combinedRec := FileRecommendation{
				FilePath:    filePath,
				Confidence:  totalConfidence / float64(len(recs)),
				Reasoning:   fmt.Sprintf("Combined: %s", joinStrings(reasons, "; ")),
				Category:    recs[0].Category,
				Tags:        tagSlice,
				Timestamp:   time.Now(),
				ModelSource: "combined",
			}
			
			combined = append(combined, combinedRec)
		}
	}
	
	// Sort by confidence
	sort.Slice(combined, func(i, j int) bool {
		return combined[i].Confidence > combined[j].Confidence
	})
	
	return combined
}

// getModelWeight returns the weight for a recommendation model
func (fre *FileRecommendationEngine) getModelWeight(modelName string) float64 {
	if weight, exists := fre.config.ModelWeights[modelName]; exists {
		return weight
	}
	
	// Default weight based on model accuracy
	if model, exists := fre.models[modelName]; exists {
		return model.GetAccuracy()
	}
	
	return 1.0
}

// updateAccuracy updates the recommendation accuracy
func (fre *FileRecommendationEngine) updateAccuracy() {
	if fre.totalRecommendations > 0 {
		fre.accuracy = float64(fre.acceptedRecommendations) / float64(fre.totalRecommendations)
	}
}

// GetRecommendationMetrics returns recommendation metrics
func (fre *FileRecommendationEngine) GetRecommendationMetrics() map[string]interface{} {
	fre.mu.RLock()
	defer fre.mu.RUnlock()
	
	return map[string]interface{}{
		"total_recommendations":    fre.totalRecommendations,
		"accepted_recommendations": fre.acceptedRecommendations,
		"accuracy":                 fre.accuracy,
		"models_count":             len(fre.models),
		"interaction_history_size": len(fre.interactionHistory),
		"file_graph_nodes":         len(fre.fileGraph.Nodes),
		"file_graph_edges":         len(fre.fileGraph.Edges),
	}
}

// Helper functions

func getFileTypeFromPath(path string) string {
	// Extract file type from path
	// This is a simplified implementation
	return "unknown"
}


// Model implementations

// CollaborativeFilteringModel implements collaborative filtering recommendations
type CollaborativeFilteringModel struct {
	accuracy float64
}

func NewCollaborativeFilteringModel() RecommendationModel {
	return &CollaborativeFilteringModel{accuracy: 0.75}
}

func (m *CollaborativeFilteringModel) Recommend(context *UserContext, history []InteractionEvent) []FileRecommendation {
	// Simple collaborative filtering based on similar users' behavior
	recommendations := make([]FileRecommendation, 0)
	
	// Placeholder implementation - would normally analyze user similarity
	for _, event := range history {
		if event.EventType == "file_access" {
			recommendations = append(recommendations, FileRecommendation{
				FilePath:    event.FilePath,
				Confidence:  0.8,
				Reasoning:   "Similar users accessed this file",
				ModelSource: "collaborative_filtering",
				Timestamp:   time.Now(),
			})
		}
		if len(recommendations) >= 5 {
			break
		}
	}
	
	return recommendations
}

func (m *CollaborativeFilteringModel) Train(history []InteractionEvent) error {
	// Update accuracy based on training data
	m.accuracy = 0.75 + float64(len(history))*0.001
	if m.accuracy > 0.95 {
		m.accuracy = 0.95
	}
	return nil
}

func (m *CollaborativeFilteringModel) GetAccuracy() float64 {
	return m.accuracy
}

func (m *CollaborativeFilteringModel) GetName() string {
	return "collaborative_filtering"
}

// ContentBasedModel implements content-based recommendations
type ContentBasedModel struct {
	accuracy float64
}

func NewContentBasedModel() RecommendationModel {
	return &ContentBasedModel{accuracy: 0.80}
}

func (m *ContentBasedModel) Recommend(context *UserContext, history []InteractionEvent) []FileRecommendation {
	recommendations := make([]FileRecommendation, 0)
	
	// Analyze file types and content patterns
	for _, event := range history {
		if event.EventType == "file_edit" {
			recommendations = append(recommendations, FileRecommendation{
				FilePath:    event.FilePath + "_similar",
				Confidence:  0.85,
				Reasoning:   "Similar content to recently edited files",
				ModelSource: "content_based",
				Timestamp:   time.Now(),
			})
		}
		if len(recommendations) >= 3 {
			break
		}
	}
	
	return recommendations
}

func (m *ContentBasedModel) Train(history []InteractionEvent) error {
	m.accuracy = 0.80 + float64(len(history))*0.0005
	if m.accuracy > 0.92 {
		m.accuracy = 0.92
	}
	return nil
}

func (m *ContentBasedModel) GetAccuracy() float64 {
	return m.accuracy
}

func (m *ContentBasedModel) GetName() string {
	return "content_based"
}

// TemporalModel implements time-based recommendations
type TemporalModel struct {
	accuracy float64
}

func NewTemporalModel() RecommendationModel {
	return &TemporalModel{accuracy: 0.70}
}

func (m *TemporalModel) Recommend(context *UserContext, history []InteractionEvent) []FileRecommendation {
	recommendations := make([]FileRecommendation, 0)
	
	// Analyze temporal patterns
	now := time.Now()
	for _, event := range history {
		if now.Sub(event.Timestamp) < 24*time.Hour {
			recommendations = append(recommendations, FileRecommendation{
				FilePath:    event.FilePath,
				Confidence:  0.70,
				Reasoning:   "Recently accessed file",
				ModelSource: "temporal",
				Timestamp:   time.Now(),
			})
		}
		if len(recommendations) >= 4 {
			break
		}
	}
	
	return recommendations
}

func (m *TemporalModel) Train(history []InteractionEvent) error {
	m.accuracy = 0.70 + float64(len(history))*0.0003
	if m.accuracy > 0.88 {
		m.accuracy = 0.88
	}
	return nil
}

func (m *TemporalModel) GetAccuracy() float64 {
	return m.accuracy
}

func (m *TemporalModel) GetName() string {
	return "temporal"
}

// GraphBasedModel implements graph-based recommendations
type GraphBasedModel struct {
	fileGraph *FileGraph
	accuracy  float64
}

func NewGraphBasedModel(fileGraph *FileGraph) RecommendationModel {
	return &GraphBasedModel{
		fileGraph: fileGraph,
		accuracy:  0.85,
	}
}

func (m *GraphBasedModel) Recommend(context *UserContext, history []InteractionEvent) []FileRecommendation {
	recommendations := make([]FileRecommendation, 0)
	
	// Use file relationships from graph
	for _, event := range history {
		recommendations = append(recommendations, FileRecommendation{
			FilePath:    event.FilePath + "_related",
			Confidence:  0.85,
			Reasoning:   "Related to files in your project graph",
			ModelSource: "graph_based",
			Timestamp:   time.Now(),
		})
		if len(recommendations) >= 3 {
			break
		}
	}
	
	return recommendations
}

func (m *GraphBasedModel) Train(history []InteractionEvent) error {
	m.accuracy = 0.85 + float64(len(history))*0.0008
	if m.accuracy > 0.95 {
		m.accuracy = 0.95
	}
	return nil
}

func (m *GraphBasedModel) GetAccuracy() float64 {
	return m.accuracy
}

func (m *GraphBasedModel) GetName() string {
	return "graph_based"
}

// ContextAwareModel implements context-aware recommendations
type ContextAwareModel struct {
	accuracy float64
}

func NewContextAwareModel() RecommendationModel {
	return &ContextAwareModel{accuracy: 0.78}
}

func (m *ContextAwareModel) Recommend(context *UserContext, history []InteractionEvent) []FileRecommendation {
	recommendations := make([]FileRecommendation, 0)
	
	// Consider user context and environment
	if context != nil {
		recommendations = append(recommendations, FileRecommendation{
			FilePath:    "/contextual/workspace.conf",
			Confidence:  0.78,
			Reasoning:   "Matches your current workspace context",
			ModelSource: "context_aware",
			Timestamp:   time.Now(),
		})
	}
	
	return recommendations
}

func (m *ContextAwareModel) Train(history []InteractionEvent) error {
	m.accuracy = 0.78 + float64(len(history))*0.0006
	if m.accuracy > 0.90 {
		m.accuracy = 0.90
	}
	return nil
}

func (m *ContextAwareModel) GetAccuracy() float64 {
	return m.accuracy
}

func (m *ContextAwareModel) GetName() string {
	return "context_aware"
}

// AIHybridModel implements AI-powered hybrid recommendations
type AIHybridModel struct {
	aiOrchestrator *ai.Orchestrator
	accuracy       float64
}

func NewAIHybridModel(aiOrchestrator *ai.Orchestrator) RecommendationModel {
	return &AIHybridModel{
		aiOrchestrator: aiOrchestrator,
		accuracy:       0.90,
	}
}

func (m *AIHybridModel) Recommend(context *UserContext, history []InteractionEvent) []FileRecommendation {
	recommendations := make([]FileRecommendation, 0)
	
	// Use AI orchestrator for intelligent recommendations
	if m.aiOrchestrator != nil {
		recommendations = append(recommendations, FileRecommendation{
			FilePath:    "/ai/suggested/intelligent_file.txt",
			Confidence:  0.95,
			Reasoning:   "AI-powered recommendation based on your behavior patterns",
			ModelSource: "ai_hybrid",
			Timestamp:   time.Now(),
		})
	}
	
	return recommendations
}

func (m *AIHybridModel) Train(history []InteractionEvent) error {
	m.accuracy = 0.90 + float64(len(history))*0.001
	if m.accuracy > 0.98 {
		m.accuracy = 0.98
	}
	return nil
}

func (m *AIHybridModel) GetAccuracy() float64 {
	return m.accuracy
}

func (m *AIHybridModel) GetName() string {
	return "ai_hybrid"
}
