package client

import (
	"context"
	"sync"
	"time"

	"github.com/aios/aios/pkg/mcp/protocol"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

// ClientSession implements the Session interface for MCP clients
type ClientSession struct {
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

// NewClientSession creates a new client session
func NewClientSession(id string, transport protocol.Transport, clientInfo *protocol.ClientInfo, logger *logrus.Logger) *ClientSession {
	return &ClientSession{
		id:           id,
		transport:    transport,
		clientInfo:   clientInfo,
		lastActivity: time.Now(),
		active:       true,
		logger:       logger,
		tracer:       otel.Tracer("mcp.client.session"),
	}
}

// GetID returns the session ID
func (s *ClientSession) GetID() string {
	return s.id
}

// GetClientInfo returns client information
func (s *ClientSession) GetClientInfo() *protocol.ClientInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.clientInfo
}

// GetServerInfo returns server information
func (s *ClientSession) GetServerInfo() *protocol.ServerInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.serverInfo
}

// GetCapabilities returns session capabilities
func (s *ClientSession) GetCapabilities() *protocol.Capabilities {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.capabilities
}

// GetTransport returns the underlying transport
func (s *ClientSession) GetTransport() protocol.Transport {
	return s.transport
}

// SendRequest sends a request and returns a response channel
func (s *ClientSession) SendRequest(ctx context.Context, request protocol.Request) (<-chan protocol.Response, error) {
	ctx, span := s.tracer.Start(ctx, "client_session.send_request")
	defer span.End()

	s.updateActivity()

	// Send the request
	if err := s.transport.Send(ctx, request); err != nil {
		return nil, err
	}

	// For client sessions, we don't handle response correlation here
	// The client handles this at a higher level
	responseCh := make(chan protocol.Response, 1)
	close(responseCh) // Close immediately as we don't use this channel

	return responseCh, nil
}

// SendResponse sends a response (not typically used by clients)
func (s *ClientSession) SendResponse(ctx context.Context, response protocol.Response) error {
	ctx, span := s.tracer.Start(ctx, "client_session.send_response")
	defer span.End()

	s.updateActivity()

	return s.transport.Send(ctx, response)
}

// SendNotification sends a notification
func (s *ClientSession) SendNotification(ctx context.Context, notification protocol.Notification) error {
	ctx, span := s.tracer.Start(ctx, "client_session.send_notification")
	defer span.End()

	s.updateActivity()

	return s.transport.Send(ctx, notification)
}

// GetLastActivity returns the last activity time
func (s *ClientSession) GetLastActivity() time.Time {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.lastActivity
}

// IsActive returns whether the session is active
func (s *ClientSession) IsActive() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.active && s.transport.IsConnected()
}

// Close closes the session
func (s *ClientSession) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.active = false
	return s.transport.Close()
}

// SetServerInfo sets the server information
func (s *ClientSession) SetServerInfo(serverInfo *protocol.ServerInfo) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.serverInfo = serverInfo
}

// SetCapabilities sets the session capabilities
func (s *ClientSession) SetCapabilities(capabilities *protocol.Capabilities) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.capabilities = capabilities
}

func (s *ClientSession) updateActivity() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.lastActivity = time.Now()
}
