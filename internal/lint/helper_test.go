// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Popeye

package lint

// import (
// 	"testing"

// 	"github.com/stretchr/testify/assert"
// 	v1 "k8s.io/api/core/v1"
// 	"k8s.io/apimachinery/pkg/api/resource"
// )

// func TestNamepaced(t *testing.T) {
// 	uu := []struct {
// 		s     string
// 		ns, n string
// 	}{
// 		{"fred/blee", "fred", "blee"},
// 		{"fred", "", "fred"},
// 	}

// 	for _, u := range uu {
// 		ns, n := namespaced(u.s)
// 		assert.Equal(t, u.ns, ns)
// 		assert.Equal(t, u.n, n)
// 	}
// }

// func TestPluralOf(t *testing.T) {
// 	uu := []struct {
// 		n     string
// 		count int
// 		e     string
// 	}{
// 		{"fred", 0, "fred"},
// 		{"fred", 1, "fred"},
// 		{"fred", 2, "freds"},
// 	}

// 	for _, u := range uu {
// 		assert.Equal(t, u.e, pluralOf(u.n, u.count))
// 	}
// }

// func TestToPerc(t *testing.T) {
// 	uu := []struct {
// 		v1, v2, e int64
// 	}{
// 		{50, 100, 50},
// 		{100, 0, 0},
// 		{100, 50, 200},
// 	}

// 	for _, u := range uu {
// 		assert.Equal(t, u.e, ToPerc(u.v1, u.v2))
// 	}
// }

// func TestIn(t *testing.T) {
// 	uu := []struct {
// 		l []string
// 		s string
// 		e bool
// 	}{
// 		{[]string{"a", "b", "c"}, "a", true},
// 		{[]string{"a", "b", "c"}, "z", false},
// 	}

// 	for _, u := range uu {
// 		assert.Equal(t, u.e, in(u.l, u.s))
// 	}
// }

// func TestToMCRatio(t *testing.T) {
// 	uu := []struct {
// 		q1, q2 resource.Quantity
// 		r      float64
// 	}{
// 		{test.ToQty("100m"), test.ToQty("200m"), 50},
// 		{test.ToQty("100m"), test.ToQty("50m"), 200},
// 		{test.ToQty("0m"), test.ToQty("5m"), 0},
// 		{test.ToQty("10m"), test.ToQty("0m"), 0},
// 	}

// 	for _, u := range uu {
// 		assert.Equal(t, u.r, toMCRatio(u.q1, u.q2))
// 	}
// }

// func TestToMEMRatio(t *testing.T) {
// 	uu := []struct {
// 		q1, q2 resource.Quantity
// 		r      float64
// 	}{
// 		{test.ToQty("10Mi"), test.ToQty("20Mi"), 50},
// 		{test.ToQty("20Mi"), test.ToQty("10Mi"), 200},
// 		{test.ToQty("0Mi"), test.ToQty("5Mi"), 0},
// 		{test.ToQty("10Mi"), test.ToQty("0Mi"), 0},
// 	}

// 	for _, u := range uu {
// 		assert.Equal(t, u.r, toMEMRatio(u.q1, u.q2))
// 	}
// }

// func TestContainerResources(t *testing.T) {
// 	uu := map[string]struct {
// 		co       v1.Container
// 		cpu, mem *resource.Quantity
// 		qos      qos
// 	}{
// 		"none": {
// 			co: makeContainer("c1", coOpts{
// 				image: "fred:1.0.1",
// 			}),
// 			qos: qosBestEffort,
// 		},
// 		"guaranteed": {
// 			co: makeContainer("c1", coOpts{
// 				image: "fred:1.0.1",
// 				rcpu:  "100m",
// 				rmem:  "10Mi",
// 				lcpu:  "100m",
// 				lmem:  "10Mi",
// 			}),
// 			cpu: makeQty("100m"),
// 			mem: makeQty("10Mi"),
// 			qos: qosGuaranteed,
// 		},
// 		"bustableLimit": {
// 			co: makeContainer("c1", coOpts{
// 				image: "fred:1.0.1",
// 				lcpu:  "100m",
// 				lmem:  "10Mi",
// 			}),
// 			cpu: makeQty("100m"),
// 			mem: makeQty("10Mi"),
// 			qos: qosBurstable,
// 		},
// 		"burstableRequest": {
// 			co: makeContainer("c1", coOpts{
// 				image: "fred:1.0.1",
// 				rcpu:  "100m",
// 				rmem:  "10Mi",
// 			}),
// 			cpu: makeQty("100m"),
// 			mem: makeQty("10Mi"),
// 			qos: qosBurstable,
// 		},
// 	}

// 	for k := range uu {
// 		u := uu[k]
// 		t.Run(k, func(t *testing.T) {
// 			cpu, mem, qos := containerResources(u.co)

// 			assert.Equal(t, cpu, u.cpu)
// 			assert.Equal(t, mem, u.mem)
// 			assert.Equal(t, u.qos, qos)
// 		})
// 	}
// }

// func TestPortAsString(t *testing.T) {
// 	uu := []struct {
// 		port v1.ServicePort
// 		e    string
// 	}{
// 		{v1.ServicePort{Protocol: v1.ProtocolTCP, Name: "p1", Port: 80}, "TCP:p1:80"},
// 		{v1.ServicePort{Protocol: v1.ProtocolUDP, Name: "", Port: 80}, "UDP::80"},
// 	}

// 	for _, u := range uu {
// 		assert.Equal(t, u.e, portAsStr(u.port))
// 	}
// }

// // ----------------------------------------------------------------------------
// // Helpers...

// func test.ToQty(s string) resource.Quantity {
// 	q, _ := resource.ParseQuantity(s)

// 	return q
// }
