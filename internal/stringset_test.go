// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Popeye

package internal_test

import (
	"testing"

	"github.com/derailed/popeye/internal"
	"github.com/stretchr/testify/assert"
)

func TestSetHas(t *testing.T) {
	ss := internal.StringSet{"a": internal.Blank}

	assert.False(t, ss.Has("b"))
	assert.True(t, ss.Has("a"))
}

func TestSetAdd(t *testing.T) {
	uu := []struct {
		ss []string
		e  internal.StringSet
	}{
		{
			[]string{"a"},
			internal.StringSet{
				"a": internal.Blank,
			},
		},
		{
			[]string{"a", "b", "c", "c"},
			internal.StringSet{
				"a": internal.Blank,
				"b": internal.Blank,
				"c": internal.Blank,
			},
		},
		{
			[]string{"a", "a", "a", "a"},
			internal.StringSet{
				"a": internal.Blank,
			},
		},
	}

	for _, u := range uu {
		ss := internal.StringSet{}
		ss.Add(u.ss...)

		assert.Equal(t, u.e, ss)
	}
}

func TestSetDiff(t *testing.T) {
	uu := []struct {
		s1, s2, e internal.StringSet
	}{
		{
			internal.StringSet{"a": internal.Blank, "b": internal.Blank},
			internal.StringSet{"a": internal.Blank},
			internal.StringSet{},
		},
		{
			internal.StringSet{"a": internal.Blank, "b": internal.Blank},
			internal.StringSet{"a": internal.Blank, "c": internal.Blank},
			internal.StringSet{"c": internal.Blank},
		},
	}

	for _, u := range uu {
		ss := u.s1.Diff(u.s2)

		assert.Equal(t, u.e, ss)
	}
}
