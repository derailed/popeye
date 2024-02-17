// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Popeye

package tally_test

import (
	"testing"

	"github.com/derailed/popeye/internal/issues/tally"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func init() {
	zerolog.SetGlobalLevel(zerolog.FatalLevel)
}

func TestLinterCompact(t *testing.T) {
	uu := map[string]struct {
		lt, e tally.Linter
	}{
		"empty": {},
		"multi": {
			lt: tally.Linter{
				"a": tally.Namespace{
					"ns1": tally.Code{
						"100": 0,
						"101": 2,
						"102": 5,
						"103": 0,
					},
					"ns2": tally.Code{
						"100": 1,
						"101": 0,
						"102": 0,
						"103": 6,
					},
				},
				"b": tally.Namespace{
					"ns1": tally.Code{
						"100": 0,
						"101": 2,
						"102": 5,
						"103": 0,
					},
					"ns3": tally.Code{
						"100": 1,
						"101": 0,
						"102": 0,
						"103": 6,
					},
				},
			},
			e: tally.Linter{
				"a": tally.Namespace{
					"ns1": tally.Code{
						"101": 2,
						"102": 5,
					},
					"ns2": tally.Code{
						"100": 1,
						"103": 6,
					},
				},
				"b": tally.Namespace{
					"ns1": tally.Code{
						"101": 2,
						"102": 5,
					},
					"ns3": tally.Code{
						"100": 1,
						"103": 6,
					},
				},
			},
		},
		"delete-ns": {
			lt: tally.Linter{
				"a": tally.Namespace{
					"ns1": tally.Code{
						"100": 0,
						"101": 0,
						"102": 0,
						"103": 0,
					},
					"ns2": tally.Code{
						"100": 1,
						"101": 0,
						"102": 0,
						"103": 6,
					},
				},
				"b": tally.Namespace{
					"ns1": tally.Code{
						"100": 0,
						"101": 0,
						"102": 0,
						"103": 0,
					},
					"ns3": tally.Code{
						"100": 1,
						"101": 0,
						"102": 0,
						"103": 6,
					},
				},
			},
			e: tally.Linter{
				"a": tally.Namespace{
					"ns2": tally.Code{
						"100": 1,
						"103": 6,
					},
				},
				"b": tally.Namespace{
					"ns3": tally.Code{
						"100": 1,
						"103": 6,
					},
				},
			},
		},
		"delete-linter": {
			lt: tally.Linter{
				"a": tally.Namespace{
					"ns1": tally.Code{
						"100": 0,
						"101": 0,
						"102": 0,
						"103": 0,
					},
					"ns2": tally.Code{
						"100": 0,
						"101": 0,
						"102": 0,
						"103": 0,
					},
				},
				"b": tally.Namespace{
					"ns1": tally.Code{
						"100": 0,
						"101": 0,
						"102": 0,
						"103": 0,
					},
					"ns3": tally.Code{
						"100": 1,
						"101": 0,
						"102": 0,
						"103": 6,
					},
				},
			},
			e: tally.Linter{
				"b": tally.Namespace{
					"ns3": tally.Code{
						"100": 1,
						"103": 6,
					},
				},
			},
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			u.lt.Compact()
			u.lt.Dump()
			assert.Equal(t, u.e, u.lt)
		})
	}
}
