package scrub

import (
	"context"

	"github.com/derailed/popeye/internal/cache"
	"github.com/derailed/popeye/internal/issues"
	"github.com/derailed/popeye/internal/sanitize"
)

// Secret represents a Secret scruber.
type Secret struct {
	*issues.Collector
	*cache.Secret
	*cache.Pod
	*cache.ServiceAccount
	*cache.Ingress
}

// NewSecret return a new Secret scruber.
func NewSecret(ctx context.Context, c *Cache, codes *issues.Codes) Sanitizer {
	s := Secret{Collector: issues.NewCollector(codes, c.config)}

	var err error
	s.Secret, err = c.secrets()
	if err != nil {
		s.AddErr(ctx, err)
	}

	s.Pod, err = c.pods()
	if err != nil {
		s.AddErr(ctx, err)
	}

	s.ServiceAccount, err = c.serviceaccounts()
	if err != nil {
		s.AddErr(ctx, err)
	}

	s.Ingress, err = c.ingresses()
	if err != nil {
		s.AddErr(ctx, err)
	}

	return &s
}

// Sanitize all available Secrets.
func (c *Secret) Sanitize(ctx context.Context) error {
	return sanitize.NewSecret(c.Collector, c).Sanitize(ctx)
}
