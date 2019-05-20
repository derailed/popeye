package cache

import (
	rbacv1 "k8s.io/api/rbac/v1"
)

// RoleBinding represents RoleBinding cache.
type RoleBinding struct {
	crbs map[string]*rbacv1.RoleBinding
}

// NewRoleBinding returns a new RoleBinding cache.
func NewRoleBinding(crbs map[string]*rbacv1.RoleBinding) *RoleBinding {
	return &RoleBinding{crbs: crbs}
}

// ListRoleBindings returns all available RoleBindings on the cluster.
func (r *RoleBinding) ListRoleBindings() map[string]*rbacv1.RoleBinding {
	return r.crbs
}
