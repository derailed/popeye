// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Popeye

package scrub

import (
	"context"

	v2 "github.com/cilium/cilium/pkg/k8s/apis/cilium.io/v2"
	"github.com/derailed/popeye/internal"
	"github.com/derailed/popeye/internal/cilium"
	"github.com/derailed/popeye/internal/cilium/lint"
	"github.com/derailed/popeye/internal/db"
	"github.com/derailed/popeye/internal/issues"
	iscrub "github.com/derailed/popeye/internal/scrub"
)

// CiliumNetworkPolicy represents a CiliumNetworkPolicy scruber.
type CiliumNetworkPolicy struct {
	*issues.Collector
	*iscrub.Cache
}

// NewCiliumNetworkPolicy returns a new instance.
func NewCiliumNetworkPolicy(ctx context.Context, c *iscrub.Cache, codes *issues.Codes) iscrub.Linter {
	return &CiliumNetworkPolicy{
		Collector: issues.NewCollector(codes, c.Config),
		Cache:     c,
	}
}

func (s *CiliumNetworkPolicy) Preloads() iscrub.Preloads {
	return iscrub.Preloads{
		cilium.CNP: db.LoadResource[*v2.CiliumNetworkPolicy],
		cilium.CEP: db.LoadResource[*v2.CiliumEndpoint],
	}
}

// Lint all available CiliumNetworkPolicys.
func (s *CiliumNetworkPolicy) Lint(ctx context.Context) error {
	for k, f := range s.Preloads() {
		if err := f(ctx, s.Loader, internal.Glossary[k]); err != nil {
			return err
		}
	}

	return lint.NewCiliumNetworkPolicy(s.Collector, s.DB).Lint(ctx)
}
