package k8s

import (
	"fmt"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	mapi "k8s.io/metrics/pkg/apis/metrics"
	mv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
	versioned "k8s.io/metrics/pkg/client/clientset/versioned"
)

type Metric struct {
	CPU, MEM string
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

// PodMetrics retrieves a given pod metrics from metrics-server.
func PodMetrics(c *Client, ns, pod string) (map[string]Metric, error) {
	mx := make(map[string]Metric)

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
		mx[c.Name] = Metric{CPU: fmt.Sprintf("%dm", cpuVal), MEM: fmt.Sprintf("%dMi", memVal)}
	}
	return mx, nil
}
