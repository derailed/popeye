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
		"cool": {
			makeSALister("sa1", saOpts{
				used: "sa1",
			}),
			0,
		},
		"notUsed": {
			makeSALister("sa1", saOpts{
				used: "sa2",
			}),
			1,
		},
		"missingSecret": {
			makeSALister("sa1", saOpts{
				used:       "sa1",
				secret:     "blee",
				pullSecret: "fred",
			}),
			2,
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			s := NewServiceAccount(issues.NewCollector(loadCodes(t)), u.lister)

			assert.Nil(t, s.Sanitize(context.Background()))
			assert.Equal(t, u.issues, len(s.Outcome()["default/sa1"]))
		})
	}
}

// ----------------------------------------------------------------------------
// Helpers...

type saOpts struct {
	used               string
	secret, pullSecret string
}

type sa struct {
	name string
	opts saOpts
}

func makeSALister(n string, opts saOpts) sa {
	return sa{name: n, opts: opts}
}

func (s sa) ActiveNamespace() string {
	return ""
}

func (s sa) ExcludedNS(ns string) bool {
	return false
}

func (s sa) ListClusterRoleBindings() map[string]*rbacv1.ClusterRoleBinding {
	return map[string]*rbacv1.ClusterRoleBinding{
		"crb1": makeCRB("crb1", s.opts.used),
	}
}

func (s sa) ListRoleBindings() map[string]*rbacv1.RoleBinding {
	return map[string]*rbacv1.RoleBinding{
		"default/rb1": makeRB("rb1", s.opts.used),
	}
}

func (s sa) ListPods() map[string]*v1.Pod {
	return map[string]*v1.Pod{
		"default/p1": makePodSa("p1", s.opts.used),
	}
}

func (s sa) IngressRefs(cache.ObjReferences) {}

func (s sa) ServiceAccountRefs(cache.ObjReferences) {}

func (s sa) PodRefs(cache.ObjReferences) {}

func (s sa) ListSecrets() map[string]*v1.Secret {
	return map[string]*v1.Secret{
		"default/s1":   makeSecret("s1"),
		"default/dks1": makeDockerSecret("dks1"),
	}
}

func (s sa) GetPod(map[string]string) *v1.Pod {
	return nil
}

func (s sa) ListServiceAccounts() map[string]*v1.ServiceAccount {
	return map[string]*v1.ServiceAccount{
		cache.FQN("default", s.name): makeSa(s.name, s.opts.secret, s.opts.pullSecret),
	}
}

func makeSa(n, secret, pullSecret string) *v1.ServiceAccount {
	sa := v1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      n,
			Namespace: "default",
		},
	}

	if secret != "" {
		sa.Secrets = []v1.ObjectReference{
			{Namespace: "default", Name: secret, Kind: "secret"},
		}
	}

	if pullSecret != "" {
		sa.ImagePullSecrets = []v1.LocalObjectReference{
			{Name: pullSecret},
		}
	}

	return &sa
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
