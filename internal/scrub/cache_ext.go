package scrub

import (
	"context"
	"sync"

	"github.com/derailed/popeye/internal"
	"github.com/derailed/popeye/internal/cache"
	"github.com/derailed/popeye/internal/client"
	"github.com/derailed/popeye/internal/dag"
)

type ext struct {
	*dial

	mx  sync.Mutex
	pdb *cache.PodDisruptionBudget
	ing *cache.Ingress
	cl  *cache.Cluster
}

func newExt(d *dial) *ext {
	return &ext{dial: d}
}

func (e *ext) cluster() (*cache.Cluster, error) {
	e.mx.Lock()
	defer e.mx.Unlock()

	if e.cl != nil {
		return e.cl, nil
	}
	ctx, cancel := e.context()
	defer cancel()
	major, minor, err := dag.ListVersion(ctx)
	e.cl = cache.NewCluster(major, minor)

	return e.cl, err
}

func (e *ext) ingresses() (*cache.Ingress, error) {
	e.mx.Lock()
	defer e.mx.Unlock()

	if e.ing != nil {
		return e.ing, nil
	}
	ctx, cancel := e.context()
	defer cancel()
	ings, err := dag.ListIngresses(ctx)
	e.ing = cache.NewIngress(ings)

	return e.ing, err
}

func (e *ext) podDisruptionBudgets() (*cache.PodDisruptionBudget, error) {
	e.mx.Lock()
	defer e.mx.Unlock()

	if e.pdb != nil {
		return e.pdb, nil
	}
	ctx, cancel := e.context()
	defer cancel()
	pdbs, err := dag.ListPodDisruptionBudgets(ctx)
	e.pdb = cache.NewPodDisruptionBudget(pdbs)

	return e.pdb, err
}

// Helpers...

func (e *ext) context() (context.Context, context.CancelFunc) {
	ctx := context.WithValue(context.Background(), internal.KeyFactory, e.factory)
	ctx = context.WithValue(ctx, internal.KeyConfig, e.config)
	if e.config.Flags.ActiveNamespace != nil {
		ctx = context.WithValue(ctx, internal.KeyNamespace, *e.config.Flags.ActiveNamespace)
	} else {
		ns, err := e.factory.Client().Config().CurrentNamespaceName()
		if err != nil {
			ns = client.AllNamespaces
		}
		ctx = context.WithValue(ctx, internal.KeyNamespace, ns)
	}

	return context.WithCancel(ctx)
}
