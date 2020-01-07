package cache

import (
	"strings"

	"github.com/derailed/popeye/internal"
	rbacv1 "k8s.io/api/rbac/v1"
)

const RoleKey = "role"

// RoleBinding represents RoleBinding cache.
type RoleBinding struct {
	rbs map[string]*rbacv1.RoleBinding
}

// NewRoleBinding returns a new RoleBinding cache.
func NewRoleBinding(rbs map[string]*rbacv1.RoleBinding) *RoleBinding {
	return &RoleBinding{rbs: rbs}
}

// ListRoleBindings returns all available RoleBindings on the cluster.
func (r *RoleBinding) ListRoleBindings() map[string]*rbacv1.RoleBinding {
	return r.rbs
}

// RoleRefs computes all role external references.
func (c *RoleBinding) RoleRefs(refs ObjReferences) {
	for fqn, rb := range c.rbs {
		key := ResFqn(strings.ToLower(rb.RoleRef.Kind), FQN(rb.Namespace, rb.RoleRef.Name))
		if c, ok := refs[key]; ok {
			c.Add(fqn)
		} else {
			refs[key] = internal.StringSet{fqn: internal.Blank}
		}
	}
}
