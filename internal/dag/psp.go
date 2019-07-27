package dag

import (
	"github.com/derailed/popeye/internal/k8s"
	"github.com/derailed/popeye/pkg/config"
	"github.com/rs/zerolog/log"
	pv1beta1 "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ListPodSecurityPolicies list all included PodSecurityPolicies.
func ListPodSecurityPolicies(c *k8s.Client, cfg *config.Config) (map[string]*pv1beta1.PodSecurityPolicy, error) {
	dps, err := listAllPodSecurityPolicys(c)
	if err != nil {
		return map[string]*pv1beta1.PodSecurityPolicy{}, err
	}

	res := make(map[string]*pv1beta1.PodSecurityPolicy, len(dps))
	for fqn, dp := range dps {
		if includeNS(c, cfg, dp.Namespace) && !cfg.ShouldExclude("deployment", fqn) {
			res[fqn] = dp
		}
	}

	return res, nil
}

// ListAllPodSecurityPolicys fetch all PodSecurityPolicys on the cluster.
func listAllPodSecurityPolicys(c *k8s.Client) (map[string]*pv1beta1.PodSecurityPolicy, error) {
	ll, err := fetchPodSecurityPolicys(c)
	if err != nil {
		log.Debug().Err(err).Msg("ListAll")
		return nil, err
	}

	dps := make(map[string]*pv1beta1.PodSecurityPolicy, len(ll.Items))
	for i := range ll.Items {
		dps[metaFQN(ll.Items[i].ObjectMeta)] = &ll.Items[i]
	}

	return dps, nil
}

// FetchPodSecurityPolicys retrieves all PodSecurityPolicys on the cluster.
func fetchPodSecurityPolicys(c *k8s.Client) (*pv1beta1.PodSecurityPolicyList, error) {
	return c.DialOrDie().ExtensionsV1beta1().PodSecurityPolicies().List(metav1.ListOptions{})
}
