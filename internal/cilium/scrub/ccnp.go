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

// CiliumClusterwideNetworkPolicy represents a CiliumClusterwideNetworkPolicy scruber.
type CiliumClusterwideNetworkPolicy struct {
	*issues.Collector
	*iscrub.Cache
}

// NewCiliumClusterwideNetworkPolicy returns a new instance.
func NewCiliumClusterwideNetworkPolicy(ctx context.Context, c *iscrub.Cache, codes *issues.Codes) iscrub.Linter {
	return &CiliumClusterwideNetworkPolicy{
		Collector: issues.NewCollector(codes, c.Config),
		Cache:     c,
	}
}

func (s *CiliumClusterwideNetworkPolicy) Preloads() iscrub.Preloads {
	return iscrub.Preloads{
		cilium.CCNP: db.LoadResource[*v2.CiliumClusterwideNetworkPolicy],
		cilium.CEP:  db.LoadResource[*v2.CiliumEndpoint],
	}
}

// Lint all available CiliumClusterwideNetworkPolicys.
func (s *CiliumClusterwideNetworkPolicy) Lint(ctx context.Context) error {
	for k, f := range s.Preloads() {
		if err := f(ctx, s.Loader, internal.Glossary[k]); err != nil {
			return err
		}
	}

	return lint.NewCiliumClusterwideNetworkPolicy(s.Collector, s.DB).Lint(ctx)
}
