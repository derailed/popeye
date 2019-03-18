package linter

import (
	v1 "k8s.io/api/core/v1"
)

// Pod represents a Pod linter.
type Pod struct {
	*Linter
}

// NewPod returns a new pod linter.
func NewPod() *Pod {
	return &Pod{new(Linter)}
}

// Pod checks
// + Running?
// + Restarts?
// + Resources
// + Probes
// o Metrics current vs set
// o Named ports
// o check container image tags
// o check for service accounts
// o check for naked pod ie no dep, rs, sts, cron
// o check for label existence
// o Recommended labels
// app.kubernetes.io/name
// app.kubernetes.io/instance
// app.kubernetes.io/version
// app.kubernetes.io/component
// app.kubernetes.io/part-of
// app.kubernetes.io/managed-by

// Lint a Pod.
func (p *Pod) Lint(po v1.Pod) {
	p.checkStatus(po.Status)

	// Check init containers status
	if len(po.Spec.InitContainers) > 0 {
		p.checkContainerStatus(po.Status.InitContainerStatuses, true)
	}
	p.checkContainerStatus(po.Status.ContainerStatuses, false)

	// Check for resources and probes
	p.checkContainers(po.Spec.Containers)
}

func (p *Pod) checkProbes(cc []v1.Container) {
	for _, c := range cc {
		if c.LivenessProbe == nil {
			p.addIssuef(InfoLevel, "%s container has no liveness probe", c.Name)
		}
		if c.ReadinessProbe == nil {
			p.addIssuef(InfoLevel, "%s container has no readiness probe", c.Name)
		}
	}
}

func (p *Pod) checkContainers(cc []v1.Container) {
	for _, c := range cc {
		l := NewContainer()
		l.Lint(c)
		p.addIssues(l.Issues()...)
	}
}

func (p *Pod) checkContainerStatus(ss []v1.ContainerStatus, isInit bool) {
	c := "container"
	if isInit {
		c = "init" + c
	}

	counts := new(containerStatusCount)
	for _, s := range ss {
		counts.rollup(s)
	}

	if issue := counts.diagnose(len(ss)); issue != nil {
		p.addIssues(issue)
	}
}

func (p *Pod) checkStatus(status v1.PodStatus) {
	switch status.Phase {
	case v1.PodRunning:
	case v1.PodSucceeded:
	default:
		p.addIssuef(ErrorLevel, "Pod is in an unhappy phase (%s)", status.Phase)
	}
}
