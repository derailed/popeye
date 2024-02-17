// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Popeye

package cache_test

import (
	"sync"
	"testing"

	"github.com/derailed/popeye/internal"
	"github.com/derailed/popeye/internal/cache"
	"github.com/derailed/popeye/internal/db"
	"github.com/derailed/popeye/internal/test"
	"github.com/stretchr/testify/assert"
	rbacv1 "k8s.io/api/rbac/v1"
)

func TestRoleRef(t *testing.T) {
	dba, err := test.NewTestDB()
	assert.NoError(t, err)
	l := db.NewLoader(dba)

	ctx := test.MakeCtx(t)
	assert.NoError(t, test.LoadDB[*rbacv1.RoleBinding](ctx, l.DB, "auth/rob/1.yaml", internal.Glossary[internal.ROB]))

	cr := cache.NewRoleBinding(dba)
	var refs sync.Map
	cr.RoleRefs(&refs)

	m, ok := refs.Load("clusterrole:cr-bozo")
	assert.True(t, ok)
	_, ok = m.(internal.StringSet)["default/rb3"]
	assert.True(t, ok)

	m, ok = refs.Load("role:default/r1")
	assert.True(t, ok)
	_, ok = m.(internal.StringSet)["default/rb1"]
	assert.True(t, ok)
}
