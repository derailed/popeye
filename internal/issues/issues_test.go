// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Popeye

package issues

import (
	"testing"

	"github.com/derailed/popeye/internal/rules"
	"github.com/derailed/popeye/types"
	"github.com/stretchr/testify/assert"
)

func TestMaxGroupSeverity(t *testing.T) {
	o := Outcome{
		"s1": Issues{
			New(types.NewGVR("fred"), Root, rules.OkLevel, "i1"),
		},
		"s2": Issues{
			New(types.NewGVR("fred"), Root, rules.OkLevel, "i1"),
			New(types.NewGVR("fred"), Root, rules.WarnLevel, "i2"),
			New(types.NewGVR("fred"), "g1", rules.WarnLevel, "i2"),
		},
	}

	assert.Equal(t, rules.OkLevel, o.MaxGroupSeverity("s1", Root))
	assert.Equal(t, rules.WarnLevel, o.MaxGroupSeverity("s2", Root))
}

func TestIssuesForGroup(t *testing.T) {
	o := Outcome{
		"s1": Issues{
			New(types.NewGVR("fred"), Root, rules.OkLevel, "i1"),
		},
		"s2": Issues{
			New(types.NewGVR("fred"), Root, rules.OkLevel, "i1"),
			New(types.NewGVR("fred"), Root, rules.WarnLevel, "i2"),
			New(types.NewGVR("fred"), "g1", rules.WarnLevel, "i3"),
			New(types.NewGVR("fred"), "g1", rules.WarnLevel, "i4"),
		},
	}

	assert.Equal(t, 1, len(o.For("s1", Root)))
	assert.Equal(t, 2, len(o.For("s2", "g1")))
}

func TestGroup(t *testing.T) {
	o := Outcome{
		"s2": Issues{
			New(types.NewGVR("fred"), Root, rules.OkLevel, "i1"),
			New(types.NewGVR("fred"), Root, rules.WarnLevel, "i2"),
			New(types.NewGVR("fred"), "g1", rules.ErrorLevel, "i2"),
		},
	}

	grp := o["s2"].Group()
	assert.Equal(t, rules.ErrorLevel, o["s2"].MaxSeverity())
	assert.Equal(t, 2, len(grp))
}
