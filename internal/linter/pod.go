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
func NewPod(c Client, l *zerolog.Logger) *Pod {
	return &Pod{newLinter(c, l)}
}

// Lint a Pod.
func (p *Pod) Lint(ctx context.Context) error {
	ll, err := p.client.ListPods()
	if err != nil {
		return err
	}

	var mx []mv1beta1.PodMetrics
	pmx := make(k8s.PodsMetrics)
	if p.client.ClusterHasMetrics() {
		if mx, err = p.client.FetchPodsMetrics(""); err != nil {
			return err
		}
		k8s.GetPodsMetrics(mx, pmx)
	}

	for nsed, po := range ll {
		p.initIssues(nsed)
		p.lint(po, pmx[nsed])
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
		c := NewContainer(p.client, p.log)
		c.checkUtilization(co, cmx)
		p.addIssuesMap(nsFQN(po), c.Issues())
	}
}

func (p *Pod) checkServiceAccount(po v1.Pod) {
	if len(po.Spec.ServiceAccountName) == 0 {
		p.addIssuef(nsFQN(po), InfoLevel, "No service account specified")
	}
}

func (p *Pod) checkContainers(po v1.Pod) {
	for _, c := range po.Spec.Containers {
		l := NewContainer(p.client, p.log)
		l.lint(c, isPartOfJob(po))
		p.addIssuesMap(nsFQN(po), l.Issues())
	}
}

func (p *Pod) checkContainerStatus(po v1.Pod) {
	limit := p.client.RestartsLimit()

	if len(po.Status.InitContainerStatuses) != 0 {
		counts := new(containerStatusCount)
		for _, s := range po.Status.InitContainerStatuses {
			counts.rollup(s)
		}
		if issue := counts.diagnose(len(po.Status.InitContainerStatuses), limit, true); issue != nil {
			p.addIssues(nsFQN(po), issue)
			return
		}
	}

	counts := new(containerStatusCount)
	for _, s := range po.Status.ContainerStatuses {
		counts.rollup(s)
	}
	if issue := counts.diagnose(len(po.Status.ContainerStatuses), limit, false); issue != nil {
		p.addIssues(nsFQN(po), issue)
	}
}

func (p *Pod) checkStatus(po v1.Pod) {
	switch po.Status.Phase {
	case v1.PodRunning:
	case v1.PodSucceeded:
	default:
		p.addIssuef(nsFQN(po), ErrorLevel, "Pod is in an unhappy phase (%s)", po.Status.Phase)
	}
}

// ----------------------------------------------------------------------------
// Helpers...

func nsFQN(po v1.Pod) string {
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
