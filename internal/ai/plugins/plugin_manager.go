package plugins

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/aios/aios/internal/ai"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

// PluginManager manages AI service plugins and extensions
type PluginManager struct {
	logger       *logrus.Logger
	tracer       trace.Tracer
	orchestrator *ai.Orchestrator

	// Plugin management
	plugins  map[string]*Plugin
	registry *PluginRegistry
	loader   *PluginLoader
	sandbox  *PluginSandbox
	mu       sync.RWMutex

	// Configuration
	config PluginManagerConfig
}

// PluginManagerConfig represents plugin manager configuration
type PluginManagerConfig struct {
	EnablePlugins       bool          `json:"enable_plugins"`
	PluginDirectory     string        `json:"plugin_directory"`
	MaxPlugins          int           `json:"max_plugins"`
	SandboxEnabled      bool          `json:"sandbox_enabled"`
	AutoLoad            bool          `json:"auto_load"`
	HealthCheckInterval time.Duration `json:"health_check_interval"`
	AllowedAPIs         []string      `json:"allowed_apis"`
	SecurityLevel       SecurityLevel `json:"security_level"`
}

// SecurityLevel represents plugin security levels
type SecurityLevel string

const (
	SecurityLevelNone     SecurityLevel = "none"
	SecurityLevelBasic    SecurityLevel = "basic"
	SecurityLevelStrict   SecurityLevel = "strict"
	SecurityLevelIsolated SecurityLevel = "isolated"
)

// Plugin represents a loaded AI service plugin
type Plugin struct {
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	Version      string                 `json:"version"`
	Author       string                 `json:"author"`
	Description  string                 `json:"description"`
	Type         PluginType             `json:"type"`
	Status       PluginStatus           `json:"status"`
	Capabilities []PluginCapability     `json:"capabilities"`
	Dependencies []PluginDependency     `json:"dependencies"`
	Config       map[string]interface{} `json:"config"`
	Metadata     PluginMetadata         `json:"metadata"`
	LoadedAt     time.Time              `json:"loaded_at"`
	LastUsed     *time.Time             `json:"last_used,omitempty"`
	UsageCount   int64                  `json:"usage_count"`

	// Runtime
	instance    PluginInstance
	healthCheck HealthChecker
}

// PluginType represents different types of plugins
type PluginType string

const (
	PluginTypeLLM        PluginType = "llm"
	PluginTypeVision     PluginType = "vision"
	PluginTypeVoice      PluginType = "voice"
	PluginTypeNLP        PluginType = "nlp"
	PluginTypeRAG        PluginType = "rag"
	PluginTypeMultiModal PluginType = "multimodal"
	PluginTypeWorkflow   PluginType = "workflow"
	PluginTypeUtility    PluginType = "utility"
	PluginTypeCustom     PluginType = "custom"
)

// PluginStatus represents plugin status
type PluginStatus string

const (
	StatusUnloaded  PluginStatus = "unloaded"
	StatusLoading   PluginStatus = "loading"
	StatusLoaded    PluginStatus = "loaded"
	StatusActive    PluginStatus = "active"
	StatusInactive  PluginStatus = "inactive"
	StatusError     PluginStatus = "error"
	StatusUnloading PluginStatus = "unloading"
)

// PluginCapability represents plugin capabilities
type PluginCapability struct {
	Name        string                `json:"name"`
	Version     string                `json:"version"`
	Description string                `json:"description"`
	Parameters  []CapabilityParameter `json:"parameters"`
	Returns     CapabilityReturn      `json:"returns"`
	Examples    []CapabilityExample   `json:"examples"`
}

// CapabilityParameter represents a capability parameter
type CapabilityParameter struct {
	Name        string      `json:"name"`
	Type        string      `json:"type"`
	Description string      `json:"description"`
	Required    bool        `json:"required"`
	Default     interface{} `json:"default,omitempty"`
	Validation  string      `json:"validation,omitempty"`
}

// CapabilityReturn represents capability return value
type CapabilityReturn struct {
	Type        string `json:"type"`
	Description string `json:"description"`
	Schema      string `json:"schema,omitempty"`
}

// CapabilityExample represents a capability usage example
type CapabilityExample struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Input       map[string]interface{} `json:"input"`
	Output      interface{}            `json:"output"`
}

// PluginDependency represents plugin dependencies
type PluginDependency struct {
	Name       string `json:"name"`
	Version    string `json:"version"`
	Type       string `json:"type"` // plugin, library, service
	Required   bool   `json:"required"`
	MinVersion string `json:"min_version,omitempty"`
	MaxVersion string `json:"max_version,omitempty"`
}

// PluginMetadata represents plugin metadata
type PluginMetadata struct {
	Homepage   string                 `json:"homepage,omitempty"`
	Repository string                 `json:"repository,omitempty"`
	License    string                 `json:"license,omitempty"`
	Tags       []string               `json:"tags"`
	Category   string                 `json:"category"`
	Rating     float64                `json:"rating"`
	Downloads  int64                  `json:"downloads"`
	Size       int64                  `json:"size"`
	Checksum   string                 `json:"checksum"`
	Signature  string                 `json:"signature,omitempty"`
	CreatedAt  time.Time              `json:"created_at"`
	UpdatedAt  time.Time              `json:"updated_at"`
	Extra      map[string]interface{} `json:"extra,omitempty"`
}

// PluginInstance represents a plugin instance interface
type PluginInstance interface {
	Initialize(ctx context.Context, config map[string]interface{}) error
	Execute(ctx context.Context, capability string, params map[string]interface{}) (interface{}, error)
	GetCapabilities() []PluginCapability
	GetStatus() PluginStatus
	Cleanup() error
}

// HealthChecker represents plugin health checking interface
type HealthChecker interface {
	CheckHealth(ctx context.Context) PluginHealth
}

// PluginHealth represents plugin health status
type PluginHealth struct {
	Status     HealthStatus           `json:"status"`
	Message    string                 `json:"message,omitempty"`
	Metrics    map[string]interface{} `json:"metrics,omitempty"`
	LastCheck  time.Time              `json:"last_check"`
	CheckCount int64                  `json:"check_count"`
	ErrorCount int64                  `json:"error_count"`
}

// HealthStatus represents health status
type HealthStatus string

const (
	HealthStatusHealthy   HealthStatus = "healthy"
	HealthStatusDegraded  HealthStatus = "degraded"
	HealthStatusUnhealthy HealthStatus = "unhealthy"
	HealthStatusUnknown   HealthStatus = "unknown"
)

// PluginRegistry manages plugin discovery and metadata
type PluginRegistry struct {
	manager      *PluginManager
	repositories []PluginRepository
	cache        map[string]*PluginInfo
	mu           sync.RWMutex
}

// PluginRepository represents a plugin repository
type PluginRepository struct {
	Name     string    `json:"name"`
	URL      string    `json:"url"`
	Type     string    `json:"type"` // local, remote, git
	Enabled  bool      `json:"enabled"`
	Priority int       `json:"priority"`
	LastSync time.Time `json:"last_sync"`
}

// PluginInfo represents plugin information from registry
type PluginInfo struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Version     string         `json:"version"`
	Description string         `json:"description"`
	Author      string         `json:"author"`
	Type        PluginType     `json:"type"`
	Metadata    PluginMetadata `json:"metadata"`
	Available   bool           `json:"available"`
	Installed   bool           `json:"installed"`
}

// PluginLoader handles plugin loading and unloading
type PluginLoader struct {
	manager *PluginManager
	mu      sync.RWMutex
}

// PluginSandbox provides security isolation for plugins
type PluginSandbox struct {
	manager        *PluginManager
	allowedAPIs    map[string]bool
	resourceLimits ResourceLimits
	mu             sync.RWMutex
}

// ResourceLimits represents resource limits for plugins
type ResourceLimits struct {
	MaxMemory        int64         `json:"max_memory"`
	MaxCPU           float64       `json:"max_cpu"`
	MaxDiskSpace     int64         `json:"max_disk_space"`
	MaxNetworkIO     int64         `json:"max_network_io"`
	MaxExecutionTime time.Duration `json:"max_execution_time"`
	MaxConcurrency   int           `json:"max_concurrency"`
}

// NewPluginManager creates a new plugin manager
func NewPluginManager(
	logger *logrus.Logger,
	orchestrator *ai.Orchestrator,
	config PluginManagerConfig,
) *PluginManager {
	manager := &PluginManager{
		logger:       logger,
		tracer:       otel.Tracer("ai.plugin_manager"),
		orchestrator: orchestrator,
		plugins:      make(map[string]*Plugin),
		config:       config,
	}

	// Initialize components
	manager.registry = &PluginRegistry{
		manager:      manager,
		repositories: make([]PluginRepository, 0),
		cache:        make(map[string]*PluginInfo),
	}

	manager.loader = &PluginLoader{
		manager: manager,
	}

	manager.sandbox = &PluginSandbox{
		manager:     manager,
		allowedAPIs: make(map[string]bool),
		resourceLimits: ResourceLimits{
			MaxMemory:        1024 * 1024 * 1024, // 1GB
			MaxCPU:           1.0,                // 1 CPU core
			MaxDiskSpace:     100 * 1024 * 1024,  // 100MB
			MaxNetworkIO:     10 * 1024 * 1024,   // 10MB
			MaxExecutionTime: 30 * time.Second,
			MaxConcurrency:   10,
		},
	}

	// Set allowed APIs
	for _, api := range config.AllowedAPIs {
		manager.sandbox.allowedAPIs[api] = true
	}

	return manager
}

// Start initializes the plugin manager
func (pm *PluginManager) Start(ctx context.Context) error {
	ctx, span := pm.tracer.Start(ctx, "plugin_manager.Start")
	defer span.End()

	pm.logger.Info("Starting plugin manager")

	if !pm.config.EnablePlugins {
		pm.logger.Info("Plugin system disabled")
		return nil
	}

	// Initialize registry
	if err := pm.registry.initialize(); err != nil {
		return fmt.Errorf("failed to initialize plugin registry: %w", err)
	}

	// Auto-load plugins if enabled
	if pm.config.AutoLoad {
		if err := pm.autoLoadPlugins(ctx); err != nil {
			pm.logger.WithError(err).Error("Failed to auto-load plugins")
		}
	}

	// Start health checking
	if pm.config.HealthCheckInterval > 0 {
		go pm.startHealthChecking()
	}

	pm.logger.WithFields(logrus.Fields{
		"plugin_count":    len(pm.plugins),
		"sandbox_enabled": pm.config.SandboxEnabled,
		"security_level":  pm.config.SecurityLevel,
	}).Info("Plugin manager started successfully")

	return nil
}

// Stop shuts down the plugin manager
func (pm *PluginManager) Stop(ctx context.Context) error {
	ctx, span := pm.tracer.Start(ctx, "plugin_manager.Stop")
	defer span.End()

	pm.logger.Info("Stopping plugin manager")

	// Unload all plugins
	pm.mu.Lock()
	for _, plugin := range pm.plugins {
		if err := pm.unloadPlugin(plugin.ID); err != nil {
			pm.logger.WithError(err).WithField("plugin_id", plugin.ID).Error("Failed to unload plugin")
		}
	}
	pm.mu.Unlock()

	pm.logger.Info("Plugin manager stopped")
	return nil
}

// LoadPlugin loads a plugin by ID
func (pm *PluginManager) LoadPlugin(ctx context.Context, pluginID string) error {
	ctx, span := pm.tracer.Start(ctx, "plugin_manager.LoadPlugin")
	defer span.End()

	if !pm.config.EnablePlugins {
		return fmt.Errorf("plugin system is disabled")
	}

	pm.logger.WithField("plugin_id", pluginID).Info("Loading plugin")

	// Check if plugin is already loaded
	pm.mu.RLock()
	if _, exists := pm.plugins[pluginID]; exists {
		pm.mu.RUnlock()
		return fmt.Errorf("plugin already loaded: %s", pluginID)
	}
	pm.mu.RUnlock()

	// Load plugin
	plugin, err := pm.loader.loadPlugin(ctx, pluginID)
	if err != nil {
		return fmt.Errorf("failed to load plugin: %w", err)
	}

	// Store plugin
	pm.mu.Lock()
	pm.plugins[pluginID] = plugin
	pm.mu.Unlock()

	pm.logger.WithFields(logrus.Fields{
		"plugin_id":   pluginID,
		"plugin_name": plugin.Name,
		"version":     plugin.Version,
	}).Info("Plugin loaded successfully")

	return nil
}

// UnloadPlugin unloads a plugin by ID
func (pm *PluginManager) UnloadPlugin(ctx context.Context, pluginID string) error {
	ctx, span := pm.tracer.Start(ctx, "plugin_manager.UnloadPlugin")
	defer span.End()

	pm.logger.WithField("plugin_id", pluginID).Info("Unloading plugin")

	pm.mu.Lock()
	defer pm.mu.Unlock()

	return pm.unloadPlugin(pluginID)
}

// ExecutePlugin executes a plugin capability
func (pm *PluginManager) ExecutePlugin(ctx context.Context, pluginID, capability string, params map[string]interface{}) (interface{}, error) {
	ctx, span := pm.tracer.Start(ctx, "plugin_manager.ExecutePlugin")
	defer span.End()

	pm.mu.RLock()
	plugin, exists := pm.plugins[pluginID]
	pm.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("plugin not found: %s", pluginID)
	}

	if plugin.Status != StatusActive {
		return nil, fmt.Errorf("plugin not active: %s", pluginID)
	}

	// Execute in sandbox if enabled
	if pm.config.SandboxEnabled {
		return pm.sandbox.executeInSandbox(ctx, plugin, capability, params)
	}

	// Direct execution
	result, err := plugin.instance.Execute(ctx, capability, params)
	if err != nil {
		return nil, err
	}

	// Update usage statistics
	pm.mu.Lock()
	plugin.UsageCount++
	now := time.Now()
	plugin.LastUsed = &now
	pm.mu.Unlock()

	return result, nil
}

// ListPlugins returns all loaded plugins
func (pm *PluginManager) ListPlugins(ctx context.Context) ([]*Plugin, error) {
	ctx, span := pm.tracer.Start(ctx, "plugin_manager.ListPlugins")
	defer span.End()

	pm.mu.RLock()
	defer pm.mu.RUnlock()

	plugins := make([]*Plugin, 0, len(pm.plugins))
	for _, plugin := range pm.plugins {
		// Return a copy
		pluginCopy := *plugin
		plugins = append(plugins, &pluginCopy)
	}

	return plugins, nil
}

// GetPluginInfo returns information about a specific plugin
func (pm *PluginManager) GetPluginInfo(ctx context.Context, pluginID string) (*Plugin, error) {
	ctx, span := pm.tracer.Start(ctx, "plugin_manager.GetPluginInfo")
	defer span.End()

	pm.mu.RLock()
	plugin, exists := pm.plugins[pluginID]
	pm.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("plugin not found: %s", pluginID)
	}

	// Return a copy
	pluginCopy := *plugin
	return &pluginCopy, nil
}

// Helper methods

func (pm *PluginManager) autoLoadPlugins(ctx context.Context) error {
	// Mock auto-loading - would scan plugin directory and load plugins
	pm.logger.Info("Auto-loading plugins")

	// Example: Load a mock plugin
	mockPlugin := &Plugin{
		ID:          "mock-llm-plugin",
		Name:        "Mock LLM Plugin",
		Version:     "1.0.0",
		Author:      "AIOS Team",
		Description: "A mock LLM plugin for demonstration",
		Type:        PluginTypeLLM,
		Status:      StatusLoaded,
		Capabilities: []PluginCapability{
			{
				Name:        "generate_text",
				Version:     "1.0.0",
				Description: "Generate text using the mock LLM",
				Parameters: []CapabilityParameter{
					{
						Name:        "prompt",
						Type:        "string",
						Description: "Input prompt for text generation",
						Required:    true,
					},
					{
						Name:        "max_tokens",
						Type:        "integer",
						Description: "Maximum number of tokens to generate",
						Required:    false,
						Default:     100,
					},
				},
				Returns: CapabilityReturn{
					Type:        "object",
					Description: "Generated text response",
				},
			},
		},
		Config:   make(map[string]interface{}),
		LoadedAt: time.Now(),
		instance: &MockPluginInstance{},
	}

	pm.mu.Lock()
	pm.plugins[mockPlugin.ID] = mockPlugin
	pm.mu.Unlock()

	pm.logger.WithField("plugin_count", 1).Info("Auto-loading completed")
	return nil
}

func (pm *PluginManager) startHealthChecking() {
	ticker := time.NewTicker(pm.config.HealthCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			pm.checkPluginHealth()
		}
	}
}

func (pm *PluginManager) checkPluginHealth() {
	pm.mu.RLock()
	plugins := make([]*Plugin, 0, len(pm.plugins))
	for _, plugin := range pm.plugins {
		plugins = append(plugins, plugin)
	}
	pm.mu.RUnlock()

	for _, plugin := range plugins {
		if plugin.healthCheck != nil {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			health := plugin.healthCheck.CheckHealth(ctx)
			cancel()

			if health.Status != HealthStatusHealthy {
				pm.logger.WithFields(logrus.Fields{
					"plugin_id": plugin.ID,
					"status":    health.Status,
					"message":   health.Message,
				}).Warn("Plugin health check failed")
			}
		}
	}
}

func (pm *PluginManager) unloadPlugin(pluginID string) error {
	plugin, exists := pm.plugins[pluginID]
	if !exists {
		return fmt.Errorf("plugin not found: %s", pluginID)
	}

	plugin.Status = StatusUnloading

	// Cleanup plugin
	if plugin.instance != nil {
		if err := plugin.instance.Cleanup(); err != nil {
			pm.logger.WithError(err).WithField("plugin_id", pluginID).Error("Plugin cleanup failed")
		}
	}

	delete(pm.plugins, pluginID)

	pm.logger.WithField("plugin_id", pluginID).Info("Plugin unloaded")
	return nil
}

// PluginRegistry methods

func (pr *PluginRegistry) initialize() error {
	// Initialize default repositories
	pr.repositories = []PluginRepository{
		{
			Name:     "local",
			URL:      "./plugins",
			Type:     "local",
			Enabled:  true,
			Priority: 1,
		},
	}

	return nil
}

// PluginLoader methods

func (pl *PluginLoader) loadPlugin(ctx context.Context, pluginID string) (*Plugin, error) {
	// Mock plugin loading - in real implementation, this would load actual plugin files
	return nil, fmt.Errorf("plugin loading not implemented: %s", pluginID)
}

// PluginSandbox methods

func (ps *PluginSandbox) executeInSandbox(ctx context.Context, plugin *Plugin, capability string, params map[string]interface{}) (interface{}, error) {
	// Mock sandbox execution - in real implementation, this would provide security isolation
	return plugin.instance.Execute(ctx, capability, params)
}

// MockPluginInstance for demonstration
type MockPluginInstance struct{}

func (mpi *MockPluginInstance) Initialize(ctx context.Context, config map[string]interface{}) error {
	return nil
}

func (mpi *MockPluginInstance) Execute(ctx context.Context, capability string, params map[string]interface{}) (interface{}, error) {
	switch capability {
	case "generate_text":
		prompt, _ := params["prompt"].(string)
		return map[string]interface{}{
			"text":       fmt.Sprintf("Mock response to: %s", prompt),
			"tokens":     50,
			"confidence": 0.85,
		}, nil
	default:
		return nil, fmt.Errorf("unsupported capability: %s", capability)
	}
}

func (mpi *MockPluginInstance) GetCapabilities() []PluginCapability {
	return []PluginCapability{}
}

func (mpi *MockPluginInstance) GetStatus() PluginStatus {
	return StatusActive
}

func (mpi *MockPluginInstance) Cleanup() error {
	return nil
}
