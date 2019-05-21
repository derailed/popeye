package dag

import (
	"github.com/derailed/popeye/internal/k8s"
	"github.com/derailed/popeye/pkg/config"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ListServiceAccounts list included ServiceAccounts.
func ListServiceAccounts(c *k8s.Client, cfg *config.Config) (map[string]*v1.ServiceAccount, error) {
	sas, err := listAllServiceAccounts(c)
	if err != nil {
		return map[string]*v1.ServiceAccount{}, err
	}

	res := make(map[string]*v1.ServiceAccount, len(sas))
	for fqn, sa := range sas {
		if includeNS(c, cfg, sa.Namespace) {
			res[fqn] = sa
		}
	}

	return res, nil
}

// ListAllServiceAccounts fetch all ServiceAccounts on the cluster.
func listAllServiceAccounts(c *k8s.Client) (map[string]*v1.ServiceAccount, error) {
	ll, err := fetchServiceAccounts(c)
	if err != nil {
		return nil, err
	}

	sas := make(map[string]*v1.ServiceAccount, len(ll.Items))
	for i := range ll.Items {
		sas[metaFQN(ll.Items[i].ObjectMeta)] = &ll.Items[i]
	}

	return sas, nil
}

// FetchServiceAccounts retrieves all ServiceAccounts on the cluster.
func fetchServiceAccounts(c *k8s.Client) (*v1.ServiceAccountList, error) {
	return c.DialOrDie().CoreV1().ServiceAccounts("").List(metav1.ListOptions{})
}
