package sanitize

import (
	"context"

	"github.com/derailed/popeye/internal/cache"
	"github.com/derailed/popeye/internal/issues"
	"github.com/derailed/popeye/internal/k8s"
	v1 "k8s.io/api/core/v1"
	mv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

const (
	cpuPodLimit = 80
	memPodLimit = 80
)

type (
	// Pod represents a Pod linter.
	Pod struct {
		*issues.Collector
		PodMXLister
	}

	// PodMXLister list available pods.
	PodMXLister interface {
		PodLimiter
		PodMetricsLister
		PodLister
	}

	// PodMetric tracks node metrics available and current range.
	PodMetric interface {
		CurrentCPU() int64
		CurrentMEM() int64
		Empty() bool
	}
)

// NewPod returns a new pod linter.
func NewPod(co *issues.Collector, lister PodMXLister) *Pod {
	return &Pod{
		Collector:   co,
		PodMXLister: lister,
	}
}

// Sanitize a Pod.
func (p *Pod) Sanitize(ctx context.Context) error {
	for fqn, po := range p.ListPods() {
		p.InitOutcome(fqn)
		p.checkStatus(po)
		p.checkContainerStatus(fqn, po)
		p.checkContainers(fqn, po)
		p.checkServiceAccount(fqn, po.Spec.ServiceAccountName)
		pmx := p.ListPodsMetrics()[fqn]
		cmx := k8s.ContainerMetrics{}
		containerMetrics(fqn, pmx, cmx)
		p.checkUtilization(fqn, po, cmx)
		// p.checkReferences(po)
	}
	return nil
}

func (p *Pod) checkUtilization(fqn string, po *v1.Pod, mx k8s.ContainerMetrics) {
	if len(mx) == 0 {
		return
	}
	for _, co := range po.Spec.Containers {
		cmx, ok := mx[co.Name]
		if !ok {
			continue
		}
		c := NewContainer(fqn, p)
		c.checkUtilization(co, cmx)
	}
}

func (p *Pod) checkServiceAccount(fqn, sa string) {
	if len(sa) == 0 {
		p.AddInfo(fqn, "No service account specified")
	}
}

func (p *Pod) checkContainers(fqn string, po *v1.Pod) {
	l := NewContainer(fqn, p)
	for _, c := range po.Spec.InitContainers {
		l.sanitize(c, false)
	}
	for _, c := range po.Spec.Containers {
		l.sanitize(c, !isPartOfJob(po))
	}
}

func (p *Pod) checkContainerStatus(fqn string, po *v1.Pod) {
	limit := p.RestartsLimit()
	for _, s := range po.Status.InitContainerStatuses {
		cs := newContainerStatus(p, fqn, len(po.Status.InitContainerStatuses), true, limit)
		cs.sanitize(s)
	}

	for _, s := range po.Status.ContainerStatuses {
		cs := newContainerStatus(p, fqn, len(po.Status.ContainerStatuses), false, limit)
		cs.sanitize(s)
	}
}

func (p *Pod) checkStatus(po *v1.Pod) {
	switch po.Status.Phase {
	case v1.PodRunning:
	case v1.PodSucceeded:
	default:
		p.AddErrorf(cache.MetaFQN(po.ObjectMeta), "Pod is in an unhappy phase (%s)", po.Status.Phase)
	}
}

// ----------------------------------------------------------------------------
// Helpers...

func containerMetrics(fqn string, pmx *mv1beta1.PodMetrics, mx k8s.ContainerMetrics) {
	for _, co := range pmx.Containers {
		mx[co.Name] = k8s.Metrics{
			CurrentCPU: *co.Usage.Cpu(),
			CurrentMEM: *co.Usage.Memory(),
		}
	}
}

func isPartOfJob(po *v1.Pod) bool {
	for _, o := range po.OwnerReferences {
		if o.Kind == "Job" {
			return true
		}
	}

	return false
}
