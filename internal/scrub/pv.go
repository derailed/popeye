package scrub

import (
	"context"

	"github.com/derailed/popeye/internal/cache"
	"github.com/derailed/popeye/internal/dag"
	"github.com/derailed/popeye/internal/issues"
	"github.com/derailed/popeye/internal/sanitize"
)

// PersistentVolume represents a PersistentVolume sanitizer.
type PersistentVolume struct {
	*issues.Collector
	*cache.PersistentVolume
	*cache.Pod
}

// NewPersistentVolume return a new PersistentVolume sanitizer.
func NewPersistentVolume(c *Cache) Sanitizer {
	p := PersistentVolume{Collector: issues.NewCollector()}

	ss, err := dag.ListPersistentVolumes(c.client, c.config)
	if err != nil {
		p.AddErr("services", err)
	}
	p.PersistentVolume = cache.NewPersistentVolume(ss)

	pod, err := c.pods()
	if err != nil {
		p.AddErr("pods", err)
	}
	p.Pod = pod

	return &p
}

// Sanitize all available PersistentVolumes.
func (s *PersistentVolume) Sanitize(ctx context.Context) error {
	return sanitize.NewPersistentVolume(s.Collector, s).Sanitize(ctx)
}
