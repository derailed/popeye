package dag

import (
	"github.com/derailed/popeye/internal/k8s"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	mv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

// ListPodsMetrics fetch all available Pod metrics on the cluster.
func ListPodsMetrics(c *k8s.Client) (map[string]*mv1beta1.PodMetrics, error) {
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
func fetchPodsMetrics(c *k8s.Client) (*mv1beta1.PodMetricsList, error) {
	vc, err := c.DialVersioned()
	if err != nil {
		return nil, err
	}

	return vc.MetricsV1beta1().PodMetricses(c.ActiveNamespace()).List(metav1.ListOptions{})
}
