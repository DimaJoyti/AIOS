package devtools

import (
	"context"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDevToolsManager(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	config := DevToolsConfig{
		Enabled: true,
		Debug: DebugConfig{
			Enabled:         true,
			Port:            2345,
			RemoteDebugging: false,
			Breakpoints:     []string{},
			WatchVariables:  []string{},
			StackTraceDepth: 10,
		},
		Profiling: ProfilingConfig{
			Enabled:            true,
			CPUProfiling:       true,
			MemoryProfiling:    true,
			GoroutineProfiling: true,
			BlockProfiling:     false,
			MutexProfiling:     false,
			ProfileDuration:    1 * time.Second,
			OutputDir:          "/tmp/test-profiles",
		},
		CodeAnalysis: CodeAnalysisConfig{
			Enabled:         true,
			StaticAnalysis:  true,
			SecurityScan:    false,
			QualityMetrics:  true,
			DependencyCheck: false,
			LintRules:       []string{"gofmt", "govet"},
			ExcludePaths:    []string{"vendor", ".git"},
		},
		Testing: TestingConfig{
			Enabled:     true,
			AutoRun:     false,
			Coverage:    true,
			Benchmarks:  false,
			Integration: false,
			E2E:         false,
			Timeout:     30 * time.Second,
			Parallel:    2,
			Verbose:     false,
		},
		Build: BuildConfig{
			Enabled:         true,
			AutoBuild:       false,
			OptimizedBuild:  false,
			CrossCompile:    false,
			TargetPlatforms: []string{"linux/amd64"},
			BuildTags:       []string{},
			LDFlags:         "",
			OutputDir:       "/tmp/test-builds",
		},
		LiveReload: LiveReloadConfig{
			Enabled:        false,
			Port:           3001,
			WatchPaths:     []string{"./"},
			IgnorePatterns: []string{"*.log"},
			Extensions:     []string{".go"},
			Delay:          100 * time.Millisecond,
		},
		LogAnalysis: LogAnalysisConfig{
			Enabled:         true,
			RealTime:        false,
			ErrorDetection:  true,
			PerformanceAnalysis: false,
			LogSources:      []string{"/tmp/test.log"},
			AlertThresholds: map[string]float64{
				"error_rate": 0.1,
			},
		},
		Metrics: MetricsConfig{
			Enabled:            true,
			CollectionInterval: 1 * time.Second,
			CustomMetrics:      false,
			PerformanceMetrics: true,
			BusinessMetrics:    false,
			ExportFormat:       "json",
		},
	}

	t.Run("NewManager", func(t *testing.T) {
		manager, err := NewManager(logger, config)
		require.NoError(t, err)
		assert.NotNil(t, manager)
		assert.Equal(t, config.Enabled, manager.config.Enabled)
	})

	t.Run("StartStop", func(t *testing.T) {
		manager, err := NewManager(logger, config)
		require.NoError(t, err)

		ctx := context.Background()

		// Start manager
		err = manager.Start(ctx)
		require.NoError(t, err)

		// Check status
		status, err := manager.GetStatus(ctx)
		require.NoError(t, err)
		assert.True(t, status.Enabled)
		assert.True(t, status.Running)

		// Stop manager
		err = manager.Stop(ctx)
		require.NoError(t, err)

		// Check status after stop
		status, err = manager.GetStatus(ctx)
		require.NoError(t, err)
		assert.True(t, status.Enabled)
		assert.False(t, status.Running)
	})

	t.Run("GetComponents", func(t *testing.T) {
		manager, err := NewManager(logger, config)
		require.NoError(t, err)

		// Test component getters
		assert.NotNil(t, manager.GetDebugger())
		assert.NotNil(t, manager.GetProfiler())
		assert.NotNil(t, manager.GetCodeAnalyzer())
		assert.NotNil(t, manager.GetTestRunner())
		assert.NotNil(t, manager.GetBuildManager())
		assert.NotNil(t, manager.GetLiveReloader())
		assert.NotNil(t, manager.GetLogAnalyzer())
		assert.NotNil(t, manager.GetMetricsCollector())
	})

	t.Run("DisabledManager", func(t *testing.T) {
		disabledConfig := config
		disabledConfig.Enabled = false

		manager, err := NewManager(logger, disabledConfig)
		require.NoError(t, err)

		ctx := context.Background()

		// Start disabled manager
		err = manager.Start(ctx)
		require.NoError(t, err)

		// Check status
		status, err := manager.GetStatus(ctx)
		require.NoError(t, err)
		assert.False(t, status.Enabled)
		assert.False(t, status.Running)
	})
}

func TestDebugger(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	config := DebugConfig{
		Enabled:         true,
		Port:            2345,
		RemoteDebugging: false,
		Breakpoints:     []string{},
		WatchVariables:  []string{},
		StackTraceDepth: 10,
	}

	t.Run("NewDebugger", func(t *testing.T) {
		debugger, err := NewDebugger(logger, config)
		require.NoError(t, err)
		assert.NotNil(t, debugger)
	})

	t.Run("StartStop", func(t *testing.T) {
		debugger, err := NewDebugger(logger, config)
		require.NoError(t, err)

		ctx := context.Background()

		err = debugger.Start(ctx)
		require.NoError(t, err)

		status, err := debugger.GetStatus(ctx)
		require.NoError(t, err)
		assert.True(t, status.Enabled)
		assert.True(t, status.Running)

		err = debugger.Stop(ctx)
		require.NoError(t, err)
	})

	t.Run("Breakpoints", func(t *testing.T) {
		debugger, err := NewDebugger(logger, config)
		require.NoError(t, err)

		ctx := context.Background()
		err = debugger.Start(ctx)
		require.NoError(t, err)
		defer debugger.Stop(ctx)

		// Set breakpoint
		bp, err := debugger.SetBreakpoint(ctx, "test.go", 42, "x > 10")
		require.NoError(t, err)
		assert.Equal(t, "test.go:42", bp.ID)
		assert.Equal(t, "test.go", bp.File)
		assert.Equal(t, 42, bp.Line)
		assert.Equal(t, "x > 10", bp.Condition)

		// List breakpoints
		breakpoints, err := debugger.ListBreakpoints(ctx)
		require.NoError(t, err)
		assert.Len(t, breakpoints, 1)

		// Remove breakpoint
		err = debugger.RemoveBreakpoint(ctx, bp.ID)
		require.NoError(t, err)

		// List breakpoints after removal
		breakpoints, err = debugger.ListBreakpoints(ctx)
		require.NoError(t, err)
		assert.Len(t, breakpoints, 0)
	})

	t.Run("DebugSessions", func(t *testing.T) {
		debugger, err := NewDebugger(logger, config)
		require.NoError(t, err)

		ctx := context.Background()
		err = debugger.Start(ctx)
		require.NoError(t, err)
		defer debugger.Stop(ctx)

		// Start debug session
		session, err := debugger.StartDebugSession(ctx, "test-target")
		require.NoError(t, err)
		assert.NotEmpty(t, session.ID)
		assert.Equal(t, "test-target", session.Target)
		assert.True(t, session.Active)

		// Get stack trace
		stackTrace, err := debugger.GetStackTrace(ctx, session.ID)
		require.NoError(t, err)
		assert.NotNil(t, stackTrace)

		// Get variables
		variables, err := debugger.GetVariables(ctx, session.ID, "local")
		require.NoError(t, err)
		assert.NotNil(t, variables)

		// Evaluate expression
		result, err := debugger.EvaluateExpression(ctx, session.ID, "1 + 1")
		require.NoError(t, err)
		assert.NotNil(t, result)

		// Stop debug session
		err = debugger.StopDebugSession(ctx, session.ID)
		require.NoError(t, err)
	})
}

func TestProfiler(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	config := ProfilingConfig{
		Enabled:            true,
		CPUProfiling:       true,
		MemoryProfiling:    true,
		GoroutineProfiling: true,
		BlockProfiling:     false,
		MutexProfiling:     false,
		ProfileDuration:    100 * time.Millisecond,
		OutputDir:          "/tmp/test-profiles",
	}

	t.Run("NewProfiler", func(t *testing.T) {
		profiler, err := NewProfiler(logger, config)
		require.NoError(t, err)
		assert.NotNil(t, profiler)
	})

	t.Run("StartStop", func(t *testing.T) {
		profiler, err := NewProfiler(logger, config)
		require.NoError(t, err)

		ctx := context.Background()

		err = profiler.Start(ctx)
		require.NoError(t, err)

		status, err := profiler.GetStatus(ctx)
		require.NoError(t, err)
		assert.True(t, status.Enabled)
		assert.True(t, status.Running)

		err = profiler.Stop(ctx)
		require.NoError(t, err)
	})

	t.Run("MemoryProfile", func(t *testing.T) {
		profiler, err := NewProfiler(logger, config)
		require.NoError(t, err)

		ctx := context.Background()
		err = profiler.Start(ctx)
		require.NoError(t, err)
		defer profiler.Stop(ctx)

		// Create memory profile
		profile, err := profiler.CreateMemoryProfile(ctx)
		require.NoError(t, err)
		assert.NotEmpty(t, profile.ID)
		assert.Equal(t, "memory", profile.Type)
		assert.False(t, profile.Active)

		// List profiles
		profiles, err := profiler.ListProfiles(ctx)
		require.NoError(t, err)
		assert.Len(t, profiles, 1)

		// Get specific profile
		retrievedProfile, err := profiler.GetProfile(ctx, profile.ID)
		require.NoError(t, err)
		assert.Equal(t, profile.ID, retrievedProfile.ID)

		// Delete profile
		err = profiler.DeleteProfile(ctx, profile.ID)
		require.NoError(t, err)
	})

	t.Run("RuntimeStats", func(t *testing.T) {
		profiler, err := NewProfiler(logger, config)
		require.NoError(t, err)

		ctx := context.Background()

		stats, err := profiler.GetRuntimeStats(ctx)
		require.NoError(t, err)
		assert.Greater(t, stats.Goroutines, 0)
		assert.Greater(t, stats.HeapAlloc, uint64(0))
	})
}
