package cache

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetHas(t *testing.T) {
	ss := StringSet{"a": Blank}

	assert.False(t, ss.Has("b"))
	assert.True(t, ss.Has("a"))
}

func TestSetAdd(t *testing.T) {
	uu := []struct {
		ss []string
		e  StringSet
	}{
		{
			[]string{"a"},
			StringSet{
				"a": Blank,
			},
		},
		{
			[]string{"a", "b", "c", "c"},
			StringSet{
				"a": Blank,
				"b": Blank,
				"c": Blank,
			},
		},
		{
			[]string{"a", "a", "a", "a"},
			StringSet{
				"a": Blank,
			},
		},
	}

	for _, u := range uu {
		ss := StringSet{}
		ss.Add(u.ss...)

		assert.Equal(t, u.e, ss)
	}
}

func TestSetDiff(t *testing.T) {
	uu := []struct {
		s1, s2, e StringSet
	}{
		{
			StringSet{"a": Blank, "b": Blank},
			StringSet{"a": Blank},
			StringSet{},
		},
		{
			StringSet{"a": Blank, "b": Blank},
			StringSet{"a": Blank, "c": Blank},
			StringSet{"c": Blank},
		},
	}

	for _, u := range uu {
		ss := u.s1.Diff(u.s2)

		assert.Equal(t, u.e, ss)
	}
}
