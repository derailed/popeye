package sanitize

import (
	"context"

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
		ListClusterMetrics() v1.ResourceList
	}

	// HorizontalPodAutoscaler represents a HorizontalPodAutoscaler linter.
	HorizontalPodAutoscaler struct {
		*issues.Collector
		HpaLister
	}

	// HpaLister list available hpas on a cluster.
	HpaLister interface {
		DeployLister
		StatefulSetLister
		ClusterMetricsLister
		NodeMetricsLister
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
	res := h.ListClusterMetrics()
	for fqn, hpa := range h.ListHorizontalPodAutoscalers() {
		h.InitOutcome(fqn)
		var rcpu, rmem resource.Quantity
		ns, _ := namespaced(fqn)
		switch hpa.Spec.ScaleTargetRef.Kind {
		case "Deployment":
			dfqn, dps := cache.FQN(ns, hpa.Spec.ScaleTargetRef.Name), h.ListDeployments()
			if dp, ok := dps[dfqn]; ok {
				rcpu, rmem, _ = podResources(dp.Spec.Template.Spec)
				current = dp.Status.AvailableReplicas
			} else {
				h.AddErrorf(fqn, "HPA %s references a deployment %s which does not exist", fqn, dfqn)
				continue
			}
		case "StatefulSet":
			sfqn, sts := cache.FQN(ns, hpa.Spec.ScaleTargetRef.Name), h.ListStatefulSets()
			if st, ok := sts[sfqn]; ok {
				rcpu, rmem, _ = podResources(st.Spec.Template.Spec)
				current = st.Status.CurrentReplicas
			} else {
				h.AddErrorf(fqn, "HPA %s references a statefulset %s which does not exist", fqn, sfqn)
				continue
			}
		}
		cpu, mem := h.checkResources(fqn, hpa.Spec.MaxReplicas, current, rcpu, rmem, res)
		tcpu.Add(*cpu)
		tmem.Add(*mem)

	}
	h.checkUtilization(tcpu, tmem, res)

	return nil
}

func (h *HorizontalPodAutoscaler) checkResources(fqn string, max, current int32, rcpu, rmem resource.Quantity, res v1.ResourceList) (tcpu, tmem *resource.Quantity) {
	acpu, amem := *res.Cpu(), *res.Memory()
	trcpu, trmem := rcpu.Copy(), rmem.Copy()
	for i := 1; i <= int(max-current); i++ {
		trcpu.Add(rcpu)
		trmem.Add(rmem)
	}
	if toMC(*trcpu) > toMC(acpu) {
		cpu := trcpu.Copy()
		cpu.Sub(acpu)
		h.AddWarnf(fqn, "Replicas (%d/%d) at burst will match/exceed cluster CPU(%s) capacity by %s", current, max, asMC(acpu), asMC(*cpu))
	}
	if toMB(*trmem) > toMB(amem) {
		mem := trmem.Copy()
		mem.Sub(amem)
		h.AddWarnf(fqn, "Replicas (%d/%d) at burst will match/exceed cluster memory(%s) capacity by %s", current, max, asMB(amem), asMB(*mem))
	}

	return trcpu, trmem
}

func (h *HorizontalPodAutoscaler) checkUtilization(tcpu, tmem resource.Quantity, res v1.ResourceList) {
	acpu, amem := *res.Cpu(), *res.Memory()

	if toMC(tcpu) > toMC(acpu) {
		cpu := tcpu.Copy()
		cpu.Sub(acpu)
		h.AddWarnf("HPA", "If ALL HPAs triggered, %s will match/exceed cluster CPU(%s) capacity by %s", asMC(tcpu), asMC(acpu), asMC(*cpu))
	}
	if toMB(tmem) > toMB(amem) {
		mem := tmem.Copy()
		mem.Sub(amem)
		h.AddWarnf("HPA", "If ALL HPAs triggered, %s will match/exceed cluster memory(%s) capacity by %s", asMB(tmem), asMB(amem), asMB(*mem))
	}
}
