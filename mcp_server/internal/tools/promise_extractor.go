package tools

import (
	"fmt"
	"log/slog"

	"github.com/giantswarm/dev-platform-kratix-promises/mcp_server/internal/clients"
	"github.com/giantswarm/dev-platform-kratix-promises/mcp_server/internal/resources"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// PromiseExtractor handles extraction of Promise metadata and schemas
type PromiseExtractor struct {
	k8sClient clients.KubernetesClientInterface
	logger    *slog.Logger
}

// NewPromiseExtractor creates a new PromiseExtractor
func NewPromiseExtractor(k8sClient clients.KubernetesClientInterface) *PromiseExtractor {
	return &PromiseExtractor{
		k8sClient: k8sClient,
		logger:    slog.Default().With("component", "promise-extractor"),
	}
}

// ListPromiseSummaries returns a list of all Kratix Promises with basic metadata
func (e *PromiseExtractor) ListPromiseSummaries() ([]*resources.PromiseSummary, error) {
	e.logger.Info("Listing all Kratix Promise summaries")

	// List all Kratix Promise resources
	promises, err := e.k8sClient.ListResources(resources.KratixPromiseGVR, "")
	if err != nil {
		e.logger.Error("Failed to list Kratix Promises", "error", err)
		return nil, fmt.Errorf("failed to list Kratix Promises: %w", err)
	}

	var summaries []*resources.PromiseSummary
	for _, promise := range promises.Items {
		summary, err := e.extractPromiseSummary(&promise)
		if err != nil {
			e.logger.Warn("Failed to extract promise summary", "promise", promise.GetName(), "error", err)
			// Continue with other promises instead of failing completely
			continue
		}
		summaries = append(summaries, summary)
	}

	e.logger.Info("Successfully extracted promise summaries", "count", len(summaries))
	return summaries, nil
}

// GetPromiseSchema returns the complete schema for a specific Promise
func (e *PromiseExtractor) GetPromiseSchema(name string) (*resources.PromiseSchema, error) {
	e.logger.Info("Getting Promise schema", "promise", name)

	// Get the specific Promise
	promise, err := e.k8sClient.GetResource(resources.KratixPromiseGVR, "", name)
	if err != nil {
		e.logger.Error("Failed to get Promise", "promise", name, "error", err)
		return nil, fmt.Errorf("failed to get Promise %s: %w", name, err)
	}

	// Extract version from label
	version := e.extractVersionLabel(promise)

	// Extract target resource info
	targetResource, err := e.extractTargetResourceInfo(promise)
	if err != nil {
		return nil, fmt.Errorf("failed to extract target resource info: %w", err)
	}

	// Extract the stored and served version schema
	versionSpec, _, err := e.extractStoredAndServedVersion(promise)
	if err != nil {
		return nil, fmt.Errorf("failed to extract version spec: %w", err)
	}

	// Extract the OpenAPI schema
	openAPISchema, _, _ := unstructured.NestedMap(versionSpec, "schema", "openAPIV3Schema")
	if openAPISchema == nil {
		return nil, fmt.Errorf("no OpenAPI schema found in Promise %s", name)
	}

	schema := &resources.PromiseSchema{
		PromiseName:    name,
		Version:        version,
		TargetResource: *targetResource,
		OpenAPISchema:  openAPISchema,
	}

	e.logger.Info("Successfully extracted Promise schema", "promise", name)
	return schema, nil
}

// extractPromiseSummary extracts basic summary information from a Promise
func (e *PromiseExtractor) extractPromiseSummary(promise *unstructured.Unstructured) (*resources.PromiseSummary, error) {
	name := promise.GetName()

	// Extract version from label
	version := e.extractVersionLabel(promise)

	// Extract target resource info
	targetResource, err := e.extractTargetResourceInfo(promise)
	if err != nil {
		return nil, fmt.Errorf("failed to extract target resource info: %w", err)
	}

	// Extract description from the stored and served version
	versionSpec, _, err := e.extractStoredAndServedVersion(promise)
	if err != nil {
		return nil, fmt.Errorf("failed to extract version spec: %w", err)
	}

	description := e.extractDescription(versionSpec)

	return &resources.PromiseSummary{
		Name:           name,
		Version:        version,
		Description:    description,
		TargetResource: *targetResource,
	}, nil
}

// extractVersionLabel extracts the version from the kratix.io/promise-version label
func (e *PromiseExtractor) extractVersionLabel(promise *unstructured.Unstructured) string {
	labels := promise.GetLabels()
	if labels != nil {
		if version, ok := labels["kratix.io/promise-version"]; ok {
			return version
		}
	}
	return "unknown"
}

// extractTargetResourceInfo extracts the target CRD information from the Promise
func (e *PromiseExtractor) extractTargetResourceInfo(promise *unstructured.Unstructured) (*resources.PromiseTargetResource, error) {
	// Navigate to .spec.api.spec
	apiSpec, found, err := unstructured.NestedMap(promise.Object, "spec", "api", "spec")
	if err != nil {
		return nil, fmt.Errorf("failed to extract API spec: %w", err)
	}
	if !found {
		return nil, fmt.Errorf("no API spec found in Promise")
	}

	// Extract basic CRD information
	group, _, _ := unstructured.NestedString(apiSpec, "group")
	scope, _, _ := unstructured.NestedString(apiSpec, "scope")

	// Extract names
	resourceNames, _, _ := unstructured.NestedMap(apiSpec, "names")
	var plural, kind string
	if resourceNames != nil {
		if p, ok := resourceNames["plural"].(string); ok {
			plural = p
		}
		if k, ok := resourceNames["kind"].(string); ok {
			kind = k
		}
	}

	// Extract version from the stored and served version
	_, versionName, err := e.extractStoredAndServedVersion(promise)
	if err != nil {
		return nil, fmt.Errorf("failed to extract version name: %w", err)
	}

	return &resources.PromiseTargetResource{
		Group:    group,
		Version:  versionName,
		Resource: plural,
		Kind:     kind,
		Scope:    scope,
	}, nil
}

// extractStoredAndServedVersion finds the version that is both storage=true AND served=true
func (e *PromiseExtractor) extractStoredAndServedVersion(promise *unstructured.Unstructured) (map[string]interface{}, string, error) {
	versions, found, err := unstructured.NestedSlice(promise.Object, "spec", "api", "spec", "versions")
	if err != nil || !found {
		return nil, "", fmt.Errorf("no versions found in Promise")
	}

	// Find version that is both storage=true AND served=true
	for _, v := range versions {
		version, ok := v.(map[string]interface{})
		if !ok {
			continue
		}

		storage, _, _ := unstructured.NestedBool(version, "storage")
		served, _, _ := unstructured.NestedBool(version, "served")
		versionName, _, _ := unstructured.NestedString(version, "name")

		if storage && served {
			return version, versionName, nil
		}
	}

	return nil, "", fmt.Errorf("no storage+served version found")
}

// extractDescription extracts description from version spec's OpenAPI schema
func (e *PromiseExtractor) extractDescription(versionSpec map[string]interface{}) string {
	description, _, _ := unstructured.NestedString(versionSpec, "schema", "openAPIV3Schema", "description")
	if description == "" {
		return "No description available"
	}
	return description
}
