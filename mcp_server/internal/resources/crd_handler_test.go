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

func TestHandleGitHubRepos_Success(t *testing.T) {
	// Setup
	mockClient := new(mocks.MockKubernetesClient)
	handler := NewCRDResourceHandler(mockClient)

	// Load test data
	testRepo, err := LoadGitHubRepoTestData()
	require.NoError(t, err, "Failed to load test data")

	// Create mock response
	mockList := CreateMockGitHubRepoList(testRepo)
	
	// Setup mock expectations
	mockClient.On("ListResources", GitHubRepoGVR, "").Return(mockList, nil)
	mockClient.On("GetClusterInfo").Return(GetMockClusterInfo())

	// Create test request
	request := mcp.ReadResourceRequest{
		Params: struct {
			URI string `json:"uri"`
			Arguments map[string]interface{} `json:"arguments,omitempty"`
		}{
			URI: "k8s://githubrepos",
		},
	}

	// Execute
	result, err := handler.HandleGitHubRepos(request)

	// Assert
	assert.NoError(t, err)
	assert.Len(t, result, 1, "Expected one result item")

	// Verify the response structure
	response, ok := result[0].(*ResourceResponse)
	require.True(t, ok, "Expected ResourceResponse type")
	
	assert.Equal(t, "promise.platform.giantswarm.io/v1beta1", response.APIVersion)
	assert.Equal(t, "GitHubRepoList", response.Kind)
	assert.Len(t, response.Items, 1, "Expected one GitHubRepo item")
	assert.Equal(t, 1, response.Metadata.Count)
	assert.Equal(t, "test-cluster-context", response.Metadata.ClusterInfo["context"])

	// Verify the actual GitHubRepo data
	repo := response.Items[0]
	assert.Equal(t, "mygodemo", repo["metadata"].(map[string]interface{})["name"])
	assert.Equal(t, "org-demotech", repo["metadata"].(map[string]interface{})["namespace"])

	// Verify mock expectations
	mockClient.AssertExpectations(t)
}

func TestHandleGitHubRepos_Empty(t *testing.T) {
	// Setup
	mockClient := new(mocks.MockKubernetesClient)
	handler := NewCRDResourceHandler(mockClient)

	// Create empty mock response
	mockList := CreateEmptyGitHubRepoList()
	
	// Setup mock expectations
	mockClient.On("ListResources", GitHubRepoGVR, "").Return(mockList, nil)
	mockClient.On("GetClusterInfo").Return(GetMockClusterInfo())

	// Create test request
	request := mcp.ReadResourceRequest{
		Params: struct {
			URI string `json:"uri"`
			Arguments map[string]interface{} `json:"arguments,omitempty"`
		}{
			URI: "k8s://githubrepos",
		},
	}

	// Execute
	result, err := handler.HandleGitHubRepos(request)

	// Assert
	assert.NoError(t, err)
	assert.Len(t, result, 1, "Expected one result item")

	// Verify the response structure
	response, ok := result[0].(*ResourceResponse)
	require.True(t, ok, "Expected ResourceResponse type")
	
	assert.Equal(t, "promise.platform.giantswarm.io/v1beta1", response.APIVersion)
	assert.Equal(t, "GitHubRepoList", response.Kind)
	assert.Len(t, response.Items, 0, "Expected no GitHubRepo items")
	assert.Equal(t, 0, response.Metadata.Count)

	// Verify mock expectations
	mockClient.AssertExpectations(t)
}

func TestHandleGitHubRepos_Forbidden(t *testing.T) {
	// Setup
	mockClient := new(mocks.MockKubernetesClient)
	handler := NewCRDResourceHandler(mockClient)

	// Create forbidden error
	forbiddenError := apierrors.NewForbidden(
		schema.GroupResource{Group: "promise.platform.giantswarm.io", Resource: "githubrepos"},
		"",
		nil,
	)
	
	// Setup mock expectations
	mockClient.On("ListResources", GitHubRepoGVR, "").Return(nil, forbiddenError)

	// Create test request
	request := mcp.ReadResourceRequest{
		Params: struct {
			URI string `json:"uri"`
			Arguments map[string]interface{} `json:"arguments,omitempty"`
		}{
			URI: "k8s://githubrepos",
		},
	}

	// Execute
	result, err := handler.HandleGitHubRepos(request)

	// Assert
	assert.NoError(t, err, "Handler should not return error, but handle it gracefully")
	assert.Len(t, result, 1, "Expected one result item with error details")

	// Verify the error response
	errorResponse, ok := result[0].(map[string]interface{})
	require.True(t, ok, "Expected error response as map")
	
	assert.Equal(t, "Access denied to Kubernetes cluster", errorResponse["error"])
	assert.Equal(t, "Forbidden", errorResponse["reason"])
	assert.Equal(t, "GitHubRepo", errorResponse["type"])

	// Verify mock expectations
	mockClient.AssertExpectations(t)
}

func TestHandleGitHubRepos_Unauthorized(t *testing.T) {
	// Setup
	mockClient := new(mocks.MockKubernetesClient)
	handler := NewCRDResourceHandler(mockClient)

	// Create unauthorized error
	unauthorizedError := apierrors.NewUnauthorized("authentication required")
	
	// Setup mock expectations
	mockClient.On("ListResources", GitHubRepoGVR, "").Return(nil, unauthorizedError)

	// Create test request
	request := mcp.ReadResourceRequest{
		Params: struct {
			URI string `json:"uri"`
			Arguments map[string]interface{} `json:"arguments,omitempty"`
		}{
			URI: "k8s://githubrepos",
		},
	}

	// Execute
	result, err := handler.HandleGitHubRepos(request)

	// Assert
	assert.NoError(t, err, "Handler should not return error, but handle it gracefully")
	assert.Len(t, result, 1, "Expected one result item with error details")

	// Verify the error response
	errorResponse, ok := result[0].(map[string]interface{})
	require.True(t, ok, "Expected error response as map")
	
	assert.Equal(t, "Authentication required for Kubernetes cluster", errorResponse["error"])
	assert.Equal(t, "Unauthorized", errorResponse["reason"])
	assert.Equal(t, "GitHubRepo", errorResponse["type"])

	// Verify mock expectations
	mockClient.AssertExpectations(t)
}

func TestHandleGitHubRepos_NotFound(t *testing.T) {
	// Setup
	mockClient := new(mocks.MockKubernetesClient)
	handler := NewCRDResourceHandler(mockClient)

	// Create not found error
	notFoundError := apierrors.NewNotFound(
		schema.GroupResource{Group: "promise.platform.giantswarm.io", Resource: "githubrepos"},
		"",
	)
	
	// Setup mock expectations
	mockClient.On("ListResources", GitHubRepoGVR, "").Return(nil, notFoundError)

	// Create test request
	request := mcp.ReadResourceRequest{
		Params: struct {
			URI string `json:"uri"`
			Arguments map[string]interface{} `json:"arguments,omitempty"`
		}{
			URI: "k8s://githubrepos",
		},
	}

	// Execute
	result, err := handler.HandleGitHubRepos(request)

	// Assert
	assert.NoError(t, err, "Handler should not return error, but handle it gracefully")
	assert.Len(t, result, 1, "Expected one result item with error details")

	// Verify the error response
	errorResponse, ok := result[0].(map[string]interface{})
	require.True(t, ok, "Expected error response as map")
	
	assert.Equal(t, "No GitHubRepo resources found", errorResponse["error"])
	assert.Equal(t, "NotFound", errorResponse["reason"])
	assert.Equal(t, "GitHubRepo", errorResponse["type"])

	// Verify mock expectations
	mockClient.AssertExpectations(t)
}

func TestHandleGitHubRepos_DataSanitization(t *testing.T) {
	// Setup
	mockClient := new(mocks.MockKubernetesClient)
	handler := NewCRDResourceHandler(mockClient)

	// Create a GitHubRepo with sensitive data
	sensitiveRepo := GetSensitiveGitHubRepo()
	mockList := CreateMockGitHubRepoList(sensitiveRepo)
	
	// Setup mock expectations
	mockClient.On("ListResources", GitHubRepoGVR, "").Return(mockList, nil)
	mockClient.On("GetClusterInfo").Return(GetMockClusterInfo())

	// Create test request
	request := mcp.ReadResourceRequest{
		Params: struct {
			URI string `json:"uri"`
			Arguments map[string]interface{} `json:"arguments,omitempty"`
		}{
			URI: "k8s://githubrepos",
		},
	}

	// Execute
	result, err := handler.HandleGitHubRepos(request)

	// Assert
	assert.NoError(t, err)
	assert.Len(t, result, 1, "Expected one result item")

	// Verify the response structure
	response, ok := result[0].(*ResourceResponse)
	require.True(t, ok, "Expected ResourceResponse type")
	
	assert.Len(t, response.Items, 1, "Expected one GitHubRepo item")
	
	// Get the sanitized repo
	repo := response.Items[0]
	metadata := repo["metadata"].(map[string]interface{})
	spec := repo["spec"].(map[string]interface{})

	// Verify sensitive metadata fields are removed
	assert.NotContains(t, metadata, "resourceVersion", "resourceVersion should be removed")
	assert.NotContains(t, metadata, "managedFields", "managedFields should be removed")
	assert.NotContains(t, metadata, "selfLink", "selfLink should be removed")

	// Verify sensitive spec fields are removed
	assert.NotContains(t, spec, "githubTokenSecretRef", "githubTokenSecretRef should be removed")
	assert.NotContains(t, spec, "registryInfoConfigMapRef", "registryInfoConfigMapRef should be removed")

	// Verify non-sensitive data is preserved
	assert.Equal(t, "sensitive-repo", metadata["name"])
	assert.Equal(t, "test-ns", metadata["namespace"])
	
	repository := spec["repository"].(map[string]interface{})
	assert.Equal(t, "test-repo", repository["name"])
	assert.Equal(t, "test-owner", repository["owner"])

	// Verify mock expectations
	mockClient.AssertExpectations(t)
}

func TestHandleGitHubRepos_MultipleResources(t *testing.T) {
	// Setup
	mockClient := new(mocks.MockKubernetesClient)
	handler := NewCRDResourceHandler(mockClient)

	// Load test data and create multiple repos
	testRepo1, err := LoadGitHubRepoTestData()
	require.NoError(t, err, "Failed to load test data")
	
	testRepo2 := GetSensitiveGitHubRepo()
	
	mockList := CreateMockGitHubRepoList(testRepo1, testRepo2)
	
	// Setup mock expectations
	mockClient.On("ListResources", GitHubRepoGVR, "").Return(mockList, nil)
	mockClient.On("GetClusterInfo").Return(GetMockClusterInfo())

	// Create test request
	request := mcp.ReadResourceRequest{
		Params: struct {
			URI string `json:"uri"`
			Arguments map[string]interface{} `json:"arguments,omitempty"`
		}{
			URI: "k8s://githubrepos",
		},
	}

	// Execute
	result, err := handler.HandleGitHubRepos(request)

	// Assert
	assert.NoError(t, err)
	assert.Len(t, result, 1, "Expected one result item")

	// Verify the response structure
	response, ok := result[0].(*ResourceResponse)
	require.True(t, ok, "Expected ResourceResponse type")
	
	assert.Equal(t, "promise.platform.giantswarm.io/v1beta1", response.APIVersion)
	assert.Equal(t, "GitHubRepoList", response.Kind)
	assert.Len(t, response.Items, 2, "Expected two GitHubRepo items")
	assert.Equal(t, 2, response.Metadata.Count)

	// Verify both repos are present
	repoNames := make([]string, 2)
	for i, item := range response.Items {
		metadata := item["metadata"].(map[string]interface{})
		repoNames[i] = metadata["name"].(string)
	}
	
	assert.Contains(t, repoNames, "mygodemo")
	assert.Contains(t, repoNames, "sensitive-repo")

	// Verify mock expectations
	mockClient.AssertExpectations(t)
} 