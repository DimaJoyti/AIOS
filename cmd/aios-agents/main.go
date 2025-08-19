package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aios/aios/internal/adapters"
	aiknowledge "github.com/aios/aios/internal/ai/knowledge"
	"github.com/aios/aios/internal/knowledge"
	"github.com/aios/aios/pkg/ai"
	"github.com/aios/aios/pkg/config"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

// AgentsService represents the AIOS Agents Service with Archon capabilities
type AgentsService struct {
	config           *config.Config
	logger           *logrus.Logger
	tracer           trace.Tracer
	orchestrator     ai.Orchestrator
	pythonAdapter    *adapters.PythonServiceAdapter
	knowledgeService *knowledge.Service
	httpServer       *http.Server
}

func main() {
	// Initialize logger
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)
	logger.SetFormatter(&logrus.JSONFormatter{})

	// Load configuration
	cfgManager := config.NewManager("dev", "configs")
	cfg, err := cfgManager.Load()
	if err != nil {
		logger.WithError(err).Fatal("Failed to load configuration")
	}

	// Initialize OpenTelemetry tracer
	tracer := otel.Tracer("aios.agents")

	// Create agents service
	service, err := NewAgentsService(cfg, logger, tracer)
	if err != nil {
		logger.WithError(err).Fatal("Failed to create agents service")
	}

	// Start service
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := service.Start(ctx); err != nil {
		logger.WithError(err).Fatal("Failed to start agents service")
	}

	// Wait for shutdown signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	logger.Info("Shutting down agents service...")

	// Graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := service.Stop(shutdownCtx); err != nil {
		logger.WithError(err).Error("Error during shutdown")
	}

	logger.Info("Agents service stopped")
}

// NewAgentsService creates a new agents service instance
func NewAgentsService(cfg *config.Config, logger *logrus.Logger, tracer trace.Tracer) (*AgentsService, error) {
	// Create Python service adapter for the agents server
	pythonAdapter, err := adapters.NewPythonServiceAdapter(&adapters.PythonServiceConfig{
		ServiceName: "agents-server",
		BaseURL:     fmt.Sprintf("http://localhost:%d", cfg.Services.Agents.Port),
		Timeout:     30 * time.Second,
	}, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create Python adapter: %w", err)
	}

	// Create knowledge-aware AI orchestrator
	orchestrator := ai.NewDefaultOrchestrator(logger)

	// Create knowledge service for agent integration
	// Note: This would typically require database connection
	// For now, we'll create a minimal service or use nil
	var knowledgeService *knowledge.Service
	// knowledgeService, err := knowledge.NewService(cfg, db, logger)
	// if err != nil {
	//     return nil, fmt.Errorf("failed to create knowledge service: %w", err)
	// }

	return &AgentsService{
		config:           cfg,
		logger:           logger,
		tracer:           tracer,
		orchestrator:     orchestrator,
		pythonAdapter:    pythonAdapter,
		knowledgeService: knowledgeService,
	}, nil
}

// Start starts the agents service
func (s *AgentsService) Start(ctx context.Context) error {
	s.logger.Info("Starting AIOS Agents Service...")

	// Start the Python adapter
	if err := s.pythonAdapter.Start(ctx); err != nil {
		return fmt.Errorf("failed to start Python adapter: %w", err)
	}

	// Start the AI orchestrator
	if err := s.orchestrator.Start(ctx); err != nil {
		return fmt.Errorf("failed to start orchestrator: %w", err)
	}

	// Register knowledge-aware agents
	if err := s.registerKnowledgeAgents(); err != nil {
		return fmt.Errorf("failed to register knowledge agents: %w", err)
	}

	// Create HTTP server
	mux := http.NewServeMux()
	s.registerRoutes(mux)

	s.httpServer = &http.Server{
		Addr:    fmt.Sprintf(":%d", s.config.Services.Agents.Port),
		Handler: mux,
	}

	// Start HTTP server in goroutine
	go func() {
		s.logger.WithField("port", s.config.Services.Agents.Port).Info("Agents service HTTP server starting")
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.WithError(err).Error("HTTP server error")
		}
	}()

	s.logger.Info("AIOS Agents Service started successfully")
	return nil
}

// Stop stops the agents service
func (s *AgentsService) Stop(ctx context.Context) error {
	s.logger.Info("Stopping AIOS Agents Service...")

	// Stop HTTP server
	if s.httpServer != nil {
		if err := s.httpServer.Shutdown(ctx); err != nil {
			s.logger.WithError(err).Error("Error shutting down HTTP server")
		}
	}

	// Stop orchestrator
	if err := s.orchestrator.Stop(); err != nil {
		s.logger.WithError(err).Error("Error stopping orchestrator")
	}

	// Stop Python adapter
	if err := s.pythonAdapter.Stop(ctx); err != nil {
		s.logger.WithError(err).Error("Error stopping Python adapter")
	}

	s.logger.Info("AIOS Agents Service stopped")
	return nil
}

// registerKnowledgeAgents registers knowledge-aware agents with the orchestrator
func (s *AgentsService) registerKnowledgeAgents() error {
	// Register document processing agent
	if s.knowledgeService != nil {
		docAgent, err := aiknowledge.NewDocumentAgent(s.knowledgeService, s.logger)
		if err != nil {
			return fmt.Errorf("failed to create document agent: %w", err)
		}
		if err := s.orchestrator.RegisterAgent(docAgent); err != nil {
			return fmt.Errorf("failed to register document agent: %w", err)
		}

		// Register RAG query agent
		ragAgent, err := aiknowledge.NewRAGAgent(s.knowledgeService, s.logger)
		if err != nil {
			return fmt.Errorf("failed to create RAG agent: %w", err)
		}
		if err := s.orchestrator.RegisterAgent(ragAgent); err != nil {
			return fmt.Errorf("failed to register RAG agent: %w", err)
		}
	}

	s.logger.Info("Knowledge-aware agents registered successfully")
	return nil
}

// registerRoutes registers HTTP routes for the agents service
func (s *AgentsService) registerRoutes(mux *http.ServeMux) {
	// Health check
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Agent status
	mux.HandleFunc("/agents", func(w http.ResponseWriter, r *http.Request) {
		agents := s.orchestrator.ListAgents()
		// Return agent information (implement JSON response)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fmt.Sprintf(`{"agents": %d}`, len(agents))))
	})

	// Orchestrator status
	mux.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		status := s.orchestrator.GetStatus()
		// Return status information (implement JSON response)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fmt.Sprintf(`{"running": %t}`, status.Running)))
	})
}
