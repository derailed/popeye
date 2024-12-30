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
	"k8s.io/apimachinery/pkg/api/resource"
)

// Deployment tracks Deployment sanitization.
type Deployment struct {
	*issues.Collector

	db *db.DB
}

// NewDeployment returns a new instance.
func NewDeployment(co *issues.Collector, db *db.DB) *Deployment {
	return &Deployment{
		Collector: co,
		db:        db,
	}
}

// Lint cleanse the resource.
func (s *Deployment) Lint(ctx context.Context) error {
	over := pullOverAllocs(ctx)
	txn, it := s.db.MustITFor(internal.Glossary[internal.DP])
	defer txn.Abort()
	for o := it.Next(); o != nil; o = it.Next() {
		dp := o.(*appsv1.Deployment)
		fqn := client.FQN(dp.Namespace, dp.Name)
		s.InitOutcome(fqn)
		ctx = internal.WithSpec(ctx, coSpecFor(fqn, dp, dp.Spec.Template.Spec))
		s.checkDeployment(ctx, dp)
		s.checkContainers(ctx, fqn, dp.Spec.Template.Spec)
		s.checkUtilization(ctx, over, dp)
	}

	return nil
}

// CheckDeployment checks if deployment contract is currently happy or not.
func (s *Deployment) checkDeployment(ctx context.Context, dp *appsv1.Deployment) {
	if dp.Spec.Replicas == nil || (dp.Spec.Replicas != nil && *dp.Spec.Replicas == 0) {
		s.AddCode(ctx, 500)
		return
	}

	if dp.Spec.Replicas != nil && *dp.Spec.Replicas != dp.Status.AvailableReplicas {
		s.AddCode(ctx, 501, *dp.Spec.Replicas, dp.Status.AvailableReplicas)
	}

	if dp.Spec.Template.Spec.ServiceAccountName == "" {
		return
	}

	saFQN := client.FQN(dp.Namespace, dp.Spec.Template.Spec.ServiceAccountName)
	if !s.db.Exists(internal.Glossary[internal.SA], saFQN) {
		s.AddCode(ctx, 507, dp.Spec.Template.Spec.ServiceAccountName)
	}
}

// CheckContainers runs thru deployment template and checks pod configuration.
func (s *Deployment) checkContainers(ctx context.Context, fqn string, spec v1.PodSpec) {
	c := NewContainer(fqn, s)
	for _, co := range spec.InitContainers {
		c.sanitize(ctx, co, false)
	}
	for _, co := range spec.Containers {
		c.sanitize(ctx, co, false)
	}
}

// CheckUtilization checks deployments requested resources vs current utilization.
func (s *Deployment) checkUtilization(ctx context.Context, over bool, dp *appsv1.Deployment) {
	mx := resourceUsage(ctx, s.db, s, dp.Namespace, dp.Spec.Selector)
	if mx.RequestCPU.IsZero() && mx.RequestMEM.IsZero() {
		return
	}
	checkCPU(ctx, s, over, mx)
	checkMEM(ctx, s, over, mx)
}

// Helpers...

// PullOverAllocs check for over allocation setting in context.
func pullOverAllocs(ctx context.Context) bool {
	over := ctx.Value(internal.KeyOverAllocs)
	if over == nil {
		return false
	}
	return over.(bool)
}

func computePodResources(spec v1.PodSpec) (cpu, mem resource.Quantity) {
	for _, co := range spec.InitContainers {
		c, m, _ := containerResources(co)
		if c != nil {
			cpu.Add(*c)
		}
		if m != nil {
			mem.Add(*m)
		}
	}

	for _, co := range spec.Containers {
		c, m, _ := containerResources(co)
		if c != nil {
			cpu.Add(*c)
		}
		if m != nil {
			mem.Add(*m)
		}
	}

	return
}
