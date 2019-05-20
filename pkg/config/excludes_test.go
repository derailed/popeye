package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExcludes(t *testing.T) {
	uu := map[string]struct {
		n string
		e bool
	}{
		"excluded": {"a", true},
		"included": {"d", false},
	}

	for k, u := range uu {
		ex := Excludes{"a", "b", "c"}
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, ex.excluded(u.n))
		})
	}
}
