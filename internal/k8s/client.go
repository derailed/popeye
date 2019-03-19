package k8s

//go:generate popeye gen

import (
	"k8s.io/client-go/kubernetes"
	metricsapi "k8s.io/metrics/pkg/apis/metrics"
)

var (
	supportedMetricsAPIVersions = []string{"v1beta1"}
)

// Client represents a Kubernetes api server client.
type Client struct {
	config *Config

	api kubernetes.Interface
}

// NewClient returns a dialable api server configuration.
func NewClient(config *Config) *Client {
	return &Client{config: config}
}

// Dial returns a handle to api server.
func (c *Client) Dial() (kubernetes.Interface, error) {
	if c.api != nil {
		return c.api, nil
	}

	cfg, err := c.config.RESTConfig()
	if err != nil {
		return nil, err
	}

	if c.api, err = kubernetes.NewForConfig(cfg); err != nil {
		return nil, err
	}
	return c.api, nil
}

// ClusterHasMetrics checks if metrics server is available on the cluster.
func (c *Client) ClusterHasMetrics() bool {
	srv, err := c.Dial()
	if err != nil {
		return false
	}
	apiGroups, err := srv.Discovery().ServerGroups()
	if err != nil {
		return false
	}

	for _, discoveredAPIGroup := range apiGroups.Groups {
		if discoveredAPIGroup.Name != metricsapi.GroupName {
			continue
		}
		for _, version := range discoveredAPIGroup.Versions {
			for _, supportedVersion := range supportedMetricsAPIVersions {
				if version.Version == supportedVersion {
					return true
				}
			}
		}
	}
	return false
}
