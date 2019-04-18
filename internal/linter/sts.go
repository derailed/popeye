package linter

import (
	"context"

	"github.com/derailed/popeye/internal/k8s"
	"github.com/rs/zerolog"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// StatefulSet represents a StatefulSet linter.
type StatefulSet struct {
	*Linter
}

// NewStatefulSet returns a new StatefulSet linter.
func NewStatefulSet(l Loader, log *zerolog.Logger) *StatefulSet {
	return &StatefulSet{NewLinter(l, log)}
}

// Lint a StatefulSet.
func (s *StatefulSet) Lint(ctx context.Context) error {
	sts, err := s.ListStatefulSets()
	if err != nil {
		return err
	}

	s.lint(sts)

	return nil
}

func (s *StatefulSet) lint(sts map[string]appsv1.StatefulSet) {
	for fqn, st := range sts {
		s.initIssues(fqn)
		s.checkStatefulSet(fqn, st)
		s.checkContainers(fqn, st)

		pmx := make(k8s.PodsMetrics)
		if err := podsMetrics(s, pmx); err != nil {
			continue
		}
		s.checkUtilization(fqn, st, pmx)
	}
}

func (s *StatefulSet) checkStatefulSet(fqn string, st appsv1.StatefulSet) {
	if st.Spec.Replicas == nil || (st.Spec.Replicas != nil && *st.Spec.Replicas == 0) {
		s.addIssue(fqn, InfoLevel, "Zero scale detected")
	}

	if st.Status.CurrentReplicas == 0 {
		s.addIssue(fqn, WarnLevel, "Used?")
	}

	if st.Status.CollisionCount != nil && *st.Status.CollisionCount > 0 {
		s.addIssuef(fqn, ErrorLevel, "ReplicaSet collisions detected %d", *st.Status.CollisionCount)
	}
}

func (s *StatefulSet) checkContainers(fqn string, st appsv1.StatefulSet) {
	spec := st.Spec.Template.Spec

	l := NewContainer(s, s.log)
	for _, co := range spec.InitContainers {
		l.lint(co, false)
	}

	for _, co := range spec.Containers {
		l.lint(co, false)
	}

	s.addContainerIssues(fqn, l.Issues())
}

func (s *StatefulSet) checkUtilization(fqn string, st appsv1.StatefulSet, pmx k8s.PodsMetrics) error {
	mx, err := statefulsetUsage(s, st, pmx)
	if err != nil {
		return err
	}

	// No resources bail!
	if mx.RequestedCPU.IsZero() && mx.RequestedMEM.IsZero() {
		return nil
	}

	cpuPerc := mx.ReqCPURatio()
	if cpuPerc > int64(s.CPUResourceLimits().Over) {
		s.addIssuef(fqn, WarnLevel, "CPU over allocated. Requested:%s - Current:%s (%s)", asMC(mx.RequestedCPU), asMC(mx.CurrentCPU), asPerc(cpuPerc))
	}

	if cpuPerc > 0 && cpuPerc < int64(s.CPUResourceLimits().Under) {
		s.addIssuef(fqn, WarnLevel, "CPU under allocated. Requested:%s - Current:%s (%s)", asMC(mx.RequestedCPU), asMC(mx.CurrentCPU), asPerc(cpuPerc))
	}

	memPerc := mx.ReqMEMRatio()
	if memPerc > int64(s.MEMResourceLimits().Over) {
		s.addIssuef(fqn, WarnLevel, "Memory over allocated. Requested:%s - Current:%s (%s)", asMB(mx.RequestedMEM), asMB(mx.CurrentMEM), asPerc(memPerc))
	}

	if memPerc > 0 && memPerc < int64(s.MEMResourceLimits().Under) {
		s.addIssuef(fqn, WarnLevel, "Memory under allocated. Requested:%s - Current:%s (%s)", asMB(mx.RequestedMEM), asMB(mx.CurrentMEM), asPerc(memPerc))
	}

	return nil
}

func statefulsetUsage(l Loader, st appsv1.StatefulSet, pmx k8s.PodsMetrics) (ConsumptionMetrics, error) {
	var mx ConsumptionMetrics

	sel, err := metav1.LabelSelectorAsSelector(st.Spec.Selector)
	if err != nil {
		return mx, err
	}
	pods, err := l.ListPodsByLabels(sel.String())
	if err != nil {
		return mx, err
	}

	rc, rm := podResources(st.Spec.Template.Spec)
	if st.Spec.Replicas != nil {
		for i := 0; i < int(*st.Spec.Replicas); i++ {
			mx.RequestedCPU.Add(rc)
			mx.RequestedMEM.Add(rm)
		}
	}

	for pfqn := range pods {
		if ccx, ok := pmx[pfqn]; ok {
			for _, cx := range ccx {
				mx.CurrentCPU.Add(cx.CurrentCPU)
				mx.CurrentMEM.Add(cx.CurrentMEM)
			}
		}
	}

	return mx, nil
}
