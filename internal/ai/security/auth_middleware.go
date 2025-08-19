package security

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

// AuthMiddleware provides authentication and authorization for AI services
type AuthMiddleware struct {
	logger      *logrus.Logger
	tracer      trace.Tracer
	jwtSecret   []byte
	apiKeys     map[string]*APIKey
	sessions    map[string]*Session
	rateLimiter *RateLimiter
	mu          sync.RWMutex
}

// APIKey represents an API key for service access
type APIKey struct {
	ID          string                 `json:"id"`
	Key         string                 `json:"key"`
	Name        string                 `json:"name"`
	Permissions []string               `json:"permissions"`
	RateLimit   int                    `json:"rate_limit"`
	ExpiresAt   *time.Time             `json:"expires_at,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	LastUsed    *time.Time             `json:"last_used,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// Session represents a user session
type Session struct {
	ID        string                 `json:"id"`
	UserID    string                 `json:"user_id"`
	Token     string                 `json:"token"`
	ExpiresAt time.Time              `json:"expires_at"`
	CreatedAt time.Time              `json:"created_at"`
	LastUsed  time.Time              `json:"last_used"`
	IPAddress string                 `json:"ip_address"`
	UserAgent string                 `json:"user_agent"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// AuthContext represents authentication context
type AuthContext struct {
	UserID      string                 `json:"user_id"`
	SessionID   string                 `json:"session_id,omitempty"`
	APIKeyID    string                 `json:"api_key_id,omitempty"`
	Permissions []string               `json:"permissions"`
	RateLimit   int                    `json:"rate_limit"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// Claims represents JWT claims
type Claims struct {
	UserID      string   `json:"user_id"`
	SessionID   string   `json:"session_id"`
	Permissions []string `json:"permissions"`
	jwt.RegisteredClaims
}

// NewAuthMiddleware creates a new authentication middleware
func NewAuthMiddleware(logger *logrus.Logger, jwtSecret string) *AuthMiddleware {
	return &AuthMiddleware{
		logger:      logger,
		tracer:      otel.Tracer("ai.auth_middleware"),
		jwtSecret:   []byte(jwtSecret),
		apiKeys:     make(map[string]*APIKey),
		sessions:    make(map[string]*Session),
		rateLimiter: NewRateLimiter(),
	}
}

// CreateAPIKey creates a new API key
func (am *AuthMiddleware) CreateAPIKey(name string, permissions []string, rateLimit int, expiresAt *time.Time) (*APIKey, error) {
	am.mu.Lock()
	defer am.mu.Unlock()

	// Generate secure API key
	keyBytes := make([]byte, 32)
	if _, err := rand.Read(keyBytes); err != nil {
		return nil, fmt.Errorf("failed to generate API key: %w", err)
	}

	keyID := generateID()
	keyString := "aios_" + hex.EncodeToString(keyBytes)

	apiKey := &APIKey{
		ID:          keyID,
		Key:         keyString,
		Name:        name,
		Permissions: permissions,
		RateLimit:   rateLimit,
		ExpiresAt:   expiresAt,
		CreatedAt:   time.Now(),
		Metadata:    make(map[string]interface{}),
	}

	am.apiKeys[keyString] = apiKey

	am.logger.WithFields(logrus.Fields{
		"api_key_id":  keyID,
		"name":        name,
		"permissions": permissions,
		"rate_limit":  rateLimit,
	}).Info("API key created")

	return apiKey, nil
}

// ValidateAPIKey validates an API key
func (am *AuthMiddleware) ValidateAPIKey(ctx context.Context, keyString string) (*AuthContext, error) {
	ctx, span := am.tracer.Start(ctx, "auth.ValidateAPIKey")
	defer span.End()

	am.mu.RLock()
	apiKey, exists := am.apiKeys[keyString]
	am.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("invalid API key")
	}

	// Check expiration
	if apiKey.ExpiresAt != nil && time.Now().After(*apiKey.ExpiresAt) {
		return nil, fmt.Errorf("API key expired")
	}

	// Check rate limit
	if !am.rateLimiter.Allow(apiKey.ID, apiKey.RateLimit) {
		return nil, fmt.Errorf("rate limit exceeded")
	}

	// Update last used
	am.mu.Lock()
	now := time.Now()
	apiKey.LastUsed = &now
	am.mu.Unlock()

	authCtx := &AuthContext{
		UserID:      apiKey.ID,
		APIKeyID:    apiKey.ID,
		Permissions: apiKey.Permissions,
		RateLimit:   apiKey.RateLimit,
		Metadata:    apiKey.Metadata,
	}

	am.logger.WithFields(logrus.Fields{
		"api_key_id": apiKey.ID,
		"name":       apiKey.Name,
	}).Debug("API key validated")

	return authCtx, nil
}

// CreateSession creates a new user session
func (am *AuthMiddleware) CreateSession(userID, ipAddress, userAgent string) (*Session, string, error) {
	am.mu.Lock()
	defer am.mu.Unlock()

	sessionID := generateID()
	expiresAt := time.Now().Add(24 * time.Hour) // 24 hour session

	// Create JWT token
	claims := &Claims{
		UserID:      userID,
		SessionID:   sessionID,
		Permissions: []string{"ai:query", "ai:chat", "ai:voice", "ai:cv"}, // Default permissions
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "aios",
			Subject:   userID,
			ID:        sessionID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(am.jwtSecret)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create token: %w", err)
	}

	session := &Session{
		ID:        sessionID,
		UserID:    userID,
		Token:     tokenString,
		ExpiresAt: expiresAt,
		CreatedAt: time.Now(),
		LastUsed:  time.Now(),
		IPAddress: ipAddress,
		UserAgent: userAgent,
		Metadata:  make(map[string]interface{}),
	}

	am.sessions[sessionID] = session

	am.logger.WithFields(logrus.Fields{
		"session_id": sessionID,
		"user_id":    userID,
		"ip_address": ipAddress,
	}).Info("Session created")

	return session, tokenString, nil
}

// ValidateToken validates a JWT token
func (am *AuthMiddleware) ValidateToken(ctx context.Context, tokenString string) (*AuthContext, error) {
	ctx, span := am.tracer.Start(ctx, "auth.ValidateToken")
	defer span.End()

	// Parse and validate token
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return am.jwtSecret, nil
	})

	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}

	// Check session exists
	am.mu.RLock()
	session, exists := am.sessions[claims.SessionID]
	am.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("session not found")
	}

	// Check session expiration
	if time.Now().After(session.ExpiresAt) {
		am.mu.Lock()
		delete(am.sessions, claims.SessionID)
		am.mu.Unlock()
		return nil, fmt.Errorf("session expired")
	}

	// Update last used
	am.mu.Lock()
	session.LastUsed = time.Now()
	am.mu.Unlock()

	authCtx := &AuthContext{
		UserID:      claims.UserID,
		SessionID:   claims.SessionID,
		Permissions: claims.Permissions,
		RateLimit:   100, // Default rate limit for sessions
		Metadata:    session.Metadata,
	}

	am.logger.WithFields(logrus.Fields{
		"user_id":    claims.UserID,
		"session_id": claims.SessionID,
	}).Debug("Token validated")

	return authCtx, nil
}

// CheckPermission checks if the auth context has a specific permission
func (am *AuthMiddleware) CheckPermission(authCtx *AuthContext, permission string) bool {
	for _, perm := range authCtx.Permissions {
		if perm == permission || perm == "*" {
			return true
		}
		// Check wildcard permissions
		if strings.HasSuffix(perm, "*") {
			prefix := strings.TrimSuffix(perm, "*")
			if strings.HasPrefix(permission, prefix) {
				return true
			}
		}
	}
	return false
}

// RevokeAPIKey revokes an API key
func (am *AuthMiddleware) RevokeAPIKey(keyString string) error {
	am.mu.Lock()
	defer am.mu.Unlock()

	if apiKey, exists := am.apiKeys[keyString]; exists {
		delete(am.apiKeys, keyString)
		am.logger.WithFields(logrus.Fields{
			"api_key_id": apiKey.ID,
			"name":       apiKey.Name,
		}).Info("API key revoked")
		return nil
	}

	return fmt.Errorf("API key not found")
}

// RevokeSession revokes a session
func (am *AuthMiddleware) RevokeSession(sessionID string) error {
	am.mu.Lock()
	defer am.mu.Unlock()

	if session, exists := am.sessions[sessionID]; exists {
		delete(am.sessions, sessionID)
		am.logger.WithFields(logrus.Fields{
			"session_id": sessionID,
			"user_id":    session.UserID,
		}).Info("Session revoked")
		return nil
	}

	return fmt.Errorf("session not found")
}

// CleanupExpired removes expired sessions and API keys
func (am *AuthMiddleware) CleanupExpired() {
	am.mu.Lock()
	defer am.mu.Unlock()

	now := time.Now()

	// Clean up expired sessions
	for sessionID, session := range am.sessions {
		if now.After(session.ExpiresAt) {
			delete(am.sessions, sessionID)
			am.logger.WithField("session_id", sessionID).Debug("Expired session cleaned up")
		}
	}

	// Clean up expired API keys
	for keyString, apiKey := range am.apiKeys {
		if apiKey.ExpiresAt != nil && now.After(*apiKey.ExpiresAt) {
			delete(am.apiKeys, keyString)
			am.logger.WithField("api_key_id", apiKey.ID).Debug("Expired API key cleaned up")
		}
	}
}

// GetStats returns authentication statistics
func (am *AuthMiddleware) GetStats() map[string]interface{} {
	am.mu.RLock()
	defer am.mu.RUnlock()

	activeAPIKeys := 0
	expiredAPIKeys := 0
	now := time.Now()

	for _, apiKey := range am.apiKeys {
		if apiKey.ExpiresAt == nil || now.Before(*apiKey.ExpiresAt) {
			activeAPIKeys++
		} else {
			expiredAPIKeys++
		}
	}

	activeSessions := 0
	expiredSessions := 0

	for _, session := range am.sessions {
		if now.Before(session.ExpiresAt) {
			activeSessions++
		} else {
			expiredSessions++
		}
	}

	return map[string]interface{}{
		"active_api_keys":  activeAPIKeys,
		"expired_api_keys": expiredAPIKeys,
		"active_sessions":  activeSessions,
		"expired_sessions": expiredSessions,
		"rate_limit_stats": am.rateLimiter.GetStats(),
	}
}

// Helper functions

func generateID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

func hashString(s string) string {
	hash := sha256.Sum256([]byte(s))
	return hex.EncodeToString(hash[:])
}
