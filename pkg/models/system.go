package models

import (
	"time"
)

// User represents a system user
type User struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	IsActive  bool      `json:"is_active"`
	IsAdmin   bool      `json:"is_admin"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// UserPreferences represents user preferences
type UserPreferences struct {
	UserID                string            `json:"user_id"`
	Theme                 string            `json:"theme"`
	Language              string            `json:"language"`
	Notifications         bool              `json:"notifications"`
	AutoSave              bool              `json:"auto_save"`
	PreferredAIModel      string            `json:"preferred_ai_model"`
	VoiceSettings         map[string]interface{} `json:"voice_settings"`
	DesktopLayout         map[string]interface{} `json:"desktop_layout"`
	AccessibilitySettings map[string]interface{} `json:"accessibility_settings"`
	CreatedAt             time.Time         `json:"created_at"`
	UpdatedAt             time.Time         `json:"updated_at"`
}

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

// Developer Tools Models

// DevToolsStatus represents the overall developer tools status
type DevToolsStatus struct {
	Enabled          bool                    `json:"enabled"`
	Running          bool                    `json:"running"`
	Debugger         *DebuggerStatus         `json:"debugger"`
	Profiler         *ProfilerStatus         `json:"profiler"`
	CodeAnalyzer     *CodeAnalyzerStatus     `json:"code_analyzer"`
	TestRunner       *TestRunnerStatus       `json:"test_runner"`
	BuildManager     *BuildManagerStatus     `json:"build_manager"`
	LiveReloader     *LiveReloaderStatus     `json:"live_reloader"`
	LogAnalyzer      *LogAnalyzerStatus      `json:"log_analyzer"`
	MetricsCollector *MetricsCollectorStatus `json:"metrics_collector"`
	Timestamp        time.Time               `json:"timestamp"`
}

// DebuggerStatus represents debugger status
type DebuggerStatus struct {
	Enabled         bool            `json:"enabled"`
	Running         bool            `json:"running"`
	Port            int             `json:"port"`
	RemoteDebugging bool            `json:"remote_debugging"`
	Breakpoints     []*Breakpoint   `json:"breakpoints"`
	Sessions        []*DebugSession `json:"sessions"`
	Timestamp       time.Time       `json:"timestamp"`
}

// Breakpoint represents a debug breakpoint
type Breakpoint struct {
	ID        string    `json:"id"`
	File      string    `json:"file"`
	Line      int       `json:"line"`
	Condition string    `json:"condition"`
	Enabled   bool      `json:"enabled"`
	HitCount  int       `json:"hit_count"`
	CreatedAt time.Time `json:"created_at"`
}

// DebugSession represents a debug session
type DebugSession struct {
	ID        string                 `json:"id"`
	Target    string                 `json:"target"`
	Active    bool                   `json:"active"`
	StartTime time.Time              `json:"start_time"`
	EndTime   time.Time              `json:"end_time,omitempty"`
	Variables map[string]interface{} `json:"variables"`
	CallStack []StackFrame           `json:"call_stack"`
}

// StackFrame represents a stack frame
type StackFrame struct {
	Function string  `json:"function"`
	File     string  `json:"file"`
	Line     int     `json:"line"`
	PC       uintptr `json:"pc"`
}

// ProfilerStatus represents profiler status
type ProfilerStatus struct {
	Enabled            bool       `json:"enabled"`
	Running            bool       `json:"running"`
	CPUProfiling       bool       `json:"cpu_profiling"`
	MemoryProfiling    bool       `json:"memory_profiling"`
	GoroutineProfiling bool       `json:"goroutine_profiling"`
	BlockProfiling     bool       `json:"block_profiling"`
	MutexProfiling     bool       `json:"mutex_profiling"`
	Profiles           []*Profile `json:"profiles"`
	Timestamp          time.Time  `json:"timestamp"`
}

// Profile represents a performance profile
type Profile struct {
	ID        string        `json:"id"`
	Type      string        `json:"type"`
	Filename  string        `json:"filename"`
	StartTime time.Time     `json:"start_time"`
	EndTime   time.Time     `json:"end_time,omitempty"`
	Duration  time.Duration `json:"duration"`
	Active    bool          `json:"active"`
}

// RuntimeStats represents runtime statistics
type RuntimeStats struct {
	Goroutines    int       `json:"goroutines"`
	CGoCalls      int64     `json:"cgo_calls"`
	HeapAlloc     uint64    `json:"heap_alloc"`
	HeapSys       uint64    `json:"heap_sys"`
	HeapIdle      uint64    `json:"heap_idle"`
	HeapInuse     uint64    `json:"heap_inuse"`
	HeapReleased  uint64    `json:"heap_released"`
	HeapObjects   uint64    `json:"heap_objects"`
	StackInuse    uint64    `json:"stack_inuse"`
	StackSys      uint64    `json:"stack_sys"`
	MSpanInuse    uint64    `json:"mspan_inuse"`
	MSpanSys      uint64    `json:"mspan_sys"`
	MCacheInuse   uint64    `json:"mcache_inuse"`
	MCacheSys     uint64    `json:"mcache_sys"`
	GCSys         uint64    `json:"gc_sys"`
	OtherSys      uint64    `json:"other_sys"`
	NextGC        uint64    `json:"next_gc"`
	LastGC        time.Time `json:"last_gc"`
	PauseTotalNs  uint64    `json:"pause_total_ns"`
	NumGC         uint32    `json:"num_gc"`
	NumForcedGC   uint32    `json:"num_forced_gc"`
	GCCPUFraction float64   `json:"gc_cpu_fraction"`
	Timestamp     time.Time `json:"timestamp"`
}

// CodeAnalyzerStatus represents code analyzer status
type CodeAnalyzerStatus struct {
	Enabled         bool            `json:"enabled"`
	Running         bool            `json:"running"`
	StaticAnalysis  bool            `json:"static_analysis"`
	SecurityScan    bool            `json:"security_scan"`
	QualityMetrics  bool            `json:"quality_metrics"`
	DependencyCheck bool            `json:"dependency_check"`
	Analyses        []*CodeAnalysis `json:"analyses"`
	Timestamp       time.Time       `json:"timestamp"`
}

// CodeAnalysis represents a code analysis result
type CodeAnalysis struct {
	ID        string        `json:"id"`
	Type      string        `json:"type"`
	Path      string        `json:"path"`
	StartTime time.Time     `json:"start_time"`
	EndTime   time.Time     `json:"end_time,omitempty"`
	Duration  time.Duration `json:"duration"`
	Status    string        `json:"status"`
	Issues    []CodeIssue   `json:"issues"`
	Metrics   CodeMetrics   `json:"metrics"`
}

// CodeIssue represents a code issue
type CodeIssue struct {
	Type       string `json:"type"`
	Severity   string `json:"severity"`
	Message    string `json:"message"`
	File       string `json:"file"`
	Line       int    `json:"line"`
	Column     int    `json:"column"`
	Rule       string `json:"rule"`
	Suggestion string `json:"suggestion"`
}

// CodeMetrics represents code quality metrics
type CodeMetrics struct {
	LinesOfCode     int     `json:"lines_of_code"`
	Functions       int     `json:"functions"`
	Complexity      float64 `json:"complexity"`
	TestCoverage    float64 `json:"test_coverage"`
	Maintainability float64 `json:"maintainability"`
	Duplication     float64 `json:"duplication"`
}

// TestRunnerStatus represents test runner status
type TestRunnerStatus struct {
	Enabled     bool       `json:"enabled"`
	Running     bool       `json:"running"`
	AutoRun     bool       `json:"auto_run"`
	Coverage    bool       `json:"coverage"`
	Benchmarks  bool       `json:"benchmarks"`
	Integration bool       `json:"integration"`
	E2E         bool       `json:"e2e"`
	TestRuns    []*TestRun `json:"test_runs"`
	Timestamp   time.Time  `json:"timestamp"`
}

// TestRun represents a test execution
type TestRun struct {
	ID        string        `json:"id"`
	Type      string        `json:"type"`
	StartTime time.Time     `json:"start_time"`
	EndTime   time.Time     `json:"end_time,omitempty"`
	Duration  time.Duration `json:"duration"`
	Status    string        `json:"status"`
	Results   TestResults   `json:"results"`
}

// TestResults represents test execution results
type TestResults struct {
	Total    int     `json:"total"`
	Passed   int     `json:"passed"`
	Failed   int     `json:"failed"`
	Skipped  int     `json:"skipped"`
	Coverage float64 `json:"coverage"`
	Output   string  `json:"output"`
}

// BuildManagerStatus represents build manager status
type BuildManagerStatus struct {
	Enabled   bool      `json:"enabled"`
	Running   bool      `json:"running"`
	AutoBuild bool      `json:"auto_build"`
	Builds    []*Build  `json:"builds"`
	Timestamp time.Time `json:"timestamp"`
}

// Build represents a build execution
type Build struct {
	ID        string        `json:"id"`
	Target    string        `json:"target"`
	StartTime time.Time     `json:"start_time"`
	EndTime   time.Time     `json:"end_time,omitempty"`
	Duration  time.Duration `json:"duration"`
	Status    string        `json:"status"`
	Output    string        `json:"output"`
	Artifacts []string      `json:"artifacts"`
}

// LiveReloaderStatus represents live reloader status
type LiveReloaderStatus struct {
	Enabled    bool      `json:"enabled"`
	Running    bool      `json:"running"`
	Port       int       `json:"port"`
	WatchPaths []string  `json:"watch_paths"`
	Reloads    int       `json:"reloads"`
	LastReload time.Time `json:"last_reload,omitempty"`
	Timestamp  time.Time `json:"timestamp"`
}

// LogAnalyzerStatus represents log analyzer status
type LogAnalyzerStatus struct {
	Enabled          bool      `json:"enabled"`
	Running          bool      `json:"running"`
	RealTime         bool      `json:"real_time"`
	ErrorDetection   bool      `json:"error_detection"`
	LogSources       []string  `json:"log_sources"`
	ErrorsDetected   int       `json:"errors_detected"`
	WarningsDetected int       `json:"warnings_detected"`
	Timestamp        time.Time `json:"timestamp"`
}

// MetricsCollectorStatus represents metrics collector status
type MetricsCollectorStatus struct {
	Enabled            bool      `json:"enabled"`
	Running            bool      `json:"running"`
	CustomMetrics      bool      `json:"custom_metrics"`
	PerformanceMetrics bool      `json:"performance_metrics"`
	BusinessMetrics    bool      `json:"business_metrics"`
	MetricsCollected   int       `json:"metrics_collected"`
	LastCollection     time.Time `json:"last_collection,omitempty"`
	Timestamp          time.Time `json:"timestamp"`
}

// Security and Privacy Models

// SecurityStatus represents the overall security status
type SecurityStatus struct {
	Enabled               bool                    `json:"enabled"`
	Running               bool                    `json:"running"`
	Authentication        *AuthenticationStatus   `json:"authentication"`
	Encryption            *EncryptionStatus       `json:"encryption"`
	Privacy               *PrivacyStatus          `json:"privacy"`
	ThreatDetection       *ThreatDetectionStatus  `json:"threat_detection"`
	Audit                 *AuditStatus            `json:"audit"`
	AccessControl         *AccessControlStatus    `json:"access_control"`
	Compliance            *ComplianceStatus       `json:"compliance"`
	IncidentResponse      *IncidentResponseStatus `json:"incident_response"`
	VulnerabilityScanning *VulnerabilityStatus    `json:"vulnerability_scanning"`
	Timestamp             time.Time               `json:"timestamp"`
}

// AuthenticationStatus represents authentication status
type AuthenticationStatus struct {
	Enabled        bool      `json:"enabled"`
	ActiveSessions int       `json:"active_sessions"`
	MFAEnabled     bool      `json:"mfa_enabled"`
	OAuthEnabled   bool      `json:"oauth_enabled"`
	LDAPEnabled    bool      `json:"ldap_enabled"`
	LastLogin      time.Time `json:"last_login,omitempty"`
	FailedAttempts int       `json:"failed_attempts"`
	Timestamp      time.Time `json:"timestamp"`
}

// EncryptionStatus represents encryption status
type EncryptionStatus struct {
	Enabled      bool      `json:"enabled"`
	Algorithm    string    `json:"algorithm"`
	KeySize      int       `json:"key_size"`
	AtRest       bool      `json:"at_rest"`
	InTransit    bool      `json:"in_transit"`
	HSMEnabled   bool      `json:"hsm_enabled"`
	LastRotation time.Time `json:"last_rotation,omitempty"`
	Timestamp    time.Time `json:"timestamp"`
}

// PrivacyStatus represents privacy status
type PrivacyStatus struct {
	Enabled           bool      `json:"enabled"`
	DataMinimization  bool      `json:"data_minimization"`
	Anonymization     bool      `json:"anonymization"`
	ConsentManagement bool      `json:"consent_management"`
	PIIDetected       int       `json:"pii_detected"`
	DataRetention     string    `json:"data_retention"`
	Timestamp         time.Time `json:"timestamp"`
}

// ThreatDetectionStatus represents threat detection status
type ThreatDetectionStatus struct {
	Enabled         bool      `json:"enabled"`
	RealTime        bool      `json:"real_time"`
	ThreatsDetected int       `json:"threats_detected"`
	HighSeverity    int       `json:"high_severity"`
	LastThreat      time.Time `json:"last_threat,omitempty"`
	Timestamp       time.Time `json:"timestamp"`
}

// AuditStatus represents audit status
type AuditStatus struct {
	Enabled       bool      `json:"enabled"`
	LogsGenerated int       `json:"logs_generated"`
	Encrypted     bool      `json:"encrypted"`
	RemoteLogging bool      `json:"remote_logging"`
	LastAudit     time.Time `json:"last_audit,omitempty"`
	Timestamp     time.Time `json:"timestamp"`
}

// AccessControlStatus represents access control status
type AccessControlStatus struct {
	Enabled     bool      `json:"enabled"`
	Model       string    `json:"model"`
	ActiveUsers int       `json:"active_users"`
	Roles       int       `json:"roles"`
	Permissions int       `json:"permissions"`
	LastAccess  time.Time `json:"last_access,omitempty"`
	Timestamp   time.Time `json:"timestamp"`
}

// ComplianceStatus represents compliance status
type ComplianceStatus struct {
	Enabled    bool      `json:"enabled"`
	Standards  []string  `json:"standards"`
	Compliant  bool      `json:"compliant"`
	LastCheck  time.Time `json:"last_check,omitempty"`
	Violations int       `json:"violations"`
	Timestamp  time.Time `json:"timestamp"`
}

// IncidentResponseStatus represents incident response status
type IncidentResponseStatus struct {
	Enabled         bool      `json:"enabled"`
	AutoResponse    bool      `json:"auto_response"`
	ActiveIncidents int       `json:"active_incidents"`
	ResolvedToday   int       `json:"resolved_today"`
	LastIncident    time.Time `json:"last_incident,omitempty"`
	Timestamp       time.Time `json:"timestamp"`
}

// VulnerabilityStatus represents vulnerability scanning status
type VulnerabilityStatus struct {
	Enabled              bool      `json:"enabled"`
	LastScan             time.Time `json:"last_scan,omitempty"`
	VulnerabilitiesFound int       `json:"vulnerabilities_found"`
	HighSeverity         int       `json:"high_severity"`
	AutoRemediation      bool      `json:"auto_remediation"`
	Timestamp            time.Time `json:"timestamp"`
}

// ThreatAnalysis represents a threat analysis result
type ThreatAnalysis struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Severity    string                 `json:"severity"`
	Source      string                 `json:"source"`
	Target      string                 `json:"target"`
	Description string                 `json:"description"`
	Indicators  []string               `json:"indicators"`
	Mitigation  string                 `json:"mitigation"`
	Confidence  float64                `json:"confidence"`
	Metadata    map[string]interface{} `json:"metadata"`
	DetectedAt  time.Time              `json:"detected_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// AuditLog represents an audit log entry
type AuditLog struct {
	ID        string                 `json:"id"`
	UserID    string                 `json:"user_id"`
	Action    string                 `json:"action"`
	Resource  string                 `json:"resource"`
	Result    string                 `json:"result"`
	IPAddress string                 `json:"ip_address"`
	UserAgent string                 `json:"user_agent"`
	Details   map[string]interface{} `json:"details"`
	Timestamp time.Time              `json:"timestamp"`
}

// AuditFilter represents audit log filter criteria
type AuditFilter struct {
	UserID    string    `json:"user_id,omitempty"`
	Action    string    `json:"action,omitempty"`
	Resource  string    `json:"resource,omitempty"`
	StartTime time.Time `json:"start_time,omitempty"`
	EndTime   time.Time `json:"end_time,omitempty"`
	Limit     int       `json:"limit,omitempty"`
}

// ComplianceReport represents a compliance validation report
type ComplianceReport struct {
	ID              string                 `json:"id"`
	Standard        string                 `json:"standard"`
	Status          string                 `json:"status"`
	Score           float64                `json:"score"`
	Violations      []ComplianceViolation  `json:"violations"`
	Recommendations []string               `json:"recommendations"`
	Metadata        map[string]interface{} `json:"metadata"`
	GeneratedAt     time.Time              `json:"generated_at"`
}

// ComplianceViolation represents a compliance violation
type ComplianceViolation struct {
	Rule        string `json:"rule"`
	Severity    string `json:"severity"`
	Description string `json:"description"`
	Resource    string `json:"resource"`
	Remediation string `json:"remediation"`
}

// Testing and Validation Models

// TestingStatus represents the overall testing status
type TestingStatus struct {
	Enabled     bool                      `json:"enabled"`
	Running     bool                      `json:"running"`
	Unit        *UnitTestingStatus        `json:"unit"`
	Integration *IntegrationTestingStatus `json:"integration"`
	E2E         *E2ETestingStatus         `json:"e2e"`
	Performance *PerformanceTestingStatus `json:"performance"`
	Security    *SecurityTestingStatus    `json:"security"`
	Validation  *ValidationStatus         `json:"validation"`
	Coverage    *CoverageStatus           `json:"coverage"`
	Timestamp   time.Time                 `json:"timestamp"`
}

// UnitTestingStatus represents unit testing status
type UnitTestingStatus struct {
	Enabled      bool          `json:"enabled"`
	TestsRun     int           `json:"tests_run"`
	TestsPassed  int           `json:"tests_passed"`
	TestsFailed  int           `json:"tests_failed"`
	TestsSkipped int           `json:"tests_skipped"`
	Coverage     float64       `json:"coverage"`
	Duration     time.Duration `json:"duration"`
	LastRun      time.Time     `json:"last_run,omitempty"`
	Timestamp    time.Time     `json:"timestamp"`
}

// IntegrationTestingStatus represents integration testing status
type IntegrationTestingStatus struct {
	Enabled     bool          `json:"enabled"`
	TestsRun    int           `json:"tests_run"`
	TestsPassed int           `json:"tests_passed"`
	TestsFailed int           `json:"tests_failed"`
	Duration    time.Duration `json:"duration"`
	LastRun     time.Time     `json:"last_run,omitempty"`
	Timestamp   time.Time     `json:"timestamp"`
}

// E2ETestingStatus represents end-to-end testing status
type E2ETestingStatus struct {
	Enabled     bool          `json:"enabled"`
	TestsRun    int           `json:"tests_run"`
	TestsPassed int           `json:"tests_passed"`
	TestsFailed int           `json:"tests_failed"`
	Duration    time.Duration `json:"duration"`
	LastRun     time.Time     `json:"last_run,omitempty"`
	Timestamp   time.Time     `json:"timestamp"`
}

// PerformanceTestingStatus represents performance testing status
type PerformanceTestingStatus struct {
	Enabled       bool          `json:"enabled"`
	TestsRun      int           `json:"tests_run"`
	BenchmarksRun int           `json:"benchmarks_run"`
	LoadTests     int           `json:"load_tests"`
	StressTests   int           `json:"stress_tests"`
	Duration      time.Duration `json:"duration"`
	LastRun       time.Time     `json:"last_run,omitempty"`
	Timestamp     time.Time     `json:"timestamp"`
}

// SecurityTestingStatus represents security testing status
type SecurityTestingStatus struct {
	Enabled              bool          `json:"enabled"`
	VulnerabilitiesFound int           `json:"vulnerabilities_found"`
	SecurityTestsRun     int           `json:"security_tests_run"`
	PenetrationTests     int           `json:"penetration_tests"`
	Duration             time.Duration `json:"duration"`
	LastRun              time.Time     `json:"last_run,omitempty"`
	Timestamp            time.Time     `json:"timestamp"`
}

// ValidationStatus represents validation status
type ValidationStatus struct {
	Enabled          bool      `json:"enabled"`
	SchemasValidated int       `json:"schemas_validated"`
	DataValidated    int       `json:"data_validated"`
	APIValidated     int       `json:"api_validated"`
	ValidationErrors int       `json:"validation_errors"`
	LastValidation   time.Time `json:"last_validation,omitempty"`
	Timestamp        time.Time `json:"timestamp"`
}

// CoverageStatus represents test coverage status
type CoverageStatus struct {
	Enabled          bool      `json:"enabled"`
	OverallCoverage  float64   `json:"overall_coverage"`
	LineCoverage     float64   `json:"line_coverage"`
	BranchCoverage   float64   `json:"branch_coverage"`
	FunctionCoverage float64   `json:"function_coverage"`
	LastAnalysis     time.Time `json:"last_analysis,omitempty"`
	Timestamp        time.Time `json:"timestamp"`
}

// TestSuiteResult represents a complete test suite execution result
type TestSuiteResult struct {
	ID        string                 `json:"id"`
	StartTime time.Time              `json:"start_time"`
	EndTime   time.Time              `json:"end_time"`
	Duration  time.Duration          `json:"duration"`
	Status    string                 `json:"status"`
	Results   map[string]*TestResult `json:"results"`
	Coverage  *CoverageReport        `json:"coverage,omitempty"`
	Metadata  map[string]interface{} `json:"metadata"`
}

// TestResult represents individual test execution result
type TestResult struct {
	ID           string                 `json:"id"`
	Type         string                 `json:"type"`
	Name         string                 `json:"name"`
	StartTime    time.Time              `json:"start_time"`
	EndTime      time.Time              `json:"end_time"`
	Duration     time.Duration          `json:"duration"`
	Status       string                 `json:"status"`
	TestsRun     int                    `json:"tests_run"`
	TestsPassed  int                    `json:"tests_passed"`
	TestsFailed  int                    `json:"tests_failed"`
	TestsSkipped int                    `json:"tests_skipped"`
	Failures     []TestFailure          `json:"failures"`
	Output       string                 `json:"output"`
	Coverage     float64                `json:"coverage"`
	Metadata     map[string]interface{} `json:"metadata"`
}

// TestFailure represents a test failure
type TestFailure struct {
	TestName   string `json:"test_name"`
	Message    string `json:"message"`
	StackTrace string `json:"stack_trace"`
	File       string `json:"file"`
	Line       int    `json:"line"`
	Expected   string `json:"expected,omitempty"`
	Actual     string `json:"actual,omitempty"`
}

// ValidationResult represents validation result
type ValidationResult struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	Target    string                 `json:"target"`
	Valid     bool                   `json:"valid"`
	Errors    []ValidationError      `json:"errors"`
	Warnings  []ValidationWarning    `json:"warnings"`
	Schema    string                 `json:"schema,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
	Metadata  map[string]interface{} `json:"metadata"`
}

// ValidationError represents a validation error
type ValidationError struct {
	Field   string      `json:"field"`
	Message string      `json:"message"`
	Code    string      `json:"code"`
	Value   interface{} `json:"value,omitempty"`
}

// ValidationWarning represents a validation warning
type ValidationWarning struct {
	Field   string      `json:"field"`
	Message string      `json:"message"`
	Code    string      `json:"code"`
	Value   interface{} `json:"value,omitempty"`
}

// CoverageReport represents test coverage report
type CoverageReport struct {
	ID               string                 `json:"id"`
	GeneratedAt      time.Time              `json:"generated_at"`
	OverallCoverage  float64                `json:"overall_coverage"`
	LineCoverage     float64                `json:"line_coverage"`
	BranchCoverage   float64                `json:"branch_coverage"`
	FunctionCoverage float64                `json:"function_coverage"`
	Files            []FileCoverage         `json:"files"`
	Packages         []PackageCoverage      `json:"packages"`
	Thresholds       map[string]float64     `json:"thresholds"`
	Metadata         map[string]interface{} `json:"metadata"`
}

// FileCoverage represents coverage for a single file
type FileCoverage struct {
	Path             string  `json:"path"`
	LineCoverage     float64 `json:"line_coverage"`
	BranchCoverage   float64 `json:"branch_coverage"`
	FunctionCoverage float64 `json:"function_coverage"`
	LinesTotal       int     `json:"lines_total"`
	LinesCovered     int     `json:"lines_covered"`
	BranchesTotal    int     `json:"branches_total"`
	BranchesCovered  int     `json:"branches_covered"`
	FunctionsTotal   int     `json:"functions_total"`
	FunctionsCovered int     `json:"functions_covered"`
}

// PackageCoverage represents coverage for a package
type PackageCoverage struct {
	Name             string         `json:"name"`
	LineCoverage     float64        `json:"line_coverage"`
	BranchCoverage   float64        `json:"branch_coverage"`
	FunctionCoverage float64        `json:"function_coverage"`
	Files            []FileCoverage `json:"files"`
}

// TestFilter represents test filtering criteria
type TestFilter struct {
	Type      string    `json:"type,omitempty"`
	Status    string    `json:"status,omitempty"`
	StartTime time.Time `json:"start_time,omitempty"`
	EndTime   time.Time `json:"end_time,omitempty"`
	Limit     int       `json:"limit,omitempty"`
	Offset    int       `json:"offset,omitempty"`
}

// Deployment and Documentation Models

// DeploymentStatus represents the overall deployment status
type DeploymentStatus struct {
	Enabled     bool               `json:"enabled"`
	Running     bool               `json:"running"`
	Environment string             `json:"environment"`
	Platform    string             `json:"platform"`
	Container   *ContainerStatus   `json:"container"`
	Kubernetes  *KubernetesStatus  `json:"kubernetes"`
	CICD        *CICDStatus        `json:"cicd"`
	Health      *HealthCheckStatus `json:"health"`
	Timestamp   time.Time          `json:"timestamp"`
}

// ContainerStatus represents container deployment status
type ContainerStatus struct {
	Enabled    bool      `json:"enabled"`
	Running    bool      `json:"running"`
	Containers int       `json:"containers"`
	Images     int       `json:"images"`
	Networks   int       `json:"networks"`
	Volumes    int       `json:"volumes"`
	LastDeploy time.Time `json:"last_deploy,omitempty"`
	Timestamp  time.Time `json:"timestamp"`
}

// KubernetesStatus represents Kubernetes deployment status
type KubernetesStatus struct {
	Enabled     bool      `json:"enabled"`
	Connected   bool      `json:"connected"`
	Namespace   string    `json:"namespace"`
	Pods        int       `json:"pods"`
	Services    int       `json:"services"`
	Deployments int       `json:"deployments"`
	LastDeploy  time.Time `json:"last_deploy,omitempty"`
	Timestamp   time.Time `json:"timestamp"`
}

// CICDStatus represents CI/CD pipeline status
type CICDStatus struct {
	Enabled      bool      `json:"enabled"`
	Provider     string    `json:"provider"`
	PipelinesRun int       `json:"pipelines_run"`
	LastPipeline time.Time `json:"last_pipeline,omitempty"`
	LastStatus   string    `json:"last_status"`
	Timestamp    time.Time `json:"timestamp"`
}

// HealthCheckStatus represents health check status
type HealthCheckStatus struct {
	Enabled      bool      `json:"enabled"`
	Healthy      bool      `json:"healthy"`
	ChecksRun    int       `json:"checks_run"`
	ChecksPassed int       `json:"checks_passed"`
	ChecksFailed int       `json:"checks_failed"`
	LastCheck    time.Time `json:"last_check,omitempty"`
	Timestamp    time.Time `json:"timestamp"`
}

// DeploymentRequest represents a deployment request
type DeploymentRequest struct {
	Version     string                 `json:"version"`
	Environment string                 `json:"environment"`
	Platform    string                 `json:"platform"`
	Config      map[string]interface{} `json:"config"`
	Variables   map[string]string      `json:"variables"`
	Secrets     map[string]string      `json:"secrets"`
	Rollback    bool                   `json:"rollback"`
	DryRun      bool                   `json:"dry_run"`
	Force       bool                   `json:"force"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// DeploymentResult represents a deployment result
type DeploymentResult struct {
	ID          string                 `json:"id"`
	Version     string                 `json:"version"`
	Environment string                 `json:"environment"`
	Platform    string                 `json:"platform"`
	StartTime   time.Time              `json:"start_time"`
	EndTime     time.Time              `json:"end_time"`
	Duration    time.Duration          `json:"duration"`
	Status      string                 `json:"status"`
	Error       string                 `json:"error,omitempty"`
	Steps       []DeploymentStep       `json:"steps"`
	Artifacts   []DeploymentArtifact   `json:"artifacts"`
	Logs        []string               `json:"logs"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// DeploymentStep represents a deployment step
type DeploymentStep struct {
	Name      string                 `json:"name"`
	Status    string                 `json:"status"`
	StartTime time.Time              `json:"start_time"`
	EndTime   time.Time              `json:"end_time"`
	Duration  time.Duration          `json:"duration"`
	Error     string                 `json:"error,omitempty"`
	Output    string                 `json:"output,omitempty"`
	Metadata  map[string]interface{} `json:"metadata"`
}

// DeploymentArtifact represents a deployment artifact
type DeploymentArtifact struct {
	Name     string    `json:"name"`
	Type     string    `json:"type"`
	Path     string    `json:"path"`
	Size     int64     `json:"size"`
	Checksum string    `json:"checksum"`
	Created  time.Time `json:"created"`
}

// RollbackRequest represents a rollback request
type RollbackRequest struct {
	TargetVersion string                 `json:"target_version"`
	Environment   string                 `json:"environment"`
	Platform      string                 `json:"platform"`
	Reason        string                 `json:"reason"`
	Force         bool                   `json:"force"`
	DryRun        bool                   `json:"dry_run"`
	Metadata      map[string]interface{} `json:"metadata"`
}

// DeploymentFilter represents deployment filtering criteria
type DeploymentFilter struct {
	Environment string    `json:"environment,omitempty"`
	Platform    string    `json:"platform,omitempty"`
	Status      string    `json:"status,omitempty"`
	Version     string    `json:"version,omitempty"`
	StartTime   time.Time `json:"start_time,omitempty"`
	EndTime     time.Time `json:"end_time,omitempty"`
	Limit       int       `json:"limit,omitempty"`
	Offset      int       `json:"offset,omitempty"`
}

// DocumentationStatus represents documentation status
type DocumentationStatus struct {
	Enabled       bool      `json:"enabled"`
	Generated     bool      `json:"generated"`
	LastGenerated time.Time `json:"last_generated,omitempty"`
	Pages         int       `json:"pages"`
	APIs          int       `json:"apis"`
	Examples      int       `json:"examples"`
	Timestamp     time.Time `json:"timestamp"`
}

// DocumentationRequest represents a documentation generation request
type DocumentationRequest struct {
	Type      string                 `json:"type"`
	Format    string                 `json:"format"`
	Output    string                 `json:"output"`
	Include   []string               `json:"include"`
	Exclude   []string               `json:"exclude"`
	Template  string                 `json:"template"`
	Variables map[string]string      `json:"variables"`
	Metadata  map[string]interface{} `json:"metadata"`
}

// DocumentationResult represents documentation generation result
type DocumentationResult struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	Format    string                 `json:"format"`
	Output    string                 `json:"output"`
	StartTime time.Time              `json:"start_time"`
	EndTime   time.Time              `json:"end_time"`
	Duration  time.Duration          `json:"duration"`
	Status    string                 `json:"status"`
	Error     string                 `json:"error,omitempty"`
	Files     []DocumentationFile    `json:"files"`
	Metadata  map[string]interface{} `json:"metadata"`
}

// DocumentationFile represents a generated documentation file
type DocumentationFile struct {
	Path     string    `json:"path"`
	Type     string    `json:"type"`
	Size     int64     `json:"size"`
	Checksum string    `json:"checksum"`
	Created  time.Time `json:"created"`
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

// ChatMessage represents a single message in a conversation
type ChatMessage struct {
	Role      string                 `json:"role"` // "system", "user", "assistant"
	Content   string                 `json:"content"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
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
	Provider     string                 `json:"provider"` // ollama, openai, etc.
	IsActive     bool                   `json:"is_active"`
	IsDefault    bool                   `json:"is_default"`
	SizeBytes    int64                  `json:"size_bytes"`
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
