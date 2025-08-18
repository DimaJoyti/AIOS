package connectors

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/aios/aios/pkg/dataintegration"
)

// RestAPIConnector implements the DataConnector interface for REST APIs
type RestAPIConnector struct {
	client    *http.Client
	baseURL   string
	headers   map[string]string
	connected bool
}

// NewRestAPIConnector creates a new REST API connector
func NewRestAPIConnector() *RestAPIConnector {
	return &RestAPIConnector{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		headers: make(map[string]string),
	}
}

// Connector Information

// GetType returns the connector type
func (rc *RestAPIConnector) GetType() string {
	return "restapi"
}

// GetName returns the connector name
func (rc *RestAPIConnector) GetName() string {
	return "REST API Connector"
}

// GetDescription returns the connector description
func (rc *RestAPIConnector) GetDescription() string {
	return "REST API connector for extracting data from RESTful web services with authentication and pagination support"
}

// GetVersion returns the connector version
func (rc *RestAPIConnector) GetVersion() string {
	return "1.0.0"
}

// GetSupportedOperations returns the list of supported operations
func (rc *RestAPIConnector) GetSupportedOperations() []string {
	return []string{
		"get_data",
		"post_data",
		"put_data",
		"delete_data",
		"list_endpoints",
		"get_schema",
		"paginated_fetch",
		"batch_request",
	}
}

// Configuration

// GetConfigSchema returns the configuration schema
func (rc *RestAPIConnector) GetConfigSchema() *dataintegration.ConnectorConfigSchema {
	return &dataintegration.ConnectorConfigSchema{
		Properties: map[string]*dataintegration.ConfigProperty{
			"base_url": {
				Type:        "string",
				Description: "Base URL of the REST API",
			},
			"api_key": {
				Type:        "string",
				Description: "API key for authentication",
				Sensitive:   true,
			},
			"auth_header": {
				Type:        "string",
				Description: "Authentication header name",
				Default:     "Authorization",
			},
			"auth_type": {
				Type:        "string",
				Description: "Authentication type (bearer, api_key, basic)",
				Default:     "bearer",
				Enum:        []string{"bearer", "api_key", "basic"},
			},
			"timeout": {
				Type:        "string",
				Description: "Request timeout (e.g., '30s', '1m')",
				Default:     "30s",
			},
			"rate_limit": {
				Type:        "integer",
				Description: "Requests per second limit",
				Default:     10,
			},
			"headers": {
				Type:        "object",
				Description: "Additional HTTP headers",
			},
			"pagination_type": {
				Type:        "string",
				Description: "Pagination type (offset, cursor, page)",
				Default:     "offset",
				Enum:        []string{"offset", "cursor", "page"},
			},
			"pagination_limit": {
				Type:        "integer",
				Description: "Default pagination limit",
				Default:     100,
			},
		},
		Required: []string{"base_url"},
	}
}

// ValidateConfig validates the configuration
func (rc *RestAPIConnector) ValidateConfig(config map[string]interface{}) error {
	if baseURL, exists := config["base_url"]; !exists || baseURL == "" {
		return fmt.Errorf("base_url is required")
	}

	// Validate URL format
	if baseURL, exists := config["base_url"]; exists {
		if _, err := url.Parse(baseURL.(string)); err != nil {
			return fmt.Errorf("invalid base_url format: %w", err)
		}
	}

	// Validate auth type
	if authType, exists := config["auth_type"]; exists {
		validTypes := []string{"bearer", "api_key", "basic"}
		authTypeStr := authType.(string)
		valid := false
		for _, validType := range validTypes {
			if authTypeStr == validType {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("invalid auth_type: %s", authTypeStr)
		}
	}

	return nil
}

// Connection Management

// Connect establishes a connection to the REST API
func (rc *RestAPIConnector) Connect(ctx context.Context, config map[string]interface{}) error {
	if err := rc.ValidateConfig(config); err != nil {
		return err
	}

	rc.baseURL = config["base_url"].(string)

	// Set up authentication
	if apiKey, exists := config["api_key"]; exists && apiKey != "" {
		authType := "bearer"
		if at, exists := config["auth_type"]; exists {
			authType = at.(string)
		}

		authHeader := "Authorization"
		if ah, exists := config["auth_header"]; exists {
			authHeader = ah.(string)
		}

		switch authType {
		case "bearer":
			rc.headers[authHeader] = fmt.Sprintf("Bearer %s", apiKey)
		case "api_key":
			rc.headers[authHeader] = apiKey.(string)
		case "basic":
			// For basic auth, apiKey should be base64 encoded username:password
			rc.headers[authHeader] = fmt.Sprintf("Basic %s", apiKey)
		}
	}

	// Set up additional headers
	if headers, exists := config["headers"]; exists {
		if headerMap, ok := headers.(map[string]interface{}); ok {
			for key, value := range headerMap {
				rc.headers[key] = value.(string)
			}
		}
	}

	// Set default content type
	if _, exists := rc.headers["Content-Type"]; !exists {
		rc.headers["Content-Type"] = "application/json"
	}

	// Set timeout
	if timeout, exists := config["timeout"]; exists {
		if timeoutStr, ok := timeout.(string); ok {
			if d, err := time.ParseDuration(timeoutStr); err == nil {
				rc.client.Timeout = d
			}
		}
	}

	// Test the connection
	if err := rc.TestConnection(ctx); err != nil {
		return fmt.Errorf("failed to connect to REST API: %w", err)
	}

	rc.connected = true
	return nil
}

// Disconnect closes the connection
func (rc *RestAPIConnector) Disconnect(ctx context.Context) error {
	rc.connected = false
	rc.headers = make(map[string]string)
	return nil
}

// IsConnected returns whether the connector is connected
func (rc *RestAPIConnector) IsConnected() bool {
	return rc.connected
}

// TestConnection tests the connection by making a simple request
func (rc *RestAPIConnector) TestConnection(ctx context.Context) error {
	// Try to make a HEAD request to the base URL
	req, err := http.NewRequestWithContext(ctx, "HEAD", rc.baseURL, nil)
	if err != nil {
		return err
	}

	// Add headers
	for key, value := range rc.headers {
		req.Header.Set(key, value)
	}

	resp, err := rc.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Accept any 2xx or 3xx status code as success
	if resp.StatusCode >= 400 {
		return fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	return nil
}

// Data Operations

// ExtractData performs data extraction from the REST API
func (rc *RestAPIConnector) ExtractData(ctx context.Context, params *dataintegration.ExtractionParams) (*dataintegration.DataExtraction, error) {
	start := time.Now()

	if !rc.connected {
		return nil, fmt.Errorf("connector not connected")
	}

	// Build request URL
	endpoint := "/"
	if params.Query != "" {
		endpoint = params.Query
	}

	requestURL := rc.baseURL + endpoint

	// Add query parameters
	if len(params.Filters) > 0 || params.Limit > 0 || params.Offset > 0 {
		u, err := url.Parse(requestURL)
		if err != nil {
			return nil, fmt.Errorf("invalid URL: %w", err)
		}

		q := u.Query()

		// Add filters as query parameters
		for key, value := range params.Filters {
			q.Set(key, fmt.Sprintf("%v", value))
		}

		// Add pagination parameters
		if params.Limit > 0 {
			q.Set("limit", fmt.Sprintf("%d", params.Limit))
		}
		if params.Offset > 0 {
			q.Set("offset", fmt.Sprintf("%d", params.Offset))
		}

		// Add sorting
		if params.SortBy != "" {
			q.Set("sort", params.SortBy)
			if params.SortOrder != "" {
				q.Set("order", params.SortOrder)
			}
		}

		u.RawQuery = q.Encode()
		requestURL = u.String()
	}

	// Make the request
	req, err := http.NewRequestWithContext(ctx, "GET", requestURL, nil)
	if err != nil {
		return nil, err
	}

	// Add headers
	for key, value := range rc.headers {
		req.Header.Set(key, value)
	}

	resp, err := rc.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Parse JSON response
	var responseData interface{}
	if err := json.Unmarshal(body, &responseData); err != nil {
		return nil, fmt.Errorf("failed to parse JSON response: %w", err)
	}

	// Convert response to data records
	records := rc.convertToDataRecords(responseData, endpoint)

	return &dataintegration.DataExtraction{
		Records:     records,
		TotalCount:  int64(len(records)),
		HasMore:     false, // TODO: Implement pagination detection
		ExtractedAt: time.Now(),
		Duration:    time.Since(start),
		Metadata: map[string]interface{}{
			"endpoint":      endpoint,
			"status_code":   resp.StatusCode,
			"content_type":  resp.Header.Get("Content-Type"),
			"response_size": len(body),
		},
	}, nil
}

// StreamData streams data (not typically supported for REST APIs)
func (rc *RestAPIConnector) StreamData(ctx context.Context, params *dataintegration.StreamParams) (<-chan *dataintegration.DataRecord, error) {
	return nil, fmt.Errorf("streaming not supported for REST API connector")
}

// Health and Monitoring

// GetHealth returns the connector health status
func (rc *RestAPIConnector) GetHealth() *dataintegration.ConnectorHealth {
	status := dataintegration.HealthStatusHealthy
	message := "REST API connector is healthy"

	if !rc.connected {
		status = dataintegration.HealthStatusUnhealthy
		message = "REST API connector not connected"
	}

	return &dataintegration.ConnectorHealth{
		Status:       status,
		Message:      message,
		Connected:    rc.connected,
		LastActivity: time.Now(),
	}
}

// GetMetrics returns connector metrics
func (rc *RestAPIConnector) GetMetrics() *dataintegration.ConnectorMetrics {
	return &dataintegration.ConnectorMetrics{
		OperationCount:   make(map[string]int64),
		AverageLatency:   make(map[string]time.Duration),
		ErrorCount:       make(map[string]int64),
		LastOperation:    time.Now(),
		ConnectionUptime: time.Hour, // Placeholder
	}
}

// Helper methods

// convertToDataRecords converts API response data to data records
func (rc *RestAPIConnector) convertToDataRecords(responseData interface{}, endpoint string) []*dataintegration.DataRecord {
	var records []*dataintegration.DataRecord

	switch data := responseData.(type) {
	case []interface{}:
		// Array of objects
		for i, item := range data {
			if itemMap, ok := item.(map[string]interface{}); ok {
				record := &dataintegration.DataRecord{
					ID:       fmt.Sprintf("%s_%d", endpoint, i),
					SourceID: "restapi",
					Data:     itemMap,
					Metadata: map[string]interface{}{
						"endpoint": endpoint,
						"index":    i,
					},
					Timestamp: time.Now(),
				}
				records = append(records, record)
			}
		}
	case map[string]interface{}:
		// Single object
		record := &dataintegration.DataRecord{
			ID:       fmt.Sprintf("%s_single", endpoint),
			SourceID: "restapi",
			Data:     data,
			Metadata: map[string]interface{}{
				"endpoint": endpoint,
				"type":     "single_object",
			},
			Timestamp: time.Now(),
		}
		records = append(records, record)
	default:
		// Primitive value or unknown structure
		record := &dataintegration.DataRecord{
			ID:       fmt.Sprintf("%s_raw", endpoint),
			SourceID: "restapi",
			Data: map[string]interface{}{
				"value": data,
			},
			Metadata: map[string]interface{}{
				"endpoint": endpoint,
				"type":     "raw_value",
			},
			Timestamp: time.Now(),
		}
		records = append(records, record)
	}

	return records
}

// makeRequest makes an HTTP request with proper headers and error handling
func (rc *RestAPIConnector) makeRequest(ctx context.Context, method, endpoint string, body interface{}) (*http.Response, error) {
	var requestBody io.Reader

	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		requestBody = bytes.NewBuffer(jsonBody)
	}

	url := rc.baseURL + endpoint
	req, err := http.NewRequestWithContext(ctx, method, url, requestBody)
	if err != nil {
		return nil, err
	}

	// Add headers
	for key, value := range rc.headers {
		req.Header.Set(key, value)
	}

	return rc.client.Do(req)
}
