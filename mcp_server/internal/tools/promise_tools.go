package tools

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/giantswarm/dev-platform-kratix-promises/mcp_server/internal/clients"
	"github.com/giantswarm/dev-platform-kratix-promises/mcp_server/internal/resources"
	"github.com/giantswarm/dev-platform-kratix-promises/mcp_server/internal/validation"
	"github.com/mark3labs/mcp-go/mcp"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// PromiseToolsHandler manages Promise-related MCP tools
type PromiseToolsHandler struct {
	extractor *PromiseExtractor
	validator *validation.SchemaValidator
	k8sClient clients.KubernetesClientInterface
	logger    *slog.Logger
}

// NewPromiseToolsHandler creates a new handler for Promise-related tools
func NewPromiseToolsHandler(k8sClient clients.KubernetesClientInterface) *PromiseToolsHandler {
	return &PromiseToolsHandler{
		extractor: NewPromiseExtractor(k8sClient),
		validator: validation.NewSchemaValidator(),
		k8sClient: k8sClient,
		logger:    slog.Default().With("component", "promise-tools"),
	}
}

// HandleListKratixPromises handles the list_building_blocks tool call
func (h *PromiseToolsHandler) HandleListKratixPromises(arguments map[string]interface{}) (*mcp.CallToolResult, error) {
	h.logger.Info("Handling list_building_blocks tool call")

	// Get all building block summaries (internally using Promise summaries)
	summaries, err := h.extractor.ListPromiseSummaries()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to list platform building blocks: %v", err)), nil
	}

	// Get cluster info for metadata
	clusterInfo := h.k8sClient.GetClusterInfo()
	currentContext := h.k8sClient.GetCurrentContext()

	// Build response with metadata
	response := map[string]interface{}{
		"building_blocks": summaries,
		"metadata": map[string]interface{}{
			"total_count":     len(summaries),
			"cluster_context": currentContext,
			"last_updated":    time.Now().Format(time.RFC3339),
			"cluster_info":    clusterInfo,
		},
	}

	// Format as JSON
	jsonResponse, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to format response: %v", err)), nil
	}

	h.logger.Info("Successfully listed platform building blocks", "count", len(summaries))
	return mcp.NewToolResultText(string(jsonResponse)), nil
}

// HandleGetPromiseSchema handles the get_building_block_schema tool call
func (h *PromiseToolsHandler) HandleGetPromiseSchema(arguments map[string]interface{}) (*mcp.CallToolResult, error) {
	h.logger.Info("Handling get_building_block_schema tool call")

	// Extract building_block_name parameter
	buildingBlockNameRaw, exists := arguments["building_block_name"]
	if !exists {
		return mcp.NewToolResultError("Missing required parameter 'building_block_name'"), nil
	}

	buildingBlockName, ok := buildingBlockNameRaw.(string)
	if !ok {
		return mcp.NewToolResultError("Parameter 'building_block_name' must be a string"), nil
	}

	// Get the Promise schema
	schema, err := h.extractor.GetPromiseSchema(buildingBlockName)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get building block schema for '%s': %v", buildingBlockName, err)), nil
	}

	// Format as JSON
	jsonResponse, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to format schema response: %v", err)), nil
	}

	h.logger.Info("Successfully retrieved building block schema", "building_block", buildingBlockName)
	return mcp.NewToolResultText(string(jsonResponse)), nil
}

// HandleValidatePromiseSpec handles the validate_building_block_spec tool call
func (h *PromiseToolsHandler) HandleValidatePromiseSpec(arguments map[string]interface{}) (*mcp.CallToolResult, error) {
	h.logger.Info("Handling validate_building_block_spec tool call")

	// Extract building_block_name parameter
	buildingBlockNameRaw, exists := arguments["building_block_name"]
	if !exists {
		return mcp.NewToolResultError("Missing required parameter 'building_block_name'"), nil
	}

	buildingBlockName, ok := buildingBlockNameRaw.(string)
	if !ok {
		return mcp.NewToolResultError("Parameter 'building_block_name' must be a string"), nil
	}

	// Extract spec parameter
	specRaw, exists := arguments["spec"]
	if !exists {
		return mcp.NewToolResultError("Missing required parameter 'spec'"), nil
	}

	specString, ok := specRaw.(string)
	if !ok {
		return mcp.NewToolResultError("Parameter 'spec' must be a JSON string"), nil
	}

	// Parse the JSON string into a map
	var spec map[string]interface{}
	if err := json.Unmarshal([]byte(specString), &spec); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid JSON in 'spec' parameter: %v", err)), nil
	}

	// Get the Promise schema
	promiseSchema, err := h.extractor.GetPromiseSchema(buildingBlockName)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get building block schema for '%s': %v", buildingBlockName, err)), nil
	}

	// Validate the spec against the schema
	validationDetails, err := h.validator.ValidateSpec(promiseSchema.OpenAPISchema, spec)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to validate spec: %v", err)), nil
	}

	// Build validation result
	result := &resources.ValidationResult{
		Valid:            len(validationDetails.Errors) == 0,
		PromiseName:      buildingBlockName,
		ValidationResult: *validationDetails,
	}

	// Format as JSON
	jsonResponse, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to format validation response: %v", err)), nil
	}

	h.logger.Info("Successfully validated building block spec",
		"building_block", buildingBlockName,
		"valid", result.Valid,
		"errors", len(validationDetails.Errors))

	return mcp.NewToolResultText(string(jsonResponse)), nil
}

// HandleCreateBuildingBlock handles the create_building_block tool call
func (h *PromiseToolsHandler) HandleCreateBuildingBlock(arguments map[string]interface{}) (*mcp.CallToolResult, error) {
	h.logger.Info("Handling create_building_block tool call")

	// Extract building_block_name parameter (this is the Promise name)
	buildingBlockNameRaw, exists := arguments["building_block_name"]
	if !exists {
		return mcp.NewToolResultError("Missing required parameter 'building_block_name'"), nil
	}

	buildingBlockName, ok := buildingBlockNameRaw.(string)
	if !ok {
		return mcp.NewToolResultError("Parameter 'building_block_name' must be a string"), nil
	}

	// Extract resource_name parameter (name for the new Custom Resource instance)
	resourceNameRaw, exists := arguments["resource_name"]
	if !exists {
		return mcp.NewToolResultError("Missing required parameter 'resource_name'"), nil
	}

	resourceName, ok := resourceNameRaw.(string)
	if !ok {
		return mcp.NewToolResultError("Parameter 'resource_name' must be a string"), nil
	}

	// Extract spec parameter
	specRaw, exists := arguments["spec"]
	if !exists {
		return mcp.NewToolResultError("Missing required parameter 'spec'"), nil
	}

	specString, ok := specRaw.(string)
	if !ok {
		return mcp.NewToolResultError("Parameter 'spec' must be a JSON string"), nil
	}

	// Extract namespace parameter (required)
	namespaceRaw, exists := arguments["namespace"]
	if !exists {
		return mcp.NewToolResultError("Missing required parameter 'namespace'"), nil
	}

	namespace, ok := namespaceRaw.(string)
	if !ok {
		return mcp.NewToolResultError("Parameter 'namespace' must be a string"), nil
	}

	// Parse the JSON string into a map
	var spec map[string]interface{}
	if err := json.Unmarshal([]byte(specString), &spec); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid JSON in 'spec' parameter: %v", err)), nil
	}

	// Get the Promise to extract target resource info and schema
	promiseSchema, err := h.extractor.GetPromiseSchema(buildingBlockName)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get building block schema for '%s': %v", buildingBlockName, err)), nil
	}

	// Validate the spec against the Promise's schema
	validationDetails, err := h.validator.ValidateSpec(promiseSchema.OpenAPISchema, spec)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to validate spec: %v", err)), nil
	}

	if len(validationDetails.Errors) > 0 {
		// Return validation errors
		result := &resources.ValidationResult{
			Valid:            false,
			PromiseName:      buildingBlockName,
			ValidationResult: *validationDetails,
		}

		jsonResponse, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to format validation response: %v", err)), nil
		}

		return mcp.NewToolResultError(fmt.Sprintf("Validation failed for building block spec:\n%s", string(jsonResponse))), nil
	}

	// Construct the target GVR from the Promise's target resource info
	targetGVR := schema.GroupVersionResource{
		Group:    promiseSchema.TargetResource.Group,
		Version:  promiseSchema.TargetResource.Version,
		Resource: promiseSchema.TargetResource.Resource,
	}

	// Check if Custom Resource already exists in the namespace
	_, err = h.k8sClient.GetResource(targetGVR, namespace, resourceName)
	if err == nil {
		return mcp.NewToolResultError(fmt.Sprintf("Resource '%s' of type '%s' already exists in namespace '%s'", resourceName, promiseSchema.TargetResource.Kind, namespace)), nil
	}

	// Create the Custom Resource object
	apiVersion := fmt.Sprintf("%s/%s", promiseSchema.TargetResource.Group, promiseSchema.TargetResource.Version)
	customResource := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": apiVersion,
			"kind":       promiseSchema.TargetResource.Kind,
			"metadata": map[string]interface{}{
				"name":      resourceName,
				"namespace": namespace,
			},
			"spec": spec,
		},
	}

	// Create the resource in Kubernetes
	created, err := h.k8sClient.CreateResource(targetGVR, namespace, customResource)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to create resource '%s' of type '%s': %v", resourceName, promiseSchema.TargetResource.Kind, err)), nil
	}

	// Build success response
	response := map[string]interface{}{
		"success":        true,
		"resource_name":  resourceName,
		"resource_type":  promiseSchema.TargetResource.Kind,
		"building_block": buildingBlockName,
		"status":         "created",
		"metadata": map[string]interface{}{
			"created_at":       time.Now().Format(time.RFC3339),
			"cluster_context":  h.k8sClient.GetCurrentContext(),
			"resource_version": created.GetResourceVersion(),
		},
		"details": map[string]interface{}{
			"api_version": created.GetAPIVersion(),
			"kind":        created.GetKind(),
			"namespace":   created.GetNamespace(),
			"uid":         created.GetUID(),
		},
	}

	// Format as JSON
	jsonResponse, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to format response: %v", err)), nil
	}

	h.logger.Info("Successfully created Custom Resource",
		"resource_name", resourceName,
		"resource_type", promiseSchema.TargetResource.Kind,
		"building_block", buildingBlockName,
		"namespace", namespace,
		"uid", created.GetUID())

	return mcp.NewToolResultText(string(jsonResponse)), nil
}
