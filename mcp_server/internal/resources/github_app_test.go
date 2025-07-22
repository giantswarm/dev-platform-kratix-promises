package resources

import (
	"testing"

	"github.com/giantswarm/dev-platform-kratix-promises/mcp_server/internal/resources/mocks"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func TestHandleGitHubApps_Success(t *testing.T) {
	// Setup
	mockClient := new(mocks.MockKubernetesClient)
	handler := NewCRDResourceHandler(mockClient)

	// Load test data
	testGitHubApp, err := LoadGitHubAppTestData()
	require.NoError(t, err, "Failed to load test data")

	// Create mock response
	mockList := CreateMockGitHubAppList(testGitHubApp)
	
	// Setup mock expectations
	mockClient.On("ListResources", GitHubAppGVR, "").Return(mockList, nil)
	mockClient.On("GetClusterInfo").Return(GetMockClusterInfo())

	// Create test request
	request := mcp.ReadResourceRequest{
		Params: struct {
			URI string `json:"uri"`
			Arguments map[string]interface{} `json:"arguments,omitempty"`
		}{
			URI: "k8s://githubapps",
		},
	}

	// Execute
	result, err := handler.HandleGitHubApps(request)

	// Assert
	assert.NoError(t, err)
	assert.Len(t, result, 1, "Expected one result item")

	// Verify the response structure
	response, ok := result[0].(*ResourceResponse)
	require.True(t, ok, "Expected ResourceResponse type")
	
	assert.Equal(t, "promise.platform.giantswarm.io/v1beta1", response.APIVersion)
	assert.Equal(t, "GitHubAppList", response.Kind)
	assert.Len(t, response.Items, 1, "Expected one GitHubApp item")
	assert.Equal(t, 1, response.Metadata.Count)
	assert.Equal(t, "test-cluster-context", response.Metadata.ClusterInfo["context"])

	// Verify the actual GitHubApp data
	githubApp := response.Items[0]
	assert.Equal(t, "mygodemo", githubApp["metadata"].(map[string]interface{})["name"])
	assert.Equal(t, "org-demotech", githubApp["metadata"].(map[string]interface{})["namespace"])

	// Verify mock expectations
	mockClient.AssertExpectations(t)
}

func TestHandleGitHubApps_Empty(t *testing.T) {
	// Setup
	mockClient := new(mocks.MockKubernetesClient)
	handler := NewCRDResourceHandler(mockClient)

	// Create empty mock response
	mockList := CreateEmptyGitHubAppList()
	
	// Setup mock expectations
	mockClient.On("ListResources", GitHubAppGVR, "").Return(mockList, nil)
	mockClient.On("GetClusterInfo").Return(GetMockClusterInfo())

	// Create test request
	request := mcp.ReadResourceRequest{
		Params: struct {
			URI string `json:"uri"`
			Arguments map[string]interface{} `json:"arguments,omitempty"`
		}{
			URI: "k8s://githubapps",
		},
	}

	// Execute
	result, err := handler.HandleGitHubApps(request)

	// Assert
	assert.NoError(t, err)
	assert.Len(t, result, 1, "Expected one result item")

	// Verify the response structure
	response, ok := result[0].(*ResourceResponse)
	require.True(t, ok, "Expected ResourceResponse type")
	
	assert.Equal(t, "promise.platform.giantswarm.io/v1beta1", response.APIVersion)
	assert.Equal(t, "GitHubAppList", response.Kind)
	assert.Len(t, response.Items, 0, "Expected no GitHubApp items")
	assert.Equal(t, 0, response.Metadata.Count)

	// Verify mock expectations
	mockClient.AssertExpectations(t)
}

func TestHandleGitHubApps_Forbidden(t *testing.T) {
	// Setup
	mockClient := new(mocks.MockKubernetesClient)
	handler := NewCRDResourceHandler(mockClient)

	// Create forbidden error
	forbiddenError := apierrors.NewForbidden(
		schema.GroupResource{Group: "promise.platform.giantswarm.io", Resource: "githubapps"},
		"",
		nil,
	)
	
	// Setup mock expectations
	mockClient.On("ListResources", GitHubAppGVR, "").Return(nil, forbiddenError)

	// Create test request
	request := mcp.ReadResourceRequest{
		Params: struct {
			URI string `json:"uri"`
			Arguments map[string]interface{} `json:"arguments,omitempty"`
		}{
			URI: "k8s://githubapps",
		},
	}

	// Execute
	result, err := handler.HandleGitHubApps(request)

	// Assert
	assert.NoError(t, err, "Handler should not return error, but handle it gracefully")
	assert.Len(t, result, 1, "Expected one result item with error details")

	// Verify the error response
	errorResponse, ok := result[0].(map[string]interface{})
	require.True(t, ok, "Expected error response as map")
	
	assert.Equal(t, "Access denied to Kubernetes cluster", errorResponse["error"])
	assert.Equal(t, "Forbidden", errorResponse["reason"])
	assert.Equal(t, "GitHubApp", errorResponse["type"])

	// Verify mock expectations
	mockClient.AssertExpectations(t)
}

func TestHandleGitHubApps_Unauthorized(t *testing.T) {
	// Setup
	mockClient := new(mocks.MockKubernetesClient)
	handler := NewCRDResourceHandler(mockClient)

	// Create unauthorized error
	unauthorizedError := apierrors.NewUnauthorized("authentication required")
	
	// Setup mock expectations
	mockClient.On("ListResources", GitHubAppGVR, "").Return(nil, unauthorizedError)

	// Create test request
	request := mcp.ReadResourceRequest{
		Params: struct {
			URI string `json:"uri"`
			Arguments map[string]interface{} `json:"arguments,omitempty"`
		}{
			URI: "k8s://githubapps",
		},
	}

	// Execute
	result, err := handler.HandleGitHubApps(request)

	// Assert
	assert.NoError(t, err, "Handler should not return error, but handle it gracefully")
	assert.Len(t, result, 1, "Expected one result item with error details")

	// Verify the error response
	errorResponse, ok := result[0].(map[string]interface{})
	require.True(t, ok, "Expected error response as map")
	
	assert.Equal(t, "Authentication required for Kubernetes cluster", errorResponse["error"])
	assert.Equal(t, "Unauthorized", errorResponse["reason"])
	assert.Equal(t, "GitHubApp", errorResponse["type"])

	// Verify mock expectations
	mockClient.AssertExpectations(t)
}

func TestHandleGitHubApps_NotFound(t *testing.T) {
	// Setup
	mockClient := new(mocks.MockKubernetesClient)
	handler := NewCRDResourceHandler(mockClient)

	// Create not found error
	notFoundError := apierrors.NewNotFound(
		schema.GroupResource{Group: "promise.platform.giantswarm.io", Resource: "githubapps"},
		"",
	)
	
	// Setup mock expectations
	mockClient.On("ListResources", GitHubAppGVR, "").Return(nil, notFoundError)

	// Create test request
	request := mcp.ReadResourceRequest{
		Params: struct {
			URI string `json:"uri"`
			Arguments map[string]interface{} `json:"arguments,omitempty"`
		}{
			URI: "k8s://githubapps",
		},
	}

	// Execute
	result, err := handler.HandleGitHubApps(request)

	// Assert
	assert.NoError(t, err, "Handler should not return error, but handle it gracefully")
	assert.Len(t, result, 1, "Expected one result item with error details")

	// Verify the error response
	errorResponse, ok := result[0].(map[string]interface{})
	require.True(t, ok, "Expected error response as map")
	
	assert.Equal(t, "No GitHubApp resources found", errorResponse["error"])
	assert.Equal(t, "NotFound", errorResponse["reason"])
	assert.Equal(t, "GitHubApp", errorResponse["type"])

	// Verify mock expectations
	mockClient.AssertExpectations(t)
}

func TestHandleGitHubApps_DataSanitization(t *testing.T) {
	// Setup
	mockClient := new(mocks.MockKubernetesClient)
	handler := NewCRDResourceHandler(mockClient)

	// Create a GitHubApp with sensitive data
	sensitiveGitHubApp := GetSensitiveGitHubApp()
	mockList := CreateMockGitHubAppList(sensitiveGitHubApp)
	
	// Setup mock expectations
	mockClient.On("ListResources", GitHubAppGVR, "").Return(mockList, nil)
	mockClient.On("GetClusterInfo").Return(GetMockClusterInfo())

	// Create test request
	request := mcp.ReadResourceRequest{
		Params: struct {
			URI string `json:"uri"`
			Arguments map[string]interface{} `json:"arguments,omitempty"`
		}{
			URI: "k8s://githubapps",
		},
	}

	// Execute
	result, err := handler.HandleGitHubApps(request)

	// Assert
	assert.NoError(t, err)
	assert.Len(t, result, 1, "Expected one result item")

	// Verify the response structure
	response, ok := result[0].(*ResourceResponse)
	require.True(t, ok, "Expected ResourceResponse type")
	
	assert.Len(t, response.Items, 1, "Expected one GitHubApp item")
	
	// Get the sanitized GitHub app
	githubApp := response.Items[0]
	metadata := githubApp["metadata"].(map[string]interface{})
	spec := githubApp["spec"].(map[string]interface{})

	// Verify sensitive metadata fields are removed
	assert.NotContains(t, metadata, "resourceVersion", "resourceVersion should be removed")
	assert.NotContains(t, metadata, "managedFields", "managedFields should be removed")
	assert.NotContains(t, metadata, "selfLink", "selfLink should be removed")

	// Verify sensitive nested spec fields are removed
	if githubRepo, found := spec["githubRepo"].(map[string]interface{}); found {
		if repoSpec, found := githubRepo["spec"].(map[string]interface{}); found {
			assert.NotContains(t, repoSpec, "githubTokenSecretRef", "githubRepo.spec.githubTokenSecretRef should be removed")
			assert.NotContains(t, repoSpec, "registryInfoConfigMapRef", "githubRepo.spec.registryInfoConfigMapRef should be removed")
		}
	}

	if appDeployment, found := spec["appDeployment"].(map[string]interface{}); found {
		if deploySpec, found := appDeployment["spec"].(map[string]interface{}); found {
			if kubeConfig, found := deploySpec["kubeConfig"].(map[string]interface{}); found {
				assert.NotContains(t, kubeConfig, "secretRef", "appDeployment.spec.kubeConfig.secretRef should be removed")
			}
		}
	}

	// Verify non-sensitive data is preserved
	assert.Equal(t, "sensitive-gh-app", metadata["name"])
	assert.Equal(t, "test-ns", metadata["namespace"])

	// Verify mock expectations
	mockClient.AssertExpectations(t)
}

func TestHandleGitHubApps_MultipleResources(t *testing.T) {
	// Setup
	mockClient := new(mocks.MockKubernetesClient)
	handler := NewCRDResourceHandler(mockClient)

	// Load test data and create multiple GitHub apps
	testGitHubApp1, err := LoadGitHubAppTestData()
	require.NoError(t, err, "Failed to load test data")
	
	testGitHubApp2 := GetSensitiveGitHubApp()
	
	mockList := CreateMockGitHubAppList(testGitHubApp1, testGitHubApp2)
	
	// Setup mock expectations
	mockClient.On("ListResources", GitHubAppGVR, "").Return(mockList, nil)
	mockClient.On("GetClusterInfo").Return(GetMockClusterInfo())

	// Create test request
	request := mcp.ReadResourceRequest{
		Params: struct {
			URI string `json:"uri"`
			Arguments map[string]interface{} `json:"arguments,omitempty"`
		}{
			URI: "k8s://githubapps",
		},
	}

	// Execute
	result, err := handler.HandleGitHubApps(request)

	// Assert
	assert.NoError(t, err)
	assert.Len(t, result, 1, "Expected one result item")

	// Verify the response structure
	response, ok := result[0].(*ResourceResponse)
	require.True(t, ok, "Expected ResourceResponse type")
	
	assert.Equal(t, "promise.platform.giantswarm.io/v1beta1", response.APIVersion)
	assert.Equal(t, "GitHubAppList", response.Kind)
	assert.Len(t, response.Items, 2, "Expected two GitHubApp items")
	assert.Equal(t, 2, response.Metadata.Count)

	// Verify both GitHub apps are present
	appNames := make([]string, 2)
	for i, item := range response.Items {
		metadata := item["metadata"].(map[string]interface{})
		appNames[i] = metadata["name"].(string)
	}
	
	assert.Contains(t, appNames, "mygodemo")
	assert.Contains(t, appNames, "sensitive-gh-app")

	// Verify mock expectations
	mockClient.AssertExpectations(t)
}

func TestHandleGitHubApps_RealDataStructure(t *testing.T) {
	// Load real test data to verify its structure
	testGitHubApp, err := LoadGitHubAppTestData()
	require.NoError(t, err, "Failed to load real GitHubApp test data")

	// Verify the loaded data has the expected structure
	assert.Equal(t, "promise.platform.giantswarm.io/v1beta1", testGitHubApp.GetAPIVersion())
	assert.Equal(t, "githubapp", testGitHubApp.GetKind())
	assert.Equal(t, "mygodemo", testGitHubApp.GetName())
	assert.Equal(t, "org-demotech", testGitHubApp.GetNamespace())

	// Verify the spec contains expected nested fields
	spec, found, err := unstructured.NestedMap(testGitHubApp.Object, "spec")
	require.NoError(t, err)
	require.True(t, found, "spec should be present")
	
	// Check for nested appDeployment and githubRepo specs
	assert.Contains(t, spec, "appDeployment", "spec should contain appDeployment")
	assert.Contains(t, spec, "githubRepo", "spec should contain githubRepo")

	// Verify nested appDeployment structure
	appDeployment, found := spec["appDeployment"].(map[string]interface{})
	require.True(t, found, "appDeployment should be present")
	assert.Contains(t, appDeployment, "name", "appDeployment should have name")
	assert.Contains(t, appDeployment, "spec", "appDeployment should have spec")

	// Verify nested githubRepo structure
	githubRepo, found := spec["githubRepo"].(map[string]interface{})
	require.True(t, found, "githubRepo should be present")
	assert.Contains(t, githubRepo, "name", "githubRepo should have name")
	assert.Contains(t, githubRepo, "spec", "githubRepo should have spec")
} 