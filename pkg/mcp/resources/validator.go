package resources

import (
	"fmt"
	"net/url"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/aios/aios/pkg/mcp/protocol"
	"github.com/sirupsen/logrus"
)

// ValidationRule represents a validation rule
type ValidationRule struct {
	Name        string
	Description string
	Validator   func(uri string) error
}

// DefaultResourceValidator implements ResourceValidator
type DefaultResourceValidator struct {
	allowedSchemes map[string]bool
	rules          []ValidationRule
	logger         *logrus.Logger
}

// NewResourceValidator creates a new resource validator
func NewResourceValidator(allowedSchemes []string, logger *logrus.Logger) (ResourceValidator, error) {
	validator := &DefaultResourceValidator{
		allowedSchemes: make(map[string]bool),
		logger:         logger,
	}

	// Set default allowed schemes if none provided
	if len(allowedSchemes) == 0 {
		allowedSchemes = []string{"file", "http", "https", "data"}
	}

	for _, scheme := range allowedSchemes {
		validator.allowedSchemes[scheme] = true
	}

	// Add default validation rules
	validator.addDefaultRules()

	return validator, nil
}

// ValidateURI validates a complete URI
func (v *DefaultResourceValidator) ValidateURI(uri string) error {
	if uri == "" {
		return fmt.Errorf("URI cannot be empty")
	}

	// Parse URI
	parsedURI, err := url.Parse(uri)
	if err != nil {
		return fmt.Errorf("invalid URI format: %w", err)
	}

	// Validate scheme
	if err := v.ValidateScheme(parsedURI.Scheme); err != nil {
		return err
	}

	// Validate path
	if err := v.ValidatePath(parsedURI.Path); err != nil {
		return err
	}

	// Apply custom validation rules
	for _, rule := range v.rules {
		if err := rule.Validator(uri); err != nil {
			return fmt.Errorf("validation rule '%s' failed: %w", rule.Name, err)
		}
	}

	return nil
}

// ValidateScheme validates a URI scheme
func (v *DefaultResourceValidator) ValidateScheme(scheme string) error {
	if scheme == "" {
		return fmt.Errorf("URI scheme cannot be empty")
	}

	scheme = strings.ToLower(scheme)
	if !v.allowedSchemes[scheme] {
		return fmt.Errorf("scheme '%s' is not allowed", scheme)
	}

	return nil
}

// ValidatePath validates a URI path
func (v *DefaultResourceValidator) ValidatePath(path string) error {
	if path == "" {
		return nil // Empty path is allowed for some schemes
	}

	// Check for path traversal attempts
	if strings.Contains(path, "..") {
		return fmt.Errorf("path traversal not allowed: %s", path)
	}

	// Check for null bytes
	if strings.Contains(path, "\x00") {
		return fmt.Errorf("null bytes not allowed in path: %s", path)
	}

	// Check for control characters
	for _, char := range path {
		if char < 32 && char != 9 && char != 10 && char != 13 { // Allow tab, LF, CR
			return fmt.Errorf("control characters not allowed in path: %s", path)
		}
	}

	return nil
}

// IsAllowedURI checks if a URI is allowed
func (v *DefaultResourceValidator) IsAllowedURI(uri string) bool {
	return v.ValidateURI(uri) == nil
}

// GetAllowedSchemes returns the list of allowed schemes
func (v *DefaultResourceValidator) GetAllowedSchemes() []string {
	schemes := make([]string, 0, len(v.allowedSchemes))
	for scheme := range v.allowedSchemes {
		schemes = append(schemes, scheme)
	}
	return schemes
}

// AddAllowedScheme adds a scheme to the allowed list
func (v *DefaultResourceValidator) AddAllowedScheme(scheme string) {
	scheme = strings.ToLower(scheme)
	v.allowedSchemes[scheme] = true
	v.logger.WithField("scheme", scheme).Debug("Added allowed scheme")
}

// RemoveAllowedScheme removes a scheme from the allowed list
func (v *DefaultResourceValidator) RemoveAllowedScheme(scheme string) {
	scheme = strings.ToLower(scheme)
	delete(v.allowedSchemes, scheme)
	v.logger.WithField("scheme", scheme).Debug("Removed allowed scheme")
}

// addDefaultRules adds default validation rules
func (v *DefaultResourceValidator) addDefaultRules() {
	// File path validation
	v.rules = append(v.rules, ValidationRule{
		Name:        "file_path_security",
		Description: "Validates file paths for security",
		Validator: func(uri string) error {
			parsedURI, err := url.Parse(uri)
			if err != nil {
				return err
			}

			if parsedURI.Scheme == "file" {
				path := parsedURI.Path

				// Check for absolute path requirements
				if !filepath.IsAbs(path) && !strings.HasPrefix(path, "./") {
					return fmt.Errorf("file paths must be absolute or relative with ./ prefix")
				}

				// Check for dangerous file extensions
				ext := strings.ToLower(filepath.Ext(path))
				dangerousExts := []string{".exe", ".bat", ".cmd", ".com", ".scr", ".pif"}
				for _, dangerous := range dangerousExts {
					if ext == dangerous {
						return fmt.Errorf("file extension '%s' is not allowed", ext)
					}
				}
			}

			return nil
		},
	})

	// URL validation
	v.rules = append(v.rules, ValidationRule{
		Name:        "url_security",
		Description: "Validates URLs for security",
		Validator: func(uri string) error {
			parsedURI, err := url.Parse(uri)
			if err != nil {
				return err
			}

			if parsedURI.Scheme == "http" || parsedURI.Scheme == "https" {
				// Check for localhost/private IP restrictions
				host := parsedURI.Hostname()
				if host == "localhost" || host == "127.0.0.1" || strings.HasPrefix(host, "192.168.") || strings.HasPrefix(host, "10.") {
					return fmt.Errorf("access to private/local addresses is restricted")
				}

				// Check for suspicious query parameters
				query := parsedURI.RawQuery
				suspiciousPatterns := []string{
					"javascript:",
					"data:",
					"vbscript:",
					"<script",
					"</script>",
				}

				for _, pattern := range suspiciousPatterns {
					if strings.Contains(strings.ToLower(query), pattern) {
						return fmt.Errorf("suspicious content detected in query parameters")
					}
				}
			}

			return nil
		},
	})

	// Data URI validation
	v.rules = append(v.rules, ValidationRule{
		Name:        "data_uri_validation",
		Description: "Validates data URIs",
		Validator: func(uri string) error {
			if strings.HasPrefix(uri, "data:") {
				// Check data URI format
				dataURIRegex := regexp.MustCompile(`^data:([^;,]+)(;[^,]*)?,(.*)`)
				if !dataURIRegex.MatchString(uri) {
					return fmt.Errorf("invalid data URI format")
				}

				// Check for reasonable size limits (1MB)
				if len(uri) > 1024*1024 {
					return fmt.Errorf("data URI too large (max 1MB)")
				}

				// Extract media type
				matches := dataURIRegex.FindStringSubmatch(uri)
				if len(matches) > 1 {
					mediaType := matches[1]

					// Allow only safe media types
					allowedTypes := []string{
						"text/plain",
						"text/html",
						"text/css",
						"text/javascript",
						"application/json",
						"application/xml",
						"image/png",
						"image/jpeg",
						"image/gif",
						"image/svg+xml",
					}

					allowed := false
					for _, allowedType := range allowedTypes {
						if mediaType == allowedType {
							allowed = true
							break
						}
					}

					if !allowed {
						return fmt.Errorf("media type '%s' is not allowed in data URIs", mediaType)
					}
				}
			}

			return nil
		},
	})

	// General URI length validation
	v.rules = append(v.rules, ValidationRule{
		Name:        "uri_length",
		Description: "Validates URI length",
		Validator: func(uri string) error {
			// Most browsers support up to 2048 characters
			if len(uri) > 2048 {
				return fmt.Errorf("URI too long (max 2048 characters)")
			}
			return nil
		},
	})

	// Character encoding validation
	v.rules = append(v.rules, ValidationRule{
		Name:        "character_encoding",
		Description: "Validates character encoding",
		Validator: func(uri string) error {
			// Check for proper UTF-8 encoding
			if !isValidUTF8(uri) {
				return fmt.Errorf("URI contains invalid UTF-8 characters")
			}
			return nil
		},
	})
}

// isValidUTF8 checks if a string is valid UTF-8
func isValidUTF8(s string) bool {
	for _, r := range s {
		if r == '\uFFFD' { // Unicode replacement character
			return false
		}
	}
	return true
}

// AddCustomRule adds a custom validation rule
func (v *DefaultResourceValidator) AddCustomRule(rule ValidationRule) {
	v.rules = append(v.rules, rule)
	v.logger.WithField("rule_name", rule.Name).Debug("Added custom validation rule")
}

// RemoveRule removes a validation rule by name
func (v *DefaultResourceValidator) RemoveRule(name string) {
	for i, rule := range v.rules {
		if rule.Name == name {
			v.rules = append(v.rules[:i], v.rules[i+1:]...)
			v.logger.WithField("rule_name", name).Debug("Removed validation rule")
			return
		}
	}
}

// GetRules returns all validation rules
func (v *DefaultResourceValidator) GetRules() []ValidationRule {
	return v.rules
}

// ValidateMultiple validates multiple URIs
func (v *DefaultResourceValidator) ValidateMultiple(uris []string) map[string]error {
	results := make(map[string]error)
	for _, uri := range uris {
		results[uri] = v.ValidateURI(uri)
	}
	return results
}

// GetValidationRules returns current validation rules as a map
func (v *DefaultResourceValidator) GetValidationRules() map[string]interface{} {
	rules := make(map[string]interface{})
	for _, rule := range v.rules {
		rules[rule.Name] = map[string]interface{}{
			"description": rule.Description,
			"validator":   rule.Validator,
		}
	}
	return rules
}

// ValidateResource validates a resource
func (v *DefaultResourceValidator) ValidateResource(resource MCPResource) error {
	return v.ValidateURI(resource.GetURI())
}

// ValidateContent validates resource content
func (v *DefaultResourceValidator) ValidateContent(content []protocol.ResourceContent) error {
	// Basic validation - in a real implementation this would be more comprehensive
	for _, c := range content {
		if c.URI == "" {
			return fmt.Errorf("resource content missing URI")
		}
	}
	return nil
}

// SetValidationRules sets custom validation rules
func (v *DefaultResourceValidator) SetValidationRules(rules map[string]interface{}) error {
	// Basic implementation - convert map back to ValidationRule structs
	v.rules = nil
	for name, ruleData := range rules {
		if ruleMap, ok := ruleData.(map[string]interface{}); ok {
			rule := ValidationRule{
				Name: name,
			}
			if desc, ok := ruleMap["description"].(string); ok {
				rule.Description = desc
			}
			v.rules = append(v.rules, rule)
		}
	}
	return nil
}
