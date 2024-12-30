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
	polv1 "k8s.io/api/policy/v1"
)

// PodDisruptionBudget represents a pdb scruber.
type PodDisruptionBudget struct {
	*issues.Collector
	*Cache
}

// NewPodDisruptionBudget return a new PodDisruptionBudget scruber.
func NewPodDisruptionBudget(_ context.Context, c *Cache, codes *issues.Codes) Linter {
	return &PodDisruptionBudget{
		Collector: issues.NewCollector(codes, c.Config),
		Cache:     c,
	}
}

func (s *PodDisruptionBudget) Preloads() Preloads {
	return Preloads{
		internal.PDB: db.LoadResource[*polv1.PodDisruptionBudget],
		internal.PO:  db.LoadResource[*v1.Pod],
	}
}

// Lint all available PodDisruptionBudgets.
func (s *PodDisruptionBudget) Lint(ctx context.Context) error {
	for k, f := range s.Preloads() {
		if err := f(ctx, s.Loader, internal.Glossary[k]); err != nil {
			return err
		}
	}

	return lint.NewPodDisruptionBudget(s.Collector, s.DB).Lint(ctx)
}
