package sanitize

import (
	"testing"

	"github.com/derailed/popeye/internal/cache"
	"github.com/derailed/popeye/internal/issues"
	"github.com/stretchr/testify/assert"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	mv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

func TestHPASanitizeDP(t *testing.T) {
	uu := map[string]struct {
		l       HpaLister
		issues  int
		hissues int
	}{
		"cool": {
			newDpHpa(
				hpaOpts{
					name: "d1",
					ccpu: "20m",
					cmem: "20Mi",
					max:  1,
					coOpts: coOpts{
						rcpu: "1m",
						rmem: "10Mi",
					},
				}),
			0,
			0,
		},
		"noDeployments": {
			newDpHpa(
				hpaOpts{
					name: "bozo",
					ccpu: "20m",
					cmem: "20Mi",
					max:  1,
					coOpts: coOpts{
						rcpu: "1m",
						rmem: "10Mi",
					},
				}),
			1,
			0,
		},
		"overCpu": {
			newDpHpa(
				hpaOpts{
					name: "d1",
					ccpu: "10m",
					cmem: "20Mi",
					max:  1,
					coOpts: coOpts{
						rcpu: "10m",
						rmem: "10Mi",
					},
				}),
			1,
			1,
		},
		"overMem": {
			newDpHpa(
				hpaOpts{
					name: "d1",
					ccpu: "10m",
					cmem: "10Mi",
					max:  1,
					coOpts: coOpts{
						rcpu: "1m",
						rmem: "10Mi",
					},
				}),
			1,
			1,
		},
	}

	ctx := makeContext("hpa")
	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			h := NewHorizontalPodAutoscaler(issues.NewCollector(loadCodes(t), makeConfig(t)), u.l)

			assert.Nil(t, h.Sanitize(ctx))
			assert.Equal(t, u.issues, len(h.Outcome()["default/h1"]))
			assert.Equal(t, u.hissues, len(h.Outcome()["HPA"]))
		})
	}
}

func TestHPASanitizeSTS(t *testing.T) {
	uu := map[string]struct {
		l       HpaLister
		issues  int
		hissues int
	}{
		"cool": {
			newStsHpa(
				hpaOpts{
					name: "sts1",
					ccpu: "10m",
					cmem: "10Mi",
					max:  1,
					coOpts: coOpts{
						rcpu: "1m",
						rmem: "1Mi",
					},
				}),
			0,
			0,
		},
		"noSTS": {
			newStsHpa(
				hpaOpts{
					name: "bozo",
					ccpu: "20m",
					cmem: "20Mi",
					max:  1,
					coOpts: coOpts{
						rcpu: "1m",
						rmem: "10Mi",
					},
				}),
			1,
			0,
		},
		"overCpu": {
			newStsHpa(
				hpaOpts{
					name: "sts1",
					ccpu: "10m",
					cmem: "10Mi",
					max:  2,
					coOpts: coOpts{
						rcpu: "10m",
						rmem: "1Mi",
						lcpu: "10m",
						lmem: "1Mi",
					},
				}),
			1,
			1,
		},
		"overMem": {
			newStsHpa(
				hpaOpts{
					name: "sts1",
					ccpu: "10m",
					cmem: "10Mi",
					max:  1,
					coOpts: coOpts{
						rcpu: "1m",
						rmem: "10Mi",
						lcpu: "1m",
						lmem: "10Mi",
					},
				}),
			1,
			1,
		},
	}

	ctx := makeContext("hpa")
	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			h := NewHorizontalPodAutoscaler(issues.NewCollector(loadCodes(t), makeConfig(t)), u.l)

			assert.Nil(t, h.Sanitize(ctx))
			assert.Equal(t, u.issues, len(h.Outcome()["default/h1"]))
			assert.Equal(t, u.hissues, len(h.Outcome()["HPA"]))
		})
	}
}

// ----------------------------------------------------------------------------
// Helpers...

type hpaOpts struct {
	coOpts
	name                     string
	refType, ref, ccpu, cmem string
	max                      int32
}

type hpa struct {
	StatefulSetLister
	DeploymentLister
	name string
	opts hpaOpts
}

func newDpHpa(opts hpaOpts) *hpa {
	h := hpa{
		DeploymentLister: makeDPLister(dpOpts{
			coOpts:    opts.coOpts,
			reps:      1,
			availReps: 1,
		}),
		name: "h1",
		opts: opts,
	}
	h.opts.refType, h.opts.ref = "Deployment", opts.name

	return &h
}

func newStsHpa(opts hpaOpts) *hpa {
	h := hpa{
		StatefulSetLister: makeSTSLister(stsOpts{
			coOpts:      opts.coOpts,
			replicas:    1,
			currentReps: 1,
		}),
		name: "h1",
		opts: opts,
	}
	h.opts.refType, h.opts.ref = "StatefulSet", opts.name

	return &h
}

func (h *hpa) ListHorizontalPodAutoscalers() map[string]*autoscalingv1.HorizontalPodAutoscaler {
	return map[string]*autoscalingv1.HorizontalPodAutoscaler{
		cache.FQN("default", h.name): makeHPA(h.name, h.opts.refType, h.opts.ref, h.opts.max),
	}
}

func (h *hpa) ListNodesMetrics() map[string]*mv1beta1.NodeMetrics {
	return map[string]*mv1beta1.NodeMetrics{}
}

func (h *hpa) ListNodes() map[string]*v1.Node {
	return map[string]*v1.Node{}
}

func (h *hpa) ListPods() map[string]*v1.Pod {
	return map[string]*v1.Pod{}
}

func (h *hpa) NodeCPULimit() float64 { return 0 }
func (h *hpa) NodeMEMLimit() float64 { return 0 }

func (h *hpa) ListPodsMetrics() map[string]*mv1beta1.PodMetrics {
	return map[string]*mv1beta1.PodMetrics{
		"default/p1": makeMxPod(h.opts.rcpu, h.opts.rmem),
	}
}

func (h *hpa) ListAvailableMetrics(map[string]*v1.Node) v1.ResourceList {
	return v1.ResourceList{
		v1.ResourceCPU:    toQty(h.opts.ccpu),
		v1.ResourceMemory: toQty(h.opts.cmem),
	}
}

func (h *hpa) GetPod(string, map[string]string) *v1.Pod {
	return &v1.Pod{}
}

func makeHPA(n, kind, dp string, max int32) *autoscalingv1.HorizontalPodAutoscaler {
	return &autoscalingv1.HorizontalPodAutoscaler{
		ObjectMeta: metav1.ObjectMeta{
			Name:      n,
			Namespace: "default",
		},
		Spec: autoscalingv1.HorizontalPodAutoscalerSpec{
			MaxReplicas: max,
			ScaleTargetRef: autoscalingv1.CrossVersionObjectReference{
				Kind: kind,
				Name: dp,
			},
		},
	}
}
