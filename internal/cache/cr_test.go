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

func TestClusterRoleAggregation(t *testing.T) {
	dba, err := test.NewTestDB()
	assert.NoError(t, err)
	l := db.NewLoader(dba)

	ctx := test.MakeCtx(t)
	assert.NoError(t, test.LoadDB[*rbacv1.ClusterRole](ctx, l.DB, "auth/cr/1.yaml", internal.Glossary[internal.CR]))

	cr := cache.NewClusterRole(dba)
	var aRefs sync.Map
	cr.AggregationMatchers(&aRefs)

	value, ok := aRefs.Load("rbac.authorization.k8s.io/aggregate-to-cr4")
	assert.True(t, ok)
	assert.Equal(t, "true", value)
}
