// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Popeye

package lint

import (
	"context"

	"github.com/derailed/popeye/internal"
	"github.com/derailed/popeye/internal/cache"
	"github.com/derailed/popeye/internal/client"
	"github.com/derailed/popeye/internal/dao"
	"github.com/derailed/popeye/internal/db"
	"github.com/derailed/popeye/internal/issues"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
)

// CronJob tracks CronJob linting.
type CronJob struct {
	*issues.Collector

	db *db.DB
}

// NewCronJob returns a new instance.
func NewCronJob(co *issues.Collector, db *db.DB) *CronJob {
	return &CronJob{
		Collector: co,
		db:        db,
	}
}

// Lint cleanse the resource.
func (s *CronJob) Lint(ctx context.Context) error {
	over := pullOverAllocs(ctx)
	txn, it := s.db.MustITFor(internal.Glossary[internal.CJOB])
	defer txn.Abort()
	for o := it.Next(); o != nil; o = it.Next() {
		cj := o.(*batchv1.CronJob)
		fqn := client.FQN(cj.Namespace, cj.Name)
		s.InitOutcome(fqn)
		ctx = internal.WithSpec(ctx, coSpecFor(fqn, cj, cj.Spec.JobTemplate.Spec.Template.Spec))
		s.checkCronJob(ctx, fqn, cj)
		s.checkContainers(ctx, fqn, cj.Spec.JobTemplate.Spec.Template.Spec)
		s.checkUtilization(ctx, over, fqn)
	}

	return nil
}

// CheckCronJob checks if CronJob contract is currently happy or not.
func (s *CronJob) checkCronJob(ctx context.Context, fqn string, cj *batchv1.CronJob) {
	checkEvents(ctx, s.Collector, internal.CJOB, "", "CronJob", fqn)

	if cj.Spec.Suspend != nil && *cj.Spec.Suspend {
		s.AddCode(ctx, 1500, cj.Kind)
	}

	if len(cj.Status.Active) == 0 {
		s.AddCode(ctx, 1501)
	}
	if cj.Status.LastSuccessfulTime == nil {
		s.AddCode(ctx, 1502)
	}

	if sa := cj.Spec.JobTemplate.Spec.Template.Spec.ServiceAccountName; sa != "" {
		saFQN := client.FQN(cj.Namespace, sa)
		if !s.db.Exists(internal.Glossary[internal.SA], saFQN) {
			s.AddCode(ctx, 307, cj.Kind, sa)
		}
	}
}

// CheckContainers runs thru CronJob template and checks pod configuration.
func (s *CronJob) checkContainers(ctx context.Context, fqn string, spec v1.PodSpec) {
	c := NewContainer(fqn, s)
	for _, co := range spec.InitContainers {
		c.sanitize(ctx, co, false)
	}
	for _, co := range spec.Containers {
		c.sanitize(ctx, co, false)
	}
}

// CheckUtilization checks CronJobs requested resources vs current utilization.
func (s *CronJob) checkUtilization(ctx context.Context, over bool, fqn string) {
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

// Helpers...

func checkEvents(ctx context.Context, ii *issues.Collector, r internal.R, kind, object, fqn string) {
	ee, err := dao.EventsFor(ctx, internal.Glossary[r], kind, object, fqn)
	if err != nil {
		ii.AddErr(ctx, err)
		return
	}
	for _, e := range ee.Issues() {
		ii.AddCode(ctx, 1503, e)
	}
}

func jobResourceUsage(dba *db.DB, jobs []*batchv1.Job) ConsumptionMetrics {
	var mx ConsumptionMetrics

	if len(jobs) == 0 {
		return mx
	}

	for _, job := range jobs {
		fqn := cache.FQN(job.Namespace, job.Name)
		cpu, mem := computePodResources(job.Spec.Template.Spec)
		mx.RequestCPU.Add(cpu)
		mx.RequestMEM.Add(mem)

		pmx, err := dba.FindPMX(fqn)
		if err != nil || pmx == nil {
			continue
		}
		for _, cx := range pmx.Containers {
			mx.CurrentCPU.Add(*cx.Usage.Cpu())
			mx.CurrentMEM.Add(*cx.Usage.Memory())
		}
	}

	return mx
}
