// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Popeye

package rules_test

import (
	"testing"

	"github.com/derailed/popeye/internal/rules"
	"github.com/stretchr/testify/assert"
)

func TestCodeFormat(t *testing.T) {
	uu := map[string]struct {
		c  rules.Code
		aa []any
		e  string
	}{
		"empty": {
			e: "[POP-100]",
		},
		"no-args": {
			c: rules.Code{Message: "bla"},
			e: "[POP-100] bla",
		},
		"args": {
			c:  rules.Code{Message: "bla %s %d"},
			aa: []any{"yo", 10},
			e:  "[POP-100] bla yo 10",
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, u.c.Format(100, u.aa...))
		})
	}
}
