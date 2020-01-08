package cache_test

import (
	"testing"

	"github.com/derailed/popeye/internal/cache"
	"github.com/stretchr/testify/assert"
	rbacv1 "k8s.io/api/rbac/v1"
)

func TestRoleRef(t *testing.T) {
	cr := cache.NewRoleBinding(makeRBMap())
	refs := make(cache.ObjReferences)
	cr.RoleRefs(refs)

	assert.Equal(t, 2, len(refs))
	m, ok := refs["clusterrole:cr1"]
	assert.True(t, ok)
	_, ok = m["rb1"]
	assert.True(t, ok)

	m, ok = refs["role:blee/r1"]
	assert.True(t, ok)
	_, ok = m["rb2"]
	assert.True(t, ok)
}

// Helpers...

func makeRBMap() map[string]*rbacv1.RoleBinding {
	return map[string]*rbacv1.RoleBinding{
		"rb1": makeRB("", "r1", "ClusterRole", "cr1"),
		"rb2": makeRB("blee", "r2", "Role", "r1"),
	}
}

func makeRB(ns, name, kind, refName string) *rbacv1.RoleBinding {
	return &rbacv1.RoleBinding{
		ObjectMeta: makeObjMeta(ns, name),
		RoleRef: rbacv1.RoleRef{
			Kind: kind,
			Name: refName,
		},
	}
}
