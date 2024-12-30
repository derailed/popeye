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

// Job represents a Job scruber.
type Job struct {
	*issues.Collector
	*Cache
}

// NewJob return a new instance.
func NewJob(_ context.Context, c *Cache, codes *issues.Codes) Linter {
	return &Job{
		Collector: issues.NewCollector(codes, c.Config),
		Cache:     c,
	}
}

func (s *Job) Preloads() Preloads {
	return Preloads{
		internal.CJOB: db.LoadResource[*batchv1.CronJob],
		internal.JOB:  db.LoadResource[*batchv1.Job],
		internal.PO:   db.LoadResource[*v1.Pod],
		internal.SA:   db.LoadResource[*v1.ServiceAccount],
		internal.PMX:  db.LoadResource[*mv1beta1.PodMetrics],
	}
}

// Lint all available Jobs.
func (s *Job) Lint(ctx context.Context) error {
	for k, f := range s.Preloads() {
		if err := f(ctx, s.Loader, internal.Glossary[k]); err != nil {
			return err
		}
	}

	return lint.NewJob(s.Collector, s.DB).Lint(ctx)
}
