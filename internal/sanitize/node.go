package sanitize

import (
	"context"

	"github.com/derailed/popeye/internal"
	"github.com/derailed/popeye/internal/client"
	"github.com/derailed/popeye/internal/issues"
	v1 "k8s.io/api/core/v1"
	mv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

const (
	// Label for master nodes since v1.6
	labelNodeRoleMaster = "node-role.kubernetes.io/master"

	// Future label for master nodes as of v1.20,
	// according to https://github.com/kubernetes/kubeadm/issues/2200
	labelNodeRoleControlPlane = "node-role.kubernetes.io/control-plane"
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

	// Node represents a Node sanitizer.
	Node struct {
		*issues.Collector
		NodeLister
	}
)

// NewNode returns a new sanitizer.
func NewNode(co *issues.Collector, lister NodeLister) *Node {
	return &Node{
		Collector:  co,
		NodeLister: lister,
	}
}

// Sanitize cleanse the resource.
func (n *Node) Sanitize(ctx context.Context) error {
	nmx := client.NodesMetrics{}
	nodesMetrics(n.ListNodes(), n.ListNodesMetrics(), nmx)

	var numMasters int
	var masterCtx context.Context
	for fqn, no := range n.ListNodes() {
		n.InitOutcome(fqn)
		ctx = internal.WithFQN(ctx, fqn)

		if n.checkMasterRole(no) {
			masterCtx = ctx
			numMasters++
		}

		ready := n.checkConditions(ctx, no)
		if ready {
			n.checkTaints(ctx, no.Spec.Taints)
			n.checkUtilization(ctx, nmx[fqn])
		}

		if n.NoConcerns(fqn) && n.Config.ExcludeFQN(internal.MustExtractSectionGVR(ctx), fqn) {
			n.ClearOutcome(fqn)
		}
	}

	if numMasters == 1 {
		n.AddCode(masterCtx, 712)
	}

	return nil
}

func (n *Node) checkTaints(ctx context.Context, taints []v1.Taint) {
	tols := n.fetchPodTolerations()
	for _, ta := range taints {
		if _, ok := tols[mkKey(ta.Key, ta.Value)]; !ok {
			n.AddCode(ctx, 700, ta.Key)
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

func (n *Node) checkConditions(ctx context.Context, no *v1.Node) bool {
	var ready bool
	if no.Spec.Unschedulable {
		n.AddCode(ctx, 711)
	}
	for _, c := range no.Status.Conditions {
		// Unknow type
		if c.Status == v1.ConditionUnknown {
			n.AddCode(ctx, 701)
			return false
		}

		// Node is not ready bail other checks
		if c.Type == v1.NodeReady && c.Status == v1.ConditionFalse {
			n.AddCode(ctx, 702)
			return ready
		}
		ready = n.statusReport(ctx, c.Type, c.Status)
	}

	return ready
}

// checkMasterRole checks whether the node is a master node.
func (n *Node) checkMasterRole(no *v1.Node) bool {
	if _, ok := no.Labels[labelNodeRoleMaster]; ok {
		return true
	}
	if _, ok := no.Labels[labelNodeRoleControlPlane]; ok {
		return true
	}

	return false
}

func (n *Node) statusReport(ctx context.Context, cond v1.NodeConditionType, status v1.ConditionStatus) bool {
	var ready bool

	// Status is good ie no condition detected -> bail!
	if status == v1.ConditionFalse {
		return true
	}

	switch cond {
	case v1.NodeMemoryPressure:
		n.AddCode(ctx, 704)
	case v1.NodeDiskPressure:
		n.AddCode(ctx, 705)
	case v1.NodePIDPressure:
		n.AddCode(ctx, 706)
	case v1.NodeNetworkUnavailable:
		n.AddCode(ctx, 707)
	case v1.NodeReady:
		ready = true
	}

	return ready
}

func (n *Node) checkUtilization(ctx context.Context, mx client.NodeMetrics) {
	if mx.Empty() {
		n.AddCode(ctx, 708)
		return
	}

	percCPU := ToPerc(toMC(mx.CurrentCPU), toMC(mx.AvailableCPU))
	cpuLimit := int64(n.NodeCPULimit())
	if percCPU > cpuLimit {
		n.AddCode(ctx, 709, cpuLimit, percCPU)
	}

	percMEM := ToPerc(toMB(mx.CurrentMEM), toMB(mx.AvailableMEM))
	memLimit := int64(n.NodeMEMLimit())
	if percMEM > memLimit {
		n.AddCode(ctx, 710, memLimit, percMEM)
	}
}

// ----------------------------------------------------------------------------
// Helpers...

func nodesMetrics(nodes map[string]*v1.Node, metrics map[string]*mv1beta1.NodeMetrics, nmx client.NodesMetrics) {
	if len(metrics) == 0 {
		return
	}

	for fqn, n := range nodes {
		nmx[fqn] = client.NodeMetrics{
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
