package sanitize

import (
	"context"
	"sync"

	"github.com/derailed/popeye/internal"
	"github.com/derailed/popeye/internal/cache"
	"github.com/derailed/popeye/internal/issues"
)

type (
	// ROLister list out roles and deps.
	ROLister interface {
		RoleLister
		ClusterRoleBindingLister
		RoleBindingLister
	}

	// Role tracks Role sanitization.
	Role struct {
		*issues.Collector
		ROLister
	}
)

// NewRole returns a new sanitizer.
func NewRole(c *issues.Collector, lister ROLister) *Role {
	return &Role{
		Collector: c,
		ROLister:  lister,
	}
}

// Sanitize cleanse the resource.
func (r *Role) Sanitize(ctx context.Context) error {
	var roRefs sync.Map
	r.ClusterRoleRefs(&roRefs)
	r.RoleRefs(&roRefs)
	r.checkInUse(ctx, &roRefs)

	return nil
}

func (r *Role) checkInUse(ctx context.Context, refs *sync.Map) {
	for fqn := range r.ListRoles() {
		r.InitOutcome(fqn)
		ctx = internal.WithFQN(ctx, fqn)

		_, ok := refs.Load(cache.ResFqn(cache.RoleKey, fqn))
		if !ok {
			r.AddCode(ctx, 400)
		}

		if r.NoConcerns(fqn) && r.Config.ExcludeFQN(internal.MustExtractSectionGVR(ctx), fqn) {
			r.ClearOutcome(fqn)
		}
	}
}
