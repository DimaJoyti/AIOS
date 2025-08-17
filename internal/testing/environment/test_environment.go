package environment

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"
)

// TestEnvironment manages test environment setup and teardown
type TestEnvironment struct {
	name        string
	config      EnvironmentConfig
	resources   []Resource
	cleanupFuncs []func() error
	mu          sync.RWMutex
	started     bool
}

// EnvironmentConfig defines test environment configuration
type EnvironmentConfig struct {
	Name            string            `json:"name"`
	Type            string            `json:"type"` // local, docker, kubernetes
	WorkingDir      string            `json:"working_dir"`
	Environment     map[string]string `json:"environment"`
	Services        []ServiceConfig   `json:"services"`
	Databases       []DatabaseConfig  `json:"databases"`
	Volumes         []VolumeConfig    `json:"volumes"`
	Networks        []NetworkConfig   `json:"networks"`
	Timeout         time.Duration     `json:"timeout"`
	CleanupOnExit   bool              `json:"cleanup_on_exit"`
	ParallelSetup   bool              `json:"parallel_setup"`
	HealthChecks    []HealthCheck     `json:"health_checks"`
}

// ServiceConfig defines a service configuration
type ServiceConfig struct {
	Name        string            `json:"name"`
	Image       string            `json:"image"`
	Ports       []PortMapping     `json:"ports"`
	Environment map[string]string `json:"environment"`
	Volumes     []string          `json:"volumes"`
	Command     []string          `json:"command"`
	HealthCheck *HealthCheck      `json:"health_check,omitempty"`
	DependsOn   []string          `json:"depends_on"`
}

// DatabaseConfig defines a database configuration
type DatabaseConfig struct {
	Name     string `json:"name"`
	Type     string `json:"type"` // postgres, mysql, mongodb, redis
	Version  string `json:"version"`
	Database string `json:"database"`
	Username string `json:"username"`
	Password string `json:"password"`
	Port     int    `json:"port"`
}

// VolumeConfig defines a volume configuration
type VolumeConfig struct {
	Name   string `json:"name"`
	Source string `json:"source"`
	Target string `json:"target"`
	Type   string `json:"type"` // bind, volume, tmpfs
}

// NetworkConfig defines a network configuration
type NetworkConfig struct {
	Name   string `json:"name"`
	Driver string `json:"driver"`
	Subnet string `json:"subnet"`
}

// PortMapping defines a port mapping
type PortMapping struct {
	Host      int    `json:"host"`
	Container int    `json:"container"`
	Protocol  string `json:"protocol"`
}

// HealthCheck defines a health check
type HealthCheck struct {
	Type     string        `json:"type"` // http, tcp, command
	Target   string        `json:"target"`
	Interval time.Duration `json:"interval"`
	Timeout  time.Duration `json:"timeout"`
	Retries  int           `json:"retries"`
}

// Resource represents a managed resource
type Resource interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	IsHealthy(ctx context.Context) (bool, error)
	GetConnectionInfo() map[string]interface{}
	GetName() string
}

// NewTestEnvironment creates a new test environment
func NewTestEnvironment(config EnvironmentConfig) *TestEnvironment {
	if config.Timeout == 0 {
		config.Timeout = 5 * time.Minute
	}
	
	return &TestEnvironment{
		name:         config.Name,
		config:       config,
		resources:    make([]Resource, 0),
		cleanupFuncs: make([]func() error, 0),
	}
}

// Start starts the test environment
func (te *TestEnvironment) Start(ctx context.Context) error {
	te.mu.Lock()
	defer te.mu.Unlock()
	
	if te.started {
		return fmt.Errorf("environment %s is already started", te.name)
	}
	
	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, te.config.Timeout)
	defer cancel()
	
	// Setup working directory
	if err := te.setupWorkingDirectory(); err != nil {
		return fmt.Errorf("failed to setup working directory: %w", err)
	}
	
	// Set environment variables
	if err := te.setupEnvironmentVariables(); err != nil {
		return fmt.Errorf("failed to setup environment variables: %w", err)
	}
	
	// Start resources
	if err := te.startResources(ctx); err != nil {
		te.cleanup()
		return fmt.Errorf("failed to start resources: %w", err)
	}
	
	// Wait for health checks
	if err := te.waitForHealthChecks(ctx); err != nil {
		te.cleanup()
		return fmt.Errorf("health checks failed: %w", err)
	}
	
	te.started = true
	return nil
}

// Stop stops the test environment
func (te *TestEnvironment) Stop(ctx context.Context) error {
	te.mu.Lock()
	defer te.mu.Unlock()
	
	if !te.started {
		return nil
	}
	
	return te.cleanup()
}

// IsReady checks if the environment is ready
func (te *TestEnvironment) IsReady(ctx context.Context) (bool, error) {
	te.mu.RLock()
	defer te.mu.RUnlock()
	
	if !te.started {
		return false, nil
	}
	
	for _, resource := range te.resources {
		healthy, err := resource.IsHealthy(ctx)
		if err != nil {
			return false, err
		}
		if !healthy {
			return false, nil
		}
	}
	
	return true, nil
}

// GetConnectionInfo returns connection information for all resources
func (te *TestEnvironment) GetConnectionInfo() map[string]map[string]interface{} {
	te.mu.RLock()
	defer te.mu.RUnlock()
	
	info := make(map[string]map[string]interface{})
	for _, resource := range te.resources {
		info[resource.GetName()] = resource.GetConnectionInfo()
	}
	
	return info
}

// AddResource adds a resource to the environment
func (te *TestEnvironment) AddResource(resource Resource) {
	te.mu.Lock()
	defer te.mu.Unlock()
	
	te.resources = append(te.resources, resource)
}

// AddCleanupFunc adds a cleanup function
func (te *TestEnvironment) AddCleanupFunc(fn func() error) {
	te.mu.Lock()
	defer te.mu.Unlock()
	
	te.cleanupFuncs = append(te.cleanupFuncs, fn)
}

// setupWorkingDirectory sets up the working directory
func (te *TestEnvironment) setupWorkingDirectory() error {
	if te.config.WorkingDir == "" {
		return nil
	}
	
	if err := os.MkdirAll(te.config.WorkingDir, 0755); err != nil {
		return err
	}
	
	return os.Chdir(te.config.WorkingDir)
}

// setupEnvironmentVariables sets up environment variables
func (te *TestEnvironment) setupEnvironmentVariables() error {
	for key, value := range te.config.Environment {
		if err := os.Setenv(key, value); err != nil {
			return err
		}
		
		// Add cleanup to restore original value
		originalValue := os.Getenv(key)
		te.cleanupFuncs = append(te.cleanupFuncs, func() error {
			if originalValue == "" {
				return os.Unsetenv(key)
			}
			return os.Setenv(key, originalValue)
		})
	}
	
	return nil
}

// startResources starts all resources
func (te *TestEnvironment) startResources(ctx context.Context) error {
	if te.config.ParallelSetup {
		return te.startResourcesParallel(ctx)
	}
	return te.startResourcesSequential(ctx)
}

// startResourcesSequential starts resources sequentially
func (te *TestEnvironment) startResourcesSequential(ctx context.Context) error {
	for _, resource := range te.resources {
		if err := resource.Start(ctx); err != nil {
			return fmt.Errorf("failed to start resource %s: %w", resource.GetName(), err)
		}
	}
	return nil
}

// startResourcesParallel starts resources in parallel
func (te *TestEnvironment) startResourcesParallel(ctx context.Context) error {
	var wg sync.WaitGroup
	errChan := make(chan error, len(te.resources))
	
	for _, resource := range te.resources {
		wg.Add(1)
		go func(r Resource) {
			defer wg.Done()
			if err := r.Start(ctx); err != nil {
				errChan <- fmt.Errorf("failed to start resource %s: %w", r.GetName(), err)
			}
		}(resource)
	}
	
	wg.Wait()
	close(errChan)
	
	// Check for errors
	for err := range errChan {
		return err
	}
	
	return nil
}

// waitForHealthChecks waits for all health checks to pass
func (te *TestEnvironment) waitForHealthChecks(ctx context.Context) error {
	for _, healthCheck := range te.config.HealthChecks {
		if err := te.waitForHealthCheck(ctx, healthCheck); err != nil {
			return err
		}
	}
	
	// Check resource health
	for _, resource := range te.resources {
		if err := te.waitForResourceHealth(ctx, resource); err != nil {
			return err
		}
	}
	
	return nil
}

// waitForHealthCheck waits for a specific health check to pass
func (te *TestEnvironment) waitForHealthCheck(ctx context.Context, check HealthCheck) error {
	ticker := time.NewTicker(check.Interval)
	defer ticker.Stop()
	
	for i := 0; i < check.Retries; i++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			healthy, err := te.performHealthCheck(ctx, check)
			if err != nil {
				continue
			}
			if healthy {
				return nil
			}
		}
	}
	
	return fmt.Errorf("health check failed after %d retries", check.Retries)
}

// waitForResourceHealth waits for a resource to become healthy
func (te *TestEnvironment) waitForResourceHealth(ctx context.Context, resource Resource) error {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	
	timeout := time.After(30 * time.Second)
	
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timeout:
			return fmt.Errorf("resource %s health check timed out", resource.GetName())
		case <-ticker.C:
			healthy, err := resource.IsHealthy(ctx)
			if err != nil {
				continue
			}
			if healthy {
				return nil
			}
		}
	}
}

// performHealthCheck performs a health check
func (te *TestEnvironment) performHealthCheck(ctx context.Context, check HealthCheck) (bool, error) {
	switch check.Type {
	case "http":
		return te.performHTTPHealthCheck(ctx, check)
	case "tcp":
		return te.performTCPHealthCheck(ctx, check)
	case "command":
		return te.performCommandHealthCheck(ctx, check)
	default:
		return false, fmt.Errorf("unsupported health check type: %s", check.Type)
	}
}

// performHTTPHealthCheck performs an HTTP health check
func (te *TestEnvironment) performHTTPHealthCheck(ctx context.Context, check HealthCheck) (bool, error) {
	// Implementation would make HTTP request to check.Target
	// For now, return true as placeholder
	return true, nil
}

// performTCPHealthCheck performs a TCP health check
func (te *TestEnvironment) performTCPHealthCheck(ctx context.Context, check HealthCheck) (bool, error) {
	// Implementation would attempt TCP connection to check.Target
	// For now, return true as placeholder
	return true, nil
}

// performCommandHealthCheck performs a command health check
func (te *TestEnvironment) performCommandHealthCheck(ctx context.Context, check HealthCheck) (bool, error) {
	// Implementation would execute command specified in check.Target
	// For now, return true as placeholder
	return true, nil
}

// cleanup performs cleanup of all resources
func (te *TestEnvironment) cleanup() error {
	var errors []error
	
	// Stop resources in reverse order
	for i := len(te.resources) - 1; i >= 0; i-- {
		if err := te.resources[i].Stop(context.Background()); err != nil {
			errors = append(errors, err)
		}
	}
	
	// Run cleanup functions in reverse order
	for i := len(te.cleanupFuncs) - 1; i >= 0; i-- {
		if err := te.cleanupFuncs[i](); err != nil {
			errors = append(errors, err)
		}
	}
	
	te.started = false
	
	if len(errors) > 0 {
		return fmt.Errorf("cleanup errors: %v", errors)
	}
	
	return nil
}

// EnvironmentManager manages multiple test environments
type EnvironmentManager struct {
	environments map[string]*TestEnvironment
	mu           sync.RWMutex
}

// NewEnvironmentManager creates a new environment manager
func NewEnvironmentManager() *EnvironmentManager {
	return &EnvironmentManager{
		environments: make(map[string]*TestEnvironment),
	}
}

// CreateEnvironment creates a new test environment
func (em *EnvironmentManager) CreateEnvironment(config EnvironmentConfig) (*TestEnvironment, error) {
	em.mu.Lock()
	defer em.mu.Unlock()
	
	if _, exists := em.environments[config.Name]; exists {
		return nil, fmt.Errorf("environment %s already exists", config.Name)
	}
	
	env := NewTestEnvironment(config)
	em.environments[config.Name] = env
	
	return env, nil
}

// GetEnvironment gets an existing test environment
func (em *EnvironmentManager) GetEnvironment(name string) (*TestEnvironment, error) {
	em.mu.RLock()
	defer em.mu.RUnlock()
	
	env, exists := em.environments[name]
	if !exists {
		return nil, fmt.Errorf("environment %s not found", name)
	}
	
	return env, nil
}

// StartEnvironment starts a test environment
func (em *EnvironmentManager) StartEnvironment(ctx context.Context, name string) error {
	env, err := em.GetEnvironment(name)
	if err != nil {
		return err
	}
	
	return env.Start(ctx)
}

// StopEnvironment stops a test environment
func (em *EnvironmentManager) StopEnvironment(ctx context.Context, name string) error {
	env, err := em.GetEnvironment(name)
	if err != nil {
		return err
	}
	
	return env.Stop(ctx)
}

// StopAllEnvironments stops all test environments
func (em *EnvironmentManager) StopAllEnvironments(ctx context.Context) error {
	em.mu.RLock()
	defer em.mu.RUnlock()
	
	var errors []error
	
	for name, env := range em.environments {
		if err := env.Stop(ctx); err != nil {
			errors = append(errors, fmt.Errorf("failed to stop environment %s: %w", name, err))
		}
	}
	
	if len(errors) > 0 {
		return fmt.Errorf("errors stopping environments: %v", errors)
	}
	
	return nil
}

// ListEnvironments lists all environments
func (em *EnvironmentManager) ListEnvironments() []string {
	em.mu.RLock()
	defer em.mu.RUnlock()
	
	names := make([]string, 0, len(em.environments))
	for name := range em.environments {
		names = append(names, name)
	}
	
	return names
}

// LoadEnvironmentFromFile loads environment configuration from file
func LoadEnvironmentFromFile(filename string) (EnvironmentConfig, error) {
	// Implementation would load from JSON/YAML file
	return EnvironmentConfig{}, fmt.Errorf("not implemented")
}

// SaveEnvironmentToFile saves environment configuration to file
func SaveEnvironmentToFile(config EnvironmentConfig, filename string) error {
	// Implementation would save to JSON/YAML file
	return fmt.Errorf("not implemented")
}
