// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Popeye

package lint

import (
	"context"

	"github.com/derailed/popeye/internal"
	"github.com/derailed/popeye/internal/client"
	"github.com/derailed/popeye/internal/dao"
	"github.com/derailed/popeye/internal/db"
	"github.com/derailed/popeye/internal/issues"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
)

// Job tracks Job linting.
type Job struct {
	*issues.Collector

	db *db.DB
}

// NewJob returns a new instance.
func NewJob(co *issues.Collector, db *db.DB) *Job {
	return &Job{
		Collector: co,
		db:        db,
	}
}

// Lint cleanse the resource.
func (s *Job) Lint(ctx context.Context) error {
	over := pullOverAllocs(ctx)
	txn, it := s.db.MustITFor(internal.Glossary[internal.JOB])
	defer txn.Abort()
	for o := it.Next(); o != nil; o = it.Next() {
		j := o.(*batchv1.Job)
		fqn := client.FQN(j.Namespace, j.Name)
		s.InitOutcome(fqn)
		ctx = internal.WithSpec(ctx, coSpecFor(fqn, j, j.Spec.Template.Spec))
		s.checkJob(ctx, fqn, j)
		s.checkContainers(ctx, fqn, j.Spec.Template.Spec)
		s.checkUtilization(ctx, over, fqn)
	}

	return nil
}

// CheckJob checks if Job contract is currently happy or not.
func (s *Job) checkJob(ctx context.Context, fqn string, j *batchv1.Job) {
	checkEvents(ctx, s.Collector, internal.JOB, dao.WarnEvt, "Job", fqn)

	if j.Spec.Suspend != nil && *j.Spec.Suspend {
		s.AddCode(ctx, 1500, j.Kind)
	}

	if sa := j.Spec.Template.Spec.ServiceAccountName; sa != "" {
		saFQN := client.FQN(j.Namespace, sa)
		if !s.db.Exists(internal.Glossary[internal.SA], saFQN) {
			s.AddCode(ctx, 307, j.Kind, sa)
		}
	}
}

// CheckContainers runs thru Job template and checks pod configuration.
func (s *Job) checkContainers(ctx context.Context, fqn string, spec v1.PodSpec) {
	c := NewContainer(fqn, s)
	for _, co := range spec.InitContainers {
		c.sanitize(ctx, co, false)
	}
	for _, co := range spec.Containers {
		c.sanitize(ctx, co, false)
	}
}

// CheckUtilization checks Jobs requested resources vs current utilization.
func (s *Job) checkUtilization(ctx context.Context, over bool, fqn string) {
	jj, err := s.db.FindJobs(fqn)
	if err != nil {
		s.AddErr(ctx, err)
		return
	}
	mx := jobResourceUsage(s.db, jj)
	if mx.RequestCPU.IsZero() && mx.RequestMEM.IsZero() {
		return
	}
	checkCPU(ctx, s, over, mx)
	checkMEM(ctx, s, over, mx)
}
