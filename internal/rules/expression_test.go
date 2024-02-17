// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Popeye

package rules

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_expressionMatch(t *testing.T) {
	uu := map[string]struct {
		exp Expression
		s   string
		e   bool
	}{
		"empty": {
			e: true,
		},
		"empty-rule": {
			s: "fred",
			e: true,
		},
		"happy": {
			exp: "fred",
			s:   "fred",
			e:   true,
		},
		"happy-rx": {
			exp: "rx:^fred",
			s:   "fred",
			e:   true,
		},
		"toast": {
			exp: "freddy",
			s:   "fred",
		},
		"toast-rx": {
			exp: "rx:freddy",
			s:   "fred",
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			ok := u.exp.match(u.s)
			assert.Equal(t, u.e, ok)
		})
	}
}
