package scrub

import (
	"context"

	"github.com/derailed/popeye/internal/cache"
	"github.com/derailed/popeye/internal/issues"
	"github.com/derailed/popeye/internal/sanitize"
	"github.com/derailed/popeye/pkg/config"
)

// Pod represents a Pod scruber.
type Pod struct {
	*issues.Collector
	*cache.Pod
	*cache.PodsMetrics
	*config.Config
	*cache.PodDisruptionBudget
	*cache.ServiceAccount
}

// NewPod return a new Pod scruber.
func NewPod(ctx context.Context, c *Cache, codes *issues.Codes) Sanitizer {
	p := Pod{
		Collector: issues.NewCollector(codes, c.config),
		Config:    c.config,
	}

	var err error
	p.Pod, err = c.pods()
	if err != nil {
		p.AddErr(ctx, err)
	}

	p.PodsMetrics, _ = c.podsMx()

	p.PodDisruptionBudget, err = c.podDisruptionBudgets()
	if err != nil {
		p.AddErr(ctx, err)
	}

	p.ServiceAccount, err = c.serviceaccounts()
	if err != nil {
		p.AddErr(ctx, err)
	}

	return &p
}

// Sanitize all available Pods.
func (p *Pod) Sanitize(ctx context.Context) error {
	return sanitize.NewPod(p.Collector, p).Sanitize(ctx)
}
