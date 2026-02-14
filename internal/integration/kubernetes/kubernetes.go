package kubernetes

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// ContainerImage represents a unique container image running in the cluster.
type ContainerImage struct {
	// Image is the full image reference (e.g. "nginx:1.25", "gcr.io/project/app:sha256-...")
	Image string `json:"image"`
	// PodName is the name of a pod using this image
	PodName string `json:"pod_name"`
	// Namespace is the namespace of the pod
	Namespace string `json:"namespace"`
	// ContainerName is the container within the pod
	ContainerName string `json:"container_name"`
	// IsInit indicates whether this is an init container
	IsInit bool `json:"is_init"`
}

// ClusterInfo holds metadata about the connected cluster.
type ClusterInfo struct {
	ServerVersion string `json:"server_version"`
	NodeCount     int    `json:"node_count"`
	NamespaceCount int   `json:"namespace_count"`
}

// Client wraps the Kubernetes client-go clientset and operates via in-cluster config.
type Client struct {
	clientset *kubernetes.Clientset
}

// NewInClusterClient creates a Client using the in-cluster service account credentials.
// This only works when the application is running inside a Kubernetes pod.
func NewInClusterClient() (*Client, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, fmt.Errorf("not running in a Kubernetes cluster (in-cluster config unavailable): %w", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kubernetes client: %w", err)
	}

	return &Client{clientset: clientset}, nil
}

// IsAvailable returns true if the application is running inside a Kubernetes cluster.
func IsAvailable() bool {
	_, err := rest.InClusterConfig()
	return err == nil
}

// GetClusterInfo returns high-level metadata about the cluster.
func (c *Client) GetClusterInfo(ctx context.Context) (*ClusterInfo, error) {
	version, err := c.clientset.Discovery().ServerVersion()
	if err != nil {
		return nil, fmt.Errorf("failed to get server version: %w", err)
	}

	nodes, err := c.clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list nodes: %w", err)
	}

	namespaces, err := c.clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list namespaces: %w", err)
	}

	return &ClusterInfo{
		ServerVersion:  version.GitVersion,
		NodeCount:      len(nodes.Items),
		NamespaceCount: len(namespaces.Items),
	}, nil
}

// ListContainerImages lists all unique container images running across all namespaces.
// It scans all pods in all namespaces and collects both regular and init containers.
func (c *Client) ListContainerImages(ctx context.Context, namespace string) ([]ContainerImage, error) {
	var pods *corev1.PodList
	var err error

	if namespace == "" {
		pods, err = c.clientset.CoreV1().Pods(metav1.NamespaceAll).List(ctx, metav1.ListOptions{})
	} else {
		pods, err = c.clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
	}
	if err != nil {
		return nil, fmt.Errorf("failed to list pods: %w", err)
	}

	var images []ContainerImage
	seen := make(map[string]bool)

	for _, pod := range pods.Items {
		for _, c := range pod.Spec.Containers {
			key := fmt.Sprintf("%s/%s/%s", pod.Namespace, pod.Name, c.Name)
			if !seen[key] {
				seen[key] = true
				images = append(images, ContainerImage{
					Image:         c.Image,
					PodName:       pod.Name,
					Namespace:     pod.Namespace,
					ContainerName: c.Name,
					IsInit:        false,
				})
			}
		}
		for _, c := range pod.Spec.InitContainers {
			key := fmt.Sprintf("%s/%s/init/%s", pod.Namespace, pod.Name, c.Name)
			if !seen[key] {
				seen[key] = true
				images = append(images, ContainerImage{
					Image:         c.Image,
					PodName:       pod.Name,
					Namespace:     pod.Namespace,
					ContainerName: c.Name,
					IsInit:        true,
				})
			}
		}
	}

	return images, nil
}

// ListNamespaces returns all namespace names in the cluster.
func (c *Client) ListNamespaces(ctx context.Context) ([]string, error) {
	namespaces, err := c.clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list namespaces: %w", err)
	}

	names := make([]string, 0, len(namespaces.Items))
	for _, ns := range namespaces.Items {
		names = append(names, ns.Name)
	}
	return names, nil
}
