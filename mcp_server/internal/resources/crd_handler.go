package resources

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/giantswarm/dev-platform-kratix-promises/mcp_server/internal/clients"
	"github.com/mark3labs/mcp-go/mcp"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// CRDResourceHandler handles Giant Swarm CRD resources for MCP
type CRDResourceHandler struct {
	k8sClient clients.KubernetesClientInterface
	logger    *slog.Logger
}

// NewCRDResourceHandler creates a new CRD resource handler
func NewCRDResourceHandler(k8sClient clients.KubernetesClientInterface) *CRDResourceHandler {
	return &CRDResourceHandler{
		k8sClient: k8sClient,
		logger:    slog.Default().With("component", "crd-handler"),
	}
}

// HandleAppDeployments handles MCP resource requests for AppDeployment CRDs
func (h *CRDResourceHandler) HandleAppDeployments(request mcp.ReadResourceRequest) ([]interface{}, error) {
	h.logger.Info("Handling AppDeployment resources request")

	// List all AppDeployment resources across all namespaces
	resources, err := h.k8sClient.ListResources(AppDeploymentGVR, "")
	if err != nil {
		return h.handleKubernetesErrorAsInterface(err, AppDeploymentType)
	}

	// Sanitize and format the response
	h.sanitizeResources(resources)
	response := h.formatResourceResponse(resources, AppDeploymentType, "")

	return []interface{}{response}, nil
}

// HandleGitHubApps handles MCP resource requests for GitHubApp CRDs
func (h *CRDResourceHandler) HandleGitHubApps(request mcp.ReadResourceRequest) ([]interface{}, error) {
	h.logger.Info("Handling GitHubApp resources request")

	// List all GitHubApp resources across all namespaces
	resources, err := h.k8sClient.ListResources(GitHubAppGVR, "")
	if err != nil {
		return h.handleKubernetesErrorAsInterface(err, GitHubAppType)
	}

	// Sanitize and format the response
	h.sanitizeResources(resources)
	response := h.formatResourceResponse(resources, GitHubAppType, "")

	return []interface{}{response}, nil
}

// HandleGitHubRepos handles MCP resource requests for GitHubRepo CRDs
func (h *CRDResourceHandler) HandleGitHubRepos(request mcp.ReadResourceRequest) ([]interface{}, error) {
	h.logger.Info("Handling GitHubRepo resources request")

	// List all GitHubRepo resources across all namespaces
	resources, err := h.k8sClient.ListResources(GitHubRepoGVR, "")
	if err != nil {
		return h.handleKubernetesErrorAsInterface(err, GitHubRepoType)
	}

	// Sanitize and format the response
	h.sanitizeResources(resources)
	response := h.formatResourceResponse(resources, GitHubRepoType, "")

	return []interface{}{response}, nil
}

// formatResourceResponse formats Kubernetes resources into a standardized response
func (h *CRDResourceHandler) formatResourceResponse(resources *unstructured.UnstructuredList, resourceType CRDResourceType, namespace string) *ResourceResponse {
	response := &ResourceResponse{
		APIVersion: "promise.platform.giantswarm.io/v1beta1",
		Kind:       string(resourceType) + "List",
		Items:      make([]map[string]interface{}, len(resources.Items)),
		Metadata: ResponseMetadata{
			Namespace:   namespace,
			Count:       len(resources.Items),
			LastUpdated: time.Now(),
			ClusterInfo: h.k8sClient.GetClusterInfo(),
		},
	}

	// Process each resource
	for i, item := range resources.Items {
		response.Items[i] = item.Object
	}

	h.logger.Debug("Formatted resource response",
		"type", resourceType,
		"count", len(resources.Items),
		"namespace", namespace)

	return response
}

// sanitizeResources removes sensitive information from Kubernetes resources
func (h *CRDResourceHandler) sanitizeResources(resources *unstructured.UnstructuredList) {
	for i := range resources.Items {
		h.sanitizeResource(&resources.Items[i])
	}
}

// sanitizeResource removes sensitive fields from a single resource
func (h *CRDResourceHandler) sanitizeResource(resource *unstructured.Unstructured) {
	// Remove sensitive secret references and tokens
	sensitiveFields := [][]string{
		// TODO: Add sensitive fields here
	}

	for _, fieldPath := range sensitiveFields {
		unstructured.RemoveNestedField(resource.Object, fieldPath...)
	}

	// Remove managed fields and other Kubernetes metadata that might be sensitive
	if metadata, found, err := unstructured.NestedMap(resource.Object, "metadata"); found && err == nil {
		delete(metadata, "managedFields")
		delete(metadata, "resourceVersion")
		delete(metadata, "selfLink")
		unstructured.SetNestedMap(resource.Object, metadata, "metadata")
	}

	h.logger.Debug("Sanitized resource",
		"name", resource.GetName(),
		"namespace", resource.GetNamespace(),
		"kind", resource.GetKind())
}

// handleKubernetesErrorAsInterface handles Kubernetes API errors and converts them to interface slice
func (h *CRDResourceHandler) handleKubernetesErrorAsInterface(err error, resourceType CRDResourceType) ([]interface{}, error) {
	var errorMessage string
	var errorDetails map[string]interface{}

	if apierrors.IsNotFound(err) {
		errorMessage = fmt.Sprintf("No %s resources found", resourceType)
		errorDetails = map[string]interface{}{
			"error":  errorMessage,
			"reason": "NotFound",
			"type":   string(resourceType),
		}
		h.logger.Info("No resources found", "type", resourceType)
	} else if apierrors.IsForbidden(err) {
		errorMessage = "Access denied to Kubernetes cluster"
		errorDetails = map[string]interface{}{
			"error":  errorMessage,
			"reason": "Forbidden",
			"type":   string(resourceType),
		}
		h.logger.Error("Access denied to Kubernetes cluster", "type", resourceType, "error", err)
	} else if apierrors.IsUnauthorized(err) {
		errorMessage = "Authentication required for Kubernetes cluster"
		errorDetails = map[string]interface{}{
			"error":  errorMessage,
			"reason": "Unauthorized",
			"type":   string(resourceType),
		}
		h.logger.Error("Authentication required", "type", resourceType, "error", err)
	} else {
		errorMessage = fmt.Sprintf("Failed to retrieve %s resources: %v", resourceType, err)
		errorDetails = map[string]interface{}{
			"error":  errorMessage,
			"reason": "InternalError",
			"type":   string(resourceType),
		}
		h.logger.Error("Failed to retrieve resources", "type", resourceType, "error", err)
	}

	return []interface{}{errorDetails}, nil
}
