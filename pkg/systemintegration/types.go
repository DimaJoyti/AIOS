package systemintegration

import (
	"context"
	"time"
)

// Enhanced System Integration types and interfaces for AIOS

// SystemIntegrationHub defines the interface for managing system integration
type SystemIntegrationHub interface {
	// Service Management
	RegisterService(service *Service) (*Service, error)
	GetService(serviceID string) (*Service, error)
	UpdateService(service *Service) (*Service, error)
	UnregisterService(serviceID string) error
	ListServices(filter *ServiceFilter) ([]*Service, error)

	// Service Discovery
	DiscoverServices(criteria *DiscoveryCriteria) ([]*Service, error)
	GetServiceEndpoints(serviceID string) ([]*ServiceEndpoint, error)
	UpdateServiceHealth(serviceID string, health *ServiceHealth) error

	// API Gateway Management
	CreateRoute(route *APIRoute) (*APIRoute, error)
	GetRoute(routeID string) (*APIRoute, error)
	UpdateRoute(route *APIRoute) (*APIRoute, error)
	DeleteRoute(routeID string) error
	ListRoutes(filter *RouteFilter) ([]*APIRoute, error)

	// Event Bus Management
	PublishEvent(event *SystemEvent) error
	SubscribeToEvents(subscription *EventSubscription) error
	UnsubscribeFromEvents(subscriptionID string) error
	GetEventHistory(filter *EventFilter) ([]*SystemEvent, error)

	// Data Flow Orchestration
	CreateDataFlow(flow *DataFlow) (*DataFlow, error)
	GetDataFlow(flowID string) (*DataFlow, error)
	UpdateDataFlow(flow *DataFlow) (*DataFlow, error)
	DeleteDataFlow(flowID string) error
	StartDataFlow(flowID string) error
	StopDataFlow(flowID string) error
	GetDataFlowStatus(flowID string) (*DataFlowStatus, error)

	// Security Integration
	AuthenticateRequest(request *AuthRequest) (*AuthResult, error)
	AuthorizeRequest(request *AuthzRequest) (*AuthzResult, error)
	ValidateToken(token string) (*TokenValidation, error)
	EnforcePolicy(policy *SecurityPolicy, context *PolicyContext) (*PolicyResult, error)

	// Monitoring Integration
	GetSystemHealth() (*SystemHealth, error)
	GetSystemMetrics(timeRange *TimeRange) (*SystemMetrics, error)
	GetServiceMetrics(serviceID string, timeRange *TimeRange) (*ServiceMetrics, error)
	CreateAlert(alert *Alert) (*Alert, error)
	GetAlerts(filter *AlertFilter) ([]*Alert, error)

	// Configuration Management
	GetConfiguration(key string) (*Configuration, error)
	SetConfiguration(config *Configuration) error
	GetServiceConfiguration(serviceID string) (*ServiceConfiguration, error)
	UpdateServiceConfiguration(serviceID string, config *ServiceConfiguration) error

	// Deployment Management
	CreateDeployment(deployment *Deployment) (*Deployment, error)
	GetDeployment(deploymentID string) (*Deployment, error)
	UpdateDeployment(deployment *Deployment) (*Deployment, error)
	GetDeploymentStatus(deploymentID string) (*DeploymentStatus, error)
	RollbackDeployment(deploymentID string) error
}

// EventHandler defines the interface for event handlers
type EventHandler interface {
	HandleEvent(ctx context.Context, event *SystemEvent) error
	GetEventTypes() []string
	GetHandlerID() string
}

// ServiceHealthChecker defines the interface for service health checking
type ServiceHealthChecker interface {
	CheckHealth(ctx context.Context, service *Service) (*ServiceHealth, error)
	GetHealthCheckInterval() time.Duration
}

// Core Types

// Service represents a system service
type Service struct {
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	Description  string                 `json:"description"`
	Type         ServiceType            `json:"type"`
	Version      string                 `json:"version"`
	Status       ServiceStatus          `json:"status"`
	Endpoints    []*ServiceEndpoint     `json:"endpoints"`
	Health       *ServiceHealth         `json:"health,omitempty"`
	Config       *ServiceConfiguration  `json:"config,omitempty"`
	Dependencies []string               `json:"dependencies,omitempty"`
	Tags         []string               `json:"tags,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	RegisteredAt time.Time              `json:"registered_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
	LastSeenAt   *time.Time             `json:"last_seen_at,omitempty"`
}

// ServiceType defines the type of service
type ServiceType string

const (
	ServiceTypeCore            ServiceType = "core"
	ServiceTypeAgent           ServiceType = "agent"
	ServiceTypeCollaboration   ServiceType = "collaboration"
	ServiceTypeIntegration     ServiceType = "integration"
	ServiceTypeDataIntegration ServiceType = "data_integration"
	ServiceTypeAPI             ServiceType = "api"
	ServiceTypeUI              ServiceType = "ui"
	ServiceTypeDatabase        ServiceType = "database"
	ServiceTypeCache           ServiceType = "cache"
	ServiceTypeQueue           ServiceType = "queue"
	ServiceTypeStorage         ServiceType = "storage"
	ServiceTypeMonitoring      ServiceType = "monitoring"
	ServiceTypeSecurity        ServiceType = "security"
	ServiceTypeExternal        ServiceType = "external"
)

// ServiceStatus defines the status of a service
type ServiceStatus string

const (
	ServiceStatusHealthy     ServiceStatus = "healthy"
	ServiceStatusDegraded    ServiceStatus = "degraded"
	ServiceStatusUnhealthy   ServiceStatus = "unhealthy"
	ServiceStatusStarting    ServiceStatus = "starting"
	ServiceStatusStopping    ServiceStatus = "stopping"
	ServiceStatusStopped     ServiceStatus = "stopped"
	ServiceStatusMaintenance ServiceStatus = "maintenance"
)

// ServiceEndpoint represents a service endpoint
type ServiceEndpoint struct {
	ID       string                 `json:"id"`
	Name     string                 `json:"name"`
	URL      string                 `json:"url"`
	Protocol string                 `json:"protocol"`
	Port     int                    `json:"port"`
	Path     string                 `json:"path"`
	Method   string                 `json:"method,omitempty"`
	Headers  map[string]string      `json:"headers,omitempty"`
	Auth     *EndpointAuth          `json:"auth,omitempty"`
	Health   *EndpointHealth        `json:"health,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// EndpointAuth contains endpoint authentication information
type EndpointAuth struct {
	Type     AuthType               `json:"type"`
	Token    string                 `json:"token,omitempty"`
	Username string                 `json:"username,omitempty"`
	Password string                 `json:"password,omitempty"`
	Headers  map[string]string      `json:"headers,omitempty"`
	Custom   map[string]interface{} `json:"custom,omitempty"`
}

// AuthType defines the authentication type
type AuthType string

const (
	AuthTypeNone   AuthType = "none"
	AuthTypeBearer AuthType = "bearer"
	AuthTypeBasic  AuthType = "basic"
	AuthTypeAPIKey AuthType = "api_key"
	AuthTypeOAuth2 AuthType = "oauth2"
	AuthTypeCustom AuthType = "custom"
)

// EndpointHealth contains endpoint health information
type EndpointHealth struct {
	Status       HealthStatus  `json:"status"`
	ResponseTime time.Duration `json:"response_time"`
	LastCheck    time.Time     `json:"last_check"`
	ErrorCount   int           `json:"error_count"`
	Message      string        `json:"message,omitempty"`
}

// ServiceHealth represents the health status of a service
type ServiceHealth struct {
	ServiceID    string                 `json:"service_id"`
	Status       HealthStatus           `json:"status"`
	Message      string                 `json:"message"`
	LastCheck    time.Time              `json:"last_check"`
	Uptime       time.Duration          `json:"uptime"`
	Checks       []*HealthCheck         `json:"checks"`
	Dependencies []*DependencyHealth    `json:"dependencies,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// HealthStatus defines the health status
type HealthStatus string

const (
	HealthStatusHealthy   HealthStatus = "healthy"
	HealthStatusDegraded  HealthStatus = "degraded"
	HealthStatusUnhealthy HealthStatus = "unhealthy"
	HealthStatusUnknown   HealthStatus = "unknown"
)

// HealthCheck represents a specific health check
type HealthCheck struct {
	Name      string                 `json:"name"`
	Status    HealthStatus           `json:"status"`
	Message   string                 `json:"message"`
	Duration  time.Duration          `json:"duration"`
	Timestamp time.Time              `json:"timestamp"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// DependencyHealth represents the health of a service dependency
type DependencyHealth struct {
	ServiceID string       `json:"service_id"`
	Status    HealthStatus `json:"status"`
	Message   string       `json:"message"`
	LastCheck time.Time    `json:"last_check"`
}

// API Gateway Types

// APIRoute represents an API route configuration
type APIRoute struct {
	ID         string                 `json:"id"`
	Name       string                 `json:"name"`
	Path       string                 `json:"path"`
	Method     string                 `json:"method"`
	ServiceID  string                 `json:"service_id"`
	Endpoint   string                 `json:"endpoint"`
	Config     *RouteConfig           `json:"config"`
	Middleware []*RouteMiddleware     `json:"middleware,omitempty"`
	RateLimit  *RateLimit             `json:"rate_limit,omitempty"`
	Auth       *RouteAuth             `json:"auth,omitempty"`
	Cache      *RouteCache            `json:"cache,omitempty"`
	Transform  *RouteTransform        `json:"transform,omitempty"`
	Status     RouteStatus            `json:"status"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt  time.Time              `json:"created_at"`
	UpdatedAt  time.Time              `json:"updated_at"`
}

// RouteStatus defines the status of a route
type RouteStatus string

const (
	RouteStatusActive   RouteStatus = "active"
	RouteStatusInactive RouteStatus = "inactive"
	RouteStatusTesting  RouteStatus = "testing"
)

// RouteConfig contains route configuration
type RouteConfig struct {
	Timeout        time.Duration          `json:"timeout"`
	RetryPolicy    *RetryPolicy           `json:"retry_policy,omitempty"`
	LoadBalancing  *LoadBalancingConfig   `json:"load_balancing,omitempty"`
	CircuitBreaker *CircuitBreakerConfig  `json:"circuit_breaker,omitempty"`
	Custom         map[string]interface{} `json:"custom,omitempty"`
}

// RouteMiddleware represents middleware configuration
type RouteMiddleware struct {
	Name    string                 `json:"name"`
	Type    MiddlewareType         `json:"type"`
	Config  map[string]interface{} `json:"config,omitempty"`
	Order   int                    `json:"order"`
	Enabled bool                   `json:"enabled"`
}

// MiddlewareType defines the type of middleware
type MiddlewareType string

const (
	MiddlewareTypeAuth       MiddlewareType = "auth"
	MiddlewareTypeRateLimit  MiddlewareType = "rate_limit"
	MiddlewareTypeLogging    MiddlewareType = "logging"
	MiddlewareTypeMetrics    MiddlewareType = "metrics"
	MiddlewareTypeTransform  MiddlewareType = "transform"
	MiddlewareTypeValidation MiddlewareType = "validation"
	MiddlewareTypeCORS       MiddlewareType = "cors"
	MiddlewareTypeCustom     MiddlewareType = "custom"
)

// RateLimit defines rate limiting configuration
type RateLimit struct {
	RequestsPerSecond int           `json:"requests_per_second"`
	BurstSize         int           `json:"burst_size"`
	WindowSize        time.Duration `json:"window_size"`
	KeyExtractor      string        `json:"key_extractor,omitempty"`
	SkipSuccessful    bool          `json:"skip_successful"`
	SkipClientErrors  bool          `json:"skip_client_errors"`
}

// RouteAuth contains route authentication configuration
type RouteAuth struct {
	Required bool                   `json:"required"`
	Types    []AuthType             `json:"types"`
	Scopes   []string               `json:"scopes,omitempty"`
	Roles    []string               `json:"roles,omitempty"`
	Custom   map[string]interface{} `json:"custom,omitempty"`
}

// RouteCache contains route caching configuration
type RouteCache struct {
	Enabled    bool          `json:"enabled"`
	TTL        time.Duration `json:"ttl"`
	KeyPattern string        `json:"key_pattern,omitempty"`
	Vary       []string      `json:"vary,omitempty"`
	Conditions []string      `json:"conditions,omitempty"`
}

// RouteTransform contains request/response transformation configuration
type RouteTransform struct {
	Request  *TransformConfig `json:"request,omitempty"`
	Response *TransformConfig `json:"response,omitempty"`
}

// TransformConfig defines transformation rules
type TransformConfig struct {
	Headers map[string]string      `json:"headers,omitempty"`
	Body    *BodyTransform         `json:"body,omitempty"`
	Custom  map[string]interface{} `json:"custom,omitempty"`
}

// BodyTransform defines body transformation rules
type BodyTransform struct {
	Type     TransformType          `json:"type"`
	Template string                 `json:"template,omitempty"`
	Mapping  map[string]string      `json:"mapping,omitempty"`
	Custom   map[string]interface{} `json:"custom,omitempty"`
}

// TransformType defines the type of transformation
type TransformType string

const (
	TransformTypeTemplate TransformType = "template"
	TransformTypeMapping  TransformType = "mapping"
	TransformTypeCustom   TransformType = "custom"
)

// RetryPolicy defines retry behavior
type RetryPolicy struct {
	MaxRetries      int           `json:"max_retries"`
	InitialDelay    time.Duration `json:"initial_delay"`
	MaxDelay        time.Duration `json:"max_delay"`
	BackoffFactor   float64       `json:"backoff_factor"`
	RetryableErrors []string      `json:"retryable_errors,omitempty"`
}

// LoadBalancingConfig defines load balancing configuration
type LoadBalancingConfig struct {
	Strategy    LoadBalancingStrategy  `json:"strategy"`
	HealthCheck bool                   `json:"health_check"`
	Targets     []*LoadBalancingTarget `json:"targets,omitempty"`
}

// LoadBalancingStrategy defines the load balancing strategy
type LoadBalancingStrategy string

const (
	LoadBalancingRoundRobin LoadBalancingStrategy = "round_robin"
	LoadBalancingWeighted   LoadBalancingStrategy = "weighted"
	LoadBalancingLeastConn  LoadBalancingStrategy = "least_conn"
	LoadBalancingIPHash     LoadBalancingStrategy = "ip_hash"
	LoadBalancingRandom     LoadBalancingStrategy = "random"
)

// LoadBalancingTarget represents a load balancing target
type LoadBalancingTarget struct {
	ServiceID string `json:"service_id"`
	Endpoint  string `json:"endpoint"`
	Weight    int    `json:"weight"`
	Enabled   bool   `json:"enabled"`
}

// CircuitBreakerConfig defines circuit breaker configuration
type CircuitBreakerConfig struct {
	Enabled          bool          `json:"enabled"`
	FailureThreshold int           `json:"failure_threshold"`
	RecoveryTimeout  time.Duration `json:"recovery_timeout"`
	HalfOpenRequests int           `json:"half_open_requests"`
	SuccessThreshold int           `json:"success_threshold"`
}

// Event Types

// SystemEvent represents a system-wide event
type SystemEvent struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	Source    string                 `json:"source"`
	Target    string                 `json:"target,omitempty"`
	Data      map[string]interface{} `json:"data"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
	TTL       time.Duration          `json:"ttl,omitempty"`
	Priority  EventPriority          `json:"priority"`
}

// EventPriority defines the priority of an event
type EventPriority string

const (
	EventPriorityLow      EventPriority = "low"
	EventPriorityNormal   EventPriority = "normal"
	EventPriorityHigh     EventPriority = "high"
	EventPriorityCritical EventPriority = "critical"
)

// EventSubscription represents an event subscription
type EventSubscription struct {
	ID         string              `json:"id"`
	EventTypes []string            `json:"event_types"`
	Source     string              `json:"source,omitempty"`
	Target     string              `json:"target,omitempty"`
	Filter     *EventFilter        `json:"filter,omitempty"`
	Handler    EventHandler        `json:"-"`
	Config     *SubscriptionConfig `json:"config,omitempty"`
	Status     SubscriptionStatus  `json:"status"`
	CreatedAt  time.Time           `json:"created_at"`
	UpdatedAt  time.Time           `json:"updated_at"`
}

// SubscriptionStatus defines the status of a subscription
type SubscriptionStatus string

const (
	SubscriptionStatusActive   SubscriptionStatus = "active"
	SubscriptionStatusInactive SubscriptionStatus = "inactive"
	SubscriptionStatusPaused   SubscriptionStatus = "paused"
)

// SubscriptionConfig contains subscription configuration
type SubscriptionConfig struct {
	DeliveryMode    DeliveryMode  `json:"delivery_mode"`
	RetryPolicy     *RetryPolicy  `json:"retry_policy,omitempty"`
	DeadLetterQueue string        `json:"dead_letter_queue,omitempty"`
	BatchSize       int           `json:"batch_size,omitempty"`
	BatchTimeout    time.Duration `json:"batch_timeout,omitempty"`
}

// DeliveryMode defines the event delivery mode
type DeliveryMode string

const (
	DeliveryModeSync  DeliveryMode = "sync"
	DeliveryModeAsync DeliveryMode = "async"
	DeliveryModeBatch DeliveryMode = "batch"
)

// TimeRange represents a time range for queries
type TimeRange struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// Data Flow Types

// DataFlow represents a data flow configuration
type DataFlow struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Source      *DataFlowNode          `json:"source"`
	Target      *DataFlowNode          `json:"target"`
	Steps       []*DataFlowStep        `json:"steps"`
	Config      *DataFlowConfig        `json:"config"`
	Schedule    *DataFlowSchedule      `json:"schedule,omitempty"`
	Status      DataFlowStatus         `json:"status"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	LastRunAt   *time.Time             `json:"last_run_at,omitempty"`
}

// DataFlowStatus defines the status of a data flow
type DataFlowStatus string

const (
	DataFlowStatusActive    DataFlowStatus = "active"
	DataFlowStatusInactive  DataFlowStatus = "inactive"
	DataFlowStatusRunning   DataFlowStatus = "running"
	DataFlowStatusCompleted DataFlowStatus = "completed"
	DataFlowStatusFailed    DataFlowStatus = "failed"
	DataFlowStatusPaused    DataFlowStatus = "paused"
)

// DataFlowNode represents a node in a data flow
type DataFlowNode struct {
	ID        string                 `json:"id"`
	Type      DataFlowNodeType       `json:"type"`
	ServiceID string                 `json:"service_id"`
	Endpoint  string                 `json:"endpoint"`
	Config    map[string]interface{} `json:"config,omitempty"`
	Auth      *EndpointAuth          `json:"auth,omitempty"`
}

// DataFlowNodeType defines the type of data flow node
type DataFlowNodeType string

const (
	DataFlowNodeTypeSource    DataFlowNodeType = "source"
	DataFlowNodeTypeTarget    DataFlowNodeType = "target"
	DataFlowNodeTypeTransform DataFlowNodeType = "transform"
	DataFlowNodeTypeFilter    DataFlowNodeType = "filter"
	DataFlowNodeTypeAggregate DataFlowNodeType = "aggregate"
	DataFlowNodeTypeValidate  DataFlowNodeType = "validate"
	DataFlowNodeTypeEnrich    DataFlowNodeType = "enrich"
)

// DataFlowStep represents a step in a data flow
type DataFlowStep struct {
	ID        string                 `json:"id"`
	Name      string                 `json:"name"`
	Type      DataFlowStepType       `json:"type"`
	Config    map[string]interface{} `json:"config"`
	Order     int                    `json:"order"`
	Enabled   bool                   `json:"enabled"`
	Condition string                 `json:"condition,omitempty"`
	OnError   ErrorHandlingStrategy  `json:"on_error"`
	Timeout   time.Duration          `json:"timeout,omitempty"`
}

// DataFlowStepType defines the type of data flow step
type DataFlowStepType string

const (
	DataFlowStepTypeExtract   DataFlowStepType = "extract"
	DataFlowStepTypeTransform DataFlowStepType = "transform"
	DataFlowStepTypeLoad      DataFlowStepType = "load"
	DataFlowStepTypeValidate  DataFlowStepType = "validate"
	DataFlowStepTypeEnrich    DataFlowStepType = "enrich"
	DataFlowStepTypeAggregate DataFlowStepType = "aggregate"
	DataFlowStepTypeFilter    DataFlowStepType = "filter"
	DataFlowStepTypeCustom    DataFlowStepType = "custom"
)

// ErrorHandlingStrategy defines how to handle errors
type ErrorHandlingStrategy string

const (
	ErrorHandlingStop       ErrorHandlingStrategy = "stop"
	ErrorHandlingSkip       ErrorHandlingStrategy = "skip"
	ErrorHandlingRetry      ErrorHandlingStrategy = "retry"
	ErrorHandlingDeadLetter ErrorHandlingStrategy = "dead_letter"
)

// DataFlowConfig contains data flow configuration
type DataFlowConfig struct {
	BatchSize     int                    `json:"batch_size"`
	Parallelism   int                    `json:"parallelism"`
	Timeout       time.Duration          `json:"timeout"`
	RetryPolicy   *RetryPolicy           `json:"retry_policy,omitempty"`
	ErrorHandling ErrorHandlingStrategy  `json:"error_handling"`
	Monitoring    *DataFlowMonitoring    `json:"monitoring,omitempty"`
	Custom        map[string]interface{} `json:"custom,omitempty"`
}

// DataFlowMonitoring contains monitoring configuration for data flows
type DataFlowMonitoring struct {
	Enabled         bool             `json:"enabled"`
	MetricsInterval time.Duration    `json:"metrics_interval"`
	AlertThresholds *AlertThresholds `json:"alert_thresholds,omitempty"`
	Notifications   []string         `json:"notifications,omitempty"`
}

// AlertThresholds defines alert thresholds for data flows
type AlertThresholds struct {
	ErrorRate      float64       `json:"error_rate"`
	ProcessingTime time.Duration `json:"processing_time"`
	ThroughputMin  int           `json:"throughput_min"`
	MemoryUsage    float64       `json:"memory_usage"`
	CPUUsage       float64       `json:"cpu_usage"`
}

// DataFlowSchedule contains scheduling configuration
type DataFlowSchedule struct {
	Type       ScheduleType `json:"type"`
	Expression string       `json:"expression"`
	Timezone   string       `json:"timezone,omitempty"`
	StartTime  *time.Time   `json:"start_time,omitempty"`
	EndTime    *time.Time   `json:"end_time,omitempty"`
	Enabled    bool         `json:"enabled"`
}

// ScheduleType defines the type of schedule
type ScheduleType string

const (
	ScheduleTypeCron     ScheduleType = "cron"
	ScheduleTypeInterval ScheduleType = "interval"
	ScheduleTypeOnce     ScheduleType = "once"
	ScheduleTypeEvent    ScheduleType = "event"
)

// Security Types

// AuthRequest represents an authentication request
type AuthRequest struct {
	Token      string                 `json:"token,omitempty"`
	Username   string                 `json:"username,omitempty"`
	Password   string                 `json:"password,omitempty"`
	Headers    map[string]string      `json:"headers,omitempty"`
	Method     string                 `json:"method"`
	Path       string                 `json:"path"`
	RemoteAddr string                 `json:"remote_addr,omitempty"`
	UserAgent  string                 `json:"user_agent,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// AuthResult represents the result of authentication
type AuthResult struct {
	Success     bool                   `json:"success"`
	UserID      string                 `json:"user_id,omitempty"`
	Username    string                 `json:"username,omitempty"`
	Roles       []string               `json:"roles,omitempty"`
	Permissions []string               `json:"permissions,omitempty"`
	Token       string                 `json:"token,omitempty"`
	ExpiresAt   *time.Time             `json:"expires_at,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	Error       string                 `json:"error,omitempty"`
}

// AuthzRequest represents an authorization request
type AuthzRequest struct {
	UserID   string                 `json:"user_id"`
	Resource string                 `json:"resource"`
	Action   string                 `json:"action"`
	Context  map[string]interface{} `json:"context,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// AuthzResult represents the result of authorization
type AuthzResult struct {
	Allowed  bool                   `json:"allowed"`
	Reason   string                 `json:"reason,omitempty"`
	Policies []string               `json:"policies,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// TokenValidation represents token validation result
type TokenValidation struct {
	Valid       bool                   `json:"valid"`
	UserID      string                 `json:"user_id,omitempty"`
	Username    string                 `json:"username,omitempty"`
	Roles       []string               `json:"roles,omitempty"`
	Permissions []string               `json:"permissions,omitempty"`
	ExpiresAt   *time.Time             `json:"expires_at,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	Error       string                 `json:"error,omitempty"`
}

// SecurityPolicy represents a security policy
type SecurityPolicy struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Type        PolicyType             `json:"type"`
	Rules       []*PolicyRule          `json:"rules"`
	Effect      PolicyEffect           `json:"effect"`
	Priority    int                    `json:"priority"`
	Enabled     bool                   `json:"enabled"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// PolicyType defines the type of security policy
type PolicyType string

const (
	PolicyTypeAccess     PolicyType = "access"
	PolicyTypeRate       PolicyType = "rate"
	PolicyTypeData       PolicyType = "data"
	PolicyTypeCompliance PolicyType = "compliance"
	PolicyTypeCustom     PolicyType = "custom"
)

// PolicyEffect defines the effect of a policy
type PolicyEffect string

const (
	PolicyEffectAllow PolicyEffect = "allow"
	PolicyEffectDeny  PolicyEffect = "deny"
)

// PolicyRule represents a rule within a security policy
type PolicyRule struct {
	ID        string                 `json:"id"`
	Condition string                 `json:"condition"`
	Action    string                 `json:"action"`
	Resource  string                 `json:"resource"`
	Effect    PolicyEffect           `json:"effect"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// PolicyContext represents the context for policy evaluation
type PolicyContext struct {
	UserID      string                 `json:"user_id,omitempty"`
	Roles       []string               `json:"roles,omitempty"`
	Resource    string                 `json:"resource"`
	Action      string                 `json:"action"`
	Environment map[string]interface{} `json:"environment,omitempty"`
	Request     map[string]interface{} `json:"request,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// PolicyResult represents the result of policy evaluation
type PolicyResult struct {
	Allowed  bool                   `json:"allowed"`
	Policies []string               `json:"policies"`
	Reason   string                 `json:"reason,omitempty"`
	Actions  []string               `json:"actions,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// Monitoring Types

// SystemHealth represents the overall system health
type SystemHealth struct {
	Status       HealthStatus           `json:"status"`
	Message      string                 `json:"message"`
	Services     []*ServiceHealth       `json:"services"`
	Dependencies []*DependencyHealth    `json:"dependencies"`
	LastCheck    time.Time              `json:"last_check"`
	Uptime       time.Duration          `json:"uptime"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// SystemMetrics contains system-wide metrics
type SystemMetrics struct {
	TimeRange        *TimeRange                 `json:"time_range"`
	RequestCount     int64                      `json:"request_count"`
	ErrorCount       int64                      `json:"error_count"`
	SuccessRate      float64                    `json:"success_rate"`
	AverageLatency   time.Duration              `json:"average_latency"`
	P95Latency       time.Duration              `json:"p95_latency"`
	P99Latency       time.Duration              `json:"p99_latency"`
	ThroughputPerSec float64                    `json:"throughput_per_sec"`
	ResourceUsage    *ResourceUsage             `json:"resource_usage"`
	ServiceMetrics   map[string]*ServiceMetrics `json:"service_metrics"`
	Metadata         map[string]interface{}     `json:"metadata,omitempty"`
}

// ServiceMetrics contains service-specific metrics
type ServiceMetrics struct {
	ServiceID        string                      `json:"service_id"`
	TimeRange        *TimeRange                  `json:"time_range"`
	RequestCount     int64                       `json:"request_count"`
	ErrorCount       int64                       `json:"error_count"`
	SuccessRate      float64                     `json:"success_rate"`
	AverageLatency   time.Duration               `json:"average_latency"`
	P95Latency       time.Duration               `json:"p95_latency"`
	P99Latency       time.Duration               `json:"p99_latency"`
	ThroughputPerSec float64                     `json:"throughput_per_sec"`
	ResourceUsage    *ResourceUsage              `json:"resource_usage"`
	EndpointMetrics  map[string]*EndpointMetrics `json:"endpoint_metrics"`
	Metadata         map[string]interface{}      `json:"metadata,omitempty"`
}

// EndpointMetrics contains endpoint-specific metrics
type EndpointMetrics struct {
	Endpoint         string        `json:"endpoint"`
	RequestCount     int64         `json:"request_count"`
	ErrorCount       int64         `json:"error_count"`
	SuccessRate      float64       `json:"success_rate"`
	AverageLatency   time.Duration `json:"average_latency"`
	P95Latency       time.Duration `json:"p95_latency"`
	P99Latency       time.Duration `json:"p99_latency"`
	ThroughputPerSec float64       `json:"throughput_per_sec"`
}

// ResourceUsage contains resource usage metrics
type ResourceUsage struct {
	CPUUsage    float64 `json:"cpu_usage"`
	MemoryUsage float64 `json:"memory_usage"`
	DiskUsage   float64 `json:"disk_usage"`
	NetworkIn   int64   `json:"network_in"`
	NetworkOut  int64   `json:"network_out"`
}

// Alert represents a system alert
type Alert struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Type        AlertType              `json:"type"`
	Severity    AlertSeverity          `json:"severity"`
	Status      AlertStatus            `json:"status"`
	Source      string                 `json:"source"`
	Target      string                 `json:"target,omitempty"`
	Condition   string                 `json:"condition"`
	Threshold   *AlertThreshold        `json:"threshold,omitempty"`
	Actions     []*AlertAction         `json:"actions"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	TriggeredAt *time.Time             `json:"triggered_at,omitempty"`
	ResolvedAt  *time.Time             `json:"resolved_at,omitempty"`
}

// AlertType defines the type of alert
type AlertType string

const (
	AlertTypeHealth      AlertType = "health"
	AlertTypePerformance AlertType = "performance"
	AlertTypeSecurity    AlertType = "security"
	AlertTypeCapacity    AlertType = "capacity"
	AlertTypeCustom      AlertType = "custom"
)

// AlertSeverity defines the severity of an alert
type AlertSeverity string

const (
	AlertSeverityLow      AlertSeverity = "low"
	AlertSeverityMedium   AlertSeverity = "medium"
	AlertSeverityHigh     AlertSeverity = "high"
	AlertSeverityCritical AlertSeverity = "critical"
)

// AlertStatus defines the status of an alert
type AlertStatus string

const (
	AlertStatusActive     AlertStatus = "active"
	AlertStatusTriggered  AlertStatus = "triggered"
	AlertStatusResolved   AlertStatus = "resolved"
	AlertStatusSuppressed AlertStatus = "suppressed"
)

// AlertThreshold defines alert threshold configuration
type AlertThreshold struct {
	Metric    string        `json:"metric"`
	Operator  string        `json:"operator"`
	Value     float64       `json:"value"`
	Duration  time.Duration `json:"duration,omitempty"`
	Frequency time.Duration `json:"frequency,omitempty"`
}

// AlertAction defines an action to take when an alert is triggered
type AlertAction struct {
	Type    AlertActionType        `json:"type"`
	Target  string                 `json:"target"`
	Config  map[string]interface{} `json:"config,omitempty"`
	Enabled bool                   `json:"enabled"`
}

// AlertActionType defines the type of alert action
type AlertActionType string

const (
	AlertActionTypeEmail     AlertActionType = "email"
	AlertActionTypeSlack     AlertActionType = "slack"
	AlertActionTypeWebhook   AlertActionType = "webhook"
	AlertActionTypePagerDuty AlertActionType = "pagerduty"
	AlertActionTypeAutoScale AlertActionType = "autoscale"
	AlertActionTypeRestart   AlertActionType = "restart"
	AlertActionTypeCustom    AlertActionType = "custom"
)

// Configuration Types

// Configuration represents a configuration entry
type Configuration struct {
	Key         string                 `json:"key"`
	Value       interface{}            `json:"value"`
	Type        ConfigurationType      `json:"type"`
	Description string                 `json:"description,omitempty"`
	Sensitive   bool                   `json:"sensitive"`
	Encrypted   bool                   `json:"encrypted"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// ConfigurationType defines the type of configuration
type ConfigurationType string

const (
	ConfigurationTypeString  ConfigurationType = "string"
	ConfigurationTypeInteger ConfigurationType = "integer"
	ConfigurationTypeFloat   ConfigurationType = "float"
	ConfigurationTypeBoolean ConfigurationType = "boolean"
	ConfigurationTypeJSON    ConfigurationType = "json"
	ConfigurationTypeSecret  ConfigurationType = "secret"
)

// ServiceConfiguration represents service-specific configuration
type ServiceConfiguration struct {
	ServiceID string                    `json:"service_id"`
	Config    map[string]*Configuration `json:"config"`
	Version   string                    `json:"version"`
	UpdatedAt time.Time                 `json:"updated_at"`
}

// Deployment Types

// Deployment represents a deployment configuration
type Deployment struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	ServiceID   string                 `json:"service_id"`
	Version     string                 `json:"version"`
	Strategy    DeploymentStrategy     `json:"strategy"`
	Config      *DeploymentConfig      `json:"config"`
	Status      DeploymentStatus       `json:"status"`
	Progress    *DeploymentProgress    `json:"progress,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	StartedAt   *time.Time             `json:"started_at,omitempty"`
	CompletedAt *time.Time             `json:"completed_at,omitempty"`
}

// DeploymentStrategy defines the deployment strategy
type DeploymentStrategy string

const (
	DeploymentStrategyRolling   DeploymentStrategy = "rolling"
	DeploymentStrategyBlueGreen DeploymentStrategy = "blue_green"
	DeploymentStrategyCanary    DeploymentStrategy = "canary"
	DeploymentStrategyRecreate  DeploymentStrategy = "recreate"
)

// DeploymentStatus defines the status of a deployment
type DeploymentStatus string

const (
	DeploymentStatusPending    DeploymentStatus = "pending"
	DeploymentStatusRunning    DeploymentStatus = "running"
	DeploymentStatusCompleted  DeploymentStatus = "completed"
	DeploymentStatusFailed     DeploymentStatus = "failed"
	DeploymentStatusRolledBack DeploymentStatus = "rolled_back"
	DeploymentStatusCancelled  DeploymentStatus = "cancelled"
)

// DeploymentConfig contains deployment configuration
type DeploymentConfig struct {
	Replicas        int                    `json:"replicas"`
	MaxUnavailable  int                    `json:"max_unavailable"`
	MaxSurge        int                    `json:"max_surge"`
	ProgressTimeout time.Duration          `json:"progress_timeout"`
	HealthCheck     *DeploymentHealthCheck `json:"health_check,omitempty"`
	Rollback        *RollbackConfig        `json:"rollback,omitempty"`
	Custom          map[string]interface{} `json:"custom,omitempty"`
}

// DeploymentHealthCheck contains health check configuration for deployments
type DeploymentHealthCheck struct {
	Enabled          bool          `json:"enabled"`
	Path             string        `json:"path"`
	Port             int           `json:"port"`
	InitialDelay     time.Duration `json:"initial_delay"`
	Interval         time.Duration `json:"interval"`
	Timeout          time.Duration `json:"timeout"`
	SuccessThreshold int           `json:"success_threshold"`
	FailureThreshold int           `json:"failure_threshold"`
}

// RollbackConfig contains rollback configuration
type RollbackConfig struct {
	Enabled          bool          `json:"enabled"`
	AutoRollback     bool          `json:"auto_rollback"`
	FailureThreshold float64       `json:"failure_threshold"`
	MonitoringWindow time.Duration `json:"monitoring_window"`
}

// DeploymentProgress contains deployment progress information
type DeploymentProgress struct {
	TotalSteps     int       `json:"total_steps"`
	CompletedSteps int       `json:"completed_steps"`
	CurrentStep    string    `json:"current_step"`
	Percentage     float64   `json:"percentage"`
	Message        string    `json:"message,omitempty"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// Filter Types

// ServiceFilter contains filters for service queries
type ServiceFilter struct {
	Type   ServiceType   `json:"type,omitempty"`
	Status ServiceStatus `json:"status,omitempty"`
	Tags   []string      `json:"tags,omitempty"`
	Search string        `json:"search,omitempty"`
	Limit  int           `json:"limit,omitempty"`
	Offset int           `json:"offset,omitempty"`
}

// RouteFilter contains filters for route queries
type RouteFilter struct {
	ServiceID string      `json:"service_id,omitempty"`
	Method    string      `json:"method,omitempty"`
	Status    RouteStatus `json:"status,omitempty"`
	Search    string      `json:"search,omitempty"`
	Limit     int         `json:"limit,omitempty"`
	Offset    int         `json:"offset,omitempty"`
}

// EventFilter contains filters for event queries
type EventFilter struct {
	Type     string        `json:"type,omitempty"`
	Source   string        `json:"source,omitempty"`
	Target   string        `json:"target,omitempty"`
	Priority EventPriority `json:"priority,omitempty"`
	Since    *time.Time    `json:"since,omitempty"`
	Until    *time.Time    `json:"until,omitempty"`
	Search   string        `json:"search,omitempty"`
	Limit    int           `json:"limit,omitempty"`
	Offset   int           `json:"offset,omitempty"`
}

// AlertFilter contains filters for alert queries
type AlertFilter struct {
	Type     AlertType     `json:"type,omitempty"`
	Severity AlertSeverity `json:"severity,omitempty"`
	Status   AlertStatus   `json:"status,omitempty"`
	Source   string        `json:"source,omitempty"`
	Since    *time.Time    `json:"since,omitempty"`
	Until    *time.Time    `json:"until,omitempty"`
	Search   string        `json:"search,omitempty"`
	Limit    int           `json:"limit,omitempty"`
	Offset   int           `json:"offset,omitempty"`
}

// DiscoveryCriteria contains criteria for service discovery
type DiscoveryCriteria struct {
	Type         ServiceType            `json:"type,omitempty"`
	Tags         []string               `json:"tags,omitempty"`
	HealthStatus HealthStatus           `json:"health_status,omitempty"`
	Version      string                 `json:"version,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}
