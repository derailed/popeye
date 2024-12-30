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
)

// Namespace represents a Namespace scruber.
type Namespace struct {
	*issues.Collector
	*Cache
}

// NewNamespace returns a new instance.
func NewNamespace(_ context.Context, c *Cache, codes *issues.Codes) Linter {
	return &Namespace{
		Collector: issues.NewCollector(codes, c.Config),
		Cache:     c,
	}
}

func (s *Namespace) Preloads() Preloads {
	return Preloads{
		internal.NS: db.LoadResource[*v1.Namespace],
		internal.PO: db.LoadResource[*v1.Pod],
		internal.SA: db.LoadResource[*v1.ServiceAccount],
	}
}

// Lint all available Namespaces.
func (s *Namespace) Lint(ctx context.Context) error {
	for k, f := range s.Preloads() {
		if err := f(ctx, s.Loader, internal.Glossary[k]); err != nil {
			return err
		}
	}

	return lint.NewNamespace(s.Collector, s.DB).Lint(ctx)
}
