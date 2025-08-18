package main

import (
	"fmt"
	"log"
	"time"

	"github.com/aios/aios/pkg/testing"
	"github.com/sirupsen/logrus"
)

func main() {
	// Initialize logger
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)
	logger.SetFormatter(&logrus.TextFormatter{
		DisableTimestamp: true,
		ForceColors:      true,
	})

	fmt.Println("🧪 AIOS Testing Framework Demo")
	fmt.Println("==============================")

	// Run the testing demo
	if err := runTestingDemo(logger); err != nil {
		log.Fatalf("Demo failed: %v", err)
	}

	fmt.Println("\n✅ Testing Framework Demo completed successfully!")
}

func runTestingDemo(logger *logrus.Logger) error {
	// Step 1: Create Test Framework
	fmt.Println("\n1. Creating Test Framework...")
	config := &testing.TestFrameworkConfig{
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

	framework := testing.NewDefaultTestFramework(config, logger)
	fmt.Println("✓ Test Framework created successfully")

	// Step 2: Create Test Suite
	fmt.Println("\n2. Creating Test Suite...")
	
	testSuite := &testing.TestSuite{
		Name:        "AIOS Core Tests",
		Description: "Core functionality tests for AIOS",
		Category:    testing.TestCategoryUnit,
		Priority:    testing.TestPriorityHigh,
		Tags:        []string{"core", "unit", "fast"},
		Config: &testing.TestSuiteConfig{
			Parallel:        true,
			MaxConcurrency:  3,
			Timeout:         5 * time.Minute,
			FailFast:        false,
			ContinueOnError: true,
			Environment:     "test",
		},
		CreatedBy: "demo-user",
	}

	createdSuite, err := framework.CreateTestSuite(testSuite)
	if err != nil {
		return fmt.Errorf("failed to create test suite: %w", err)
	}

	fmt.Printf("   ✓ Test Suite created: %s (ID: %s)\n", 
		createdSuite.Name, createdSuite.ID)

	// Step 3: Create Test Case
	fmt.Println("\n3. Creating Test Case...")
	
	testCase := &testing.TestCase{
		Name:        "Basic Functionality Test",
		Description: "Test basic AIOS functionality",
		Type:        testing.TestTypeFunctional,
		Priority:    testing.TestPriorityHigh,
		Tags:        []string{"basic", "functionality"},
		Steps: []*testing.TestStep{
			{
				ID:          "step-1",
				Name:        "Initialize System",
				Description: "Initialize the AIOS system",
				Type:        testing.TestStepTypeSetup,
				Action:      "initialize_system",
				Timeout:     30 * time.Second,
				Order:       1,
				Enabled:     true,
				OnFailure:   testing.TestStepFailureActionStop,
			},
			{
				ID:          "step-2",
				Name:        "Test Core Function",
				Description: "Test core system function",
				Type:        testing.TestStepTypeAction,
				Action:      "test_core_function",
				Timeout:     10 * time.Second,
				Order:       2,
				Enabled:     true,
				OnFailure:   testing.TestStepFailureActionStop,
			},
		},
		Assertions: []*testing.TestAssertion{
			{
				ID:       "assertion-1",
				Name:     "System Initialized",
				Type:     testing.TestAssertionTypeExists,
				Field:    "system.status",
				Operator: testing.TestAssertionOperatorEQ,
				Expected: "initialized",
				Message:  "System should be initialized",
				Critical: true,
			},
		},
		Timeout: 2 * time.Minute,
		Retries: 1,
	}

	createdTestCase, err := framework.AddTestCase(createdSuite.ID, testCase)
	if err != nil {
		return fmt.Errorf("failed to create test case: %w", err)
	}

	fmt.Printf("   ✓ Test Case created: %s (ID: %s)\n", 
		createdTestCase.Name, createdTestCase.ID)

	// Step 4: Create Test Data
	fmt.Println("\n4. Creating Test Data...")
	
	testData := &testing.TestData{
		Name:   "Demo Test Data",
		Type:   testing.TestDataTypeStatic,
		Source: testing.TestDataSourceInline,
		Data: map[string]interface{}{
			"test_config": map[string]interface{}{
				"timeout":     30,
				"retry_count": 3,
				"debug_mode":  true,
			},
			"test_users": []map[string]interface{}{
				{
					"id":   1,
					"name": "Test User 1",
					"role": "admin",
				},
				{
					"id":   2,
					"name": "Test User 2",
					"role": "user",
				},
			},
		},
	}

	createdTestData, err := framework.CreateTestData(testData)
	if err != nil {
		return fmt.Errorf("failed to create test data: %w", err)
	}

	fmt.Printf("   ✓ Test Data created: %s (ID: %s)\n", 
		createdTestData.Name, createdTestData.ID)

	// Step 5: Create Mock Service
	fmt.Println("\n5. Creating Mock Service...")
	
	mockService := &testing.MockService{
		Name: "Demo Mock API",
		Type: testing.MockServiceTypeHTTP,
		Endpoints: []*testing.MockEndpoint{
			{
				ID:     "health-check",
				Path:   "/health",
				Method: "GET",
				Response: &testing.MockResponse{
					StatusCode: 200,
					Headers: map[string]string{
						"Content-Type": "application/json",
					},
					Body: map[string]interface{}{
						"status":  "healthy",
						"version": "1.0.0",
						"uptime":  "24h",
					},
				},
				Delay: 50 * time.Millisecond,
			},
		},
		Config: &testing.MockServiceConfig{
			Port:        8080,
			Host:        "localhost",
			TLS:         false,
			Timeout:     30 * time.Second,
			Persistence: false,
			Logging:     true,
		},
	}

	createdMockService, err := framework.CreateMock(mockService)
	if err != nil {
		return fmt.Errorf("failed to create mock service: %w", err)
	}

	fmt.Printf("   ✓ Mock Service created: %s (ID: %s)\n", 
		createdMockService.Name, createdMockService.ID)

	// Step 6: Execute Test Suite
	fmt.Println("\n6. Executing Test Suite...")
	
	executionConfig := &testing.TestExecutionConfig{
		Environment:     "test",
		Parallel:        true,
		MaxConcurrency:  3,
		Timeout:         10 * time.Minute,
		FailFast:        false,
		ContinueOnError: true,
		Tags:            []string{"core"},
		Variables: map[string]interface{}{
			"test_environment": "demo",
			"debug_enabled":    true,
		},
	}

	execution, err := framework.RunTestSuite(createdSuite.ID, executionConfig)
	if err != nil {
		return fmt.Errorf("failed to start test execution: %w", err)
	}

	fmt.Printf("   ✓ Test Suite execution started: %s\n", execution.ID)

	// Wait for execution to complete
	time.Sleep(3 * time.Second)

	// Get execution status
	executionStatus, err := framework.GetTestExecution(execution.ID)
	if err != nil {
		return fmt.Errorf("failed to get execution status: %w", err)
	}

	fmt.Printf("   ✓ Test Suite Status: %s\n", executionStatus.Status)
	if executionStatus.Summary != nil {
		fmt.Printf("     - Total Tests: %d\n", executionStatus.Summary.TotalTests)
		fmt.Printf("     - Passed: %d\n", executionStatus.Summary.PassedTests)
		fmt.Printf("     - Failed: %d\n", executionStatus.Summary.FailedTests)
		fmt.Printf("     - Success Rate: %.1f%%\n", executionStatus.Summary.SuccessRate*100)
		fmt.Printf("     - Duration: %s\n", executionStatus.Summary.Duration)
	}

	// Step 7: Performance Testing
	fmt.Println("\n7. Running Performance Test...")
	
	loadTestConfig := &testing.LoadTestConfig{
		Name:        "Demo Load Test",
		Description: "Simple load test demonstration",
		Target: &testing.TestTarget{
			Type:     testing.TestTargetTypeHTTP,
			URL:      "http://localhost:8080",
			Protocol: "http",
			Timeout:  10 * time.Second,
		},
		LoadPattern:     testing.LoadPatternConstant,
		Duration:        30 * time.Second,
		MaxUsers:        10,
		RequestsPerSec:  5,
		Environment:     "test",
	}

	loadTestResult, err := framework.RunLoadTest(loadTestConfig)
	if err != nil {
		return fmt.Errorf("failed to run load test: %w", err)
	}

	fmt.Printf("   ✓ Load Test completed: %s\n", loadTestResult.ID)
	fmt.Printf("     - Status: %s\n", loadTestResult.Status)
	fmt.Printf("     - Duration: %s\n", loadTestResult.Duration)
	if loadTestResult.Summary != nil {
		fmt.Printf("     - Total Requests: %d\n", loadTestResult.Summary.TotalRequests)
		fmt.Printf("     - Success Rate: %.1f%%\n", (1.0-loadTestResult.Summary.ErrorRate)*100)
		fmt.Printf("     - Average Response Time: %s\n", loadTestResult.Summary.AverageResponseTime)
		fmt.Printf("     - Throughput: %.1f req/sec\n", loadTestResult.Summary.Throughput)
	}

	// Step 8: Generate Test Report
	fmt.Println("\n8. Generating Test Report...")
	
	testReport, err := framework.GenerateTestReport(execution.ID)
	if err != nil {
		return fmt.Errorf("failed to generate test report: %w", err)
	}

	fmt.Printf("   ✓ Test Report generated: %s\n", testReport.ID)
	fmt.Printf("     - Title: %s\n", testReport.Title)
	fmt.Printf("     - Environment: %s\n", testReport.Environment)
	fmt.Printf("     - Generated At: %s\n", testReport.GeneratedAt.Format("2006-01-02 15:04:05"))

	// Step 9: List Test Suites
	fmt.Println("\n9. Listing Test Suites...")
	
	suites, err := framework.ListTestSuites(&testing.TestSuiteFilter{
		Category: testing.TestCategoryUnit,
		Limit:    5,
	})
	if err != nil {
		return fmt.Errorf("failed to list test suites: %w", err)
	}

	fmt.Printf("   ✓ Test Suites found: %d\n", len(suites))
	for i, suite := range suites {
		fmt.Printf("     %d. %s (%s) - %s\n", 
			i+1, suite.Name, suite.Category, suite.Priority)
	}

	return nil
}
