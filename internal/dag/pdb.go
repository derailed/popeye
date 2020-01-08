package dag

import (
	"github.com/derailed/popeye/internal/k8s"
	"github.com/derailed/popeye/pkg/config"
	pv1beta1 "k8s.io/api/policy/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ListPodDisruptionBudgets list all included PodDisruptionBudgets.
func ListPodDisruptionBudgets(c *k8s.Client, cfg *config.Config) (map[string]*pv1beta1.PodDisruptionBudget, error) {
	pdbs, err := listAllPodDisruptionBudgets(c)
	if err != nil {
		return nil, err
	}

	res := make(map[string]*pv1beta1.PodDisruptionBudget, len(pdbs))
	for fqn, pdb := range pdbs {
		if includeNS(c, pdb.Namespace) {
			res[fqn] = pdb
		}
	}

	return res, nil
}

// ListAllPodDisruptionBudgets fetch all PodDisruptionBudgets on the cluster.
func listAllPodDisruptionBudgets(c *k8s.Client) (map[string]*pv1beta1.PodDisruptionBudget, error) {
	ll, err := fetchPodDisruptionBudgets(c)
	if err != nil {
		return nil, err
	}

	pdbs := make(map[string]*pv1beta1.PodDisruptionBudget, len(ll.Items))
	for i := range ll.Items {
		pdbs[metaFQN(ll.Items[i].ObjectMeta)] = &ll.Items[i]
	}

	return pdbs, nil
}

// fetchPodDisruptionBudgets retrieves all PodDisruptionBudgets on the cluster.
func fetchPodDisruptionBudgets(c *k8s.Client) (*pv1beta1.PodDisruptionBudgetList, error) {
	return c.DialOrDie().PolicyV1beta1().PodDisruptionBudgets(c.ActiveNamespace()).List(metav1.ListOptions{})
}
