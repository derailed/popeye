package dag

import (
	"github.com/derailed/popeye/internal/k8s"
	"github.com/derailed/popeye/pkg/config"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ListRoleBindings list included RoleBindings.
func ListRoleBindings(c *k8s.Client, cfg *config.Config) (map[string]*rbacv1.RoleBinding, error) {
	sas, err := listAllRoleBindings(c)
	if err != nil {
		return nil, err
	}

	res := make(map[string]*rbacv1.RoleBinding, len(sas))
	for fqn, sa := range sas {
		if c.IsActiveNamespace(sa.Namespace) && !cfg.ExcludedNS(sa.Namespace) {
			res[fqn] = sa
		}
	}

	return res, nil
}

// ListAllRoleBindings fetch all RoleBindings on the cluster.
func listAllRoleBindings(c *k8s.Client) (map[string]*rbacv1.RoleBinding, error) {
	ll, err := fetchRoleBindings(c)
	if err != nil {
		return nil, err
	}

	sas := make(map[string]*rbacv1.RoleBinding, len(ll.Items))
	for i := range ll.Items {
		sas[MetaFQN(ll.Items[i].ObjectMeta)] = &ll.Items[i]
	}

	return sas, nil
}

// FetchRoleBindings retrieves all RoleBindings on the cluster.
func fetchRoleBindings(c *k8s.Client) (*rbacv1.RoleBindingList, error) {
	return c.DialOrDie().RbacV1().RoleBindings("").List(metav1.ListOptions{})
}
