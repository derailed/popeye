package sanitize

import (
	"context"
	"testing"

	"github.com/derailed/popeye/internal/issues"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	pv1beta1 "k8s.io/api/policy/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

func TestPodSanitize(t *testing.T) {
	uu := map[string]struct {
		lister PodMXLister
		issues int
	}{
		"cool": {
			makePodLister("p1", podOpts{
				pods: map[string]*v1.Pod{
					"default/p1": makeFullPod("p1", podOpts{
						serviceAcct: "fred",
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
						phase: v1.PodRunning,
					}),
				},
			}),
			0,
		},
		"unhappy": {
			makePodLister("p1", podOpts{
				pods: map[string]*v1.Pod{
					"default/p1": makeFullPod("p1", podOpts{
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
			1,
		},
		"noSA": {
			makePodLister("p1", podOpts{
				pods: map[string]*v1.Pod{
					"default/p1": makeFullPod("p1", podOpts{
						coOpts: coOpts{
							rcpu: "100m",
							rmem: "20Mi",
							lcpu: "100m",
							lmem: "200Mi",
						},
						phase: v1.PodRunning,
						csOpts: csOpts{
							ready:    true,
							restarts: 0,
							state:    running,
						},
					}),
				},
			}),
			1,
		},
	}

	for k, u := range uu {
		t.Run(k, func(t *testing.T) {
			p := NewPod(issues.NewCollector(), u.lister)
			p.Sanitize(context.Background())

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
	}

	pod struct {
		opts podOpts
	}
)

func makePodLister(n string, opts podOpts) *pod {
	return &pod{
		opts: opts,
	}
}

func (p *pod) ListPods() map[string]*v1.Pod {
	return p.opts.pods
}

func (p *pod) GetPod(map[string]string) *v1.Pod {
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

func (p *pod) ListPodsMetrics() map[string]*v1beta1.PodMetrics {
	return map[string]*v1beta1.PodMetrics{
		"default/p1": makeMxPod("p1", "10m", "10Mi"),
	}
}

func (p *pod) ForLabels(l map[string]string) *pv1beta1.PodDisruptionBudget {
	return &pv1beta1.PodDisruptionBudget{}
}

func (p *pod) ListPodDisruptionBudgets() map[string]*pv1beta1.PodDisruptionBudget {
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

func makePhasePod(n string, p v1.PodPhase) *v1.Pod {
	po := makePod(n)
	po.Status.Phase = p

	return po
}

func makeMxPod(name, cpu, mem string) *v1beta1.PodMetrics {
	return &v1beta1.PodMetrics{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: "default",
		},
		Containers: []v1beta1.ContainerMetrics{
			{Name: "i1", Usage: makeRes(cpu, mem)},
			{Name: "c1", Usage: makeRes(cpu, mem)},
		},
	}
}

func makeFullPod(n string, opts podOpts) *v1.Pod {
	po := makePod(n)
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
