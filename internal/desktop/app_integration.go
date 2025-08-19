package desktop

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
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

// AppIntegration provides AI-powered application management and file operations
type AppIntegration struct {
	logger     *logrus.Logger
	tracer     trace.Tracer
	nlpService ai.NaturalLanguageService
	llmService ai.LanguageModelService
	cvService  ai.ComputerVisionService

	// Application management
	installedApps map[string]*Application
	runningApps   map[string]*RunningApplication
	appCategories map[string][]string
	mu            sync.RWMutex

	// File system integration
	fileIndex   map[string]*FileInfo
	recentFiles []*FileInfo
	bookmarks   []*Bookmark

	// Configuration
	config AppIntegrationConfig
}

// AppIntegrationConfig represents configuration for app integration
type AppIntegrationConfig struct {
	AppSearchPaths          []string      `json:"app_search_paths"`
	FileIndexPaths          []string      `json:"file_index_paths"`
	MaxRecentFiles          int           `json:"max_recent_files"`
	IndexUpdateInterval     time.Duration `json:"index_update_interval"`
	EnableSmartSuggestions  bool          `json:"enable_smart_suggestions"`
	EnableContextualActions bool          `json:"enable_contextual_actions"`
}

// Application represents an installed application
type Application struct {
	ID             string                 `json:"id"`
	Name           string                 `json:"name"`
	DisplayName    string                 `json:"display_name"`
	Description    string                 `json:"description"`
	ExecutablePath string                 `json:"executable_path"`
	IconPath       string                 `json:"icon_path"`
	Categories     []string               `json:"categories"`
	Keywords       []string               `json:"keywords"`
	Version        string                 `json:"version"`
	InstallDate    time.Time              `json:"install_date"`
	LastUsed       *time.Time             `json:"last_used,omitempty"`
	UsageCount     int                    `json:"usage_count"`
	Rating         float64                `json:"rating"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

// RunningApplication represents a currently running application
type RunningApplication struct {
	App         *Application           `json:"app"`
	ProcessID   int                    `json:"process_id"`
	WindowID    string                 `json:"window_id,omitempty"`
	StartTime   time.Time              `json:"start_time"`
	MemoryUsage int64                  `json:"memory_usage"`
	CPUUsage    float64                `json:"cpu_usage"`
	Status      string                 `json:"status"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// FileInfo represents file system information
type FileInfo struct {
	Path         string                 `json:"path"`
	Name         string                 `json:"name"`
	Size         int64                  `json:"size"`
	ModTime      time.Time              `json:"mod_time"`
	IsDirectory  bool                   `json:"is_directory"`
	MimeType     string                 `json:"mime_type"`
	Tags         []string               `json:"tags"`
	Description  string                 `json:"description,omitempty"`
	Thumbnail    string                 `json:"thumbnail,omitempty"`
	AccessCount  int                    `json:"access_count"`
	LastAccessed *time.Time             `json:"last_accessed,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// Bookmark represents a file system bookmark
type Bookmark struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Path        string                 `json:"path"`
	Description string                 `json:"description,omitempty"`
	Tags        []string               `json:"tags"`
	CreatedAt   time.Time              `json:"created_at"`
	AccessCount int                    `json:"access_count"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// SmartSuggestion represents an AI-generated suggestion
type SmartSuggestion struct {
	Type        string                 `json:"type"` // "app", "file", "action", "workflow"
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Action      string                 `json:"action"`
	Target      string                 `json:"target"`
	Confidence  float64                `json:"confidence"`
	Reasoning   string                 `json:"reasoning"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// NewAppIntegration creates a new application integration instance
func NewAppIntegration(
	logger *logrus.Logger,
	nlpService ai.NaturalLanguageService,
	llmService ai.LanguageModelService,
	cvService ai.ComputerVisionService,
	config AppIntegrationConfig,
) *AppIntegration {
	ai := &AppIntegration{
		logger:        logger,
		tracer:        otel.Tracer("desktop.app_integration"),
		nlpService:    nlpService,
		llmService:    llmService,
		cvService:     cvService,
		installedApps: make(map[string]*Application),
		runningApps:   make(map[string]*RunningApplication),
		appCategories: make(map[string][]string),
		fileIndex:     make(map[string]*FileInfo),
		recentFiles:   make([]*FileInfo, 0),
		bookmarks:     make([]*Bookmark, 0),
		config:        config,
	}

	// Initialize application discovery
	go ai.discoverApplications()

	// Initialize file indexing
	if config.IndexUpdateInterval > 0 {
		go ai.startFileIndexing()
	}

	return ai
}

// LaunchApplication launches an application by name or intent
func (ai *AppIntegration) LaunchApplication(ctx context.Context, query string) (*RunningApplication, error) {
	ctx, span := ai.tracer.Start(ctx, "app_integration.LaunchApplication")
	defer span.End()

	ai.logger.WithField("query", query).Info("Launching application")

	// Parse intent to find the application
	intent, err := ai.nlpService.ParseIntent(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to parse intent: %w", err)
	}

	// Find the best matching application
	app := ai.findBestApplication(intent)
	if app == nil {
		return nil, fmt.Errorf("no suitable application found for: %s", query)
	}

	// Launch the application
	cmd := exec.CommandContext(ctx, app.ExecutablePath)
	err = cmd.Start()
	if err != nil {
		return nil, fmt.Errorf("failed to launch %s: %w", app.Name, err)
	}

	// Create running application record
	runningApp := &RunningApplication{
		App:       app,
		ProcessID: cmd.Process.Pid,
		StartTime: time.Now(),
		Status:    "running",
		Metadata:  make(map[string]interface{}),
	}

	ai.mu.Lock()
	ai.runningApps[app.ID] = runningApp
	app.UsageCount++
	now := time.Now()
	app.LastUsed = &now
	ai.mu.Unlock()

	ai.logger.WithFields(logrus.Fields{
		"app_name":   app.Name,
		"process_id": cmd.Process.Pid,
	}).Info("Application launched successfully")

	return runningApp, nil
}

// SearchApplications searches for applications using natural language
func (ai *AppIntegration) SearchApplications(ctx context.Context, query string) ([]*Application, error) {
	ctx, span := ai.tracer.Start(ctx, "app_integration.SearchApplications")
	defer span.End()

	ai.mu.RLock()
	defer ai.mu.RUnlock()

	var results []*Application
	queryLower := strings.ToLower(query)

	// Score applications based on relevance
	type appScore struct {
		app   *Application
		score float64
	}

	var scored []appScore

	for _, app := range ai.installedApps {
		score := ai.calculateAppRelevance(app, queryLower)
		if score > 0 {
			scored = append(scored, appScore{app: app, score: score})
		}
	}

	// Sort by score (descending)
	sort.Slice(scored, func(i, j int) bool {
		return scored[i].score > scored[j].score
	})

	// Extract applications
	for _, item := range scored {
		results = append(results, item.app)
	}

	ai.logger.WithFields(logrus.Fields{
		"query":         query,
		"results_count": len(results),
	}).Info("Application search completed")

	return results, nil
}

// SearchFiles searches for files using natural language
func (ai *AppIntegration) SearchFiles(ctx context.Context, query string) ([]*FileInfo, error) {
	ctx, span := ai.tracer.Start(ctx, "app_integration.SearchFiles")
	defer span.End()

	ai.mu.RLock()
	defer ai.mu.RUnlock()

	var results []*FileInfo
	queryLower := strings.ToLower(query)

	// Score files based on relevance
	type fileScore struct {
		file  *FileInfo
		score float64
	}

	var scored []fileScore

	for _, file := range ai.fileIndex {
		score := ai.calculateFileRelevance(file, queryLower)
		if score > 0 {
			scored = append(scored, fileScore{file: file, score: score})
		}
	}

	// Sort by score (descending)
	sort.Slice(scored, func(i, j int) bool {
		return scored[i].score > scored[j].score
	})

	// Extract files
	for _, item := range scored {
		results = append(results, item.file)
	}

	ai.logger.WithFields(logrus.Fields{
		"query":         query,
		"results_count": len(results),
	}).Info("File search completed")

	return results, nil
}

// GetSmartSuggestions generates AI-powered suggestions based on context
func (ai *AppIntegration) GetSmartSuggestions(ctx context.Context, context map[string]interface{}) ([]*SmartSuggestion, error) {
	ctx, span := ai.tracer.Start(ctx, "app_integration.GetSmartSuggestions")
	defer span.End()

	if !ai.config.EnableSmartSuggestions {
		return []*SmartSuggestion{}, nil
	}

	var suggestions []*SmartSuggestion

	// Analyze current context
	contextPrompt := ai.buildContextPrompt(context)

	// Generate suggestions using LLM
	response, err := ai.llmService.ProcessQuery(ctx, fmt.Sprintf(
		"Based on the current context: %s\nSuggest 3-5 relevant applications, files, or actions the user might want to perform. Format as JSON array with type, title, description, action, confidence fields.",
		contextPrompt,
	))

	if err != nil {
		ai.logger.WithError(err).Error("Failed to generate smart suggestions")
		return ai.getFallbackSuggestions(), nil
	}

	// Parse LLM response (simplified - would need proper JSON parsing)
	suggestions = ai.parseSuggestionsFromLLM(response.Text)

	// Add usage-based suggestions
	suggestions = append(suggestions, ai.getUsageBasedSuggestions()...)

	// Add time-based suggestions
	suggestions = append(suggestions, ai.getTimeBasedSuggestions()...)

	ai.logger.WithField("suggestions_count", len(suggestions)).Info("Smart suggestions generated")

	return suggestions, nil
}

// GetRunningApplications returns currently running applications
func (ai *AppIntegration) GetRunningApplications(ctx context.Context) ([]*RunningApplication, error) {
	ctx, span := ai.tracer.Start(ctx, "app_integration.GetRunningApplications")
	defer span.End()

	ai.mu.RLock()
	defer ai.mu.RUnlock()

	var running []*RunningApplication
	for _, app := range ai.runningApps {
		// Create a copy
		appCopy := *app
		running = append(running, &appCopy)
	}

	return running, nil
}

// Helper methods

func (ai *AppIntegration) discoverApplications() {
	ai.logger.Info("Starting application discovery")

	for _, searchPath := range ai.config.AppSearchPaths {
		ai.scanApplicationPath(searchPath)
	}

	ai.logger.WithField("apps_found", len(ai.installedApps)).Info("Application discovery completed")
}

func (ai *AppIntegration) scanApplicationPath(path string) {
	err := filepath.Walk(path, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Continue on errors
		}

		// Look for executable files or .desktop files
		if ai.isApplicationFile(filePath, info) {
			app := ai.createApplicationFromFile(filePath, info)
			if app != nil {
				ai.mu.Lock()
				ai.installedApps[app.ID] = app
				ai.mu.Unlock()
			}
		}

		return nil
	})

	if err != nil {
		ai.logger.WithError(err).WithField("path", path).Error("Failed to scan application path")
	}
}

func (ai *AppIntegration) isApplicationFile(path string, info os.FileInfo) bool {
	if info.IsDir() {
		return false
	}

	// Check for executable files
	if info.Mode()&0111 != 0 {
		return true
	}

	// Check for .desktop files (Linux)
	if strings.HasSuffix(path, ".desktop") {
		return true
	}

	// Check for .app bundles (macOS)
	if strings.HasSuffix(path, ".app") {
		return true
	}

	return false
}

func (ai *AppIntegration) createApplicationFromFile(path string, info os.FileInfo) *Application {
	name := strings.TrimSuffix(info.Name(), filepath.Ext(info.Name()))

	return &Application{
		ID:             generateAppID(path),
		Name:           name,
		DisplayName:    name,
		ExecutablePath: path,
		InstallDate:    info.ModTime(),
		UsageCount:     0,
		Rating:         0.0,
		Categories:     []string{"Other"},
		Keywords:       []string{name},
		Metadata:       make(map[string]interface{}),
	}
}

func (ai *AppIntegration) findBestApplication(intent *models.IntentAnalysis) *Application {
	ai.mu.RLock()
	defer ai.mu.RUnlock()

	var bestApp *Application
	var bestScore float64

	// Extract application name from entities
	var targetApp string
	for _, entity := range intent.Entities {
		if entity.Type == "application" || entity.Type == "target" {
			targetApp = entity.Text
			break
		}
	}

	if targetApp == "" {
		return nil
	}

	// Find best matching application
	for _, app := range ai.installedApps {
		score := ai.calculateAppRelevance(app, strings.ToLower(targetApp))
		if score > bestScore {
			bestScore = score
			bestApp = app
		}
	}

	return bestApp
}

func (ai *AppIntegration) calculateAppRelevance(app *Application, query string) float64 {
	score := 0.0

	// Exact name match
	if strings.ToLower(app.Name) == query {
		score += 100.0
	}

	// Name contains query
	if strings.Contains(strings.ToLower(app.Name), query) {
		score += 50.0
	}

	// Display name contains query
	if strings.Contains(strings.ToLower(app.DisplayName), query) {
		score += 40.0
	}

	// Keywords match
	for _, keyword := range app.Keywords {
		if strings.Contains(strings.ToLower(keyword), query) {
			score += 20.0
		}
	}

	// Usage frequency bonus
	score += float64(app.UsageCount) * 0.1

	// Recent usage bonus
	if app.LastUsed != nil && time.Since(*app.LastUsed) < 24*time.Hour {
		score += 10.0
	}

	return score
}

func (ai *AppIntegration) calculateFileRelevance(file *FileInfo, query string) float64 {
	score := 0.0

	// Exact name match
	if strings.ToLower(file.Name) == query {
		score += 100.0
	}

	// Name contains query
	if strings.Contains(strings.ToLower(file.Name), query) {
		score += 50.0
	}

	// Path contains query
	if strings.Contains(strings.ToLower(file.Path), query) {
		score += 30.0
	}

	// Tags match
	for _, tag := range file.Tags {
		if strings.Contains(strings.ToLower(tag), query) {
			score += 20.0
		}
	}

	// Access frequency bonus
	score += float64(file.AccessCount) * 0.1

	// Recent access bonus
	if file.LastAccessed != nil && time.Since(*file.LastAccessed) < 24*time.Hour {
		score += 10.0
	}

	return score
}

func (ai *AppIntegration) startFileIndexing() {
	ticker := time.NewTicker(ai.config.IndexUpdateInterval)
	defer ticker.Stop()

	// Initial indexing
	ai.updateFileIndex()

	for {
		select {
		case <-ticker.C:
			ai.updateFileIndex()
		}
	}
}

func (ai *AppIntegration) updateFileIndex() {
	ai.logger.Info("Updating file index")

	for _, indexPath := range ai.config.FileIndexPaths {
		ai.scanFilePath(indexPath)
	}

	ai.logger.WithField("files_indexed", len(ai.fileIndex)).Info("File index updated")
}

func (ai *AppIntegration) scanFilePath(path string) {
	err := filepath.Walk(path, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Continue on errors
		}

		fileInfo := &FileInfo{
			Path:        filePath,
			Name:        info.Name(),
			Size:        info.Size(),
			ModTime:     info.ModTime(),
			IsDirectory: info.IsDir(),
			Tags:        []string{},
			AccessCount: 0,
			Metadata:    make(map[string]interface{}),
		}

		ai.mu.Lock()
		ai.fileIndex[filePath] = fileInfo
		ai.mu.Unlock()

		return nil
	})

	if err != nil {
		ai.logger.WithError(err).WithField("path", path).Error("Failed to scan file path")
	}
}

func (ai *AppIntegration) buildContextPrompt(context map[string]interface{}) string {
	var prompt strings.Builder

	prompt.WriteString("Current context: ")

	if currentTime, ok := context["time"]; ok {
		prompt.WriteString(fmt.Sprintf("Time: %v. ", currentTime))
	}

	if currentDir, ok := context["current_directory"]; ok {
		prompt.WriteString(fmt.Sprintf("Directory: %v. ", currentDir))
	}

	if recentApps, ok := context["recent_applications"]; ok {
		prompt.WriteString(fmt.Sprintf("Recent apps: %v. ", recentApps))
	}

	return prompt.String()
}

func (ai *AppIntegration) parseSuggestionsFromLLM(response string) []*SmartSuggestion {
	// TODO: Implement proper JSON parsing of LLM response
	// This is a simplified mock implementation
	return []*SmartSuggestion{
		{
			Type:        "app",
			Title:       "Open Text Editor",
			Description: "Based on your recent file activity",
			Action:      "launch_app",
			Target:      "text_editor",
			Confidence:  0.85,
			Reasoning:   "User has been working with text files",
		},
	}
}

func (ai *AppIntegration) getFallbackSuggestions() []*SmartSuggestion {
	return []*SmartSuggestion{
		{
			Type:        "app",
			Title:       "File Manager",
			Description: "Browse your files",
			Action:      "launch_app",
			Target:      "file_manager",
			Confidence:  0.7,
			Reasoning:   "General utility",
		},
	}
}

func (ai *AppIntegration) getUsageBasedSuggestions() []*SmartSuggestion {
	// TODO: Implement usage-based suggestions
	return []*SmartSuggestion{}
}

func (ai *AppIntegration) getTimeBasedSuggestions() []*SmartSuggestion {
	// TODO: Implement time-based suggestions
	return []*SmartSuggestion{}
}

func generateAppID(path string) string {
	return fmt.Sprintf("app_%x", []byte(path))
}
