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
func NewHorizontalPodAutoscaler(c *k8s.Client, cfg *config.Config) Sanitizer {
	p := HorizontalPodAutoscaler{Collector: issues.NewCollector(), Config: cfg}

	ss, err := dag.ListHorizontalPodAutoscalers(c, cfg)
	if err != nil {
		p.AddErr("services", err)
	}
	p.HorizontalPodAutoscaler = cache.NewHorizontalPodAutoscaler(ss)

	dps, err := dag.ListDeployments(c, cfg)
	if err != nil {
		p.AddErr("deployments", err)
	}
	p.Deployment = cache.NewDeployment(dps)

	sts, err := dag.ListStatefulSets(c, cfg)
	if err != nil {
		p.AddErr("statefulsets", err)
	}
	p.StatefulSet = cache.NewStatefulSet(sts)

	nmx, err := dag.ListNodesMetrics(c)
	if err != nil {
		p.AddErr("nodemetrics", err)
	}
	p.NodesMetrics = cache.NewNodesMetrics(nmx)

	pods, err := dag.ListPods(c, cfg)
	if err != nil {
		p.AddErr("pods", err)
	}
	p.Pod = cache.NewPod(pods)

	pmx, err := dag.ListPodsMetrics(c)
	if err != nil {
		p.AddErr("podsmetrics", err)
	}
	p.PodsMetrics = cache.NewPodsMetrics(pmx)

	return &p
}

// Sanitize all available HorizontalPodAutoscalers.
func (h *HorizontalPodAutoscaler) Sanitize(ctx context.Context) error {
	return sanitize.NewHorizontalPodAutoscaler(h.Collector, h).Sanitize(ctx)
}
