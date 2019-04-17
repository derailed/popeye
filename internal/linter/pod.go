package linter

import (
	"context"

	"github.com/derailed/popeye/internal/k8s"
	"github.com/rs/zerolog"
	v1 "k8s.io/api/core/v1"
	mv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

// Pod checks
// o check for naked pod ie no dep, rs, sts, cron
// o check for label existence
// o Recommended labels
// app.kubernetes.io/name
// app.kubernetes.io/instance
// app.kubernetes.io/version
// app.kubernetes.io/component
// app.kubernetes.io/part-of
// app.kubernetes.io/managed-by

const (
	cpuPodLimit = 80
	memPodLimit = 80
)

type (
	// Pod represents a Pod linter.
	Pod struct {
		*Linter
	}

	// PodMetric tracks node metrics available and current range.
	PodMetric interface {
		CurrentCPU() int64
		CurrentMEM() int64
		Empty() bool
	}
)

// NewPod returns a new pod linter.
func NewPod(l Loader, log *zerolog.Logger) *Pod {
	return &Pod{NewLinter(l, log)}
}

// Lint a Pod.
func (p *Pod) Lint(ctx context.Context) error {
	pods, err := p.ListPods()
	if err != nil {
		return err
	}

	var mx []mv1beta1.PodMetrics
	pmx := make(k8s.PodsMetrics)
	if ok, _ := p.ClusterHasMetrics(); ok {
		if mx, err = p.FetchPodsMetrics(""); err != nil {
			return err
		}
		p.ListPodsMetrics(mx, pmx)
	}

	for fqn, po := range pods {
		p.initIssues(fqn)
		p.lint(po, pmx[fqn])
	}

	return nil
}

func (p *Pod) lint(po v1.Pod, mx k8s.ContainerMetrics) {
	p.checkStatus(po)
	p.checkContainerStatus(po)
	p.checkContainers(po)
	p.checkServiceAccount(po)
	p.checkUtilization(po, mx)
}

func (p *Pod) checkUtilization(po v1.Pod, mx k8s.ContainerMetrics) {
	if len(mx) == 0 {
		return
	}

	for _, co := range po.Spec.Containers {
		cmx, ok := mx[co.Name]
		if !ok {
			continue
		}
		c := NewContainer(p.Loader, p.log)
		c.checkUtilization(co, cmx)

		p.addContainerIssues(metaFQN(po.ObjectMeta), c.Issues())
	}
}

func (p *Pod) checkServiceAccount(po v1.Pod) {
	if len(po.Spec.ServiceAccountName) == 0 {
		p.addIssuef(metaFQN(po.ObjectMeta), InfoLevel, "No service account specified")
	}
}

func (p *Pod) checkContainers(po v1.Pod) {
	l := NewContainer(p.Loader, p.log)
	for _, c := range po.Spec.InitContainers {
		l.lint(c, false)
	}

	for _, c := range po.Spec.Containers {
		l.lint(c, !isPartOfJob(po))
	}

	p.addContainerIssues(metaFQN(po.ObjectMeta), l.Issues())
}

func (p *Pod) checkContainerStatus(po v1.Pod) {
	limit := p.RestartsLimit()

	if len(po.Status.InitContainerStatuses) != 0 {
		for _, s := range po.Status.InitContainerStatuses {
			counts := new(containerStatusCount)
			counts.rollup(s)
			if issue := counts.diagnose(len(po.Status.InitContainerStatuses), limit, true); issue != nil {
				p.addIssues(metaFQN(po.ObjectMeta), issue)
				return
			}
		}
	}

	for _, s := range po.Status.ContainerStatuses {
		counts := new(containerStatusCount)
		counts.rollup(s)
		if issue := counts.diagnose(len(po.Status.ContainerStatuses), limit, false); issue != nil {
			p.addIssues(metaFQN(po.ObjectMeta), issue)
		}
	}
}

func (p *Pod) checkStatus(po v1.Pod) {
	switch po.Status.Phase {
	case v1.PodRunning:
	case v1.PodSucceeded:
	default:
		p.addIssuef(metaFQN(po.ObjectMeta), ErrorLevel, "Pod is in an unhappy phase (%s)", po.Status.Phase)
	}
}

// ----------------------------------------------------------------------------
// Helpers...

func isPartOfJob(po v1.Pod) bool {
	for _, o := range po.OwnerReferences {
		if o.Kind == "Job" {
			return true
		}
	}

	return false
}

func podsMetrics(l Loader, pmx k8s.PodsMetrics) error {
	var (
		mx  []mv1beta1.PodMetrics
		err error
	)
	if ok, _ := l.ClusterHasMetrics(); ok {
		if mx, err = l.FetchPodsMetrics(""); err != nil {
			return err
		}
		l.ListPodsMetrics(mx, pmx)
	}

	return nil
}
