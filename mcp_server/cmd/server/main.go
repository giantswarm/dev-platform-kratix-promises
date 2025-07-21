package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/giantswarm/dev-platform-kratix-promises/mcp_server/internal/config"
	"github.com/giantswarm/dev-platform-kratix-promises/mcp_server/internal/logging"
	"github.com/giantswarm/dev-platform-kratix-promises/mcp_server/internal/server"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables from .env file if it exists
	if err := godotenv.Load(); err != nil {
		// It's OK if .env doesn't exist, just log and continue
		slog.Debug("No .env file found, using environment variables and defaults")
	}

	// Load configuration
	cfg := config.Load()
	if err := cfg.Validate(); err != nil {
		slog.Error("Invalid configuration", "error", err)
		os.Exit(1)
	}

	// Setup logging
	logging.SetupLogger(cfg.LogLevel, cfg.LogFormat)
	logger := slog.Default().With("component", "main")

	logger.Info("Starting IDP MCP Server",
		"app_name", cfg.AppName,
		"version", cfg.AppVersion,
		"log_level", cfg.LogLevel)

	// Create and initialize the server
	srv := server.New(cfg)
	if err := srv.Initialize(); err != nil {
		logger.Error("Failed to initialize server", "error", err)
		os.Exit(1)
	}

	// Setup graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle shutdown signals
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

		sig := <-sigChan
		logger.Info("Received shutdown signal", "signal", sig)
		cancel()
	}()

	// Run the server
	if err := srv.Run(ctx); err != nil {
		logger.Error("Server error", "error", err)
		os.Exit(1)
	}

	// Graceful shutdown
	logger.Info("Server stopped gracefully")
}
