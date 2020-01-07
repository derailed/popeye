package scrub

import (
	"context"

	"github.com/derailed/popeye/internal/cache"
	"github.com/derailed/popeye/internal/issues"
	"github.com/derailed/popeye/internal/sanitize"
)

// PersistentVolumeClaim represents a PersistentVolumeClaim scruber.
type PersistentVolumeClaim struct {
	*issues.Collector
	*cache.PersistentVolumeClaim
	*cache.Pod
}

// NewPersistentVolumeClaim return a new PersistentVolumeClaim scruber.
func NewPersistentVolumeClaim(ctx context.Context, c *Cache, codes *issues.Codes) Sanitizer {
	p := PersistentVolumeClaim{
		Collector: issues.NewCollector(codes, c.config),
	}

	var err error
	p.PersistentVolumeClaim, err = c.persistentvolumeclaims()
	if err != nil {
		p.AddErr(ctx, err)
	}

	p.Pod, err = c.pods()
	if err != nil {
		p.AddErr(ctx, err)
	}

	return &p
}

// Sanitize all available PersistentVolumeClaims.
func (s *PersistentVolumeClaim) Sanitize(ctx context.Context) error {
	return sanitize.NewPersistentVolumeClaim(s.Collector, s).Sanitize(ctx)
}
