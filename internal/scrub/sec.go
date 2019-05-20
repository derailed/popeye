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

// Secret represents a Secret sanitizer.
type Secret struct {
	*issues.Collector
	*cache.Secret
	*cache.Pod
	*cache.ServiceAccount
}

// NewSecret return a new Secret sanitizer.
func NewSecret(c *k8s.Client, cfg *config.Config) Sanitizer {
	s := Secret{Collector: issues.NewCollector()}

	secs, err := dag.ListSecrets(c, cfg)
	if err != nil {
		s.AddErr("secrets", err)
	}
	pods, err := dag.ListPods(c, cfg)
	if err != nil {
		s.AddErr("pods", err)
	}
	sas, err := dag.ListServiceAccounts(c, cfg)
	if err != nil {
		s.AddErr("serviceaccounts", err)
	}

	s.Secret = cache.NewSecret(secs)
	s.Pod = cache.NewPod(pods)
	s.ServiceAccount = cache.NewServiceAccount(sas)
	return &s
}

// Sanitize all available Secrets.
func (c *Secret) Sanitize(ctx context.Context) error {
	return sanitize.NewSecret(c.Collector, c).Sanitize(ctx)
}
