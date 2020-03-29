package dag

import (
	"context"

	"github.com/derailed/popeye/internal/client"
	"github.com/derailed/popeye/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	mv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

// ListPodsMetrics fetch all available Pod metrics on the cluster.
func ListPodsMetrics(c types.Connection) (map[string]*mv1beta1.PodMetrics, error) {
	ll, err := fetchPodsMetrics(c)
	if err != nil {
		return map[string]*mv1beta1.PodMetrics{}, err
	}

	pmx := make(map[string]*mv1beta1.PodMetrics, len(ll.Items))
	for i := range ll.Items {
		pmx[metaFQN(ll.Items[i].ObjectMeta)] = &ll.Items[i]
	}

	return pmx, nil
}

// FetchPodsMetrics retrieves all Pod metrics on the cluster.
func fetchPodsMetrics(c types.Connection) (*mv1beta1.PodMetricsList, error) {
	vc, err := c.MXDial()
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), client.CallTimeout)
	defer cancel()
	return vc.MetricsV1beta1().PodMetricses(c.ActiveNamespace()).List(ctx, metav1.ListOptions{})
}
