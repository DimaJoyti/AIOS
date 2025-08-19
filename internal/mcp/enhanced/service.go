package enhanced

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/aios/aios/internal/ai/knowledge"
	knowledgeService "github.com/aios/aios/internal/knowledge"
	"github.com/aios/aios/pkg/config"
	"github.com/aios/aios/pkg/mcp/server"
	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

// Service provides enhanced MCP functionality with knowledge integration
type Service struct {
	config           *config.Config
	logger           *logrus.Logger
	tracer           trace.Tracer
	db               *sqlx.DB
	mcpServer        *server.MCPServer
	knowledgeService *knowledgeService.Service
	knowledgeAgent   *knowledge.DocumentAgent
	ragAgent         *knowledge.RAGAgent
	sessionManager   *SessionManager
	toolRegistry     *EnhancedToolRegistry
	streamingHandler *StreamingHandler
	httpServer       *http.Server
}

// NewService creates a new enhanced MCP service instance
func NewService(config *config.Config, db *sqlx.DB, knowledgeService *knowledgeService.Service, logger *logrus.Logger) (*Service, error) {
	tracer := otel.Tracer("mcp.enhanced")

	// Create knowledge agents
	knowledgeAgent, err := knowledge.NewDocumentAgent(knowledgeService, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create knowledge agent: %w", err)
	}

	ragAgent, err := knowledge.NewRAGAgent(knowledgeService, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create RAG agent: %w", err)
	}

	// Create session manager
	sessionManager := NewSessionManager(logger)

	// Create enhanced tool registry
	toolRegistry := NewEnhancedToolRegistry(sessionManager, knowledgeAgent, ragAgent, knowledgeService, logger)

	// Create streaming handler
	streamingHandler := NewStreamingHandler(sessionManager, toolRegistry, logger)

	// Create MCP server
	mcpServer, err := server.NewMCPServer(&server.ServerConfig{
		Address: "0.0.0.0",
		Port:    config.Services.MCP.Port,
		Metadata: map[string]interface{}{
			"name":        "aios-enhanced-mcp",
			"version":     "1.0.0",
			"description": "Enhanced MCP Server with Knowledge Integration",
		},
	}, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create MCP server: %w", err)
	}

	service := &Service{
		config:           config,
		logger:           logger,
		tracer:           tracer,
		db:               db,
		mcpServer:        mcpServer,
		knowledgeService: knowledgeService,
		knowledgeAgent:   knowledgeAgent,
		ragAgent:         ragAgent,
		sessionManager:   sessionManager,
		toolRegistry:     toolRegistry,
		streamingHandler: streamingHandler,
	}

	return service, nil
}

// Start starts the enhanced MCP service
func (s *Service) Start(ctx context.Context) error {
	s.logger.Info("Starting Enhanced MCP Service...")

	// Start MCP server
	if err := s.mcpServer.Start(ctx); err != nil {
		return fmt.Errorf("failed to start MCP server: %w", err)
	}

	// Setup HTTP server for management and WebSocket
	router := mux.NewRouter()
	s.setupRoutes(router)

	s.httpServer = &http.Server{
		Addr:         fmt.Sprintf(":%d", s.config.Services.MCP.Port+1), // Management port
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start HTTP server in goroutine
	go func() {
		s.logger.WithField("port", s.config.Services.MCP.Port+1).Info("Enhanced MCP management server starting")
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.WithError(err).Error("HTTP server error")
		}
	}()

	s.logger.Info("Enhanced MCP Service started successfully")
	return nil
}

// Stop stops the enhanced MCP service
func (s *Service) Stop(ctx context.Context) error {
	s.logger.Info("Stopping Enhanced MCP Service...")

	// Stop HTTP server
	if s.httpServer != nil {
		if err := s.httpServer.Shutdown(ctx); err != nil {
			s.logger.WithError(err).Error("Error shutting down HTTP server")
		}
	}

	// Stop MCP server
	if err := s.mcpServer.Stop(context.Background()); err != nil {
		s.logger.WithError(err).Error("Error stopping MCP server")
	}

	s.logger.Info("Enhanced MCP Service stopped")
	return nil
}

// setupRoutes sets up HTTP routes
func (s *Service) setupRoutes(router *mux.Router) {
	// Health check
	router.HandleFunc("/health", s.healthHandler).Methods("GET")

	// Tool management
	router.HandleFunc("/tools", s.toolsHandler).Methods("GET")
	router.HandleFunc("/tools/execute", s.executeToolHandler).Methods("POST")

	// Session management
	router.HandleFunc("/sessions", s.sessionsHandler).Methods("GET", "POST")
	router.HandleFunc("/sessions/{id}", s.sessionHandler).Methods("GET", "DELETE")

	// WebSocket for real-time communication
	router.HandleFunc("/ws", s.websocketHandler)

	// Enhanced session management
	router.HandleFunc("/sessions/{id}/context", s.sessionContextHandler).Methods("GET", "PUT")
	router.HandleFunc("/sessions/{id}/memory", s.sessionMemoryHandler).Methods("GET", "POST")

	// Status and metrics
	router.HandleFunc("/status", s.statusHandler).Methods("GET")
	router.HandleFunc("/metrics", s.metricsHandler).Methods("GET")
}

// healthHandler handles health check requests
func (s *Service) healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
}

// toolsHandler handles tool listing requests
func (s *Service) toolsHandler(w http.ResponseWriter, r *http.Request) {
	tools := s.toolRegistry.GetToolList()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"tools": tools,
		"count": len(tools),
	})
}

// executeToolHandler handles tool execution requests
func (s *Service) executeToolHandler(w http.ResponseWriter, r *http.Request) {
	ctx, span := s.tracer.Start(r.Context(), "mcp.execute_tool")
	defer span.End()

	var req MCPToolRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Execute tool (need session ID - create one if not provided)
	sessionID := "default-session" // TODO: Get from request or create new session
	result, err := s.toolRegistry.ExecuteTool(ctx, sessionID, req.ToolName, req.Arguments)
	if err != nil {
		response := MCPToolResponse{
			Success: false,
			Error:   err.Error(),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	response := MCPToolResponse{
		Success: true,
		Result:  result,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// sessionsHandler handles session management requests
func (s *Service) sessionsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		// List sessions
		sessions := []map[string]interface{}{
			{
				"id":     "default",
				"status": "active",
				"tools":  len(s.toolRegistry.tools),
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"sessions": sessions,
		})

	case "POST":
		// Create session
		sessionID := fmt.Sprintf("session_%d", time.Now().Unix())
		session := map[string]interface{}{
			"id":         sessionID,
			"status":     "active",
			"created_at": time.Now(),
			"tools":      len(s.toolRegistry.tools),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(session)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// sessionHandler handles individual session requests
func (s *Service) sessionHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sessionID := vars["id"]

	switch r.Method {
	case "GET":
		session := map[string]interface{}{
			"id":     sessionID,
			"status": "active",
			"tools":  len(s.toolRegistry.tools),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(session)

	case "DELETE":
		w.WriteHeader(http.StatusNoContent)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// websocketHandler handles WebSocket connections
func (s *Service) websocketHandler(w http.ResponseWriter, r *http.Request) {
	// Delegate to streaming handler
	s.streamingHandler.HandleWebSocket(w, r)
}

// statusHandler handles status requests
func (s *Service) statusHandler(w http.ResponseWriter, r *http.Request) {
	status := map[string]interface{}{
		"status":     "running",
		"tools":      len(s.toolRegistry.tools),
		"mcp_server": "active",
		"uptime":     time.Since(time.Now()).String(), // This would be actual uptime
		"version":    "1.0.0",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

// metricsHandler handles metrics requests
func (s *Service) metricsHandler(w http.ResponseWriter, r *http.Request) {
	metrics := map[string]interface{}{
		"tool_executions": 0, // This would be actual metrics
		"active_sessions": 1,
		"total_requests":  0,
		"error_rate":      0.0,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}

// sessionContextHandler handles session context requests
func (s *Service) sessionContextHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sessionID := vars["id"]

	switch r.Method {
	case "GET":
		session, err := s.sessionManager.GetSession(sessionID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(session.Context)

	case "PUT":
		var updates map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		if err := s.sessionManager.UpdateSessionContext(sessionID, updates); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "updated"})
	}
}

// sessionMemoryHandler handles session memory requests
func (s *Service) sessionMemoryHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sessionID := vars["id"]

	switch r.Method {
	case "GET":
		query := r.URL.Query().Get("query")
		limit := 10
		if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
			if l, err := fmt.Sscanf(limitStr, "%d", &limit); err == nil && l == 1 {
				// limit parsed successfully
			}
		}

		var memory []MemoryItem
		var err error
		if query != "" {
			memory, err = s.sessionManager.GetRelevantMemory(sessionID, query, limit)
		} else {
			session, err := s.sessionManager.GetSession(sessionID)
			if err != nil {
				http.Error(w, err.Error(), http.StatusNotFound)
				return
			}
			memory = session.Memory.ShortTerm
			if len(memory) > limit {
				memory = memory[:limit]
			}
		}

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(memory)

	case "POST":
		var item MemoryItem
		if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		if err := s.sessionManager.AddToMemory(sessionID, item); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "added"})
	}
}

// startCleanupRoutines starts background cleanup routines
func (s *Service) startCleanupRoutines(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// Cleanup expired sessions
			s.sessionManager.CleanupExpiredSessions(24 * time.Hour)

			// Cleanup inactive connections
			s.streamingHandler.CleanupInactiveConnections(30 * time.Minute)
		}
	}
}
