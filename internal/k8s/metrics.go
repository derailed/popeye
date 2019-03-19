package k8s

import (
	"fmt"
	"strings"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	mapi "k8s.io/metrics/pkg/apis/metrics"
	mv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
	versioned "k8s.io/metrics/pkg/client/clientset/versioned"
)

// ----------------------------------------------------------------------------

// PodMetric records a pod core metrics.
type PodMetric struct {
	CPU, MEM int64
}

// CurrentCPU returns the current CPU reading.
func (p PodMetric) CurrentCPU() int64 {
	return p.CPU
}

// CurrentMEM returns the current MEM reading.
func (p PodMetric) CurrentMEM() int64 {
	return p.MEM
}

// Empty return true if no metrics are present.
func (p PodMetric) Empty() bool {
	return p.CPU == 0 && p.MEM == 0
}

// ----------------------------------------------------------------------------

// NodeMetric records a node core metrics.
type NodeMetric struct {
	PodMetric
	AvailCPU, AvailMEM int64
}

// MaxCPU returns the max available CPU on this node.
func (n NodeMetric) MaxCPU() int64 {
	return n.AvailCPU
}

// MaxMEM returns the max available memory on this node.
func (n NodeMetric) MaxMEM() int64 {
	return n.AvailMEM
}

// Empty return true if no metrics are present.
func (n NodeMetric) Empty() bool {
	return n.PodMetric.Empty() && n.AvailCPU == 0 && n.AvailMEM == 0
}

// ----------------------------------------------------------------------------

// NodeMetrics retrieves a given node metrics from metrics-server.
func NodeMetrics(c *Client, no v1.Node) (NodeMetric, error) {
	var mx NodeMetric

	vClient, err := vClient(c)
	if err != nil {
		return mx, err
	}

	vmx, err := vClient.Metrics().NodeMetricses().List(metav1.ListOptions{
		FieldSelector: "metadata.name=" + no.Name,
	})
	if err != nil {
		return mx, err
	}

	var mmx mapi.NodeMetricsList
	if err = mv1beta1.Convert_v1beta1_NodeMetricsList_To_metrics_NodeMetricsList(vmx, &mmx, nil); err != nil {
		return mx, err
	}

	if len(mmx.Items) == 0 {
		return mx, fmt.Errorf("no metrics found for node %s", no)
	}

	metrics := mmx.Items[0]
	mx.CPU = metrics.Usage.Cpu().MilliValue()
	mx.MEM = metrics.Usage.Memory().Value() / (1024 * 1024)

	cpuQty := no.Status.Allocatable["cpu"]
	if cpu, ok := cpuQty.AsInt64(); ok {
		mx.AvailCPU = cpu * 1000 // unit of millicores
	}

	memQty := no.Status.Allocatable["memory"]
	if mem, ok := memQty.AsInt64(); ok {
		mx.AvailMEM = mem / (1024 * 1024)
	}

	return mx, nil
}

// PodMetrics retrieves a given pod metrics from metrics-server.
func PodMetrics(c *Client, ns, pod string) (map[string]PodMetric, error) {
	mx := make(map[string]PodMetric)

	vClient, err := vClient(c)
	if err != nil {
		return mx, err
	}

	fields := []string{"metadata.namespace=" + ns, "metadata.name=" + pod}
	vmx, err := vClient.Metrics().PodMetricses(ns).List(metav1.ListOptions{
		FieldSelector: strings.Join(fields, ","),
	})
	if err != nil {
		return mx, err
	}

	var mmx mapi.PodMetricsList
	if err = mv1beta1.Convert_v1beta1_PodMetricsList_To_metrics_PodMetricsList(vmx, &mmx, nil); err != nil {
		return mx, err
	}

	if len(mmx.Items) == 0 {
		return mx, fmt.Errorf("no metrics found for pod %s/%s", ns, pod)
	}

	for _, c := range mmx.Items[0].Containers {
		cpuVal := c.Usage.Cpu().MilliValue()
		memVal := c.Usage.Memory().Value() / (1024 * 1024)
		mx[c.Name] = PodMetric{CPU: cpuVal, MEM: memVal}
	}
	return mx, nil
}

func vClient(c *Client) (*versioned.Clientset, error) {
	restCfg, err := c.config.RESTConfig()
	if err != nil {
		return nil, err
	}

	client, err := versioned.NewForConfig(restCfg)
	if err != nil {
		return nil, err
	}
	return client, nil
}
