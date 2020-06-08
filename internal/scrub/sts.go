package scrub

import (
	"context"

	"github.com/derailed/popeye/internal/cache"
	"github.com/derailed/popeye/internal/issues"
	"github.com/derailed/popeye/internal/sanitize"
	"github.com/derailed/popeye/pkg/config"
)

// StatefulSet represents a StatefulSet scruber.
type StatefulSet struct {
	*issues.Collector
	*cache.Pod
	*cache.StatefulSet
	*cache.PodsMetrics
	*cache.ServiceAccount
	*config.Config
}

// NewStatefulSet return a new StatefulSet scruber.
func NewStatefulSet(ctx context.Context, c *Cache, codes *issues.Codes) Sanitizer {
	s := StatefulSet{
		Collector: issues.NewCollector(codes, c.config),
		Config:    c.config,
	}

	var err error
	s.StatefulSet, err = c.statefulsets()
	if err != nil {
		s.AddErr(ctx, err)
	}

	s.Pod, err = c.pods()
	if err != nil {
		s.AddErr(ctx, err)
	}

	s.PodsMetrics, _ = c.podsMx()

	s.ServiceAccount, err = c.serviceaccounts()
	if err != nil {
		s.AddErr(ctx, err)
	}

	return &s
}

// Sanitize all available StatefulSets.
func (c *StatefulSet) Sanitize(ctx context.Context) error {
	return sanitize.NewStatefulSet(c.Collector, c).Sanitize(ctx)
}
