package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/aios/aios/internal/system"
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
