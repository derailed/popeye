// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Popeye

package scrub

import (
	"context"

	"github.com/derailed/popeye/internal"
	"github.com/derailed/popeye/internal/db"
	"github.com/derailed/popeye/internal/issues"
	"github.com/derailed/popeye/internal/lint"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
)

// ReplicaSet represents a ReplicaSet scruber.
type ReplicaSet struct {
	*issues.Collector
	*Cache
}

// NewReplicaSet returns a new instance.
func NewReplicaSet(_ context.Context, c *Cache, codes *issues.Codes) Linter {
	return &ReplicaSet{
		Collector: issues.NewCollector(codes, c.Config),
		Cache:     c,
	}
}

func (s *ReplicaSet) Preloads() Preloads {
	return Preloads{
		internal.RS: db.LoadResource[*appsv1.ReplicaSet],
		internal.PO: db.LoadResource[*v1.Pod],
	}
}

// Lint all available ReplicaSets.
func (s *ReplicaSet) Lint(ctx context.Context) error {
	for k, f := range s.Preloads() {
		if err := f(ctx, s.Loader, internal.Glossary[k]); err != nil {
			return err
		}
	}

	return lint.NewReplicaSet(s.Collector, s.DB).Lint(ctx)
}
