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

// Pod represents a Pod sanitizer.
type Pod struct {
	*issues.Collector
	*cache.Pod
	*cache.PodsMetrics
	*config.Config
}

// NewPod return a new Pod sanitizer.
func NewPod(c *k8s.Client, cfg *config.Config) Sanitizer {
	p := Pod{Collector: issues.NewCollector(), Config: cfg}

	pods, err := dag.ListPods(c, cfg)
	if err != nil {
		p.AddErr("pods", err)
	}
	pmx, err := dag.ListPodsMetrics(c)
	if err != nil {
		p.AddInfof("podmetrics", "No metric-server detected %v", err)
	}
	p.Pod, p.PodsMetrics = cache.NewPod(pods), cache.NewPodsMetrics(pmx)

	return &p
}

// Sanitize all available Pods.
func (p *Pod) Sanitize(ctx context.Context) error {
	return sanitize.NewPod(p.Collector, p).Sanitize(ctx)
}
