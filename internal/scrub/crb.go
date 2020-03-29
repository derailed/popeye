package scrub

import (
	"context"

	"github.com/derailed/popeye/internal/cache"
	"github.com/derailed/popeye/internal/issues"
	"github.com/derailed/popeye/internal/sanitize"
	"github.com/derailed/popeye/pkg/config"
	"github.com/derailed/popeye/types"
)

// ClusterRoleBinding represents a ClusterRoleBinding scruber.
type ClusterRoleBinding struct {
	client types.Connection
	*config.Config
	*issues.Collector

	*cache.ClusterRoleBinding
	*cache.ClusterRole
	*cache.Role
}

// NewClusterRoleBinding return a new ClusterRoleBinding scruber.
func NewClusterRoleBinding(ctx context.Context, c *Cache, codes *issues.Codes) Sanitizer {
	crb := ClusterRoleBinding{
		client:    c.factory.Client(),
		Config:    c.config,
		Collector: issues.NewCollector(codes, c.config),
	}

	var err error
	crb.ClusterRoleBinding, err = c.clusterrolebindings()
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

// Sanitize all available ClusterRoleBindings.
func (c *ClusterRoleBinding) Sanitize(ctx context.Context) error {
	return sanitize.NewClusterRoleBinding(c.Collector, c).Sanitize(ctx)
}
