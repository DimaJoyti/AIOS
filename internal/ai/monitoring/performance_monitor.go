package monitoring

import (
	"context"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

// PerformanceMonitor tracks AI service performance metrics
type PerformanceMonitor struct {
	logger *logrus.Logger
	tracer trace.Tracer
	meter  metric.Meter

	// Metrics
	requestCounter    metric.Int64Counter
	responseTime      metric.Float64Histogram
	tokenCounter      metric.Int64Counter
	errorCounter      metric.Int64Counter
	cacheHitCounter   metric.Int64Counter
	cacheMissCounter  metric.Int64Counter
	modelUsageCounter metric.Int64Counter

	// Internal tracking
	mu                sync.RWMutex
	serviceMetrics    map[string]*ServiceMetrics
	modelMetrics      map[string]*ModelMetrics
	performanceAlerts []PerformanceAlert
}

// ServiceMetrics tracks metrics for a specific AI service
type ServiceMetrics struct {
	ServiceName      string                 `json:"service_name"`
	RequestCount     int64                  `json:"request_count"`
	SuccessCount     int64                  `json:"success_count"`
	ErrorCount       int64                  `json:"error_count"`
	TotalLatency     time.Duration          `json:"total_latency"`
	AverageLatency   time.Duration          `json:"average_latency"`
	MinLatency       time.Duration          `json:"min_latency"`
	MaxLatency       time.Duration          `json:"max_latency"`
	ThroughputPerSec float64                `json:"throughput_per_sec"`
	CacheHitRate     float64                `json:"cache_hit_rate"`
	CacheHits        int64                  `json:"cache_hits"`
	CacheMisses      int64                  `json:"cache_misses"`
	LastUpdated      time.Time              `json:"last_updated"`
	Metadata         map[string]any `json:"metadata,omitempty"`
}

// ModelMetrics tracks metrics for a specific AI model
type ModelMetrics struct {
	ModelID          string                 `json:"model_id"`
	RequestCount     int64                  `json:"request_count"`
	TokensGenerated  int64                  `json:"tokens_generated"`
	TokensPerSecond  float64                `json:"tokens_per_second"`
	AverageLatency   time.Duration          `json:"average_latency"`
	MemoryUsage      int64                  `json:"memory_usage"`
	GPUUtilization   float64                `json:"gpu_utilization"`
	LoadTime         time.Duration          `json:"load_time"`
	LastUsed         time.Time              `json:"last_used"`
	Metadata         map[string]any `json:"metadata,omitempty"`
}

// PerformanceAlert represents a performance alert
type PerformanceAlert struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Severity    string                 `json:"severity"`
	Service     string                 `json:"service"`
	Model       string                 `json:"model,omitempty"`
	Message     string                 `json:"message"`
	Threshold   float64                `json:"threshold"`
	ActualValue float64                `json:"actual_value"`
	Timestamp   time.Time              `json:"timestamp"`
	Metadata    map[string]any `json:"metadata,omitempty"`
}

// NewPerformanceMonitor creates a new performance monitor
func NewPerformanceMonitor(logger *logrus.Logger) (*PerformanceMonitor, error) {
	tracer := otel.Tracer("ai.performance_monitor")
	meter := otel.Meter("ai.performance_monitor")

	// Initialize metrics
	requestCounter, err := meter.Int64Counter(
		"ai_requests_total",
		metric.WithDescription("Total number of AI service requests"),
	)
	if err != nil {
		return nil, err
	}

	responseTime, err := meter.Float64Histogram(
		"ai_response_time_seconds",
		metric.WithDescription("AI service response time in seconds"),
		metric.WithUnit("s"),
	)
	if err != nil {
		return nil, err
	}

	tokenCounter, err := meter.Int64Counter(
		"ai_tokens_total",
		metric.WithDescription("Total number of tokens processed"),
	)
	if err != nil {
		return nil, err
	}

	errorCounter, err := meter.Int64Counter(
		"ai_errors_total",
		metric.WithDescription("Total number of AI service errors"),
	)
	if err != nil {
		return nil, err
	}

	cacheHitCounter, err := meter.Int64Counter(
		"ai_cache_hits_total",
		metric.WithDescription("Total number of cache hits"),
	)
	if err != nil {
		return nil, err
	}

	cacheMissCounter, err := meter.Int64Counter(
		"ai_cache_misses_total",
		metric.WithDescription("Total number of cache misses"),
	)
	if err != nil {
		return nil, err
	}

	modelUsageCounter, err := meter.Int64Counter(
		"ai_model_usage_total",
		metric.WithDescription("Total number of model usage requests"),
	)
	if err != nil {
		return nil, err
	}

	return &PerformanceMonitor{
		logger:            logger,
		tracer:            tracer,
		meter:             meter,
		requestCounter:    requestCounter,
		responseTime:      responseTime,
		tokenCounter:      tokenCounter,
		errorCounter:      errorCounter,
		cacheHitCounter:   cacheHitCounter,
		cacheMissCounter:  cacheMissCounter,
		modelUsageCounter: modelUsageCounter,
		serviceMetrics:    make(map[string]*ServiceMetrics),
		modelMetrics:      make(map[string]*ModelMetrics),
		performanceAlerts: make([]PerformanceAlert, 0),
	}, nil
}

// RecordRequest records a service request
func (pm *PerformanceMonitor) RecordRequest(ctx context.Context, serviceName, modelID string, latency time.Duration, tokens int64, success bool, cached bool) {
	ctx, span := pm.tracer.Start(ctx, "performance_monitor.RecordRequest")
	defer span.End()

	// Record OpenTelemetry metrics
	labels := []attribute.KeyValue{
		attribute.String("service", serviceName),
		attribute.String("model", modelID),
		attribute.Bool("success", success),
		attribute.Bool("cached", cached),
	}

	pm.requestCounter.Add(ctx, 1, metric.WithAttributes(labels...))
	pm.responseTime.Record(ctx, latency.Seconds(), metric.WithAttributes(labels...))

	if tokens > 0 {
		pm.tokenCounter.Add(ctx, tokens, metric.WithAttributes(labels...))
	}

	if !success {
		pm.errorCounter.Add(ctx, 1, metric.WithAttributes(labels...))
	}

	if cached {
		pm.cacheHitCounter.Add(ctx, 1, metric.WithAttributes(labels...))
	} else {
		pm.cacheMissCounter.Add(ctx, 1, metric.WithAttributes(labels...))
	}

	if modelID != "" {
		pm.modelUsageCounter.Add(ctx, 1, metric.WithAttributes(
			attribute.String("model", modelID),
		))
	}

	// Update internal metrics
	pm.updateServiceMetrics(serviceName, latency, tokens, success, cached)
	if modelID != "" {
		pm.updateModelMetrics(modelID, latency, tokens)
	}

	// Check for performance alerts
	pm.checkPerformanceAlerts(serviceName, modelID, latency, success)
}

// GetServiceMetrics returns metrics for a specific service
func (pm *PerformanceMonitor) GetServiceMetrics(serviceName string) *ServiceMetrics {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	if metrics, exists := pm.serviceMetrics[serviceName]; exists {
		// Return a copy
		metricsCopy := *metrics
		return &metricsCopy
	}
	return nil
}

// GetModelMetrics returns metrics for a specific model
func (pm *PerformanceMonitor) GetModelMetrics(modelID string) *ModelMetrics {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	if metrics, exists := pm.modelMetrics[modelID]; exists {
		// Return a copy
		metricsCopy := *metrics
		return &metricsCopy
	}
	return nil
}

// GetAllServiceMetrics returns metrics for all services
func (pm *PerformanceMonitor) GetAllServiceMetrics() map[string]*ServiceMetrics {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	result := make(map[string]*ServiceMetrics)
	for name, metrics := range pm.serviceMetrics {
		metricsCopy := *metrics
		result[name] = &metricsCopy
	}
	return result
}

// GetAllModelMetrics returns metrics for all models
func (pm *PerformanceMonitor) GetAllModelMetrics() map[string]*ModelMetrics {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	result := make(map[string]*ModelMetrics)
	for id, metrics := range pm.modelMetrics {
		metricsCopy := *metrics
		result[id] = &metricsCopy
	}
	return result
}

// GetPerformanceAlerts returns current performance alerts
func (pm *PerformanceMonitor) GetPerformanceAlerts() []PerformanceAlert {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	// Return a copy of alerts
	alerts := make([]PerformanceAlert, len(pm.performanceAlerts))
	copy(alerts, pm.performanceAlerts)
	return alerts
}

// Internal methods

func (pm *PerformanceMonitor) updateServiceMetrics(serviceName string, latency time.Duration, _ int64, success bool, cached bool) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	metrics, exists := pm.serviceMetrics[serviceName]
	if !exists {
		metrics = &ServiceMetrics{
			ServiceName: serviceName,
			MinLatency:  latency,
			MaxLatency:  latency,
			Metadata:    make(map[string]any),
		}
		pm.serviceMetrics[serviceName] = metrics
	}

	metrics.RequestCount++
	if success {
		metrics.SuccessCount++
	} else {
		metrics.ErrorCount++
	}

	metrics.TotalLatency += latency
	metrics.AverageLatency = metrics.TotalLatency / time.Duration(metrics.RequestCount)

	if latency < metrics.MinLatency {
		metrics.MinLatency = latency
	}
	if latency > metrics.MaxLatency {
		metrics.MaxLatency = latency
	}

	if cached {
		metrics.CacheHits++
	} else {
		metrics.CacheMisses++
	}

	totalCacheRequests := metrics.CacheHits + metrics.CacheMisses
	if totalCacheRequests > 0 {
		metrics.CacheHitRate = float64(metrics.CacheHits) / float64(totalCacheRequests)
	}

	metrics.LastUpdated = time.Now()

	// Calculate throughput (requests per second over last minute)
	// This is a simplified calculation
	metrics.ThroughputPerSec = float64(metrics.RequestCount) / time.Since(metrics.LastUpdated.Add(-time.Minute)).Seconds()
}

func (pm *PerformanceMonitor) updateModelMetrics(modelID string, latency time.Duration, tokens int64) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	metrics, exists := pm.modelMetrics[modelID]
	if !exists {
		metrics = &ModelMetrics{
			ModelID:  modelID,
			Metadata: make(map[string]any),
		}
		pm.modelMetrics[modelID] = metrics
	}

	metrics.RequestCount++
	metrics.TokensGenerated += tokens
	metrics.LastUsed = time.Now()

	// Update average latency
	if metrics.RequestCount == 1 {
		metrics.AverageLatency = latency
	} else {
		metrics.AverageLatency = (metrics.AverageLatency*time.Duration(metrics.RequestCount-1) + latency) / time.Duration(metrics.RequestCount)
	}

	// Calculate tokens per second
	if latency > 0 {
		metrics.TokensPerSecond = float64(tokens) / latency.Seconds()
	}
}

func (pm *PerformanceMonitor) checkPerformanceAlerts(serviceName, modelID string, latency time.Duration, success bool) {
	// Check for high latency alerts
	if latency > 5*time.Second {
		alert := PerformanceAlert{
			ID:          generateAlertID(),
			Type:        "high_latency",
			Severity:    "warning",
			Service:     serviceName,
			Model:       modelID,
			Message:     "High response latency detected",
			Threshold:   5.0,
			ActualValue: latency.Seconds(),
			Timestamp:   time.Now(),
		}
		pm.addAlert(alert)
	}

	// Check for error rate alerts
	if !success {
		serviceMetrics := pm.serviceMetrics[serviceName]
		if serviceMetrics != nil && serviceMetrics.RequestCount > 10 {
			errorRate := float64(serviceMetrics.ErrorCount) / float64(serviceMetrics.RequestCount)
			if errorRate > 0.1 { // 10% error rate
				alert := PerformanceAlert{
					ID:          generateAlertID(),
					Type:        "high_error_rate",
					Severity:    "critical",
					Service:     serviceName,
					Model:       modelID,
					Message:     "High error rate detected",
					Threshold:   0.1,
					ActualValue: errorRate,
					Timestamp:   time.Now(),
				}
				pm.addAlert(alert)
			}
		}
	}
}

func (pm *PerformanceMonitor) addAlert(alert PerformanceAlert) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pm.performanceAlerts = append(pm.performanceAlerts, alert)

	// Keep only last 100 alerts
	if len(pm.performanceAlerts) > 100 {
		pm.performanceAlerts = pm.performanceAlerts[len(pm.performanceAlerts)-100:]
	}

	pm.logger.WithFields(logrus.Fields{
		"alert_id":   alert.ID,
		"type":       alert.Type,
		"severity":   alert.Severity,
		"service":    alert.Service,
		"model":      alert.Model,
		"threshold":  alert.Threshold,
		"actual":     alert.ActualValue,
	}).Warn("Performance alert triggered")
}

func generateAlertID() string {
	return time.Now().Format("20060102150405") + "_" + randomString(6)
}

func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(result)
}
