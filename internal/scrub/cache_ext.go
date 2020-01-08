package scrub

import (
	"github.com/derailed/popeye/internal/cache"
	"github.com/derailed/popeye/internal/dag"
)

type ext struct {
	*dial

	pdb *cache.PodDisruptionBudget
	ing *cache.Ingress
	cl  *cache.Cluster
}

func newExt(d *dial) *ext {
	return &ext{dial: d}
}

func (e *ext) cluster() (*cache.Cluster, error) {
	if e.cl != nil {
		return e.cl, nil
	}
	major, minor, err := dag.ListVersion(e.client, e.config)
	e.cl = cache.NewCluster(major, minor)

	return e.cl, err
}

func (e *ext) ingresses() (*cache.Ingress, error) {
	if e.ing != nil {
		return e.ing, nil
	}
	ings, err := dag.ListIngresses(e.client, e.config)
	e.ing = cache.NewIngress(ings)

	return e.ing, err
}

func (e *ext) podDisruptionBudgets() (*cache.PodDisruptionBudget, error) {
	if e.pdb != nil {
		return e.pdb, nil
	}
	pdbs, err := dag.ListPodDisruptionBudgets(e.client, e.config)
	e.pdb = cache.NewPodDisruptionBudget(pdbs)

	return e.pdb, err
}
