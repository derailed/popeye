package scrub

import (
	"context"

	"github.com/derailed/popeye/internal/cache"
	"github.com/derailed/popeye/internal/issues"
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
func NewStatefulSet(c *Cache) Sanitizer {
	s := StatefulSet{
		Collector: issues.NewCollector(),
		Config:    c.config,
	}

	sts, err := c.statefulsets()
	if err != nil {
		s.AddErr("statefulsets", err)
	}
	s.StatefulSet = sts

	pod, err := c.pods()
	if err != nil {
		s.AddErr("pods", err)
	}
	s.Pod = pod

	pmx, err := c.podsMx()
	if err != nil {
		s.AddInfof("podmetrics", "No metric-server detected %v", err)
	}
	s.PodsMetrics = pmx

	return &s
}

// Sanitize all available StatefulSets.
func (c *StatefulSet) Sanitize(ctx context.Context) error {
	return sanitize.NewStatefulSet(c.Collector, c).Sanitize(ctx)
}
