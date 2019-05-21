package scrub

import (
	"context"

	"github.com/derailed/popeye/internal/cache"
	"github.com/derailed/popeye/internal/issues"
	"github.com/derailed/popeye/internal/sanitize"
	"github.com/derailed/popeye/pkg/config"
)

// Pod represents a Pod sanitizer.
type Pod struct {
	*issues.Collector
	*cache.Pod
	*cache.PodsMetrics
	*config.Config
}

// NewPod return a new Pod sanitizer.
func NewPod(c *Cache) Sanitizer {
	p := Pod{
		Collector: issues.NewCollector(),
		Config:    c.config,
	}

	pod, err := c.pods()
	if err != nil {
		p.AddErr("pods", err)
	}
	p.Pod = pod

	pmx, err := c.podsMx()
	if err != nil {
		p.AddErr("podmetrics", err)
	}
	p.PodsMetrics = pmx

	return &p
}

// Sanitize all available Pods.
func (p *Pod) Sanitize(ctx context.Context) error {
	return sanitize.NewPod(p.Collector, p).Sanitize(ctx)
}
