package scrub

import (
	"context"

	"github.com/derailed/popeye/internal/cache"
	"github.com/derailed/popeye/internal/dag"
	"github.com/derailed/popeye/internal/issues"
	"github.com/derailed/popeye/internal/sanitize"
	"github.com/derailed/popeye/pkg/config"
)

// HorizontalPodAutoscaler represents a HorizontalPodAutoscaler scruber.
type HorizontalPodAutoscaler struct {
	*issues.Collector
	*cache.HorizontalPodAutoscaler
	*cache.Pod
	*cache.Node
	*cache.PodsMetrics
	*cache.NodesMetrics
	*cache.Deployment
	*cache.StatefulSet
	*config.Config
}

// NewHorizontalPodAutoscaler return a new HorizontalPodAutoscaler scruber.
func NewHorizontalPodAutoscaler(ctx context.Context, c *Cache, codes *issues.Codes) Sanitizer {
	h := HorizontalPodAutoscaler{
		Collector: issues.NewCollector(codes, c.config),
		Config:    c.config,
	}

	var err error
	ss, err := dag.ListHorizontalPodAutoscalers(c.factory, c.config)
	if err != nil {
		h.AddErr(ctx, err)
	}
	h.HorizontalPodAutoscaler = cache.NewHorizontalPodAutoscaler(ss)

	h.Deployment, err = c.deployments()
	if err != nil {
		h.AddErr(ctx, err)
	}

	h.StatefulSet, err = c.statefulsets()
	if err != nil {
		h.AddErr(ctx, err)
	}

	h.Node, err = c.nodes()
	if err != nil {
		h.AddCode(ctx, 402, err)
	}

	h.NodesMetrics, _ = c.nodesMx()

	h.Pod, err = c.pods()
	if err != nil {
		h.AddErr(ctx, err)
	}

	h.PodsMetrics, _ = c.podsMx()

	return &h
}

// Sanitize all available HorizontalPodAutoscalers.
func (h *HorizontalPodAutoscaler) Sanitize(ctx context.Context) error {
	return sanitize.NewHorizontalPodAutoscaler(h.Collector, h).Sanitize(ctx)
}
