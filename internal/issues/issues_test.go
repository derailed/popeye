package issues

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMaxGroupSeverity(t *testing.T) {
	o := Outcome{
		"s1": Issues{
			New(Root, OkLevel, "i1"),
		},
		"s2": Issues{
			New(Root, OkLevel, "i1"),
			New(Root, WarnLevel, "i2"),
			New("g1", WarnLevel, "i2"),
		},
	}

	assert.Equal(t, OkLevel, o.MaxGroupSeverity("s1", Root))
	assert.Equal(t, WarnLevel, o.MaxGroupSeverity("s2", Root))
}

func TestIssuesForGroup(t *testing.T) {
	o := Outcome{
		"s1": Issues{
			New(Root, OkLevel, "i1"),
		},
		"s2": Issues{
			New(Root, OkLevel, "i1"),
			New(Root, WarnLevel, "i2"),
			New("g1", WarnLevel, "i3"),
			New("g1", WarnLevel, "i4"),
		},
	}

	assert.Equal(t, 1, len(o.For("s1", Root)))
	assert.Equal(t, 2, len(o.For("s2", "g1")))
}

func TestGroup(t *testing.T) {
	o := Outcome{
		"s2": Issues{
			New(Root, OkLevel, "i1"),
			New(Root, WarnLevel, "i2"),
			New("g1", ErrorLevel, "i2"),
		},
	}

	grp := o["s2"].Group()
	assert.Equal(t, ErrorLevel, o["s2"].MaxSeverity())
	assert.Equal(t, 2, len(grp))
}
