package linter

import (
	"context"

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
	pmx := make(PodsMetrics)
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

func (p *Pod) lint(po v1.Pod, mx ContainerMetrics) {
	p.checkStatus(po)
	p.checkContainerStatus(po)
	p.checkContainers(po)
	p.checkServiceAccount(po)
	p.checkUtilization(po, mx)
}

func (p *Pod) checkUtilization(po v1.Pod, mx ContainerMetrics) {
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

		p.addIssuesMap(podFQN(po), c.Issues())
	}
}

func (p *Pod) checkServiceAccount(po v1.Pod) {
	if len(po.Spec.ServiceAccountName) == 0 {
		p.addIssuef(podFQN(po), InfoLevel, "No service account specified")
	}
}

func (p *Pod) checkContainers(po v1.Pod) {
	for _, c := range po.Spec.Containers {
		l := NewContainer(p.Loader, p.log)
		l.lint(c, isPartOfJob(po))

		p.addIssuesMap(podFQN(po), l.Issues())
	}
}

func (p *Pod) checkContainerStatus(po v1.Pod) {
	limit := p.RestartsLimit()

	if len(po.Status.InitContainerStatuses) != 0 {

		for _, s := range po.Status.InitContainerStatuses {
			counts := new(containerStatusCount)
			counts.rollup(s)
			if issue := counts.diagnose(len(po.Status.InitContainerStatuses), limit, true); issue != nil {
				p.addIssues(podFQN(po), issue)
				return
			}
		}
	}

	for _, s := range po.Status.ContainerStatuses {
		counts := new(containerStatusCount)
		counts.rollup(s)
		if issue := counts.diagnose(len(po.Status.ContainerStatuses), limit, false); issue != nil {
			p.addIssues(podFQN(po), issue)
		}
	}
}

func (p *Pod) checkStatus(po v1.Pod) {
	switch po.Status.Phase {
	case v1.PodRunning:
	case v1.PodSucceeded:
	default:
		p.addIssuef(podFQN(po), ErrorLevel, "Pod is in an unhappy phase (%s)", po.Status.Phase)
	}
}

// ----------------------------------------------------------------------------
// Helpers...

func podFQN(po v1.Pod) string {
	return po.Namespace + "/" + po.Name
}

func isPartOfJob(po v1.Pod) bool {
	for _, o := range po.OwnerReferences {
		if o.Kind == "Job" {
			return true
		}
	}

	return false
}
