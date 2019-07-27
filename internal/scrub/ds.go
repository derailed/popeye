package scrub

import (
	"context"

	"github.com/derailed/popeye/internal/cache"
	"github.com/derailed/popeye/internal/issues"
	"github.com/derailed/popeye/internal/k8s"
	"github.com/derailed/popeye/internal/sanitize"
	"github.com/derailed/popeye/pkg/config"
)

// DaemonSet represents a DaemonSet sanitizer.
type DaemonSet struct {
	*issues.Collector
	*cache.DaemonSet
	*cache.PodsMetrics
	*cache.Pod
	*config.Config

	client *k8s.Client
}

// NewDaemonSet return a new DaemonSet sanitizer.
func NewDaemonSet(c *Cache, codes *issues.Codes) Sanitizer {
	d := DaemonSet{
		client:    c.client,
		Config:    c.config,
		Collector: issues.NewCollector(codes),
	}

	ds, err := c.daemonSets()
	if err != nil {
		d.AddErr("daemonSets", err)
	}
	d.DaemonSet = ds

	pmx, err := c.podsMx()
	if err != nil {
		d.AddCode(402, "podmetrics", err)
	}
	d.PodsMetrics = pmx

	pod, err := c.pods()
	if err != nil {
		d.AddErr("pods", err)
	}
	d.Pod = pod

	return &d
}

// Sanitize all available DaemonSets.
func (d *DaemonSet) Sanitize(ctx context.Context) error {
	return sanitize.NewDaemonSet(d.Collector, d).Sanitize(ctx)
}
