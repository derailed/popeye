package scrub

import (
	"github.com/derailed/popeye/internal/cache"
	"github.com/derailed/popeye/internal/dag"
)

type policy struct {
	*dial

	np  *cache.NetworkPolicy
	psp *cache.PodSecurityPolicy
}

func newPolicy(d *dial) *policy {
	return &policy{dial: d}
}

func (p *policy) podsecuritypolicies() (*cache.PodSecurityPolicy, error) {
	if p.psp != nil {
		return p.psp, nil
	}
	psps, err := dag.ListPodSecurityPolicies(p.factory, p.config)
	p.psp = cache.NewPodSecurityPolicy(psps)

	return p.psp, err
}

func (p *policy) networkpolicies() (*cache.NetworkPolicy, error) {
	if p.np != nil {
		return p.np, nil
	}
	nps, err := dag.ListNetworkPolicies(p.factory, p.config)
	p.np = cache.NewNetworkPolicy(nps)

	return p.np, err
}
