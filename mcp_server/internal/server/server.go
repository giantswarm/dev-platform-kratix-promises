package server

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/giantswarm/dev-platform-kratix-promises/mcp_server/internal/clients"
	"github.com/giantswarm/dev-platform-kratix-promises/mcp_server/internal/config"
	"github.com/giantswarm/dev-platform-kratix-promises/mcp_server/internal/resources"
	"github.com/giantswarm/dev-platform-kratix-promises/mcp_server/internal/tools"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// Server wraps the MCP server with our configuration
type Server struct {
	config    *config.Config
	mcpServer *server.MCPServer
	k8sClient *clients.KubernetesClient
	logger    *slog.Logger
}

// New creates a new MCP server instance
func New(cfg *config.Config) *Server {
	logger := slog.Default().With("component", "mcp-server")

	return &Server{
		config: cfg,
		logger: logger,
	}
}

// Initialize sets up the MCP server with tools, resources, and prompts
func (s *Server) Initialize() error {
	s.logger.Info("Initializing MCP server",
		"app_name", config.AppName,
		"version", config.Version)

	// Create the MCP server with resource capabilities enabled
	mcpServer := server.NewMCPServer(
		config.AppName,
		config.Version,
		server.WithResourceCapabilities(false, false), // subscribe=false, listChanged=false for now
	)

	// Initialize Kubernetes client
	k8sClient, err := clients.NewKubernetesClient(s.config)
	if err != nil {
		s.logger.Error("Failed to initialize Kubernetes client", "error", err)
		return fmt.Errorf("failed to initialize Kubernetes client: %w", err)
	}
	s.k8sClient = k8sClient

	// Create CRD resource handler
	resourceHandler := resources.NewCRDResourceHandler(k8sClient)

	// Register MCP resources for Giant Swarm CRDs
	s.registerResources(mcpServer, resourceHandler)

	// Register MCP tools for Promise operations
	s.registerTools(mcpServer, k8sClient)

	// TODO: In the future, we'll register more tools here for:
	// - High-level platform capabilities
	// - Kubernetes API interactions beyond Promises
	// - GitHub repository operations
	// - IDP platform specific operations

	s.logger.Info("MCP server initialized with Kubernetes CRD resources",
		"context", k8sClient.GetCurrentContext())

	s.mcpServer = mcpServer
	return nil
}

// registerResources registers all MCP resources with the server
func (s *Server) registerResources(mcpServer *server.MCPServer, resourceHandler *resources.CRDResourceHandler) {
	// Register AppDeployment resources
	appDeploymentResource := mcp.NewResource(
		"k8s://appdeployments",
		"App Deployments",
		mcp.WithResourceDescription("Giant Swarm application deployment resources"),
		mcp.WithMIMEType("application/json"),
	)
	mcpServer.AddResource(appDeploymentResource, resourceHandler.HandleAppDeployments)
	s.logger.Info("Registered MCP resources for App Deployments")

	// Register GitHubApp resources
	githubAppResource := mcp.NewResource(
		"k8s://githubapps",
		"GitHub Apps",
		mcp.WithResourceDescription("Giant Swarm GitHub application resources"),
		mcp.WithMIMEType("application/json"),
	)
	mcpServer.AddResource(githubAppResource, resourceHandler.HandleGitHubApps)
	s.logger.Info("Registered MCP resources for GitHub Apps")

	// Register GitHubRepo resources
	githubRepoResource := mcp.NewResource(
		"k8s://githubrepos",
		"GitHub Repositories",
		mcp.WithResourceDescription("Giant Swarm GitHub repository resources"),
		mcp.WithMIMEType("application/json"),
	)
	mcpServer.AddResource(githubRepoResource, resourceHandler.HandleGitHubRepos)
	s.logger.Info("Registered MCP resources for GitHub Repositories")
}

// registerTools registers all MCP tools with the server
func (s *Server) registerTools(mcpServer *server.MCPServer, k8sClient *clients.KubernetesClient) {
	// Create Promise tools handler (internal implementation uses Kratix Promises)
	promiseTools := tools.NewPromiseToolsHandler(k8sClient)

	// Register list_building_blocks tool
	listTool := mcp.NewTool("list_building_blocks",
		mcp.WithDescription("List all available platform building blocks in the cluster"),
	)
	mcpServer.AddTool(listTool, promiseTools.HandleListKratixPromises)
	s.logger.Info("Registered MCP tool: list_building_blocks")

	// Register get_building_block_schema tool
	getSchemaTool := mcp.NewTool("get_building_block_schema",
		mcp.WithDescription("Get the complete OpenAPI schema for a specific platform building block"),
		mcp.WithString("building_block_name", mcp.Description("Name of the building block to get schema for"), mcp.Required()),
	)
	mcpServer.AddTool(getSchemaTool, promiseTools.HandleGetPromiseSchema)
	s.logger.Info("Registered MCP tool: get_building_block_schema")

	// Register validate_building_block_spec tool
	validateTool := mcp.NewTool("validate_building_block_spec",
		mcp.WithDescription("Validate a resource specification against a platform building block's schema"),
		mcp.WithString("building_block_name", mcp.Description("Name of the building block to validate against"), mcp.Required()),
		mcp.WithString("spec", mcp.Description("JSON string containing the resource specification to validate"), mcp.Required()),
	)
	mcpServer.AddTool(validateTool, promiseTools.HandleValidatePromiseSpec)
	s.logger.Info("Registered MCP tool: validate_building_block_spec")

	// Register create_building_block tool
	createTool := mcp.NewTool("create_building_block",
		mcp.WithDescription("Create a new Custom Resource instance based on a platform building block schema"),
		mcp.WithString("building_block_name", mcp.Description("Name of the building block (Promise) to use for schema"), mcp.Required()),
		mcp.WithString("resource_name", mcp.Description("Name for the new Custom Resource instance"), mcp.Required()),
		mcp.WithString("spec", mcp.Description("JSON specification for the Custom Resource following the building block's schema"), mcp.Required()),
		mcp.WithString("namespace", mcp.Description("Target namespace for the Custom Resource"), mcp.Required()),
	)
	mcpServer.AddTool(createTool, promiseTools.HandleCreateBuildingBlock)
	s.logger.Info("Registered MCP tool: create_building_block")

	s.logger.Info("Successfully registered all platform building block MCP tools", "tools_count", 4)
}

// Run starts the MCP server
func (s *Server) Run(ctx context.Context) error {
	if s.mcpServer == nil {
		return fmt.Errorf("server not initialized")
	}

	s.logger.Info("Starting MCP server", "host", s.config.Host, "port", s.config.Port)

	// Run the server using the correct serve function (this blocks until interrupted)
	if err := server.ServeStdio(s.mcpServer); err != nil {
		s.logger.Error("MCP server error", "error", err)
		return fmt.Errorf("failed to serve MCP server: %w", err)
	}

	return nil
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Info("Shutting down MCP server")

	// TODO: Add proper shutdown logic when needed
	return nil
}
