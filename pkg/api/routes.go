package api

import (
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/aios/aios/internal/system"
	"github.com/aios/aios/pkg/models"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

// RegisterSystemRoutes registers all system-related API routes
func RegisterSystemRoutes(router *mux.Router, systemManager *system.Manager, logger *logrus.Logger) {
	tracer := otel.Tracer("api-routes")

	api := router.PathPrefix("/api/v1").Subrouter()

	// System status endpoints
	api.HandleFunc("/system/status", handleSystemStatus(systemManager, logger, tracer)).Methods("GET")
	api.HandleFunc("/system/resources", handleResourceStatus(systemManager, logger, tracer)).Methods("GET")
	api.HandleFunc("/system/optimize", handleOptimizeSystem(systemManager, logger, tracer)).Methods("POST")

	// Resource management endpoints
	api.HandleFunc("/resources/cpu", handleCPUInfo(systemManager, logger, tracer)).Methods("GET")
	api.HandleFunc("/resources/memory", handleMemoryInfo(systemManager, logger, tracer)).Methods("GET")
	api.HandleFunc("/resources/disk", handleDiskInfo(systemManager, logger, tracer)).Methods("GET")

	// Security endpoints
	api.HandleFunc("/security/status", handleSecurityStatus(systemManager, logger, tracer)).Methods("GET")
	api.HandleFunc("/security/threats", handleThreatAnalysis(systemManager, logger, tracer)).Methods("GET")

	// File system AI endpoints
	api.HandleFunc("/filesystem/analyze", handleFileSystemAnalysis(systemManager, logger, tracer)).Methods("POST")
	api.HandleFunc("/filesystem/organize", handleFileSystemOrganization(systemManager, logger, tracer)).Methods("POST")

	// AI services endpoints
	api.HandleFunc("/ai/chat", handleAIChat(systemManager, logger, tracer)).Methods("POST")
	api.HandleFunc("/ai/vision", handleAIVision(systemManager, logger, tracer)).Methods("POST")
	api.HandleFunc("/ai/optimize", handleAIOptimize(systemManager, logger, tracer)).Methods("POST")
	api.HandleFunc("/ai/status", handleAIStatus(systemManager, logger, tracer)).Methods("GET")
	api.HandleFunc("/ai/models", handleAIModels(systemManager, logger, tracer)).Methods("GET")
	api.HandleFunc("/ai/workflow", handleAIWorkflow(systemManager, logger, tracer)).Methods("POST")
}

// RegisterHealthRoutes registers health check endpoints
func RegisterHealthRoutes(router *mux.Router, logger *logrus.Logger) {
	router.HandleFunc("/health", handleHealthCheck(logger)).Methods("GET")
	router.HandleFunc("/ready", handleReadinessCheck(logger)).Methods("GET")
	router.HandleFunc("/version", handleVersionCheck(logger)).Methods("GET")
}

// handleSystemStatus returns the overall system status
func handleSystemStatus(systemManager *system.Manager, logger *logrus.Logger, tracer trace.Tracer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, span := tracer.Start(r.Context(), "api.handleSystemStatus")
		defer span.End()

		status, err := systemManager.GetSystemStatus(ctx)
		if err != nil {
			logger.WithError(err).Error("Failed to get system status")
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(status); err != nil {
			logger.WithError(err).Error("Failed to encode system status response")
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	}
}

// handleResourceStatus returns detailed resource information
func handleResourceStatus(systemManager *system.Manager, logger *logrus.Logger, tracer trace.Tracer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, span := tracer.Start(r.Context(), "api.handleResourceStatus")
		defer span.End()

		resourceManager := systemManager.GetResourceManager()
		status, err := resourceManager.GetStatus(ctx)
		if err != nil {
			logger.WithError(err).Error("Failed to get resource status")
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(status); err != nil {
			logger.WithError(err).Error("Failed to encode resource status response")
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	}
}

// handleOptimizeSystem triggers system optimization
func handleOptimizeSystem(systemManager *system.Manager, logger *logrus.Logger, tracer trace.Tracer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, span := tracer.Start(r.Context(), "api.handleOptimizeSystem")
		defer span.End()

		optimizationAI := systemManager.GetOptimizationAI()
		if err := optimizationAI.RunOptimization(ctx); err != nil {
			logger.WithError(err).Error("Failed to run system optimization")
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		response := map[string]interface{}{
			"message":   "System optimization started",
			"timestamp": time.Now(),
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			logger.WithError(err).Error("Failed to encode optimization response")
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	}
}

// handleCPUInfo returns detailed CPU information
func handleCPUInfo(systemManager *system.Manager, logger *logrus.Logger, tracer trace.Tracer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, span := tracer.Start(r.Context(), "api.handleCPUInfo")
		defer span.End()

		resourceManager := systemManager.GetResourceManager()
		cpuInfo, err := resourceManager.GetCPUInfo(ctx)
		if err != nil {
			logger.WithError(err).Error("Failed to get CPU information")
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(cpuInfo); err != nil {
			logger.WithError(err).Error("Failed to encode CPU info response")
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	}
}

// handleMemoryInfo returns detailed memory information
func handleMemoryInfo(systemManager *system.Manager, logger *logrus.Logger, tracer trace.Tracer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, span := tracer.Start(r.Context(), "api.handleMemoryInfo")
		defer span.End()

		resourceManager := systemManager.GetResourceManager()
		memoryInfo, err := resourceManager.GetMemoryInfo(ctx)
		if err != nil {
			logger.WithError(err).Error("Failed to get memory information")
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(memoryInfo); err != nil {
			logger.WithError(err).Error("Failed to encode memory info response")
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	}
}

// handleDiskInfo returns detailed disk information
func handleDiskInfo(systemManager *system.Manager, logger *logrus.Logger, tracer trace.Tracer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, span := tracer.Start(r.Context(), "api.handleDiskInfo")
		defer span.End()

		resourceManager := systemManager.GetResourceManager()
		diskInfo, err := resourceManager.GetDiskInfo(ctx)
		if err != nil {
			logger.WithError(err).Error("Failed to get disk information")
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(diskInfo); err != nil {
			logger.WithError(err).Error("Failed to encode disk info response")
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	}
}

// handleSecurityStatus returns security status information
func handleSecurityStatus(systemManager *system.Manager, logger *logrus.Logger, tracer trace.Tracer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, span := tracer.Start(r.Context(), "api.handleSecurityStatus")
		defer span.End()

		securityManager := systemManager.GetSecurityManager()
		status, err := securityManager.GetStatus(ctx)
		if err != nil {
			logger.WithError(err).Error("Failed to get security status")
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(status); err != nil {
			logger.WithError(err).Error("Failed to encode security status response")
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	}
}

// handleThreatAnalysis returns threat analysis results
func handleThreatAnalysis(systemManager *system.Manager, logger *logrus.Logger, tracer trace.Tracer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, span := tracer.Start(r.Context(), "api.handleThreatAnalysis")
		defer span.End()

		securityManager := systemManager.GetSecurityManager()
		threats, err := securityManager.AnalyzeThreats(ctx)
		if err != nil {
			logger.WithError(err).Error("Failed to analyze threats")
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(threats); err != nil {
			logger.WithError(err).Error("Failed to encode threat analysis response")
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	}
}

// handleFileSystemAnalysis analyzes file system patterns
func handleFileSystemAnalysis(systemManager *system.Manager, logger *logrus.Logger, tracer trace.Tracer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, span := tracer.Start(r.Context(), "api.handleFileSystemAnalysis")
		defer span.End()

		var request struct {
			Path string `json:"path"`
		}

		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			logger.WithError(err).Error("Failed to decode file system analysis request")
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		fileSystemAI := systemManager.GetFileSystemAI()
		analysis, err := fileSystemAI.AnalyzePath(ctx, request.Path)
		if err != nil {
			logger.WithError(err).Error("Failed to analyze file system path")
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(analysis); err != nil {
			logger.WithError(err).Error("Failed to encode file system analysis response")
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	}
}

// handleFileSystemOrganization organizes files using AI
func handleFileSystemOrganization(systemManager *system.Manager, logger *logrus.Logger, tracer trace.Tracer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, span := tracer.Start(r.Context(), "api.handleFileSystemOrganization")
		defer span.End()

		var request struct {
			Path   string `json:"path"`
			DryRun bool   `json:"dry_run"`
		}

		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			logger.WithError(err).Error("Failed to decode file system organization request")
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		fileSystemAI := systemManager.GetFileSystemAI()
		result, err := fileSystemAI.OrganizePath(ctx, request.Path, request.DryRun)
		if err != nil {
			logger.WithError(err).Error("Failed to organize file system path")
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(result); err != nil {
			logger.WithError(err).Error("Failed to encode file system organization response")
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	}
}

// handleHealthCheck returns basic health status
func handleHealthCheck(logger *logrus.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"status":    "healthy",
			"timestamp": time.Now(),
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			logger.WithError(err).Error("Failed to encode health check response")
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	}
}

// handleReadinessCheck returns readiness status
func handleReadinessCheck(logger *logrus.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"status":    "ready",
			"timestamp": time.Now(),
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			logger.WithError(err).Error("Failed to encode readiness check response")
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	}
}

// handleVersionCheck returns version information
func handleVersionCheck(logger *logrus.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// These would be injected at build time
		response := map[string]interface{}{
			"version":    "dev",
			"commit":     "unknown",
			"build_time": "unknown",
			"timestamp":  time.Now(),
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			logger.WithError(err).Error("Failed to encode version check response")
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	}
}

// AI Service Handlers

// handleAIChat handles AI chat requests
func handleAIChat(systemManager *system.Manager, logger *logrus.Logger, tracer trace.Tracer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, span := tracer.Start(r.Context(), "api.handleAIChat")
		defer span.End()

		var request struct {
			Message        string                 `json:"message"`
			ConversationID string                 `json:"conversation_id,omitempty"`
			Context        map[string]interface{} `json:"context,omitempty"`
		}

		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			logger.WithError(err).Error("Failed to decode AI chat request")
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		// Create AI request
		aiRequest := &models.AIRequest{
			ID:        uuid.New().String(),
			Type:      "chat",
			Input:     request.Message,
			Context:   request.Context,
			Timeout:   30 * time.Second,
			Timestamp: time.Now(),
		}

		if request.ConversationID != "" {
			if aiRequest.Context == nil {
				aiRequest.Context = make(map[string]interface{})
			}
			aiRequest.Context["conversation_id"] = request.ConversationID
		}

		// Process request through AI orchestrator
		orchestrator := systemManager.GetAIOrchestrator()
		response, err := orchestrator.ProcessRequest(ctx, aiRequest)
		if err != nil {
			logger.WithError(err).Error("Failed to process AI chat request")
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			logger.WithError(err).Error("Failed to encode AI chat response")
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	}
}

// handleAIVision handles AI vision requests
func handleAIVision(systemManager *system.Manager, logger *logrus.Logger, tracer trace.Tracer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, span := tracer.Start(r.Context(), "api.handleAIVision")
		defer span.End()

		// Parse multipart form for image upload
		if err := r.ParseMultipartForm(10 << 20); err != nil { // 10MB max
			logger.WithError(err).Error("Failed to parse multipart form")
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		file, _, err := r.FormFile("image")
		if err != nil {
			logger.WithError(err).Error("Failed to get image file")
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}
		defer file.Close()

		// Read image data
		imageData, err := io.ReadAll(file)
		if err != nil {
			logger.WithError(err).Error("Failed to read image data")
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// Get task parameter
		task := r.FormValue("task")
		if task == "" {
			task = "analyze_screen"
		}

		// Create AI request
		aiRequest := &models.AIRequest{
			ID:    uuid.New().String(),
			Type:  "vision",
			Input: imageData,
			Parameters: map[string]interface{}{
				"task": task,
			},
			Timeout:   60 * time.Second,
			Timestamp: time.Now(),
		}

		// Process request through AI orchestrator
		orchestrator := systemManager.GetAIOrchestrator()
		response, err := orchestrator.ProcessRequest(ctx, aiRequest)
		if err != nil {
			logger.WithError(err).Error("Failed to process AI vision request")
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			logger.WithError(err).Error("Failed to encode AI vision response")
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	}
}

// handleAIOptimize handles AI optimization requests
func handleAIOptimize(systemManager *system.Manager, logger *logrus.Logger, tracer trace.Tracer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, span := tracer.Start(r.Context(), "api.handleAIOptimize")
		defer span.End()

		var request struct {
			Task       string                 `json:"task"`
			Parameters map[string]interface{} `json:"parameters,omitempty"`
		}

		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			logger.WithError(err).Error("Failed to decode AI optimization request")
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		if request.Task == "" {
			request.Task = "analyze"
		}

		// Create AI request
		aiRequest := &models.AIRequest{
			ID:         uuid.New().String(),
			Type:       "optimization",
			Parameters: request.Parameters,
			Timeout:    60 * time.Second,
			Timestamp:  time.Now(),
		}

		if aiRequest.Parameters == nil {
			aiRequest.Parameters = make(map[string]interface{})
		}
		aiRequest.Parameters["task"] = request.Task

		// Process request through AI orchestrator
		orchestrator := systemManager.GetAIOrchestrator()
		response, err := orchestrator.ProcessRequest(ctx, aiRequest)
		if err != nil {
			logger.WithError(err).Error("Failed to process AI optimization request")
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			logger.WithError(err).Error("Failed to encode AI optimization response")
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	}
}

// handleAIStatus returns the status of all AI services
func handleAIStatus(systemManager *system.Manager, logger *logrus.Logger, tracer trace.Tracer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, span := tracer.Start(r.Context(), "api.handleAIStatus")
		defer span.End()

		orchestrator := systemManager.GetAIOrchestrator()
		status, err := orchestrator.GetServiceStatus(ctx)
		if err != nil {
			logger.WithError(err).Error("Failed to get AI service status")
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(status); err != nil {
			logger.WithError(err).Error("Failed to encode AI status response")
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	}
}

// handleAIModels returns available AI models
func handleAIModels(systemManager *system.Manager, logger *logrus.Logger, tracer trace.Tracer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, span := tracer.Start(r.Context(), "api.handleAIModels")
		defer span.End()

		// TODO: Get models from model manager
		// For now, return mock models
		models := []models.AIModel{
			{
				ID:           "llama2",
				Name:         "Llama 2",
				Version:      "7b",
				Type:         "llm",
				Size:         3800000000,
				Description:  "Llama 2 7B parameter model for text generation",
				Capabilities: []string{"text-generation", "chat", "code"},
				Status:       "loaded",
				CreatedAt:    time.Now().Add(-24 * time.Hour),
				UpdatedAt:    time.Now(),
			},
			{
				ID:           "cv-model",
				Name:         "Computer Vision Model",
				Version:      "1.0",
				Type:         "cv",
				Size:         500000000,
				Description:  "Computer vision model for UI analysis",
				Capabilities: []string{"object-detection", "ocr", "ui-analysis"},
				Status:       "available",
				CreatedAt:    time.Now().Add(-12 * time.Hour),
				UpdatedAt:    time.Now(),
			},
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(models); err != nil {
			logger.WithError(err).Error("Failed to encode AI models response")
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	}
}

// handleAIWorkflow handles AI workflow execution requests
func handleAIWorkflow(systemManager *system.Manager, logger *logrus.Logger, tracer trace.Tracer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, span := tracer.Start(r.Context(), "api.handleAIWorkflow")
		defer span.End()

		var workflow models.AIWorkflow
		if err := json.NewDecoder(r.Body).Decode(&workflow); err != nil {
			logger.WithError(err).Error("Failed to decode AI workflow request")
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		// Set workflow ID if not provided
		if workflow.ID == "" {
			workflow.ID = uuid.New().String()
		}

		// Set default timeout if not provided
		if workflow.Timeout == 0 {
			workflow.Timeout = 5 * time.Minute
		}

		workflow.Timestamp = time.Now()

		// Execute workflow through AI orchestrator
		orchestrator := systemManager.GetAIOrchestrator()
		result, err := orchestrator.ManageWorkflow(ctx, &workflow)
		if err != nil {
			logger.WithError(err).Error("Failed to execute AI workflow")
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(result); err != nil {
			logger.WithError(err).Error("Failed to encode AI workflow response")
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	}
}
