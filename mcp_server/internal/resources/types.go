package resources

import (
	"time"

	"k8s.io/apimachinery/pkg/runtime/schema"
)

// Giant Swarm CRD Group Version Resources
var (
	AppDeploymentGVR = schema.GroupVersionResource{
		Group:    "promise.platform.giantswarm.io",
		Version:  "v1beta1",
		Resource: "appdeployments",
	}
	GitHubAppGVR = schema.GroupVersionResource{
		Group:    "promise.platform.giantswarm.io",
		Version:  "v1beta1",
		Resource: "githubapps",
	}
	GitHubRepoGVR = schema.GroupVersionResource{
		Group:    "promise.platform.giantswarm.io",
		Version:  "v1beta1",
		Resource: "githubrepos",
	}
	// Kratix Promise GVR for discovering promises
	KratixPromiseGVR = schema.GroupVersionResource{
		Group:    "platform.kratix.io",
		Version:  "v1alpha1",
		Resource: "promises",
	}
)

// ResourceResponse represents the formatted response for MCP resources
type ResourceResponse struct {
	APIVersion string                   `json:"apiVersion"`
	Kind       string                   `json:"kind"`
	Items      []map[string]interface{} `json:"items"`
	Metadata   ResponseMetadata         `json:"metadata"`
}

// ResponseMetadata contains metadata about the resource response
type ResponseMetadata struct {
	Namespace   string                 `json:"namespace,omitempty"`
	Count       int                    `json:"count"`
	LastUpdated time.Time              `json:"lastUpdated"`
	ClusterInfo map[string]interface{} `json:"clusterInfo"`
}

// CRDResourceType represents the different types of CRDs we handle
type CRDResourceType string

const (
	AppDeploymentType CRDResourceType = "AppDeployment"
	GitHubAppType     CRDResourceType = "GitHubApp"
	GitHubRepoType    CRDResourceType = "GitHubRepo"
)

// PromiseGVR represents a Group/Version/Resource for JSON serialization
type PromiseGVR struct {
	Group    string `json:"group"`
	Version  string `json:"version"`
	Resource string `json:"resource"`
}

// PromiseSummary represents high-level Promise information
type PromiseSummary struct {
	Name           string                `json:"name"`
	Version        string                `json:"version"`
	Description    string                `json:"description"`
	TargetResource PromiseTargetResource `json:"target_resource"`
}

// PromiseTargetResource represents the CRD that the Promise creates
type PromiseTargetResource struct {
	Group    string `json:"group"`
	Version  string `json:"version"`
	Resource string `json:"resource"`
	Kind     string `json:"kind"`
	Scope    string `json:"scope"`
}

// PromiseSchema represents the full schema information
type PromiseSchema struct {
	PromiseName    string                 `json:"promise_name"`
	Version        string                 `json:"version"`
	TargetResource PromiseTargetResource  `json:"target_resource"`
	OpenAPISchema  map[string]interface{} `json:"openapi_schema"`
}

// ValidationResult represents schema validation results
type ValidationResult struct {
	Valid            bool              `json:"valid"`
	PromiseName      string            `json:"promise_name"`
	ValidationResult ValidationDetails `json:"validation_result"`
}

// ValidationDetails contains detailed validation information
type ValidationDetails struct {
	Errors   []ValidationError `json:"errors"`
	Warnings []ValidationError `json:"warnings"`
}

// ValidationError represents a single validation error
type ValidationError struct {
	Field   string      `json:"field"`
	Message string      `json:"message"`
	Value   interface{} `json:"value"`
}

// PlatformCapability represents a high-level platform capability
type PlatformCapability struct {
	Name        string   `json:"name"`
	Groups      []string `json:"groups"`
	Description string   `json:"description"`
	Prompt      string   `json:"prompt"`
}

// CapabilityGroup represents a logical grouping of capabilities
type CapabilityGroup struct {
	Name         string   `json:"name"`
	Description  string   `json:"description"`
	Capabilities []string `json:"capabilities"`
}

// PlatformCapabilitiesResponse represents the response for listing platform capabilities
type PlatformCapabilitiesResponse struct {
	Capabilities []PlatformCapability `json:"capabilities"`
	Groups       []CapabilityGroup    `json:"groups"`
	Metadata     map[string]interface{} `json:"metadata"`
}
