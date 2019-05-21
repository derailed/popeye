package scrub

import (
	"github.com/derailed/popeye/internal/cache"
	"github.com/derailed/popeye/internal/dag"
	"github.com/derailed/popeye/internal/k8s"
	"github.com/derailed/popeye/pkg/config"
)

// Cache caches commonly used resources.
type Cache struct {
	client *k8s.Client
	config *config.Config
	pod    *cache.Pod
	podMx  *cache.PodsMetrics
	nodeMx *cache.NodesMetrics
	dp     *cache.Deployment
	sts    *cache.StatefulSet
	sa     *cache.ServiceAccount
}

// NewCache returns a new resource cache
func NewCache(c *k8s.Client, cfg *config.Config) *Cache {
	return &Cache{client: c, config: cfg}
}

// Pods retrieves pods from cache if present or populate if not.
func (c *Cache) pods() (*cache.Pod, error) {
	if c.pod != nil {
		return c.pod, nil
	}

	pods, err := dag.ListPods(c.client, c.config)
	if err != nil {
		return nil, err
	}
	c.pod = cache.NewPod(pods)

	return c.pod, nil
}

// PodsMx retrieves pods metrics from cache if present or populate if not.
func (c *Cache) podsMx() (*cache.PodsMetrics, error) {
	if c.podMx != nil {
		return c.podMx, nil
	}

	pmx, err := dag.ListPodsMetrics(c.client)
	if err != nil {
		return nil, err
	}
	c.podMx = cache.NewPodsMetrics(pmx)

	return c.podMx, nil
}

// NodesMx retrieves nodes metrics from cache if present or populate if not.
func (c *Cache) nodesMx() (*cache.NodesMetrics, error) {
	if c.nodeMx != nil {
		return c.nodeMx, nil
	}

	nmx, err := dag.ListNodesMetrics(c.client)
	if err != nil {
		return nil, err
	}
	c.nodeMx = cache.NewNodesMetrics(nmx)

	return c.nodeMx, nil
}

// Deployments retrieves deployments from cache if present or populate if not.
func (c *Cache) deployments() (*cache.Deployment, error) {
	if c.dp != nil {
		return c.dp, nil
	}

	dps, err := dag.ListDeployments(c.client, c.config)
	if err != nil {
		return nil, err
	}
	c.dp = cache.NewDeployment(dps)

	return c.dp, nil
}

// Deployments retrieves deployments from cache if present or populate if not.
func (c *Cache) statefulsets() (*cache.StatefulSet, error) {
	if c.sts != nil {
		return c.sts, nil
	}

	sts, err := dag.ListStatefulSets(c.client, c.config)
	if err != nil {
		return nil, err
	}
	c.sts = cache.NewStatefulSet(sts)

	return c.sts, nil
}

// ServiceAccount retrieves serviceaccounts from cache if present or populate if not.
func (c *Cache) serviceaccounts() (*cache.ServiceAccount, error) {
	if c.sa != nil {
		return c.sa, nil
	}

	sas, err := dag.ListServiceAccounts(c.client, c.config)
	if err != nil {
		return nil, err
	}
	c.sa = cache.NewServiceAccount(sas)

	return c.sa, nil
}
