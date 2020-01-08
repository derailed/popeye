package sanitize

import (
	"context"
	"errors"

	"github.com/derailed/popeye/internal"
	"github.com/derailed/popeye/internal/issues"
	"github.com/derailed/popeye/internal/k8s"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
)

type (
	// DaemonSet tracks DaemonSet sanitization.
	DaemonSet struct {
		*issues.Collector
		DaemonSetLister
	}

	// DaemonLister list DaemonSets.
	DaemonLister interface {
		ListDaemonSets() map[string]*appsv1.DaemonSet
	}

	// DaemonSetLister list available DaemonSets on a cluster.
	DaemonSetLister interface {
		PodLimiter
		PodsMetricsLister
		PodSelectorLister
		ConfigLister
		DaemonLister
	}
)

// NewDaemonSet returns a new sanitizer.
func NewDaemonSet(co *issues.Collector, lister DaemonSetLister) *DaemonSet {
	return &DaemonSet{
		Collector:       co,
		DaemonSetLister: lister,
	}
}

// Sanitize cleanse the resource.
func (d *DaemonSet) Sanitize(ctx context.Context) error {
	over := pullOverAllocs(ctx)
	for fqn, ds := range d.ListDaemonSets() {
		d.InitOutcome(fqn)
		ctx = internal.WithFQN(ctx, fqn)

		d.checkDeprecation(ctx, ds)
		d.checkContainers(ctx, ds.Spec.Template.Spec)
		pmx := k8s.PodsMetrics{}
		podsMetrics(d, pmx)
		d.checkUtilization(ctx, over, ds, pmx)

		if d.NoConcerns(fqn) && d.Config.ExcludeFQN(internal.MustExtractSection(ctx), fqn) {
			d.ClearOutcome(fqn)
		}
	}

	return nil
}

func (d *DaemonSet) checkDeprecation(ctx context.Context, ds *appsv1.DaemonSet) {
	const current = "apps/v1"

	rev, err := resourceRev(internal.MustExtractFQN(ctx), ds.Annotations)
	if err != nil {
		rev = revFromLink(ds.SelfLink)
		if rev == "" {
			d.AddCode(ctx, 404, errors.New("Unable to assert resource version"))
			return
		}
	}
	if rev != current {
		d.AddCode(ctx, 403, "DaemonSet", rev, current)
	}
}

// CheckContainers runs thru deployment template and checks pod configuration.
func (d *DaemonSet) checkContainers(ctx context.Context, spec v1.PodSpec) {
	c := NewContainer(internal.MustExtractFQN(ctx), d)
	for _, co := range spec.InitContainers {
		c.sanitize(ctx, co, false)
	}
	for _, co := range spec.Containers {
		c.sanitize(ctx, co, false)
	}
}

// CheckUtilization checks deployments requested resources vs current utilization.
func (d *DaemonSet) checkUtilization(ctx context.Context, over bool, ds *appsv1.DaemonSet, pmx k8s.PodsMetrics) {
	mx := d.daemonsetUsage(ds, pmx)
	if mx.RequestCPU.IsZero() && mx.RequestMEM.IsZero() {
		return
	}

	checkCPU(ctx, d, over, mx)
	checkMEM(ctx, d, over, mx)
}

// DaemonSetUsage finds deployment running pods and compute current vs requested resource usage.
func (d *DaemonSet) daemonsetUsage(ds *appsv1.DaemonSet, pmx k8s.PodsMetrics) ConsumptionMetrics {
	var mx ConsumptionMetrics
	for pfqn, pod := range d.ListPodsBySelector(ds.Spec.Selector) {
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
