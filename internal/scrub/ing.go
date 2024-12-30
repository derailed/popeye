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

// Ingress represents a Ingress scruber.
type Ingress struct {
	*issues.Collector
	*Cache
}

// NewIngress return a new instance.
func NewIngress(_ context.Context, c *Cache, codes *issues.Codes) Linter {
	return &Ingress{
		Collector: issues.NewCollector(codes, c.Config),
		Cache:     c,
	}
}

func (s *Ingress) Preloads() Preloads {
	return Preloads{
		internal.ING: db.LoadResource[*netv1.Ingress],
		internal.SVC: db.LoadResource[*v1.Service],
	}
}

// Lint all available Ingress.
func (s *Ingress) Lint(ctx context.Context) error {
	for k, f := range s.Preloads() {
		if err := f(ctx, s.Loader, internal.Glossary[k]); err != nil {
			return err
		}
	}

	return lint.NewIngress(s.Collector, s.DB).Lint(ctx)
}
