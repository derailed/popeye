package issues

import (
	"testing"

	"github.com/derailed/popeye/pkg/config"
	"github.com/stretchr/testify/assert"
)

func TestMaxGroupSeverity(t *testing.T) {
	o := Outcome{
		"s1": Issues{
			New(Root, config.OkLevel, "i1"),
		},
		"s2": Issues{
			New(Root, config.OkLevel, "i1"),
			New(Root, config.WarnLevel, "i2"),
			New("g1", config.WarnLevel, "i2"),
		},
	}

	assert.Equal(t, config.OkLevel, o.MaxGroupSeverity("s1", Root))
	assert.Equal(t, config.WarnLevel, o.MaxGroupSeverity("s2", Root))
}

func TestIssuesForGroup(t *testing.T) {
	o := Outcome{
		"s1": Issues{
			New(Root, config.OkLevel, "i1"),
		},
		"s2": Issues{
			New(Root, config.OkLevel, "i1"),
			New(Root, config.WarnLevel, "i2"),
			New("g1", config.WarnLevel, "i3"),
			New("g1", config.WarnLevel, "i4"),
		},
	}

	assert.Equal(t, 1, len(o.For("s1", Root)))
	assert.Equal(t, 2, len(o.For("s2", "g1")))
}

func TestGroup(t *testing.T) {
	o := Outcome{
		"s2": Issues{
			New(Root, config.OkLevel, "i1"),
			New(Root, config.WarnLevel, "i2"),
			New("g1", config.ErrorLevel, "i2"),
		},
	}

	grp := o["s2"].Group()
	assert.Equal(t, config.ErrorLevel, o["s2"].MaxSeverity())
	assert.Equal(t, 2, len(grp))
}
