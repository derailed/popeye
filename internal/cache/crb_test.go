package cache_test

import (
	"testing"

	"github.com/derailed/popeye/internal/cache"
	"github.com/stretchr/testify/assert"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestClusterRoleRef(t *testing.T) {
	cr := cache.NewClusterRoleBinding(makeCRBMap())
	refs := make(cache.ObjReferences)
	cr.ClusterRoleRefs(refs)

	assert.Equal(t, 2, len(refs))
	m, ok := refs["clusterrole:cr1"]
	assert.True(t, ok)
	_, ok = m["crb1"]
	assert.True(t, ok)

	m, ok = refs["role:blee/r1"]
	assert.True(t, ok)
	_, ok = m["crb2"]
	assert.True(t, ok)
}

// Helpers...

func makeCRBMap() map[string]*rbacv1.ClusterRoleBinding {
	return map[string]*rbacv1.ClusterRoleBinding{
		"crb1": makeCRB("", "crb1", "ClusterRole", "cr1"),
		"crb2": makeCRB("blee", "crb2", "Role", "r1"),
	}
}

func makeCRB(ns, name, kind, refName string) *rbacv1.ClusterRoleBinding {
	return &rbacv1.ClusterRoleBinding{
		ObjectMeta: makeObjMeta(ns, name),
		RoleRef: rbacv1.RoleRef{
			Kind: kind,
			Name: refName,
		},
	}
}

func makeObjMeta(ns, n string) metav1.ObjectMeta {
	m := metav1.ObjectMeta{Name: n}
	if ns != "" {
		m.Namespace = ns
	}

	return m
}
