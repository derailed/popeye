package scrub

import (
	"context"

	"github.com/derailed/popeye/internal/cache"
	"github.com/derailed/popeye/internal/dag"
	"github.com/derailed/popeye/internal/issues"
	"github.com/derailed/popeye/internal/k8s"
	"github.com/derailed/popeye/internal/sanitize"
	"github.com/derailed/popeye/pkg/config"
)

// Node represents a Node sanitizer.
type Node struct {
	*issues.Collector
	*cache.Node
	*cache.Pod
	*cache.NodesMetrics
	*config.Config
}

// NewNode return a new Node sanitizer.
func NewNode(c *k8s.Client, cfg *config.Config) Sanitizer {
	n := Node{Collector: issues.NewCollector(), Config: cfg}

	nn, err := dag.ListNodes(c, cfg)
	if err != nil {
		n.AddErr("nodes", err)
	}
	n.Node = cache.NewNode(nn)

	pp, err := dag.ListPods(c, cfg)
	if err != nil {
		n.AddErr("pod", err)
	}
	n.Pod = cache.NewPod(pp)

	nmx, err := dag.ListNodesMetrics(c)
	if err != nil {
		n.AddErr("nodemetrics", err)
	}
	n.NodesMetrics = cache.NewNodesMetrics(nmx)

	return &n
}

// Sanitize all available Nodes.
func (n *Node) Sanitize(ctx context.Context) error {
	return sanitize.NewNode(n.Collector, n).Sanitize(ctx)
}
