// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Popeye

package rules

import (
	"regexp/syntax"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_rxMatch(t *testing.T) {
	uu := map[string]struct {
		exp, name string
		e         bool
		err       error
	}{
		"empty": {},
		"match-exact": {
			exp:  "rx:blee",
			name: "blee",
			e:    true,
		},
		"exclude-all": {
			exp:  "rx:.*",
			name: "blee",
			e:    true,
		},
		"match-dash": {
			exp:  "rx:blee",
			name: "blee-aeou",
			e:    true,
		},
		"no-match-dash": {
			exp:  "rx:blee-",
			name: "blee1",
		},
		"no-match-dash-wild": {
			exp:  "rx:fred1*-blee",
			name: "fred1blee",
		},
		"match-dash-wild": {
			exp:  "rx:fred-*",
			name: "fred-1blee",
			e:    true,
		},
		"match-ns": {
			exp:  `rx:default\/\w+\.v1`,
			name: "default/cm.v1",
			e:    true,
		},
		"match-ns1": {
			exp:  `rx:\.v\d+`,
			name: "default/cm.v1",
			e:    true,
		},
		"match-ns2": {
			exp:  `rx:\.v\d+`,
			name: "default/cm.v2",
			e:    true,
		},
		"match-ns3": {
			exp:  `rx:\.v\d+`,
			name: "fred/cm.v2",
			e:    true,
		},
		"match-slash": {
			exp:  "rx:kube*",
			name: "kube-system/eks-certificates-controller",
			e:    true,
		},
		"wild-version": {
			exp:  `rx:fred\.v\d+`,
			name: "kube-system/fred.v23",
			e:    true,
		},
		"wild-version-1": {
			exp:  `rx:fred.+\.v\d+`,
			name: "kube-system/fredblee.v23",
			e:    true,
		},
		"wild_version-2": {
			exp:  `rx:fred.+\.v\d+`,
			name: "kube-system/fredblee.v2.3",
			e:    true,
		},
		"starts-with": {
			exp:  `rx:\Ans`,
			name: "ns-1",
			e:    true,
		},
		"no-match-starts-with": {
			exp:  `rx:\Ans`,
			name: "ans-1",
		},
		"toast-rx": {
			exp:  `rx:\yns`,
			name: "ans-1",
			err:  &syntax.Error{Code: "invalid escape sequence", Expr: "\\y"},
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			ok, err := rxMatch(u.exp, u.name)
			assert.Equal(t, u.err, err)
			assert.Equal(t, u.e, ok)
		})
	}
}
