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

// NetworkPolicy represents a NetworkPolicy scruber.
type NetworkPolicy struct {
	*issues.Collector
	*Cache
}

// NewNetworkPolicy return a new instance.
func NewNetworkPolicy(_ context.Context, c *Cache, codes *issues.Codes) Linter {
	return &NetworkPolicy{
		Collector: issues.NewCollector(codes, c.Config),
		Cache:     c,
	}
}

func (s *NetworkPolicy) Preloads() Preloads {
	return Preloads{
		internal.NP: db.LoadResource[*netv1.NetworkPolicy],
		internal.NS: db.LoadResource[*v1.Namespace],
		internal.PO: db.LoadResource[*v1.Pod],
	}
}

// Lint all available NetworkPolicies.
func (s *NetworkPolicy) Lint(ctx context.Context) error {
	for k, f := range s.Preloads() {
		if err := f(ctx, s.Loader, internal.Glossary[k]); err != nil {
			return err
		}
	}

	return lint.NewNetworkPolicy(s.Collector, s.DB).Lint(ctx)
}
