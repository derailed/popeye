package linter

import (
	"context"
	"testing"

	m "github.com/petergtz/pegomock"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	mv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

func TestPoLinter(t *testing.T) {
	mkl := NewMockLoader()
	m.When(mkl.ListPods()).ThenReturn(map[string]v1.Pod{
		"default/p1": makePod("p1"),
		"default/p2": makePod("p2"),
	}, nil)
	m.When(mkl.ClusterHasMetrics()).ThenReturn(true, nil)
	m.When(mkl.FetchPodsMetrics("")).ThenReturn([]mv1beta1.PodMetrics{
		makeMxPod("p1", "50m", "1Mi"),
		makeMxPod("p2", "50m", "1Mi"),
	}, nil)

	l := NewPod(mkl, nil)
	l.Lint(context.Background())

	assert.Equal(t, 2, len(l.Issues()))
	assert.Equal(t, 0, len(l.Issues()["p1"]))
	assert.Equal(t, 0, len(l.Issues()["p2"]))

	mkl.VerifyWasCalledOnce().ListPods()
	mkl.VerifyWasCalledOnce().ClusterHasMetrics()
	mkl.VerifyWasCalledOnce().FetchPodsMetrics("")
}

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
		po := makePod("p1")
		po.Status = v1.PodStatus{
			Phase: u.phase,
		}

		l := NewPod(nil, nil)
		l.checkStatus(po)

		fqn := podFQN(po)
		assert.Equal(t, u.issues, len(l.Issues()[fqn]))
		if len(l.Issues()[fqn]) != 0 {
			assert.Equal(t, u.severity, l.MaxSeverity(fqn))
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
		{ready: false, state: v1.ContainerState{Waiting: new(v1.ContainerStateWaiting)}, issues: 1, severity: ErrorLevel},
		{ready: false, state: v1.ContainerState{Terminated: new(v1.ContainerStateTerminated)}, issues: 0, severity: WarnLevel},
	}

	for _, u := range uu {
		po := makePod("p1")
		po.Status = v1.PodStatus{
			ContainerStatuses: []v1.ContainerStatus{
				{
					State: u.state,
					Ready: u.ready,
				},
			},
		}

		mkl := NewMockLoader()
		m.When(mkl.RestartsLimit()).ThenReturn(1)

		l := NewPod(mkl, nil)
		l.checkContainerStatus(po)

		fqn := podFQN(po)
		assert.Equal(t, u.issues, len(l.Issues()[fqn]))
		if len(l.Issues()[fqn]) != 0 {
			assert.Equal(t, u.severity, l.Issues()[fqn][0].Severity())
		}
		mkl.VerifyWasCalledOnce().RestartsLimit()
	}
}

func TestPoCheckContainers(t *testing.T) {
	uu := []struct {
		request, limit      bool
		liveness, readiness bool
		issues              int
		severity            Level
	}{
		{issues: 2, severity: WarnLevel},
		{issues: 2, readiness: true, severity: WarnLevel},
		{issues: 2, liveness: true, severity: WarnLevel},
		{issues: 1, liveness: true, readiness: true, severity: WarnLevel},
		{issues: 1, limit: true, severity: WarnLevel},
		{issues: 1, limit: true, readiness: true, severity: WarnLevel},
		{issues: 1, limit: true, liveness: true, severity: WarnLevel},
		{issues: 0, limit: true, liveness: true, readiness: true},
		{issues: 2, request: true, severity: WarnLevel},
		{issues: 2, request: true, readiness: true, severity: WarnLevel},
		{issues: 2, request: true, liveness: true, severity: WarnLevel},
		{issues: 1, request: true, liveness: true, readiness: true, severity: WarnLevel},
		{issues: 1, request: true, limit: true, severity: WarnLevel},
		{issues: 1, request: true, limit: true, readiness: true, severity: WarnLevel},
		{issues: 1, request: true, limit: true, liveness: true, severity: WarnLevel},
		{issues: 0, request: true, limit: true, liveness: true, readiness: true},
	}

	for _, u := range uu {
		po := makePod("p1")
		po.Spec = v1.PodSpec{
			Containers: []v1.Container{
				{Name: "c1", Image: "fred:1.2.3"},
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

		l := NewPod(nil, nil)
		l.checkContainers(po)

		fqn := podFQN(po)
		assert.Equal(t, u.issues, len(l.Issues()[fqn]))
		if len(l.Issues()[fqn]) != 0 {
			assert.Equal(t, u.severity, l.MaxSeverity(fqn))
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
		po := makePod("p1")
		if u.sa != "" {
			po.Spec.ServiceAccountName = u.sa
		}
		fqn := podFQN(po)

		l := NewPod(nil, nil)
		l.checkServiceAccount(po)

		assert.Equal(t, u.issues, len(l.Issues()[fqn]))
		if len(l.Issues()[fqn]) != 0 {
			assert.Equal(t, u.severity, l.MaxSeverity(fqn))
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

	mkl := NewMockLoader()
	l := NewPod(mkl, nil)
	l.lint(po, nil)

	assert.True(t, l.NoIssues("p1"))
}

func TestPoUtilization(t *testing.T) {
	uu := []struct {
		mx     Metrics
		res    v1.ResourceRequirements
		issues int
		level  Level
	}{
		// Under the request (Burstable)
		{
			mx: Metrics{CurrentCPU: 50, CurrentMEM: 15e6},
			res: v1.ResourceRequirements{
				Requests: makeRes("1", "10Mi"),
				Limits:   makeRes("200m", "20Mi"),
			},
			issues: 0,
		},
		// Under the limit (Burstable)
		{
			mx: Metrics{CurrentCPU: 200, CurrentMEM: 5e6},
			res: v1.ResourceRequirements{
				Requests: makeRes("100m", "10Mi"),
				Limits:   makeRes("500m", "20Mi"),
			},
			issues: 0,
		},
		// Over the request CPU
		{
			mx: Metrics{CurrentCPU: 200, CurrentMEM: 5e6},
			res: v1.ResourceRequirements{
				Requests: makeRes("100m", "10Mi"),
			},
			issues: 1,
		},
		// Over the request MEM
		{
			mx: Metrics{CurrentCPU: 50, CurrentMEM: 15e6},
			res: v1.ResourceRequirements{
				Requests: makeRes("100m", "10Mi"),
			},
			issues: 1,
		},
		// Over the limit CPU (Guaranteed)
		{
			mx: Metrics{CurrentCPU: 200, CurrentMEM: 5e6},
			res: v1.ResourceRequirements{
				Limits: makeRes("100m", "20Mi"),
			},
			issues: 1,
		},
		// Over the limit MEM (Guaranteed)
		{
			mx: Metrics{CurrentCPU: 50, CurrentMEM: 40e6},
			res: v1.ResourceRequirements{
				Limits: makeRes("100m", "20Mi"),
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
		if u.res.Requests != nil {
			resReq.Requests = u.res.Requests
		}
		if u.res.Limits != nil {
			resReq.Limits = u.res.Limits
		}
		co.Resources = resReq
		po.Spec = v1.PodSpec{Containers: []v1.Container{co}}

		mkl := NewMockLoader()
		m.When(mkl.PodCPULimit()).ThenReturn(float64(80))
		m.When(mkl.PodMEMLimit()).ThenReturn(float64(80))

		l := NewPod(mkl, nil)
		l.checkUtilization(po, ContainerMetrics{"c1": u.mx})

		assert.Equal(t, u.issues, len(l.Issues()))
		mkl.VerifyWasCalledOnce().PodCPULimit()
		mkl.VerifyWasCalledOnce().PodMEMLimit()
	}
}

// ----------------------------------------------------------------------------
// Helpers...

func makePod(n string) v1.Pod {
	po := v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      n,
			Namespace: "default",
		},
	}

	return po
}
