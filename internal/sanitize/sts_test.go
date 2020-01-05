package sanitize

import (
	"context"
	"testing"

	"github.com/derailed/popeye/internal/cache"
	"github.com/derailed/popeye/internal/issues"
	"github.com/derailed/popeye/pkg/config"
	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	mv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

func TestSTSSanitizer(t *testing.T) {
	uu := map[string]struct {
		lister StatefulSetLister
		issues issues.Issues
	}{
		"good": {
			lister: makeSTSLister(stsOpts{
				coOpts:      coOpts{rcpu: "100m", rmem: "10Mi"},
				replicas:    1,
				currentReps: 1,
				rev:         "apps/v1",
				ccpu:        "100m", cmem: "10Mi",
			}),
			issues: issues.Issues{},
		},
		"deprecated": {
			lister: makeSTSLister(stsOpts{
				coOpts:      coOpts{rcpu: "100m", rmem: "10Mi"},
				replicas:    1,
				currentReps: 1,
				rev:         "extensions/v1",
				ccpu:        "100m", cmem: "10Mi",
			}),
			issues: issues.Issues{
				issues.Issue{Group: "__root__", Level: 2, Message: `[POP-403] Deprecated StatefulSet API group "extensions/v1". Use "apps/v1" instead`},
			},
		},
		"used?": {
			lister: makeSTSLister(stsOpts{
				coOpts:      coOpts{rcpu: "100m", rmem: "10Mi"},
				replicas:    1,
				currentReps: 0,
				rev:         "apps/v1",
				ccpu:        "100m", cmem: "10Mi",
			}),
			issues: issues.Issues{
				issues.New(issues.Root, issues.WarnLevel, "[POP-501] Used? No available replicas found"),
			},
		},
		"zeroReplicas": {
			lister: makeSTSLister(stsOpts{
				coOpts:      coOpts{rcpu: "100m", rmem: "10Mi"},
				replicas:    0,
				currentReps: 1,
				rev:         "apps/v1",
				ccpu:        "100m", cmem: "10Mi",
			}),
			issues: issues.Issues{
				issues.New(issues.Root, issues.WarnLevel, "[POP-500] Zero scale detected"),
			},
		},
		"collisions": {
			lister: makeSTSLister(stsOpts{
				coOpts:      coOpts{rcpu: "100m", rmem: "10Mi"},
				replicas:    1,
				currentReps: 1,
				collisions:  1,
				rev:         "apps/v1",
				ccpu:        "100m", cmem: "10Mi",
			}),
			issues: issues.Issues{
				issues.New(issues.Root, issues.ErrorLevel, "[POP-502] ReplicaSet collisions detected (1)"),
			},
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			sts := NewStatefulSet(issues.NewCollector(loadCodes(t)), u.lister)

			assert.Nil(t, sts.Sanitize(context.TODO()))
			assert.Equal(t, u.issues, sts.Outcome()["default/sts1"])
		})
	}
}

func TestSTSSanitizerUtilization(t *testing.T) {
	uu := map[string]struct {
		lister StatefulSetLister
		issues issues.Issues
	}{
		"bestEffort": {
			lister: makeSTSLister(stsOpts{
				replicas:    1,
				currentReps: 1,
				rev:         "apps/v1",
				ccpu:        "200m", cmem: "10Mi",
			}),
			issues: issues.Issues{},
		},
		"underCPUBurstable": {
			lister: makeSTSLister(stsOpts{
				coOpts: coOpts{
					rcpu: "100m", rmem: "10Mi",
				},
				replicas:    1,
				currentReps: 1,
				rev:         "apps/v1",
				ccpu:        "200m", cmem: "10Mi",
			}),
			issues: issues.Issues{
				issues.New(issues.Root, issues.WarnLevel, "[POP-503] At current load, CPU under allocated. Current:400m vs Requested:200m (200.00%)"),
			},
		},
		"underCPUGuaranteed": {
			lister: makeSTSLister(stsOpts{
				coOpts: coOpts{
					rcpu: "100m", rmem: "10Mi",
					lcpu: "100m", lmem: "10Mi",
				},
				replicas:    1,
				currentReps: 1,
				rev:         "apps/v1",
				ccpu:        "200m", cmem: "10Mi",
			}),
			issues: issues.Issues{
				issues.New(issues.Root, issues.WarnLevel, "[POP-503] At current load, CPU under allocated. Current:400m vs Requested:200m (200.00%)"),
			},
		},
		"overCPUBurstable": {
			lister: makeSTSLister(stsOpts{
				coOpts: coOpts{
					rcpu: "400m", rmem: "10Mi",
				},
				replicas:    1,
				currentReps: 1,
				rev:         "apps/v1",
				ccpu:        "100m", cmem: "10Mi",
			}),
			issues: issues.Issues{
				issues.New(issues.Root, issues.WarnLevel, "[POP-504] At current load, CPU over allocated. Current:200m vs Requested:800m (400.00%)"),
			},
		},
		"overCPUGuarenteed": {
			lister: makeSTSLister(stsOpts{
				coOpts: coOpts{
					rcpu: "400m", rmem: "10Mi",
					lcpu: "400m", lmem: "10Mi",
				},
				replicas:    1,
				currentReps: 1,
				rev:         "apps/v1",
				ccpu:        "100m", cmem: "10Mi",
			}),
			issues: issues.Issues{
				issues.New(issues.Root, issues.WarnLevel, "[POP-504] At current load, CPU over allocated. Current:200m vs Requested:800m (400.00%)"),
			},
		},
		"underMEMBurstable": {
			lister: makeSTSLister(stsOpts{
				coOpts: coOpts{
					rcpu: "100m", rmem: "10Mi",
				},
				replicas:    1,
				currentReps: 1,
				rev:         "apps/v1",
				ccpu:        "100m", cmem: "20Mi",
			}),
			issues: issues.Issues{
				issues.New(issues.Root, issues.WarnLevel, "[POP-505] At current load, Memory under allocated. Current:40Mi vs Requested:20Mi (200.00%)"),
			},
		},
		"underMEMGuaranteed": {
			lister: makeSTSLister(stsOpts{
				coOpts: coOpts{
					rcpu: "100m", rmem: "10Mi",
					lcpu: "100m", lmem: "10Mi",
				},
				replicas:    1,
				currentReps: 1,
				rev:         "apps/v1",
				ccpu:        "100m", cmem: "20Mi",
			}),
			issues: issues.Issues{
				issues.New(issues.Root, issues.WarnLevel, "[POP-505] At current load, Memory under allocated. Current:40Mi vs Requested:20Mi (200.00%)"),
			},
		},
		"overMEMBurstable": {
			lister: makeSTSLister(stsOpts{
				coOpts: coOpts{
					rcpu: "100m", rmem: "100Mi",
				},
				replicas:    1,
				currentReps: 1,
				rev:         "apps/v1",
				ccpu:        "100m", cmem: "20Mi",
			}),
			issues: issues.Issues{
				issues.New(issues.Root, issues.WarnLevel, "[POP-506] At current load, Memory over allocated. Current:40Mi vs Requested:200Mi (500.00%)"),
			},
		},
		"overMEMGuaranteed": {
			lister: makeSTSLister(stsOpts{
				coOpts: coOpts{
					rcpu: "100m", rmem: "100Mi",
					lcpu: "100m", lmem: "100Mi",
				},
				replicas:    1,
				currentReps: 1,
				rev:         "apps/v1",
				ccpu:        "100m", cmem: "20Mi",
			}),
			issues: issues.Issues{
				issues.New(issues.Root, issues.WarnLevel, "[POP-506] At current load, Memory over allocated. Current:40Mi vs Requested:200Mi (500.00%)"),
			},
		},
	}

	ctx := context.WithValue(context.Background(), PopeyeKey("OverAllocs"), true)
	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			sts := NewStatefulSet(issues.NewCollector(loadCodes(t)), u.lister)

			assert.Nil(t, sts.Sanitize(ctx))
			assert.Equal(t, u.issues, sts.Outcome()["default/sts1"])
		})
	}
}

// ----------------------------------------------------------------------------
// Helpers...

type (
	stsOpts struct {
		coOpts
		replicas    int32
		currentReps int32
		collisions  int32
		ccpu, cmem  string
		rev         string
	}

	sts struct {
		name string
		opts stsOpts
	}
)

func makeSTSLister(opts stsOpts) *sts {
	return &sts{
		name: "sts1",
		opts: opts,
	}
}

func (s *sts) CPUResourceLimits() config.Allocations {
	return config.Allocations{
		UnderPerc: 100,
		OverPerc:  50,
	}
}

func (s *sts) MEMResourceLimits() config.Allocations {
	return config.Allocations{
		UnderPerc: 100,
		OverPerc:  50,
	}
}

func (*sts) RestartsLimit() int {
	return 10
}

func (*sts) PodCPULimit() float64 {
	return 100
}

func (*sts) PodMEMLimit() float64 {
	return 100
}

func (s *sts) ListStatefulSets() map[string]*appsv1.StatefulSet {
	return map[string]*appsv1.StatefulSet{
		cache.FQN("default", s.name): makeSTS(s.name, s.opts),
	}
}

func (s *sts) ListPodsBySelector(sel *metav1.LabelSelector) map[string]*v1.Pod {
	return map[string]*v1.Pod{
		"default/p1": makeFullPod(podOpts{
			coOpts: coOpts{
				rcpu: s.opts.rcpu,
				rmem: s.opts.rmem,
				lcpu: s.opts.lcpu,
				lmem: s.opts.lmem,
			}}),
	}
}

func (s *sts) ListPodsMetrics() map[string]*mv1beta1.PodMetrics {
	return map[string]*mv1beta1.PodMetrics{
		"default/p1": makeMxPod(s.opts.ccpu, s.opts.cmem),
	}
}

func makeSTS(n string, opts stsOpts) *appsv1.StatefulSet {
	return &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      n,
			Namespace: "default",
			SelfLink:  "/api/" + opts.rev,
		},
		Spec: appsv1.StatefulSetSpec{
			Replicas: &opts.replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"fred": "blee",
				},
			},
			Template: v1.PodTemplateSpec{
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:  "c1",
							Image: "fred:0.0.1",
							Resources: v1.ResourceRequirements{
								Requests: v1.ResourceList{
									v1.ResourceCPU:    toQty(opts.coOpts.rcpu),
									v1.ResourceMemory: toQty(opts.coOpts.rmem),
								},
								Limits: v1.ResourceList{
									v1.ResourceCPU:    toQty(opts.coOpts.lcpu),
									v1.ResourceMemory: toQty(opts.coOpts.lmem),
								},
							},
						},
					},
					InitContainers: []v1.Container{
						{
							Name:  "i1",
							Image: "fred:0.0.1",
							Resources: v1.ResourceRequirements{
								Requests: v1.ResourceList{
									v1.ResourceCPU:    toQty(opts.coOpts.rcpu),
									v1.ResourceMemory: toQty(opts.coOpts.rmem),
								},
								Limits: v1.ResourceList{
									v1.ResourceCPU:    toQty(opts.coOpts.lcpu),
									v1.ResourceMemory: toQty(opts.coOpts.lmem),
								},
							},
						},
					},
				},
			},
		},
		Status: appsv1.StatefulSetStatus{
			CurrentReplicas: opts.currentReps,
			CollisionCount:  &opts.collisions,
		},
	}
}
