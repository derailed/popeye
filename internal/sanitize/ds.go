package sanitize

import (
	"context"
	"errors"

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

// NewDaemonSet returns a new DaemonSet sanitizer.
func NewDaemonSet(co *issues.Collector, lister DaemonSetLister) *DaemonSet {
	return &DaemonSet{
		Collector:       co,
		DaemonSetLister: lister,
	}
}

// Sanitize configmaps.
func (d *DaemonSet) Sanitize(ctx context.Context) error {
	over := pullOverAllocs(ctx)
	for fqn, ds := range d.ListDaemonSets() {
		d.InitOutcome(fqn)
		d.checkDeprecation(fqn, ds)
		d.checkContainers(fqn, ds.Spec.Template.Spec)

		pmx := k8s.PodsMetrics{}
		podsMetrics(d, pmx)
		d.checkUtilization(over, fqn, ds, pmx)
	}

	return nil
}

func (d *DaemonSet) checkDeprecation(fqn string, ds *appsv1.DaemonSet) {
	const current = "apps/v1"

	rev, err := resourceRev(fqn, ds.Annotations)
	if err != nil {
		rev = revFromLink(ds.SelfLink)
		if rev == "" {
			d.AddCode(404, fqn, errors.New("Unable to assert resource version"))
			return
		}
	}
	if rev != current {
		d.AddCode(403, fqn, "DaemonSet", rev, current)
	}
}

// CheckContainers runs thru deployment template and checks pod configuration.
func (d *DaemonSet) checkContainers(fqn string, spec v1.PodSpec) {
	c := NewContainer(fqn, d)
	for _, co := range spec.InitContainers {
		c.sanitize(co, false)
	}
	for _, co := range spec.Containers {
		c.sanitize(co, false)
	}
}

// CheckUtilization checks deployments requested resources vs current utilization.
func (d *DaemonSet) checkUtilization(over bool, fqn string, ds *appsv1.DaemonSet, pmx k8s.PodsMetrics) {
	mx := d.daemonsetUsage(ds, pmx)
	if mx.RequestCPU.IsZero() && mx.RequestMEM.IsZero() {
		return
	}

	checkCPU(d, over, fqn, mx)
	checkMEM(d, over, fqn, mx)
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
