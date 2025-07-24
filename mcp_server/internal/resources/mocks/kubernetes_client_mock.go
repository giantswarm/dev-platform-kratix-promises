package mocks

import (
	"github.com/giantswarm/dev-platform-kratix-promises/mcp_server/internal/clients"
	"github.com/stretchr/testify/mock"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// MockKubernetesClient is a mock implementation of the KubernetesClient interface
type MockKubernetesClient struct {
	mock.Mock
}

// Ensure MockKubernetesClient implements the interface
var _ clients.KubernetesClientInterface = (*MockKubernetesClient)(nil)

// ListResources mocks the ListResources method
func (m *MockKubernetesClient) ListResources(gvr schema.GroupVersionResource, namespace string) (*unstructured.UnstructuredList, error) {
	args := m.Called(gvr, namespace)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*unstructured.UnstructuredList), args.Error(1)
}

// GetResource mocks the GetResource method
func (m *MockKubernetesClient) GetResource(gvr schema.GroupVersionResource, namespace, name string) (*unstructured.Unstructured, error) {
	args := m.Called(gvr, namespace, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*unstructured.Unstructured), args.Error(1)
}

// CreateResource mocks the CreateResource method
func (m *MockKubernetesClient) CreateResource(gvr schema.GroupVersionResource, namespace string, obj *unstructured.Unstructured) (*unstructured.Unstructured, error) {
	args := m.Called(gvr, namespace, obj)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*unstructured.Unstructured), args.Error(1)
}

// DeleteResource mocks the DeleteResource method
func (m *MockKubernetesClient) DeleteResource(gvr schema.GroupVersionResource, namespace, name string) error {
	args := m.Called(gvr, namespace, name)
	return args.Error(0)
}

// GetCurrentContext mocks the GetCurrentContext method
func (m *MockKubernetesClient) GetCurrentContext() string {
	args := m.Called()
	return args.String(0)
}

// GetClusterInfo mocks the GetClusterInfo method
func (m *MockKubernetesClient) GetClusterInfo() map[string]interface{} {
	args := m.Called()
	return args.Get(0).(map[string]interface{})
} 