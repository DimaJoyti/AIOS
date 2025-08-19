package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aios/aios/internal/knowledge"
	"github.com/aios/aios/internal/mcp/enhanced"
	"github.com/aios/aios/pkg/config"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
)

func main() {
	// Initialize logger
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)
	logger.SetFormatter(&logrus.JSONFormatter{})

	// Load configuration (mock for now)
	cfg := &config.Config{
		Services: config.ServicesConfig{
			Knowledge: config.KnowledgeServiceConfig{
				Enabled: true,
				Host:    "0.0.0.0",
				Port:    8181,
			},
			MCP: config.MCPServiceConfig{
				Enabled: true,
				Host:    "0.0.0.0",
				Port:    8051,
			},
		},
	}

	// Create database connection
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:password@localhost:5432/aios?sslmode=disable"
	}

	db, err := sqlx.Connect("postgres", dbURL)
	if err != nil {
		logger.WithError(err).Fatal("Failed to connect to database")
	}
	defer db.Close()

	// Test database connection
	if err := db.Ping(); err != nil {
		logger.WithError(err).Fatal("Failed to ping database")
	}

	logger.Info("Database connection established")

	// Create knowledge service first
	knowledgeService, err := knowledge.NewService(cfg, db, logger)
	if err != nil {
		logger.WithError(err).Fatal("Failed to create knowledge service")
	}

	// Create enhanced MCP service
	service, err := enhanced.NewService(cfg, db, knowledgeService, logger)
	if err != nil {
		logger.WithError(err).Fatal("Failed to create enhanced MCP service")
	}

	// Start services
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start knowledge service
	if err := knowledgeService.Start(ctx); err != nil {
		logger.WithError(err).Fatal("Failed to start knowledge service")
	}

	// Start MCP service
	if err := service.Start(ctx); err != nil {
		logger.WithError(err).Fatal("Failed to start enhanced MCP service")
	}

	// Wait for shutdown signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	logger.Info("Shutting down enhanced MCP service...")

	// Graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := service.Stop(shutdownCtx); err != nil {
		logger.WithError(err).Error("Error during MCP service shutdown")
	}

	if err := knowledgeService.Stop(shutdownCtx); err != nil {
		logger.WithError(err).Error("Error during knowledge service shutdown")
	}

	logger.Info("Enhanced MCP service stopped")
}
