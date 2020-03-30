package cache

import (
	"strings"
	"sync"

	"github.com/derailed/popeye/internal"
	rbacv1 "k8s.io/api/rbac/v1"
)

// ClusterRoleBinding represents ClusterRoleBinding cache.
type ClusterRoleBinding struct {
	crbs map[string]*rbacv1.ClusterRoleBinding
}

// NewClusterRoleBinding returns a new ClusterRoleBinding cache.
func NewClusterRoleBinding(crbs map[string]*rbacv1.ClusterRoleBinding) *ClusterRoleBinding {
	return &ClusterRoleBinding{crbs: crbs}
}

// ListClusterRoleBindings returns all available ClusterRoleBindings on the cluster.
func (c *ClusterRoleBinding) ListClusterRoleBindings() map[string]*rbacv1.ClusterRoleBinding {
	return c.crbs
}

// ClusterRoleRefs computes all clusterrole external references.
func (c *ClusterRoleBinding) ClusterRoleRefs(refs *sync.Map) {
	for fqn, crb := range c.crbs {
		key := ResFqn(strings.ToLower(crb.RoleRef.Kind), FQN(crb.Namespace, crb.RoleRef.Name))
		if c, ok := refs.Load(key); ok {
			c.(internal.StringSet).Add(fqn)
		} else {
			refs.Store(key, internal.StringSet{fqn: internal.Blank})
		}
	}
}
