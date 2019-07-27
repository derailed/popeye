package scrub

import (
	"context"

	"github.com/derailed/popeye/internal/cache"
	"github.com/derailed/popeye/internal/issues"
	"github.com/derailed/popeye/internal/k8s"
	"github.com/derailed/popeye/internal/sanitize"
	"github.com/derailed/popeye/pkg/config"
)

// Deployment represents a Deployment sanitizer.
type Deployment struct {
	*issues.Collector
	*cache.Deployment
	*cache.PodsMetrics
	*cache.Pod
	*config.Config

	client *k8s.Client
}

// NewDeployment return a new Deployment sanitizer.
func NewDeployment(c *Cache, codes *issues.Codes) Sanitizer {
	d := Deployment{
		client:    c.client,
		Config:    c.config,
		Collector: issues.NewCollector(codes),
	}

	dps, err := c.deployments()
	if err != nil {
		d.AddErr("deployments", err)
	}
	d.Deployment = dps

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

// Sanitize all available Deployments.
func (d *Deployment) Sanitize(ctx context.Context) error {
	return sanitize.NewDeployment(d.Collector, d).Sanitize(ctx)
}
