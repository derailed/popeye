package scrub

import (
	"context"

	"github.com/derailed/popeye/internal/cache"
	"github.com/derailed/popeye/internal/issues"
	"github.com/derailed/popeye/internal/k8s"
	"github.com/derailed/popeye/internal/sanitize"
	"github.com/derailed/popeye/pkg/config"
)

// ClusterRole represents a ClusterRole scruber.
type ClusterRole struct {
	client *k8s.Client
	*config.Config
	*issues.Collector

	*cache.ClusterRole
	*cache.ClusterRoleBinding
	*cache.RoleBinding
}

// NewClusterRole return a new ClusterRole scruber.
func NewClusterRole(ctx context.Context, c *Cache, codes *issues.Codes) Sanitizer {
	crb := ClusterRole{
		client:    c.client,
		Config:    c.config,
		Collector: issues.NewCollector(codes, c.config),
	}

	var err error
	crb.ClusterRole, err = c.clusterroles()
	if err != nil {
		crb.AddErr(ctx, err)
	}

	crb.ClusterRoleBinding, err = c.clusterrolebindings()
	if err != nil {
		crb.AddCode(ctx, 402, err)
	}

	crb.RoleBinding, err = c.rolebindings()
	if err != nil {
		crb.AddErr(ctx, err)
	}

	return &crb
}

// Sanitize all available ClusterRoles.
func (c *ClusterRole) Sanitize(ctx context.Context) error {
	return sanitize.NewClusterRole(c.Collector, c).Sanitize(ctx)
}
