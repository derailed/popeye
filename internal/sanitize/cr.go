package sanitize

import (
	"context"
	"sync"

	"github.com/derailed/popeye/internal"
	"github.com/derailed/popeye/internal/cache"
	"github.com/derailed/popeye/internal/issues"
)

type (
	// CRLister lists roles and rolebindings.
	CRLister interface {
		ClusterRoleLister
		ClusterRoleBindingLister
		RoleBindingLister
	}

	// ClusterRole tracks ClusterRole sanitization.
	ClusterRole struct {
		*issues.Collector
		CRLister
	}
)

// NewClusterRole returns a new ClusterRole sanitizer.
func NewClusterRole(c *issues.Collector, lister CRLister) *ClusterRole {
	return &ClusterRole{
		Collector: c,
		CRLister:  lister,
	}
}

// Sanitize a configmap.
func (c *ClusterRole) Sanitize(ctx context.Context) error {
	var crRefs sync.Map
	c.ClusterRoleRefs(&crRefs)
	c.RoleRefs(&crRefs)
	c.checkInUse(ctx, &crRefs)

	return nil
}

func (c *ClusterRole) checkInUse(ctx context.Context, refs *sync.Map) {
	for fqn := range c.ListClusterRoles() {
		c.InitOutcome(fqn)
		ctx = internal.WithFQN(ctx, fqn)

		_, ok := refs.Load(cache.ResFqn(cache.ClusterRoleKey, fqn))
		if !ok {
			c.AddCode(ctx, 400)
		}
		if c.NoConcerns(fqn) && c.Config.ExcludeFQN(internal.MustExtractSectionGVR(ctx), fqn) {
			c.ClearOutcome(fqn)
		}
	}
}
