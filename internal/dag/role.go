package dag

import (
	"github.com/derailed/popeye/internal/k8s"
	"github.com/derailed/popeye/pkg/config"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ListRoles list included Roles.
func ListRoles(c *k8s.Client, cfg *config.Config) (map[string]*rbacv1.Role, error) {
	ros, err := listAllRoles(c)
	if err != nil {
		return nil, err
	}

	res := make(map[string]*rbacv1.Role, len(ros))
	for fqn, ro := range ros {
		if includeNS(c, ro.Namespace) {
			res[fqn] = ro
		}
	}

	return res, nil
}

// ListAllRoles fetch all Roles on the cluster.
func listAllRoles(c *k8s.Client) (map[string]*rbacv1.Role, error) {
	ll, err := fetchRoles(c)
	if err != nil {
		return nil, err
	}

	ros := make(map[string]*rbacv1.Role, len(ll.Items))
	for i := range ll.Items {
		ros[metaFQN(ll.Items[i].ObjectMeta)] = &ll.Items[i]
	}

	return ros, nil
}

// FetchRoleBindings retrieves all RoleBindings on the cluster.
func fetchRoles(c *k8s.Client) (*rbacv1.RoleList, error) {
	return c.DialOrDie().RbacV1().Roles(c.ActiveNamespace()).List(metav1.ListOptions{})
}
