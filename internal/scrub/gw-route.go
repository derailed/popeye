// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Popeye

package scrub

import (
	"context"

	"github.com/derailed/popeye/internal"
	"github.com/derailed/popeye/internal/db"
	"github.com/derailed/popeye/internal/issues"
	"github.com/derailed/popeye/internal/lint"
	gwv1 "sigs.k8s.io/gateway-api/apis/v1"
)

// HTTPRoute represents a HTTPRoute scruber.
type HTTPRoute struct {
	*issues.Collector
	*Cache
}

// NewHTTPRoute return a new instance.
func NewHTTPRoute(_ context.Context, c *Cache, codes *issues.Codes) Linter {
	return &HTTPRoute{
		Collector: issues.NewCollector(codes, c.Config),
		Cache:     c,
	}
}

func (s *HTTPRoute) Preloads() Preloads {
	return Preloads{
		internal.GW:  db.LoadResource[*gwv1.Gateway],
		internal.GWC: db.LoadResource[*gwv1.GatewayClass],
		internal.GWR: db.LoadResource[*gwv1.HTTPRoute],
	}
}

// Lint all available HTTPRoute.
func (s *HTTPRoute) Lint(ctx context.Context) error {
	for k, f := range s.Preloads() {
		if err := f(ctx, s.Loader, internal.Glossary[k]); err != nil {
			return err
		}
	}

	return lint.NewHTTPRoute(s.Collector, s.DB).Lint(ctx)
}
