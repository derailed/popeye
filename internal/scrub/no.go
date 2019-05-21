package scrub

import (
	"context"

	"github.com/derailed/popeye/internal/cache"
	"github.com/derailed/popeye/internal/dag"
	"github.com/derailed/popeye/internal/issues"
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
func NewNode(c *Cache) Sanitizer {
	n := Node{
		Collector: issues.NewCollector(),
		Config:    c.config,
	}

	nn, err := dag.ListNodes(c.client, c.config)
	if err != nil {
		n.AddErr("nodes", err)
	}
	n.Node = cache.NewNode(nn)

	pod, err := c.pods()
	if err != nil {
		n.AddErr("pods", err)
	}
	n.Pod = pod

	nmx, err := c.nodesMx()
	if err != nil {
		n.AddErr("nodemetrics", err)
	}
	n.NodesMetrics = nmx

	return &n
}

// Sanitize all available Nodes.
func (n *Node) Sanitize(ctx context.Context) error {
	return sanitize.NewNode(n.Collector, n).Sanitize(ctx)
}
