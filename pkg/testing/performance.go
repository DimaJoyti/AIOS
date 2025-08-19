package testing

import (
	"time"
)

// Performance Testing Types and Configurations

// PerformanceTestConfig contains configuration for performance tests
type PerformanceTestConfig struct {
	Name         string                 `json:"name"`
	Description  string                 `json:"description"`
	Type         PerformanceTestType    `json:"type"`
	Target       *TestTarget            `json:"target"`
	Duration     time.Duration          `json:"duration"`
	Users        int                    `json:"users"`
	RampUpTime   time.Duration          `json:"ramp_up_time"`
	RampDownTime time.Duration          `json:"ramp_down_time"`
	ThinkTime    time.Duration          `json:"think_time"`
	Scenarios    []*PerformanceScenario `json:"scenarios"`
	Thresholds   *PerformanceThresholds `json:"thresholds"`
	Monitoring   *PerformanceMonitoring `json:"monitoring,omitempty"`
	Environment  string                 `json:"environment"`
	Tags         []string               `json:"tags,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// LoadTestConfig contains configuration for load tests
type LoadTestConfig struct {
	Name           string                 `json:"name"`
	Description    string                 `json:"description"`
	Target         *TestTarget            `json:"target"`
	LoadPattern    LoadPattern            `json:"load_pattern"`
	Duration       time.Duration          `json:"duration"`
	MaxUsers       int                    `json:"max_users"`
	RampUpTime     time.Duration          `json:"ramp_up_time"`
	SustainTime    time.Duration          `json:"sustain_time"`
	RampDownTime   time.Duration          `json:"ramp_down_time"`
	RequestsPerSec int                    `json:"requests_per_sec"`
	Scenarios      []*LoadScenario        `json:"scenarios"`
	Thresholds     *LoadTestThresholds    `json:"thresholds"`
	Monitoring     *PerformanceMonitoring `json:"monitoring,omitempty"`
	Environment    string                 `json:"environment"`
	Tags           []string               `json:"tags,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

// StressTestConfig contains configuration for stress tests
type StressTestConfig struct {
	Name          string                 `json:"name"`
	Description   string                 `json:"description"`
	Target        *TestTarget            `json:"target"`
	StressPattern StressPattern          `json:"stress_pattern"`
	Duration      time.Duration          `json:"duration"`
	StartUsers    int                    `json:"start_users"`
	MaxUsers      int                    `json:"max_users"`
	UserIncrement int                    `json:"user_increment"`
	IncrementTime time.Duration          `json:"increment_time"`
	Scenarios     []*StressScenario      `json:"scenarios"`
	Thresholds    *StressTestThresholds  `json:"thresholds"`
	Monitoring    *PerformanceMonitoring `json:"monitoring,omitempty"`
	Environment   string                 `json:"environment"`
	Tags          []string               `json:"tags,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// IntegrationTestConfig contains configuration for integration tests
type IntegrationTestConfig struct {
	Name         string                     `json:"name"`
	Description  string                     `json:"description"`
	Services     []*IntegrationService      `json:"services"`
	Scenarios    []*IntegrationScenario     `json:"scenarios"`
	Dependencies []*ServiceDependency       `json:"dependencies"`
	DataFlow     []*DataFlowTest            `json:"data_flow"`
	Contracts    []*ContractTest            `json:"contracts"`
	Thresholds   *IntegrationTestThresholds `json:"thresholds"`
	Environment  string                     `json:"environment"`
	Tags         []string                   `json:"tags,omitempty"`
	Metadata     map[string]interface{}     `json:"metadata,omitempty"`
}

// Test Target Types

// TestTarget represents a test target
type TestTarget struct {
	Type     TestTargetType         `json:"type"`
	URL      string                 `json:"url"`
	Host     string                 `json:"host"`
	Port     int                    `json:"port"`
	Protocol string                 `json:"protocol"`
	Path     string                 `json:"path"`
	Headers  map[string]string      `json:"headers,omitempty"`
	Auth     *TestAuth              `json:"auth,omitempty"`
	TLS      *TLSConfig             `json:"tls,omitempty"`
	Timeout  time.Duration          `json:"timeout"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// TestTargetType defines the type of test target
type TestTargetType string

const (
	TestTargetTypeHTTP      TestTargetType = "http"
	TestTargetTypeGRPC      TestTargetType = "grpc"
	TestTargetTypeWebSocket TestTargetType = "websocket"
	TestTargetTypeDatabase  TestTargetType = "database"
	TestTargetTypeQueue     TestTargetType = "queue"
	TestTargetTypeCustom    TestTargetType = "custom"
)

// TestAuth represents authentication configuration for tests
type TestAuth struct {
	Type     TestAuthType           `json:"type"`
	Username string                 `json:"username,omitempty"`
	Password string                 `json:"password,omitempty"`
	Token    string                 `json:"token,omitempty"`
	Headers  map[string]string      `json:"headers,omitempty"`
	Custom   map[string]interface{} `json:"custom,omitempty"`
}

// TestAuthType defines the type of authentication
type TestAuthType string

const (
	TestAuthTypeNone   TestAuthType = "none"
	TestAuthTypeBasic  TestAuthType = "basic"
	TestAuthTypeBearer TestAuthType = "bearer"
	TestAuthTypeAPIKey TestAuthType = "api_key"
	TestAuthTypeOAuth2 TestAuthType = "oauth2"
	TestAuthTypeCustom TestAuthType = "custom"
)

// TLSConfig represents TLS configuration for tests
type TLSConfig struct {
	Enabled            bool   `json:"enabled"`
	InsecureSkipVerify bool   `json:"insecure_skip_verify"`
	CertFile           string `json:"cert_file,omitempty"`
	KeyFile            string `json:"key_file,omitempty"`
	CAFile             string `json:"ca_file,omitempty"`
}

// Performance Test Types

// PerformanceTestType defines the type of performance test
type PerformanceTestType string

const (
	PerformanceTestTypeLoad      PerformanceTestType = "load"
	PerformanceTestTypeStress    PerformanceTestType = "stress"
	PerformanceTestTypeSpike     PerformanceTestType = "spike"
	PerformanceTestTypeVolume    PerformanceTestType = "volume"
	PerformanceTestTypeEndurance PerformanceTestType = "endurance"
	PerformanceTestTypeBaseline  PerformanceTestType = "baseline"
)

// LoadPattern defines the load pattern for tests
type LoadPattern string

const (
	LoadPatternConstant   LoadPattern = "constant"
	LoadPatternRampUp     LoadPattern = "ramp_up"
	LoadPatternRampDown   LoadPattern = "ramp_down"
	LoadPatternSteps      LoadPattern = "steps"
	LoadPatternSpike      LoadPattern = "spike"
	LoadPatternSinusoidal LoadPattern = "sinusoidal"
)

// StressPattern defines the stress pattern for tests
type StressPattern string

const (
	StressPatternLinear      StressPattern = "linear"
	StressPatternExponential StressPattern = "exponential"
	StressPatternLogarithmic StressPattern = "logarithmic"
	StressPatternCustom      StressPattern = "custom"
)

// Scenario Types

// PerformanceScenario represents a performance test scenario
type PerformanceScenario struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Weight      float64                `json:"weight"`
	Steps       []*PerformanceStep     `json:"steps"`
	Variables   map[string]interface{} `json:"variables,omitempty"`
	Enabled     bool                   `json:"enabled"`
}

// LoadScenario represents a load test scenario
type LoadScenario struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Weight      float64                `json:"weight"`
	Steps       []*LoadStep            `json:"steps"`
	Variables   map[string]interface{} `json:"variables,omitempty"`
	Enabled     bool                   `json:"enabled"`
}

// StressScenario represents a stress test scenario
type StressScenario struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Weight      float64                `json:"weight"`
	Steps       []*StressStep          `json:"steps"`
	Variables   map[string]interface{} `json:"variables,omitempty"`
	Enabled     bool                   `json:"enabled"`
}

// IntegrationScenario represents an integration test scenario
type IntegrationScenario struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Services    []string               `json:"services"`
	Steps       []*IntegrationStep     `json:"steps"`
	Variables   map[string]interface{} `json:"variables,omitempty"`
	Enabled     bool                   `json:"enabled"`
}

// Step Types

// PerformanceStep represents a step in a performance test
type PerformanceStep struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Type        PerformanceStepType    `json:"type"`
	Request     *TestRequest           `json:"request,omitempty"`
	ThinkTime   time.Duration          `json:"think_time"`
	Timeout     time.Duration          `json:"timeout"`
	Retries     int                    `json:"retries"`
	Variables   map[string]interface{} `json:"variables,omitempty"`
	Validations []*StepValidation      `json:"validations,omitempty"`
	Order       int                    `json:"order"`
	Enabled     bool                   `json:"enabled"`
}

// LoadStep represents a step in a load test
type LoadStep struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Type        LoadStepType           `json:"type"`
	Request     *TestRequest           `json:"request,omitempty"`
	ThinkTime   time.Duration          `json:"think_time"`
	Timeout     time.Duration          `json:"timeout"`
	Retries     int                    `json:"retries"`
	Variables   map[string]interface{} `json:"variables,omitempty"`
	Validations []*StepValidation      `json:"validations,omitempty"`
	Order       int                    `json:"order"`
	Enabled     bool                   `json:"enabled"`
}

// StressStep represents a step in a stress test
type StressStep struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Type        StressStepType         `json:"type"`
	Request     *TestRequest           `json:"request,omitempty"`
	ThinkTime   time.Duration          `json:"think_time"`
	Timeout     time.Duration          `json:"timeout"`
	Retries     int                    `json:"retries"`
	Variables   map[string]interface{} `json:"variables,omitempty"`
	Validations []*StepValidation      `json:"validations,omitempty"`
	Order       int                    `json:"order"`
	Enabled     bool                   `json:"enabled"`
}

// IntegrationStep represents a step in an integration test
type IntegrationStep struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Type        IntegrationStepType    `json:"type"`
	ServiceID   string                 `json:"service_id"`
	Request     *TestRequest           `json:"request,omitempty"`
	Timeout     time.Duration          `json:"timeout"`
	Retries     int                    `json:"retries"`
	Variables   map[string]interface{} `json:"variables,omitempty"`
	Validations []*StepValidation      `json:"validations,omitempty"`
	Order       int                    `json:"order"`
	Enabled     bool                   `json:"enabled"`
}

// Step Type Enums

// PerformanceStepType defines the type of performance test step
type PerformanceStepType string

const (
	PerformanceStepTypeHTTP     PerformanceStepType = "http"
	PerformanceStepTypeGRPC     PerformanceStepType = "grpc"
	PerformanceStepTypeDatabase PerformanceStepType = "database"
	PerformanceStepTypeCustom   PerformanceStepType = "custom"
)

// LoadStepType defines the type of load test step
type LoadStepType string

const (
	LoadStepTypeHTTP     LoadStepType = "http"
	LoadStepTypeGRPC     LoadStepType = "grpc"
	LoadStepTypeDatabase LoadStepType = "database"
	LoadStepTypeCustom   LoadStepType = "custom"
)

// StressStepType defines the type of stress test step
type StressStepType string

const (
	StressStepTypeHTTP     StressStepType = "http"
	StressStepTypeGRPC     StressStepType = "grpc"
	StressStepTypeDatabase StressStepType = "database"
	StressStepTypeCustom   StressStepType = "custom"
)

// IntegrationStepType defines the type of integration test step
type IntegrationStepType string

const (
	IntegrationStepTypeHTTP     IntegrationStepType = "http"
	IntegrationStepTypeGRPC     IntegrationStepType = "grpc"
	IntegrationStepTypeDatabase IntegrationStepType = "database"
	IntegrationStepTypeEvent    IntegrationStepType = "event"
	IntegrationStepTypeCustom   IntegrationStepType = "custom"
)

// Request and Validation Types

// TestRequest represents a test request
type TestRequest struct {
	Method     string                 `json:"method"`
	URL        string                 `json:"url"`
	Headers    map[string]string      `json:"headers,omitempty"`
	Body       interface{}            `json:"body,omitempty"`
	Parameters map[string]interface{} `json:"parameters,omitempty"`
	Auth       *TestAuth              `json:"auth,omitempty"`
	Timeout    time.Duration          `json:"timeout"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// StepValidation represents validation for a test step
type StepValidation struct {
	ID       string                `json:"id"`
	Type     StepValidationType    `json:"type"`
	Field    string                `json:"field"`
	Operator TestAssertionOperator `json:"operator"`
	Expected interface{}           `json:"expected"`
	Message  string                `json:"message,omitempty"`
	Critical bool                  `json:"critical"`
}

// StepValidationType defines the type of step validation
type StepValidationType string

const (
	StepValidationTypeResponse     StepValidationType = "response"
	StepValidationTypeStatusCode   StepValidationType = "status_code"
	StepValidationTypeHeader       StepValidationType = "header"
	StepValidationTypeBody         StepValidationType = "body"
	StepValidationTypeResponseTime StepValidationType = "response_time"
	StepValidationTypeCustom       StepValidationType = "custom"
)

// Threshold Types

// PerformanceThresholds defines thresholds for performance tests
type PerformanceThresholds struct {
	ResponseTime   *ThresholdConfig            `json:"response_time,omitempty"`
	Throughput     *ThresholdConfig            `json:"throughput,omitempty"`
	ErrorRate      *ThresholdConfig            `json:"error_rate,omitempty"`
	CPUUsage       *ThresholdConfig            `json:"cpu_usage,omitempty"`
	MemoryUsage    *ThresholdConfig            `json:"memory_usage,omitempty"`
	NetworkLatency *ThresholdConfig            `json:"network_latency,omitempty"`
	Custom         map[string]*ThresholdConfig `json:"custom,omitempty"`
}

// LoadTestThresholds defines thresholds for load tests
type LoadTestThresholds struct {
	ResponseTime  *ThresholdConfig            `json:"response_time,omitempty"`
	Throughput    *ThresholdConfig            `json:"throughput,omitempty"`
	ErrorRate     *ThresholdConfig            `json:"error_rate,omitempty"`
	Concurrency   *ThresholdConfig            `json:"concurrency,omitempty"`
	ResourceUsage *ThresholdConfig            `json:"resource_usage,omitempty"`
	Custom        map[string]*ThresholdConfig `json:"custom,omitempty"`
}

// StressTestThresholds defines thresholds for stress tests
type StressTestThresholds struct {
	ResponseTime  *ThresholdConfig            `json:"response_time,omitempty"`
	Throughput    *ThresholdConfig            `json:"throughput,omitempty"`
	ErrorRate     *ThresholdConfig            `json:"error_rate,omitempty"`
	BreakingPoint *ThresholdConfig            `json:"breaking_point,omitempty"`
	RecoveryTime  *ThresholdConfig            `json:"recovery_time,omitempty"`
	ResourceUsage *ThresholdConfig            `json:"resource_usage,omitempty"`
	Custom        map[string]*ThresholdConfig `json:"custom,omitempty"`
}

// IntegrationTestThresholds defines thresholds for integration tests
type IntegrationTestThresholds struct {
	ResponseTime    *ThresholdConfig            `json:"response_time,omitempty"`
	DataConsistency *ThresholdConfig            `json:"data_consistency,omitempty"`
	ServiceHealth   *ThresholdConfig            `json:"service_health,omitempty"`
	ErrorRate       *ThresholdConfig            `json:"error_rate,omitempty"`
	Custom          map[string]*ThresholdConfig `json:"custom,omitempty"`
}

// ThresholdConfig defines a threshold configuration
type ThresholdConfig struct {
	Min      float64 `json:"min,omitempty"`
	Max      float64 `json:"max,omitempty"`
	Target   float64 `json:"target,omitempty"`
	Critical bool    `json:"critical"`
	Unit     string  `json:"unit,omitempty"`
}

// Monitoring Types

// PerformanceMonitoring defines monitoring configuration for performance tests
type PerformanceMonitoring struct {
	Enabled            bool                   `json:"enabled"`
	Interval           time.Duration          `json:"interval"`
	Metrics            []string               `json:"metrics"`
	SystemMetrics      bool                   `json:"system_metrics"`
	ApplicationMetrics bool                   `json:"application_metrics"`
	CustomMetrics      map[string]interface{} `json:"custom_metrics,omitempty"`
	Alerts             []*MonitoringAlert     `json:"alerts,omitempty"`
}

// MonitoringAlert defines an alert for monitoring
type MonitoringAlert struct {
	ID        string   `json:"id"`
	Name      string   `json:"name"`
	Metric    string   `json:"metric"`
	Condition string   `json:"condition"`
	Threshold float64  `json:"threshold"`
	Actions   []string `json:"actions"`
	Enabled   bool     `json:"enabled"`
}

// Integration Test Types

// IntegrationService represents a service in integration tests
type IntegrationService struct {
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	Type         string                 `json:"type"`
	URL          string                 `json:"url"`
	Health       string                 `json:"health"`
	Dependencies []string               `json:"dependencies,omitempty"`
	Config       map[string]interface{} `json:"config,omitempty"`
}

// ServiceDependency represents a service dependency
type ServiceDependency struct {
	ServiceID   string `json:"service_id"`
	DependsOnID string `json:"depends_on_id"`
	Type        string `json:"type"`
	Required    bool   `json:"required"`
	HealthCheck bool   `json:"health_check"`
}

// DataFlowTest represents a data flow test
type DataFlowTest struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Source      string                 `json:"source"`
	Target      string                 `json:"target"`
	Data        map[string]interface{} `json:"data"`
	Validations []*DataValidation      `json:"validations"`
	Timeout     time.Duration          `json:"timeout"`
}

// DataValidation represents data validation in integration tests
type DataValidation struct {
	ID       string                `json:"id"`
	Field    string                `json:"field"`
	Type     DataValidationType    `json:"type"`
	Operator TestAssertionOperator `json:"operator"`
	Expected interface{}           `json:"expected"`
	Message  string                `json:"message,omitempty"`
}

// DataValidationType defines the type of data validation
type DataValidationType string

const (
	DataValidationTypeExists   DataValidationType = "exists"
	DataValidationTypeEquals   DataValidationType = "equals"
	DataValidationTypeContains DataValidationType = "contains"
	DataValidationTypeFormat   DataValidationType = "format"
	DataValidationTypeRange    DataValidationType = "range"
	DataValidationTypeCustom   DataValidationType = "custom"
)

// ContractTest represents a contract test
type ContractTest struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Provider    string                 `json:"provider"`
	Consumer    string                 `json:"consumer"`
	Contract    map[string]interface{} `json:"contract"`
	Validations []*ContractValidation  `json:"validations"`
}

// ContractValidation represents contract validation
type ContractValidation struct {
	ID       string                 `json:"id"`
	Field    string                 `json:"field"`
	Type     ContractValidationType `json:"type"`
	Expected interface{}            `json:"expected"`
	Message  string                 `json:"message,omitempty"`
}

// ContractValidationType defines the type of contract validation
type ContractValidationType string

const (
	ContractValidationTypeSchema   ContractValidationType = "schema"
	ContractValidationTypeResponse ContractValidationType = "response"
	ContractValidationTypeRequest  ContractValidationType = "request"
	ContractValidationTypeCustom   ContractValidationType = "custom"
)

// Result Types

// PerformanceTestResult contains results of performance tests
type PerformanceTestResult struct {
	ID              string                  `json:"id"`
	TestID          string                  `json:"test_id"`
	Status          TestResultStatus        `json:"status"`
	StartTime       time.Time               `json:"start_time"`
	EndTime         time.Time               `json:"end_time"`
	Duration        time.Duration           `json:"duration"`
	Summary         *PerformanceTestSummary `json:"summary"`
	Metrics         *PerformanceMetrics     `json:"metrics"`
	ScenarioResults []*ScenarioResult       `json:"scenario_results"`
	Thresholds      *ThresholdResults       `json:"thresholds"`
	Environment     string                  `json:"environment"`
	Metadata        map[string]interface{}  `json:"metadata,omitempty"`
}

// LoadTestResult contains results of load tests
type LoadTestResult struct {
	ID              string                 `json:"id"`
	TestID          string                 `json:"test_id"`
	Status          TestResultStatus       `json:"status"`
	StartTime       time.Time              `json:"start_time"`
	EndTime         time.Time              `json:"end_time"`
	Duration        time.Duration          `json:"duration"`
	Summary         *LoadTestSummary       `json:"summary"`
	Metrics         *LoadTestMetrics       `json:"metrics"`
	ScenarioResults []*ScenarioResult      `json:"scenario_results"`
	Thresholds      *ThresholdResults      `json:"thresholds"`
	Environment     string                 `json:"environment"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

// StressTestResult contains results of stress tests
type StressTestResult struct {
	ID              string                 `json:"id"`
	TestID          string                 `json:"test_id"`
	Status          TestResultStatus       `json:"status"`
	StartTime       time.Time              `json:"start_time"`
	EndTime         time.Time              `json:"end_time"`
	Duration        time.Duration          `json:"duration"`
	Summary         *StressTestSummary     `json:"summary"`
	Metrics         *StressTestMetrics     `json:"metrics"`
	ScenarioResults []*ScenarioResult      `json:"scenario_results"`
	Thresholds      *ThresholdResults      `json:"thresholds"`
	Environment     string                 `json:"environment"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

// IntegrationTestResult contains results of integration tests
type IntegrationTestResult struct {
	ID              string                  `json:"id"`
	TestID          string                  `json:"test_id"`
	Status          TestResultStatus        `json:"status"`
	StartTime       time.Time               `json:"start_time"`
	EndTime         time.Time               `json:"end_time"`
	Duration        time.Duration           `json:"duration"`
	Summary         *IntegrationTestSummary `json:"summary"`
	ServiceResults  []*ServiceTestResult    `json:"service_results"`
	DataFlowResults []*DataFlowTestResult   `json:"data_flow_results"`
	ContractResults []*ContractTestResult   `json:"contract_results"`
	Thresholds      *ThresholdResults       `json:"thresholds"`
	Environment     string                  `json:"environment"`
	Metadata        map[string]interface{}  `json:"metadata,omitempty"`
}

// SystemIntegrationReport contains system integration validation results
type SystemIntegrationReport struct {
	ID           string                      `json:"id"`
	Status       TestResultStatus            `json:"status"`
	StartTime    time.Time                   `json:"start_time"`
	EndTime      time.Time                   `json:"end_time"`
	Duration     time.Duration               `json:"duration"`
	Services     []*ServiceIntegrationResult `json:"services"`
	Dependencies []*DependencyTestResult     `json:"dependencies"`
	DataFlows    []*DataFlowTestResult       `json:"data_flows"`
	HealthChecks []*HealthCheckResult        `json:"health_checks"`
	Summary      *SystemIntegrationSummary   `json:"summary"`
	Environment  string                      `json:"environment"`
	Metadata     map[string]interface{}      `json:"metadata,omitempty"`
}

// Summary Types

// PerformanceTestSummary contains summary of performance test results
type PerformanceTestSummary struct {
	TotalRequests       int64         `json:"total_requests"`
	SuccessfulRequests  int64         `json:"successful_requests"`
	FailedRequests      int64         `json:"failed_requests"`
	ErrorRate           float64       `json:"error_rate"`
	AverageResponseTime time.Duration `json:"average_response_time"`
	MinResponseTime     time.Duration `json:"min_response_time"`
	MaxResponseTime     time.Duration `json:"max_response_time"`
	P50ResponseTime     time.Duration `json:"p50_response_time"`
	P95ResponseTime     time.Duration `json:"p95_response_time"`
	P99ResponseTime     time.Duration `json:"p99_response_time"`
	Throughput          float64       `json:"throughput"`
	ConcurrentUsers     int           `json:"concurrent_users"`
}

// LoadTestSummary contains summary of load test results
type LoadTestSummary struct {
	TotalRequests       int64         `json:"total_requests"`
	SuccessfulRequests  int64         `json:"successful_requests"`
	FailedRequests      int64         `json:"failed_requests"`
	ErrorRate           float64       `json:"error_rate"`
	AverageResponseTime time.Duration `json:"average_response_time"`
	P95ResponseTime     time.Duration `json:"p95_response_time"`
	P99ResponseTime     time.Duration `json:"p99_response_time"`
	Throughput          float64       `json:"throughput"`
	MaxConcurrentUsers  int           `json:"max_concurrent_users"`
	PeakThroughput      float64       `json:"peak_throughput"`
}

// StressTestSummary contains summary of stress test results
type StressTestSummary struct {
	TotalRequests       int64         `json:"total_requests"`
	SuccessfulRequests  int64         `json:"successful_requests"`
	FailedRequests      int64         `json:"failed_requests"`
	ErrorRate           float64       `json:"error_rate"`
	AverageResponseTime time.Duration `json:"average_response_time"`
	P95ResponseTime     time.Duration `json:"p95_response_time"`
	P99ResponseTime     time.Duration `json:"p99_response_time"`
	Throughput          float64       `json:"throughput"`
	MaxConcurrentUsers  int           `json:"max_concurrent_users"`
	BreakingPoint       int           `json:"breaking_point"`
	RecoveryTime        time.Duration `json:"recovery_time"`
}

// IntegrationTestSummary contains summary of integration test results
type IntegrationTestSummary struct {
	TotalServices       int     `json:"total_services"`
	HealthyServices     int     `json:"healthy_services"`
	UnhealthyServices   int     `json:"unhealthy_services"`
	TotalDataFlows      int     `json:"total_data_flows"`
	SuccessfulDataFlows int     `json:"successful_data_flows"`
	FailedDataFlows     int     `json:"failed_data_flows"`
	TotalContracts      int     `json:"total_contracts"`
	ValidContracts      int     `json:"valid_contracts"`
	InvalidContracts    int     `json:"invalid_contracts"`
	OverallHealth       float64 `json:"overall_health"`
}

// SystemIntegrationSummary contains summary of system integration results
type SystemIntegrationSummary struct {
	TotalServices       int     `json:"total_services"`
	HealthyServices     int     `json:"healthy_services"`
	UnhealthyServices   int     `json:"unhealthy_services"`
	TotalDependencies   int     `json:"total_dependencies"`
	ValidDependencies   int     `json:"valid_dependencies"`
	InvalidDependencies int     `json:"invalid_dependencies"`
	TotalDataFlows      int     `json:"total_data_flows"`
	ActiveDataFlows     int     `json:"active_data_flows"`
	FailedDataFlows     int     `json:"failed_data_flows"`
	SystemHealth        float64 `json:"system_health"`
	IntegrationScore    float64 `json:"integration_score"`
}

// Metrics Types

// PerformanceMetrics contains detailed performance metrics
type PerformanceMetrics struct {
	ResponseTimes  *ResponseTimeMetrics   `json:"response_times"`
	Throughput     *ThroughputMetrics     `json:"throughput"`
	ErrorRates     *ErrorRateMetrics      `json:"error_rates"`
	ResourceUsage  *ResourceUsageMetrics  `json:"resource_usage"`
	NetworkMetrics *NetworkMetrics        `json:"network_metrics"`
	CustomMetrics  map[string]interface{} `json:"custom_metrics,omitempty"`
}

// LoadTestMetrics contains detailed load test metrics
type LoadTestMetrics struct {
	ResponseTimes *ResponseTimeMetrics   `json:"response_times"`
	Throughput    *ThroughputMetrics     `json:"throughput"`
	ErrorRates    *ErrorRateMetrics      `json:"error_rates"`
	Concurrency   *ConcurrencyMetrics    `json:"concurrency"`
	ResourceUsage *ResourceUsageMetrics  `json:"resource_usage"`
	CustomMetrics map[string]interface{} `json:"custom_metrics,omitempty"`
}

// StressTestMetrics contains detailed stress test metrics
type StressTestMetrics struct {
	ResponseTimes *ResponseTimeMetrics   `json:"response_times"`
	Throughput    *ThroughputMetrics     `json:"throughput"`
	ErrorRates    *ErrorRateMetrics      `json:"error_rates"`
	Concurrency   *ConcurrencyMetrics    `json:"concurrency"`
	ResourceUsage *ResourceUsageMetrics  `json:"resource_usage"`
	BreakingPoint *BreakingPointMetrics  `json:"breaking_point"`
	CustomMetrics map[string]interface{} `json:"custom_metrics,omitempty"`
}

// Detailed Metrics Types

// ResponseTimeMetrics contains response time metrics
type ResponseTimeMetrics struct {
	Average time.Duration `json:"average"`
	Min     time.Duration `json:"min"`
	Max     time.Duration `json:"max"`
	P50     time.Duration `json:"p50"`
	P90     time.Duration `json:"p90"`
	P95     time.Duration `json:"p95"`
	P99     time.Duration `json:"p99"`
	P999    time.Duration `json:"p999"`
	StdDev  time.Duration `json:"std_dev"`
}

// ThroughputMetrics contains throughput metrics
type ThroughputMetrics struct {
	Average float64 `json:"average"`
	Min     float64 `json:"min"`
	Max     float64 `json:"max"`
	Peak    float64 `json:"peak"`
	Total   int64   `json:"total"`
}

// ErrorRateMetrics contains error rate metrics
type ErrorRateMetrics struct {
	Overall        float64            `json:"overall"`
	ByStatusCode   map[string]float64 `json:"by_status_code"`
	ByErrorType    map[string]float64 `json:"by_error_type"`
	TimeoutRate    float64            `json:"timeout_rate"`
	ConnectionRate float64            `json:"connection_rate"`
}

// ConcurrencyMetrics contains concurrency metrics
type ConcurrencyMetrics struct {
	Average int `json:"average"`
	Min     int `json:"min"`
	Max     int `json:"max"`
	Peak    int `json:"peak"`
}

// ResourceUsageMetrics contains resource usage metrics
type ResourceUsageMetrics struct {
	CPU     *ResourceMetric `json:"cpu"`
	Memory  *ResourceMetric `json:"memory"`
	Disk    *ResourceMetric `json:"disk"`
	Network *ResourceMetric `json:"network"`
}

// ResourceMetric contains metrics for a specific resource
type ResourceMetric struct {
	Average float64 `json:"average"`
	Min     float64 `json:"min"`
	Max     float64 `json:"max"`
	Peak    float64 `json:"peak"`
	Unit    string  `json:"unit"`
}

// NetworkMetrics contains network-related metrics
type NetworkMetrics struct {
	Latency    *ResponseTimeMetrics `json:"latency"`
	Bandwidth  *ThroughputMetrics   `json:"bandwidth"`
	PacketLoss float64              `json:"packet_loss"`
}

// BreakingPointMetrics contains breaking point metrics for stress tests
type BreakingPointMetrics struct {
	MaxUsers        int           `json:"max_users"`
	MaxThroughput   float64       `json:"max_throughput"`
	BreakingPoint   int           `json:"breaking_point"`
	RecoveryTime    time.Duration `json:"recovery_time"`
	DegradationRate float64       `json:"degradation_rate"`
}

// Result Detail Types

// ScenarioResult contains results for a specific scenario
type ScenarioResult struct {
	ScenarioID   string                 `json:"scenario_id"`
	Status       TestResultStatus       `json:"status"`
	Duration     time.Duration          `json:"duration"`
	Requests     int64                  `json:"requests"`
	Errors       int64                  `json:"errors"`
	ErrorRate    float64                `json:"error_rate"`
	ResponseTime *ResponseTimeMetrics   `json:"response_time"`
	Throughput   float64                `json:"throughput"`
	StepResults  []*StepResult          `json:"step_results"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// StepResult contains results for a specific step
type StepResult struct {
	StepID       string                 `json:"step_id"`
	Status       TestResultStatus       `json:"status"`
	Duration     time.Duration          `json:"duration"`
	Requests     int64                  `json:"requests"`
	Errors       int64                  `json:"errors"`
	ErrorRate    float64                `json:"error_rate"`
	ResponseTime *ResponseTimeMetrics   `json:"response_time"`
	Validations  []*ValidationResult    `json:"validations"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// ValidationResult contains results for a specific validation
type ValidationResult struct {
	ValidationID string           `json:"validation_id"`
	Status       TestResultStatus `json:"status"`
	Expected     interface{}      `json:"expected"`
	Actual       interface{}      `json:"actual"`
	Message      string           `json:"message,omitempty"`
	Error        string           `json:"error,omitempty"`
}

// ThresholdResults contains threshold evaluation results
type ThresholdResults struct {
	Overall    ThresholdStatus             `json:"overall"`
	Individual map[string]*ThresholdResult `json:"individual"`
}

// ThresholdResult contains result for a specific threshold
type ThresholdResult struct {
	Name     string          `json:"name"`
	Status   ThresholdStatus `json:"status"`
	Expected float64         `json:"expected"`
	Actual   float64         `json:"actual"`
	Unit     string          `json:"unit,omitempty"`
	Message  string          `json:"message,omitempty"`
	Critical bool            `json:"critical"`
}

// ThresholdStatus defines the status of threshold evaluation
type ThresholdStatus string

const (
	ThresholdStatusPassed  ThresholdStatus = "passed"
	ThresholdStatusFailed  ThresholdStatus = "failed"
	ThresholdStatusWarning ThresholdStatus = "warning"
)

// Service and Dependency Test Results

// ServiceTestResult contains results for a specific service test
type ServiceTestResult struct {
	ServiceID    string                 `json:"service_id"`
	Status       TestResultStatus       `json:"status"`
	Health       string                 `json:"health"`
	ResponseTime time.Duration          `json:"response_time"`
	Availability float64                `json:"availability"`
	Endpoints    []*EndpointTestResult  `json:"endpoints"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// EndpointTestResult contains results for a specific endpoint test
type EndpointTestResult struct {
	EndpointID   string           `json:"endpoint_id"`
	Status       TestResultStatus `json:"status"`
	ResponseTime time.Duration    `json:"response_time"`
	StatusCode   int              `json:"status_code"`
	Error        string           `json:"error,omitempty"`
}

// DependencyTestResult contains results for dependency tests
type DependencyTestResult struct {
	ServiceID    string           `json:"service_id"`
	DependencyID string           `json:"dependency_id"`
	Status       TestResultStatus `json:"status"`
	Health       string           `json:"health"`
	ResponseTime time.Duration    `json:"response_time"`
	Error        string           `json:"error,omitempty"`
}

// DataFlowTestResult contains results for data flow tests
type DataFlowTestResult struct {
	DataFlowID       string                 `json:"data_flow_id"`
	Status           TestResultStatus       `json:"status"`
	Duration         time.Duration          `json:"duration"`
	RecordsProcessed int64                  `json:"records_processed"`
	RecordsFailed    int64                  `json:"records_failed"`
	Validations      []*ValidationResult    `json:"validations"`
	Error            string                 `json:"error,omitempty"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
}

// ContractTestResult contains results for contract tests
type ContractTestResult struct {
	ContractID  string                 `json:"contract_id"`
	Status      TestResultStatus       `json:"status"`
	Provider    string                 `json:"provider"`
	Consumer    string                 `json:"consumer"`
	Validations []*ValidationResult    `json:"validations"`
	Error       string                 `json:"error,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// ServiceIntegrationResult contains results for service integration
type ServiceIntegrationResult struct {
	ServiceID    string                  `json:"service_id"`
	Status       TestResultStatus        `json:"status"`
	Health       string                  `json:"health"`
	Dependencies []*DependencyTestResult `json:"dependencies"`
	Endpoints    []*EndpointTestResult   `json:"endpoints"`
	Metadata     map[string]interface{}  `json:"metadata,omitempty"`
}

// HealthCheckResult contains results for health checks
type HealthCheckResult struct {
	ServiceID    string           `json:"service_id"`
	CheckName    string           `json:"check_name"`
	Status       TestResultStatus `json:"status"`
	ResponseTime time.Duration    `json:"response_time"`
	Message      string           `json:"message,omitempty"`
	Error        string           `json:"error,omitempty"`
}
