// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Popeye

package issues

import (
	"testing"

	"github.com/derailed/popeye/internal/rules"
	"github.com/derailed/popeye/types"
	"github.com/stretchr/testify/assert"
)

func TestIsSubIssues(t *testing.T) {
	uu := map[string]struct {
		i Issue
		e bool
	}{
		"root":  {New(types.NewGVR("fred"), Root, rules.WarnLevel, "blah"), false},
		"rootf": {Newf(types.NewGVR("fred"), Root, rules.WarnLevel, "blah %s", "blee"), false},
		"sub":   {New(types.NewGVR("fred"), "sub1", rules.WarnLevel, "blah"), true},
		"subf":  {Newf(types.NewGVR("fred"), "sub1", rules.WarnLevel, "blah %s", "blee"), true},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, u.i.IsSubIssue())
		})
	}
}

func TestBlank(t *testing.T) {
	uu := map[string]struct {
		i Issue
		e bool
	}{
		"blank":    {Issue{}, true},
		"notBlank": {New(types.NewGVR("fred"), Root, rules.WarnLevel, "blah"), false},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, u.i.Blank())
		})
	}
}
