package testing

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/aios/aios/internal/ai"
	"github.com/aios/aios/internal/ai/monitoring"
	"github.com/sirupsen/logrus"
)

// AITestSuite provides comprehensive testing for AI services
type AITestSuite struct {
	t                  *testing.T
	logger             *logrus.Logger
	performanceMonitor *monitoring.PerformanceMonitor
	services           map[string]interface{}
	testResults        []TestResult
	mu                 sync.RWMutex
}

// TestResult represents the result of a test
type TestResult struct {
	TestName     string                 `json:"test_name"`
	Service      string                 `json:"service"`
	Success      bool                   `json:"success"`
	Duration     time.Duration          `json:"duration"`
	ErrorMessage string                 `json:"error_message,omitempty"`
	Metrics      map[string]interface{} `json:"metrics,omitempty"`
	Timestamp    time.Time              `json:"timestamp"`
}

// TestConfig represents test configuration
type TestConfig struct {
	ConcurrentUsers     int           `json:"concurrent_users"`
	TestDuration        time.Duration `json:"test_duration"`
	RequestsPerSecond   int           `json:"requests_per_second"`
	TimeoutPerRequest   time.Duration `json:"timeout_per_request"`
	EnableLoadTesting   bool          `json:"enable_load_testing"`
	EnableStressTesting bool          `json:"enable_stress_testing"`
}

// NewAITestSuite creates a new AI test suite
func NewAITestSuite(t *testing.T, logger *logrus.Logger) (*AITestSuite, error) {
	perfMonitor, err := monitoring.NewPerformanceMonitor(logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create performance monitor: %w", err)
	}

	return &AITestSuite{
		t:                  t,
		logger:             logger,
		performanceMonitor: perfMonitor,
		services:           make(map[string]interface{}),
		testResults:        make([]TestResult, 0),
	}, nil
}

// RegisterService registers a service for testing
func (suite *AITestSuite) RegisterService(name string, service interface{}) {
	suite.services[name] = service
}

// RunAllTests runs all available tests
func (suite *AITestSuite) RunAllTests(ctx context.Context, config TestConfig) error {
	suite.logger.Info("Starting comprehensive AI service tests")

	// Unit tests
	if err := suite.runUnitTests(ctx); err != nil {
		return fmt.Errorf("unit tests failed: %w", err)
	}

	// Integration tests
	if err := suite.runIntegrationTests(ctx); err != nil {
		return fmt.Errorf("integration tests failed: %w", err)
	}

	// Performance tests
	if err := suite.runPerformanceTests(ctx, config); err != nil {
		return fmt.Errorf("performance tests failed: %w", err)
	}

	// Load tests
	if config.EnableLoadTesting {
		if err := suite.runLoadTests(ctx, config); err != nil {
			return fmt.Errorf("load tests failed: %w", err)
		}
	}

	// Stress tests
	if config.EnableStressTesting {
		if err := suite.runStressTests(ctx, config); err != nil {
			return fmt.Errorf("stress tests failed: %w", err)
		}
	}

	suite.logger.Info("All AI service tests completed")
	return nil
}

// Unit Tests

func (suite *AITestSuite) runUnitTests(ctx context.Context) error {
	suite.logger.Info("Running unit tests")

	// Test LLM Service
	if llmService, exists := suite.services["llm"]; exists {
		if err := suite.testLLMService(ctx, llmService); err != nil {
			return err
		}
	}

	// Test Voice Service
	if voiceService, exists := suite.services["voice"]; exists {
		if err := suite.testVoiceService(ctx, voiceService); err != nil {
			return err
		}
	}

	// Test CV Service
	if cvService, exists := suite.services["cv"]; exists {
		if err := suite.testCVService(ctx, cvService); err != nil {
			return err
		}
	}

	// Test NLP Service
	if nlpService, exists := suite.services["nlp"]; exists {
		if err := suite.testNLPService(ctx, nlpService); err != nil {
			return err
		}
	}

	return nil
}

func (suite *AITestSuite) testLLMService(ctx context.Context, service interface{}) error {
	llmService, ok := service.(ai.LanguageModelService)
	if !ok {
		return fmt.Errorf("invalid LLM service type")
	}

	start := time.Now()
	testName := "LLM_ProcessQuery"

	// Test basic query processing
	response, err := llmService.ProcessQuery(ctx, "What is artificial intelligence?")
	duration := time.Since(start)

	result := TestResult{
		TestName:  testName,
		Service:   "llm",
		Success:   err == nil,
		Duration:  duration,
		Timestamp: time.Now(),
		Metrics: map[string]interface{}{
			"response_length": 0,
			"tokens":          0,
		},
	}

	if err != nil {
		result.ErrorMessage = err.Error()
		suite.logger.WithError(err).Error("LLM service test failed")
	} else {
		result.Metrics["response_length"] = len(response.Text)
		result.Metrics["tokens"] = response.TokensUsed
		suite.logger.WithFields(logrus.Fields{
			"response_length": len(response.Text),
			"tokens":          response.TokensUsed,
			"duration":        duration,
		}).Info("LLM service test passed")
	}

	suite.addTestResult(result)

	// Record performance metrics
	suite.performanceMonitor.RecordRequest(ctx, "llm", "test-model", duration, int64(result.Metrics["tokens"].(int)), err == nil, false)

	return nil
}

func (suite *AITestSuite) testVoiceService(ctx context.Context, service interface{}) error {
	voiceService, ok := service.(ai.VoiceService)
	if !ok {
		return fmt.Errorf("invalid voice service type")
	}

	start := time.Now()
	testName := "Voice_SpeechToText"

	// Test speech to text with dummy audio data
	dummyAudio := make([]byte, 1024)
	recognition, err := voiceService.SpeechToText(ctx, dummyAudio)
	duration := time.Since(start)

	result := TestResult{
		TestName:  testName,
		Service:   "voice",
		Success:   err == nil,
		Duration:  duration,
		Timestamp: time.Now(),
		Metrics: map[string]interface{}{
			"audio_size": len(dummyAudio),
			"confidence": 0.0,
		},
	}

	if err != nil {
		result.ErrorMessage = err.Error()
	} else {
		result.Metrics["confidence"] = recognition.Confidence
		result.Metrics["text_length"] = len(recognition.Text)
	}

	suite.addTestResult(result)
	suite.performanceMonitor.RecordRequest(ctx, "voice", "whisper", duration, 0, err == nil, false)

	return nil
}

func (suite *AITestSuite) testCVService(ctx context.Context, service interface{}) error {
	cvService, ok := service.(ai.ComputerVisionService)
	if !ok {
		return fmt.Errorf("invalid CV service type")
	}

	start := time.Now()
	testName := "CV_AnalyzeImage"

	// Test image analysis with dummy image data
	dummyImage := make([]byte, 2048)
	analysis, err := cvService.AnalyzeScreen(ctx, dummyImage)
	duration := time.Since(start)

	result := TestResult{
		TestName:  testName,
		Service:   "cv",
		Success:   err == nil,
		Duration:  duration,
		Timestamp: time.Now(),
		Metrics: map[string]interface{}{
			"image_size":    len(dummyImage),
			"objects_found": 0,
		},
	}

	if err != nil {
		result.ErrorMessage = err.Error()
	} else {
		result.Metrics["elements_found"] = len(analysis.Elements)
		result.Metrics["text_regions"] = len(analysis.Text)
		result.Metrics["actions_found"] = len(analysis.Actions)
	}

	suite.addTestResult(result)
	suite.performanceMonitor.RecordRequest(ctx, "cv", "yolo", duration, 0, err == nil, false)

	return nil
}

func (suite *AITestSuite) testNLPService(ctx context.Context, service interface{}) error {
	nlpService, ok := service.(ai.NaturalLanguageService)
	if !ok {
		return fmt.Errorf("invalid NLP service type")
	}

	start := time.Now()
	testName := "NLP_ParseIntent"

	// Test intent analysis
	analysis, err := nlpService.ParseIntent(ctx, "Please open the calculator application")
	duration := time.Since(start)

	result := TestResult{
		TestName:  testName,
		Service:   "nlp",
		Success:   err == nil,
		Duration:  duration,
		Timestamp: time.Now(),
		Metrics: map[string]interface{}{
			"entities_found": 0,
			"confidence":     0.0,
		},
	}

	if err != nil {
		result.ErrorMessage = err.Error()
	} else {
		result.Metrics["entities_found"] = len(analysis.Entities)
		result.Metrics["confidence"] = analysis.Confidence
		result.Metrics["intent"] = analysis.Intent
	}

	suite.addTestResult(result)
	suite.performanceMonitor.RecordRequest(ctx, "nlp", "bert", duration, 0, err == nil, false)

	return nil
}

// Integration Tests

func (suite *AITestSuite) runIntegrationTests(ctx context.Context) error {
	suite.logger.Info("Running integration tests")

	// Test service interactions
	if err := suite.testServiceIntegration(ctx); err != nil {
		return err
	}

	// Test caching integration
	if err := suite.testCachingIntegration(ctx); err != nil {
		return err
	}

	return nil
}

func (suite *AITestSuite) testServiceIntegration(ctx context.Context) error {
	// Test LLM + NLP integration
	llmService, llmExists := suite.services["llm"]
	nlpService, nlpExists := suite.services["nlp"]

	if llmExists && nlpExists {
		start := time.Now()
		testName := "Integration_LLM_NLP"

		// First analyze intent
		nlp := nlpService.(ai.NaturalLanguageService)
		intent, err := nlp.ParseIntent(ctx, "What's the weather like today?")
		if err != nil {
			return err
		}

		// Then generate response based on intent
		llm := llmService.(ai.LanguageModelService)
		response, err := llm.ProcessQuery(ctx, fmt.Sprintf("Generate a response for intent: %s", intent.Intent))
		duration := time.Since(start)

		result := TestResult{
			TestName:  testName,
			Service:   "integration",
			Success:   err == nil,
			Duration:  duration,
			Timestamp: time.Now(),
			Metrics: map[string]interface{}{
				"intent":          intent.Intent,
				"response_length": len(response.Text),
			},
		}

		if err != nil {
			result.ErrorMessage = err.Error()
		}

		suite.addTestResult(result)
	}

	return nil
}

func (suite *AITestSuite) testCachingIntegration(ctx context.Context) error {
	// Test cache hit/miss scenarios
	if llmService, exists := suite.services["llm"]; exists {
		llm := llmService.(ai.LanguageModelService)

		// First request (cache miss)
		start := time.Now()
		_, err := llm.ProcessQuery(ctx, "Test caching query")
		firstDuration := time.Since(start)

		if err != nil {
			return err
		}

		// Second request (should be cache hit)
		start = time.Now()
		_, err = llm.ProcessQuery(ctx, "Test caching query")
		secondDuration := time.Since(start)

		if err != nil {
			return err
		}

		result := TestResult{
			TestName:  "Cache_Performance",
			Service:   "cache",
			Success:   secondDuration < firstDuration,
			Duration:  secondDuration,
			Timestamp: time.Now(),
			Metrics: map[string]interface{}{
				"first_request_ms":  firstDuration.Milliseconds(),
				"second_request_ms": secondDuration.Milliseconds(),
				"cache_speedup":     float64(firstDuration.Milliseconds()) / float64(secondDuration.Milliseconds()),
			},
		}

		suite.addTestResult(result)
	}

	return nil
}

// Performance Tests

func (suite *AITestSuite) runPerformanceTests(ctx context.Context, config TestConfig) error {
	suite.logger.Info("Running performance tests")

	// Test response time under normal load
	if err := suite.testResponseTime(ctx, config); err != nil {
		return err
	}

	// Test throughput
	if err := suite.testThroughput(ctx, config); err != nil {
		return err
	}

	return nil
}

func (suite *AITestSuite) testResponseTime(ctx context.Context, config TestConfig) error {
	if llmService, exists := suite.services["llm"]; exists {
		llm := llmService.(ai.LanguageModelService)

		var totalDuration time.Duration
		successCount := 0
		testCount := 10

		for i := 0; i < testCount; i++ {
			start := time.Now()
			_, err := llm.ProcessQuery(ctx, fmt.Sprintf("Performance test query %d", i))
			duration := time.Since(start)
			totalDuration += duration

			if err == nil {
				successCount++
			}
		}

		avgDuration := totalDuration / time.Duration(testCount)
		successRate := float64(successCount) / float64(testCount)

		result := TestResult{
			TestName:  "Performance_ResponseTime",
			Service:   "llm",
			Success:   avgDuration < config.TimeoutPerRequest && successRate > 0.95,
			Duration:  avgDuration,
			Timestamp: time.Now(),
			Metrics: map[string]interface{}{
				"average_duration_ms": avgDuration.Milliseconds(),
				"success_rate":        successRate,
				"test_count":          testCount,
			},
		}

		suite.addTestResult(result)
	}

	return nil
}

func (suite *AITestSuite) testThroughput(ctx context.Context, config TestConfig) error {
	// Implement throughput testing
	suite.logger.Info("Testing throughput")
	return nil
}

// Load and Stress Tests

func (suite *AITestSuite) runLoadTests(ctx context.Context, config TestConfig) error {
	suite.logger.WithField("concurrent_users", config.ConcurrentUsers).Info("Running load tests")

	if llmService, exists := suite.services["llm"]; exists {
		llm := llmService.(ai.LanguageModelService)

		var wg sync.WaitGroup
		results := make(chan TestResult, config.ConcurrentUsers)

		// Launch concurrent users
		for i := 0; i < config.ConcurrentUsers; i++ {
			wg.Add(1)
			go func(userID int) {
				defer wg.Done()

				start := time.Now()
				_, err := llm.ProcessQuery(ctx, fmt.Sprintf("Load test query from user %d", userID))
				duration := time.Since(start)

				results <- TestResult{
					TestName:  fmt.Sprintf("LoadTest_User_%d", userID),
					Service:   "llm",
					Success:   err == nil,
					Duration:  duration,
					Timestamp: time.Now(),
					Metrics: map[string]interface{}{
						"user_id": userID,
					},
				}
			}(i)
		}

		// Wait for all users to complete
		go func() {
			wg.Wait()
			close(results)
		}()

		// Collect results
		successCount := 0
		totalDuration := time.Duration(0)

		for result := range results {
			suite.addTestResult(result)
			if result.Success {
				successCount++
			}
			totalDuration += result.Duration
		}

		suite.logger.WithFields(logrus.Fields{
			"concurrent_users": config.ConcurrentUsers,
			"success_rate":     float64(successCount) / float64(config.ConcurrentUsers),
			"avg_duration":     totalDuration / time.Duration(config.ConcurrentUsers),
		}).Info("Load test completed")
	}

	return nil
}

func (suite *AITestSuite) runStressTests(ctx context.Context, config TestConfig) error {
	suite.logger.Info("Running stress tests with high load")

	// Stress test with 2x the normal concurrent users
	stressConfig := config
	stressConfig.ConcurrentUsers *= 2

	return suite.runLoadTests(ctx, stressConfig)
}

// Helper methods

func (suite *AITestSuite) addTestResult(result TestResult) {
	suite.mu.Lock()
	defer suite.mu.Unlock()
	suite.testResults = append(suite.testResults, result)
}

// GetTestResults returns all test results
func (suite *AITestSuite) GetTestResults() []TestResult {
	suite.mu.RLock()
	defer suite.mu.RUnlock()

	results := make([]TestResult, len(suite.testResults))
	copy(results, suite.testResults)
	return results
}

// GenerateReport generates a comprehensive test report
func (suite *AITestSuite) GenerateReport() map[string]interface{} {
	results := suite.GetTestResults()

	totalTests := len(results)
	successfulTests := 0
	totalDuration := time.Duration(0)

	serviceResults := make(map[string][]TestResult)

	for _, result := range results {
		if result.Success {
			successfulTests++
		}
		totalDuration += result.Duration

		serviceResults[result.Service] = append(serviceResults[result.Service], result)
	}

	return map[string]interface{}{
		"total_tests":         totalTests,
		"successful_tests":    successfulTests,
		"success_rate":        float64(successfulTests) / float64(totalTests),
		"total_duration":      totalDuration,
		"average_duration":    totalDuration / time.Duration(totalTests),
		"service_results":     serviceResults,
		"performance_metrics": suite.performanceMonitor.GetAllServiceMetrics(),
	}
}
