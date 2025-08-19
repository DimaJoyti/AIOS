package resources

import (
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// MetricsSnapshot represents a snapshot of resource metrics
type MetricsSnapshot struct {
	// Access metrics
	TotalAccesses      int64   `json:"total_accesses"`
	SuccessfulAccesses int64   `json:"successful_accesses"`
	FailedAccesses     int64   `json:"failed_accesses"`
	SuccessRate        float64 `json:"success_rate"`

	// Performance metrics
	AverageLatency time.Duration `json:"average_latency"`
	MinLatency     time.Duration `json:"min_latency"`
	MaxLatency     time.Duration `json:"max_latency"`
	TotalLatency   time.Duration `json:"total_latency"`

	// Cache metrics
	CacheHits    int64   `json:"cache_hits"`
	CacheMisses  int64   `json:"cache_misses"`
	CacheHitRate float64 `json:"cache_hit_rate"`

	// Resource metrics
	TotalResources int   `json:"total_resources"`
	TotalSize      int64 `json:"total_size"`
	AverageSize    int64 `json:"average_size"`

	// Operation metrics
	OperationCounts    map[string]int64         `json:"operation_counts"`
	OperationLatencies map[string]time.Duration `json:"operation_latencies"`

	// Error metrics
	ErrorCounts map[string]int64 `json:"error_counts"`
	ErrorsByURI map[string]int64 `json:"errors_by_uri"`

	// Time metrics
	StartTime      time.Time `json:"start_time"`
	LastAccess     time.Time `json:"last_access"`
	CollectionTime time.Time `json:"collection_time"`
}

// DefaultResourceMetrics implements ResourceMetrics
type DefaultResourceMetrics struct {
	mu         sync.RWMutex
	startTime  time.Time
	lastAccess time.Time

	// Access counters
	totalAccesses      int64
	successfulAccesses int64
	failedAccesses     int64

	// Latency tracking
	totalLatency time.Duration
	minLatency   time.Duration
	maxLatency   time.Duration
	latencyCount int64

	// Cache tracking
	cacheHits   int64
	cacheMisses int64

	// Resource tracking
	resourceSizes map[string]int64
	totalSize     int64

	// Operation tracking
	operationCounts       map[string]int64
	operationLatencies    map[string]time.Duration
	operationLatencyCount map[string]int64

	// Error tracking
	errorCounts map[string]int64
	errorsByURI map[string]int64

	logger *logrus.Logger
}

// NewResourceMetrics creates a new resource metrics collector
func NewResourceMetrics(logger *logrus.Logger) (ResourceMetrics, error) {
	return &DefaultResourceMetrics{
		startTime:             time.Now(),
		resourceSizes:         make(map[string]int64),
		operationCounts:       make(map[string]int64),
		operationLatencies:    make(map[string]time.Duration),
		operationLatencyCount: make(map[string]int64),
		errorCounts:           make(map[string]int64),
		errorsByURI:           make(map[string]int64),
		logger:                logger,
	}, nil
}

// RecordResourceAccess records a resource access operation
func (m *DefaultResourceMetrics) RecordResourceAccess(uri, operation string, duration time.Duration, success bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.totalAccesses++
	m.lastAccess = time.Now()

	if success {
		m.successfulAccesses++
	} else {
		m.failedAccesses++
		m.errorsByURI[uri]++
	}

	// Update latency metrics
	m.totalLatency += duration
	m.latencyCount++

	if m.minLatency == 0 || duration < m.minLatency {
		m.minLatency = duration
	}
	if duration > m.maxLatency {
		m.maxLatency = duration
	}

	// Update operation metrics
	m.operationCounts[operation]++
	m.operationLatencies[operation] += duration
	m.operationLatencyCount[operation]++
}

// RecordResourceSize records the size of a resource
func (m *DefaultResourceMetrics) RecordResourceSize(uri string, size int64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	oldSize, exists := m.resourceSizes[uri]
	if exists {
		m.totalSize -= oldSize
	}

	m.resourceSizes[uri] = size
	m.totalSize += size
}

// RecordCacheHit records a cache hit or miss
func (m *DefaultResourceMetrics) RecordCacheHit(uri string, hit bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if hit {
		m.cacheHits++
	} else {
		m.cacheMisses++
	}
}

// RecordError records an error for a specific operation
func (m *DefaultResourceMetrics) RecordError(uri, operation string, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	errorType := "unknown"
	if err != nil {
		errorType = err.Error()
		if len(errorType) > 100 {
			errorType = errorType[:100] + "..."
		}
	}

	m.errorCounts[errorType]++
	m.errorsByURI[uri]++

	m.logger.WithFields(logrus.Fields{
		"uri":       uri,
		"operation": operation,
		"error":     errorType,
	}).Debug("Resource error recorded")
}

// GetMetrics returns a snapshot of current metrics
func (m *DefaultResourceMetrics) GetMetrics() *MetricsSnapshot {
	m.mu.RLock()
	defer m.mu.RUnlock()

	snapshot := &MetricsSnapshot{
		TotalAccesses:      m.totalAccesses,
		SuccessfulAccesses: m.successfulAccesses,
		FailedAccesses:     m.failedAccesses,
		CacheHits:          m.cacheHits,
		CacheMisses:        m.cacheMisses,
		TotalResources:     len(m.resourceSizes),
		TotalSize:          m.totalSize,
		TotalLatency:       m.totalLatency,
		MinLatency:         m.minLatency,
		MaxLatency:         m.maxLatency,
		StartTime:          m.startTime,
		LastAccess:         m.lastAccess,
		CollectionTime:     time.Now(),
		OperationCounts:    make(map[string]int64),
		OperationLatencies: make(map[string]time.Duration),
		ErrorCounts:        make(map[string]int64),
		ErrorsByURI:        make(map[string]int64),
	}

	// Calculate derived metrics
	if m.totalAccesses > 0 {
		snapshot.SuccessRate = float64(m.successfulAccesses) / float64(m.totalAccesses)
	}

	if m.latencyCount > 0 {
		snapshot.AverageLatency = m.totalLatency / time.Duration(m.latencyCount)
	}

	if m.cacheHits+m.cacheMisses > 0 {
		snapshot.CacheHitRate = float64(m.cacheHits) / float64(m.cacheHits+m.cacheMisses)
	}

	if len(m.resourceSizes) > 0 {
		snapshot.AverageSize = m.totalSize / int64(len(m.resourceSizes))
	}

	// Copy maps to avoid race conditions
	for op, count := range m.operationCounts {
		snapshot.OperationCounts[op] = count
	}

	for op, latency := range m.operationLatencies {
		if count := m.operationLatencyCount[op]; count > 0 {
			snapshot.OperationLatencies[op] = latency / time.Duration(count)
		}
	}

	for errorType, count := range m.errorCounts {
		snapshot.ErrorCounts[errorType] = count
	}

	for uri, count := range m.errorsByURI {
		snapshot.ErrorsByURI[uri] = count
	}

	return snapshot
}

// Reset resets all metrics
func (m *DefaultResourceMetrics) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.startTime = time.Now()
	m.lastAccess = time.Time{}
	m.totalAccesses = 0
	m.successfulAccesses = 0
	m.failedAccesses = 0
	m.totalLatency = 0
	m.minLatency = 0
	m.maxLatency = 0
	m.latencyCount = 0
	m.cacheHits = 0
	m.cacheMisses = 0
	m.totalSize = 0

	m.resourceSizes = make(map[string]int64)
	m.operationCounts = make(map[string]int64)
	m.operationLatencies = make(map[string]time.Duration)
	m.operationLatencyCount = make(map[string]int64)
	m.errorCounts = make(map[string]int64)
	m.errorsByURI = make(map[string]int64)

	m.logger.Info("Resource metrics reset")
}

// MetricsReporter provides periodic metrics reporting
type MetricsReporter struct {
	metrics  ResourceMetrics
	logger   *logrus.Logger
	interval time.Duration
	stopChan chan struct{}
	running  bool
	mu       sync.Mutex
}

// NewMetricsReporter creates a new metrics reporter
func NewMetricsReporter(metrics ResourceMetrics, interval time.Duration, logger *logrus.Logger) *MetricsReporter {
	return &MetricsReporter{
		metrics:  metrics,
		logger:   logger,
		interval: interval,
		stopChan: make(chan struct{}),
	}
}

// Start starts periodic metrics reporting
func (r *MetricsReporter) Start() {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.running {
		return
	}

	r.running = true
	go r.reportLoop()
}

// Stop stops metrics reporting
func (r *MetricsReporter) Stop() {
	r.mu.Lock()
	defer r.mu.Unlock()

	if !r.running {
		return
	}

	r.running = false
	close(r.stopChan)
}

// reportLoop runs the periodic reporting loop
func (r *MetricsReporter) reportLoop() {
	ticker := time.NewTicker(r.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			r.reportMetrics()
		case <-r.stopChan:
			return
		}
	}
}

// reportMetrics logs current metrics
func (r *MetricsReporter) reportMetrics() {
	snapshot := r.metrics.(*DefaultResourceMetrics).GetMetrics()

	r.logger.WithFields(logrus.Fields{
		"total_accesses":  snapshot.TotalAccesses,
		"success_rate":    snapshot.SuccessRate,
		"cache_hit_rate":  snapshot.CacheHitRate,
		"average_latency": snapshot.AverageLatency,
		"total_resources": snapshot.TotalResources,
		"total_size":      snapshot.TotalSize,
		"uptime":          time.Since(snapshot.StartTime),
	}).Info("Resource metrics report")
}

// GetOverallStats returns overall resource statistics
func (m *DefaultResourceMetrics) GetOverallStats() map[string]interface{} {
	snapshot := m.GetMetrics()
	return map[string]interface{}{
		"total_accesses":      snapshot.TotalAccesses,
		"successful_accesses": snapshot.SuccessfulAccesses,
		"failed_accesses":     snapshot.FailedAccesses,
		"cache_hit_rate":      snapshot.CacheHitRate,
		"average_latency":     snapshot.AverageLatency,
		"total_size":          snapshot.TotalSize,
		"uptime":              time.Since(snapshot.StartTime).String(),
	}
}

// GetResourceStats returns statistics for a specific resource
func (m *DefaultResourceMetrics) GetResourceStats(uri string) map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if size, exists := m.resourceSizes[uri]; exists {
		return map[string]interface{}{
			"uri":        uri,
			"size_bytes": size,
			"errors":     m.errorsByURI[uri],
		}
	}

	return map[string]interface{}{
		"uri":    uri,
		"exists": false,
	}
}

// GetTopResources returns the most accessed resources
func (m *DefaultResourceMetrics) GetTopResources(limit int) []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	type resourceData struct {
		uri  string
		size int64
	}

	var resources []resourceData
	for uri, size := range m.resourceSizes {
		resources = append(resources, resourceData{uri: uri, size: size})
	}

	// Simple sorting by size (largest first)
	for i := 0; i < len(resources)-1; i++ {
		for j := i + 1; j < len(resources); j++ {
			if resources[i].size < resources[j].size {
				resources[i], resources[j] = resources[j], resources[i]
			}
		}
	}

	result := make([]string, 0, limit)
	for i := 0; i < len(resources) && i < limit; i++ {
		result = append(result, resources[i].uri)
	}

	return result
}
