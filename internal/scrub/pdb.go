package scrub

import (
	"context"

	"github.com/derailed/popeye/internal/cache"
	"github.com/derailed/popeye/internal/dag"
	"github.com/derailed/popeye/internal/issues"
	"github.com/derailed/popeye/internal/sanitize"
)

// PodDisruptionBudget represents a pdb sanitizer.
type PodDisruptionBudget struct {
	*issues.Collector
	*cache.Pod
	*cache.PodDisruptionBudget
}

// NewPodDisruptionBudget return a new PodDisruptionBudget sanitizer.
func NewPodDisruptionBudget(c *Cache, codes *issues.Codes) Sanitizer {
	s := PodDisruptionBudget{Collector: issues.NewCollector(codes)}

	cms, err := dag.ListPodDisruptionBudgets(c.client, c.config)
	if err != nil {
		s.AddErr("PodDisruptionBudget", err)
	}
	s.PodDisruptionBudget = cache.NewPodDisruptionBudget(cms)

	pod, err := c.pods()
	if err != nil {
		s.AddErr("pods", err)
	}
	s.Pod = pod

	return &s
}

// Sanitize all available PodDisruptionBudgets.
func (c *PodDisruptionBudget) Sanitize(ctx context.Context) error {
	return sanitize.NewPodDisruptionBudget(c.Collector, c).Sanitize(ctx)
}
