package enhanced

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

// StreamingHandler manages WebSocket connections for real-time MCP communication
type StreamingHandler struct {
	connections    map[string]*WebSocketConnection
	sessionManager *SessionManager
	toolRegistry   *EnhancedToolRegistry
	upgrader       websocket.Upgrader
	mutex          sync.RWMutex
	logger         *logrus.Logger
	tracer         trace.Tracer
}

// WebSocketConnection represents a WebSocket connection with session info
type WebSocketConnection struct {
	ID        string
	SessionID string
	Conn      *websocket.Conn
	Send      chan []byte
	Hub       *StreamingHandler
	IsActive  bool
	CreatedAt time.Time
	LastPing  time.Time
}

// MCPMessage represents an MCP protocol message
type MCPMessage struct {
	Type      string                 `json:"type"`
	ID        string                 `json:"id,omitempty"`
	Method    string                 `json:"method,omitempty"`
	Params    map[string]interface{} `json:"params,omitempty"`
	Result    interface{}            `json:"result,omitempty"`
	Error     *MCPError              `json:"error,omitempty"`
	SessionID string                 `json:"session_id,omitempty"`
}

// MCPError represents an MCP error
type MCPError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// StreamingToolRequest represents a streaming tool request
type StreamingToolRequest struct {
	ToolName   string                 `json:"tool_name"`
	Parameters map[string]interface{} `json:"parameters"`
	SessionID  string                 `json:"session_id"`
	RequestID  string                 `json:"request_id"`
	Streaming  bool                   `json:"streaming"`
}

// StreamingToolResponse represents a streaming tool response
type StreamingToolResponse struct {
	RequestID   string                 `json:"request_id"`
	Type        string                 `json:"type"` // start, data, progress, complete, error
	Data        interface{}            `json:"data,omitempty"`
	Error       string                 `json:"error,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	Timestamp   time.Time              `json:"timestamp"`
	IsComplete  bool                   `json:"is_complete"`
}

// NewStreamingHandler creates a new streaming handler
func NewStreamingHandler(sessionManager *SessionManager, toolRegistry *EnhancedToolRegistry, logger *logrus.Logger) *StreamingHandler {
	return &StreamingHandler{
		connections:    make(map[string]*WebSocketConnection),
		sessionManager: sessionManager,
		toolRegistry:   toolRegistry,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				// TODO: Implement proper origin checking
				return true
			},
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
		logger: logger,
		tracer: otel.Tracer("mcp.enhanced.streaming"),
	}
}

// HandleWebSocket handles WebSocket connections
func (sh *StreamingHandler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	ctx, span := sh.tracer.Start(r.Context(), "streaming_handler.handle_websocket")
	defer span.End()

	conn, err := sh.upgrader.Upgrade(w, r, nil)
	if err != nil {
		sh.logger.WithError(err).Error("Failed to upgrade WebSocket connection")
		return
	}

	// Create connection
	connectionID := fmt.Sprintf("conn_%d", time.Now().UnixNano())
	wsConn := &WebSocketConnection{
		ID:        connectionID,
		Conn:      conn,
		Send:      make(chan []byte, 256),
		Hub:       sh,
		IsActive:  true,
		CreatedAt: time.Now(),
		LastPing:  time.Now(),
	}

	// Register connection
	sh.mutex.Lock()
	sh.connections[connectionID] = wsConn
	sh.mutex.Unlock()

	sh.logger.WithField("connection_id", connectionID).Info("New WebSocket connection established")

	// Start goroutines for reading and writing
	go wsConn.writePump()
	go wsConn.readPump(ctx)
}

// readPump handles reading from the WebSocket connection
func (c *WebSocketConnection) readPump(ctx context.Context) {
	defer func() {
		c.Hub.unregisterConnection(c.ID)
		c.Conn.Close()
	}()

	// Set read deadline and pong handler
	c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.Conn.SetPongHandler(func(string) error {
		c.LastPing = time.Now()
		c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				c.Hub.logger.WithError(err).Error("WebSocket error")
			}
			break
		}

		// Process message
		c.Hub.processMessage(ctx, c, message)
	}
}

// writePump handles writing to the WebSocket connection
func (c *WebSocketConnection) writePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
				c.Hub.logger.WithError(err).Error("Failed to write WebSocket message")
				return
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// processMessage processes incoming WebSocket messages
func (sh *StreamingHandler) processMessage(ctx context.Context, conn *WebSocketConnection, message []byte) {
	ctx, span := sh.tracer.Start(ctx, "streaming_handler.process_message")
	defer span.End()

	var mcpMsg MCPMessage
	if err := json.Unmarshal(message, &mcpMsg); err != nil {
		sh.sendError(conn, "", fmt.Sprintf("Invalid JSON: %v", err))
		return
	}

	sh.logger.WithFields(logrus.Fields{
		"connection_id": conn.ID,
		"message_type":  mcpMsg.Type,
		"method":        mcpMsg.Method,
	}).Debug("Processing WebSocket message")

	switch mcpMsg.Type {
	case "request":
		sh.handleRequest(ctx, conn, &mcpMsg)
	case "notification":
		sh.handleNotification(ctx, conn, &mcpMsg)
	case "session_init":
		sh.handleSessionInit(ctx, conn, &mcpMsg)
	default:
		sh.sendError(conn, mcpMsg.ID, fmt.Sprintf("Unknown message type: %s", mcpMsg.Type))
	}
}

// handleRequest handles MCP requests
func (sh *StreamingHandler) handleRequest(ctx context.Context, conn *WebSocketConnection, msg *MCPMessage) {
	switch msg.Method {
	case "tools/call":
		sh.handleToolCall(ctx, conn, msg)
	case "tools/list":
		sh.handleToolList(ctx, conn, msg)
	case "session/info":
		sh.handleSessionInfo(ctx, conn, msg)
	default:
		sh.sendError(conn, msg.ID, fmt.Sprintf("Unknown method: %s", msg.Method))
	}
}

// handleToolCall handles tool execution requests
func (sh *StreamingHandler) handleToolCall(ctx context.Context, conn *WebSocketConnection, msg *MCPMessage) {
	if conn.SessionID == "" {
		sh.sendError(conn, msg.ID, "No active session")
		return
	}

	toolName, _ := msg.Params["name"].(string)
	arguments, _ := msg.Params["arguments"].(map[string]interface{})
	streaming, _ := msg.Params["streaming"].(bool)

	if toolName == "" {
		sh.sendError(conn, msg.ID, "Tool name is required")
		return
	}

	// Send start response for streaming
	if streaming {
		startResponse := StreamingToolResponse{
			RequestID:  msg.ID,
			Type:       "start",
			Data:       fmt.Sprintf("Starting tool execution: %s", toolName),
			Timestamp:  time.Now(),
			IsComplete: false,
		}
		sh.sendStreamingResponse(conn, &startResponse)
	}

	// Execute tool
	result, err := sh.toolRegistry.ExecuteTool(ctx, conn.SessionID, toolName, arguments)
	if err != nil {
		if streaming {
			errorResponse := StreamingToolResponse{
				RequestID:  msg.ID,
				Type:       "error",
				Error:      err.Error(),
				Timestamp:  time.Now(),
				IsComplete: true,
			}
			sh.sendStreamingResponse(conn, &errorResponse)
		} else {
			sh.sendError(conn, msg.ID, err.Error())
		}
		return
	}

	// Handle streaming results
	if result.IsStreaming && result.StreamData != nil {
		go sh.handleStreamingResult(conn, msg.ID, result.StreamData)
	} else {
		// Send regular response
		response := MCPMessage{
			Type:   "response",
			ID:     msg.ID,
			Result: result,
		}
		sh.sendMessage(conn, &response)
	}
}

// handleStreamingResult handles streaming tool results
func (sh *StreamingHandler) handleStreamingResult(conn *WebSocketConnection, requestID string, stream chan interface{}) {
	for data := range stream {
		streamResponse, ok := data.(StreamingResponse)
		if !ok {
			// Convert to streaming response
			streamResponse = StreamingResponse{
				Type:      "data",
				Data:      data,
				Timestamp: time.Now(),
			}
		}

		toolResponse := StreamingToolResponse{
			RequestID:  requestID,
			Type:       streamResponse.Type,
			Data:       streamResponse.Data,
			Metadata:   streamResponse.Metadata,
			Timestamp:  streamResponse.Timestamp,
			IsComplete: streamResponse.Type == "complete",
		}

		sh.sendStreamingResponse(conn, &toolResponse)

		if streamResponse.Type == "complete" {
			break
		}
	}
}

// handleToolList handles tool listing requests
func (sh *StreamingHandler) handleToolList(ctx context.Context, conn *WebSocketConnection, msg *MCPMessage) {
	tools := sh.toolRegistry.GetToolList()
	
	// Convert to MCP tool format
	mcpTools := make([]map[string]interface{}, 0, len(tools))
	for _, tool := range tools {
		mcpTool := map[string]interface{}{
			"name":         tool.Name,
			"description":  tool.Description,
			"inputSchema": tool.Parameters,
		}
		mcpTools = append(mcpTools, mcpTool)
	}

	response := MCPMessage{
		Type: "response",
		ID:   msg.ID,
		Result: map[string]interface{}{
			"tools": mcpTools,
		},
	}
	sh.sendMessage(conn, &response)
}

// handleSessionInfo handles session information requests
func (sh *StreamingHandler) handleSessionInfo(ctx context.Context, conn *WebSocketConnection, msg *MCPMessage) {
	if conn.SessionID == "" {
		sh.sendError(conn, msg.ID, "No active session")
		return
	}

	session, err := sh.sessionManager.GetSession(conn.SessionID)
	if err != nil {
		sh.sendError(conn, msg.ID, err.Error())
		return
	}

	response := MCPMessage{
		Type:   "response",
		ID:     msg.ID,
		Result: session,
	}
	sh.sendMessage(conn, &response)
}

// handleSessionInit handles session initialization
func (sh *StreamingHandler) handleSessionInit(ctx context.Context, conn *WebSocketConnection, msg *MCPMessage) {
	clientID, _ := msg.Params["client_id"].(string)
	if clientID == "" {
		clientID = conn.ID
	}

	// Create new session
	session, err := sh.sessionManager.CreateSession(ctx, clientID)
	if err != nil {
		sh.sendError(conn, msg.ID, fmt.Sprintf("Failed to create session: %v", err))
		return
	}

	conn.SessionID = session.ID

	response := MCPMessage{
		Type: "response",
		ID:   msg.ID,
		Result: map[string]interface{}{
			"session_id": session.ID,
			"capabilities": session.Capabilities,
		},
	}
	sh.sendMessage(conn, &response)
}

// handleNotification handles MCP notifications
func (sh *StreamingHandler) handleNotification(ctx context.Context, conn *WebSocketConnection, msg *MCPMessage) {
	// Handle notifications like progress updates, status changes, etc.
	sh.logger.WithFields(logrus.Fields{
		"connection_id": conn.ID,
		"method":        msg.Method,
	}).Debug("Received notification")
}

// sendMessage sends a message to a WebSocket connection
func (sh *StreamingHandler) sendMessage(conn *WebSocketConnection, msg *MCPMessage) {
	data, err := json.Marshal(msg)
	if err != nil {
		sh.logger.WithError(err).Error("Failed to marshal message")
		return
	}

	select {
	case conn.Send <- data:
	default:
		sh.logger.Warn("WebSocket send channel full, closing connection")
		sh.unregisterConnection(conn.ID)
	}
}

// sendStreamingResponse sends a streaming response
func (sh *StreamingHandler) sendStreamingResponse(conn *WebSocketConnection, response *StreamingToolResponse) {
	msg := MCPMessage{
		Type:   "streaming_response",
		Result: response,
	}
	sh.sendMessage(conn, &msg)
}

// sendError sends an error message
func (sh *StreamingHandler) sendError(conn *WebSocketConnection, requestID, errorMsg string) {
	msg := MCPMessage{
		Type: "response",
		ID:   requestID,
		Error: &MCPError{
			Code:    -1,
			Message: errorMsg,
		},
	}
	sh.sendMessage(conn, &msg)
}

// unregisterConnection removes a connection
func (sh *StreamingHandler) unregisterConnection(connectionID string) {
	sh.mutex.Lock()
	defer sh.mutex.Unlock()

	if conn, exists := sh.connections[connectionID]; exists {
		conn.IsActive = false
		close(conn.Send)
		delete(sh.connections, connectionID)
		sh.logger.WithField("connection_id", connectionID).Info("WebSocket connection closed")
	}
}

// BroadcastToSession broadcasts a message to all connections in a session
func (sh *StreamingHandler) BroadcastToSession(sessionID string, message interface{}) {
	sh.mutex.RLock()
	defer sh.mutex.RUnlock()

	msg := MCPMessage{
		Type:   "notification",
		Method: "session/broadcast",
		Params: map[string]interface{}{
			"data": message,
		},
	}

	for _, conn := range sh.connections {
		if conn.SessionID == sessionID && conn.IsActive {
			sh.sendMessage(conn, &msg)
		}
	}
}

// GetConnectionStats returns connection statistics
func (sh *StreamingHandler) GetConnectionStats() map[string]interface{} {
	sh.mutex.RLock()
	defer sh.mutex.RUnlock()

	activeConnections := 0
	sessionCounts := make(map[string]int)

	for _, conn := range sh.connections {
		if conn.IsActive {
			activeConnections++
			if conn.SessionID != "" {
				sessionCounts[conn.SessionID]++
			}
		}
	}

	return map[string]interface{}{
		"total_connections":  len(sh.connections),
		"active_connections": activeConnections,
		"sessions":           len(sessionCounts),
		"session_counts":     sessionCounts,
	}
}

// CleanupInactiveConnections removes inactive connections
func (sh *StreamingHandler) CleanupInactiveConnections(maxAge time.Duration) {
	sh.mutex.Lock()
	defer sh.mutex.Unlock()

	now := time.Now()
	for connectionID, conn := range sh.connections {
		if !conn.IsActive || now.Sub(conn.LastPing) > maxAge {
			conn.IsActive = false
			if conn.Send != nil {
				close(conn.Send)
			}
			delete(sh.connections, connectionID)
			sh.logger.WithField("connection_id", connectionID).Info("Cleaned up inactive connection")
		}
	}
}
