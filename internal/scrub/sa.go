package scrub

import (
	"context"

	"github.com/derailed/popeye/internal/cache"
	"github.com/derailed/popeye/internal/dag"
	"github.com/derailed/popeye/internal/issues"
	"github.com/derailed/popeye/internal/sanitize"
)

// ServiceAccount represents a ServiceAccount sanitizer.
type ServiceAccount struct {
	*issues.Collector
	*cache.ServiceAccount
	*cache.Pod
	*cache.ClusterRoleBinding
	*cache.RoleBinding
	*cache.Secret
	*cache.Ingress
}

// NewServiceAccount return a new ServiceAccount sanitizer.
func NewServiceAccount(c *Cache, codes *issues.Codes) Sanitizer {
	s := ServiceAccount{Collector: issues.NewCollector(codes)}

	sas, err := c.serviceaccounts()
	if err != nil {
		s.AddErr("serviceaccounts", err)
	}
	s.ServiceAccount = sas

	pod, err := c.pods()
	if err != nil {
		s.AddErr("pods", err)
	}
	s.Pod = pod

	crbs, err := dag.ListClusterRoleBindings(c.client, c.config)
	if err != nil {
		s.AddErr("clusterrolebindings", err)
	}
	s.ClusterRoleBinding = cache.NewClusterRoleBinding(crbs)

	rbs, err := dag.ListRoleBindings(c.client, c.config)
	if err != nil {
		s.AddErr("rolebindings", err)
	}
	s.RoleBinding = cache.NewRoleBinding(rbs)

	secrets, err := dag.ListSecrets(c.client, c.config)
	if err != nil {
		s.AddErr("secrets", err)
	}
	s.Secret = cache.NewSecret(secrets)

	ing, err := c.ingresses()
	if err != nil {
		s.AddErr("ingresses", err)
	}
	s.Ingress = ing

	return &s
}

// Sanitize all available ServiceAccounts.
func (s *ServiceAccount) Sanitize(ctx context.Context) error {
	return sanitize.NewServiceAccount(s.Collector, s).Sanitize(ctx)
}
