package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

var (
	Version   = "dev"
	Commit    = "unknown"
	BuildTime = "unknown"
)

type AssistantServer struct {
	httpServer *http.Server
	logger     *logrus.Logger
	tracer     trace.Tracer
	upgrader   websocket.Upgrader
}

func main() {
	var rootCmd = &cobra.Command{
		Use:   "aios-assistant",
		Short: "AIOS AI Assistant Service",
		Long:  "The AI assistant service for natural language interaction with the AIOS system",
		Run:   runAssistant,
	}

	rootCmd.Flags().String("config", "", "config file (default is $HOME/.aios.yaml)")
	rootCmd.Flags().String("log-level", "info", "log level (debug, info, warn, error)")
	rootCmd.Flags().String("bind-addr", ":8081", "HTTP server bind address")

	viper.BindPFlags(rootCmd.Flags())

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func runAssistant(cmd *cobra.Command, args []string) {
	// Initialize configuration
	initConfig()

	// Initialize logger
	logger := initLogger()

	// Initialize tracing
	tracer := otel.Tracer("aios-assistant")

	// Create server
	server := &AssistantServer{
		logger: logger,
		tracer: tracer,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow all origins in development
			},
		},
	}

	// Start server
	if err := server.Start(); err != nil {
		logger.WithError(err).Fatal("Failed to start assistant server")
	}

	// Wait for shutdown signal
	server.WaitForShutdown()
}

func (s *AssistantServer) Start() error {
	// Initialize HTTP router
	router := mux.NewRouter()

	// Add middleware
	router.Use(s.loggingMiddleware)
	router.Use(s.corsMiddleware)
	router.Use(otelhttp.NewMiddleware("aios-assistant"))

	// Register API routes
	s.registerRoutes(router)

	// Create HTTP server
	s.httpServer = &http.Server{
		Addr:         viper.GetString("bind-addr"),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start HTTP server
	go func() {
		s.logger.WithField("addr", s.httpServer.Addr).Info("Starting AI assistant server")
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.WithError(err).Fatal("Assistant server failed")
		}
	}()

	s.logger.Info("AI assistant service started successfully")
	return nil
}

func (s *AssistantServer) registerRoutes(router *mux.Router) {
	api := router.PathPrefix("/api/v1").Subrouter()

	// Health endpoints
	api.HandleFunc("/health", s.handleHealth).Methods("GET")
	api.HandleFunc("/ready", s.handleReady).Methods("GET")

	// Assistant endpoints
	api.HandleFunc("/chat", s.handleChat).Methods("POST")
	api.HandleFunc("/voice", s.handleVoice).Methods("POST")
	api.HandleFunc("/commands", s.handleCommands).Methods("POST")

	// WebSocket endpoint for real-time communication
	api.HandleFunc("/ws", s.handleWebSocket).Methods("GET")

	// Static file serving for assistant UI
	router.PathPrefix("/").Handler(http.FileServer(http.Dir("./web/assistant/dist/")))
}

func (s *AssistantServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	_, span := s.tracer.Start(r.Context(), "assistant.handleHealth")
	defer span.End()

	response := map[string]interface{}{
		"status":    "healthy",
		"service":   "aios-assistant",
		"version":   Version,
		"timestamp": time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	if err := writeJSON(w, response); err != nil {
		s.logger.WithError(err).Error("Failed to write health response")
	}
}

func (s *AssistantServer) handleReady(w http.ResponseWriter, r *http.Request) {
	_, span := s.tracer.Start(r.Context(), "assistant.handleReady")
	defer span.End()

	response := map[string]interface{}{
		"status":    "ready",
		"service":   "aios-assistant",
		"timestamp": time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	if err := writeJSON(w, response); err != nil {
		s.logger.WithError(err).Error("Failed to write ready response")
	}
}

func (s *AssistantServer) handleChat(w http.ResponseWriter, r *http.Request) {
	_, span := s.tracer.Start(r.Context(), "assistant.handleChat")
	defer span.End()

	var request struct {
		Message string `json:"message"`
		Context string `json:"context,omitempty"`
	}

	if err := readJSON(r, &request); err != nil {
		s.logger.WithError(err).Error("Failed to decode chat request")
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	// TODO: Implement actual AI chat processing
	response := map[string]interface{}{
		"response":  fmt.Sprintf("I received your message: %s", request.Message),
		"timestamp": time.Now(),
		"context":   request.Context,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := writeJSON(w, response); err != nil {
		s.logger.WithError(err).Error("Failed to write chat response")
	}
}

func (s *AssistantServer) handleVoice(w http.ResponseWriter, r *http.Request) {
	_, span := s.tracer.Start(r.Context(), "assistant.handleVoice")
	defer span.End()

	// TODO: Implement voice processing
	response := map[string]interface{}{
		"message":   "Voice processing not yet implemented",
		"timestamp": time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	if err := writeJSON(w, response); err != nil {
		s.logger.WithError(err).Error("Failed to write voice response")
	}
}

func (s *AssistantServer) handleCommands(w http.ResponseWriter, r *http.Request) {
	_, span := s.tracer.Start(r.Context(), "assistant.handleCommands")
	defer span.End()

	var request struct {
		Command string   `json:"command"`
		Args    []string `json:"args,omitempty"`
	}

	if err := readJSON(r, &request); err != nil {
		s.logger.WithError(err).Error("Failed to decode command request")
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	// TODO: Implement command processing
	response := map[string]interface{}{
		"result":    fmt.Sprintf("Command '%s' processed", request.Command),
		"timestamp": time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	if err := writeJSON(w, response); err != nil {
		s.logger.WithError(err).Error("Failed to write command response")
	}
}

func (s *AssistantServer) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	_, span := s.tracer.Start(r.Context(), "assistant.handleWebSocket")
	defer span.End()

	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		s.logger.WithError(err).Error("Failed to upgrade WebSocket connection")
		return
	}
	defer conn.Close()

	s.logger.Info("WebSocket connection established")

	// Handle WebSocket messages
	for {
		var message map[string]interface{}
		if err := conn.ReadJSON(&message); err != nil {
			s.logger.WithError(err).Debug("WebSocket connection closed")
			break
		}

		// Echo message back (TODO: implement actual processing)
		response := map[string]interface{}{
			"type":      "response",
			"message":   fmt.Sprintf("Received: %v", message),
			"timestamp": time.Now(),
		}

		if err := conn.WriteJSON(response); err != nil {
			s.logger.WithError(err).Error("Failed to write WebSocket response")
			break
		}
	}
}

func (s *AssistantServer) WaitForShutdown() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	s.logger.Info("Shutting down AI assistant service...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := s.httpServer.Shutdown(ctx); err != nil {
		s.logger.WithError(err).Error("Failed to shutdown assistant server")
	}

	s.logger.Info("AI assistant service shutdown complete")
}

func (s *AssistantServer) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		duration := time.Since(start)

		s.logger.WithFields(logrus.Fields{
			"method":   r.Method,
			"path":     r.URL.Path,
			"duration": duration,
		}).Info("Request processed")
	})
}

func (s *AssistantServer) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func initConfig() {
	if cfgFile := viper.GetString("config"); cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			log.Fatal(err)
		}
		viper.AddConfigPath(home)
		viper.AddConfigPath(".")
		viper.AddConfigPath("./configs")
		viper.SetConfigName(".aios")
		viper.SetConfigType("yaml")
	}

	viper.AutomaticEnv()
	viper.SetEnvPrefix("AIOS")

	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}

func initLogger() *logrus.Logger {
	logger := logrus.New()

	level, err := logrus.ParseLevel(viper.GetString("log-level"))
	if err != nil {
		level = logrus.InfoLevel
	}
	logger.SetLevel(level)

	logger.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: time.RFC3339,
	})

	return logger
}

// Helper functions
func readJSON(r *http.Request, v interface{}) error {
	return json.NewDecoder(r.Body).Decode(v)
}

func writeJSON(w http.ResponseWriter, v interface{}) error {
	return json.NewEncoder(w).Encode(v)
}
