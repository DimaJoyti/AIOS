package adapters

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

// PythonServiceConfig holds configuration for Python service adapter
type PythonServiceConfig struct {
	ServiceName string
	BaseURL     string
	Timeout     time.Duration
	RetryCount  int
	RetryDelay  time.Duration
}

// PythonServiceAdapter provides a Go interface to Python services
type PythonServiceAdapter struct {
	config     *PythonServiceConfig
	httpClient *http.Client
	logger     *logrus.Logger
	tracer     trace.Tracer
}

// ServiceResponse represents a generic response from Python services
type ServiceResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Message string      `json:"message,omitempty"`
}

// NewPythonServiceAdapter creates a new Python service adapter
func NewPythonServiceAdapter(config *PythonServiceConfig, logger *logrus.Logger) (*PythonServiceAdapter, error) {
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}
	if config.RetryCount == 0 {
		config.RetryCount = 3
	}
	if config.RetryDelay == 0 {
		config.RetryDelay = 1 * time.Second
	}

	httpClient := &http.Client{
		Timeout: config.Timeout,
	}

	return &PythonServiceAdapter{
		config:     config,
		httpClient: httpClient,
		logger:     logger,
		tracer:     otel.Tracer("adapters.python"),
	}, nil
}

// Start initializes the Python service adapter
func (a *PythonServiceAdapter) Start(ctx context.Context) error {
	a.logger.WithField("service", a.config.ServiceName).Info("Starting Python service adapter")
	
	// Wait for service to be ready
	return a.waitForService(ctx)
}

// Stop shuts down the Python service adapter
func (a *PythonServiceAdapter) Stop(ctx context.Context) error {
	a.logger.WithField("service", a.config.ServiceName).Info("Stopping Python service adapter")
	return nil
}

// Get performs a GET request to the Python service
func (a *PythonServiceAdapter) Get(ctx context.Context, endpoint string) (*ServiceResponse, error) {
	return a.makeRequest(ctx, "GET", endpoint, nil)
}

// Post performs a POST request to the Python service
func (a *PythonServiceAdapter) Post(ctx context.Context, endpoint string, data interface{}) (*ServiceResponse, error) {
	return a.makeRequest(ctx, "POST", endpoint, data)
}

// Put performs a PUT request to the Python service
func (a *PythonServiceAdapter) Put(ctx context.Context, endpoint string, data interface{}) (*ServiceResponse, error) {
	return a.makeRequest(ctx, "PUT", endpoint, data)
}

// Delete performs a DELETE request to the Python service
func (a *PythonServiceAdapter) Delete(ctx context.Context, endpoint string) (*ServiceResponse, error) {
	return a.makeRequest(ctx, "DELETE", endpoint, nil)
}

// makeRequest performs an HTTP request with retries and tracing
func (a *PythonServiceAdapter) makeRequest(ctx context.Context, method, endpoint string, data interface{}) (*ServiceResponse, error) {
	ctx, span := a.tracer.Start(ctx, fmt.Sprintf("python_service.%s", method))
	defer span.End()

	url := fmt.Sprintf("%s%s", a.config.BaseURL, endpoint)
	
	var body io.Reader
	if data != nil {
		jsonData, err := json.Marshal(data)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request data: %w", err)
		}
		body = bytes.NewReader(jsonData)
	}

	var lastErr error
	for attempt := 0; attempt < a.config.RetryCount; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(a.config.RetryDelay):
			}
		}

		req, err := http.NewRequestWithContext(ctx, method, url, body)
		if err != nil {
			lastErr = fmt.Errorf("failed to create request: %w", err)
			continue
		}

		if data != nil {
			req.Header.Set("Content-Type", "application/json")
		}

		resp, err := a.httpClient.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("request failed: %w", err)
			a.logger.WithError(err).WithField("attempt", attempt+1).Warn("Request attempt failed")
			continue
		}

		defer resp.Body.Close()

		responseBody, err := io.ReadAll(resp.Body)
		if err != nil {
			lastErr = fmt.Errorf("failed to read response body: %w", err)
			continue
		}

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			var serviceResp ServiceResponse
			if err := json.Unmarshal(responseBody, &serviceResp); err != nil {
				// If JSON parsing fails, treat as raw response
				serviceResp = ServiceResponse{
					Success: true,
					Data:    string(responseBody),
				}
			}
			return &serviceResp, nil
		}

		lastErr = fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(responseBody))
		a.logger.WithField("status", resp.StatusCode).WithField("attempt", attempt+1).Warn("Request returned error status")
	}

	return nil, fmt.Errorf("all retry attempts failed, last error: %w", lastErr)
}

// waitForService waits for the Python service to become available
func (a *PythonServiceAdapter) waitForService(ctx context.Context) error {
	healthEndpoint := "/health"
	maxWait := 60 * time.Second
	checkInterval := 2 * time.Second

	ctx, cancel := context.WithTimeout(ctx, maxWait)
	defer cancel()

	ticker := time.NewTicker(checkInterval)
	defer ticker.Stop()

	a.logger.WithField("service", a.config.ServiceName).Info("Waiting for Python service to become ready")

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("timeout waiting for service %s to become ready", a.config.ServiceName)
		case <-ticker.C:
			resp, err := a.Get(ctx, healthEndpoint)
			if err == nil && resp.Success {
				a.logger.WithField("service", a.config.ServiceName).Info("Python service is ready")
				return nil
			}
			a.logger.WithField("service", a.config.ServiceName).Debug("Service not ready yet, retrying...")
		}
	}
}

// IsHealthy checks if the Python service is healthy
func (a *PythonServiceAdapter) IsHealthy(ctx context.Context) bool {
	resp, err := a.Get(ctx, "/health")
	return err == nil && resp.Success
}
