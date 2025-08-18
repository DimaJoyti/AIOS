package client

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"sync"

	"github.com/aios/aios/pkg/mcp/protocol"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

// ClientTransport implements the Transport interface for MCP clients
type ClientTransport struct {
	conn      net.Conn
	reader    *bufio.Reader
	writer    *bufio.Writer
	logger    *logrus.Logger
	tracer    trace.Tracer
	connected bool
	mu        sync.RWMutex
}

// NewClientTransport creates a new client transport
func NewClientTransport(conn net.Conn, logger *logrus.Logger) (protocol.Transport, error) {
	if conn == nil {
		return nil, fmt.Errorf("connection cannot be nil")
	}

	return &ClientTransport{
		conn:      conn,
		reader:    bufio.NewReader(conn),
		writer:    bufio.NewWriter(conn),
		logger:    logger,
		tracer:    otel.Tracer("mcp.client.transport"),
		connected: true,
	}, nil
}

// Send sends a message through the transport
func (t *ClientTransport) Send(ctx context.Context, message protocol.Message) error {
	ctx, span := t.tracer.Start(ctx, "client_transport.send")
	defer span.End()

	t.mu.Lock()
	defer t.mu.Unlock()

	if !t.connected {
		return fmt.Errorf("transport is not connected")
	}

	// Serialize message to JSON
	data, err := t.serializeMessage(message)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to serialize message: %w", err)
	}

	// Write message with newline delimiter
	if _, err := t.writer.Write(data); err != nil {
		span.RecordError(err)
		t.connected = false
		return fmt.Errorf("failed to write message: %w", err)
	}

	if _, err := t.writer.WriteString("\n"); err != nil {
		span.RecordError(err)
		t.connected = false
		return fmt.Errorf("failed to write delimiter: %w", err)
	}

	// Flush the writer
	if err := t.writer.Flush(); err != nil {
		span.RecordError(err)
		t.connected = false
		return fmt.Errorf("failed to flush writer: %w", err)
	}

	t.logger.WithFields(logrus.Fields{
		"message_type":   message.GetType(),
		"message_method": message.GetMethod(),
		"message_id":     message.GetID(),
		"data_size":      len(data),
	}).Debug("Message sent")

	return nil
}

// Receive receives messages from the transport
func (t *ClientTransport) Receive(ctx context.Context) (<-chan protocol.Message, error) {
	ctx, span := t.tracer.Start(ctx, "client_transport.receive")
	defer span.End()

	messageCh := make(chan protocol.Message, 10)

	go func() {
		defer close(messageCh)

		for {
			select {
			case <-ctx.Done():
				return
			default:
				// Read message
				message, err := t.readMessage()
				if err != nil {
					t.logger.WithError(err).Error("Failed to read message")
					t.mu.Lock()
					t.connected = false
					t.mu.Unlock()
					return
				}

				if message != nil {
					select {
					case messageCh <- message:
					case <-ctx.Done():
						return
					}
				}
			}
		}
	}()

	return messageCh, nil
}

// Close closes the transport
func (t *ClientTransport) Close() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if !t.connected {
		return nil
	}

	t.connected = false

	if t.conn != nil {
		return t.conn.Close()
	}

	return nil
}

// GetRemoteAddress returns the remote address
func (t *ClientTransport) GetRemoteAddress() string {
	if t.conn != nil {
		return t.conn.RemoteAddr().String()
	}
	return ""
}

// IsConnected returns whether the transport is connected
func (t *ClientTransport) IsConnected() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.connected
}

// Helper methods (similar to server transport)

func (t *ClientTransport) serializeMessage(message protocol.Message) ([]byte, error) {
	// Create a map representation of the message
	messageMap := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      message.GetID(),
		"method":  message.GetMethod(),
	}

	// Add type-specific fields
	switch msg := message.(type) {
	case protocol.Request:
		messageMap["params"] = json.RawMessage(msg.GetParams())
		if msg.GetRequestID() != "" {
			messageMap["id"] = msg.GetRequestID()
		}

	case protocol.Response:
		messageMap["id"] = msg.GetRequestID()
		delete(messageMap, "method") // Responses don't have methods

		if msg.IsSuccess() {
			messageMap["result"] = json.RawMessage(msg.GetResult())
		} else {
			messageMap["error"] = msg.GetError()
		}

	case protocol.Notification:
		delete(messageMap, "id") // Notifications don't have IDs
		messageMap["params"] = json.RawMessage(msg.GetParams())
	}

	return json.Marshal(messageMap)
}

func (t *ClientTransport) readMessage() (protocol.Message, error) {
	// Read line from connection
	line, err := t.reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("failed to read line: %w", err)
	}

	// Remove newline
	if len(line) > 0 && line[len(line)-1] == '\n' {
		line = line[:len(line)-1]
	}
	if len(line) > 0 && line[len(line)-1] == '\r' {
		line = line[:len(line)-1]
	}

	if len(line) == 0 {
		return nil, nil // Empty line, skip
	}

	// Parse JSON
	var rawMessage map[string]interface{}
	if err := json.Unmarshal([]byte(line), &rawMessage); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Determine message type and create appropriate message
	return t.parseMessage(rawMessage)
}

func (t *ClientTransport) parseMessage(rawMessage map[string]interface{}) (protocol.Message, error) {
	// Check for required fields
	jsonrpc, ok := rawMessage["jsonrpc"].(string)
	if !ok || jsonrpc != "2.0" {
		return nil, fmt.Errorf("invalid or missing jsonrpc field")
	}

	_, hasMethod := rawMessage["method"].(string)
	_, hasID := rawMessage["id"]
	_, hasResult := rawMessage["result"]
	_, hasError := rawMessage["error"]

	// Determine message type
	if hasMethod && hasID {
		// Request
		return t.parseRequest(rawMessage)
	} else if hasMethod && !hasID {
		// Notification
		return t.parseNotification(rawMessage)
	} else if (hasResult || hasError) && hasID {
		// Response
		return t.parseResponse(rawMessage)
	}

	return nil, fmt.Errorf("unable to determine message type")
}

func (t *ClientTransport) parseRequest(rawMessage map[string]interface{}) (protocol.Request, error) {
	method := rawMessage["method"].(string)
	id := fmt.Sprintf("%v", rawMessage["id"])

	var params json.RawMessage
	if p, exists := rawMessage["params"]; exists {
		var err error
		params, err = json.Marshal(p)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal params: %w", err)
		}
	}

	request := &protocol.MCPRequest{
		BaseMessage: protocol.BaseMessage{
			ID:     id,
			Type:   protocol.MessageTypeRequest,
			Method: method,
			Params: params,
		},
		RequestID: id,
	}

	return request, nil
}

func (t *ClientTransport) parseNotification(rawMessage map[string]interface{}) (protocol.Notification, error) {
	method := rawMessage["method"].(string)

	var params json.RawMessage
	if p, exists := rawMessage["params"]; exists {
		var err error
		params, err = json.Marshal(p)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal params: %w", err)
		}
	}

	notification := &protocol.MCPNotification{
		BaseMessage: protocol.BaseMessage{
			ID:     "", // Notifications don't have IDs
			Type:   protocol.MessageTypeNotification,
			Method: method,
			Params: params,
		},
		NotificationType: method,
	}

	return notification, nil
}

func (t *ClientTransport) parseResponse(rawMessage map[string]interface{}) (protocol.Response, error) {
	requestID := fmt.Sprintf("%v", rawMessage["id"])

	response := &protocol.MCPResponse{
		BaseMessage: protocol.BaseMessage{
			ID:     requestID,
			Type:   protocol.MessageTypeResponse,
			Method: "response",
		},
		RequestID: requestID,
	}

	// Check for result or error
	if result, hasResult := rawMessage["result"]; hasResult {
		var err error
		response.Result, err = json.Marshal(result)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal result: %w", err)
		}
	} else if errorData, hasError := rawMessage["error"]; hasError {
		errorMap, ok := errorData.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid error format")
		}

		mcpError := &protocol.MCPError{}

		if code, ok := errorMap["code"].(float64); ok {
			mcpError.Code = int(code)
		}

		if message, ok := errorMap["message"].(string); ok {
			mcpError.Message = message
		}

		if data, ok := errorMap["data"]; ok {
			var err error
			mcpError.Data, err = json.Marshal(data)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal error data: %w", err)
			}
		}

		response.Error = mcpError
	}

	return response, nil
}
