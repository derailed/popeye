package dag

import (
	"github.com/derailed/popeye/internal/k8s"
	"github.com/derailed/popeye/pkg/config"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ListNamespaces list all included Namespaces.
func ListNamespaces(c *k8s.Client, cfg *config.Config) (map[string]*v1.Namespace, error) {
	secs, err := listAllNamespaces(c)
	if err != nil {
		return nil, err
	}

	res := make(map[string]*v1.Namespace, len(secs))
	for fqn, sec := range secs {
		if c.IsActiveNamespace(sec.Namespace) && !cfg.ExcludedNS(sec.Namespace) {
			res[fqn] = sec
		}
	}

	return res, nil
}

// ListAllNamespaces fetch all Namespaces on the cluster.
func listAllNamespaces(c *k8s.Client) (map[string]*v1.Namespace, error) {
	ll, err := fetchNamespaces(c)
	if err != nil {
		return nil, err
	}

	secs := make(map[string]*v1.Namespace, len(ll.Items))
	for i := range ll.Items {
		secs[MetaFQN(ll.Items[i].ObjectMeta)] = &ll.Items[i]
	}

	return secs, nil
}

// FetchNamespaces retrieves all Namespaces on the cluster.
func fetchNamespaces(c *k8s.Client) (*v1.NamespaceList, error) {
	return c.DialOrDie().CoreV1().Namespaces().List(metav1.ListOptions{})
}
