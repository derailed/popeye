// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Popeye

package lint

import (
	"context"

	"github.com/derailed/popeye/internal"
	"github.com/derailed/popeye/internal/client"
	"github.com/derailed/popeye/internal/db"
	"github.com/derailed/popeye/internal/issues"
	appsv1 "k8s.io/api/apps/v1"
)

type (
	// CollectorLimiter represents a collector with resource allocation limits.
	CollectorLimiter interface {
		// Collector an issue collector.
		Collector

		// ConfigLister track config options.
		ConfigLister
	}

	// StatefulSet represents a StatefulSet linter.
	StatefulSet struct {
		*issues.Collector

		db *db.DB
	}
)

// NewStatefulSet returns a new instance.
func NewStatefulSet(co *issues.Collector, db *db.DB) *StatefulSet {
	return &StatefulSet{
		Collector: co,
		db:        db,
	}
}

// Lint cleanse the resource.
func (s *StatefulSet) Lint(ctx context.Context) error {
	over := pullOverAllocs(ctx)
	txn, it := s.db.MustITFor(internal.Glossary[internal.STS])
	defer txn.Abort()
	for o := it.Next(); o != nil; o = it.Next() {
		sts := o.(*appsv1.StatefulSet)
		fqn := client.FQN(sts.Namespace, sts.Name)
		s.InitOutcome(fqn)
		ctx = internal.WithSpec(ctx, coSpecFor(fqn, sts, sts.Spec.Template.Spec))

		s.checkStatefulSet(ctx, sts)
		s.checkContainers(ctx, fqn, sts)
		s.checkUtilization(ctx, over, sts)
	}

	return nil
}

func (s *StatefulSet) checkStatefulSet(ctx context.Context, sts *appsv1.StatefulSet) {
	if sts.Spec.Replicas == nil || (sts.Spec.Replicas != nil && *sts.Spec.Replicas == 0) {
		s.AddCode(ctx, 500)
		return
	}

	if sts.Spec.Replicas != nil && *sts.Spec.Replicas != sts.Status.ReadyReplicas {
		s.AddCode(ctx, 501, *sts.Spec.Replicas, sts.Status.ReadyReplicas)
	}

	if sts.Spec.Template.Spec.ServiceAccountName == "" {
		return
	}

	saFQN := client.FQN(sts.Namespace, sts.Spec.Template.Spec.ServiceAccountName)
	if !s.db.Exists(internal.Glossary[internal.SA], saFQN) {
		s.AddCode(ctx, 507, sts.Spec.Template.Spec.ServiceAccountName)
	}
}

func (s *StatefulSet) checkContainers(ctx context.Context, fqn string, st *appsv1.StatefulSet) {
	spec := st.Spec.Template.Spec

	l := NewContainer(fqn, s)
	for _, co := range spec.InitContainers {
		l.sanitize(ctx, co, false)
	}

	for _, co := range spec.Containers {
		l.sanitize(ctx, co, false)
	}
}

func checkCPU(ctx context.Context, c CollectorLimiter, over bool, mx ConsumptionMetrics) {
	cpuPerc := mx.ReqCPURatio()
	if cpuPerc > 1 && cpuPerc > float64(c.CPUResourceLimits().UnderPerc) {
		c.AddCode(ctx, 503, asMC(mx.CurrentCPU), asMC(mx.RequestCPU), asPerc(cpuPerc))
		return
	}

	if over && cpuPerc > 0 && cpuPerc < float64(c.CPUResourceLimits().OverPerc) {
		c.AddCode(ctx, 504, asMC(mx.CurrentCPU), asMC(mx.RequestCPU), asPerc(mx.ReqAbsCPURatio()))
	}
}

func checkMEM(ctx context.Context, c CollectorLimiter, over bool, mx ConsumptionMetrics) {
	memPerc := mx.ReqMEMRatio()
	if memPerc > 1 && memPerc > float64(c.MEMResourceLimits().UnderPerc) {
		c.AddCode(ctx, 505, asMB(mx.CurrentMEM), asMB(mx.RequestMEM), asPerc(memPerc))
		return
	}

	if over && memPerc < float64(c.MEMResourceLimits().OverPerc) {
		c.AddCode(ctx, 506, asMB(mx.CurrentMEM), asMB(mx.RequestMEM), asPerc(mx.ReqAbsMEMRatio()))
	}
}

func (s *StatefulSet) checkUtilization(ctx context.Context, over bool, sts *appsv1.StatefulSet) {
	mx := resourceUsage(ctx, s.db, s, sts.Namespace, sts.Spec.Selector)
	if mx.RequestCPU.IsZero() && mx.RequestMEM.IsZero() {
		return
	}

	checkCPU(ctx, s, over, mx)
	checkMEM(ctx, s, over, mx)
}
