package scrub

import (
	"context"
	"sync"

	"github.com/derailed/popeye/internal"
	"github.com/derailed/popeye/internal/cache"
	"github.com/derailed/popeye/internal/client"
	"github.com/derailed/popeye/internal/dag"
	"github.com/rs/zerolog/log"
)

type core struct {
	*dial

	mx        sync.Mutex
	namespace *cache.Namespace
	cm        *cache.ConfigMap
	pod       *cache.Pod
	node      *cache.Node
	sa        *cache.ServiceAccount
	pv        *cache.PersistentVolume
	pvc       *cache.PersistentVolumeClaim
	sec       *cache.Secret
	svc       *cache.Service
	ep        *cache.Endpoints
}

func newCore(d *dial) *core {
	return &core{dial: d}
}

func (c *core) services() (*cache.Service, error) {
	log.Debug().Msgf("CACHE-SVC")
	defer log.Debug().Msgf("  SVC-DONE")
	c.mx.Lock()
	defer c.mx.Unlock()

	if c.svc != nil {
		return c.svc, nil
	}
	ctx, cancel := c.context()
	defer cancel()
	ss, err := dag.ListServices(ctx)
	c.svc = cache.NewService(ss)

	return c.svc, err
}

func (c *core) endpoints() (*cache.Endpoints, error) {
	log.Debug().Msgf("CACHE-EP")
	defer log.Debug().Msgf("  EP-DONE")

	c.mx.Lock()
	defer c.mx.Unlock()

	if c.ep != nil {
		return c.ep, nil
	}
	ctx, cancel := c.context()
	defer cancel()
	eps, err := dag.ListEndpoints(ctx)
	c.ep = cache.NewEndpoints(eps)

	return c.ep, err
}

func (c *core) secrets() (*cache.Secret, error) {
	log.Debug().Msgf("CACHE-SEC")
	defer log.Debug().Msgf("  SEC-DONE")

	c.mx.Lock()
	defer c.mx.Unlock()

	if c.sec != nil {
		return c.sec, nil
	}
	ctx, cancel := c.context()
	defer cancel()
	secs, err := dag.ListSecrets(ctx)
	c.sec = cache.NewSecret(secs)

	return c.sec, err
}

func (c *core) persistentvolumes() (*cache.PersistentVolume, error) {
	log.Debug().Msgf("CACHE-PV")
	defer log.Debug().Msgf("  PV-DONE")

	c.mx.Lock()
	defer c.mx.Unlock()

	if c.pv != nil {
		return c.pv, nil
	}
	ctx, cancel := c.context()
	defer cancel()
	pvs, err := dag.ListPersistentVolumes(ctx)
	c.pv = cache.NewPersistentVolume(pvs)

	return c.pv, err
}

func (c *core) persistentvolumeclaims() (*cache.PersistentVolumeClaim, error) {
	log.Debug().Msgf("CACHE-PVC")
	defer log.Debug().Msgf("  PVC-DONE")

	c.mx.Lock()
	defer c.mx.Unlock()

	if c.pvc != nil {
		return c.pvc, nil
	}
	ctx, cancel := c.context()
	defer cancel()
	pvcs, err := dag.ListPersistentVolumeClaims(ctx)
	c.pvc = cache.NewPersistentVolumeClaim(pvcs)

	return c.pvc, err
}

func (c *core) configmaps() (*cache.ConfigMap, error) {
	log.Debug().Msgf("CACHE-CM")
	defer log.Debug().Msgf("  CM-DONE")
	c.mx.Lock()
	defer c.mx.Unlock()

	if c.cm != nil {
		return c.cm, nil
	}
	ctx, cancel := c.context()
	defer cancel()
	cms, err := dag.ListConfigMaps(ctx)
	c.cm = cache.NewConfigMap(cms)

	return c.cm, err
}

func (c *core) namespaces() (*cache.Namespace, error) {
	log.Debug().Msgf("CACHE-NS")
	defer log.Debug().Msgf("  NS-DONE")

	c.mx.Lock()
	defer c.mx.Unlock()

	if c.namespace != nil {
		return c.namespace, nil
	}
	ctx, cancel := c.context()
	defer cancel()
	nss, err := dag.ListNamespaces(ctx)
	c.namespace = cache.NewNamespace(nss)

	return c.namespace, err
}

func (c *core) nodes() (*cache.Node, error) {
	log.Debug().Msgf("CACHE-NO")
	defer log.Debug().Msgf("  NO-DONE")

	c.mx.Lock()
	defer c.mx.Unlock()

	if c.node != nil {
		return c.node, nil
	}
	ctx, cancel := c.context()
	defer cancel()
	nodes, err := dag.ListNodes(ctx)
	c.node = cache.NewNode(nodes)

	return c.node, err
}

func (c *core) pods() (*cache.Pod, error) {
	log.Debug().Msgf("CACHE-POD")
	defer log.Debug().Msgf("  POD-DONE")

	c.mx.Lock()
	defer c.mx.Unlock()

	if c.pod != nil {
		return c.pod, nil
	}
	ctx, cancel := c.context()
	defer cancel()
	pods, err := dag.ListPods(ctx)
	c.pod = cache.NewPod(pods)

	return c.pod, err
}

func (c *core) serviceaccounts() (*cache.ServiceAccount, error) {
	log.Debug().Msgf("CACHE-SA")
	defer log.Debug().Msgf("  SA-DONE")

	c.mx.Lock()
	defer c.mx.Unlock()

	if c.sa != nil {
		return c.sa, nil
	}
	ctx, cancel := c.context()
	defer cancel()
	sas, err := dag.ListServiceAccounts(ctx)
	c.sa = cache.NewServiceAccount(sas)

	return c.sa, err
}

// Helpers...

func (c *core) context() (context.Context, context.CancelFunc) {
	ctx := context.WithValue(context.Background(), internal.KeyFactory, c.factory)
	ctx = context.WithValue(ctx, internal.KeyConfig, c.config)
	if c.config.Flags.ActiveNamespace != nil {
		ctx = context.WithValue(ctx, internal.KeyNamespace, *c.config.Flags.ActiveNamespace)
	} else {
		ns, err := c.factory.Client().Config().CurrentNamespaceName()
		if err != nil {
			ns = client.AllNamespaces
		}
		ctx = context.WithValue(ctx, internal.KeyNamespace, ns)
	}
	return context.WithCancel(ctx)
}
