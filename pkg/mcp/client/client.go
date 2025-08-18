package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/aios/aios/pkg/mcp/protocol"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// MCPClient implements an MCP client
type MCPClient struct {
	config          *ClientConfig
	session         protocol.Session
	transport       protocol.Transport
	pendingRequests map[string]chan protocol.Response
	logger          *logrus.Logger
	tracer          trace.Tracer
	connected       bool
	mu              sync.RWMutex
}

// ClientConfig represents MCP client configuration
type ClientConfig struct {
	ServerAddress  string                 `json:"server_address"`
	ServerPort     int                    `json:"server_port"`
	ClientInfo     protocol.ClientInfo    `json:"client_info"`
	Capabilities   protocol.Capabilities  `json:"capabilities"`
	ConnectTimeout time.Duration          `json:"connect_timeout"`
	RequestTimeout time.Duration          `json:"request_timeout"`
	RetryAttempts  int                    `json:"retry_attempts"`
	RetryDelay     time.Duration          `json:"retry_delay"`
	EnableMetrics  bool                   `json:"enable_metrics"`
	Metadata       map[string]interface{} `json:"metadata"`
}

// NewMCPClient creates a new MCP client
func NewMCPClient(config *ClientConfig, logger *logrus.Logger) (*MCPClient, error) {
	if config == nil {
		return nil, fmt.Errorf("client config cannot be nil")
	}

	// Set defaults
	if config.ServerAddress == "" {
		config.ServerAddress = "localhost"
	}
	if config.ServerPort == 0 {
		config.ServerPort = 8080
	}
	if config.ConnectTimeout == 0 {
		config.ConnectTimeout = 30 * time.Second
	}
	if config.RequestTimeout == 0 {
		config.RequestTimeout = 30 * time.Second
	}
	if config.RetryAttempts == 0 {
		config.RetryAttempts = 3
	}
	if config.RetryDelay == 0 {
		config.RetryDelay = 1 * time.Second
	}

	client := &MCPClient{
		config:          config,
		pendingRequests: make(map[string]chan protocol.Response),
		logger:          logger,
		tracer:          otel.Tracer("mcp.client"),
		connected:       false,
	}

	return client, nil
}

// Connect connects to the MCP server
func (c *MCPClient) Connect(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.connected {
		return fmt.Errorf("client is already connected")
	}

	ctx, span := c.tracer.Start(ctx, "mcp_client.connect")
	defer span.End()

	c.logger.WithFields(logrus.Fields{
		"server_address": c.config.ServerAddress,
		"server_port":    c.config.ServerPort,
	}).Info("Connecting to MCP server")

	// Create connection with timeout
	connectCtx, cancel := context.WithTimeout(ctx, c.config.ConnectTimeout)
	defer cancel()

	address := fmt.Sprintf("%s:%d", c.config.ServerAddress, c.config.ServerPort)

	var conn net.Conn
	var err error

	// Retry connection
	for attempt := 0; attempt < c.config.RetryAttempts; attempt++ {
		conn, err = (&net.Dialer{}).DialContext(connectCtx, "tcp", address)
		if err == nil {
			break
		}

		if attempt < c.config.RetryAttempts-1 {
			c.logger.WithError(err).WithField("attempt", attempt+1).Warn("Connection attempt failed, retrying")
			time.Sleep(c.config.RetryDelay)
		}
	}

	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to connect to server: %w", err)
	}

	// Create transport
	transport, err := NewClientTransport(conn, c.logger)
	if err != nil {
		conn.Close()
		return fmt.Errorf("failed to create transport: %w", err)
	}

	c.transport = transport

	// Create session
	session := NewClientSession(uuid.New().String(), transport, &c.config.ClientInfo, c.logger)
	c.session = session

	// Start message handling
	go c.handleMessages(ctx)

	// Initialize session
	if err := c.initialize(ctx); err != nil {
		c.transport.Close()
		return fmt.Errorf("failed to initialize session: %w", err)
	}

	c.connected = true

	span.SetAttributes(
		attribute.String("server.address", c.config.ServerAddress),
		attribute.Int("server.port", c.config.ServerPort),
	)

	c.logger.WithField("session_id", session.GetID()).Info("Connected to MCP server")

	return nil
}

// Disconnect disconnects from the MCP server
func (c *MCPClient) Disconnect(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.connected {
		return nil
	}

	ctx, span := c.tracer.Start(ctx, "mcp_client.disconnect")
	defer span.End()

	c.logger.Info("Disconnecting from MCP server")

	// Close session
	if c.session != nil {
		if err := c.session.Close(); err != nil {
			c.logger.WithError(err).Error("Failed to close session")
		}
	}

	// Close transport
	if c.transport != nil {
		if err := c.transport.Close(); err != nil {
			c.logger.WithError(err).Error("Failed to close transport")
		}
	}

	// Cancel pending requests
	for requestID, responseCh := range c.pendingRequests {
		close(responseCh)
		delete(c.pendingRequests, requestID)
	}

	c.connected = false

	c.logger.Info("Disconnected from MCP server")

	return nil
}

// IsConnected returns whether the client is connected
func (c *MCPClient) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.connected
}

// Ping sends a ping request to the server
func (c *MCPClient) Ping(ctx context.Context, message string) (*protocol.PingResult, error) {
	params := protocol.PingParams{
		Message: message,
	}

	response, err := c.sendRequest(ctx, protocol.MethodPing, params)
	if err != nil {
		return nil, err
	}

	var result protocol.PingResult
	if err := parseResponse(response, &result); err != nil {
		return nil, fmt.Errorf("failed to parse ping response: %w", err)
	}

	return &result, nil
}

// ListResources lists available resources
func (c *MCPClient) ListResources(ctx context.Context, cursor string) (*protocol.ListResourcesResult, error) {
	params := protocol.ListResourcesParams{
		Cursor: cursor,
	}

	response, err := c.sendRequest(ctx, protocol.MethodListResources, params)
	if err != nil {
		return nil, err
	}

	var result protocol.ListResourcesResult
	if err := parseResponse(response, &result); err != nil {
		return nil, fmt.Errorf("failed to parse list resources response: %w", err)
	}

	return &result, nil
}

// ReadResource reads a resource
func (c *MCPClient) ReadResource(ctx context.Context, uri string) (*protocol.ReadResourceResult, error) {
	params := protocol.ReadResourceParams{
		URI: uri,
	}

	response, err := c.sendRequest(ctx, protocol.MethodReadResource, params)
	if err != nil {
		return nil, err
	}

	var result protocol.ReadResourceResult
	if err := parseResponse(response, &result); err != nil {
		return nil, fmt.Errorf("failed to parse read resource response: %w", err)
	}

	return &result, nil
}

// ListTools lists available tools
func (c *MCPClient) ListTools(ctx context.Context, cursor string) (*protocol.ListToolsResult, error) {
	params := protocol.ListToolsParams{
		Cursor: cursor,
	}

	response, err := c.sendRequest(ctx, protocol.MethodListTools, params)
	if err != nil {
		return nil, err
	}

	var result protocol.ListToolsResult
	if err := parseResponse(response, &result); err != nil {
		return nil, fmt.Errorf("failed to parse list tools response: %w", err)
	}

	return &result, nil
}

// CallTool calls a tool
func (c *MCPClient) CallTool(ctx context.Context, name string, arguments map[string]interface{}) (*protocol.CallToolResult, error) {
	params := protocol.CallToolParams{
		Name:      name,
		Arguments: arguments,
	}

	response, err := c.sendRequest(ctx, protocol.MethodCallTool, params)
	if err != nil {
		return nil, err
	}

	var result protocol.CallToolResult
	if err := parseResponse(response, &result); err != nil {
		return nil, fmt.Errorf("failed to parse call tool response: %w", err)
	}

	return &result, nil
}

// ListPrompts lists available prompts
func (c *MCPClient) ListPrompts(ctx context.Context, cursor string) (*protocol.ListPromptsResult, error) {
	params := protocol.ListPromptsParams{
		Cursor: cursor,
	}

	response, err := c.sendRequest(ctx, protocol.MethodListPrompts, params)
	if err != nil {
		return nil, err
	}

	var result protocol.ListPromptsResult
	if err := parseResponse(response, &result); err != nil {
		return nil, fmt.Errorf("failed to parse list prompts response: %w", err)
	}

	return &result, nil
}

// GetPrompt gets a prompt
func (c *MCPClient) GetPrompt(ctx context.Context, name string, arguments map[string]interface{}) (*protocol.GetPromptResult, error) {
	params := protocol.GetPromptParams{
		Name:      name,
		Arguments: arguments,
	}

	response, err := c.sendRequest(ctx, protocol.MethodGetPrompt, params)
	if err != nil {
		return nil, err
	}

	var result protocol.GetPromptResult
	if err := parseResponse(response, &result); err != nil {
		return nil, fmt.Errorf("failed to parse get prompt response: %w", err)
	}

	return &result, nil
}

// SendNotification sends a notification to the server
func (c *MCPClient) SendNotification(ctx context.Context, method string, params interface{}) error {
	notification, err := protocol.NewNotification(method, params)
	if err != nil {
		return fmt.Errorf("failed to create notification: %w", err)
	}

	return c.session.SendNotification(ctx, notification)
}

// Helper methods

func (c *MCPClient) initialize(ctx context.Context) error {
	params := protocol.InitializeParams{
		ProtocolVersion: protocol.MCPVersion,
		Capabilities:    c.config.Capabilities,
		ClientInfo:      c.config.ClientInfo,
	}

	response, err := c.sendRequest(ctx, protocol.MethodInitialize, params)
	if err != nil {
		return fmt.Errorf("initialize request failed: %w", err)
	}

	var result protocol.InitializeResult
	if err := parseResponse(response, &result); err != nil {
		return fmt.Errorf("failed to parse initialize response: %w", err)
	}

	// Update session with server info
	if clientSession, ok := c.session.(*ClientSession); ok {
		clientSession.SetServerInfo(&result.ServerInfo)
		clientSession.SetCapabilities(&result.Capabilities)
	}

	// Send initialized notification
	return c.SendNotification(ctx, protocol.MethodInitialized, nil)
}

func (c *MCPClient) sendRequest(ctx context.Context, method string, params interface{}) (protocol.Response, error) {
	if !c.connected {
		return nil, fmt.Errorf("client is not connected")
	}

	request, err := protocol.NewRequest(method, params)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Create response channel
	responseCh := make(chan protocol.Response, 1)

	c.mu.Lock()
	c.pendingRequests[request.GetRequestID()] = responseCh
	c.mu.Unlock()

	// Send request
	_, err = c.session.SendRequest(ctx, request)
	if err != nil {
		c.mu.Lock()
		delete(c.pendingRequests, request.GetRequestID())
		c.mu.Unlock()
		close(responseCh)
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	// Wait for response with timeout
	requestCtx, cancel := context.WithTimeout(ctx, c.config.RequestTimeout)
	defer cancel()

	select {
	case response := <-responseCh:
		c.mu.Lock()
		delete(c.pendingRequests, request.GetRequestID())
		c.mu.Unlock()

		if response == nil {
			return nil, fmt.Errorf("received nil response")
		}

		if !response.IsSuccess() {
			return nil, fmt.Errorf("request failed: %s", response.GetError().Message)
		}

		return response, nil

	case <-requestCtx.Done():
		c.mu.Lock()
		delete(c.pendingRequests, request.GetRequestID())
		c.mu.Unlock()
		close(responseCh)

		return nil, fmt.Errorf("request timeout")
	}
}

func (c *MCPClient) handleMessages(ctx context.Context) {
	messageCh, err := c.transport.Receive(ctx)
	if err != nil {
		c.logger.WithError(err).Error("Failed to start receiving messages")
		return
	}

	for {
		select {
		case <-ctx.Done():
			return
		case message, ok := <-messageCh:
			if !ok {
				// Channel closed, connection ended
				c.mu.Lock()
				c.connected = false
				c.mu.Unlock()
				return
			}

			c.handleMessage(ctx, message)
		}
	}
}

func (c *MCPClient) handleMessage(ctx context.Context, message protocol.Message) {
	switch message.GetType() {
	case protocol.MessageTypeResponse:
		c.handleResponse(message.(protocol.Response))
	case protocol.MessageTypeNotification:
		c.handleNotification(message.(protocol.Notification))
	default:
		c.logger.WithField("message_type", message.GetType()).Warn("Unknown message type received")
	}
}

func (c *MCPClient) handleResponse(response protocol.Response) {
	requestID := response.GetRequestID()

	c.mu.RLock()
	responseCh, exists := c.pendingRequests[requestID]
	c.mu.RUnlock()

	if !exists {
		c.logger.WithField("request_id", requestID).Warn("Received response for unknown request")
		return
	}

	select {
	case responseCh <- response:
	default:
		c.logger.WithField("request_id", requestID).Warn("Response channel is full or closed")
	}
}

func (c *MCPClient) handleNotification(notification protocol.Notification) {
	c.logger.WithFields(logrus.Fields{
		"method": notification.GetMethod(),
		"type":   notification.GetNotificationType(),
	}).Debug("Received notification")

	// Handle specific notifications
	switch notification.GetMethod() {
	case protocol.MethodNotificationProgress:
		// Handle progress notifications
	case protocol.MethodNotificationMessage:
		// Handle message notifications
	case protocol.MethodNotificationCancelled:
		// Handle cancellation notifications
	default:
		c.logger.WithField("method", notification.GetMethod()).Debug("Unhandled notification")
	}
}

func parseResponse(response protocol.Response, target interface{}) error {
	if !response.IsSuccess() {
		return fmt.Errorf("response error: %s", response.GetError().Message)
	}

	return json.Unmarshal(response.GetResult(), target)
}
