package dag

import (
	"github.com/derailed/popeye/internal/k8s"
	"github.com/derailed/popeye/pkg/config"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ListNamespaces list all included Namespaces.
func ListNamespaces(c *k8s.Client, cfg *config.Config) (map[string]*v1.Namespace, error) {
	nss, err := listAllNamespaces(c)
	if err != nil {
		return nil, err
	}

	res := make(map[string]*v1.Namespace, len(nss))
	for fqn, ns := range nss {
		if includeNS(c, cfg, ns.Name) {
			res[fqn] = ns
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

	nss := make(map[string]*v1.Namespace, len(ll.Items))
	for i := range ll.Items {
		nss[metaFQN(ll.Items[i].ObjectMeta)] = &ll.Items[i]
	}

	return nss, nil
}

// FetchNamespaces retrieves all Namespaces on the cluster.
func fetchNamespaces(c *k8s.Client) (*v1.NamespaceList, error) {
	return c.DialOrDie().CoreV1().Namespaces().List(metav1.ListOptions{})
}
