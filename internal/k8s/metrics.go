package k8s

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	mv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
	versioned "k8s.io/metrics/pkg/client/clientset/versioned"
)

const megaByte = 1024 * 1024

type (
	// NodeMetrics describes raw node metrics.
	NodeMetrics struct {
		CurrentCPU int64
		CurrentMEM float64
		AvailCPU   int64
		AvailMEM   float64
		TotalCPU   int64
		TotalMEM   float64
	}

	// NodesMetrics tracks usage metrics per nodes.
	NodesMetrics map[string]NodeMetrics

	// Metrics represent an aggregation of all pod containers metrics.
	Metrics struct {
		CurrentCPU int64
		CurrentMEM float64
	}

	// PodsMetrics tracks usage metrics per pods.
	PodsMetrics map[string]ContainerMetrics

	// ContainerMetrics tracks container metrics
	ContainerMetrics map[string]Metrics
)

// Empty checks if we have any metrics.
func (n NodeMetrics) Empty() bool {
	return n == NodeMetrics{}
}

// Empty checks if we have any metrics.
func (m Metrics) Empty() bool {
	return m == Metrics{}
}

// GetNodesMetrics retrieves metrics for a given set of nodes.
func GetNodesMetrics(nodes []v1.Node, metrics []mv1beta1.NodeMetrics, mmx NodesMetrics) {
	for _, n := range nodes {
		mmx[n.Name] = NodeMetrics{
			AvailCPU: n.Status.Allocatable.Cpu().MilliValue(),
			AvailMEM: asMi(n.Status.Allocatable.Memory().Value()),
			TotalCPU: n.Status.Capacity.Cpu().MilliValue(),
			TotalMEM: asMi(n.Status.Capacity.Memory().Value()),
		}
	}

	for _, c := range metrics {
		if mx, ok := mmx[c.Name]; ok {
			mx.CurrentCPU = c.Usage.Cpu().MilliValue()
			mx.CurrentMEM = asMi(c.Usage.Memory().Value())
			mmx[c.Name] = mx
		}
	}
}

// FetchNodesMetrics retrieve metrics from metrics server.
func FetchNodesMetrics(c *Client) ([]mv1beta1.NodeMetrics, error) {
	vClient, err := vClient(c)
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
func FetchPodsMetrics(c *Client, ns string) ([]mv1beta1.PodMetrics, error) {
	client, err := vClient(c)
	if err != nil {
		return nil, err
	}

	list, err := client.Metrics().PodMetricses(ns).List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return list.Items, nil
}

// GetPodsMetrics retrieves metrics for all pods in a given namespace.
func GetPodsMetrics(pods []mv1beta1.PodMetrics, mmx PodsMetrics) {
	// Compute all pod's containers metrics.
	for _, p := range pods {
		mx := make(ContainerMetrics, len(p.Containers))
		for _, c := range p.Containers {
			mx[c.Name] = Metrics{
				CurrentCPU: c.Usage.Cpu().MilliValue(),
				CurrentMEM: asMi(c.Usage.Memory().Value()),
			}
		}
		mmx[namespacedName(p)] = mx
	}
}

// ----------------------------------------------------------------------------
// Helpers...

func asMi(v int64) float64 {
	return float64(v) / megaByte
}

func namespacedName(mx mv1beta1.PodMetrics) string {
	return mx.Namespace + "/" + mx.Name
}

func vClient(c *Client) (*versioned.Clientset, error) {
	restCfg, err := c.Config.RESTConfig()
	if err != nil {
		return nil, err
	}

	client, err := versioned.NewForConfig(restCfg)
	if err != nil {
		return nil, err
	}
	return client, nil
}
