package dag

import (
	"github.com/derailed/popeye/internal/k8s"
	"github.com/derailed/popeye/pkg/config"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ListDeployments list all included Deployments.
func ListDeployments(c *k8s.Client, cfg *config.Config) (map[string]*appsv1.Deployment, error) {
	dps, err := listAllDeployments(c)
	if err != nil {
		return nil, err
	}

	res := make(map[string]*appsv1.Deployment, len(dps))
	for fqn, dp := range dps {
		if c.IsActiveNamespace(dp.Namespace) && !cfg.ExcludedNS(dp.Namespace) {
			res[fqn] = dp
		}
	}

	return res, nil
}

// ListAllDeployments fetch all Deployments on the cluster.
func listAllDeployments(c *k8s.Client) (map[string]*appsv1.Deployment, error) {
	ll, err := fetchDeployments(c)
	if err != nil {
		return nil, err
	}

	dps := make(map[string]*appsv1.Deployment, len(ll.Items))
	for i := range ll.Items {
		dps[metaFQN(ll.Items[i].ObjectMeta)] = &ll.Items[i]
	}

	return dps, nil
}

// FetchDeployments retrieves all Deployments on the cluster.
func fetchDeployments(c *k8s.Client) (*appsv1.DeploymentList, error) {
	return c.DialOrDie().AppsV1().Deployments("").List(metav1.ListOptions{})
}
