package server

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/aios/aios/pkg/mcp/protocol"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

// DefaultSessionManager implements the SessionManager interface
type DefaultSessionManager struct {
	config   *SessionManagerConfig
	sessions map[string]protocol.Session
	logger   *logrus.Logger
	tracer   trace.Tracer
	mu       sync.RWMutex
}

// SessionManagerConfig represents session manager configuration
type SessionManagerConfig struct {
	MaxSessions     int           `json:"max_sessions"`
	SessionTimeout  time.Duration `json:"session_timeout"`
	CleanupInterval time.Duration `json:"cleanup_interval"`
}

// NewSessionManager creates a new session manager
func NewSessionManager(config *SessionManagerConfig, logger *logrus.Logger) (protocol.SessionManager, error) {
	if config.MaxSessions <= 0 {
		config.MaxSessions = 1000
	}
	if config.SessionTimeout <= 0 {
		config.SessionTimeout = 30 * time.Minute
	}
	if config.CleanupInterval <= 0 {
		config.CleanupInterval = 5 * time.Minute
	}

	manager := &DefaultSessionManager{
		config:   config,
		sessions: make(map[string]protocol.Session),
		logger:   logger,
		tracer:   otel.Tracer("mcp.session_manager"),
	}

	// Start cleanup routine
	go manager.cleanupRoutine()

	return manager, nil
}

// CreateSession creates a new session
func (sm *DefaultSessionManager) CreateSession(ctx context.Context, transport protocol.Transport, clientInfo *protocol.ClientInfo) (protocol.Session, error) {
	ctx, span := sm.tracer.Start(ctx, "session_manager.create_session")
	defer span.End()

	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Check session limit
	if len(sm.sessions) >= sm.config.MaxSessions {
		return nil, fmt.Errorf("maximum number of sessions reached: %d", sm.config.MaxSessions)
	}

	// Create new session
	sessionID := uuid.New().String()
	session := NewMCPSession(sessionID, transport, clientInfo, sm.logger)

	sm.sessions[sessionID] = session

	sm.logger.WithFields(logrus.Fields{
		"session_id":     sessionID,
		"remote_addr":    transport.GetRemoteAddress(),
		"total_sessions": len(sm.sessions),
	}).Info("Session created")

	return session, nil
}

// GetSession retrieves a session by ID
func (sm *DefaultSessionManager) GetSession(sessionID string) (protocol.Session, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	session, exists := sm.sessions[sessionID]
	if !exists {
		return nil, fmt.Errorf("session not found: %s", sessionID)
	}

	return session, nil
}

// ListSessions returns all active sessions
func (sm *DefaultSessionManager) ListSessions() []protocol.Session {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	sessions := make([]protocol.Session, 0, len(sm.sessions))
	for _, session := range sm.sessions {
		sessions = append(sessions, session)
	}

	return sessions
}

// CloseSession closes a session
func (sm *DefaultSessionManager) CloseSession(sessionID string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	session, exists := sm.sessions[sessionID]
	if !exists {
		return fmt.Errorf("session not found: %s", sessionID)
	}

	if err := session.Close(); err != nil {
		sm.logger.WithError(err).WithField("session_id", sessionID).Error("Failed to close session")
	}

	delete(sm.sessions, sessionID)

	sm.logger.WithFields(logrus.Fields{
		"session_id":     sessionID,
		"total_sessions": len(sm.sessions),
	}).Info("Session closed")

	return nil
}

// CloseAllSessions closes all sessions
func (sm *DefaultSessionManager) CloseAllSessions() error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	var lastErr error
	for sessionID, session := range sm.sessions {
		if err := session.Close(); err != nil {
			sm.logger.WithError(err).WithField("session_id", sessionID).Error("Failed to close session")
			lastErr = err
		}
	}

	sm.sessions = make(map[string]protocol.Session)

	sm.logger.Info("All sessions closed")

	return lastErr
}

// GetSessionCount returns the number of active sessions
func (sm *DefaultSessionManager) GetSessionCount() int {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	return len(sm.sessions)
}

func (sm *DefaultSessionManager) cleanupRoutine() {
	ticker := time.NewTicker(sm.config.CleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		sm.cleanupInactiveSessions()
	}
}

func (sm *DefaultSessionManager) cleanupInactiveSessions() {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	now := time.Now()
	var toRemove []string

	for sessionID, session := range sm.sessions {
		if !session.IsActive() || now.Sub(session.GetLastActivity()) > sm.config.SessionTimeout {
			toRemove = append(toRemove, sessionID)
		}
	}

	for _, sessionID := range toRemove {
		session := sm.sessions[sessionID]
		if err := session.Close(); err != nil {
			sm.logger.WithError(err).WithField("session_id", sessionID).Error("Failed to close inactive session")
		}
		delete(sm.sessions, sessionID)

		sm.logger.WithField("session_id", sessionID).Info("Cleaned up inactive session")
	}

	if len(toRemove) > 0 {
		sm.logger.WithField("cleaned_sessions", len(toRemove)).Debug("Session cleanup completed")
	}
}

// MCPSession implements the Session interface
type MCPSession struct {
	id           string
	transport    protocol.Transport
	clientInfo   *protocol.ClientInfo
	serverInfo   *protocol.ServerInfo
	capabilities *protocol.Capabilities
	lastActivity time.Time
	active       bool
	logger       *logrus.Logger
	tracer       trace.Tracer
	mu           sync.RWMutex
}

// NewMCPSession creates a new MCP session
func NewMCPSession(id string, transport protocol.Transport, clientInfo *protocol.ClientInfo, logger *logrus.Logger) *MCPSession {
	return &MCPSession{
		id:           id,
		transport:    transport,
		clientInfo:   clientInfo,
		lastActivity: time.Now(),
		active:       true,
		logger:       logger,
		tracer:       otel.Tracer("mcp.session"),
	}
}

// GetID returns the session ID
func (s *MCPSession) GetID() string {
	return s.id
}

// GetClientInfo returns client information
func (s *MCPSession) GetClientInfo() *protocol.ClientInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.clientInfo
}

// GetServerInfo returns server information
func (s *MCPSession) GetServerInfo() *protocol.ServerInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.serverInfo
}

// GetCapabilities returns session capabilities
func (s *MCPSession) GetCapabilities() *protocol.Capabilities {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.capabilities
}

// GetTransport returns the underlying transport
func (s *MCPSession) GetTransport() protocol.Transport {
	return s.transport
}

// SendRequest sends a request and returns a response channel
func (s *MCPSession) SendRequest(ctx context.Context, request protocol.Request) (<-chan protocol.Response, error) {
	ctx, span := s.tracer.Start(ctx, "session.send_request")
	defer span.End()

	s.updateActivity()

	// Send the request
	if err := s.transport.Send(ctx, request); err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	// Create response channel
	responseCh := make(chan protocol.Response, 1)

	// TODO: Implement request/response correlation
	// For now, we'll return an empty channel
	go func() {
		defer close(responseCh)
		// This would normally wait for the corresponding response
		// and send it through the channel
	}()

	return responseCh, nil
}

// SendResponse sends a response
func (s *MCPSession) SendResponse(ctx context.Context, response protocol.Response) error {
	ctx, span := s.tracer.Start(ctx, "session.send_response")
	defer span.End()

	s.updateActivity()

	return s.transport.Send(ctx, response)
}

// SendNotification sends a notification
func (s *MCPSession) SendNotification(ctx context.Context, notification protocol.Notification) error {
	ctx, span := s.tracer.Start(ctx, "session.send_notification")
	defer span.End()

	s.updateActivity()

	return s.transport.Send(ctx, notification)
}

// GetLastActivity returns the last activity time
func (s *MCPSession) GetLastActivity() time.Time {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.lastActivity
}

// IsActive returns whether the session is active
func (s *MCPSession) IsActive() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.active && s.transport.IsConnected()
}

// Close closes the session
func (s *MCPSession) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.active = false
	return s.transport.Close()
}

// SetServerInfo sets the server information
func (s *MCPSession) SetServerInfo(serverInfo *protocol.ServerInfo) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.serverInfo = serverInfo
}

// SetCapabilities sets the session capabilities
func (s *MCPSession) SetCapabilities(capabilities *protocol.Capabilities) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.capabilities = capabilities
}

func (s *MCPSession) updateActivity() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.lastActivity = time.Now()
}
