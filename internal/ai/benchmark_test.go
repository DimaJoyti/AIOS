package ai

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/aios/aios/pkg/models"
	"github.com/sirupsen/logrus"
)

// BenchmarkAIServices provides comprehensive benchmarks for AI services
func BenchmarkAIServices(b *testing.B) {
	config := AIServiceConfig{
		OllamaHost:          "localhost",
		OllamaPort:          11434,
		OllamaTimeout:       30 * time.Second,
		ModelsPath:          "/tmp/benchmark-models",
		DefaultModel:        "test-model",
		MaxTokens:           100,
		Temperature:         0.7,
		CVEnabled:           true,
		CVModelPath:         "/tmp/benchmark-cv-models",
		ConfidenceThreshold: 0.8,
		VoiceEnabled:        true,
		VoiceModelPath:      "/tmp/benchmark-voice-models",
		SampleRate:          16000,
		NLPEnabled:          true,
		NLPModelPath:        "/tmp/benchmark-nlp-models",
	}

	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	ctx := context.Background()

	b.Run("LLM Service", func(b *testing.B) {
		llmService := NewLLMService(config, logger)

		b.Run("ProcessQuery", func(b *testing.B) {
			queries := []string{
				"Hello, how are you?",
				"What is the weather like today?",
				"Explain quantum computing in simple terms",
				"Write a Python function to sort a list",
				"What are the benefits of renewable energy?",
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				query := queries[i%len(queries)]
				_, err := llmService.ProcessQuery(ctx, query)
				if err != nil {
					b.Errorf("LLM query failed: %v", err)
				}
			}
		})

		b.Run("ProcessQueryConcurrent", func(b *testing.B) {
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					_, err := llmService.ProcessQuery(ctx, "Concurrent test query")
					if err != nil {
						b.Errorf("Concurrent LLM query failed: %v", err)
					}
				}
			})
		})

		b.Run("GenerateCode", func(b *testing.B) {
			prompts := []string{
				"Write a function to calculate fibonacci numbers",
				"Create a REST API endpoint in Go",
				"Implement a binary search algorithm",
				"Write a SQL query to find top customers",
				"Create a React component for a login form",
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				prompt := prompts[i%len(prompts)]
				_, err := llmService.GenerateCode(ctx, prompt)
				if err != nil {
					b.Errorf("Code generation failed: %v", err)
				}
			}
		})

		b.Run("AnalyzeText", func(b *testing.B) {
			texts := []string{
				"This is a sample text for analysis.",
				"The quick brown fox jumps over the lazy dog.",
				"Artificial intelligence is transforming the world.",
				"Climate change is a global challenge requiring immediate action.",
				"Technology has revolutionized how we communicate and work.",
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				text := texts[i%len(texts)]
				_, err := llmService.AnalyzeText(ctx, text)
				if err != nil {
					b.Errorf("Text analysis failed: %v", err)
				}
			}
		})

		b.Run("GetModels", func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err := llmService.GetModels(ctx)
				if err != nil {
					b.Errorf("Get models failed: %v", err)
				}
			}
		})
	})

	b.Run("Computer Vision Service", func(b *testing.B) {
		cvService := NewCVService(config, logger)

		// Create test image data of different sizes
		smallImage := make([]byte, 1024)    // 1KB
		mediumImage := make([]byte, 10240)  // 10KB
		largeImage := make([]byte, 102400)  // 100KB

		b.Run("DetectObjects", func(b *testing.B) {
			images := [][]byte{smallImage, mediumImage, largeImage}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				image := images[i%len(images)]
				_, err := cvService.DetectObjects(ctx, image)
				if err != nil {
					b.Errorf("Object detection failed: %v", err)
				}
			}
		})

		b.Run("DetectObjectsConcurrent", func(b *testing.B) {
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					_, err := cvService.DetectObjects(ctx, mediumImage)
					if err != nil {
						b.Errorf("Concurrent object detection failed: %v", err)
					}
				}
			})
		})

		b.Run("RecognizeText", func(b *testing.B) {
			images := [][]byte{smallImage, mediumImage, largeImage}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				image := images[i%len(images)]
				_, err := cvService.RecognizeText(ctx, image)
				if err != nil {
					b.Errorf("Text recognition failed: %v", err)
				}
			}
		})

		b.Run("ClassifyImage", func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err := cvService.ClassifyImage(ctx, mediumImage)
				if err != nil {
					b.Errorf("Image classification failed: %v", err)
				}
			}
		})

		b.Run("AnalyzeScreen", func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err := cvService.AnalyzeScreen(ctx, largeImage)
				if err != nil {
					b.Errorf("Screen analysis failed: %v", err)
				}
			}
		})
	})

	b.Run("Voice Service", func(b *testing.B) {
		voiceService := NewVoiceService(config, logger)

		// Create test audio data of different durations
		shortAudio := make([]byte, 16000)   // 1 second at 16kHz
		mediumAudio := make([]byte, 48000)  // 3 seconds at 16kHz
		longAudio := make([]byte, 160000)   // 10 seconds at 16kHz

		b.Run("SpeechToText", func(b *testing.B) {
			audios := [][]byte{shortAudio, mediumAudio, longAudio}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				audio := audios[i%len(audios)]
				_, err := voiceService.SpeechToText(ctx, audio)
				if err != nil {
					b.Errorf("Speech-to-text failed: %v", err)
				}
			}
		})

		b.Run("SpeechToTextConcurrent", func(b *testing.B) {
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					_, err := voiceService.SpeechToText(ctx, mediumAudio)
					if err != nil {
						b.Errorf("Concurrent speech-to-text failed: %v", err)
					}
				}
			})
		})

		b.Run("TextToSpeech", func(b *testing.B) {
			texts := []string{
				"Hello",
				"Hello, how are you today?",
				"This is a longer text that will take more time to synthesize into speech.",
				"The quick brown fox jumps over the lazy dog. This sentence contains every letter of the alphabet.",
				"Artificial intelligence and machine learning are transforming the way we interact with technology.",
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				text := texts[i%len(texts)]
				_, err := voiceService.TextToSpeech(ctx, text)
				if err != nil {
					b.Errorf("Text-to-speech failed: %v", err)
				}
			}
		})

		b.Run("DetectWakeWord", func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err := voiceService.DetectWakeWord(ctx, shortAudio)
				if err != nil {
					b.Errorf("Wake word detection failed: %v", err)
				}
			}
		})

		b.Run("ProcessVoiceCommand", func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err := voiceService.ProcessVoiceCommand(ctx, mediumAudio)
				if err != nil {
					b.Errorf("Voice command processing failed: %v", err)
				}
			}
		})
	})

	b.Run("NLP Service", func(b *testing.B) {
		nlpService := NewNLPService(config, logger)

		texts := []string{
			"Open the terminal",
			"I want to search for documents",
			"Please close the application",
			"Can you help me with this task?",
			"Set the volume to 50 percent",
			"What time is it right now?",
			"Show me the weather forecast",
			"I'm feeling happy today",
			"This is a terrible experience",
			"The product quality is excellent",
		}

		b.Run("ParseIntent", func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				text := texts[i%len(texts)]
				_, err := nlpService.ParseIntent(ctx, text)
				if err != nil {
					b.Errorf("Intent parsing failed: %v", err)
				}
			}
		})

		b.Run("ParseIntentConcurrent", func(b *testing.B) {
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					text := texts[0] // Use first text for consistency
					_, err := nlpService.ParseIntent(ctx, text)
					if err != nil {
						b.Errorf("Concurrent intent parsing failed: %v", err)
					}
				}
			})
		})

		b.Run("ExtractEntities", func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				text := texts[i%len(texts)]
				_, err := nlpService.ExtractEntities(ctx, text)
				if err != nil {
					b.Errorf("Entity extraction failed: %v", err)
				}
			}
		})

		b.Run("AnalyzeSentiment", func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				text := texts[i%len(texts)]
				_, err := nlpService.AnalyzeSentiment(ctx, text)
				if err != nil {
					b.Errorf("Sentiment analysis failed: %v", err)
				}
			}
		})
	})

	b.Run("AI Orchestrator", func(b *testing.B) {
		orchestrator := NewOrchestrator(config, logger)

		b.Run("GetServiceStatus", func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err := orchestrator.GetServiceStatus(ctx)
				if err != nil {
					b.Errorf("Get service status failed: %v", err)
				}
			}
		})

		b.Run("ProcessRequest", func(b *testing.B) {
			requests := []*models.AIRequest{
				{
					ID:   "bench-1",
					Type: "llm_query",
					Input: map[string]interface{}{"query": "Hello world"},
				},
				{
					ID:   "bench-2",
					Type: "nlp_intent",
					Input: map[string]interface{}{"text": "Open terminal"},
				},
				{
					ID:   "bench-3",
					Type: "cv_analyze",
					Input: map[string]interface{}{"image": make([]byte, 1024)},
				},
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				request := requests[i%len(requests)]
				request.ID = fmt.Sprintf("bench-%d", i)
				request.Timestamp = time.Now()

				_, err := orchestrator.ProcessRequest(ctx, request)
				if err != nil {
					b.Errorf("Process request failed: %v", err)
				}
			}
		})

		b.Run("ProcessRequestConcurrent", func(b *testing.B) {
			b.RunParallel(func(pb *testing.PB) {
				i := 0
				for pb.Next() {
					request := &models.AIRequest{
						ID:   fmt.Sprintf("concurrent-bench-%d", i),
						Type: "llm_query",
						Input: map[string]interface{}{"query": "Concurrent test"},
						Timestamp: time.Now(),
					}

					_, err := orchestrator.ProcessRequest(ctx, request)
					if err != nil {
						b.Errorf("Concurrent process request failed: %v", err)
					}
					i++
				}
			})
		})
	})
}

// BenchmarkMemoryUsage benchmarks memory usage of AI services
func BenchmarkMemoryUsage(b *testing.B) {
	config := AIServiceConfig{
		OllamaHost:   "localhost",
		OllamaPort:   11434,
		CVEnabled:    true,
		VoiceEnabled: true,
		NLPEnabled:   true,
		MaxTokens:    50, // Smaller to reduce memory usage
	}

	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	ctx := context.Background()

	b.Run("LLM Service Memory", func(b *testing.B) {
		b.ReportAllocs()
		llmService := NewLLMService(config, logger)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := llmService.ProcessQuery(ctx, "Memory test query")
			if err != nil {
				b.Errorf("LLM query failed: %v", err)
			}
		}
	})

	b.Run("CV Service Memory", func(b *testing.B) {
		b.ReportAllocs()
		cvService := NewCVService(config, logger)
		testImage := make([]byte, 1024)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := cvService.DetectObjects(ctx, testImage)
			if err != nil {
				b.Errorf("Object detection failed: %v", err)
			}
		}
	})

	b.Run("Voice Service Memory", func(b *testing.B) {
		b.ReportAllocs()
		voiceService := NewVoiceService(config, logger)
		testAudio := make([]byte, 16000)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := voiceService.SpeechToText(ctx, testAudio)
			if err != nil {
				b.Errorf("Speech-to-text failed: %v", err)
			}
		}
	})

	b.Run("NLP Service Memory", func(b *testing.B) {
		b.ReportAllocs()
		nlpService := NewNLPService(config, logger)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := nlpService.ParseIntent(ctx, "Memory test intent")
			if err != nil {
				b.Errorf("Intent parsing failed: %v", err)
			}
		}
	})
}
