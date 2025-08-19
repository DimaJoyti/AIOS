package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aios/aios/internal/ai"
	"github.com/aios/aios/pkg/config"
	"github.com/aios/aios/pkg/database"
	"github.com/sirupsen/logrus"
)

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

	// Initialize database connection
	dbConfig := database.Config{
		Host:            cfg.Database.Host,
		Port:            cfg.Database.Port,
		User:            cfg.Database.User,
		Password:        cfg.Database.Password,
		Database:        cfg.Database.Name,
		SSLMode:         cfg.Database.SSLMode,
		MaxOpenConns:    cfg.Database.MaxOpenConns,
		MaxIdleConns:    cfg.Database.MaxIdleConns,
		ConnMaxLifetime: cfg.Database.ConnMaxLifetime,
	}

	db, err := database.NewConnection(dbConfig)
	if err != nil {
		logger.WithError(err).Fatal("Failed to connect to database")
	}
	defer db.Close()

	// Create AI service
	aiService, err := ai.NewService(cfg, db, logger)
	if err != nil {
		logger.WithError(err).Fatal("Failed to create AI service")
	}

	// Start service
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := aiService.Start(ctx); err != nil {
		logger.WithError(err).Fatal("Failed to start AI service")
	}

	// Wait for shutdown signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	logger.Info("Shutting down AI service...")

	// Graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := aiService.Stop(shutdownCtx); err != nil {
		logger.WithError(err).Error("Error during shutdown")
	}

	logger.Info("AI service stopped")
}
