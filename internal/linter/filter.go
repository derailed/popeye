package linter

import (
	"fmt"

	v1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	mv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

var systemNS = []string{"kube-system", "kube-public"}

// Filter represents a Kubernetes resources filter based on configuration.
type Filter struct {
	Spinach
	Fetcher

	allPods map[string]v1.Pod
	allNSs  map[string]v1.Namespace
	allEPs  map[string]v1.Endpoints
	allCRBs map[string]rbacv1.ClusterRoleBinding
	allRBs  map[string]rbacv1.RoleBinding
	allCMs  map[string]v1.ConfigMap
	allSecs map[string]v1.Secret
	allSAs  map[string]v1.ServiceAccount
	allSVCs map[string]v1.Service
}

// NewFilter returns a new Kubernetes resource filter.
func NewFilter(f Fetcher, s Spinach) *Filter {
	return &Filter{Fetcher: f, Spinach: s}
}

// ListNodesMetrics retrieves metrics for a given set of nodes.
func (*Filter) ListNodesMetrics(nodes []v1.Node, metrics []mv1beta1.NodeMetrics, mmx NodesMetrics) {
	for _, n := range nodes {
		mmx[n.Name] = NodeMetrics{
			AvailCPU: n.Status.Allocatable.Cpu().MilliValue(),
			AvailMEM: asMi(n.Status.Allocatable.Memory().Value()),
			TotalCPU: n.Status.Capacity.Cpu().MilliValue(),
			TotalMEM: asMi(n.Status.Capacity.Memory().Value()),
		}
	}

	for _, c := range metrics {
		if mx, ok := mmx[c.Name]; ok {
			mx.CurrentCPU = c.Usage.Cpu().MilliValue()
			mx.CurrentMEM = asMi(c.Usage.Memory().Value())
			mmx[c.Name] = mx
		}
	}
}

// ListPodsMetrics retrieves metrics for all pods in a given namespace.
func (*Filter) ListPodsMetrics(pods []mv1beta1.PodMetrics, mmx PodsMetrics) {
	// Compute all pod's containers metrics.
	for _, p := range pods {
		mx := make(ContainerMetrics, len(p.Containers))
		for _, c := range p.Containers {
			mx[c.Name] = Metrics{
				CurrentCPU: c.Usage.Cpu().MilliValue(),
				CurrentMEM: asMi(c.Usage.Memory().Value()),
			}
		}
		mmx[mxFQN(p)] = mx
	}
}

// ListRoleBindings lists all available RBs in the allowed namespaces.
func (f *Filter) ListRoleBindings() (map[string]rbacv1.RoleBinding, error) {
	rbs, err := f.ListAllRoleBindings()
	if err != nil {
		return nil, err
	}

	res := make(map[string]rbacv1.RoleBinding, len(rbs))
	for fqn, rb := range rbs {
		if f.matchActiveNS(rb.Namespace) && !f.ExcludedNS(rb.Namespace) {
			res[fqn] = rb
		}
	}

	return res, nil
}

// ListAllRoleBindings returns all RoleBindings.
func (f *Filter) ListAllRoleBindings() (map[string]rbacv1.RoleBinding, error) {
	if f.allRBs != nil {
		return f.allRBs, nil
	}

	crbs, err := f.FetchRoleBindings()
	if err != nil {
		return nil, err
	}

	f.allRBs = make(map[string]rbacv1.RoleBinding, len(crbs.Items))
	for _, rb := range crbs.Items {
		f.allRBs[fqn(rb.Namespace, rb.Name)] = rb
	}

	return f.allRBs, nil
}

// ListAllClusterRoleBindings returns a ClusterRoleBindings.
func (f *Filter) ListAllClusterRoleBindings() (map[string]rbacv1.ClusterRoleBinding, error) {
	if f.allCRBs != nil {
		return f.allCRBs, nil
	}

	ll, err := f.FetchClusterRoleBindings()
	if err != nil {
		return nil, err
	}

	f.allCRBs = make(map[string]rbacv1.ClusterRoleBinding, len(ll.Items))
	for _, crb := range ll.Items {
		f.allCRBs[crb.Name] = crb
	}

	return f.allCRBs, nil
}

// ListAllEndpoints returns all the  endpoints on a cluster.
func (f *Filter) ListAllEndpoints() (map[string]v1.Endpoints, error) {
	if f.allEPs != nil {
		return f.allEPs, nil
	}

	ll, err := f.FetchEndpoints()
	if err != nil {
		return nil, err
	}

	f.allEPs = make(map[string]v1.Endpoints, len(ll.Items))
	for _, ep := range ll.Items {
		f.allEPs[fqn(ep.Namespace, ep.Name)] = ep
	}

	return f.allEPs, nil
}

// ListEndpoints returns a collection of endpoints in allowed namespaces.
func (f *Filter) ListEndpoints() (map[string]v1.Endpoints, error) {
	eps, err := f.ListAllEndpoints()
	if err != nil {
		return nil, err
	}

	res := make(map[string]v1.Endpoints, len(eps))
	for fqn, ep := range eps {
		if !f.ExcludedNS(ep.Namespace) {
			res[fqn] = ep
		}
	}

	return res, nil
}

// GetEndpoints returns a endpoint instance if present or an error if not.
func (f *Filter) GetEndpoints(svcFQN string) (*v1.Endpoints, error) {
	eps, err := f.ListEndpoints()
	if err != nil {
		return nil, err
	}

	if ep, ok := eps[svcFQN]; ok {
		return &ep, nil
	}

	return nil, nil
}

// ListServices lists services in tolerated namespaces.
func (f *Filter) ListServices() (map[string]v1.Service, error) {
	svcs, err := f.ListAllServices()
	if err != nil {
		return nil, err
	}

	res := make(map[string]v1.Service, len(svcs))
	for fqn, svc := range svcs {
		if f.matchActiveNS(svc.Namespace) && !f.ExcludedNS(svc.Namespace) {
			res[fqn] = svc
		}
	}

	return res, nil
}

// ListAllServices lists services in the cluster.
func (f *Filter) ListAllServices() (map[string]v1.Service, error) {
	if f.allSVCs != nil {
		return f.allSVCs, nil
	}

	svcs, err := f.FetchServices()
	if err != nil {
		return nil, err
	}

	f.allSVCs = make(map[string]v1.Service, len(svcs.Items))
	for _, svc := range svcs.Items {
		f.allSVCs[fqn(svc.Namespace, svc.Name)] = svc
	}

	return f.allSVCs, nil
}

// ListNodes list all available nodes on the cluster.
func (f *Filter) ListNodes() ([]v1.Node, error) {
	ll, err := f.FetchNodes()
	if err != nil {
		return nil, err
	}

	nodes := make([]v1.Node, 0, len(ll.Items))
	for _, no := range ll.Items {
		if !f.ExcludedNode(no.Name) {
			nodes = append(nodes, no)
		}
	}

	return nodes, nil
}

// GetPod returns a pod via a label query.
func (f *Filter) GetPod(sel map[string]string) (*v1.Pod, error) {
	pods, err := f.ListPods()
	if err != nil {
		return nil, err
	}

	for _, po := range pods {
		var found int
		for k, v := range sel {
			if pv, ok := po.Labels[k]; ok && pv == v {
				found++
			}
		}
		if found == len(sel) {
			return &po, nil
		}
	}

	return nil, fmt.Errorf("No pods match service selector")
}

// PodsNamespaces fetch a list of all namespaces used by pods.
func (f *Filter) PodsNamespaces(nss []string) {
	pods, err := f.ListPods()
	if err != nil {
		return
	}

	set := make(map[string]struct{})
	for _, p := range pods {
		set[p.Namespace] = struct{}{}
	}

	var i int
	for k := range set {
		nss[i] = k
		i++
	}
}

// ListPods list all available pods.
func (f *Filter) ListPods() (map[string]v1.Pod, error) {
	pods, err := f.ListAllPods()
	if err != nil {
		return nil, err
	}

	res := make(map[string]v1.Pod, len(pods))
	for fqn, po := range pods {
		if f.matchActiveNS(po.Namespace) && !f.ExcludedNS(po.Namespace) {
			res[fqn] = po
		}
	}

	return res, nil
}

// ListAllPods fetch all pods on the cluster.
func (f *Filter) ListAllPods() (map[string]v1.Pod, error) {
	if len(f.allPods) != 0 {
		return f.allPods, nil
	}

	ll, err := f.FetchPods()
	if err != nil {
		return nil, err
	}

	f.allPods = make(map[string]v1.Pod, len(ll.Items))
	for _, po := range ll.Items {
		f.allPods[podFQN(po)] = po
	}

	return f.allPods, nil
}

// ListConfigMaps list all included ConfigMaps.
func (f *Filter) ListConfigMaps() (map[string]v1.ConfigMap, error) {
	cms, err := f.ListAllConfigMaps()
	if err != nil {
		return nil, err
	}

	res := make(map[string]v1.ConfigMap, len(cms))
	for fqn, cm := range cms {
		if f.matchActiveNS(cm.Namespace) && !f.ExcludedNS(cm.Namespace) {
			res[fqn] = cm
		}
	}

	return res, nil
}

// ListAllConfigMaps fetch all configmaps on the cluster.
func (f *Filter) ListAllConfigMaps() (map[string]v1.ConfigMap, error) {
	if len(f.allCMs) != 0 {
		return f.allCMs, nil
	}

	ll, err := f.FetchConfigMaps()
	if err != nil {
		return nil, err
	}

	f.allCMs = make(map[string]v1.ConfigMap, len(ll.Items))
	for _, cm := range ll.Items {
		f.allCMs[fqn(cm.Namespace, cm.Name)] = cm
	}

	return f.allCMs, nil
}

// ListSecrets list included Secrets.
func (f *Filter) ListSecrets() (map[string]v1.Secret, error) {
	secs, err := f.ListAllSecrets()
	if err != nil {
		return nil, err
	}

	res := make(map[string]v1.Secret, len(secs))
	for fqn, sec := range secs {
		if f.matchActiveNS(sec.Namespace) && !f.ExcludedNS(sec.Namespace) {
			res[fqn] = sec
		}
	}

	return res, nil
}

// ListAllSecrets fetch all secrets on the cluster.
func (f *Filter) ListAllSecrets() (map[string]v1.Secret, error) {
	if len(f.allSecs) != 0 {
		return f.allSecs, nil
	}

	ll, err := f.FetchSecrets()
	if err != nil {
		return nil, err
	}

	f.allSecs = make(map[string]v1.Secret, len(ll.Items))
	for _, sec := range ll.Items {
		f.allSecs[fqn(sec.Namespace, sec.Name)] = sec
	}

	return f.allSecs, nil
}

// ListServiceAccounts list included ServiceAccounts.
func (f *Filter) ListServiceAccounts() (map[string]v1.ServiceAccount, error) {
	sas, err := f.ListAllServiceAccounts()
	if err != nil {
		return nil, err
	}

	res := make(map[string]v1.ServiceAccount, len(sas))
	for fqn, sa := range sas {
		if f.matchActiveNS(sa.Namespace) && !f.ExcludedNS(sa.Namespace) {
			res[fqn] = sa
		}
	}

	return res, nil
}

// ListAllServiceAccounts fetch all ServiceAccount on the cluster.
func (f *Filter) ListAllServiceAccounts() (map[string]v1.ServiceAccount, error) {
	if len(f.allSAs) != 0 {
		return f.allSAs, nil
	}

	ll, err := f.FetchServiceAccounts()
	if err != nil {
		return nil, err
	}

	f.allSAs = make(map[string]v1.ServiceAccount, len(ll.Items))
	for _, sa := range ll.Items {
		f.allSAs[fqn(sa.Namespace, sa.Name)] = sa
	}

	return f.allSAs, nil
}

// ListNamespaces lists all available namespaces.
func (f *Filter) ListNamespaces() (map[string]v1.Namespace, error) {
	nss, err := f.ListAllNamespaces()
	if err != nil {
		return nil, nil
	}

	res := make(map[string]v1.Namespace, len(nss))
	for n, ns := range nss {
		if f.matchActiveNS(n) && !f.ExcludedNS(n) {
			res[n] = ns
		}
	}

	return res, nil
}

// ListAllNamespaces fetch all namespaces on this cluster.
func (f *Filter) ListAllNamespaces() (map[string]v1.Namespace, error) {
	if len(f.allNSs) != 0 {
		return f.allNSs, nil
	}

	nn, err := f.FetchNamespaces()
	if err != nil {
		return nil, err
	}

	f.allNSs = make(map[string]v1.Namespace, len(nn.Items))
	for _, ns := range nn.Items {
		f.allNSs[ns.Name] = ns
	}

	return f.allNSs, nil
}

func (f *Filter) matchActiveNS(ns string) bool {
	if f.ActiveNamespace() == "" {
		return true
	}

	return ns == f.ActiveNamespace()
}

// ----------------------------------------------------------------------------
// Helpers...

func mxFQN(mx mv1beta1.PodMetrics) string {
	return mx.Namespace + "/" + mx.Name
}

func fqn(ns, n string) string {
	return ns + "/" + n
}

func isSystemNS(ns string) bool {
	for _, n := range systemNS {
		if n == ns {
			return true
		}
	}
	return false
}
