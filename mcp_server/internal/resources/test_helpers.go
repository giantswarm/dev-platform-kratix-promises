package resources

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/yaml"
)

// GetTestDataPath returns the path to the testdata directory
func GetTestDataPath() string {
	_, filename, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(filename), "testdata")
}

// LoadGitHubRepoTestData loads the sample GitHubRepo resource from test data
func LoadGitHubRepoTestData() (*unstructured.Unstructured, error) {
	testDataPath := GetTestDataPath()
	filePath := filepath.Join(testDataPath, "github_repo_sample.yaml")
	
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read test data file: %w", err)
	}

	// Convert YAML to JSON first
	jsonData, err := yaml.YAMLToJSON(data)
	if err != nil {
		return nil, fmt.Errorf("failed to convert YAML to JSON: %w", err)
	}

	// Parse into unstructured object
	var obj map[string]interface{}
	if err := json.Unmarshal(jsonData, &obj); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	return &unstructured.Unstructured{Object: obj}, nil
}

// CreateMockGitHubRepoList creates a mock UnstructuredList containing GitHubRepo resources
func CreateMockGitHubRepoList(repos ...*unstructured.Unstructured) *unstructured.UnstructuredList {
	list := &unstructured.UnstructuredList{
		Object: map[string]interface{}{
			"apiVersion": "promise.platform.giantswarm.io/v1beta1",
			"kind":       "GitHubRepoList",
		},
		Items: make([]unstructured.Unstructured, len(repos)),
	}

	for i, repo := range repos {
		list.Items[i] = *repo
	}

	return list
}

// CreateEmptyGitHubRepoList creates an empty UnstructuredList for GitHubRepo resources
func CreateEmptyGitHubRepoList() *unstructured.UnstructuredList {
	return &unstructured.UnstructuredList{
		Object: map[string]interface{}{
			"apiVersion": "promise.platform.giantswarm.io/v1beta1",
			"kind":       "GitHubRepoList",
		},
		Items: []unstructured.Unstructured{},
	}
}

// GetMockClusterInfo returns mock cluster information for testing
func GetMockClusterInfo() map[string]interface{} {
	return map[string]interface{}{
		"context": "test-cluster-context",
		"server":  "https://test-cluster.example.com:443",
		"timeout": "30s",
	}
}

// GetSensitiveGitHubRepo creates a GitHubRepo resource with sensitive data for sanitization testing
func GetSensitiveGitHubRepo() *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "promise.platform.giantswarm.io/v1beta1",
			"kind":       "githubrepo",
			"metadata": map[string]interface{}{
				"name":            "sensitive-repo",
				"namespace":       "test-ns",
				"resourceVersion": "123456", // Should be removed
				"managedFields":   []interface{}{}, // Should be removed
				"selfLink":        "/api/v1/...", // Should be removed
			},
			"spec": map[string]interface{}{
				"githubTokenSecretRef": map[string]interface{}{ // Should be removed
					"name": "secret-token",
				},
				"registryInfoConfigMapRef": map[string]interface{}{ // Should be removed
					"name": "registry-config",
				},
				"repository": map[string]interface{}{
					"name":  "test-repo",
					"owner": "test-owner",
				},
			},
		},
	}
} 