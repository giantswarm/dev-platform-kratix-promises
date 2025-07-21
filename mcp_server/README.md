# IDP MCP Server

A Model Context Protocol (MCP) server implementation for the IDP demo platform, providing AI systems with tools to interact with Kubernetes APIs and GitHub repositories.

## Overview

This MCP server provides access to Giant Swarm platform Custom Resource Definitions (CRDs) through the Model Context Protocol. It connects to your current Kubernetes cluster context and exposes three main CRD resource types as MCP resources.

## Features

- ✅ **MCP Protocol Support**: Full JSON-RPC 2.0 over stdio implementation
- ✅ **Kubernetes Integration**: Connects to K8s cluster using current kubeconfig context
- ✅ **Giant Swarm CRDs**: Access to AppDeployment, GitHubApp, and GitHubRepo resources
- ✅ **Security**: Automatic sanitization of sensitive data (secrets, tokens)
- ✅ **Structured Logging**: Using Go's `slog` package with JSON/text output
- ✅ **Configuration Management**: Environment variable based configuration
- ✅ **Graceful Shutdown**: Proper signal handling for clean shutdown
- ✅ **Docker Support**: Production-ready containerization
- 🚧 **Tools**: Not implemented yet (future: Kubernetes operations)
- 🚧 **GitHub Integration**: Not implemented yet (future: Git repository operations)

## Available MCP Resources

The server exposes the following MCP resources:

### 1. App Deployments (`k8s://appdeployments`)
- **Description**: Giant Swarm application deployment resources
- **API**: `promise.platform.giantswarm.io/v1beta1/appdeployments`
- **Scope**: Cluster-wide (all namespaces)

### 2. GitHub Apps (`k8s://githubapps`)
- **Description**: Giant Swarm GitHub application resources  
- **API**: `promise.platform.giantswarm.io/v1beta1/githubapps`
- **Scope**: Cluster-wide (all namespaces)

### 3. GitHub Repositories (`k8s://githubrepos`)
- **Description**: Giant Swarm GitHub repository resources
- **API**: `promise.platform.giantswarm.io/v1beta1/githubrepos`
- **Scope**: Cluster-wide (all namespaces)

## Quick Start

### Prerequisites

- Go 1.21 or later
- Access to a Kubernetes cluster with Giant Swarm CRDs installed
- Valid kubeconfig file (typically at `~/.kube/config`)

### Installation

```bash
# Clone and build
cd mcp_server
go mod tidy
go build -o mcp-server ./cmd/server

# Run the server
./mcp-server
```

### Configuration

Configure the server using environment variables:

```bash
# Basic configuration
export APP_NAME="idp-mcp-server"
export APP_VERSION="1.0.0"
export LOG_LEVEL="info"
export LOG_FORMAT="json"

# Kubernetes configuration
export KUBE_CONFIG_PATH="/path/to/kubeconfig"  # Optional, defaults to ~/.kube/config
export KUBE_CONTEXT="my-cluster-context"      # Optional, uses current context
export K8S_TIMEOUT="30s"                      # Optional, default 30s

# Or create a .env file
echo "LOG_LEVEL=debug" > .env
echo "KUBE_CONTEXT=my-cluster" >> .env
```

## Testing

### Manual Testing

You can test the MCP server by sending JSON-RPC requests over stdin:

```bash
# Start the server
./mcp-server

# Send initialization request (in another terminal)
echo '{"jsonrpc": "2.0", "id": 1, "method": "initialize", "params": {"clientInfo": {"name": "test-client", "version": "1.0.0"}}}' | ./mcp-server

# List available resources
echo '{"jsonrpc": "2.0", "id": 2, "method": "resources/list", "params": {}}' | ./mcp-server

# Read AppDeployment resources
echo '{"jsonrpc": "2.0", "id": 3, "method": "resources/read", "params": {"uri": "k8s://appdeployments"}}' | ./mcp-server
```

### Build and Test

```bash
# Build the application
go build -o mcp-server ./cmd/server

# Run tests
go test ./...

# Test compilation of all packages
go build ./...
```

## Security Features

### Data Sanitization

The server automatically removes sensitive information from Kubernetes resources:

- **Secret References**: `githubTokenSecretRef`, `passwordSecretRef`, etc.
- **Kubernetes Metadata**: `managedFields`, `resourceVersion`, `selfLink`
- **Token Values**: Direct token or password values in specifications

### Access Control

- Uses your current Kubernetes RBAC permissions
- Only reads resources (no write operations)
- Connects using your existing kubeconfig authentication

## Error Handling

The server provides comprehensive error handling for common scenarios:

- **NotFound**: When no CRD resources exist
- **Forbidden**: When RBAC denies access
- **Unauthorized**: When authentication fails  
- **InternalError**: For other Kubernetes API errors

## Development

### Project Structure

```
mcp_server/
├── cmd/server/           # Application entry point
├── internal/
│   ├── config/          # Configuration management
│   ├── logging/         # Structured logging setup
│   ├── clients/         # Kubernetes client wrapper
│   ├── resources/       # CRD resource handlers and types
│   └── server/          # MCP server implementation
├── Dockerfile           # Container image
└── README.md           # This file
```

### Adding New CRD Types

To add support for new CRDs:

1. Add the GVR definition in `internal/resources/types.go`
2. Create a handler method in `internal/resources/crd_handler.go`
3. Register the resource in `internal/server/server.go`

## Docker

```bash
# Build image
docker build -t idp-mcp-server .

# Run container (requires kubeconfig mount)
docker run --rm \
  -v ~/.kube/config:/root/.kube/config:ro \
  -e KUBE_CONTEXT=my-cluster \
  idp-mcp-server
```

## Troubleshooting

### Common Issues

1. **"Failed to load kubeconfig"**
   - Verify `KUBE_CONFIG_PATH` points to valid kubeconfig
   - Ensure file permissions allow reading

2. **"Access denied to Kubernetes cluster"**
   - Check RBAC permissions for your user/service account
   - Verify you can run `kubectl get appdeployments` manually

3. **"No resources found"**  
   - Confirm Giant Swarm CRDs are installed in your cluster
   - Verify resources exist: `kubectl get appdeployments --all-namespaces`

4. **"Authentication required"**
   - Check if your kubeconfig token is valid
   - Try `kubectl cluster-info` to test connectivity

### Debug Mode

Run with debug logging to see detailed information:

```bash
LOG_LEVEL=debug ./mcp-server
```

## Future Enhancements

- **Tools**: Kubernetes operations (create, update, delete resources)
- **GitHub Integration**: Repository management and operations
- **Namespace Filtering**: Limit resources to specific namespaces
- **Resource Watching**: Real-time updates for resource changes
- **Custom Queries**: Filter resources by labels, fields, etc.

## Contributing

1. Follow Go best practices and conventions
2. Add tests for new functionality
3. Update documentation for user-facing changes
4. Use structured logging with appropriate levels 