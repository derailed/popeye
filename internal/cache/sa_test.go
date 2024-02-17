// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Popeye

package cache

import (
	"sync"
	"testing"

	"github.com/derailed/popeye/internal"
	"github.com/derailed/popeye/internal/db"
	"github.com/derailed/popeye/internal/test"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
)

func TestServiceAccountRefs(t *testing.T) {
	dba, err := test.NewTestDB()
	assert.NoError(t, err)
	l := db.NewLoader(dba)

	ctx := test.MakeCtx(t)
	assert.NoError(t, test.LoadDB[*v1.ServiceAccount](ctx, l.DB, "core/sa/1.yaml", internal.Glossary[internal.SA]))

	uu := []struct {
		keys []string
	}{
		{
			[]string{
				"sec:default/s1",
				"sec:default/bozo",
			},
		},
	}

	var refs sync.Map
	sa := NewServiceAccount(dba)
	assert.NoError(t, sa.ServiceAccountRefs(&refs))
	for _, u := range uu {
		for _, k := range u.keys {
			v, ok := refs.Load(k)
			assert.True(t, ok)
			assert.Equal(t, internal.AllKeys, v.(internal.StringSet))
		}
	}
}
