package testing

import (
	"context"
	"fmt"
	"os/exec"
	"sync"
	"time"

	"github.com/aios/aios/pkg/models"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

// UnitTester handles unit test execution
type UnitTester struct {
	logger  *logrus.Logger
	tracer  trace.Tracer
	config  UnitTestingConfig
	mu      sync.RWMutex
	running bool
	stopCh  chan struct{}
}

// NewUnitTester creates a new unit tester
func NewUnitTester(logger *logrus.Logger, config UnitTestingConfig) (*UnitTester, error) {
	tracer := otel.Tracer("unit-tester")

	return &UnitTester{
		logger: logger,
		tracer: tracer,
		config: config,
		stopCh: make(chan struct{}),
	}, nil
}

// Start initializes the unit tester
func (ut *UnitTester) Start(ctx context.Context) error {
	ut.mu.Lock()
	defer ut.mu.Unlock()

	if !ut.config.Enabled {
		ut.logger.Info("Unit tester is disabled")
		return nil
	}

	ut.running = true
	ut.logger.Info("Unit tester started")
	return nil
}

// Stop shuts down the unit tester
func (ut *UnitTester) Stop(ctx context.Context) error {
	ut.mu.Lock()
	defer ut.mu.Unlock()

	if !ut.running {
		return nil
	}

	close(ut.stopCh)
	ut.running = false
	ut.logger.Info("Unit tester stopped")
	return nil
}

// GetStatus returns the current unit testing status
func (ut *UnitTester) GetStatus(ctx context.Context) (*models.UnitTestingStatus, error) {
	ut.mu.RLock()
	defer ut.mu.RUnlock()

	return &models.UnitTestingStatus{
		Enabled:   ut.config.Enabled,
		Timestamp: time.Now(),
	}, nil
}

// RunTests executes unit tests
func (ut *UnitTester) RunTests(ctx context.Context) (*models.TestResult, error) {
	ctx, span := ut.tracer.Start(ctx, "unitTester.RunTests")
	defer span.End()

	ut.logger.Info("Running unit tests")

	result := &models.TestResult{
		ID:        fmt.Sprintf("unit-%d", time.Now().Unix()),
		Type:      "unit",
		Name:      "Unit Tests",
		StartTime: time.Now(),
		Status:    "running",
	}

	// Build test command
	args := []string{"test"}
	if ut.config.Verbose {
		args = append(args, "-v")
	}
	if ut.config.Race {
		args = append(args, "-race")
	}
	if ut.config.ShortMode {
		args = append(args, "-short")
	}
	if ut.config.Parallel > 0 {
		args = append(args, fmt.Sprintf("-parallel=%d", ut.config.Parallel))
	}
	args = append(args, "./...")

	// Execute tests
	cmd := exec.CommandContext(ctx, "go", args...)
	output, err := cmd.CombinedOutput()

	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)
	result.Output = string(output)

	if err != nil {
		result.Status = "failed"
		result.TestsFailed = 1
	} else {
		result.Status = "passed"
		result.TestsPassed = 1
	}

	result.TestsRun = result.TestsPassed + result.TestsFailed

	ut.logger.WithFields(logrus.Fields{
		"duration": result.Duration,
		"status":   result.Status,
	}).Info("Unit tests completed")

	return result, nil
}

// IntegrationTester handles integration test execution
type IntegrationTester struct {
	logger  *logrus.Logger
	tracer  trace.Tracer
	config  IntegrationTestingConfig
	mu      sync.RWMutex
	running bool
	stopCh  chan struct{}
}

// NewIntegrationTester creates a new integration tester
func NewIntegrationTester(logger *logrus.Logger, config IntegrationTestingConfig) (*IntegrationTester, error) {
	tracer := otel.Tracer("integration-tester")

	return &IntegrationTester{
		logger: logger,
		tracer: tracer,
		config: config,
		stopCh: make(chan struct{}),
	}, nil
}

// Start initializes the integration tester
func (it *IntegrationTester) Start(ctx context.Context) error {
	it.mu.Lock()
	defer it.mu.Unlock()

	if !it.config.Enabled {
		it.logger.Info("Integration tester is disabled")
		return nil
	}

	it.running = true
	it.logger.Info("Integration tester started")
	return nil
}

// Stop shuts down the integration tester
func (it *IntegrationTester) Stop(ctx context.Context) error {
	it.mu.Lock()
	defer it.mu.Unlock()

	if !it.running {
		return nil
	}

	close(it.stopCh)
	it.running = false
	it.logger.Info("Integration tester stopped")
	return nil
}

// GetStatus returns the current integration testing status
func (it *IntegrationTester) GetStatus(ctx context.Context) (*models.IntegrationTestingStatus, error) {
	it.mu.RLock()
	defer it.mu.RUnlock()

	return &models.IntegrationTestingStatus{
		Enabled:   it.config.Enabled,
		Timestamp: time.Now(),
	}, nil
}

// RunTests executes integration tests
func (it *IntegrationTester) RunTests(ctx context.Context) (*models.TestResult, error) {
	ctx, span := it.tracer.Start(ctx, "integrationTester.RunTests")
	defer span.End()

	it.logger.Info("Running integration tests")

	result := &models.TestResult{
		ID:        fmt.Sprintf("integration-%d", time.Now().Unix()),
		Type:      "integration",
		Name:      "Integration Tests",
		StartTime: time.Now(),
		Status:    "running",
	}

	// Execute integration tests
	cmd := exec.CommandContext(ctx, "go", "test", "-tags=integration", "./tests/integration/...")
	output, err := cmd.CombinedOutput()

	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)
	result.Output = string(output)

	if err != nil {
		result.Status = "failed"
		result.TestsFailed = 1
	} else {
		result.Status = "passed"
		result.TestsPassed = 1
	}

	result.TestsRun = result.TestsPassed + result.TestsFailed

	it.logger.WithFields(logrus.Fields{
		"duration": result.Duration,
		"status":   result.Status,
	}).Info("Integration tests completed")

	return result, nil
}

// E2ETester handles end-to-end test execution
type E2ETester struct {
	logger  *logrus.Logger
	tracer  trace.Tracer
	config  E2ETestingConfig
	mu      sync.RWMutex
	running bool
	stopCh  chan struct{}
}

// NewE2ETester creates a new E2E tester
func NewE2ETester(logger *logrus.Logger, config E2ETestingConfig) (*E2ETester, error) {
	tracer := otel.Tracer("e2e-tester")

	return &E2ETester{
		logger: logger,
		tracer: tracer,
		config: config,
		stopCh: make(chan struct{}),
	}, nil
}

// Start initializes the E2E tester
func (et *E2ETester) Start(ctx context.Context) error {
	et.mu.Lock()
	defer et.mu.Unlock()

	if !et.config.Enabled {
		et.logger.Info("E2E tester is disabled")
		return nil
	}

	et.running = true
	et.logger.Info("E2E tester started")
	return nil
}

// Stop shuts down the E2E tester
func (et *E2ETester) Stop(ctx context.Context) error {
	et.mu.Lock()
	defer et.mu.Unlock()

	if !et.running {
		return nil
	}

	close(et.stopCh)
	et.running = false
	et.logger.Info("E2E tester stopped")
	return nil
}

// GetStatus returns the current E2E testing status
func (et *E2ETester) GetStatus(ctx context.Context) (*models.E2ETestingStatus, error) {
	et.mu.RLock()
	defer et.mu.RUnlock()

	return &models.E2ETestingStatus{
		Enabled:   et.config.Enabled,
		Timestamp: time.Now(),
	}, nil
}

// RunTests executes E2E tests
func (et *E2ETester) RunTests(ctx context.Context) (*models.TestResult, error) {
	ctx, span := et.tracer.Start(ctx, "e2eTester.RunTests")
	defer span.End()

	et.logger.Info("Running E2E tests")

	result := &models.TestResult{
		ID:        fmt.Sprintf("e2e-%d", time.Now().Unix()),
		Type:      "e2e",
		Name:      "End-to-End Tests",
		StartTime: time.Now(),
		Status:    "running",
	}

	// Execute E2E tests
	cmd := exec.CommandContext(ctx, "go", "test", "-tags=e2e", "./tests/e2e/...")
	output, err := cmd.CombinedOutput()

	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)
	result.Output = string(output)

	if err != nil {
		result.Status = "failed"
		result.TestsFailed = 1
	} else {
		result.Status = "passed"
		result.TestsPassed = 1
	}

	result.TestsRun = result.TestsPassed + result.TestsFailed

	et.logger.WithFields(logrus.Fields{
		"duration": result.Duration,
		"status":   result.Status,
	}).Info("E2E tests completed")

	return result, nil
}
