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
	v1 "k8s.io/api/core/v1"
)

// DaemonSet tracks DaemonSet sanitization.
type DaemonSet struct {
	*issues.Collector

	db *db.DB
}

// NewDaemonSet returns a new instance.
func NewDaemonSet(co *issues.Collector, db *db.DB) *DaemonSet {
	return &DaemonSet{
		Collector: co,
		db:        db,
	}
}

// Lint cleanse the resource.
func (s *DaemonSet) Lint(ctx context.Context) error {
	over := pullOverAllocs(ctx)
	txn, it := s.db.MustITFor(internal.Glossary[internal.DS])
	defer txn.Abort()
	for o := it.Next(); o != nil; o = it.Next() {
		ds := o.(*appsv1.DaemonSet)
		fqn := client.FQN(ds.Namespace, ds.Name)
		s.InitOutcome(fqn)
		ctx = internal.WithSpec(ctx, coSpecFor(fqn, ds, ds.Spec.Template.Spec))

		s.checkDaemonSet(ctx, ds)
		s.checkContainers(ctx, fqn, ds.Spec.Template.Spec)
		s.checkUtilization(ctx, over, ds)
	}

	return nil
}

func (s *DaemonSet) checkDaemonSet(ctx context.Context, ds *appsv1.DaemonSet) {
	if ds.Spec.Template.Spec.ServiceAccountName == "" {
		return
	}
	_, err := s.db.Find(internal.Glossary[internal.SA], client.FQN(ds.Namespace, ds.Spec.Template.Spec.ServiceAccountName))
	if err != nil {
		s.AddCode(ctx, 507, ds.Spec.Template.Spec.ServiceAccountName)
	}
}

// CheckContainers runs thru deployment template and checks pod configuration.
func (s *DaemonSet) checkContainers(ctx context.Context, fqn string, spec v1.PodSpec) {
	c := NewContainer(fqn, s)
	for _, co := range spec.InitContainers {
		c.sanitize(ctx, co, false)
	}
	for _, co := range spec.Containers {
		c.sanitize(ctx, co, false)
	}
}

// CheckUtilization checks deployments requested resources vs current utilization.
func (s *DaemonSet) checkUtilization(ctx context.Context, over bool, ds *appsv1.DaemonSet) {
	mx := resourceUsage(ctx, s.db, s, ds.Namespace, ds.Spec.Selector)
	if mx.RequestCPU.IsZero() && mx.RequestMEM.IsZero() {
		return
	}

	checkCPU(ctx, s, over, mx)
	checkMEM(ctx, s, over, mx)
}
