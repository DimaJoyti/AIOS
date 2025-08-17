package performance

import (
	"context"
	"fmt"
	"math"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

// LoadTester provides comprehensive load testing capabilities
type LoadTester struct {
	config LoadTestConfig
	stats  *LoadTestStats
	mu     sync.RWMutex
}

// LoadTestConfig defines load test configuration
type LoadTestConfig struct {
	Duration        time.Duration `json:"duration"`
	Concurrency     int           `json:"concurrency"`
	RampUpTime      time.Duration `json:"ramp_up_time"`
	RampDownTime    time.Duration `json:"ramp_down_time"`
	RequestRate     int           `json:"request_rate"` // requests per second
	MaxRequests     int64         `json:"max_requests"`
	Timeout         time.Duration `json:"timeout"`
	ThinkTime       time.Duration `json:"think_time"`
	FailureThreshold float64      `json:"failure_threshold"` // percentage
}

// LoadTestStats tracks load test statistics
type LoadTestStats struct {
	StartTime       time.Time     `json:"start_time"`
	EndTime         time.Time     `json:"end_time"`
	Duration        time.Duration `json:"duration"`
	TotalRequests   int64         `json:"total_requests"`
	SuccessRequests int64         `json:"success_requests"`
	FailedRequests  int64         `json:"failed_requests"`
	RequestsPerSec  float64       `json:"requests_per_sec"`
	AvgResponseTime time.Duration `json:"avg_response_time"`
	MinResponseTime time.Duration `json:"min_response_time"`
	MaxResponseTime time.Duration `json:"max_response_time"`
	P50ResponseTime time.Duration `json:"p50_response_time"`
	P95ResponseTime time.Duration `json:"p95_response_time"`
	P99ResponseTime time.Duration `json:"p99_response_time"`
	ErrorRate       float64       `json:"error_rate"`
	Throughput      float64       `json:"throughput"` // bytes per second
	ResponseTimes   []time.Duration `json:"-"`
	Errors          []LoadTestError `json:"errors"`
	CPUUsage        []float64     `json:"cpu_usage"`
	MemoryUsage     []int64       `json:"memory_usage"`
}

// LoadTestError represents an error during load testing
type LoadTestError struct {
	Timestamp time.Time `json:"timestamp"`
	Error     string    `json:"error"`
	Count     int       `json:"count"`
}

// LoadTestResult represents the result of a load test
type LoadTestResult struct {
	Config LoadTestConfig `json:"config"`
	Stats  LoadTestStats  `json:"stats"`
	Passed bool           `json:"passed"`
	Report string         `json:"report"`
}

// TestFunction represents a function to be load tested
type TestFunction func(ctx context.Context) error

// NewLoadTester creates a new load tester
func NewLoadTester(config LoadTestConfig) *LoadTester {
	return &LoadTester{
		config: config,
		stats: &LoadTestStats{
			ResponseTimes: make([]time.Duration, 0),
			Errors:        make([]LoadTestError, 0),
			CPUUsage:      make([]float64, 0),
			MemoryUsage:   make([]int64, 0),
			MinResponseTime: time.Hour, // Initialize to high value
		},
	}
}

// RunLoadTest executes a load test
func (lt *LoadTester) RunLoadTest(ctx context.Context, testFunc TestFunction) (*LoadTestResult, error) {
	lt.stats.StartTime = time.Now()
	
	// Create context with timeout
	testCtx, cancel := context.WithTimeout(ctx, lt.config.Duration)
	defer cancel()
	
	// Start monitoring
	monitorCtx, monitorCancel := context.WithCancel(testCtx)
	defer monitorCancel()
	go lt.monitorResources(monitorCtx)
	
	// Execute load test
	if err := lt.executeLoadTest(testCtx, testFunc); err != nil {
		return nil, err
	}
	
	lt.stats.EndTime = time.Now()
	lt.stats.Duration = lt.stats.EndTime.Sub(lt.stats.StartTime)
	
	// Calculate final statistics
	lt.calculateFinalStats()
	
	// Generate result
	result := &LoadTestResult{
		Config: lt.config,
		Stats:  *lt.stats,
		Passed: lt.evaluateResult(),
		Report: lt.generateReport(),
	}
	
	return result, nil
}

// executeLoadTest executes the actual load test
func (lt *LoadTester) executeLoadTest(ctx context.Context, testFunc TestFunction) error {
	var wg sync.WaitGroup
	requestChan := make(chan struct{}, lt.config.Concurrency)
	
	// Rate limiter
	ticker := time.NewTicker(time.Second / time.Duration(lt.config.RequestRate))
	defer ticker.Stop()
	
	// Ramp up workers
	for i := 0; i < lt.config.Concurrency; i++ {
		wg.Add(1)
		go lt.worker(ctx, &wg, requestChan, testFunc)
		
		// Ramp up delay
		if lt.config.RampUpTime > 0 {
			rampDelay := lt.config.RampUpTime / time.Duration(lt.config.Concurrency)
			time.Sleep(rampDelay)
		}
	}
	
	// Send requests
	go func() {
		defer close(requestChan)
		
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if lt.config.MaxRequests > 0 && atomic.LoadInt64(&lt.stats.TotalRequests) >= lt.config.MaxRequests {
					return
				}
				
				select {
				case requestChan <- struct{}{}:
				default:
					// Channel full, skip this request
				}
			}
		}
	}()
	
	// Wait for completion
	wg.Wait()
	
	return nil
}

// worker executes individual test requests
func (lt *LoadTester) worker(ctx context.Context, wg *sync.WaitGroup, requestChan <-chan struct{}, testFunc TestFunction) {
	defer wg.Done()
	
	for {
		select {
		case <-ctx.Done():
			return
		case _, ok := <-requestChan:
			if !ok {
				return
			}
			
			// Execute test function
			start := time.Now()
			err := testFunc(ctx)
			responseTime := time.Since(start)
			
			// Record statistics
			lt.recordRequest(responseTime, err)
			
			// Think time
			if lt.config.ThinkTime > 0 {
				time.Sleep(lt.config.ThinkTime)
			}
		}
	}
}

// recordRequest records the result of a single request
func (lt *LoadTester) recordRequest(responseTime time.Duration, err error) {
	lt.mu.Lock()
	defer lt.mu.Unlock()
	
	atomic.AddInt64(&lt.stats.TotalRequests, 1)
	
	if err != nil {
		atomic.AddInt64(&lt.stats.FailedRequests, 1)
		lt.recordError(err)
	} else {
		atomic.AddInt64(&lt.stats.SuccessRequests, 1)
	}
	
	// Record response time
	lt.stats.ResponseTimes = append(lt.stats.ResponseTimes, responseTime)
	
	// Update min/max response times
	if responseTime < lt.stats.MinResponseTime {
		lt.stats.MinResponseTime = responseTime
	}
	if responseTime > lt.stats.MaxResponseTime {
		lt.stats.MaxResponseTime = responseTime
	}
}

// recordError records an error
func (lt *LoadTester) recordError(err error) {
	errorMsg := err.Error()
	
	// Find existing error or create new one
	for i := range lt.stats.Errors {
		if lt.stats.Errors[i].Error == errorMsg {
			lt.stats.Errors[i].Count++
			return
		}
	}
	
	// New error
	lt.stats.Errors = append(lt.stats.Errors, LoadTestError{
		Timestamp: time.Now(),
		Error:     errorMsg,
		Count:     1,
	})
}

// monitorResources monitors system resources during the test
func (lt *LoadTester) monitorResources(ctx context.Context) {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// Monitor CPU and memory usage
			var m runtime.MemStats
			runtime.ReadMemStats(&m)
			
			lt.mu.Lock()
			lt.stats.MemoryUsage = append(lt.stats.MemoryUsage, int64(m.Alloc))
			// CPU usage would require additional system calls
			lt.stats.CPUUsage = append(lt.stats.CPUUsage, 0.0) // Placeholder
			lt.mu.Unlock()
		}
	}
}

// calculateFinalStats calculates final statistics
func (lt *LoadTester) calculateFinalStats() {
	lt.mu.Lock()
	defer lt.mu.Unlock()
	
	if lt.stats.Duration > 0 {
		lt.stats.RequestsPerSec = float64(lt.stats.TotalRequests) / lt.stats.Duration.Seconds()
	}
	
	if lt.stats.TotalRequests > 0 {
		lt.stats.ErrorRate = float64(lt.stats.FailedRequests) / float64(lt.stats.TotalRequests) * 100
	}
	
	// Calculate response time percentiles
	if len(lt.stats.ResponseTimes) > 0 {
		// Sort response times
		responseTimes := make([]time.Duration, len(lt.stats.ResponseTimes))
		copy(responseTimes, lt.stats.ResponseTimes)
		
		// Simple sort (for production, use sort.Slice)
		for i := 0; i < len(responseTimes); i++ {
			for j := i + 1; j < len(responseTimes); j++ {
				if responseTimes[i] > responseTimes[j] {
					responseTimes[i], responseTimes[j] = responseTimes[j], responseTimes[i]
				}
			}
		}
		
		// Calculate percentiles
		lt.stats.P50ResponseTime = responseTimes[len(responseTimes)*50/100]
		lt.stats.P95ResponseTime = responseTimes[len(responseTimes)*95/100]
		lt.stats.P99ResponseTime = responseTimes[len(responseTimes)*99/100]
		
		// Calculate average
		var total time.Duration
		for _, rt := range responseTimes {
			total += rt
		}
		lt.stats.AvgResponseTime = total / time.Duration(len(responseTimes))
	}
}

// evaluateResult evaluates if the load test passed
func (lt *LoadTester) evaluateResult() bool {
	// Check error rate threshold
	if lt.stats.ErrorRate > lt.config.FailureThreshold {
		return false
	}
	
	// Check if test completed successfully
	if lt.stats.TotalRequests == 0 {
		return false
	}
	
	return true
}

// generateReport generates a human-readable report
func (lt *LoadTester) generateReport() string {
	report := fmt.Sprintf(`
Load Test Report
================
Duration: %v
Total Requests: %d
Success Rate: %.2f%%
Error Rate: %.2f%%
Requests/sec: %.2f
Average Response Time: %v
Min Response Time: %v
Max Response Time: %v
P50 Response Time: %v
P95 Response Time: %v
P99 Response Time: %v

Errors:
`,
		lt.stats.Duration,
		lt.stats.TotalRequests,
		float64(lt.stats.SuccessRequests)/float64(lt.stats.TotalRequests)*100,
		lt.stats.ErrorRate,
		lt.stats.RequestsPerSec,
		lt.stats.AvgResponseTime,
		lt.stats.MinResponseTime,
		lt.stats.MaxResponseTime,
		lt.stats.P50ResponseTime,
		lt.stats.P95ResponseTime,
		lt.stats.P99ResponseTime,
	)
	
	for _, err := range lt.stats.Errors {
		report += fmt.Sprintf("  %s: %d occurrences\n", err.Error, err.Count)
	}
	
	return report
}

// StressTest performs a stress test with increasing load
func (lt *LoadTester) StressTest(ctx context.Context, testFunc TestFunction, maxConcurrency int) ([]*LoadTestResult, error) {
	results := make([]*LoadTestResult, 0)
	
	for concurrency := 1; concurrency <= maxConcurrency; concurrency *= 2 {
		config := lt.config
		config.Concurrency = concurrency
		config.Duration = time.Minute // Shorter duration for stress test
		
		stressTester := NewLoadTester(config)
		result, err := stressTester.RunLoadTest(ctx, testFunc)
		if err != nil {
			return results, err
		}
		
		results = append(results, result)
		
		// Stop if error rate is too high
		if result.Stats.ErrorRate > 50 {
			break
		}
	}
	
	return results, nil
}

// SpikeTest performs a spike test with sudden load increases
func (lt *LoadTester) SpikeTest(ctx context.Context, testFunc TestFunction) (*LoadTestResult, error) {
	// Implement spike test logic
	// This would involve sudden increases in load followed by normal load
	
	config := lt.config
	config.RampUpTime = 0 // No ramp up for spike test
	config.Duration = time.Minute * 5
	
	spikeTester := NewLoadTester(config)
	return spikeTester.RunLoadTest(ctx, testFunc)
}

// VolumeTest performs a volume test with sustained high load
func (lt *LoadTester) VolumeTest(ctx context.Context, testFunc TestFunction) (*LoadTestResult, error) {
	config := lt.config
	config.Duration = time.Hour // Long duration for volume test
	config.Concurrency = int(math.Max(float64(config.Concurrency), 100)) // High concurrency
	
	volumeTester := NewLoadTester(config)
	return volumeTester.RunLoadTest(ctx, testFunc)
}

// BenchmarkComparison compares performance between different implementations
type BenchmarkComparison struct {
	Name    string
	Results []*LoadTestResult
}

// CompareBenchmarks compares multiple benchmark results
func CompareBenchmarks(comparisons []BenchmarkComparison) string {
	report := "Benchmark Comparison\n"
	report += "===================\n\n"
	
	for _, comp := range comparisons {
		report += fmt.Sprintf("%s:\n", comp.Name)
		for i, result := range comp.Results {
			report += fmt.Sprintf("  Run %d: %.2f req/s, %.2f%% errors, %v avg response\n",
				i+1,
				result.Stats.RequestsPerSec,
				result.Stats.ErrorRate,
				result.Stats.AvgResponseTime,
			)
		}
		report += "\n"
	}
	
	return report
}
