package desktop

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/aios/aios/pkg/models"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

// NotificationManager handles desktop notifications
type NotificationManager struct {
	logger        *logrus.Logger
	tracer        trace.Tracer
	config        NotificationConfig
	notifications map[string]*models.Notification
	mu            sync.RWMutex
	running       bool
	stopCh        chan struct{}
}

// NewNotificationManager creates a new notification manager
func NewNotificationManager(logger *logrus.Logger, config NotificationConfig) (*NotificationManager, error) {
	tracer := otel.Tracer("notification-manager")

	return &NotificationManager{
		logger:        logger,
		tracer:        tracer,
		config:        config,
		notifications: make(map[string]*models.Notification),
		stopCh:        make(chan struct{}),
	}, nil
}

// Start initializes the notification manager
func (nm *NotificationManager) Start(ctx context.Context) error {
	ctx, span := nm.tracer.Start(ctx, "notificationManager.Start")
	defer span.End()

	nm.mu.Lock()
	defer nm.mu.Unlock()

	if nm.running {
		return fmt.Errorf("notification manager is already running")
	}

	nm.logger.Info("Starting notification manager")

	// Start cleanup routine
	go nm.cleanupNotifications()

	nm.running = true
	nm.logger.Info("Notification manager started successfully")

	return nil
}

// Stop shuts down the notification manager
func (nm *NotificationManager) Stop(ctx context.Context) error {
	ctx, span := nm.tracer.Start(ctx, "notificationManager.Stop")
	defer span.End()

	nm.mu.Lock()
	defer nm.mu.Unlock()

	if !nm.running {
		return nil
	}

	nm.logger.Info("Stopping notification manager")

	close(nm.stopCh)
	nm.running = false
	nm.logger.Info("Notification manager stopped")

	return nil
}

// GetStatus returns the current notification manager status
func (nm *NotificationManager) GetStatus(ctx context.Context) (*models.NotificationStatus, error) {
	ctx, span := nm.tracer.Start(ctx, "notificationManager.GetStatus")
	defer span.End()

	nm.mu.RLock()
	defer nm.mu.RUnlock()

	notifications := make([]*models.Notification, 0, len(nm.notifications))
	recentCount := 0

	for _, notification := range nm.notifications {
		notifications = append(notifications, notification)
		if time.Since(notification.CreatedAt) < 24*time.Hour {
			recentCount++
		}
	}

	return &models.NotificationStatus{
		Running:           nm.running,
		NotificationCount: len(nm.notifications),
		Notifications:     notifications,
		RecentCount:       recentCount,
		Timestamp:         time.Now(),
	}, nil
}

// ShowNotification displays a new notification
func (nm *NotificationManager) ShowNotification(ctx context.Context, notification *models.Notification) error {
	ctx, span := nm.tracer.Start(ctx, "notificationManager.ShowNotification")
	defer span.End()

	if !nm.config.Enabled {
		return nil // Notifications disabled
	}

	nm.logger.WithFields(logrus.Fields{
		"title":    notification.Title,
		"category": notification.Category,
		"priority": notification.Priority,
	}).Info("Showing notification")

	nm.mu.Lock()
	defer nm.mu.Unlock()

	// Generate ID if not provided
	if notification.ID == "" {
		notification.ID = uuid.New().String()
	}

	// Set creation time
	notification.CreatedAt = time.Now()

	// Set default timeout if not provided
	if notification.Timeout == 0 {
		notification.Timeout = nm.config.Timeout
	}

	// Store notification
	nm.notifications[notification.ID] = notification

	// TODO: Display notification in desktop environment
	// This would involve:
	// - Creating notification popup
	// - Playing notification sound
	// - Showing in notification center

	// Auto-dismiss after timeout (if not persistent)
	if !notification.Persistent && notification.Timeout > 0 {
		go nm.autoDismiss(notification.ID, notification.Timeout)
	}

	nm.logger.WithFields(logrus.Fields{
		"notification_id": notification.ID,
		"title":           notification.Title,
	}).Info("Notification shown")

	return nil
}

// DismissNotification dismisses a notification
func (nm *NotificationManager) DismissNotification(ctx context.Context, id string) error {
	ctx, span := nm.tracer.Start(ctx, "notificationManager.DismissNotification")
	defer span.End()

	nm.logger.WithField("notification_id", id).Info("Dismissing notification")

	nm.mu.Lock()
	defer nm.mu.Unlock()

	notification, exists := nm.notifications[id]
	if !exists {
		return fmt.Errorf("notification %s not found", id)
	}

	// Mark as dismissed
	notification.Dismissed = true
	now := time.Now()
	notification.DismissedAt = &now

	nm.logger.WithFields(logrus.Fields{
		"notification_id": id,
		"title":           notification.Title,
	}).Info("Notification dismissed")

	return nil
}

// DismissAllNotifications dismisses all notifications
func (nm *NotificationManager) DismissAllNotifications(ctx context.Context) error {
	ctx, span := nm.tracer.Start(ctx, "notificationManager.DismissAllNotifications")
	defer span.End()

	nm.logger.Info("Dismissing all notifications")

	nm.mu.Lock()
	defer nm.mu.Unlock()

	count := 0
	now := time.Now()

	for _, notification := range nm.notifications {
		if !notification.Dismissed {
			notification.Dismissed = true
			notification.DismissedAt = &now
			count++
		}
	}

	nm.logger.WithField("dismissed_count", count).Info("All notifications dismissed")

	return nil
}

// GetNotifications returns all notifications
func (nm *NotificationManager) GetNotifications(ctx context.Context) ([]*models.Notification, error) {
	ctx, span := nm.tracer.Start(ctx, "notificationManager.GetNotifications")
	defer span.End()

	nm.mu.RLock()
	defer nm.mu.RUnlock()

	notifications := make([]*models.Notification, 0, len(nm.notifications))
	for _, notification := range nm.notifications {
		notifications = append(notifications, notification)
	}

	return notifications, nil
}

// GetActiveNotifications returns non-dismissed notifications
func (nm *NotificationManager) GetActiveNotifications(ctx context.Context) ([]*models.Notification, error) {
	ctx, span := nm.tracer.Start(ctx, "notificationManager.GetActiveNotifications")
	defer span.End()

	nm.mu.RLock()
	defer nm.mu.RUnlock()

	notifications := []*models.Notification{}
	for _, notification := range nm.notifications {
		if !notification.Dismissed {
			notifications = append(notifications, notification)
		}
	}

	return notifications, nil
}

// GetNotification returns a specific notification
func (nm *NotificationManager) GetNotification(ctx context.Context, id string) (*models.Notification, error) {
	ctx, span := nm.tracer.Start(ctx, "notificationManager.GetNotification")
	defer span.End()

	nm.mu.RLock()
	defer nm.mu.RUnlock()

	notification, exists := nm.notifications[id]
	if !exists {
		return nil, fmt.Errorf("notification %s not found", id)
	}

	return notification, nil
}

// ExecuteNotificationAction executes a notification action
func (nm *NotificationManager) ExecuteNotificationAction(ctx context.Context, notificationID, actionID string) error {
	ctx, span := nm.tracer.Start(ctx, "notificationManager.ExecuteNotificationAction")
	defer span.End()

	nm.logger.WithFields(logrus.Fields{
		"notification_id": notificationID,
		"action_id":       actionID,
	}).Info("Executing notification action")

	nm.mu.RLock()
	notification, exists := nm.notifications[notificationID]
	nm.mu.RUnlock()

	if !exists {
		return fmt.Errorf("notification %s not found", notificationID)
	}

	// Find action
	var action *models.NotificationAction
	for _, a := range notification.Actions {
		if a.ID == actionID {
			action = &a
			break
		}
	}

	if action == nil {
		return fmt.Errorf("action %s not found in notification %s", actionID, notificationID)
	}

	// TODO: Execute the actual action
	// This would involve calling the appropriate system function

	// Auto-dismiss notification after action
	nm.DismissNotification(ctx, notificationID)

	nm.logger.WithFields(logrus.Fields{
		"notification_id": notificationID,
		"action_id":       actionID,
		"action_label":    action.Label,
	}).Info("Notification action executed")

	return nil
}

// CreateSystemNotification creates a system notification
func (nm *NotificationManager) CreateSystemNotification(title, body, category string) *models.Notification {
	return &models.Notification{
		Title:      title,
		Body:       body,
		Category:   category,
		Priority:   "normal",
		Source:     "system",
		Persistent: false,
		Actions:    []models.NotificationAction{},
	}
}

// CreateAppNotification creates an application notification
func (nm *NotificationManager) CreateAppNotification(title, body, source string, actions []models.NotificationAction) *models.Notification {
	return &models.Notification{
		Title:      title,
		Body:       body,
		Category:   "application",
		Priority:   "normal",
		Source:     source,
		Persistent: false,
		Actions:    actions,
	}
}

// Helper methods

func (nm *NotificationManager) autoDismiss(id string, timeout time.Duration) {
	time.Sleep(timeout)

	nm.mu.Lock()
	defer nm.mu.Unlock()

	if notification, exists := nm.notifications[id]; exists && !notification.Dismissed {
		notification.Dismissed = true
		now := time.Now()
		notification.DismissedAt = &now

		nm.logger.WithField("notification_id", id).Debug("Notification auto-dismissed")
	}
}

func (nm *NotificationManager) cleanupNotifications() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			nm.performCleanup()

		case <-nm.stopCh:
			nm.logger.Debug("Notification cleanup stopped")
			return
		}
	}
}

func (nm *NotificationManager) performCleanup() {
	nm.mu.Lock()
	defer nm.mu.Unlock()

	// Remove old dismissed notifications (older than 7 days)
	cutoff := time.Now().Add(-7 * 24 * time.Hour)
	removed := 0

	for id, notification := range nm.notifications {
		if notification.Dismissed && notification.DismissedAt != nil && notification.DismissedAt.Before(cutoff) {
			delete(nm.notifications, id)
			removed++
		}
	}

	if removed > 0 {
		nm.logger.WithField("removed_count", removed).Debug("Old notifications cleaned up")
	}
}
