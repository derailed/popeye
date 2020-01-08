package scrub

import (
	"context"

	"github.com/derailed/popeye/internal/cache"
	"github.com/derailed/popeye/internal/issues"
	"github.com/derailed/popeye/internal/k8s"
	"github.com/derailed/popeye/internal/sanitize"
	"github.com/derailed/popeye/pkg/config"
)

// Ingress represents a Ingress scruber.
type Ingress struct {
	*issues.Collector
	*cache.Ingress
	*config.Config

	client *k8s.Client
}

// NewIngress return a new Ingress scruber.
func NewIngress(ctx context.Context, c *Cache, codes *issues.Codes) Sanitizer {
	d := Ingress{
		client:    c.client,
		Config:    c.config,
		Collector: issues.NewCollector(codes, c.config),
	}

	var err error
	d.Ingress, err = c.ingresses()
	if err != nil {
		d.AddErr(ctx, err)
	}

	return &d
}

// Sanitize all available Ingresss.
func (i *Ingress) Sanitize(ctx context.Context) error {
	return sanitize.NewIngress(i.Collector, i).Sanitize(ctx)
}
