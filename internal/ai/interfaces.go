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

	// ProcessQueryStream processes a query with streaming response
	ProcessQueryStream(ctx context.Context, query string) (<-chan *models.LLMStreamChunk, error)

	// GenerateCode generates code based on a prompt
	GenerateCode(ctx context.Context, prompt string) (*models.CodeResponse, error)

	// GenerateCodeStream generates code with streaming response
	GenerateCodeStream(ctx context.Context, prompt string) (<-chan *models.CodeStreamChunk, error)

	// AnalyzeText analyzes text for various insights
	AnalyzeText(ctx context.Context, text string) (*models.TextAnalysis, error)

	// Chat maintains a conversation context
	Chat(ctx context.Context, message string, conversationID string) (*models.ChatResponse, error)

	// ChatWithHistory maintains a conversation with full message history
	ChatWithHistory(ctx context.Context, messages []models.ChatMessage) (*models.ChatResponse, error)

	// ChatStream maintains a conversation context with streaming
	ChatStream(ctx context.Context, message string, conversationID string) (<-chan *models.ChatStreamChunk, error)

	// Summarize creates a summary of the given text
	Summarize(ctx context.Context, text string) (*models.SummaryResponse, error)

	// Translate translates text between languages
	Translate(ctx context.Context, text, fromLang, toLang string) (*models.TranslationResponse, error)

	// FunctionCall executes a function call based on the model's response
	FunctionCall(ctx context.Context, functionName string, parameters map[string]any) (*models.FunctionCallResponse, error)

	// EmbedText generates embeddings for text
	EmbedText(ctx context.Context, text string) (*models.EmbeddingResponse, error)

	// BatchEmbed generates embeddings for multiple texts
	BatchEmbed(ctx context.Context, texts []string) (*models.BatchEmbeddingResponse, error)

	// GetModels returns available language models
	GetModels(ctx context.Context) ([]models.AIModel, error)

	// LoadModel loads a specific model
	LoadModel(ctx context.Context, modelName string) error

	// UnloadModel unloads a specific model
	UnloadModel(ctx context.Context, modelName string) error

	// GetModelInfo gets detailed information about a model
	GetModelInfo(ctx context.Context, modelName string) (*models.ModelInfo, error)
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

// MultiModalService defines the interface for multi-modal AI operations
type MultiModalService interface {
	// ProcessMultiModal processes requests involving multiple modalities
	ProcessMultiModal(ctx context.Context, request *models.MultiModalRequest) (*models.MultiModalResponse, error)

	// GenerateImageFromText generates images from text descriptions
	GenerateImageFromText(ctx context.Context, prompt string) (*models.ImageGenerationResponse, error)

	// DescribeImage generates text descriptions from images
	DescribeImage(ctx context.Context, image []byte) (*models.ImageDescriptionResponse, error)

	// AnalyzeVideoContent analyzes video content
	AnalyzeVideoContent(ctx context.Context, video []byte) (*models.VideoAnalysisResponse, error)

	// CrossModalSearch performs search across different modalities
	CrossModalSearch(ctx context.Context, query string, modalities []string) (*models.CrossModalSearchResponse, error)
}

// RAGService defines the interface for Retrieval Augmented Generation
type RAGService interface {
	// IndexDocuments indexes documents for retrieval
	IndexDocuments(ctx context.Context, documents []models.Document) error

	// SearchDocuments searches for relevant documents
	SearchDocuments(ctx context.Context, query string, limit int) (*models.DocumentSearchResponse, error)

	// GenerateWithContext generates responses using retrieved context
	GenerateWithContext(ctx context.Context, query string, context []models.Document) (*models.RAGResponse, error)

	// UpdateIndex updates the document index
	UpdateIndex(ctx context.Context, documentID string, document models.Document) error

	// DeleteFromIndex removes documents from the index
	DeleteFromIndex(ctx context.Context, documentIDs []string) error
}

// WorkflowEngine defines the interface for AI workflow management
type WorkflowEngine interface {
	// ExecuteWorkflow executes a complex AI workflow
	ExecuteWorkflow(ctx context.Context, workflow *models.AIWorkflow) (*models.WorkflowResult, error)

	// CreateWorkflow creates a new workflow template
	CreateWorkflow(ctx context.Context, workflow *models.WorkflowTemplate) error

	// GetWorkflowStatus gets the status of a running workflow
	GetWorkflowStatus(ctx context.Context, workflowID string) (*models.WorkflowStatus, error)

	// CancelWorkflow cancels a running workflow
	CancelWorkflow(ctx context.Context, workflowID string) error

	// ListWorkflows lists available workflow templates
	ListWorkflows(ctx context.Context) ([]models.WorkflowTemplate, error)
}

// CacheManager defines the interface for AI caching operations
type CacheManager interface {
	// Get retrieves a cached result
	Get(ctx context.Context, key string) (interface{}, bool, error)

	// Set stores a result in cache
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error

	// Delete removes a cached result
	Delete(ctx context.Context, key string) error

	// Clear clears all cached results
	Clear(ctx context.Context) error

	// GetStats returns cache statistics
	GetStats(ctx context.Context) (*models.CacheStats, error)
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
	CVEnabled           bool    `yaml:"cv_enabled"`
	CVModelPath         string  `yaml:"cv_model_path"`
	ConfidenceThreshold float64 `yaml:"confidence_threshold"`
	MaxImageSize        string  `yaml:"max_image_size"`

	// Voice processing configuration
	VoiceEnabled   bool   `yaml:"voice_enabled"`
	VoiceModelPath string `yaml:"voice_model_path"`
	WakeWord       string `yaml:"wake_word"`
	SampleRate     int    `yaml:"sample_rate"`

	// NLP configuration
	NLPEnabled   bool   `yaml:"nlp_enabled"`
	NLPModelPath string `yaml:"nlp_model_path"`
	IntentModel  string `yaml:"intent_model"`
	EntityModel  string `yaml:"entity_model"`

	// Multi-modal configuration
	MultiModalEnabled  bool   `yaml:"multimodal_enabled"`
	MultiModalPath     string `yaml:"multimodal_path"`
	ImageGenModel      string `yaml:"image_gen_model"`
	VideoAnalysisModel string `yaml:"video_analysis_model"`

	// RAG configuration
	RAGEnabled     bool   `yaml:"rag_enabled"`
	VectorDBPath   string `yaml:"vector_db_path"`
	VectorDBType   string `yaml:"vector_db_type"` // chroma, pinecone, weaviate
	EmbeddingModel string `yaml:"embedding_model"`
	ChunkSize      int    `yaml:"chunk_size"`
	ChunkOverlap   int    `yaml:"chunk_overlap"`

	// Caching configuration
	CacheEnabled        bool          `yaml:"cache_enabled"`
	CacheType           string        `yaml:"cache_type"` // memory, redis, distributed
	CacheTTL            time.Duration `yaml:"cache_ttl"`
	SemanticCache       bool          `yaml:"semantic_cache"`
	CacheMaxSize        int64         `yaml:"cache_max_size"`
	CacheEvictionPolicy string        `yaml:"cache_eviction_policy"`

	// Workflow configuration
	WorkflowEnabled  bool          `yaml:"workflow_enabled"`
	WorkflowTimeout  time.Duration `yaml:"workflow_timeout"`
	MaxWorkflowSteps int           `yaml:"max_workflow_steps"`
	WorkflowRetries  int           `yaml:"workflow_retries"`

	// Performance configuration
	MaxConcurrentRequests int           `yaml:"max_concurrent_requests"`
	RequestTimeout        time.Duration `yaml:"request_timeout"`
	ModelCacheSize        int           `yaml:"model_cache_size"`
	GPUEnabled            bool          `yaml:"gpu_enabled"`
	GPUMemoryFraction     float64       `yaml:"gpu_memory_fraction"`
	BatchSize             int           `yaml:"batch_size"`

	// Security configuration
	EnableSandbox     bool          `yaml:"enable_sandbox"`
	AllowedOperations []string      `yaml:"allowed_operations"`
	DataRetention     time.Duration `yaml:"data_retention"`
	EncryptionEnabled bool          `yaml:"encryption_enabled"`
	AuditLogging      bool          `yaml:"audit_logging"`

	// Provider configuration
	Providers map[string]ProviderConfig `yaml:"providers"`
}

// ProviderConfig represents configuration for AI providers
type ProviderConfig struct {
	Type     string                 `yaml:"type"` // ollama, openai, anthropic, etc.
	Endpoint string                 `yaml:"endpoint"`
	APIKey   string                 `yaml:"api_key"`
	Models   []string               `yaml:"models"`
	Enabled  bool                   `yaml:"enabled"`
	Priority int                    `yaml:"priority"`
	Config   map[string]interface{} `yaml:"config,omitempty"`
}
