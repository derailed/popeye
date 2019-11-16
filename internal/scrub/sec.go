package scrub

import (
	"context"

	"github.com/derailed/popeye/internal/cache"
	"github.com/derailed/popeye/internal/dag"
	"github.com/derailed/popeye/internal/issues"
	"github.com/derailed/popeye/internal/sanitize"
)

// Secret represents a Secret sanitizer.
type Secret struct {
	*issues.Collector
	*cache.Secret
	*cache.Pod
	*cache.ServiceAccount
	*cache.Ingress
}

// NewSecret return a new Secret sanitizer.
func NewSecret(c *Cache, codes *issues.Codes) Sanitizer {
	s := Secret{Collector: issues.NewCollector(codes)}

	secs, err := dag.ListSecrets(c.client, c.config)
	if err != nil {
		s.AddErr("secrets", err)
	}
	s.Secret = cache.NewSecret(secs)

	pod, err := c.pods()
	if err != nil {
		s.AddErr("pods", err)
	}
	s.Pod = pod

	sas, err := c.serviceaccounts()
	if err != nil {
		s.AddErr("serviceaccounts", err)
	}
	s.ServiceAccount = sas

	ing, err := c.ingresses()
	if err != nil {
		s.AddErr("ingresses", err)
	}
	s.Ingress = ing

	return &s
}

// Sanitize all available Secrets.
func (c *Secret) Sanitize(ctx context.Context) error {
	return sanitize.NewSecret(c.Collector, c).Sanitize(ctx)
}
