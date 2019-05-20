package cache

import (
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
