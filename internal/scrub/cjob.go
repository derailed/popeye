// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Popeye

package scrub

import (
	"context"

	"github.com/derailed/popeye/internal"
	"github.com/derailed/popeye/internal/db"
	"github.com/derailed/popeye/internal/issues"
	"github.com/derailed/popeye/internal/lint"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	mv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

// CronJob represents a CronJob scruber.
type CronJob struct {
	*issues.Collector
	*Cache
}

// NewCronJob return a new instance.
func NewCronJob(_ context.Context, c *Cache, codes *issues.Codes) Linter {
	return &CronJob{
		Collector: issues.NewCollector(codes, c.Config),
		Cache:     c,
	}
}

func (s *CronJob) Preloads() Preloads {
	return Preloads{
		internal.CJOB: db.LoadResource[*batchv1.CronJob],
		internal.JOB:  db.LoadResource[*batchv1.Job],
		internal.PO:   db.LoadResource[*v1.Pod],
		internal.SA:   db.LoadResource[*v1.ServiceAccount],
		internal.PMX:  db.LoadResource[*mv1beta1.PodMetrics],
	}
}

// Lint all available CronJobs.
func (s *CronJob) Lint(ctx context.Context) error {
	for k, f := range s.Preloads() {
		if err := f(ctx, s.Loader, internal.Glossary[k]); err != nil {
			return err
		}
	}

	return lint.NewCronJob(s.Collector, s.DB).Lint(ctx)
}
