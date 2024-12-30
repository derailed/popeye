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

// Service represents a Service scruber.
type Service struct {
	*issues.Collector
	*Cache
}

// NewService return a new instance.
func NewService(_ context.Context, c *Cache, codes *issues.Codes) Linter {
	return &Service{
		Collector: issues.NewCollector(codes, c.Config),
		Cache:     c,
	}
}

func (s *Service) Preloads() Preloads {
	return Preloads{
		internal.SVC: db.LoadResource[*v1.Service],
		internal.PO:  db.LoadResource[*v1.Pod],
		internal.EP:  db.LoadResource[*v1.Endpoints],
	}
}

// Lint all available Services.
func (s *Service) Lint(ctx context.Context) error {
	for k, f := range s.Preloads() {
		if err := f(ctx, s.Loader, internal.Glossary[k]); err != nil {
			return err
		}
	}

	return lint.NewService(s.Collector, s.DB).Lint(ctx)
}
