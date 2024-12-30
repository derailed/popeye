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
	mv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

// Node represents a Node scruber.
type Node struct {
	*issues.Collector
	*Cache
}

// NewNode return a new instance.
func NewNode(_ context.Context, c *Cache, codes *issues.Codes) Linter {
	return &Node{
		Collector: issues.NewCollector(codes, c.Config),
		Cache:     c,
	}
}

func (s *Node) Preloads() Preloads {
	return Preloads{
		internal.NO:  db.LoadResource[*v1.Node],
		internal.PO:  db.LoadResource[*v1.Pod],
		internal.NMX: db.LoadResource[*mv1beta1.NodeMetrics],
	}
}

// Lint all available Nodes.
func (s *Node) Lint(ctx context.Context) error {
	for k, f := range s.Preloads() {
		if err := f(ctx, s.Loader, internal.Glossary[k]); err != nil {
			return err
		}
	}

	return lint.NewNode(s.Collector, s.DB).Lint(ctx)
}
