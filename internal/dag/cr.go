package dag

import (
	"github.com/derailed/popeye/internal/k8s"
	"github.com/derailed/popeye/pkg/config"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ListClusterRoles list included ClusterRoles.
func ListClusterRoles(c *k8s.Client, cfg *config.Config) (map[string]*rbacv1.ClusterRole, error) {
	crs, err := listAllClusterRoles(c)
	if err != nil {
		return map[string]*rbacv1.ClusterRole{}, err
	}
	res := make(map[string]*rbacv1.ClusterRole, len(crs))
	for fqn, cr := range crs {
		res[fqn] = cr
	}

	return res, nil
}

// ListAllClusterRoles fetch all ClusterRoles on the cluster.
func listAllClusterRoles(c *k8s.Client) (map[string]*rbacv1.ClusterRole, error) {
	ll, err := fetchClusterRoles(c)
	if err != nil {
		return nil, err
	}

	crs := make(map[string]*rbacv1.ClusterRole, len(ll.Items))
	for i := range ll.Items {
		crs[metaFQN(ll.Items[i].ObjectMeta)] = &ll.Items[i]
	}

	return crs, nil
}

// FetchClusterRoles retrieves all ClusterRoles on the cluster.
func fetchClusterRoles(c *k8s.Client) (*rbacv1.ClusterRoleList, error) {
	return c.DialOrDie().RbacV1().ClusterRoles().List(metav1.ListOptions{})
}
