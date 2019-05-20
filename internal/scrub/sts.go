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

// StatefulSet represents a StatefulSet sanitizer.
type StatefulSet struct {
	*issues.Collector
	*cache.Pod
	*cache.StatefulSet
	*cache.PodsMetrics
	*config.Config
}

// NewStatefulSet return a new StatefulSet sanitizer.
func NewStatefulSet(c *k8s.Client, cfg *config.Config) Sanitizer {
	s := StatefulSet{Collector: issues.NewCollector(), Config: cfg}

	sts, err := dag.ListStatefulSets(c, cfg)
	if err != nil {
		s.AddErr("configmaps", err)
	}
	s.StatefulSet = cache.NewStatefulSet(sts)

	pods, err := dag.ListPods(c, cfg)
	if err != nil {
		s.AddErr("pods", err)
	}
	s.Pod = cache.NewPod(pods)

	pmx, err := dag.ListPodsMetrics(c)
	if err != nil {
		s.AddErr("podmetrics", err)
	}
	s.PodsMetrics = cache.NewPodsMetrics(pmx)

	return &s
}

// Sanitize all available StatefulSets.
func (c *StatefulSet) Sanitize(ctx context.Context) error {
	return sanitize.NewStatefulSet(c.Collector, c).Sanitize(ctx)
}
