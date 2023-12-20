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

// Role represents a Role scruber.
type Role struct {
	client types.Connection
	*config.Config
	*issues.Collector

	*cache.Role
	*cache.ClusterRoleBinding
	*cache.RoleBinding
}

// NewRole return a new Role scruber.
func NewRole(ctx context.Context, c *Cache, codes *issues.Codes) Sanitizer {
	ro := Role{
		client:    c.factory.Client(),
		Config:    c.config,
		Collector: issues.NewCollector(codes, c.config),
	}

	var err error
	ro.Role, err = c.roles()
	if err != nil {
		ro.AddErr(ctx, err)
	}

	ro.ClusterRoleBinding, err = c.clusterrolebindings()
	if err != nil {
		ro.AddErr(ctx, err)
	}

	ro.RoleBinding, err = c.rolebindings()
	if err != nil {
		ro.AddErr(ctx, err)
	}

	return &ro
}

// Sanitize all available Roles.
func (c *Role) Sanitize(ctx context.Context) error {
	return sanitize.NewRole(c.Collector, c).Sanitize(ctx)
}
