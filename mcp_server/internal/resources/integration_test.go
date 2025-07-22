package resources

import (
	"testing"

	"github.com/giantswarm/dev-platform-kratix-promises/mcp_server/internal/resources/mocks"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// TestGitHubRepoResourceIntegration tests the full flow from MCP request to response
func TestGitHubRepoResourceIntegration(t *testing.T) {
	// Load real test data
	testRepo, err := LoadGitHubRepoTestData()
	require.NoError(t, err, "Failed to load real GitHubRepo test data")

	// Verify the loaded data has the expected structure
	assert.Equal(t, "promise.platform.giantswarm.io/v1beta1", testRepo.GetAPIVersion())
	assert.Equal(t, "githubrepo", testRepo.GetKind())
	assert.Equal(t, "mygodemo", testRepo.GetName())
	assert.Equal(t, "org-demotech", testRepo.GetNamespace())

	// Verify the spec contains expected fields
	spec, found, err := unstructured.NestedMap(testRepo.Object, "spec")
	require.NoError(t, err)
	require.True(t, found, "spec should be present")
	
	// Check repository information
	repository, found := spec["repository"].(map[string]interface{})
	require.True(t, found, "repository spec should be present")
	assert.Equal(t, "mygodemo", repository["name"])
	assert.Equal(t, "DemoTechInc", repository["owner"])
	assert.Equal(t, "giantswarm/devplatform-template-go-service", repository["templateSource"])

	// Setup mock client with real data
	mockClient := new(mocks.MockKubernetesClient)
	handler := NewCRDResourceHandler(mockClient)

	mockList := CreateMockGitHubRepoList(testRepo)
	mockClient.On("ListResources", GitHubRepoGVR, "").Return(mockList, nil)
	mockClient.On("GetClusterInfo").Return(GetMockClusterInfo())

	// Create MCP request
	request := mcp.ReadResourceRequest{
		Params: struct {
			URI string `json:"uri"`
			Arguments map[string]interface{} `json:"arguments,omitempty"`
		}{
			URI: "k8s://githubrepos",
		},
	}

	// Execute the full MCP resource flow
	result, err := handler.HandleGitHubRepos(request)
	require.NoError(t, err)
	require.Len(t, result, 1, "Expected one result item")

	// Verify the MCP response structure
	response, ok := result[0].(*ResourceResponse)
	require.True(t, ok, "Expected ResourceResponse type")

	// Verify response metadata
	assert.Equal(t, "promise.platform.giantswarm.io/v1beta1", response.APIVersion)
	assert.Equal(t, "GitHubRepoList", response.Kind)
	assert.Equal(t, 1, response.Metadata.Count)
	assert.NotZero(t, response.Metadata.LastUpdated)
	assert.Equal(t, "test-cluster-context", response.Metadata.ClusterInfo["context"])

	// Verify the actual GitHubRepo data in the response
	require.Len(t, response.Items, 1, "Expected one GitHubRepo item")
	githubrepo := response.Items[0]

	// Check metadata
	metadata := githubrepo["metadata"].(map[string]interface{})
	assert.Equal(t, "mygodemo", metadata["name"])
	assert.Equal(t, "org-demotech", metadata["namespace"])
	assert.Contains(t, metadata, "creationTimestamp")
	assert.Contains(t, metadata, "labels")

	// Check spec
	responseSpec := githubrepo["spec"].(map[string]interface{})
	responseRepository := responseSpec["repository"].(map[string]interface{})
	assert.Equal(t, "mygodemo", responseRepository["name"])
	assert.Equal(t, "DemoTechInc", responseRepository["owner"])
	assert.Equal(t, "A go service demo", responseRepository["description"])

	// Check status (if present)
	if status, found := githubrepo["status"].(map[string]interface{}); found {
		assert.Contains(t, status, "conditions")
		assert.Contains(t, status, "message")
	}

	// Verify sensitive data was sanitized
	assert.NotContains(t, metadata, "resourceVersion", "resourceVersion should be sanitized")
	assert.NotContains(t, metadata, "managedFields", "managedFields should be sanitized")
	assert.NotContains(t, responseSpec, "githubTokenSecretRef", "githubTokenSecretRef should be sanitized")

	// Verify mock expectations were met
	mockClient.AssertExpectations(t)
}

// TestMCPResourceResponseFormat verifies the response format matches MCP specification
func TestMCPResourceResponseFormat(t *testing.T) {
	mockClient := new(mocks.MockKubernetesClient)
	handler := NewCRDResourceHandler(mockClient)

	// Test with multiple resources
	testRepo1, err := LoadGitHubRepoTestData()
	require.NoError(t, err)
	
	testRepo2 := GetSensitiveGitHubRepo()
	mockList := CreateMockGitHubRepoList(testRepo1, testRepo2)

	mockClient.On("ListResources", GitHubRepoGVR, "").Return(mockList, nil)
	mockClient.On("GetClusterInfo").Return(GetMockClusterInfo())

	request := mcp.ReadResourceRequest{
		Params: struct {
			URI string `json:"uri"`
			Arguments map[string]interface{} `json:"arguments,omitempty"`
		}{
			URI: "k8s://githubrepos",
		},
	}

	result, err := handler.HandleGitHubRepos(request)
	require.NoError(t, err)
	require.Len(t, result, 1)

	response := result[0].(*ResourceResponse)

	// Verify MCP response format compliance
	assert.NotEmpty(t, response.APIVersion, "APIVersion should be set")
	assert.NotEmpty(t, response.Kind, "Kind should be set")
	assert.NotNil(t, response.Items, "Items should not be nil")
	assert.NotZero(t, response.Metadata.Count, "Count should reflect actual items")
	assert.Equal(t, len(response.Items), response.Metadata.Count, "Count should match items length")

	// Verify all items have required Kubernetes resource fields
	for i, item := range response.Items {
		assert.Contains(t, item, "apiVersion", "Item %d should have apiVersion", i)
		assert.Contains(t, item, "kind", "Item %d should have kind", i)
		assert.Contains(t, item, "metadata", "Item %d should have metadata", i)
		assert.Contains(t, item, "spec", "Item %d should have spec", i)

		metadata := item["metadata"].(map[string]interface{})
		assert.Contains(t, metadata, "name", "Item %d metadata should have name", i)
		assert.Contains(t, metadata, "namespace", "Item %d metadata should have namespace", i)
	}

	mockClient.AssertExpectations(t)
}

// TestAppDeploymentResourceIntegration tests the full flow from MCP request to response for AppDeployment
func TestAppDeploymentResourceIntegration(t *testing.T) {
	// Load real test data
	testAppDeployment, err := LoadAppDeploymentTestData()
	require.NoError(t, err, "Failed to load real AppDeployment test data")

	// Verify the loaded data has the expected structure
	assert.Equal(t, "promise.platform.giantswarm.io/v1beta1", testAppDeployment.GetAPIVersion())
	assert.Equal(t, "appdeployment", testAppDeployment.GetKind())
	assert.Equal(t, "mygodemo", testAppDeployment.GetName())
	assert.Equal(t, "org-demotech", testAppDeployment.GetNamespace())

	// Setup mock client with real data
	mockClient := new(mocks.MockKubernetesClient)
	handler := NewCRDResourceHandler(mockClient)

	mockList := CreateMockAppDeploymentList(testAppDeployment)
	mockClient.On("ListResources", AppDeploymentGVR, "").Return(mockList, nil)
	mockClient.On("GetClusterInfo").Return(GetMockClusterInfo())

	// Create MCP request
	request := mcp.ReadResourceRequest{
		Params: struct {
			URI string `json:"uri"`
			Arguments map[string]interface{} `json:"arguments,omitempty"`
		}{
			URI: "k8s://appdeployments",
		},
	}

	// Execute the full MCP resource flow
	result, err := handler.HandleAppDeployments(request)
	require.NoError(t, err)
	require.Len(t, result, 1, "Expected one result item")

	// Verify the MCP response structure
	response, ok := result[0].(*ResourceResponse)
	require.True(t, ok, "Expected ResourceResponse type")

	// Verify response metadata
	assert.Equal(t, "promise.platform.giantswarm.io/v1beta1", response.APIVersion)
	assert.Equal(t, "AppDeploymentList", response.Kind)
	assert.Equal(t, 1, response.Metadata.Count)
	assert.NotZero(t, response.Metadata.LastUpdated)

	// Verify the actual AppDeployment data in the response
	require.Len(t, response.Items, 1, "Expected one AppDeployment item")
	appDeployment := response.Items[0]

	// Check metadata
	metadata := appDeployment["metadata"].(map[string]interface{})
	assert.Equal(t, "mygodemo", metadata["name"])
	assert.Equal(t, "org-demotech", metadata["namespace"])

	// Check spec
	responseSpec := appDeployment["spec"].(map[string]interface{})
	assert.Equal(t, "mygodemo.demotech-rds.awsprod.gigantic.io", responseSpec["ingressHost"])
	assert.Equal(t, "default", responseSpec["targetNamespace"])

	// Verify sensitive data was sanitized
	if kubeConfig, found := responseSpec["kubeConfig"].(map[string]interface{}); found {
		assert.NotContains(t, kubeConfig, "secretRef", "kubeConfig.secretRef should be sanitized")
	}

	mockClient.AssertExpectations(t)
}

// TestGitHubAppResourceIntegration tests the full flow from MCP request to response for GitHubApp
func TestGitHubAppResourceIntegration(t *testing.T) {
	// Load real test data
	testGitHubApp, err := LoadGitHubAppTestData()
	require.NoError(t, err, "Failed to load real GitHubApp test data")

	// Verify the loaded data has the expected structure
	assert.Equal(t, "promise.platform.giantswarm.io/v1beta1", testGitHubApp.GetAPIVersion())
	assert.Equal(t, "githubapp", testGitHubApp.GetKind())
	assert.Equal(t, "mygodemo", testGitHubApp.GetName())
	assert.Equal(t, "org-demotech", testGitHubApp.GetNamespace())

	// Setup mock client with real data
	mockClient := new(mocks.MockKubernetesClient)
	handler := NewCRDResourceHandler(mockClient)

	mockList := CreateMockGitHubAppList(testGitHubApp)
	mockClient.On("ListResources", GitHubAppGVR, "").Return(mockList, nil)
	mockClient.On("GetClusterInfo").Return(GetMockClusterInfo())

	// Create MCP request
	request := mcp.ReadResourceRequest{
		Params: struct {
			URI string `json:"uri"`
			Arguments map[string]interface{} `json:"arguments,omitempty"`
		}{
			URI: "k8s://githubapps",
		},
	}

	// Execute the full MCP resource flow
	result, err := handler.HandleGitHubApps(request)
	require.NoError(t, err)
	require.Len(t, result, 1, "Expected one result item")

	// Verify the MCP response structure
	response, ok := result[0].(*ResourceResponse)
	require.True(t, ok, "Expected ResourceResponse type")

	// Verify response metadata
	assert.Equal(t, "promise.platform.giantswarm.io/v1beta1", response.APIVersion)
	assert.Equal(t, "GitHubAppList", response.Kind)
	assert.Equal(t, 1, response.Metadata.Count)
	assert.NotZero(t, response.Metadata.LastUpdated)

	// Verify the actual GitHubApp data in the response
	require.Len(t, response.Items, 1, "Expected one GitHubApp item")
	githubApp := response.Items[0]

	// Check metadata
	metadata := githubApp["metadata"].(map[string]interface{})
	assert.Equal(t, "mygodemo", metadata["name"])
	assert.Equal(t, "org-demotech", metadata["namespace"])

	// Check nested specs exist
	responseSpec := githubApp["spec"].(map[string]interface{})
	assert.Contains(t, responseSpec, "appDeployment")
	assert.Contains(t, responseSpec, "githubRepo")

	// Verify sensitive data was sanitized in nested structures
	if githubRepo, found := responseSpec["githubRepo"].(map[string]interface{}); found {
		if repoSpec, found := githubRepo["spec"].(map[string]interface{}); found {
			assert.NotContains(t, repoSpec, "githubTokenSecretRef", "Nested githubTokenSecretRef should be sanitized")
		}
	}

	mockClient.AssertExpectations(t)
}

// TestMultipleCRDTypesIntegration tests MCP responses for all three CRD types
func TestMultipleCRDTypesIntegration(t *testing.T) {
	// Load test data for all types
	testRepo, err := LoadGitHubRepoTestData()
	require.NoError(t, err)
	
	testAppDeployment, err := LoadAppDeploymentTestData()
	require.NoError(t, err)
	
	testGitHubApp, err := LoadGitHubAppTestData()
	require.NoError(t, err)

	// Test each resource type
	testCases := []struct {
		name         string
		uri          string
		gvr          schema.GroupVersionResource
		listCreator  func() *unstructured.UnstructuredList
		expectedKind string
	}{
		{
			name:         "GitHubRepo",
			uri:          "k8s://githubrepos",
			gvr:          GitHubRepoGVR,
			listCreator:  func() *unstructured.UnstructuredList { return CreateMockGitHubRepoList(testRepo) },
			expectedKind: "GitHubRepoList",
		},
		{
			name:         "AppDeployment",
			uri:          "k8s://appdeployments",
			gvr:          AppDeploymentGVR,
			listCreator:  func() *unstructured.UnstructuredList { return CreateMockAppDeploymentList(testAppDeployment) },
			expectedKind: "AppDeploymentList",
		},
		{
			name:         "GitHubApp",
			uri:          "k8s://githubapps",
			gvr:          GitHubAppGVR,
			listCreator:  func() *unstructured.UnstructuredList { return CreateMockGitHubAppList(testGitHubApp) },
			expectedKind: "GitHubAppList",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockClient := new(mocks.MockKubernetesClient)
			handler := NewCRDResourceHandler(mockClient)

			mockList := tc.listCreator()
			mockClient.On("ListResources", tc.gvr, "").Return(mockList, nil)
			mockClient.On("GetClusterInfo").Return(GetMockClusterInfo())

			request := mcp.ReadResourceRequest{
				Params: struct {
					URI string `json:"uri"`
					Arguments map[string]interface{} `json:"arguments,omitempty"`
				}{
					URI: tc.uri,
				},
			}

			// Execute based on resource type
			var result []interface{}
			var err error
			switch tc.name {
			case "GitHubRepo":
				result, err = handler.HandleGitHubRepos(request)
			case "AppDeployment":
				result, err = handler.HandleAppDeployments(request)
			case "GitHubApp":
				result, err = handler.HandleGitHubApps(request)
			}

			require.NoError(t, err)
			require.Len(t, result, 1)

			response, ok := result[0].(*ResourceResponse)
			require.True(t, ok, "Expected ResourceResponse type")

			assert.Equal(t, "promise.platform.giantswarm.io/v1beta1", response.APIVersion)
			assert.Equal(t, tc.expectedKind, response.Kind)
			assert.Equal(t, 1, response.Metadata.Count)
			assert.Len(t, response.Items, 1)

			mockClient.AssertExpectations(t)
		})
	}
}

// TestRealDataConsistency verifies the test data matches the expected Giant Swarm CRD structures
func TestRealDataConsistency(t *testing.T) {
	// Test GitHubRepo data consistency
	t.Run("GitHubRepo", func(t *testing.T) {
		testRepo, err := LoadGitHubRepoTestData()
		require.NoError(t, err)

		// Verify this is the expected resource from the kubectl command
		assert.Equal(t, "mygodemo", testRepo.GetName())
		assert.Equal(t, "org-demotech", testRepo.GetNamespace())
		assert.Equal(t, "promise.platform.giantswarm.io/v1beta1", testRepo.GetAPIVersion())
		assert.Equal(t, "githubrepo", testRepo.GetKind())

		// Verify required spec fields
		spec, found, err := unstructured.NestedMap(testRepo.Object, "spec")
		require.NoError(t, err)
		require.True(t, found)

		// Check for expected spec structure
		expectedSpecFields := []string{"repository", "backstageCatalogEntity"}
		for _, field := range expectedSpecFields {
			assert.Contains(t, spec, field, "spec should contain %s", field)
		}

		// Check repository details
		repository := spec["repository"].(map[string]interface{})
		assert.Equal(t, "mygodemo", repository["name"])
		assert.Equal(t, "DemoTechInc", repository["owner"])
		assert.Equal(t, "private", repository["visibility"])

		// Check that sensitive fields exist in raw data (they should be removed during processing)
		assert.Contains(t, spec, "githubTokenSecretRef", "Raw data should contain sensitive fields")
		assert.Contains(t, spec, "registryInfoConfigMapRef", "Raw data should contain sensitive fields")
	})

	// Test AppDeployment data consistency
	t.Run("AppDeployment", func(t *testing.T) {
		testAppDeployment, err := LoadAppDeploymentTestData()
		require.NoError(t, err)

		assert.Equal(t, "mygodemo", testAppDeployment.GetName())
		assert.Equal(t, "org-demotech", testAppDeployment.GetNamespace())
		assert.Equal(t, "promise.platform.giantswarm.io/v1beta1", testAppDeployment.GetAPIVersion())
		assert.Equal(t, "appdeployment", testAppDeployment.GetKind())

		// Verify spec structure
		spec, found, err := unstructured.NestedMap(testAppDeployment.Object, "spec")
		require.NoError(t, err)
		require.True(t, found)

		expectedSpecFields := []string{"ingressHost", "targetNamespace", "kubeConfig"}
		for _, field := range expectedSpecFields {
			assert.Contains(t, spec, field, "spec should contain %s", field)
		}

		// Check that sensitive fields exist in raw data
		kubeConfig := spec["kubeConfig"].(map[string]interface{})
		assert.Contains(t, kubeConfig, "secretRef", "Raw data should contain sensitive secretRef")
	})

	// Test GitHubApp data consistency
	t.Run("GitHubApp", func(t *testing.T) {
		testGitHubApp, err := LoadGitHubAppTestData()
		require.NoError(t, err)

		assert.Equal(t, "mygodemo", testGitHubApp.GetName())
		assert.Equal(t, "org-demotech", testGitHubApp.GetNamespace())
		assert.Equal(t, "promise.platform.giantswarm.io/v1beta1", testGitHubApp.GetAPIVersion())
		assert.Equal(t, "githubapp", testGitHubApp.GetKind())

		// Verify nested structure
		spec, found, err := unstructured.NestedMap(testGitHubApp.Object, "spec")
		require.NoError(t, err)
		require.True(t, found)

		assert.Contains(t, spec, "appDeployment", "spec should contain appDeployment")
		assert.Contains(t, spec, "githubRepo", "spec should contain githubRepo")
	})
} 