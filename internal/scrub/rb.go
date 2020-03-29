package scrub

import (
	"context"

	"github.com/derailed/popeye/internal/cache"
	"github.com/derailed/popeye/internal/issues"
	"github.com/derailed/popeye/internal/sanitize"
	"github.com/derailed/popeye/pkg/config"
	"github.com/derailed/popeye/types"
)

// RoleBinding represents a RoleBinding scruber.
type RoleBinding struct {
	client types.Connection
	*config.Config
	*issues.Collector

	*cache.RoleBinding
	*cache.ClusterRole
	*cache.Role
}

// NewRoleBinding return a new RoleBinding scruber.
func NewRoleBinding(ctx context.Context, c *Cache, codes *issues.Codes) Sanitizer {
	crb := RoleBinding{
		client:    c.factory.Client(),
		Config:    c.config,
		Collector: issues.NewCollector(codes, c.config),
	}

	var err error
	crb.RoleBinding, err = c.rolebindings()
	if err != nil {
		crb.AddErr(ctx, err)
	}

	crb.ClusterRole, err = c.clusterroles()
	if err != nil {
		crb.AddCode(ctx, 402, err)
	}

	crb.Role, err = c.roles()
	if err != nil {
		crb.AddErr(ctx, err)
	}

	return &crb
}

// Sanitize all available RoleBindings.
func (c *RoleBinding) Sanitize(ctx context.Context) error {
	return sanitize.NewRoleBinding(c.Collector, c).Sanitize(ctx)
}
