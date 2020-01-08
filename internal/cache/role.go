package cache

import (
	rbacv1 "k8s.io/api/rbac/v1"
)

// Role represents Role cache.
type Role struct {
	ros map[string]*rbacv1.Role
}

// NewRole returns a new Role cache.
func NewRole(ros map[string]*rbacv1.Role) *Role {
	return &Role{ros: ros}
}

// ListRoles returns all available Roles on the cluster.
func (r *Role) ListRoles() map[string]*rbacv1.Role {
	return r.ros
}
