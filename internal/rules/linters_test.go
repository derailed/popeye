// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Popeye

package rules

import (
	"testing"

	"github.com/derailed/popeye/types"
	"github.com/stretchr/testify/assert"
)

func Test_lintersMatch(t *testing.T) {
	uu := map[string]struct {
		linters Linters
		spec    Spec
		glob    bool
		e       bool
	}{
		"empty": {},
		"empty-rule": {
			spec: Spec{
				GVR:  types.NewGVR("v1/pods"),
				FQN:  "ns1/p1",
				Code: 100,
			},
		},
		"missing": {
			linters: Linters{
				"pods": LinterExcludes{
					Instances: Excludes{
						{
							FQNs: expressions{"ns1", "ns2"},
						},
					},
				},
			},
			spec: Spec{
				GVR:  types.NewGVR("v1/configmaps"),
				FQN:  "ns1/p1",
				Code: 100,
			},
		},

		"happy": {
			linters: Linters{
				"pods": LinterExcludes{
					Instances: Excludes{
						{
							FQNs: expressions{"rx:^ns1", "ns2"},
						},
					},
				},
			},
			spec: Spec{
				GVR:  types.NewGVR("v1/pods"),
				FQN:  "ns1/p1",
				Code: 100,
			},
			e: true,
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, u.linters.Match(u.spec, u.glob))
		})
	}
}
