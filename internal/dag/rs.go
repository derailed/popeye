package dag

import (
	"github.com/derailed/popeye/internal/k8s"
	"github.com/derailed/popeye/pkg/config"
	"github.com/rs/zerolog/log"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ListReplicaSets list all included ReplicaSets.
func ListReplicaSets(c *k8s.Client, cfg *config.Config) (map[string]*appsv1.ReplicaSet, error) {
	dps, err := listAllReplicaSets(c)
	if err != nil {
		return map[string]*appsv1.ReplicaSet{}, err
	}

	res := make(map[string]*appsv1.ReplicaSet, len(dps))
	for fqn, dp := range dps {
		if includeNS(c, cfg, dp.Namespace) && !cfg.ShouldExclude("replicaset", fqn) {
			res[fqn] = dp
		}
	}

	return res, nil
}

// ListAllReplicaSets fetch all ReplicaSets on the cluster.
func listAllReplicaSets(c *k8s.Client) (map[string]*appsv1.ReplicaSet, error) {
	ll, err := fetchReplicaSets(c)
	if err != nil {
		log.Debug().Err(err).Msg("ListAll")
		return nil, err
	}

	dps := make(map[string]*appsv1.ReplicaSet, len(ll.Items))
	for i := range ll.Items {
		dps[metaFQN(ll.Items[i].ObjectMeta)] = &ll.Items[i]
	}

	return dps, nil
}

// FetchReplicaSets retrieves all ReplicaSets on the cluster.
func fetchReplicaSets(c *k8s.Client) (*appsv1.ReplicaSetList, error) {
	return c.DialOrDie().AppsV1().ReplicaSets(c.ActiveNamespace()).List(metav1.ListOptions{})
}
