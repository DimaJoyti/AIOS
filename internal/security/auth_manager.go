package security

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"sync"
	"time"

	"github.com/aios/aios/pkg/models"
	"github.com/golang-jwt/jwt/v5"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/crypto/bcrypt"
)

// AuthManager handles authentication and authorization
type AuthManager struct {
	logger         *logrus.Logger
	tracer         trace.Tracer
	config         AuthConfig
	sessions       map[string]*Session
	users          map[string]*User
	refreshTokens  map[string]*RefreshToken
	mfaTokens      map[string]*MFAToken
	mu             sync.RWMutex
	running        bool
	stopCh         chan struct{}
}

// Session represents an active user session
type Session struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	IPAddress string    `json:"ip_address"`
	UserAgent string    `json:"user_agent"`
	CreatedAt time.Time `json:"created_at"`
	LastUsed  time.Time `json:"last_used"`
}

// User represents a user account
type User struct {
	ID           string            `json:"id"`
	Username     string            `json:"username"`
	Email        string            `json:"email"`
	PasswordHash string            `json:"password_hash"`
	Roles        []string          `json:"roles"`
	Permissions  []string          `json:"permissions"`
	MFAEnabled   bool              `json:"mfa_enabled"`
	MFASecret    string            `json:"mfa_secret"`
	Metadata     map[string]string `json:"metadata"`
	CreatedAt    time.Time         `json:"created_at"`
	UpdatedAt    time.Time         `json:"updated_at"`
	LastLogin    time.Time         `json:"last_login"`
	FailedLogins int               `json:"failed_logins"`
	Locked       bool              `json:"locked"`
	Active       bool              `json:"active"`
}

// RefreshToken represents a refresh token
type RefreshToken struct {
	Token     string    `json:"token"`
	UserID    string    `json:"user_id"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

// MFAToken represents a multi-factor authentication token
type MFAToken struct {
	Token     string    `json:"token"`
	UserID    string    `json:"user_id"`
	Method    string    `json:"method"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
	Used      bool      `json:"used"`
}

// LoginRequest represents a login request
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	MFAToken string `json:"mfa_token,omitempty"`
}

// LoginResponse represents a login response
type LoginResponse struct {
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
	User         *User  `json:"user"`
}

// NewAuthManager creates a new authentication manager
func NewAuthManager(logger *logrus.Logger, config AuthConfig) (*AuthManager, error) {
	tracer := otel.Tracer("auth-manager")

	return &AuthManager{
		logger:        logger,
		tracer:        tracer,
		config:        config,
		sessions:      make(map[string]*Session),
		users:         make(map[string]*User),
		refreshTokens: make(map[string]*RefreshToken),
		mfaTokens:     make(map[string]*MFAToken),
		stopCh:        make(chan struct{}),
	}, nil
}

// Start initializes the authentication manager
func (am *AuthManager) Start(ctx context.Context) error {
	ctx, span := am.tracer.Start(ctx, "authManager.Start")
	defer span.End()

	am.mu.Lock()
	defer am.mu.Unlock()

	if am.running {
		return fmt.Errorf("auth manager is already running")
	}

	if !am.config.Enabled {
		am.logger.Info("Authentication manager is disabled")
		return nil
	}

	am.logger.Info("Starting authentication manager")

	// Initialize default admin user if no users exist
	if len(am.users) == 0 {
		if err := am.createDefaultAdmin(); err != nil {
			return fmt.Errorf("failed to create default admin user: %w", err)
		}
	}

	// Start session cleanup
	go am.cleanupSessions()

	am.running = true
	am.logger.Info("Authentication manager started successfully")

	return nil
}

// Stop shuts down the authentication manager
func (am *AuthManager) Stop(ctx context.Context) error {
	ctx, span := am.tracer.Start(ctx, "authManager.Stop")
	defer span.End()

	am.mu.Lock()
	defer am.mu.Unlock()

	if !am.running {
		return nil
	}

	am.logger.Info("Stopping authentication manager")

	close(am.stopCh)
	am.running = false
	am.logger.Info("Authentication manager stopped")

	return nil
}

// GetStatus returns the current authentication status
func (am *AuthManager) GetStatus(ctx context.Context) (*models.AuthenticationStatus, error) {
	ctx, span := am.tracer.Start(ctx, "authManager.GetStatus")
	defer span.End()

	am.mu.RLock()
	defer am.mu.RUnlock()

	activeSessions := 0
	failedAttempts := 0
	var lastLogin time.Time

	for _, session := range am.sessions {
		if session.ExpiresAt.After(time.Now()) {
			activeSessions++
		}
	}

	for _, user := range am.users {
		failedAttempts += user.FailedLogins
		if user.LastLogin.After(lastLogin) {
			lastLogin = user.LastLogin
		}
	}

	return &models.AuthenticationStatus{
		Enabled:        am.config.Enabled,
		ActiveSessions: activeSessions,
		MFAEnabled:     am.config.MFA.Enabled,
		OAuthEnabled:   am.config.OAuth.Enabled,
		LDAPEnabled:    am.config.LDAP.Enabled,
		LastLogin:      lastLogin,
		FailedAttempts: failedAttempts,
		Timestamp:      time.Now(),
	}, nil
}

// Login authenticates a user and creates a session
func (am *AuthManager) Login(ctx context.Context, req LoginRequest) (*LoginResponse, error) {
	ctx, span := am.tracer.Start(ctx, "authManager.Login")
	defer span.End()

	am.mu.Lock()
	defer am.mu.Unlock()

	// Find user
	user, exists := am.users[req.Username]
	if !exists {
		am.logger.WithField("username", req.Username).Warn("Login attempt with invalid username")
		return nil, fmt.Errorf("invalid credentials")
	}

	// Check if user is locked
	if user.Locked {
		am.logger.WithField("username", req.Username).Warn("Login attempt for locked user")
		return nil, fmt.Errorf("account is locked")
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		user.FailedLogins++
		if user.FailedLogins >= 5 {
			user.Locked = true
			am.logger.WithField("username", req.Username).Warn("User account locked due to failed login attempts")
		}
		return nil, fmt.Errorf("invalid credentials")
	}

	// Check MFA if enabled
	if user.MFAEnabled && am.config.MFA.Enabled {
		if req.MFAToken == "" {
			return nil, fmt.Errorf("MFA token required")
		}
		if !am.validateMFAToken(user.ID, req.MFAToken) {
			return nil, fmt.Errorf("invalid MFA token")
		}
	}

	// Reset failed login attempts
	user.FailedLogins = 0
	user.LastLogin = time.Now()

	// Generate JWT token
	token, err := am.generateJWT(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	// Generate refresh token
	refreshToken, err := am.generateRefreshToken(user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Create session
	session := &Session{
		ID:        generateSessionID(),
		UserID:    user.ID,
		Token:     token,
		ExpiresAt: time.Now().Add(am.config.SessionTimeout),
		CreatedAt: time.Now(),
		LastUsed:  time.Now(),
	}

	am.sessions[session.ID] = session

	am.logger.WithFields(logrus.Fields{
		"user_id":    user.ID,
		"username":   user.Username,
		"session_id": session.ID,
	}).Info("User logged in successfully")

	return &LoginResponse{
		Token:        token,
		RefreshToken: refreshToken,
		ExpiresAt:    session.ExpiresAt,
		User:         user,
	}, nil
}

// Logout invalidates a user session
func (am *AuthManager) Logout(ctx context.Context, sessionID string) error {
	ctx, span := am.tracer.Start(ctx, "authManager.Logout")
	defer span.End()

	am.mu.Lock()
	defer am.mu.Unlock()

	session, exists := am.sessions[sessionID]
	if !exists {
		return fmt.Errorf("session not found")
	}

	delete(am.sessions, sessionID)

	am.logger.WithFields(logrus.Fields{
		"user_id":    session.UserID,
		"session_id": sessionID,
	}).Info("User logged out")

	return nil
}

// ValidateToken validates a JWT token
func (am *AuthManager) ValidateToken(ctx context.Context, tokenString string) (*User, error) {
	ctx, span := am.tracer.Start(ctx, "authManager.ValidateToken")
	defer span.End()

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(am.config.JWTSecret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("token is not valid")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	userID, ok := claims["user_id"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid user_id in token")
	}

	am.mu.RLock()
	user, exists := am.users[userID]
	am.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("user not found")
	}

	return user, nil
}

// Helper methods

func (am *AuthManager) createDefaultAdmin() error {
	adminPassword := "admin123" // TODO: Generate secure password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(adminPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	admin := &User{
		ID:           "admin",
		Username:     "admin",
		Email:        "admin@aios.local",
		PasswordHash: string(hashedPassword),
		Roles:        []string{"admin"},
		Permissions:  []string{"*"},
		MFAEnabled:   false,
		Active:       true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	am.users[admin.Username] = admin
	am.logger.Info("Default admin user created")

	return nil
}

func (am *AuthManager) generateJWT(user *User) (string, error) {
	claims := jwt.MapClaims{
		"user_id":  user.ID,
		"username": user.Username,
		"roles":    user.Roles,
		"exp":      time.Now().Add(am.config.SessionTimeout).Unix(),
		"iat":      time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(am.config.JWTSecret))
}

func (am *AuthManager) generateRefreshToken(userID string) (string, error) {
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", err
	}

	token := base64.URLEncoding.EncodeToString(tokenBytes)
	refreshToken := &RefreshToken{
		Token:     token,
		UserID:    userID,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour), // 7 days
		CreatedAt: time.Now(),
	}

	am.refreshTokens[token] = refreshToken
	return token, nil
}

func (am *AuthManager) validateMFAToken(userID, token string) bool {
	// TODO: Implement actual MFA validation (TOTP, SMS, etc.)
	return token == "123456" // Mock validation
}

func (am *AuthManager) cleanupSessions() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			am.mu.Lock()
			now := time.Now()
			for id, session := range am.sessions {
				if session.ExpiresAt.Before(now) {
					delete(am.sessions, id)
					am.logger.WithField("session_id", id).Debug("Expired session cleaned up")
				}
			}
			am.mu.Unlock()

		case <-am.stopCh:
			am.logger.Debug("Session cleanup stopped")
			return
		}
	}
}

func generateSessionID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return base64.URLEncoding.EncodeToString(bytes)
}
