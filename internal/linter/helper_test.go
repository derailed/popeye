package linter

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToPerc(t *testing.T) {
	uu := []struct {
		v1, v2, e float64
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
