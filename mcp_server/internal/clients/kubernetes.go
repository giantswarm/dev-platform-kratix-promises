package clients

import (
	"context"
	"fmt"
	"log/slog"
	"path/filepath"
	"time"

	"github.com/giantswarm/dev-platform-kratix-promises/mcp_server/internal/config"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// KubernetesClient provides access to Kubernetes cluster resources
type KubernetesClient struct {
	client    dynamic.Interface
	config    *rest.Config
	timeout   time.Duration
	logger    *slog.Logger
	context   string
}

// NewKubernetesClient creates a new Kubernetes client using kubeconfig
func NewKubernetesClient(cfg *config.Config) (*KubernetesClient, error) {
	logger := slog.Default().With("component", "k8s-client")

	// Parse timeout
	timeout, err := time.ParseDuration(cfg.K8sTimeout)
	if err != nil {
		logger.Warn("Invalid K8S_TIMEOUT, using default", "timeout", cfg.K8sTimeout, "default", "30s")
		timeout = 30 * time.Second
	}

	// Determine kubeconfig path
	kubeConfigPath := cfg.KubeConfigPath
	if kubeConfigPath == "" {
		if home := homedir.HomeDir(); home != "" {
			kubeConfigPath = filepath.Join(home, ".kube", "config")
		}
	}

	logger.Info("Loading Kubernetes configuration", 
		"kubeconfig", kubeConfigPath, 
		"context", cfg.KubeContext)

	// Load kubeconfig
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	if kubeConfigPath != "" {
		loadingRules.ExplicitPath = kubeConfigPath
	}

	configOverrides := &clientcmd.ConfigOverrides{}
	if cfg.KubeContext != "" {
		configOverrides.CurrentContext = cfg.KubeContext
	}

	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		loadingRules,
		configOverrides,
	)

	// Build REST config
	restConfig, err := kubeConfig.ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load kubeconfig: %w", err)
	}

	// Set timeout
	restConfig.Timeout = timeout

	// Create dynamic client
	dynamicClient, err := dynamic.NewForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kubernetes client: %w", err)
	}

	// Get current context info
	rawConfig, err := kubeConfig.RawConfig()
	if err != nil {
		logger.Warn("Failed to get raw kubeconfig", "error", err)
	}

	currentContext := rawConfig.CurrentContext
	if cfg.KubeContext != "" {
		currentContext = cfg.KubeContext
	}

	logger.Info("Successfully connected to Kubernetes cluster", 
		"context", currentContext,
		"server", restConfig.Host)

	return &KubernetesClient{
		client:  dynamicClient,
		config:  restConfig,
		timeout: timeout,
		logger:  logger,
		context: currentContext,
	}, nil
}

// ListResources lists all resources of a specific Group/Version/Resource
func (k *KubernetesClient) ListResources(gvr schema.GroupVersionResource, namespace string) (*unstructured.UnstructuredList, error) {
	ctx, cancel := context.WithTimeout(context.Background(), k.timeout)
	defer cancel()

	k.logger.Debug("Listing Kubernetes resources", 
		"gvr", gvr.String(), 
		"namespace", namespace)

	var resourceInterface dynamic.ResourceInterface
	if namespace == "" {
		// List cluster-wide resources
		resourceInterface = k.client.Resource(gvr)
	} else {
		// List namespaced resources
		resourceInterface = k.client.Resource(gvr).Namespace(namespace)
	}

	result, err := resourceInterface.List(ctx, metav1.ListOptions{})
	if err != nil {
		k.logger.Error("Failed to list resources", 
			"gvr", gvr.String(), 
			"namespace", namespace, 
			"error", err)
		return nil, fmt.Errorf("failed to list %s: %w", gvr.String(), err)
	}

	k.logger.Debug("Successfully listed resources", 
		"gvr", gvr.String(), 
		"namespace", namespace, 
		"count", len(result.Items))

	return result, nil
}

// CreateResource creates a new resource in the cluster
func (k *KubernetesClient) CreateResource(gvr schema.GroupVersionResource, namespace string, obj *unstructured.Unstructured) (*unstructured.Unstructured, error) {
	ctx, cancel := context.WithTimeout(context.Background(), k.timeout)
	defer cancel()

	k.logger.Debug("Creating Kubernetes resource", 
		"gvr", gvr.String(), 
		"namespace", namespace, 
		"name", obj.GetName())

	var resourceInterface dynamic.ResourceInterface
	if namespace == "" {
		resourceInterface = k.client.Resource(gvr)
	} else {
		resourceInterface = k.client.Resource(gvr).Namespace(namespace)
	}

	result, err := resourceInterface.Create(ctx, obj, metav1.CreateOptions{})
	if err != nil {
		k.logger.Error("Failed to create resource", 
			"gvr", gvr.String(), 
			"namespace", namespace, 
			"name", obj.GetName(), 
			"error", err)
		return nil, fmt.Errorf("failed to create %s/%s: %w", gvr.String(), obj.GetName(), err)
	}

	k.logger.Info("Successfully created resource", 
		"gvr", gvr.String(), 
		"namespace", namespace, 
		"name", result.GetName(),
		"uid", result.GetUID())

	return result, nil
}

// GetResource gets a specific resource by name
func (k *KubernetesClient) GetResource(gvr schema.GroupVersionResource, namespace, name string) (*unstructured.Unstructured, error) {
	ctx, cancel := context.WithTimeout(context.Background(), k.timeout)
	defer cancel()

	k.logger.Debug("Getting Kubernetes resource", 
		"gvr", gvr.String(), 
		"namespace", namespace, 
		"name", name)

	var resourceInterface dynamic.ResourceInterface
	if namespace == "" {
		resourceInterface = k.client.Resource(gvr)
	} else {
		resourceInterface = k.client.Resource(gvr).Namespace(namespace)
	}

	result, err := resourceInterface.Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		k.logger.Error("Failed to get resource", 
			"gvr", gvr.String(), 
			"namespace", namespace, 
			"name", name, 
			"error", err)
		return nil, fmt.Errorf("failed to get %s/%s: %w", gvr.String(), name, err)
	}

	k.logger.Debug("Successfully got resource", 
		"gvr", gvr.String(), 
		"namespace", namespace, 
		"name", name)

	return result, nil
}

// GetCurrentContext returns the current Kubernetes context
func (k *KubernetesClient) GetCurrentContext() string {
	return k.context
}

// GetClusterInfo returns basic cluster information
func (k *KubernetesClient) GetClusterInfo() map[string]interface{} {
	return map[string]interface{}{
		"context": k.context,
		"server":  k.config.Host,
		"timeout": k.timeout.String(),
	}
} 