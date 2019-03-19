package linter

import (
	v1 "k8s.io/api/core/v1"
)

const (
	// BOZO!! Set in a config file?
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
func NewPod() *Pod {
	return &Pod{new(Linter)}
}

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

// Lint a Pod.
func (p *Pod) Lint(po v1.Pod, mx map[string]PodMetric) {
	p.checkStatus(po.Status)

	if len(po.Spec.InitContainers) > 0 {
		p.checkContainerStatus(po.Status.InitContainerStatuses, true)
	}
	p.checkContainerStatus(po.Status.ContainerStatuses, false)

	p.checkContainers(po.Spec.Containers)
	p.checkServiceAccount(po.Spec)
	p.checkUtilization(po, mx)
}

func (p *Pod) checkUtilization(po v1.Pod, mx map[string]PodMetric) {
	if mx == nil || len(mx) == 0 {
		return
	}

	for _, co := range po.Spec.InitContainers {
		cmx, ok := mx[co.Name]
		if !ok {
			continue
		}
		l := NewContainer()
		l.checkUtilization(co, cmx)
		p.addIssues(l.Issues()...)
	}

	for _, co := range po.Spec.Containers {
		cmx, ok := mx[co.Name]
		if !ok {
			continue
		}
		l := NewContainer()
		l.checkUtilization(co, cmx)
		p.addIssues(l.Issues()...)
	}
}

func (p *Pod) checkServiceAccount(spec v1.PodSpec) {
	if len(spec.ServiceAccountName) == 0 {
		p.addIssuef(InfoLevel, "No service account specified")
	}
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

	if issue := counts.diagnose(len(ss), isInit); issue != nil {
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
