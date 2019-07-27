package scrub

import (
	"context"

	"github.com/derailed/popeye/internal/cache"
	"github.com/derailed/popeye/internal/dag"
	"github.com/derailed/popeye/internal/issues"
	"github.com/derailed/popeye/internal/sanitize"
)

// PersistentVolumeClaim represents a PersistentVolumeClaim sanitizer.
type PersistentVolumeClaim struct {
	*issues.Collector
	*cache.PersistentVolumeClaim
	*cache.Pod
}

// NewPersistentVolumeClaim return a new PersistentVolumeClaim sanitizer.
func NewPersistentVolumeClaim(c *Cache, codes *issues.Codes) Sanitizer {
	p := PersistentVolumeClaim{
		Collector: issues.NewCollector(codes),
	}

	ss, err := dag.ListPersistentVolumeClaims(c.client, c.config)
	if err != nil {
		p.AddErr("services", err)
	}
	p.PersistentVolumeClaim = cache.NewPersistentVolumeClaim(ss)

	pod, err := c.pods()
	if err != nil {
		p.AddErr("pods", err)
	}
	p.Pod = pod

	return &p
}

// Sanitize all available PersistentVolumeClaims.
func (s *PersistentVolumeClaim) Sanitize(ctx context.Context) error {
	return sanitize.NewPersistentVolumeClaim(s.Collector, s).Sanitize(ctx)
}
