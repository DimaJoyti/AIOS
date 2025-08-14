package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
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

type DesktopServer struct {
	httpServer *http.Server
	logger     *logrus.Logger
	tracer     trace.Tracer
}

func main() {
	var rootCmd = &cobra.Command{
		Use:   "aios-desktop",
		Short: "AIOS Desktop Environment",
		Long:  "The AI-aware desktop environment for the AIOS system",
		Run:   runDesktop,
	}

	rootCmd.Flags().String("config", "", "config file (default is $HOME/.aios.yaml)")
	rootCmd.Flags().String("log-level", "info", "log level (debug, info, warn, error)")
	rootCmd.Flags().String("bind-addr", ":8082", "HTTP server bind address")

	viper.BindPFlags(rootCmd.Flags())

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func runDesktop(cmd *cobra.Command, args []string) {
	// Initialize configuration
	initConfig()

	// Initialize logger
	logger := initLogger()

	// Initialize tracing
	tracer := otel.Tracer("aios-desktop")

	// Create server
	server := &DesktopServer{
		logger: logger,
		tracer: tracer,
	}

	// Start server
	if err := server.Start(); err != nil {
		logger.WithError(err).Fatal("Failed to start desktop server")
	}

	// Wait for shutdown signal
	server.WaitForShutdown()
}

func (s *DesktopServer) Start() error {
	// Initialize HTTP router
	router := mux.NewRouter()

	// Add middleware
	router.Use(s.loggingMiddleware)
	router.Use(s.corsMiddleware)
	router.Use(otelhttp.NewMiddleware("aios-desktop"))

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
		s.logger.WithField("addr", s.httpServer.Addr).Info("Starting desktop environment server")
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.WithError(err).Fatal("Desktop server failed")
		}
	}()

	s.logger.Info("Desktop environment started successfully")
	return nil
}

func (s *DesktopServer) registerRoutes(router *mux.Router) {
	api := router.PathPrefix("/api/v1").Subrouter()

	// Health endpoints
	api.HandleFunc("/health", s.handleHealth).Methods("GET")
	api.HandleFunc("/ready", s.handleReady).Methods("GET")

	// Desktop environment endpoints
	api.HandleFunc("/windows", s.handleWindows).Methods("GET")
	api.HandleFunc("/workspaces", s.handleWorkspaces).Methods("GET")
	api.HandleFunc("/applications", s.handleApplications).Methods("GET")
	api.HandleFunc("/settings", s.handleSettings).Methods("GET", "POST")

	// Window management endpoints
	api.HandleFunc("/windows/{id}/focus", s.handleWindowFocus).Methods("POST")
	api.HandleFunc("/windows/{id}/close", s.handleWindowClose).Methods("POST")
	api.HandleFunc("/windows/{id}/minimize", s.handleWindowMinimize).Methods("POST")
	api.HandleFunc("/windows/{id}/maximize", s.handleWindowMaximize).Methods("POST")

	// Workspace management endpoints
	api.HandleFunc("/workspaces/{id}/switch", s.handleWorkspaceSwitch).Methods("POST")
	api.HandleFunc("/workspaces/{id}/windows", s.handleWorkspaceWindows).Methods("GET")

	// Application launcher endpoints
	api.HandleFunc("/applications/{id}/launch", s.handleApplicationLaunch).Methods("POST")
	api.HandleFunc("/applications/search", s.handleApplicationSearch).Methods("GET")

	// Static file serving for desktop UI
	router.PathPrefix("/").Handler(http.FileServer(http.Dir("./web/desktop/dist/")))
}

func (s *DesktopServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	ctx, span := s.tracer.Start(r.Context(), "desktop.handleHealth")
	defer span.End()

	response := map[string]interface{}{
		"status":    "healthy",
		"service":   "aios-desktop",
		"version":   Version,
		"timestamp": time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	if err := writeJSON(w, response); err != nil {
		s.logger.WithError(err).Error("Failed to write health response")
	}
}

func (s *DesktopServer) handleReady(w http.ResponseWriter, r *http.Request) {
	ctx, span := s.tracer.Start(r.Context(), "desktop.handleReady")
	defer span.End()

	response := map[string]interface{}{
		"status":    "ready",
		"service":   "aios-desktop",
		"timestamp": time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	if err := writeJSON(w, response); err != nil {
		s.logger.WithError(err).Error("Failed to write ready response")
	}
}

func (s *DesktopServer) handleWindows(w http.ResponseWriter, r *http.Request) {
	ctx, span := s.tracer.Start(r.Context(), "desktop.handleWindows")
	defer span.End()

	// TODO: Implement actual window management
	windows := []map[string]interface{}{
		{
			"id":          "window-1",
			"title":       "Terminal",
			"application": "gnome-terminal",
			"workspace":   1,
			"focused":     true,
			"minimized":   false,
			"maximized":   false,
		},
		{
			"id":          "window-2",
			"title":       "Firefox",
			"application": "firefox",
			"workspace":   1,
			"focused":     false,
			"minimized":   false,
			"maximized":   true,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	if err := writeJSON(w, windows); err != nil {
		s.logger.WithError(err).Error("Failed to write windows response")
	}
}

func (s *DesktopServer) handleWorkspaces(w http.ResponseWriter, r *http.Request) {
	ctx, span := s.tracer.Start(r.Context(), "desktop.handleWorkspaces")
	defer span.End()

	// TODO: Implement actual workspace management
	workspaces := []map[string]interface{}{
		{
			"id":      1,
			"name":    "Main",
			"active":  true,
			"windows": 2,
		},
		{
			"id":      2,
			"name":    "Development",
			"active":  false,
			"windows": 0,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	if err := writeJSON(w, workspaces); err != nil {
		s.logger.WithError(err).Error("Failed to write workspaces response")
	}
}

func (s *DesktopServer) handleApplications(w http.ResponseWriter, r *http.Request) {
	ctx, span := s.tracer.Start(r.Context(), "desktop.handleApplications")
	defer span.End()

	// TODO: Implement actual application discovery
	applications := []map[string]interface{}{
		{
			"id":          "firefox",
			"name":        "Firefox",
			"description": "Web Browser",
			"icon":        "/icons/firefox.png",
			"category":    "Internet",
		},
		{
			"id":          "terminal",
			"name":        "Terminal",
			"description": "Command Line Interface",
			"icon":        "/icons/terminal.png",
			"category":    "System",
		},
	}

	w.Header().Set("Content-Type", "application/json")
	if err := writeJSON(w, applications); err != nil {
		s.logger.WithError(err).Error("Failed to write applications response")
	}
}

func (s *DesktopServer) handleSettings(w http.ResponseWriter, r *http.Request) {
	ctx, span := s.tracer.Start(r.Context(), "desktop.handleSettings")
	defer span.End()

	if r.Method == "GET" {
		// TODO: Implement settings retrieval
		settings := map[string]interface{}{
			"theme":           "dark",
			"wallpaper":       "/wallpapers/default.jpg",
			"ai_assistant":    true,
			"voice_commands":  false,
			"auto_organize":   true,
		}

		w.Header().Set("Content-Type", "application/json")
		if err := writeJSON(w, settings); err != nil {
			s.logger.WithError(err).Error("Failed to write settings response")
		}
	} else {
		// TODO: Implement settings update
		response := map[string]interface{}{
			"message":   "Settings updated successfully",
			"timestamp": time.Now(),
		}

		w.Header().Set("Content-Type", "application/json")
		if err := writeJSON(w, response); err != nil {
			s.logger.WithError(err).Error("Failed to write settings update response")
		}
	}
}

func (s *DesktopServer) handleWindowFocus(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement window focus
	s.handleGenericAction(w, r, "focus")
}

func (s *DesktopServer) handleWindowClose(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement window close
	s.handleGenericAction(w, r, "close")
}

func (s *DesktopServer) handleWindowMinimize(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement window minimize
	s.handleGenericAction(w, r, "minimize")
}

func (s *DesktopServer) handleWindowMaximize(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement window maximize
	s.handleGenericAction(w, r, "maximize")
}

func (s *DesktopServer) handleWorkspaceSwitch(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement workspace switching
	s.handleGenericAction(w, r, "switch")
}

func (s *DesktopServer) handleWorkspaceWindows(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement workspace window listing
	windows := []map[string]interface{}{}

	w.Header().Set("Content-Type", "application/json")
	if err := writeJSON(w, windows); err != nil {
		s.logger.WithError(err).Error("Failed to write workspace windows response")
	}
}

func (s *DesktopServer) handleApplicationLaunch(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement application launching
	s.handleGenericAction(w, r, "launch")
}

func (s *DesktopServer) handleApplicationSearch(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement application search
	query := r.URL.Query().Get("q")
	results := []map[string]interface{}{
		{
			"id":    "firefox",
			"name":  "Firefox",
			"score": 0.9,
			"query": query,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	if err := writeJSON(w, results); err != nil {
		s.logger.WithError(err).Error("Failed to write application search response")
	}
}

func (s *DesktopServer) handleGenericAction(w http.ResponseWriter, r *http.Request, action string) {
	vars := mux.Vars(r)
	id := vars["id"]

	response := map[string]interface{}{
		"action":    action,
		"target":    id,
		"success":   true,
		"timestamp": time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	if err := writeJSON(w, response); err != nil {
		s.logger.WithError(err).Error("Failed to write action response")
	}
}

func (s *DesktopServer) WaitForShutdown() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	s.logger.Info("Shutting down desktop environment...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := s.httpServer.Shutdown(ctx); err != nil {
		s.logger.WithError(err).Error("Failed to shutdown desktop server")
	}

	s.logger.Info("Desktop environment shutdown complete")
}

func (s *DesktopServer) loggingMiddleware(next http.Handler) http.Handler {
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

func (s *DesktopServer) corsMiddleware(next http.Handler) http.Handler {
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

// Helper function
func writeJSON(w http.ResponseWriter, v interface{}) error {
	// TODO: Implement JSON writing with proper error handling
	return nil
}
