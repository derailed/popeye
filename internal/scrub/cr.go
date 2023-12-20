// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Popeye

package scrub

import (
	"context"

	"github.com/derailed/popeye/internal/cache"
	"github.com/derailed/popeye/internal/issues"
	"github.com/derailed/popeye/internal/sanitize"
	"github.com/derailed/popeye/pkg/config"
	"github.com/derailed/popeye/types"
)

// ClusterRole represents a ClusterRole scruber.
type ClusterRole struct {
	client types.Connection
	*config.Config
	*issues.Collector

	*cache.ClusterRole
	*cache.ClusterRoleBinding
	*cache.RoleBinding
}

// NewClusterRole return a new ClusterRole scruber.
func NewClusterRole(ctx context.Context, c *Cache, codes *issues.Codes) Sanitizer {
	cr := ClusterRole{
		client:    c.factory.Client(),
		Config:    c.config,
		Collector: issues.NewCollector(codes, c.config),
	}

	var err error
	cr.ClusterRole, err = c.clusterroles()
	if err != nil {
		cr.AddErr(ctx, err)
	}

	cr.ClusterRoleBinding, err = c.clusterrolebindings()
	if err != nil {
		cr.AddErr(ctx, err)
	}

	cr.RoleBinding, err = c.rolebindings()
	if err != nil {
		cr.AddErr(ctx, err)
	}

	return &cr
}

// Sanitize all available ClusterRoles.
func (c *ClusterRole) Sanitize(ctx context.Context) error {
	return sanitize.NewClusterRole(c.Collector, c).Sanitize(ctx)
}
