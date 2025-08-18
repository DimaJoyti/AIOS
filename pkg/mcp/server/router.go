package server

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/aios/aios/pkg/mcp/protocol"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

// DefaultMessageRouter implements the MessageRouter interface
type DefaultMessageRouter struct {
	handlers map[string]protocol.Handler
	logger   *logrus.Logger
	tracer   trace.Tracer
	mu       sync.RWMutex
}

// NewMessageRouter creates a new message router
func NewMessageRouter(logger *logrus.Logger) (protocol.MessageRouter, error) {
	return &DefaultMessageRouter{
		handlers: make(map[string]protocol.Handler),
		logger:   logger,
		tracer:   otel.Tracer("mcp.router"),
	}, nil
}

// RegisterHandler registers a handler for specific methods
func (r *DefaultMessageRouter) RegisterHandler(methods []string, handler protocol.Handler) error {
	if handler == nil {
		return fmt.Errorf("handler cannot be nil")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	for _, method := range methods {
		if method == "" {
			return fmt.Errorf("method cannot be empty")
		}

		if _, exists := r.handlers[method]; exists {
			return fmt.Errorf("handler for method %s already exists", method)
		}

		r.handlers[method] = handler
	}

	r.logger.WithFields(logrus.Fields{
		"methods": methods,
		"handler": fmt.Sprintf("%T", handler),
	}).Debug("Handler registered")

	return nil
}

// UnregisterHandler unregisters a handler
func (r *DefaultMessageRouter) UnregisterHandler(methods []string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, method := range methods {
		delete(r.handlers, method)
	}

	r.logger.WithField("methods", methods).Debug("Handler unregistered")

	return nil
}

// RouteRequest routes a request to the appropriate handler
func (r *DefaultMessageRouter) RouteRequest(ctx context.Context, session protocol.Session, request protocol.Request) (protocol.Response, error) {
	ctx, span := r.tracer.Start(ctx, "message_router.route_request")
	defer span.End()

	method := request.GetMethod()

	r.mu.RLock()
	handler, exists := r.handlers[method]
	r.mu.RUnlock()

	if !exists {
		return protocol.NewErrorResponse(
			request.GetRequestID(),
			protocol.ErrorCodeMethodNotFound,
			fmt.Sprintf("method not found: %s", method),
			nil,
		)
	}

	// Handle the request
	response, err := handler.HandleRequest(ctx, session, request)
	if err != nil {
		return protocol.NewErrorResponse(
			request.GetRequestID(),
			protocol.ErrorCodeInternalError,
			err.Error(),
			nil,
		)
	}

	return response, nil
}

// RouteNotification routes a notification to the appropriate handler
func (r *DefaultMessageRouter) RouteNotification(ctx context.Context, session protocol.Session, notification protocol.Notification) error {
	ctx, span := r.tracer.Start(ctx, "message_router.route_notification")
	defer span.End()

	method := notification.GetMethod()

	r.mu.RLock()
	handler, exists := r.handlers[method]
	r.mu.RUnlock()

	if !exists {
		r.logger.WithField("method", method).Warn("No handler found for notification")
		return nil // Notifications don't return errors
	}

	return handler.HandleNotification(ctx, session, notification)
}

// GetRegisteredMethods returns all registered methods
func (r *DefaultMessageRouter) GetRegisteredMethods() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	methods := make([]string, 0, len(r.handlers))
	for method := range r.handlers {
		methods = append(methods, method)
	}

	return methods
}

// InitializeHandler handles the initialize method
type InitializeHandler struct {
	config *ServerConfig
	logger *logrus.Logger
	tracer trace.Tracer
}

// NewInitializeHandler creates a new initialize handler
func NewInitializeHandler(config *ServerConfig, logger *logrus.Logger) *InitializeHandler {
	return &InitializeHandler{
		config: config,
		logger: logger,
		tracer: otel.Tracer("mcp.handlers.initialize"),
	}
}

// HandleRequest handles an initialize request
func (h *InitializeHandler) HandleRequest(ctx context.Context, session protocol.Session, request protocol.Request) (protocol.Response, error) {
	ctx, span := h.tracer.Start(ctx, "initialize_handler.handle_request")
	defer span.End()

	// Parse initialize parameters
	var params protocol.InitializeParams
	if err := parseParams(request.GetParams(), &params); err != nil {
		return protocol.NewErrorResponse(
			request.GetRequestID(),
			protocol.ErrorCodeInvalidParams,
			fmt.Sprintf("invalid initialize params: %v", err),
			nil,
		)
	}

	// Validate protocol version
	if params.ProtocolVersion != protocol.MCPVersion {
		return protocol.NewErrorResponse(
			request.GetRequestID(),
			protocol.ErrorCodeInvalidRequest,
			fmt.Sprintf("unsupported protocol version: %s", params.ProtocolVersion),
			nil,
		)
	}

	// Create server info
	serverInfo := &protocol.ServerInfo{
		Name:            "AIOS MCP Server",
		Version:         "1.0.0",
		ProtocolVersion: protocol.MCPVersion,
		Capabilities:    h.config.Protocol.Capabilities,
	}

	// Update session with server info and capabilities
	if mcpSession, ok := session.(*MCPSession); ok {
		mcpSession.SetServerInfo(serverInfo)
		mcpSession.SetCapabilities(&h.config.Protocol.Capabilities)
	}

	// Create initialize result
	result := protocol.InitializeResult{
		ProtocolVersion: protocol.MCPVersion,
		Capabilities:    h.config.Protocol.Capabilities,
		ServerInfo:      *serverInfo,
	}

	h.logger.WithFields(logrus.Fields{
		"session_id":       session.GetID(),
		"client_name":      params.ClientInfo.Name,
		"client_version":   params.ClientInfo.Version,
		"protocol_version": params.ProtocolVersion,
	}).Info("Session initialized")

	return protocol.NewResponse(request.GetRequestID(), result)
}

// HandleNotification handles initialize notifications
func (h *InitializeHandler) HandleNotification(ctx context.Context, session protocol.Session, notification protocol.Notification) error {
	// Initialize method doesn't handle notifications
	return nil
}

// GetSupportedMethods returns the methods this handler supports
func (h *InitializeHandler) GetSupportedMethods() []string {
	return []string{protocol.MethodInitialize}
}

// PingHandler handles the ping method
type PingHandler struct {
	logger *logrus.Logger
	tracer trace.Tracer
}

// NewPingHandler creates a new ping handler
func NewPingHandler(logger *logrus.Logger) *PingHandler {
	return &PingHandler{
		logger: logger,
		tracer: otel.Tracer("mcp.handlers.ping"),
	}
}

// HandleRequest handles a ping request
func (h *PingHandler) HandleRequest(ctx context.Context, session protocol.Session, request protocol.Request) (protocol.Response, error) {
	ctx, span := h.tracer.Start(ctx, "ping_handler.handle_request")
	defer span.End()

	// Parse ping parameters
	var params protocol.PingParams
	if err := parseParams(request.GetParams(), &params); err != nil {
		// Ping can work without parameters
		params = protocol.PingParams{}
	}

	// Create ping result
	result := protocol.PingResult{
		Message: params.Message,
	}

	if result.Message == "" {
		result.Message = "pong"
	}

	h.logger.WithFields(logrus.Fields{
		"session_id": session.GetID(),
		"message":    params.Message,
	}).Debug("Ping received")

	return protocol.NewResponse(request.GetRequestID(), result)
}

// HandleNotification handles ping notifications
func (h *PingHandler) HandleNotification(ctx context.Context, session protocol.Session, notification protocol.Notification) error {
	// Ping method doesn't handle notifications
	return nil
}

// GetSupportedMethods returns the methods this handler supports
func (h *PingHandler) GetSupportedMethods() []string {
	return []string{protocol.MethodPing}
}

// Helper functions

func parseParams(paramsJSON []byte, target interface{}) error {
	if len(paramsJSON) == 0 {
		return nil
	}

	if err := json.Unmarshal(paramsJSON, target); err != nil {
		return fmt.Errorf("failed to unmarshal params: %w", err)
	}

	return nil
}

// BaseHandler provides common functionality for handlers
type BaseHandler struct {
	logger *logrus.Logger
	tracer trace.Tracer
}

// NewBaseHandler creates a new base handler
func NewBaseHandler(name string, logger *logrus.Logger) *BaseHandler {
	return &BaseHandler{
		logger: logger,
		tracer: otel.Tracer(fmt.Sprintf("mcp.handlers.%s", name)),
	}
}

// LogRequest logs a request
func (h *BaseHandler) LogRequest(session protocol.Session, request protocol.Request) {
	h.logger.WithFields(logrus.Fields{
		"session_id": session.GetID(),
		"method":     request.GetMethod(),
		"request_id": request.GetRequestID(),
	}).Debug("Handling request")
}

// LogNotification logs a notification
func (h *BaseHandler) LogNotification(session protocol.Session, notification protocol.Notification) {
	h.logger.WithFields(logrus.Fields{
		"session_id": session.GetID(),
		"method":     notification.GetMethod(),
	}).Debug("Handling notification")
}

// CreateErrorResponse creates an error response
func (h *BaseHandler) CreateErrorResponse(requestID string, code int, message string, data interface{}) (protocol.Response, error) {
	return protocol.NewErrorResponse(requestID, code, message, data)
}

// CreateSuccessResponse creates a success response
func (h *BaseHandler) CreateSuccessResponse(requestID string, result interface{}) (protocol.Response, error) {
	return protocol.NewResponse(requestID, result)
}
