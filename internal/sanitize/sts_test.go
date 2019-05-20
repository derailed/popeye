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
		issues map[string]int
	}{
		"good": {
			makeSTSLister("sts1", stsOpts{
				coOpts: coOpts{
					rcpu: "100m",
					rmem: "10Mi",
				},
				replicas:    1,
				currentReps: 1,
			}),
			map[string]int{"default/sts1": 0},
		},
		"noReplicas": {
			makeSTSLister("sts1", stsOpts{
				coOpts: coOpts{
					rcpu: "100m",
					rmem: "10Mi",
				},
				replicas:    0,
				currentReps: 0,
			}),
			map[string]int{"default/sts1": 2},
		},
		"collisions": {
			makeSTSLister("sts1", stsOpts{
				coOpts: coOpts{
					rcpu: "100m",
					rmem: "10Mi",
				},
				replicas:    1,
				currentReps: 1,
				collisions:  1,
			}),
			map[string]int{"default/sts1": 1},
		},
		"overCPU": {
			makeSTSLister("sts1", stsOpts{
				coOpts: coOpts{
					rcpu: "200m",
					rmem: "10Mi",
					lcpu: "200m",
					lmem: "10Mi",
				},
				replicas:    1,
				currentReps: 1,
			}),
			map[string]int{"default/sts1": 1},
		},
		"overMem": {
			makeSTSLister("sts1", stsOpts{
				coOpts: coOpts{
					rcpu: "100m",
					rmem: "20Mi",
					lcpu: "100m",
					lmem: "20Mi",
				},
				replicas:    1,
				currentReps: 1,
			}),
			map[string]int{"default/sts1": 1},
		},
		"underCPU": {
			makeSTSLister("sts1", stsOpts{
				coOpts: coOpts{
					rcpu: "10m",
					rmem: "10Mi",
					lcpu: "10m",
					lmem: "10Mi",
				},
				replicas:    1,
				currentReps: 1,
			}),
			map[string]int{"default/sts1": 1},
		},
		"underMem": {
			makeSTSLister("sts1", stsOpts{
				coOpts: coOpts{
					rcpu: "100m",
					rmem: "1Mi",
					lcpu: "100m",
					lmem: "1Mi",
				},
				replicas:    1,
				currentReps: 1,
			}),
			map[string]int{"default/sts1": 1},
		},
	}

	for k, u := range uu {
		t.Run(k, func(t *testing.T) {
			s := NewStatefulSet(issues.NewCollector(), u.lister)
			s.Sanitize(context.Background())

			for sts, v := range u.issues {
				assert.Equal(t, v, len(s.Outcome()[sts]))
			}
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
	}

	sts struct {
		name string
		opts stsOpts
	}
)

func makeSTSLister(n string, opts stsOpts) *sts {
	return &sts{
		name: n,
		opts: opts,
	}
}

func (s *sts) CPUResourceLimits() config.Allocations {
	return config.Allocations{
		Under: 50,
		Over:  100,
	}
}

func (s *sts) MEMResourceLimits() config.Allocations {
	return config.Allocations{
		Under: 50,
		Over:  100,
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
		"default/p1": makePod("p1"),
	}
}

func (s *sts) ListPodsMetrics() map[string]*mv1beta1.PodMetrics {
	return map[string]*mv1beta1.PodMetrics{
		"default/p1": makeMxPod("1", "100m", "10Mi"),
	}
}

func makeSTS(n string, opts stsOpts) *appsv1.StatefulSet {
	return &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      n,
			Namespace: "default",
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
