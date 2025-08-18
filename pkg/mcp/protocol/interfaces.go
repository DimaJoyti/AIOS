package protocol

import (
	"context"
	"encoding/json"
	"time"
)

// MCPVersion represents the MCP protocol version
const MCPVersion = "2024-11-05"

// MessageType represents the type of MCP message
type MessageType string

const (
	MessageTypeRequest      MessageType = "request"
	MessageTypeResponse     MessageType = "response"
	MessageTypeNotification MessageType = "notification"
)

// Message represents a base MCP message
type Message interface {
	GetID() string
	GetType() MessageType
	GetMethod() string
	GetParams() json.RawMessage
	GetTimestamp() time.Time
	Validate() error
}

// Request represents an MCP request message
type Request interface {
	Message
	GetRequestID() string
	SetRequestID(id string)
	WantsResponse() bool
}

// Response represents an MCP response message
type Response interface {
	Message
	GetRequestID() string
	GetResult() json.RawMessage
	GetError() *MCPError
	IsSuccess() bool
}

// Notification represents an MCP notification message
type Notification interface {
	Message
	GetNotificationType() string
}

// MCPError represents an MCP protocol error
type MCPError struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data,omitempty"`
}

// Error implements the error interface
func (e *MCPError) Error() string {
	return e.Message
}

// Standard MCP error codes
const (
	ErrorCodeParseError     = -32700
	ErrorCodeInvalidRequest = -32600
	ErrorCodeMethodNotFound = -32601
	ErrorCodeInvalidParams  = -32602
	ErrorCodeInternalError  = -32603
	ErrorCodeServerError    = -32000
	ErrorCodeUnauthorized   = -32001
	ErrorCodeForbidden      = -32002
	ErrorCodeNotFound       = -32003
	ErrorCodeTimeout        = -32004
	ErrorCodeRateLimited    = -32005
)

// Transport represents the transport layer for MCP communication
type Transport interface {
	// Send sends a message through the transport
	Send(ctx context.Context, message Message) error

	// Receive receives messages from the transport
	Receive(ctx context.Context) (<-chan Message, error)

	// Close closes the transport
	Close() error

	// GetRemoteAddress returns the remote address
	GetRemoteAddress() string

	// IsConnected returns whether the transport is connected
	IsConnected() bool
}

// Session represents an MCP session
type Session interface {
	// GetID returns the session ID
	GetID() string

	// GetClientInfo returns client information
	GetClientInfo() *ClientInfo

	// GetServerInfo returns server information
	GetServerInfo() *ServerInfo

	// GetCapabilities returns session capabilities
	GetCapabilities() *Capabilities

	// GetTransport returns the underlying transport
	GetTransport() Transport

	// SendRequest sends a request and returns a response channel
	SendRequest(ctx context.Context, request Request) (<-chan Response, error)

	// SendResponse sends a response
	SendResponse(ctx context.Context, response Response) error

	// SendNotification sends a notification
	SendNotification(ctx context.Context, notification Notification) error

	// GetLastActivity returns the last activity time
	GetLastActivity() time.Time

	// IsActive returns whether the session is active
	IsActive() bool

	// Close closes the session
	Close() error
}

// ClientInfo represents information about an MCP client
type ClientInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// ServerInfo represents information about an MCP server
type ServerInfo struct {
	Name         string       `json:"name"`
	Version      string       `json:"version"`
	ProtocolVersion string    `json:"protocolVersion"`
	Capabilities Capabilities `json:"capabilities"`
}

// Capabilities represents MCP capabilities
type Capabilities struct {
	Logging      *LoggingCapability      `json:"logging,omitempty"`
	Prompts      *PromptsCapability      `json:"prompts,omitempty"`
	Resources    *ResourcesCapability    `json:"resources,omitempty"`
	Tools        *ToolsCapability        `json:"tools,omitempty"`
	Sampling     *SamplingCapability     `json:"sampling,omitempty"`
	Experimental map[string]interface{}  `json:"experimental,omitempty"`
}

// LoggingCapability represents logging capabilities
type LoggingCapability struct {
	Enabled bool `json:"enabled"`
}

// PromptsCapability represents prompts capabilities
type PromptsCapability struct {
	ListChanged bool `json:"listChanged,omitempty"`
}

// ResourcesCapability represents resources capabilities
type ResourcesCapability struct {
	Subscribe   bool `json:"subscribe,omitempty"`
	ListChanged bool `json:"listChanged,omitempty"`
}

// ToolsCapability represents tools capabilities
type ToolsCapability struct {
	ListChanged bool `json:"listChanged,omitempty"`
}

// SamplingCapability represents sampling capabilities
type SamplingCapability struct {
	Enabled bool `json:"enabled"`
}

// Handler represents a handler for MCP messages
type Handler interface {
	// HandleRequest handles an incoming request
	HandleRequest(ctx context.Context, session Session, request Request) (Response, error)

	// HandleNotification handles an incoming notification
	HandleNotification(ctx context.Context, session Session, notification Notification) error

	// GetSupportedMethods returns the methods this handler supports
	GetSupportedMethods() []string
}

// Middleware represents middleware for MCP message processing
type Middleware interface {
	// ProcessRequest processes a request before handling
	ProcessRequest(ctx context.Context, session Session, request Request, next func(context.Context, Session, Request) (Response, error)) (Response, error)

	// ProcessNotification processes a notification before handling
	ProcessNotification(ctx context.Context, session Session, notification Notification, next func(context.Context, Session, Notification) error) error
}

// SessionManager manages MCP sessions
type SessionManager interface {
	// CreateSession creates a new session
	CreateSession(ctx context.Context, transport Transport, clientInfo *ClientInfo) (Session, error)

	// GetSession retrieves a session by ID
	GetSession(sessionID string) (Session, error)

	// ListSessions returns all active sessions
	ListSessions() []Session

	// CloseSession closes a session
	CloseSession(sessionID string) error

	// CloseAllSessions closes all sessions
	CloseAllSessions() error

	// GetSessionCount returns the number of active sessions
	GetSessionCount() int
}

// MessageRouter routes messages to appropriate handlers
type MessageRouter interface {
	// RegisterHandler registers a handler for specific methods
	RegisterHandler(methods []string, handler Handler) error

	// UnregisterHandler unregisters a handler
	UnregisterHandler(methods []string) error

	// RouteRequest routes a request to the appropriate handler
	RouteRequest(ctx context.Context, session Session, request Request) (Response, error)

	// RouteNotification routes a notification to the appropriate handler
	RouteNotification(ctx context.Context, session Session, notification Notification) error

	// GetRegisteredMethods returns all registered methods
	GetRegisteredMethods() []string
}

// ProtocolValidator validates MCP protocol compliance
type ProtocolValidator interface {
	// ValidateMessage validates a message
	ValidateMessage(message Message) error

	// ValidateRequest validates a request
	ValidateRequest(request Request) error

	// ValidateResponse validates a response
	ValidateResponse(response Response) error

	// ValidateNotification validates a notification
	ValidateNotification(notification Notification) error

	// ValidateCapabilities validates capabilities
	ValidateCapabilities(capabilities *Capabilities) error
}

// SecurityManager manages MCP security
type SecurityManager interface {
	// AuthenticateClient authenticates a client
	AuthenticateClient(ctx context.Context, clientInfo *ClientInfo, credentials map[string]interface{}) error

	// AuthorizeRequest authorizes a request
	AuthorizeRequest(ctx context.Context, session Session, request Request) error

	// AuthorizeNotification authorizes a notification
	AuthorizeNotification(ctx context.Context, session Session, notification Notification) error

	// GetPermissions returns permissions for a session
	GetPermissions(session Session) []string

	// ValidatePermission validates a specific permission
	ValidatePermission(session Session, permission string) bool
}

// MetricsCollector collects MCP metrics
type MetricsCollector interface {
	// RecordRequest records a request metric
	RecordRequest(method string, duration time.Duration, success bool)

	// RecordNotification records a notification metric
	RecordNotification(method string, success bool)

	// RecordSession records session metrics
	RecordSession(action string, sessionID string)

	// RecordError records an error metric
	RecordError(errorCode int, method string)

	// GetMetrics returns current metrics
	GetMetrics() map[string]interface{}
}

// EventEmitter emits MCP events
type EventEmitter interface {
	// EmitSessionCreated emits a session created event
	EmitSessionCreated(session Session)

	// EmitSessionClosed emits a session closed event
	EmitSessionClosed(sessionID string)

	// EmitRequestReceived emits a request received event
	EmitRequestReceived(session Session, request Request)

	// EmitResponseSent emits a response sent event
	EmitResponseSent(session Session, response Response)

	// EmitNotificationReceived emits a notification received event
	EmitNotificationReceived(session Session, notification Notification)

	// EmitError emits an error event
	EmitError(session Session, err error)
}

// ProtocolConfig represents MCP protocol configuration
type ProtocolConfig struct {
	Version              string                 `json:"version"`
	MaxMessageSize       int                    `json:"max_message_size"`
	RequestTimeout       time.Duration          `json:"request_timeout"`
	SessionTimeout       time.Duration          `json:"session_timeout"`
	MaxConcurrentSessions int                   `json:"max_concurrent_sessions"`
	EnableCompression    bool                   `json:"enable_compression"`
	EnableEncryption     bool                   `json:"enable_encryption"`
	LogLevel             string                 `json:"log_level"`
	Capabilities         Capabilities           `json:"capabilities"`
	Security             SecurityConfig         `json:"security"`
	Metadata             map[string]interface{} `json:"metadata"`
}

// SecurityConfig represents security configuration
type SecurityConfig struct {
	EnableAuthentication bool                   `json:"enable_authentication"`
	EnableAuthorization  bool                   `json:"enable_authorization"`
	AllowedClients       []string               `json:"allowed_clients"`
	RequiredPermissions  []string               `json:"required_permissions"`
	TokenValidation      TokenValidationConfig  `json:"token_validation"`
	RateLimit            RateLimitConfig        `json:"rate_limit"`
	Metadata             map[string]interface{} `json:"metadata"`
}

// TokenValidationConfig represents token validation configuration
type TokenValidationConfig struct {
	Enabled    bool          `json:"enabled"`
	Algorithm  string        `json:"algorithm"`
	SecretKey  string        `json:"secret_key"`
	Expiration time.Duration `json:"expiration"`
}

// RateLimitConfig represents rate limiting configuration
type RateLimitConfig struct {
	Enabled         bool          `json:"enabled"`
	RequestsPerMinute int         `json:"requests_per_minute"`
	BurstSize       int           `json:"burst_size"`
	WindowSize      time.Duration `json:"window_size"`
}
