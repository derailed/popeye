package dag

import (
	"github.com/derailed/popeye/internal/k8s"
	"github.com/derailed/popeye/pkg/config"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ListStatefulSets list available StatefulSets.
func ListStatefulSets(c *k8s.Client, cfg *config.Config) (map[string]*appsv1.StatefulSet, error) {
	sts, err := listAllStatefulSets(c)
	if err != nil {
		return nil, err
	}

	res := make(map[string]*appsv1.StatefulSet, len(sts))
	for fqn, st := range sts {
		if c.IsActiveNamespace(st.Namespace) && !cfg.ExcludedNS(st.Namespace) {
			res[fqn] = st
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

	sts := make(map[string]*appsv1.StatefulSet, len(ll.Items))
	for i := range ll.Items {
		sts[metaFQN(ll.Items[i].ObjectMeta)] = &ll.Items[i]
	}

	return sts, nil
}

// FetchStatefulSets retrieves all StatefulSets on the cluster.
func fetchStatefulSets(c *k8s.Client) (*appsv1.StatefulSetList, error) {
	return c.DialOrDie().AppsV1().StatefulSets("").List(metav1.ListOptions{})
}
