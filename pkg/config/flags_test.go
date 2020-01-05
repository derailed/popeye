package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOutputFormat(t *testing.T) {
	uu := map[string]struct {
		f Flags
		e string
	}{
		"standard": {Flags{Output: strPtr("standard")}, "standard"},
		"blank":    {Flags{Output: strPtr("")}, "cool"},
		"nil":      {Flags{}, "cool"},
		"blee":     {Flags{Output: strPtr("blee")}, "blee"},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, u.f.OutputFormat())
		})
	}
}
