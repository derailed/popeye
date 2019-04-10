package linter

import (
	"fmt"
	"testing"

	m "github.com/petergtz/pegomock"
	pegomock "github.com/petergtz/pegomock"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
)

func TestIsSystemNS(t *testing.T) {
	assert.True(t, isSystemNS("kube-system"))
	assert.False(t, isSystemNS("fred"))
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

	po, err := NewFilter(mkf, mkk).ListNamespaces()

	assert.Nil(t, err)
	assert.NotNil(t, po)
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
	po, err := NewFilter(mkf, mkk).ListAllNamespaces()

	assert.Nil(t, err)
	assert.NotNil(t, po)
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

	po, err := NewFilter(mkf, mkk).ListServiceAccounts()

	assert.Nil(t, err)
	assert.NotNil(t, po)
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
	po, err := NewFilter(mkf, mkk).ListAllServiceAccounts()

	assert.Nil(t, err)
	assert.NotNil(t, po)
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

	po, err := NewFilter(mkf, mkk).ListSecrets()

	assert.Nil(t, err)
	assert.NotNil(t, po)
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
	po, err := NewFilter(mkf, mkk).ListAllSecrets()

	assert.Nil(t, err)
	assert.NotNil(t, po)
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

	po, err := NewFilter(mkf, mkk).ListConfigMaps()

	assert.Nil(t, err)
	assert.NotNil(t, po)
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
	po, err := NewFilter(mkf, mkk).ListAllConfigMaps()

	assert.Nil(t, err)
	assert.NotNil(t, po)
	mkf.VerifyWasCalledOnce().FetchConfigMaps()
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

	po, err := NewFilter(mkf, mkk).ListPods()

	assert.Nil(t, err)
	assert.NotNil(t, po)
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

	po, err := NewFilter(mkf, mkk).ListAllPods()

	assert.Nil(t, err)
	assert.NotNil(t, po)
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

	rbs, err := NewFilter(mkf, mkk).ListNodes()

	assert.Nil(t, err)
	assert.Equal(t, 2, len(rbs))
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

	rbs, err := NewFilter(mkf, mkk).ListServices()

	assert.Nil(t, err)
	assert.Equal(t, 2, len(rbs))
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

	rbs, err := NewFilter(mkf, mkk).ListAllServices()

	assert.Nil(t, err)
	assert.Equal(t, 2, len(rbs))
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

	rbs, err := NewFilter(mkf, mkk).ListAllClusterRoleBindings()

	assert.Nil(t, err)
	assert.Equal(t, 2, len(rbs))
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

	rbs, err := NewFilter(mkf, mkk).ListRoleBindings()

	assert.Nil(t, err)
	assert.Equal(t, 2, len(rbs))
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

	rbs, err := NewFilter(mkf, mkk).ListAllRoleBindings()

	assert.Nil(t, err)
	assert.Equal(t, 2, len(rbs))
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
			fmt.Errorf("Unable to find ep for service default/s1"),
			1,
			true,
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
		m.When(mkl.FetchServices()).ThenReturn(&v1.ServiceList{Items: []v1.Service{u.svc}}, nil)

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
		mkl.VerifyWasCalled(pegomock.Times(u.count)).FetchServices()
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
