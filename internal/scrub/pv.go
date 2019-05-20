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

// PersistentVolume represents a PersistentVolume sanitizer.
type PersistentVolume struct {
	*issues.Collector
	*cache.PersistentVolume
	*cache.Pod
}

// NewPersistentVolume return a new PersistentVolume sanitizer.
func NewPersistentVolume(c *k8s.Client, cfg *config.Config) Sanitizer {
	p := PersistentVolume{Collector: issues.NewCollector()}

	ss, err := dag.ListPersistentVolumes(c, cfg)
	if err != nil {
		p.AddErr("services", err)
	}
	p.PersistentVolume = cache.NewPersistentVolume(ss)

	pp, err := dag.ListPods(c, cfg)
	if err != nil {
		p.AddErr("pod", err)
	}
	p.Pod = cache.NewPod(pp)

	return &p
}

// Sanitize all available PersistentVolumes.
func (s *PersistentVolume) Sanitize(ctx context.Context) error {
	return sanitize.NewPersistentVolume(s.Collector, s).Sanitize(ctx)
}
