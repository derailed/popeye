package issues

import (
	"testing"

	"github.com/derailed/popeye/pkg/config"
	"github.com/stretchr/testify/assert"
)

func TestIsSubIssues(t *testing.T) {
	uu := map[string]struct {
		i Issue
		e bool
	}{
		"root":  {New(Root, config.WarnLevel, "blah"), false},
		"rootf": {Newf(Root, config.WarnLevel, "blah %s", "blee"), false},
		"sub":   {New("sub1", config.WarnLevel, "blah"), true},
		"subf":  {Newf("sub1", config.WarnLevel, "blah %s", "blee"), true},
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
		"notBlank": {New(Root, config.WarnLevel, "blah"), false},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, u.i.Blank())
		})
	}
}
