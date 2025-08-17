package config

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

// Manager handles configuration loading and management
type Manager struct {
	viper       *viper.Viper
	environment string
	configPath  string
}

// Config represents the complete application configuration
type Config struct {
	Environment string            `mapstructure:"environment"`
	Version     string            `mapstructure:"version"`
	Server      ServerConfig      `mapstructure:"server"`
	Metrics     MetricsConfig     `mapstructure:"metrics"`
	Logging     LoggingConfig     `mapstructure:"logging"`
	Database    DatabaseConfig    `mapstructure:"database"`
	Redis       RedisConfig       `mapstructure:"redis"`
	AI          AIConfig          `mapstructure:"ai"`
	Security    SecurityConfig    `mapstructure:"security"`
	Tracing     TracingConfig     `mapstructure:"tracing"`
	Features    FeaturesConfig    `mapstructure:"features"`
	Storage     StorageConfig     `mapstructure:"storage"`
	Performance PerformanceConfig `mapstructure:"performance"`
	Health      HealthConfig      `mapstructure:"health"`
	Backup      BackupConfig      `mapstructure:"backup"`
	Monitoring  MonitoringConfig  `mapstructure:"monitoring"`
}

// ServerConfig contains HTTP server configuration
type ServerConfig struct {
	Host            string        `mapstructure:"host"`
	Port            int           `mapstructure:"port"`
	ReadTimeout     time.Duration `mapstructure:"read_timeout"`
	WriteTimeout    time.Duration `mapstructure:"write_timeout"`
	IdleTimeout     time.Duration `mapstructure:"idle_timeout"`
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout"`
	EnablePprof     bool          `mapstructure:"enable_pprof"`
	PprofPort       int           `mapstructure:"pprof_port"`
}

// MetricsConfig contains metrics configuration
type MetricsConfig struct {
	Enabled         bool          `mapstructure:"enabled"`
	Host            string        `mapstructure:"host"`
	Port            int           `mapstructure:"port"`
	Path            string        `mapstructure:"path"`
	CollectInterval time.Duration `mapstructure:"collect_interval"`
}

// LoggingConfig contains logging configuration
type LoggingConfig struct {
	Level            string         `mapstructure:"level"`
	Format           string         `mapstructure:"format"`
	Output           string         `mapstructure:"output"`
	EnableCaller     bool           `mapstructure:"enable_caller"`
	EnableStacktrace bool           `mapstructure:"enable_stacktrace"`
	Development      bool           `mapstructure:"development"`
	Sampling         SamplingConfig `mapstructure:"sampling"`
}

// SamplingConfig contains log sampling configuration
type SamplingConfig struct {
	Initial    int `mapstructure:"initial"`
	Thereafter int `mapstructure:"thereafter"`
}

// DatabaseConfig contains database configuration
type DatabaseConfig struct {
	Driver             string        `mapstructure:"driver"`
	Host               string        `mapstructure:"host"`
	Port               int           `mapstructure:"port"`
	Name               string        `mapstructure:"name"`
	User               string        `mapstructure:"user"`
	Password           string        `mapstructure:"password"`
	SSLMode            string        `mapstructure:"ssl_mode"`
	MaxOpenConns       int           `mapstructure:"max_open_conns"`
	MaxIdleConns       int           `mapstructure:"max_idle_conns"`
	ConnMaxLifetime    time.Duration `mapstructure:"conn_max_lifetime"`
	ConnMaxIdleTime    time.Duration `mapstructure:"conn_max_idle_time"`
	LogQueries         bool          `mapstructure:"log_queries"`
	SlowQueryThreshold time.Duration `mapstructure:"slow_query_threshold"`
}

// RedisConfig contains Redis configuration
type RedisConfig struct {
	Host         string        `mapstructure:"host"`
	Port         int           `mapstructure:"port"`
	Password     string        `mapstructure:"password"`
	DB           int           `mapstructure:"db"`
	PoolSize     int           `mapstructure:"pool_size"`
	MinIdleConns int           `mapstructure:"min_idle_conns"`
	DialTimeout  time.Duration `mapstructure:"dial_timeout"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
	PoolTimeout  time.Duration `mapstructure:"pool_timeout"`
	IdleTimeout  time.Duration `mapstructure:"idle_timeout"`
	MaxRetries   int           `mapstructure:"max_retries"`
}

// AIConfig contains AI services configuration
type AIConfig struct {
	Enabled      bool               `mapstructure:"enabled"`
	Debug        bool               `mapstructure:"debug"`
	LLM          LLMConfig          `mapstructure:"llm"`
	CV           CVConfig           `mapstructure:"cv"`
	Optimization OptimizationConfig `mapstructure:"optimization"`
}

// LLMConfig contains language model configuration
type LLMConfig struct {
	Provider      string        `mapstructure:"provider"`
	Endpoint      string        `mapstructure:"endpoint"`
	Model         string        `mapstructure:"model"`
	MaxTokens     int           `mapstructure:"max_tokens"`
	Temperature   float64       `mapstructure:"temperature"`
	Timeout       time.Duration `mapstructure:"timeout"`
	RetryAttempts int           `mapstructure:"retry_attempts"`
}

// CVConfig contains computer vision configuration
type CVConfig struct {
	Enabled             bool    `mapstructure:"enabled"`
	Provider            string  `mapstructure:"provider"`
	ModelsPath          string  `mapstructure:"models_path"`
	ConfidenceThreshold float64 `mapstructure:"confidence_threshold"`
}

// OptimizationConfig contains optimization service configuration
type OptimizationConfig struct {
	Enabled            bool          `mapstructure:"enabled"`
	CollectionInterval time.Duration `mapstructure:"collection_interval"`
	AnalysisInterval   time.Duration `mapstructure:"analysis_interval"`
	AutoApply          bool          `mapstructure:"auto_apply"`
}

// SecurityConfig contains security configuration
type SecurityConfig struct {
	JWT          JWTConfig          `mapstructure:"jwt"`
	Encryption   EncryptionConfig   `mapstructure:"encryption"`
	CORS         CORSConfig         `mapstructure:"cors"`
	RateLimiting RateLimitingConfig `mapstructure:"rate_limiting"`
	Auth         AuthConfig         `mapstructure:"auth"`
}

// JWTConfig contains JWT configuration
type JWTConfig struct {
	Secret          string        `mapstructure:"secret"`
	Issuer          string        `mapstructure:"issuer"`
	Audience        string        `mapstructure:"audience"`
	AccessTokenTTL  time.Duration `mapstructure:"access_token_ttl"`
	RefreshTokenTTL time.Duration `mapstructure:"refresh_token_ttl"`
}

// EncryptionConfig contains encryption configuration
type EncryptionConfig struct {
	Key       string `mapstructure:"key"`
	Algorithm string `mapstructure:"algorithm"`
}

// CORSConfig contains CORS configuration
type CORSConfig struct {
	Enabled          bool     `mapstructure:"enabled"`
	AllowedOrigins   []string `mapstructure:"allowed_origins"`
	AllowedMethods   []string `mapstructure:"allowed_methods"`
	AllowedHeaders   []string `mapstructure:"allowed_headers"`
	AllowCredentials bool     `mapstructure:"allow_credentials"`
	MaxAge           int      `mapstructure:"max_age"`
}

// RateLimitingConfig contains rate limiting configuration
type RateLimitingConfig struct {
	Enabled           bool `mapstructure:"enabled"`
	RequestsPerMinute int  `mapstructure:"requests_per_minute"`
	Burst             int  `mapstructure:"burst"`
}

// AuthConfig contains authentication configuration
type AuthConfig struct {
	SessionTimeout   time.Duration `mapstructure:"session_timeout"`
	MaxLoginAttempts int           `mapstructure:"max_login_attempts"`
	LockoutDuration  time.Duration `mapstructure:"lockout_duration"`
}

// TracingConfig contains tracing configuration
type TracingConfig struct {
	Enabled        bool         `mapstructure:"enabled"`
	ServiceName    string       `mapstructure:"service_name"`
	ServiceVersion string       `mapstructure:"service_version"`
	Environment    string       `mapstructure:"environment"`
	Jaeger         JaegerConfig `mapstructure:"jaeger"`
	OTLP           OTLPConfig   `mapstructure:"otlp"`
}

// JaegerConfig contains Jaeger configuration
type JaegerConfig struct {
	Endpoint     string  `mapstructure:"endpoint"`
	SamplerType  string  `mapstructure:"sampler_type"`
	SamplerParam float64 `mapstructure:"sampler_param"`
}

// OTLPConfig contains OTLP configuration
type OTLPConfig struct {
	Endpoint string `mapstructure:"endpoint"`
	Insecure bool   `mapstructure:"insecure"`
}

// FeaturesConfig contains feature flags
type FeaturesConfig struct {
	VoiceControl         bool `mapstructure:"voice_control"`
	ComputerVision       bool `mapstructure:"computer_vision"`
	PredictiveFS         bool `mapstructure:"predictive_fs"`
	DeveloperTools       bool `mapstructure:"developer_tools"`
	AIOptimization       bool `mapstructure:"ai_optimization"`
	AdvancedSecurity     bool `mapstructure:"advanced_security"`
	ExperimentalFeatures bool `mapstructure:"experimental_features"`
}

// StorageConfig contains storage configuration
type StorageConfig struct {
	Cloud  CloudStorageConfig `mapstructure:"cloud"`
	Models ModelStorageConfig `mapstructure:"models"`
}

// CloudStorageConfig contains cloud storage configuration
type CloudStorageConfig struct {
	Provider    string `mapstructure:"provider"`
	Bucket      string `mapstructure:"bucket"`
	Region      string `mapstructure:"region"`
	AccessKey   string `mapstructure:"access_key"`
	SecretKey   string `mapstructure:"secret_key"`
	MaxFileSize string `mapstructure:"max_file_size"`
	Encryption  bool   `mapstructure:"encryption"`
}

// ModelStorageConfig contains model storage configuration
type ModelStorageConfig struct {
	Provider     string `mapstructure:"provider"`
	Bucket       string `mapstructure:"bucket"`
	CacheSize    string `mapstructure:"cache_size"`
	AutoDownload bool   `mapstructure:"auto_download"`
	Encryption   bool   `mapstructure:"encryption"`
}

// PerformanceConfig contains performance configuration
type PerformanceConfig struct {
	GC          GCConfig          `mapstructure:"gc"`
	Concurrency ConcurrencyConfig `mapstructure:"concurrency"`
	Cache       CacheConfig       `mapstructure:"cache"`
}

// GCConfig contains garbage collection configuration
type GCConfig struct {
	TargetPercentage int    `mapstructure:"target_percentage"`
	MaxHeapSize      string `mapstructure:"max_heap_size"`
}

// ConcurrencyConfig contains concurrency configuration
type ConcurrencyConfig struct {
	MaxGoroutines  int `mapstructure:"max_goroutines"`
	WorkerPoolSize int `mapstructure:"worker_pool_size"`
}

// CacheConfig contains cache configuration
type CacheConfig struct {
	DefaultTTL      time.Duration `mapstructure:"default_ttl"`
	MaxSize         string        `mapstructure:"max_size"`
	CleanupInterval time.Duration `mapstructure:"cleanup_interval"`
}

// HealthConfig contains health check configuration
type HealthConfig struct {
	Enabled          bool               `mapstructure:"enabled"`
	Endpoint         string             `mapstructure:"endpoint"`
	DetailedEndpoint string             `mapstructure:"detailed_endpoint"`
	CheckInterval    time.Duration      `mapstructure:"check_interval"`
	Checks           HealthChecksConfig `mapstructure:"checks"`
}

// HealthChecksConfig contains individual health check configuration
type HealthChecksConfig struct {
	Database     bool `mapstructure:"database"`
	Redis        bool `mapstructure:"redis"`
	AIServices   bool `mapstructure:"ai_services"`
	ExternalAPIs bool `mapstructure:"external_apis"`
}

// BackupConfig contains backup configuration
type BackupConfig struct {
	Enabled       bool                `mapstructure:"enabled"`
	Schedule      string              `mapstructure:"schedule"`
	RetentionDays int                 `mapstructure:"retention_days"`
	Storage       BackupStorageConfig `mapstructure:"storage"`
}

// BackupStorageConfig contains backup storage configuration
type BackupStorageConfig struct {
	Provider               string `mapstructure:"provider"`
	Bucket                 string `mapstructure:"bucket"`
	Encryption             bool   `mapstructure:"encryption"`
	CrossRegionReplication bool   `mapstructure:"cross_region_replication"`
}

// MonitoringConfig contains monitoring configuration
type MonitoringConfig struct {
	Prometheus PrometheusConfig `mapstructure:"prometheus"`
	Grafana    GrafanaConfig    `mapstructure:"grafana"`
	Alerting   AlertingConfig   `mapstructure:"alerting"`
}

// PrometheusConfig contains Prometheus configuration
type PrometheusConfig struct {
	Enabled        bool   `mapstructure:"enabled"`
	ScrapeInterval string `mapstructure:"scrape_interval"`
	Retention      string `mapstructure:"retention"`
}

// GrafanaConfig contains Grafana configuration
type GrafanaConfig struct {
	Enabled       bool   `mapstructure:"enabled"`
	AdminPassword string `mapstructure:"admin_password"`
}

// AlertingConfig contains alerting configuration
type AlertingConfig struct {
	Enabled      bool          `mapstructure:"enabled"`
	WebhookURL   string        `mapstructure:"webhook_url"`
	PagerDutyKey string        `mapstructure:"pagerduty_key"`
	Rules        AlertingRules `mapstructure:"rules"`
}

// AlertingRules contains alerting rules
type AlertingRules struct {
	HighCPU      int `mapstructure:"high_cpu"`
	HighMemory   int `mapstructure:"high_memory"`
	HighDisk     int `mapstructure:"high_disk"`
	ErrorRate    int `mapstructure:"error_rate"`
	ResponseTime int `mapstructure:"response_time"`
}

// NewManager creates a new configuration manager
func NewManager(environment, configPath string) *Manager {
	v := viper.New()

	return &Manager{
		viper:       v,
		environment: environment,
		configPath:  configPath,
	}
}

// Load loads the configuration from files and environment variables
func (m *Manager) Load() (*Config, error) {
	// Set default configuration path if not provided
	if m.configPath == "" {
		m.configPath = "configs"
	}

	// Set configuration file name based on environment
	configFile := fmt.Sprintf("environments/%s.yaml", m.environment)

	// Set up viper
	m.viper.SetConfigName(strings.TrimSuffix(configFile, filepath.Ext(configFile)))
	m.viper.SetConfigType("yaml")
	m.viper.AddConfigPath(m.configPath)
	m.viper.AddConfigPath(".")
	m.viper.AddConfigPath("./configs")
	m.viper.AddConfigPath("/etc/aios")
	m.viper.AddConfigPath("$HOME/.aios")

	// Enable environment variable support
	m.viper.AutomaticEnv()
	m.viper.SetEnvPrefix("AIOS")
	m.viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))

	// Read configuration file
	if err := m.viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			return nil, fmt.Errorf("configuration file not found: %s", configFile)
		}
		return nil, fmt.Errorf("failed to read configuration file: %w", err)
	}

	// Unmarshal configuration
	var config Config
	if err := m.viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal configuration: %w", err)
	}

	// Validate configuration
	if err := m.validate(&config); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return &config, nil
}

// validate validates the configuration
func (m *Manager) validate(config *Config) error {
	// Validate required fields
	if config.Environment == "" {
		return fmt.Errorf("environment is required")
	}

	if config.Server.Port <= 0 || config.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", config.Server.Port)
	}

	if config.Database.Host == "" {
		return fmt.Errorf("database host is required")
	}

	if config.Security.JWT.Secret == "" {
		return fmt.Errorf("JWT secret is required")
	}

	if config.Security.Encryption.Key == "" {
		return fmt.Errorf("encryption key is required")
	}

	// Validate encryption key length (must be 32 bytes for AES-256)
	if len(config.Security.Encryption.Key) != 32 {
		return fmt.Errorf("encryption key must be exactly 32 characters long")
	}

	return nil
}

// GetString returns a string configuration value
func (m *Manager) GetString(key string) string {
	return m.viper.GetString(key)
}

// GetInt returns an integer configuration value
func (m *Manager) GetInt(key string) int {
	return m.viper.GetInt(key)
}

// GetBool returns a boolean configuration value
func (m *Manager) GetBool(key string) bool {
	return m.viper.GetBool(key)
}

// GetDuration returns a duration configuration value
func (m *Manager) GetDuration(key string) time.Duration {
	return m.viper.GetDuration(key)
}

// Set sets a configuration value
func (m *Manager) Set(key string, value interface{}) {
	m.viper.Set(key, value)
}

// IsSet checks if a configuration key is set
func (m *Manager) IsSet(key string) bool {
	return m.viper.IsSet(key)
}

// WatchConfig watches for configuration file changes
func (m *Manager) WatchConfig(callback func()) {
	m.viper.WatchConfig()
	m.viper.OnConfigChange(func(e fsnotify.Event) {
		if callback != nil {
			callback()
		}
	})
}

// GetEnvironment returns the current environment
func (m *Manager) GetEnvironment() string {
	return m.environment
}

// IsDevelopment returns true if running in development environment
func (m *Manager) IsDevelopment() bool {
	return m.environment == "development"
}

// IsStaging returns true if running in staging environment
func (m *Manager) IsStaging() bool {
	return m.environment == "staging"
}

// IsProduction returns true if running in production environment
func (m *Manager) IsProduction() bool {
	return m.environment == "production"
}

// GetDatabaseURL returns the database connection URL
func (config *Config) GetDatabaseURL() string {
	return fmt.Sprintf("%s://%s:%s@%s:%d/%s?sslmode=%s",
		config.Database.Driver,
		config.Database.User,
		config.Database.Password,
		config.Database.Host,
		config.Database.Port,
		config.Database.Name,
		config.Database.SSLMode,
	)
}

// GetRedisURL returns the Redis connection URL
func (config *Config) GetRedisURL() string {
	if config.Redis.Password != "" {
		return fmt.Sprintf("redis://:%s@%s:%d/%d",
			config.Redis.Password,
			config.Redis.Host,
			config.Redis.Port,
			config.Redis.DB,
		)
	}
	return fmt.Sprintf("redis://%s:%d/%d",
		config.Redis.Host,
		config.Redis.Port,
		config.Redis.DB,
	)
}

// GetServerAddress returns the server address
func (config *Config) GetServerAddress() string {
	return fmt.Sprintf("%s:%d", config.Server.Host, config.Server.Port)
}

// GetMetricsAddress returns the metrics server address
func (config *Config) GetMetricsAddress() string {
	return fmt.Sprintf("%s:%d", config.Metrics.Host, config.Metrics.Port)
}
