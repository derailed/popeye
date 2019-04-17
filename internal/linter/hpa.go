package linter

import (
	"context"

	"github.com/derailed/popeye/internal/k8s"
	"github.com/rs/zerolog"
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

// HorizontalPodAutoscaler represents a HorizontalPodAutoscaler linter.
type HorizontalPodAutoscaler struct {
	*Linter
}

// NewHorizontalPodAutoscaler returns a new ServiceAccount linter.
func NewHorizontalPodAutoscaler(l Loader, log *zerolog.Logger) *HorizontalPodAutoscaler {
	return &HorizontalPodAutoscaler{NewLinter(l, log)}
}

// Lint a serviceaccount.
func (h *HorizontalPodAutoscaler) Lint(ctx context.Context) error {
	// h.pigs()

	hpas, err := h.ListHorizontalPodAutoscalers()
	if err != nil {
		return err
	}

	dps, err := h.ListDeployments()
	if err != nil {
		return err
	}

	sts, err := h.ListStatefulSets()
	if err != nil {
		return err
	}

	nmx := make(k8s.NodesMetrics)
	if err := nodeMetrics(h, nmx); err != nil {
		return err
	}

	acpu, amem := clusterCapacity(nmx)
	h.lint(hpas, dps, sts, acpu, amem)

	return nil
}

func (h *HorizontalPodAutoscaler) lint(hpas map[string]autoscalingv1.HorizontalPodAutoscaler,
	dps map[string]appsv1.Deployment, sts map[string]appsv1.StatefulSet, acpu, amem resource.Quantity) {
	var tcpu, tmem resource.Quantity

	var current int32
	for id, hpa := range hpas {
		h.initIssues(id)
		var rcpu, rmem resource.Quantity
		ns, _ := namespaced(id)
		switch hpa.Spec.ScaleTargetRef.Kind {
		case "Deployment":
			dp, ok := dps[fqn(ns, hpa.Spec.ScaleTargetRef.Name)]
			if !ok {
				h.addIssue(id, ErrorLevel, "Used?")
				continue
			}
			rcpu, rmem = podResources(dp.Spec.Template.Spec)
			current = dp.Status.AvailableReplicas
		case "StatefulSet":
			st, ok := sts[fqn(ns, hpa.Spec.ScaleTargetRef.Name)]
			if !ok {
				h.addIssue(id, ErrorLevel, "Used?")
				continue
			}
			rcpu, rmem = podResources(st.Spec.Template.Spec)
			current = st.Status.CurrentReplicas
		}

		trcpu, trmem := rcpu.Copy(), rmem.Copy()
		for i := 1; i < int(hpa.Spec.MaxReplicas-current); i++ {
			trcpu.Add(rcpu)
			trmem.Add(rmem)
		}

		if toMC(*trcpu) > toMC(acpu) {
			cpu := trcpu.Copy()
			cpu.Sub(acpu)
			h.addIssuef(id, WarnLevel, "Replicas (%d/%d) at burst will exceed cluster CPU capacity by %s", current, hpa.Spec.MaxReplicas, asMC(*cpu))
		}
		if toMB(*trmem) > toMB(amem) {
			mem := trmem.Copy()
			mem.Sub(amem)
			h.addIssuef(id, WarnLevel, "Replicas (%d/%d) at burst will exceed cluster MEM capacity by %s", current, hpa.Spec.MaxReplicas, asMB(*mem))
		}

		tcpu.Add(*trcpu)
		tmem.Add(*trmem)
	}

	if toMC(tcpu) > toMC(acpu) {
		cpu := tcpu.Copy()
		cpu.Sub(acpu)
		h.addIssuef("hpas", WarnLevel, "If ALL HPAs triggered, %s will exceed cluster CPU capacity by %s", asMC(tcpu), asMC(*cpu))
	}
	if toMB(tmem) > toMB(amem) {
		mem := tmem.Copy()
		mem.Sub(amem)
		h.addIssuef("hpas", WarnLevel, "If ALL HPAs triggered, %s will exceed cluster MEM capacity by %s", asMB(tmem), asMB(*mem))
	}
}

// ----------------------------------------------------------------------------
// Helpers...

func podResources(spec v1.PodSpec) (cpu, mem resource.Quantity) {
	for _, co := range spec.InitContainers {
		c, m, _ := containerResources(co)
		cpu.Add(c)
		mem.Add(m)
	}
	for _, co := range spec.Containers {
		c, m, _ := containerResources(co)
		cpu.Add(c)
		mem.Add(m)
	}

	return
}
