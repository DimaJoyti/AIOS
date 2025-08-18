package server

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/aios/aios/pkg/mcp/protocol"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// MCPServer implements an MCP server
type MCPServer struct {
	config           *ServerConfig
	sessionManager   protocol.SessionManager
	messageRouter    protocol.MessageRouter
	securityManager  protocol.SecurityManager
	metricsCollector protocol.MetricsCollector
	eventEmitter     protocol.EventEmitter
	logger           *logrus.Logger
	tracer           trace.Tracer
	listener         net.Listener
	running          bool
	mu               sync.RWMutex
}

// ServerConfig represents MCP server configuration
type ServerConfig struct {
	Address         string                  `json:"address"`
	Port            int                     `json:"port"`
	Protocol        protocol.ProtocolConfig `json:"protocol"`
	TLS             TLSConfig               `json:"tls"`
	MaxConnections  int                     `json:"max_connections"`
	ReadTimeout     time.Duration           `json:"read_timeout"`
	WriteTimeout    time.Duration           `json:"write_timeout"`
	IdleTimeout     time.Duration           `json:"idle_timeout"`
	ShutdownTimeout time.Duration           `json:"shutdown_timeout"`
	EnableMetrics   bool                    `json:"enable_metrics"`
	EnableEvents    bool                    `json:"enable_events"`
	Middleware      []string                `json:"middleware"`
	Metadata        map[string]interface{}  `json:"metadata"`
}

// TLSConfig represents TLS configuration
type TLSConfig struct {
	Enabled  bool   `json:"enabled"`
	CertFile string `json:"cert_file"`
	KeyFile  string `json:"key_file"`
	CAFile   string `json:"ca_file"`
}

// NewMCPServer creates a new MCP server
func NewMCPServer(config *ServerConfig, logger *logrus.Logger) (*MCPServer, error) {
	if config == nil {
		return nil, fmt.Errorf("server config cannot be nil")
	}

	// Set defaults
	if config.Address == "" {
		config.Address = "localhost"
	}
	if config.Port == 0 {
		config.Port = 8080
	}
	if config.MaxConnections == 0 {
		config.MaxConnections = 1000
	}
	if config.ReadTimeout == 0 {
		config.ReadTimeout = 30 * time.Second
	}
	if config.WriteTimeout == 0 {
		config.WriteTimeout = 30 * time.Second
	}
	if config.IdleTimeout == 0 {
		config.IdleTimeout = 120 * time.Second
	}
	if config.ShutdownTimeout == 0 {
		config.ShutdownTimeout = 30 * time.Second
	}

	server := &MCPServer{
		config:  config,
		logger:  logger,
		tracer:  otel.Tracer("mcp.server"),
		running: false,
	}

	// Initialize components
	if err := server.initializeComponents(); err != nil {
		return nil, fmt.Errorf("failed to initialize server components: %w", err)
	}

	return server, nil
}

// Start starts the MCP server
func (s *MCPServer) Start(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return fmt.Errorf("server is already running")
	}

	ctx, span := s.tracer.Start(ctx, "mcp_server.start")
	defer span.End()

	s.logger.WithFields(logrus.Fields{
		"address": s.config.Address,
		"port":    s.config.Port,
	}).Info("Starting MCP server")

	// Create listener
	address := fmt.Sprintf("%s:%d", s.config.Address, s.config.Port)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to create listener: %w", err)
	}

	s.listener = listener
	s.running = true

	span.SetAttributes(
		attribute.String("server.address", s.config.Address),
		attribute.Int("server.port", s.config.Port),
	)

	// Start accepting connections
	go s.acceptConnections(ctx)

	s.logger.WithField("address", address).Info("MCP server started successfully")

	return nil
}

// Stop stops the MCP server
func (s *MCPServer) Stop(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return nil
	}

	ctx, span := s.tracer.Start(ctx, "mcp_server.stop")
	defer span.End()

	s.logger.Info("Stopping MCP server")

	// Set shutdown timeout
	shutdownCtx, cancel := context.WithTimeout(ctx, s.config.ShutdownTimeout)
	defer cancel()

	// Close listener
	if s.listener != nil {
		if err := s.listener.Close(); err != nil {
			s.logger.WithError(err).Error("Failed to close listener")
		}
	}

	// Close all sessions
	if s.sessionManager != nil {
		if err := s.sessionManager.CloseAllSessions(); err != nil {
			s.logger.WithError(err).Error("Failed to close all sessions")
		}
	}

	// Wait for shutdown or timeout
	done := make(chan struct{})
	go func() {
		// Additional cleanup can be done here
		close(done)
	}()

	select {
	case <-done:
		s.logger.Info("MCP server stopped gracefully")
	case <-shutdownCtx.Done():
		s.logger.Warn("MCP server shutdown timeout exceeded")
	}

	s.running = false

	return nil
}

// IsRunning returns whether the server is running
func (s *MCPServer) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.running
}

// GetConfig returns the server configuration
func (s *MCPServer) GetConfig() *ServerConfig {
	return s.config
}

// GetSessionManager returns the session manager
func (s *MCPServer) GetSessionManager() protocol.SessionManager {
	return s.sessionManager
}

// GetMessageRouter returns the message router
func (s *MCPServer) GetMessageRouter() protocol.MessageRouter {
	return s.messageRouter
}

// RegisterHandler registers a message handler
func (s *MCPServer) RegisterHandler(methods []string, handler protocol.Handler) error {
	return s.messageRouter.RegisterHandler(methods, handler)
}

// UnregisterHandler unregisters a message handler
func (s *MCPServer) UnregisterHandler(methods []string) error {
	return s.messageRouter.UnregisterHandler(methods)
}

// GetMetrics returns server metrics
func (s *MCPServer) GetMetrics() map[string]interface{} {
	if s.metricsCollector == nil {
		return make(map[string]interface{})
	}
	return s.metricsCollector.GetMetrics()
}

// Helper methods

func (s *MCPServer) initializeComponents() error {
	// Initialize session manager
	sessionManager, err := NewSessionManager(&SessionManagerConfig{
		MaxSessions:    s.config.MaxConnections,
		SessionTimeout: s.config.IdleTimeout,
	}, s.logger)
	if err != nil {
		return fmt.Errorf("failed to create session manager: %w", err)
	}
	s.sessionManager = sessionManager

	// Initialize message router
	messageRouter, err := NewMessageRouter(s.logger)
	if err != nil {
		return fmt.Errorf("failed to create message router: %w", err)
	}
	s.messageRouter = messageRouter

	// Initialize security manager if enabled
	if s.config.Protocol.Security.EnableAuthentication || s.config.Protocol.Security.EnableAuthorization {
		// TODO: Implement protocol.SecurityManager
		// For now, we'll skip security manager initialization
		s.logger.Info("Security enabled but not yet implemented")
	}

	// Initialize metrics collector if enabled
	if s.config.EnableMetrics {
		// TODO: Implement metrics collector
		// metricsCollector, err := NewMetricsCollector(s.logger)
		// if err != nil {
		//     return fmt.Errorf("failed to create metrics collector: %w", err)
		// }
		// s.metricsCollector = metricsCollector
		s.logger.Info("Metrics collection enabled but not yet implemented")
	}

	// Initialize event emitter if enabled
	if s.config.EnableEvents {
		// TODO: Implement event emitter
		// eventEmitter, err := NewEventEmitter(s.logger)
		// if err != nil {
		//     return fmt.Errorf("failed to create event emitter: %w", err)
		// }
		// s.eventEmitter = eventEmitter
		s.logger.Info("Event emission enabled but not yet implemented")
	}

	// Register default handlers
	if err := s.registerDefaultHandlers(); err != nil {
		return fmt.Errorf("failed to register default handlers: %w", err)
	}

	return nil
}

func (s *MCPServer) registerDefaultHandlers() error {
	// Register initialize handler
	initHandler := NewInitializeHandler(s.config, s.logger)
	if err := s.messageRouter.RegisterHandler([]string{protocol.MethodInitialize}, initHandler); err != nil {
		return fmt.Errorf("failed to register initialize handler: %w", err)
	}

	// Register ping handler
	pingHandler := NewPingHandler(s.logger)
	if err := s.messageRouter.RegisterHandler([]string{protocol.MethodPing}, pingHandler); err != nil {
		return fmt.Errorf("failed to register ping handler: %w", err)
	}

	return nil
}

func (s *MCPServer) acceptConnections(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			conn, err := s.listener.Accept()
			if err != nil {
				if s.running {
					s.logger.WithError(err).Error("Failed to accept connection")
				}
				continue
			}

			// Handle connection in a separate goroutine
			go s.handleConnection(ctx, conn)
		}
	}
}

func (s *MCPServer) handleConnection(ctx context.Context, conn net.Conn) {
	ctx, span := s.tracer.Start(ctx, "mcp_server.handle_connection")
	defer span.End()

	remoteAddr := conn.RemoteAddr().String()
	s.logger.WithField("remote_addr", remoteAddr).Debug("New connection accepted")

	span.SetAttributes(
		attribute.String("connection.remote_addr", remoteAddr),
	)

	// Set connection timeouts
	if s.config.ReadTimeout > 0 {
		conn.SetReadDeadline(time.Now().Add(s.config.ReadTimeout))
	}
	if s.config.WriteTimeout > 0 {
		conn.SetWriteDeadline(time.Now().Add(s.config.WriteTimeout))
	}

	// Create transport
	transport, err := NewTCPTransport(conn, s.logger)
	if err != nil {
		s.logger.WithError(err).Error("Failed to create transport")
		conn.Close()
		return
	}

	// Create session
	session, err := s.sessionManager.CreateSession(ctx, transport, nil)
	if err != nil {
		s.logger.WithError(err).Error("Failed to create session")
		transport.Close()
		return
	}

	// Emit session created event
	if s.eventEmitter != nil {
		s.eventEmitter.EmitSessionCreated(session)
	}

	// Record session metrics
	if s.metricsCollector != nil {
		s.metricsCollector.RecordSession("created", session.GetID())
	}

	// Handle session messages
	s.handleSession(ctx, session)

	// Cleanup
	s.sessionManager.CloseSession(session.GetID())

	// Emit session closed event
	if s.eventEmitter != nil {
		s.eventEmitter.EmitSessionClosed(session.GetID())
	}

	// Record session metrics
	if s.metricsCollector != nil {
		s.metricsCollector.RecordSession("closed", session.GetID())
	}

	s.logger.WithFields(logrus.Fields{
		"session_id":  session.GetID(),
		"remote_addr": remoteAddr,
	}).Debug("Session closed")
}

func (s *MCPServer) handleSession(ctx context.Context, session protocol.Session) {
	transport := session.GetTransport()

	// Start receiving messages
	messageCh, err := transport.Receive(ctx)
	if err != nil {
		s.logger.WithError(err).Error("Failed to start receiving messages")
		return
	}

	for {
		select {
		case <-ctx.Done():
			return
		case message, ok := <-messageCh:
			if !ok {
				// Channel closed, session ended
				return
			}

			// Handle the message
			s.handleMessage(ctx, session, message)
		}
	}
}

func (s *MCPServer) handleMessage(ctx context.Context, session protocol.Session, message protocol.Message) {
	ctx, span := s.tracer.Start(ctx, "mcp_server.handle_message")
	defer span.End()

	span.SetAttributes(
		attribute.String("message.type", string(message.GetType())),
		attribute.String("message.method", message.GetMethod()),
		attribute.String("session.id", session.GetID()),
	)

	switch message.GetType() {
	case protocol.MessageTypeRequest:
		s.handleRequest(ctx, session, message.(protocol.Request))
	case protocol.MessageTypeNotification:
		s.handleNotification(ctx, session, message.(protocol.Notification))
	default:
		s.logger.WithField("message_type", message.GetType()).Warn("Unknown message type")
	}
}

func (s *MCPServer) handleRequest(ctx context.Context, session protocol.Session, request protocol.Request) {
	start := time.Now()

	// Emit request received event
	if s.eventEmitter != nil {
		s.eventEmitter.EmitRequestReceived(session, request)
	}

	// Route request to handler
	response, err := s.messageRouter.RouteRequest(ctx, session, request)
	if err != nil {
		// Create error response
		response, _ = protocol.NewErrorResponse(
			request.GetRequestID(),
			protocol.ErrorCodeInternalError,
			err.Error(),
			nil,
		)
	}

	// Send response
	if err := session.SendResponse(ctx, response); err != nil {
		s.logger.WithError(err).Error("Failed to send response")
	}

	// Emit response sent event
	if s.eventEmitter != nil {
		s.eventEmitter.EmitResponseSent(session, response)
	}

	// Record metrics
	if s.metricsCollector != nil {
		duration := time.Since(start)
		success := response.IsSuccess()
		s.metricsCollector.RecordRequest(request.GetMethod(), duration, success)

		if !success {
			s.metricsCollector.RecordError(response.GetError().Code, request.GetMethod())
		}
	}
}

func (s *MCPServer) handleNotification(ctx context.Context, session protocol.Session, notification protocol.Notification) {
	// Emit notification received event
	if s.eventEmitter != nil {
		s.eventEmitter.EmitNotificationReceived(session, notification)
	}

	// Route notification to handler
	if err := s.messageRouter.RouteNotification(ctx, session, notification); err != nil {
		s.logger.WithError(err).Error("Failed to handle notification")
	}

	// Record metrics
	if s.metricsCollector != nil {
		s.metricsCollector.RecordNotification(notification.GetMethod(), true)
	}
}

// Helper constructors for missing components

// TODO: Implement proper protocol.MetricsCollector and protocol.EventEmitter
// These implementations don't match the protocol interfaces

// NewMetricsCollector creates a new metrics collector
// func NewMetricsCollector(logger *logrus.Logger) (protocol.MetricsCollector, error) {
//     return &DefaultMetricsCollector{
//         logger: logger,
//     }, nil
// }

// NewEventEmitter creates a new event emitter
// func NewEventEmitter(logger *logrus.Logger) (protocol.EventEmitter, error) {
//     return &DefaultEventEmitter{
//         logger: logger,
//     }, nil
// }

// TODO: Implement proper DefaultMetricsCollector that matches protocol.MetricsCollector interface

// DefaultMetricsCollector implements protocol.MetricsCollector
// type DefaultMetricsCollector struct {
//     logger *logrus.Logger
// }

// GetMetrics returns metrics
// func (mc *DefaultMetricsCollector) GetMetrics() map[string]interface{} {
//     return make(map[string]interface{})
// }

// RecordRequest records a request metric
// func (mc *DefaultMetricsCollector) RecordRequest(method string, duration time.Duration, success bool) {
//     mc.logger.WithFields(logrus.Fields{
//         "method":   method,
//         "duration": duration,
//         "success":  success,
//     }).Debug("Request metric recorded")
// }

// RecordError records an error metric
// func (mc *DefaultMetricsCollector) RecordError(code int, method string) {
//     mc.logger.WithFields(logrus.Fields{
//         "code":   code,
//         "method": method,
//     }).Debug("Error metric recorded")
// }

// RecordNotification records a notification metric
// func (mc *DefaultMetricsCollector) RecordNotification(method string, success bool) {
//     mc.logger.WithFields(logrus.Fields{
//         "method":  method,
//         "success": success,
//     }).Debug("Notification metric recorded")
// }

// RecordSession records a session metric
// func (mc *DefaultMetricsCollector) RecordSession(event, sessionID string) {
//     mc.logger.WithFields(logrus.Fields{
//         "event":      event,
//         "session_id": sessionID,
//     }).Debug("Session metric recorded")
// }

// TODO: Implement proper DefaultEventEmitter that matches protocol.EventEmitter interface

// DefaultEventEmitter implements protocol.EventEmitter
// type DefaultEventEmitter struct {
//     logger *logrus.Logger
// }

// EmitRequestReceived emits a request received event
// func (ee *DefaultEventEmitter) EmitRequestReceived(session protocol.Session, request protocol.Request) {
//     ee.logger.WithFields(logrus.Fields{
//         "session_id": session.GetID(),
//         "method":     request.GetMethod(),
//     }).Debug("Request received event")
// }

// EmitResponseSent emits a response sent event
// func (ee *DefaultEventEmitter) EmitResponseSent(session protocol.Session, response protocol.Response) {
//     ee.logger.WithFields(logrus.Fields{
//         "session_id": session.GetID(),
//         "success":    response.IsSuccess(),
//     }).Debug("Response sent event")
// }

// EmitNotificationReceived emits a notification received event
// func (ee *DefaultEventEmitter) EmitNotificationReceived(session protocol.Session, notification protocol.Notification) {
//     ee.logger.WithFields(logrus.Fields{
//         "session_id": session.GetID(),
//         "method":     notification.GetMethod(),
//     }).Debug("Notification received event")
// }

// EmitSessionCreated emits a session created event
// func (ee *DefaultEventEmitter) EmitSessionCreated(session protocol.Session) {
//     ee.logger.WithField("session_id", session.GetID()).Debug("Session created event")
// }

// EmitSessionClosed emits a session closed event
// func (ee *DefaultEventEmitter) EmitSessionClosed(sessionID string) {
//     ee.logger.WithField("session_id", sessionID).Debug("Session closed event")
// }

// EmitError emits an error event
// func (ee *DefaultEventEmitter) EmitError(session protocol.Session, err error) {
//     ee.logger.WithFields(logrus.Fields{
//         "session_id": session.GetID(),
//         "error":      err.Error(),
//     }).Debug("Error event")
// }
