package scrub

import (
	"context"

	"github.com/derailed/popeye/internal/cache"
	"github.com/derailed/popeye/internal/issues"
	"github.com/derailed/popeye/internal/k8s"
	"github.com/derailed/popeye/internal/sanitize"
	"github.com/derailed/popeye/pkg/config"
)

// Cluster represents a Cluster scruber.
type Cluster struct {
	*issues.Collector
	*cache.Cluster
	*config.Config

	client *k8s.Client
}

// NewCluster return a new Cluster scruber.
func NewCluster(ctx context.Context, c *Cache, codes *issues.Codes) Sanitizer {
	cl := Cluster{
		client:    c.client,
		Config:    c.config,
		Collector: issues.NewCollector(codes, c.config),
	}

	var err error
	cl.Cluster, err = c.cluster()
	if err != nil {
		cl.AddErr(ctx, err)
	}

	return &cl
}

// Sanitize all available Clusters.
func (d *Cluster) Sanitize(ctx context.Context) error {
	return sanitize.NewCluster(d.Collector, d).Sanitize(ctx)
}
