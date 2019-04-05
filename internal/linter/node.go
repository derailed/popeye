package linter

import (
	"context"
	"fmt"
	"math"

	"github.com/derailed/popeye/internal/k8s"
	"github.com/rs/zerolog"
	v1 "k8s.io/api/core/v1"
	mv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

const (
	// BOZO!! Set in a config file?
	cpuLimit = 80
	memLimit = 80
)

type (
	tolerations map[string]struct{}

	// Node represents a Node linter.
	Node struct {
		*Linter
	}
)

// NewNode returns a new Node linter.
func NewNode(c Client, l *zerolog.Logger) *Node {
	return &Node{newLinter(c, l)}
}

// Lint a Node.
func (n *Node) Lint(ctx context.Context) error {
	nodes, err := n.client.ListNodes()
	if err != nil {
		return err
	}

	tt := n.fetchPodTolerations()

	var mx []mv1beta1.NodeMetrics
	nmx := make(k8s.NodesMetrics)
	if n.client.ClusterHasMetrics() {
		if mx, err = n.client.FetchNodesMetrics(); err != nil {
			return err
		}
		k8s.GetNodesMetrics(nodes, mx, nmx)
	}

	for _, no := range nodes {
		n.initIssues(no.Name)
		n.lint(no, nmx[no.Name], tt)
	}

	return nil
}

func (n *Node) lint(no v1.Node, mx k8s.NodeMetrics, t tolerations) {
	ready := n.checkConditions(no)
	if ready {
		n.checkTaints(no, t)
		n.checkUtilization(no.Name, mx)
	}
}

func (n *Node) checkTaints(no v1.Node, t tolerations) {
	for _, ta := range no.Spec.Taints {
		if _, ok := t[mkKey(ta.Key, ta.Value)]; !ok {
			n.addIssuef(no.Name, WarnLevel, "Found taint `%s but no pod can tolerate", ta.Key)
		}
	}
}

func (n *Node) fetchPodTolerations() tolerations {
	tt := tolerations{}
	pods, err := n.client.ListAllPods()
	if err != nil {
		n.addIssuef("", ErrorLevel, "Unable to list all pods %s", err)
	}
	fmt.Println(len(pods))
	for _, po := range pods {
		for _, t := range po.Spec.Tolerations {
			tt[mkKey(t.Key, t.Value)] = struct{}{}
		}
	}

	return tt
}

func mkKey(k, v string) string {
	return k + ":" + v
}

func (n *Node) checkConditions(no v1.Node) bool {
	var ready bool
	for _, c := range no.Status.Conditions {
		// Unknow type
		if c.Status == v1.ConditionUnknown {
			n.addIssuef(no.Name, ErrorLevel, "Unable to assess node condition `%s", c.Type)
			continue
		}

		// Node is not ready bail other checks
		if c.Status == v1.ConditionFalse {
			if c.Type == v1.NodeReady {
				n.addIssuef(no.Name, ErrorLevel, "Node is not in ready state")
				return ready
			}
			continue
		}

		switch c.Type {
		case v1.NodeOutOfDisk:
			n.addIssue(no.Name, ErrorLevel, "Out of disk space")
		case v1.NodeMemoryPressure:
			n.addIssue(no.Name, WarnLevel, "Insuficient memory")
		case v1.NodeDiskPressure:
			n.addIssue(no.Name, WarnLevel, "Insuficient disk space")
		case v1.NodePIDPressure:
			n.addIssue(no.Name, ErrorLevel, "Insuficient PIDS on node")
		case v1.NodeNetworkUnavailable:
			n.addIssue(no.Name, ErrorLevel, "No network configured on node")
		case v1.NodeReady:
			ready = true
		}
	}

	return ready
}

func (n *Node) checkUtilization(no string, mx k8s.NodeMetrics) {
	if mx.Empty() {
		n.addIssuef(no, WarnLevel, "No node metrics available")
		return
	}

	percCPU := ToPerc(float64(mx.CurrentCPU), float64(mx.AvailCPU))
	cpuLimit := n.client.NodeCPULimit()
	if math.Round(percCPU) > cpuLimit {
		n.addIssuef(no, WarnLevel, "CPU threshold (%0.f%%) reached %0.f%%", cpuLimit, percCPU)
	}

	percMEM := ToPerc(float64(mx.CurrentMEM), float64(mx.AvailMEM))
	memLimit := n.client.NodeMEMLimit()
	if math.Round(percMEM) > memLimit {
		n.addIssuef(no, WarnLevel, "Memory threshold (%0.f%%) reached %0.f%%", memLimit, percMEM)
	}
}
