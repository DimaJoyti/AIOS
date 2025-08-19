package unit

import (
	"context"
	"testing"
	"time"

	"github.com/aios/aios/internal/ai/services"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockModelProvider implements the ModelProvider interface for testing
type MockModelProvider struct {
	mock.Mock
}

func (m *MockModelProvider) GetName() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockModelProvider) GetModels() []string {
	args := m.Called()
	return args.Get(0).([]string)
}

func (m *MockModelProvider) IsAvailable() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockModelProvider) GenerateText(ctx context.Context, request *services.TextGenerationRequest) (*services.TextGenerationResponse, error) {
	args := m.Called(ctx, request)
	return args.Get(0).(*services.TextGenerationResponse), args.Error(1)
}

func (m *MockModelProvider) GenerateImage(ctx context.Context, request *services.ImageGenerationRequest) (*services.ImageGenerationResponse, error) {
	args := m.Called(ctx, request)
	return args.Get(0).(*services.ImageGenerationResponse), args.Error(1)
}

func (m *MockModelProvider) ProcessAudio(ctx context.Context, request *services.AudioProcessingRequest) (*services.AudioProcessingResponse, error) {
	args := m.Called(ctx, request)
	return args.Get(0).(*services.AudioProcessingResponse), args.Error(1)
}

func (m *MockModelProvider) ProcessVideo(ctx context.Context, request *services.VideoProcessingRequest) (*services.VideoProcessingResponse, error) {
	args := m.Called(ctx, request)
	return args.Get(0).(*services.VideoProcessingResponse), args.Error(1)
}

func (m *MockModelProvider) GetUsage() *services.ProviderUsage {
	args := m.Called()
	return args.Get(0).(*services.ProviderUsage)
}

func (m *MockModelProvider) GetHealth() *services.ProviderHealth {
	args := m.Called()
	return args.Get(0).(*services.ProviderHealth)
}

func TestModelManager(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel) // Reduce noise in tests

	t.Run("NewModelManager", func(t *testing.T) {
		mm := services.NewModelManager(logger)
		assert.NotNil(t, mm)
	})

	t.Run("RegisterModel", func(t *testing.T) {
		mm := services.NewModelManager(logger)

		model := &services.AIModel{
			ID:       "test-model",
			Name:     "Test Model",
			Provider: "test-provider",
			Type:     "text",
			Version:  "1.0",
			Status:   "active",
		}

		err := mm.RegisterModel(model)
		assert.NoError(t, err)

		// Verify model was registered
		retrievedModel, err := mm.GetModel("test-model")
		assert.NoError(t, err)
		assert.Equal(t, model.ID, retrievedModel.ID)
		assert.Equal(t, model.Name, retrievedModel.Name)
	})

	t.Run("RegisterProvider", func(t *testing.T) {
		mm := services.NewModelManager(logger)
		mockProvider := &MockModelProvider{}

		mockProvider.On("GetName").Return("test-provider")

		err := mm.RegisterProvider(mockProvider)
		assert.NoError(t, err)

		mockProvider.AssertExpectations(t)
	})

	t.Run("GetModel_NotFound", func(t *testing.T) {
		mm := services.NewModelManager(logger)

		_, err := mm.GetModel("non-existent")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "model not found")
	})

	t.Run("ListModels", func(t *testing.T) {
		mm := services.NewModelManager(logger)

		// Initially empty
		models := mm.ListModels()
		assert.Empty(t, models)

		// Add a model
		model := &services.AIModel{
			ID:       "test-model",
			Name:     "Test Model",
			Provider: "test-provider",
			Type:     "text",
			Status:   "active",
		}

		err := mm.RegisterModel(model)
		require.NoError(t, err)

		// Should now have one model
		models = mm.ListModels()
		assert.Len(t, models, 1)
		assert.Equal(t, "test-model", models[0].ID)
	})

	t.Run("GetModelsByType", func(t *testing.T) {
		mm := services.NewModelManager(logger)

		// Add models of different types
		textModel := &services.AIModel{
			ID:     "text-model",
			Name:   "Text Model",
			Type:   "text",
			Status: "active",
		}

		imageModel := &services.AIModel{
			ID:     "image-model",
			Name:   "Image Model",
			Type:   "image",
			Status: "active",
		}

		require.NoError(t, mm.RegisterModel(textModel))
		require.NoError(t, mm.RegisterModel(imageModel))

		// Get text models
		textModels := mm.GetModelsByType("text")
		assert.Len(t, textModels, 1)
		assert.Equal(t, "text-model", textModels[0].ID)

		// Get image models
		imageModels := mm.GetModelsByType("image")
		assert.Len(t, imageModels, 1)
		assert.Equal(t, "image-model", imageModels[0].ID)

		// Get non-existent type
		audioModels := mm.GetModelsByType("audio")
		assert.Empty(t, audioModels)
	})

	t.Run("GenerateText", func(t *testing.T) {
		mm := services.NewModelManager(logger)
		mockProvider := &MockModelProvider{}

		// Setup mock
		mockProvider.On("GetName").Return("test-provider")

		// Register provider
		err := mm.RegisterProvider(mockProvider)
		require.NoError(t, err)

		// Register model
		model := &services.AIModel{
			ID:       "test-model",
			Provider: "test-provider",
			Type:     "text",
			Status:   "active",
		}
		err = mm.RegisterModel(model)
		require.NoError(t, err)

		// Setup text generation mock
		request := &services.TextGenerationRequest{
			ModelID: "test-model",
			Prompt:  "Test prompt",
		}

		expectedResponse := &services.TextGenerationResponse{
			Text:         "Generated text",
			FinishReason: "stop",
			Usage: services.TokenUsage{
				PromptTokens:     10,
				CompletionTokens: 20,
				TotalTokens:      30,
			},
			ModelID:  "test-model",
			Provider: "test-provider",
			Latency:  100 * time.Millisecond,
			Cost:     0.001,
		}

		mockProvider.On("GenerateText", mock.Anything, request).Return(expectedResponse, nil)

		// Execute
		ctx := context.Background()
		response, err := mm.GenerateText(ctx, request)

		// Verify
		assert.NoError(t, err)
		assert.Equal(t, expectedResponse.Text, response.Text)
		assert.Equal(t, expectedResponse.Usage.TotalTokens, response.Usage.TotalTokens)

		mockProvider.AssertExpectations(t)
	})
}

func TestPromptManager(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	t.Run("NewPromptManager", func(t *testing.T) {
		pm := services.NewPromptManager(logger)
		assert.NotNil(t, pm)
	})

	t.Run("CreateTemplate", func(t *testing.T) {
		pm := services.NewPromptManager(logger)

		template := &services.PromptTemplate{
			ID:          "test-template",
			Name:        "Test Template",
			Description: "A test template",
			Category:    "test",
			Template:    "Hello {{.name}}!",
			Variables: []services.PromptVariable{
				{
					Name:        "name",
					Type:        "string",
					Description: "Name to greet",
					Required:    true,
				},
			},
		}

		err := pm.CreateTemplate(template)
		assert.NoError(t, err)

		// Verify template was created
		retrievedTemplate, err := pm.GetTemplate("test-template")
		assert.NoError(t, err)
		assert.Equal(t, template.ID, retrievedTemplate.ID)
		assert.Equal(t, template.Name, retrievedTemplate.Name)
	})

	t.Run("RenderTemplate", func(t *testing.T) {
		pm := services.NewPromptManager(logger)

		template := &services.PromptTemplate{
			ID:       "greeting-template",
			Name:     "Greeting Template",
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

		err := pm.CreateTemplate(template)
		require.NoError(t, err)

		// Render with variables
		variables := map[string]interface{}{
			"name": "Alice",
			"age":  30,
		}

		ctx := context.Background()
		result, err := pm.RenderTemplate(ctx, "greeting-template", variables)

		assert.NoError(t, err)
		assert.Equal(t, "Hello Alice, you are 30 years old!", result)
	})

	t.Run("RenderTemplate_MissingRequiredVariable", func(t *testing.T) {
		pm := services.NewPromptManager(logger)

		template := &services.PromptTemplate{
			ID:       "required-template",
			Name:     "Required Template",
			Template: "Hello {{.name}}!",
			Variables: []services.PromptVariable{
				{
					Name:     "name",
					Type:     "string",
					Required: true,
				},
			},
		}

		err := pm.CreateTemplate(template)
		require.NoError(t, err)

		// Try to render without required variable
		variables := map[string]interface{}{}

		ctx := context.Background()
		_, err = pm.RenderTemplate(ctx, "required-template", variables)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "required variable")
	})

	t.Run("ListTemplates", func(t *testing.T) {
		pm := services.NewPromptManager(logger)

		// Initially empty
		templates := pm.ListTemplates()
		assert.Empty(t, templates)

		// Add a template
		template := &services.PromptTemplate{
			ID:       "test-template",
			Name:     "Test Template",
			Template: "Test {{.var}}",
		}

		err := pm.CreateTemplate(template)
		require.NoError(t, err)

		// Should now have one template
		templates = pm.ListTemplates()
		assert.Len(t, templates, 1)
		assert.Equal(t, "test-template", templates[0].ID)
	})
}

func TestModelCache(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	t.Run("NewModelCache", func(t *testing.T) {
		cache := services.NewModelCache(logger)
		assert.NotNil(t, cache)
	})

	t.Run("TextResponse_CacheHit", func(t *testing.T) {
		cache := services.NewModelCache(logger)

		request := &services.TextGenerationRequest{
			ModelID: "test-model",
			Prompt:  "Test prompt",
		}

		response := &services.TextGenerationResponse{
			Text:     "Cached response",
			ModelID:  "test-model",
			Provider: "test-provider",
		}

		// Cache the response
		cache.SetTextResponse(request, response)

		// Retrieve from cache
		cachedResponse := cache.GetTextResponse(request)
		assert.NotNil(t, cachedResponse)
		assert.Equal(t, response.Text, cachedResponse.Text)
	})

	t.Run("TextResponse_CacheMiss", func(t *testing.T) {
		cache := services.NewModelCache(logger)

		request := &services.TextGenerationRequest{
			ModelID: "test-model",
			Prompt:  "Non-existent prompt",
		}

		// Should return nil for cache miss
		cachedResponse := cache.GetTextResponse(request)
		assert.Nil(t, cachedResponse)
	})

	t.Run("CacheStats", func(t *testing.T) {
		cache := services.NewModelCache(logger)

		// Initial stats
		stats := cache.GetStats()
		assert.Equal(t, int64(0), stats.TotalRequests)
		assert.Equal(t, int64(0), stats.CacheHits)
		assert.Equal(t, int64(0), stats.CacheMisses)

		request := &services.TextGenerationRequest{
			ModelID: "test-model",
			Prompt:  "Test prompt",
		}

		// Cache miss
		cache.GetTextResponse(request)
		stats = cache.GetStats()
		assert.Equal(t, int64(1), stats.TotalRequests)
		assert.Equal(t, int64(0), stats.CacheHits)
		assert.Equal(t, int64(1), stats.CacheMisses)

		// Cache the response
		response := &services.TextGenerationResponse{
			Text: "Test response",
		}
		cache.SetTextResponse(request, response)

		// Cache hit
		cache.GetTextResponse(request)
		stats = cache.GetStats()
		assert.Equal(t, int64(2), stats.TotalRequests)
		assert.Equal(t, int64(1), stats.CacheHits)
		assert.Equal(t, int64(1), stats.CacheMisses)
	})
}
