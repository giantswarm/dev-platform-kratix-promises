package tools

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/giantswarm/dev-platform-kratix-promises/mcp_server/internal/resources"
	"github.com/giantswarm/dev-platform-kratix-promises/mcp_server/internal/resources/mocks"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPromiseToolsHandler_HandleListKratixPromises_Success(t *testing.T) {
	// Setup
	mockClient := new(mocks.MockKubernetesClient)
	handler := NewPromiseToolsHandler(mockClient)

	// Load test data for all three promises
	appDeploymentData, err := LoadPromiseTestData("appdeployment")
	require.NoError(t, err, "Failed to load appdeployment test data")

	githubAppData, err := LoadPromiseTestData("githubapp")
	require.NoError(t, err, "Failed to load githubapp test data")

	gitHubRepoData, err := LoadPromiseTestData("githubrepo")
	require.NoError(t, err, "Failed to load githubrepo test data")

	// Create mock response with all three promises
	mockList := CreateMockPromiseList(
		appDeploymentData.Promise,
		githubAppData.Promise,
		gitHubRepoData.Promise,
	)

	// Setup mock expectations
	mockClient.On("ListResources", resources.KratixPromiseGVR, "").Return(mockList, nil)
	mockClient.On("GetClusterInfo").Return(GetMockClusterInfo())
	mockClient.On("GetCurrentContext").Return("test-cluster-context")

	// Execute
	arguments := map[string]interface{}{}
	result, err := handler.HandleListKratixPromises(arguments)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.False(t, result.IsError)
	assert.NotEmpty(t, result.Content)

	// Get the text content
	textContent, ok := result.Content[0].(mcp.TextContent)
	require.True(t, ok, "Expected TextContent")
	text := textContent.Text

	// Parse response JSON
	var response map[string]interface{}
	err = json.Unmarshal([]byte(text), &response)
	require.NoError(t, err)

	// Verify response structure
	assert.Contains(t, response, "building_blocks")
	assert.Contains(t, response, "metadata")

	buildingBlocks, ok := response["building_blocks"].([]interface{})
	require.True(t, ok, "building_blocks should be an array")
	assert.Len(t, buildingBlocks, 3, "Expected 3 building_blocks")

	metadata, ok := response["metadata"].(map[string]interface{})
	require.True(t, ok, "metadata should be an object")
	assert.Equal(t, float64(3), metadata["total_count"])
	assert.Equal(t, "test-cluster-context", metadata["cluster_context"])

	// Verify promise names
	promiseNames := make([]string, len(buildingBlocks))
	for i, p := range buildingBlocks {
		promise, ok := p.(map[string]interface{})
		require.True(t, ok)
		promiseNames[i] = promise["name"].(string)
	}
	assert.Contains(t, promiseNames, "appdeployment")
	assert.Contains(t, promiseNames, "githubapp")
	assert.Contains(t, promiseNames, "githubrepo")

	// Verify mock expectations
	mockClient.AssertExpectations(t)
}

func TestPromiseToolsHandler_HandleGetPromiseSchema_Success(t *testing.T) {
	// Setup
	mockClient := new(mocks.MockKubernetesClient)
	handler := NewPromiseToolsHandler(mockClient)

	// Load test data
	testData, err := LoadPromiseTestData("appdeployment")
	require.NoError(t, err, "Failed to load test data")

	// Setup mock expectations - GetPromiseSchema calls GetResource directly
	mockClient.On("GetResource", resources.KratixPromiseGVR, "", "appdeployment").Return(testData.Promise, nil)

	// Execute
	arguments := map[string]interface{}{
		"building_block_name": "appdeployment",
	}
	result, err := handler.HandleGetPromiseSchema(arguments)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.False(t, result.IsError)

	// Get the text content
	textContent, ok := result.Content[0].(mcp.TextContent)
	require.True(t, ok, "Expected TextContent")
	text := textContent.Text

	// Parse response JSON
	var schema resources.PromiseSchema
	err = json.Unmarshal([]byte(text), &schema)
	require.NoError(t, err)

	// Verify schema structure
	assert.Equal(t, "appdeployment", schema.PromiseName)
	assert.NotNil(t, schema.OpenAPISchema)
	assert.Contains(t, schema.OpenAPISchema, "properties")

	// Navigate to the spec properties for verification
	rootProperties, ok := schema.OpenAPISchema["properties"].(map[string]interface{})
	require.True(t, ok, "Expected root properties")
	specSchema, ok := rootProperties["spec"].(map[string]interface{})
	require.True(t, ok, "Expected spec schema")
	properties, ok := specSchema["properties"].(map[string]interface{})
	require.True(t, ok, "Expected properties in spec schema")
	assert.Contains(t, properties, "database")
	assert.Contains(t, properties, "ingressHost")
	assert.Contains(t, properties, "values")

	// Verify mock expectations
	mockClient.AssertExpectations(t)
}

func TestPromiseToolsHandler_HandleGetPromiseSchema_NotFound(t *testing.T) {
	// Setup
	mockClient := new(mocks.MockKubernetesClient)
	handler := NewPromiseToolsHandler(mockClient)

	// Setup mock expectations - return error for non-existent promise
	mockClient.On("GetResource", resources.KratixPromiseGVR, "", "nonexistent").Return(nil, assert.AnError)

	// Execute
	arguments := map[string]interface{}{
		"building_block_name": "nonexistent",
	}
	result, err := handler.HandleGetPromiseSchema(arguments)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.IsError)

	// Get the text content (error message)
	textContent, ok := result.Content[0].(mcp.TextContent)
	require.True(t, ok, "Expected TextContent")
	text := textContent.Text
	assert.Contains(t, text, "Failed to get building block schema for 'nonexistent'")

	// Verify mock expectations
	mockClient.AssertExpectations(t)
}

func TestPromiseToolsHandler_HandleValidatePromiseSpec_Valid(t *testing.T) {
	// Setup
	mockClient := new(mocks.MockKubernetesClient)
	handler := NewPromiseToolsHandler(mockClient)

	// Load test data
	testData, err := LoadPromiseTestData("appdeployment")
	require.NoError(t, err, "Failed to load test data")

	// Setup mock expectations
	mockClient.On("GetResource", resources.KratixPromiseGVR, "", "appdeployment").Return(testData.Promise, nil)

	// Convert valid spec to JSON string - wrap in spec field to match schema
	wrappedValidSpec := map[string]interface{}{
		"spec": testData.ValidSpec,
	}
	validSpecJSON, err := json.Marshal(wrappedValidSpec)
	require.NoError(t, err)

	// Execute
	arguments := map[string]interface{}{
		"building_block_name": "appdeployment",
		"spec":                string(validSpecJSON),
	}
	result, err := handler.HandleValidatePromiseSpec(arguments)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.False(t, result.IsError)

	// Get the text content
	textContent, ok := result.Content[0].(mcp.TextContent)
	require.True(t, ok, "Expected TextContent")
	text := textContent.Text

	// Parse response JSON
	var validationResult resources.ValidationResult
	err = json.Unmarshal([]byte(text), &validationResult)
	require.NoError(t, err)

	// Verify validation result
	assert.True(t, validationResult.Valid, "Valid specification should pass validation")
	assert.Equal(t, "appdeployment", validationResult.PromiseName)
	assert.Empty(t, validationResult.ValidationResult.Errors, "Valid specification should have no errors")

	// Verify mock expectations
	mockClient.AssertExpectations(t)
}

func TestPromiseToolsHandler_HandleValidatePromiseSpec_Invalid(t *testing.T) {
	// Setup
	mockClient := new(mocks.MockKubernetesClient)
	handler := NewPromiseToolsHandler(mockClient)

	// Load test data
	testData, err := LoadPromiseTestData("appdeployment")
	require.NoError(t, err, "Failed to load test data")

	// Setup mock expectations
	mockClient.On("GetResource", resources.KratixPromiseGVR, "", "appdeployment").Return(testData.Promise, nil)

	// Convert invalid spec to JSON string - wrap in spec field to match schema
	wrappedInvalidSpec := map[string]interface{}{
		"spec": testData.InvalidSpec,
	}
	invalidSpecJSON, err := json.Marshal(wrappedInvalidSpec)
	require.NoError(t, err)

	// Execute
	arguments := map[string]interface{}{
		"building_block_name": "appdeployment",
		"spec":                string(invalidSpecJSON),
	}
	result, err := handler.HandleValidatePromiseSpec(arguments)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.False(t, result.IsError)

	// Get the text content
	textContent, ok := result.Content[0].(mcp.TextContent)
	require.True(t, ok, "Expected TextContent")
	text := textContent.Text

	// Parse response JSON
	var validationResult resources.ValidationResult
	err = json.Unmarshal([]byte(text), &validationResult)
	require.NoError(t, err)

	// Verify validation result
	assert.False(t, validationResult.Valid, "Invalid specification should fail validation")
	assert.Equal(t, "appdeployment", validationResult.PromiseName)
	assert.NotEmpty(t, validationResult.ValidationResult.Errors, "Invalid specification should have errors")

	// Verify that specific validation errors are present
	errorMessages := make([]string, len(validationResult.ValidationResult.Errors))
	for i, err := range validationResult.ValidationResult.Errors {
		errorMessages[i] = err.Message
	}

	// We expect validation errors related to missing required fields and invalid patterns
	found := false
	for _, msg := range errorMessages {
		if strings.Contains(msg, "required") || strings.Contains(msg, "pattern") || strings.Contains(msg, "missing") {
			found = true
			break
		}
	}
	assert.True(t, found, "Expected validation errors related to missing required fields or invalid patterns")

	// Verify mock expectations
	mockClient.AssertExpectations(t)
}

func TestPromiseToolsHandler_HandleValidatePromiseSpec_MissingParameters(t *testing.T) {
	// Setup
	mockClient := new(mocks.MockKubernetesClient)
	handler := NewPromiseToolsHandler(mockClient)

	tests := []struct {
		name      string
		arguments map[string]interface{}
		errorMsg  string
	}{
		{
			name:      "missing building_block_name",
			arguments: map[string]interface{}{"spec": `{"test": "value"}`},
			errorMsg:  "Missing required parameter 'building_block_name'",
		},
		{
			name:      "missing spec",
			arguments: map[string]interface{}{"building_block_name": "appdeployment"},
			errorMsg:  "Missing required parameter 'spec'",
		},
		{
			name:      "invalid JSON spec",
			arguments: map[string]interface{}{"building_block_name": "appdeployment", "spec": `{invalid json`},
			errorMsg:  "Invalid JSON in 'spec' parameter",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Execute
			result, err := handler.HandleValidatePromiseSpec(tt.arguments)

			// Assert
			assert.NoError(t, err)
			assert.NotNil(t, result)
			assert.True(t, result.IsError)

			// Get the error message
			textContent, ok := result.Content[0].(mcp.TextContent)
			require.True(t, ok, "Expected TextContent")
			text := textContent.Text
			assert.Contains(t, text, tt.errorMsg)
		})
	}
}
