package dag

import (
	"github.com/derailed/popeye/internal/k8s"
	"github.com/derailed/popeye/pkg/config"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ListServices list all included Services.
func ListServices(c *k8s.Client, cfg *config.Config) (map[string]*v1.Service, error) {
	svcs, err := listAllServices(c)
	if err != nil {
		return nil, err
	}

	res := make(map[string]*v1.Service, len(svcs))
	for fqn, svc := range svcs {
		if includeNS(c, cfg, svc.Namespace) && !cfg.ExcludedService(svc.Name) {
			res[fqn] = svc
		}
	}

	return res, nil
}

// ListAllServices fetch all Services on the cluster.
func listAllServices(c *k8s.Client) (map[string]*v1.Service, error) {
	ll, err := fetchServices(c)
	if err != nil {
		return nil, err
	}

	svcs := make(map[string]*v1.Service, len(ll.Items))
	for i := range ll.Items {
		svcs[metaFQN(ll.Items[i].ObjectMeta)] = &ll.Items[i]
	}

	return svcs, nil
}

// FetchServices retrieves all Services on the cluster.
func fetchServices(c *k8s.Client) (*v1.ServiceList, error) {
	return c.DialOrDie().CoreV1().Services("").List(metav1.ListOptions{})
}
