package dag

import (
	"github.com/derailed/popeye/internal/k8s"
	"github.com/derailed/popeye/pkg/config"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ListLimitRanges list all included LimitRanges.
func ListLimitRanges(c *k8s.Client, cfg *config.Config) (map[string]*v1.LimitRange, error) {
	lrs, err := listAllLimitRanges(c)
	if err != nil {
		return nil, err
	}

	res := make(map[string]*v1.LimitRange, len(lrs))
	for fqn, lr := range lrs {
		if includeNS(c, lr.Namespace) {
			res[fqn] = lr
		}
	}

	return res, nil
}

// ListAllLimitRanges fetch all LimitRanges on the cluster.
func listAllLimitRanges(c *k8s.Client) (map[string]*v1.LimitRange, error) {
	ll, err := fetchLimitRanges(c)
	if err != nil {
		return nil, err
	}

	lrs := make(map[string]*v1.LimitRange, len(ll.Items))
	for i := range ll.Items {
		lrs[metaFQN(ll.Items[i].ObjectMeta)] = &ll.Items[i]
	}

	return lrs, nil
}

// fetchLimitRanges retrieves all LimitRanges on the cluster.
func fetchLimitRanges(c *k8s.Client) (*v1.LimitRangeList, error) {
	return c.DialOrDie().CoreV1().LimitRanges(c.ActiveNamespace()).List(metav1.ListOptions{})
}
