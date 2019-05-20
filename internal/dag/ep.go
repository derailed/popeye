package dag

import (
	"github.com/derailed/popeye/internal/k8s"
	"github.com/derailed/popeye/pkg/config"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ListEndpoints list all included Endpoints.
func ListEndpoints(c *k8s.Client, cfg *config.Config) (map[string]*v1.Endpoints, error) {
	cms, err := listAllEndpoints(c)
	if err != nil {
		return nil, err
	}

	res := make(map[string]*v1.Endpoints, len(cms))
	for fqn, cm := range cms {
		if c.IsActiveNamespace(cm.Namespace) && !cfg.ExcludedNS(cm.Namespace) {
			res[fqn] = cm
		}
	}

	return res, nil
}

// ListAllEndpoints fetch all Endpoints on the cluster.
func listAllEndpoints(c *k8s.Client) (map[string]*v1.Endpoints, error) {
	ll, err := fetchEndpoints(c)
	if err != nil {
		return nil, err
	}

	cms := make(map[string]*v1.Endpoints, len(ll.Items))
	for i := range ll.Items {
		cms[MetaFQN(ll.Items[i].ObjectMeta)] = &ll.Items[i]
	}

	return cms, nil
}

// FetchEndpoints retrieves all Endpoints on the cluster.
func fetchEndpoints(c *k8s.Client) (*v1.EndpointsList, error) {
	return c.DialOrDie().CoreV1().Endpoints("").List(metav1.ListOptions{})
}
