// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Popeye

package tally_test

import (
	"testing"

	"github.com/derailed/popeye/internal/issues/tally"
	"github.com/derailed/popeye/internal/rules"
	"github.com/stretchr/testify/assert"
)

func TestCodeMerge(t *testing.T) {
	uu := map[string]struct {
		c1, c2, e tally.Code
	}{
		"empty": {},
		"empty-1": {
			c1: tally.Code{},
			c2: tally.Code{
				"100": 1,
				"101": 2,
				"102": 5,
				"103": 6,
			},
			e: tally.Code{
				"100": 1,
				"101": 2,
				"102": 5,
				"103": 6,
			},
		},
		"empty-2": {
			c1: tally.Code{
				"100": 1,
				"101": 2,
				"102": 5,
				"103": 6,
			},
			c2: tally.Code{},
			e: tally.Code{
				"100": 1,
				"101": 2,
				"102": 5,
				"103": 6,
			},
		},

		"same": {
			c1: tally.Code{
				"100": 1,
				"101": 2,
				"102": 5,
				"103": 6,
			},
			c2: tally.Code{
				"100": 1,
				"101": 2,
				"102": 5,
				"103": 6,
			},
			e: tally.Code{
				"100": 2,
				"101": 4,
				"102": 10,
				"103": 12,
			},
		},
		"delta": {
			c1: tally.Code{
				"100": 1,
				"101": 2,
				"102": 5,
				"103": 6,
			},
			c2: tally.Code{
				"102": 5,
				"200": 1,
				"201": 2,
				"203": 6,
			},
			e: tally.Code{
				"100": 1,
				"101": 2,
				"102": 10,
				"103": 6,
				"200": 1,
				"201": 2,
				"203": 6,
			},
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			u.c1.Merge(u.c2)
			assert.Equal(t, u.e, u.c1)
		})
	}
}

func TestCodeCompact(t *testing.T) {
	uu := map[string]struct {
		c, e tally.Code
	}{
		"empty": {},
		"none": {
			c: tally.Code{
				"100": 1,
				"101": 2,
				"102": 5,
				"103": 6,
			},
			e: tally.Code{
				"100": 1,
				"101": 2,
				"102": 5,
				"103": 6,
			},
		},
		"happy": {
			c: tally.Code{
				"100": 1,
				"101": 2,
				"200": 0,
				"201": 6,
				"202": 0,
				"203": 0,
			},
			e: tally.Code{
				"100": 1,
				"101": 2,
				"201": 6,
			},
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			u.c.Compact()
			assert.Equal(t, u.e, u.c)
		})
	}
}

func TestCodeRollup(t *testing.T) {
	uu := map[string]struct {
		c tally.Code
		e tally.SevScore
	}{
		"empty": {},
		"plain": {
			c: tally.Code{
				"100": 1,
				"101": 2,
				"102": 5,
				"103": 6,
			},
			e: tally.SevScore{
				0: 6,
				1: 5,
				2: 2,
				3: 1,
			},
		},
		"singles": {
			c: tally.Code{
				"100": 1,
				"101": 2,
				"200": 5,
				"201": 6,
				"202": 20,
				"203": 10,
			},
			e: tally.SevScore{
				0: 10,
				1: 20,
				2: 8,
				3: 6,
			},
		},
	}

	g := rules.Glossary{
		100: {
			Severity: rules.ErrorLevel,
		},
		101: {
			Severity: rules.WarnLevel,
		},
		102: {
			Severity: rules.InfoLevel,
		},
		103: {
			Severity: rules.OkLevel,
		},
		200: {
			Severity: rules.ErrorLevel,
		},
		201: {
			Severity: rules.WarnLevel,
		},
		202: {
			Severity: rules.InfoLevel,
		},
		203: {
			Severity: rules.OkLevel,
		},
	}
	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, u.c.Rollup(g))
		})
	}
}
