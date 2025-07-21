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
	Namespace    string                 `json:"namespace,omitempty"`
	Count        int                    `json:"count"`
	LastUpdated  time.Time              `json:"lastUpdated"`
	ClusterInfo  map[string]interface{} `json:"clusterInfo"`
}

// CRDResourceType represents the different types of CRDs we handle
type CRDResourceType string

const (
	AppDeploymentType CRDResourceType = "AppDeployment"
	GitHubAppType     CRDResourceType = "GitHubApp"
	GitHubRepoType    CRDResourceType = "GitHubRepo"
) 