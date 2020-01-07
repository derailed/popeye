package sanitize

import (
	"context"

	"github.com/derailed/popeye/internal"
	"github.com/derailed/popeye/internal/issues"
	rbacv1 "k8s.io/api/rbac/v1"
)

type (
	ClusterRoleLister interface {
		ListClusterRoles() map[string]*rbacv1.ClusterRole
	}

	RoleLister interface {
		ListRoles() map[string]*rbacv1.Role
	}

	CRBLister interface {
		ClusterRoleBindingLister
		ClusterRoleLister
		RoleLister
	}

	// ClusterRoleBinding tracks ClusterRoleBinding sanitization.
	ClusterRoleBinding struct {
		*issues.Collector
		CRBLister
	}
)

// NewClusterRoleBinding returns a new ClusterRoleBinding sanitizer.
func NewClusterRoleBinding(c *issues.Collector, lister CRBLister) *ClusterRoleBinding {
	return &ClusterRoleBinding{
		Collector: c,
		CRBLister: lister,
	}
}

// Sanitize a configmap.
func (c *ClusterRoleBinding) Sanitize(ctx context.Context) error {
	c.checkInUse(ctx)

	return nil
}

func (c *ClusterRoleBinding) checkInUse(ctx context.Context) {
	for fqn, crb := range c.ListClusterRoleBindings() {
		c.InitOutcome(fqn)
		ctx = internal.WithFQN(ctx, fqn)

		switch crb.RoleRef.Kind {
		case "ClusterRole":
			if _, ok := c.ListClusterRoles()[crb.RoleRef.Name]; !ok {
				c.AddCode(ctx, 1300, crb.RoleRef.Kind, crb.RoleRef.Name)
			}
		case "Role":
			if _, ok := c.ListRoles()[crb.RoleRef.Name]; !ok {
				c.AddCode(ctx, 1300, crb.RoleRef.Kind, crb.RoleRef.Name)
			}
		}

		if c.NoConcerns(fqn) && c.Config.ExcludeFQN(internal.MustExtractSection(ctx), fqn) {
			c.ClearOutcome(fqn)
		}
	}
}
