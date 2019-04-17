package linter

import (
	"github.com/derailed/popeye/internal/k8s"
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
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
	allPVs  map[string]v1.PersistentVolume
	allPVCs map[string]v1.PersistentVolumeClaim
	allHPAs map[string]autoscalingv1.HorizontalPodAutoscaler
	allDPs  map[string]appsv1.Deployment
	allSTs  map[string]appsv1.StatefulSet
}

// NewFilter returns a new Kubernetes resource filter.
func NewFilter(f Fetcher, s Spinach) *Filter {
	return &Filter{Fetcher: f, Spinach: s}
}

// ListStatefulSets lists all available StatefulSets in allowed namespace.
func (f *Filter) ListStatefulSets() (map[string]appsv1.StatefulSet, error) {
	sts, err := f.ListAllStatefulSets()
	if err != nil {
		return nil, err
	}

	res := make(map[string]appsv1.StatefulSet, len(sts))
	for fqn, st := range sts {
		if f.matchActiveNS(st.Namespace) && !f.ExcludedNS(st.Namespace) {
			res[fqn] = st
		}
	}

	return res, nil
}

// ListAllStatefulSets returns all StatefulSets on cluster.
func (f *Filter) ListAllStatefulSets() (map[string]appsv1.StatefulSet, error) {
	if f.allSTs != nil {
		return f.allSTs, nil
	}

	sts, err := f.FetchStatefulSets()
	if err != nil {
		return nil, err
	}

	f.allSTs = make(map[string]appsv1.StatefulSet, len(sts.Items))
	for _, st := range sts.Items {
		f.allSTs[metaFQN(st.ObjectMeta)] = st
	}

	return f.allSTs, nil
}

// ListDeployments lists all available Deployments in allowed namespace.
func (f *Filter) ListDeployments() (map[string]appsv1.Deployment, error) {
	dps, err := f.ListAllDeployments()
	if err != nil {
		return nil, err
	}

	res := make(map[string]appsv1.Deployment, len(dps))
	for fqn, dp := range dps {
		if f.matchActiveNS(dp.Namespace) && !f.ExcludedNS(dp.Namespace) {
			res[fqn] = dp
		}
	}

	return res, nil
}

// ListAllDeployments returns all Deployments on cluster.
func (f *Filter) ListAllDeployments() (map[string]appsv1.Deployment, error) {
	if f.allDPs != nil {
		return f.allDPs, nil
	}

	dps, err := f.FetchDeployments()
	if err != nil {
		return nil, err
	}

	f.allDPs = make(map[string]appsv1.Deployment, len(dps.Items))
	for _, dp := range dps.Items {
		f.allDPs[metaFQN(dp.ObjectMeta)] = dp
	}

	return f.allDPs, nil
}

// ListHorizontalPodAutoscalers lists all available PVCs on the cluster.
func (f *Filter) ListHorizontalPodAutoscalers() (map[string]autoscalingv1.HorizontalPodAutoscaler, error) {
	hpas, err := f.ListAllHorizontalPodAutoscalers()
	if err != nil {
		return nil, err
	}

	res := make(map[string]autoscalingv1.HorizontalPodAutoscaler, len(hpas))
	for fqn, hpa := range hpas {
		if f.matchActiveNS(hpa.Namespace) && !f.ExcludedNS(hpa.Namespace) {
			res[fqn] = hpa
		}
	}

	return res, nil
}

// ListAllHorizontalPodAutoscalers returns all HorizontalPodAutoscaler.
func (f *Filter) ListAllHorizontalPodAutoscalers() (map[string]autoscalingv1.HorizontalPodAutoscaler, error) {
	if f.allHPAs != nil {
		return f.allHPAs, nil
	}

	hpas, err := f.FetchHorizontalPodAutoscalers()
	if err != nil {
		return nil, err
	}

	f.allHPAs = make(map[string]autoscalingv1.HorizontalPodAutoscaler, len(hpas.Items))
	for _, hpa := range hpas.Items {
		f.allHPAs[metaFQN(hpa.ObjectMeta)] = hpa
	}

	return f.allHPAs, nil
}

// ListPersistentVolumeClaims lists all available PVCs on the cluster.
func (f *Filter) ListPersistentVolumeClaims() (map[string]v1.PersistentVolumeClaim, error) {
	pvcs, err := f.ListAllPersistentVolumeClaims()
	if err != nil {
		return nil, err
	}

	res := make(map[string]v1.PersistentVolumeClaim, len(pvcs))
	for fqn, pvc := range pvcs {
		if f.matchActiveNS(pvc.Namespace) && !f.ExcludedNS(pvc.Namespace) {
			res[fqn] = pvc
		}
	}

	return res, nil
}

// ListAllPersistentVolumeClaims returns all PersistentVolumeClaims.
func (f *Filter) ListAllPersistentVolumeClaims() (map[string]v1.PersistentVolumeClaim, error) {
	if f.allPVCs != nil {
		return f.allPVCs, nil
	}

	pvcs, err := f.FetchPersistentVolumeClaims()
	if err != nil {
		return nil, err
	}

	f.allPVCs = make(map[string]v1.PersistentVolumeClaim, len(pvcs.Items))
	for _, pv := range pvcs.Items {
		f.allPVCs[metaFQN(pv.ObjectMeta)] = pv
	}

	return f.allPVCs, nil
}

// ListPersistentVolumes returns all PersistentVolumes.
func (f *Filter) ListPersistentVolumes() (map[string]v1.PersistentVolume, error) {
	if f.allPVs != nil {
		return f.allPVs, nil
	}

	pvs, err := f.FetchPersistentVolumes()
	if err != nil {
		return nil, err
	}

	f.allPVs = make(map[string]v1.PersistentVolume, len(pvs.Items))
	for _, pv := range pvs.Items {
		f.allPVs[metaFQN(pv.ObjectMeta)] = pv
	}

	return f.allPVs, nil
}

// ListNodesMetrics retrieves metrics for a given set of nodes.
func (*Filter) ListNodesMetrics(nodes []v1.Node, metrics []mv1beta1.NodeMetrics, nmx k8s.NodesMetrics) {
	for _, n := range nodes {
		nmx[n.Name] = k8s.NodeMetrics{
			AvailableCPU: *n.Status.Allocatable.Cpu(),
			AvailableMEM: *n.Status.Allocatable.Memory(),
			TotalCPU:     *n.Status.Capacity.Cpu(),
			TotalMEM:     *n.Status.Capacity.Memory(),
		}
	}

	for _, c := range metrics {
		if mx, ok := nmx[c.Name]; ok {
			mx.CurrentCPU = *c.Usage.Cpu()
			mx.CurrentMEM = *c.Usage.Memory()
			nmx[c.Name] = mx
		}
	}
}

// ListPodsMetrics retrieves metrics for all pods in a given namespace.
func (*Filter) ListPodsMetrics(pods []mv1beta1.PodMetrics, nmx k8s.PodsMetrics) {
	// Compute all pod's containers metrics.
	for _, p := range pods {
		mx := make(k8s.ContainerMetrics, len(p.Containers))
		for _, c := range p.Containers {
			mx[c.Name] = k8s.Metrics{
				CurrentCPU: *c.Usage.Cpu(),
				CurrentMEM: *c.Usage.Memory(),
			}
		}
		nmx[mxFQN(p)] = mx
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
		f.allRBs[metaFQN(rb.ObjectMeta)] = rb
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
		f.allEPs[metaFQN(ep.ObjectMeta)] = ep
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
		f.allSVCs[metaFQN(svc.ObjectMeta)] = svc
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

	return nil, nil
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

// ListPodsByLabels retrieves all pods matching a label selector in the allowed namespaces.
func (f *Filter) ListPodsByLabels(sel string) (map[string]v1.Pod, error) {
	pods, err := f.FetchPodsByLabels(sel)
	if err != nil {
		return nil, err
	}

	res := make(map[string]v1.Pod, len(pods.Items))
	for _, po := range pods.Items {
		if f.matchActiveNS(po.Namespace) && !f.ExcludedNS(po.Namespace) {
			res[metaFQN(po.ObjectMeta)] = po
		}
	}

	return res, nil

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
		f.allPods[metaFQN(po.ObjectMeta)] = po
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
		f.allCMs[metaFQN(cm.ObjectMeta)] = cm
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
		f.allSecs[metaFQN(sec.ObjectMeta)] = sec
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
		f.allSAs[metaFQN(sa.ObjectMeta)] = sa
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

func isSystemNS(ns string) bool {
	for _, n := range systemNS {
		if n == ns {
			return true
		}
	}
	return false
}
