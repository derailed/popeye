package linter

import (
	"math"

	v1 "k8s.io/api/core/v1"
)

const (
	// BOZO!! Set in a config file?
	cpuLimit = 80
	memLimit = 80
)

type (
	// Node represents a Node linter.
	Node struct {
		*Linter
	}

	// NodeMetric tracks node metrics available and current range.
	NodeMetric interface {
		CurrentCPU() int64
		CurrentMEM() int64
		MaxCPU() int64
		MaxMEM() int64
		Empty() bool
	}
)

// NewNode returns a new Node linter.
func NewNode() *Node {
	return &Node{new(Linter)}
}

// Lint a Node.
func (n *Node) Lint(no v1.Node, mx NodeMetric) {
	n.checkUtilization(no.Name, mx)
}

func (n *Node) checkUtilization(no string, mx NodeMetric) {
	// No metrics bail out!
	if mx.Empty() {
		return
	}

	percCPU := math.Round(float64(mx.CurrentCPU()) / float64(mx.MaxCPU()) * 100)
	if percCPU >= cpuLimit {
		n.addIssuef(WarnLevel, "CPU threshold reached on node `%s (%0.f%%)", no, percCPU)
	}

	percMEM := math.Round(float64(mx.CurrentMEM()) / float64(mx.MaxMEM()) * 100)
	if percMEM >= memLimit {
		n.addIssuef(WarnLevel, "Memory threshold reached on node `%s (%0.f%%)", no, percMEM)
	}
}
