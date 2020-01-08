package cache

import (
	rbacv1 "k8s.io/api/rbac/v1"
)

// ClusterRoleKey tracks ClusterRole resource references
const ClusterRoleKey = "clusterrole"

// ClusterRole represents ClusterRole cache.
type ClusterRole struct {
	crs map[string]*rbacv1.ClusterRole
}

// NewClusterRole returns a new ClusterRole cache.
func NewClusterRole(crs map[string]*rbacv1.ClusterRole) *ClusterRole {
	return &ClusterRole{crs: crs}
}

// ListClusterRoles returns all available ClusterRoles on the cluster.
func (c *ClusterRole) ListClusterRoles() map[string]*rbacv1.ClusterRole {
	return c.crs
}
