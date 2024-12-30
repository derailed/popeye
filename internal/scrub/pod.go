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
	polv1 "k8s.io/api/policy/v1"
	mv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

// Pod represents a Pod scruber.
type Pod struct {
	*issues.Collector
	*Cache
}

// NewPod return a new instance.
func NewPod(_ context.Context, c *Cache, codes *issues.Codes) Linter {
	return &Pod{
		Collector: issues.NewCollector(codes, c.Config),
		Cache:     c,
	}
}

func (s *Pod) Preloads() Preloads {
	return Preloads{
		internal.PO:  db.LoadResource[*v1.Pod],
		internal.SA:  db.LoadResource[*v1.ServiceAccount],
		internal.PDB: db.LoadResource[*polv1.PodDisruptionBudget],
		internal.NP:  db.LoadResource[*netv1.NetworkPolicy],
		internal.PMX: db.LoadResource[*mv1beta1.PodMetrics],
	}
}

// Lint all available Pods.
func (s *Pod) Lint(ctx context.Context) error {
	for k, f := range s.Preloads() {
		if err := f(ctx, s.Loader, internal.Glossary[k]); err != nil {
			return err
		}
	}

	return lint.NewPod(s.Collector, s.DB).Lint(ctx)
}
