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

// Gateway represents a Gateway scruber.
type Gateway struct {
	*issues.Collector
	*Cache
}

// NewGateway return a new instance.
func NewGateway(_ context.Context, c *Cache, codes *issues.Codes) Linter {
	return &Gateway{
		Collector: issues.NewCollector(codes, c.Config),
		Cache:     c,
	}
}

func (s *Gateway) Preloads() Preloads {
	return Preloads{
		internal.GW:  db.LoadResource[*gwv1.Gateway],
		internal.GWC: db.LoadResource[*gwv1.GatewayClass],
	}
}

// Lint all available Gateway.
func (s *Gateway) Lint(ctx context.Context) error {
	for k, f := range s.Preloads() {
		if err := f(ctx, s.Loader, internal.Glossary[k]); err != nil {
			return err
		}
	}

	return lint.NewGateway(s.Collector, s.DB).Lint(ctx)
}
