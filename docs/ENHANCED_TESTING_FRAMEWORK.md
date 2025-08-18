# AIOS Enhanced Testing Framework

## Overview

The AIOS Enhanced Testing Framework provides comprehensive testing capabilities for the entire AIOS ecosystem. Built with enterprise-grade features, it supports unit testing, integration testing, performance testing, security testing, and end-to-end testing with advanced reporting, metrics, and CI/CD integration.

## 🏗️ Architecture

### Core Components

```
AIOS Enhanced Testing Framework
├── Test Framework Core (test management and execution)
├── Test Suite Management (organized test collections)
├── Test Case Management (individual test definitions)
├── Test Data Management (test data and fixtures)
├── Mock & Stub Services (service virtualization)
├── Performance Testing (load, stress, spike testing)
├── Integration Testing (system and contract testing)
├── Security Testing (vulnerability and compliance testing)
├── Test Execution Engine (parallel and distributed execution)
├── Test Reporting (comprehensive reports and analytics)
├── Test Metrics & Analytics (performance and trend analysis)
└── CI/CD Integration (pipeline integration and automation)
```

### Key Features

- **🧪 Comprehensive Testing**: Unit, integration, performance, security, and E2E testing
- **🚀 Parallel Execution**: Concurrent test execution with configurable concurrency
- **📊 Advanced Reporting**: Detailed reports with metrics, trends, and analytics
- **🎭 Service Virtualization**: Mock services and stubs for isolated testing
- **⚡ Performance Testing**: Load, stress, and spike testing capabilities
- **🔒 Security Testing**: Vulnerability scanning and compliance validation
- **📈 Test Analytics**: Metrics, trends, and performance analysis
- **🔄 CI/CD Integration**: Seamless integration with deployment pipelines
- **🎯 Test Data Management**: Comprehensive test data and fixture management

## 🚀 Quick Start

### Basic Test Framework Setup

```go
package main

import (
    "time"
    "github.com/aios/aios/pkg/testing"
    "github.com/sirupsen/logrus"
)

func main() {
    logger := logrus.New()
    
    // Create test framework
    config := &testing.TestFrameworkConfig{
        MaxTestSuites:   1000,
        MaxTestCases:    10000,
        MaxExecutions:   100000,
        DefaultTimeout:  30 * time.Minute,
        MaxConcurrency:  10,
        RetentionPeriod: 30 * 24 * time.Hour,
        EnableMetrics:   true,
        EnableReporting: true,
        ArtifactStorage: "./test-artifacts",
    }
    
    framework := testing.NewDefaultTestFramework(config, logger)
    
    // Create test suite
    suite := &testing.TestSuite{
        Name:        "AIOS Unit Tests",
        Description: "Comprehensive unit tests for AIOS components",
        Category:    testing.TestCategoryUnit,
        Priority:    testing.TestPriorityHigh,
        Tags:        []string{"unit", "core", "fast"},
        Config: &testing.TestSuiteConfig{
            Parallel:        true,
            MaxConcurrency:  5,
            Timeout:         10 * time.Minute,
            FailFast:        false,
            ContinueOnError: true,
            Environment:     "test",
        },
    }
    
    createdSuite, err := framework.CreateTestSuite(suite)
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("Test suite created: %s\n", createdSuite.ID)
}
```

## 🧪 Test Suite Management

### Creating Test Suites

Organize tests into logical suites with comprehensive configuration:

```go
// Unit Test Suite
unitTestSuite := &testing.TestSuite{
    Name:        "AIOS Unit Tests",
    Description: "Comprehensive unit tests for AIOS components",
    Category:    testing.TestCategoryUnit,
    Priority:    testing.TestPriorityHigh,
    Tags:        []string{"unit", "core", "fast"},
    Config: &testing.TestSuiteConfig{
        Parallel:        true,
        MaxConcurrency:  5,
        Timeout:         10 * time.Minute,
        FailFast:        false,
        ContinueOnError: true,
        Environment:     "test",
        RetryPolicy: &testing.RetryPolicy{
            MaxRetries:    2,
            InitialDelay:  1 * time.Second,
            MaxDelay:      10 * time.Second,
            BackoffFactor: 2.0,
        },
    },
    Setup: &testing.TestSetup{
        Steps: []*testing.TestStep{
            {
                Name:    "Initialize Test Environment",
                Type:    testing.TestStepTypeSetup,
                Action:  "initialize_environment",
                Timeout: 30 * time.Second,
            },
        },
        Timeout: 60 * time.Second,
    },
    Teardown: &testing.TestTeardown{
        Steps: []*testing.TestStep{
            {
                Name:    "Cleanup Test Environment",
                Type:    testing.TestStepTypeTeardown,
                Action:  "cleanup_environment",
                Timeout: 30 * time.Second,
            },
        },
        Timeout:   60 * time.Second,
        AlwaysRun: true,
    },
    CreatedBy: "test-engineer",
}

createdSuite, err := framework.CreateTestSuite(unitTestSuite)
```

### Managing Test Suites

```go
// List test suites with filtering
suites, err := framework.ListTestSuites(&testing.TestSuiteFilter{
    Category: testing.TestCategoryUnit,
    Priority: testing.TestPriorityHigh,
    Tags:     []string{"core"},
    Search:   "AIOS",
    Limit:    10,
})

// Get specific test suite
suite, err := framework.GetTestSuite(suiteID)

// Delete test suite
err = framework.DeleteTestSuite(suiteID)
```

## 📝 Test Case Management

### Creating Test Cases

Define comprehensive test cases with steps, assertions, and configurations:

```go
// Unit Test Case - Agent Service
agentUnitTestCase := &testing.TestCase{
    Name:        "Agent Service Unit Test",
    Description: "Test agent creation, configuration, and basic operations",
    Type:        testing.TestTypeFunctional,
    Priority:    testing.TestPriorityHigh,
    Tags:        []string{"agent", "unit", "core"},
    Steps: []*testing.TestStep{
        {
            ID:          "setup-agent",
            Name:        "Setup Agent Environment",
            Description: "Initialize agent testing environment",
            Type:        testing.TestStepTypeSetup,
            Action:      "initialize_agent_environment",
            Input: map[string]interface{}{
                "config": map[string]interface{}{
                    "model":        "gpt-4",
                    "max_tokens":   2000,
                    "temperature":  0.7,
                },
            },
            Timeout:   30 * time.Second,
            Order:     1,
            Enabled:   true,
            OnFailure: testing.TestStepFailureActionStop,
        },
        {
            ID:          "create-agent",
            Name:        "Create Agent",
            Description: "Create a new agent instance",
            Type:        testing.TestStepTypeAction,
            Action:      "create_agent",
            Input: map[string]interface{}{
                "name":         "Test Agent",
                "description":  "Agent for unit testing",
                "capabilities": []string{"chat", "analysis"},
            },
            Expected: map[string]interface{}{
                "status": "created",
                "id":     "not_empty",
            },
            Timeout:   10 * time.Second,
            Order:     2,
            Enabled:   true,
            OnFailure: testing.TestStepFailureActionStop,
        },
    },
    Assertions: []*testing.TestAssertion{
        {
            ID:       "agent-created",
            Name:     "Agent Created Successfully",
            Type:     testing.TestAssertionTypeExists,
            Field:    "agent.id",
            Operator: testing.TestAssertionOperatorNE,
            Expected: "",
            Message:  "Agent should be created with valid ID",
            Critical: true,
        },
        {
            ID:       "agent-status",
            Name:     "Agent Status Active",
            Type:     testing.TestAssertionTypeEquals,
            Field:    "agent.status",
            Operator: testing.TestAssertionOperatorEQ,
            Expected: "active",
            Message:  "Agent should be in active status",
            Critical: true,
        },
    ],
    Timeout: 2 * time.Minute,
    Retries: 2,
}

createdTestCase, err := framework.AddTestCase(suiteID, agentUnitTestCase)
```

## 🎭 Mock Services and Test Data

### Creating Mock Services

Set up service virtualization for isolated testing:

```go
// HTTP Mock Service
mockService := &testing.MockService{
    Name: "Mock External API",
    Type: testing.MockServiceTypeHTTP,
    Endpoints: []*testing.MockEndpoint{
        {
            ID:     "get-user",
            Path:   "/api/users/{id}",
            Method: "GET",
            Response: &testing.MockResponse{
                StatusCode: 200,
                Headers: map[string]string{
                    "Content-Type": "application/json",
                },
                Body: map[string]interface{}{
                    "id":    "{id}",
                    "name":  "Test User",
                    "email": "test@example.com",
                    "status": "active",
                },
            },
            Delay: 100 * time.Millisecond,
        },
    },
    Behaviors: []*testing.MockBehavior{
        {
            ID:        "rate-limit",
            Name:      "Rate Limiting",
            Condition: "request_count > 100",
            Action:    "return_429",
            Parameters: map[string]interface{}{
                "status_code": 429,
                "message":     "Rate limit exceeded",
            },
            Enabled: true,
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
```

### Managing Test Data

Create and manage comprehensive test data:

```go
// Static Test Data
testData := &testing.TestData{
    Name:   "Agent Test Data",
    Type:   testing.TestDataTypeStatic,
    Source: testing.TestDataSourceInline,
    Data: map[string]interface{}{
        "test_agents": []map[string]interface{}{
            {
                "name":         "Customer Support Agent",
                "model":        "gpt-4",
                "capabilities": []string{"chat", "email", "analysis"},
                "config": map[string]interface{}{
                    "max_tokens":   2000,
                    "temperature":  0.7,
                    "top_p":        0.9,
                },
            },
        },
        "test_scenarios": []map[string]interface{}{
            {
                "name":        "Customer Inquiry",
                "description": "Handle customer support inquiry",
                "input":       "I need help with my account",
                "expected":    "helpful_response",
            },
        },
    },
    Schema: map[string]interface{}{
        "type": "object",
        "properties": map[string]interface{}{
            "test_agents": map[string]interface{}{
                "type": "array",
                "items": map[string]interface{}{
                    "type": "object",
                    "required": []string{"name", "model", "capabilities"},
                },
            },
        },
    },
}

createdTestData, err := framework.CreateTestData(testData)
```

## ⚡ Performance Testing

### Load Testing

Configure and execute comprehensive load tests:

```go
// Load Test Configuration
loadTestConfig := &testing.LoadTestConfig{
    Name:        "AIOS API Load Test",
    Description: "Load test for AIOS API endpoints",
    Target: &testing.TestTarget{
        Type:     testing.TestTargetTypeHTTP,
        URL:      "http://localhost:8080",
        Protocol: "http",
        Timeout:  30 * time.Second,
    },
    LoadPattern:     testing.LoadPatternRampUp,
    Duration:        10 * time.Minute,
    MaxUsers:        500,
    RampUpTime:      2 * time.Minute,
    SustainTime:     6 * time.Minute,
    RampDownTime:    2 * time.Minute,
    RequestsPerSec:  100,
    Scenarios: []*testing.LoadScenario{
        {
            ID:          "api-load-scenario",
            Name:        "API Load Scenario",
            Description: "Load test API endpoints",
            Weight:      0.7, // 70% of traffic
            Steps: []*testing.LoadStep{
                {
                    ID:   "get-agents",
                    Name: "Get Agents",
                    Type: testing.LoadStepTypeHTTP,
                    Request: &testing.TestRequest{
                        Method: "GET",
                        URL:    "/api/v1/agents",
                        Headers: map[string]string{
                            "Accept": "application/json",
                        },
                        Timeout: 10 * time.Second,
                    },
                    Validations: []*testing.StepValidation{
                        {
                            Type:     testing.StepValidationTypeStatusCode,
                            Operator: testing.TestAssertionOperatorEQ,
                            Expected: 200,
                            Critical: true,
                        },
                    },
                    ThinkTime: 1 * time.Second,
                    Order:     1,
                    Enabled:   true,
                },
            },
            Enabled: true,
        },
    },
    Thresholds: &testing.LoadTestThresholds{
        ResponseTime: &testing.ThresholdConfig{
            Max:      1000.0, // 1 second
            Critical: true,
            Unit:     "ms",
        },
        Throughput: &testing.ThresholdConfig{
            Min:      80.0, // 80 req/sec
            Critical: false,
            Unit:     "req/sec",
        },
        ErrorRate: &testing.ThresholdConfig{
            Max:      0.05, // 5%
            Critical: true,
            Unit:     "%",
        },
    },
    Environment: "performance",
}

loadTestResult, err := framework.RunLoadTest(loadTestConfig)
```

## 🔗 Integration Testing

### System Integration Testing

Test integration between AIOS components:

```go
// Integration Test Configuration
integrationTestConfig := &testing.IntegrationTestConfig{
    Name:        "AIOS System Integration Test",
    Description: "Comprehensive integration test for AIOS components",
    Services: []*testing.IntegrationService{
        {
            ID:     "agent-service",
            Name:   "Agent Service",
            Type:   "microservice",
            URL:    "http://agent-service:8080",
            Health: "/health",
            Dependencies: []string{"database", "cache"},
        },
        {
            ID:     "collaboration-service",
            Name:   "Collaboration Service",
            Type:   "microservice",
            URL:    "http://collaboration-service:8081",
            Health: "/health",
            Dependencies: []string{"agent-service", "websocket"},
        },
    ],
    Scenarios: []*testing.IntegrationScenario{
        {
            ID:          "agent-collaboration-flow",
            Name:        "Agent-Collaboration Flow",
            Description: "Test agent and collaboration service integration",
            Services:    []string{"agent-service", "collaboration-service"},
            Steps: []*testing.IntegrationStep{
                {
                    ID:        "create-agent",
                    Name:      "Create Agent",
                    Type:      testing.IntegrationStepTypeHTTP,
                    ServiceID: "agent-service",
                    Request: &testing.TestRequest{
                        Method: "POST",
                        URL:    "/api/v1/agents",
                        Body: map[string]interface{}{
                            "name":         "Integration Test Agent",
                            "capabilities": []string{"chat", "collaboration"},
                        },
                    },
                    Timeout: 30 * time.Second,
                    Order:   1,
                    Enabled: true,
                },
            ],
            Enabled: true,
        },
    },
    Environment: "integration",
}

integrationTestResult, err := framework.RunIntegrationTest(integrationTestConfig)
```

## 🧪 Testing

Run the comprehensive test suite:

```bash
# Run all tests
go test ./pkg/testing/...

# Run with race detection
go test -race ./pkg/testing/...

# Run integration tests
go test -tags=integration ./pkg/testing/...

# Run comprehensive testing example
go run examples/comprehensive_testing_example.go
```

## 📖 Examples

See the complete example in `examples/comprehensive_testing_example.go` for a comprehensive demonstration including:

- Test framework setup and configuration
- Test suite and case management
- Mock services and test data
- Performance and load testing
- Integration and system testing
- Test execution and reporting
- Metrics and analytics

## 🤝 Contributing

1. Follow established testing patterns and interfaces
2. Add comprehensive tests for new testing features
3. Update documentation for API changes
4. Ensure proper error handling and logging
5. Use OpenTelemetry for observability
6. Implement proper test isolation and cleanup

## 📄 License

This Enhanced Testing Framework is part of the AIOS project and follows the same licensing terms.
