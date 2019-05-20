package dag

import (
	"github.com/derailed/popeye/internal/k8s"
	"github.com/derailed/popeye/pkg/config"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ListStatefulSets list available StatefulSets.
func ListStatefulSets(c *k8s.Client, cfg *config.Config) (map[string]*appsv1.StatefulSet, error) {
	sas, err := listAllStatefulSets(c)
	if err != nil {
		return nil, err
	}

	res := make(map[string]*appsv1.StatefulSet, len(sas))
	for fqn, sa := range sas {
		if c.IsActiveNamespace(sa.Namespace) && !cfg.ExcludedNS(sa.Namespace) {
			res[fqn] = sa
		}
	}

	return res, nil
}

// ListAllStatefulSets fetch all StatefulSets on the cluster.
func listAllStatefulSets(c *k8s.Client) (map[string]*appsv1.StatefulSet, error) {
	ll, err := fetchStatefulSets(c)
	if err != nil {
		return nil, err
	}

	sas := make(map[string]*appsv1.StatefulSet, len(ll.Items))
	for i := range ll.Items {
		sas[MetaFQN(ll.Items[i].ObjectMeta)] = &ll.Items[i]
	}

	return sas, nil
}

// FetchStatefulSets retrieves all StatefulSets on the cluster.
func fetchStatefulSets(c *k8s.Client) (*appsv1.StatefulSetList, error) {
	return c.DialOrDie().AppsV1().StatefulSets("").List(metav1.ListOptions{})
}
