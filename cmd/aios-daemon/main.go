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

	"github.com/aios/aios/internal/system"
	"github.com/aios/aios/pkg/api"
	"github.com/aios/aios/pkg/utils"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
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

type Server struct {
	httpServer    *http.Server
	metricsServer *http.Server
	logger        *logrus.Logger
	tracer        trace.Tracer
	systemManager *system.Manager
}

func main() {
	var rootCmd = &cobra.Command{
		Use:   "aios-daemon",
		Short: "AIOS System Daemon",
		Long:  "The main system daemon for the AI-powered Operating System",
		Run:   runDaemon,
	}

	rootCmd.Flags().String("config", "", "config file (default is $HOME/.aios.yaml)")
	rootCmd.Flags().String("log-level", "info", "log level (debug, info, warn, error)")
	rootCmd.Flags().String("bind-addr", ":8080", "HTTP server bind address")
	rootCmd.Flags().String("metrics-addr", ":9090", "Metrics server bind address")

	viper.BindPFlags(rootCmd.Flags())

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func runDaemon(cmd *cobra.Command, args []string) {
	// Initialize configuration
	initConfig()

	// Initialize logger
	logger := initLogger()

	// Initialize tracing
	tracer := otel.Tracer("aios-daemon")

	// Initialize system manager
	systemManager, err := system.NewManager(logger)
	if err != nil {
		logger.WithError(err).Fatal("Failed to initialize system manager")
	}

	// Create server
	server := &Server{
		logger:        logger,
		tracer:        tracer,
		systemManager: systemManager,
	}

	// Start server
	if err := server.Start(); err != nil {
		logger.WithError(err).Fatal("Failed to start server")
	}

	// Wait for shutdown signal
	server.WaitForShutdown()
}

func (s *Server) Start() error {
	ctx := context.Background()

	// Initialize HTTP router
	router := mux.NewRouter()

	// Add middleware
	router.Use(utils.LoggingMiddleware(s.logger))
	router.Use(utils.CORSMiddleware())
	router.Use(otelhttp.NewMiddleware("aios-daemon"))

	// Register API routes
	api.RegisterSystemRoutes(router, s.systemManager, s.logger)
	api.RegisterHealthRoutes(router, s.logger)

	// Create HTTP server
	s.httpServer = &http.Server{
		Addr:         viper.GetString("bind-addr"),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Create metrics server
	metricsRouter := mux.NewRouter()
	metricsRouter.Handle("/metrics", promhttp.Handler())

	s.metricsServer = &http.Server{
		Addr:    viper.GetString("metrics-addr"),
		Handler: metricsRouter,
	}

	// Start system manager
	if err := s.systemManager.Start(ctx); err != nil {
		return fmt.Errorf("failed to start system manager: %w", err)
	}

	// Start HTTP server
	go func() {
		s.logger.WithField("addr", s.httpServer.Addr).Info("Starting HTTP server")
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.WithError(err).Fatal("HTTP server failed")
		}
	}()

	// Start metrics server
	go func() {
		s.logger.WithField("addr", s.metricsServer.Addr).Info("Starting metrics server")
		if err := s.metricsServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.WithError(err).Fatal("Metrics server failed")
		}
	}()

	s.logger.Info("AIOS daemon started successfully")
	return nil
}

func (s *Server) WaitForShutdown() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	s.logger.Info("Shutting down AIOS daemon...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown HTTP server
	if err := s.httpServer.Shutdown(ctx); err != nil {
		s.logger.WithError(err).Error("Failed to shutdown HTTP server")
	}

	// Shutdown metrics server
	if err := s.metricsServer.Shutdown(ctx); err != nil {
		s.logger.WithError(err).Error("Failed to shutdown metrics server")
	}

	// Shutdown system manager
	if err := s.systemManager.Stop(ctx); err != nil {
		s.logger.WithError(err).Error("Failed to shutdown system manager")
	}

	s.logger.Info("AIOS daemon shutdown complete")
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
