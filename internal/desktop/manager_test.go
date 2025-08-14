package desktop

import (
	"context"
	"testing"
	"time"

	"github.com/aios/aios/internal/ai"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewManager(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	// Create AI orchestrator
	aiConfig := ai.AIServiceConfig{
		OllamaHost:    "localhost",
		OllamaPort:    11434,
		OllamaTimeout: 30 * time.Second,
		DefaultModel:  "llama2",
		MaxTokens:     2048,
		Temperature:   0.7,
	}
	aiOrchestrator := ai.NewOrchestrator(aiConfig, logger)

	// Create desktop config
	config := DesktopConfig{
		Theme: "dark",
		AIAssistant: AIAssistantConfig{
			VoiceEnabled:    true,
			WakeWord:        "aios",
			Language:        "en-US",
			AutoSuggestions: true,
			ContextAware:    true,
		},
		WindowManager: WindowManagerConfig{
			TilingEnabled:     true,
			SmartGaps:         true,
			BorderWidth:       2,
			AnimationSpeed:    1.0,
			FocusFollowsMouse: false,
			AutoTiling:        true,
		},
		Workspaces: WorkspaceConfig{
			DefaultCount:  4,
			AutoSwitch:    true,
			SmartNaming:   true,
			PersistLayout: true,
			WrapAround:    true,
		},
		Notifications: NotificationConfig{
			Enabled:       true,
			Position:      "top-right",
			Timeout:       5 * time.Second,
			MaxVisible:    5,
			SmartGrouping: true,
			AIFiltering:   true,
		},
	}

	manager, err := NewManager(logger, aiOrchestrator, config)
	require.NoError(t, err)
	assert.NotNil(t, manager)
	assert.NotNil(t, manager.windowManager)
	assert.NotNil(t, manager.workspaceManager)
	assert.NotNil(t, manager.appLauncher)
	assert.NotNil(t, manager.themeManager)
	assert.NotNil(t, manager.notificationMgr)
	assert.Equal(t, config, manager.config)
}

func TestManagerStartStop(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	// Create AI orchestrator
	aiConfig := ai.AIServiceConfig{}
	aiOrchestrator := ai.NewOrchestrator(aiConfig, logger)

	// Create desktop config
	config := DesktopConfig{
		Theme: "dark",
		WindowManager: WindowManagerConfig{
			TilingEnabled: true,
			SmartGaps:     true,
			BorderWidth:   2,
		},
		Workspaces: WorkspaceConfig{
			DefaultCount: 4,
		},
		Notifications: NotificationConfig{
			Enabled: true,
			Timeout: 5 * time.Second,
		},
	}

	manager, err := NewManager(logger, aiOrchestrator, config)
	require.NoError(t, err)

	ctx := context.Background()

	// Test start
	err = manager.Start(ctx)
	require.NoError(t, err)
	assert.True(t, manager.running)

	// Test double start (should fail)
	err = manager.Start(ctx)
	assert.Error(t, err)

	// Test stop
	err = manager.Stop(ctx)
	require.NoError(t, err)
	assert.False(t, manager.running)

	// Test double stop (should not fail)
	err = manager.Stop(ctx)
	assert.NoError(t, err)
}

func TestManagerGetStatus(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	// Create AI orchestrator
	aiConfig := ai.AIServiceConfig{}
	aiOrchestrator := ai.NewOrchestrator(aiConfig, logger)

	// Create desktop config
	config := DesktopConfig{
		Theme: "dark",
		WindowManager: WindowManagerConfig{
			TilingEnabled: true,
		},
		Workspaces: WorkspaceConfig{
			DefaultCount: 4,
		},
		Notifications: NotificationConfig{
			Enabled: true,
		},
	}

	manager, err := NewManager(logger, aiOrchestrator, config)
	require.NoError(t, err)

	ctx := context.Background()

	// Start manager
	err = manager.Start(ctx)
	require.NoError(t, err)
	defer manager.Stop(ctx)

	// Get status
	status, err := manager.GetStatus(ctx)
	require.NoError(t, err)
	assert.NotNil(t, status)
	assert.True(t, status.Running)
	assert.Equal(t, "dark", status.Theme)
	assert.NotNil(t, status.Windows)
	assert.NotNil(t, status.Workspaces)
	assert.NotNil(t, status.Applications)
	assert.NotNil(t, status.Themes)
	assert.NotNil(t, status.Notifications)
	assert.NotNil(t, status.Performance)
}

func TestManagerProcessAICommand(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	// Create AI orchestrator
	aiConfig := ai.AIServiceConfig{}
	aiOrchestrator := ai.NewOrchestrator(aiConfig, logger)

	// Initialize AI orchestrator
	ctx := context.Background()
	err := aiOrchestrator.Initialize(ctx)
	require.NoError(t, err)

	// Create desktop config
	config := DesktopConfig{
		Theme: "dark",
		WindowManager: WindowManagerConfig{
			TilingEnabled: true,
		},
		Workspaces: WorkspaceConfig{
			DefaultCount: 4,
		},
		Notifications: NotificationConfig{
			Enabled: true,
		},
	}

	manager, err := NewManager(logger, aiOrchestrator, config)
	require.NoError(t, err)

	// Start manager
	err = manager.Start(ctx)
	require.NoError(t, err)
	defer manager.Stop(ctx)

	// Process AI command
	response, err := manager.ProcessAICommand(ctx, "switch to workspace 2")
	if err != nil {
		// Skip test if Ollama is not available
		t.Skip("Skipping AI command test - Ollama not available")
	}
	assert.NotNil(t, response)
	assert.Equal(t, "chat", response.Type)
}

func TestManagerComponentAccess(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	// Create AI orchestrator
	aiConfig := ai.AIServiceConfig{}
	aiOrchestrator := ai.NewOrchestrator(aiConfig, logger)

	// Create desktop config
	config := DesktopConfig{
		Theme: "dark",
		WindowManager: WindowManagerConfig{
			TilingEnabled: true,
		},
		Workspaces: WorkspaceConfig{
			DefaultCount: 4,
		},
		Notifications: NotificationConfig{
			Enabled: true,
		},
	}

	manager, err := NewManager(logger, aiOrchestrator, config)
	require.NoError(t, err)

	// Test component getters
	assert.NotNil(t, manager.GetWindowManager())
	assert.NotNil(t, manager.GetWorkspaceManager())
	assert.NotNil(t, manager.GetApplicationLauncher())
	assert.NotNil(t, manager.GetThemeManager())
	assert.NotNil(t, manager.GetNotificationManager())
}

func BenchmarkManagerGetStatus(b *testing.B) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel) // Reduce logging for benchmark

	// Create AI orchestrator
	aiConfig := ai.AIServiceConfig{}
	aiOrchestrator := ai.NewOrchestrator(aiConfig, logger)

	// Create desktop config
	config := DesktopConfig{
		Theme: "dark",
		WindowManager: WindowManagerConfig{
			TilingEnabled: true,
		},
		Workspaces: WorkspaceConfig{
			DefaultCount: 4,
		},
		Notifications: NotificationConfig{
			Enabled: true,
		},
	}

	manager, err := NewManager(logger, aiOrchestrator, config)
	require.NoError(b, err)

	ctx := context.Background()

	// Start manager
	err = manager.Start(ctx)
	require.NoError(b, err)
	defer manager.Stop(ctx)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := manager.GetStatus(ctx)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

func TestManagerIntegration(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	// Create AI orchestrator
	aiConfig := ai.AIServiceConfig{}
	aiOrchestrator := ai.NewOrchestrator(aiConfig, logger)

	// Initialize AI orchestrator
	ctx := context.Background()
	err := aiOrchestrator.Initialize(ctx)
	require.NoError(t, err)

	// Create desktop config
	config := DesktopConfig{
		Theme: "dark",
		WindowManager: WindowManagerConfig{
			TilingEnabled: true,
			SmartGaps:     true,
			BorderWidth:   2,
		},
		Workspaces: WorkspaceConfig{
			DefaultCount:  4,
			SmartNaming:   true,
			PersistLayout: true,
		},
		Notifications: NotificationConfig{
			Enabled:       true,
			Position:      "top-right",
			Timeout:       5 * time.Second,
			SmartGrouping: true,
		},
	}

	manager, err := NewManager(logger, aiOrchestrator, config)
	require.NoError(t, err)

	// Start manager
	err = manager.Start(ctx)
	require.NoError(t, err)
	defer manager.Stop(ctx)

	// Test full workflow
	t.Run("WindowManagement", func(t *testing.T) {
		windowManager := manager.GetWindowManager()

		// List windows
		windows, err := windowManager.ListWindows(ctx)
		require.NoError(t, err)
		assert.Len(t, windows, 2) // Mock windows

		// Focus a window
		if len(windows) > 0 {
			err = windowManager.FocusWindow(ctx, windows[0].ID)
			assert.NoError(t, err)
		}
	})

	t.Run("WorkspaceManagement", func(t *testing.T) {
		workspaceManager := manager.GetWorkspaceManager()

		// List workspaces
		workspaces, err := workspaceManager.ListWorkspaces(ctx)
		require.NoError(t, err)
		assert.Len(t, workspaces, 4) // Default count

		// Switch workspace
		err = workspaceManager.SwitchWorkspace(ctx, 2)
		assert.NoError(t, err)

		// Create new workspace
		workspace, err := workspaceManager.CreateWorkspace(ctx, "Test Workspace")
		require.NoError(t, err)
		assert.Equal(t, "Test Workspace", workspace.Name)
	})

	t.Run("ApplicationLauncher", func(t *testing.T) {
		appLauncher := manager.GetApplicationLauncher()

		// List applications
		apps, err := appLauncher.ListApplications(ctx)
		require.NoError(t, err)
		assert.Len(t, apps, 3) // Mock applications

		// Search applications
		results, err := appLauncher.SearchApplications(ctx, "firefox")
		require.NoError(t, err)
		assert.Len(t, results, 1)
		assert.Equal(t, "firefox", results[0].ID)
	})

	t.Run("ThemeManagement", func(t *testing.T) {
		themeManager := manager.GetThemeManager()

		// List themes
		themes, err := themeManager.ListThemes(ctx)
		require.NoError(t, err)
		assert.Len(t, themes, 3) // Default themes

		// Switch theme
		err = themeManager.SetTheme(ctx, "light")
		assert.NoError(t, err)

		// Get active theme
		activeTheme, err := themeManager.GetActiveTheme(ctx)
		require.NoError(t, err)
		assert.Equal(t, "light", activeTheme.ID)
	})

	t.Run("NotificationManagement", func(t *testing.T) {
		notificationManager := manager.GetNotificationManager()

		// Create notification
		notification := notificationManager.CreateSystemNotification(
			"Test Notification",
			"This is a test notification",
			"system",
		)

		err = notificationManager.ShowNotification(ctx, notification)
		assert.NoError(t, err)

		// Get notifications
		notifications, err := notificationManager.GetNotifications(ctx)
		require.NoError(t, err)
		assert.Len(t, notifications, 1)

		// Dismiss notification
		err = notificationManager.DismissNotification(ctx, notification.ID)
		assert.NoError(t, err)
	})
}
