// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Popeye

package tally_test

import (
	"testing"

	"github.com/derailed/popeye/internal/issues/tally"
	"github.com/stretchr/testify/assert"
)

func TestNSMerge(t *testing.T) {
	uu := map[string]struct {
		ns1, ns2, e tally.Namespace
	}{
		"empty": {},
		"one-way": {
			ns1: tally.Namespace{
				"ns1": tally.Code{
					"100": 1,
					"101": 2,
					"102": 5,
					"103": 6,
				},
			},
			ns2: tally.Namespace{},
			e: tally.Namespace{
				"ns1": tally.Code{
					"100": 1,
					"101": 2,
					"102": 5,
					"103": 6,
				},
			},
		},
		"union": {
			ns1: tally.Namespace{
				"ns1": tally.Code{
					"100": 1,
					"101": 2,
					"102": 5,
					"103": 6,
				},
			},
			ns2: tally.Namespace{
				"ns2": tally.Code{
					"100": 1,
					"101": 2,
					"102": 5,
					"103": 6,
				},
			},
			e: tally.Namespace{
				"ns1": tally.Code{
					"100": 1,
					"101": 2,
					"102": 5,
					"103": 6,
				},
				"ns2": tally.Code{
					"100": 1,
					"101": 2,
					"102": 5,
					"103": 6,
				},
			},
		},
		"intersect": {
			ns1: tally.Namespace{
				"ns1": tally.Code{
					"100": 1,
					"101": 2,
					"102": 5,
					"103": 6,
				},
			},
			ns2: tally.Namespace{
				"ns1": tally.Code{
					"100": 1,
					"101": 2,
					"102": 5,
					"103": 6,
				},
				"ns2": tally.Code{
					"100": 1,
					"101": 2,
					"102": 5,
					"103": 6,
				},
			},
			e: tally.Namespace{
				"ns1": tally.Code{
					"100": 2,
					"101": 4,
					"102": 10,
					"103": 12,
				},
				"ns2": tally.Code{
					"100": 1,
					"101": 2,
					"102": 5,
					"103": 6,
				},
			},
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			u.ns1.Merge(u.ns2)
			assert.Equal(t, u.e, u.ns1)
		})
	}
}

func TestNSCompact(t *testing.T) {
	uu := map[string]struct {
		ns1, e tally.Namespace
	}{
		"empty": {},
		"multi": {
			ns1: tally.Namespace{
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
			e: tally.Namespace{
				"ns1": tally.Code{
					"101": 2,
					"102": 5,
				},
				"ns2": tally.Code{
					"100": 1,
					"103": 6,
				},
			},
		},
		"single": {
			ns1: tally.Namespace{
				"ns1": tally.Code{
					"100": 1,
					"101": 0,
					"102": 5,
					"103": 6,
				},
			},
			e: tally.Namespace{
				"ns1": tally.Code{
					"100": 1,
					"102": 5,
					"103": 6,
				},
			},
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			u.ns1.Compact()
			assert.Equal(t, u.e, u.ns1)
		})
	}
}
