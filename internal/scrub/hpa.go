package scrub

import (
	"context"

	"github.com/derailed/popeye/internal/cache"
	"github.com/derailed/popeye/internal/dag"
	"github.com/derailed/popeye/internal/issues"
	"github.com/derailed/popeye/internal/sanitize"
	"github.com/derailed/popeye/pkg/config"
)

// HorizontalPodAutoscaler represents a HorizontalPodAutoscaler sanitizer.
type HorizontalPodAutoscaler struct {
	*issues.Collector
	*cache.HorizontalPodAutoscaler
	*cache.Pod
	*cache.PodsMetrics
	*cache.Deployment
	*cache.StatefulSet
	*cache.NodesMetrics
	*config.Config
}

// NewHorizontalPodAutoscaler return a new HorizontalPodAutoscaler sanitizer.
func NewHorizontalPodAutoscaler(c *Cache) Sanitizer {
	h := HorizontalPodAutoscaler{
		Collector: issues.NewCollector(),
		Config:    c.config,
	}

	ss, err := dag.ListHorizontalPodAutoscalers(c.client, c.config)
	if err != nil {
		h.AddErr("services", err)
	}
	h.HorizontalPodAutoscaler = cache.NewHorizontalPodAutoscaler(ss)

	dps, err := c.deployments()
	if err != nil {
		h.AddErr("deployments", err)
	}
	h.Deployment = dps

	sts, err := c.statefulsets()
	if err != nil {
		h.AddErr("statefulsets", err)
	}
	h.StatefulSet = sts

	nmx, err := c.nodesMx()
	if err != nil {
		h.AddInfof("nodemetrics", "No metric-server detected %v", err)
	}
	h.NodesMetrics = nmx

	pod, err := c.pods()
	if err != nil {
		h.AddErr("pods", err)
	}
	h.Pod = pod

	pmx, err := c.podsMx()
	if err != nil {
		h.AddInfof("podmetrics", "No metric-server detected %v", err)
	}
	h.PodsMetrics = pmx

	return &h
}

// Sanitize all available HorizontalPodAutoscalers.
func (h *HorizontalPodAutoscaler) Sanitize(ctx context.Context) error {
	return sanitize.NewHorizontalPodAutoscaler(h.Collector, h).Sanitize(ctx)
}
