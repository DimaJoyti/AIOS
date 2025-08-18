package deployment

import (
	"time"
)

// AIOS Deployment Management Types and Configurations

// DeploymentManager provides comprehensive deployment management capabilities
type DeploymentManager interface {
	// Deployment Management
	CreateDeployment(deployment *Deployment) (*Deployment, error)
	GetDeployment(deploymentID string) (*Deployment, error)
	UpdateDeployment(deployment *Deployment) (*Deployment, error)
	DeleteDeployment(deploymentID string) error
	ListDeployments(filter *DeploymentFilter) ([]*Deployment, error)

	// Deployment Execution
	StartDeployment(deploymentID string) (*Deployment, error)
	StopDeployment(executionID string) error
	GetDeploymentExecution(executionID string) (*Deployment, error)
	ListDeploymentExecutions(filter *DeploymentFilter) ([]*Deployment, error)

	// Environment Management
	CreateEnvironment(environment *Environment) (*Environment, error)
	GetEnvironment(environmentID string) (*Environment, error)
	UpdateEnvironment(environment *Environment) (*Environment, error)
	DeleteEnvironment(environmentID string) error
	ListEnvironments(filter *EnvironmentFilter) ([]*Environment, error)

	// Pipeline Management
	CreatePipeline(pipeline *DeploymentConfig) (*DeploymentConfig, error)
	GetPipeline(pipelineID string) (*DeploymentConfig, error)
	UpdatePipeline(pipeline *DeploymentConfig) (*DeploymentConfig, error)
	DeletePipeline(pipelineID string) error
	ExecutePipeline(pipelineID string, config *DeploymentConfig) (*Deployment, error)

	// Rollback Management
	CreateRollback(rollback *Deployment) (*Deployment, error)
	ExecuteRollback(rollbackID string) (*Deployment, error)
	GetRollbackExecution(executionID string) (*Deployment, error)

	// Health and Monitoring
	GetDeploymentHealth(deploymentID string) (*HealthCheck, error)
	GetEnvironmentHealth(environmentID string) (*HealthCheck, error)
	GetDeploymentMetrics(deploymentID string, timeRange *TimeRange) (*PerformanceMonitoring, error)
}

// Core Deployment Types

// Deployment represents a deployment configuration
type Deployment struct {
	ID            string                  `json:"id"`
	Name          string                  `json:"name"`
	Description   string                  `json:"description"`
	ApplicationID string                  `json:"application_id"`
	Version       string                  `json:"version"`
	EnvironmentID string                  `json:"environment_id"`
	Strategy      DeploymentStrategy      `json:"strategy"`
	Config        *DeploymentConfig       `json:"config"`
	Artifacts     []*DeploymentArtifact   `json:"artifacts"`
	Dependencies  []*DeploymentDependency `json:"dependencies,omitempty"`
	HealthChecks  []*HealthCheck          `json:"health_checks,omitempty"`
	Notifications *NotificationConfig     `json:"notifications,omitempty"`
	Status        DeploymentStatus        `json:"status"`
	Tags          []string                `json:"tags,omitempty"`
	Metadata      map[string]interface{}  `json:"metadata,omitempty"`
	CreatedAt     time.Time               `json:"created_at"`
	UpdatedAt     time.Time               `json:"updated_at"`
	CreatedBy     string                  `json:"created_by"`
}

// DeploymentStrategy defines the deployment strategy
type DeploymentStrategy string

const (
	DeploymentStrategyRolling   DeploymentStrategy = "rolling"
	DeploymentStrategyBlueGreen DeploymentStrategy = "blue_green"
	DeploymentStrategyCanary    DeploymentStrategy = "canary"
	DeploymentStrategyRecreate  DeploymentStrategy = "recreate"
	DeploymentStrategyA_B       DeploymentStrategy = "a_b"
	DeploymentStrategyCustom    DeploymentStrategy = "custom"
)

// DeploymentStatus defines the status of a deployment
type DeploymentStatus string

const (
	DeploymentStatusDraft       DeploymentStatus = "draft"
	DeploymentStatusReady       DeploymentStatus = "ready"
	DeploymentStatusDeploying   DeploymentStatus = "deploying"
	DeploymentStatusDeployed    DeploymentStatus = "deployed"
	DeploymentStatusFailed      DeploymentStatus = "failed"
	DeploymentStatusRollingBack DeploymentStatus = "rolling_back"
	DeploymentStatusRolledBack  DeploymentStatus = "rolled_back"
	DeploymentStatusArchived    DeploymentStatus = "archived"
)

// DeploymentConfig contains deployment configuration
type DeploymentConfig struct {
	Replicas            int                    `json:"replicas"`
	MaxUnavailable      int                    `json:"max_unavailable"`
	MaxSurge            int                    `json:"max_surge"`
	ProgressTimeout     time.Duration          `json:"progress_timeout"`
	RollbackTimeout     time.Duration          `json:"rollback_timeout"`
	HealthCheckTimeout  time.Duration          `json:"health_check_timeout"`
	PreDeploymentSteps  []*DeploymentStep      `json:"pre_deployment_steps,omitempty"`
	PostDeploymentSteps []*DeploymentStep      `json:"post_deployment_steps,omitempty"`
	RollbackSteps       []*DeploymentStep      `json:"rollback_steps,omitempty"`
	Resources           *ResourceRequirements  `json:"resources,omitempty"`
	Scaling             *ScalingConfig         `json:"scaling,omitempty"`
	Security            *SecurityConfig        `json:"security,omitempty"`
	Networking          *NetworkingConfig      `json:"networking,omitempty"`
	Storage             *StorageConfig         `json:"storage,omitempty"`
	Environment         map[string]string      `json:"environment,omitempty"`
	Secrets             map[string]string      `json:"secrets,omitempty"`
	ConfigMaps          map[string]interface{} `json:"config_maps,omitempty"`
	Custom              map[string]interface{} `json:"custom,omitempty"`
}

// DeploymentArtifact represents a deployment artifact
type DeploymentArtifact struct {
	ID         string                 `json:"id"`
	Name       string                 `json:"name"`
	Type       ArtifactType           `json:"type"`
	Source     string                 `json:"source"`
	Version    string                 `json:"version"`
	Checksum   string                 `json:"checksum"`
	Size       int64                  `json:"size"`
	Registry   string                 `json:"registry,omitempty"`
	Repository string                 `json:"repository,omitempty"`
	Tag        string                 `json:"tag,omitempty"`
	Path       string                 `json:"path,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt  time.Time              `json:"created_at"`
}

// ArtifactType defines the type of deployment artifact
type ArtifactType string

const (
	ArtifactTypeContainer ArtifactType = "container"
	ArtifactTypeBinary    ArtifactType = "binary"
	ArtifactTypePackage   ArtifactType = "package"
	ArtifactTypeArchive   ArtifactType = "archive"
	ArtifactTypeConfig    ArtifactType = "config"
	ArtifactTypeCustom    ArtifactType = "custom"
)

// DeploymentDependency represents a deployment dependency
type DeploymentDependency struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Type        DependencyType         `json:"type"`
	Target      string                 `json:"target"`
	Version     string                 `json:"version,omitempty"`
	Required    bool                   `json:"required"`
	HealthCheck bool                   `json:"health_check"`
	Timeout     time.Duration          `json:"timeout"`
	RetryPolicy *RetryPolicy           `json:"retry_policy,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// DependencyType defines the type of deployment dependency
type DependencyType string

const (
	DependencyTypeService  DependencyType = "service"
	DependencyTypeDatabase DependencyType = "database"
	DependencyTypeQueue    DependencyType = "queue"
	DependencyTypeCache    DependencyType = "cache"
	DependencyTypeStorage  DependencyType = "storage"
	DependencyTypeExternal DependencyType = "external"
	DependencyTypeCustom   DependencyType = "custom"
)

// DeploymentStep represents a deployment step
type DeploymentStep struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Type        DeploymentStepType     `json:"type"`
	Command     string                 `json:"command,omitempty"`
	Script      string                 `json:"script,omitempty"`
	Parameters  map[string]interface{} `json:"parameters,omitempty"`
	Timeout     time.Duration          `json:"timeout"`
	RetryPolicy *RetryPolicy           `json:"retry_policy,omitempty"`
	OnFailure   StepFailureAction      `json:"on_failure"`
	Order       int                    `json:"order"`
	Enabled     bool                   `json:"enabled"`
	Condition   string                 `json:"condition,omitempty"`
}

// DeploymentStepType defines the type of deployment step
type DeploymentStepType string

const (
	DeploymentStepTypeCommand      DeploymentStepType = "command"
	DeploymentStepTypeScript       DeploymentStepType = "script"
	DeploymentStepTypeHealthCheck  DeploymentStepType = "health_check"
	DeploymentStepTypeNotification DeploymentStepType = "notification"
	DeploymentStepTypeWait         DeploymentStepType = "wait"
	DeploymentStepTypeCustom       DeploymentStepType = "custom"
)

// StepFailureAction defines what to do when a step fails
type StepFailureAction string

const (
	StepFailureActionStop     StepFailureAction = "stop"
	StepFailureActionContinue StepFailureAction = "continue"
	StepFailureActionRetry    StepFailureAction = "retry"
	StepFailureActionRollback StepFailureAction = "rollback"
)

// Environment Types

// Environment represents a deployment environment
type Environment struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Type        EnvironmentType        `json:"type"`
	Status      EnvironmentStatus      `json:"status"`
	Config      *EnvironmentConfig     `json:"config"`
	Resources   *ResourceRequirements  `json:"resources,omitempty"`
	Security    *SecurityConfig        `json:"security,omitempty"`
	Networking  *NetworkingConfig      `json:"networking,omitempty"`
	Monitoring  *PerformanceMonitoring `json:"monitoring,omitempty"`
	Tags        []string               `json:"tags,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	CreatedBy   string                 `json:"created_by"`
}

// EnvironmentType defines the type of environment
type EnvironmentType string

const (
	EnvironmentTypeDevelopment EnvironmentType = "development"
	EnvironmentTypeStaging     EnvironmentType = "staging"
	EnvironmentTypeProduction  EnvironmentType = "production"
	EnvironmentTypeTesting     EnvironmentType = "testing"
	EnvironmentTypeDemo        EnvironmentType = "demo"
	EnvironmentTypeCustom      EnvironmentType = "custom"
)

// EnvironmentStatus defines the status of an environment
type EnvironmentStatus string

const (
	EnvironmentStatusActive      EnvironmentStatus = "active"
	EnvironmentStatusInactive    EnvironmentStatus = "inactive"
	EnvironmentStatusMaintenance EnvironmentStatus = "maintenance"
	EnvironmentStatusError       EnvironmentStatus = "error"
)

// EnvironmentConfig contains environment configuration
type EnvironmentConfig struct {
	Region            string                 `json:"region"`
	Zone              string                 `json:"zone,omitempty"`
	Provider          string                 `json:"provider"`
	Cluster           string                 `json:"cluster,omitempty"`
	Namespace         string                 `json:"namespace,omitempty"`
	AutoScaling       bool                   `json:"auto_scaling"`
	LoadBalancing     bool                   `json:"load_balancing"`
	HighAvailability  bool                   `json:"high_availability"`
	BackupEnabled     bool                   `json:"backup_enabled"`
	MonitoringEnabled bool                   `json:"monitoring_enabled"`
	LoggingEnabled    bool                   `json:"logging_enabled"`
	Environment       map[string]string      `json:"environment,omitempty"`
	Secrets           map[string]string      `json:"secrets,omitempty"`
	Custom            map[string]interface{} `json:"custom,omitempty"`
}

// Configuration Types

// ResourceRequirements defines resource requirements
type ResourceRequirements struct {
	CPU    *ResourceSpec `json:"cpu,omitempty"`
	Memory *ResourceSpec `json:"memory,omitempty"`
	Disk   *ResourceSpec `json:"disk,omitempty"`
	GPU    *ResourceSpec `json:"gpu,omitempty"`
}

// ResourceSpec defines a resource specification
type ResourceSpec struct {
	Requests string `json:"requests,omitempty"`
	Limits   string `json:"limits,omitempty"`
}

// ScalingConfig defines scaling configuration
type ScalingConfig struct {
	Enabled     bool                   `json:"enabled"`
	MinReplicas int                    `json:"min_replicas"`
	MaxReplicas int                    `json:"max_replicas"`
	Metrics     []*ScalingMetric       `json:"metrics,omitempty"`
	Behavior    *ScalingBehavior       `json:"behavior,omitempty"`
	Custom      map[string]interface{} `json:"custom,omitempty"`
}

// ScalingMetric defines a scaling metric
type ScalingMetric struct {
	Type      ScalingMetricType `json:"type"`
	Name      string            `json:"name"`
	Target    float64           `json:"target"`
	Threshold float64           `json:"threshold,omitempty"`
}

// ScalingMetricType defines the type of scaling metric
type ScalingMetricType string

const (
	ScalingMetricTypeCPU    ScalingMetricType = "cpu"
	ScalingMetricTypeMemory ScalingMetricType = "memory"
	ScalingMetricTypeCustom ScalingMetricType = "custom"
)

// ScalingBehavior defines scaling behavior
type ScalingBehavior struct {
	ScaleUp   *ScalingPolicy `json:"scale_up,omitempty"`
	ScaleDown *ScalingPolicy `json:"scale_down,omitempty"`
}

// ScalingPolicy defines a scaling policy
type ScalingPolicy struct {
	StabilizationWindow time.Duration  `json:"stabilization_window"`
	SelectPolicy        string         `json:"select_policy"`
	Policies            []*ScalingRule `json:"policies,omitempty"`
}

// ScalingRule defines a scaling rule
type ScalingRule struct {
	Type          string        `json:"type"`
	Value         int           `json:"value"`
	PeriodSeconds time.Duration `json:"period_seconds"`
}

// SecurityConfig defines security configuration
type SecurityConfig struct {
	RunAsUser                int64                 `json:"run_as_user,omitempty"`
	RunAsGroup               int64                 `json:"run_as_group,omitempty"`
	RunAsNonRoot             bool                  `json:"run_as_non_root"`
	ReadOnlyRootFS           bool                  `json:"read_only_root_fs"`
	AllowPrivilegeEscalation bool                  `json:"allow_privilege_escalation"`
	Capabilities             *SecurityCapabilities `json:"capabilities,omitempty"`
	SELinux                  *SELinuxOptions       `json:"selinux,omitempty"`
	SeccompProfile           *SeccompProfile       `json:"seccomp_profile,omitempty"`
	AppArmorProfile          *AppArmorProfile      `json:"apparmor_profile,omitempty"`
}

// SecurityCapabilities defines security capabilities
type SecurityCapabilities struct {
	Add  []string `json:"add,omitempty"`
	Drop []string `json:"drop,omitempty"`
}

// SELinuxOptions defines SELinux options
type SELinuxOptions struct {
	User  string `json:"user,omitempty"`
	Role  string `json:"role,omitempty"`
	Type  string `json:"type,omitempty"`
	Level string `json:"level,omitempty"`
}

// SeccompProfile defines seccomp profile
type SeccompProfile struct {
	Type             string `json:"type"`
	LocalhostProfile string `json:"localhost_profile,omitempty"`
}

// AppArmorProfile defines AppArmor profile
type AppArmorProfile struct {
	Type             string `json:"type"`
	LocalhostProfile string `json:"localhost_profile,omitempty"`
}

// NetworkingConfig defines networking configuration
type NetworkingConfig struct {
	Ports           []*Port                `json:"ports,omitempty"`
	ServiceType     ServiceType            `json:"service_type,omitempty"`
	LoadBalancer    *LoadBalancerConfig    `json:"load_balancer,omitempty"`
	Ingress         *IngressConfig         `json:"ingress,omitempty"`
	NetworkPolicies []*NetworkPolicy       `json:"network_policies,omitempty"`
	DNS             *DNSConfig             `json:"dns,omitempty"`
	Custom          map[string]interface{} `json:"custom,omitempty"`
}

// Port defines a network port
type Port struct {
	Name       string   `json:"name,omitempty"`
	Port       int      `json:"port"`
	TargetPort int      `json:"target_port,omitempty"`
	Protocol   Protocol `json:"protocol"`
}

// Protocol defines network protocol
type Protocol string

const (
	ProtocolTCP  Protocol = "TCP"
	ProtocolUDP  Protocol = "UDP"
	ProtocolSCTP Protocol = "SCTP"
)

// ServiceType defines service type
type ServiceType string

const (
	ServiceTypeClusterIP    ServiceType = "ClusterIP"
	ServiceTypeNodePort     ServiceType = "NodePort"
	ServiceTypeLoadBalancer ServiceType = "LoadBalancer"
	ServiceTypeExternalName ServiceType = "ExternalName"
)

// LoadBalancerConfig defines load balancer configuration
type LoadBalancerConfig struct {
	Type        LoadBalancerType         `json:"type"`
	Algorithm   LoadBalancerAlgorithm    `json:"algorithm"`
	HealthCheck *LoadBalancerHealthCheck `json:"health_check,omitempty"`
	Sticky      bool                     `json:"sticky"`
	Timeout     time.Duration            `json:"timeout"`
	Custom      map[string]interface{}   `json:"custom,omitempty"`
}

// LoadBalancerType defines load balancer type
type LoadBalancerType string

const (
	LoadBalancerTypeApplication LoadBalancerType = "application"
	LoadBalancerTypeNetwork     LoadBalancerType = "network"
	LoadBalancerTypeClassic     LoadBalancerType = "classic"
)

// LoadBalancerAlgorithm defines load balancer algorithm
type LoadBalancerAlgorithm string

const (
	LoadBalancerAlgorithmRoundRobin LoadBalancerAlgorithm = "round_robin"
	LoadBalancerAlgorithmLeastConn  LoadBalancerAlgorithm = "least_conn"
	LoadBalancerAlgorithmIPHash     LoadBalancerAlgorithm = "ip_hash"
	LoadBalancerAlgorithmWeighted   LoadBalancerAlgorithm = "weighted"
)

// LoadBalancerHealthCheck defines load balancer health check
type LoadBalancerHealthCheck struct {
	Path               string        `json:"path"`
	Port               int           `json:"port"`
	Protocol           Protocol      `json:"protocol"`
	Interval           time.Duration `json:"interval"`
	Timeout            time.Duration `json:"timeout"`
	HealthyThreshold   int           `json:"healthy_threshold"`
	UnhealthyThreshold int           `json:"unhealthy_threshold"`
}

// IngressConfig defines ingress configuration
type IngressConfig struct {
	Enabled     bool                   `json:"enabled"`
	Class       string                 `json:"class,omitempty"`
	Host        string                 `json:"host,omitempty"`
	Paths       []*IngressPath         `json:"paths,omitempty"`
	TLS         *IngressTLS            `json:"tls,omitempty"`
	Annotations map[string]string      `json:"annotations,omitempty"`
	Custom      map[string]interface{} `json:"custom,omitempty"`
}

// IngressPath defines an ingress path
type IngressPath struct {
	Path        string   `json:"path"`
	PathType    PathType `json:"path_type"`
	ServiceName string   `json:"service_name"`
	ServicePort int      `json:"service_port"`
}

// PathType defines ingress path type
type PathType string

const (
	PathTypeExact                  PathType = "Exact"
	PathTypePrefix                 PathType = "Prefix"
	PathTypeImplementationSpecific PathType = "ImplementationSpecific"
)

// IngressTLS defines ingress TLS configuration
type IngressTLS struct {
	Enabled    bool     `json:"enabled"`
	SecretName string   `json:"secret_name,omitempty"`
	Hosts      []string `json:"hosts,omitempty"`
}

// NetworkPolicy defines a network policy
type NetworkPolicy struct {
	Name     string                 `json:"name"`
	Ingress  []*NetworkPolicyRule   `json:"ingress,omitempty"`
	Egress   []*NetworkPolicyRule   `json:"egress,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// NetworkPolicyRule defines a network policy rule
type NetworkPolicyRule struct {
	Ports []*NetworkPolicyPort `json:"ports,omitempty"`
	From  []*NetworkPolicyPeer `json:"from,omitempty"`
	To    []*NetworkPolicyPeer `json:"to,omitempty"`
}

// NetworkPolicyPort defines a network policy port
type NetworkPolicyPort struct {
	Protocol Protocol `json:"protocol,omitempty"`
	Port     int      `json:"port,omitempty"`
}

// NetworkPolicyPeer defines a network policy peer
type NetworkPolicyPeer struct {
	PodSelector       map[string]string `json:"pod_selector,omitempty"`
	NamespaceSelector map[string]string `json:"namespace_selector,omitempty"`
	IPBlock           *IPBlock          `json:"ip_block,omitempty"`
}

// IPBlock defines an IP block
type IPBlock struct {
	CIDR   string   `json:"cidr"`
	Except []string `json:"except,omitempty"`
}

// DNSConfig defines DNS configuration
type DNSConfig struct {
	Policy DNSPolicy              `json:"policy,omitempty"`
	Config *DNSConfigOptions      `json:"config,omitempty"`
	Custom map[string]interface{} `json:"custom,omitempty"`
}

// DNSPolicy defines DNS policy
type DNSPolicy string

const (
	DNSPolicyClusterFirst            DNSPolicy = "ClusterFirst"
	DNSPolicyClusterFirstWithHostNet DNSPolicy = "ClusterFirstWithHostNet"
	DNSPolicyDefault                 DNSPolicy = "Default"
	DNSPolicyNone                    DNSPolicy = "None"
)

// DNSConfigOptions defines DNS config options
type DNSConfigOptions struct {
	Nameservers []string           `json:"nameservers,omitempty"`
	Searches    []string           `json:"searches,omitempty"`
	Options     []*DNSConfigOption `json:"options,omitempty"`
}

// DNSConfigOption defines a DNS config option
type DNSConfigOption struct {
	Name  string `json:"name"`
	Value string `json:"value,omitempty"`
}

// StorageConfig defines storage configuration
type StorageConfig struct {
	Volumes      []*Volume              `json:"volumes,omitempty"`
	VolumeMounts []*VolumeMount         `json:"volume_mounts,omitempty"`
	Custom       map[string]interface{} `json:"custom,omitempty"`
}

// Volume defines a storage volume
type Volume struct {
	Name         string                 `json:"name"`
	Type         VolumeType             `json:"type"`
	Size         string                 `json:"size,omitempty"`
	StorageClass string                 `json:"storage_class,omitempty"`
	AccessModes  []AccessMode           `json:"access_modes,omitempty"`
	Source       map[string]interface{} `json:"source,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// VolumeType defines volume type
type VolumeType string

const (
	VolumeTypePersistentVolumeClaim VolumeType = "persistent_volume_claim"
	VolumeTypeEmptyDir              VolumeType = "empty_dir"
	VolumeTypeHostPath              VolumeType = "host_path"
	VolumeTypeConfigMap             VolumeType = "config_map"
	VolumeTypeSecret                VolumeType = "secret"
	VolumeTypeCustom                VolumeType = "custom"
)

// AccessMode defines volume access mode
type AccessMode string

const (
	AccessModeReadWriteOnce AccessMode = "ReadWriteOnce"
	AccessModeReadOnlyMany  AccessMode = "ReadOnlyMany"
	AccessModeReadWriteMany AccessMode = "ReadWriteMany"
)

// VolumeMount defines a volume mount
type VolumeMount struct {
	Name      string `json:"name"`
	MountPath string `json:"mount_path"`
	SubPath   string `json:"sub_path,omitempty"`
	ReadOnly  bool   `json:"read_only"`
}

// Supporting Types

// RetryPolicy defines retry behavior
type RetryPolicy struct {
	MaxRetries      int           `json:"max_retries"`
	InitialDelay    time.Duration `json:"initial_delay"`
	MaxDelay        time.Duration `json:"max_delay"`
	BackoffFactor   float64       `json:"backoff_factor"`
	RetryableErrors []string      `json:"retryable_errors,omitempty"`
}

// NotificationConfig defines notification configuration
type NotificationConfig struct {
	Enabled   bool                   `json:"enabled"`
	Channels  []*NotificationChannel `json:"channels,omitempty"`
	Events    []NotificationEvent    `json:"events,omitempty"`
	Templates map[string]string      `json:"templates,omitempty"`
}

// NotificationChannel defines a notification channel
type NotificationChannel struct {
	Type   NotificationChannelType `json:"type"`
	Target string                  `json:"target"`
	Config map[string]interface{}  `json:"config,omitempty"`
}

// NotificationChannelType defines notification channel type
type NotificationChannelType string

const (
	NotificationChannelTypeEmail   NotificationChannelType = "email"
	NotificationChannelTypeSlack   NotificationChannelType = "slack"
	NotificationChannelTypeWebhook NotificationChannelType = "webhook"
	NotificationChannelTypeSMS     NotificationChannelType = "sms"
	NotificationChannelTypeCustom  NotificationChannelType = "custom"
)

// NotificationEvent defines notification events
type NotificationEvent string

const (
	NotificationEventDeploymentStarted   NotificationEvent = "deployment_started"
	NotificationEventDeploymentCompleted NotificationEvent = "deployment_completed"
	NotificationEventDeploymentFailed    NotificationEvent = "deployment_failed"
	NotificationEventRollbackStarted     NotificationEvent = "rollback_started"
	NotificationEventRollbackCompleted   NotificationEvent = "rollback_completed"
	NotificationEventHealthCheckFailed   NotificationEvent = "health_check_failed"
)

// HealthCheck defines a health check
type HealthCheck struct {
	ID               string                 `json:"id"`
	Name             string                 `json:"name"`
	Type             HealthCheckType        `json:"type"`
	Path             string                 `json:"path,omitempty"`
	Port             int                    `json:"port,omitempty"`
	Protocol         Protocol               `json:"protocol,omitempty"`
	Command          []string               `json:"command,omitempty"`
	InitialDelay     time.Duration          `json:"initial_delay"`
	Interval         time.Duration          `json:"interval"`
	Timeout          time.Duration          `json:"timeout"`
	SuccessThreshold int                    `json:"success_threshold"`
	FailureThreshold int                    `json:"failure_threshold"`
	Headers          map[string]string      `json:"headers,omitempty"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
}

// HealthCheckType defines health check type
type HealthCheckType string

const (
	HealthCheckTypeHTTP   HealthCheckType = "http"
	HealthCheckTypeTCP    HealthCheckType = "tcp"
	HealthCheckTypeExec   HealthCheckType = "exec"
	HealthCheckTypeGRPC   HealthCheckType = "grpc"
	HealthCheckTypeCustom HealthCheckType = "custom"
)

// TimeRange represents a time range for queries
type TimeRange struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// PerformanceMonitoring defines performance monitoring configuration
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

// Filter Types

// DeploymentFilter contains filters for deployment queries
type DeploymentFilter struct {
	ApplicationID string             `json:"application_id,omitempty"`
	EnvironmentID string             `json:"environment_id,omitempty"`
	Strategy      DeploymentStrategy `json:"strategy,omitempty"`
	Status        DeploymentStatus   `json:"status,omitempty"`
	Version       string             `json:"version,omitempty"`
	Tags          []string           `json:"tags,omitempty"`
	Search        string             `json:"search,omitempty"`
	Limit         int                `json:"limit,omitempty"`
	Offset        int                `json:"offset,omitempty"`
}

// EnvironmentFilter contains filters for environment queries
type EnvironmentFilter struct {
	Type   EnvironmentType   `json:"type,omitempty"`
	Status EnvironmentStatus `json:"status,omitempty"`
	Tags   []string          `json:"tags,omitempty"`
	Search string            `json:"search,omitempty"`
	Limit  int               `json:"limit,omitempty"`
	Offset int               `json:"offset,omitempty"`
}
