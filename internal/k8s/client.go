package k8s

import (
	"github.com/derailed/popeye/pkg/config"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	clientcmd "k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	mv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
	versioned "k8s.io/metrics/pkg/client/clientset/versioned"
)

// Client represents an api server client.
type Client struct {
	flags        *config.Flags
	clientConfig clientcmd.ClientConfig
	rawConfig    *clientcmdapi.Config
	restConfig   *rest.Config
	api          kubernetes.Interface
}

// NewClient create Kubernetes apiservice client.
func NewClient(flags *config.Flags) *Client {
	return &Client{flags: flags}
}

// DialOrDie returns an api server client connection or dies.
func (c *Client) DialOrDie() kubernetes.Interface {
	client, err := c.Dial()
	if err != nil {
		panic(err)
	}
	return client
}

// Dial returns a handle to api server.
func (c *Client) Dial() (kubernetes.Interface, error) {
	if c.api != nil {
		return c.api, nil
	}

	cfg, err := c.RESTConfig()
	if err != nil {
		return nil, err
	}

	if c.api, err = kubernetes.NewForConfig(cfg); err != nil {
		return nil, err
	}
	return c.api, nil
}

// ActiveNamespace returns the active namespace from the args or kubeconfig.
// It returns all namespace is none is found.
func (c *Client) ActiveNamespace() string {
	if c.flags.AllNamespaces != nil && *c.flags.AllNamespaces {
		return ""
	}

	if isSet(c.flags.Namespace) {
		return *c.flags.Namespace
	}

	cfg, err := c.RawConfig()
	if err != nil {
		return ""
	}

	ctx := cfg.CurrentContext
	if isSet(c.flags.Context) {
		ctx = *c.flags.Context
	}

	if c, ok := cfg.Contexts[ctx]; ok {
		return c.Namespace
	}

	return ""
}

// IsActiveNamespace check if current ns is equal to specified namespace if a namespace is set.
func (c *Client) IsActiveNamespace(ns string) bool {
	if c.ActiveNamespace() == "" {
		return true
	}

	return ns == c.ActiveNamespace()
}

// ActiveCluster get the current cluster name.
func (c *Client) ActiveCluster() string {
	if isSet(c.flags.ClusterName) {
		return *c.flags.ClusterName
	}

	if isSet(c.flags.K8sPopeyeClusterName) {
		return *c.flags.ClusterName
	}

	cfg, err := c.RawConfig()
	if err != nil {
		return "n/a"
	}

	ctx := cfg.CurrentContext
	if isSet(c.flags.Context) {
		ctx = *c.flags.Context
	}

	if ctx, ok := cfg.Contexts[ctx]; ok {
		return ctx.Cluster
	}

	return "n/a"
}

// RawConfig fetch the current kubeconfig with no overrides.
func (c *Client) RawConfig() (clientcmdapi.Config, error) {
	if c.rawConfig != nil {
		return *c.rawConfig, nil
	}

	c.ensureClientConfig()
	raw, err := c.clientConfig.RawConfig()
	if err != nil {
		return clientcmdapi.Config{}, err
	}
	c.rawConfig = &raw

	return *c.rawConfig, nil
}

// RESTConfig fetch the current REST api-server connection.
func (c *Client) RESTConfig() (*rest.Config, error) {
	if c.restConfig != nil {
		return c.restConfig, nil
	}

	var err error
	c.ensureClientConfig()
	if c.restConfig, err = c.flags.ToRESTConfig(); err != nil {
		return nil, err
	}

	return c.restConfig, nil
}

func (c *Client) ensureClientConfig() {
	if c.clientConfig == nil {
		c.clientConfig = c.flags.ToRawKubeConfigLoader()
	}
}

var supportedMetricsAPIVersions = []string{"v1beta1"}

// ClusterHasMetrics checks if metrics server is available on the cluster.
func (c *Client) ClusterHasMetrics() (bool, error) {
	apiGroups, err := c.DialOrDie().Discovery().ServerGroups()
	if err != nil {
		return false, err
	}

	for _, discoveredAPIGroup := range apiGroups.Groups {
		if discoveredAPIGroup.Name != mv1beta1.GroupName {
			continue
		}
		for _, version := range discoveredAPIGroup.Versions {
			for _, supportedVersion := range supportedMetricsAPIVersions {
				if version.Version == supportedVersion {
					return true, nil
				}
			}
		}
	}

	return false, nil
}

// DialVersioned connects to the versioned client to pull cluster metrics.
func (c *Client) DialVersioned() (*versioned.Clientset, error) {
	restCfg, err := c.RESTConfig()
	if err != nil {
		return nil, err
	}

	client, err := versioned.NewForConfig(restCfg)
	if err != nil {
		return nil, err
	}
	return client, nil
}

// isSet checks if a string flag is set.
func isSet(s *string) bool {
	return s != nil && *s != ""
}
