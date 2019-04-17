package linter

import (
	"testing"

	"github.com/derailed/popeye/internal/k8s"
	m "github.com/petergtz/pegomock"
	pegomock "github.com/petergtz/pegomock"
	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	v1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

func TestIsSystemNS(t *testing.T) {
	assert.True(t, isSystemNS("kube-system"))
	assert.False(t, isSystemNS("fred"))
}

func TestListPersistentVolumeClaims(t *testing.T) {
	mkf := NewMockFetcher()
	m.When(mkf.FetchPersistentVolumeClaims()).ThenReturn(&v1.PersistentVolumeClaimList{
		Items: []v1.PersistentVolumeClaim{
			makePVC("p1", v1.ClaimBound),
			makePVC("p2", v1.ClaimBound),
		},
	}, nil)
	m.When(mkf.ActiveNamespace()).ThenReturn("")

	mkk := NewMockSpinach()
	m.When(mkk.ExcludedNS("default")).ThenReturn(false)

	po, err := NewFilter(mkf, mkk).ListPersistentVolumeClaims()

	assert.Nil(t, err)
	assert.NotNil(t, po)
	mkf.VerifyWasCalledOnce().FetchPersistentVolumeClaims()
	mkf.VerifyWasCalled(pegomock.Times(2)).ActiveNamespace()
	mkk.VerifyWasCalled(pegomock.Times(2)).ExcludedNS("default")
}

func TestListAllPersistentVolumeClaims(t *testing.T) {
	mkf := NewMockFetcher()
	m.When(mkf.FetchPersistentVolumeClaims()).ThenReturn(&v1.PersistentVolumeClaimList{
		Items: []v1.PersistentVolumeClaim{
			makePVC("p1", v1.ClaimBound),
			makePVC("p2", v1.ClaimBound),
		},
	}, nil)

	mkk := NewMockSpinach()
	ll, err := NewFilter(mkf, mkk).ListAllPersistentVolumeClaims()

	assert.Nil(t, err)
	assert.NotNil(t, ll)
	mkf.VerifyWasCalledOnce().FetchPersistentVolumeClaims()
}

func TestListPersistentVolumes(t *testing.T) {
	mkf := NewMockFetcher()
	m.When(mkf.FetchPersistentVolumes()).ThenReturn(&v1.PersistentVolumeList{
		Items: []v1.PersistentVolume{
			makePV("p1", v1.VolumeBound),
			makePV("p2", v1.VolumeBound),
		},
	}, nil)

	mkk := NewMockSpinach()
	ll, err := NewFilter(mkf, mkk).ListPersistentVolumes()

	assert.Nil(t, err)
	assert.NotNil(t, ll)
	mkf.VerifyWasCalledOnce().FetchPersistentVolumes()
}

func TestListHorizontalPodAutoscalers(t *testing.T) {
	mkf := NewMockFetcher()
	m.When(mkf.FetchHorizontalPodAutoscalers()).ThenReturn(&autoscalingv1.HorizontalPodAutoscalerList{
		Items: []autoscalingv1.HorizontalPodAutoscaler{
			makeHPA("h1", "Deployment", "d1", 1),
			makeHPA("h2", "Deployment", "d2", 1),
		},
	}, nil)
	m.When(mkf.ActiveNamespace()).ThenReturn("")

	mkk := NewMockSpinach()
	m.When(mkk.ExcludedNS("default")).ThenReturn(false)

	po, err := NewFilter(mkf, mkk).ListHorizontalPodAutoscalers()

	assert.Nil(t, err)
	assert.NotNil(t, po)
	mkf.VerifyWasCalledOnce().FetchHorizontalPodAutoscalers()
	mkf.VerifyWasCalled(pegomock.Times(2)).ActiveNamespace()
	mkk.VerifyWasCalled(pegomock.Times(2)).ExcludedNS("default")
}

func TestListAllHorizontalPodAutoscalers(t *testing.T) {
	mkf := NewMockFetcher()
	m.When(mkf.FetchHorizontalPodAutoscalers()).ThenReturn(&autoscalingv1.HorizontalPodAutoscalerList{
		Items: []autoscalingv1.HorizontalPodAutoscaler{
			makeHPA("h1", "Deployment", "d1", 1),
			makeHPA("h2", "Deployment", "d2", 1),
		},
	}, nil)

	mkk := NewMockSpinach()
	ll, err := NewFilter(mkf, mkk).ListAllHorizontalPodAutoscalers()

	assert.Nil(t, err)
	assert.NotNil(t, ll)
	mkf.VerifyWasCalledOnce().FetchHorizontalPodAutoscalers()
}

func TestListDeployments(t *testing.T) {
	mkf := NewMockFetcher()
	m.When(mkf.FetchDeployments()).ThenReturn(&appsv1.DeploymentList{
		Items: []appsv1.Deployment{
			makeDP("s1", "100m", "1Mi"),
			makeDP("s2", "100m", "1Mi"),
		},
	}, nil)
	m.When(mkf.ActiveNamespace()).ThenReturn("")

	mkk := NewMockSpinach()
	m.When(mkk.ExcludedNS("default")).ThenReturn(false)

	po, err := NewFilter(mkf, mkk).ListDeployments()

	assert.Nil(t, err)
	assert.NotNil(t, po)
	mkf.VerifyWasCalledOnce().FetchDeployments()
	mkf.VerifyWasCalled(pegomock.Times(2)).ActiveNamespace()
	mkk.VerifyWasCalled(pegomock.Times(2)).ExcludedNS("default")
}

func TestListAllDeployments(t *testing.T) {
	mkf := NewMockFetcher()
	m.When(mkf.FetchDeployments()).ThenReturn(&appsv1.DeploymentList{
		Items: []appsv1.Deployment{
			makeDP("s1", "1m", "1Mi"),
			makeDP("s2", "1m", "1Mi"),
		},
	}, nil)

	mkk := NewMockSpinach()
	ll, err := NewFilter(mkf, mkk).ListAllDeployments()

	assert.Nil(t, err)
	assert.NotNil(t, ll)
	mkf.VerifyWasCalledOnce().FetchDeployments()
}

func TestListStatefulSets(t *testing.T) {
	mkf := NewMockFetcher()
	m.When(mkf.FetchStatefulSets()).ThenReturn(&appsv1.StatefulSetList{
		Items: []appsv1.StatefulSet{
			makeSTS("s1", "100m", "1Mi"),
			makeSTS("s2", "100m", "1Mi"),
		},
	}, nil)
	m.When(mkf.ActiveNamespace()).ThenReturn("")

	mkk := NewMockSpinach()
	m.When(mkk.ExcludedNS("default")).ThenReturn(false)

	po, err := NewFilter(mkf, mkk).ListStatefulSets()

	assert.Nil(t, err)
	assert.NotNil(t, po)
	mkf.VerifyWasCalledOnce().FetchStatefulSets()
	mkf.VerifyWasCalled(pegomock.Times(2)).ActiveNamespace()
	mkk.VerifyWasCalled(pegomock.Times(2)).ExcludedNS("default")
}

func TestListAllStatefulSets(t *testing.T) {
	mkf := NewMockFetcher()
	m.When(mkf.FetchStatefulSets()).ThenReturn(&appsv1.StatefulSetList{
		Items: []appsv1.StatefulSet{
			makeSTS("s1", "1m", "1Mi"),
			makeSTS("s2", "1m", "1Mi"),
		},
	}, nil)

	mkk := NewMockSpinach()
	ll, err := NewFilter(mkf, mkk).ListAllStatefulSets()

	assert.Nil(t, err)
	assert.NotNil(t, ll)
	mkf.VerifyWasCalledOnce().FetchStatefulSets()
}

func TestListNamespaces(t *testing.T) {
	mkf := NewMockFetcher()
	m.When(mkf.FetchNamespaces()).ThenReturn(&v1.NamespaceList{
		Items: []v1.Namespace{
			makeNS("n1", true),
			makeNS("n2", true),
		},
	}, nil)
	m.When(mkf.ActiveNamespace()).ThenReturn("")

	mkk := NewMockSpinach()
	m.When(mkk.ExcludedNS("n1")).ThenReturn(false)
	m.When(mkk.ExcludedNS("n2")).ThenReturn(false)

	ll, err := NewFilter(mkf, mkk).ListNamespaces()

	assert.Nil(t, err)
	assert.NotNil(t, ll)
	mkf.VerifyWasCalledOnce().FetchNamespaces()
	mkf.VerifyWasCalled(pegomock.Times(2)).ActiveNamespace()
	mkk.VerifyWasCalled(pegomock.Times(1)).ExcludedNS("n1")
	mkk.VerifyWasCalled(pegomock.Times(1)).ExcludedNS("n2")
}

func TestListAllNamespaces(t *testing.T) {
	mkf := NewMockFetcher()
	m.When(mkf.FetchNamespaces()).ThenReturn(&v1.NamespaceList{
		Items: []v1.Namespace{
			makeNS("n1", true),
			makeNS("n2", true),
		},
	}, nil)

	mkk := NewMockSpinach()
	ll, err := NewFilter(mkf, mkk).ListAllNamespaces()

	assert.Nil(t, err)
	assert.NotNil(t, ll)
	mkf.VerifyWasCalledOnce().FetchNamespaces()
}

func TestListServiceAccounts(t *testing.T) {
	mkf := NewMockFetcher()
	m.When(mkf.FetchServiceAccounts()).ThenReturn(&v1.ServiceAccountList{
		Items: []v1.ServiceAccount{
			makeSA("s1"),
			makeSA("s2"),
		},
	}, nil)
	m.When(mkf.ActiveNamespace()).ThenReturn("default")

	mkk := NewMockSpinach()
	m.When(mkk.ExcludedNS("s1")).ThenReturn(false)
	m.When(mkk.ExcludedNS("s2")).ThenReturn(false)

	ll, err := NewFilter(mkf, mkk).ListServiceAccounts()

	assert.Nil(t, err)
	assert.NotNil(t, ll)
	mkf.VerifyWasCalledOnce().FetchServiceAccounts()
	mkf.VerifyWasCalled(pegomock.Times(4)).ActiveNamespace()
	mkk.VerifyWasCalled(pegomock.Times(2)).ExcludedNS("default")
}

func TestListAllServiceAccounts(t *testing.T) {
	mkf := NewMockFetcher()
	m.When(mkf.FetchServiceAccounts()).ThenReturn(&v1.ServiceAccountList{
		Items: []v1.ServiceAccount{
			makeSA("s1"),
			makeSA("s2"),
		},
	}, nil)

	mkk := NewMockSpinach()
	ll, err := NewFilter(mkf, mkk).ListAllServiceAccounts()

	assert.Nil(t, err)
	assert.NotNil(t, ll)
	mkf.VerifyWasCalledOnce().FetchServiceAccounts()
}

func TestListSecrets(t *testing.T) {
	mkf := NewMockFetcher()
	m.When(mkf.FetchSecrets()).ThenReturn(&v1.SecretList{
		Items: []v1.Secret{
			makeSec("s1"),
			makeSec("s2"),
		},
	}, nil)
	m.When(mkf.ActiveNamespace()).ThenReturn("default")

	mkk := NewMockSpinach()
	m.When(mkk.ExcludedNS("s1")).ThenReturn(false)
	m.When(mkk.ExcludedNS("s2")).ThenReturn(false)

	ll, err := NewFilter(mkf, mkk).ListSecrets()

	assert.Nil(t, err)
	assert.NotNil(t, ll)
	mkf.VerifyWasCalledOnce().FetchSecrets()
	mkf.VerifyWasCalled(pegomock.Times(4)).ActiveNamespace()
	mkk.VerifyWasCalled(pegomock.Times(2)).ExcludedNS("default")
}

func TestListAllSecrets(t *testing.T) {
	mkf := NewMockFetcher()
	m.When(mkf.FetchSecrets()).ThenReturn(&v1.SecretList{
		Items: []v1.Secret{
			makeSec("s1"),
			makeSec("s2"),
		},
	}, nil)

	mkk := NewMockSpinach()
	ll, err := NewFilter(mkf, mkk).ListAllSecrets()

	assert.Nil(t, err)
	assert.NotNil(t, ll)
	mkf.VerifyWasCalledOnce().FetchSecrets()
}

func TestListConfigMaps(t *testing.T) {
	mkf := NewMockFetcher()
	m.When(mkf.FetchConfigMaps()).ThenReturn(&v1.ConfigMapList{
		Items: []v1.ConfigMap{
			makeCM("cm1"),
			makeCM("cm2"),
		},
	}, nil)
	m.When(mkf.ActiveNamespace()).ThenReturn("default")

	mkk := NewMockSpinach()
	m.When(mkk.ExcludedNS("p1")).ThenReturn(false)
	m.When(mkk.ExcludedNS("p2")).ThenReturn(false)

	ll, err := NewFilter(mkf, mkk).ListConfigMaps()

	assert.Nil(t, err)
	assert.NotNil(t, ll)
	mkf.VerifyWasCalledOnce().FetchConfigMaps()
	mkf.VerifyWasCalled(pegomock.Times(4)).ActiveNamespace()
	mkk.VerifyWasCalled(pegomock.Times(2)).ExcludedNS("default")
}

func TestListAllConfigMaps(t *testing.T) {
	mkf := NewMockFetcher()
	m.When(mkf.FetchConfigMaps()).ThenReturn(&v1.ConfigMapList{
		Items: []v1.ConfigMap{
			makeCM("cm1"),
			makeCM("cm2"),
		},
	}, nil)

	mkk := NewMockSpinach()
	ll, err := NewFilter(mkf, mkk).ListAllConfigMaps()

	assert.Nil(t, err)
	assert.NotNil(t, ll)
	mkf.VerifyWasCalledOnce().FetchConfigMaps()
}

func TestListPodByLabels(t *testing.T) {
	mkf := NewMockFetcher()
	m.When(mkf.FetchPodsByLabels("app=blee")).ThenReturn(&v1.PodList{
		Items: []v1.Pod{
			makePodLabel("p1"),
			makePodLabel("p2"),
		},
	}, nil)
	m.When(mkf.ActiveNamespace()).ThenReturn("default")

	mkk := NewMockSpinach()
	m.When(mkk.ExcludedNS("p1")).ThenReturn(false)
	m.When(mkk.ExcludedNS("p2")).ThenReturn(false)

	ll, err := NewFilter(mkf, mkk).ListPodsByLabels("app=blee")

	assert.Nil(t, err)
	assert.NotNil(t, ll)
	mkf.VerifyWasCalledOnce().FetchPodsByLabels("app=blee")
	mkf.VerifyWasCalled(pegomock.Times(4)).ActiveNamespace()
	mkk.VerifyWasCalled(pegomock.Times(2)).ExcludedNS("default")
}

func TestListPod(t *testing.T) {
	mkf := NewMockFetcher()
	m.When(mkf.FetchPods()).ThenReturn(&v1.PodList{
		Items: []v1.Pod{
			makePodLabel("p1"),
			makePodLabel("p2"),
		},
	}, nil)
	m.When(mkf.ActiveNamespace()).ThenReturn("default")

	mkk := NewMockSpinach()
	m.When(mkk.ExcludedNS("p1")).ThenReturn(false)
	m.When(mkk.ExcludedNS("p2")).ThenReturn(false)

	ll, err := NewFilter(mkf, mkk).ListPods()

	assert.Nil(t, err)
	assert.NotNil(t, ll)
	mkf.VerifyWasCalledOnce().FetchPods()
	mkf.VerifyWasCalled(pegomock.Times(4)).ActiveNamespace()
	mkk.VerifyWasCalled(pegomock.Times(2)).ExcludedNS("default")
}

func TestListAllPods(t *testing.T) {
	mkf := NewMockFetcher()
	m.When(mkf.FetchPods()).ThenReturn(&v1.PodList{
		Items: []v1.Pod{
			makePodLabel("p1"),
			makePodLabel("p2"),
		},
	}, nil)

	mkk := NewMockSpinach()

	ll, err := NewFilter(mkf, mkk).ListAllPods()

	assert.Nil(t, err)
	assert.NotNil(t, ll)
	mkf.VerifyWasCalledOnce().FetchPods()
}

func TestGetPod(t *testing.T) {
	mkf := NewMockFetcher()
	m.When(mkf.FetchPods()).ThenReturn(&v1.PodList{
		Items: []v1.Pod{
			makePodLabel("p1"),
			makePodLabel("p2"),
		},
	}, nil)
	m.When(mkf.ActiveNamespace()).ThenReturn("default")

	mkk := NewMockSpinach()
	m.When(mkk.ExcludedNS("p1")).ThenReturn(false)
	m.When(mkk.ExcludedNS("p2")).ThenReturn(false)

	po, err := NewFilter(mkf, mkk).GetPod(map[string]string{})

	assert.Nil(t, err)
	assert.NotNil(t, po)
	mkf.VerifyWasCalledOnce().FetchPods()
	mkf.VerifyWasCalled(pegomock.Times(4)).ActiveNamespace()
	mkk.VerifyWasCalled(pegomock.Times(2)).ExcludedNS("default")
}

func TestPodsNamespaces(t *testing.T) {
	mkf := NewMockFetcher()
	m.When(mkf.FetchPods()).ThenReturn(&v1.PodList{
		Items: []v1.Pod{
			makePod("p1"),
			makePod("p2"),
		},
	}, nil)

	mkk := NewMockSpinach()
	m.When(mkk.ExcludedNS("default")).ThenReturn(false)

	nss := make([]string, 1)
	NewFilter(mkf, mkk).PodsNamespaces(nss)

	assert.Equal(t, []string{"default"}, nss)
	mkf.VerifyWasCalledOnce().FetchPods()
	mkk.VerifyWasCalled(pegomock.Times(2)).ExcludedNS("default")
}

func TestListNodes(t *testing.T) {
	mkf := NewMockFetcher()
	m.When(mkf.FetchNodes()).ThenReturn(&v1.NodeList{
		Items: []v1.Node{
			makeNode("n1"),
			makeNode("n2"),
		},
	}, nil)

	mkk := NewMockSpinach()
	m.When(mkk.ExcludedNode("n1")).ThenReturn(false)
	m.When(mkk.ExcludedNode("n2")).ThenReturn(false)

	ll, err := NewFilter(mkf, mkk).ListNodes()

	assert.Nil(t, err)
	assert.Equal(t, 2, len(ll))
	mkf.VerifyWasCalledOnce().FetchNodes()
	mkk.VerifyWasCalledOnce().ExcludedNode("n1")
	mkk.VerifyWasCalledOnce().ExcludedNode("n2")
}

func TestListServices(t *testing.T) {
	mkf := NewMockFetcher()
	m.When(mkf.FetchServices()).ThenReturn(&v1.ServiceList{
		Items: []v1.Service{
			makeSvc("s1"),
			makeSvc("s2"),
		},
	}, nil)

	mkk := NewMockSpinach()
	m.When(mkk.ExcludedNS("default")).ThenReturn(false)

	ll, err := NewFilter(mkf, mkk).ListServices()

	assert.Nil(t, err)
	assert.Equal(t, 2, len(ll))
	mkf.VerifyWasCalledOnce().FetchServices()
}

func TestListAllServices(t *testing.T) {
	mkf := NewMockFetcher()
	m.When(mkf.FetchServices()).ThenReturn(&v1.ServiceList{
		Items: []v1.Service{
			makeSvc("s1"),
			makeSvc("s2"),
		},
	}, nil)

	mkk := NewMockSpinach()
	m.When(mkk.ExcludedNS("default")).ThenReturn(false)

	ll, err := NewFilter(mkf, mkk).ListAllServices()

	assert.Nil(t, err)
	assert.Equal(t, 2, len(ll))
	mkf.VerifyWasCalledOnce().FetchServices()
}

func TestListClusterRoleBindings(t *testing.T) {
	mkf := NewMockFetcher()
	m.When(mkf.FetchClusterRoleBindings()).ThenReturn(&rbacv1.ClusterRoleBindingList{
		Items: []rbacv1.ClusterRoleBinding{
			makeCRB("crb1", "sa1"),
			makeCRB("crb2", "sa1"),
		},
	}, nil)

	mkk := NewMockSpinach()
	m.When(mkk.ExcludedNS("default")).ThenReturn(false)

	ll, err := NewFilter(mkf, mkk).ListAllClusterRoleBindings()

	assert.Nil(t, err)
	assert.Equal(t, 2, len(ll))
	mkf.VerifyWasCalledOnce().FetchClusterRoleBindings()
}

func TestListRoleBindings(t *testing.T) {
	mkf := NewMockFetcher()
	m.When(mkf.FetchRoleBindings()).ThenReturn(&rbacv1.RoleBindingList{
		Items: []rbacv1.RoleBinding{
			makeRB("rb1", "sa1"),
			makeRB("rb2", "sa1"),
		},
	}, nil)

	mkk := NewMockSpinach()
	m.When(mkk.ExcludedNS("default")).ThenReturn(false)

	ll, err := NewFilter(mkf, mkk).ListRoleBindings()

	assert.Nil(t, err)
	assert.Equal(t, 2, len(ll))
	mkf.VerifyWasCalledOnce().FetchRoleBindings()
	mkk.VerifyWasCalled(pegomock.Times(2)).ExcludedNS("default")
}

func TestListAllRoleBindings(t *testing.T) {
	mkf := NewMockFetcher()
	m.When(mkf.FetchRoleBindings()).ThenReturn(&rbacv1.RoleBindingList{
		Items: []rbacv1.RoleBinding{
			makeRB("rb1", "sa1"),
			makeRB("rb2", "sa1"),
		},
	}, nil)

	mkk := NewMockSpinach()
	m.When(mkk.ExcludedNS("default")).ThenReturn(false)

	ll, err := NewFilter(mkf, mkk).ListAllRoleBindings()

	assert.Nil(t, err)
	assert.Equal(t, 2, len(ll))
	mkf.VerifyWasCalledOnce().FetchRoleBindings()
}

func TestGetEndPoints(t *testing.T) {
	uu := []struct {
		ep    v1.Endpoints
		svc   v1.Service
		err   error
		count int
		nilOk bool
	}{
		// Matching EP => All good!
		{
			makeEp("s1", "1.2.3.4"),
			makeSvcType("s1", v1.ServiceTypeClusterIP, map[string]string{"app": "blee"}),
			nil,
			0,
			true,
		},
		// Missing EP but service has selector => yield error
		{
			makeEp("s2", "1.2.3.4"),
			makeSvcType("s1", v1.ServiceTypeClusterIP, map[string]string{"app": "blee"}),
			nil,
			0,
			false,
		},
		// Missing EP but no selectors => no error
		{
			makeEp("s2", "1.2.3.4"),
			makeSvcType("s1", v1.ServiceTypeClusterIP, nil),
			nil,
			1,
			false,
		},
	}

	for _, u := range uu {
		mkl := NewMockFetcher()
		m.When(mkl.FetchEndpoints()).ThenReturn(&v1.EndpointsList{Items: []v1.Endpoints{u.ep}}, nil)

		mkf := NewMockSpinach()

		c := NewFilter(mkl, mkf)
		ep, err := c.GetEndpoints("default/s1")

		assert.Equal(t, u.err, err)
		if u.err == nil && u.nilOk {
			assert.Equal(t, &u.ep, ep)
		} else {
			assert.Nil(t, ep)
		}
		mkl.VerifyWasCalledOnce().FetchEndpoints()
	}
}

func TestPodsMetrics(t *testing.T) {
	metrics := v1beta1.PodMetricsList{
		Items: []v1beta1.PodMetrics{
			makeMxPod("p1", "1", "4Gi"),
			makeMxPod("p2", "50m", "1Mi"),
		},
	}

	mmx := make(k8s.PodsMetrics)
	var f *Filter
	f.ListPodsMetrics(metrics.Items, mmx)
	assert.Equal(t, 2, len(mmx))

	mx, ok := mmx["default/p1"]
	assert.True(t, ok)
	assert.Equal(t, toQty("1"), mx["c1"].CurrentCPU)
	assert.Equal(t, toQty("4Gi"), mx["c1"].CurrentMEM)
}

func BenchmarkPodsMetrics(b *testing.B) {
	metrics := v1beta1.PodMetricsList{
		Items: []v1beta1.PodMetrics{
			makeMxPod("p1", "1", "4Gi"),
			makeMxPod("p2", "50m", "1Mi"),
			makeMxPod("p3", "50m", "1Mi"),
		},
	}
	mmx := make(k8s.PodsMetrics, 3)

	var f *Filter

	b.ResetTimer()
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		f.ListPodsMetrics(metrics.Items, mmx)
	}
}

func TestNodesMetrics(t *testing.T) {
	nodes := v1.NodeList{
		Items: []v1.Node{
			makeNodeMX("n1", "32", "128Gi", "50m", "2Mi"),
			makeNodeMX("n2", "8", "4Gi", "50m", "2Mi"),
		},
	}

	metrics := v1beta1.NodeMetricsList{
		Items: []v1beta1.NodeMetrics{
			makeMxNode("n1", "10", "8Gi"),
			makeMxNode("n2", "50m", "1Mi"),
		},
	}

	mmx := make(k8s.NodesMetrics)
	var f *Filter
	f.ListNodesMetrics(nodes.Items, metrics.Items, mmx)

	assert.Equal(t, 2, len(mmx))
	mx, ok := mmx["n1"]
	assert.True(t, ok)
	assert.Equal(t, toQty("32"), mx.TotalCPU)
	assert.Equal(t, toQty("128Gi"), mx.TotalMEM)
	assert.Equal(t, toQty("50m"), mx.AvailableCPU)
	assert.Equal(t, toQty("2Mi"), mx.AvailableMEM)
	assert.Equal(t, toQty("10"), mx.CurrentCPU)
	assert.Equal(t, toQty("8Gi"), mx.CurrentMEM)
}

func BenchmarkNodesMetrics(b *testing.B) {
	nodes := v1.NodeList{
		Items: []v1.Node{
			makeNodeMX("n1", "100m", "4Mi", "50m", "2Mi"),
			makeNodeMX("n2", "100m", "4Mi", "50m", "2Mi"),
		},
	}

	metrics := v1beta1.NodeMetricsList{
		Items: []v1beta1.NodeMetrics{
			makeMxNode("n1", "50m", "1Mi"),
			makeMxNode("n2", "50m", "1Mi"),
		},
	}

	mmx := make(k8s.NodesMetrics)
	var f *Filter

	b.ResetTimer()
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		f.ListNodesMetrics(nodes.Items, metrics.Items, mmx)
	}
}

// ----------------------------------------------------------------------------
// Helpers...

func makeSvcType(s string, kind v1.ServiceType, sel map[string]string) v1.Service {
	svc := makeSvc(s)
	svc.Spec = v1.ServiceSpec{
		Selector: sel,
		Type:     kind,
	}

	return svc
}

func makePodLabel(n string) v1.Pod {
	po := makePod(n)
	po.ObjectMeta.Labels = map[string]string{
		"l1": "v1",
	}
	return po
}

func makeNodeMX(name, tcpu, tmem, acpu, amem string) v1.Node {
	return v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Status: v1.NodeStatus{
			Capacity:    makeRes(tcpu, tmem),
			Allocatable: makeRes(acpu, amem),
		},
	}
}

func makeMxNode(name, cpu, mem string) v1beta1.NodeMetrics {
	return v1beta1.NodeMetrics{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Usage: makeRes(cpu, mem),
	}
}

func makeRes(c, m string) v1.ResourceList {
	cpu, _ := resource.ParseQuantity(c)
	mem, _ := resource.ParseQuantity(m)

	return v1.ResourceList{
		v1.ResourceCPU:    cpu,
		v1.ResourceMemory: mem,
	}
}
