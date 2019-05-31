package sanitize

import (
	"context"
	"testing"

	"github.com/derailed/popeye/internal/cache"
	"github.com/derailed/popeye/internal/issues"
	"github.com/derailed/popeye/internal/k8s"
	"github.com/derailed/popeye/pkg/config"
	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	mv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

func TestDPSanitize(t *testing.T) {
	uu := map[string]struct {
		dpl    DeploymentLister
		issues int
	}{
		"good": {
			makeDPLister("d1", dpOpts{
				reps:      1,
				availReps: 1,
				coOpts: coOpts{
					image: "fred:0.0.1",
					rcpu:  "10m",
					rmem:  "10Mi",
					lcpu:  "10m",
					lmem:  "10Mi",
				},
				ccpu: "10m",
				cmem: "10Mi",
			}),
			0,
		},
		"noReps": {
			makeDPLister("d1", dpOpts{
				reps:      0,
				availReps: 1,
				coOpts: coOpts{
					image: "fred:0.0.1",
					rcpu:  "10m",
					rmem:  "10Mi",
					lcpu:  "10m",
					lmem:  "10Mi",
				},
				ccpu: "10m",
				cmem: "10Mi",
			}),
			1,
		},
		"noAvailReps": {
			makeDPLister("d1", dpOpts{
				reps:       1,
				availReps:  0,
				collisions: 0,
				coOpts: coOpts{
					image: "fred:0.0.1",
					rcpu:  "10m",
					rmem:  "10Mi",
					lcpu:  "10m",
					lmem:  "10Mi",
				},
				ccpu: "10m",
				cmem: "10Mi",
			}),
			1,
		},
		"collisions": {
			makeDPLister("d1", dpOpts{
				reps:       1,
				availReps:  1,
				collisions: 1,
				coOpts: coOpts{
					image: "fred:0.0.1",
					rcpu:  "10m",
					rmem:  "10Mi",
					lcpu:  "10m",
					lmem:  "10Mi",
				},
				ccpu: "10m",
				cmem: "10Mi",
			}),
			1,
		},
		"cpuOver": {
			makeDPLister("d1", dpOpts{
				reps:       1,
				availReps:  1,
				collisions: 0,
				coOpts: coOpts{
					image: "fred:0.0.1",
					rcpu:  "20m",
					rmem:  "10Mi",
					lcpu:  "20m",
					lmem:  "10Mi",
				},
				ccpu: "10m",
				cmem: "10Mi",
			}),
			1,
		},
		"cpuUnder": {
			makeDPLister("d1", dpOpts{
				reps:       1,
				availReps:  1,
				collisions: 0,
				coOpts: coOpts{
					image: "fred:0.0.1",
					rcpu:  "1m",
					rmem:  "10Mi",
					lcpu:  "10m",
					lmem:  "10Mi",
				},
				ccpu: "10m",
				cmem: "10Mi",
			}),
			1,
		},
		"memOver": {
			makeDPLister("d1", dpOpts{
				reps:       1,
				availReps:  1,
				collisions: 0,
				coOpts: coOpts{
					image: "fred:0.0.1",
					rcpu:  "10m",
					rmem:  "20Mi",
					lcpu:  "10m",
					lmem:  "20Mi",
				},
				ccpu: "10m",
				cmem: "10Mi",
			}),
			1,
		},
		"memUnder": {
			makeDPLister("d1", dpOpts{
				reps:       1,
				availReps:  1,
				collisions: 0,
				coOpts: coOpts{
					image: "fred:0.0.1",
					rcpu:  "10m",
					rmem:  "2Mi",
					lcpu:  "10m",
					lmem:  "20Mi",
				},
				ccpu: "10m",
				cmem: "10Mi",
			}),
			1,
		},
	}

	for k, u := range uu {
		t.Run(k, func(t *testing.T) {
			dp := NewDeployment(issues.NewCollector(), u.dpl)
			dp.Sanitize(context.Background())

			assert.Equal(t, u.issues, len(dp.Outcome()["default/d1"]))
		})
	}
}

// ----------------------------------------------------------------------------
// Helpers...

type (
	dpOpts struct {
		coOpts
		reps       int32
		availReps  int32
		collisions int32
		ccpu, cmem string
	}

	dp struct {
		name string
		opts dpOpts
	}
)

func makeDPLister(n string, opts dpOpts) *dp {
	return &dp{
		name: n,
		opts: opts,
	}
}

func (d *dp) CPUResourceLimits() config.Allocations {
	return config.Allocations{
		Over:  100,
		Under: 50,
	}
}

func (d *dp) MEMResourceLimits() config.Allocations {
	return config.Allocations{
		Over:  100,
		Under: 50,
	}
}

func (d *dp) ListPodsBySelector(sel *metav1.LabelSelector) map[string]*v1.Pod {
	return map[string]*v1.Pod{
		"default/p1": makeFullPod("p1", podOpts{
			coOpts: d.opts.coOpts,
		}),
	}
}

func (d *dp) RestartsLimit() int {
	return 10
}

func (d *dp) PodCPULimit() float64 {
	return 100
}

func (d *dp) PodMEMLimit() float64 {
	return 100
}

func (d *dp) ListPodsMetrics() map[string]*mv1beta1.PodMetrics {
	return map[string]*mv1beta1.PodMetrics{
		cache.FQN("default", "p1"): makeMxPod("p1", d.opts.ccpu, d.opts.cmem),
	}
}

func (d *dp) ListDeployments() map[string]*appsv1.Deployment {
	return map[string]*appsv1.Deployment{
		cache.FQN("default", d.name): makeDP(d.name, d.opts),
	}
}

func makeContainerMx(n, cpu, mem string) k8s.ContainerMetrics {
	return k8s.ContainerMetrics{
		n: k8s.Metrics{
			CurrentCPU: toQty(cpu),
			CurrentMEM: toQty(mem),
		},
	}
}

func makeDP(n string, o dpOpts) *appsv1.Deployment {
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      n,
			Namespace: "default",
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &o.reps,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"fred": "blee",
				},
			},
			Template: v1.PodTemplateSpec{
				Spec: v1.PodSpec{
					InitContainers: []v1.Container{
						makeContainer("i1", o.coOpts),
					},
					Containers: []v1.Container{
						makeContainer("c1", o.coOpts),
					},
				},
			},
		},
		Status: appsv1.DeploymentStatus{
			AvailableReplicas: o.availReps,
			CollisionCount:    &o.collisions,
		},
	}
}

// func makePodRes(n, cpu, mem string) *v1.Pod {
// 	po := makePod(n)
// 	po.Spec.Containers = []v1.Container{
// 		{
// 			Name:  "c1",
// 			Image: "fred:1.2.3",
// 			Resources: v1.ResourceRequirements{
// 				Requests: makeRes("cpu", cpu),
// 				Limits:   makeRes("mem", mem),
// 			},
// 		},
// 	}
// 	po.Spec.InitContainers = []v1.Container{
// 		{
// 			Name:  "ic1",
// 			Image: "fred:1.2.3",
// 			Resources: v1.ResourceRequirements{
// 				Requests: makeRes("cpu", cpu),
// 				Limits:   makeRes("mem", mem),
// 			},
// 		},
// 	}

// 	return po
// }
