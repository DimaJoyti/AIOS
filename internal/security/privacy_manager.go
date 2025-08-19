package security

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/aios/aios/pkg/models"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

// PrivacyManager handles privacy protection and data anonymization
type PrivacyManager struct {
	logger      *logrus.Logger
	tracer      trace.Tracer
	config      PrivacyConfig
	piiPatterns []*regexp.Regexp
	mu          sync.RWMutex
	running     bool
	stopCh      chan struct{}
}

// NewPrivacyManager creates a new privacy manager
func NewPrivacyManager(logger *logrus.Logger, config PrivacyConfig) (*PrivacyManager, error) {
	tracer := otel.Tracer("privacy-manager")

	pm := &PrivacyManager{
		logger: logger,
		tracer: tracer,
		config: config,
		stopCh: make(chan struct{}),
	}

	if config.Enabled {
		pm.initializePIIPatterns()
	}

	return pm, nil
}

// Start initializes the privacy manager
func (pm *PrivacyManager) Start(ctx context.Context) error {
	ctx, span := pm.tracer.Start(ctx, "privacyManager.Start")
	defer span.End()

	pm.mu.Lock()
	defer pm.mu.Unlock()

	if pm.running {
		return fmt.Errorf("privacy manager is already running")
	}

	if !pm.config.Enabled {
		pm.logger.Info("Privacy manager is disabled")
		return nil
	}

	pm.logger.Info("Starting privacy manager")

	pm.running = true
	pm.logger.Info("Privacy manager started successfully")

	return nil
}

// Stop shuts down the privacy manager
func (pm *PrivacyManager) Stop(ctx context.Context) error {
	ctx, span := pm.tracer.Start(ctx, "privacyManager.Stop")
	defer span.End()

	pm.mu.Lock()
	defer pm.mu.Unlock()

	if !pm.running {
		return nil
	}

	pm.logger.Info("Stopping privacy manager")

	close(pm.stopCh)
	pm.running = false
	pm.logger.Info("Privacy manager stopped")

	return nil
}

// GetStatus returns the current privacy status
func (pm *PrivacyManager) GetStatus(ctx context.Context) (*models.PrivacyStatus, error) {
	ctx, span := pm.tracer.Start(ctx, "privacyManager.GetStatus")
	defer span.End()

	pm.mu.RLock()
	defer pm.mu.RUnlock()

	return &models.PrivacyStatus{
		Enabled:           pm.config.Enabled,
		DataMinimization:  pm.config.DataMinimization,
		Anonymization:     pm.config.Anonymization,
		ConsentManagement: pm.config.ConsentManagement,
		PIIDetected:       0, // TODO: Track actual PII detections
		DataRetention:     pm.config.DataRetention.String(),
		Timestamp:         time.Now(),
	}, nil
}

// Anonymize anonymizes personal data in the given input
func (pm *PrivacyManager) Anonymize(ctx context.Context, data interface{}) (interface{}, error) {
	ctx, span := pm.tracer.Start(ctx, "privacyManager.Anonymize")
	defer span.End()

	if !pm.config.Anonymization {
		return data, nil
	}

	switch v := data.(type) {
	case string:
		return pm.anonymizeString(v), nil
	case map[string]interface{}:
		return pm.anonymizeMap(v), nil
	case []interface{}:
		return pm.anonymizeSlice(v), nil
	default:
		return data, nil
	}
}

// DetectPII detects personally identifiable information in text
func (pm *PrivacyManager) DetectPII(ctx context.Context, text string) ([]string, error) {
	ctx, span := pm.tracer.Start(ctx, "privacyManager.DetectPII")
	defer span.End()

	if !pm.config.PIIDetection {
		return nil, nil
	}

	var detected []string

	for _, pattern := range pm.piiPatterns {
		matches := pattern.FindAllString(text, -1)
		detected = append(detected, matches...)
	}

	return detected, nil
}

// Helper methods

func (pm *PrivacyManager) initializePIIPatterns() {
	patterns := []string{
		// Email addresses
		`[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`,
		// Phone numbers (US format)
		`\b\d{3}-\d{3}-\d{4}\b`,
		`\b\(\d{3}\)\s*\d{3}-\d{4}\b`,
		// Social Security Numbers
		`\b\d{3}-\d{2}-\d{4}\b`,
		// Credit Card Numbers (basic pattern)
		`\b\d{4}[\s-]?\d{4}[\s-]?\d{4}[\s-]?\d{4}\b`,
		// IP Addresses
		`\b\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}\b`,
	}

	for _, pattern := range patterns {
		if regex, err := regexp.Compile(pattern); err == nil {
			pm.piiPatterns = append(pm.piiPatterns, regex)
		}
	}
}

func (pm *PrivacyManager) anonymizeString(text string) string {
	result := text

	// Anonymize email addresses
	emailRegex := regexp.MustCompile(`[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`)
	result = emailRegex.ReplaceAllString(result, "[EMAIL]")

	// Anonymize phone numbers
	phoneRegex := regexp.MustCompile(`\b\d{3}-\d{3}-\d{4}\b`)
	result = phoneRegex.ReplaceAllString(result, "[PHONE]")

	// Anonymize SSNs
	ssnRegex := regexp.MustCompile(`\b\d{3}-\d{2}-\d{4}\b`)
	result = ssnRegex.ReplaceAllString(result, "[SSN]")

	// Anonymize credit card numbers
	ccRegex := regexp.MustCompile(`\b\d{4}[\s-]?\d{4}[\s-]?\d{4}[\s-]?\d{4}\b`)
	result = ccRegex.ReplaceAllString(result, "[CREDIT_CARD]")

	return result
}

func (pm *PrivacyManager) anonymizeMap(data map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	for key, value := range data {
		// Check if key contains sensitive information
		lowerKey := strings.ToLower(key)
		if pm.isSensitiveKey(lowerKey) {
			result[key] = "[REDACTED]"
			continue
		}

		// Recursively anonymize values
		switch v := value.(type) {
		case string:
			result[key] = pm.anonymizeString(v)
		case map[string]interface{}:
			result[key] = pm.anonymizeMap(v)
		case []interface{}:
			result[key] = pm.anonymizeSlice(v)
		default:
			result[key] = value
		}
	}

	return result
}

func (pm *PrivacyManager) anonymizeSlice(data []interface{}) []interface{} {
	result := make([]interface{}, len(data))

	for i, item := range data {
		switch v := item.(type) {
		case string:
			result[i] = pm.anonymizeString(v)
		case map[string]interface{}:
			result[i] = pm.anonymizeMap(v)
		case []interface{}:
			result[i] = pm.anonymizeSlice(v)
		default:
			result[i] = item
		}
	}

	return result
}

func (pm *PrivacyManager) isSensitiveKey(key string) bool {
	sensitiveKeys := []string{
		"password", "passwd", "pwd",
		"secret", "token", "key",
		"ssn", "social_security",
		"credit_card", "creditcard", "cc",
		"phone", "telephone", "mobile",
		"email", "mail",
		"address", "addr",
		"birthday", "birthdate", "dob",
	}

	for _, sensitive := range sensitiveKeys {
		if strings.Contains(key, sensitive) {
			return true
		}
	}

	return false
}

// ThreatDetector handles threat detection and analysis
type ThreatDetector struct {
	logger  *logrus.Logger
	tracer  trace.Tracer
	config  ThreatDetectionConfig
	threats map[string]*models.ThreatAnalysis
	mu      sync.RWMutex
	running bool
	stopCh  chan struct{}
}

// NewThreatDetector creates a new threat detector
func NewThreatDetector(logger *logrus.Logger, config ThreatDetectionConfig) (*ThreatDetector, error) {
	tracer := otel.Tracer("threat-detector")

	return &ThreatDetector{
		logger:  logger,
		tracer:  tracer,
		config:  config,
		threats: make(map[string]*models.ThreatAnalysis),
		stopCh:  make(chan struct{}),
	}, nil
}

// Start initializes the threat detector
func (td *ThreatDetector) Start(ctx context.Context) error {
	ctx, span := td.tracer.Start(ctx, "threatDetector.Start")
	defer span.End()

	td.mu.Lock()
	defer td.mu.Unlock()

	if td.running {
		return fmt.Errorf("threat detector is already running")
	}

	if !td.config.Enabled {
		td.logger.Info("Threat detector is disabled")
		return nil
	}

	td.logger.Info("Starting threat detector")

	// Start real-time monitoring if enabled
	if td.config.RealTime {
		go td.monitorThreats()
	}

	td.running = true
	td.logger.Info("Threat detector started successfully")

	return nil
}

// Stop shuts down the threat detector
func (td *ThreatDetector) Stop(ctx context.Context) error {
	ctx, span := td.tracer.Start(ctx, "threatDetector.Stop")
	defer span.End()

	td.mu.Lock()
	defer td.mu.Unlock()

	if !td.running {
		return nil
	}

	td.logger.Info("Stopping threat detector")

	close(td.stopCh)
	td.running = false
	td.logger.Info("Threat detector stopped")

	return nil
}

// GetStatus returns the current threat detection status
func (td *ThreatDetector) GetStatus(ctx context.Context) (*models.ThreatDetectionStatus, error) {
	ctx, span := td.tracer.Start(ctx, "threatDetector.GetStatus")
	defer span.End()

	td.mu.RLock()
	defer td.mu.RUnlock()

	threatsDetected := len(td.threats)
	highSeverity := 0
	var lastThreat time.Time

	for _, threat := range td.threats {
		if threat.Severity == "high" || threat.Severity == "critical" {
			highSeverity++
		}
		if threat.DetectedAt.After(lastThreat) {
			lastThreat = threat.DetectedAt
		}
	}

	return &models.ThreatDetectionStatus{
		Enabled:         td.config.Enabled,
		RealTime:        td.config.RealTime,
		ThreatsDetected: threatsDetected,
		HighSeverity:    highSeverity,
		LastThreat:      lastThreat,
		Timestamp:       time.Now(),
	}, nil
}

// AnalyzeThreats performs threat analysis
func (td *ThreatDetector) AnalyzeThreats(ctx context.Context) ([]*models.ThreatAnalysis, error) {
	ctx, span := td.tracer.Start(ctx, "threatDetector.AnalyzeThreats")
	defer span.End()

	td.mu.RLock()
	defer td.mu.RUnlock()

	threats := make([]*models.ThreatAnalysis, 0, len(td.threats))
	for _, threat := range td.threats {
		threats = append(threats, threat)
	}

	return threats, nil
}

func (td *ThreatDetector) monitorThreats() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// TODO: Implement actual threat monitoring
			td.logger.Debug("Monitoring for threats")

		case <-td.stopCh:
			td.logger.Debug("Threat monitoring stopped")
			return
		}
	}
}
