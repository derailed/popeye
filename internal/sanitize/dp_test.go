package sanitize

import (
	"context"
	"testing"

	"github.com/derailed/popeye/internal"
	"github.com/derailed/popeye/internal/cache"
	"github.com/derailed/popeye/internal/client"
	"github.com/derailed/popeye/internal/issues"
	"github.com/derailed/popeye/pkg/config"
	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	mv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

func TestRevFromLink(t *testing.T) {
	uu := map[string]struct {
		l, e string
	}{
		"single.namespaced": {
			"/api/v1/namespaces/fred/pods/blee",
			"v1",
		},
		"single.notnamespaced": {
			"/api/v1/pv/blee",
			"v1",
		},
		"multi.namespaced": {
			"/api/extensions/v1beta1/namespaces/fred/pods/blee",
			"extensions/v1beta1",
		},
		"multi.notnamespaced": {
			"/api/rbac.authorization.k8s.io/v1beta1/blee/duh",
			"rbac.authorization.k8s.io/v1beta1",
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, revFromLink(u.l))
		})
	}
}

func TestDPSanitize(t *testing.T) {
	uu := map[string]struct {
		lister DPLister
		issues issues.Issues
	}{
		"good": {
			lister: makeDPLister(dpOpts{
				rev:       "apps/v1",
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
			issues: issues.Issues{},
		},
		"deprecated": {
			lister: makeDPLister(dpOpts{
				rev:       "extensions/v1",
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
			issues: issues.Issues{
				issues.New(
					client.NewGVR("apps/v1/deployments"),
					issues.Root,
					config.WarnLevel,
					`[POP-403] Deprecated Deployment API group "extensions/v1". Use "apps/v1" instead`,
				),
			},
		},
		"zeroReps": {
			lister: makeDPLister(dpOpts{
				rev:       "apps/v1",
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
			issues: issues.Issues{
				issues.New(client.NewGVR("apps/v1/deployments"), issues.Root, config.WarnLevel, "[POP-500] Zero scale detected"),
			},
		},
		"noAvailReps": {
			lister: makeDPLister(dpOpts{
				rev:        "apps/v1",
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
			issues: issues.Issues{
				issues.New(client.NewGVR("apps/v1/deployments"), issues.Root, config.ErrorLevel, "[POP-501] Unhealthy 1 desired but have 0 available"),
			},
		},
	}

	ctx := makeContext("apps/v1/deployments", "deployment")
	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			dp := NewDeployment(issues.NewCollector(loadCodes(t), makeConfig(t)), u.lister)

			assert.Nil(t, dp.Sanitize(ctx))
			assert.Equal(t, u.issues, dp.Outcome()["default/d1"])
		})
	}
}

func TestDPSanitizeUtilization(t *testing.T) {
	uu := map[string]struct {
		lister DPLister
		issues issues.Issues
	}{
		"bestEffort": {
			lister: makeDPLister(dpOpts{
				rev:        "apps/v1",
				reps:       2,
				availReps:  2,
				collisions: 0,
				coOpts: coOpts{
					image: "fred:0.0.1",
				},
				ccpu: "10m",
				cmem: "10Mi",
			}),
			issues: issues.Issues{
				issues.New(client.NewGVR("containers"), "i1", config.WarnLevel, "[POP-106] No resources requests/limits defined"),
				issues.New(client.NewGVR("containers"), "c1", config.WarnLevel, "[POP-106] No resources requests/limits defined"),
			},
		},
		"cpuUnderBurstable": {
			lister: makeDPLister(dpOpts{
				rev:        "apps/v1",
				reps:       2,
				availReps:  2,
				collisions: 0,
				coOpts: coOpts{
					image: "fred:0.0.1",
					rcpu:  "5m",
					rmem:  "10Mi",
					lcpu:  "10m",
					lmem:  "10Mi",
				},
				ccpu: "10m",
				cmem: "10Mi",
			}),
			issues: issues.Issues{
				issues.New(client.NewGVR("containers"), issues.Root, config.WarnLevel, "[POP-503] At current load, CPU under allocated. Current:20m vs Requested:10m (200.00%)"),
			},
		},
		"cpuUnderGuaranteed": {
			lister: makeDPLister(dpOpts{
				rev:        "apps/v1",
				reps:       2,
				availReps:  2,
				collisions: 0,
				coOpts: coOpts{
					image: "fred:0.0.1",
					rcpu:  "5m",
					rmem:  "10Mi",
					lcpu:  "5m",
					lmem:  "10Mi",
				},
				ccpu: "10m",
				cmem: "10Mi",
			}),
			issues: issues.Issues{
				issues.New(client.NewGVR("containers"), issues.Root, config.WarnLevel, "[POP-503] At current load, CPU under allocated. Current:20m vs Requested:10m (200.00%)"),
			},
		},
		"cpuOverBustable": {
			lister: makeDPLister(dpOpts{
				rev:        "apps/v1",
				reps:       2,
				availReps:  2,
				collisions: 0,
				coOpts: coOpts{
					image: "fred:0.0.1",
					rcpu:  "30m",
					rmem:  "10Mi",
					lcpu:  "50m",
					lmem:  "10Mi",
				},
				ccpu: "10m",
				cmem: "10Mi",
			}),
			issues: issues.Issues{
				issues.New(client.NewGVR("containers"), issues.Root, config.WarnLevel, "[POP-504] At current load, CPU over allocated. Current:20m vs Requested:60m (300.00%)"),
			},
		},
		"cpuOverGuaranteed": {
			lister: makeDPLister(dpOpts{
				rev:        "apps/v1",
				reps:       2,
				availReps:  2,
				collisions: 0,
				coOpts: coOpts{
					image: "fred:0.0.1",
					rcpu:  "30m",
					rmem:  "10Mi",
					lcpu:  "30m",
					lmem:  "10Mi",
				},
				ccpu: "10m",
				cmem: "10Mi",
			}),
			issues: issues.Issues{
				issues.New(client.NewGVR("containers"), issues.Root, config.WarnLevel, "[POP-504] At current load, CPU over allocated. Current:20m vs Requested:60m (300.00%)"),
			},
		},
		"memUnderBurstable": {
			lister: makeDPLister(dpOpts{
				rev:        "apps/v1",
				reps:       2,
				availReps:  2,
				collisions: 0,
				coOpts: coOpts{
					image: "fred:0.0.1",
					rcpu:  "10m",
					rmem:  "5Mi",
					lcpu:  "20m",
					lmem:  "20Mi",
				},
				ccpu: "10m",
				cmem: "10Mi",
			}),
			issues: issues.Issues{
				issues.New(client.NewGVR("containers"), issues.Root, config.WarnLevel, "[POP-505] At current load, Memory under allocated. Current:20Mi vs Requested:10Mi (200.00%)"),
			},
		},
		"memUnderGuaranteed": {
			lister: makeDPLister(dpOpts{
				rev:        "apps/v1",
				reps:       2,
				availReps:  2,
				collisions: 0,
				coOpts: coOpts{
					image: "fred:0.0.1",
					rcpu:  "10m",
					rmem:  "5Mi",
					lcpu:  "10m",
					lmem:  "5Mi",
				},
				ccpu: "10m",
				cmem: "10Mi",
			}),
			issues: issues.Issues{
				issues.New(client.NewGVR("containers"), issues.Root, config.WarnLevel, "[POP-505] At current load, Memory under allocated. Current:20Mi vs Requested:10Mi (200.00%)"),
			},
		},
		"memOverBurstable": {
			lister: makeDPLister(dpOpts{
				rev:        "apps/v1",
				reps:       2,
				availReps:  2,
				collisions: 0,
				coOpts: coOpts{
					image: "fred:0.0.1",
					rcpu:  "10m",
					rmem:  "30Mi",
					lcpu:  "20m",
					lmem:  "60Mi",
				},
				ccpu: "10m",
				cmem: "10Mi",
			}),
			issues: issues.Issues{
				issues.New(client.NewGVR("containers"), issues.Root, config.WarnLevel, "[POP-506] At current load, Memory over allocated. Current:20Mi vs Requested:60Mi (300.00%)"),
			},
		},
		"memOverGuaranteed": {
			lister: makeDPLister(dpOpts{
				rev:        "apps/v1",
				reps:       2,
				availReps:  2,
				collisions: 0,
				coOpts: coOpts{
					image: "fred:0.0.1",
					rcpu:  "10m",
					rmem:  "30Mi",
					lcpu:  "10m",
					lmem:  "30Mi",
				},
				ccpu: "10m",
				cmem: "10Mi",
			}),
			issues: issues.Issues{
				issues.New(client.NewGVR("containers"), issues.Root, config.WarnLevel, "[POP-506] At current load, Memory over allocated. Current:20Mi vs Requested:60Mi (300.00%)"),
			},
		},
	}

	ctx := makeContext("containers", "deploy")
	ctx = context.WithValue(ctx, internal.KeyOverAllocs, true)
	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			dp := NewDeployment(issues.NewCollector(loadCodes(t), makeConfig(t)), u.lister)

			assert.Nil(t, dp.Sanitize(ctx))
			assert.Equal(t, u.issues, dp.Outcome()["default/d1"])
		})
	}
}

// ----------------------------------------------------------------------------
// Helpers...

type (
	dpOpts struct {
		coOpts
		rev        string
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

var _ DPLister = (*dp)(nil)

func makeDPLister(opts dpOpts) *dp {
	return &dp{
		name: "d1",
		opts: opts,
	}
}

func (d *dp) CPUResourceLimits() config.Allocations {
	return config.Allocations{
		UnderPerc: 100,
		OverPerc:  50,
	}
}

func (d *dp) MEMResourceLimits() config.Allocations {
	return config.Allocations{
		UnderPerc: 100,
		OverPerc:  50,
	}
}

func (d *dp) ListPodsBySelector(ns string, sel *metav1.LabelSelector) map[string]*v1.Pod {
	return map[string]*v1.Pod{
		"default/p1": makeFullPod(podOpts{
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

func (d *dp) ListServiceAccounts() map[string]*v1.ServiceAccount {
	return nil
}

func (d *dp) ListPodsMetrics() map[string]*mv1beta1.PodMetrics {
	return map[string]*mv1beta1.PodMetrics{
		cache.FQN("default", "p1"): makeMxPod(d.opts.ccpu, d.opts.cmem),
	}
}

func (d *dp) ListDeployments() map[string]*appsv1.Deployment {
	return map[string]*appsv1.Deployment{
		cache.FQN("default", d.name): makeDP(d.name, d.opts),
	}
}

func (d *dp) DeploymentPreferredRev() string {
	return "apps/v1"
}

func makeDP(n string, o dpOpts) *appsv1.Deployment {
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      n,
			Namespace: "default",
			SelfLink:  "/api/" + o.rev,
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
