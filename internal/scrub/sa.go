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
	netv1 "k8s.io/api/networking/v1"
	rbacv1 "k8s.io/api/rbac/v1"
)

// ServiceAccount represents a ServiceAccount scruber.
type ServiceAccount struct {
	*issues.Collector
	*Cache
}

// NewServiceAccount returns a new instance.
func NewServiceAccount(_ context.Context, c *Cache, codes *issues.Codes) Linter {
	return &ServiceAccount{
		Collector: issues.NewCollector(codes, c.Config),
		Cache:     c,
	}
}

func (s *ServiceAccount) Preloads() Preloads {
	return Preloads{
		internal.SA:  db.LoadResource[*v1.ServiceAccount],
		internal.PO:  db.LoadResource[*v1.Pod],
		internal.ROB: db.LoadResource[*rbacv1.RoleBinding],
		internal.CRB: db.LoadResource[*rbacv1.ClusterRoleBinding],
		internal.SEC: db.LoadResource[*v1.Secret],
		internal.ING: db.LoadResource[*netv1.Ingress],
	}
}

// Lint all available ServiceAccounts.
func (s *ServiceAccount) Lint(ctx context.Context) error {
	for k, f := range s.Preloads() {
		if err := f(ctx, s.Loader, internal.Glossary[k]); err != nil {
			return err
		}
	}

	return lint.NewServiceAccount(s.Collector, s.DB).Lint(ctx)
}
