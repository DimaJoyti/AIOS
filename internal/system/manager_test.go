package system

import (
	"context"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewManager(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	manager, err := NewManager(logger)
	require.NoError(t, err)
	assert.NotNil(t, manager)
	assert.NotNil(t, manager.logger)
	assert.NotNil(t, manager.tracer)
	assert.NotNil(t, manager.resourceManager)
	assert.NotNil(t, manager.fileSystemAI)
	assert.NotNil(t, manager.securityManager)
	assert.NotNil(t, manager.optimizationAI)
	assert.False(t, manager.running)
}

func TestManagerStartStop(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	manager, err := NewManager(logger)
	require.NoError(t, err)

	ctx := context.Background()

	// Test start
	err = manager.Start(ctx)
	require.NoError(t, err)
	assert.True(t, manager.running)

	// Test double start (should not error)
	err = manager.Start(ctx)
	assert.Error(t, err)

	// Test stop
	err = manager.Stop(ctx)
	require.NoError(t, err)
	assert.False(t, manager.running)

	// Test double stop (should not error)
	err = manager.Stop(ctx)
	require.NoError(t, err)
}

func TestGetSystemStatus(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	manager, err := NewManager(logger)
	require.NoError(t, err)

	ctx := context.Background()

	// Test status when not running
	_, err = manager.GetSystemStatus(ctx)
	assert.Error(t, err)

	// Start manager
	err = manager.Start(ctx)
	require.NoError(t, err)
	defer manager.Stop(ctx)

	// Test status when running
	status, err := manager.GetSystemStatus(ctx)
	require.NoError(t, err)
	assert.NotNil(t, status)
	assert.True(t, status.Running)
	assert.Equal(t, "dev", status.Version)
	assert.NotNil(t, status.Resources)
	assert.NotNil(t, status.Security)
	assert.NotNil(t, status.Optimization)
	assert.NotZero(t, status.Timestamp)
}

func TestManagerComponents(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	manager, err := NewManager(logger)
	require.NoError(t, err)

	// Test component getters
	assert.NotNil(t, manager.GetResourceManager())
	assert.NotNil(t, manager.GetFileSystemAI())
	assert.NotNil(t, manager.GetSecurityManager())
	assert.NotNil(t, manager.GetOptimizationAI())
}

func TestManagerConcurrency(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	manager, err := NewManager(logger)
	require.NoError(t, err)

	ctx := context.Background()

	// Start manager
	err = manager.Start(ctx)
	require.NoError(t, err)
	defer manager.Stop(ctx)

	// Test concurrent access to system status
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			defer func() { done <- true }()
			
			status, err := manager.GetSystemStatus(ctx)
			assert.NoError(t, err)
			assert.NotNil(t, status)
			assert.True(t, status.Running)
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		select {
		case <-done:
			// Success
		case <-time.After(5 * time.Second):
			t.Fatal("Timeout waiting for concurrent operations")
		}
	}
}

func BenchmarkGetSystemStatus(b *testing.B) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel) // Reduce logging for benchmark

	manager, err := NewManager(logger)
	require.NoError(b, err)

	ctx := context.Background()
	err = manager.Start(ctx)
	require.NoError(b, err)
	defer manager.Stop(ctx)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := manager.GetSystemStatus(ctx)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}
