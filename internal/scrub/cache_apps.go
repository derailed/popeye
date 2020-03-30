package scrub

import (
	"sync"

	"github.com/derailed/popeye/internal/cache"
	"github.com/derailed/popeye/internal/dag"
)

type apps struct {
	*dial

	mx  sync.Mutex
	dp  *cache.Deployment
	ds  *cache.DaemonSet
	sts *cache.StatefulSet
	rs  *cache.ReplicaSet
}

func newApps(d *dial) *apps {
	return &apps{dial: d}
}

func (a *apps) deployments() (*cache.Deployment, error) {
	a.mx.Lock()
	defer a.mx.Unlock()

	if a.dp != nil {
		return a.dp, nil
	}
	dps, err := dag.ListDeployments(a.factory, a.config)
	a.dp = cache.NewDeployment(dps)

	return a.dp, err
}

func (a *apps) replicasets() (*cache.ReplicaSet, error) {
	a.mx.Lock()
	defer a.mx.Unlock()

	if a.rs != nil {
		return a.rs, nil
	}
	rss, err := dag.ListReplicaSets(a.factory, a.config)
	a.rs = cache.NewReplicaSet(rss)

	return a.rs, err
}

func (a *apps) daemonSets() (*cache.DaemonSet, error) {
	a.mx.Lock()
	defer a.mx.Unlock()

	if a.ds != nil {
		return a.ds, nil
	}
	ds, err := dag.ListDaemonSets(a.factory, a.config)
	a.ds = cache.NewDaemonSet(ds)

	return a.ds, err
}

func (a *apps) statefulsets() (*cache.StatefulSet, error) {
	a.mx.Lock()
	defer a.mx.Unlock()

	if a.sts != nil {
		return a.sts, nil
	}

	sts, err := dag.ListStatefulSets(a.factory, a.config)
	a.sts = cache.NewStatefulSet(sts)

	return a.sts, err
}
