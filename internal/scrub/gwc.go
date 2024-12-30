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

// GatewayClass represents a GatewayClass scruber.
type GatewayClass struct {
	*issues.Collector
	*Cache
}

// NewGatewayClass return a new instance.
func NewGatewayClass(_ context.Context, c *Cache, codes *issues.Codes) Linter {
	return &GatewayClass{
		Collector: issues.NewCollector(codes, c.Config),
		Cache:     c,
	}
}

func (s *GatewayClass) Preloads() Preloads {
	return Preloads{
		internal.GW:  db.LoadResource[*gwv1.Gateway],
		internal.GWC: db.LoadResource[*gwv1.GatewayClass],
	}
}

// Lint all available GatewayClass.
func (s *GatewayClass) Lint(ctx context.Context) error {
	for k, f := range s.Preloads() {
		if err := f(ctx, s.Loader, internal.Glossary[k]); err != nil {
			return err
		}
	}

	return lint.NewGatewayClass(s.Collector, s.DB).Lint(ctx)
}
