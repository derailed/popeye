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

// CiliumEndpoint represents a CiliumEndpoint scruber.
type CiliumEndpoint struct {
	*issues.Collector
	*iscrub.Cache
}

// NewCiliumEndpoint returns a new instance.
func NewCiliumEndpoint(ctx context.Context, c *iscrub.Cache, codes *issues.Codes) iscrub.Linter {
	return &CiliumEndpoint{
		Collector: issues.NewCollector(codes, c.Config),
		Cache:     c,
	}
}

func (s *CiliumEndpoint) Preloads() iscrub.Preloads {
	return iscrub.Preloads{
		cilium.CEP:  db.LoadResource[*v2.CiliumEndpoint],
		cilium.CID:  db.LoadResource[*v2.CiliumIdentity],
		internal.PO: db.LoadResource[*v1.Pod],
		internal.NO: db.LoadResource[*v1.Node],
		internal.SA: db.LoadResource[*v1.ServiceAccount],
		internal.NS: db.LoadResource[*v1.Namespace],
	}
}

// Lint all available CiliumEndpoints.
func (s *CiliumEndpoint) Lint(ctx context.Context) error {
	for k, f := range s.Preloads() {
		if err := f(ctx, s.Loader, internal.Glossary[k]); err != nil {
			return err
		}
	}

	return lint.NewCiliumEndpoint(s.Collector, s.DB).Lint(ctx)
}
