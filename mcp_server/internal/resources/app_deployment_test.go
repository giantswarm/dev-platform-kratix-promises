package resources

import (
	"testing"

	"github.com/giantswarm/dev-platform-kratix-promises/mcp_server/internal/resources/mocks"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func TestHandleAppDeployments_Success(t *testing.T) {
	// Setup
	mockClient := new(mocks.MockKubernetesClient)
	handler := NewCRDResourceHandler(mockClient)

	// Load test data
	testAppDeployment, err := LoadAppDeploymentTestData()
	require.NoError(t, err, "Failed to load test data")

	// Create mock response
	mockList := CreateMockAppDeploymentList(testAppDeployment)

	// Setup mock expectations
	mockClient.On("ListResources", AppDeploymentGVR, "").Return(mockList, nil)
	mockClient.On("GetClusterInfo").Return(GetMockClusterInfo())

	// Create test request
	request := mcp.ReadResourceRequest{
		Params: struct {
			URI       string                 `json:"uri"`
			Arguments map[string]interface{} `json:"arguments,omitempty"`
		}{
			URI: "k8s://appdeployments",
		},
	}

	// Execute
	result, err := handler.HandleAppDeployments(request)

	// Assert
	assert.NoError(t, err)
	assert.Len(t, result, 1, "Expected one result item")

	// Verify the response structure
	response, ok := result[0].(*ResourceResponse)
	require.True(t, ok, "Expected ResourceResponse type")

	assert.Equal(t, "promise.platform.giantswarm.io/v1beta1", response.APIVersion)
	assert.Equal(t, "AppDeploymentList", response.Kind)
	assert.Len(t, response.Items, 1, "Expected one AppDeployment item")
	assert.Equal(t, 1, response.Metadata.Count)
	assert.Equal(t, "test-cluster-context", response.Metadata.ClusterInfo["context"])

	// Verify the actual AppDeployment data
	appDeployment := response.Items[0]
	assert.Equal(t, "mygodemo", appDeployment["metadata"].(map[string]interface{})["name"])
	assert.Equal(t, "org-demotech", appDeployment["metadata"].(map[string]interface{})["namespace"])

	// Verify mock expectations
	mockClient.AssertExpectations(t)
}

func TestHandleAppDeployments_Empty(t *testing.T) {
	// Setup
	mockClient := new(mocks.MockKubernetesClient)
	handler := NewCRDResourceHandler(mockClient)

	// Create empty mock response
	mockList := CreateEmptyAppDeploymentList()

	// Setup mock expectations
	mockClient.On("ListResources", AppDeploymentGVR, "").Return(mockList, nil)
	mockClient.On("GetClusterInfo").Return(GetMockClusterInfo())

	// Create test request
	request := mcp.ReadResourceRequest{
		Params: struct {
			URI       string                 `json:"uri"`
			Arguments map[string]interface{} `json:"arguments,omitempty"`
		}{
			URI: "k8s://appdeployments",
		},
	}

	// Execute
	result, err := handler.HandleAppDeployments(request)

	// Assert
	assert.NoError(t, err)
	assert.Len(t, result, 1, "Expected one result item")

	// Verify the response structure
	response, ok := result[0].(*ResourceResponse)
	require.True(t, ok, "Expected ResourceResponse type")

	assert.Equal(t, "promise.platform.giantswarm.io/v1beta1", response.APIVersion)
	assert.Equal(t, "AppDeploymentList", response.Kind)
	assert.Len(t, response.Items, 0, "Expected no AppDeployment items")
	assert.Equal(t, 0, response.Metadata.Count)

	// Verify mock expectations
	mockClient.AssertExpectations(t)
}

func TestHandleAppDeployments_Forbidden(t *testing.T) {
	// Setup
	mockClient := new(mocks.MockKubernetesClient)
	handler := NewCRDResourceHandler(mockClient)

	// Create forbidden error
	forbiddenError := apierrors.NewForbidden(
		schema.GroupResource{Group: "promise.platform.giantswarm.io", Resource: "appdeployments"},
		"",
		nil,
	)

	// Setup mock expectations
	mockClient.On("ListResources", AppDeploymentGVR, "").Return(nil, forbiddenError)

	// Create test request
	request := mcp.ReadResourceRequest{
		Params: struct {
			URI       string                 `json:"uri"`
			Arguments map[string]interface{} `json:"arguments,omitempty"`
		}{
			URI: "k8s://appdeployments",
		},
	}

	// Execute
	result, err := handler.HandleAppDeployments(request)

	// Assert
	assert.NoError(t, err, "Handler should not return error, but handle it gracefully")
	assert.Len(t, result, 1, "Expected one result item with error details")

	// Verify the error response
	errorResponse, ok := result[0].(map[string]interface{})
	require.True(t, ok, "Expected error response as map")

	assert.Equal(t, "Access denied to Kubernetes cluster", errorResponse["error"])
	assert.Equal(t, "Forbidden", errorResponse["reason"])
	assert.Equal(t, "AppDeployment", errorResponse["type"])

	// Verify mock expectations
	mockClient.AssertExpectations(t)
}

func TestHandleAppDeployments_Unauthorized(t *testing.T) {
	// Setup
	mockClient := new(mocks.MockKubernetesClient)
	handler := NewCRDResourceHandler(mockClient)

	// Create unauthorized error
	unauthorizedError := apierrors.NewUnauthorized("authentication required")

	// Setup mock expectations
	mockClient.On("ListResources", AppDeploymentGVR, "").Return(nil, unauthorizedError)

	// Create test request
	request := mcp.ReadResourceRequest{
		Params: struct {
			URI       string                 `json:"uri"`
			Arguments map[string]interface{} `json:"arguments,omitempty"`
		}{
			URI: "k8s://appdeployments",
		},
	}

	// Execute
	result, err := handler.HandleAppDeployments(request)

	// Assert
	assert.NoError(t, err, "Handler should not return error, but handle it gracefully")
	assert.Len(t, result, 1, "Expected one result item with error details")

	// Verify the error response
	errorResponse, ok := result[0].(map[string]interface{})
	require.True(t, ok, "Expected error response as map")

	assert.Equal(t, "Authentication required for Kubernetes cluster", errorResponse["error"])
	assert.Equal(t, "Unauthorized", errorResponse["reason"])
	assert.Equal(t, "AppDeployment", errorResponse["type"])

	// Verify mock expectations
	mockClient.AssertExpectations(t)
}

func TestHandleAppDeployments_NotFound(t *testing.T) {
	// Setup
	mockClient := new(mocks.MockKubernetesClient)
	handler := NewCRDResourceHandler(mockClient)

	// Create not found error
	notFoundError := apierrors.NewNotFound(
		schema.GroupResource{Group: "promise.platform.giantswarm.io", Resource: "appdeployments"},
		"",
	)

	// Setup mock expectations
	mockClient.On("ListResources", AppDeploymentGVR, "").Return(nil, notFoundError)

	// Create test request
	request := mcp.ReadResourceRequest{
		Params: struct {
			URI       string                 `json:"uri"`
			Arguments map[string]interface{} `json:"arguments,omitempty"`
		}{
			URI: "k8s://appdeployments",
		},
	}

	// Execute
	result, err := handler.HandleAppDeployments(request)

	// Assert
	assert.NoError(t, err, "Handler should not return error, but handle it gracefully")
	assert.Len(t, result, 1, "Expected one result item with error details")

	// Verify the error response
	errorResponse, ok := result[0].(map[string]interface{})
	require.True(t, ok, "Expected error response as map")

	assert.Equal(t, "No AppDeployment resources found", errorResponse["error"])
	assert.Equal(t, "NotFound", errorResponse["reason"])
	assert.Equal(t, "AppDeployment", errorResponse["type"])

	// Verify mock expectations
	mockClient.AssertExpectations(t)
}

func TestHandleAppDeployments_DataSanitization(t *testing.T) {
	// Setup
	mockClient := new(mocks.MockKubernetesClient)
	handler := NewCRDResourceHandler(mockClient)

	// Create an AppDeployment with sensitive data
	sensitiveAppDeployment := GetSensitiveAppDeployment()
	mockList := CreateMockAppDeploymentList(sensitiveAppDeployment)

	// Setup mock expectations
	mockClient.On("ListResources", AppDeploymentGVR, "").Return(mockList, nil)
	mockClient.On("GetClusterInfo").Return(GetMockClusterInfo())

	// Create test request
	request := mcp.ReadResourceRequest{
		Params: struct {
			URI       string                 `json:"uri"`
			Arguments map[string]interface{} `json:"arguments,omitempty"`
		}{
			URI: "k8s://appdeployments",
		},
	}

	// Execute
	result, err := handler.HandleAppDeployments(request)

	// Assert
	assert.NoError(t, err)
	assert.Len(t, result, 1, "Expected one result item")

	// Verify the response structure
	response, ok := result[0].(*ResourceResponse)
	require.True(t, ok, "Expected ResourceResponse type")

	assert.Len(t, response.Items, 1, "Expected one AppDeployment item")

	// Get the sanitized app deployment
	appDeployment := response.Items[0]
	metadata := appDeployment["metadata"].(map[string]interface{})
	spec := appDeployment["spec"].(map[string]interface{})

	// Verify sensitive metadata fields are removed
	assert.NotContains(t, metadata, "resourceVersion", "resourceVersion should be removed")
	assert.NotContains(t, metadata, "managedFields", "managedFields should be removed")
	assert.NotContains(t, metadata, "selfLink", "selfLink should be removed")

	// Verify secret references are preserved (not removed)
	if kubeConfig, found := spec["kubeConfig"].(map[string]interface{}); found {
		assert.Contains(t, kubeConfig, "secretRef", "kubeConfig.secretRef should be preserved")
	}

	// Verify non-sensitive data is preserved
	assert.Equal(t, "sensitive-app", metadata["name"])
	assert.Equal(t, "test-ns", metadata["namespace"])
	assert.Equal(t, "test-app.example.com", spec["ingressHost"])

	// Verify mock expectations
	mockClient.AssertExpectations(t)
}

func TestHandleAppDeployments_MultipleResources(t *testing.T) {
	// Setup
	mockClient := new(mocks.MockKubernetesClient)
	handler := NewCRDResourceHandler(mockClient)

	// Load test data and create multiple app deployments
	testAppDeployment1, err := LoadAppDeploymentTestData()
	require.NoError(t, err, "Failed to load test data")

	testAppDeployment2 := GetSensitiveAppDeployment()

	mockList := CreateMockAppDeploymentList(testAppDeployment1, testAppDeployment2)

	// Setup mock expectations
	mockClient.On("ListResources", AppDeploymentGVR, "").Return(mockList, nil)
	mockClient.On("GetClusterInfo").Return(GetMockClusterInfo())

	// Create test request
	request := mcp.ReadResourceRequest{
		Params: struct {
			URI       string                 `json:"uri"`
			Arguments map[string]interface{} `json:"arguments,omitempty"`
		}{
			URI: "k8s://appdeployments",
		},
	}

	// Execute
	result, err := handler.HandleAppDeployments(request)

	// Assert
	assert.NoError(t, err)
	assert.Len(t, result, 1, "Expected one result item")

	// Verify the response structure
	response, ok := result[0].(*ResourceResponse)
	require.True(t, ok, "Expected ResourceResponse type")

	assert.Equal(t, "promise.platform.giantswarm.io/v1beta1", response.APIVersion)
	assert.Equal(t, "AppDeploymentList", response.Kind)
	assert.Len(t, response.Items, 2, "Expected two AppDeployment items")
	assert.Equal(t, 2, response.Metadata.Count)

	// Verify both app deployments are present
	appNames := make([]string, 2)
	for i, item := range response.Items {
		metadata := item["metadata"].(map[string]interface{})
		appNames[i] = metadata["name"].(string)
	}

	assert.Contains(t, appNames, "mygodemo")
	assert.Contains(t, appNames, "sensitive-app")

	// Verify mock expectations
	mockClient.AssertExpectations(t)
}
