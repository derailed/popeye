package k8s

//go:generate popeye gen

import (
	"fmt"

	"github.com/derailed/popeye/internal/config"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	metricsapi "k8s.io/metrics/pkg/apis/metrics"
	mv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

var (
	supportedMetricsAPIVersions = []string{"v1beta1"}
	systemNS                    = []string{"kube-system", "kube-public"}
)

// Client represents a Kubernetes api server client.
type Client struct {
	*config.Config

	api        kubernetes.Interface
	pods       []v1.Pod
	namespaces []v1.Namespace
}

// NewClient returns a dialable api server configuration.
func NewClient(config *config.Config) *Client {
	return &Client{Config: config}
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

	cfg, err := c.Config.RESTConfig()
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

// FetchNodesMetrics fetch all node metrics.
func (c *Client) FetchNodesMetrics() ([]mv1beta1.NodeMetrics, error) {
	return FetchNodesMetrics(c)
}

// FetchPodsMetrics fetch all pods metrics in a given namespace.
func (c *Client) FetchPodsMetrics(ns string) ([]mv1beta1.PodMetrics, error) {
	return FetchPodsMetrics(c, ns)
}

// InUseNamespaces returns a list of namespaces referenced by pods.
func (c *Client) InUseNamespaces(nss []string) {
	pods, err := c.ListPods()
	if err != nil {
		return
	}

	ll := make(map[string]struct{})
	for _, p := range pods {
		ll[p.Namespace] = struct{}{}
	}

	var i int
	for k := range ll {
		nss[i] = k
		i++
	}
}

// GetEndpoints returns a endpoint by name.
func (c *Client) GetEndpoints(ns, n string) (*v1.Endpoints, error) {
	ep, err := c.DialOrDie().CoreV1().Endpoints(ns).Get(n, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return ep, nil
}

// ListServices list all available services in a given namespace.
func (c *Client) ListServices(ns string) ([]v1.Service, error) {
	ll, err := c.DialOrDie().CoreV1().Services(ns).List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	return ll.Items, nil
}

// ListNodes list all available nodes on the cluster.
func (c *Client) ListNodes() ([]v1.Node, error) {
	ll, err := c.DialOrDie().CoreV1().Nodes().List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	nodes := make([]v1.Node, 0, len(ll.Items))
	for _, no := range ll.Items {
		if !c.Config.ExcludedNode(no.Name) {
			nodes = append(nodes, no)
		}
	}

	return nodes, nil
}

// GetPod returns a pod via a label query.
func (c *Client) GetPod(sel string) (*v1.Pod, error) {
	pods, err := c.DialOrDie().CoreV1().Pods("").List(metav1.ListOptions{
		LabelSelector: sel,
	})
	if err != nil {
		return nil, err
	}

	if len(pods.Items) == 0 {
		return nil, fmt.Errorf("No pods match service selector")
	}

	return &pods.Items[0], nil
}

// ListPods list all available pods.
func (c *Client) ListPods() ([]v1.Pod, error) {
	if len(c.pods) != 0 {
		return c.pods, nil
	}

	ll, err := c.DialOrDie().CoreV1().
		Pods(c.Config.ActiveNamespace()).
		List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	c.pods = make([]v1.Pod, 0, len(ll.Items))
	for _, po := range ll.Items {
		if !c.Config.ExcludedNS(po.Namespace) {
			c.pods = append(c.pods, po)
		}
	}

	return c.pods, nil
}

// ListNS lists all available namespaces.
func (c *Client) ListNS() ([]v1.Namespace, error) {
	if len(c.namespaces) != 0 {
		return c.namespaces, nil
	}

	var (
		nn  *v1.NamespaceList
		err error
	)
	if ns := c.Config.ActiveNamespace(); ns == "" {
		nn, err = c.DialOrDie().CoreV1().Namespaces().List(metav1.ListOptions{})
	} else {
		var n *v1.Namespace
		n, err = c.DialOrDie().CoreV1().Namespaces().Get(ns, metav1.GetOptions{})
		nn = &v1.NamespaceList{Items: []v1.Namespace{*n}}
	}

	if err != nil {
		return nil, err
	}

	c.namespaces = make([]v1.Namespace, 0, len(nn.Items))
	for _, ns := range nn.Items {
		if !c.Config.ExcludedNS(ns.Name) {
			c.namespaces = append(c.namespaces, ns)
		}
	}

	return c.namespaces, nil
}

func isSystemNS(ns string) bool {
	for _, n := range systemNS {
		if n == ns {
			return true
		}
	}
	return false
}
