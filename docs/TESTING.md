# AIOS Testing Framework

This document describes the comprehensive testing framework implemented for AIOS, providing advanced testing capabilities including unit testing, integration testing, performance testing, contract testing, and property-based testing.

## Overview

The AIOS Testing Framework is a comprehensive testing solution that provides:

- **Enhanced Test Utilities**: Advanced test helpers, mocks, and fixtures
- **Coverage Analysis**: Detailed coverage reporting with threshold enforcement
- **Performance Testing**: Load testing, stress testing, and benchmarking
- **Contract Testing**: API contract validation
- **Property-Based Testing**: Generative testing for edge cases
- **Test Environment Management**: Automated test environment setup and teardown
- **Advanced Reporting**: Multi-format test reports with visualization

## Quick Start

### Running Tests

```bash
# Run all tests using the enhanced framework
make test-enhanced

# Run specific test suites
make test-unit          # Unit tests only
make test-integration   # Integration tests only
make test-e2e          # End-to-end tests only
make test-performance  # Performance tests only
make test-contract     # Contract tests only

# Run tests with different modes
make test-fast-enhanced    # Fast tests only
make test-slow-enhanced    # Slow tests only
make test-parallel         # Parallel execution
make test-dry-run         # Show what would run
```

### Using the Test Runner

```bash
# Direct test runner usage
go run scripts/test-runner.go -config=configs/testing.yaml -suite=all

# With specific options
go run scripts/test-runner.go \
  -config=configs/testing.yaml \
  -suite=unit \
  -parallel=true \
  -verbose=true \
  -timeout=5m
```

## Framework Components

### 1. Test Utilities (`internal/testing/utils`)

Enhanced test utilities providing common testing patterns:

```go
// Create a test helper
helper := utils.NewTestHelper(t)

// Create temporary files and directories
tempDir := helper.TempDir()
tempFile := helper.TempFile("test content")

// Assert conditions with timeout
helper.AssertEventually(func() bool {
    return someCondition()
}, time.Second*5, "condition should become true")

// Run concurrent tests
helper.ConcurrentTest(
    func() { /* test 1 */ },
    func() { /* test 2 */ },
    func() { /* test 3 */ },
)
```

### 2. Mock Generation (`internal/testing/mocks`)

Automated mock generation and management:

```go
// Create mock generator
mockGen := mocks.NewMockGenerator()

// Build mocks with fluent interface
mock := mocks.NewMockBuilder(mockObj).
    On("Method", arg1, arg2).
    Return(result).
    Times(1).
    Build()

// HTTP mocking
httpMock := mocks.NewHTTPMock()
httpMock.SetResponse("/api/test", &mocks.HTTPResponse{
    StatusCode: 200,
    Body:       []byte(`{"status": "ok"}`),
})
```

### 3. Test Fixtures (`internal/testing/fixtures`)

Comprehensive test data management:

```go
// Load fixtures
fixtureManager := fixtures.NewFixtureManager("testdata/fixtures")
var users []User
err := fixtureManager.LoadFixture("users", &users)

// Build test data
user := fixtures.NewTestDataBuilder().
    WithField("username", "testuser").
    WithRandomString("email", 10).
    WithTimestamp("created_at").
    Build(&User{})
```

### 4. Coverage Analysis (`internal/testing/coverage`)

Advanced coverage analysis with threshold enforcement:

```go
// Analyze coverage
analyzer := coverage.NewCoverageAnalyzer(".")
report, err := analyzer.AnalyzeCoverage()

// Check thresholds
exitCode, err := analyzer.ValidateCoverage()

// Generate reports
err = analyzer.GenerateHTMLReport("coverage.html")
err = analyzer.ExportReport(report, "json", "coverage.json")
```

### 5. Performance Testing (`internal/testing/performance`)

Comprehensive load and performance testing:

```go
// Configure load test
config := performance.LoadTestConfig{
    Duration:    time.Minute,
    Concurrency: 10,
    RequestRate: 100,
}

// Create load tester
loadTester := performance.NewLoadTester(config)

// Run load test
result, err := loadTester.RunLoadTest(ctx, func(ctx context.Context) error {
    // Your test function here
    return nil
})

// Run stress test
results, err := loadTester.StressTest(ctx, testFunc, 100)
```

### 6. Contract Testing (`internal/testing/contract`)

API contract validation:

```go
// Create contract tester
contractTester := contract.NewContractTester("http://localhost:8080")

// Define contract
contract := contract.Contract{
    Name:     "User Registration",
    Endpoint: "/api/v1/users/register",
    Method:   "POST",
    Request: &contract.ContractRequest{
        Schema: &contract.JSONSchema{
            Type: "object",
            Properties: map[string]*contract.JSONSchema{
                "username": {Type: "string"},
                "email":    {Type: "string"},
            },
            Required: []string{"username", "email"},
        },
    },
    Response: &contract.ContractResponse{
        StatusCode: 201,
        Schema: &contract.JSONSchema{
            Type: "object",
            Properties: map[string]*contract.JSONSchema{
                "id":       {Type: "string"},
                "username": {Type: "string"},
            },
        },
    },
}

// Test contract
result, err := contractTester.TestContract(ctx, contract)
```

### 7. Property-Based Testing (`internal/testing/property`)

Generative testing for edge cases:

```go
// Create property tester
config := property.PropertyConfig{
    MaxTests:   100,
    MaxShrinks: 100,
}
propertyTester := property.NewPropertyTester(config)

// Define property
prop := property.ForAll(
    []property.Generator{
        property.NewIntGenerator(0, 100),
        property.NewStringGenerator(1, 50),
    },
    func(args ...interface{}) bool {
        num := args[0].(int)
        str := args[1].(string)
        // Your property test here
        return len(str) > 0 && num >= 0
    },
)

// Test property
result := propertyTester.TestProperty(t, prop)
```

### 8. Test Environment Management (`internal/testing/environment`)

Automated test environment setup:

```go
// Configure environment
envConfig := environment.EnvironmentConfig{
    Name: "test-env",
    Services: []environment.ServiceConfig{
        {
            Name:  "postgres",
            Image: "postgres:15-alpine",
            Ports: []environment.PortMapping{
                {Host: 5433, Container: 5432},
            },
        },
    },
}

// Create and start environment
envManager := environment.NewEnvironmentManager()
env, err := envManager.CreateEnvironment(envConfig)
err = env.Start(ctx)

// Cleanup
defer env.Stop(ctx)
```

## Configuration

### Testing Configuration (`configs/testing.yaml`)

The main configuration file defines all testing parameters:

```yaml
# Project settings
project_root: "."
test_data_path: "testdata"
coverage_threshold: 80.0
parallel_execution: true

# Enabled features
enabled_features:
  - "coverage"
  - "contract"
  - "property"
  - "performance"

# Environment configuration
environment:
  name: "aios-test-env"
  services:
    - name: "test-database"
      image: "postgres:15-alpine"
      ports:
        - host: 5433
          container: 5432

# Coverage thresholds
coverage:
  overall: 80.0
  package: 75.0
  function: 70.0
```

## Test Environment

### Docker Compose Setup

The test environment uses Docker Compose for service orchestration:

```bash
# Start test environment
make test-env-setup

# Stop test environment
make test-env-teardown

# Reset test environment
make test-env-reset
```

Services included:
- PostgreSQL (test database)
- Redis (caching)
- RabbitMQ (message queue)
- Elasticsearch (search)
- MinIO (object storage)
- Jaeger (tracing)
- Prometheus (metrics)

### Environment Variables

```bash
AIOS_ENV=test
AIOS_LOG_LEVEL=error
AIOS_TEST_MODE=true
DATABASE_URL=postgres://test_user:test_pass@localhost:5433/aios_test
REDIS_URL=redis://localhost:6380
```

## Test Data Management

### Fixtures

Test fixtures are stored in `testdata/fixtures/`:

- `users.json` - Test user data
- `ai_models.yaml` - AI model configurations
- `contracts/` - API contract definitions

### Generating Test Data

```go
// Using faker
generator := fixtures.NewTestDataGenerator()
email := generator.RandomEmail()
name := generator.RandomString(10)

// Using fixtures
var users []User
fixtureManager.LoadFixture("users", &users)
```

## Reporting

### Test Reports

The framework generates multiple report formats:

- **HTML**: Interactive web reports
- **JSON**: Machine-readable results
- **XML**: CI/CD integration
- **JUnit**: Jenkins/CI compatibility

### Coverage Reports

- Line coverage
- Function coverage
- Branch coverage
- Statement coverage
- Trend analysis

### Performance Reports

- Request/response metrics
- Latency percentiles
- Throughput analysis
- Resource utilization

## CI/CD Integration

### GitHub Actions

```yaml
- name: Run Tests
  run: make test-enhanced

- name: Upload Coverage
  uses: codecov/codecov-action@v3
  with:
    file: ./coverage.out

- name: Upload Test Reports
  uses: actions/upload-artifact@v3
  with:
    name: test-reports
    path: test-reports/
```

### Make Targets

```bash
make ci-test        # CI-optimized test run
make ci-coverage    # Coverage for CI
make ci-lint        # Linting for CI
```

## Best Practices

### Writing Tests

1. **Use descriptive test names**
2. **Follow AAA pattern** (Arrange, Act, Assert)
3. **Use test helpers** for common operations
4. **Mock external dependencies**
5. **Test edge cases** with property-based testing

### Test Organization

1. **Group tests by feature**
2. **Use build tags** for different test types
3. **Separate unit and integration tests**
4. **Use fixtures** for test data

### Performance Testing

1. **Start with baseline measurements**
2. **Test realistic scenarios**
3. **Monitor resource usage**
4. **Set appropriate thresholds**

## Troubleshooting

### Common Issues

1. **Test environment not starting**
   - Check Docker is running
   - Verify port availability
   - Check service health

2. **Coverage below threshold**
   - Add missing tests
   - Review coverage report
   - Exclude non-testable code

3. **Flaky tests**
   - Use proper synchronization
   - Avoid time-dependent tests
   - Use test isolation

### Debug Mode

```bash
# Run tests with debugging
make test-debug

# Verbose output
go run scripts/test-runner.go -verbose=true

# Dry run to see what would execute
make test-dry-run
```

## Contributing

When adding new tests:

1. Follow the existing patterns
2. Add appropriate fixtures
3. Update documentation
4. Ensure tests are deterministic
5. Add performance benchmarks for critical paths

For more information, see the [Contributing Guide](CONTRIBUTING.md).
