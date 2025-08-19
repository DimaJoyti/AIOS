package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/aios/aios/internal/ai/providers"
	"github.com/aios/aios/internal/ai/services"
	"github.com/aios/aios/pkg/config"
	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

// Service provides AI functionality
type Service struct {
	config       *config.Config
	logger       *logrus.Logger
	tracer       trace.Tracer
	db           *sqlx.DB
	orchestrator *services.AIOrchestrator
	httpServer   *http.Server
}

// NewService creates a new AI service instance
func NewService(config *config.Config, db *sqlx.DB, logger *logrus.Logger) (*Service, error) {
	tracer := otel.Tracer("ai.service")

	// Create orchestrator configuration
	orchestratorConfig := &services.OrchestratorConfig{
		DefaultModel:        "gpt-3.5-turbo",
		MaxConcurrentTasks:  10,
		DefaultTimeout:      60 * time.Second,
		EnableSafetyFilter:  true,
		EnableAnalytics:     true,
		CacheEnabled:        true,
		CacheTTL:            1 * time.Hour,
		RateLimitPerMinute:  100,
		CostLimitPerHour:    10.0,
		EnableLoadBalancing: true,
	}

	// Create AI orchestrator
	orchestrator := services.NewAIOrchestrator(orchestratorConfig, logger)

	// Register AI providers
	if err := registerProviders(orchestrator, logger); err != nil {
		return nil, fmt.Errorf("failed to register AI providers: %w", err)
	}

	// Register default models
	if err := registerModels(orchestrator, logger); err != nil {
		return nil, fmt.Errorf("failed to register AI models: %w", err)
	}

	// Register default prompt templates
	if err := registerPromptTemplates(orchestrator, logger); err != nil {
		return nil, fmt.Errorf("failed to register prompt templates: %w", err)
	}

	return &Service{
		config:       config,
		logger:       logger,
		tracer:       tracer,
		db:           db,
		orchestrator: orchestrator,
	}, nil
}

// Start starts the AI service
func (s *Service) Start(ctx context.Context) error {
	s.logger.Info("Starting AI Service...")

	// Setup HTTP routes
	router := mux.NewRouter()

	// AI endpoints
	router.HandleFunc("/api/v1/ai/generate", s.handleGenerate).Methods("POST")
	router.HandleFunc("/api/v1/ai/chat", s.handleChat).Methods("POST")
	router.HandleFunc("/api/v1/ai/templates", s.handleTemplates).Methods("GET", "POST")
	router.HandleFunc("/api/v1/ai/templates/{id}", s.handleTemplate).Methods("GET", "PUT", "DELETE")
	router.HandleFunc("/api/v1/ai/templates/{id}/execute", s.handleTemplateExecute).Methods("POST")
	router.HandleFunc("/api/v1/ai/chains", s.handleChains).Methods("GET", "POST")
	router.HandleFunc("/api/v1/ai/chains/{id}/execute", s.handleChainExecute).Methods("POST")
	router.HandleFunc("/api/v1/ai/models", s.handleModels).Methods("GET")
	router.HandleFunc("/api/v1/ai/models/{id}", s.handleModel).Methods("GET", "PUT")
	router.HandleFunc("/api/v1/ai/providers", s.handleProviders).Methods("GET")
	router.HandleFunc("/api/v1/ai/analytics", s.handleAnalytics).Methods("GET")
	router.HandleFunc("/api/v1/ai/health", s.handleHealth).Methods("GET")

	// Health check endpoint
	router.HandleFunc("/health", s.handleHealthCheck).Methods("GET")

	// Create HTTP server
	s.httpServer = &http.Server{
		Addr:    fmt.Sprintf(":%d", 8182), // Default AI service port
		Handler: router,
	}

	// Start HTTP server
	go func() {
		s.logger.WithField("port", 8182).Info("AI HTTP server starting")
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.WithError(err).Error("AI HTTP server error")
		}
	}()

	s.logger.Info("AI Service started successfully")
	return nil
}

// Stop stops the AI service
func (s *Service) Stop(ctx context.Context) error {
	s.logger.Info("Stopping AI Service...")

	// Stop HTTP server
	if s.httpServer != nil {
		if err := s.httpServer.Shutdown(ctx); err != nil {
			s.logger.WithError(err).Error("Error stopping AI HTTP server")
		}
	}

	s.logger.Info("AI Service stopped")
	return nil
}

// handleGenerate handles text generation requests
func (s *Service) handleGenerate(w http.ResponseWriter, r *http.Request) {
	ctx, span := s.tracer.Start(r.Context(), "ai.service.handle_generate")
	defer span.End()

	var request services.AIRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Set request type if not specified
	if request.Type == "" {
		request.Type = "text"
	}

	// Process request
	response, err := s.orchestrator.ProcessRequest(ctx, &request)
	if err != nil {
		s.logger.WithError(err).Error("Failed to process AI request")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleChat handles chat completion requests
func (s *Service) handleChat(w http.ResponseWriter, r *http.Request) {
	ctx, span := s.tracer.Start(r.Context(), "ai.service.handle_chat")
	defer span.End()

	var request struct {
		Messages     []services.ChatMessage `json:"messages"`
		ModelID      string                 `json:"model_id,omitempty"`
		SystemPrompt string                 `json:"system_prompt,omitempty"`
		Config       *services.ModelConfig  `json:"config,omitempty"`
		UserID       string                 `json:"user_id,omitempty"`
		SessionID    string                 `json:"session_id,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Convert to AI request
	aiRequest := &services.AIRequest{
		Type:         "text",
		ModelID:      request.ModelID,
		SystemPrompt: request.SystemPrompt,
		Messages:     request.Messages,
		Config:       request.Config,
		UserID:       request.UserID,
		SessionID:    request.SessionID,
	}

	// Process request
	response, err := s.orchestrator.ProcessRequest(ctx, aiRequest)
	if err != nil {
		s.logger.WithError(err).Error("Failed to process chat request")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleTemplates handles prompt template requests
func (s *Service) handleTemplates(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		templates := s.orchestrator.GetPromptManager().ListTemplates()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(templates)

	case "POST":
		var template services.PromptTemplate
		if err := json.NewDecoder(r.Body).Decode(&template); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		if err := s.orchestrator.GetPromptManager().CreateTemplate(&template); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(template)
	}
}

// handleTemplate handles individual template requests
func (s *Service) handleTemplate(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	templateID := vars["id"]

	switch r.Method {
	case "GET":
		template, err := s.orchestrator.GetPromptManager().GetTemplate(templateID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(template)

	case "PUT":
		var template services.PromptTemplate
		if err := json.NewDecoder(r.Body).Decode(&template); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		template.ID = templateID
		if err := s.orchestrator.GetPromptManager().CreateTemplate(&template); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(template)

	case "DELETE":
		// Implementation for template deletion
		w.WriteHeader(http.StatusNoContent)
	}
}

// handleTemplateExecute handles template execution requests
func (s *Service) handleTemplateExecute(w http.ResponseWriter, r *http.Request) {
	ctx, span := s.tracer.Start(r.Context(), "ai.service.handle_template_execute")
	defer span.End()

	vars := mux.Vars(r)
	templateID := vars["id"]

	var request struct {
		Variables map[string]interface{} `json:"variables"`
		UserID    string                 `json:"user_id,omitempty"`
		SessionID string                 `json:"session_id,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Create AI request
	aiRequest := &services.AIRequest{
		Type:       "text",
		TemplateID: templateID,
		Variables:  request.Variables,
		UserID:     request.UserID,
		SessionID:  request.SessionID,
	}

	// Process request
	response, err := s.orchestrator.ProcessRequest(ctx, aiRequest)
	if err != nil {
		s.logger.WithError(err).Error("Failed to execute template")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleChains handles prompt chain requests
func (s *Service) handleChains(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		// Implementation for listing chains
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]interface{}{})

	case "POST":
		var chain services.PromptChain
		if err := json.NewDecoder(r.Body).Decode(&chain); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		if err := s.orchestrator.GetPromptManager().CreateChain(&chain); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(chain)
	}
}

// handleChainExecute handles chain execution requests
func (s *Service) handleChainExecute(w http.ResponseWriter, r *http.Request) {
	ctx, span := s.tracer.Start(r.Context(), "ai.service.handle_chain_execute")
	defer span.End()

	vars := mux.Vars(r)
	chainID := vars["id"]

	var request struct {
		Variables map[string]interface{} `json:"variables"`
		UserID    string                 `json:"user_id,omitempty"`
		SessionID string                 `json:"session_id,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Create AI request
	aiRequest := &services.AIRequest{
		Type:      "text",
		ChainID:   chainID,
		Variables: request.Variables,
		UserID:    request.UserID,
		SessionID: request.SessionID,
	}

	// Process request
	response, err := s.orchestrator.ProcessRequest(ctx, aiRequest)
	if err != nil {
		s.logger.WithError(err).Error("Failed to execute chain")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleModels handles model management requests
func (s *Service) handleModels(w http.ResponseWriter, r *http.Request) {
	models := s.orchestrator.GetModelManager().ListModels()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models)
}

// handleModel handles individual model requests
func (s *Service) handleModel(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	modelID := vars["id"]

	switch r.Method {
	case "GET":
		model, err := s.orchestrator.GetModelManager().GetModel(modelID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(model)

	case "PUT":
		var config services.ModelConfig
		if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		if err := s.orchestrator.GetModelManager().UpdateModelConfig(modelID, config); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

// handleProviders handles provider status requests
func (s *Service) handleProviders(w http.ResponseWriter, r *http.Request) {
	health := s.orchestrator.GetModelManager().GetProviderHealth()
	usage := s.orchestrator.GetModelManager().GetUsageStats()

	response := map[string]interface{}{
		"health": health,
		"usage":  usage,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleAnalytics handles analytics requests
func (s *Service) handleAnalytics(w http.ResponseWriter, r *http.Request) {
	analytics := s.orchestrator.GetAnalytics()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(analytics)
}

// handleHealth handles health check requests
func (s *Service) handleHealth(w http.ResponseWriter, r *http.Request) {
	health := s.orchestrator.GetHealth()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(health)
}

// handleHealthCheck handles simple health check
func (s *Service) handleHealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "healthy",
		"service": "ai",
	})
}

// registerProviders registers AI providers
func registerProviders(orchestrator *services.AIOrchestrator, logger *logrus.Logger) error {
	// Register OpenAI provider if API key is available
	if apiKey := getEnvOrDefault("OPENAI_API_KEY", ""); apiKey != "" {
		openaiProvider := providers.NewOpenAIProvider(apiKey, logger)
		if err := orchestrator.GetModelManager().RegisterProvider(openaiProvider); err != nil {
			return fmt.Errorf("failed to register OpenAI provider: %w", err)
		}
		logger.Info("OpenAI provider registered")
	}

	// Register other providers as needed
	// anthropicProvider := providers.NewAnthropicProvider(apiKey, logger)
	// localProvider := providers.NewLocalProvider(config, logger)

	return nil
}

// registerModels registers default AI models
func registerModels(orchestrator *services.AIOrchestrator, logger *logrus.Logger) error {
	models := []*services.AIModel{
		{
			ID:           "gpt-4",
			Name:         "GPT-4",
			Provider:     "openai",
			Type:         "text",
			Version:      "1.0",
			Capabilities: []string{"text_generation", "chat", "reasoning"},
			Config: services.ModelConfig{
				MaxTokens:   4096,
				Temperature: 0.7,
			},
			Limits: services.ModelLimits{
				RequestsPerMinute: 60,
				TokensPerMinute:   40000,
				TimeoutSeconds:    60,
			},
			Pricing: services.ModelPricing{
				InputTokenCost:  0.03 / 1000,
				OutputTokenCost: 0.06 / 1000,
				Currency:        "USD",
			},
			Status: "active",
		},
		{
			ID:           "gpt-3.5-turbo",
			Name:         "GPT-3.5 Turbo",
			Provider:     "openai",
			Type:         "text",
			Version:      "1.0",
			Capabilities: []string{"text_generation", "chat"},
			Config: services.ModelConfig{
				MaxTokens:   4096,
				Temperature: 0.7,
			},
			Limits: services.ModelLimits{
				RequestsPerMinute: 100,
				TokensPerMinute:   90000,
				TimeoutSeconds:    30,
			},
			Pricing: services.ModelPricing{
				InputTokenCost:  0.0015 / 1000,
				OutputTokenCost: 0.002 / 1000,
				Currency:        "USD",
			},
			Status: "active",
		},
	}

	for _, model := range models {
		if err := orchestrator.GetModelManager().RegisterModel(model); err != nil {
			return fmt.Errorf("failed to register model %s: %w", model.ID, err)
		}
	}

	logger.WithField("count", len(models)).Info("AI models registered")
	return nil
}

// registerPromptTemplates registers default prompt templates
func registerPromptTemplates(orchestrator *services.AIOrchestrator, logger *logrus.Logger) error {
	templates := []*services.PromptTemplate{
		{
			ID:          "summarize",
			Name:        "Text Summarization",
			Description: "Summarizes long text into key points",
			Category:    "text_processing",
			Template:    "Please summarize the following text in {{.max_points}} key points:\n\n{{.text}}",
			Variables: []services.PromptVariable{
				{
					Name:        "text",
					Type:        "string",
					Description: "Text to summarize",
					Required:    true,
				},
				{
					Name:         "max_points",
					Type:         "number",
					Description:  "Maximum number of summary points",
					Required:     false,
					DefaultValue: 5,
				},
			},
			Config: services.PromptConfig{
				ModelID:     "gpt-3.5-turbo",
				Temperature: 0.3,
				MaxTokens:   500,
			},
			Tags: []string{"summarization", "text_processing"},
		},
		{
			ID:          "translate",
			Name:        "Language Translation",
			Description: "Translates text between languages",
			Category:    "translation",
			Template:    "Translate the following text from {{.source_language}} to {{.target_language}}:\n\n{{.text}}",
			Variables: []services.PromptVariable{
				{
					Name:        "text",
					Type:        "string",
					Description: "Text to translate",
					Required:    true,
				},
				{
					Name:        "source_language",
					Type:        "string",
					Description: "Source language",
					Required:    true,
				},
				{
					Name:        "target_language",
					Type:        "string",
					Description: "Target language",
					Required:    true,
				},
			},
			Config: services.PromptConfig{
				ModelID:     "gpt-3.5-turbo",
				Temperature: 0.1,
				MaxTokens:   1000,
			},
			Tags: []string{"translation", "language"},
		},
	}

	for _, template := range templates {
		if err := orchestrator.GetPromptManager().CreateTemplate(template); err != nil {
			return fmt.Errorf("failed to register template %s: %w", template.ID, err)
		}
	}

	logger.WithField("count", len(templates)).Info("Prompt templates registered")
	return nil
}

// getEnvOrDefault gets environment variable or returns default value
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
