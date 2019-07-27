package sanitize

import (
	"context"
	"errors"

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
		s.InitOutcome(fqn)
		s.checkDeprecation(fqn, st)
		s.checkStatefulSet(fqn, st)
		s.checkContainers(fqn, st)
		s.checkUtilization(over, fqn, st, pmx)
	}

	return nil
}

func (s *StatefulSet) checkDeprecation(fqn string, st *appsv1.StatefulSet) {
	const current = "apps/v1"

	rev, err := resourceRev(fqn, st.Annotations)
	if err != nil {
		rev = revFromLink(st.SelfLink)
		if rev == "" {
			s.AddCode(404, fqn, errors.New("Unable to assert resource version"))
			return
		}
	}

	if rev != current {
		s.AddCode(403, fqn, "StatefulSet", rev, current)
	}
}

func (s *StatefulSet) checkStatefulSet(fqn string, st *appsv1.StatefulSet) {
	if st.Spec.Replicas == nil || (st.Spec.Replicas != nil && *st.Spec.Replicas == 0) {
		s.AddCode(500, fqn)
	}

	if st.Status.CurrentReplicas == 0 {
		s.AddCode(501, fqn)
	}

	if st.Status.CollisionCount != nil && *st.Status.CollisionCount > 0 {
		s.AddCode(502, fqn, *st.Status.CollisionCount)
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
		s.AddCode(503, fqn, asMC(mx.CurrentCPU), asMC(mx.RequestCPU), asPerc(cpuPerc))
	} else if over && cpuPerc > 0 && cpuPerc < float64(s.CPUResourceLimits().OverPerc) {
		s.AddCode(504, fqn, asMC(mx.CurrentCPU), asMC(mx.RequestCPU), asPerc(cpuPerc))
	}

	memPerc := mx.ReqMEMRatio()
	if memPerc > float64(s.MEMResourceLimits().UnderPerc) {
		s.AddCode(505, fqn, asMB(mx.CurrentMEM), asMB(mx.RequestMEM), asPerc(memPerc))
	} else if over && memPerc > 0 && memPerc < float64(s.MEMResourceLimits().OverPerc) {
		s.AddCode(506, fqn, asMB(mx.CurrentMEM), asMB(mx.RequestMEM), asPerc(memPerc))
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
