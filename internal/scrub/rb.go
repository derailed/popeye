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

// RoleBinding represents a RoleBinding scruber.
type RoleBinding struct {
	*issues.Collector
	*Cache
}

// NewRoleBinding returns a new instance.
func NewRoleBinding(_ context.Context, c *Cache, codes *issues.Codes) Linter {
	return &RoleBinding{
		Collector: issues.NewCollector(codes, c.Config),
		Cache:     c,
	}
}

func (s *RoleBinding) Preloads() Preloads {
	return Preloads{
		internal.ROB: db.LoadResource[*rbacv1.RoleBinding],
		internal.RO:  db.LoadResource[*rbacv1.Role],
		internal.CR:  db.LoadResource[*rbacv1.ClusterRole],
		internal.CRB: db.LoadResource[*rbacv1.ClusterRoleBinding],
		internal.SA:  db.LoadResource[*v1.ServiceAccount],
	}
}

// Lint all available RoleBindings.
func (s *RoleBinding) Lint(ctx context.Context) error {
	for k, f := range s.Preloads() {
		if err := f(ctx, s.Loader, internal.Glossary[k]); err != nil {
			return err
		}
	}

	return lint.NewRoleBinding(s.Collector, s.DB).Lint(ctx)
}
