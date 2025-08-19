package collaboration

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
)

// Communication methods for collaboration engine

// SendMessage sends a message to a channel
func (ce *DefaultCollaborationEngine) SendMessage(message *Message) (*Message, error) {
	_, span := ce.tracer.Start(context.Background(), "collaboration.send_message")
	defer span.End()

	if message.ID == "" {
		message.ID = uuid.New().String()
	}

	// Set timestamp
	message.CreatedAt = time.Now()

	// Validate message
	if err := ce.validateMessage(message); err != nil {
		return nil, fmt.Errorf("message validation failed: %w", err)
	}

	// Verify channel exists
	ce.mu.RLock()
	channel, exists := ce.channels[message.ChannelID]
	ce.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("channel not found: %s", message.ChannelID)
	}

	// Check if user is member of channel
	isMember := false
	for _, memberID := range channel.Members {
		if memberID == message.UserID {
			isMember = true
			break
		}
	}

	if !isMember && channel.Type != ChannelTypePublic {
		return nil, fmt.Errorf("user %s is not a member of channel %s", message.UserID, message.ChannelID)
	}

	ce.mu.Lock()
	ce.messages[message.ID] = message
	ce.messagesByChannel[message.ChannelID] = append(ce.messagesByChannel[message.ChannelID], message.ID)

	// Update channel's last message
	channel.LastMessage = message
	channel.UpdatedAt = time.Now()
	ce.mu.Unlock()

	// Log activity
	ce.logActivity(&Activity{
		Type:        ActivityTypeMessage,
		UserID:      message.UserID,
		Username:    message.Username,
		TeamID:      channel.TeamID,
		Action:      "sent",
		Resource:    &ActivityResource{Type: "message", ID: message.ID, Name: "message"},
		Description: fmt.Sprintf("Sent message in #%s", channel.Name),
		Timestamp:   time.Now(),
	})

	span.SetAttributes(
		attribute.String("message.id", message.ID),
		attribute.String("channel.id", message.ChannelID),
		attribute.String("user.id", message.UserID),
		attribute.String("message.type", string(message.Type)),
	)

	ce.logger.WithFields(map[string]interface{}{
		"message_id": message.ID,
		"channel_id": message.ChannelID,
		"user_id":    message.UserID,
		"type":       message.Type,
	}).Info("Message sent successfully")

	return message, nil
}

// GetMessages retrieves messages from a channel
func (ce *DefaultCollaborationEngine) GetMessages(channelID string, filter *MessageFilter) ([]*Message, error) {
	_, span := ce.tracer.Start(context.Background(), "collaboration.get_messages")
	defer span.End()

	// Verify channel exists
	ce.mu.RLock()
	_, exists := ce.channels[channelID]
	if !exists {
		ce.mu.RUnlock()
		return nil, fmt.Errorf("channel not found: %s", channelID)
	}

	var messages []*Message
	if messageIDs, exists := ce.messagesByChannel[channelID]; exists {
		for _, messageID := range messageIDs {
			if message, exists := ce.messages[messageID]; exists {
				if ce.matchesMessageFilter(message, filter) {
					messages = append(messages, message)
				}
			}
		}
	}
	ce.mu.RUnlock()

	// Sort messages by creation time (newest first)
	sort.Slice(messages, func(i, j int) bool {
		return messages[i].CreatedAt.After(messages[j].CreatedAt)
	})

	// Apply pagination
	if filter != nil {
		if filter.Offset > 0 && filter.Offset < len(messages) {
			messages = messages[filter.Offset:]
		}
		if filter.Limit > 0 && filter.Limit < len(messages) {
			messages = messages[:filter.Limit]
		}
	}

	span.SetAttributes(
		attribute.String("channel.id", channelID),
		attribute.Int("messages.count", len(messages)),
	)

	return messages, nil
}

// CreateChannel creates a new communication channel
func (ce *DefaultCollaborationEngine) CreateChannel(channel *Channel) (*Channel, error) {
	_, span := ce.tracer.Start(context.Background(), "collaboration.create_channel")
	defer span.End()

	if channel.ID == "" {
		channel.ID = uuid.New().String()
	}

	// Set timestamps
	now := time.Now()
	channel.CreatedAt = now
	channel.UpdatedAt = now

	// Set default settings if not provided
	if channel.Settings == nil {
		channel.Settings = &ChannelSettings{
			IsArchived:       false,
			AllowThreads:     true,
			AllowFileUploads: true,
			RetentionDays:    90,
		}
	}

	// Validate channel
	if err := ce.validateChannel(channel); err != nil {
		return nil, fmt.Errorf("channel validation failed: %w", err)
	}

	// Verify team exists
	ce.mu.RLock()
	team, exists := ce.teams[channel.TeamID]
	ce.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("team not found: %s", channel.TeamID)
	}

	ce.mu.Lock()
	ce.channels[channel.ID] = channel
	ce.channelsByTeam[channel.TeamID] = append(ce.channelsByTeam[channel.TeamID], channel.ID)
	ce.messagesByChannel[channel.ID] = []string{}
	ce.mu.Unlock()

	// Log activity
	ce.logActivity(&Activity{
		Type:        ActivityTypeTeam,
		UserID:      channel.CreatedBy,
		TeamID:      channel.TeamID,
		Action:      "channel_created",
		Resource:    &ActivityResource{Type: "channel", ID: channel.ID, Name: channel.Name},
		Description: fmt.Sprintf("Created channel #%s in team '%s'", channel.Name, team.Name),
		Timestamp:   now,
	})

	span.SetAttributes(
		attribute.String("channel.id", channel.ID),
		attribute.String("channel.name", channel.Name),
		attribute.String("channel.type", string(channel.Type)),
		attribute.String("team.id", channel.TeamID),
	)

	ce.logger.WithFields(map[string]interface{}{
		"channel_id":   channel.ID,
		"channel_name": channel.Name,
		"channel_type": channel.Type,
		"team_id":      channel.TeamID,
	}).Info("Channel created successfully")

	return channel, nil
}

// GetChannel retrieves a channel by ID
func (ce *DefaultCollaborationEngine) GetChannel(channelID string) (*Channel, error) {
	_, span := ce.tracer.Start(context.Background(), "collaboration.get_channel")
	defer span.End()

	ce.mu.RLock()
	channel, exists := ce.channels[channelID]
	ce.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("channel not found: %s", channelID)
	}

	span.SetAttributes(attribute.String("channel.id", channelID))

	return channel, nil
}

// ListChannels lists channels for a team
func (ce *DefaultCollaborationEngine) ListChannels(teamID string, filter *ChannelFilter) ([]*Channel, error) {
	_, span := ce.tracer.Start(context.Background(), "collaboration.list_channels")
	defer span.End()

	// Verify team exists
	ce.mu.RLock()
	_, exists := ce.teams[teamID]
	if !exists {
		ce.mu.RUnlock()
		return nil, fmt.Errorf("team not found: %s", teamID)
	}

	var channels []*Channel
	if channelIDs, exists := ce.channelsByTeam[teamID]; exists {
		for _, channelID := range channelIDs {
			if channel, exists := ce.channels[channelID]; exists {
				if ce.matchesChannelFilter(channel, filter) {
					channels = append(channels, channel)
				}
			}
		}
	}
	ce.mu.RUnlock()

	// Sort channels by name
	sort.Slice(channels, func(i, j int) bool {
		return channels[i].Name < channels[j].Name
	})

	// Apply pagination
	if filter != nil {
		if filter.Offset > 0 && filter.Offset < len(channels) {
			channels = channels[filter.Offset:]
		}
		if filter.Limit > 0 && filter.Limit < len(channels) {
			channels = channels[:filter.Limit]
		}
	}

	span.SetAttributes(
		attribute.String("team.id", teamID),
		attribute.Int("channels.count", len(channels)),
	)

	return channels, nil
}

// SendNotification sends a notification to a user
func (ce *DefaultCollaborationEngine) SendNotification(notification *Notification) error {
	_, span := ce.tracer.Start(context.Background(), "collaboration.send_notification")
	defer span.End()

	if notification.ID == "" {
		notification.ID = uuid.New().String()
	}

	// Set timestamp
	notification.CreatedAt = time.Now()

	// Validate notification
	if err := ce.validateNotification(notification); err != nil {
		return fmt.Errorf("notification validation failed: %w", err)
	}

	ce.mu.Lock()
	ce.notifications[notification.ID] = notification
	ce.notificationsByUser[notification.UserID] = append(ce.notificationsByUser[notification.UserID], notification.ID)
	ce.mu.Unlock()

	// Log activity
	ce.logActivity(&Activity{
		Type:        ActivityTypeNotification,
		UserID:      notification.UserID,
		Action:      "received",
		Resource:    &ActivityResource{Type: "notification", ID: notification.ID, Name: notification.Title},
		Description: fmt.Sprintf("Received notification: %s", notification.Title),
		Timestamp:   time.Now(),
	})

	span.SetAttributes(
		attribute.String("notification.id", notification.ID),
		attribute.String("user.id", notification.UserID),
		attribute.String("notification.type", string(notification.Type)),
		attribute.String("notification.priority", string(notification.Priority)),
	)

	ce.logger.WithFields(map[string]interface{}{
		"notification_id": notification.ID,
		"user_id":         notification.UserID,
		"type":            notification.Type,
		"priority":        notification.Priority,
		"title":           notification.Title,
	}).Info("Notification sent successfully")

	return nil
}

// GetNotifications retrieves notifications for a user
func (ce *DefaultCollaborationEngine) GetNotifications(userID string, filter *NotificationFilter) ([]*Notification, error) {
	_, span := ce.tracer.Start(context.Background(), "collaboration.get_notifications")
	defer span.End()

	ce.mu.RLock()
	var notifications []*Notification
	if notificationIDs, exists := ce.notificationsByUser[userID]; exists {
		for _, notificationID := range notificationIDs {
			if notification, exists := ce.notifications[notificationID]; exists {
				if ce.matchesNotificationFilter(notification, filter) {
					notifications = append(notifications, notification)
				}
			}
		}
	}
	ce.mu.RUnlock()

	// Sort notifications by creation time (newest first)
	sort.Slice(notifications, func(i, j int) bool {
		return notifications[i].CreatedAt.After(notifications[j].CreatedAt)
	})

	// Apply pagination
	if filter != nil {
		if filter.Offset > 0 && filter.Offset < len(notifications) {
			notifications = notifications[filter.Offset:]
		}
		if filter.Limit > 0 && filter.Limit < len(notifications) {
			notifications = notifications[:filter.Limit]
		}
	}

	span.SetAttributes(
		attribute.String("user.id", userID),
		attribute.Int("notifications.count", len(notifications)),
	)

	return notifications, nil
}

// MarkNotificationRead marks a notification as read
func (ce *DefaultCollaborationEngine) MarkNotificationRead(notificationID string) error {
	_, span := ce.tracer.Start(context.Background(), "collaboration.mark_notification_read")
	defer span.End()

	ce.mu.Lock()
	notification, exists := ce.notifications[notificationID]
	if !exists {
		ce.mu.Unlock()
		return fmt.Errorf("notification not found: %s", notificationID)
	}

	notification.IsRead = true
	now := time.Now()
	notification.ReadAt = &now
	ce.mu.Unlock()

	span.SetAttributes(attribute.String("notification.id", notificationID))

	ce.logger.WithField("notification_id", notificationID).Info("Notification marked as read")

	return nil
}

// GetNotificationSettings retrieves notification settings for a user
func (ce *DefaultCollaborationEngine) GetNotificationSettings(userID string) (*NotificationSettings, error) {
	_, span := ce.tracer.Start(context.Background(), "collaboration.get_notification_settings")
	defer span.End()

	ce.mu.RLock()
	settings, exists := ce.notificationSettings[userID]
	ce.mu.RUnlock()

	if !exists {
		// Return default settings
		settings = &NotificationSettings{
			UserID: userID,
			EmailNotifications: &EmailNotificationSettings{
				Enabled:    true,
				Frequency:  "immediate",
				Categories: []string{"mention", "assignment", "deadline"},
			},
			PushNotifications: &PushNotificationSettings{
				Enabled:    true,
				Categories: []string{"mention", "assignment"},
				Sound:      true,
				Vibration:  true,
			},
			InAppNotifications: &InAppNotificationSettings{
				Enabled:    true,
				Categories: []string{"mention", "assignment", "deadline", "approval"},
				ShowBadge:  true,
				AutoRead:   false,
			},
			CategorySettings: map[string]*CategorySettings{
				"mention": {
					Enabled:  true,
					Priority: NotificationPriorityHigh,
					Channels: []string{"email", "push", "in_app"},
				},
				"assignment": {
					Enabled:  true,
					Priority: NotificationPriorityNormal,
					Channels: []string{"email", "in_app"},
				},
				"deadline": {
					Enabled:  true,
					Priority: NotificationPriorityHigh,
					Channels: []string{"email", "push", "in_app"},
				},
			},
			UpdatedAt: time.Now(),
		}
	}

	span.SetAttributes(attribute.String("user.id", userID))

	return settings, nil
}

// UpdateNotificationSettings updates notification settings for a user
func (ce *DefaultCollaborationEngine) UpdateNotificationSettings(userID string, settings *NotificationSettings) error {
	_, span := ce.tracer.Start(context.Background(), "collaboration.update_notification_settings")
	defer span.End()

	settings.UserID = userID
	settings.UpdatedAt = time.Now()

	ce.mu.Lock()
	ce.notificationSettings[userID] = settings
	ce.mu.Unlock()

	span.SetAttributes(attribute.String("user.id", userID))

	ce.logger.WithField("user_id", userID).Info("Notification settings updated successfully")

	return nil
}
