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
