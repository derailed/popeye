package linter

import (
	"context"
	"testing"

	m "github.com/petergtz/pegomock"
	pegomock "github.com/petergtz/pegomock"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestSALint(t *testing.T) {
	mkc := NewMockClient()
	m.When(mkc.ActiveNamespace()).ThenReturn("")
	m.When(mkc.ListAllCRBs()).ThenReturn(map[string]rbacv1.ClusterRoleBinding{
		"crb1": makeCrb("crb1", "sa1"),
	}, nil)
	m.When(mkc.ListRBs()).ThenReturn(map[string]rbacv1.RoleBinding{
		"rb1": makeRb("rb1", "sa1"),
	}, nil)

	m.When(mkc.ListAllPods()).ThenReturn(map[string]v1.Pod{
		"p1": makePodSa("p1", "sa2"),
	}, nil)

	s := NewSA(mkc, nil)
	s.Lint(context.Background())

	assert.Equal(t, 1, len(s.Issues()["default/sa1"]))
	mkc.VerifyWasCalledOnce().ListAllCRBs()
	mkc.VerifyWasCalledOnce().ListRBs()
	mkc.VerifyWasCalledOnce().ListAllPods()
}

func TestSACheckDead(t *testing.T) {
	uu := []struct {
		crbs  map[string]rbacv1.ClusterRoleBinding
		rbs   map[string]rbacv1.RoleBinding
		pods  map[string]v1.Pod
		issue int
	}{
		{
			crbs:  map[string]rbacv1.ClusterRoleBinding{"crb1": makeCrb("crb1", "sa1")},
			rbs:   map[string]rbacv1.RoleBinding{"default/rb1": makeRb("rb1", "sa2")},
			pods:  map[string]v1.Pod{"p1": makePodSa("p1", "sa1")},
			issue: 1,
		},
		{
			crbs: map[string]rbacv1.ClusterRoleBinding{"crb1": makeCrb("crb1", "sa2")},
			rbs:  map[string]rbacv1.RoleBinding{"default/rb1": makeRb("rb1", "sa2")},
			pods: map[string]v1.Pod{
				"p1": makePodSa("p1", "sa2"),
			},
			issue: 0,
		},
	}

	mks := NewMockClient()
	m.When(mks.ExcludedNS("default")).ThenReturn(false)
	for _, u := range uu {
		s := NewSA(mks, nil)
		s.checkDead(u.pods, u.crbs, u.rbs)

		assert.Equal(t, u.issue, len(s.Issues()["default/sa2"]))
	}
	mks.VerifyWasCalled(pegomock.Times(6)).ExcludedNS("default")
}

// ----------------------------------------------------------------------------
// Helpers...

func makePodSa(s, sa string) v1.Pod {
	po := makePod(s)
	po.Spec.ServiceAccountName = sa

	return po
}

func makeCrb(s, sa string) rbacv1.ClusterRoleBinding {
	return rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: s,
		},
		Subjects: []rbacv1.Subject{
			{
				Kind: "ServiceAccount",
				Name: sa,
			},
		},
	}
}

func makeRb(s, sa string) rbacv1.RoleBinding {
	return rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      s,
			Namespace: "default",
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      sa,
				Namespace: "default",
			},
		},
	}
}
