package tools

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/yaml"
)

// TestPromiseData holds test data for a specific Promise
type TestPromiseData struct {
	Name        string
	Promise     *unstructured.Unstructured
	ValidSpec   map[string]interface{}
	InvalidSpec map[string]interface{}
}

// LoadPromiseTestData loads a Promise object from YAML test data
func LoadPromiseTestData(promiseName string) (*TestPromiseData, error) {
	// Load Promise YAML
	promisePath := filepath.Join("testdata", fmt.Sprintf("%s-promise.yaml", promiseName))
	promiseBytes, err := ioutil.ReadFile(promisePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read Promise file %s: %w", promisePath, err)
	}

	var promise unstructured.Unstructured
	if err := yaml.Unmarshal(promiseBytes, &promise); err != nil {
		return nil, fmt.Errorf("failed to unmarshal Promise YAML: %w", err)
	}

	// Load valid spec
	validSpecPath := filepath.Join("testdata", fmt.Sprintf("%s-spec-valid.json", promiseName))
	validSpecBytes, err := ioutil.ReadFile(validSpecPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read valid spec file %s: %w", validSpecPath, err)
	}

	var validSpec map[string]interface{}
	if err := json.Unmarshal(validSpecBytes, &validSpec); err != nil {
		return nil, fmt.Errorf("failed to unmarshal valid spec JSON: %w", err)
	}

	// Load invalid spec
	invalidSpecPath := filepath.Join("testdata", fmt.Sprintf("%s-spec-invalid.json", promiseName))
	invalidSpecBytes, err := ioutil.ReadFile(invalidSpecPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read invalid spec file %s: %w", invalidSpecPath, err)
	}

	var invalidSpec map[string]interface{}
	if err := json.Unmarshal(invalidSpecBytes, &invalidSpec); err != nil {
		return nil, fmt.Errorf("failed to unmarshal invalid spec JSON: %w", err)
	}

	return &TestPromiseData{
		Name:        promiseName,
		Promise:     &promise,
		ValidSpec:   validSpec,
		InvalidSpec: invalidSpec,
	}, nil
}

// CreateMockPromiseList creates a mock list of Promise objects for testing
func CreateMockPromiseList(promises ...*unstructured.Unstructured) *unstructured.UnstructuredList {
	list := &unstructured.UnstructuredList{
		Object: map[string]interface{}{
			"apiVersion": "platform.kratix.io/v1alpha1",
			"kind":       "PromiseList",
			"metadata": map[string]interface{}{
				"resourceVersion": "1234567",
			},
		},
	}

	for _, promise := range promises {
		list.Items = append(list.Items, *promise)
	}

	return list
}

// ExtractPromiseSchemaFromTestData extracts the OpenAPI schema from a test Promise
func ExtractPromiseSchemaFromTestData(promise *unstructured.Unstructured) (map[string]interface{}, error) {
	// Navigate to .spec.api.spec.versions[0].schema.openAPIV3Schema
	api, found, err := unstructured.NestedMap(promise.Object, "spec", "api")
	if err != nil || !found {
		return nil, fmt.Errorf("failed to get api spec: %w", err)
	}

	spec, found, err := unstructured.NestedMap(api, "spec")
	if err != nil || !found {
		return nil, fmt.Errorf("failed to get spec: %w", err)
	}

	versions, found, err := unstructured.NestedSlice(spec, "versions")
	if err != nil || !found || len(versions) == 0 {
		return nil, fmt.Errorf("failed to get versions: %w", err)
	}

	firstVersion, ok := versions[0].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid version format")
	}

	schema, found, err := unstructured.NestedMap(firstVersion, "schema", "openAPIV3Schema")
	if err != nil || !found {
		return nil, fmt.Errorf("failed to get openAPIV3Schema: %w", err)
	}

	return schema, nil
}

// GetMockClusterInfo returns mock cluster information for testing
func GetMockClusterInfo() map[string]interface{} {
	return map[string]interface{}{
		"context":     "test-cluster-context",
		"server":      "https://test-cluster.example.com",
		"version":     "v1.28.0",
		"environment": "test",
	}
}
