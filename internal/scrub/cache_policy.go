package scrub

import (
	"context"
	"sync"

	"github.com/derailed/popeye/internal"
	"github.com/derailed/popeye/internal/cache"
	"github.com/derailed/popeye/internal/client"
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
	ctx, cancel := p.context()
	defer cancel()
	psps, err := dag.ListPodSecurityPolicies(ctx)
	p.psp = cache.NewPodSecurityPolicy(psps)

	return p.psp, err
}

func (p *policy) networkpolicies() (*cache.NetworkPolicy, error) {
	p.mx.Lock()
	defer p.mx.Unlock()

	if p.np != nil {
		return p.np, nil
	}
	ctx, cancel := p.context()
	defer cancel()
	nps, err := dag.ListNetworkPolicies(ctx)
	p.np = cache.NewNetworkPolicy(nps)

	return p.np, err
}

// Helpers...

func (p *policy) context() (context.Context, context.CancelFunc) {
	ctx := context.WithValue(context.Background(), internal.KeyFactory, p.factory)
	ctx = context.WithValue(ctx, internal.KeyConfig, p.config)
	if p.config.Flags.ActiveNamespace != nil {
		ctx = context.WithValue(ctx, internal.KeyNamespace, *p.config.Flags.ActiveNamespace)
	} else {
		ns, err := p.factory.Client().Config().CurrentNamespaceName()
		if err != nil {
			ns = client.AllNamespaces
		}
		ctx = context.WithValue(ctx, internal.KeyNamespace, ns)
	}

	return context.WithCancel(ctx)
}
