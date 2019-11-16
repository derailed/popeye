package sanitize

import (
	"context"

	"github.com/derailed/popeye/internal/issues"
	"github.com/derailed/popeye/internal/k8s"
	v1 "k8s.io/api/core/v1"
	mv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

type (
	tolerations map[string]struct{}

	// NodeLimiter tracks metrics limit range.
	NodeLimiter interface {
		NodeCPULimit() float64
		NodeMEMLimit() float64
	}

	// NodeLister lists available nodes.
	NodeLister interface {
		NodeMetricsLister
		PodLister
		NodeLimiter
		ListNodes() map[string]*v1.Node
	}

	// NodeMetricsLister handle
	NodeMetricsLister interface {
		ListNodesMetrics() map[string]*mv1beta1.NodeMetrics
	}

	// Node represents a Node linter.
	Node struct {
		*issues.Collector
		NodeLister
	}
)

// NewNode returns a new Node linter.
func NewNode(co *issues.Collector, lister NodeLister) *Node {
	return &Node{
		Collector:  co,
		NodeLister: lister,
	}
}

// Sanitize a Node.
func (n *Node) Sanitize(ctx context.Context) error {
	nmx := k8s.NodesMetrics{}
	nodesMetrics(n.ListNodes(), n.ListNodesMetrics(), nmx)
	for fqn, no := range n.ListNodes() {
		n.InitOutcome(fqn)
		ready := n.checkConditions(no)
		if ready {
			n.checkTaints(fqn, no.Spec.Taints)
			n.checkUtilization(fqn, nmx[fqn])
		}
	}

	return nil
}

func (n *Node) checkTaints(fqn string, taints []v1.Taint) {
	tols := n.fetchPodTolerations()
	for _, ta := range taints {
		if _, ok := tols[mkKey(ta.Key, ta.Value)]; !ok {
			n.AddCode(700, fqn, ta.Key)
		}
	}
}

func (n *Node) fetchPodTolerations() tolerations {
	tt := tolerations{}
	for _, po := range n.ListPods() {
		for _, t := range po.Spec.Tolerations {
			tt[mkKey(t.Key, t.Value)] = struct{}{}
		}
	}

	return tt
}

func mkKey(k, v string) string {
	return k + ":" + v
}

func (n *Node) checkConditions(no *v1.Node) bool {
	var ready bool
	for _, c := range no.Status.Conditions {
		// Unknow type
		if c.Status == v1.ConditionUnknown {
			n.AddCode(701, no.Name)
			continue
		}

		// Node is not ready bail other checks
		if c.Type == v1.NodeReady && c.Status == v1.ConditionFalse {
			n.AddCode(702, no.Name)
			return ready
		}
		ready = n.statusReport(no.Name, c.Type)
	}

	return ready
}

func (n *Node) statusReport(node string, cond v1.NodeConditionType) bool {
	var ready bool

	switch cond {
	case v1.NodeOutOfDisk:
		n.AddCode(703, node)
	case v1.NodeMemoryPressure:
		n.AddCode(704, node)
	case v1.NodeDiskPressure:
		n.AddCode(705, node)
	case v1.NodePIDPressure:
		n.AddCode(706, node)
	case v1.NodeNetworkUnavailable:
		n.AddCode(707, node)
	case v1.NodeReady:
		ready = true
	}

	return ready
}

func (n *Node) checkUtilization(no string, mx k8s.NodeMetrics) {
	if mx.Empty() {
		n.AddCode(708, no)
		return
	}

	percCPU := ToPerc(toMC(mx.CurrentCPU), toMC(mx.AvailableCPU))
	cpuLimit := int64(n.NodeCPULimit())
	if percCPU > cpuLimit {
		n.AddCode(709, no, cpuLimit, percCPU)
	}

	percMEM := ToPerc(toMB(mx.CurrentMEM), toMB(mx.AvailableMEM))
	memLimit := int64(n.NodeMEMLimit())
	if percMEM > memLimit {
		n.AddCode(710, no, memLimit, percMEM)
	}
}

// ----------------------------------------------------------------------------
// Helpers...

func nodesMetrics(nodes map[string]*v1.Node, metrics map[string]*mv1beta1.NodeMetrics, nmx k8s.NodesMetrics) {
	if len(metrics) == 0 {
		return
	}

	for fqn, n := range nodes {
		nmx[fqn] = k8s.NodeMetrics{
			AvailableCPU: *n.Status.Allocatable.Cpu(),
			AvailableMEM: *n.Status.Allocatable.Memory(),
			TotalCPU:     *n.Status.Capacity.Cpu(),
			TotalMEM:     *n.Status.Capacity.Memory(),
		}
	}

	for fqn, c := range metrics {
		if mx, ok := nmx[fqn]; ok {
			mx.CurrentCPU = *c.Usage.Cpu()
			mx.CurrentMEM = *c.Usage.Memory()
			nmx[fqn] = mx
		}
	}
}
