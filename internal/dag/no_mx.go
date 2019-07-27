package dag

import (
	"github.com/derailed/popeye/internal/k8s"
	"github.com/rs/zerolog/log"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	mv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

// ListNodesMetrics fetch all available Node metrics on the cluster.
func ListNodesMetrics(c *k8s.Client) (map[string]*mv1beta1.NodeMetrics, error) {
	ll, err := fetchNodesMetrics(c)
	if err != nil {
		log.Debug().Err(err).Msg("ListAll")
		return map[string]*mv1beta1.NodeMetrics{}, err
	}

	pmx := make(map[string]*mv1beta1.NodeMetrics, len(ll.Items))
	for i := range ll.Items {
		pmx[metaFQN(ll.Items[i].ObjectMeta)] = &ll.Items[i]
	}

	return pmx, nil
}

// FetchNodesMetrics retrieves all Node metrics on the cluster.
func fetchNodesMetrics(c *k8s.Client) (*mv1beta1.NodeMetricsList, error) {
	vc, err := c.DialVersioned()
	if err != nil {
		return nil, err
	}

	return vc.MetricsV1beta1().NodeMetricses().List(metav1.ListOptions{})
}
