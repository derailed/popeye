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

	for fqn, st := range s.ListStatefulSets() {
		s.checkStatefulSet(fqn, st)
		s.checkContainers(fqn, st)
		s.checkUtilization(fqn, st, pmx)
	}

	return nil
}

func (s *StatefulSet) checkStatefulSet(fqn string, st *appsv1.StatefulSet) {
	if st.Spec.Replicas == nil || (st.Spec.Replicas != nil && *st.Spec.Replicas == 0) {
		s.AddInfo(fqn, "Zero scale detected")
	}

	if st.Status.CurrentReplicas == 0 {
		s.AddWarn(fqn, "Used?")
	}

	if st.Status.CollisionCount != nil && *st.Status.CollisionCount > 0 {
		s.AddErrorf(fqn, "ReplicaSet collisions detected %d", *st.Status.CollisionCount)
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

func (s *StatefulSet) checkUtilization(fqn string, st *appsv1.StatefulSet, pmx k8s.PodsMetrics) error {
	mx, err := s.statefulsetUsage(st, pmx)
	if err != nil {
		return err
	}

	// No resources bail!
	if mx.RequestedCPU.IsZero() && mx.RequestedMEM.IsZero() {
		return nil
	}

	cpuPerc := mx.ReqCPURatio()
	if cpuPerc > int64(s.CPUResourceLimits().Over) {
		s.AddWarnf(fqn, "CPU over allocated. Requested:%s - Current:%s (%s)", asMC(mx.RequestedCPU), asMC(mx.CurrentCPU), asPerc(cpuPerc))
	}

	if cpuPerc > 0 && cpuPerc < int64(s.CPUResourceLimits().Under) {
		s.AddWarnf(fqn, "CPU under allocated. Requested:%s - Current:%s (%s)", asMC(mx.RequestedCPU), asMC(mx.CurrentCPU), asPerc(cpuPerc))
	}

	memPerc := mx.ReqMEMRatio()
	if memPerc > int64(s.MEMResourceLimits().Over) {
		s.AddWarnf(fqn, "Memory over allocated. Requested:%s - Current:%s (%s)", asMB(mx.RequestedMEM), asMB(mx.CurrentMEM), asPerc(memPerc))
	}

	if memPerc > 0 && memPerc < int64(s.MEMResourceLimits().Under) {
		s.AddWarnf(fqn, "Memory under allocated. Requested:%s - Current:%s (%s)", asMB(mx.RequestedMEM), asMB(mx.CurrentMEM), asPerc(memPerc))
	}

	return nil
}

func (s *StatefulSet) statefulsetUsage(st *appsv1.StatefulSet, pmx k8s.PodsMetrics) (ConsumptionMetrics, error) {
	var mx ConsumptionMetrics
	rc, rm := podResources(st.Spec.Template.Spec)
	if st.Spec.Replicas != nil {
		for i := 0; i < int(*st.Spec.Replicas); i++ {
			mx.RequestedCPU.Add(rc)
			mx.RequestedMEM.Add(rm)
		}
	}

	for pfqn := range s.ListPodsBySelector(st.Spec.Selector) {
		if ccx, ok := pmx[pfqn]; ok {
			for _, cx := range ccx {
				mx.CurrentCPU.Add(cx.CurrentCPU)
				mx.CurrentMEM.Add(cx.CurrentMEM)
			}
		}
	}

	return mx, nil
}
