package k8s

import (
	v1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	clientcmd "k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	mv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
	versioned "k8s.io/metrics/pkg/client/clientset/versioned"
)

// Client represents an api server client.
type Client struct {
	Flags *Flags

	clientConfig clientcmd.ClientConfig
	rawConfig    *clientcmdapi.Config
	restConfig   *rest.Config
	api          kubernetes.Interface
}

// NewClient create Kubernetes apiservice client.
func NewClient(flags *Flags) *Client {
	return &Client{Flags: flags}
}

// FetchNamespaces retrieves all namespaces on the cluster.
func (c *Client) FetchNamespaces() (*v1.NamespaceList, error) {
	return c.DialOrDie().CoreV1().Namespaces().List(metav1.ListOptions{})
}

// FetchServiceAccounts retrieves all serviceaccounts on the cluster.
func (c *Client) FetchServiceAccounts() (*v1.ServiceAccountList, error) {
	return c.DialOrDie().CoreV1().ServiceAccounts("").List(metav1.ListOptions{})
}

// FetchConfigMaps retrieves all configmaps on the cluster.
func (c *Client) FetchConfigMaps() (*v1.ConfigMapList, error) {
	return c.DialOrDie().CoreV1().ConfigMaps("").List(metav1.ListOptions{})
}

// FetchSecrets retrieves all secrets on the cluster.
func (c *Client) FetchSecrets() (*v1.SecretList, error) {
	return c.DialOrDie().CoreV1().Secrets("").List(metav1.ListOptions{})
}

// FetchPods retrieves all pods on the cluster.
func (c *Client) FetchPods() (*v1.PodList, error) {
	return c.DialOrDie().CoreV1().Pods("").List(metav1.ListOptions{})
}

// FetchNodes retrieves all nodes on the cluster.
func (c *Client) FetchNodes() (*v1.NodeList, error) {
	return c.DialOrDie().CoreV1().Nodes().List(metav1.ListOptions{})
}

// FetchRoleBindings retrieves all RoleBindings on the cluster.
func (c *Client) FetchRoleBindings() (*rbacv1.RoleBindingList, error) {
	return c.DialOrDie().RbacV1().RoleBindings("").List(metav1.ListOptions{})
}

// FetchClusterRoleBindings retrieves all CRoleBindings on the cluster.
func (c *Client) FetchClusterRoleBindings() (*rbacv1.ClusterRoleBindingList, error) {
	return c.DialOrDie().RbacV1().ClusterRoleBindings().List(metav1.ListOptions{})
}

// FetchServices retrieves all services on the cluster.
func (c *Client) FetchServices() (*v1.ServiceList, error) {
	return c.DialOrDie().CoreV1().Services("").List(metav1.ListOptions{})
}

// FetchEndpoints retrieves all endpoints on the cluster.
func (c *Client) FetchEndpoints() (*v1.EndpointsList, error) {
	return c.DialOrDie().CoreV1().Endpoints(c.ActiveNamespace()).List(metav1.ListOptions{})
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
	if IsSet(c.Flags.Namespace) {
		return *c.Flags.Namespace
	}

	cfg, err := c.RawConfig()
	if err != nil {
		return "n/a"
	}

	ctx := cfg.CurrentContext
	if IsSet(c.Flags.Context) {
		ctx = *c.Flags.Context
	}

	if c, ok := cfg.Contexts[ctx]; ok {
		return c.Namespace
	}

	return ""
}

// ActiveCluster get the current cluster name.
func (c *Client) ActiveCluster() string {
	if IsSet(c.Flags.ClusterName) {
		return *c.Flags.ClusterName
	}

	cfg, err := c.RawConfig()
	if err != nil {
		return "n/a"
	}

	ctx := cfg.CurrentContext
	if IsSet(c.Flags.Context) {
		ctx = *c.Flags.Context
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

	err := c.ensureClientConfig()
	if err != nil {
		return clientcmdapi.Config{}, err
	}

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

	err := c.ensureClientConfig()
	if err != nil {
		return nil, err
	}

	if c.restConfig, err = c.Flags.ToRESTConfig(); err != nil {
		return nil, err
	}

	return c.restConfig, nil
}

func (c *Client) ensureClientConfig() error {
	if c.clientConfig == nil {
		c.clientConfig = c.Flags.ToRawKubeConfigLoader()
	}
	return nil
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

// FetchNodesMetrics retrieve metrics from metrics server.
func (c *Client) FetchNodesMetrics() ([]mv1beta1.NodeMetrics, error) {
	vClient, err := c.vClient()
	if err != nil {
		return nil, err
	}

	list, err := vClient.Metrics().NodeMetricses().List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	return list.Items, nil
}

// FetchPodsMetrics return all metrics for pods in a given namespace.
func (c *Client) FetchPodsMetrics(ns string) ([]mv1beta1.PodMetrics, error) {
	client, err := c.vClient()
	if err != nil {
		return nil, err
	}

	list, err := client.Metrics().PodMetricses(ns).List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	return list.Items, nil
}

func (c *Client) vClient() (*versioned.Clientset, error) {
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
