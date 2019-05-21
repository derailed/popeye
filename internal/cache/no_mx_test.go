package cache

import (
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	mv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

func TestListClusterMetrics(t *testing.T) {
	uu := map[string]struct {
		nmx map[string]*mv1beta1.NodeMetrics
		e   v1.ResourceList
	}{
		"cool": {
			map[string]*mv1beta1.NodeMetrics{
				"n1": {
					Usage: v1.ResourceList{
						v1.ResourceCPU:    toQty("100m"),
						v1.ResourceMemory: toQty("100Mi"),
					},
				},
				"n2": {
					Usage: v1.ResourceList{
						v1.ResourceCPU:    toQty("100m"),
						v1.ResourceMemory: toQty("100Mi"),
					},
				},
			},
			v1.ResourceList{
				v1.ResourceCPU:    toQty("200m"),
				v1.ResourceMemory: toQty("200Mi"),
			},
		},
	}

	for k, u := range uu {
		t.Run(k, func(t *testing.T) {
			n := NewNodesMetrics(u.nmx)
			res := n.ListClusterMetrics()
			assert.Equal(t, u.e.Cpu().Value(), res.Cpu().Value())
			assert.Equal(t, u.e.Memory().Value(), res.Memory().Value())
		})
	}
}

// ----------------------------------------------------------------------------
// Helpers...

func toQty(s string) resource.Quantity {
	q, _ := resource.ParseQuantity(s)

	return q
}
