package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExcludes(t *testing.T) {
	uu := map[string]struct {
		s Excludes
		n string
		e bool
	}{
		"empty":            {Excludes{"aa", "bb", "cc"}, "", false},
		"stringMatch":      {Excludes{"aa", "bb", "cc"}, "aa", true},
		"stringNoMatch":    {Excludes{"aa", "bb", "cc"}, "dd", false},
		"rxBleeMatch":      {Excludes{"rx:blee"}, "blee", true},
		"rxBleeNoMatch":    {Excludes{"rx:blee"}, "ble", false},
		"rxFred-Match":     {Excludes{"rx:blee", "rx:fred-*"}, "fred-aeuo", true},
		"rxFredNoMatch":    {Excludes{"rx:blee*", "rx:fred-"}, "fred", false},
		"fredbleeNoMatch":  {Excludes{"rx:blee", "duh", "rx:fred1*-blee"}, "fredblee", false},
		"fred1bleeMatch":   {Excludes{"blee*", "duh", "rx:fred1.*blee"}, "fred1duhblee", true},
		"fred1-duh-bleeNM": {Excludes{"blee*", "duh", "fred", `fred1\-*`}, "fred1-duh-blee", false},
		"fred1-duh-bleeM":  {Excludes{"rx:blee*", "duh", "fred", "rx:fred1-*"}, "fred1-duh-blee", true},
		"d":                {Excludes{"a", "b", "c"}, "d", false},
		"p-":               {Excludes{"rx:p-"}, "p-1", true},
		"p1":               {Excludes{"rx:p-"}, "p1", false},
		"p-1":              {Excludes{"rx:p-.*"}, "p-1", true},
	}

	for k, u := range uu {
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, u.s.excluded(u.n))
		})
	}
}
