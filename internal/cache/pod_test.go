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
	v1 "k8s.io/api/core/v1"
)

func TestPodRef(t *testing.T) {
	dba, err := test.NewTestDB()
	assert.NoError(t, err)
	l := db.NewLoader(dba)

	ctx := test.MakeCtx(t)
	assert.NoError(t, test.LoadDB[*v1.Pod](ctx, l.DB, "core/pod/1.yaml", internal.Glossary[internal.PO]))

	cr := cache.NewPod(dba)
	var refs sync.Map
	assert.NoError(t, cr.PodRefs(&refs))

	uu := map[string]struct {
		k  string
		vv []string
	}{
		"ns": {
			k: "ns",
		},
		"cm1-env": {
			k:  "cm:default/cm1",
			vv: []string{"blee", "ns"},
		},
		"cm3-vol": {
			k:  "cm:default/cm3",
			vv: []string{"k1", "k2", "k3", "k4"},
		},
		"cm4-env-from": {
			k: "cm:default/cm4",
		},
		"sec1-vol": {
			k:  "sec:default/sec1",
			vv: []string{"k1"},
		},
		"sec2-env": {
			k:  "sec:default/sec2",
			vv: []string{"ca.crt", "fred", "k1", "namespace"},
		},
		"sec3-img-pull": {
			k: "sec:default/sec3",
		},
		"sec4-env-from": {
			k: "sec:default/sec4",
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			m, ok := refs.Load(u.k)
			assert.True(t, ok)
			for _, k := range u.vv {
				_, ok = m.(internal.StringSet)[k]
				assert.True(t, ok)
			}
		})
	}
}
