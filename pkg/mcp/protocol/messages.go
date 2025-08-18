package protocol

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// BaseMessage implements the common Message interface
type BaseMessage struct {
	ID        string          `json:"id"`
	Type      MessageType     `json:"type"`
	Method    string          `json:"method"`
	Params    json.RawMessage `json:"params,omitempty"`
	Timestamp time.Time       `json:"timestamp"`
}

// GetID returns the message ID
func (m *BaseMessage) GetID() string {
	return m.ID
}

// GetType returns the message type
func (m *BaseMessage) GetType() MessageType {
	return m.Type
}

// GetMethod returns the method name
func (m *BaseMessage) GetMethod() string {
	return m.Method
}

// GetParams returns the parameters
func (m *BaseMessage) GetParams() json.RawMessage {
	return m.Params
}

// GetTimestamp returns the timestamp
func (m *BaseMessage) GetTimestamp() time.Time {
	return m.Timestamp
}

// Validate validates the base message
func (m *BaseMessage) Validate() error {
	if m.ID == "" {
		return fmt.Errorf("message ID cannot be empty")
	}
	if m.Type == "" {
		return fmt.Errorf("message type cannot be empty")
	}
	if m.Method == "" {
		return fmt.Errorf("method cannot be empty")
	}
	return nil
}

// MCPRequest implements the Request interface
type MCPRequest struct {
	BaseMessage
	RequestID string `json:"requestId,omitempty"`
}

// NewRequest creates a new MCP request
func NewRequest(method string, params interface{}) (*MCPRequest, error) {
	paramsJSON, err := json.Marshal(params)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal params: %w", err)
	}

	return &MCPRequest{
		BaseMessage: BaseMessage{
			ID:        uuid.New().String(),
			Type:      MessageTypeRequest,
			Method:    method,
			Params:    paramsJSON,
			Timestamp: time.Now(),
		},
		RequestID: uuid.New().String(),
	}, nil
}

// GetRequestID returns the request ID
func (r *MCPRequest) GetRequestID() string {
	if r.RequestID != "" {
		return r.RequestID
	}
	return r.ID
}

// SetRequestID sets the request ID
func (r *MCPRequest) SetRequestID(id string) {
	r.RequestID = id
}

// WantsResponse returns whether this request expects a response
func (r *MCPRequest) WantsResponse() bool {
	return true // All requests expect responses in MCP
}

// MCPResponse implements the Response interface
type MCPResponse struct {
	BaseMessage
	RequestID string          `json:"requestId"`
	Result    json.RawMessage `json:"result,omitempty"`
	Error     *MCPError       `json:"error,omitempty"`
}

// NewResponse creates a new MCP response
func NewResponse(requestID string, result interface{}) (*MCPResponse, error) {
	resultJSON, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal result: %w", err)
	}

	return &MCPResponse{
		BaseMessage: BaseMessage{
			ID:        uuid.New().String(),
			Type:      MessageTypeResponse,
			Method:    "response",
			Timestamp: time.Now(),
		},
		RequestID: requestID,
		Result:    resultJSON,
	}, nil
}

// NewErrorResponse creates a new MCP error response
func NewErrorResponse(requestID string, code int, message string, data interface{}) (*MCPResponse, error) {
	var dataJSON json.RawMessage
	if data != nil {
		var err error
		dataJSON, err = json.Marshal(data)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal error data: %w", err)
		}
	}

	return &MCPResponse{
		BaseMessage: BaseMessage{
			ID:        uuid.New().String(),
			Type:      MessageTypeResponse,
			Method:    "response",
			Timestamp: time.Now(),
		},
		RequestID: requestID,
		Error: &MCPError{
			Code:    code,
			Message: message,
			Data:    dataJSON,
		},
	}, nil
}

// GetRequestID returns the request ID this response is for
func (r *MCPResponse) GetRequestID() string {
	return r.RequestID
}

// GetResult returns the result
func (r *MCPResponse) GetResult() json.RawMessage {
	return r.Result
}

// GetError returns the error
func (r *MCPResponse) GetError() *MCPError {
	return r.Error
}

// IsSuccess returns whether the response is successful
func (r *MCPResponse) IsSuccess() bool {
	return r.Error == nil
}

// MCPNotification implements the Notification interface
type MCPNotification struct {
	BaseMessage
	NotificationType string `json:"notificationType"`
}

// NewNotification creates a new MCP notification
func NewNotification(method string, params interface{}) (*MCPNotification, error) {
	paramsJSON, err := json.Marshal(params)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal params: %w", err)
	}

	return &MCPNotification{
		BaseMessage: BaseMessage{
			ID:        uuid.New().String(),
			Type:      MessageTypeNotification,
			Method:    method,
			Params:    paramsJSON,
			Timestamp: time.Now(),
		},
		NotificationType: method,
	}, nil
}

// GetNotificationType returns the notification type
func (n *MCPNotification) GetNotificationType() string {
	return n.NotificationType
}

// Standard MCP method names
const (
	MethodInitialize           = "initialize"
	MethodInitialized          = "initialized"
	MethodPing                 = "ping"
	MethodListResources        = "resources/list"
	MethodReadResource         = "resources/read"
	MethodSubscribeResource    = "resources/subscribe"
	MethodUnsubscribeResource  = "resources/unsubscribe"
	MethodListTools            = "tools/list"
	MethodCallTool             = "tools/call"
	MethodListPrompts          = "prompts/list"
	MethodGetPrompt            = "prompts/get"
	MethodSampling             = "sampling/createMessage"
	MethodLogging              = "logging/setLevel"
	MethodNotificationProgress = "notifications/progress"
	MethodNotificationMessage  = "notifications/message"
	MethodNotificationCancelled = "notifications/cancelled"
)

// InitializeParams represents parameters for the initialize method
type InitializeParams struct {
	ProtocolVersion string       `json:"protocolVersion"`
	Capabilities    Capabilities `json:"capabilities"`
	ClientInfo      ClientInfo   `json:"clientInfo"`
}

// InitializeResult represents the result of the initialize method
type InitializeResult struct {
	ProtocolVersion string       `json:"protocolVersion"`
	Capabilities    Capabilities `json:"capabilities"`
	ServerInfo      ServerInfo   `json:"serverInfo"`
}

// PingParams represents parameters for the ping method
type PingParams struct {
	Message string `json:"message,omitempty"`
}

// PingResult represents the result of the ping method
type PingResult struct {
	Message string `json:"message,omitempty"`
}

// ListResourcesParams represents parameters for listing resources
type ListResourcesParams struct {
	Cursor string `json:"cursor,omitempty"`
}

// ListResourcesResult represents the result of listing resources
type ListResourcesResult struct {
	Resources []Resource `json:"resources"`
	NextCursor string    `json:"nextCursor,omitempty"`
}

// Resource represents an MCP resource
type Resource struct {
	URI         string                 `json:"uri"`
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	MimeType    string                 `json:"mimeType,omitempty"`
	Annotations map[string]interface{} `json:"annotations,omitempty"`
}

// ReadResourceParams represents parameters for reading a resource
type ReadResourceParams struct {
	URI string `json:"uri"`
}

// ReadResourceResult represents the result of reading a resource
type ReadResourceResult struct {
	Contents []ResourceContent `json:"contents"`
}

// ResourceContent represents the content of a resource
type ResourceContent struct {
	URI      string `json:"uri"`
	MimeType string `json:"mimeType,omitempty"`
	Text     string `json:"text,omitempty"`
	Blob     string `json:"blob,omitempty"` // Base64 encoded binary data
}

// ListToolsParams represents parameters for listing tools
type ListToolsParams struct {
	Cursor string `json:"cursor,omitempty"`
}

// ListToolsResult represents the result of listing tools
type ListToolsResult struct {
	Tools      []Tool `json:"tools"`
	NextCursor string `json:"nextCursor,omitempty"`
}

// Tool represents an MCP tool
type Tool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	InputSchema map[string]interface{} `json:"inputSchema"`
}

// CallToolParams represents parameters for calling a tool
type CallToolParams struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments,omitempty"`
}

// CallToolResult represents the result of calling a tool
type CallToolResult struct {
	Content []ToolContent `json:"content"`
	IsError bool          `json:"isError,omitempty"`
}

// ToolContent represents the content returned by a tool
type ToolContent struct {
	Type     string                 `json:"type"`
	Text     string                 `json:"text,omitempty"`
	Data     interface{}            `json:"data,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// ListPromptsParams represents parameters for listing prompts
type ListPromptsParams struct {
	Cursor string `json:"cursor,omitempty"`
}

// ListPromptsResult represents the result of listing prompts
type ListPromptsResult struct {
	Prompts    []Prompt `json:"prompts"`
	NextCursor string   `json:"nextCursor,omitempty"`
}

// Prompt represents an MCP prompt
type Prompt struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	Arguments   []PromptArgument       `json:"arguments,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// PromptArgument represents an argument for a prompt
type PromptArgument struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Required    bool   `json:"required,omitempty"`
}

// GetPromptParams represents parameters for getting a prompt
type GetPromptParams struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments,omitempty"`
}

// GetPromptResult represents the result of getting a prompt
type GetPromptResult struct {
	Description string         `json:"description,omitempty"`
	Messages    []PromptMessage `json:"messages"`
}

// PromptMessage represents a message in a prompt
type PromptMessage struct {
	Role    string                 `json:"role"`
	Content PromptContent          `json:"content"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// PromptContent represents the content of a prompt message
type PromptContent struct {
	Type     string                 `json:"type"`
	Text     string                 `json:"text,omitempty"`
	ImageURL string                 `json:"image_url,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// ProgressNotificationParams represents parameters for progress notifications
type ProgressNotificationParams struct {
	ProgressToken interface{} `json:"progressToken"`
	Progress      float64     `json:"progress"`
	Total         float64     `json:"total,omitempty"`
}

// MessageNotificationParams represents parameters for message notifications
type MessageNotificationParams struct {
	Level   string `json:"level"`
	Logger  string `json:"logger,omitempty"`
	Data    interface{} `json:"data"`
}

// CancelledNotificationParams represents parameters for cancelled notifications
type CancelledNotificationParams struct {
	RequestID string `json:"requestId"`
	Reason    string `json:"reason,omitempty"`
}
