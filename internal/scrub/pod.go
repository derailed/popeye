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
	*cache.PodDisruptionBudget
}

// NewPod return a new Pod sanitizer.
func NewPod(c *Cache, codes *issues.Codes) Sanitizer {
	p := Pod{
		Collector: issues.NewCollector(codes),
		Config:    c.config,
	}

	pod, err := c.pods()
	if err != nil {
		p.AddErr("pods", err)
	}
	p.Pod = pod

	pmx, err := c.podsMx()
	if err != nil {
		p.AddCode(402, "podmetrics", err)
	}
	p.PodsMetrics = pmx

	pdb, err := c.podDisruptionBudgets()
	if err != nil {
		p.AddErr("podDisruptionBudget", err)
	}
	p.PodDisruptionBudget = pdb

	return &p
}

// Sanitize all available Pods.
func (p *Pod) Sanitize(ctx context.Context) error {
	return sanitize.NewPod(p.Collector, p).Sanitize(ctx)
}
