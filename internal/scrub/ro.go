// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Popeye

package scrub

import (
	"context"

	"github.com/derailed/popeye/internal"
	"github.com/derailed/popeye/internal/db"
	"github.com/derailed/popeye/internal/issues"
	"github.com/derailed/popeye/internal/lint"
	rbacv1 "k8s.io/api/rbac/v1"
)

// Role represents a Role scruber.
type Role struct {
	*issues.Collector
	*Cache
}

// NewRole returns a new instance.
func NewRole(_ context.Context, c *Cache, codes *issues.Codes) Linter {
	return &Role{
		Collector: issues.NewCollector(codes, c.Config),
		Cache:     c,
	}
}

func (s *Role) Preloads() Preloads {
	return Preloads{
		internal.RO:  db.LoadResource[*rbacv1.Role],
		internal.ROB: db.LoadResource[*rbacv1.RoleBinding],
		internal.CRB: db.LoadResource[*rbacv1.ClusterRoleBinding],
	}
}

// Lint all available Roles.
func (s *Role) Lint(ctx context.Context) error {
	for k, f := range s.Preloads() {
		if err := f(ctx, s.Loader, internal.Glossary[k]); err != nil {
			return err
		}
	}

	return lint.NewRole(s.Collector, s.DB).Lint(ctx)
}
