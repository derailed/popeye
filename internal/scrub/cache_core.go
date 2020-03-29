package scrub

import (
	"github.com/derailed/popeye/internal/cache"
	"github.com/derailed/popeye/internal/dag"
)

type core struct {
	*dial

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
	if c.svc != nil {
		return c.svc, nil
	}
	ss, err := dag.ListServices(c.factory, c.config)
	c.svc = cache.NewService(ss)

	return c.svc, err
}

func (c *core) endpoints() (*cache.Endpoints, error) {
	if c.ep != nil {
		return c.ep, nil
	}
	eps, err := dag.ListEndpoints(c.factory, c.config)
	c.ep = cache.NewEndpoints(eps)

	return c.ep, err
}

func (c *core) secrets() (*cache.Secret, error) {
	if c.sec != nil {
		return c.sec, nil
	}
	secs, err := dag.ListSecrets(c.factory, c.config)
	c.sec = cache.NewSecret(secs)

	return c.sec, err
}

func (c *core) persistentvolumes() (*cache.PersistentVolume, error) {
	if c.pv != nil {
		return c.pv, nil
	}
	pvs, err := dag.ListPersistentVolumes(c.factory, c.config)
	c.pv = cache.NewPersistentVolume(pvs)

	return c.pv, err
}

func (c *core) persistentvolumeclaims() (*cache.PersistentVolumeClaim, error) {
	if c.pvc != nil {
		return c.pvc, nil
	}
	pvcs, err := dag.ListPersistentVolumeClaims(c.factory, c.config)
	c.pvc = cache.NewPersistentVolumeClaim(pvcs)

	return c.pvc, err
}

func (c *core) configmaps() (*cache.ConfigMap, error) {
	if c.cm != nil {
		return c.cm, nil
	}
	cms, err := dag.ListConfigMaps(c.factory, c.config)
	c.cm = cache.NewConfigMap(cms)

	return c.cm, err
}

func (c *core) namespaces() (*cache.Namespace, error) {
	if c.namespace != nil {
		return c.namespace, nil
	}
	nss, err := dag.ListNamespaces(c.factory, c.config)
	c.namespace = cache.NewNamespace(nss)

	return c.namespace, err
}

func (c *core) nodes() (*cache.Node, error) {
	if c.node != nil {
		return c.node, nil
	}
	nodes, err := dag.ListNodes(c.factory, c.config)
	c.node = cache.NewNode(nodes)

	return c.node, err
}

func (c *core) pods() (*cache.Pod, error) {
	if c.pod != nil {
		return c.pod, nil
	}
	pods, err := dag.ListPods(c.factory, c.config)
	c.pod = cache.NewPod(pods)

	return c.pod, err
}

func (c *core) serviceaccounts() (*cache.ServiceAccount, error) {
	if c.sa != nil {
		return c.sa, nil
	}
	sas, err := dag.ListServiceAccounts(c.factory, c.config)
	c.sa = cache.NewServiceAccount(sas)

	return c.sa, err
}
