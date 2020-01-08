package scrub

import (
	"github.com/derailed/popeye/internal/cache"
	"github.com/derailed/popeye/internal/dag"
)

type rbac struct {
	*dial

	crb *cache.ClusterRoleBinding
	cr  *cache.ClusterRole
	rb  *cache.RoleBinding
	ro  *cache.Role
}

func newRBAC(d *dial) *rbac {
	return &rbac{dial: d}
}

func (r *rbac) roles() (*cache.Role, error) {
	if r.ro != nil {
		return r.ro, nil
	}
	ros, err := dag.ListRoles(r.client, r.config)
	r.ro = cache.NewRole(ros)

	return r.ro, err
}

func (r *rbac) rolebindings() (*cache.RoleBinding, error) {
	if r.rb != nil {
		return r.rb, nil
	}
	rbs, err := dag.ListRoleBindings(r.client, r.config)
	r.rb = cache.NewRoleBinding(rbs)

	return r.rb, err
}

func (r *rbac) clusterroles() (*cache.ClusterRole, error) {
	if r.cr != nil {
		return r.cr, nil
	}
	crs, err := dag.ListClusterRoles(r.client, r.config)
	r.cr = cache.NewClusterRole(crs)

	return r.cr, err
}

func (r *rbac) clusterrolebindings() (*cache.ClusterRoleBinding, error) {
	if r.crb != nil {
		return r.crb, nil
	}
	crbs, err := dag.ListClusterRoleBindings(r.client, r.config)
	r.crb = cache.NewClusterRoleBinding(crbs)

	return r.crb, err
}
