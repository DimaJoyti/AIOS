package models

import (
	"time"
)

// SystemStatus represents the overall system status
type SystemStatus struct {
	Running      bool                     `json:"running"`
	Version      string                   `json:"version"`
	Uptime       time.Duration            `json:"uptime"`
	Resources    *ResourceStatus          `json:"resources"`
	Security     *SecurityStatus          `json:"security"`
	Optimization *OptimizationStatus      `json:"optimization"`
	Services     map[string]ServiceStatus `json:"services"`
	Timestamp    time.Time                `json:"timestamp"`
}

// ResourceStatus represents system resource utilization
type ResourceStatus struct {
	CPU     *CPUStatus     `json:"cpu"`
	Memory  *MemoryStatus  `json:"memory"`
	Disk    *DiskStatus    `json:"disk"`
	Network *NetworkStatus `json:"network"`
}

// CPUStatus represents CPU utilization information
type CPUStatus struct {
	Usage       float64   `json:"usage"` // Percentage
	Cores       int       `json:"cores"`
	Temperature float64   `json:"temperature"` // Celsius
	Frequency   float64   `json:"frequency"`   // MHz
	LoadAvg     []float64 `json:"load_avg"`    // 1, 5, 15 minute averages
}

// MemoryStatus represents memory utilization information
type MemoryStatus struct {
	Total     uint64      `json:"total"`     // Bytes
	Used      uint64      `json:"used"`      // Bytes
	Available uint64      `json:"available"` // Bytes
	Usage     float64     `json:"usage"`     // Percentage
	Swap      *SwapStatus `json:"swap"`
}

// SwapStatus represents swap memory information
type SwapStatus struct {
	Total uint64  `json:"total"` // Bytes
	Used  uint64  `json:"used"`  // Bytes
	Usage float64 `json:"usage"` // Percentage
}

// DiskStatus represents disk utilization information
type DiskStatus struct {
	Filesystems []FilesystemStatus `json:"filesystems"`
	IOStats     *DiskIOStats       `json:"io_stats"`
}

// FilesystemStatus represents individual filesystem status
type FilesystemStatus struct {
	Device     string  `json:"device"`
	Mountpoint string  `json:"mountpoint"`
	Type       string  `json:"type"`
	Total      uint64  `json:"total"`     // Bytes
	Used       uint64  `json:"used"`      // Bytes
	Available  uint64  `json:"available"` // Bytes
	Usage      float64 `json:"usage"`     // Percentage
}

// DiskIOStats represents disk I/O statistics
type DiskIOStats struct {
	ReadBytes  uint64 `json:"read_bytes"`
	WriteBytes uint64 `json:"write_bytes"`
	ReadOps    uint64 `json:"read_ops"`
	WriteOps   uint64 `json:"write_ops"`
}

// NetworkStatus represents network utilization information
type NetworkStatus struct {
	Interfaces  []NetworkInterface  `json:"interfaces"`
	Connections *NetworkConnections `json:"connections"`
}

// NetworkInterface represents network interface statistics
type NetworkInterface struct {
	Name        string `json:"name"`
	BytesRecv   uint64 `json:"bytes_recv"`
	BytesSent   uint64 `json:"bytes_sent"`
	PacketsRecv uint64 `json:"packets_recv"`
	PacketsSent uint64 `json:"packets_sent"`
	Errors      uint64 `json:"errors"`
	Drops       uint64 `json:"drops"`
}

// NetworkConnections represents network connection statistics
type NetworkConnections struct {
	TCP       uint64 `json:"tcp"`
	UDP       uint64 `json:"udp"`
	Listening uint64 `json:"listening"`
}

// SecurityStatus represents security system status
type SecurityStatus struct {
	ThreatLevel    string           `json:"threat_level"` // low, medium, high, critical
	ActiveThreats  int              `json:"active_threats"`
	BlockedAttacks int              `json:"blocked_attacks"`
	LastScan       time.Time        `json:"last_scan"`
	Firewall       *FirewallStatus  `json:"firewall"`
	Antivirus      *AntivirusStatus `json:"antivirus"`
}

// FirewallStatus represents firewall status
type FirewallStatus struct {
	Enabled      bool  `json:"enabled"`
	Rules        int   `json:"rules"`
	BlockedIPs   int   `json:"blocked_ips"`
	AllowedPorts []int `json:"allowed_ports"`
	BlockedPorts []int `json:"blocked_ports"`
}

// AntivirusStatus represents antivirus status
type AntivirusStatus struct {
	Enabled          bool      `json:"enabled"`
	LastUpdate       time.Time `json:"last_update"`
	DefinitionsCount int       `json:"definitions_count"`
	QuarantinedFiles int       `json:"quarantined_files"`
}

// OptimizationStatus represents AI optimization status
type OptimizationStatus struct {
	Enabled          bool                         `json:"enabled"`
	LastOptimization time.Time                    `json:"last_optimization"`
	OptimizationsRun int                          `json:"optimizations_run"`
	PerformanceGain  float64                      `json:"performance_gain"` // Percentage
	Recommendations  []OptimizationRecommendation `json:"recommendations"`
}

// OptimizationRecommendation represents an AI optimization recommendation
type OptimizationRecommendation struct {
	ID          string    `json:"id"`
	Type        string    `json:"type"`     // cpu, memory, disk, network
	Priority    string    `json:"priority"` // low, medium, high
	Description string    `json:"description"`
	Impact      string    `json:"impact"` // Expected impact description
	Applied     bool      `json:"applied"`
	CreatedAt   time.Time `json:"created_at"`
}

// ServiceStatus represents individual service status
type ServiceStatus struct {
	Name      string        `json:"name"`
	Status    string        `json:"status"` // running, stopped, error
	Health    string        `json:"health"` // healthy, unhealthy, unknown
	Uptime    time.Duration `json:"uptime"`
	CPU       float64       `json:"cpu"`    // Percentage
	Memory    uint64        `json:"memory"` // Bytes
	LastCheck time.Time     `json:"last_check"`
}

// AIModelStatus represents AI model status
type AIModelStatus struct {
	Name         string        `json:"name"`
	Version      string        `json:"version"`
	Status       string        `json:"status"` // loaded, loading, error
	Type         string        `json:"type"`   // llm, cv, optimization
	Size         uint64        `json:"size"`   // Bytes
	LoadTime     time.Duration `json:"load_time"`
	LastUsed     time.Time     `json:"last_used"`
	RequestCount int64         `json:"request_count"`
	AvgLatency   time.Duration `json:"avg_latency"`
}

// FileSystemAnalysis represents file system AI analysis results
type FileSystemAnalysis struct {
	Path            string           `json:"path"`
	TotalFiles      int              `json:"total_files"`
	TotalSize       uint64           `json:"total_size"`
	FileTypes       map[string]int   `json:"file_types"`
	LargestFiles    []FileInfo       `json:"largest_files"`
	DuplicateFiles  []DuplicateGroup `json:"duplicate_files"`
	UnusedFiles     []FileInfo       `json:"unused_files"`
	Recommendations []string         `json:"recommendations"`
	AnalyzedAt      time.Time        `json:"analyzed_at"`
}

// FileInfo represents file information
type FileInfo struct {
	Path        string    `json:"path"`
	Size        uint64    `json:"size"`
	ModTime     time.Time `json:"mod_time"`
	AccessTime  time.Time `json:"access_time"`
	Type        string    `json:"type"`
	Permissions string    `json:"permissions"`
}

// DuplicateGroup represents a group of duplicate files
type DuplicateGroup struct {
	Hash  string     `json:"hash"`
	Size  uint64     `json:"size"`
	Files []FileInfo `json:"files"`
}

// ThreatAnalysis represents security threat analysis
type ThreatAnalysis struct {
	Threats         []ThreatInfo `json:"threats"`
	RiskScore       float64      `json:"risk_score"` // 0-100
	Severity        string       `json:"severity"`   // low, medium, high, critical
	AnalyzedAt      time.Time    `json:"analyzed_at"`
	Recommendations []string     `json:"recommendations"`
}

// ThreatInfo represents individual threat information
type ThreatInfo struct {
	ID          string    `json:"id"`
	Type        string    `json:"type"`     // malware, intrusion, anomaly
	Severity    string    `json:"severity"` // low, medium, high, critical
	Source      string    `json:"source"`   // IP, process, file
	Description string    `json:"description"`
	FirstSeen   time.Time `json:"first_seen"`
	LastSeen    time.Time `json:"last_seen"`
	Count       int       `json:"count"`
	Blocked     bool      `json:"blocked"`
}

// PerformanceMetrics represents system performance metrics
type PerformanceMetrics struct {
	Timestamp    time.Time     `json:"timestamp"`
	CPUUsage     float64       `json:"cpu_usage"`
	MemoryUsage  float64       `json:"memory_usage"`
	DiskUsage    float64       `json:"disk_usage"`
	NetworkIn    uint64        `json:"network_in"`
	NetworkOut   uint64        `json:"network_out"`
	ResponseTime time.Duration `json:"response_time"`
	Throughput   float64       `json:"throughput"` // Requests per second
}

// Configuration represents system configuration
type Configuration struct {
	Server     ServerConfig     `json:"server"`
	Database   DatabaseConfig   `json:"database"`
	Redis      RedisConfig      `json:"redis"`
	AI         AIConfig         `json:"ai"`
	Security   SecurityConfig   `json:"security"`
	Logging    LoggingConfig    `json:"logging"`
	Tracing    TracingConfig    `json:"tracing"`
	FileSystem FileSystemConfig `json:"filesystem"`
	Desktop    DesktopConfig    `json:"desktop"`
}

// ServerConfig represents server configuration
type ServerConfig struct {
	Host         string        `json:"host"`
	Port         int           `json:"port"`
	MetricsPort  int           `json:"metrics_port"`
	ReadTimeout  time.Duration `json:"read_timeout"`
	WriteTimeout time.Duration `json:"write_timeout"`
	IdleTimeout  time.Duration `json:"idle_timeout"`
}

// DatabaseConfig represents database configuration
type DatabaseConfig struct {
	Host            string        `json:"host"`
	Port            int           `json:"port"`
	Name            string        `json:"name"`
	User            string        `json:"user"`
	Password        string        `json:"password"`
	SSLMode         string        `json:"ssl_mode"`
	MaxOpenConns    int           `json:"max_open_conns"`
	MaxIdleConns    int           `json:"max_idle_conns"`
	ConnMaxLifetime time.Duration `json:"conn_max_lifetime"`
}

// RedisConfig represents Redis configuration
type RedisConfig struct {
	Host         string `json:"host"`
	Port         int    `json:"port"`
	Password     string `json:"password"`
	DB           int    `json:"db"`
	PoolSize     int    `json:"pool_size"`
	MinIdleConns int    `json:"min_idle_conns"`
}

// AIConfig represents AI services configuration
type AIConfig struct {
	Ollama       OllamaConfig `json:"ollama"`
	ModelsPath   string       `json:"models_path"`
	DefaultModel string       `json:"default_model"`
	MaxTokens    int          `json:"max_tokens"`
	Temperature  float64      `json:"temperature"`
}

// OllamaConfig represents Ollama configuration
type OllamaConfig struct {
	Host    string        `json:"host"`
	Port    int           `json:"port"`
	Timeout time.Duration `json:"timeout"`
}

// SecurityConfig represents security configuration
type SecurityConfig struct {
	JWTSecret      string          `json:"jwt_secret"`
	SessionTimeout time.Duration   `json:"session_timeout"`
	RateLimit      RateLimitConfig `json:"rate_limit"`
	CORS           CORSConfig      `json:"cors"`
}

// RateLimitConfig represents rate limiting configuration
type RateLimitConfig struct {
	RequestsPerMinute int `json:"requests_per_minute"`
	Burst             int `json:"burst"`
}

// CORSConfig represents CORS configuration
type CORSConfig struct {
	AllowedOrigins []string `json:"allowed_origins"`
	AllowedMethods []string `json:"allowed_methods"`
	AllowedHeaders []string `json:"allowed_headers"`
}

// LoggingConfig represents logging configuration
type LoggingConfig struct {
	Level      string `json:"level"`
	Format     string `json:"format"`
	Output     string `json:"output"`
	FilePath   string `json:"file_path"`
	MaxSize    string `json:"max_size"`
	MaxBackups int    `json:"max_backups"`
	MaxAge     int    `json:"max_age"`
}

// TracingConfig represents tracing configuration
type TracingConfig struct {
	Enabled        bool    `json:"enabled"`
	JaegerEndpoint string  `json:"jaeger_endpoint"`
	ServiceName    string  `json:"service_name"`
	SampleRate     float64 `json:"sample_rate"`
}

// FileSystemConfig represents file system configuration
type FileSystemConfig struct {
	WatchPaths          []string      `json:"watch_paths"`
	IgnorePatterns      []string      `json:"ignore_patterns"`
	AnalysisInterval    time.Duration `json:"analysis_interval"`
	OrganizationEnabled bool          `json:"organization_enabled"`
}

// DesktopConfig represents desktop environment configuration
type DesktopConfig struct {
	Theme         string              `json:"theme"`
	AIAssistant   AIAssistantConfig   `json:"ai_assistant"`
	WindowManager WindowManagerConfig `json:"window_manager"`
}

// AIAssistantConfig represents AI assistant configuration
type AIAssistantConfig struct {
	VoiceEnabled bool   `json:"voice_enabled"`
	WakeWord     string `json:"wake_word"`
	Language     string `json:"language"`
}

// WindowManagerConfig represents window manager configuration
type WindowManagerConfig struct {
	TilingEnabled bool `json:"tiling_enabled"`
	SmartGaps     bool `json:"smart_gaps"`
	BorderWidth   int  `json:"border_width"`
}

// AI Service Models

// LLMResponse represents a response from a language model
type LLMResponse struct {
	Text           string                 `json:"text"`
	Confidence     float64                `json:"confidence"`
	TokensUsed     int                    `json:"tokens_used"`
	Model          string                 `json:"model"`
	ProcessingTime time.Duration          `json:"processing_time"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
	Timestamp      time.Time              `json:"timestamp"`
}

// CodeResponse represents a code generation response
type CodeResponse struct {
	Code        string    `json:"code"`
	Language    string    `json:"language"`
	Explanation string    `json:"explanation"`
	Confidence  float64   `json:"confidence"`
	Suggestions []string  `json:"suggestions,omitempty"`
	Timestamp   time.Time `json:"timestamp"`
}

// TextAnalysis represents text analysis results
type TextAnalysis struct {
	Summary    string                 `json:"summary"`
	Keywords   []string               `json:"keywords"`
	Entities   []NamedEntity          `json:"entities"`
	Sentiment  SentimentScore         `json:"sentiment"`
	Language   string                 `json:"language"`
	Complexity float64                `json:"complexity"`
	Topics     []string               `json:"topics"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
	Timestamp  time.Time              `json:"timestamp"`
}

// ChatResponse represents a chat conversation response
type ChatResponse struct {
	Message        string                 `json:"message"`
	ConversationID string                 `json:"conversation_id"`
	Context        map[string]interface{} `json:"context,omitempty"`
	Suggestions    []string               `json:"suggestions,omitempty"`
	Actions        []ActionSuggestion     `json:"actions,omitempty"`
	Confidence     float64                `json:"confidence"`
	Timestamp      time.Time              `json:"timestamp"`
}

// SummaryResponse represents a text summarization response
type SummaryResponse struct {
	Summary     string    `json:"summary"`
	KeyPoints   []string  `json:"key_points"`
	Length      int       `json:"length"`
	Compression float64   `json:"compression_ratio"`
	Timestamp   time.Time `json:"timestamp"`
}

// TranslationResponse represents a translation response
type TranslationResponse struct {
	TranslatedText string    `json:"translated_text"`
	FromLanguage   string    `json:"from_language"`
	ToLanguage     string    `json:"to_language"`
	Confidence     float64   `json:"confidence"`
	Timestamp      time.Time `json:"timestamp"`
}

// AIModel represents an AI model
type AIModel struct {
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	Version      string                 `json:"version"`
	Type         string                 `json:"type"` // llm, cv, voice, etc.
	Size         int64                  `json:"size"`
	Description  string                 `json:"description"`
	Capabilities []string               `json:"capabilities"`
	Status       string                 `json:"status"` // available, loaded, loading, error
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
}

// ScreenAnalysis represents screen analysis results
type ScreenAnalysis struct {
	Elements      []UIElement            `json:"elements"`
	Layout        LayoutInfo             `json:"layout"`
	Text          []TextRegion           `json:"text"`
	Actions       []PossibleAction       `json:"actions"`
	Accessibility AccessibilityInfo      `json:"accessibility"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
	Timestamp     time.Time              `json:"timestamp"`
}

// UIElements represents detected UI elements
type UIElements struct {
	Buttons    []UIElement `json:"buttons"`
	TextFields []UIElement `json:"text_fields"`
	Images     []UIElement `json:"images"`
	Links      []UIElement `json:"links"`
	Menus      []UIElement `json:"menus"`
	Windows    []UIElement `json:"windows"`
	Other      []UIElement `json:"other"`
}

// UIElement represents a single UI element
type UIElement struct {
	ID         string                 `json:"id"`
	Type       string                 `json:"type"`
	Text       string                 `json:"text,omitempty"`
	Bounds     Rectangle              `json:"bounds"`
	Confidence float64                `json:"confidence"`
	Properties map[string]interface{} `json:"properties,omitempty"`
}

// Rectangle represents a rectangular area
type Rectangle struct {
	X      int `json:"x"`
	Y      int `json:"y"`
	Width  int `json:"width"`
	Height int `json:"height"`
}

// TextRecognition represents OCR results
type TextRecognition struct {
	Text       string       `json:"text"`
	Regions    []TextRegion `json:"regions"`
	Language   string       `json:"language"`
	Confidence float64      `json:"confidence"`
	Timestamp  time.Time    `json:"timestamp"`
}

// TextRegion represents a region of recognized text
type TextRegion struct {
	Text       string    `json:"text"`
	Bounds     Rectangle `json:"bounds"`
	Confidence float64   `json:"confidence"`
	Language   string    `json:"language,omitempty"`
}

// ImageClassification represents image classification results
type ImageClassification struct {
	Classes    []ClassificationResult `json:"classes"`
	TopClass   string                 `json:"top_class"`
	Confidence float64                `json:"confidence"`
	Timestamp  time.Time              `json:"timestamp"`
}

// ClassificationResult represents a single classification result
type ClassificationResult struct {
	Class       string  `json:"class"`
	Confidence  float64 `json:"confidence"`
	Probability float64 `json:"probability"`
}

// ObjectDetection represents object detection results
type ObjectDetection struct {
	Objects   []DetectedObject `json:"objects"`
	Count     int              `json:"count"`
	Timestamp time.Time        `json:"timestamp"`
}

// DetectedObject represents a detected object
type DetectedObject struct {
	Class      string                 `json:"class"`
	Confidence float64                `json:"confidence"`
	Bounds     Rectangle              `json:"bounds"`
	Properties map[string]interface{} `json:"properties,omitempty"`
}

// LayoutAnalysis represents layout analysis results
type LayoutAnalysis struct {
	Structure   LayoutStructure        `json:"structure"`
	Hierarchy   []LayoutNode           `json:"hierarchy"`
	Patterns    []LayoutPattern        `json:"patterns"`
	Suggestions []LayoutSuggestion     `json:"suggestions"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	Timestamp   time.Time              `json:"timestamp"`
}

// LayoutStructure represents the overall layout structure
type LayoutStructure struct {
	Type      string      `json:"type"` // grid, flex, absolute, etc.
	Columns   int         `json:"columns,omitempty"`
	Rows      int         `json:"rows,omitempty"`
	Regions   []Rectangle `json:"regions"`
	Alignment string      `json:"alignment"`
	Spacing   int         `json:"spacing"`
}

// LayoutNode represents a node in the layout hierarchy
type LayoutNode struct {
	ID       string       `json:"id"`
	Type     string       `json:"type"`
	Bounds   Rectangle    `json:"bounds"`
	Children []LayoutNode `json:"children,omitempty"`
	Parent   string       `json:"parent,omitempty"`
}

// LayoutPattern represents a detected layout pattern
type LayoutPattern struct {
	Type        string      `json:"type"`
	Confidence  float64     `json:"confidence"`
	Description string      `json:"description"`
	Examples    []Rectangle `json:"examples"`
}

// LayoutSuggestion represents a layout improvement suggestion
type LayoutSuggestion struct {
	Type        string `json:"type"`
	Description string `json:"description"`
	Impact      string `json:"impact"`
	Priority    string `json:"priority"`
}

// ImageDescription represents an image description
type ImageDescription struct {
	Description string                 `json:"description"`
	Details     []string               `json:"details"`
	Objects     []string               `json:"objects"`
	Scene       string                 `json:"scene"`
	Mood        string                 `json:"mood,omitempty"`
	Colors      []string               `json:"colors"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	Timestamp   time.Time              `json:"timestamp"`
}

// ImageComparison represents image comparison results
type ImageComparison struct {
	Similarity   float64                `json:"similarity"`
	Differences  []ImageDifference      `json:"differences"`
	MatchRegions []Rectangle            `json:"match_regions"`
	Analysis     string                 `json:"analysis"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	Timestamp    time.Time              `json:"timestamp"`
}

// ImageDifference represents a difference between two images
type ImageDifference struct {
	Type        string    `json:"type"`
	Region      Rectangle `json:"region"`
	Description string    `json:"description"`
	Severity    string    `json:"severity"`
}

// Supporting types for AI services

// NamedEntity represents a named entity in text
type NamedEntity struct {
	Text       string  `json:"text"`
	Type       string  `json:"type"` // PERSON, ORGANIZATION, LOCATION, etc.
	Confidence float64 `json:"confidence"`
	StartPos   int     `json:"start_pos"`
	EndPos     int     `json:"end_pos"`
}

// SentimentScore represents sentiment analysis results
type SentimentScore struct {
	Score      float64 `json:"score"` // -1.0 to 1.0
	Label      string  `json:"label"` // positive, negative, neutral
	Confidence float64 `json:"confidence"`
}

// ActionSuggestion represents a suggested action
type ActionSuggestion struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Description string                 `json:"description"`
	Command     string                 `json:"command,omitempty"`
	Parameters  map[string]interface{} `json:"parameters,omitempty"`
	Confidence  float64                `json:"confidence"`
}

// LayoutInfo represents layout information
type LayoutInfo struct {
	Type        string      `json:"type"`
	Dimensions  Rectangle   `json:"dimensions"`
	Regions     []Rectangle `json:"regions"`
	Orientation string      `json:"orientation"`
	Density     float64     `json:"density"`
}

// PossibleAction represents a possible action on a UI element
type PossibleAction struct {
	Type        string                 `json:"type"` // click, type, scroll, etc.
	Target      string                 `json:"target"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters,omitempty"`
	Confidence  float64                `json:"confidence"`
}

// AccessibilityInfo represents accessibility information
type AccessibilityInfo struct {
	Score       float64  `json:"score"`
	Issues      []string `json:"issues"`
	Suggestions []string `json:"suggestions"`
	Compliance  string   `json:"compliance"` // WCAG level
}

// Desktop Environment Models

// DesktopStatus represents the overall desktop environment status
type DesktopStatus struct {
	Running       bool                 `json:"running"`
	Version       string               `json:"version"`
	Theme         string               `json:"theme"`
	Windows       *WindowManagerStatus `json:"windows"`
	Workspaces    *WorkspaceStatus     `json:"workspaces"`
	Applications  *ApplicationStatus   `json:"applications"`
	Themes        *ThemeStatus         `json:"themes"`
	Notifications *NotificationStatus  `json:"notifications"`
	Performance   *DesktopPerformance  `json:"performance"`
	Timestamp     time.Time            `json:"timestamp"`
}

// DesktopPerformance represents desktop performance metrics
type DesktopPerformance struct {
	FPS           float64       `json:"fps"`
	MemoryUsage   int64         `json:"memory_usage"`
	CPUUsage      float64       `json:"cpu_usage"`
	GPUUsage      float64       `json:"gpu_usage"`
	CompositorLag time.Duration `json:"compositor_lag"`
}

// Window represents a desktop window
type Window struct {
	ID          string                 `json:"id"`
	Title       string                 `json:"title"`
	Application string                 `json:"application"`
	PID         int                    `json:"pid"`
	Position    Position               `json:"position"`
	Size        Size                   `json:"size"`
	Workspace   int                    `json:"workspace"`
	Focused     bool                   `json:"focused"`
	Visible     bool                   `json:"visible"`
	Minimized   bool                   `json:"minimized"`
	Maximized   bool                   `json:"maximized"`
	Fullscreen  bool                   `json:"fullscreen"`
	Tags        []string               `json:"tags"`
	Properties  map[string]interface{} `json:"properties,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	LastFocused time.Time              `json:"last_focused"`
}

// Position represents a 2D position
type Position struct {
	X int `json:"x"`
	Y int `json:"y"`
}

// Size represents 2D dimensions
type Size struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

// WindowManagerStatus represents window manager status
type WindowManagerStatus struct {
	Running     bool                `json:"running"`
	WindowCount int                 `json:"window_count"`
	Windows     []*Window           `json:"windows"`
	Layouts     []*WindowLayout     `json:"layouts"`
	Rules       []WindowRule        `json:"rules"`
	Config      WindowManagerConfig `json:"config"`
	Timestamp   time.Time           `json:"timestamp"`
}

// WindowLayout represents a saved window layout
type WindowLayout struct {
	ID        string             `json:"id"`
	Name      string             `json:"name"`
	Windows   []WindowLayoutItem `json:"windows"`
	CreatedAt time.Time          `json:"created_at"`
	UpdatedAt time.Time          `json:"updated_at"`
}

// WindowLayoutItem represents a window in a layout
type WindowLayoutItem struct {
	WindowID  string   `json:"window_id"`
	Position  Position `json:"position"`
	Size      Size     `json:"size"`
	Workspace int      `json:"workspace"`
	Visible   bool     `json:"visible"`
}

// WindowRule represents a window management rule
type WindowRule struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Priority    int    `json:"priority"`
	Enabled     bool   `json:"enabled"`
	AIGenerated bool   `json:"ai_generated"`
}

// Workspace represents a desktop workspace
type Workspace struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Active      bool      `json:"active"`
	WindowCount int       `json:"window_count"`
	Windows     []string  `json:"windows"` // Window IDs
	Layout      string    `json:"layout"`
	CreatedAt   time.Time `json:"created_at"`
	LastUsed    time.Time `json:"last_used"`
}

// WorkspaceStatus represents workspace manager status
type WorkspaceStatus struct {
	Running         bool         `json:"running"`
	ActiveWorkspace int          `json:"active_workspace"`
	WorkspaceCount  int          `json:"workspace_count"`
	Workspaces      []*Workspace `json:"workspaces"`
	Timestamp       time.Time    `json:"timestamp"`
}

// Application represents a desktop application
type Application struct {
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	DisplayName  string                 `json:"display_name"`
	Description  string                 `json:"description"`
	Icon         string                 `json:"icon"`
	Category     string                 `json:"category"`
	Executable   string                 `json:"executable"`
	Keywords     []string               `json:"keywords"`
	MimeTypes    []string               `json:"mime_types"`
	Running      bool                   `json:"running"`
	Windows      []string               `json:"windows"` // Window IDs
	LaunchCount  int                    `json:"launch_count"`
	LastLaunched time.Time              `json:"last_launched"`
	Properties   map[string]interface{} `json:"properties,omitempty"`
}

// ApplicationStatus represents application launcher status
type ApplicationStatus struct {
	Running          bool           `json:"running"`
	ApplicationCount int            `json:"application_count"`
	Applications     []*Application `json:"applications"`
	RecentApps       []*Application `json:"recent_apps"`
	FavoriteApps     []*Application `json:"favorite_apps"`
	Timestamp        time.Time      `json:"timestamp"`
}

// Theme represents a desktop theme
type Theme struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Author      string                 `json:"author"`
	Version     string                 `json:"version"`
	Colors      map[string]string      `json:"colors"`
	Fonts       map[string]string      `json:"fonts"`
	Icons       string                 `json:"icons"`
	Wallpaper   string                 `json:"wallpaper"`
	Active      bool                   `json:"active"`
	Properties  map[string]interface{} `json:"properties,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
}

// ThemeStatus represents theme manager status
type ThemeStatus struct {
	Running     bool      `json:"running"`
	ActiveTheme string    `json:"active_theme"`
	ThemeCount  int       `json:"theme_count"`
	Themes      []*Theme  `json:"themes"`
	Timestamp   time.Time `json:"timestamp"`
}

// Notification represents a desktop notification
type Notification struct {
	ID          string                 `json:"id"`
	Title       string                 `json:"title"`
	Body        string                 `json:"body"`
	Icon        string                 `json:"icon"`
	Category    string                 `json:"category"`
	Priority    string                 `json:"priority"` // low, normal, high, critical
	Timeout     time.Duration          `json:"timeout"`
	Actions     []NotificationAction   `json:"actions"`
	Source      string                 `json:"source"`
	Persistent  bool                   `json:"persistent"`
	Dismissed   bool                   `json:"dismissed"`
	Properties  map[string]interface{} `json:"properties,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	DismissedAt *time.Time             `json:"dismissed_at,omitempty"`
}

// NotificationAction represents an action in a notification
type NotificationAction struct {
	ID    string `json:"id"`
	Label string `json:"label"`
	Icon  string `json:"icon,omitempty"`
}

// NotificationStatus represents notification manager status
type NotificationStatus struct {
	Running           bool            `json:"running"`
	NotificationCount int             `json:"notification_count"`
	Notifications     []*Notification `json:"notifications"`
	RecentCount       int             `json:"recent_count"`
	Timestamp         time.Time       `json:"timestamp"`
}
