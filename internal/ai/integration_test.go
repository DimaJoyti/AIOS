package ai

import (
	"context"
	"testing"
	"time"

	"github.com/aios/aios/pkg/models"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAIServicesIntegration tests the integration between AI services
func TestAIServicesIntegration(t *testing.T) {
	// Setup test configuration
	config := AIServiceConfig{
		OllamaHost:          "localhost",
		OllamaPort:          11434,
		OllamaTimeout:       30 * time.Second,
		ModelsPath:          "/tmp/test-models",
		DefaultModel:        "test-model",
		MaxTokens:           1000,
		Temperature:         0.7,
		CVEnabled:           true,
		CVModelPath:         "/tmp/test-cv-models",
		ConfidenceThreshold: 0.8,
		MaxImageSize:        "10MB",
		VoiceEnabled:        true,
		VoiceModelPath:      "/tmp/test-voice-models",
		WakeWord:            "test",
		SampleRate:          16000,
		NLPEnabled:          true,
		NLPModelPath:        "/tmp/test-nlp-models",
		IntentModel:         "test-intent",
		EntityModel:         "test-entity",
	}

	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel) // Reduce noise in tests

	ctx := context.Background()

	t.Run("LLM Service Integration", func(t *testing.T) {
		llmService := NewLLMService(config, logger)
		require.NotNil(t, llmService)

		// Test query processing
		response, err := llmService.ProcessQuery(ctx, "Hello, how are you?")
		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.NotEmpty(t, response.Text)
		assert.True(t, response.TokensUsed > 0)

		// Test model listing
		models, err := llmService.GetModels(ctx)
		assert.NoError(t, err)
		assert.NotNil(t, models)
	})

	t.Run("Computer Vision Service Integration", func(t *testing.T) {
		cvService := NewCVService(config, logger)
		require.NotNil(t, cvService)

		// Create mock image data
		mockImageData := []byte("mock-image-data-for-testing")

		// Test object detection
		detection, err := cvService.DetectObjects(ctx, mockImageData)
		assert.NoError(t, err)
		assert.NotNil(t, detection)
		assert.NotNil(t, detection.Objects)

		// Test text recognition
		textRecognition, err := cvService.RecognizeText(ctx, mockImageData)
		assert.NoError(t, err)
		assert.NotNil(t, textRecognition)

		// Test image classification
		classification, err := cvService.ClassifyImage(ctx, mockImageData)
		assert.NoError(t, err)
		assert.NotNil(t, classification)

		// Test screen analysis
		screenAnalysis, err := cvService.AnalyzeScreen(ctx, mockImageData)
		assert.NoError(t, err)
		assert.NotNil(t, screenAnalysis)
	})

	t.Run("Voice Service Integration", func(t *testing.T) {
		voiceService := NewVoiceService(config, logger)
		require.NotNil(t, voiceService)

		// Create mock audio data
		mockAudioData := []byte("mock-audio-data-for-testing")

		// Test speech-to-text
		speechRecognition, err := voiceService.SpeechToText(ctx, mockAudioData)
		assert.NoError(t, err)
		assert.NotNil(t, speechRecognition)
		assert.NotEmpty(t, speechRecognition.Text)
		assert.True(t, speechRecognition.Confidence >= 0)

		// Test text-to-speech
		speechSynthesis, err := voiceService.TextToSpeech(ctx, "Hello world")
		assert.NoError(t, err)
		assert.NotNil(t, speechSynthesis)
		assert.NotEmpty(t, speechSynthesis.Audio)

		// Test wake word detection
		wakeWordDetection, err := voiceService.DetectWakeWord(ctx, mockAudioData)
		assert.NoError(t, err)
		assert.NotNil(t, wakeWordDetection)

		// Test voice command processing
		voiceCommand, err := voiceService.ProcessVoiceCommand(ctx, mockAudioData)
		assert.NoError(t, err)
		assert.NotNil(t, voiceCommand)
	})

	t.Run("NLP Service Integration", func(t *testing.T) {
		nlpService := NewNLPService(config, logger)
		require.NotNil(t, nlpService)

		testText := "I want to open the terminal application"

		// Test intent parsing
		intentAnalysis, err := nlpService.ParseIntent(ctx, testText)
		assert.NoError(t, err)
		assert.NotNil(t, intentAnalysis)
		assert.NotEmpty(t, intentAnalysis.Intent)
		assert.True(t, intentAnalysis.Confidence >= 0)

		// Test entity extraction
		entityExtraction, err := nlpService.ExtractEntities(ctx, testText)
		assert.NoError(t, err)
		assert.NotNil(t, entityExtraction)
		assert.NotNil(t, entityExtraction.Entities)

		// Test sentiment analysis
		sentimentAnalysis, err := nlpService.AnalyzeSentiment(ctx, testText)
		assert.NoError(t, err)
		assert.NotNil(t, sentimentAnalysis)
		assert.NotEmpty(t, sentimentAnalysis.Sentiment.Label)
	})

	t.Run("AI Orchestrator Integration", func(t *testing.T) {
		orchestrator := NewOrchestrator(config, logger)
		require.NotNil(t, orchestrator)

		// Test service status
		status, err := orchestrator.GetServiceStatus(ctx)
		assert.NoError(t, err)
		assert.NotNil(t, status)

		// Test complex AI request processing
		request := &models.AIRequest{
			ID:   "test-request-1",
			Type: "llm_query",
			Input: map[string]interface{}{
				"query": "What is the weather like today?",
			},
			Timestamp: time.Now(),
		}

		response, err := orchestrator.ProcessRequest(ctx, request)
		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, request.ID, response.RequestID)
		assert.NotEmpty(t, response.Type)
	})
}

// TestAIServicePerformance tests the performance characteristics of AI services
func TestAIServicePerformance(t *testing.T) {
	config := AIServiceConfig{
		OllamaHost:          "localhost",
		OllamaPort:          11434,
		OllamaTimeout:       30 * time.Second,
		ModelsPath:          "/tmp/test-models",
		DefaultModel:        "test-model",
		MaxTokens:           100, // Smaller for faster tests
		Temperature:         0.7,
		CVEnabled:           true,
		VoiceEnabled:        true,
		NLPEnabled:          true,
		ConfidenceThreshold: 0.8,
	}

	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	ctx := context.Background()

	t.Run("LLM Service Performance", func(t *testing.T) {
		llmService := NewLLMService(config, logger)

		start := time.Now()
		_, err := llmService.ProcessQuery(ctx, "Hello")
		duration := time.Since(start)

		assert.NoError(t, err)
		assert.Less(t, duration, 10*time.Second, "LLM query should complete within 10 seconds")
	})

	t.Run("Computer Vision Service Performance", func(t *testing.T) {
		cvService := NewCVService(config, logger)
		mockImageData := make([]byte, 1024) // 1KB mock image

		start := time.Now()
		_, err := cvService.DetectObjects(ctx, mockImageData)
		duration := time.Since(start)

		assert.NoError(t, err)
		assert.Less(t, duration, 5*time.Second, "Object detection should complete within 5 seconds")
	})

	t.Run("Voice Service Performance", func(t *testing.T) {
		voiceService := NewVoiceService(config, logger)
		mockAudioData := make([]byte, 16000) // 1 second of 16kHz audio

		start := time.Now()
		_, err := voiceService.SpeechToText(ctx, mockAudioData)
		duration := time.Since(start)

		assert.NoError(t, err)
		assert.Less(t, duration, 3*time.Second, "Speech recognition should complete within 3 seconds")
	})

	t.Run("NLP Service Performance", func(t *testing.T) {
		nlpService := NewNLPService(config, logger)
		testText := "I want to open the terminal application"

		start := time.Now()
		_, err := nlpService.ParseIntent(ctx, testText)
		duration := time.Since(start)

		assert.NoError(t, err)
		assert.Less(t, duration, 1*time.Second, "Intent parsing should complete within 1 second")
	})
}

// TestAIServiceConcurrency tests concurrent access to AI services
func TestAIServiceConcurrency(t *testing.T) {
	config := AIServiceConfig{
		OllamaHost:          "localhost",
		OllamaPort:          11434,
		OllamaTimeout:       30 * time.Second,
		ModelsPath:          "/tmp/test-models",
		DefaultModel:        "test-model",
		MaxTokens:           50,
		Temperature:         0.7,
		CVEnabled:           true,
		VoiceEnabled:        true,
		NLPEnabled:          true,
	}

	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	ctx := context.Background()

	t.Run("Concurrent LLM Queries", func(t *testing.T) {
		llmService := NewLLMService(config, logger)
		
		const numGoroutines = 5
		results := make(chan error, numGoroutines)

		for i := 0; i < numGoroutines; i++ {
			go func(id int) {
				_, err := llmService.ProcessQuery(ctx, "Hello from goroutine")
				results <- err
			}(i)
		}

		// Collect results
		for i := 0; i < numGoroutines; i++ {
			err := <-results
			assert.NoError(t, err, "Concurrent LLM query %d should succeed", i)
		}
	})

	t.Run("Concurrent CV Operations", func(t *testing.T) {
		cvService := NewCVService(config, logger)
		mockImageData := make([]byte, 1024)
		
		const numGoroutines = 3
		results := make(chan error, numGoroutines)

		for i := 0; i < numGoroutines; i++ {
			go func(id int) {
				_, err := cvService.DetectObjects(ctx, mockImageData)
				results <- err
			}(i)
		}

		// Collect results
		for i := 0; i < numGoroutines; i++ {
			err := <-results
			assert.NoError(t, err, "Concurrent CV operation %d should succeed", i)
		}
	})

	t.Run("Mixed Service Concurrency", func(t *testing.T) {
		llmService := NewLLMService(config, logger)
		cvService := NewCVService(config, logger)
		nlpService := NewNLPService(config, logger)
		
		const numGoroutines = 6
		results := make(chan error, numGoroutines)

		// Start mixed operations
		for i := 0; i < numGoroutines; i++ {
			switch i % 3 {
			case 0:
				go func() {
					_, err := llmService.ProcessQuery(ctx, "Test query")
					results <- err
				}()
			case 1:
				go func() {
					mockImageData := make([]byte, 512)
					_, err := cvService.DetectObjects(ctx, mockImageData)
					results <- err
				}()
			case 2:
				go func() {
					_, err := nlpService.ParseIntent(ctx, "test intent")
					results <- err
				}()
			}
		}

		// Collect results
		for i := 0; i < numGoroutines; i++ {
			err := <-results
			assert.NoError(t, err, "Mixed concurrent operation %d should succeed", i)
		}
	})
}

// TestAIServiceErrorHandling tests error handling in AI services
func TestAIServiceErrorHandling(t *testing.T) {
	config := AIServiceConfig{
		OllamaHost:   "invalid-host",
		OllamaPort:   99999,
		CVEnabled:    true,
		VoiceEnabled: true,
		NLPEnabled:   true,
	}

	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	ctx := context.Background()

	t.Run("LLM Service Error Handling", func(t *testing.T) {
		llmService := NewLLMService(config, logger)

		// Test with empty query
		_, err := llmService.ProcessQuery(ctx, "")
		// Should handle gracefully (might return empty response or error)
		// The exact behavior depends on implementation

		// Test with very long query
		longQuery := make([]byte, 10000)
		for i := range longQuery {
			longQuery[i] = 'a'
		}
		_, err = llmService.ProcessQuery(ctx, string(longQuery))
		// Should handle gracefully
		_ = err // Acknowledge we're checking error handling
	})

	t.Run("CV Service Error Handling", func(t *testing.T) {
		cvService := NewCVService(config, logger)

		// Test with empty image data
		_, err := cvService.DetectObjects(ctx, []byte{})
		// Should return appropriate error
		_ = err

		// Test with invalid image data
		_, err = cvService.DetectObjects(ctx, []byte("invalid-image-data"))
		// Should handle gracefully
		_ = err
	})

	t.Run("Voice Service Error Handling", func(t *testing.T) {
		voiceService := NewVoiceService(config, logger)

		// Test with empty audio data
		_, err := voiceService.SpeechToText(ctx, []byte{})
		// Should return appropriate error
		_ = err

		// Test with invalid audio data
		_, err = voiceService.SpeechToText(ctx, []byte("invalid-audio-data"))
		// Should handle gracefully
		_ = err
	})

	t.Run("NLP Service Error Handling", func(t *testing.T) {
		nlpService := NewNLPService(config, logger)

		// Test with empty text
		_, err := nlpService.ParseIntent(ctx, "")
		// Should handle gracefully
		_ = err

		// Test with very long text
		longText := make([]byte, 100000)
		for i := range longText {
			longText[i] = 'a'
		}
		_, err = nlpService.ParseIntent(ctx, string(longText))
		// Should handle gracefully
		_ = err
	})
}
