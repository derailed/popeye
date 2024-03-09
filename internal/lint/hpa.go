// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Popeye

package lint

import (
	"context"
	"strings"

	"github.com/derailed/popeye/internal"
	"github.com/derailed/popeye/internal/cache"
	"github.com/derailed/popeye/internal/client"
	"github.com/derailed/popeye/internal/db"
	"github.com/derailed/popeye/internal/issues"
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

// HorizontalPodAutoscaler represents a HorizontalPodAutoscaler linter.
type HorizontalPodAutoscaler struct {
	*issues.Collector

	db *db.DB
}

// NewHorizontalPodAutoscaler returns a new instance.
func NewHorizontalPodAutoscaler(co *issues.Collector, db *db.DB) *HorizontalPodAutoscaler {
	return &HorizontalPodAutoscaler{
		Collector: co,
		db:        db,
	}
}

// Lint sanitizes an hpa.
func (h *HorizontalPodAutoscaler) Lint(ctx context.Context) error {
	var (
		tcpu, tmem resource.Quantity
		current    int32
	)
	res, err := cache.ListAvailableMetrics(h.db)
	if err != nil {
		return err
	}
	txn, it := h.db.MustITFor(internal.Glossary[internal.HPA])
	defer txn.Abort()
	for o := it.Next(); o != nil; o = it.Next() {
		hpa := o.(*autoscalingv1.HorizontalPodAutoscaler)
		fqn := client.FQN(hpa.Namespace, hpa.Name)
		h.InitOutcome(fqn)
		ctx = internal.WithSpec(ctx, SpecFor(fqn, hpa))
		var rcpu, rmem resource.Quantity
		ns, _ := namespaced(fqn)
		switch hpa.Spec.ScaleTargetRef.Kind {
		case "Deployment":
			rfqn := cache.FQN(ns, hpa.Spec.ScaleTargetRef.Name)
			if o, err := h.db.Find(internal.Glossary[internal.DP], rfqn); err == nil {
				dp := o.(*appsv1.Deployment)
				rcpu, rmem = podResources(dp.Spec.Template.Spec)
				current = dp.Status.AvailableReplicas
			} else {
				h.AddCode(ctx, 600, fqn, strings.ToLower(hpa.Spec.ScaleTargetRef.Kind), rfqn)
				continue
			}

		case "ReplicaSet":
			rfqn := cache.FQN(ns, hpa.Spec.ScaleTargetRef.Name)
			if o, err := h.db.Find(internal.Glossary[internal.RS], rfqn); err == nil {
				rs := o.(*appsv1.ReplicaSet)
				rcpu, rmem = podResources(rs.Spec.Template.Spec)
				current = rs.Status.AvailableReplicas
			} else {
				h.AddCode(ctx, 600, fqn, strings.ToLower(hpa.Spec.ScaleTargetRef.Kind), rfqn)
				continue
			}

		case "StatefulSet":
			rfqn := cache.FQN(ns, hpa.Spec.ScaleTargetRef.Name)
			if o, err := h.db.Find(internal.Glossary[internal.STS], rfqn); err == nil {
				sts := o.(*appsv1.StatefulSet)
				rcpu, rmem = podResources(sts.Spec.Template.Spec)
				current = sts.Status.CurrentReplicas
			} else {
				h.AddCode(ctx, 600, fqn, strings.ToLower(hpa.Spec.ScaleTargetRef.Kind), rfqn)
				continue
			}
		}

		rList := v1.ResourceList{v1.ResourceCPU: rcpu, v1.ResourceMemory: rmem}
		list := h.checkResources(ctx, hpa.Spec.MaxReplicas, current, rList, res)
		tcpu.Add(*list.Cpu())
		tmem.Add(*list.Memory())
	}
	h.checkUtilization(ctx, tcpu, tmem, res)

	return nil
}

func (h *HorizontalPodAutoscaler) checkResources(ctx context.Context, max, current int32, rList, res v1.ResourceList) v1.ResourceList {
	rcpu, rmem := rList.Cpu(), rList.Memory()
	acpu, amem := *res.Cpu(), *res.Memory()
	trcpu, trmem := rcpu.DeepCopy(), rmem.DeepCopy()
	for i := 1; i <= int(max-current); i++ {
		trcpu.Add(*rcpu)
		trmem.Add(*rmem)
	}
	if toMC(trcpu) > toMC(acpu) {
		cpu := trcpu.DeepCopy()
		cpu.Sub(acpu)
		h.AddCode(ctx, 602, current, max, asMC(acpu), asMC(cpu))
	}
	if toMB(trmem) > toMB(amem) {
		mem := trmem.DeepCopy()
		mem.Sub(amem)
		h.AddCode(ctx, 603, current, max, asMB(amem), asMB(mem))
	}

	return v1.ResourceList{v1.ResourceCPU: trcpu, v1.ResourceMemory: trmem}
}

func (h *HorizontalPodAutoscaler) checkUtilization(ctx context.Context, tcpu, tmem resource.Quantity, res v1.ResourceList) {
	acpu, amem := *res.Cpu(), *res.Memory()
	ctx = internal.WithSpec(ctx, SpecFor("HPA", nil))
	if toMC(tcpu) > toMC(acpu) {
		cpu := tcpu.DeepCopy()
		cpu.Sub(acpu)
		h.AddCode(ctx, 604, asMC(tcpu), asMC(acpu), asMC(cpu))
	}
	if toMB(tmem) > toMB(amem) {
		mem := tmem.DeepCopy()
		mem.Sub(amem)
		h.AddCode(ctx, 605, asMB(tmem), asMB(amem), asMB(mem))
	}
}
