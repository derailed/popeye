package scrub

import (
	"context"

	"github.com/derailed/popeye/internal/cache"
	"github.com/derailed/popeye/internal/dag"
	"github.com/derailed/popeye/internal/issues"
	"github.com/derailed/popeye/internal/sanitize"
)

// ServiceAccouunt represents a ServiceAccouunt sanitizer.
type ServiceAccouunt struct {
	*issues.Collector
	*cache.ServiceAccount
	*cache.Pod
	*cache.ClusterRoleBinding
	*cache.RoleBinding
}

// NewServiceAccount return a new ServiceAccouunt sanitizer.
func NewServiceAccount(c *Cache) Sanitizer {
	s := ServiceAccouunt{Collector: issues.NewCollector()}

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

	return &s
}

// Sanitize all available ServiceAccouunts.
func (s *ServiceAccouunt) Sanitize(ctx context.Context) error {
	return sanitize.NewServiceAccount(s.Collector, s).Sanitize(ctx)
}
