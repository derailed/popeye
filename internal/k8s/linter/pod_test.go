package linter

import (
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestPoCheckStatus(t *testing.T) {
	uu := []struct {
		phase    v1.PodPhase
		issues   int
		severity Level
	}{
		{phase: v1.PodPending, issues: 1, severity: ErrorLevel},
		{phase: v1.PodRunning, issues: 0},
		{phase: v1.PodSucceeded, issues: 0},
		{phase: v1.PodFailed, issues: 1, severity: ErrorLevel},
		{phase: v1.PodUnknown, issues: 1, severity: ErrorLevel},
	}

	for _, u := range uu {
		po := v1.Pod{
			Status: v1.PodStatus{
				Phase: u.phase,
			},
		}
		l := NewPod()
		l.checkStatus(po.Status)
		assert.Equal(t, u.issues, len(l.Issues()))
		if len(l.Issues()) != 0 {
			assert.Equal(t, u.severity, l.Issues()[0].Severity())
		}
	}
}

func TestPoCheckContainerStatus(t *testing.T) {
	uu := []struct {
		state    v1.ContainerState
		ready    bool
		issues   int
		severity Level
	}{
		{ready: true, state: v1.ContainerState{Running: new(v1.ContainerStateRunning)}, issues: 0},
		{ready: false, state: v1.ContainerState{Running: new(v1.ContainerStateRunning)}, issues: 1, severity: ErrorLevel},
		{ready: false, state: v1.ContainerState{Waiting: new(v1.ContainerStateWaiting)}, issues: 1, severity: WarnLevel},
		{ready: false, state: v1.ContainerState{Terminated: new(v1.ContainerStateTerminated)}, issues: 1, severity: WarnLevel},
	}

	for _, u := range uu {
		po := v1.Pod{
			Status: v1.PodStatus{
				ContainerStatuses: []v1.ContainerStatus{
					{
						State: u.state,
						Ready: u.ready,
					},
				},
			},
		}

		l := NewPod()
		l.checkContainerStatus(po.Status.ContainerStatuses, false)
		assert.Equal(t, u.issues, len(l.Issues()))
		if len(l.Issues()) != 0 {
			assert.Equal(t, u.severity, l.Issues()[0].Severity())
		}
	}
}

func TestPoCheckContainers(t *testing.T) {
	uu := []struct {
		request, limit      bool
		liveness, readiness bool
		issues              int
		severity            Level
	}{
		{issues: 3, severity: InfoLevel},
		{readiness: true, issues: 2, severity: InfoLevel},
		{liveness: true, issues: 2, severity: InfoLevel},
		{liveness: true, readiness: true, issues: 1, severity: InfoLevel},
		{limit: true, readiness: false, issues: 2, severity: InfoLevel},
		{limit: true, readiness: true, issues: 1, severity: InfoLevel},
		{limit: true, liveness: true, issues: 1, severity: InfoLevel},
		{limit: true, liveness: true, readiness: true, issues: 0},
		{request: true, issues: 2, severity: InfoLevel},
		{request: true, readiness: true, issues: 1, severity: InfoLevel},
		{request: true, liveness: true, issues: 1, severity: InfoLevel},
		{request: true, liveness: true, readiness: true, issues: 0},
		{request: true, limit: true, issues: 2, severity: InfoLevel},
		{request: true, limit: true, readiness: true, issues: 1, severity: InfoLevel},
		{request: true, limit: true, liveness: true, issues: 1, severity: InfoLevel},
		{request: true, limit: true, liveness: true, readiness: true, issues: 0},
	}

	for _, u := range uu {
		po := v1.Pod{
			Spec: v1.PodSpec{
				Containers: []v1.Container{
					{Name: "c1", Image: "fred:1.2.3"},
				},
			},
		}
		if u.request {
			po.Spec.Containers[0].Resources = v1.ResourceRequirements{
				Requests: v1.ResourceList{
					v1.ResourceCPU: resource.Quantity{},
				},
			}
		}
		if u.limit {
			po.Spec.Containers[0].Resources = v1.ResourceRequirements{
				Limits: v1.ResourceList{
					v1.ResourceCPU: resource.Quantity{},
				},
			}
		}
		if u.liveness {
			po.Spec.Containers[0].LivenessProbe = &v1.Probe{}
		}
		if u.readiness {
			po.Spec.Containers[0].ReadinessProbe = &v1.Probe{}
		}

		l := NewPod()
		l.checkContainers(po.Spec.Containers)
		assert.Equal(t, u.issues, len(l.Issues()))
		if len(l.Issues()) != 0 {
			assert.Equal(t, u.severity, l.Issues()[0].Severity())
		}
	}
}

func TestPoCheckProbes(t *testing.T) {
	uu := []struct {
		liveness, readiness bool
		issues              int
		severity            Level
	}{
		{issues: 2, severity: InfoLevel},
		{liveness: true, readiness: true, issues: 0},
		{liveness: true, issues: 1, severity: InfoLevel},
		{readiness: true, issues: 1, severity: InfoLevel},
	}

	for _, u := range uu {
		po := v1.Pod{
			Spec: v1.PodSpec{
				Containers: []v1.Container{
					{Name: "c1"},
				},
			},
		}
		if u.liveness {
			po.Spec.Containers[0].LivenessProbe = &v1.Probe{}
		}
		if u.readiness {
			po.Spec.Containers[0].ReadinessProbe = &v1.Probe{}
		}

		l := NewPod()
		l.checkProbes(po.Spec.Containers)
		assert.Equal(t, u.issues, len(l.Issues()))
		if len(l.Issues()) != 0 {
			assert.Equal(t, u.severity, l.Issues()[0].Severity())
		}
	}
}

func TestPoCheckServiceAccount(t *testing.T) {
	uu := []struct {
		sa       string
		issues   int
		severity Level
	}{
		{issues: 1, severity: InfoLevel},
		{sa: "fred", issues: 0},
	}

	for _, u := range uu {
		po := v1.Pod{
			Spec: v1.PodSpec{
				ServiceAccountName: u.sa,
			},
		}

		l := NewPod()
		l.checkServiceAccount(po.Spec)
		assert.Equal(t, u.issues, len(l.Issues()))
		if len(l.Issues()) != 0 {
			assert.Equal(t, u.severity, l.Issues()[0].Severity())
		}
	}
}

func TestPoLint(t *testing.T) {
	po := v1.Pod{
		Spec: v1.PodSpec{
			ServiceAccountName: "fred",
			Containers: []v1.Container{
				{
					Name:  "c1",
					Image: "fred:1.2.3",
					Resources: v1.ResourceRequirements{
						Requests: v1.ResourceList{
							v1.ResourceCPU: resource.Quantity{},
						},
					},
					LivenessProbe:  &v1.Probe{},
					ReadinessProbe: &v1.Probe{},
				},
			},
			InitContainers: []v1.Container{
				{
					Name:  "ic1",
					Image: "fred:1.2.3",
					Resources: v1.ResourceRequirements{
						Requests: v1.ResourceList{
							v1.ResourceCPU: resource.Quantity{},
						},
					},
				},
			},
		},
		Status: v1.PodStatus{
			Phase: v1.PodRunning,
			ContainerStatuses: []v1.ContainerStatus{
				{
					Ready: true,
					State: v1.ContainerState{
						Running: &v1.ContainerStateRunning{},
					},
				},
			},
			InitContainerStatuses: []v1.ContainerStatus{
				{
					Ready: true,
					State: v1.ContainerState{
						Terminated: &v1.ContainerStateTerminated{},
					},
				},
			},
		},
	}

	l := NewPod()
	l.Lint(po, nil)
	assert.True(t, l.NoIssues())
}

type podMetrics struct {
	cpu, mem int64
}

func (n podMetrics) CurrentCPU() int64 {
	return n.cpu
}
func (n podMetrics) CurrentMEM() int64 {
	return n.mem
}
func (n podMetrics) Empty() bool {
	return n.cpu == 0 && n.mem == 0
}

func TestPoUtilization(t *testing.T) {
	type (
		metrix struct {
			cpu string
			mem string
		}
		res struct {
			requests *metrix
			limits   *metrix
		}
	)
	uu := []struct {
		mx     podMetrics
		res    res
		issues int
		level  Level
	}{
		// Under the request (Burstable)
		{
			mx: podMetrics{cpu: 50, mem: 15e6},
			res: res{
				requests: &metrix{cpu: "1", mem: "10Mi"},
				limits:   &metrix{cpu: "200m", mem: "20Mi"},
			},
			issues: 0,
		},
		// Under the limit (Burstable)
		{
			mx: podMetrics{cpu: 200, mem: 5e6},
			res: res{
				requests: &metrix{cpu: "100m", mem: "10Mi"},
				limits:   &metrix{cpu: "500m", mem: "20Mi"},
			},
			issues: 0,
		},
		// Over the request CPU
		{
			mx: podMetrics{cpu: 200, mem: 5e6},
			res: res{
				requests: &metrix{cpu: "100m", mem: "10Mi"},
			},
			issues: 1,
		},
		// Over the request MEM
		{
			mx: podMetrics{cpu: 50, mem: 15e6},
			res: res{
				requests: &metrix{cpu: "100m", mem: "10Mi"},
			},
			issues: 1,
		},
		// Over the limit CPU (Guaranteed)
		{
			mx: podMetrics{cpu: 200, mem: 5e6},
			res: res{
				limits: &metrix{cpu: "100m", mem: "20Mi"},
			},
			issues: 1,
		},
		// Over the limit MEM (Guaranteed)
		{
			mx: podMetrics{cpu: 50, mem: 40e6},
			res: res{
				limits: &metrix{cpu: "100m", mem: "20Mi"},
			},
			issues: 1,
		},
	}

	for _, u := range uu {
		po := v1.Pod{
			ObjectMeta: metav1.ObjectMeta{Name: "fred"},
		}

		co := v1.Container{
			Name:  "c1",
			Image: "fred:1.2.3",
		}

		var resReq v1.ResourceRequirements
		if u.res.requests != nil {
			cpu := resource.MustParse(u.res.requests.cpu)
			mem := resource.MustParse(u.res.requests.mem)
			resReq.Requests = v1.ResourceList{
				v1.ResourceCPU:    cpu,
				v1.ResourceMemory: mem,
			}
		}
		if u.res.limits != nil {
			cpu := resource.MustParse(u.res.limits.cpu)
			mem := resource.MustParse(u.res.limits.mem)
			resReq.Limits = v1.ResourceList{
				v1.ResourceCPU:    cpu,
				v1.ResourceMemory: mem,
			}
		}
		co.Resources = resReq
		po.Spec = v1.PodSpec{Containers: []v1.Container{co}}

		l := NewPod()
		l.checkUtilization(po, map[string]PodMetric{"c1": u.mx})
		assert.Equal(t, u.issues, len(l.Issues()))
	}
}
