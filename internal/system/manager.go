package system

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/aios/aios/pkg/models"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

// Manager handles core system operations and AI integration
type Manager struct {
	logger          *logrus.Logger
	tracer          trace.Tracer
	resourceManager *ResourceManager
	fileSystemAI    *FileSystemAI
	securityManager *SecurityManager
	optimizationAI  *OptimizationAI
	mu              sync.RWMutex
	running         bool
	stopCh          chan struct{}
}

// NewManager creates a new system manager instance
func NewManager(logger *logrus.Logger) (*Manager, error) {
	tracer := otel.Tracer("system-manager")

	resourceManager, err := NewResourceManager(logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource manager: %w", err)
	}

	fileSystemAI, err := NewFileSystemAI(logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create filesystem AI: %w", err)
	}

	securityManager, err := NewSecurityManager(logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create security manager: %w", err)
	}

	optimizationAI, err := NewOptimizationAI(logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create optimization AI: %w", err)
	}

	return &Manager{
		logger:          logger,
		tracer:          tracer,
		resourceManager: resourceManager,
		fileSystemAI:    fileSystemAI,
		securityManager: securityManager,
		optimizationAI:  optimizationAI,
		stopCh:          make(chan struct{}),
	}, nil
}

// Start initializes and starts all system components
func (m *Manager) Start(ctx context.Context) error {
	ctx, span := m.tracer.Start(ctx, "system.Manager.Start")
	defer span.End()

	m.mu.Lock()
	defer m.mu.Unlock()

	if m.running {
		return fmt.Errorf("system manager is already running")
	}

	m.logger.Info("Starting system manager...")

	// Start resource manager
	if err := m.resourceManager.Start(ctx); err != nil {
		return fmt.Errorf("failed to start resource manager: %w", err)
	}

	// Start filesystem AI
	if err := m.fileSystemAI.Start(ctx); err != nil {
		return fmt.Errorf("failed to start filesystem AI: %w", err)
	}

	// Start security manager
	if err := m.securityManager.Start(ctx); err != nil {
		return fmt.Errorf("failed to start security manager: %w", err)
	}

	// Start optimization AI
	if err := m.optimizationAI.Start(ctx); err != nil {
		return fmt.Errorf("failed to start optimization AI: %w", err)
	}

	// Start monitoring goroutines
	go m.monitorSystem()
	go m.performOptimizations()

	m.running = true
	m.logger.Info("System manager started successfully")

	return nil
}

// Stop gracefully shuts down all system components
func (m *Manager) Stop(ctx context.Context) error {
	ctx, span := m.tracer.Start(ctx, "system.Manager.Stop")
	defer span.End()

	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.running {
		return nil
	}

	m.logger.Info("Stopping system manager...")

	// Signal stop to monitoring goroutines
	close(m.stopCh)

	// Stop components in reverse order
	if err := m.optimizationAI.Stop(ctx); err != nil {
		m.logger.WithError(err).Error("Failed to stop optimization AI")
	}

	if err := m.securityManager.Stop(ctx); err != nil {
		m.logger.WithError(err).Error("Failed to stop security manager")
	}

	if err := m.fileSystemAI.Stop(ctx); err != nil {
		m.logger.WithError(err).Error("Failed to stop filesystem AI")
	}

	if err := m.resourceManager.Stop(ctx); err != nil {
		m.logger.WithError(err).Error("Failed to stop resource manager")
	}

	m.running = false
	m.logger.Info("System manager stopped")

	return nil
}

// GetSystemStatus returns the current system status
func (m *Manager) GetSystemStatus(ctx context.Context) (*models.SystemStatus, error) {
	ctx, span := m.tracer.Start(ctx, "system.Manager.GetSystemStatus")
	defer span.End()

	m.mu.RLock()
	defer m.mu.RUnlock()

	if !m.running {
		return nil, fmt.Errorf("system manager is not running")
	}

	resourceStatus, err := m.resourceManager.GetStatus(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get resource status: %w", err)
	}

	securityStatus, err := m.securityManager.GetStatus(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get security status: %w", err)
	}

	optimizationStatus, err := m.optimizationAI.GetStatus(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get optimization status: %w", err)
	}

	return &models.SystemStatus{
		Running:      m.running,
		Version:      "dev",
		Uptime:       time.Since(time.Now().Add(-1 * time.Hour)), // Mock uptime
		Resources:    resourceStatus,
		Security:     securityStatus,
		Optimization: optimizationStatus,
		Services:     make(map[string]models.ServiceStatus),
		Timestamp:    time.Now(),
	}, nil
}

// GetResourceManager returns the resource manager instance
func (m *Manager) GetResourceManager() *ResourceManager {
	return m.resourceManager
}

// GetFileSystemAI returns the filesystem AI instance
func (m *Manager) GetFileSystemAI() *FileSystemAI {
	return m.fileSystemAI
}

// GetSecurityManager returns the security manager instance
func (m *Manager) GetSecurityManager() *SecurityManager {
	return m.securityManager
}

// GetOptimizationAI returns the optimization AI instance
func (m *Manager) GetOptimizationAI() *OptimizationAI {
	return m.optimizationAI
}

// monitorSystem continuously monitors system health and performance
func (m *Manager) monitorSystem() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

			status, err := m.GetSystemStatus(ctx)
			if err != nil {
				m.logger.WithError(err).Error("Failed to get system status during monitoring")
			} else {
				m.logger.WithFields(logrus.Fields{
					"cpu_usage":    status.Resources.CPU.Usage,
					"memory_usage": status.Resources.Memory.Usage,
					"disk_usage":   status.Resources.Disk.Filesystems[0].Usage,
				}).Debug("System status update")
			}

			cancel()

		case <-m.stopCh:
			m.logger.Debug("System monitoring stopped")
			return
		}
	}
}

// performOptimizations runs periodic system optimizations
func (m *Manager) performOptimizations() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)

			if err := m.optimizationAI.RunOptimization(ctx); err != nil {
				m.logger.WithError(err).Error("Failed to run system optimization")
			}

			cancel()

		case <-m.stopCh:
			m.logger.Debug("System optimization stopped")
			return
		}
	}
}
