package scrub

import (
	"sync"

	"github.com/derailed/popeye/internal/cache"
	"github.com/derailed/popeye/internal/dag"
)

type policy struct {
	*dial

	mx  sync.Mutex
	np  *cache.NetworkPolicy
	psp *cache.PodSecurityPolicy
}

func newPolicy(d *dial) *policy {
	return &policy{dial: d}
}

func (p *policy) podsecuritypolicies() (*cache.PodSecurityPolicy, error) {
	p.mx.Lock()
	defer p.mx.Unlock()

	if p.psp != nil {
		return p.psp, nil
	}
	psps, err := dag.ListPodSecurityPolicies(p.factory, p.config)
	p.psp = cache.NewPodSecurityPolicy(psps)

	return p.psp, err
}

func (p *policy) networkpolicies() (*cache.NetworkPolicy, error) {
	p.mx.Lock()
	defer p.mx.Unlock()

	if p.np != nil {
		return p.np, nil
	}
	nps, err := dag.ListNetworkPolicies(p.factory, p.config)
	p.np = cache.NewNetworkPolicy(nps)

	return p.np, err
}
