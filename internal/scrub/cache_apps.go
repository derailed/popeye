package scrub

import (
	"context"
	"sync"

	"github.com/derailed/popeye/internal"
	"github.com/derailed/popeye/internal/cache"
	"github.com/derailed/popeye/internal/client"
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
	ctx, cancel := a.context()
	defer cancel()
	dps, err := dag.ListDeployments(ctx)
	a.dp = cache.NewDeployment(dps)

	return a.dp, err
}

func (a *apps) replicasets() (*cache.ReplicaSet, error) {
	a.mx.Lock()
	defer a.mx.Unlock()

	if a.rs != nil {
		return a.rs, nil
	}
	ctx, cancel := a.context()
	defer cancel()
	rss, err := dag.ListReplicaSets(ctx)
	a.rs = cache.NewReplicaSet(rss)

	return a.rs, err
}

func (a *apps) daemonSets() (*cache.DaemonSet, error) {
	a.mx.Lock()
	defer a.mx.Unlock()

	if a.ds != nil {
		return a.ds, nil
	}
	ctx, cancel := a.context()
	defer cancel()
	ds, err := dag.ListDaemonSets(ctx)
	a.ds = cache.NewDaemonSet(ds)

	return a.ds, err
}

func (a *apps) statefulsets() (*cache.StatefulSet, error) {
	a.mx.Lock()
	defer a.mx.Unlock()

	if a.sts != nil {
		return a.sts, nil
	}

	ctx, cancel := a.context()
	defer cancel()
	sts, err := dag.ListStatefulSets(ctx)
	a.sts = cache.NewStatefulSet(sts)

	return a.sts, err
}

// Helpers...

func (a *apps) context() (context.Context, context.CancelFunc) {
	ctx := context.WithValue(context.Background(), internal.KeyFactory, a.factory)
	ctx = context.WithValue(ctx, internal.KeyConfig, a.config)
	if a.config.Flags.ActiveNamespace != nil {
		ctx = context.WithValue(ctx, internal.KeyNamespace, *a.config.Flags.ActiveNamespace)
	} else {
		ns, err := a.factory.Client().Config().CurrentNamespaceName()
		if err != nil {
			ns = client.AllNamespaces
		}
		ctx = context.WithValue(ctx, internal.KeyNamespace, ns)
	}

	return context.WithCancel(ctx)
}
