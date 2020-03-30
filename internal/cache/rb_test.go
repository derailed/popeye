package cache_test

import (
	"sync"
	"testing"

	"github.com/derailed/popeye/internal"
	"github.com/derailed/popeye/internal/cache"
	"github.com/stretchr/testify/assert"
	rbacv1 "k8s.io/api/rbac/v1"
)

func TestRoleRef(t *testing.T) {
	cr := cache.NewRoleBinding(makeRBMap())
	var refs sync.Map
	cr.RoleRefs(&refs)

	m, ok := refs.Load("clusterrole:cr1")
	assert.True(t, ok)
	_, ok = m.(internal.StringSet)["rb1"]
	assert.True(t, ok)

	m, ok = refs.Load("role:blee/r1")
	assert.True(t, ok)
	_, ok = m.(internal.StringSet)["rb2"]
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
