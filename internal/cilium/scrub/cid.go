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
	v1 "k8s.io/api/core/v1"
)

// CiliumIdentity represents a CiliumIdentity scruber.
type CiliumIdentity struct {
	*issues.Collector
	*iscrub.Cache
}

// NewCiliumIdentity returns a new instance.
func NewCiliumIdentity(ctx context.Context, c *iscrub.Cache, codes *issues.Codes) iscrub.Linter {
	return &CiliumIdentity{
		Collector: issues.NewCollector(codes, c.Config),
		Cache:     c,
	}
}

func (s *CiliumIdentity) Preloads() iscrub.Preloads {
	return iscrub.Preloads{
		cilium.CID:  db.LoadResource[*v2.CiliumIdentity],
		cilium.CEP:  db.LoadResource[*v2.CiliumEndpoint],
		internal.SA: db.LoadResource[*v1.ServiceAccount],
		internal.NS: db.LoadResource[*v1.Namespace],
	}
}

// Lint all available CiliumIdentities.
func (s *CiliumIdentity) Lint(ctx context.Context) error {
	for k, f := range s.Preloads() {
		if err := f(ctx, s.Loader, internal.Glossary[k]); err != nil {
			return err
		}
	}

	return lint.NewCiliumIdentity(s.Collector, s.DB).Lint(ctx)
}
