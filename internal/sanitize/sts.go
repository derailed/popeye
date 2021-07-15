package sanitize

import (
	"context"

	"github.com/derailed/popeye/internal"
	"github.com/derailed/popeye/internal/client"
	"github.com/derailed/popeye/internal/issues"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
)

type (
	// CollectorLimiter represents a collector with resource allocation limits.
	CollectorLimiter interface {
		// Collector an issue collector.
		Collector

		// ConfigLister track config options.
		ConfigLister
	}

	// StatefulSetLister handles statefulsets.
	StatefulSetLister interface {
		PodLimiter
		ConfigLister
		PodSelectorLister
		PodsMetricsLister

		ListStatefulSets() map[string]*appsv1.StatefulSet
		ListServiceAccounts() map[string]*v1.ServiceAccount
	}

	// StatefulSet represents a StatefulSet sanitizer.
	StatefulSet struct {
		*issues.Collector
		StatefulSetLister
	}
)

// NewStatefulSet returns a new sanitizer.
func NewStatefulSet(co *issues.Collector, lister StatefulSetLister) *StatefulSet {
	return &StatefulSet{
		Collector:         co,
		StatefulSetLister: lister,
	}
}

// Sanitize cleanse the resource.
func (s *StatefulSet) Sanitize(ctx context.Context) error {
	pmx := client.PodsMetrics{}
	podsMetrics(s, pmx)

	over := pullOverAllocs(ctx)
	for fqn, st := range s.ListStatefulSets() {
		s.InitOutcome(fqn)
		ctx = internal.WithFQN(ctx, fqn)

		s.checkDeprecation(ctx, st)
		s.checkStatefulSet(ctx, st)
		s.checkContainers(ctx, st)
		s.checkUtilization(ctx, over, st, pmx)

		if s.NoConcerns(fqn) && s.Config.ExcludeFQN(internal.MustExtractSectionGVR(ctx), fqn) {
			s.ClearOutcome(fqn)
		}
	}

	return nil
}

func (s *StatefulSet) checkDeprecation(ctx context.Context, st *appsv1.StatefulSet) {
	const current = "apps/v1"

	rev, err := resourceRev(internal.MustExtractFQN(ctx), "StatefulSet", st.Annotations)
	if err != nil {
		if rev = revFromLink(st.SelfLink); rev == "" {
			return
		}
	}

	if rev != current {
		s.AddCode(ctx, 403, "StatefulSet", rev, current)
	}
}

func (s *StatefulSet) checkStatefulSet(ctx context.Context, sts *appsv1.StatefulSet) {
	if sts.Spec.Replicas == nil || (sts.Spec.Replicas != nil && *sts.Spec.Replicas == 0) {
		s.AddCode(ctx, 500)
		return
	}

	if sts.Spec.Replicas != nil && *sts.Spec.Replicas != sts.Status.ReadyReplicas {
		s.AddCode(ctx, 501, *sts.Spec.Replicas, sts.Status.ReadyReplicas)
	}

	if sts.Spec.Template.Spec.ServiceAccountName == "" {
		return
	}

	if _, ok := s.ListServiceAccounts()[client.FQN(sts.Namespace, sts.Spec.Template.Spec.ServiceAccountName)]; !ok {
		s.AddCode(ctx, 507, sts.Spec.Template.Spec.ServiceAccountName)
	}

}

func (s *StatefulSet) checkContainers(ctx context.Context, st *appsv1.StatefulSet) {
	spec := st.Spec.Template.Spec

	l := NewContainer(internal.MustExtractFQN(ctx), s)
	for _, co := range spec.InitContainers {
		l.sanitize(ctx, co, false)
	}

	for _, co := range spec.Containers {
		l.sanitize(ctx, co, false)
	}
}

func checkCPU(ctx context.Context, c CollectorLimiter, over bool, mx ConsumptionMetrics) {
	cpuPerc := mx.ReqCPURatio()
	if cpuPerc > 1 && cpuPerc > float64(c.CPUResourceLimits().UnderPerc) {
		c.AddCode(ctx, 503, asMC(mx.CurrentCPU), asMC(mx.RequestCPU), asPerc(cpuPerc))
		return
	}

	if over && cpuPerc > 0 && cpuPerc < float64(c.CPUResourceLimits().OverPerc) {
		c.AddCode(ctx, 504, asMC(mx.CurrentCPU), asMC(mx.RequestCPU), asPerc(mx.ReqAbsCPURatio()))
	}
}

func checkMEM(ctx context.Context, c CollectorLimiter, over bool, mx ConsumptionMetrics) {
	memPerc := mx.ReqMEMRatio()
	if memPerc > 1 && memPerc > float64(c.MEMResourceLimits().UnderPerc) {
		c.AddCode(ctx, 505, asMB(mx.CurrentMEM), asMB(mx.RequestMEM), asPerc(memPerc))
		return
	}

	if over && memPerc < float64(c.MEMResourceLimits().OverPerc) {
		c.AddCode(ctx, 506, asMB(mx.CurrentMEM), asMB(mx.RequestMEM), asPerc(mx.ReqAbsMEMRatio()))
	}
}

func (s *StatefulSet) checkUtilization(ctx context.Context, over bool, st *appsv1.StatefulSet, pmx client.PodsMetrics) {
	mx := s.statefulsetUsage(st, pmx)
	if mx.RequestCPU.IsZero() && mx.RequestMEM.IsZero() {
		return
	}

	checkCPU(ctx, s, over, mx)
	checkMEM(ctx, s, over, mx)
}

func (s *StatefulSet) statefulsetUsage(st *appsv1.StatefulSet, pmx client.PodsMetrics) ConsumptionMetrics {
	var mx ConsumptionMetrics
	for pfqn, pod := range s.ListPodsBySelector(st.Namespace, st.Spec.Selector) {
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
