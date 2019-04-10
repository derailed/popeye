package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExcludes(t *testing.T) {
	uu := []struct {
		n string
		e bool
	}{
		{"a", true},
		{"d", false},
	}

	for _, u := range uu {
		ex := Excludes{"a", "b", "c"}
		assert.Equal(t, u.e, ex.excluded(u.n))
	}
}
