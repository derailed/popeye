package dag

import (
	"github.com/derailed/popeye/internal/k8s"
	"github.com/derailed/popeye/pkg/config"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ListReplicaSets list all included ReplicaSets.
func ListReplicaSets(c *k8s.Client, cfg *config.Config) (map[string]*appsv1.ReplicaSet, error) {
	rss, err := listAllReplicaSets(c)
	if err != nil {
		return map[string]*appsv1.ReplicaSet{}, err
	}

	res := make(map[string]*appsv1.ReplicaSet, len(rss))
	for fqn, rs := range rss {
		if includeNS(c, rs.Namespace) {
			res[fqn] = rs
		}
	}

	return res, nil
}

// ListAllReplicaSets fetch all ReplicaSets on the cluster.
func listAllReplicaSets(c *k8s.Client) (map[string]*appsv1.ReplicaSet, error) {
	ll, err := fetchReplicaSets(c)
	if err != nil {
		return nil, err
	}

	rss := make(map[string]*appsv1.ReplicaSet, len(ll.Items))
	for i := range ll.Items {
		rss[metaFQN(ll.Items[i].ObjectMeta)] = &ll.Items[i]
	}

	return rss, nil
}

// FetchReplicaSets retrieves all ReplicaSets on the cluster.
func fetchReplicaSets(c *k8s.Client) (*appsv1.ReplicaSetList, error) {
	return c.DialOrDie().AppsV1().ReplicaSets(c.ActiveNamespace()).List(metav1.ListOptions{})
}
