package cache

import (
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
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
				v1.ResourceCPU:    toQty("400m"),
				v1.ResourceMemory: toQty("300Mi"),
			},
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			n := NewNodesMetrics(map[string]*mv1beta1.NodeMetrics{})
			res := n.ListAvailableMetrics(u.nn)
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

func makeNodeMx(n, cpu, mem string) *v1.Node {
	return &v1.Node{
		ObjectMeta: metav1.ObjectMeta{Name: n},
		Status: v1.NodeStatus{
			Allocatable: v1.ResourceList{
				v1.ResourceCPU:    toQty(cpu),
				v1.ResourceMemory: toQty(mem),
			},
		},
	}
}
