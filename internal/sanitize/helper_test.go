package sanitize

import (
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func TestNamepaced(t *testing.T) {
	uu := []struct {
		s     string
		ns, n string
	}{
		{"fred/blee", "fred", "blee"},
		{"fred", "", "fred"},
	}

	for _, u := range uu {
		ns, n := namespaced(u.s)
		assert.Equal(t, u.ns, ns)
		assert.Equal(t, u.n, n)
	}
}

func TestPluralOf(t *testing.T) {
	uu := []struct {
		n     string
		count int
		e     string
	}{
		{"fred", 0, "fred"},
		{"fred", 1, "fred"},
		{"fred", 2, "freds"},
	}

	for _, u := range uu {
		assert.Equal(t, u.e, pluralOf(u.n, u.count))
	}
}

func TestToPerc(t *testing.T) {
	uu := []struct {
		v1, v2, e int64
	}{
		{50, 100, 50},
		{100, 0, 0},
		{100, 50, 200},
	}

	for _, u := range uu {
		assert.Equal(t, u.e, ToPerc(u.v1, u.v2))
	}
}

func TestIn(t *testing.T) {
	uu := []struct {
		l []string
		s string
		e bool
	}{
		{[]string{"a", "b", "c"}, "a", true},
		{[]string{"a", "b", "c"}, "z", false},
	}

	for _, u := range uu {
		assert.Equal(t, u.e, in(u.l, u.s))
	}
}

func TestToMCRatio(t *testing.T) {
	uu := []struct {
		q1, q2 resource.Quantity
		r      int64
	}{
		{toQty("100m"), toQty("200m"), 50},
		{toQty("100m"), toQty("50m"), 200},
		{toQty("0m"), toQty("5m"), 0},
		{toQty("10m"), toQty("0m"), 0},
	}

	for _, u := range uu {
		assert.Equal(t, u.r, toMCRatio(u.q1, u.q2))
	}
}

func TestToMEMRatio(t *testing.T) {
	uu := []struct {
		q1, q2 resource.Quantity
		r      int64
	}{
		{toQty("10Mi"), toQty("20Mi"), 50},
		{toQty("20Mi"), toQty("10Mi"), 200},
		{toQty("0Mi"), toQty("5Mi"), 0},
		{toQty("10Mi"), toQty("0Mi"), 0},
	}

	for _, u := range uu {
		assert.Equal(t, u.r, toMEMRatio(u.q1, u.q2))
	}
}

func TestContainerResources(t *testing.T) {
	uu := []struct {
		co    v1.Container
		res   v1.ResourceList
		burst bool
	}{
		{
			makeContainer("c1", coOpts{
				image: "fred:1.0.1",
				rcpu:  "100m",
				rmem:  "10Mi",
				lcpu:  "100m",
				lmem:  "10Mi",
			}),
			makeRes("100m", "10Mi"),
			true,
		},
		{
			makeContainer("c1", coOpts{
				image: "fred:1.0.1",
				lcpu:  "100m",
				lmem:  "10Mi",
			}),
			makeRes("100m", "10Mi"),
			false,
		},
		{
			makeContainer("c1", coOpts{
				image: "fred:1.0.1",
				rcpu:  "100m",
				rmem:  "10Mi",
			}),
			makeRes("100m", "10Mi"),
			false,
		},
	}

	for _, u := range uu {
		cpu, mem, burst := containerResources(u.co)
		assert.Equal(t, cpu, u.res.Cpu())
		assert.Equal(t, mem, u.res.Memory())
		assert.Equal(t, u.burst, burst)
	}
}

func TestPortAsString(t *testing.T) {
	uu := []struct {
		port v1.ServicePort
		e    string
	}{
		{v1.ServicePort{Protocol: v1.ProtocolTCP, Name: "p1", Port: 80}, "TCP:p1:80"},
		{v1.ServicePort{Protocol: v1.ProtocolUDP, Name: "", Port: 80}, "UDP::80"},
	}

	for _, u := range uu {
		assert.Equal(t, u.e, portAsStr(u.port))
	}
}

// ----------------------------------------------------------------------------
// Helpers...

func toQty(s string) resource.Quantity {
	q, _ := resource.ParseQuantity(s)

	return q
}
