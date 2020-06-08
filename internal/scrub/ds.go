package scrub

import (
	"context"

	"github.com/derailed/popeye/internal/cache"
	"github.com/derailed/popeye/internal/issues"
	"github.com/derailed/popeye/internal/sanitize"
	"github.com/derailed/popeye/pkg/config"
	"github.com/derailed/popeye/types"
)

// DaemonSet represents a DaemonSet scruber.
type DaemonSet struct {
	*issues.Collector
	*cache.DaemonSet
	*cache.PodsMetrics
	*cache.Pod
	*cache.ServiceAccount
	*config.Config

	client types.Connection
}

// NewDaemonSet return a new DaemonSet scruber.
func NewDaemonSet(ctx context.Context, c *Cache, codes *issues.Codes) Sanitizer {
	d := DaemonSet{
		client:    c.factory.Client(),
		Config:    c.config,
		Collector: issues.NewCollector(codes, c.config),
	}

	var err error
	d.DaemonSet, err = c.daemonSets()
	if err != nil {
		d.AddErr(ctx, err)
	}

	d.Pod, err = c.pods()
	if err != nil {
		d.AddErr(ctx, err)
	}
	d.PodsMetrics, _ = c.podsMx()

	d.ServiceAccount, err = c.serviceaccounts()
	if err != nil {
		d.AddErr(ctx, err)
	}

	return &d
}

// Sanitize all available DaemonSets.
func (d *DaemonSet) Sanitize(ctx context.Context) error {
	return sanitize.NewDaemonSet(d.Collector, d).Sanitize(ctx)
}
