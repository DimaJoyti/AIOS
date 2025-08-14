package models

import (
	"time"
)

// SystemStatus represents the overall system status
type SystemStatus struct {
	Running      bool                    `json:"running"`
	Version      string                  `json:"version"`
	Uptime       time.Duration           `json:"uptime"`
	Resources    *ResourceStatus         `json:"resources"`
	Security     *SecurityStatus         `json:"security"`
	Optimization *OptimizationStatus     `json:"optimization"`
	Services     map[string]ServiceStatus `json:"services"`
	Timestamp    time.Time               `json:"timestamp"`
}

// ResourceStatus represents system resource utilization
type ResourceStatus struct {
	CPU    *CPUStatus    `json:"cpu"`
	Memory *MemoryStatus `json:"memory"`
	Disk   *DiskStatus   `json:"disk"`
	Network *NetworkStatus `json:"network"`
}

// CPUStatus represents CPU utilization information
type CPUStatus struct {
	Usage       float64   `json:"usage"`        // Percentage
	Cores       int       `json:"cores"`
	Temperature float64   `json:"temperature"`  // Celsius
	Frequency   float64   `json:"frequency"`    // MHz
	LoadAvg     []float64 `json:"load_avg"`     // 1, 5, 15 minute averages
}

// MemoryStatus represents memory utilization information
type MemoryStatus struct {
	Total     uint64  `json:"total"`      // Bytes
	Used      uint64  `json:"used"`       // Bytes
	Available uint64  `json:"available"`  // Bytes
	Usage     float64 `json:"usage"`      // Percentage
	Swap      *SwapStatus `json:"swap"`
}

// SwapStatus represents swap memory information
type SwapStatus struct {
	Total uint64  `json:"total"`  // Bytes
	Used  uint64  `json:"used"`   // Bytes
	Usage float64 `json:"usage"`  // Percentage
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
	Total      uint64  `json:"total"`      // Bytes
	Used       uint64  `json:"used"`       // Bytes
	Available  uint64  `json:"available"`  // Bytes
	Usage      float64 `json:"usage"`      // Percentage
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
	Interfaces []NetworkInterface `json:"interfaces"`
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
	ThreatLevel    string           `json:"threat_level"`    // low, medium, high, critical
	ActiveThreats  int              `json:"active_threats"`
	BlockedAttacks int              `json:"blocked_attacks"`
	LastScan       time.Time        `json:"last_scan"`
	Firewall       *FirewallStatus  `json:"firewall"`
	Antivirus      *AntivirusStatus `json:"antivirus"`
}

// FirewallStatus represents firewall status
type FirewallStatus struct {
	Enabled       bool   `json:"enabled"`
	Rules         int    `json:"rules"`
	BlockedIPs    int    `json:"blocked_ips"`
	AllowedPorts  []int  `json:"allowed_ports"`
	BlockedPorts  []int  `json:"blocked_ports"`
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
	Enabled           bool                    `json:"enabled"`
	LastOptimization  time.Time               `json:"last_optimization"`
	OptimizationsRun  int                     `json:"optimizations_run"`
	PerformanceGain   float64                 `json:"performance_gain"`  // Percentage
	Recommendations   []OptimizationRecommendation `json:"recommendations"`
}

// OptimizationRecommendation represents an AI optimization recommendation
type OptimizationRecommendation struct {
	ID          string    `json:"id"`
	Type        string    `json:"type"`        // cpu, memory, disk, network
	Priority    string    `json:"priority"`    // low, medium, high
	Description string    `json:"description"`
	Impact      string    `json:"impact"`      // Expected impact description
	Applied     bool      `json:"applied"`
	CreatedAt   time.Time `json:"created_at"`
}

// ServiceStatus represents individual service status
type ServiceStatus struct {
	Name      string    `json:"name"`
	Status    string    `json:"status"`    // running, stopped, error
	Health    string    `json:"health"`    // healthy, unhealthy, unknown
	Uptime    time.Duration `json:"uptime"`
	CPU       float64   `json:"cpu"`       // Percentage
	Memory    uint64    `json:"memory"`    // Bytes
	LastCheck time.Time `json:"last_check"`
}

// AIModelStatus represents AI model status
type AIModelStatus struct {
	Name         string    `json:"name"`
	Version      string    `json:"version"`
	Status       string    `json:"status"`       // loaded, loading, error
	Type         string    `json:"type"`         // llm, cv, optimization
	Size         uint64    `json:"size"`         // Bytes
	LoadTime     time.Duration `json:"load_time"`
	LastUsed     time.Time `json:"last_used"`
	RequestCount int64     `json:"request_count"`
	AvgLatency   time.Duration `json:"avg_latency"`
}

// FileSystemAnalysis represents file system AI analysis results
type FileSystemAnalysis struct {
	Path              string                 `json:"path"`
	TotalFiles        int                    `json:"total_files"`
	TotalSize         uint64                 `json:"total_size"`
	FileTypes         map[string]int         `json:"file_types"`
	LargestFiles      []FileInfo             `json:"largest_files"`
	DuplicateFiles    []DuplicateGroup       `json:"duplicate_files"`
	UnusedFiles       []FileInfo             `json:"unused_files"`
	Recommendations   []string               `json:"recommendations"`
	AnalyzedAt        time.Time              `json:"analyzed_at"`
}

// FileInfo represents file information
type FileInfo struct {
	Path         string    `json:"path"`
	Size         uint64    `json:"size"`
	ModTime      time.Time `json:"mod_time"`
	AccessTime   time.Time `json:"access_time"`
	Type         string    `json:"type"`
	Permissions  string    `json:"permissions"`
}

// DuplicateGroup represents a group of duplicate files
type DuplicateGroup struct {
	Hash  string     `json:"hash"`
	Size  uint64     `json:"size"`
	Files []FileInfo `json:"files"`
}

// ThreatAnalysis represents security threat analysis
type ThreatAnalysis struct {
	Threats       []ThreatInfo `json:"threats"`
	RiskScore     float64      `json:"risk_score"`     // 0-100
	Severity      string       `json:"severity"`       // low, medium, high, critical
	AnalyzedAt    time.Time    `json:"analyzed_at"`
	Recommendations []string   `json:"recommendations"`
}

// ThreatInfo represents individual threat information
type ThreatInfo struct {
	ID          string    `json:"id"`
	Type        string    `json:"type"`        // malware, intrusion, anomaly
	Severity    string    `json:"severity"`    // low, medium, high, critical
	Source      string    `json:"source"`      // IP, process, file
	Description string    `json:"description"`
	FirstSeen   time.Time `json:"first_seen"`
	LastSeen    time.Time `json:"last_seen"`
	Count       int       `json:"count"`
	Blocked     bool      `json:"blocked"`
}

// PerformanceMetrics represents system performance metrics
type PerformanceMetrics struct {
	Timestamp    time.Time `json:"timestamp"`
	CPUUsage     float64   `json:"cpu_usage"`
	MemoryUsage  float64   `json:"memory_usage"`
	DiskUsage    float64   `json:"disk_usage"`
	NetworkIn    uint64    `json:"network_in"`
	NetworkOut   uint64    `json:"network_out"`
	ResponseTime time.Duration `json:"response_time"`
	Throughput   float64   `json:"throughput"`    // Requests per second
}

// Configuration represents system configuration
type Configuration struct {
	Server       ServerConfig       `json:"server"`
	Database     DatabaseConfig     `json:"database"`
	Redis        RedisConfig        `json:"redis"`
	AI           AIConfig           `json:"ai"`
	Security     SecurityConfig     `json:"security"`
	Logging      LoggingConfig      `json:"logging"`
	Tracing      TracingConfig      `json:"tracing"`
	FileSystem   FileSystemConfig   `json:"filesystem"`
	Desktop      DesktopConfig      `json:"desktop"`
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
	Ollama      OllamaConfig `json:"ollama"`
	ModelsPath  string       `json:"models_path"`
	DefaultModel string      `json:"default_model"`
	MaxTokens   int          `json:"max_tokens"`
	Temperature float64      `json:"temperature"`
}

// OllamaConfig represents Ollama configuration
type OllamaConfig struct {
	Host    string        `json:"host"`
	Port    int           `json:"port"`
	Timeout time.Duration `json:"timeout"`
}

// SecurityConfig represents security configuration
type SecurityConfig struct {
	JWTSecret      string        `json:"jwt_secret"`
	SessionTimeout time.Duration `json:"session_timeout"`
	RateLimit      RateLimitConfig `json:"rate_limit"`
	CORS           CORSConfig    `json:"cors"`
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
	Enabled         bool    `json:"enabled"`
	JaegerEndpoint  string  `json:"jaeger_endpoint"`
	ServiceName     string  `json:"service_name"`
	SampleRate      float64 `json:"sample_rate"`
}

// FileSystemConfig represents file system configuration
type FileSystemConfig struct {
	WatchPaths         []string      `json:"watch_paths"`
	IgnorePatterns     []string      `json:"ignore_patterns"`
	AnalysisInterval   time.Duration `json:"analysis_interval"`
	OrganizationEnabled bool         `json:"organization_enabled"`
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
