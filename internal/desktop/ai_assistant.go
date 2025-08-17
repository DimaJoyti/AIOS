package desktop

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/aios/aios/internal/ai"
	"github.com/aios/aios/internal/ai/security"
	"github.com/aios/aios/pkg/models"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

// AIAssistant provides the main AI assistant interface for the desktop
type AIAssistant struct {
	logger         *logrus.Logger
	tracer         trace.Tracer
	llmService     ai.LanguageModelService
	voiceService   ai.VoiceService
	nlpService     ai.NaturalLanguageService
	cvService      ai.ComputerVisionService
	authMiddleware *security.AuthMiddleware

	// State management
	isListening      bool
	conversationMode bool
	currentSession   *AssistantSession
	mu               sync.RWMutex

	// Configuration
	config AssistantConfig

	// Event channels
	voiceEvents    chan VoiceEvent
	textEvents     chan TextEvent
	responseEvents chan ResponseEvent
	systemEvents   chan SystemEvent
}

// AssistantConfig represents configuration for the AI assistant
type AssistantConfig struct {
	WakeWord            string        `json:"wake_word"`
	VoiceEnabled        bool          `json:"voice_enabled"`
	ContinuousListening bool          `json:"continuous_listening"`
	ResponseTimeout     time.Duration `json:"response_timeout"`
	MaxConversationAge  time.Duration `json:"max_conversation_age"`
	DefaultPersonality  string        `json:"default_personality"`
	SystemPrompt        string        `json:"system_prompt"`
	EnableScreenContext bool          `json:"enable_screen_context"`
	EnableFileContext   bool          `json:"enable_file_context"`
}

// AssistantSession represents an active conversation session
type AssistantSession struct {
	ID           string                 `json:"id"`
	UserID       string                 `json:"user_id"`
	StartTime    time.Time              `json:"start_time"`
	LastActivity time.Time              `json:"last_activity"`
	Messages     []models.ChatMessage   `json:"messages"`
	Context      map[string]interface{} `json:"context"`
	Personality  string                 `json:"personality"`
	mu           sync.RWMutex
}

// Event types for the assistant
type VoiceEvent struct {
	Type      string    `json:"type"` // "wake_word", "speech_start", "speech_end", "audio_data"
	AudioData []byte    `json:"audio_data,omitempty"`
	Text      string    `json:"text,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

type TextEvent struct {
	Type      string                 `json:"type"` // "user_input", "command"
	Text      string                 `json:"text"`
	Context   map[string]interface{} `json:"context,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
}

type ResponseEvent struct {
	Type      string                 `json:"type"` // "text", "voice", "action", "error"
	Text      string                 `json:"text,omitempty"`
	AudioData []byte                 `json:"audio_data,omitempty"`
	Actions   []AssistantAction      `json:"actions,omitempty"`
	Error     string                 `json:"error,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
}

type SystemEvent struct {
	Type      string                 `json:"type"` // "session_start", "session_end", "mode_change", "error"
	Message   string                 `json:"message"`
	Data      map[string]interface{} `json:"data,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
}

// AssistantAction represents an action the assistant can perform
type AssistantAction struct {
	Type       string                 `json:"type"`
	Target     string                 `json:"target"`
	Parameters map[string]interface{} `json:"parameters"`
	Confidence float64                `json:"confidence"`
}

// NewAIAssistant creates a new AI assistant instance
func NewAIAssistant(
	logger *logrus.Logger,
	llmService ai.LanguageModelService,
	voiceService ai.VoiceService,
	nlpService ai.NaturalLanguageService,
	cvService ai.ComputerVisionService,
	authMiddleware *security.AuthMiddleware,
	config AssistantConfig,
) *AIAssistant {
	assistant := &AIAssistant{
		logger:         logger,
		tracer:         otel.Tracer("desktop.ai_assistant"),
		llmService:     llmService,
		voiceService:   voiceService,
		nlpService:     nlpService,
		cvService:      cvService,
		authMiddleware: authMiddleware,
		config:         config,
		voiceEvents:    make(chan VoiceEvent, 100),
		textEvents:     make(chan TextEvent, 100),
		responseEvents: make(chan ResponseEvent, 100),
		systemEvents:   make(chan SystemEvent, 100),
	}

	// Start event processing
	go assistant.processEvents()

	return assistant
}

// StartListening begins voice listening mode
func (a *AIAssistant) StartListening(ctx context.Context, userID string) error {
	ctx, span := a.tracer.Start(ctx, "ai_assistant.StartListening")
	defer span.End()

	a.mu.Lock()
	defer a.mu.Unlock()

	if a.isListening {
		return fmt.Errorf("already listening")
	}

	a.isListening = true
	a.logger.WithField("user_id", userID).Info("AI Assistant started listening")

	// Create or resume session
	if a.currentSession == nil || a.currentSession.UserID != userID {
		a.currentSession = &AssistantSession{
			ID:           generateSessionID(),
			UserID:       userID,
			StartTime:    time.Now(),
			LastActivity: time.Now(),
			Messages:     make([]models.ChatMessage, 0),
			Context:      make(map[string]interface{}),
			Personality:  a.config.DefaultPersonality,
		}

		// Add system prompt
		if a.config.SystemPrompt != "" {
			a.currentSession.Messages = append(a.currentSession.Messages, models.ChatMessage{
				Role:    "system",
				Content: a.config.SystemPrompt,
			})
		}
	}

	// Send system event
	a.systemEvents <- SystemEvent{
		Type:      "session_start",
		Message:   "AI Assistant listening started",
		Data:      map[string]interface{}{"user_id": userID, "session_id": a.currentSession.ID},
		Timestamp: time.Now(),
	}

	return nil
}

// StopListening stops voice listening mode
func (a *AIAssistant) StopListening(ctx context.Context) error {
	ctx, span := a.tracer.Start(ctx, "ai_assistant.StopListening")
	defer span.End()

	a.mu.Lock()
	defer a.mu.Unlock()

	if !a.isListening {
		return fmt.Errorf("not currently listening")
	}

	a.isListening = false
	a.logger.Info("AI Assistant stopped listening")

	// Send system event
	a.systemEvents <- SystemEvent{
		Type:      "session_pause",
		Message:   "AI Assistant listening stopped",
		Timestamp: time.Now(),
	}

	return nil
}

// ProcessVoiceInput processes voice input from the user
func (a *AIAssistant) ProcessVoiceInput(ctx context.Context, audioData []byte) error {
	ctx, span := a.tracer.Start(ctx, "ai_assistant.ProcessVoiceInput")
	defer span.End()

	if !a.config.VoiceEnabled {
		return fmt.Errorf("voice input is disabled")
	}

	// Send voice event
	a.voiceEvents <- VoiceEvent{
		Type:      "audio_data",
		AudioData: audioData,
		Timestamp: time.Now(),
	}

	return nil
}

// ProcessTextInput processes text input from the user
func (a *AIAssistant) ProcessTextInput(ctx context.Context, text string, context map[string]interface{}) error {
	ctx, span := a.tracer.Start(ctx, "ai_assistant.ProcessTextInput")
	defer span.End()

	a.logger.WithField("text", text[:minInt(100, len(text))]).Info("Processing text input")

	// Send text event
	a.textEvents <- TextEvent{
		Type:      "user_input",
		Text:      text,
		Context:   context,
		Timestamp: time.Now(),
	}

	return nil
}

// GetResponseEvents returns the response events channel
func (a *AIAssistant) GetResponseEvents() <-chan ResponseEvent {
	return a.responseEvents
}

// GetSystemEvents returns the system events channel
func (a *AIAssistant) GetSystemEvents() <-chan SystemEvent {
	return a.systemEvents
}

// GetCurrentSession returns the current session information
func (a *AIAssistant) GetCurrentSession() *AssistantSession {
	a.mu.RLock()
	defer a.mu.RUnlock()

	if a.currentSession == nil {
		return nil
	}

	// Return a copy without the mutex
	session := AssistantSession{
		ID:           a.currentSession.ID,
		UserID:       a.currentSession.UserID,
		StartTime:    a.currentSession.StartTime,
		LastActivity: a.currentSession.LastActivity,
		Messages:     make([]models.ChatMessage, len(a.currentSession.Messages)),
		Context:      make(map[string]interface{}),
		Personality:  a.currentSession.Personality,
	}

	copy(session.Messages, a.currentSession.Messages)
	for k, v := range a.currentSession.Context {
		session.Context[k] = v
	}

	return &session
}

// Event processing

func (a *AIAssistant) processEvents() {
	for {
		select {
		case voiceEvent := <-a.voiceEvents:
			a.handleVoiceEvent(voiceEvent)
		case textEvent := <-a.textEvents:
			a.handleTextEvent(textEvent)
		}
	}
}

func (a *AIAssistant) handleVoiceEvent(event VoiceEvent) {
	ctx := context.Background()
	ctx, span := a.tracer.Start(ctx, "ai_assistant.handleVoiceEvent")
	defer span.End()

	switch event.Type {
	case "audio_data":
		// Convert speech to text
		recognition, err := a.voiceService.SpeechToText(ctx, event.AudioData)
		if err != nil {
			a.logger.WithError(err).Error("Failed to convert speech to text")
			a.responseEvents <- ResponseEvent{
				Type:      "error",
				Error:     "Failed to process voice input",
				Timestamp: time.Now(),
			}
			return
		}

		if recognition.Text != "" {
			// Process the recognized text
			a.handleTextEvent(TextEvent{
				Type:      "user_input",
				Text:      recognition.Text,
				Context:   map[string]interface{}{"source": "voice", "confidence": recognition.Confidence},
				Timestamp: time.Now(),
			})
		}
	}
}

func (a *AIAssistant) handleTextEvent(event TextEvent) {
	ctx := context.Background()
	ctx, span := a.tracer.Start(ctx, "ai_assistant.handleTextEvent")
	defer span.End()

	switch event.Type {
	case "user_input":
		a.processUserInput(ctx, event.Text, event.Context)
	case "command":
		a.processCommand(ctx, event.Text, event.Context)
	}
}

func (a *AIAssistant) processUserInput(ctx context.Context, text string, inputContext map[string]interface{}) {
	a.mu.Lock()
	if a.currentSession == nil {
		a.mu.Unlock()
		a.logger.Error("No active session for user input")
		return
	}

	// Add user message to conversation
	a.currentSession.Messages = append(a.currentSession.Messages, models.ChatMessage{
		Role:    "user",
		Content: text,
	})
	a.currentSession.LastActivity = time.Now()
	a.mu.Unlock()

	// Analyze intent first
	intent, err := a.nlpService.ParseIntent(ctx, text)
	if err != nil {
		a.logger.WithError(err).Error("Failed to analyze intent")
	}

	// Gather context if enabled
	var contextInfo strings.Builder
	if a.config.EnableScreenContext {
		// TODO: Capture and analyze current screen
		contextInfo.WriteString("Current screen context available. ")
	}

	if a.config.EnableFileContext {
		// TODO: Analyze current file/directory context
		contextInfo.WriteString("File system context available. ")
	}

	// Prepare enhanced prompt
	enhancedText := text
	if contextInfo.Len() > 0 {
		enhancedText = fmt.Sprintf("Context: %s\nUser: %s", contextInfo.String(), text)
	}

	// Generate response using LLM
	a.mu.RLock()
	messages := make([]models.ChatMessage, len(a.currentSession.Messages))
	copy(messages, a.currentSession.Messages)
	a.mu.RUnlock()

	// Update the last message with enhanced context
	if len(messages) > 0 {
		messages[len(messages)-1].Content = enhancedText
	}

	response, err := a.llmService.ChatWithHistory(ctx, messages)
	if err != nil {
		a.logger.WithError(err).Error("Failed to generate response")
		a.responseEvents <- ResponseEvent{
			Type:      "error",
			Error:     "Failed to generate response",
			Timestamp: time.Now(),
		}
		return
	}

	// Add assistant response to conversation
	a.mu.Lock()
	a.currentSession.Messages = append(a.currentSession.Messages, models.ChatMessage{
		Role:      "assistant",
		Content:   response.Message,
		Timestamp: time.Now(),
	})
	a.mu.Unlock()

	// Determine response actions
	actions := a.extractActions(intent, response.Message)

	// Send response event
	responseEvent := ResponseEvent{
		Type:    "text",
		Text:    response.Message,
		Actions: actions,
		Metadata: map[string]interface{}{
			"intent": intent,
			"tokens": response.Context["tokens"],
			"model":  response.Context["model"],
			"source": inputContext["source"],
		},
		Timestamp: time.Now(),
	}

	// Generate voice response if voice input was used
	if source, ok := inputContext["source"]; ok && source == "voice" && a.config.VoiceEnabled {
		synthesis, err := a.voiceService.TextToSpeech(ctx, response.Message)
		if err != nil {
			a.logger.WithError(err).Error("Failed to generate voice response")
		} else {
			responseEvent.Type = "voice"
			responseEvent.AudioData = synthesis.Audio
		}
	}

	a.responseEvents <- responseEvent
}

func (a *AIAssistant) processCommand(ctx context.Context, command string, context map[string]interface{}) {
	// Process system commands
	switch strings.ToLower(command) {
	case "clear_conversation":
		a.clearConversation()
	case "change_personality":
		if personality, ok := context["personality"].(string); ok {
			a.changePersonality(personality)
		}
	case "enable_voice":
		a.config.VoiceEnabled = true
	case "disable_voice":
		a.config.VoiceEnabled = false
	}
}

func (a *AIAssistant) extractActions(intent *models.IntentAnalysis, responseText string) []AssistantAction {
	var actions []AssistantAction

	if intent == nil {
		return actions
	}

	// Extract actions based on intent
	switch intent.Intent {
	case "open_application":
		for _, entity := range intent.Entities {
			if entity.Type == "application" {
				actions = append(actions, AssistantAction{
					Type:       "open_application",
					Target:     entity.Text,
					Parameters: map[string]interface{}{},
					Confidence: entity.Confidence,
				})
			}
		}
	case "system_control":
		actions = append(actions, AssistantAction{
			Type:       "system_action",
			Target:     "system",
			Parameters: map[string]interface{}{"intent": intent.Intent},
			Confidence: intent.Confidence,
		})
	case "file_operation":
		actions = append(actions, AssistantAction{
			Type:       "file_action",
			Target:     "filesystem",
			Parameters: map[string]interface{}{"entities": intent.Entities},
			Confidence: intent.Confidence,
		})
	}

	return actions
}

func (a *AIAssistant) clearConversation() {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.currentSession != nil {
		a.currentSession.Messages = a.currentSession.Messages[:1] // Keep system prompt
		a.logger.Info("Conversation cleared")
	}
}

func (a *AIAssistant) changePersonality(personality string) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.currentSession != nil {
		a.currentSession.Personality = personality
		a.logger.WithField("personality", personality).Info("Personality changed")
	}
}

// Helper functions

func generateSessionID() string {
	return fmt.Sprintf("session_%d", time.Now().UnixNano())
}

// min function removed - use minInt instead
