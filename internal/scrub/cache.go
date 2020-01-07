package scrub

import (
	"github.com/derailed/popeye/internal/cache"
	"github.com/derailed/popeye/internal/dag"
	"github.com/derailed/popeye/internal/k8s"
	"github.com/derailed/popeye/pkg/config"
)

// Cache tracks commonly used resources.
type Cache struct {
	client *k8s.Client
	config *config.Config

	namespace *cache.Namespace
	cm        *cache.ConfigMap
	pod       *cache.Pod
	node      *cache.Node
	nodeMx    *cache.NodesMetrics
	podMx     *cache.PodsMetrics
	sa        *cache.ServiceAccount
	cl        *cache.Cluster
	dp        *cache.Deployment
	ds        *cache.DaemonSet
	sts       *cache.StatefulSet
	rs        *cache.ReplicaSet
	pdb       *cache.PodDisruptionBudget
	ing       *cache.Ingress
	np        *cache.NetworkPolicy
	psp       *cache.PodSecurityPolicy
	pv        *cache.PersistentVolume
	pvc       *cache.PersistentVolumeClaim
	crb       *cache.ClusterRoleBinding
	cr        *cache.ClusterRole
	rb        *cache.RoleBinding
	ro        *cache.Role
	sec       *cache.Secret
	svc       *cache.Service
	ep        *cache.Endpoints
}

// NewCache returns a new resource cache
func NewCache(c *k8s.Client, cfg *config.Config) *Cache {
	return &Cache{client: c, config: cfg}
}

func (c *Cache) cluster() (*cache.Cluster, error) {
	if c.cl != nil {
		return c.cl, nil
	}
	major, minor, err := dag.ListVersion(c.client, c.config)
	c.cl = cache.NewCluster(major, minor)

	return c.cl, err
}

func (c *Cache) services() (*cache.Service, error) {
	if c.svc != nil {
		return c.svc, nil
	}
	ss, err := dag.ListServices(c.client, c.config)
	c.svc = cache.NewService(ss)

	return c.svc, err
}

func (c *Cache) endpoints() (*cache.Endpoints, error) {
	if c.ep != nil {
		return c.ep, nil
	}
	eps, err := dag.ListEndpoints(c.client, c.config)
	c.ep = cache.NewEndpoints(eps)

	return c.ep, err
}

func (c *Cache) secrets() (*cache.Secret, error) {
	if c.sec != nil {
		return c.sec, nil
	}
	secs, err := dag.ListSecrets(c.client, c.config)
	c.sec = cache.NewSecret(secs)

	return c.sec, err
}

func (c *Cache) roles() (*cache.Role, error) {
	if c.ro != nil {
		return c.ro, nil
	}
	ros, err := dag.ListRoles(c.client, c.config)
	c.ro = cache.NewRole(ros)

	return c.ro, err
}

func (c *Cache) rolebindings() (*cache.RoleBinding, error) {
	if c.rb != nil {
		return c.rb, nil
	}
	rbs, err := dag.ListRoleBindings(c.client, c.config)
	c.rb = cache.NewRoleBinding(rbs)

	return c.rb, err
}

func (c *Cache) clusterroles() (*cache.ClusterRole, error) {
	if c.cr != nil {
		return c.cr, nil
	}
	crs, err := dag.ListClusterRoles(c.client, c.config)
	c.cr = cache.NewClusterRole(crs)

	return c.cr, err
}

func (c *Cache) clusterrolebindings() (*cache.ClusterRoleBinding, error) {
	if c.crb != nil {
		return c.crb, nil
	}
	crbs, err := dag.ListClusterRoleBindings(c.client, c.config)
	c.crb = cache.NewClusterRoleBinding(crbs)

	return c.crb, err
}

func (c *Cache) persistentvolumes() (*cache.PersistentVolume, error) {
	if c.pv != nil {
		return c.pv, nil
	}
	pvs, err := dag.ListPersistentVolumes(c.client, c.config)
	c.pv = cache.NewPersistentVolume(pvs)

	return c.pv, err
}

func (c *Cache) persistentvolumeclaims() (*cache.PersistentVolumeClaim, error) {
	if c.pvc != nil {
		return c.pvc, nil
	}
	pvcs, err := dag.ListPersistentVolumeClaims(c.client, c.config)
	c.pvc = cache.NewPersistentVolumeClaim(pvcs)

	return c.pvc, err
}

func (c *Cache) configmaps() (*cache.ConfigMap, error) {
	if c.cm != nil {
		return c.cm, nil
	}
	cms, err := dag.ListConfigMaps(c.client, c.config)
	c.cm = cache.NewConfigMap(cms)

	return c.cm, err
}

func (c *Cache) podsecuritypolicies() (*cache.PodSecurityPolicy, error) {
	if c.psp != nil {
		return c.psp, nil
	}
	psps, err := dag.ListPodSecurityPolicies(c.client, c.config)
	c.psp = cache.NewPodSecurityPolicy(psps)

	return c.psp, err
}

func (c *Cache) networkpolicies() (*cache.NetworkPolicy, error) {
	if c.np != nil {
		return c.np, nil
	}
	nps, err := dag.ListNetworkPolicies(c.client, c.config)
	c.np = cache.NewNetworkPolicy(nps)

	return c.np, err
}

func (c *Cache) namespaces() (*cache.Namespace, error) {
	if c.namespace != nil {
		return c.namespace, nil
	}
	nss, err := dag.ListNamespaces(c.client, c.config)
	c.namespace = cache.NewNamespace(nss)

	return c.namespace, err
}

func (c *Cache) nodes() (*cache.Node, error) {
	if c.node != nil {
		return c.node, nil
	}
	nodes, err := dag.ListNodes(c.client, c.config)
	c.node = cache.NewNode(nodes)

	return c.node, err
}

func (c *Cache) pods() (*cache.Pod, error) {
	if c.pod != nil {
		return c.pod, nil
	}
	pods, err := dag.ListPods(c.client, c.config)
	c.pod = cache.NewPod(pods)

	return c.pod, err
}

func (c *Cache) podsMx() (*cache.PodsMetrics, error) {
	if c.podMx != nil {
		return c.podMx, nil
	}
	pmx, err := dag.ListPodsMetrics(c.client)
	c.podMx = cache.NewPodsMetrics(pmx)

	return c.podMx, err
}

func (c *Cache) nodesMx() (*cache.NodesMetrics, error) {
	if c.nodeMx != nil {
		return c.nodeMx, nil
	}
	nmx, err := dag.ListNodesMetrics(c.client)
	c.nodeMx = cache.NewNodesMetrics(nmx)

	return c.nodeMx, err
}

func (c *Cache) ingresses() (*cache.Ingress, error) {
	if c.ing != nil {
		return c.ing, nil
	}
	ings, err := dag.ListIngresses(c.client, c.config)
	c.ing = cache.NewIngress(ings)

	return c.ing, err
}

func (c *Cache) deployments() (*cache.Deployment, error) {
	if c.dp != nil {
		return c.dp, nil
	}
	dps, err := dag.ListDeployments(c.client, c.config)
	c.dp = cache.NewDeployment(dps)

	return c.dp, err
}

func (c *Cache) replicasets() (*cache.ReplicaSet, error) {
	if c.rs != nil {
		return c.rs, nil
	}
	rss, err := dag.ListReplicaSets(c.client, c.config)
	c.rs = cache.NewReplicaSet(rss)

	return c.rs, err
}

func (c *Cache) daemonSets() (*cache.DaemonSet, error) {
	if c.ds != nil {
		return c.ds, nil
	}
	ds, err := dag.ListDaemonSets(c.client, c.config)
	c.ds = cache.NewDaemonSet(ds)

	return c.ds, err
}

func (c *Cache) statefulsets() (*cache.StatefulSet, error) {
	if c.sts != nil {
		return c.sts, nil
	}

	sts, err := dag.ListStatefulSets(c.client, c.config)
	c.sts = cache.NewStatefulSet(sts)

	return c.sts, err
}

func (c *Cache) serviceaccounts() (*cache.ServiceAccount, error) {
	if c.sa != nil {
		return c.sa, nil
	}
	sas, err := dag.ListServiceAccounts(c.client, c.config)
	c.sa = cache.NewServiceAccount(sas)

	return c.sa, err
}

func (c *Cache) podDisruptionBudgets() (*cache.PodDisruptionBudget, error) {
	if c.pdb != nil {
		return c.pdb, nil
	}
	pdbs, err := dag.ListPodDisruptionBudgets(c.client, c.config)
	c.pdb = cache.NewPodDisruptionBudget(pdbs)

	return c.pdb, err
}
