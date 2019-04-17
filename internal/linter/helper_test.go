package linter

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/api/resource"
)

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

func TestToRatio(t *testing.T) {
	uu := []struct {
		q1, q2 resource.Quantity
		r      int64
	}{
		{toQty("10Mi"), toQty("20Mi"), 50},
		{toQty("500m"), toQty("5"), 10},
	}

	for _, u := range uu {
		assert.Equal(t, u.r, toRatio(u.q1, u.q2))
	}
}

func toQty(s string) resource.Quantity {
	q, _ := resource.ParseQuantity(s)

	return q
}
