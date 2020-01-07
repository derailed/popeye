package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRxMatch(t *testing.T) {
	uu := map[string]struct {
		exp, name string
		e         bool
	}{
		"match": {
			exp:  "rx:blee",
			name: "blee",
			e:    true,
		},
		"match_dash": {
			exp:  "rx:blee",
			name: "blee-aeou",
			e:    true,
		},
		"no_match_dash": {
			exp:  "rx:blee-",
			name: "blee1",
		},
		"no_match_dash_wild": {
			exp:  "rx:fred1*-blee",
			name: "fred1blee",
		},
		"match_dash_wild": {
			exp:  "rx:fred-*",
			name: "fred-1blee",
			e:    true,
		},
		"match_ns": {
			exp:  `rx:default\/\w+\.v1`,
			name: "default/cm.v1",
			e:    true,
		},
		"match_ns1": {
			exp:  `rx:\.v\d+`,
			name: "default/cm.v1",
			e:    true,
		},
		"match_ns2": {
			exp:  `rx:\.v\d+`,
			name: "default/cm.v2",
			e:    true,
		},
		"match_ns3": {
			exp:  `rx:\.v\d+`,
			name: "fred/cm.v2",
			e:    true,
		},

		"match_slash": {
			exp:  "rx:kube*",
			name: "kube-system/eks-certificates-controller",
			e:    true,
		},
		"wild_version": {
			exp:  `rx:fred\.v\d+`,
			name: "kube-system/fred.v23",
			e:    true,
		},
		"wild_version_1": {
			exp:  `rx:fred.+\.v\d+`,
			name: "kube-system/fredblee.v23",
			e:    true,
		},
		"wild_version_2": {
			exp:  `rx:fred.+\.v\d+`,
			name: "kube-system/fredblee.v2.3",
			e:    true,
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, rxMatch(u.exp, u.name))
		})
	}
}
