package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/giantswarm/dev-platform-kratix-promises/mcp_server/internal/config"
	"github.com/giantswarm/dev-platform-kratix-promises/mcp_server/internal/logging"
	"github.com/giantswarm/dev-platform-kratix-promises/mcp_server/internal/server"
	"github.com/joho/godotenv"
)

func printHelp() {
	fmt.Printf(`%s v%s - Platform Building Blocks MCP Server

USAGE:
    %s [OPTIONS]

DESCRIPTION:
    An MCP (Model Context Protocol) server that provides tools for discovering,
    managing, and validating platform building blocks in a Kubernetes cluster.
    
    Building blocks are reusable platform components that developers can use
    to create applications and infrastructure resources.
    
    This server exposes four main tools:
    - list_building_blocks: Discover all available platform building blocks
    - get_building_block_schema: Get the OpenAPI schema for a specific building block
    - validate_building_block_spec: Validate resource specifications against building block schemas
    - create_building_block: Create Custom Resource instances using building block schemas

OPTIONS:
    -h, --help          Show this help message and exit

CONFIGURATION:
    The server is configured using environment variables or a .env file:

    Server Configuration:
        MCP_HOST            Server host address (default: localhost)
        MCP_PORT            Server port number (default: 8080)

    Logging Configuration:
        LOG_LEVEL           Log level: debug, info, warn, error (default: info)
        LOG_FORMAT          Log format: json, text (default: json)

    Kubernetes Configuration:
        KUBECONFIG          Path to kubeconfig file (default: ~/.kube/config)
        KUBE_CONTEXT        Kubernetes context to use (default: current context)
        K8S_TIMEOUT         Kubernetes API timeout (default: 30s)

EXAMPLES:
    # Start server with default configuration
    %s

    # Start server on specific port
    MCP_PORT=9090 %s

    # Start server with debug logging
    LOG_LEVEL=debug %s

    # Start server with specific kubeconfig
    KUBECONFIG=/path/to/config %s

    # View configuration file example
    cat > .env << EOF
    MCP_HOST=0.0.0.0
    MCP_PORT=8080
    LOG_LEVEL=info
    LOG_FORMAT=json
    KUBECONFIG=~/.kube/config
    KUBE_CONTEXT=my-cluster
    K8S_TIMEOUT=30s
    EOF

ENVIRONMENT FILES:
    The server will automatically load configuration from a .env file in the
    current directory if it exists. Environment variables take precedence
    over .env file values.

MCP PROTOCOL:
    This server implements the Model Context Protocol (MCP) which allows
    AI assistants to interact with external tools and data sources. The
    server exposes platform building blocks and validation capabilities
    to AI agents for infrastructure management tasks.

`, config.AppName, config.Version, config.AppName, config.AppName, config.AppName, config.AppName, config.AppName)
}

func main() {
	// Parse command line arguments
	var showHelp bool
	flag.BoolVar(&showHelp, "h", false, "Show help message")
	flag.BoolVar(&showHelp, "help", false, "Show help message")

	// Custom usage function
	flag.Usage = func() {
		printHelp()
	}

	flag.Parse()

	// Show help if requested
	if showHelp {
		printHelp()
		os.Exit(0)
	}

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
		"app_name", config.AppName,
		"version", config.Version,
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
