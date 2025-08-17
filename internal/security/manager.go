package security

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"github.com/aios/aios/pkg/models"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

// Manager handles security and privacy operations
type Manager struct {
	logger               *logrus.Logger
	tracer               trace.Tracer
	config               SecurityConfig
	authManager          *AuthManager
	encryptionManager    *EncryptionManager
	privacyManager       *PrivacyManager
	threatDetector       *ThreatDetector
	auditLogger          *AuditLogger
	accessController     *AccessController
	complianceManager    *ComplianceManager
	incidentResponder    *IncidentResponder
	vulnerabilityScanner *VulnerabilityScanner
	mu                   sync.RWMutex
	running              bool
	stopCh               chan struct{}
}

// SecurityConfig represents security configuration
type SecurityConfig struct {
	Enabled               bool                   `yaml:"enabled"`
	Authentication        AuthConfig             `yaml:"authentication"`
	Encryption            EncryptionConfig       `yaml:"encryption"`
	Privacy               PrivacyConfig          `yaml:"privacy"`
	ThreatDetection       ThreatDetectionConfig  `yaml:"threat_detection"`
	Audit                 AuditConfig            `yaml:"audit"`
	AccessControl         AccessControlConfig    `yaml:"access_control"`
	Compliance            ComplianceConfig       `yaml:"compliance"`
	IncidentResponse      IncidentResponseConfig `yaml:"incident_response"`
	VulnerabilityScanning VulnerabilityConfig    `yaml:"vulnerability_scanning"`
	RateLimiting          RateLimitConfig        `yaml:"rate_limiting"`
	CORS                  CORSConfig             `yaml:"cors"`
	TLS                   TLSConfig              `yaml:"tls"`
}

// AuthConfig represents authentication configuration
type AuthConfig struct {
	Enabled        bool                 `yaml:"enabled"`
	JWTSecret      string               `yaml:"jwt_secret"`
	SessionTimeout time.Duration        `yaml:"session_timeout"`
	MFA            MFAConfig            `yaml:"mfa"`
	OAuth          OAuthConfig          `yaml:"oauth"`
	LDAP           LDAPConfig           `yaml:"ldap"`
	PasswordPolicy PasswordPolicyConfig `yaml:"password_policy"`
}

// MFAConfig represents multi-factor authentication configuration
type MFAConfig struct {
	Enabled     bool          `yaml:"enabled"`
	Methods     []string      `yaml:"methods"`
	Required    bool          `yaml:"required"`
	GracePeriod time.Duration `yaml:"grace_period"`
}

// OAuthConfig represents OAuth configuration
type OAuthConfig struct {
	Enabled     bool              `yaml:"enabled"`
	Providers   map[string]string `yaml:"providers"`
	RedirectURL string            `yaml:"redirect_url"`
	Scopes      []string          `yaml:"scopes"`
}

// LDAPConfig represents LDAP configuration
type LDAPConfig struct {
	Enabled      bool   `yaml:"enabled"`
	Server       string `yaml:"server"`
	Port         int    `yaml:"port"`
	BaseDN       string `yaml:"base_dn"`
	BindDN       string `yaml:"bind_dn"`
	BindPassword string `yaml:"bind_password"`
	UserFilter   string `yaml:"user_filter"`
	GroupFilter  string `yaml:"group_filter"`
}

// PasswordPolicyConfig represents password policy configuration
type PasswordPolicyConfig struct {
	MinLength        int           `yaml:"min_length"`
	RequireUppercase bool          `yaml:"require_uppercase"`
	RequireLowercase bool          `yaml:"require_lowercase"`
	RequireNumbers   bool          `yaml:"require_numbers"`
	RequireSymbols   bool          `yaml:"require_symbols"`
	MaxAge           time.Duration `yaml:"max_age"`
	HistorySize      int           `yaml:"history_size"`
}

// EncryptionConfig represents encryption configuration
type EncryptionConfig struct {
	Enabled     bool          `yaml:"enabled"`
	Algorithm   string        `yaml:"algorithm"`
	KeySize     int           `yaml:"key_size"`
	AtRest      bool          `yaml:"at_rest"`
	InTransit   bool          `yaml:"in_transit"`
	KeyRotation time.Duration `yaml:"key_rotation"`
	HSMEnabled  bool          `yaml:"hsm_enabled"`
	HSMConfig   HSMConfig     `yaml:"hsm_config"`
}

// HSMConfig represents Hardware Security Module configuration
type HSMConfig struct {
	Provider string `yaml:"provider"`
	Endpoint string `yaml:"endpoint"`
	KeyID    string `yaml:"key_id"`
}

// PrivacyConfig represents privacy configuration
type PrivacyConfig struct {
	Enabled           bool          `yaml:"enabled"`
	DataMinimization  bool          `yaml:"data_minimization"`
	Anonymization     bool          `yaml:"anonymization"`
	Pseudonymization  bool          `yaml:"pseudonymization"`
	DataRetention     time.Duration `yaml:"data_retention"`
	ConsentManagement bool          `yaml:"consent_management"`
	RightToErasure    bool          `yaml:"right_to_erasure"`
	DataPortability   bool          `yaml:"data_portability"`
	PIIDetection      bool          `yaml:"pii_detection"`
}

// ThreatDetectionConfig represents threat detection configuration
type ThreatDetectionConfig struct {
	Enabled            bool               `yaml:"enabled"`
	RealTime           bool               `yaml:"real_time"`
	MachineLearning    bool               `yaml:"machine_learning"`
	BehavioralAnalysis bool               `yaml:"behavioral_analysis"`
	NetworkMonitoring  bool               `yaml:"network_monitoring"`
	FileIntegrity      bool               `yaml:"file_integrity"`
	AlertThresholds    map[string]float64 `yaml:"alert_thresholds"`
	ResponseActions    []string           `yaml:"response_actions"`
}

// AuditConfig represents audit configuration
type AuditConfig struct {
	Enabled         bool          `yaml:"enabled"`
	LogLevel        string        `yaml:"log_level"`
	RetentionPeriod time.Duration `yaml:"retention_period"`
	Encryption      bool          `yaml:"encryption"`
	Integrity       bool          `yaml:"integrity"`
	RemoteLogging   bool          `yaml:"remote_logging"`
	SIEMIntegration bool          `yaml:"siem_integration"`
}

// AccessControlConfig represents access control configuration
type AccessControlConfig struct {
	Enabled        bool                `yaml:"enabled"`
	Model          string              `yaml:"model"` // RBAC, ABAC, etc.
	DefaultPolicy  string              `yaml:"default_policy"`
	Roles          map[string][]string `yaml:"roles"`
	Permissions    map[string][]string `yaml:"permissions"`
	SessionTimeout time.Duration       `yaml:"session_timeout"`
	MaxSessions    int                 `yaml:"max_sessions"`
}

// ComplianceConfig represents compliance configuration
type ComplianceConfig struct {
	Enabled    bool     `yaml:"enabled"`
	Standards  []string `yaml:"standards"` // GDPR, HIPAA, SOC2, etc.
	Reporting  bool     `yaml:"reporting"`
	Monitoring bool     `yaml:"monitoring"`
	Automation bool     `yaml:"automation"`
}

// IncidentResponseConfig represents incident response configuration
type IncidentResponseConfig struct {
	Enabled         bool              `yaml:"enabled"`
	AutoResponse    bool              `yaml:"auto_response"`
	NotificationURL string            `yaml:"notification_url"`
	EscalationRules []EscalationRule  `yaml:"escalation_rules"`
	Playbooks       map[string]string `yaml:"playbooks"`
}

// EscalationRule represents an escalation rule
type EscalationRule struct {
	Severity  string        `yaml:"severity"`
	TimeLimit time.Duration `yaml:"time_limit"`
	Contacts  []string      `yaml:"contacts"`
	Actions   []string      `yaml:"actions"`
}

// VulnerabilityConfig represents vulnerability scanning configuration
type VulnerabilityConfig struct {
	Enabled         bool          `yaml:"enabled"`
	ScanInterval    time.Duration `yaml:"scan_interval"`
	ScanTypes       []string      `yaml:"scan_types"`
	AutoRemediation bool          `yaml:"auto_remediation"`
	ReportingURL    string        `yaml:"reporting_url"`
}

// RateLimitConfig represents rate limiting configuration
type RateLimitConfig struct {
	Enabled           bool           `yaml:"enabled"`
	RequestsPerMinute int            `yaml:"requests_per_minute"`
	Burst             int            `yaml:"burst"`
	IPWhitelist       []string       `yaml:"ip_whitelist"`
	EndpointLimits    map[string]int `yaml:"endpoint_limits"`
}

// CORSConfig represents CORS configuration
type CORSConfig struct {
	Enabled        bool     `yaml:"enabled"`
	AllowedOrigins []string `yaml:"allowed_origins"`
	AllowedMethods []string `yaml:"allowed_methods"`
	AllowedHeaders []string `yaml:"allowed_headers"`
	MaxAge         int      `yaml:"max_age"`
}

// TLSConfig represents TLS configuration
type TLSConfig struct {
	Enabled      bool     `yaml:"enabled"`
	CertFile     string   `yaml:"cert_file"`
	KeyFile      string   `yaml:"key_file"`
	MinVersion   string   `yaml:"min_version"`
	CipherSuites []string `yaml:"cipher_suites"`
	HSTS         bool     `yaml:"hsts"`
}

// NewManager creates a new security manager
func NewManager(logger *logrus.Logger, config SecurityConfig) (*Manager, error) {
	tracer := otel.Tracer("security-manager")

	// Initialize components
	authManager, err := NewAuthManager(logger, config.Authentication)
	if err != nil {
		return nil, fmt.Errorf("failed to create auth manager: %w", err)
	}

	encryptionManager, err := NewEncryptionManager(logger, config.Encryption)
	if err != nil {
		return nil, fmt.Errorf("failed to create encryption manager: %w", err)
	}

	privacyManager, err := NewPrivacyManager(logger, config.Privacy)
	if err != nil {
		return nil, fmt.Errorf("failed to create privacy manager: %w", err)
	}

	threatDetector, err := NewThreatDetector(logger, config.ThreatDetection)
	if err != nil {
		return nil, fmt.Errorf("failed to create threat detector: %w", err)
	}

	auditLogger, err := NewAuditLogger(logger, config.Audit)
	if err != nil {
		return nil, fmt.Errorf("failed to create audit logger: %w", err)
	}

	accessController, err := NewAccessController(logger, config.AccessControl)
	if err != nil {
		return nil, fmt.Errorf("failed to create access controller: %w", err)
	}

	complianceManager, err := NewComplianceManager(logger, config.Compliance)
	if err != nil {
		return nil, fmt.Errorf("failed to create compliance manager: %w", err)
	}

	incidentResponder, err := NewIncidentResponder(logger, config.IncidentResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to create incident responder: %w", err)
	}

	vulnerabilityScanner, err := NewVulnerabilityScanner(logger, config.VulnerabilityScanning)
	if err != nil {
		return nil, fmt.Errorf("failed to create vulnerability scanner: %w", err)
	}

	return &Manager{
		logger:               logger,
		tracer:               tracer,
		config:               config,
		authManager:          authManager,
		encryptionManager:    encryptionManager,
		privacyManager:       privacyManager,
		threatDetector:       threatDetector,
		auditLogger:          auditLogger,
		accessController:     accessController,
		complianceManager:    complianceManager,
		incidentResponder:    incidentResponder,
		vulnerabilityScanner: vulnerabilityScanner,
		stopCh:               make(chan struct{}),
	}, nil
}

// Start initializes the security manager
func (m *Manager) Start(ctx context.Context) error {
	ctx, span := m.tracer.Start(ctx, "security.Manager.Start")
	defer span.End()

	m.mu.Lock()
	defer m.mu.Unlock()

	if m.running {
		return fmt.Errorf("security manager is already running")
	}

	if !m.config.Enabled {
		m.logger.Info("Security manager is disabled")
		return nil
	}

	m.logger.Info("Starting security manager")

	// Start components
	if err := m.authManager.Start(ctx); err != nil {
		return fmt.Errorf("failed to start auth manager: %w", err)
	}

	if err := m.encryptionManager.Start(ctx); err != nil {
		return fmt.Errorf("failed to start encryption manager: %w", err)
	}

	if err := m.privacyManager.Start(ctx); err != nil {
		return fmt.Errorf("failed to start privacy manager: %w", err)
	}

	if err := m.threatDetector.Start(ctx); err != nil {
		return fmt.Errorf("failed to start threat detector: %w", err)
	}

	if err := m.auditLogger.Start(ctx); err != nil {
		return fmt.Errorf("failed to start audit logger: %w", err)
	}

	if err := m.accessController.Start(ctx); err != nil {
		return fmt.Errorf("failed to start access controller: %w", err)
	}

	if err := m.complianceManager.Start(ctx); err != nil {
		return fmt.Errorf("failed to start compliance manager: %w", err)
	}

	if err := m.incidentResponder.Start(ctx); err != nil {
		return fmt.Errorf("failed to start incident responder: %w", err)
	}

	if err := m.vulnerabilityScanner.Start(ctx); err != nil {
		return fmt.Errorf("failed to start vulnerability scanner: %w", err)
	}

	// Start monitoring
	go m.monitorSecurity()

	m.running = true
	m.logger.Info("Security manager started successfully")

	return nil
}

// Stop shuts down the security manager
func (m *Manager) Stop(ctx context.Context) error {
	ctx, span := m.tracer.Start(ctx, "security.Manager.Stop")
	defer span.End()

	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.running {
		return nil
	}

	m.logger.Info("Stopping security manager")

	// Stop components in reverse order
	if err := m.vulnerabilityScanner.Stop(ctx); err != nil {
		m.logger.WithError(err).Error("Failed to stop vulnerability scanner")
	}

	if err := m.incidentResponder.Stop(ctx); err != nil {
		m.logger.WithError(err).Error("Failed to stop incident responder")
	}

	if err := m.complianceManager.Stop(ctx); err != nil {
		m.logger.WithError(err).Error("Failed to stop compliance manager")
	}

	if err := m.accessController.Stop(ctx); err != nil {
		m.logger.WithError(err).Error("Failed to stop access controller")
	}

	if err := m.auditLogger.Stop(ctx); err != nil {
		m.logger.WithError(err).Error("Failed to stop audit logger")
	}

	if err := m.threatDetector.Stop(ctx); err != nil {
		m.logger.WithError(err).Error("Failed to stop threat detector")
	}

	if err := m.privacyManager.Stop(ctx); err != nil {
		m.logger.WithError(err).Error("Failed to stop privacy manager")
	}

	if err := m.encryptionManager.Stop(ctx); err != nil {
		m.logger.WithError(err).Error("Failed to stop encryption manager")
	}

	if err := m.authManager.Stop(ctx); err != nil {
		m.logger.WithError(err).Error("Failed to stop auth manager")
	}

	close(m.stopCh)
	m.running = false
	m.logger.Info("Security manager stopped")

	return nil
}

// GetStatus returns the current security status
func (m *Manager) GetStatus(ctx context.Context) (*models.SecurityStatus, error) {
	ctx, span := m.tracer.Start(ctx, "security.Manager.GetStatus")
	defer span.End()

	m.mu.RLock()
	defer m.mu.RUnlock()

	if !m.config.Enabled {
		return &models.SecurityStatus{
			Enabled:   false,
			Running:   false,
			Timestamp: time.Now(),
		}, nil
	}

	// Get component statuses
	authStatus, _ := m.authManager.GetStatus(ctx)
	encryptionStatus, _ := m.encryptionManager.GetStatus(ctx)
	privacyStatus, _ := m.privacyManager.GetStatus(ctx)
	threatStatus, _ := m.threatDetector.GetStatus(ctx)
	auditStatus, _ := m.auditLogger.GetStatus(ctx)
	accessStatus, _ := m.accessController.GetStatus(ctx)
	complianceStatus, _ := m.complianceManager.GetStatus(ctx)
	incidentStatus, _ := m.incidentResponder.GetStatus(ctx)
	vulnerabilityStatus, _ := m.vulnerabilityScanner.GetStatus(ctx)

	return &models.SecurityStatus{
		Enabled:               m.config.Enabled,
		Running:               m.running,
		Authentication:        authStatus,
		Encryption:            encryptionStatus,
		Privacy:               privacyStatus,
		ThreatDetection:       threatStatus,
		Audit:                 auditStatus,
		AccessControl:         accessStatus,
		Compliance:            complianceStatus,
		IncidentResponse:      incidentStatus,
		VulnerabilityScanning: vulnerabilityStatus,
		Timestamp:             time.Now(),
	}, nil
}

// AnalyzeThreats performs threat analysis
func (m *Manager) AnalyzeThreats(ctx context.Context) ([]*models.ThreatAnalysis, error) {
	ctx, span := m.tracer.Start(ctx, "security.Manager.AnalyzeThreats")
	defer span.End()

	return m.threatDetector.AnalyzeThreats(ctx)
}

// GetAuditLogs returns audit logs
func (m *Manager) GetAuditLogs(ctx context.Context, filter models.AuditFilter) ([]*models.AuditLog, error) {
	ctx, span := m.tracer.Start(ctx, "security.Manager.GetAuditLogs")
	defer span.End()

	return m.auditLogger.GetLogs(ctx, filter)
}

// CheckAccess checks if a user has access to a resource
func (m *Manager) CheckAccess(ctx context.Context, userID, resource, action string) (bool, error) {
	ctx, span := m.tracer.Start(ctx, "security.Manager.CheckAccess")
	defer span.End()

	return m.accessController.CheckAccess(ctx, userID, resource, action)
}

// EncryptData encrypts data
func (m *Manager) EncryptData(ctx context.Context, data []byte) ([]byte, error) {
	ctx, span := m.tracer.Start(ctx, "security.Manager.EncryptData")
	defer span.End()

	return m.encryptionManager.Encrypt(ctx, data)
}

// DecryptData decrypts data
func (m *Manager) DecryptData(ctx context.Context, encryptedData []byte) ([]byte, error) {
	ctx, span := m.tracer.Start(ctx, "security.Manager.DecryptData")
	defer span.End()

	return m.encryptionManager.Decrypt(ctx, encryptedData)
}

// AnonymizeData anonymizes personal data
func (m *Manager) AnonymizeData(ctx context.Context, data interface{}) (interface{}, error) {
	ctx, span := m.tracer.Start(ctx, "security.Manager.AnonymizeData")
	defer span.End()

	return m.privacyManager.Anonymize(ctx, data)
}

// ValidateCompliance validates compliance with regulations
func (m *Manager) ValidateCompliance(ctx context.Context, standard string) (*models.ComplianceReport, error) {
	ctx, span := m.tracer.Start(ctx, "security.Manager.ValidateCompliance")
	defer span.End()

	return m.complianceManager.ValidateCompliance(ctx, standard)
}

// GetAuthManager returns the authentication manager
func (m *Manager) GetAuthManager() *AuthManager {
	return m.authManager
}

// GetEncryptionManager returns the encryption manager
func (m *Manager) GetEncryptionManager() *EncryptionManager {
	return m.encryptionManager
}

// GetPrivacyManager returns the privacy manager
func (m *Manager) GetPrivacyManager() *PrivacyManager {
	return m.privacyManager
}

// GetThreatDetector returns the threat detector
func (m *Manager) GetThreatDetector() *ThreatDetector {
	return m.threatDetector
}

// GetAuditLogger returns the audit logger
func (m *Manager) GetAuditLogger() *AuditLogger {
	return m.auditLogger
}

// GetAccessController returns the access controller
func (m *Manager) GetAccessController() *AccessController {
	return m.accessController
}

// GetComplianceManager returns the compliance manager
func (m *Manager) GetComplianceManager() *ComplianceManager {
	return m.complianceManager
}

// GetIncidentResponder returns the incident responder
func (m *Manager) GetIncidentResponder() *IncidentResponder {
	return m.incidentResponder
}

// GetVulnerabilityScanner returns the vulnerability scanner
func (m *Manager) GetVulnerabilityScanner() *VulnerabilityScanner {
	return m.vulnerabilityScanner
}

// monitorSecurity continuously monitors security status
func (m *Manager) monitorSecurity() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			ctx := context.Background()
			status, err := m.GetStatus(ctx)
			if err != nil {
				m.logger.WithError(err).Error("Failed to get security status")
				continue
			}

			// Log status for monitoring
			m.logger.WithFields(logrus.Fields{
				"enabled": status.Enabled,
				"running": status.Running,
			}).Debug("Security status")

			// Check for threats
			threats, err := m.AnalyzeThreats(ctx)
			if err != nil {
				m.logger.WithError(err).Error("Failed to analyze threats")
				continue
			}

			// Handle high-severity threats
			for _, threat := range threats {
				if threat.Severity == "high" || threat.Severity == "critical" {
					m.logger.WithFields(logrus.Fields{
						"threat_id": threat.ID,
						"severity":  threat.Severity,
						"type":      threat.Type,
					}).Warn("High-severity threat detected")

					// Trigger incident response
					if m.config.IncidentResponse.AutoResponse {
						if err := m.incidentResponder.HandleThreat(ctx, threat); err != nil {
							m.logger.WithError(err).Error("Failed to handle threat")
						}
					}
				}
			}

		case <-m.stopCh:
			m.logger.Debug("Security monitoring stopped")
			return
		}
	}
}

// GenerateSecureToken generates a cryptographically secure token
func GenerateSecureToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// HashPassword hashes a password using SHA-256
func HashPassword(password string) string {
	hash := sha256.Sum256([]byte(password))
	return hex.EncodeToString(hash[:])
}
