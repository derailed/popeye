package sanitize

import (
	"context"
	"testing"

	"github.com/derailed/popeye/internal/cache"
	"github.com/derailed/popeye/internal/issues"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	mv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
	v1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

func TestNodeSanitizer(t *testing.T) {
	uu := map[string]struct {
		lister NodeLister
		issues int
	}{
		"good": {
			makeNodeLister("n1", nodeOpts{
				nodes: map[string]*v1.Node{
					"n1": makeNode("n1", "1000m", "200Mi"),
				},
				metrics: map[string]*mv1beta1.NodeMetrics{
					"n1": makeNodeMX("500m", "100Mi"),
				},
			}),
			0,
		},
		"noMetrics": {
			makeNodeLister("n1", nodeOpts{
				noMetrics: true,
				nodes: map[string]*v1.Node{
					"n1": makeNode("n1", "", ""),
				},
			}),
			1,
		},
		"overCPU": {
			makeNodeLister("n1", nodeOpts{
				nodes: map[string]*v1.Node{
					"n1": makeNode("n1", "1000m", "200Mi"),
				},
				metrics: map[string]*mv1beta1.NodeMetrics{
					"n1": makeNodeMX("2000m", "100Mi"),
				},
			}),
			1,
		},
		"overMem": {
			makeNodeLister("n1", nodeOpts{
				nodes: map[string]*v1.Node{
					"n1": makeNode("n1", "1", "100Mi"),
				},
				metrics: map[string]*mv1beta1.NodeMetrics{
					"n1": makeNodeMX("500m", "250Mi"),
				},
			}),
			1,
		},
		"missingToleration": {
			makeNodeLister("n1", nodeOpts{
				nodes: map[string]*v1.Node{
					"n1": makeTaintedNode("n1", "fred", "blee"),
				},
				pods: map[string]*v1.Pod{
					cache.FQN("default", "p1"): makePod("p1"),
					cache.FQN("default", "p2"): makePodToleration("p2", "k1", "v1"),
				},
				metrics: map[string]*mv1beta1.NodeMetrics{
					"n1": makeNodeMX("10m", "10Mi"),
				},
			}),
			1,
		},
		"notReady": {
			makeNodeLister("n1", nodeOpts{
				nodes: map[string]*v1.Node{
					"n1": makeCondNode("n1", v1.NodeReady, v1.ConditionFalse),
				},
				metrics: map[string]*mv1beta1.NodeMetrics{
					"n1": makeNodeMX("500m", "100Mi"),
				},
			}),
			1,
		},
		"unknownState": {
			makeNodeLister("n1", nodeOpts{
				nodes: map[string]*v1.Node{
					"n1": makeCondNode("n1", v1.NodeReady, v1.ConditionUnknown),
				},
				metrics: map[string]*mv1beta1.NodeMetrics{
					"n1": makeNodeMX("500m", "100Mi"),
				},
			}),
			1,
		},
		"outOfDisk": {
			makeNodeLister("n1", nodeOpts{
				nodes: map[string]*v1.Node{
					"n1": makeCondNode("n1", v1.NodeOutOfDisk, v1.ConditionTrue),
				},
				metrics: map[string]*mv1beta1.NodeMetrics{
					"n1": makeNodeMX("500m", "100Mi"),
				},
			}),
			1,
		},
		"outOfMemory": {
			makeNodeLister("n1", nodeOpts{
				nodes: map[string]*v1.Node{
					"n1": makeCondNode("n1", v1.NodeMemoryPressure, v1.ConditionTrue),
				},
				metrics: map[string]*mv1beta1.NodeMetrics{
					"n1": makeNodeMX("500m", "100Mi"),
				},
			}),
			1,
		},
		"diskPressure": {
			makeNodeLister("n1", nodeOpts{
				nodes: map[string]*v1.Node{
					"n1": makeCondNode("n1", v1.NodeDiskPressure, v1.ConditionTrue),
				},
				metrics: map[string]*mv1beta1.NodeMetrics{
					"n1": makeNodeMX("500m", "100Mi"),
				},
			}),
			1,
		},
		"outOfPID": {
			makeNodeLister("n1", nodeOpts{
				nodes: map[string]*v1.Node{
					"n1": makeCondNode("n1", v1.NodePIDPressure, v1.ConditionTrue),
				},
				metrics: map[string]*mv1beta1.NodeMetrics{
					"n1": makeNodeMX("500m", "100Mi"),
				},
			}),
			1,
		},
		"noNetwork": {
			makeNodeLister("n1", nodeOpts{
				nodes: map[string]*v1.Node{
					"n1": makeCondNode("n1", v1.NodeNetworkUnavailable, v1.ConditionTrue),
				},
				metrics: map[string]*mv1beta1.NodeMetrics{
					"n1": makeNodeMX("500m", "100Mi"),
				},
			}),
			1,
		},
	}

	for k, u := range uu {
		t.Run(k, func(t *testing.T) {
			n := NewNode(issues.NewCollector(), u.lister)
			n.Sanitize(context.Background())

			assert.Equal(t, u.issues, len(n.Outcome()["n1"]))
		})
	}
}

// ----------------------------------------------------------------------------
// Helpers...

type (
	nodeOpts struct {
		noMetrics bool
		nodes     map[string]*v1.Node
		metrics   map[string]*v1beta1.NodeMetrics
		pods      map[string]*v1.Pod
	}

	node struct {
		name string
		opts nodeOpts
	}
)

func makeNodeLister(n string, opts nodeOpts) *node {
	return &node{
		name: n,
		opts: opts,
	}
}

func (*node) RestartsLimit() int {
	return 10
}

func (*node) PodCPULimit() float64 {
	return 90
}

func (*node) PodMEMLimit() float64 {
	return 90
}

func (*node) NodeCPULimit() float64 {
	return 90
}

func (*node) NodeMEMLimit() float64 {
	return 90
}

func (n *node) ListNodesMetrics() map[string]*v1beta1.NodeMetrics {
	if n.opts.noMetrics {
		return map[string]*v1beta1.NodeMetrics{}
	}

	return n.opts.metrics
}

func (n *node) ListPods() map[string]*v1.Pod {
	return n.opts.pods
}

func (n *node) ListPodsMetrics() map[string]*v1beta1.PodMetrics {
	return map[string]*v1beta1.PodMetrics{}
}

func makePodToleration(n, k, v string) *v1.Pod {
	p := makePod(n)
	p.Spec.Tolerations = []v1.Toleration{
		{Key: k, Value: v},
	}
	return p
}

func (n *node) ListNodes() map[string]*v1.Node {
	return n.opts.nodes
}

func makeCondNode(n string, c v1.NodeConditionType, s v1.ConditionStatus) *v1.Node {
	no := makeNode(n, "100m", "100Mi")
	no.Status = v1.NodeStatus{
		Conditions: []v1.NodeCondition{
			{Type: c, Status: s},
		},
	}
	return no
}

func makeNode(n, cpu, mem string) *v1.Node {
	no := v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: n,
		},
		Spec: v1.NodeSpec{},
		Status: v1.NodeStatus{
			Conditions: []v1.NodeCondition{
				{Type: v1.NodeReady, Status: v1.ConditionTrue},
			},
		},
	}

	if cpu != "" {
		no.Status.Allocatable = v1.ResourceList{
			v1.ResourceCPU:    toQty(cpu),
			v1.ResourceMemory: toQty(mem),
		}
	}

	return &no
}

func makeTaintedNode(n, k, v string) *v1.Node {
	no := makeNode(n, "100m", "100Mi")
	no.Spec.Taints = []v1.Taint{
		{Key: k, Value: v},
	}
	return no
}

func makeMxNode(name, cpu, mem string) *v1beta1.NodeMetrics {
	return &v1beta1.NodeMetrics{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Usage: makeRes(cpu, mem),
	}
}

func makeNodeMX(cpu, mem string) *v1beta1.NodeMetrics {
	return &v1beta1.NodeMetrics{
		Usage: v1.ResourceList{
			v1.ResourceCPU:    toQty(cpu),
			v1.ResourceMemory: toQty(mem),
		},
	}
}
