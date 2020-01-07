package dag

import (
	"github.com/derailed/popeye/internal/k8s"
	"github.com/derailed/popeye/pkg/config"
	nv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ListNetworkPolicies list all included NetworkPolicies.
func ListNetworkPolicies(c *k8s.Client, cfg *config.Config) (map[string]*nv1.NetworkPolicy, error) {
	dps, err := listAllNetworkPolicies(c)
	if err != nil {
		return map[string]*nv1.NetworkPolicy{}, err
	}

	res := make(map[string]*nv1.NetworkPolicy, len(dps))
	for fqn, dp := range dps {
		if includeNS(c, dp.Namespace) {
			res[fqn] = dp
		}
	}

	return res, nil
}

// ListAllNetworkPolicies fetch all NetworkPolicies on the cluster.
func listAllNetworkPolicies(c *k8s.Client) (map[string]*nv1.NetworkPolicy, error) {
	ll, err := fetchNetworkPolicies(c)
	if err != nil {
		return nil, err
	}

	dps := make(map[string]*nv1.NetworkPolicy, len(ll.Items))
	for i := range ll.Items {
		dps[metaFQN(ll.Items[i].ObjectMeta)] = &ll.Items[i]
	}

	return dps, nil
}

// FetchNetworkPolicies retrieves all NetworkPolicies on the cluster.
func fetchNetworkPolicies(c *k8s.Client) (*nv1.NetworkPolicyList, error) {
	return c.DialOrDie().NetworkingV1().NetworkPolicies(c.ActiveNamespace()).List(metav1.ListOptions{})
}
