package dag

import (
	"context"

	"github.com/derailed/popeye/internal/client"
	"github.com/derailed/popeye/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	mv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

// ListNodesMetrics fetch all available Node metrics on the cluster.
func ListNodesMetrics(c types.Connection) (map[string]*mv1beta1.NodeMetrics, error) {
	ll, err := fetchNodesMetrics(c)
	if err != nil {
		return map[string]*mv1beta1.NodeMetrics{}, err
	}

	pmx := make(map[string]*mv1beta1.NodeMetrics, len(ll.Items))
	for i := range ll.Items {
		pmx[metaFQN(ll.Items[i].ObjectMeta)] = &ll.Items[i]
	}

	return pmx, nil
}

// FetchNodesMetrics retrieves all Node metrics on the cluster.
func fetchNodesMetrics(c types.Connection) (*mv1beta1.NodeMetricsList, error) {
	vc, err := c.MXDial()
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), client.CallTimeout)
	defer cancel()
	return vc.MetricsV1beta1().NodeMetricses().List(ctx, metav1.ListOptions{})
}
