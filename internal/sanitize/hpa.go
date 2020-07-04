package sanitize

import (
	"context"

	"github.com/derailed/popeye/internal"
	"github.com/derailed/popeye/internal/cache"
	"github.com/derailed/popeye/internal/issues"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	mv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

type (
	// PodMetricsLister handles pods metrics.
	PodMetricsLister interface {
		ListPodsMetrics() map[string]*mv1beta1.PodMetrics
	}

	// ClusterMetricsLister handles cluster metrics.
	ClusterMetricsLister interface {
		ListAvailableMetrics(map[string]*v1.Node) v1.ResourceList
	}

	// HorizontalPodAutoscaler represents a HorizontalPodAutoscaler linter.
	HorizontalPodAutoscaler struct {
		*issues.Collector
		HpaLister
	}

	// HpaLister list available hpas on a cluster.
	HpaLister interface {
		NodeLister
		DeploymentLister
		StatefulSetLister
		ClusterMetricsLister
		ListHorizontalPodAutoscalers() map[string]*autoscalingv1.HorizontalPodAutoscaler
	}
)

// NewHorizontalPodAutoscaler returns a new ServiceAccount linter.
func NewHorizontalPodAutoscaler(co *issues.Collector, lister HpaLister) *HorizontalPodAutoscaler {
	return &HorizontalPodAutoscaler{
		Collector: co,
		HpaLister: lister,
	}
}

// Sanitize an horizontalpodautoscaler.
func (h *HorizontalPodAutoscaler) Sanitize(ctx context.Context) error {
	var (
		tcpu, tmem resource.Quantity
		current    int32
	)
	res := h.ListAvailableMetrics(h.ListNodes())
	for fqn, hpa := range h.ListHorizontalPodAutoscalers() {
		h.InitOutcome(fqn)
		ctx = internal.WithFQN(ctx, fqn)
		var rcpu, rmem resource.Quantity
		ns, _ := namespaced(fqn)
		switch hpa.Spec.ScaleTargetRef.Kind {
		case "Deployment":
			dpFqn, dps := cache.FQN(ns, hpa.Spec.ScaleTargetRef.Name), h.ListDeployments()
			if dp, ok := dps[dpFqn]; ok {
				rcpu, rmem = podResources(dp.Spec.Template.Spec)
				current = dp.Status.AvailableReplicas
			} else {
				h.AddCode(ctx, 600, fqn, dpFqn)
				continue
			}
		case "StatefulSet":
			stsFqn, sts := cache.FQN(ns, hpa.Spec.ScaleTargetRef.Name), h.ListStatefulSets()
			if st, ok := sts[stsFqn]; ok {
				rcpu, rmem = podResources(st.Spec.Template.Spec)
				current = st.Status.CurrentReplicas
			} else {
				h.AddCode(ctx, 601, fqn, stsFqn)
				continue
			}
		}

		rList := v1.ResourceList{v1.ResourceCPU: rcpu, v1.ResourceMemory: rmem}
		list := h.checkResources(ctx, hpa.Spec.MaxReplicas, current, rList, res)
		tcpu.Add(*list.Cpu())
		tmem.Add(*list.Memory())

		if h.NoConcerns(fqn) && h.Config.ExcludeFQN(internal.MustExtractSectionGVR(ctx), fqn) {
			h.ClearOutcome(fqn)
		}
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
	ctx = internal.WithFQN(ctx, "HPA")
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
