package sanitize

import (
	"context"

	"github.com/derailed/popeye/internal/issues"
	"github.com/derailed/popeye/internal/k8s"
	appsv1 "k8s.io/api/apps/v1"
)

type (
	// StatefulSetLister handles statefulsets.
	StatefulSetLister interface {
		PodLimiter
		ConfigLister
		PodSelectorLister
		PodsMetricsLister

		ListStatefulSets() map[string]*appsv1.StatefulSet
	}

	// StatefulSet represents a StatefulSet sanitizer.
	StatefulSet struct {
		*issues.Collector
		StatefulSetLister
	}
)

// NewStatefulSet returns a new StatefulSet linter.
func NewStatefulSet(co *issues.Collector, lister StatefulSetLister) *StatefulSet {
	return &StatefulSet{
		Collector:         co,
		StatefulSetLister: lister,
	}
}

// Sanitize a StatefulSet.
func (s *StatefulSet) Sanitize(ctx context.Context) error {
	pmx := k8s.PodsMetrics{}
	podsMetrics(s, pmx)

	over := pullOverAllocs(ctx)
	for fqn, st := range s.ListStatefulSets() {
		s.checkStatefulSet(fqn, st)
		s.checkContainers(fqn, st)
		s.checkUtilization(over, fqn, st, pmx)
	}

	return nil
}

func (s *StatefulSet) checkStatefulSet(fqn string, st *appsv1.StatefulSet) {
	if st.Spec.Replicas == nil || (st.Spec.Replicas != nil && *st.Spec.Replicas == 0) {
		s.AddInfo(fqn, "Zero scale detected")
	}

	if st.Status.CurrentReplicas == 0 {
		s.AddWarn(fqn, "Used? No available replicas found")
	}

	if st.Status.CollisionCount != nil && *st.Status.CollisionCount > 0 {
		s.AddErrorf(fqn, "ReplicaSet collisions detected (%d)", *st.Status.CollisionCount)
	}
}

func (s *StatefulSet) checkContainers(fqn string, st *appsv1.StatefulSet) {
	spec := st.Spec.Template.Spec

	l := NewContainer(fqn, s)
	for _, co := range spec.InitContainers {
		l.sanitize(co, false)
	}

	for _, co := range spec.Containers {
		l.sanitize(co, false)
	}
}

func (s *StatefulSet) checkUtilization(over bool, fqn string, st *appsv1.StatefulSet, pmx k8s.PodsMetrics) error {
	mx := s.statefulsetUsage(st, pmx)
	if mx.RequestCPU.IsZero() && mx.RequestMEM.IsZero() {
		return nil
	}

	cpuPerc := mx.ReqCPURatio()
	if cpuPerc > float64(s.CPUResourceLimits().UnderPerc) {
		s.AddWarnf(fqn, utilFmt, "CPU under allocated", asMC(mx.CurrentCPU), asMC(mx.RequestCPU), asPerc(cpuPerc))
	} else if over && cpuPerc > 0 && cpuPerc < float64(s.CPUResourceLimits().OverPerc) {
		s.AddWarnf(fqn, utilFmt, "CPU over allocated", asMC(mx.CurrentCPU), asMC(mx.RequestCPU), asPerc(cpuPerc))
	}

	memPerc := mx.ReqMEMRatio()
	if memPerc > float64(s.MEMResourceLimits().UnderPerc) {
		s.AddWarnf(fqn, utilFmt, "Memory under allocated", asMB(mx.CurrentMEM), asMB(mx.RequestMEM), asPerc(memPerc))
	} else if over && memPerc > 0 && memPerc < float64(s.MEMResourceLimits().OverPerc) {
		s.AddWarnf(fqn, utilFmt, "Memory over allocated", asMB(mx.CurrentMEM), asMB(mx.RequestMEM), asPerc(memPerc))
	}

	return nil
}

func (s *StatefulSet) statefulsetUsage(st *appsv1.StatefulSet, pmx k8s.PodsMetrics) ConsumptionMetrics {
	var mx ConsumptionMetrics
	for pfqn, pod := range s.ListPodsBySelector(st.Spec.Selector) {
		cpu, mem := computePodResources(pod.Spec)
		mx.QOS = pod.Status.QOSClass
		mx.RequestCPU.Add(cpu)
		mx.RequestMEM.Add(mem)

		ccx, ok := pmx[pfqn]
		if !ok {
			continue
		}
		for _, cx := range ccx {
			mx.CurrentCPU.Add(cx.CurrentCPU)
			mx.CurrentMEM.Add(cx.CurrentMEM)
		}
	}

	return mx
}
