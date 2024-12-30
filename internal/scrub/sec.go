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
)

// Secret represents a Secret scruber.
type Secret struct {
	*issues.Collector
	*Cache
}

// NewSecret return a new Secret scruber.
func NewSecret(_ context.Context, c *Cache, codes *issues.Codes) Linter {
	return &Secret{
		Collector: issues.NewCollector(codes, c.Config),
		Cache:     c,
	}
}

func (s *Secret) Preloads() Preloads {
	return Preloads{
		internal.SEC: db.LoadResource[*v1.Secret],
		internal.PO:  db.LoadResource[*v1.Pod],
		internal.SA:  db.LoadResource[*v1.ServiceAccount],
		internal.ING: db.LoadResource[*netv1.Ingress],
	}
}

// Lint all available Secrets.
func (s *Secret) Lint(ctx context.Context) error {
	for k, f := range s.Preloads() {
		if err := f(ctx, s.Loader, internal.Glossary[k]); err != nil {
			return err
		}
	}

	return lint.NewSecret(s.Collector, s.DB).Lint(ctx)
}
