package dag

import (
	"github.com/derailed/popeye/internal/k8s"
	"github.com/derailed/popeye/pkg/config"
	"github.com/rs/zerolog/log"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ListDaemonSets list all included DaemonSets.
func ListDaemonSets(c *k8s.Client, cfg *config.Config) (map[string]*appsv1.DaemonSet, error) {
	dps, err := listAllDaemonSets(c)
	if err != nil {
		return map[string]*appsv1.DaemonSet{}, err
	}

	res := make(map[string]*appsv1.DaemonSet, len(dps))
	for fqn, dp := range dps {
		if includeNS(c, cfg, dp.Namespace) && !cfg.ShouldExclude("daemonset", fqn) {
			res[fqn] = dp
		}
	}

	return res, nil
}

// ListAllDaemonSets fetch all DaemonSets on the cluster.
func listAllDaemonSets(c *k8s.Client) (map[string]*appsv1.DaemonSet, error) {
	ll, err := fetchDaemonSets(c)
	if err != nil {
		log.Debug().Err(err).Msg("ListAll")
		return nil, err
	}

	dps := make(map[string]*appsv1.DaemonSet, len(ll.Items))
	for i := range ll.Items {
		dps[metaFQN(ll.Items[i].ObjectMeta)] = &ll.Items[i]
	}

	return dps, nil
}

// FetchDaemonSets retrieves all DaemonSets on the cluster.
func fetchDaemonSets(c *k8s.Client) (*appsv1.DaemonSetList, error) {
	return c.DialOrDie().AppsV1().DaemonSets(c.ActiveNamespace()).List(metav1.ListOptions{})
}
