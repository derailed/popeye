package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExcludes(t *testing.T) {
	uu := map[string]struct {
		s Exclude
		n string
		e bool
	}{
		"empty":            {Exclude{"aa", "bb", "cc"}, "", false},
		"stringMatch":      {Exclude{"aa", "bb", "cc"}, "aa", true},
		"stringNoMatch":    {Exclude{"aa", "bb", "cc"}, "dd", false},
		"rxBleeMatch":      {Exclude{"rx:blee"}, "blee", true},
		"rxBleeNoMatch":    {Exclude{"rx:blee"}, "ble", false},
		"rxFred-Match":     {Exclude{"rx:blee", "rx:fred-*"}, "fred-aeuo", true},
		"rxFredNoMatch":    {Exclude{"rx:blee*", "rx:fred-"}, "fred", false},
		"fredbleeNoMatch":  {Exclude{"rx:blee", "duh", "rx:fred1*-blee"}, "fredblee", false},
		"fred1bleeMatch":   {Exclude{"blee*", "duh", "rx:fred1.*blee"}, "fred1duhblee", true},
		"fred1-duh-bleeNM": {Exclude{"blee*", "duh", "fred", `fred1\-*`}, "fred1-duh-blee", false},
		"fred1-duh-bleeM":  {Exclude{"rx:blee*", "duh", "fred", "rx:fred1-*"}, "fred1-duh-blee", true},
		"d":                {Exclude{"a", "b", "c"}, "d", false},
		"p-":               {Exclude{"rx:p-"}, "p-1", true},
		"p1":               {Exclude{"rx:p-"}, "p1", false},
		"p-1":              {Exclude{"rx:p-.*"}, "p-1", true},
		"toast":            {Exclude{`rx:($x`}, "blee", false},
		"cm1":              {Exclude{`rx:default\/\w+\.v1`}, "default/cm.v1", true},
		"cm2":              {Exclude{`rx:*\.v\d+`}, "default/cm.v1", true},
		"cm3":              {Exclude{`rx:*\.v\d+`}, "default/cm.v2", true},
		"cmNS1":            {Exclude{`rx:*\.v\d+`}, "fred/cm.v2", true},
		"cmNS2":            {Exclude{`rx:fred*\.v\d+`}, "default/cm.v2", false},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, u.s.ShouldExclude(u.n))
		})
	}
}
