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

// ClusterRoleBinding represents a ClusterRoleBinding scruber.
type ClusterRoleBinding struct {
	*issues.Collector
	*Cache
}

// NewClusterRoleBinding returns a new instance.
func NewClusterRoleBinding(_ context.Context, c *Cache, codes *issues.Codes) Linter {
	return &ClusterRoleBinding{
		Collector: issues.NewCollector(codes, c.Config),
		Cache:     c,
	}
}

func (s *ClusterRoleBinding) Preloads() Preloads {
	return Preloads{
		internal.CRB: db.LoadResource[*rbacv1.ClusterRoleBinding],
		internal.CR:  db.LoadResource[*rbacv1.ClusterRole],
		internal.RO:  db.LoadResource[*rbacv1.Role],
		internal.SA:  db.LoadResource[*v1.ServiceAccount],
	}
}

// Lint all available ClusterRoleBindings.
func (s *ClusterRoleBinding) Lint(ctx context.Context) error {
	for k, f := range s.Preloads() {
		if err := f(ctx, s.Loader, internal.Glossary[k]); err != nil {
			return err
		}
	}

	return lint.NewClusterRoleBinding(s.Collector, s.DB).Lint(ctx)
}
