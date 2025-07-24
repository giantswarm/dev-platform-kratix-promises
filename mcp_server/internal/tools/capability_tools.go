package tools

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/giantswarm/dev-platform-kratix-promises/mcp_server/internal/clients"
	"github.com/giantswarm/dev-platform-kratix-promises/mcp_server/internal/resources"
	"github.com/mark3labs/mcp-go/mcp"
)

// CapabilityToolsHandler manages high-level platform capability tools
type CapabilityToolsHandler struct {
	k8sClient    clients.KubernetesClientInterface
	logger       *slog.Logger
	capabilities []resources.PlatformCapability
	groups       []resources.CapabilityGroup
}

// NewCapabilityToolsHandler creates a new handler for platform capability tools
func NewCapabilityToolsHandler(k8sClient clients.KubernetesClientInterface) *CapabilityToolsHandler {
	handler := &CapabilityToolsHandler{
		k8sClient: k8sClient,
		logger:    slog.Default().With("component", "capability-tools"),
	}

	// Initialize predefined capabilities
	handler.initializeCapabilities()
	return handler
}

// HandleListPlatformCapabilities handles the list_platform_capabilities tool call
func (h *CapabilityToolsHandler) HandleListPlatformCapabilities(arguments map[string]interface{}) (*mcp.CallToolResult, error) {
	h.logger.Info("Handling list_platform_capabilities tool call")

	// Build response
	response := &resources.PlatformCapabilitiesResponse{
		Capabilities: h.capabilities,
		Groups:       h.groups,
		Metadata: map[string]interface{}{
			"total_capabilities": len(h.capabilities),
			"total_groups":       len(h.groups),
			"last_updated":       time.Now().Format(time.RFC3339),
		},
	}

	// Format as JSON
	jsonResponse, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to format response: %v", err)), nil
	}

	h.logger.Info("Successfully listed platform capabilities",
		"capabilities_count", len(h.capabilities),
		"groups_count", len(h.groups))

	return mcp.NewToolResultText(string(jsonResponse)), nil
}

// initializeCapabilities sets up predefined platform capabilities
func (h *CapabilityToolsHandler) initializeCapabilities() {
	// Define capability groups
	h.groups = []resources.CapabilityGroup{
		{
			Name:         "microservice-development",
			Description:  "Capabilities for microservice development and deployment",
			Capabilities: []string{"create-python-microservice-with-database", "create-go-microservice-with-database"},
		},
	}

	// Define capabilities with structured prompts
	h.capabilities = []resources.PlatformCapability{
		{
			Name:        "create-python-microservice-with-database",
			Groups:      []string{"microservice-development"},
			Description: "Create a complete python microservice with backend services and database connectivity",
			Prompt: `To create a python microservice with database, follow these steps:

1. First, get the available building blocks:
   - Call 'list_building_blocks' to see available options
   - Look for 'githubapp' building block

2. Create a GitHubApp buuilding block that will initialize a repository with application template and then deploy it:
   - Call 'get_building_block_schema' with building_block_name: "githubapp"
   - Call 'create_building_block' with:
     * building_block_name: "githubapp"
     * resource_name: "{app-name}-app"
     * namespace: "{target-namespace}"
     * spec: Include both githubRepo and appDeployment configurations
       - githubRepo: repository setup with template
       - appDeployment: database configuration and Helm release setup
       - use 'giantswarm/devplatform-template-go-service' as templateSource

3. The githubapp building block will automatically:
   - Create a GitHub repository from template
   - Set up CI/CD pipeline
   - Deploy the application with database connectivity
   - Configure ingress and monitoring

Parameters to customize:
- app-name: Name for your application
- target-namespace: Kubernetes namespace for deployment
- repository-owner: GitHub organization or user
- database-engine: Database type (postgresql, mysql, etc.)
- ingress-host: Domain for your application

Example spec structure:
{
"appDeployment": {
    "name": "kubedemo",
    "spec": {
        "database": {
            "engine": "aurora-postgresql",
            "eso": {
                "clusterSsaField": "demotech_rcc",
                "tenantCluster": {
                    "apiServerEndpoint": "demotech-rds-apiserver-852993111.eu-central-1.elb.amazonaws.com",
                    "clusterName": "demotech-rds",
                    "enabled": true
                }
            },
            "providerConfigRef": {
                "name": "demotech-rcc-postgresql-provider-config"
            }
        },
        "ingressHost": "kubedemo.demotech-rds.awsprod.gigantic.io",
        "interval": "1m",
        "kubeConfig": {
            "secretRef": {
                "name": "demotech-rds-kubeconfig"
            }
        },
        "storageNamespace": "default",
        "suspend": false,
        "targetNamespace": "default",
        "timeout": "3m",
        "version": "\u003e=0.1.0-0"
    }
},
"githubRepo": {
    "name": "kubedemo",
    "spec": {
        "backstageCatalogEntity": {
            "lifecycle": "experimental",
            "owner": "team-service-engineering"
        },
        "githubTokenSecretRef": {
            "name": "github-app-secret"
        },
        "registryInfoConfigMapRef": {
            "name": "github-oci-registry-info"
        },
        "repository": {
            "description": "KubeCon 2025 London",
            "name": "kubedemo",
            "owner": "DemoTechInc",
            "templateSource": "giantswarm/devplatform-template-go-service",
            "visibility": "private"
        }
    }
}`,
	},
  }

	h.logger.Info("Initialized platform capabilities",
		"capabilities_count", len(h.capabilities),
		"groups_count", len(h.groups))
}
