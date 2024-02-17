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

func TestClusterRoleRef(t *testing.T) {
	dba, err := test.NewTestDB()
	assert.NoError(t, err)
	l := db.NewLoader(dba)

	ctx := test.MakeCtx(t)
	assert.NoError(t, test.LoadDB[*rbacv1.ClusterRoleBinding](ctx, l.DB, "auth/crb/1.yaml", internal.Glossary[internal.CRB]))

	cr := cache.NewClusterRoleBinding(dba)
	var refs sync.Map
	cr.ClusterRoleRefs(&refs)

	m, ok := refs.Load("clusterrole:cr1")
	assert.True(t, ok)
	_, ok = m.(internal.StringSet)["crb1"]
	assert.True(t, ok)

	m, ok = refs.Load("role:r1")
	assert.True(t, ok)

	_, ok = m.(internal.StringSet)["crb3"]
	assert.True(t, ok)
}
