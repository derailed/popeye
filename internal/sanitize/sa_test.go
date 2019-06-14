package sanitize

import (
	"context"
	"testing"

	"github.com/derailed/popeye/internal/cache"
	"github.com/derailed/popeye/internal/issues"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestSASanitize(t *testing.T) {
	uu := map[string]struct {
		lister ServiceAccountLister
		issues int
	}{
		"cool":    {makeSALister("sa1", "sa1"), 0},
		"notUsed": {makeSALister("sa1", "sa2"), 1},
	}

	for k, u := range uu {
		t.Run(k, func(t *testing.T) {
			s := NewServiceAccount(issues.NewCollector(), u.lister)
			s.Sanitize(context.Background())

			assert.Equal(t, u.issues, len(s.Outcome()["default/sa1"]))
		})
	}
}

// ----------------------------------------------------------------------------
// Helpers...

type sa struct {
	name string
	used string
}

func makeSALister(n, used string) sa {
	return sa{name: n, used: used}
}

func (s sa) ActiveNamespace() string {
	return ""
}

func (s sa) ExcludedNS(ns string) bool {
	return false
}

func (s sa) ListClusterRoleBindings() map[string]*rbacv1.ClusterRoleBinding {
	return map[string]*rbacv1.ClusterRoleBinding{
		"crb1": makeCRB("crb1", s.used),
	}
}

func (s sa) ListRoleBindings() map[string]*rbacv1.RoleBinding {
	return map[string]*rbacv1.RoleBinding{
		"default/rb1": makeRB("rb1", s.used),
	}
}

func (s sa) ListPods() map[string]*v1.Pod {
	return map[string]*v1.Pod{
		"default/p1": makePodSa("p1", s.used),
	}
}

func (s sa) GetPod(map[string]string) *v1.Pod {
	return nil
}

func (s sa) ListServiceAccounts() map[string]*v1.ServiceAccount {
	return map[string]*v1.ServiceAccount{
		cache.FQN("default", s.name): makeSa(s.name),
	}
}

func makeSa(n string) *v1.ServiceAccount {
	return &v1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      n,
			Namespace: "default",
		},
	}
}

func makePodSa(s, sa string) *v1.Pod {
	po := makePod(s)
	po.Spec.ServiceAccountName = sa

	return po
}

func makeCRB(s, sa string) *rbacv1.ClusterRoleBinding {
	return &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: s,
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

func makeRB(s, sa string) *rbacv1.RoleBinding {
	return &rbacv1.RoleBinding{
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
