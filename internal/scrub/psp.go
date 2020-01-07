package scrub

import (
	"context"

	"github.com/derailed/popeye/internal/cache"
	"github.com/derailed/popeye/internal/issues"
	"github.com/derailed/popeye/internal/k8s"
	"github.com/derailed/popeye/internal/sanitize"
	"github.com/derailed/popeye/pkg/config"
)

// PodSecurityPolicy represents a PodSecurityPolicy scruber.
type PodSecurityPolicy struct {
	*issues.Collector
	*cache.PodSecurityPolicy
	*config.Config

	client *k8s.Client
}

// NewPodSecurityPolicy return a new PodSecurityPolicy scruber.
func NewPodSecurityPolicy(ctx context.Context, c *Cache, codes *issues.Codes) Sanitizer {
	p := PodSecurityPolicy{
		client:    c.client,
		Config:    c.config,
		Collector: issues.NewCollector(codes, c.config),
	}

	var err error
	p.PodSecurityPolicy, err = c.podsecuritypolicies()
	if err != nil {
		p.AddErr(ctx, err)
	}

	return &p
}

// Sanitize all available PodSecurityPolicys.
func (p *PodSecurityPolicy) Sanitize(ctx context.Context) error {
	return sanitize.NewPodSecurityPolicy(p.Collector, p).Sanitize(ctx)
}
