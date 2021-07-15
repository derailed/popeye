package sanitize

import (
	"testing"

	"github.com/derailed/popeye/internal"
	"github.com/derailed/popeye/internal/issues"
	"github.com/derailed/popeye/pkg/config"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	polv1beta1 "k8s.io/api/policy/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

func TestPodCheckSecure(t *testing.T) {
	uu := map[string]struct {
		pod    v1.Pod
		issues int
	}{
		"cool_1": {
			pod:    makeSecPod(SecNonRootSet, SecNonRootSet, SecNonRootSet, SecNonRootSet),
			issues: 0,
		},
		"cool_2": {
			pod:    makeSecPod(SecNonRootSet, SecNonRootUnset, SecNonRootUnset, SecNonRootUnset),
			issues: 0,
		},
		"cool_3": {
			pod:    makeSecPod(SecNonRootUnset, SecNonRootSet, SecNonRootSet, SecNonRootSet),
			issues: 0,
		},
		"cool_4": {
			pod:    makeSecPod(SecNonRootUndefined, SecNonRootSet, SecNonRootSet, SecNonRootSet),
			issues: 0,
		},
		"cool_5": {
			pod:    makeSecPod(SecNonRootSet, SecNonRootUndefined, SecNonRootUndefined, SecNonRootUndefined),
			issues: 0,
		},
		"hacked_1": {
			pod:    makeSecPod(SecNonRootUndefined, SecNonRootUndefined, SecNonRootUndefined, SecNonRootUndefined),
			issues: 4,
		},
		"hacked_2": {
			pod:    makeSecPod(SecNonRootUndefined, SecNonRootUnset, SecNonRootUndefined, SecNonRootUndefined),
			issues: 4,
		},
		"hacked_3": {
			pod:    makeSecPod(SecNonRootUndefined, SecNonRootSet, SecNonRootUndefined, SecNonRootUndefined),
			issues: 3,
		},
		"hacked_4": {
			pod:    makeSecPod(SecNonRootUndefined, SecNonRootUnset, SecNonRootSet, SecNonRootUndefined),
			issues: 3,
		},
		"toast": {
			pod:    makeSecPod(SecNonRootUndefined, SecNonRootUndefined, SecNonRootUndefined, SecNonRootUndefined),
			issues: 4,
		},
	}

	ctx := makeContext("v1/pods", "po")
	ctx = internal.WithFQN(ctx, "default/p1")
	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			p := NewPod(issues.NewCollector(loadCodes(t), makeConfig(t)), nil)

			p.checkSecure(ctx, "default/p1", u.pod.Spec)
			assert.Equal(t, u.issues, len(p.Outcome()["default/p1"]))
		})
	}
}

func TestPodSanitize(t *testing.T) {
	uu := map[string]struct {
		lister PodMXLister
		issues int
	}{
		"cool": {
			makePodLister(podOpts{
				pods: map[string]*v1.Pod{
					"default/p1": makeFullPod(podOpts{
						serviceAcct: "fred",
						certs:       false,
						coOpts: coOpts{
							rcpu: "100m",
							rmem: "20Mi",
							lcpu: "100m",
							lmem: "200Mi",
						},
						csOpts: csOpts{
							ready:    true,
							restarts: 0,
							state:    running,
						},
						phase:      v1.PodRunning,
						controlled: true,
					}),
				},
			}),
			0,
		},
		"unhappy": {
			makePodLister(podOpts{
				pods: map[string]*v1.Pod{
					"default/p1": makeFullPod(podOpts{
						coOpts: coOpts{
							rcpu: "100m",
							rmem: "20Mi",
							lcpu: "100m",
							lmem: "200Mi",
						},
						csOpts: csOpts{
							ready:    true,
							restarts: 0,
							state:    running,
						},
						serviceAcct: "fred",
						phase:       v1.PodPending,
					}),
				},
			}),
			2,
		},
		"defaultSA": {
			makePodLister(podOpts{
				pods: map[string]*v1.Pod{
					"default/p1": makeFullPod(podOpts{
						coOpts: coOpts{
							rcpu: "100m",
							rmem: "20Mi",
							lcpu: "100m",
							lmem: "200Mi",
						},
						serviceAcct: "default",
						certs:       true,
						phase:       v1.PodRunning,
						csOpts: csOpts{
							ready:    true,
							restarts: 0,
							state:    running,
						},
						controlled: true,
					}),
				},
			}),
			2,
		},
	}

	ctx := makeContext("v1/pods", "po")
	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			p := NewPod(issues.NewCollector(loadCodes(t), makeConfig(t)), u.lister)

			assert.Nil(t, p.Sanitize(ctx))
			assert.Equal(t, u.issues, len(p.Outcome()["default/p1"]))
		})
	}
}

// ----------------------------------------------------------------------------
// Helpers...

type (
	podOpts struct {
		coOpts
		csOpts
		phase       v1.PodPhase
		pods        map[string]*v1.Pod
		serviceAcct string
		certs       bool
		controlled  bool
	}

	pod struct {
		opts podOpts
	}
)

func makePodLister(opts podOpts) *pod {
	return &pod{
		opts: opts,
	}
}

func (p *pod) ListPods() map[string]*v1.Pod {
	return p.opts.pods
}

func (p *pod) ListServiceAccounts() map[string]*v1.ServiceAccount {
	return make(map[string]*v1.ServiceAccount)
}

func (p *pod) GetPod(string, map[string]string) *v1.Pod {
	return nil
}

func (*pod) RestartsLimit() int {
	return 10
}

func (*pod) PodCPULimit() float64 {
	return 90
}

func (*pod) PodMEMLimit() float64 {
	return 90
}

func (*pod) CPUResourceLimits() config.Allocations {
	return config.Allocations{
		UnderPerc: 100,
		OverPerc:  50,
	}
}

func (*pod) MEMResourceLimits() config.Allocations {
	return config.Allocations{
		UnderPerc: 100,
		OverPerc:  50,
	}
}

func (p *pod) ListPodsMetrics() map[string]*v1beta1.PodMetrics {
	return map[string]*v1beta1.PodMetrics{
		"default/p1": makeMxPod("10m", "10Mi"),
	}
}

func (p *pod) ForLabels(l map[string]string) *polv1beta1.PodDisruptionBudget {
	return &polv1beta1.PodDisruptionBudget{}
}

func (p *pod) ListPodDisruptionBudgets() map[string]*polv1beta1.PodDisruptionBudget {
	return nil
}

func makePod(n string) *v1.Pod {
	po := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      n,
			Namespace: "default",
		},
	}

	po.Status.Phase = v1.PodRunning

	return po
}

func makeMxPod(cpu, mem string) *v1beta1.PodMetrics {
	return &v1beta1.PodMetrics{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "p1",
			Namespace: "default",
		},
		Containers: []v1beta1.ContainerMetrics{
			{Name: "i1", Usage: makeRes(cpu, mem)},
			{Name: "c1", Usage: makeRes(cpu, mem)},
		},
	}
}

func makeFullPod(opts podOpts) *v1.Pod {
	po := makePod("p1")
	po.Spec = v1.PodSpec{
		InitContainers: []v1.Container{
			makeContainer("i1", coOpts{
				image: "fred:0.0.1",
				rcpu:  opts.rcpu,
				rmem:  opts.rmem,
				lcpu:  opts.lcpu,
				lmem:  opts.lmem,
			}),
		},
		Containers: []v1.Container{
			makeContainer("c1", coOpts{
				image: "fred:0.0.1",
				rcpu:  opts.rcpu,
				rmem:  opts.rmem,
				lcpu:  opts.lcpu,
				lmem:  opts.lmem,
				lprob: true,
				rprob: true,
			}),
		},
	}
	if opts.serviceAcct != "" {
		po.Spec.ServiceAccountName = opts.serviceAcct
	}
	po.Spec.AutomountServiceAccountToken = &opts.certs

	if opts.controlled {
		truthful := true
		po.OwnerReferences = append(po.OwnerReferences, metav1.OwnerReference{
			Kind:       "ReplicaSet",
			Name:       "mock-replica-set",
			Controller: &truthful,
		})
	}

	po.Status = v1.PodStatus{
		Phase: opts.phase,
		InitContainerStatuses: []v1.ContainerStatus{
			makeCS("i1", opts.csOpts),
		},
		ContainerStatuses: []v1.ContainerStatus{
			makeCS("c1", opts.csOpts),
		},
	}

	return po
}

const (
	running int = iota
	waiting
	terminated
)

type csOpts struct {
	ready    bool
	restarts int32
	state    int
}

func makeCS(n string, opts csOpts) v1.ContainerStatus {
	cs := v1.ContainerStatus{
		Name:         n,
		Ready:        opts.ready,
		RestartCount: opts.restarts,
	}

	switch opts.state {
	case waiting:
		cs.State = v1.ContainerState{
			Waiting: &v1.ContainerStateWaiting{},
		}
	case terminated:
		cs.State = v1.ContainerState{
			Terminated: &v1.ContainerStateTerminated{},
		}
	default:
		cs.State = v1.ContainerState{
			Running: &v1.ContainerStateRunning{},
		}
	}

	return cs
}

func makeSecCO(name string, level NonRootUser) v1.Container {
	t, f := true, false
	secCtx := v1.SecurityContext{}
	// nolint:exhaustive
	switch level {
	case SecNonRootUnset:
		secCtx.RunAsNonRoot = &f
	case SecNonRootSet:
		secCtx.RunAsNonRoot = &t
	}

	return v1.Container{Name: name, SecurityContext: &secCtx}
}

func makeSecPod(pod, init, co1, co2 NonRootUser) v1.Pod {
	t, f := true, false

	secCtx := v1.PodSecurityContext{}
	// nolint:exhaustive
	switch pod {
	case SecNonRootUnset:
		secCtx.RunAsNonRoot = &f
	case SecNonRootSet:
		secCtx.RunAsNonRoot = &t
	}

	var auto bool
	return v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: "p1",
		},
		Spec: v1.PodSpec{
			AutomountServiceAccountToken: &auto,
			InitContainers:               []v1.Container{makeSecCO("i1", init)},
			Containers: []v1.Container{
				makeSecCO("co1", co1),
				makeSecCO("co2", co2),
			},
			SecurityContext: &secCtx,
		},
	}
}
