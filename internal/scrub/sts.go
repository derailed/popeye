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
	mv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

// StatefulSet represents a StatefulSet scruber.
type StatefulSet struct {
	*issues.Collector
	*Cache
}

// NewStatefulSet return a new StatefulSet scruber.
func NewStatefulSet(_ context.Context, c *Cache, codes *issues.Codes) Linter {
	return &StatefulSet{
		Collector: issues.NewCollector(codes, c.Config),
		Cache:     c,
	}
}

func (s *StatefulSet) Preloads() Preloads {
	return Preloads{
		internal.STS: db.LoadResource[*appsv1.StatefulSet],
		internal.PO:  db.LoadResource[*v1.Pod],
		internal.SA:  db.LoadResource[*v1.ServiceAccount],
		internal.PMX: db.LoadResource[*mv1beta1.PodMetrics],
	}
}

// Lint all available StatefulSets.
func (s *StatefulSet) Lint(ctx context.Context) error {
	for k, f := range s.Preloads() {
		if err := f(ctx, s.Loader, internal.Glossary[k]); err != nil {
			return err
		}
	}

	return lint.NewStatefulSet(s.Collector, s.DB).Lint(ctx)
}
