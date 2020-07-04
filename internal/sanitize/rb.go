package sanitize

import (
	"context"

	"github.com/derailed/popeye/internal"
	"github.com/derailed/popeye/internal/cache"
	"github.com/derailed/popeye/internal/issues"
)

type (
	// RBLister represents RB dependencies.
	RBLister interface {
		RoleBindingLister
		ClusterRoleLister
		RoleLister
	}

	// RoleBinding tracks RoleBinding sanitization.
	RoleBinding struct {
		*issues.Collector
		RBLister
	}
)

// NewRoleBinding returns a new  sanitizer.
func NewRoleBinding(c *issues.Collector, lister RBLister) *RoleBinding {
	return &RoleBinding{
		Collector: c,
		RBLister:  lister,
	}
}

// Sanitize cleanse the resource..
func (r *RoleBinding) Sanitize(ctx context.Context) error {
	for fqn, rb := range r.ListRoleBindings() {
		r.InitOutcome(fqn)
		ctx = internal.WithFQN(ctx, fqn)

		switch rb.RoleRef.Kind {
		case "ClusterRole":
			if _, ok := r.ListClusterRoles()[rb.RoleRef.Name]; !ok {
				r.AddCode(ctx, 1300, rb.RoleRef.Kind, rb.RoleRef.Name)
			}
		case "Role":
			rFQN := cache.FQN(rb.Namespace, rb.RoleRef.Name)
			if _, ok := r.ListRoles()[rFQN]; !ok {
				r.AddCode(ctx, 1300, rb.RoleRef.Kind, rFQN)
			}
		}

		if r.NoConcerns(fqn) && r.Config.ExcludeFQN(internal.MustExtractSectionGVR(ctx), fqn) {
			r.ClearOutcome(fqn)
		}
	}
	return nil
}
