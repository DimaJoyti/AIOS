package server

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"github.com/aios/aios/pkg/mcp/protocol"
	"github.com/sirupsen/logrus"
)

// SecurityManager handles authentication and authorization for MCP server
type SecurityManager interface {
	// Authentication
	Authenticate(ctx context.Context, credentials *Credentials) (*AuthResult, error)
	ValidateToken(ctx context.Context, token string) (*TokenInfo, error)
	RefreshToken(ctx context.Context, refreshToken string) (*AuthResult, error)
	RevokeToken(ctx context.Context, token string) error
	
	// Authorization
	Authorize(ctx context.Context, user *User, resource string, action string) error
	CheckPermission(ctx context.Context, user *User, permission string) bool
	
	// User management
	CreateUser(ctx context.Context, user *User) error
	GetUser(ctx context.Context, userID string) (*User, error)
	UpdateUser(ctx context.Context, user *User) error
	DeleteUser(ctx context.Context, userID string) error
	
	// Session management
	CreateSession(ctx context.Context, user *User) (*Session, error)
	GetSession(ctx context.Context, sessionID string) (*Session, error)
	InvalidateSession(ctx context.Context, sessionID string) error
}

// Credentials represents authentication credentials
type Credentials struct {
	Type     string                 `json:"type"`     // "password", "token", "api_key"
	Username string                 `json:"username,omitempty"`
	Password string                 `json:"password,omitempty"`
	Token    string                 `json:"token,omitempty"`
	APIKey   string                 `json:"api_key,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// AuthResult represents authentication result
type AuthResult struct {
	Success      bool      `json:"success"`
	User         *User     `json:"user,omitempty"`
	AccessToken  string    `json:"access_token,omitempty"`
	RefreshToken string    `json:"refresh_token,omitempty"`
	ExpiresAt    time.Time `json:"expires_at,omitempty"`
	Permissions  []string  `json:"permissions,omitempty"`
}

// TokenInfo represents token information
type TokenInfo struct {
	UserID      string    `json:"user_id"`
	Username    string    `json:"username"`
	Permissions []string  `json:"permissions"`
	ExpiresAt   time.Time `json:"expires_at"`
	Valid       bool      `json:"valid"`
}

// User represents a user
type User struct {
	ID          string                 `json:"id"`
	Username    string                 `json:"username"`
	Email       string                 `json:"email,omitempty"`
	PasswordHash string                `json:"password_hash,omitempty"`
	Permissions []string               `json:"permissions"`
	Roles       []string               `json:"roles"`
	Active      bool                   `json:"active"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	LastLogin   *time.Time             `json:"last_login,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// Session represents a user session
type Session struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Token     string    `json:"token"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
	Active    bool      `json:"active"`
	IPAddress string    `json:"ip_address,omitempty"`
	UserAgent string    `json:"user_agent,omitempty"`
}

// SecurityConfig represents security configuration
type SecurityConfig struct {
	EnableAuthentication bool          `json:"enable_authentication"`
	EnableAuthorization  bool          `json:"enable_authorization"`
	TokenExpiration      time.Duration `json:"token_expiration"`
	RefreshExpiration    time.Duration `json:"refresh_expiration"`
	SessionExpiration    time.Duration `json:"session_expiration"`
	MaxSessions          int           `json:"max_sessions"`
	RequireHTTPS         bool          `json:"require_https"`
	AllowedOrigins       []string      `json:"allowed_origins"`
	RateLimiting         bool          `json:"rate_limiting"`
}

// DefaultSecurityManager implements SecurityManager
type DefaultSecurityManager struct {
	config   *SecurityConfig
	users    map[string]*User
	sessions map[string]*Session
	tokens   map[string]*TokenInfo
	mu       sync.RWMutex
	logger   *logrus.Logger
}

// NewSecurityManager creates a new security manager
func NewSecurityManager(config *SecurityConfig, logger *logrus.Logger) (SecurityManager, error) {
	return &DefaultSecurityManager{
		config:   config,
		users:    make(map[string]*User),
		sessions: make(map[string]*Session),
		tokens:   make(map[string]*TokenInfo),
		logger:   logger,
	}, nil
}

// Authenticate authenticates a user
func (sm *DefaultSecurityManager) Authenticate(ctx context.Context, credentials *Credentials) (*AuthResult, error) {
	switch credentials.Type {
	case "password":
		return sm.authenticatePassword(ctx, credentials)
	case "token":
		return sm.authenticateToken(ctx, credentials)
	case "api_key":
		return sm.authenticateAPIKey(ctx, credentials)
	default:
		return &AuthResult{Success: false}, fmt.Errorf("unsupported credential type: %s", credentials.Type)
	}
}

// authenticatePassword authenticates using username/password
func (sm *DefaultSecurityManager) authenticatePassword(ctx context.Context, credentials *Credentials) (*AuthResult, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	
	// Find user by username
	var user *User
	for _, u := range sm.users {
		if u.Username == credentials.Username {
			user = u
			break
		}
	}
	
	if user == nil {
		return &AuthResult{Success: false}, fmt.Errorf("user not found")
	}
	
	if !user.Active {
		return &AuthResult{Success: false}, fmt.Errorf("user account is disabled")
	}
	
	// Verify password
	if !sm.verifyPassword(credentials.Password, user.PasswordHash) {
		return &AuthResult{Success: false}, fmt.Errorf("invalid password")
	}
	
	// Generate tokens
	accessToken, err := sm.generateToken(user)
	if err != nil {
		return &AuthResult{Success: false}, err
	}
	
	refreshToken, err := sm.generateRefreshToken(user)
	if err != nil {
		return &AuthResult{Success: false}, err
	}
	
	// Update last login
	now := time.Now()
	user.LastLogin = &now
	
	return &AuthResult{
		Success:      true,
		User:         user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    time.Now().Add(sm.config.TokenExpiration),
		Permissions:  user.Permissions,
	}, nil
}

// authenticateToken authenticates using a token
func (sm *DefaultSecurityManager) authenticateToken(ctx context.Context, credentials *Credentials) (*AuthResult, error) {
	tokenInfo, err := sm.ValidateToken(ctx, credentials.Token)
	if err != nil {
		return &AuthResult{Success: false}, err
	}
	
	if !tokenInfo.Valid {
		return &AuthResult{Success: false}, fmt.Errorf("invalid token")
	}
	
	user, err := sm.GetUser(ctx, tokenInfo.UserID)
	if err != nil {
		return &AuthResult{Success: false}, err
	}
	
	return &AuthResult{
		Success:     true,
		User:        user,
		Permissions: tokenInfo.Permissions,
	}, nil
}

// authenticateAPIKey authenticates using an API key
func (sm *DefaultSecurityManager) authenticateAPIKey(ctx context.Context, credentials *Credentials) (*AuthResult, error) {
	// Simple API key validation (in production, use proper key management)
	if credentials.APIKey == "" {
		return &AuthResult{Success: false}, fmt.Errorf("API key required")
	}
	
	// For demo purposes, accept any non-empty API key
	// In production, validate against stored API keys
	return &AuthResult{
		Success:     true,
		Permissions: []string{"read", "write"},
	}, nil
}

// ValidateToken validates a token
func (sm *DefaultSecurityManager) ValidateToken(ctx context.Context, token string) (*TokenInfo, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	
	tokenInfo, exists := sm.tokens[token]
	if !exists {
		return &TokenInfo{Valid: false}, fmt.Errorf("token not found")
	}
	
	if time.Now().After(tokenInfo.ExpiresAt) {
		return &TokenInfo{Valid: false}, fmt.Errorf("token expired")
	}
	
	return tokenInfo, nil
}

// RefreshToken refreshes an access token
func (sm *DefaultSecurityManager) RefreshToken(ctx context.Context, refreshToken string) (*AuthResult, error) {
	// Simple refresh token validation
	// In production, implement proper refresh token management
	return &AuthResult{
		Success:     false,
	}, fmt.Errorf("refresh token not implemented")
}

// RevokeToken revokes a token
func (sm *DefaultSecurityManager) RevokeToken(ctx context.Context, token string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	
	delete(sm.tokens, token)
	return nil
}

// Authorize checks if a user is authorized for a resource/action
func (sm *DefaultSecurityManager) Authorize(ctx context.Context, user *User, resource string, action string) error {
	if !sm.config.EnableAuthorization {
		return nil
	}
	
	// Simple permission check
	requiredPermission := fmt.Sprintf("%s:%s", resource, action)
	
	for _, permission := range user.Permissions {
		if permission == requiredPermission || permission == "*" {
			return nil
		}
	}
	
	return fmt.Errorf("access denied: insufficient permissions")
}

// CheckPermission checks if a user has a specific permission
func (sm *DefaultSecurityManager) CheckPermission(ctx context.Context, user *User, permission string) bool {
	for _, p := range user.Permissions {
		if p == permission || p == "*" {
			return true
		}
	}
	return false
}

// CreateUser creates a new user
func (sm *DefaultSecurityManager) CreateUser(ctx context.Context, user *User) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	
	if user.ID == "" {
		user.ID = generateID()
	}
	
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	
	sm.users[user.ID] = user
	return nil
}

// GetUser gets a user by ID
func (sm *DefaultSecurityManager) GetUser(ctx context.Context, userID string) (*User, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	
	user, exists := sm.users[userID]
	if !exists {
		return nil, fmt.Errorf("user not found: %s", userID)
	}
	
	return user, nil
}

// UpdateUser updates a user
func (sm *DefaultSecurityManager) UpdateUser(ctx context.Context, user *User) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	
	existing, exists := sm.users[user.ID]
	if !exists {
		return fmt.Errorf("user not found: %s", user.ID)
	}
	
	user.CreatedAt = existing.CreatedAt
	user.UpdatedAt = time.Now()
	
	sm.users[user.ID] = user
	return nil
}

// DeleteUser deletes a user
func (sm *DefaultSecurityManager) DeleteUser(ctx context.Context, userID string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	
	delete(sm.users, userID)
	return nil
}

// CreateSession creates a new session
func (sm *DefaultSecurityManager) CreateSession(ctx context.Context, user *User) (*Session, error) {
	session := &Session{
		ID:        generateID(),
		UserID:    user.ID,
		Token:     generateToken(),
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(sm.config.SessionExpiration),
		Active:    true,
	}
	
	sm.mu.Lock()
	defer sm.mu.Unlock()
	
	sm.sessions[session.ID] = session
	return session, nil
}

// GetSession gets a session by ID
func (sm *DefaultSecurityManager) GetSession(ctx context.Context, sessionID string) (*Session, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	
	session, exists := sm.sessions[sessionID]
	if !exists {
		return nil, fmt.Errorf("session not found: %s", sessionID)
	}
	
	if time.Now().After(session.ExpiresAt) {
		return nil, fmt.Errorf("session expired")
	}
	
	return session, nil
}

// InvalidateSession invalidates a session
func (sm *DefaultSecurityManager) InvalidateSession(ctx context.Context, sessionID string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	
	if session, exists := sm.sessions[sessionID]; exists {
		session.Active = false
	}
	
	return nil
}

// Helper functions

func (sm *DefaultSecurityManager) verifyPassword(password, hash string) bool {
	// Simple password verification (in production, use proper hashing)
	return sm.hashPassword(password) == hash
}

func (sm *DefaultSecurityManager) hashPassword(password string) string {
	hash := sha256.Sum256([]byte(password))
	return hex.EncodeToString(hash[:])
}

func (sm *DefaultSecurityManager) generateToken(user *User) (string, error) {
	token := generateToken()
	
	tokenInfo := &TokenInfo{
		UserID:      user.ID,
		Username:    user.Username,
		Permissions: user.Permissions,
		ExpiresAt:   time.Now().Add(sm.config.TokenExpiration),
		Valid:       true,
	}
	
	sm.mu.Lock()
	sm.tokens[token] = tokenInfo
	sm.mu.Unlock()
	
	return token, nil
}

func (sm *DefaultSecurityManager) generateRefreshToken(user *User) (string, error) {
	return generateToken(), nil
}

func generateID() string {
	return generateToken()[:16]
}

func generateToken() string {
	bytes := make([]byte, 32)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// AuthenticateClient authenticates a client (protocol.SecurityManager implementation)
func (sm *DefaultSecurityManager) AuthenticateClient(ctx context.Context, clientInfo *protocol.ClientInfo, credentials map[string]interface{}) error {
	// Simple implementation - for production use proper authentication
	if clientInfo == nil {
		return fmt.Errorf("client info required")
	}
	return nil
}

// AuthorizeRequest authorizes a request (protocol.SecurityManager implementation)
func (sm *DefaultSecurityManager) AuthorizeRequest(ctx context.Context, session protocol.Session, request protocol.Request) error {
	if !sm.config.EnableAuthorization {
		return nil
	}
	// Simple implementation - for production use proper authorization
	return nil
}

// AuthorizeNotification authorizes a notification (protocol.SecurityManager implementation)
func (sm *DefaultSecurityManager) AuthorizeNotification(ctx context.Context, session protocol.Session, notification protocol.Notification) error {
	if !sm.config.EnableAuthorization {
		return nil
	}
	// Simple implementation - for production use proper authorization
	return nil
}

// GetPermissions returns permissions for a session (protocol.SecurityManager implementation)
func (sm *DefaultSecurityManager) GetPermissions(session protocol.Session) []string {
	// Simple implementation - return default permissions
	return []string{"read", "write"}
}

// ValidatePermission validates a specific permission (protocol.SecurityManager implementation)
func (sm *DefaultSecurityManager) ValidatePermission(session protocol.Session, permission string) bool {
	permissions := sm.GetPermissions(session)
	for _, p := range permissions {
		if p == permission || p == "*" {
			return true
		}
	}
	return false
}
