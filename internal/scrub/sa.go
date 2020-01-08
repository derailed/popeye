package scrub

import (
	"context"

	"github.com/derailed/popeye/internal/cache"
	"github.com/derailed/popeye/internal/issues"
	"github.com/derailed/popeye/internal/sanitize"
)

// ServiceAccount represents a ServiceAccount scruber.
type ServiceAccount struct {
	*issues.Collector
	*cache.ServiceAccount
	*cache.Pod
	*cache.ClusterRoleBinding
	*cache.RoleBinding
	*cache.Secret
	*cache.Ingress
}

// NewServiceAccount return a new ServiceAccount scruber.
func NewServiceAccount(ctx context.Context, c *Cache, codes *issues.Codes) Sanitizer {
	s := ServiceAccount{Collector: issues.NewCollector(codes, c.config)}

	var err error
	s.ServiceAccount, err = c.serviceaccounts()
	if err != nil {
		s.AddErr(ctx, err)
	}

	s.Pod, err = c.pods()
	if err != nil {
		s.AddErr(ctx, err)
	}

	s.ClusterRoleBinding, err = c.clusterrolebindings()
	if err != nil {
		s.AddErr(ctx, err)
	}

	s.RoleBinding, err = c.rolebindings()
	if err != nil {
		s.AddErr(ctx, err)
	}

	s.Secret, err = c.secrets()
	if err != nil {
		s.AddErr(ctx, err)
	}

	s.Ingress, err = c.ingresses()
	if err != nil {
		s.AddErr(ctx, err)
	}

	return &s
}

// Sanitize all available ServiceAccounts.
func (s *ServiceAccount) Sanitize(ctx context.Context) error {
	return sanitize.NewServiceAccount(s.Collector, s).Sanitize(ctx)
}
