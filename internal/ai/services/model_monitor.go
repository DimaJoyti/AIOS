package services

import (
	"fmt"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

// ModelMonitor monitors AI model performance and health
type ModelMonitor struct {
	metrics   map[string]*MonitoringMetrics
	alerts    map[string]*AlertRule
	history   map[string][]*MetricSnapshot
	mutex     sync.RWMutex
	logger    *logrus.Logger
	tracer    trace.Tracer
	alertChan chan *Alert
}

// MonitoringMetrics represents comprehensive monitoring metrics for a model
type MonitoringMetrics struct {
	ModelID                string        `json:"model_id"`
	RequestsPerSecond      float64       `json:"requests_per_second"`
	AverageLatency         time.Duration `json:"average_latency"`
	P95Latency             time.Duration `json:"p95_latency"`
	P99Latency             time.Duration `json:"p99_latency"`
	ErrorRate              float64       `json:"error_rate"`
	SuccessRate            float64       `json:"success_rate"`
	ThroughputTokensPerSec float64       `json:"throughput_tokens_per_sec"`
	CostPerHour            float64       `json:"cost_per_hour"`
	MemoryUsage            float64       `json:"memory_usage"`
	CPUUsage               float64       `json:"cpu_usage"`
	QueueLength            int           `json:"queue_length"`
	ActiveConnections      int           `json:"active_connections"`
	LastHealthCheck        time.Time     `json:"last_health_check"`
	HealthStatus           string        `json:"health_status"` // healthy, degraded, unhealthy
	Uptime                 time.Duration `json:"uptime"`
	LastError              string        `json:"last_error"`
	ErrorCount             int64         `json:"error_count"`
	TotalRequests          int64         `json:"total_requests"`
	LastUpdate             time.Time     `json:"last_update"`
}

// MetricSnapshot represents a point-in-time snapshot of metrics
type MetricSnapshot struct {
	Timestamp time.Time          `json:"timestamp"`
	Metrics   *MonitoringMetrics `json:"metrics"`
}

// AlertRule defines conditions for triggering alerts
type AlertRule struct {
	ID          string                 `json:"id"`
	ModelID     string                 `json:"model_id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Metric      string                 `json:"metric"`
	Operator    string                 `json:"operator"` // gt, lt, eq, gte, lte
	Threshold   float64                `json:"threshold"`
	Duration    time.Duration          `json:"duration"`
	Severity    string                 `json:"severity"` // low, medium, high, critical
	Enabled     bool                   `json:"enabled"`
	Actions     []string               `json:"actions"`
	Metadata    map[string]interface{} `json:"metadata"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// Alert represents a triggered alert
type Alert struct {
	ID          string                 `json:"id"`
	RuleID      string                 `json:"rule_id"`
	ModelID     string                 `json:"model_id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Severity    string                 `json:"severity"`
	Status      string                 `json:"status"` // firing, resolved
	Value       float64                `json:"value"`
	Threshold   float64                `json:"threshold"`
	Metadata    map[string]interface{} `json:"metadata"`
	FiredAt     time.Time              `json:"fired_at"`
	ResolvedAt  *time.Time             `json:"resolved_at,omitempty"`
}

// NewModelMonitor creates a new model monitor
func NewModelMonitor(logger *logrus.Logger) *ModelMonitor {
	return &ModelMonitor{
		metrics:   make(map[string]*MonitoringMetrics),
		alerts:    make(map[string]*AlertRule),
		history:   make(map[string][]*MetricSnapshot),
		logger:    logger,
		tracer:    otel.Tracer("ai.services.model_monitor"),
		alertChan: make(chan *Alert, 100),
	}
}

// RecordRequest records a request for monitoring
func (mm *ModelMonitor) RecordRequest(modelID string, latency time.Duration, cost float64) {
	mm.mutex.Lock()
	defer mm.mutex.Unlock()

	metrics, exists := mm.metrics[modelID]
	if !exists {
		metrics = &MonitoringMetrics{
			ModelID:      modelID,
			HealthStatus: "healthy",
			LastUpdate:   time.Now(),
		}
		mm.metrics[modelID] = metrics
	}

	// Update metrics
	metrics.TotalRequests++
	metrics.LastUpdate = time.Now()

	// Update latency metrics (simplified - in production, use proper percentile calculation)
	if metrics.AverageLatency == 0 {
		metrics.AverageLatency = latency
	} else {
		// Simple moving average
		metrics.AverageLatency = (metrics.AverageLatency + latency) / 2
	}

	// Update cost metrics
	metrics.CostPerHour += cost

	// Calculate requests per second (simplified)
	if metrics.TotalRequests > 1 {
		duration := time.Since(metrics.LastUpdate.Add(-time.Minute))
		if duration > 0 {
			metrics.RequestsPerSecond = float64(metrics.TotalRequests) / duration.Seconds()
		}
	}

	// Update success rate
	metrics.SuccessRate = float64(metrics.TotalRequests-metrics.ErrorCount) / float64(metrics.TotalRequests)
	metrics.ErrorRate = float64(metrics.ErrorCount) / float64(metrics.TotalRequests)

	// Check alerts
	mm.checkAlerts(modelID, metrics)
}

// RecordError records an error for monitoring
func (mm *ModelMonitor) RecordError(modelID string, err error) {
	mm.mutex.Lock()
	defer mm.mutex.Unlock()

	metrics, exists := mm.metrics[modelID]
	if !exists {
		metrics = &MonitoringMetrics{
			ModelID:      modelID,
			HealthStatus: "healthy",
			LastUpdate:   time.Now(),
		}
		mm.metrics[modelID] = metrics
	}

	metrics.ErrorCount++
	metrics.LastError = err.Error()
	metrics.LastUpdate = time.Now()

	// Update error rate
	if metrics.TotalRequests > 0 {
		metrics.ErrorRate = float64(metrics.ErrorCount) / float64(metrics.TotalRequests)
		metrics.SuccessRate = 1.0 - metrics.ErrorRate
	}

	// Update health status based on error rate
	if metrics.ErrorRate > 0.5 {
		metrics.HealthStatus = "unhealthy"
	} else if metrics.ErrorRate > 0.1 {
		metrics.HealthStatus = "degraded"
	}

	// Check alerts
	mm.checkAlerts(modelID, metrics)

	mm.logger.WithFields(logrus.Fields{
		"model_id":   modelID,
		"error":      err.Error(),
		"error_rate": metrics.ErrorRate,
	}).Warn("Model error recorded")
}

// GetMetrics returns monitoring metrics for a model
func (mm *ModelMonitor) GetMetrics(modelID string) *MonitoringMetrics {
	mm.mutex.RLock()
	defer mm.mutex.RUnlock()

	if metrics, exists := mm.metrics[modelID]; exists {
		return metrics
	}
	return nil
}

// GetAllMetrics returns monitoring metrics for all models
func (mm *ModelMonitor) GetAllMetrics() map[string]*MonitoringMetrics {
	mm.mutex.RLock()
	defer mm.mutex.RUnlock()

	result := make(map[string]*MonitoringMetrics)
	for k, v := range mm.metrics {
		result[k] = v
	}
	return result
}

// AddAlertRule adds a new alert rule
func (mm *ModelMonitor) AddAlertRule(rule *AlertRule) {
	mm.mutex.Lock()
	defer mm.mutex.Unlock()

	rule.CreatedAt = time.Now()
	rule.UpdatedAt = time.Now()
	mm.alerts[rule.ID] = rule

	mm.logger.WithFields(logrus.Fields{
		"rule_id":  rule.ID,
		"model_id": rule.ModelID,
		"metric":   rule.Metric,
	}).Info("Alert rule added")
}

// RemoveAlertRule removes an alert rule
func (mm *ModelMonitor) RemoveAlertRule(ruleID string) {
	mm.mutex.Lock()
	defer mm.mutex.Unlock()

	delete(mm.alerts, ruleID)
	mm.logger.WithField("rule_id", ruleID).Info("Alert rule removed")
}

// GetAlertRules returns all alert rules
func (mm *ModelMonitor) GetAlertRules() map[string]*AlertRule {
	mm.mutex.RLock()
	defer mm.mutex.RUnlock()

	result := make(map[string]*AlertRule)
	for k, v := range mm.alerts {
		result[k] = v
	}
	return result
}

// GetAlerts returns the alert channel for consuming alerts
func (mm *ModelMonitor) GetAlerts() <-chan *Alert {
	return mm.alertChan
}

// TakeSnapshot takes a snapshot of current metrics
func (mm *ModelMonitor) TakeSnapshot(modelID string) {
	mm.mutex.Lock()
	defer mm.mutex.Unlock()

	metrics, exists := mm.metrics[modelID]
	if !exists {
		return
	}

	snapshot := &MetricSnapshot{
		Timestamp: time.Now(),
		Metrics:   metrics,
	}

	// Add to history
	if _, exists := mm.history[modelID]; !exists {
		mm.history[modelID] = make([]*MetricSnapshot, 0)
	}

	mm.history[modelID] = append(mm.history[modelID], snapshot)

	// Keep only last 1000 snapshots
	if len(mm.history[modelID]) > 1000 {
		mm.history[modelID] = mm.history[modelID][1:]
	}
}

// GetHistory returns metric history for a model
func (mm *ModelMonitor) GetHistory(modelID string, duration time.Duration) []*MetricSnapshot {
	mm.mutex.RLock()
	defer mm.mutex.RUnlock()

	history, exists := mm.history[modelID]
	if !exists {
		return nil
	}

	cutoff := time.Now().Add(-duration)
	var result []*MetricSnapshot

	for _, snapshot := range history {
		if snapshot.Timestamp.After(cutoff) {
			result = append(result, snapshot)
		}
	}

	return result
}

// checkAlerts checks if any alert rules are triggered
func (mm *ModelMonitor) checkAlerts(modelID string, metrics *MonitoringMetrics) {
	for _, rule := range mm.alerts {
		if !rule.Enabled || rule.ModelID != modelID {
			continue
		}

		value := mm.getMetricValue(metrics, rule.Metric)
		triggered := mm.evaluateCondition(value, rule.Operator, rule.Threshold)

		if triggered {
			alert := &Alert{
				ID:          generateAlertID(),
				RuleID:      rule.ID,
				ModelID:     modelID,
				Name:        rule.Name,
				Description: rule.Description,
				Severity:    rule.Severity,
				Status:      "firing",
				Value:       value,
				Threshold:   rule.Threshold,
				FiredAt:     time.Now(),
				Metadata:    make(map[string]interface{}),
			}

			// Send alert (non-blocking)
			select {
			case mm.alertChan <- alert:
				mm.logger.WithFields(logrus.Fields{
					"alert_id":  alert.ID,
					"rule_id":   rule.ID,
					"model_id":  modelID,
					"metric":    rule.Metric,
					"value":     value,
					"threshold": rule.Threshold,
				}).Warn("Alert triggered")
			default:
				mm.logger.Warn("Alert channel full, dropping alert")
			}
		}
	}
}

// getMetricValue extracts a specific metric value
func (mm *ModelMonitor) getMetricValue(metrics *MonitoringMetrics, metricName string) float64 {
	switch metricName {
	case "error_rate":
		return metrics.ErrorRate
	case "success_rate":
		return metrics.SuccessRate
	case "average_latency":
		return float64(metrics.AverageLatency.Milliseconds())
	case "requests_per_second":
		return metrics.RequestsPerSecond
	case "cost_per_hour":
		return metrics.CostPerHour
	case "memory_usage":
		return metrics.MemoryUsage
	case "cpu_usage":
		return metrics.CPUUsage
	case "queue_length":
		return float64(metrics.QueueLength)
	case "active_connections":
		return float64(metrics.ActiveConnections)
	default:
		return 0
	}
}

// evaluateCondition evaluates an alert condition
func (mm *ModelMonitor) evaluateCondition(value float64, operator string, threshold float64) bool {
	switch operator {
	case "gt":
		return value > threshold
	case "gte":
		return value >= threshold
	case "lt":
		return value < threshold
	case "lte":
		return value <= threshold
	case "eq":
		return value == threshold
	default:
		return false
	}
}

// generateAlertID generates a unique alert ID
func generateAlertID() string {
	return fmt.Sprintf("alert_%d", time.Now().UnixNano())
}

// StartMonitoring starts the monitoring background processes
func (mm *ModelMonitor) StartMonitoring() {
	// Start snapshot routine
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()

		for range ticker.C {
			mm.mutex.RLock()
			modelIDs := make([]string, 0, len(mm.metrics))
			for modelID := range mm.metrics {
				modelIDs = append(modelIDs, modelID)
			}
			mm.mutex.RUnlock()

			for _, modelID := range modelIDs {
				mm.TakeSnapshot(modelID)
			}
		}
	}()

	// Start health check routine
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			mm.performHealthChecks()
		}
	}()

	mm.logger.Info("Model monitoring started")
}

// performHealthChecks performs health checks on all models
func (mm *ModelMonitor) performHealthChecks() {
	mm.mutex.Lock()
	defer mm.mutex.Unlock()

	now := time.Now()
	for _, metrics := range mm.metrics {
		// Check if model has been inactive for too long
		if now.Sub(metrics.LastUpdate) > 5*time.Minute {
			metrics.HealthStatus = "unhealthy"
		} else if metrics.ErrorRate > 0.1 {
			metrics.HealthStatus = "degraded"
		} else {
			metrics.HealthStatus = "healthy"
		}

		metrics.LastHealthCheck = now
	}
}

// GetHealthSummary returns a summary of model health
func (mm *ModelMonitor) GetHealthSummary() map[string]interface{} {
	mm.mutex.RLock()
	defer mm.mutex.RUnlock()

	summary := map[string]interface{}{
		"total_models":     len(mm.metrics),
		"healthy_models":   0,
		"degraded_models":  0,
		"unhealthy_models": 0,
		"total_requests":   int64(0),
		"total_errors":     int64(0),
		"average_latency":  time.Duration(0),
	}

	var totalLatency time.Duration
	for _, metrics := range mm.metrics {
		summary["total_requests"] = summary["total_requests"].(int64) + metrics.TotalRequests
		summary["total_errors"] = summary["total_errors"].(int64) + metrics.ErrorCount
		totalLatency += metrics.AverageLatency

		switch metrics.HealthStatus {
		case "healthy":
			summary["healthy_models"] = summary["healthy_models"].(int) + 1
		case "degraded":
			summary["degraded_models"] = summary["degraded_models"].(int) + 1
		case "unhealthy":
			summary["unhealthy_models"] = summary["unhealthy_models"].(int) + 1
		}
	}

	if len(mm.metrics) > 0 {
		summary["average_latency"] = totalLatency / time.Duration(len(mm.metrics))
	}

	return summary
}
