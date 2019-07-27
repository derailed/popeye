package dag

import (
	"fmt"

	"github.com/derailed/popeye/internal/k8s"
	"github.com/derailed/popeye/pkg/config"
	"github.com/rs/zerolog/log"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ListDeployments list all included Deployments.
func ListDeployments(c *k8s.Client, cfg *config.Config) (map[string]*appsv1.Deployment, error) {
	dps, err := listAllDeployments(c)
	if err != nil {
		return map[string]*appsv1.Deployment{}, err
	}

	res := make(map[string]*appsv1.Deployment, len(dps))
	for fqn, dp := range dps {
		if includeNS(c, cfg, dp.Namespace) && !cfg.ShouldExclude("deployment", fqn) {
			res[fqn] = dp
		}
	}

	return res, nil
}

// ListAllDeployments fetch all Deployments on the cluster.
func listAllDeployments(c *k8s.Client) (map[string]*appsv1.Deployment, error) {
	ll, err := fetchDeployments(c)
	if err != nil {
		log.Debug().Err(err).Msg("ListAll")
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
	return c.DialOrDie().AppsV1().Deployments(c.ActiveNamespace()).List(metav1.ListOptions{})
}

func preferredRev(c *k8s.Client, group string) (string, error) {
	apiGroups, err := c.DialOrDie().Discovery().ServerGroups()
	if err != nil {
		return "", err
	}

	for _, grp := range apiGroups.Groups {
		if grp.Name != group {
			continue
		}
		return grp.PreferredVersion.GroupVersion, nil
	}

	return "", fmt.Errorf("No matching API group %s", group)
}
