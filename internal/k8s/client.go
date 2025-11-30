package k8s

import (
	"context"
	"fmt"
	"os"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// Client wraps the Kubernetes client.
type Client struct {
	clientset kubernetes.Interface
}

// NewClient creates a new Kubernetes client.
// It tries to use the in-cluster config first, and falls back to KUBECONFIG if set.
func NewClient() (*Client, error) {
	var config *rest.Config
	var err error

	// Try to load from KUBECONFIG environment variable first (for local dev)
	if kubeConfigPath := os.Getenv("KUBECONFIG"); kubeConfigPath != "" {
		config, err = clientcmd.BuildConfigFromFlags("", kubeConfigPath)
	} else {
		// Fallback to in-cluster config
		config, err = rest.InClusterConfig()
	}

	if err != nil {
		return nil, fmt.Errorf("failed to load kubernetes config: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes clientset: %w", err)
	}

	return &Client{clientset: clientset}, nil
}

// GetIngressNodeNames retrieves the names of nodes matching the label.

func (c *Client) GetIngressNodeNames(ctx context.Context, labelSelector string) ([]string, error) {

	// Parse label selector? For simplicity, we assume the input is "key=value" or just "key".

	// The standard metav1.ListOptions expects a string like "key=value,key2=value2".

	// The config passes just the key (default "node-role.kubernetes.io/ingress").

	// We need to construct a selector "key=true" or "key exists".

	// If the label is just a key (no '='), we assume existence check.

	var selector string

	if strings.Contains(labelSelector, "=") {

		selector = labelSelector

	} else {

		selector = fmt.Sprintf("%s", labelSelector) // Existence check: just the key name works in some clients, but standard is "key" or "!key".

		// Actually for existence: "key".

	}

	nodes, err := c.clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{

		LabelSelector: selector,
	})

	if err != nil {

		return nil, fmt.Errorf("failed to list nodes with selector %q: %w", selector, err)

	}

	var names []string

	for _, node := range nodes.Items {

		names = append(names, node.Name)

	}

	return names, nil

}
