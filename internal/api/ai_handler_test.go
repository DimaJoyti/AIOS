package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/aios/aios/internal/ai"
	"github.com/aios/aios/pkg/models"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock implementations for testing

type MockAIOrchestrator struct {
	mock.Mock
}

func (m *MockAIOrchestrator) ProcessRequest(ctx context.Context, request *models.AIRequest) (*models.AIResponse, error) {
	args := m.Called(ctx, request)
	return args.Get(0).(*models.AIResponse), args.Error(1)
}

func (m *MockAIOrchestrator) GetServiceStatus(ctx context.Context) (*models.AIServiceStatus, error) {
	args := m.Called(ctx)
	return args.Get(0).(*models.AIServiceStatus), args.Error(1)
}

func (m *MockAIOrchestrator) RouteRequest(ctx context.Context, request *models.AIRequest) (string, error) {
	args := m.Called(ctx, request)
	return args.String(0), args.Error(1)
}

func (m *MockAIOrchestrator) AggregateResults(ctx context.Context, results []models.AIResult) (*models.AggregatedResult, error) {
	args := m.Called(ctx, results)
	return args.Get(0).(*models.AggregatedResult), args.Error(1)
}

func (m *MockAIOrchestrator) ManageWorkflow(ctx context.Context, workflow *models.AIWorkflow) (*models.WorkflowResult, error) {
	args := m.Called(ctx, workflow)
	return args.Get(0).(*models.WorkflowResult), args.Error(1)
}

type MockLLMService struct {
	mock.Mock
}

func (m *MockLLMService) ProcessQuery(ctx context.Context, query string) (*models.LLMResponse, error) {
	args := m.Called(ctx, query)
	return args.Get(0).(*models.LLMResponse), args.Error(1)
}

func (m *MockLLMService) ProcessQueryStream(ctx context.Context, query string) (<-chan *models.LLMStreamChunk, error) {
	args := m.Called(ctx, query)
	return args.Get(0).(<-chan *models.LLMStreamChunk), args.Error(1)
}

func (m *MockLLMService) GenerateCode(ctx context.Context, prompt string) (*models.CodeResponse, error) {
	args := m.Called(ctx, prompt)
	return args.Get(0).(*models.CodeResponse), args.Error(1)
}

func (m *MockLLMService) GenerateCodeStream(ctx context.Context, prompt string) (<-chan *models.CodeStreamChunk, error) {
	args := m.Called(ctx, prompt)
	return args.Get(0).(<-chan *models.CodeStreamChunk), args.Error(1)
}

func (m *MockLLMService) AnalyzeText(ctx context.Context, text string) (*models.TextAnalysis, error) {
	args := m.Called(ctx, text)
	return args.Get(0).(*models.TextAnalysis), args.Error(1)
}

func (m *MockLLMService) Chat(ctx context.Context, message string, conversationID string) (*models.ChatResponse, error) {
	args := m.Called(ctx, message, conversationID)
	return args.Get(0).(*models.ChatResponse), args.Error(1)
}

func (m *MockLLMService) ChatWithHistory(ctx context.Context, messages []models.ChatMessage) (*models.ChatResponse, error) {
	args := m.Called(ctx, messages)
	return args.Get(0).(*models.ChatResponse), args.Error(1)
}

func (m *MockLLMService) ChatStream(ctx context.Context, message string, conversationID string) (<-chan *models.ChatStreamChunk, error) {
	args := m.Called(ctx, message, conversationID)
	return args.Get(0).(<-chan *models.ChatStreamChunk), args.Error(1)
}

func (m *MockLLMService) Summarize(ctx context.Context, text string) (*models.SummaryResponse, error) {
	args := m.Called(ctx, text)
	return args.Get(0).(*models.SummaryResponse), args.Error(1)
}

func (m *MockLLMService) Translate(ctx context.Context, text, fromLang, toLang string) (*models.TranslationResponse, error) {
	args := m.Called(ctx, text, fromLang, toLang)
	return args.Get(0).(*models.TranslationResponse), args.Error(1)
}

func (m *MockLLMService) FunctionCall(ctx context.Context, functionName string, parameters map[string]any) (*models.FunctionCallResponse, error) {
	args := m.Called(ctx, functionName, parameters)
	return args.Get(0).(*models.FunctionCallResponse), args.Error(1)
}

func (m *MockLLMService) EmbedText(ctx context.Context, text string) (*models.EmbeddingResponse, error) {
	args := m.Called(ctx, text)
	return args.Get(0).(*models.EmbeddingResponse), args.Error(1)
}

func (m *MockLLMService) BatchEmbed(ctx context.Context, texts []string) (*models.BatchEmbeddingResponse, error) {
	args := m.Called(ctx, texts)
	return args.Get(0).(*models.BatchEmbeddingResponse), args.Error(1)
}

func (m *MockLLMService) GetModels(ctx context.Context) ([]models.AIModel, error) {
	args := m.Called(ctx)
	return args.Get(0).([]models.AIModel), args.Error(1)
}

func (m *MockLLMService) LoadModel(ctx context.Context, modelName string) error {
	args := m.Called(ctx, modelName)
	return args.Error(0)
}

func (m *MockLLMService) UnloadModel(ctx context.Context, modelName string) error {
	args := m.Called(ctx, modelName)
	return args.Error(0)
}

func (m *MockLLMService) GetModelInfo(ctx context.Context, modelName string) (*models.ModelInfo, error) {
	args := m.Called(ctx, modelName)
	return args.Get(0).(*models.ModelInfo), args.Error(1)
}

type MockCVService struct {
	mock.Mock
}

func (m *MockCVService) AnalyzeScreen(ctx context.Context, screenshot []byte) (*models.ScreenAnalysis, error) {
	args := m.Called(ctx, screenshot)
	return args.Get(0).(*models.ScreenAnalysis), args.Error(1)
}

func (m *MockCVService) DetectUI(ctx context.Context, image []byte) (*models.UIElements, error) {
	args := m.Called(ctx, image)
	return args.Get(0).(*models.UIElements), args.Error(1)
}

func (m *MockCVService) RecognizeText(ctx context.Context, image []byte) (*models.TextRecognition, error) {
	args := m.Called(ctx, image)
	return args.Get(0).(*models.TextRecognition), args.Error(1)
}

func (m *MockCVService) ClassifyImage(ctx context.Context, image []byte) (*models.ImageClassification, error) {
	args := m.Called(ctx, image)
	return args.Get(0).(*models.ImageClassification), args.Error(1)
}

func (m *MockCVService) DetectObjects(ctx context.Context, image []byte) (*models.ObjectDetection, error) {
	args := m.Called(ctx, image)
	return args.Get(0).(*models.ObjectDetection), args.Error(1)
}

func (m *MockCVService) AnalyzeLayout(ctx context.Context, image []byte) (*models.LayoutAnalysis, error) {
	args := m.Called(ctx, image)
	return args.Get(0).(*models.LayoutAnalysis), args.Error(1)
}

func (m *MockCVService) GenerateDescription(ctx context.Context, image []byte) (*models.ImageDescription, error) {
	args := m.Called(ctx, image)
	return args.Get(0).(*models.ImageDescription), args.Error(1)
}

func (m *MockCVService) CompareImages(ctx context.Context, image1, image2 []byte) (*models.ImageComparison, error) {
	args := m.Called(ctx, image1, image2)
	return args.Get(0).(*models.ImageComparison), args.Error(1)
}

type MockVoiceService struct {
	mock.Mock
}

func (m *MockVoiceService) SpeechToText(ctx context.Context, audio []byte) (*models.SpeechRecognition, error) {
	args := m.Called(ctx, audio)
	return args.Get(0).(*models.SpeechRecognition), args.Error(1)
}

func (m *MockVoiceService) TextToSpeech(ctx context.Context, text string) (*models.SpeechSynthesis, error) {
	args := m.Called(ctx, text)
	return args.Get(0).(*models.SpeechSynthesis), args.Error(1)
}

func (m *MockVoiceService) DetectWakeWord(ctx context.Context, audio []byte) (*models.WakeWordDetection, error) {
	args := m.Called(ctx, audio)
	return args.Get(0).(*models.WakeWordDetection), args.Error(1)
}

func (m *MockVoiceService) AnalyzeVoice(ctx context.Context, audio []byte) (*models.VoiceAnalysis, error) {
	args := m.Called(ctx, audio)
	return args.Get(0).(*models.VoiceAnalysis), args.Error(1)
}

func (m *MockVoiceService) ProcessVoiceCommand(ctx context.Context, audio []byte) (*models.VoiceCommand, error) {
	args := m.Called(ctx, audio)
	return args.Get(0).(*models.VoiceCommand), args.Error(1)
}

type MockNLPService struct {
	mock.Mock
}

func (m *MockNLPService) ParseIntent(ctx context.Context, text string) (*models.IntentAnalysis, error) {
	args := m.Called(ctx, text)
	return args.Get(0).(*models.IntentAnalysis), args.Error(1)
}

func (m *MockNLPService) ExtractEntities(ctx context.Context, text string) (*models.EntityExtraction, error) {
	args := m.Called(ctx, text)
	return args.Get(0).(*models.EntityExtraction), args.Error(1)
}

func (m *MockNLPService) AnalyzeSentiment(ctx context.Context, text string) (*models.SentimentAnalysis, error) {
	args := m.Called(ctx, text)
	return args.Get(0).(*models.SentimentAnalysis), args.Error(1)
}

func (m *MockNLPService) GenerateResponse(ctx context.Context, intent *models.IntentAnalysis, context map[string]interface{}) (*models.NLResponse, error) {
	args := m.Called(ctx, intent, context)
	return args.Get(0).(*models.NLResponse), args.Error(1)
}

func (m *MockNLPService) ParseCommand(ctx context.Context, text string) (*models.CommandParsing, error) {
	args := m.Called(ctx, text)
	return args.Get(0).(*models.CommandParsing), args.Error(1)
}

func (m *MockNLPService) ValidateCommand(ctx context.Context, command *models.CommandParsing) (*models.CommandValidation, error) {
	args := m.Called(ctx, command)
	return args.Get(0).(*models.CommandValidation), args.Error(1)
}

// Test setup helper
func setupAIHandler() (*AIHandler, *MockAIOrchestrator, *MockLLMService, *MockCVService, *MockVoiceService, *MockNLPService) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel) // Reduce noise in tests

	mockOrchestrator := &MockAIOrchestrator{}
	mockLLM := &MockLLMService{}
	mockCV := &MockCVService{}
	mockVoice := &MockVoiceService{}
	mockNLP := &MockNLPService{}

	handler := NewAIHandler(
		ai.AIOrchestrator(mockOrchestrator),
		ai.LanguageModelService(mockLLM),
		ai.ComputerVisionService(mockCV),
		ai.VoiceService(mockVoice),
		ai.NaturalLanguageService(mockNLP),
		logger,
	)

	return handler, mockOrchestrator, mockLLM, mockCV, mockVoice, mockNLP
}

// Test cases

func TestProcessAIRequest(t *testing.T) {
	handler, mockOrchestrator, _, _, _, _ := setupAIHandler()

	// Setup mock response
	expectedResponse := &models.AIResponse{
		RequestID:      "test-request-id",
		Type:           "llm_query",
		Result:         map[string]interface{}{"text": "Test response"},
		Confidence:     0.95,
		ProcessingTime: time.Millisecond * 100,
		Timestamp:      time.Now(),
	}

	mockOrchestrator.On("ProcessRequest", mock.Anything, mock.AnythingOfType("*models.AIRequest")).Return(expectedResponse, nil)

	// Create test request
	request := models.AIRequest{
		ID:    "test-request-id",
		Type:  "llm_query",
		Input: map[string]interface{}{"query": "Test query"},
	}

	requestBody, _ := json.Marshal(request)
	req := httptest.NewRequest("POST", "/ai/process", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	
	rr := httptest.NewRecorder()

	// Execute request
	handler.ProcessAIRequest(rr, req)

	// Verify response
	assert.Equal(t, http.StatusOK, rr.Code)
	
	var response models.AIResponse
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, expectedResponse.RequestID, response.RequestID)
	assert.Equal(t, expectedResponse.Type, response.Type)

	mockOrchestrator.AssertExpectations(t)
}

func TestProcessLLMQuery(t *testing.T) {
	handler, _, mockLLM, _, _, _ := setupAIHandler()

	// Setup mock response
	expectedResponse := &models.LLMResponse{
		Text:           "Test LLM response",
		Model:          "test-model",
		TokensUsed:     50,
		Confidence:     0.95,
		ProcessingTime: time.Millisecond * 100,
		Timestamp:      time.Now(),
	}

	mockLLM.On("ProcessQuery", mock.Anything, "Test query").Return(expectedResponse, nil)

	// Create test request
	request := map[string]interface{}{
		"query": "Test query",
		"model": "test-model",
	}

	requestBody, _ := json.Marshal(request)
	req := httptest.NewRequest("POST", "/ai/llm/query", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	
	rr := httptest.NewRecorder()

	// Execute request
	handler.ProcessLLMQuery(rr, req)

	// Verify response
	assert.Equal(t, http.StatusOK, rr.Code)
	
	var response models.LLMResponse
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, expectedResponse.Text, response.Text)
	assert.Equal(t, expectedResponse.Model, response.Model)

	mockLLM.AssertExpectations(t)
}

func TestDetectObjects(t *testing.T) {
	handler, _, _, mockCV, _, _ := setupAIHandler()

	// Setup mock response
	expectedResponse := &models.ObjectDetection{
		Objects: []models.DetectedObject{
			{
				Class:      "person",
				Confidence: 0.95,
				Bounds:     models.Rectangle{X: 100, Y: 100, Width: 200, Height: 300},
			},
		},
		Count:     1,
		Timestamp: time.Now(),
	}

	mockCV.On("DetectObjects", mock.Anything, mock.AnythingOfType("[]uint8")).Return(expectedResponse, nil)

	// Create test request with mock image data
	imageData := []byte("mock-image-data")
	req := httptest.NewRequest("POST", "/ai/cv/detect-objects", bytes.NewBuffer(imageData))
	req.Header.Set("Content-Type", "image/jpeg")
	
	rr := httptest.NewRecorder()

	// Execute request
	handler.DetectObjects(rr, req)

	// Verify response
	assert.Equal(t, http.StatusOK, rr.Code)
	
	var response models.ObjectDetection
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, len(expectedResponse.Objects), len(response.Objects))
	assert.Equal(t, expectedResponse.Objects[0].Class, response.Objects[0].Class)

	mockCV.AssertExpectations(t)
}

func TestSpeechToText(t *testing.T) {
	handler, _, _, _, mockVoice, _ := setupAIHandler()

	// Setup mock response
	expectedResponse := &models.SpeechRecognition{
		Text:       "Hello world",
		Confidence: 0.95,
		Language:   "en",
		Duration:   time.Second * 2,
		Timestamp:  time.Now(),
	}

	mockVoice.On("SpeechToText", mock.Anything, mock.AnythingOfType("[]uint8")).Return(expectedResponse, nil)

	// Create test request with mock audio data
	audioData := []byte("mock-audio-data")
	req := httptest.NewRequest("POST", "/ai/voice/speech-to-text", bytes.NewBuffer(audioData))
	req.Header.Set("Content-Type", "audio/wav")
	
	rr := httptest.NewRecorder()

	// Execute request
	handler.SpeechToText(rr, req)

	// Verify response
	assert.Equal(t, http.StatusOK, rr.Code)
	
	var response models.SpeechRecognition
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, expectedResponse.Text, response.Text)
	assert.Equal(t, expectedResponse.Confidence, response.Confidence)

	mockVoice.AssertExpectations(t)
}

func TestParseIntent(t *testing.T) {
	handler, _, _, _, _, mockNLP := setupAIHandler()

	// Setup mock response
	expectedResponse := &models.IntentAnalysis{
		Intent:     "open_application",
		Confidence: 0.88,
		Entities: []models.NamedEntity{
			{
				Text:       "terminal",
				Type:       "APPLICATION",
				Confidence: 0.85,
			},
		},
		Timestamp: time.Now(),
	}

	mockNLP.On("ParseIntent", mock.Anything, "open terminal").Return(expectedResponse, nil)

	// Create test request
	request := map[string]interface{}{
		"text": "open terminal",
	}

	requestBody, _ := json.Marshal(request)
	req := httptest.NewRequest("POST", "/ai/nlp/parse-intent", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	
	rr := httptest.NewRecorder()

	// Execute request
	handler.ParseIntent(rr, req)

	// Verify response
	assert.Equal(t, http.StatusOK, rr.Code)
	
	var response models.IntentAnalysis
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, expectedResponse.Intent, response.Intent)
	assert.Equal(t, expectedResponse.Confidence, response.Confidence)

	mockNLP.AssertExpectations(t)
}

func TestRouteRegistration(t *testing.T) {
	handler, _, _, _, _, _ := setupAIHandler()
	
	router := mux.NewRouter()
	handler.RegisterRoutes(router)

	// Test that routes are registered with correct methods
	routes := []struct {
		path   string
		method string
	}{
		{"/ai/process", "POST"},
		{"/ai/status", "GET"},
		{"/ai/llm/query", "POST"},
		{"/ai/llm/chat", "POST"},
		{"/ai/llm/models", "GET"},
		{"/ai/cv/analyze-screen", "POST"},
		{"/ai/cv/detect-objects", "POST"},
		{"/ai/cv/recognize-text", "POST"},
		{"/ai/cv/classify-image", "POST"},
		{"/ai/cv/generate-image", "POST"},
		{"/ai/voice/speech-to-text", "POST"},
		{"/ai/voice/text-to-speech", "POST"},
		{"/ai/voice/detect-wake-word", "POST"},
		{"/ai/voice/process-command", "POST"},
		{"/ai/nlp/parse-intent", "POST"},
		{"/ai/nlp/extract-entities", "POST"},
		{"/ai/nlp/analyze-sentiment", "POST"},
		{"/ai/nlp/summarize", "POST"},
	}

	for _, route := range routes {
		req := httptest.NewRequest(route.method, route.path, nil)
		match := &mux.RouteMatch{}
		matched := router.Match(req, match)
		assert.True(t, matched, "Route %s with method %s should be registered", route.path, route.method)
	}
}
