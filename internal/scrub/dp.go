package scrub

import (
	"context"

	"github.com/derailed/popeye/internal/cache"
	"github.com/derailed/popeye/internal/issues"
	"github.com/derailed/popeye/internal/k8s"
	"github.com/derailed/popeye/internal/sanitize"
	"github.com/derailed/popeye/pkg/config"
)

// Deployment represents a Deployment scruber.
type Deployment struct {
	*issues.Collector
	*cache.Deployment
	*cache.PodsMetrics
	*cache.Pod
	*config.Config

	client *k8s.Client
}

// NewDeployment return a new Deployment scruber.
func NewDeployment(ctx context.Context, c *Cache, codes *issues.Codes) Sanitizer {
	d := Deployment{
		client:    c.client,
		Config:    c.config,
		Collector: issues.NewCollector(codes, c.config),
	}

	var err error
	d.Deployment, err = c.deployments()
	if err != nil {
		d.AddErr(ctx, err)
	}

	d.PodsMetrics, _ = c.podsMx()

	d.Pod, err = c.pods()
	if err != nil {
		d.AddErr(ctx, err)
	}

	return &d
}

// Sanitize all available Deployments.
func (d *Deployment) Sanitize(ctx context.Context) error {
	return sanitize.NewDeployment(d.Collector, d).Sanitize(ctx)
}
