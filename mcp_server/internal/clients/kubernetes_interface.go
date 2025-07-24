package clients

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// KubernetesClientInterface defines the interface for Kubernetes client operations
type KubernetesClientInterface interface {
	ListResources(gvr schema.GroupVersionResource, namespace string) (*unstructured.UnstructuredList, error)
	GetResource(gvr schema.GroupVersionResource, namespace, name string) (*unstructured.Unstructured, error)
	CreateResource(gvr schema.GroupVersionResource, namespace string, obj *unstructured.Unstructured) (*unstructured.Unstructured, error)
	DeleteResource(gvr schema.GroupVersionResource, namespace, name string) error
	GetCurrentContext() string
	GetClusterInfo() map[string]interface{}
} 