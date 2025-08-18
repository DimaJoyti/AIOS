package testing

import (
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDefaultTestFramework(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel) // Reduce noise in tests

	config := &TestFrameworkConfig{
		MaxTestSuites:   100,
		MaxTestCases:    1000,
		MaxExecutions:   10000,
		DefaultTimeout:  30 * time.Minute,
		MaxConcurrency:  5,
		RetentionPeriod: 7 * 24 * time.Hour,
		EnableMetrics:   true,
		EnableReporting: true,
		ArtifactStorage: "./test-artifacts",
	}

	framework := NewDefaultTestFramework(config, logger)
	require.NotNil(t, framework)

	// Verify framework is properly initialized by testing basic functionality
	// We can't access internal fields, so we test the interface works
}

func TestCreateTestSuite(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	config := &TestFrameworkConfig{
		MaxTestSuites:   100,
		MaxTestCases:    1000,
		MaxExecutions:   10000,
		DefaultTimeout:  30 * time.Minute,
		MaxConcurrency:  5,
		RetentionPeriod: 7 * 24 * time.Hour,
		EnableMetrics:   true,
		EnableReporting: true,
		ArtifactStorage: "./test-artifacts",
	}

	framework := NewDefaultTestFramework(config, logger)

	testSuite := &TestSuite{
		Name:        "Test Suite",
		Description: "Test suite for testing",
		Category:    TestCategoryUnit,
		Priority:    TestPriorityHigh,
		Tags:        []string{"test", "unit"},
		Config: &TestSuiteConfig{
			Parallel:        true,
			MaxConcurrency:  3,
			Timeout:         5 * time.Minute,
			FailFast:        false,
			ContinueOnError: true,
			Environment:     "test",
		},
		CreatedBy: "test-user",
	}

	createdSuite, err := framework.CreateTestSuite(testSuite)
	require.NoError(t, err)
	require.NotNil(t, createdSuite)

	// Verify suite was created with proper values
	assert.NotEmpty(t, createdSuite.ID)
	assert.Equal(t, "Test Suite", createdSuite.Name)
	assert.Equal(t, "Test suite for testing", createdSuite.Description)
	assert.Equal(t, TestCategoryUnit, createdSuite.Category)
	assert.Equal(t, TestPriorityHigh, createdSuite.Priority)
	assert.Equal(t, []string{"test", "unit"}, createdSuite.Tags)
	assert.Equal(t, "test-user", createdSuite.CreatedBy)
	assert.False(t, createdSuite.CreatedAt.IsZero())
	assert.False(t, createdSuite.UpdatedAt.IsZero())
}

func TestAddTestCase(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	config := &TestFrameworkConfig{
		MaxTestSuites:   100,
		MaxTestCases:    1000,
		MaxExecutions:   10000,
		DefaultTimeout:  30 * time.Minute,
		MaxConcurrency:  5,
		RetentionPeriod: 7 * 24 * time.Hour,
		EnableMetrics:   true,
		EnableReporting: true,
		ArtifactStorage: "./test-artifacts",
	}

	framework := NewDefaultTestFramework(config, logger)

	// First create a test suite
	testSuite := &TestSuite{
		Name:        "Test Suite",
		Description: "Test suite for testing",
		Category:    TestCategoryUnit,
		Priority:    TestPriorityHigh,
		Tags:        []string{"test", "unit"},
		Config: &TestSuiteConfig{
			Parallel:        true,
			MaxConcurrency:  3,
			Timeout:         5 * time.Minute,
			FailFast:        false,
			ContinueOnError: true,
			Environment:     "test",
		},
		CreatedBy: "test-user",
	}

	createdSuite, err := framework.CreateTestSuite(testSuite)
	require.NoError(t, err)

	// Now add a test case
	testCase := &TestCase{
		Name:        "Test Case",
		Description: "Test case for testing",
		Type:        TestTypeFunctional,
		Priority:    TestPriorityHigh,
		Tags:        []string{"test", "functional"},
		Steps: []*TestStep{
			{
				ID:          "step-1",
				Name:        "Test Step",
				Description: "Test step for testing",
				Type:        TestStepTypeAction,
				Action:      "test_action",
				Timeout:     30 * time.Second,
				Order:       1,
				Enabled:     true,
				OnFailure:   TestStepFailureActionStop,
			},
		},
		Assertions: []*TestAssertion{
			{
				ID:       "assertion-1",
				Name:     "Test Assertion",
				Type:     TestAssertionTypeExists,
				Field:    "test.field",
				Operator: TestAssertionOperatorEQ,
				Expected: "test_value",
				Message:  "Test assertion message",
				Critical: true,
			},
		},
		Timeout: 2 * time.Minute,
		Retries: 1,
	}

	createdTestCase, err := framework.AddTestCase(createdSuite.ID, testCase)
	require.NoError(t, err)
	require.NotNil(t, createdTestCase)

	// Verify test case was created with proper values
	assert.NotEmpty(t, createdTestCase.ID)
	assert.Equal(t, "Test Case", createdTestCase.Name)
	assert.Equal(t, "Test case for testing", createdTestCase.Description)
	assert.Equal(t, TestTypeFunctional, createdTestCase.Type)
	assert.Equal(t, TestPriorityHigh, createdTestCase.Priority)
	assert.Equal(t, []string{"test", "functional"}, createdTestCase.Tags)
	assert.Len(t, createdTestCase.Steps, 1)
	assert.Len(t, createdTestCase.Assertions, 1)
	assert.Equal(t, 2*time.Minute, createdTestCase.Timeout)
	assert.Equal(t, 1, createdTestCase.Retries)
	assert.False(t, createdTestCase.CreatedAt.IsZero())
	assert.False(t, createdTestCase.UpdatedAt.IsZero())
}

func TestCreateTestData(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	config := &TestFrameworkConfig{
		MaxTestSuites:   100,
		MaxTestCases:    1000,
		MaxExecutions:   10000,
		DefaultTimeout:  30 * time.Minute,
		MaxConcurrency:  5,
		RetentionPeriod: 7 * 24 * time.Hour,
		EnableMetrics:   true,
		EnableReporting: true,
		ArtifactStorage: "./test-artifacts",
	}

	framework := NewDefaultTestFramework(config, logger)

	testData := &TestData{
		Name:   "Test Data",
		Type:   TestDataTypeStatic,
		Source: TestDataSourceInline,
		Data: map[string]interface{}{
			"test_key":    "test_value",
			"test_number": 42,
			"test_array":  []string{"item1", "item2"},
		},
	}

	createdTestData, err := framework.CreateTestData(testData)
	require.NoError(t, err)
	require.NotNil(t, createdTestData)

	// Verify test data was created with proper values
	assert.NotEmpty(t, createdTestData.ID)
	assert.Equal(t, "Test Data", createdTestData.Name)
	assert.Equal(t, TestDataTypeStatic, createdTestData.Type)
	assert.Equal(t, TestDataSourceInline, createdTestData.Source)
	assert.NotNil(t, createdTestData.Data)
	assert.False(t, createdTestData.CreatedAt.IsZero())
	assert.False(t, createdTestData.UpdatedAt.IsZero())
}

func TestCreateMockService(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	config := &TestFrameworkConfig{
		MaxTestSuites:   100,
		MaxTestCases:    1000,
		MaxExecutions:   10000,
		DefaultTimeout:  30 * time.Minute,
		MaxConcurrency:  5,
		RetentionPeriod: 7 * 24 * time.Hour,
		EnableMetrics:   true,
		EnableReporting: true,
		ArtifactStorage: "./test-artifacts",
	}

	framework := NewDefaultTestFramework(config, logger)

	mockService := &MockService{
		Name: "Test Mock Service",
		Type: MockServiceTypeHTTP,
		Endpoints: []*MockEndpoint{
			{
				ID:     "test-endpoint",
				Path:   "/test",
				Method: "GET",
				Response: &MockResponse{
					StatusCode: 200,
					Headers: map[string]string{
						"Content-Type": "application/json",
					},
					Body: map[string]interface{}{
						"status": "ok",
					},
				},
				Delay: 100 * time.Millisecond,
			},
		},
		Config: &MockServiceConfig{
			Port:        8080,
			Host:        "localhost",
			TLS:         false,
			Timeout:     30 * time.Second,
			Persistence: false,
			Logging:     true,
		},
	}

	createdMockService, err := framework.CreateMock(mockService)
	require.NoError(t, err)
	require.NotNil(t, createdMockService)

	// Verify mock service was created with proper values
	assert.NotEmpty(t, createdMockService.ID)
	assert.Equal(t, "Test Mock Service", createdMockService.Name)
	assert.Equal(t, MockServiceTypeHTTP, createdMockService.Type)
	assert.Len(t, createdMockService.Endpoints, 1)
	assert.NotNil(t, createdMockService.Config)
	assert.False(t, createdMockService.CreatedAt.IsZero())
	assert.False(t, createdMockService.UpdatedAt.IsZero())
}
