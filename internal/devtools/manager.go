package devtools

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/aios/aios/pkg/models"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

// Manager handles developer tools and debugging capabilities
type Manager struct {
	logger           *logrus.Logger
	tracer           trace.Tracer
	config           DevToolsConfig
	debugger         *Debugger
	profiler         *Profiler
	codeAnalyzer     *CodeAnalyzer
	testRunner       *TestRunner
	buildManager     *BuildManager
	liveReloader     *LiveReloader
	logAnalyzer      *LogAnalyzer
	metricsCollector *MetricsCollector
	mu               sync.RWMutex
	running          bool
	stopCh           chan struct{}
}

// DevToolsConfig represents developer tools configuration
type DevToolsConfig struct {
	Enabled      bool               `yaml:"enabled"`
	Debug        DebugConfig        `yaml:"debug"`
	Profiling    ProfilingConfig    `yaml:"profiling"`
	CodeAnalysis CodeAnalysisConfig `yaml:"code_analysis"`
	Testing      TestingConfig      `yaml:"testing"`
	Build        BuildConfig        `yaml:"build"`
	LiveReload   LiveReloadConfig   `yaml:"live_reload"`
	LogAnalysis  LogAnalysisConfig  `yaml:"log_analysis"`
	Metrics      MetricsConfig      `yaml:"metrics"`
	Environment  EnvironmentConfig  `yaml:"environment"`
	Security     SecurityConfig     `yaml:"security"`
}

// DebugConfig represents debugging configuration
type DebugConfig struct {
	Enabled         bool     `yaml:"enabled"`
	Port            int      `yaml:"port"`
	RemoteDebugging bool     `yaml:"remote_debugging"`
	Breakpoints     []string `yaml:"breakpoints"`
	WatchVariables  []string `yaml:"watch_variables"`
	StackTraceDepth int      `yaml:"stack_trace_depth"`
}

// ProfilingConfig represents profiling configuration
type ProfilingConfig struct {
	Enabled            bool          `yaml:"enabled"`
	CPUProfiling       bool          `yaml:"cpu_profiling"`
	MemoryProfiling    bool          `yaml:"memory_profiling"`
	GoroutineProfiling bool          `yaml:"goroutine_profiling"`
	BlockProfiling     bool          `yaml:"block_profiling"`
	MutexProfiling     bool          `yaml:"mutex_profiling"`
	ProfileDuration    time.Duration `yaml:"profile_duration"`
	OutputDir          string        `yaml:"output_dir"`
}

// CodeAnalysisConfig represents code analysis configuration
type CodeAnalysisConfig struct {
	Enabled         bool     `yaml:"enabled"`
	StaticAnalysis  bool     `yaml:"static_analysis"`
	SecurityScan    bool     `yaml:"security_scan"`
	QualityMetrics  bool     `yaml:"quality_metrics"`
	DependencyCheck bool     `yaml:"dependency_check"`
	LintRules       []string `yaml:"lint_rules"`
	ExcludePaths    []string `yaml:"exclude_paths"`
}

// TestingConfig represents testing configuration
type TestingConfig struct {
	Enabled     bool          `yaml:"enabled"`
	AutoRun     bool          `yaml:"auto_run"`
	Coverage    bool          `yaml:"coverage"`
	Benchmarks  bool          `yaml:"benchmarks"`
	Integration bool          `yaml:"integration"`
	E2E         bool          `yaml:"e2e"`
	Timeout     time.Duration `yaml:"timeout"`
	Parallel    int           `yaml:"parallel"`
	Verbose     bool          `yaml:"verbose"`
}

// BuildConfig represents build configuration
type BuildConfig struct {
	Enabled         bool     `yaml:"enabled"`
	AutoBuild       bool     `yaml:"auto_build"`
	OptimizedBuild  bool     `yaml:"optimized_build"`
	CrossCompile    bool     `yaml:"cross_compile"`
	TargetPlatforms []string `yaml:"target_platforms"`
	BuildTags       []string `yaml:"build_tags"`
	LDFlags         string   `yaml:"ld_flags"`
	OutputDir       string   `yaml:"output_dir"`
}

// LiveReloadConfig represents live reload configuration
type LiveReloadConfig struct {
	Enabled        bool          `yaml:"enabled"`
	Port           int           `yaml:"port"`
	WatchPaths     []string      `yaml:"watch_paths"`
	IgnorePatterns []string      `yaml:"ignore_patterns"`
	Extensions     []string      `yaml:"extensions"`
	Delay          time.Duration `yaml:"delay"`
}

// LogAnalysisConfig represents log analysis configuration
type LogAnalysisConfig struct {
	Enabled             bool               `yaml:"enabled"`
	RealTime            bool               `yaml:"real_time"`
	ErrorDetection      bool               `yaml:"error_detection"`
	PerformanceAnalysis bool               `yaml:"performance_analysis"`
	LogSources          []string           `yaml:"log_sources"`
	AlertThresholds     map[string]float64 `yaml:"alert_thresholds"`
}

// MetricsConfig represents metrics configuration
type MetricsConfig struct {
	Enabled            bool          `yaml:"enabled"`
	CollectionInterval time.Duration `yaml:"collection_interval"`
	CustomMetrics      bool          `yaml:"custom_metrics"`
	PerformanceMetrics bool          `yaml:"performance_metrics"`
	BusinessMetrics    bool          `yaml:"business_metrics"`
	ExportFormat       string        `yaml:"export_format"`
}

// EnvironmentConfig represents development environment configuration
type EnvironmentConfig struct {
	Mode            string            `yaml:"mode"`
	HotReload       bool              `yaml:"hot_reload"`
	MockServices    bool              `yaml:"mock_services"`
	DatabaseSeeding bool              `yaml:"database_seeding"`
	FeatureFlags    map[string]bool   `yaml:"feature_flags"`
	Environment     map[string]string `yaml:"environment"`
}

// SecurityConfig represents security configuration for dev tools
type SecurityConfig struct {
	Enabled               bool     `yaml:"enabled"`
	VulnerabilityScanning bool     `yaml:"vulnerability_scanning"`
	SecretDetection       bool     `yaml:"secret_detection"`
	AccessControl         bool     `yaml:"access_control"`
	AuditLogging          bool     `yaml:"audit_logging"`
	AllowedIPs            []string `yaml:"allowed_ips"`
}

// NewManager creates a new developer tools manager
func NewManager(logger *logrus.Logger, config DevToolsConfig) (*Manager, error) {
	tracer := otel.Tracer("devtools-manager")

	// Initialize components
	debugger, err := NewDebugger(logger, config.Debug)
	if err != nil {
		return nil, fmt.Errorf("failed to create debugger: %w", err)
	}

	profiler, err := NewProfiler(logger, config.Profiling)
	if err != nil {
		return nil, fmt.Errorf("failed to create profiler: %w", err)
	}

	codeAnalyzer, err := NewCodeAnalyzer(logger, config.CodeAnalysis)
	if err != nil {
		return nil, fmt.Errorf("failed to create code analyzer: %w", err)
	}

	testRunner, err := NewTestRunner(logger, config.Testing)
	if err != nil {
		return nil, fmt.Errorf("failed to create test runner: %w", err)
	}

	buildManager, err := NewBuildManager(logger, config.Build)
	if err != nil {
		return nil, fmt.Errorf("failed to create build manager: %w", err)
	}

	liveReloader, err := NewLiveReloader(logger, config.LiveReload)
	if err != nil {
		return nil, fmt.Errorf("failed to create live reloader: %w", err)
	}

	logAnalyzer, err := NewLogAnalyzer(logger, config.LogAnalysis)
	if err != nil {
		return nil, fmt.Errorf("failed to create log analyzer: %w", err)
	}

	metricsCollector, err := NewMetricsCollector(logger, config.Metrics)
	if err != nil {
		return nil, fmt.Errorf("failed to create metrics collector: %w", err)
	}

	return &Manager{
		logger:           logger,
		tracer:           tracer,
		config:           config,
		debugger:         debugger,
		profiler:         profiler,
		codeAnalyzer:     codeAnalyzer,
		testRunner:       testRunner,
		buildManager:     buildManager,
		liveReloader:     liveReloader,
		logAnalyzer:      logAnalyzer,
		metricsCollector: metricsCollector,
		stopCh:           make(chan struct{}),
	}, nil
}

// Start initializes the developer tools manager
func (m *Manager) Start(ctx context.Context) error {
	ctx, span := m.tracer.Start(ctx, "devtools.Manager.Start")
	defer span.End()

	m.mu.Lock()
	defer m.mu.Unlock()

	if m.running {
		return fmt.Errorf("developer tools manager is already running")
	}

	if !m.config.Enabled {
		m.logger.Info("Developer tools are disabled")
		return nil
	}

	m.logger.Info("Starting developer tools manager")

	// Start components
	if err := m.debugger.Start(ctx); err != nil {
		return fmt.Errorf("failed to start debugger: %w", err)
	}

	if err := m.profiler.Start(ctx); err != nil {
		return fmt.Errorf("failed to start profiler: %w", err)
	}

	if err := m.codeAnalyzer.Start(ctx); err != nil {
		return fmt.Errorf("failed to start code analyzer: %w", err)
	}

	if err := m.testRunner.Start(ctx); err != nil {
		return fmt.Errorf("failed to start test runner: %w", err)
	}

	if err := m.buildManager.Start(ctx); err != nil {
		return fmt.Errorf("failed to start build manager: %w", err)
	}

	if err := m.liveReloader.Start(ctx); err != nil {
		return fmt.Errorf("failed to start live reloader: %w", err)
	}

	if err := m.logAnalyzer.Start(ctx); err != nil {
		return fmt.Errorf("failed to start log analyzer: %w", err)
	}

	if err := m.metricsCollector.Start(ctx); err != nil {
		return fmt.Errorf("failed to start metrics collector: %w", err)
	}

	// Start monitoring
	go m.monitorDevTools()

	m.running = true
	m.logger.Info("Developer tools manager started successfully")

	return nil
}

// Stop shuts down the developer tools manager
func (m *Manager) Stop(ctx context.Context) error {
	ctx, span := m.tracer.Start(ctx, "devtools.Manager.Stop")
	defer span.End()

	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.running {
		return nil
	}

	m.logger.Info("Stopping developer tools manager")

	// Stop components in reverse order
	if err := m.metricsCollector.Stop(ctx); err != nil {
		m.logger.WithError(err).Error("Failed to stop metrics collector")
	}

	if err := m.logAnalyzer.Stop(ctx); err != nil {
		m.logger.WithError(err).Error("Failed to stop log analyzer")
	}

	if err := m.liveReloader.Stop(ctx); err != nil {
		m.logger.WithError(err).Error("Failed to stop live reloader")
	}

	if err := m.buildManager.Stop(ctx); err != nil {
		m.logger.WithError(err).Error("Failed to stop build manager")
	}

	if err := m.testRunner.Stop(ctx); err != nil {
		m.logger.WithError(err).Error("Failed to stop test runner")
	}

	if err := m.codeAnalyzer.Stop(ctx); err != nil {
		m.logger.WithError(err).Error("Failed to stop code analyzer")
	}

	if err := m.profiler.Stop(ctx); err != nil {
		m.logger.WithError(err).Error("Failed to stop profiler")
	}

	if err := m.debugger.Stop(ctx); err != nil {
		m.logger.WithError(err).Error("Failed to stop debugger")
	}

	close(m.stopCh)
	m.running = false
	m.logger.Info("Developer tools manager stopped")

	return nil
}

// GetStatus returns the current developer tools status
func (m *Manager) GetStatus(ctx context.Context) (*models.DevToolsStatus, error) {
	ctx, span := m.tracer.Start(ctx, "devtools.Manager.GetStatus")
	defer span.End()

	m.mu.RLock()
	defer m.mu.RUnlock()

	if !m.config.Enabled {
		return &models.DevToolsStatus{
			Enabled:   false,
			Running:   false,
			Timestamp: time.Now(),
		}, nil
	}

	// Get component statuses
	debuggerStatus, _ := m.debugger.GetStatus(ctx)
	profilerStatus, _ := m.profiler.GetStatus(ctx)
	codeAnalyzerStatus, _ := m.codeAnalyzer.GetStatus(ctx)
	testRunnerStatus, _ := m.testRunner.GetStatus(ctx)
	buildManagerStatus, _ := m.buildManager.GetStatus(ctx)
	liveReloaderStatus, _ := m.liveReloader.GetStatus(ctx)
	logAnalyzerStatus, _ := m.logAnalyzer.GetStatus(ctx)
	metricsCollectorStatus, _ := m.metricsCollector.GetStatus(ctx)

	return &models.DevToolsStatus{
		Enabled:          m.config.Enabled,
		Running:          m.running,
		Debugger:         debuggerStatus,
		Profiler:         profilerStatus,
		CodeAnalyzer:     codeAnalyzerStatus,
		TestRunner:       testRunnerStatus,
		BuildManager:     buildManagerStatus,
		LiveReloader:     liveReloaderStatus,
		LogAnalyzer:      logAnalyzerStatus,
		MetricsCollector: metricsCollectorStatus,
		Timestamp:        time.Now(),
	}, nil
}

// GetDebugger returns the debugger instance
func (m *Manager) GetDebugger() *Debugger {
	return m.debugger
}

// GetProfiler returns the profiler instance
func (m *Manager) GetProfiler() *Profiler {
	return m.profiler
}

// GetCodeAnalyzer returns the code analyzer instance
func (m *Manager) GetCodeAnalyzer() *CodeAnalyzer {
	return m.codeAnalyzer
}

// GetTestRunner returns the test runner instance
func (m *Manager) GetTestRunner() *TestRunner {
	return m.testRunner
}

// GetBuildManager returns the build manager instance
func (m *Manager) GetBuildManager() *BuildManager {
	return m.buildManager
}

// GetLiveReloader returns the live reloader instance
func (m *Manager) GetLiveReloader() *LiveReloader {
	return m.liveReloader
}

// GetLogAnalyzer returns the log analyzer instance
func (m *Manager) GetLogAnalyzer() *LogAnalyzer {
	return m.logAnalyzer
}

// GetMetricsCollector returns the metrics collector instance
func (m *Manager) GetMetricsCollector() *MetricsCollector {
	return m.metricsCollector
}

// monitorDevTools continuously monitors developer tools
func (m *Manager) monitorDevTools() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			ctx := context.Background()
			status, err := m.GetStatus(ctx)
			if err != nil {
				m.logger.WithError(err).Error("Failed to get developer tools status")
				continue
			}

			// Log status for monitoring
			m.logger.WithFields(logrus.Fields{
				"enabled": status.Enabled,
				"running": status.Running,
			}).Debug("Developer tools status")

		case <-m.stopCh:
			m.logger.Debug("Developer tools monitoring stopped")
			return
		}
	}
}
