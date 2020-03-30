package scrub

import (
	"sync"

	"github.com/derailed/popeye/internal/cache"
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
	major, minor, err := dag.ListVersion(e.factory.Client(), e.config)
	e.cl = cache.NewCluster(major, minor)

	return e.cl, err
}

func (e *ext) ingresses() (*cache.Ingress, error) {
	e.mx.Lock()
	defer e.mx.Unlock()

	if e.ing != nil {
		return e.ing, nil
	}
	ings, err := dag.ListIngresses(e.factory, e.config)
	e.ing = cache.NewIngress(ings)

	return e.ing, err
}

func (e *ext) podDisruptionBudgets() (*cache.PodDisruptionBudget, error) {
	e.mx.Lock()
	defer e.mx.Unlock()

	if e.pdb != nil {
		return e.pdb, nil
	}
	pdbs, err := dag.ListPodDisruptionBudgets(e.factory, e.config)
	e.pdb = cache.NewPodDisruptionBudget(pdbs)

	return e.pdb, err
}
