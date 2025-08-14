package ai

import (
	"context"
	"time"

	"github.com/aios/aios/pkg/models"
)

// LanguageModelService defines the interface for language model operations
type LanguageModelService interface {
	// ProcessQuery processes a natural language query and returns a response
	ProcessQuery(ctx context.Context, query string) (*models.LLMResponse, error)
	
	// GenerateCode generates code based on a prompt
	GenerateCode(ctx context.Context, prompt string) (*models.CodeResponse, error)
	
	// AnalyzeText analyzes text for various insights
	AnalyzeText(ctx context.Context, text string) (*models.TextAnalysis, error)
	
	// Chat maintains a conversation context
	Chat(ctx context.Context, message string, conversationID string) (*models.ChatResponse, error)
	
	// Summarize creates a summary of the given text
	Summarize(ctx context.Context, text string) (*models.SummaryResponse, error)
	
	// Translate translates text between languages
	Translate(ctx context.Context, text, fromLang, toLang string) (*models.TranslationResponse, error)
	
	// GetModels returns available language models
	GetModels(ctx context.Context) ([]models.AIModel, error)
	
	// LoadModel loads a specific model
	LoadModel(ctx context.Context, modelName string) error
	
	// UnloadModel unloads a specific model
	UnloadModel(ctx context.Context, modelName string) error
}

// ComputerVisionService defines the interface for computer vision operations
type ComputerVisionService interface {
	// AnalyzeScreen analyzes a screenshot for UI elements and content
	AnalyzeScreen(ctx context.Context, screenshot []byte) (*models.ScreenAnalysis, error)
	
	// DetectUI detects UI elements in an image
	DetectUI(ctx context.Context, image []byte) (*models.UIElements, error)
	
	// RecognizeText performs OCR on an image
	RecognizeText(ctx context.Context, image []byte) (*models.TextRecognition, error)
	
	// ClassifyImage classifies the content of an image
	ClassifyImage(ctx context.Context, image []byte) (*models.ImageClassification, error)
	
	// DetectObjects detects objects in an image
	DetectObjects(ctx context.Context, image []byte) (*models.ObjectDetection, error)
	
	// AnalyzeLayout analyzes the layout structure of a UI
	AnalyzeLayout(ctx context.Context, image []byte) (*models.LayoutAnalysis, error)
	
	// GenerateDescription generates a natural language description of an image
	GenerateDescription(ctx context.Context, image []byte) (*models.ImageDescription, error)
	
	// CompareImages compares two images for similarity
	CompareImages(ctx context.Context, image1, image2 []byte) (*models.ImageComparison, error)
}

// SystemOptimizationService defines the interface for AI-powered system optimization
type SystemOptimizationService interface {
	// AnalyzePerformance analyzes current system performance
	AnalyzePerformance(ctx context.Context) (*models.PerformanceReport, error)
	
	// OptimizeResources optimizes system resources based on constraints
	OptimizeResources(ctx context.Context, constraints *models.ResourceConstraints) error
	
	// PredictUsage predicts future resource usage
	PredictUsage(ctx context.Context, timeframe time.Duration) (*models.UsagePrediction, error)
	
	// GenerateRecommendations generates optimization recommendations
	GenerateRecommendations(ctx context.Context) ([]models.OptimizationRecommendation, error)
	
	// ApplyOptimization applies a specific optimization
	ApplyOptimization(ctx context.Context, optimizationID string) error
	
	// MonitorHealth continuously monitors system health
	MonitorHealth(ctx context.Context) (*models.HealthReport, error)
	
	// PredictFailures predicts potential system failures
	PredictFailures(ctx context.Context) (*models.FailurePrediction, error)
	
	// OptimizeWorkload optimizes workload distribution
	OptimizeWorkload(ctx context.Context, workload *models.WorkloadSpec) (*models.WorkloadOptimization, error)
}

// VoiceService defines the interface for voice processing operations
type VoiceService interface {
	// SpeechToText converts speech audio to text
	SpeechToText(ctx context.Context, audio []byte) (*models.SpeechRecognition, error)
	
	// TextToSpeech converts text to speech audio
	TextToSpeech(ctx context.Context, text string) (*models.SpeechSynthesis, error)
	
	// DetectWakeWord detects wake words in audio
	DetectWakeWord(ctx context.Context, audio []byte) (*models.WakeWordDetection, error)
	
	// AnalyzeVoice analyzes voice characteristics
	AnalyzeVoice(ctx context.Context, audio []byte) (*models.VoiceAnalysis, error)
	
	// ProcessVoiceCommand processes a voice command
	ProcessVoiceCommand(ctx context.Context, audio []byte) (*models.VoiceCommand, error)
}

// NaturalLanguageService defines the interface for natural language processing
type NaturalLanguageService interface {
	// ParseIntent extracts intent from natural language
	ParseIntent(ctx context.Context, text string) (*models.IntentAnalysis, error)
	
	// ExtractEntities extracts named entities from text
	ExtractEntities(ctx context.Context, text string) (*models.EntityExtraction, error)
	
	// AnalyzeSentiment analyzes sentiment of text
	AnalyzeSentiment(ctx context.Context, text string) (*models.SentimentAnalysis, error)
	
	// GenerateResponse generates a natural language response
	GenerateResponse(ctx context.Context, intent *models.IntentAnalysis, context map[string]interface{}) (*models.NLResponse, error)
	
	// ParseCommand parses a natural language command
	ParseCommand(ctx context.Context, text string) (*models.CommandParsing, error)
	
	// ValidateCommand validates if a command is safe to execute
	ValidateCommand(ctx context.Context, command *models.CommandParsing) (*models.CommandValidation, error)
}

// ModelManager defines the interface for AI model management
type ModelManager interface {
	// ListModels lists all available models
	ListModels(ctx context.Context) ([]models.AIModel, error)
	
	// LoadModel loads a model into memory
	LoadModel(ctx context.Context, modelID string) error
	
	// UnloadModel unloads a model from memory
	UnloadModel(ctx context.Context, modelID string) error
	
	// GetModelStatus gets the status of a specific model
	GetModelStatus(ctx context.Context, modelID string) (*models.ModelStatus, error)
	
	// UpdateModel updates a model to a new version
	UpdateModel(ctx context.Context, modelID, version string) error
	
	// DeleteModel deletes a model
	DeleteModel(ctx context.Context, modelID string) error
	
	// GetModelMetrics gets performance metrics for a model
	GetModelMetrics(ctx context.Context, modelID string) (*models.ModelMetrics, error)
	
	// OptimizeModel optimizes a model for better performance
	OptimizeModel(ctx context.Context, modelID string, optimizationParams *models.OptimizationParams) error
}

// AIOrchestrator defines the interface for coordinating AI services
type AIOrchestrator interface {
	// ProcessRequest processes a complex AI request that may involve multiple services
	ProcessRequest(ctx context.Context, request *models.AIRequest) (*models.AIResponse, error)
	
	// GetServiceStatus gets the status of all AI services
	GetServiceStatus(ctx context.Context) (*models.AIServiceStatus, error)
	
	// RouteRequest routes a request to the appropriate AI service
	RouteRequest(ctx context.Context, request *models.AIRequest) (string, error)
	
	// AggregateResults aggregates results from multiple AI services
	AggregateResults(ctx context.Context, results []models.AIResult) (*models.AggregatedResult, error)
	
	// ManageWorkflow manages complex AI workflows
	ManageWorkflow(ctx context.Context, workflow *models.AIWorkflow) (*models.WorkflowResult, error)
}

// AIServiceConfig represents configuration for AI services
type AIServiceConfig struct {
	// Ollama configuration
	OllamaHost    string        `yaml:"ollama_host"`
	OllamaPort    int           `yaml:"ollama_port"`
	OllamaTimeout time.Duration `yaml:"ollama_timeout"`
	
	// Model configuration
	ModelsPath   string  `yaml:"models_path"`
	DefaultModel string  `yaml:"default_model"`
	MaxTokens    int     `yaml:"max_tokens"`
	Temperature  float64 `yaml:"temperature"`
	
	// Computer vision configuration
	CVEnabled           bool   `yaml:"cv_enabled"`
	CVModelPath         string `yaml:"cv_model_path"`
	ConfidenceThreshold float64 `yaml:"confidence_threshold"`
	MaxImageSize        string `yaml:"max_image_size"`
	
	// Voice processing configuration
	VoiceEnabled    bool   `yaml:"voice_enabled"`
	VoiceModelPath  string `yaml:"voice_model_path"`
	WakeWord        string `yaml:"wake_word"`
	SampleRate      int    `yaml:"sample_rate"`
	
	// Performance configuration
	MaxConcurrentRequests int           `yaml:"max_concurrent_requests"`
	RequestTimeout        time.Duration `yaml:"request_timeout"`
	ModelCacheSize        int           `yaml:"model_cache_size"`
	
	// Security configuration
	EnableSandbox     bool     `yaml:"enable_sandbox"`
	AllowedOperations []string `yaml:"allowed_operations"`
	DataRetention     time.Duration `yaml:"data_retention"`
}
