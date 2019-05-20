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

// PersistentVolumeClaim represents a PersistentVolumeClaim sanitizer.
type PersistentVolumeClaim struct {
	*issues.Collector
	*cache.PersistentVolumeClaim
	*cache.Pod
}

// NewPersistentVolumeClaim return a new PersistentVolumeClaim sanitizer.
func NewPersistentVolumeClaim(c *k8s.Client, cfg *config.Config) Sanitizer {
	p := PersistentVolumeClaim{Collector: issues.NewCollector()}

	ss, err := dag.ListPersistentVolumeClaims(c, cfg)
	if err != nil {
		p.AddErr("services", err)
	}
	p.PersistentVolumeClaim = cache.NewPersistentVolumeClaim(ss)

	pp, err := dag.ListPods(c, cfg)
	if err != nil {
		p.AddErr("pod", err)
	}
	p.Pod = cache.NewPod(pp)

	return &p
}

// Sanitize all available PersistentVolumeClaims.
func (s *PersistentVolumeClaim) Sanitize(ctx context.Context) error {
	return sanitize.NewPersistentVolumeClaim(s.Collector, s).Sanitize(ctx)
}
