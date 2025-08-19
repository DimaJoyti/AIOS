package system

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/aios/aios/internal/ai"
	"github.com/aios/aios/pkg/models"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

// FileOrganizationEngine provides intelligent file organization
type FileOrganizationEngine struct {
	logger *logrus.Logger
	tracer trace.Tracer
	config OrganizationConfig
	mu     sync.RWMutex

	// AI integration
	aiOrchestrator *ai.Orchestrator

	// Organization strategies
	strategies map[string]OrganizationStrategy

	// Learning data
	organizationHistory []OrganizationEvent
	userPreferences     *OrganizationPreferences
	categoryRules       []CategoryRule

	// Performance metrics
	successRate        float64
	totalOrganizations int
	successfulOrgs     int
}

// OrganizationConfig defines organization engine configuration
type OrganizationConfig struct {
	DefaultStrategy    string         `json:"default_strategy"`
	AutoOrganize       bool           `json:"auto_organize"`
	BackupBeforeMove   bool           `json:"backup_before_move"`
	ConflictResolution string         `json:"conflict_resolution"` // "skip", "rename", "overwrite"
	CategoryThreshold  float64        `json:"category_threshold"`
	LearningEnabled    bool           `json:"learning_enabled"`
	CustomRules        []CategoryRule `json:"custom_rules"`
	AIAssisted         bool           `json:"ai_assisted"`
}

// OrganizationStrategy interface for different organization approaches
type OrganizationStrategy interface {
	Organize(files []FileInfo, targetDir string) (*OrganizationPlan, error)
	GetName() string
	GetDescription() string
	SupportsLearning() bool
}

// OrganizationPlan represents a file organization plan
type OrganizationPlan struct {
	ID            string          `json:"id"`
	Strategy      string          `json:"strategy"`
	SourceDir     string          `json:"source_dir"`
	TargetDir     string          `json:"target_dir"`
	Operations    []FileOperation `json:"operations"`
	Confidence    float64         `json:"confidence"`
	Reasoning     string          `json:"reasoning"`
	CreatedAt     time.Time       `json:"created_at"`
	EstimatedTime time.Duration   `json:"estimated_time"`
	RiskLevel     string          `json:"risk_level"` // "low", "medium", "high"
}

// FileOperation represents a single file operation
type FileOperation struct {
	Type         string                 `json:"type"` // "move", "copy", "rename", "delete", "create_dir"
	SourcePath   string                 `json:"source_path"`
	TargetPath   string                 `json:"target_path"`
	Confidence   float64                `json:"confidence"`
	Reasoning    string                 `json:"reasoning"`
	Dependencies []string               `json:"dependencies"`
	Metadata     map[string]interface{} `json:"metadata"`
}

// OrganizationEvent represents an organization event
type OrganizationEvent struct {
	PlanID       string                 `json:"plan_id"`
	Strategy     string                 `json:"strategy"`
	Timestamp    time.Time              `json:"timestamp"`
	Duration     time.Duration          `json:"duration"`
	FilesCount   int                    `json:"files_count"`
	Success      bool                   `json:"success"`
	Error        string                 `json:"error,omitempty"`
	UserFeedback string                 `json:"user_feedback,omitempty"`
	Context      map[string]interface{} `json:"context"`
}

// OrganizationPreferences stores user organization preferences
type OrganizationPreferences struct {
	PreferredStrategy   string                 `json:"preferred_strategy"`
	CategoryPreferences map[string]string      `json:"category_preferences"`
	FolderStructure     map[string]interface{} `json:"folder_structure"`
	NamingConventions   map[string]string      `json:"naming_conventions"`
	AutoRules           []AutoRule             `json:"auto_rules"`
	LastUpdated         time.Time              `json:"last_updated"`
}

// CategoryRule defines how files should be categorized
type CategoryRule struct {
	ID          string          `json:"id"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Conditions  []RuleCondition `json:"conditions"`
	Actions     []RuleAction    `json:"actions"`
	Priority    int             `json:"priority"`
	Enabled     bool            `json:"enabled"`
	CreatedAt   time.Time       `json:"created_at"`
	UsageCount  int             `json:"usage_count"`
}

// RuleCondition defines when a rule should apply
type RuleCondition struct {
	Type     string      `json:"type"`     // "extension", "size", "name", "content", "metadata"
	Operator string      `json:"operator"` // "equals", "contains", "matches", "greater_than"
	Value    interface{} `json:"value"`
	Weight   float64     `json:"weight"`
}

// RuleAction defines what should happen when a rule matches
type RuleAction struct {
	Type       string                 `json:"type"` // "move_to", "rename", "tag", "categorize"
	Parameters map[string]interface{} `json:"parameters"`
}

// AutoRule defines automatic organization rules
type AutoRule struct {
	ID         string          `json:"id"`
	Name       string          `json:"name"`
	Trigger    string          `json:"trigger"` // "file_created", "file_modified", "periodic"
	Conditions []RuleCondition `json:"conditions"`
	Actions    []RuleAction    `json:"actions"`
	Enabled    bool            `json:"enabled"`
	LastRun    time.Time       `json:"last_run"`
	RunCount   int             `json:"run_count"`
}

// FileInfo represents file information for organization
type FileInfo struct {
	Path          string                 `json:"path"`
	Name          string                 `json:"name"`
	Extension     string                 `json:"extension"`
	Size          int64                  `json:"size"`
	ModTime       time.Time              `json:"mod_time"`
	AccessTime    time.Time              `json:"access_time"`
	MimeType      string                 `json:"mime_type"`
	Category      string                 `json:"category"`
	Tags          []string               `json:"tags"`
	Importance    float64                `json:"importance"`
	Relationships []string               `json:"relationships"`
	Metadata      map[string]interface{} `json:"metadata"`
}

// NewFileOrganizationEngine creates a new file organization engine
func NewFileOrganizationEngine(logger *logrus.Logger, config OrganizationConfig, aiOrchestrator *ai.Orchestrator) *FileOrganizationEngine {
	tracer := otel.Tracer("file-organization-engine")

	engine := &FileOrganizationEngine{
		logger:              logger,
		tracer:              tracer,
		config:              config,
		aiOrchestrator:      aiOrchestrator,
		strategies:          make(map[string]OrganizationStrategy),
		organizationHistory: make([]OrganizationEvent, 0),
		userPreferences: &OrganizationPreferences{
			PreferredStrategy:   config.DefaultStrategy,
			CategoryPreferences: make(map[string]string),
			FolderStructure:     make(map[string]interface{}),
			NamingConventions:   make(map[string]string),
			AutoRules:           make([]AutoRule, 0),
			LastUpdated:         time.Now(),
		},
		categoryRules: make([]CategoryRule, 0),
	}

	// Initialize organization strategies
	engine.initializeStrategies()

	// Load default category rules
	engine.loadDefaultCategoryRules()

	return engine
}

// initializeStrategies initializes the organization strategies
func (foe *FileOrganizationEngine) initializeStrategies() {
	// TODO: Implement strategy constructors
	// foe.strategies["by_type"] = NewTypeBasedStrategy()
	// foe.strategies["by_date"] = NewDateBasedStrategy()
	// foe.strategies["by_project"] = NewProjectBasedStrategy()
	// foe.strategies["by_size"] = NewSizeBasedStrategy()
	// foe.strategies["semantic"] = NewSemanticStrategy()

	// if foe.aiOrchestrator != nil {
	//	foe.strategies["ai_smart"] = NewAISmartStrategy(foe.aiOrchestrator)
	// }
}

// loadDefaultCategoryRules loads default categorization rules
func (foe *FileOrganizationEngine) loadDefaultCategoryRules() {
	defaultRules := []CategoryRule{
		{
			ID:          "documents",
			Name:        "Documents",
			Description: "Text documents and office files",
			Conditions: []RuleCondition{
				{
					Type:     "extension",
					Operator: "in",
					Value:    []string{".txt", ".doc", ".docx", ".pdf", ".rtf", ".odt"},
					Weight:   1.0,
				},
			},
			Actions: []RuleAction{
				{
					Type: "move_to",
					Parameters: map[string]interface{}{
						"target_dir": "Documents",
					},
				},
			},
			Priority:  1,
			Enabled:   true,
			CreatedAt: time.Now(),
		},
		{
			ID:          "images",
			Name:        "Images",
			Description: "Image files",
			Conditions: []RuleCondition{
				{
					Type:     "extension",
					Operator: "in",
					Value:    []string{".jpg", ".jpeg", ".png", ".gif", ".bmp", ".svg", ".webp"},
					Weight:   1.0,
				},
			},
			Actions: []RuleAction{
				{
					Type: "move_to",
					Parameters: map[string]interface{}{
						"target_dir": "Images",
					},
				},
			},
			Priority:  1,
			Enabled:   true,
			CreatedAt: time.Now(),
		},
		{
			ID:          "videos",
			Name:        "Videos",
			Description: "Video files",
			Conditions: []RuleCondition{
				{
					Type:     "extension",
					Operator: "in",
					Value:    []string{".mp4", ".avi", ".mkv", ".mov", ".wmv", ".flv", ".webm"},
					Weight:   1.0,
				},
			},
			Actions: []RuleAction{
				{
					Type: "move_to",
					Parameters: map[string]interface{}{
						"target_dir": "Videos",
					},
				},
			},
			Priority:  1,
			Enabled:   true,
			CreatedAt: time.Now(),
		},
		{
			ID:          "audio",
			Name:        "Audio",
			Description: "Audio files",
			Conditions: []RuleCondition{
				{
					Type:     "extension",
					Operator: "in",
					Value:    []string{".mp3", ".wav", ".flac", ".aac", ".ogg", ".m4a"},
					Weight:   1.0,
				},
			},
			Actions: []RuleAction{
				{
					Type: "move_to",
					Parameters: map[string]interface{}{
						"target_dir": "Audio",
					},
				},
			},
			Priority:  1,
			Enabled:   true,
			CreatedAt: time.Now(),
		},
		{
			ID:          "code",
			Name:        "Code",
			Description: "Source code files",
			Conditions: []RuleCondition{
				{
					Type:     "extension",
					Operator: "in",
					Value:    []string{".go", ".py", ".js", ".ts", ".java", ".cpp", ".c", ".h", ".rs", ".rb"},
					Weight:   1.0,
				},
			},
			Actions: []RuleAction{
				{
					Type: "move_to",
					Parameters: map[string]interface{}{
						"target_dir": "Code",
					},
				},
			},
			Priority:  1,
			Enabled:   true,
			CreatedAt: time.Now(),
		},
	}

	foe.categoryRules = append(foe.categoryRules, defaultRules...)
}

// CreateOrganizationPlan creates an organization plan for files
func (foe *FileOrganizationEngine) CreateOrganizationPlan(ctx context.Context, sourceDir string, files []FileInfo, strategy string) (*OrganizationPlan, error) {
	ctx, span := foe.tracer.Start(ctx, "fileOrganizationEngine.CreateOrganizationPlan")
	defer span.End()

	foe.mu.RLock()
	defer foe.mu.RUnlock()

	// Select strategy
	if strategy == "" {
		strategy = foe.config.DefaultStrategy
	}

	orgStrategy, exists := foe.strategies[strategy]
	if !exists {
		return nil, fmt.Errorf("unknown organization strategy: %s", strategy)
	}

	// Categorize files first
	categorizedFiles := foe.categorizeFiles(files)

	// Create organization plan
	plan, err := orgStrategy.Organize(categorizedFiles, sourceDir)
	if err != nil {
		return nil, fmt.Errorf("failed to create organization plan: %w", err)
	}

	// Enhance plan with AI insights if available
	if foe.config.AIAssisted && foe.aiOrchestrator != nil {
		plan = foe.enhancePlanWithAI(ctx, plan, categorizedFiles)
	}

	// Validate plan
	if err := foe.validatePlan(plan); err != nil {
		return nil, fmt.Errorf("invalid organization plan: %w", err)
	}

	foe.logger.WithFields(logrus.Fields{
		"plan_id":    plan.ID,
		"strategy":   strategy,
		"operations": len(plan.Operations),
		"confidence": plan.Confidence,
	}).Debug("Organization plan created")

	return plan, nil
}

// categorizeFiles categorizes files based on rules
func (foe *FileOrganizationEngine) categorizeFiles(files []FileInfo) []FileInfo {
	categorizedFiles := make([]FileInfo, len(files))
	copy(categorizedFiles, files)

	for i := range categorizedFiles {
		category := foe.determineFileCategory(&categorizedFiles[i])
		categorizedFiles[i].Category = category
	}

	return categorizedFiles
}

// determineFileCategory determines the category for a file
func (foe *FileOrganizationEngine) determineFileCategory(file *FileInfo) string {
	bestCategory := "uncategorized"
	bestScore := 0.0

	for _, rule := range foe.categoryRules {
		if !rule.Enabled {
			continue
		}

		score := foe.evaluateRule(&rule, file)
		if score > bestScore && score >= foe.config.CategoryThreshold {
			bestScore = score
			bestCategory = rule.Name
		}
	}

	return bestCategory
}

// evaluateRule evaluates a categorization rule against a file
func (foe *FileOrganizationEngine) evaluateRule(rule *CategoryRule, file *FileInfo) float64 {
	totalWeight := 0.0
	matchedWeight := 0.0

	for _, condition := range rule.Conditions {
		weight := condition.Weight
		if weight == 0 {
			weight = 1.0
		}
		totalWeight += weight

		if foe.evaluateCondition(&condition, file) {
			matchedWeight += weight
		}
	}

	if totalWeight == 0 {
		return 0
	}

	return matchedWeight / totalWeight
}

// evaluateCondition evaluates a single condition
func (foe *FileOrganizationEngine) evaluateCondition(condition *RuleCondition, file *FileInfo) bool {
	switch condition.Type {
	case "extension":
		return foe.evaluateExtensionCondition(condition, file.Extension)
	case "size":
		return foe.evaluateSizeCondition(condition, file.Size)
	case "name":
		return foe.evaluateNameCondition(condition, file.Name)
	case "content":
		return foe.evaluateContentCondition(condition, file)
	default:
		return false
	}
}

// evaluateExtensionCondition evaluates extension-based conditions
func (foe *FileOrganizationEngine) evaluateExtensionCondition(condition *RuleCondition, extension string) bool {
	switch condition.Operator {
	case "equals":
		if value, ok := condition.Value.(string); ok {
			return extension == value
		}
	case "in":
		if values, ok := condition.Value.([]string); ok {
			for _, value := range values {
				if extension == value {
					return true
				}
			}
		}
	case "contains":
		if value, ok := condition.Value.(string); ok {
			return strings.Contains(extension, value)
		}
	}
	return false
}

// evaluateSizeCondition evaluates size-based conditions
func (foe *FileOrganizationEngine) evaluateSizeCondition(condition *RuleCondition, size int64) bool {
	if value, ok := condition.Value.(float64); ok {
		intValue := int64(value)
		switch condition.Operator {
		case "greater_than":
			return size > intValue
		case "less_than":
			return size < intValue
		case "equals":
			return size == intValue
		}
	}
	return false
}

// evaluateNameCondition evaluates name-based conditions
func (foe *FileOrganizationEngine) evaluateNameCondition(condition *RuleCondition, name string) bool {
	if value, ok := condition.Value.(string); ok {
		switch condition.Operator {
		case "contains":
			return strings.Contains(strings.ToLower(name), strings.ToLower(value))
		case "starts_with":
			return strings.HasPrefix(strings.ToLower(name), strings.ToLower(value))
		case "ends_with":
			return strings.HasSuffix(strings.ToLower(name), strings.ToLower(value))
		}
	}
	return false
}

// evaluateContentCondition evaluates content-based conditions
func (foe *FileOrganizationEngine) evaluateContentCondition(condition *RuleCondition, file *FileInfo) bool {
	// This would require content analysis
	// For now, return false
	return false
}

// enhancePlanWithAI enhances the organization plan with AI insights
func (foe *FileOrganizationEngine) enhancePlanWithAI(ctx context.Context, plan *OrganizationPlan, files []FileInfo) *OrganizationPlan {
	// Create AI request for plan enhancement
	aiRequest := &models.AIRequest{
		ID:    fmt.Sprintf("org-enhance-%s", plan.ID),
		Type:  "enhancement",
		Input: fmt.Sprintf("Enhance file organization plan with %d operations", len(plan.Operations)),
		Parameters: map[string]interface{}{
			"task":  "organization_enhancement",
			"plan":  plan,
			"files": files,
		},
		Timeout:   5 * time.Second,
		Timestamp: time.Now(),
	}

	_, err := foe.aiOrchestrator.ProcessRequest(ctx, aiRequest)
	if err != nil {
		foe.logger.WithError(err).Debug("Failed to enhance plan with AI")
		return plan
	}

	// Parse AI response and enhance plan
	// This would involve more sophisticated AI integration
	return plan
}

// validatePlan validates an organization plan
func (foe *FileOrganizationEngine) validatePlan(plan *OrganizationPlan) error {
	if plan == nil {
		return fmt.Errorf("plan is nil")
	}

	if len(plan.Operations) == 0 {
		return fmt.Errorf("plan has no operations")
	}

	// Check for conflicts
	targetPaths := make(map[string]bool)
	for _, op := range plan.Operations {
		if op.Type == "move" || op.Type == "copy" {
			if targetPaths[op.TargetPath] {
				return fmt.Errorf("conflicting target path: %s", op.TargetPath)
			}
			targetPaths[op.TargetPath] = true
		}
	}

	return nil
}

// ExecutePlan executes an organization plan
func (foe *FileOrganizationEngine) ExecutePlan(ctx context.Context, plan *OrganizationPlan, dryRun bool) (*OrganizationEvent, error) {
	ctx, span := foe.tracer.Start(ctx, "fileOrganizationEngine.ExecutePlan")
	defer span.End()

	start := time.Now()

	event := &OrganizationEvent{
		PlanID:     plan.ID,
		Strategy:   plan.Strategy,
		Timestamp:  start,
		FilesCount: len(plan.Operations),
		Success:    true,
		Context: map[string]interface{}{
			"dry_run": dryRun,
		},
	}

	if dryRun {
		foe.logger.Info("Dry run mode - no files will be moved")
		event.Duration = time.Since(start)
		return event, nil
	}

	// Execute operations
	for _, operation := range plan.Operations {
		if err := foe.executeOperation(&operation); err != nil {
			event.Success = false
			event.Error = err.Error()
			break
		}
	}

	event.Duration = time.Since(start)

	// Record event
	foe.recordOrganizationEvent(*event)

	foe.logger.WithFields(logrus.Fields{
		"plan_id":  plan.ID,
		"success":  event.Success,
		"duration": event.Duration,
	}).Info("Organization plan executed")

	return event, nil
}

// executeOperation executes a single file operation
func (foe *FileOrganizationEngine) executeOperation(operation *FileOperation) error {
	switch operation.Type {
	case "move":
		return foe.moveFile(operation.SourcePath, operation.TargetPath)
	case "copy":
		return foe.copyFile(operation.SourcePath, operation.TargetPath)
	case "rename":
		return foe.renameFile(operation.SourcePath, operation.TargetPath)
	case "create_dir":
		return foe.createDirectory(operation.TargetPath)
	default:
		return fmt.Errorf("unknown operation type: %s", operation.Type)
	}
}

// File operation implementations (simplified)

func (foe *FileOrganizationEngine) moveFile(source, target string) error {
	// Implementation would move the file
	foe.logger.WithFields(logrus.Fields{
		"source": source,
		"target": target,
	}).Debug("Moving file")
	return nil
}

func (foe *FileOrganizationEngine) copyFile(source, target string) error {
	// Implementation would copy the file
	foe.logger.WithFields(logrus.Fields{
		"source": source,
		"target": target,
	}).Debug("Copying file")
	return nil
}

func (foe *FileOrganizationEngine) renameFile(source, target string) error {
	// Implementation would rename the file
	foe.logger.WithFields(logrus.Fields{
		"source": source,
		"target": target,
	}).Debug("Renaming file")
	return nil
}

func (foe *FileOrganizationEngine) createDirectory(path string) error {
	// Implementation would create the directory
	foe.logger.WithField("path", path).Debug("Creating directory")
	return nil
}

// recordOrganizationEvent records an organization event
func (foe *FileOrganizationEngine) recordOrganizationEvent(event OrganizationEvent) {
	foe.mu.Lock()
	defer foe.mu.Unlock()

	foe.organizationHistory = append(foe.organizationHistory, event)
	foe.totalOrganizations++

	if event.Success {
		foe.successfulOrgs++
	}

	// Update success rate
	foe.successRate = float64(foe.successfulOrgs) / float64(foe.totalOrganizations)

	// Maintain history size
	if len(foe.organizationHistory) > 1000 {
		foe.organizationHistory = foe.organizationHistory[100:]
	}
}

// GetOrganizationMetrics returns organization metrics
func (foe *FileOrganizationEngine) GetOrganizationMetrics() map[string]interface{} {
	foe.mu.RLock()
	defer foe.mu.RUnlock()

	return map[string]interface{}{
		"total_organizations": foe.totalOrganizations,
		"success_rate":        foe.successRate,
		"strategies_count":    len(foe.strategies),
		"rules_count":         len(foe.categoryRules),
	}
}
