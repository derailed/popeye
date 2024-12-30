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

// Deployment represents a Deployment scruber.
type Deployment struct {
	*issues.Collector
	*Cache
}

// NewDeployment returns a new instance.
func NewDeployment(_ context.Context, c *Cache, codes *issues.Codes) Linter {
	return &Deployment{
		Collector: issues.NewCollector(codes, c.Config),
		Cache:     c,
	}
}

func (s *Deployment) Preloads() Preloads {
	return Preloads{
		internal.DP:  db.LoadResource[*appsv1.Deployment],
		internal.PO:  db.LoadResource[*v1.Pod],
		internal.SA:  db.LoadResource[*v1.ServiceAccount],
		internal.PMX: db.LoadResource[*mv1beta1.PodMetrics],
	}
}

// Lint all available Deployments.
func (s *Deployment) Lint(ctx context.Context) error {
	for k, f := range s.Preloads() {
		if err := f(ctx, s.Loader, internal.Glossary[k]); err != nil {
			return err
		}
	}

	return lint.NewDeployment(s.Collector, s.DB).Lint(ctx)
}
