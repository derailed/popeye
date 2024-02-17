// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Popeye

package cache

import (
	"testing"

	"github.com/derailed/popeye/internal"
	"github.com/derailed/popeye/internal/db"
	"github.com/derailed/popeye/internal/test"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	mv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

func TestClusterAllocatableMetrics(t *testing.T) {
	uu := map[string]struct {
		nn map[string]*v1.Node
		e  v1.ResourceList
	}{
		"cool": {
			nn: map[string]*v1.Node{
				"n1": makeNodeMx("n1", "100m", "100Mi"),
				"n2": makeNodeMx("n2", "300m", "200Mi"),
			},
			e: v1.ResourceList{
				v1.ResourceCPU:    test.ToQty("2"),
				v1.ResourceMemory: test.ToQty("200Mi"),
			},
		},
	}

	dba, err := test.NewTestDB()
	assert.NoError(t, err)
	l := db.NewLoader(dba)

	ctx := test.MakeCtx(t)
	assert.NoError(t, test.LoadDB[*mv1beta1.NodeMetrics](ctx, l.DB, "mx/node/1.yaml", internal.Glossary[internal.NMX]))
	assert.NoError(t, test.LoadDB[*v1.Node](ctx, l.DB, "core/node/1.yaml", internal.Glossary[internal.NO]))

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			res, err := ListAvailableMetrics(dba)
			assert.NoError(t, err)

			assert.Equal(t, u.e.Cpu().Value(), res.Cpu().Value())
			assert.Equal(t, u.e.Memory().Value(), res.Memory().Value())
		})
	}
}

// ----------------------------------------------------------------------------
// Helpers...

func makeNodeMx(n, cpu, mem string) *v1.Node {
	return &v1.Node{
		ObjectMeta: metav1.ObjectMeta{Name: n},
		Status: v1.NodeStatus{
			Allocatable: v1.ResourceList{
				v1.ResourceCPU:    test.ToQty(cpu),
				v1.ResourceMemory: test.ToQty(mem),
			},
		},
	}
}
