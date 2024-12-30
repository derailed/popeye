// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Popeye

package scrub

import (
	"context"

	"github.com/derailed/popeye/internal"
	"github.com/derailed/popeye/internal/db"
	"github.com/derailed/popeye/internal/issues"
	"github.com/derailed/popeye/internal/lint"
	v1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
)

// ClusterRole represents a ClusterRole scruber.
type ClusterRole struct {
	*issues.Collector
	*Cache
}

// NewClusterRole returns a new instance.
func NewClusterRole(_ context.Context, c *Cache, codes *issues.Codes) Linter {
	return &ClusterRole{
		Collector: issues.NewCollector(codes, c.Config),
		Cache:     c,
	}
}

func (s *ClusterRole) Preloads() Preloads {
	return Preloads{
		internal.CR:  db.LoadResource[*rbacv1.ClusterRole],
		internal.CRB: db.LoadResource[*rbacv1.ClusterRoleBinding],
		internal.RO:  db.LoadResource[*rbacv1.Role],
		internal.SA:  db.LoadResource[*v1.ServiceAccount],
	}
}

// Lint all available ClusterRoles.
func (s *ClusterRole) Lint(ctx context.Context) error {
	for k, f := range s.Preloads() {
		if err := f(ctx, s.Loader, internal.Glossary[k]); err != nil {
			return err
		}
	}

	return lint.NewClusterRole(s.Collector, s.DB).Lint(ctx)
}
