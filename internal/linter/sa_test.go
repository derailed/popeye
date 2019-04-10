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
	mkl := NewMockLoader()
	m.When(mkl.ActiveNamespace()).ThenReturn("")
	m.When(mkl.ListAllCRBs()).ThenReturn(map[string]rbacv1.ClusterRoleBinding{
		"crb1": makeCRB("crb1", "sa1"),
	}, nil)
	m.When(mkl.ListRBs()).ThenReturn(map[string]rbacv1.RoleBinding{
		"rb1": makeRB("rb1", "sa1"),
	}, nil)

	m.When(mkl.ListAllPods()).ThenReturn(map[string]v1.Pod{
		"p1": makePodSa("p1", "sa2"),
	}, nil)

	s := NewSA(mkl, nil)
	s.Lint(context.Background())

	assert.Equal(t, 1, len(s.Issues()["default/sa1"]))
	mkl.VerifyWasCalledOnce().ListAllCRBs()
	mkl.VerifyWasCalledOnce().ListRBs()
	mkl.VerifyWasCalledOnce().ListAllPods()
}

func TestSACheckDead(t *testing.T) {
	uu := []struct {
		crbs  map[string]rbacv1.ClusterRoleBinding
		rbs   map[string]rbacv1.RoleBinding
		pods  map[string]v1.Pod
		issue int
	}{
		{
			crbs:  map[string]rbacv1.ClusterRoleBinding{"crb1": makeCRB("crb1", "sa1")},
			rbs:   map[string]rbacv1.RoleBinding{"default/rb1": makeRB("rb1", "sa2")},
			pods:  map[string]v1.Pod{"p1": makePodSa("p1", "sa1")},
			issue: 1,
		},
		{
			crbs: map[string]rbacv1.ClusterRoleBinding{"crb1": makeCRB("crb1", "sa2")},
			rbs:  map[string]rbacv1.RoleBinding{"default/rb1": makeRB("rb1", "sa2")},
			pods: map[string]v1.Pod{
				"p1": makePodSa("p1", "sa2"),
			},
			issue: 0,
		},
	}

	mkl := NewMockLoader()
	m.When(mkl.ExcludedNS("default")).ThenReturn(false)
	for _, u := range uu {
		s := NewSA(mkl, nil)
		s.checkDead(u.pods, u.crbs, u.rbs)

		assert.Equal(t, u.issue, len(s.Issues()["default/sa2"]))
	}
	mkl.VerifyWasCalled(pegomock.Times(6)).ExcludedNS("default")
}

// ----------------------------------------------------------------------------
// Helpers...

func makePodSa(s, sa string) v1.Pod {
	po := makePod(s)
	po.Spec.ServiceAccountName = sa

	return po
}

func makeCRB(s, sa string) rbacv1.ClusterRoleBinding {
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

func makeRB(s, sa string) rbacv1.RoleBinding {
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
