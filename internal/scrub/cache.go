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
	ds     *cache.DaemonSet
	sts    *cache.StatefulSet
	sa     *cache.ServiceAccount
	rs     *cache.ReplicaSet
	pdb    *cache.PodDisruptionBudget
	ing    *cache.Ingress
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
	c.pod = cache.NewPod(pods)

	return c.pod, err
}

// PodsMx retrieves pods metrics from cache if present or populate if not.
func (c *Cache) podsMx() (*cache.PodsMetrics, error) {
	if c.podMx != nil {
		return c.podMx, nil
	}
	pmx, err := dag.ListPodsMetrics(c.client)
	c.podMx = cache.NewPodsMetrics(pmx)

	return c.podMx, err
}

// NodesMx retrieves nodes metrics from cache if present or populate if not.
func (c *Cache) nodesMx() (*cache.NodesMetrics, error) {
	if c.nodeMx != nil {
		return c.nodeMx, nil
	}
	nmx, err := dag.ListNodesMetrics(c.client)
	c.nodeMx = cache.NewNodesMetrics(nmx)

	return c.nodeMx, err
}

// Ingresses retrieves ingress from cache if present or populate if not.
func (c *Cache) ingresses() (*cache.Ingress, error) {
	if c.ing != nil {
		return c.ing, nil
	}
	ings, err := dag.ListIngresses(c.client, c.config)
	c.ing = cache.NewIngress(ings)

	return c.ing, err
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
	rev, err := dag.DeploymentPreferredRev(c.client)
	if err != nil {
		return nil, err
	}
	c.dp = cache.NewDeployment(dps, rev)

	return c.dp, err
}

// ReplicaSets retrieves rs from cache if present or populate if not.
func (c *Cache) replicasets() (*cache.ReplicaSet, error) {
	if c.rs != nil {
		return c.rs, nil
	}
	rss, err := dag.ListReplicaSets(c.client, c.config)
	c.rs = cache.NewReplicaSet(rss)

	return c.rs, err
}

// DaemonSet retrieves ds from cache if present or populate if not.
func (c *Cache) daemonSets() (*cache.DaemonSet, error) {
	if c.ds != nil {
		return c.ds, nil
	}
	ds, err := dag.ListDaemonSets(c.client, c.config)
	c.ds = cache.NewDaemonSet(ds)

	return c.ds, err
}

// StatefulSets retrieves sts from cache if present or populate if not.
func (c *Cache) statefulsets() (*cache.StatefulSet, error) {
	if c.sts != nil {
		return c.sts, nil
	}

	sts, err := dag.ListStatefulSets(c.client, c.config)
	c.sts = cache.NewStatefulSet(sts)

	return c.sts, err
}

// ServiceAccount retrieves serviceaccounts from cache if present or populate if not.
func (c *Cache) serviceaccounts() (*cache.ServiceAccount, error) {
	if c.sa != nil {
		return c.sa, nil
	}
	sas, err := dag.ListServiceAccounts(c.client, c.config)
	c.sa = cache.NewServiceAccount(sas)

	return c.sa, err
}

// PodDisruptionBudgets retrieves podDisruptionBudgets from cache if present or populate if not.
func (c *Cache) podDisruptionBudgets() (*cache.PodDisruptionBudget, error) {
	if c.pdb != nil {
		return c.pdb, nil
	}
	pdbs, err := dag.ListPodDisruptionBudgets(c.client, c.config)
	c.pdb = cache.NewPodDisruptionBudget(pdbs)

	return c.pdb, err
}
