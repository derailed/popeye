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

// ServiceAccouunt represents a ServiceAccouunt sanitizer.
type ServiceAccouunt struct {
	*issues.Collector
	*cache.ServiceAccount
	*cache.Pod
	*cache.ClusterRoleBinding
	*cache.RoleBinding
}

// NewServiceAccount return a new ServiceAccouunt sanitizer.
func NewServiceAccount(c *k8s.Client, cfg *config.Config) Sanitizer {
	p := ServiceAccouunt{Collector: issues.NewCollector()}

	ss, err := dag.ListServiceAccounts(c, cfg)
	if err != nil {
		p.AddErr("serviceaccounts", err)
	}
	p.ServiceAccount = cache.NewServiceAccount(ss)

	pp, err := dag.ListPods(c, cfg)
	if err != nil {
		p.AddErr("pod", err)
	}
	p.Pod = cache.NewPod(pp)

	crbs, err := dag.ListClusterRoleBindings(c, cfg)
	if err != nil {
		p.AddErr("clusterrolebindings", err)
	}
	p.ClusterRoleBinding = cache.NewClusterRoleBinding(crbs)

	rbs, err := dag.ListRoleBindings(c, cfg)
	if err != nil {
		p.AddErr("rolebindings", err)
	}
	p.RoleBinding = cache.NewRoleBinding(rbs)

	return &p
}

// Sanitize all available ServiceAccouunts.
func (s *ServiceAccouunt) Sanitize(ctx context.Context) error {
	return sanitize.NewServiceAccount(s.Collector, s).Sanitize(ctx)
}
