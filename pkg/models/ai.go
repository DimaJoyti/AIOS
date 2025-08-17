package models

import (
	"time"
)

// AI Service Request/Response Models

// PerformanceReport represents a system performance analysis report
type PerformanceReport struct {
	OverallScore    float64                     `json:"overall_score"`
	CPUAnalysis     CPUAnalysis                 `json:"cpu_analysis"`
	MemoryAnalysis  MemoryAnalysis              `json:"memory_analysis"`
	DiskAnalysis    DiskAnalysis                `json:"disk_analysis"`
	NetworkAnalysis NetworkAnalysis             `json:"network_analysis"`
	Bottlenecks     []PerformanceBottleneck     `json:"bottlenecks"`
	Recommendations []PerformanceRecommendation `json:"recommendations"`
	Timestamp       time.Time                   `json:"timestamp"`
}

// CPUAnalysis represents CPU performance analysis
type CPUAnalysis struct {
	UtilizationTrend []float64         `json:"utilization_trend"`
	AverageLoad      float64           `json:"average_load"`
	PeakLoad         float64           `json:"peak_load"`
	EfficiencyScore  float64           `json:"efficiency_score"`
	Processes        []ProcessAnalysis `json:"top_processes"`
}

// MemoryAnalysis represents memory performance analysis
type MemoryAnalysis struct {
	UsageTrend         []float64         `json:"usage_trend"`
	FragmentationLevel float64           `json:"fragmentation_level"`
	CacheEfficiency    float64           `json:"cache_efficiency"`
	SwapUsage          float64           `json:"swap_usage"`
	LeakSuspects       []ProcessAnalysis `json:"leak_suspects"`
}

// DiskAnalysis represents disk performance analysis
type DiskAnalysis struct {
	IOPSTrend          []float64 `json:"iops_trend"`
	ThroughputTrend    []float64 `json:"throughput_trend"`
	LatencyTrend       []float64 `json:"latency_trend"`
	FragmentationLevel float64   `json:"fragmentation_level"`
	HealthScore        float64   `json:"health_score"`
}

// NetworkAnalysis represents network performance analysis
type NetworkAnalysis struct {
	BandwidthUsage  []float64 `json:"bandwidth_usage"`
	LatencyTrend    []float64 `json:"latency_trend"`
	PacketLoss      float64   `json:"packet_loss"`
	ConnectionCount int       `json:"connection_count"`
	ThroughputScore float64   `json:"throughput_score"`
}

// ProcessAnalysis represents individual process analysis
type ProcessAnalysis struct {
	PID         int     `json:"pid"`
	Name        string  `json:"name"`
	CPUUsage    float64 `json:"cpu_usage"`
	MemoryUsage int64   `json:"memory_usage"`
	IOUsage     float64 `json:"io_usage"`
	Priority    int     `json:"priority"`
	Trend       string  `json:"trend"` // increasing, decreasing, stable
}

// PerformanceBottleneck represents a performance bottleneck
type PerformanceBottleneck struct {
	Type        string  `json:"type"`     // cpu, memory, disk, network
	Severity    string  `json:"severity"` // low, medium, high, critical
	Description string  `json:"description"`
	Impact      float64 `json:"impact"` // 0-100
	Source      string  `json:"source"`
	Suggestion  string  `json:"suggestion"`
}

// PerformanceRecommendation represents a performance improvement recommendation
type PerformanceRecommendation struct {
	ID          string  `json:"id"`
	Type        string  `json:"type"`
	Priority    string  `json:"priority"`
	Description string  `json:"description"`
	Impact      float64 `json:"expected_impact"`
	Effort      string  `json:"effort"` // low, medium, high
	Risk        string  `json:"risk"`   // low, medium, high
}

// ResourceConstraints represents resource optimization constraints
type ResourceConstraints struct {
	MaxCPUUsage     float64 `json:"max_cpu_usage"`
	MaxMemoryUsage  float64 `json:"max_memory_usage"`
	MaxDiskUsage    float64 `json:"max_disk_usage"`
	MaxNetworkUsage float64 `json:"max_network_usage"`
	PowerSaving     bool    `json:"power_saving"`
	PerformanceMode string  `json:"performance_mode"` // balanced, performance, power_save
}

// UsagePrediction represents predicted resource usage
type UsagePrediction struct {
	Timeframe    time.Duration     `json:"timeframe"`
	CPUTrend     []PredictionPoint `json:"cpu_trend"`
	MemoryTrend  []PredictionPoint `json:"memory_trend"`
	DiskTrend    []PredictionPoint `json:"disk_trend"`
	NetworkTrend []PredictionPoint `json:"network_trend"`
	Confidence   float64           `json:"confidence"`
	Factors      []string          `json:"influencing_factors"`
}

// PredictionPoint represents a single prediction point
type PredictionPoint struct {
	Timestamp  time.Time `json:"timestamp"`
	Value      float64   `json:"value"`
	Confidence float64   `json:"confidence"`
}

// HealthReport represents system health analysis
type HealthReport struct {
	OverallHealth   float64            `json:"overall_health"` // 0-100
	ComponentHealth map[string]float64 `json:"component_health"`
	Issues          []HealthIssue      `json:"issues"`
	Warnings        []HealthWarning    `json:"warnings"`
	Recommendations []string           `json:"recommendations"`
	Timestamp       time.Time          `json:"timestamp"`
}

// HealthIssue represents a health issue
type HealthIssue struct {
	Component   string `json:"component"`
	Severity    string `json:"severity"`
	Description string `json:"description"`
	Impact      string `json:"impact"`
	Resolution  string `json:"resolution"`
}

// HealthWarning represents a health warning
type HealthWarning struct {
	Component    string  `json:"component"`
	Type         string  `json:"type"`
	Description  string  `json:"description"`
	Threshold    float64 `json:"threshold"`
	CurrentValue float64 `json:"current_value"`
}

// FailurePrediction represents predicted system failures
type FailurePrediction struct {
	Predictions []FailureRisk `json:"predictions"`
	OverallRisk float64       `json:"overall_risk"`
	Timeframe   time.Duration `json:"timeframe"`
	Confidence  float64       `json:"confidence"`
	Timestamp   time.Time     `json:"timestamp"`
}

// FailureRisk represents a specific failure risk
type FailureRisk struct {
	Component     string        `json:"component"`
	FailureType   string        `json:"failure_type"`
	Probability   float64       `json:"probability"`
	Impact        string        `json:"impact"`
	TimeToFailure time.Duration `json:"time_to_failure"`
	Indicators    []string      `json:"indicators"`
	Prevention    []string      `json:"prevention_steps"`
}

// WorkloadSpec represents a workload specification
type WorkloadSpec struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Type        string                 `json:"type"` // cpu_intensive, memory_intensive, io_intensive
	Priority    int                    `json:"priority"`
	Resources   ResourceRequirements   `json:"resources"`
	Constraints WorkloadConstraints    `json:"constraints"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// ResourceRequirements represents resource requirements
type ResourceRequirements struct {
	CPU     float64 `json:"cpu"`     // cores
	Memory  int64   `json:"memory"`  // bytes
	Disk    int64   `json:"disk"`    // bytes
	Network float64 `json:"network"` // mbps
}

// WorkloadConstraints represents workload constraints
type WorkloadConstraints struct {
	MaxLatency    time.Duration `json:"max_latency"`
	MinThroughput float64       `json:"min_throughput"`
	Deadline      time.Time     `json:"deadline,omitempty"`
	Dependencies  []string      `json:"dependencies,omitempty"`
}

// WorkloadOptimization represents workload optimization results
type WorkloadOptimization struct {
	OriginalSpec    WorkloadSpec     `json:"original_spec"`
	OptimizedSpec   WorkloadSpec     `json:"optimized_spec"`
	Improvements    []Improvement    `json:"improvements"`
	ExpectedGains   PerformanceGains `json:"expected_gains"`
	Recommendations []string         `json:"recommendations"`
	Timestamp       time.Time        `json:"timestamp"`
}

// Improvement represents a specific improvement
type Improvement struct {
	Type        string  `json:"type"`
	Description string  `json:"description"`
	Impact      float64 `json:"impact"`
	Confidence  float64 `json:"confidence"`
}

// PerformanceGains represents expected performance gains
type PerformanceGains struct {
	CPUEfficiency    float64 `json:"cpu_efficiency"`
	MemoryEfficiency float64 `json:"memory_efficiency"`
	IOEfficiency     float64 `json:"io_efficiency"`
	OverallGain      float64 `json:"overall_gain"`
}

// Voice and Speech Models

// SpeechRecognition represents speech recognition results
type SpeechRecognition struct {
	Text       string            `json:"text"`
	Confidence float64           `json:"confidence"`
	Language   string            `json:"language"`
	Duration   time.Duration     `json:"duration"`
	Words      []WordRecognition `json:"words,omitempty"`
	Timestamp  time.Time         `json:"timestamp"`
}

// WordRecognition represents individual word recognition
type WordRecognition struct {
	Word       string        `json:"word"`
	Confidence float64       `json:"confidence"`
	StartTime  time.Duration `json:"start_time"`
	EndTime    time.Duration `json:"end_time"`
}

// SpeechSynthesis represents speech synthesis results
type SpeechSynthesis struct {
	Audio      []byte        `json:"audio"`
	Format     string        `json:"format"` // wav, mp3, etc.
	Duration   time.Duration `json:"duration"`
	SampleRate int           `json:"sample_rate"`
	Timestamp  time.Time     `json:"timestamp"`
}

// WakeWordDetection represents wake word detection results
type WakeWordDetection struct {
	Detected   bool          `json:"detected"`
	WakeWord   string        `json:"wake_word"`
	Confidence float64       `json:"confidence"`
	StartTime  time.Duration `json:"start_time"`
	EndTime    time.Duration `json:"end_time"`
	Timestamp  time.Time     `json:"timestamp"`
}

// VoiceAnalysis represents voice analysis results
type VoiceAnalysis struct {
	SpeakerID string                 `json:"speaker_id,omitempty"`
	Gender    string                 `json:"gender,omitempty"`
	Age       int                    `json:"age,omitempty"`
	Emotion   string                 `json:"emotion,omitempty"`
	Stress    float64                `json:"stress_level"`
	Clarity   float64                `json:"clarity"`
	Pace      float64                `json:"pace"` // words per minute
	Volume    float64                `json:"volume"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
}

// VoiceCommand represents a processed voice command
type VoiceCommand struct {
	Command    string                 `json:"command"`
	Intent     string                 `json:"intent"`
	Entities   []NamedEntity          `json:"entities"`
	Parameters map[string]interface{} `json:"parameters,omitempty"`
	Confidence float64                `json:"confidence"`
	Action     string                 `json:"action,omitempty"`
	Timestamp  time.Time              `json:"timestamp"`
}

// Natural Language Processing Models

// IntentAnalysis represents intent analysis results
type IntentAnalysis struct {
	Intent     string                 `json:"intent"`
	Confidence float64                `json:"confidence"`
	Entities   []NamedEntity          `json:"entities"`
	Context    map[string]interface{} `json:"context,omitempty"`
	Timestamp  time.Time              `json:"timestamp"`
}

// EntityExtraction represents entity extraction results
type EntityExtraction struct {
	Entities  []NamedEntity    `json:"entities"`
	Relations []EntityRelation `json:"relations,omitempty"`
	Timestamp time.Time        `json:"timestamp"`
}

// EntityRelation represents a relationship between entities
type EntityRelation struct {
	Subject    NamedEntity `json:"subject"`
	Predicate  string      `json:"predicate"`
	Object     NamedEntity `json:"object"`
	Confidence float64     `json:"confidence"`
}

// SentimentAnalysis represents sentiment analysis results
type SentimentAnalysis struct {
	Sentiment SentimentScore    `json:"sentiment"`
	Emotions  []EmotionScore    `json:"emotions,omitempty"`
	Aspects   []AspectSentiment `json:"aspects,omitempty"`
	Timestamp time.Time         `json:"timestamp"`
}

// EmotionScore represents emotion detection results
type EmotionScore struct {
	Emotion    string  `json:"emotion"`
	Score      float64 `json:"score"`
	Confidence float64 `json:"confidence"`
}

// AspectSentiment represents aspect-based sentiment
type AspectSentiment struct {
	Aspect    string         `json:"aspect"`
	Sentiment SentimentScore `json:"sentiment"`
}

// NLResponse represents a natural language response
type NLResponse struct {
	Text       string                 `json:"text"`
	Intent     string                 `json:"intent"`
	Context    map[string]interface{} `json:"context,omitempty"`
	Actions    []ActionSuggestion     `json:"actions,omitempty"`
	Confidence float64                `json:"confidence"`
	Timestamp  time.Time              `json:"timestamp"`
}

// CommandParsing represents command parsing results
type CommandParsing struct {
	Command    string                 `json:"command"`
	Action     string                 `json:"action"`
	Target     string                 `json:"target,omitempty"`
	Parameters map[string]interface{} `json:"parameters,omitempty"`
	Confidence float64                `json:"confidence"`
	Safe       bool                   `json:"safe"`
	Timestamp  time.Time              `json:"timestamp"`
}

// CommandValidation represents command validation results
type CommandValidation struct {
	Valid       bool                   `json:"valid"`
	Safe        bool                   `json:"safe"`
	Reason      string                 `json:"reason,omitempty"`
	Risk        string                 `json:"risk"` // low, medium, high
	Risks       []string               `json:"risks,omitempty"`
	Warnings    []string               `json:"warnings,omitempty"`
	Suggestions []string               `json:"suggestions,omitempty"`
	Confidence  float64                `json:"confidence"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	Timestamp   time.Time              `json:"timestamp"`
}

// AI Service Management Models

// ModelStatus represents the status of an AI model
type ModelStatus struct {
	ID           string                 `json:"id"`
	Status       string                 `json:"status"` // loading, loaded, unloaded, error
	LoadTime     time.Duration          `json:"load_time,omitempty"`
	MemoryUsage  int64                  `json:"memory_usage"`
	RequestCount int64                  `json:"request_count"`
	ErrorCount   int64                  `json:"error_count"`
	LastUsed     time.Time              `json:"last_used,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	Timestamp    time.Time              `json:"timestamp"`
}

// ModelMetrics represents performance metrics for an AI model
type ModelMetrics struct {
	ModelID        string        `json:"model_id"`
	RequestCount   int64         `json:"request_count"`
	SuccessCount   int64         `json:"success_count"`
	ErrorCount     int64         `json:"error_count"`
	AverageLatency time.Duration `json:"average_latency"`
	P95Latency     time.Duration `json:"p95_latency"`
	P99Latency     time.Duration `json:"p99_latency"`
	ThroughputRPS  float64       `json:"throughput_rps"`
	MemoryUsage    int64         `json:"memory_usage"`
	CPUUsage       float64       `json:"cpu_usage"`
	GPUUsage       float64       `json:"gpu_usage,omitempty"`
	Timestamp      time.Time     `json:"timestamp"`
}

// OptimizationParams represents model optimization parameters
type OptimizationParams struct {
	Target       string                 `json:"target"`              // latency, throughput, memory
	Precision    string                 `json:"precision,omitempty"` // fp32, fp16, int8
	BatchSize    int                    `json:"batch_size,omitempty"`
	MaxLength    int                    `json:"max_length,omitempty"`
	Quantization bool                   `json:"quantization,omitempty"`
	Pruning      bool                   `json:"pruning,omitempty"`
	Distillation bool                   `json:"distillation,omitempty"`
	Custom       map[string]interface{} `json:"custom,omitempty"`
}

// AIRequest represents a complex AI request
type AIRequest struct {
	ID         string                 `json:"id"`
	Type       string                 `json:"type"` // chat, vision, voice, optimization
	Input      interface{}            `json:"input"`
	Parameters map[string]interface{} `json:"parameters,omitempty"`
	Context    map[string]interface{} `json:"context,omitempty"`
	Priority   int                    `json:"priority,omitempty"`
	Timeout    time.Duration          `json:"timeout,omitempty"`
	Timestamp  time.Time              `json:"timestamp"`
}

// AIResponse represents a response from AI services
type AIResponse struct {
	RequestID      string                 `json:"request_id"`
	Type           string                 `json:"type"`
	Result         interface{}            `json:"result"`
	Confidence     float64                `json:"confidence"`
	ProcessingTime time.Duration          `json:"processing_time"`
	Model          string                 `json:"model,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
	Timestamp      time.Time              `json:"timestamp"`
}

// AIServiceStatus represents the status of all AI services
type AIServiceStatus struct {
	Services      map[string]ServiceHealth `json:"services"`
	Models        []ModelStatus            `json:"models"`
	OverallHealth string                   `json:"overall_health"` // healthy, degraded, unhealthy
	Timestamp     time.Time                `json:"timestamp"`
}

// ServiceHealth represents the health of a specific service
type ServiceHealth struct {
	Name         string        `json:"name"`
	Status       string        `json:"status"` // healthy, degraded, unhealthy
	Uptime       time.Duration `json:"uptime"`
	RequestCount int64         `json:"request_count"`
	ErrorRate    float64       `json:"error_rate"`
	Latency      time.Duration `json:"average_latency"`
	LastCheck    time.Time     `json:"last_check"`
}

// AIResult represents a result from an AI service
type AIResult struct {
	ServiceID      string                 `json:"service_id"`
	Type           string                 `json:"type"`
	Result         interface{}            `json:"result"`
	Confidence     float64                `json:"confidence"`
	ProcessingTime time.Duration          `json:"processing_time"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
	Timestamp      time.Time              `json:"timestamp"`
}

// AggregatedResult represents aggregated results from multiple AI services
type AggregatedResult struct {
	Results    []AIResult             `json:"results"`
	Consensus  interface{}            `json:"consensus,omitempty"`
	Confidence float64                `json:"confidence"`
	Method     string                 `json:"aggregation_method"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
	Timestamp  time.Time              `json:"timestamp"`
}

// AIWorkflow represents a complex AI workflow
type AIWorkflow struct {
	ID         string                 `json:"id"`
	Name       string                 `json:"name"`
	Steps      []WorkflowStep         `json:"steps"`
	Input      interface{}            `json:"input"`
	Parameters map[string]interface{} `json:"parameters,omitempty"`
	Timeout    time.Duration          `json:"timeout,omitempty"`
	Timestamp  time.Time              `json:"timestamp"`
}

// WorkflowStep represents a step in an AI workflow
type WorkflowStep struct {
	ID           string                 `json:"id"`
	Type         string                 `json:"type"`
	Service      string                 `json:"service"`
	Input        interface{}            `json:"input,omitempty"`
	Parameters   map[string]interface{} `json:"parameters,omitempty"`
	Dependencies []string               `json:"dependencies,omitempty"`
	Timeout      time.Duration          `json:"timeout,omitempty"`
}

// WorkflowResult represents the result of a workflow execution
type WorkflowResult struct {
	WorkflowID string                 `json:"workflow_id"`
	Status     string                 `json:"status"` // completed, failed, partial
	Results    map[string]interface{} `json:"results"`
	Errors     []WorkflowError        `json:"errors,omitempty"`
	Duration   time.Duration          `json:"duration"`
	Timestamp  time.Time              `json:"timestamp"`
}

// WorkflowError represents an error in workflow execution
type WorkflowError struct {
	StepID      string    `json:"step_id"`
	Error       string    `json:"error"`
	Recoverable bool      `json:"recoverable"`
	Timestamp   time.Time `json:"timestamp"`
}

// Enhanced AI Models for Streaming and Advanced Features

// LLMStreamChunk represents a chunk of streaming LLM response
type LLMStreamChunk struct {
	ID        string                 `json:"id"`
	Content   string                 `json:"content"`
	Delta     string                 `json:"delta"`
	Finished  bool                   `json:"finished"`
	TokenID   int                    `json:"token_id,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
}

// CodeStreamChunk represents a chunk of streaming code generation
type CodeStreamChunk struct {
	ID          string                 `json:"id"`
	Code        string                 `json:"code"`
	Delta       string                 `json:"delta"`
	Language    string                 `json:"language"`
	Finished    bool                   `json:"finished"`
	Explanation string                 `json:"explanation,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	Timestamp   time.Time              `json:"timestamp"`
}

// ChatStreamChunk represents a chunk of streaming chat response
type ChatStreamChunk struct {
	ID             string                 `json:"id"`
	ConversationID string                 `json:"conversation_id"`
	Content        string                 `json:"content"`
	Delta          string                 `json:"delta"`
	Role           string                 `json:"role"`
	Finished       bool                   `json:"finished"`
	FunctionCall   *FunctionCall          `json:"function_call,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
	Timestamp      time.Time              `json:"timestamp"`
}

// FunctionCall represents a function call request
type FunctionCall struct {
	Name       string                 `json:"name"`
	Parameters map[string]interface{} `json:"parameters"`
	ID         string                 `json:"id,omitempty"`
}

// FunctionCallResponse represents the response from a function call
type FunctionCallResponse struct {
	ID        string                 `json:"id"`
	Name      string                 `json:"name"`
	Result    interface{}            `json:"result"`
	Success   bool                   `json:"success"`
	Error     string                 `json:"error,omitempty"`
	Duration  time.Duration          `json:"duration"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
}

// EmbeddingResponse represents text embedding response
type EmbeddingResponse struct {
	Text      string    `json:"text"`
	Embedding []float64 `json:"embedding"`
	Model     string    `json:"model"`
	Dimension int       `json:"dimension"`
	Timestamp time.Time `json:"timestamp"`
}

// BatchEmbeddingResponse represents batch text embedding response
type BatchEmbeddingResponse struct {
	Texts      []string    `json:"texts"`
	Embeddings [][]float64 `json:"embeddings"`
	Model      string      `json:"model"`
	Dimension  int         `json:"dimension"`
	Timestamp  time.Time   `json:"timestamp"`
}

// ModelInfo represents detailed model information
type ModelInfo struct {
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	Version      string                 `json:"version"`
	Type         string                 `json:"type"`
	Provider     string                 `json:"provider"`
	Size         int64                  `json:"size"`
	Parameters   int64                  `json:"parameters"`
	Description  string                 `json:"description"`
	Capabilities []string               `json:"capabilities"`
	Languages    []string               `json:"languages,omitempty"`
	MaxTokens    int                    `json:"max_tokens"`
	ContextSize  int                    `json:"context_size"`
	Precision    string                 `json:"precision"`
	Quantization string                 `json:"quantization,omitempty"`
	Hardware     []string               `json:"hardware"`
	License      string                 `json:"license"`
	Status       string                 `json:"status"`
	LoadTime     time.Duration          `json:"load_time,omitempty"`
	MemoryUsage  int64                  `json:"memory_usage"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
}

// Multi-Modal AI Models

// MultiModalRequest represents a multi-modal AI request
type MultiModalRequest struct {
	ID         string                 `json:"id"`
	Text       string                 `json:"text,omitempty"`
	Images     [][]byte               `json:"images,omitempty"`
	Audio      []byte                 `json:"audio,omitempty"`
	Video      []byte                 `json:"video,omitempty"`
	Modalities []string               `json:"modalities"`
	Task       string                 `json:"task"`
	Parameters map[string]interface{} `json:"parameters,omitempty"`
	Timestamp  time.Time              `json:"timestamp"`
}

// MultiModalResponse represents a multi-modal AI response
type MultiModalResponse struct {
	ID         string                 `json:"id"`
	Text       string                 `json:"text,omitempty"`
	Images     [][]byte               `json:"images,omitempty"`
	Audio      []byte                 `json:"audio,omitempty"`
	Confidence float64                `json:"confidence"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
	Timestamp  time.Time              `json:"timestamp"`
}

// ImageGenerationResponse represents image generation results
type ImageGenerationResponse struct {
	Images    [][]byte               `json:"images"`
	Prompt    string                 `json:"prompt"`
	Model     string                 `json:"model"`
	Width     int                    `json:"width"`
	Height    int                    `json:"height"`
	Steps     int                    `json:"steps"`
	Guidance  float64                `json:"guidance"`
	Seed      int64                  `json:"seed,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
}

// ImageDescriptionResponse represents image description results
type ImageDescriptionResponse struct {
	Description string                 `json:"description"`
	Tags        []string               `json:"tags,omitempty"`
	Objects     []string               `json:"objects,omitempty"`
	Confidence  float64                `json:"confidence"`
	Model       string                 `json:"model"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	Timestamp   time.Time              `json:"timestamp"`
}

// VideoAnalysisResponse represents video analysis results
type VideoAnalysisResponse struct {
	Summary    string                 `json:"summary"`
	Scenes     []VideoScene           `json:"scenes"`
	Objects    []string               `json:"objects"`
	Activities []string               `json:"activities"`
	Duration   time.Duration          `json:"duration"`
	FrameRate  float64                `json:"frame_rate"`
	Resolution string                 `json:"resolution"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
	Timestamp  time.Time              `json:"timestamp"`
}

// VideoScene represents a scene in video analysis
type VideoScene struct {
	StartTime   time.Duration `json:"start_time"`
	EndTime     time.Duration `json:"end_time"`
	Description string        `json:"description"`
	Objects     []string      `json:"objects"`
	Activities  []string      `json:"activities"`
	Confidence  float64       `json:"confidence"`
}

// CrossModalSearchResponse represents cross-modal search results
type CrossModalSearchResponse struct {
	Results   []CrossModalResult     `json:"results"`
	Query     string                 `json:"query"`
	Total     int                    `json:"total"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
}

// CrossModalResult represents a single cross-modal search result
type CrossModalResult struct {
	ID       string                 `json:"id"`
	Type     string                 `json:"type"` // text, image, audio, video
	Content  interface{}            `json:"content"`
	Score    float64                `json:"score"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// RAG (Retrieval Augmented Generation) Models

// Document represents a document for RAG indexing
type Document struct {
	ID       string                 `json:"id"`
	Title    string                 `json:"title"`
	Content  string                 `json:"content"`
	Type     string                 `json:"type"` // text, pdf, html, etc.
	Source   string                 `json:"source"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
	Created  time.Time              `json:"created"`
	Updated  time.Time              `json:"updated"`
}

// DocumentSearchResponse represents document search results
type DocumentSearchResponse struct {
	Documents []DocumentResult       `json:"documents"`
	Query     string                 `json:"query"`
	Total     int                    `json:"total"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
}

// DocumentResult represents a single document search result
type DocumentResult struct {
	Document Document `json:"document"`
	Score    float64  `json:"score"`
	Snippet  string   `json:"snippet,omitempty"`
}

// RAGResponse represents a RAG generation response
type RAGResponse struct {
	Response  string                 `json:"response"`
	Sources   []DocumentResult       `json:"sources"`
	Query     string                 `json:"query"`
	Model     string                 `json:"model"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
}

// WorkflowTemplate represents a workflow template
type WorkflowTemplate struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Steps       []WorkflowStepTemplate `json:"steps"`
	Parameters  map[string]interface{} `json:"parameters,omitempty"`
	Created     time.Time              `json:"created"`
	Updated     time.Time              `json:"updated"`
}

// WorkflowStepTemplate represents a step in a workflow template
type WorkflowStepTemplate struct {
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	Type         string                 `json:"type"` // llm, cv, voice, etc.
	Service      string                 `json:"service"`
	Parameters   map[string]interface{} `json:"parameters,omitempty"`
	Dependencies []string               `json:"dependencies,omitempty"`
	Timeout      time.Duration          `json:"timeout,omitempty"`
}

// WorkflowStatus represents the status of a running workflow
type WorkflowStatus struct {
	ID          string                 `json:"id"`
	Status      string                 `json:"status"` // running, completed, failed, cancelled
	Progress    float64                `json:"progress"`
	CurrentStep string                 `json:"current_step,omitempty"`
	Results     map[string]interface{} `json:"results,omitempty"`
	Error       string                 `json:"error,omitempty"`
	Started     time.Time              `json:"started"`
	Updated     time.Time              `json:"updated"`
}

// CacheStats represents cache statistics
type CacheStats struct {
	HitRate     float64   `json:"hit_rate"`
	MissRate    float64   `json:"miss_rate"`
	TotalHits   int64     `json:"total_hits"`
	TotalMisses int64     `json:"total_misses"`
	Size        int64     `json:"size"`
	MaxSize     int64     `json:"max_size"`
	Evictions   int64     `json:"evictions"`
	Timestamp   time.Time `json:"timestamp"`
}

// Enhanced Computer Vision Models

// ObjectDetectionResponse represents object detection results
type ObjectDetectionResponse struct {
	Objects        []AdvancedDetectedObject `json:"objects"`
	TotalCount     int                      `json:"total_count"`
	Model          string                   `json:"model"`
	Confidence     float64                  `json:"confidence"`
	ProcessingTime time.Duration            `json:"processing_time"`
	Metadata       map[string]interface{}   `json:"metadata,omitempty"`
	Timestamp      time.Time                `json:"timestamp"`
}

// AdvancedDetectedObject represents a detected object with segmentation
type AdvancedDetectedObject struct {
	ID           string                 `json:"id"`
	Class        string                 `json:"class"`
	Confidence   float64                `json:"confidence"`
	BoundingBox  BoundingBox            `json:"bounding_box"`
	Segmentation []Point                `json:"segmentation,omitempty"`
	Attributes   map[string]interface{} `json:"attributes,omitempty"`
}

// BoundingBox represents object bounding box coordinates
type BoundingBox struct {
	X      int `json:"x"`
	Y      int `json:"y"`
	Width  int `json:"width"`
	Height int `json:"height"`
}

// Point represents a 2D coordinate point
type Point struct {
	X int `json:"x"`
	Y int `json:"y"`
}
