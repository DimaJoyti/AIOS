package performance

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/aios/aios/internal/ai/services"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

// BenchmarkModelManager tests the performance of the model manager
func BenchmarkModelManager(b *testing.B) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	mm := services.NewModelManager(logger)

	// Register test models
	for i := 0; i < 100; i++ {
		model := &services.AIModel{
			ID:       fmt.Sprintf("model-%d", i),
			Name:     fmt.Sprintf("Model %d", i),
			Provider: "test-provider",
			Type:     "text",
			Status:   "active",
		}
		require.NoError(b, mm.RegisterModel(model))
	}

	b.ResetTimer()

	b.Run("GetModel", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				modelID := fmt.Sprintf("model-%d", i%100)
				_, err := mm.GetModel(modelID)
				if err != nil {
					b.Fatal(err)
				}
				i++
			}
		})
	})

	b.Run("ListModels", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				models := mm.ListModels()
				if len(models) != 100 {
					b.Fatalf("Expected 100 models, got %d", len(models))
				}
			}
		})
	})

	b.Run("GetModelsByType", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				models := mm.GetModelsByType("text")
				if len(models) != 100 {
					b.Fatalf("Expected 100 text models, got %d", len(models))
				}
			}
		})
	})
}

// BenchmarkModelCache tests the performance of the model cache
func BenchmarkModelCache(b *testing.B) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	cache := services.NewModelCache(logger)

	// Pre-populate cache with test data
	for i := 0; i < 1000; i++ {
		request := &services.TextGenerationRequest{
			ModelID: "test-model",
			Prompt:  fmt.Sprintf("Test prompt %d", i),
		}

		response := &services.TextGenerationResponse{
			Text:     fmt.Sprintf("Response %d", i),
			ModelID:  "test-model",
			Provider: "test-provider",
		}

		cache.SetTextResponse(request, response)
	}

	b.ResetTimer()

	b.Run("CacheHit", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				request := &services.TextGenerationRequest{
					ModelID: "test-model",
					Prompt:  fmt.Sprintf("Test prompt %d", i%1000),
				}

				response := cache.GetTextResponse(request)
				if response == nil {
					b.Fatal("Expected cache hit")
				}
				i++
			}
		})
	})

	b.Run("CacheMiss", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				request := &services.TextGenerationRequest{
					ModelID: "test-model",
					Prompt:  fmt.Sprintf("Non-existent prompt %d", i+10000),
				}

				response := cache.GetTextResponse(request)
				if response != nil {
					b.Fatal("Expected cache miss")
				}
				i++
			}
		})
	})

	b.Run("CacheSet", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				request := &services.TextGenerationRequest{
					ModelID: "test-model",
					Prompt:  fmt.Sprintf("New prompt %d", i),
				}

				response := &services.TextGenerationResponse{
					Text:     fmt.Sprintf("New response %d", i),
					ModelID:  "test-model",
					Provider: "test-provider",
				}

				cache.SetTextResponse(request, response)
				i++
			}
		})
	})
}

// BenchmarkPromptManager tests the performance of the prompt manager
func BenchmarkPromptManager(b *testing.B) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	pm := services.NewPromptManager(logger)

	// Register test templates
	for i := 0; i < 100; i++ {
		template := &services.PromptTemplate{
			ID:       fmt.Sprintf("template-%d", i),
			Name:     fmt.Sprintf("Template %d", i),
			Template: "Hello {{.name}}, you are {{.age}} years old!",
			Variables: []services.PromptVariable{
				{
					Name:     "name",
					Type:     "string",
					Required: true,
				},
				{
					Name:     "age",
					Type:     "number",
					Required: true,
				},
			},
		}
		require.NoError(b, pm.CreateTemplate(template))
	}

	b.ResetTimer()

	b.Run("GetTemplate", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				templateID := fmt.Sprintf("template-%d", i%100)
				_, err := pm.GetTemplate(templateID)
				if err != nil {
					b.Fatal(err)
				}
				i++
			}
		})
	})

	b.Run("RenderTemplate", func(b *testing.B) {
		ctx := context.Background()
		variables := map[string]interface{}{
			"name": "Alice",
			"age":  30,
		}

		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				templateID := fmt.Sprintf("template-%d", i%100)
				_, err := pm.RenderTemplate(ctx, templateID, variables)
				if err != nil {
					b.Fatal(err)
				}
				i++
			}
		})
	})

	b.Run("ListTemplates", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				templates := pm.ListTemplates()
				if len(templates) != 100 {
					b.Fatalf("Expected 100 templates, got %d", len(templates))
				}
			}
		})
	})
}

// BenchmarkConcurrentOperations tests concurrent access patterns
func BenchmarkConcurrentOperations(b *testing.B) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	mm := services.NewModelManager(logger)
	cache := services.NewModelCache(logger)
	pm := services.NewPromptManager(logger)

	// Setup test data
	for i := 0; i < 10; i++ {
		model := &services.AIModel{
			ID:       fmt.Sprintf("model-%d", i),
			Name:     fmt.Sprintf("Model %d", i),
			Provider: "test-provider",
			Type:     "text",
			Status:   "active",
		}
		require.NoError(b, mm.RegisterModel(model))

		template := &services.PromptTemplate{
			ID:       fmt.Sprintf("template-%d", i),
			Name:     fmt.Sprintf("Template %d", i),
			Template: "Test {{.value}}",
			Variables: []services.PromptVariable{
				{
					Name:     "value",
					Type:     "string",
					Required: true,
				},
			},
		}
		require.NoError(b, pm.CreateTemplate(template))
	}

	b.ResetTimer()

	b.Run("MixedOperations", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				switch i % 4 {
				case 0:
					// Model operation
					modelID := fmt.Sprintf("model-%d", i%10)
					mm.GetModel(modelID)
				case 1:
					// Cache operation
					request := &services.TextGenerationRequest{
						ModelID: "test-model",
						Prompt:  fmt.Sprintf("Prompt %d", i),
					}
					cache.GetTextResponse(request)
				case 2:
					// Template operation
					templateID := fmt.Sprintf("template-%d", i%10)
					pm.GetTemplate(templateID)
				case 3:
					// List operation
					mm.ListModels()
				}
				i++
			}
		})
	})
}

// LoadTest simulates realistic load patterns
func BenchmarkLoadTest(b *testing.B) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	mm := services.NewModelManager(logger)
	cache := services.NewModelCache(logger)

	// Setup realistic test data
	for i := 0; i < 50; i++ {
		model := &services.AIModel{
			ID:       fmt.Sprintf("model-%d", i),
			Name:     fmt.Sprintf("Model %d", i),
			Provider: "test-provider",
			Type:     "text",
			Status:   "active",
		}
		require.NoError(b, mm.RegisterModel(model))
	}

	b.ResetTimer()

	b.Run("RealisticLoad", func(b *testing.B) {
		const numWorkers = 100
		const requestsPerWorker = 1000

		var wg sync.WaitGroup
		start := time.Now()

		for w := 0; w < numWorkers; w++ {
			wg.Add(1)
			go func(workerID int) {
				defer wg.Done()

				for i := 0; i < requestsPerWorker; i++ {
					// Simulate realistic operations

					// 60% cache lookups
					if i%10 < 6 {
						request := &services.TextGenerationRequest{
							ModelID: fmt.Sprintf("model-%d", i%50),
							Prompt:  fmt.Sprintf("Worker %d request %d", workerID, i),
						}
						cache.GetTextResponse(request)
					}

					// 30% model lookups
					if i%10 >= 6 && i%10 < 9 {
						modelID := fmt.Sprintf("model-%d", i%50)
						mm.GetModel(modelID)
					}

					// 10% list operations
					if i%10 == 9 {
						mm.ListModels()
					}
				}
			}(w)
		}

		wg.Wait()
		duration := time.Since(start)

		totalRequests := numWorkers * requestsPerWorker
		requestsPerSecond := float64(totalRequests) / duration.Seconds()

		b.Logf("Processed %d requests in %v (%.2f req/s)",
			totalRequests, duration, requestsPerSecond)
	})
}

// BenchmarkMemoryUsage tests memory efficiency
func BenchmarkMemoryUsage(b *testing.B) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	b.Run("CacheMemoryUsage", func(b *testing.B) {
		cache := services.NewModelCache(logger)

		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			request := &services.TextGenerationRequest{
				ModelID: "test-model",
				Prompt:  fmt.Sprintf("Memory test prompt %d", i),
			}

			response := &services.TextGenerationResponse{
				Text:     fmt.Sprintf("Memory test response %d with some longer content to test memory usage patterns", i),
				ModelID:  "test-model",
				Provider: "test-provider",
			}

			cache.SetTextResponse(request, response)

			// Periodically check cache stats
			if i%1000 == 0 {
				stats := cache.GetStats()
				b.Logf("Cache size: %d entries", stats.TotalSize)
			}
		}
	})
}
