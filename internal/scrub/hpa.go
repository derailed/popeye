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
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	v1 "k8s.io/api/core/v1"
	mv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

// HorizontalPodAutoscaler represents a HorizontalPodAutoscaler scruber.
type HorizontalPodAutoscaler struct {
	*issues.Collector
	*Cache
}

// NewHorizontalPodAutoscaler returns a new instance.
func NewHorizontalPodAutoscaler(_ context.Context, c *Cache, codes *issues.Codes) Linter {
	return &HorizontalPodAutoscaler{
		Collector: issues.NewCollector(codes, c.Config),
		Cache:     c,
	}
}

func (s *HorizontalPodAutoscaler) Preloads() Preloads {
	return Preloads{
		internal.HPA: db.LoadResource[*autoscalingv1.HorizontalPodAutoscaler],
		internal.DP:  db.LoadResource[*appsv1.Deployment],
		internal.STS: db.LoadResource[*appsv1.StatefulSet],
		internal.RS:  db.LoadResource[*appsv1.ReplicaSet],
		internal.NO:  db.LoadResource[*v1.Node],
		internal.PO:  db.LoadResource[*v1.Pod],
		internal.SA:  db.LoadResource[*v1.ServiceAccount],
		internal.PMX: db.LoadResource[*mv1beta1.PodMetrics],
		internal.NMX: db.LoadResource[*mv1beta1.NodeMetrics],
	}
}

// Lint all available HorizontalPodAutoscalers.
func (s *HorizontalPodAutoscaler) Lint(ctx context.Context) error {
	for k, f := range s.Preloads() {
		if err := f(ctx, s.Loader, internal.Glossary[k]); err != nil {
			return err
		}
	}

	return lint.NewHorizontalPodAutoscaler(s.Collector, s.DB).Lint(ctx)
}
