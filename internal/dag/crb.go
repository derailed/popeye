package dag

import (
	"github.com/derailed/popeye/internal/k8s"
	"github.com/derailed/popeye/pkg/config"
	"github.com/rs/zerolog/log"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ListClusterRoleBindings list included ClusterRoleBindings.
func ListClusterRoleBindings(c *k8s.Client, cfg *config.Config) (map[string]*rbacv1.ClusterRoleBinding, error) {
	crbs, err := listAllClusterRoleBindings(c)
	if err != nil {
		return map[string]*rbacv1.ClusterRoleBinding{}, err
	}
	res := make(map[string]*rbacv1.ClusterRoleBinding, len(crbs))
	for fqn, crb := range crbs {
		if !cfg.ShouldExclude("clusterrolebinding", fqn) {
			res[fqn] = crb
		}
	}

	return res, nil
}

// ListAllClusterRoleBindings fetch all ClusterRoleBindings on the cluster.
func listAllClusterRoleBindings(c *k8s.Client) (map[string]*rbacv1.ClusterRoleBinding, error) {
	ll, err := fetchClusterRoleBindings(c)
	if err != nil {
		log.Debug().Err(err).Msg("ListAll")
		return nil, err
	}

	crbs := make(map[string]*rbacv1.ClusterRoleBinding, len(ll.Items))
	for i := range ll.Items {
		crbs[metaFQN(ll.Items[i].ObjectMeta)] = &ll.Items[i]
	}

	return crbs, nil
}

// FetchClusterRoleBindings retrieves all ClusterRoleBindings on the cluster.
func fetchClusterRoleBindings(c *k8s.Client) (*rbacv1.ClusterRoleBindingList, error) {
	return c.DialOrDie().RbacV1().ClusterRoleBindings().List(metav1.ListOptions{})
}
