package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/aios/aios/internal/ai"
	"github.com/aios/aios/pkg/models"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// AIHandler handles AI service API endpoints
type AIHandler struct {
	orchestrator ai.AIOrchestrator
	llmService   ai.LanguageModelService
	cvService    ai.ComputerVisionService
	voiceService ai.VoiceService
	nlpService   ai.NaturalLanguageService
	logger       *logrus.Logger
	tracer       trace.Tracer
}

// NewAIHandler creates a new AI API handler
func NewAIHandler(
	orchestrator ai.AIOrchestrator,
	llmService ai.LanguageModelService,
	cvService ai.ComputerVisionService,
	voiceService ai.VoiceService,
	nlpService ai.NaturalLanguageService,
	logger *logrus.Logger,
) *AIHandler {
	return &AIHandler{
		orchestrator: orchestrator,
		llmService:   llmService,
		cvService:    cvService,
		voiceService: voiceService,
		nlpService:   nlpService,
		logger:       logger,
		tracer:       otel.Tracer("api.ai_handler"),
	}
}

// RegisterRoutes registers AI API routes
func (h *AIHandler) RegisterRoutes(router *mux.Router) {
	// AI Orchestrator routes
	router.HandleFunc("/ai/process", h.ProcessAIRequest).Methods("POST")
	router.HandleFunc("/ai/status", h.GetAIStatus).Methods("GET")

	// LLM routes
	router.HandleFunc("/ai/llm/query", h.ProcessLLMQuery).Methods("POST")
	router.HandleFunc("/ai/llm/chat", h.ProcessLLMChat).Methods("POST")
	router.HandleFunc("/ai/llm/models", h.GetLLMModels).Methods("GET")

	// Computer Vision routes
	router.HandleFunc("/ai/cv/analyze-screen", h.AnalyzeScreen).Methods("POST")
	router.HandleFunc("/ai/cv/detect-objects", h.DetectObjects).Methods("POST")
	router.HandleFunc("/ai/cv/recognize-text", h.RecognizeText).Methods("POST")
	router.HandleFunc("/ai/cv/classify-image", h.ClassifyImage).Methods("POST")
	router.HandleFunc("/ai/cv/generate-image", h.GenerateImage).Methods("POST")

	// Voice routes
	router.HandleFunc("/ai/voice/speech-to-text", h.SpeechToText).Methods("POST")
	router.HandleFunc("/ai/voice/text-to-speech", h.TextToSpeech).Methods("POST")
	router.HandleFunc("/ai/voice/detect-wake-word", h.DetectWakeWord).Methods("POST")
	router.HandleFunc("/ai/voice/process-command", h.ProcessVoiceCommand).Methods("POST")

	// NLP routes
	router.HandleFunc("/ai/nlp/parse-intent", h.ParseIntent).Methods("POST")
	router.HandleFunc("/ai/nlp/extract-entities", h.ExtractEntities).Methods("POST")
	router.HandleFunc("/ai/nlp/analyze-sentiment", h.AnalyzeSentiment).Methods("POST")
	router.HandleFunc("/ai/nlp/summarize", h.SummarizeText).Methods("POST")
}

// ProcessAIRequest handles complex AI requests through the orchestrator
func (h *AIHandler) ProcessAIRequest(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "ai.ProcessAIRequest")
	defer span.End()

	var request models.AIRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	span.SetAttributes(
		attribute.String("request_type", request.Type),
		attribute.String("request_id", request.ID),
	)

	h.logger.WithFields(logrus.Fields{
		"request_id":   request.ID,
		"request_type": request.Type,
	}).Info("Processing AI request")

	response, err := h.orchestrator.ProcessRequest(ctx, &request)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Failed to process AI request", err)
		return
	}

	h.respondWithJSON(w, http.StatusOK, response)
}

// GetAIStatus returns the status of all AI services
func (h *AIHandler) GetAIStatus(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "ai.GetAIStatus")
	defer span.End()

	status, err := h.orchestrator.GetServiceStatus(ctx)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Failed to get AI status", err)
		return
	}

	h.respondWithJSON(w, http.StatusOK, status)
}

// ProcessLLMQuery handles LLM query requests
func (h *AIHandler) ProcessLLMQuery(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "ai.ProcessLLMQuery")
	defer span.End()

	var request struct {
		Query       string                 `json:"query"`
		Model       string                 `json:"model,omitempty"`
		Parameters  map[string]interface{} `json:"parameters,omitempty"`
		Temperature float64                `json:"temperature,omitempty"`
		MaxTokens   int                    `json:"max_tokens,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	span.SetAttributes(
		attribute.String("query", request.Query[:min(100, len(request.Query))]),
		attribute.String("model", request.Model),
	)

	response, err := h.llmService.ProcessQuery(ctx, request.Query)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Failed to process LLM query", err)
		return
	}

	h.respondWithJSON(w, http.StatusOK, response)
}

// ProcessLLMChat handles LLM chat requests
func (h *AIHandler) ProcessLLMChat(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "ai.ProcessLLMChat")
	defer span.End()

	var request struct {
		Messages []models.ChatMessage `json:"messages"`
		Model    string               `json:"model,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	span.SetAttributes(
		attribute.Int("message_count", len(request.Messages)),
		attribute.String("model", request.Model),
	)

	response, err := h.llmService.ChatWithHistory(ctx, request.Messages)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Failed to process chat", err)
		return
	}

	h.respondWithJSON(w, http.StatusOK, response)
}

// GetLLMModels returns available LLM models
func (h *AIHandler) GetLLMModels(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "ai.GetLLMModels")
	defer span.End()

	models, err := h.llmService.GetModels(ctx)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Failed to get models", err)
		return
	}

	h.respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"models": models,
		"count":  len(models),
	})
}

// AnalyzeScreen handles screen analysis requests
func (h *AIHandler) AnalyzeScreen(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "ai.AnalyzeScreen")
	defer span.End()

	imageData, err := h.readImageFromRequest(r)
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Failed to read image data", err)
		return
	}

	span.SetAttributes(attribute.Int("image_size", len(imageData)))

	analysis, err := h.cvService.AnalyzeScreen(ctx, imageData)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Failed to analyze screen", err)
		return
	}

	h.respondWithJSON(w, http.StatusOK, analysis)
}

// DetectObjects handles object detection requests
func (h *AIHandler) DetectObjects(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "ai.DetectObjects")
	defer span.End()

	imageData, err := h.readImageFromRequest(r)
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Failed to read image data", err)
		return
	}

	span.SetAttributes(attribute.Int("image_size", len(imageData)))

	detection, err := h.cvService.DetectObjects(ctx, imageData)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Failed to detect objects", err)
		return
	}

	h.respondWithJSON(w, http.StatusOK, detection)
}

// RecognizeText handles OCR requests
func (h *AIHandler) RecognizeText(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "ai.RecognizeText")
	defer span.End()

	imageData, err := h.readImageFromRequest(r)
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Failed to read image data", err)
		return
	}

	span.SetAttributes(attribute.Int("image_size", len(imageData)))

	recognition, err := h.cvService.RecognizeText(ctx, imageData)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Failed to recognize text", err)
		return
	}

	h.respondWithJSON(w, http.StatusOK, recognition)
}

// ClassifyImage handles image classification requests
func (h *AIHandler) ClassifyImage(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "ai.ClassifyImage")
	defer span.End()

	imageData, err := h.readImageFromRequest(r)
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Failed to read image data", err)
		return
	}

	span.SetAttributes(attribute.Int("image_size", len(imageData)))

	classification, err := h.cvService.ClassifyImage(ctx, imageData)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Failed to classify image", err)
		return
	}

	h.respondWithJSON(w, http.StatusOK, classification)
}

// GenerateImage handles image generation requests
func (h *AIHandler) GenerateImage(w http.ResponseWriter, r *http.Request) {
	_, span := h.tracer.Start(r.Context(), "ai.GenerateImage")
	defer span.End()

	var request struct {
		Prompt     string                 `json:"prompt"`
		Parameters map[string]interface{} `json:"parameters,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	span.SetAttributes(attribute.String("prompt", request.Prompt))

	// Computer Vision service doesn't have GenerateImage - this should be handled by MultiModalService
	// For now, return an error indicating this feature is not available
	h.respondWithError(w, http.StatusNotImplemented, "Image generation not available in CV service", fmt.Errorf("use multimodal service for image generation"))
}

// SpeechToText handles speech-to-text requests
func (h *AIHandler) SpeechToText(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "ai.SpeechToText")
	defer span.End()

	audioData, err := h.readAudioFromRequest(r)
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Failed to read audio data", err)
		return
	}

	span.SetAttributes(attribute.Int("audio_size", len(audioData)))

	recognition, err := h.voiceService.SpeechToText(ctx, audioData)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Failed to process speech-to-text", err)
		return
	}

	h.respondWithJSON(w, http.StatusOK, recognition)
}

// TextToSpeech handles text-to-speech requests
func (h *AIHandler) TextToSpeech(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "ai.TextToSpeech")
	defer span.End()

	var request struct {
		Text  string `json:"text"`
		Voice string `json:"voice,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	span.SetAttributes(attribute.String("text", request.Text[:min(100, len(request.Text))]))

	synthesis, err := h.voiceService.TextToSpeech(ctx, request.Text)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Failed to process text-to-speech", err)
		return
	}

	h.respondWithJSON(w, http.StatusOK, synthesis)
}

// DetectWakeWord handles wake word detection requests
func (h *AIHandler) DetectWakeWord(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "ai.DetectWakeWord")
	defer span.End()

	audioData, err := h.readAudioFromRequest(r)
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Failed to read audio data", err)
		return
	}

	span.SetAttributes(attribute.Int("audio_size", len(audioData)))

	detection, err := h.voiceService.DetectWakeWord(ctx, audioData)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Failed to detect wake word", err)
		return
	}

	h.respondWithJSON(w, http.StatusOK, detection)
}

// ProcessVoiceCommand handles voice command processing requests
func (h *AIHandler) ProcessVoiceCommand(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "ai.ProcessVoiceCommand")
	defer span.End()

	audioData, err := h.readAudioFromRequest(r)
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Failed to read audio data", err)
		return
	}

	span.SetAttributes(attribute.Int("audio_size", len(audioData)))

	command, err := h.voiceService.ProcessVoiceCommand(ctx, audioData)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Failed to process voice command", err)
		return
	}

	h.respondWithJSON(w, http.StatusOK, command)
}

// ParseIntent handles intent parsing requests
func (h *AIHandler) ParseIntent(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "ai.ParseIntent")
	defer span.End()

	var request struct {
		Text string `json:"text"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	span.SetAttributes(attribute.String("text", request.Text))

	intent, err := h.nlpService.ParseIntent(ctx, request.Text)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Failed to parse intent", err)
		return
	}

	h.respondWithJSON(w, http.StatusOK, intent)
}

// ExtractEntities handles entity extraction requests
func (h *AIHandler) ExtractEntities(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "ai.ExtractEntities")
	defer span.End()

	var request struct {
		Text string `json:"text"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	span.SetAttributes(attribute.String("text", request.Text))

	entities, err := h.nlpService.ExtractEntities(ctx, request.Text)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Failed to extract entities", err)
		return
	}

	h.respondWithJSON(w, http.StatusOK, entities)
}

// AnalyzeSentiment handles sentiment analysis requests
func (h *AIHandler) AnalyzeSentiment(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "ai.AnalyzeSentiment")
	defer span.End()

	var request struct {
		Text string `json:"text"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	span.SetAttributes(attribute.String("text", request.Text))

	sentiment, err := h.nlpService.AnalyzeSentiment(ctx, request.Text)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Failed to analyze sentiment", err)
		return
	}

	h.respondWithJSON(w, http.StatusOK, sentiment)
}

// SummarizeText handles text summarization requests
func (h *AIHandler) SummarizeText(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.tracer.Start(r.Context(), "ai.SummarizeText")
	defer span.End()

	var request struct {
		Text      string `json:"text"`
		MaxLength int    `json:"max_length,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	span.SetAttributes(
		attribute.String("text", request.Text[:min(100, len(request.Text))]),
		attribute.Int("max_length", request.MaxLength),
	)

	summary, err := h.llmService.Summarize(ctx, request.Text)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Failed to summarize text", err)
		return
	}

	h.respondWithJSON(w, http.StatusOK, summary)
}

// Helper functions

func (h *AIHandler) readAudioFromRequest(r *http.Request) ([]byte, error) {
	// Check content type
	contentType := r.Header.Get("Content-Type")
	if contentType != "audio/wav" && contentType != "audio/mp3" && contentType != "application/octet-stream" {
		return nil, fmt.Errorf("unsupported content type: %s", contentType)
	}

	// Read audio data
	audioData, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read request body: %w", err)
	}

	if len(audioData) == 0 {
		return nil, fmt.Errorf("empty audio data")
	}

	return audioData, nil
}

func (h *AIHandler) readImageFromRequest(r *http.Request) ([]byte, error) {
	// Check content type
	contentType := r.Header.Get("Content-Type")
	if contentType != "image/jpeg" && contentType != "image/png" && contentType != "application/octet-stream" {
		return nil, fmt.Errorf("unsupported content type: %s", contentType)
	}

	// Read image data
	imageData, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read request body: %w", err)
	}

	if len(imageData) == 0 {
		return nil, fmt.Errorf("empty image data")
	}

	return imageData, nil
}

func (h *AIHandler) respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {
		h.logger.WithError(err).Error("Failed to marshal JSON response")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

func (h *AIHandler) respondWithError(w http.ResponseWriter, code int, message string, err error) {
	h.logger.WithError(err).Error(message)

	errorResponse := map[string]interface{}{
		"error":     message,
		"timestamp": time.Now().UTC(),
	}

	if err != nil {
		errorResponse["details"] = err.Error()
	}

	h.respondWithJSON(w, code, errorResponse)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
