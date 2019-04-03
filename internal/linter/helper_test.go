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
		assert.Equal(t, u.e, toPerc(u.v1, u.v2))
	}
}
