package security

import (
	"context"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSecurityManager(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	config := SecurityConfig{
		Enabled: true,
		Authentication: AuthConfig{
			Enabled:        true,
			JWTSecret:      "test-secret",
			SessionTimeout: 1 * time.Hour,
			MFA: MFAConfig{
				Enabled: false,
			},
			OAuth: OAuthConfig{
				Enabled: false,
			},
			LDAP: LDAPConfig{
				Enabled: false,
			},
			PasswordPolicy: PasswordPolicyConfig{
				MinLength:        8,
				RequireUppercase: true,
				RequireLowercase: true,
				RequireNumbers:   true,
				RequireSymbols:   false,
				MaxAge:           90 * 24 * time.Hour,
				HistorySize:      5,
			},
		},
		Encryption: EncryptionConfig{
			Enabled:     true,
			Algorithm:   "AES-256-GCM",
			KeySize:     256,
			AtRest:      true,
			InTransit:   true,
			KeyRotation: 24 * time.Hour,
			HSMEnabled:  false,
		},
		Privacy: PrivacyConfig{
			Enabled:           true,
			DataMinimization:  true,
			Anonymization:     true,
			Pseudonymization:  false,
			DataRetention:     30 * 24 * time.Hour,
			ConsentManagement: true,
			RightToErasure:    true,
			DataPortability:   true,
			PIIDetection:      true,
		},
		ThreatDetection: ThreatDetectionConfig{
			Enabled:           true,
			RealTime:          false,
			MachineLearning:   false,
			BehavioralAnalysis: true,
			NetworkMonitoring: true,
			FileIntegrity:     true,
			AlertThresholds: map[string]float64{
				"anomaly_score": 0.8,
			},
			ResponseActions: []string{"log", "alert"},
		},
		Audit: AuditConfig{
			Enabled:         true,
			LogLevel:        "info",
			RetentionPeriod: 90 * 24 * time.Hour,
			Encryption:      true,
			Integrity:       true,
			RemoteLogging:   false,
			SIEMIntegration: false,
		},
		AccessControl: AccessControlConfig{
			Enabled:       true,
			Model:         "RBAC",
			DefaultPolicy: "deny",
			Roles: map[string][]string{
				"admin": {"*"},
				"user":  {"read", "write"},
			},
			Permissions: map[string][]string{
				"system": {"read", "write", "admin"},
				"data":   {"read", "write"},
			},
			SessionTimeout: 8 * time.Hour,
			MaxSessions:    5,
		},
		Compliance: ComplianceConfig{
			Enabled:    true,
			Standards:  []string{"GDPR", "SOC2"},
			Reporting:  true,
			Monitoring: true,
			Automation: false,
		},
		IncidentResponse: IncidentResponseConfig{
			Enabled:      true,
			AutoResponse: false,
		},
		VulnerabilityScanning: VulnerabilityConfig{
			Enabled:         true,
			ScanInterval:    24 * time.Hour,
			ScanTypes:       []string{"static", "dynamic"},
			AutoRemediation: false,
		},
	}

	t.Run("NewManager", func(t *testing.T) {
		manager, err := NewManager(logger, config)
		require.NoError(t, err)
		assert.NotNil(t, manager)
		assert.Equal(t, config.Enabled, manager.config.Enabled)
	})

	t.Run("StartStop", func(t *testing.T) {
		manager, err := NewManager(logger, config)
		require.NoError(t, err)

		ctx := context.Background()

		// Start manager
		err = manager.Start(ctx)
		require.NoError(t, err)

		// Check status
		status, err := manager.GetStatus(ctx)
		require.NoError(t, err)
		assert.True(t, status.Enabled)
		assert.True(t, status.Running)

		// Stop manager
		err = manager.Stop(ctx)
		require.NoError(t, err)

		// Check status after stop
		status, err = manager.GetStatus(ctx)
		require.NoError(t, err)
		assert.True(t, status.Enabled)
		assert.False(t, status.Running)
	})

	t.Run("GetComponents", func(t *testing.T) {
		manager, err := NewManager(logger, config)
		require.NoError(t, err)

		// Test component getters
		assert.NotNil(t, manager.GetAuthManager())
		assert.NotNil(t, manager.GetEncryptionManager())
		assert.NotNil(t, manager.GetPrivacyManager())
		assert.NotNil(t, manager.GetThreatDetector())
		assert.NotNil(t, manager.GetAuditLogger())
		assert.NotNil(t, manager.GetAccessController())
		assert.NotNil(t, manager.GetComplianceManager())
		assert.NotNil(t, manager.GetIncidentResponder())
		assert.NotNil(t, manager.GetVulnerabilityScanner())
	})

	t.Run("DisabledManager", func(t *testing.T) {
		disabledConfig := config
		disabledConfig.Enabled = false

		manager, err := NewManager(logger, disabledConfig)
		require.NoError(t, err)

		ctx := context.Background()

		// Start disabled manager
		err = manager.Start(ctx)
		require.NoError(t, err)

		// Check status
		status, err := manager.GetStatus(ctx)
		require.NoError(t, err)
		assert.False(t, status.Enabled)
		assert.False(t, status.Running)
	})
}

func TestAuthManager(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	config := AuthConfig{
		Enabled:        true,
		JWTSecret:      "test-secret",
		SessionTimeout: 1 * time.Hour,
		MFA: MFAConfig{
			Enabled: false,
		},
		OAuth: OAuthConfig{
			Enabled: false,
		},
		LDAP: LDAPConfig{
			Enabled: false,
		},
		PasswordPolicy: PasswordPolicyConfig{
			MinLength:        8,
			RequireUppercase: true,
			RequireLowercase: true,
			RequireNumbers:   true,
			RequireSymbols:   false,
		},
	}

	t.Run("NewAuthManager", func(t *testing.T) {
		authManager, err := NewAuthManager(logger, config)
		require.NoError(t, err)
		assert.NotNil(t, authManager)
	})

	t.Run("StartStop", func(t *testing.T) {
		authManager, err := NewAuthManager(logger, config)
		require.NoError(t, err)

		ctx := context.Background()

		err = authManager.Start(ctx)
		require.NoError(t, err)

		status, err := authManager.GetStatus(ctx)
		require.NoError(t, err)
		assert.True(t, status.Enabled)

		err = authManager.Stop(ctx)
		require.NoError(t, err)
	})

	t.Run("Login", func(t *testing.T) {
		authManager, err := NewAuthManager(logger, config)
		require.NoError(t, err)

		ctx := context.Background()
		err = authManager.Start(ctx)
		require.NoError(t, err)
		defer authManager.Stop(ctx)

		// Test login with default admin user
		loginReq := LoginRequest{
			Username: "admin",
			Password: "admin123",
		}

		response, err := authManager.Login(ctx, loginReq)
		require.NoError(t, err)
		assert.NotEmpty(t, response.Token)
		assert.NotEmpty(t, response.RefreshToken)
		assert.Equal(t, "admin", response.User.Username)
	})

	t.Run("ValidateToken", func(t *testing.T) {
		authManager, err := NewAuthManager(logger, config)
		require.NoError(t, err)

		ctx := context.Background()
		err = authManager.Start(ctx)
		require.NoError(t, err)
		defer authManager.Stop(ctx)

		// Login to get a token
		loginReq := LoginRequest{
			Username: "admin",
			Password: "admin123",
		}

		response, err := authManager.Login(ctx, loginReq)
		require.NoError(t, err)

		// Validate the token
		user, err := authManager.ValidateToken(ctx, response.Token)
		require.NoError(t, err)
		assert.Equal(t, "admin", user.Username)
	})
}

func TestEncryptionManager(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	config := EncryptionConfig{
		Enabled:     true,
		Algorithm:   "AES-256-GCM",
		KeySize:     256,
		AtRest:      true,
		InTransit:   true,
		KeyRotation: 24 * time.Hour,
		HSMEnabled:  false,
	}

	t.Run("NewEncryptionManager", func(t *testing.T) {
		encManager, err := NewEncryptionManager(logger, config)
		require.NoError(t, err)
		assert.NotNil(t, encManager)
	})

	t.Run("EncryptDecrypt", func(t *testing.T) {
		encManager, err := NewEncryptionManager(logger, config)
		require.NoError(t, err)

		ctx := context.Background()
		err = encManager.Start(ctx)
		require.NoError(t, err)
		defer encManager.Stop(ctx)

		// Test data encryption and decryption
		plaintext := []byte("Hello, World! This is a test message.")

		encrypted, err := encManager.Encrypt(ctx, plaintext)
		require.NoError(t, err)
		assert.NotEqual(t, plaintext, encrypted)

		decrypted, err := encManager.Decrypt(ctx, encrypted)
		require.NoError(t, err)
		assert.Equal(t, plaintext, decrypted)
	})

	t.Run("EncryptDecryptString", func(t *testing.T) {
		encManager, err := NewEncryptionManager(logger, config)
		require.NoError(t, err)

		ctx := context.Background()
		err = encManager.Start(ctx)
		require.NoError(t, err)
		defer encManager.Stop(ctx)

		// Test string encryption and decryption
		plaintext := "Hello, World! This is a test message."

		encrypted, err := encManager.EncryptString(ctx, plaintext)
		require.NoError(t, err)
		assert.NotEqual(t, plaintext, encrypted)

		decrypted, err := encManager.DecryptString(ctx, encrypted)
		require.NoError(t, err)
		assert.Equal(t, plaintext, decrypted)
	})
}
