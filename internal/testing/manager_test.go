package testing

import (
	"context"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTestingManager(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	config := TestingConfig{
		Enabled: true,
		UnitTesting: UnitTestingConfig{
			Enabled:  true,
			Parallel: 2,
			Timeout:  30 * time.Second,
			Verbose:  false,
			FailFast: false,
			Race:     false,
		},
		Integration: IntegrationTestingConfig{
			Enabled:         true,
			Timeout:         5 * time.Minute,
			SetupTimeout:    30 * time.Second,
			TeardownTimeout: 30 * time.Second,
			DatabaseTests:   true,
			APITests:        true,
			ServiceTests:    true,
			ExternalDeps:    false,
			Environment:     "test",
		},
		E2E: E2ETestingConfig{
			Enabled:     false, // Disabled for unit tests
			Browser:     "chrome",
			Headless:    true,
			Timeout:     10 * time.Minute,
			Screenshots: false,
			Videos:      false,
			BaseURL:     "http://localhost:8080",
			Retries:     2,
			Parallel:    1,
		},
		Performance: PerformanceTestingConfig{
			Enabled:       true,
			LoadTesting:   true,
			StressTesting: false,
			Benchmarks:    true,
			Profiling:     true,
			Duration:      1 * time.Minute,
			Concurrency:   10,
			RampUp:        10 * time.Second,
			Thresholds: map[string]float64{
				"response_time": 1000,
				"error_rate":    0.01,
			},
		},
		Security: SecurityTestingConfig{
			Enabled:             true,
			VulnerabilityScans:  true,
			PenetrationTesting:  false,
			AuthenticationTests: true,
			AuthorizationTests:  true,
			InputValidation:     true,
			SQLInjection:        true,
			XSS:                 true,
			CSRF:                true,
			SecurityHeaders:     true,
			TLSTests:            true,
			Tools:               []string{"gosec", "nancy"},
		},
		Validation: ValidationConfig{
			Enabled:          true,
			SchemaValidation: true,
			DataValidation:   true,
			APIValidation:    true,
			ConfigValidation: true,
			BusinessRules:    true,
			Constraints:      []string{"required", "format", "range"},
			CustomRules:      []string{"business_logic", "data_integrity"},
		},
		Coverage: CoverageConfig{
			Enabled:       true,
			MinCoverage:   80.0,
			FailOnLow:     false,
			ExcludePaths:  []string{"vendor", "mocks", "test"},
			IncludePaths:  []string{"internal", "pkg"},
			ReportFormats: []string{"html", "json"},
			OutputDir:     "./coverage",
		},
		Reporting: ReportingConfig{
			Enabled:   true,
			Formats:   []string{"junit", "html", "json"},
			OutputDir: "./test-reports",
			JUnit:     true,
			HTML:      true,
			JSON:      true,
			Allure:    false,
			Slack:     false,
			Email:     false,
		},
		CI: CIConfig{
			Enabled:        true,
			Provider:       "github",
			Pipeline:       "test",
			Stages:         []string{"unit", "integration", "security"},
			Artifacts:      true,
			Notifications:  true,
			FailureActions: []string{"notify", "rollback"},
			SuccessActions: []string{"deploy"},
		},
		Quality: QualityConfig{
			Enabled: true,
			Gates: map[string]float64{
				"coverage":        80.0,
				"duplication":     5.0,
				"maintainability": 70.0,
			},
			Metrics:         []string{"coverage", "complexity", "duplication"},
			Linting:         true,
			StaticAnalysis:  true,
			Complexity:      true,
			Duplication:     true,
			Maintainability: true,
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
		assert.NotNil(t, manager.GetUnitTester())
		assert.NotNil(t, manager.GetIntegrationTester())
		assert.NotNil(t, manager.GetE2ETester())
		assert.NotNil(t, manager.GetPerformanceTester())
		assert.NotNil(t, manager.GetSecurityTester())
		assert.NotNil(t, manager.GetValidationEngine())
		assert.NotNil(t, manager.GetCoverageAnalyzer())
		assert.NotNil(t, manager.GetTestReporter())
	})

	t.Run("RunTestSuite", func(t *testing.T) {
		manager, err := NewManager(logger, config)
		require.NoError(t, err)

		ctx := context.Background()
		err = manager.Start(ctx)
		require.NoError(t, err)
		defer manager.Stop(ctx)

		// Run unit tests
		result, err := manager.RunTestSuite(ctx, "unit")
		require.NoError(t, err)
		assert.Equal(t, "unit", result.Type)
		assert.NotEmpty(t, result.ID)

		// Run integration tests
		result, err = manager.RunTestSuite(ctx, "integration")
		require.NoError(t, err)
		assert.Equal(t, "integration", result.Type)

		// Run performance tests
		result, err = manager.RunTestSuite(ctx, "performance")
		require.NoError(t, err)
		assert.Equal(t, "performance", result.Type)

		// Run security tests
		result, err = manager.RunTestSuite(ctx, "security")
		require.NoError(t, err)
		assert.Equal(t, "security", result.Type)

		// Test invalid suite type
		_, err = manager.RunTestSuite(ctx, "invalid")
		assert.Error(t, err)
	})

	t.Run("ValidateData", func(t *testing.T) {
		manager, err := NewManager(logger, config)
		require.NoError(t, err)

		ctx := context.Background()
		err = manager.Start(ctx)
		require.NoError(t, err)
		defer manager.Stop(ctx)

		// Test data validation
		data := map[string]interface{}{
			"name":  "test",
			"value": 42,
		}

		result, err := manager.ValidateData(ctx, data, "test-schema")
		require.NoError(t, err)
		assert.True(t, result.Valid)
		assert.Equal(t, "data", result.Type)
		assert.Equal(t, "test-schema", result.Schema)
	})

	t.Run("ValidateAPI", func(t *testing.T) {
		manager, err := NewManager(logger, config)
		require.NoError(t, err)

		ctx := context.Background()
		err = manager.Start(ctx)
		require.NoError(t, err)
		defer manager.Stop(ctx)

		// Test API validation
		result, err := manager.ValidateAPI(ctx, "/api/v1/test")
		require.NoError(t, err)
		assert.True(t, result.Valid)
		assert.Equal(t, "api", result.Type)
		assert.Equal(t, "/api/v1/test", result.Target)
	})

	t.Run("GetCoverageReport", func(t *testing.T) {
		manager, err := NewManager(logger, config)
		require.NoError(t, err)

		ctx := context.Background()
		err = manager.Start(ctx)
		require.NoError(t, err)
		defer manager.Stop(ctx)

		// Get coverage report
		report, err := manager.GetCoverageReport(ctx)
		require.NoError(t, err)
		assert.NotEmpty(t, report.ID)
		assert.Greater(t, report.OverallCoverage, 0.0)
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

func TestUnitTester(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	config := UnitTestingConfig{
		Enabled:  true,
		Parallel: 2,
		Timeout:  30 * time.Second,
		Verbose:  true,
		FailFast: false,
		Race:     false,
	}

	t.Run("NewUnitTester", func(t *testing.T) {
		tester, err := NewUnitTester(logger, config)
		require.NoError(t, err)
		assert.NotNil(t, tester)
	})

	t.Run("StartStop", func(t *testing.T) {
		tester, err := NewUnitTester(logger, config)
		require.NoError(t, err)

		ctx := context.Background()

		err = tester.Start(ctx)
		require.NoError(t, err)

		status, err := tester.GetStatus(ctx)
		require.NoError(t, err)
		assert.True(t, status.Enabled)

		err = tester.Stop(ctx)
		require.NoError(t, err)
	})

	t.Run("RunTests", func(t *testing.T) {
		tester, err := NewUnitTester(logger, config)
		require.NoError(t, err)

		ctx := context.Background()
		err = tester.Start(ctx)
		require.NoError(t, err)
		defer tester.Stop(ctx)

		// Note: This will actually try to run go test, which may fail
		// In a real test environment, we'd mock the command execution
		result, err := tester.RunTests(ctx)
		require.NoError(t, err)
		assert.Equal(t, "unit", result.Type)
		assert.NotEmpty(t, result.ID)
		assert.Greater(t, result.TestsRun, 0)
	})
}

func TestValidationEngine(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	config := ValidationConfig{
		Enabled:          true,
		SchemaValidation: true,
		DataValidation:   true,
		APIValidation:    true,
		ConfigValidation: true,
		BusinessRules:    true,
	}

	t.Run("NewValidationEngine", func(t *testing.T) {
		engine, err := NewValidationEngine(logger, config)
		require.NoError(t, err)
		assert.NotNil(t, engine)
	})

	t.Run("ValidateData", func(t *testing.T) {
		engine, err := NewValidationEngine(logger, config)
		require.NoError(t, err)

		ctx := context.Background()
		err = engine.Start(ctx)
		require.NoError(t, err)
		defer engine.Stop(ctx)

		data := map[string]interface{}{
			"id":   1,
			"name": "test",
		}

		result, err := engine.ValidateData(ctx, data, "user-schema")
		require.NoError(t, err)
		assert.True(t, result.Valid)
		assert.Equal(t, "data", result.Type)
		assert.Equal(t, "user-schema", result.Schema)
		assert.Empty(t, result.Errors)
	})

	t.Run("ValidateAPI", func(t *testing.T) {
		engine, err := NewValidationEngine(logger, config)
		require.NoError(t, err)

		ctx := context.Background()
		err = engine.Start(ctx)
		require.NoError(t, err)
		defer engine.Stop(ctx)

		result, err := engine.ValidateAPI(ctx, "/api/v1/users")
		require.NoError(t, err)
		assert.True(t, result.Valid)
		assert.Equal(t, "api", result.Type)
		assert.Equal(t, "/api/v1/users", result.Target)
		assert.Empty(t, result.Errors)
	})
}
