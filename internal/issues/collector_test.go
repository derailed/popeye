package issues

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNoConcerns(t *testing.T) {
	uu := map[string]struct {
		issues []Issue
		e      bool
	}{
		"noIssue": {
			e: true,
		},
		"issues": {
			issues: []Issue{
				New(Root, InfoLevel, "blee"),
				New(Root, WarnLevel, "blee"),
			},
		},
	}

	for k, u := range uu {
		t.Run(k, func(t *testing.T) {
			c := NewCollector()
			c.addIssue("fred", u.issues...)

			assert.Equal(t, u.e, c.NoConcerns("fred"))
		})
	}
}

func TestMaxSeverity(t *testing.T) {
	uu := map[string]struct {
		issues   []Issue
		section  string
		severity Level
		count    int
	}{
		"noIssue": {
			section:  Root,
			severity: OkLevel,
			count:    0,
		},
		"mix": {
			issues: []Issue{
				New(Root, InfoLevel, "blee"),
				New(Root, WarnLevel, "blee"),
			},
			section:  Root,
			severity: WarnLevel,
			count:    2,
		},
		"same": {
			issues: []Issue{
				New(Root, InfoLevel, "blee"),
				New(Root, InfoLevel, "blee"),
			},
			section:  Root,
			severity: InfoLevel,
			count:    2,
		},
		"error": {
			issues: []Issue{
				New(Root, ErrorLevel, "blee"),
				New(Root, InfoLevel, "blee"),
				New(Root, InfoLevel, "blee"),
			},
			section:  Root,
			severity: ErrorLevel,
			count:    3,
		},
	}

	for k, u := range uu {
		t.Run(k, func(t *testing.T) {
			c := NewCollector()
			c.addIssue(u.section, u.issues...)

			assert.Equal(t, u.count, len(c.outcomes[u.section]))
			assert.Equal(t, u.severity, c.MaxSeverity(u.section))
		})
	}
}

func TestAddErr(t *testing.T) {
	uu := map[string]struct {
		errors  []error
		section string
		count   int
	}{
		"one": {
			errors: []error{
				errors.New("blee"),
			},
			section: Root,
			count:   1,
		},
		"many": {
			errors: []error{
				errors.New("blee"),
				errors.New("duh"),
			},
			section: Root,
			count:   2,
		},
	}

	for k, u := range uu {
		t.Run(k, func(t *testing.T) {
			c := NewCollector()
			c.AddErr(u.section, u.errors...)

			assert.Equal(t, u.count, len(c.outcomes[u.section]))
			assert.Equal(t, ErrorLevel, c.MaxSeverity(u.section))
		})
	}
}

func TestAddOkIssue(t *testing.T) {
	group := "g1"
	c := NewCollector()
	c.AddOk(group, "blee")
	c.AddOkf(group, "blee %s", "duh")

	assert.Equal(t, 2, len(c.outcomes[group]))
	assert.Equal(t, "blee", c.outcomes[group][0].Message)
	assert.Equal(t, "blee duh", c.outcomes[group][1].Message)
	assert.Equal(t, OkLevel, c.MaxSeverity(group))
}

func TestAddInfoIssue(t *testing.T) {
	group := "g1"
	c := NewCollector()
	c.AddInfo(group, "blee")
	c.AddInfof(group, "blee %s", "duh")

	assert.Equal(t, 2, len(c.outcomes[group]))
	assert.Equal(t, "blee", c.outcomes[group][0].Message)
	assert.Equal(t, "blee duh", c.outcomes[group][1].Message)
	assert.Equal(t, InfoLevel, c.MaxSeverity(group))
}

func TestAddWarnIssue(t *testing.T) {
	group := "g1"
	c := NewCollector()
	c.AddWarn(group, "blee")
	c.AddWarnf(group, "blee %s", "duh")

	assert.Equal(t, 2, len(c.outcomes[group]))
	assert.Equal(t, "blee", c.outcomes[group][0].Message)
	assert.Equal(t, "blee duh", c.outcomes[group][1].Message)
	assert.Equal(t, WarnLevel, c.MaxSeverity(group))
}

func TestAddErrorIssue(t *testing.T) {
	group := "g1"
	c := NewCollector()
	c.AddError(group, "blee")
	c.AddErrorf(group, "blee %s", "duh")

	assert.Equal(t, 2, len(c.outcomes[group]))
	assert.Equal(t, "blee", c.outcomes[group][0].Message)
	assert.Equal(t, "blee duh", c.outcomes[group][1].Message)
	assert.Equal(t, ErrorLevel, c.MaxSeverity(group))
}

func TestAddSubOk(t *testing.T) {
	section, group := "s1", "g1"
	c := NewCollector()
	c.AddSubOk(section, group, "blee")
	c.AddSubOkf(section, group, "blee %s", "duh")

	ii := c.outcomes.For(section, group)
	assert.Equal(t, 2, len(ii))
	assert.Equal(t, "blee", ii[0].Message)
	assert.Equal(t, "blee duh", ii[1].Message)
	assert.Equal(t, OkLevel, c.outcomes.MaxGroupSeverity(section, group))
}

func TestAddSubInfo(t *testing.T) {
	section, group := "s1", "g1"
	c := NewCollector()
	c.AddSubInfo(section, group, "blee")
	c.AddSubInfof(section, group, "blee %s", "duh")

	ii := c.outcomes.For(section, group)
	assert.Equal(t, 2, len(ii))
	assert.Equal(t, "blee", ii[0].Message)
	assert.Equal(t, "blee duh", ii[1].Message)
	assert.Equal(t, InfoLevel, c.outcomes.MaxGroupSeverity(section, group))
}

func TestAddSubWarn(t *testing.T) {
	section, group := "s1", "g1"
	c := NewCollector()
	c.AddSubWarn(section, group, "blee")
	c.AddSubWarnf(section, group, "blee %s", "duh")

	ii := c.outcomes.For(section, group)
	assert.Equal(t, 2, len(ii))
	assert.Equal(t, "blee", ii[0].Message)
	assert.Equal(t, "blee duh", ii[1].Message)
	assert.Equal(t, WarnLevel, c.outcomes.MaxGroupSeverity(section, group))
}

func TestAddSubError(t *testing.T) {
	section, group := "s1", "g1"
	c := NewCollector()
	c.AddSubError(section, group, "blee")
	c.AddSubErrorf(section, group, "blee %s", "duh")

	ii := c.outcomes.For(section, group)
	assert.Equal(t, 2, len(ii))
	assert.Equal(t, "blee", ii[0].Message)
	assert.Equal(t, "blee duh", ii[1].Message)
	assert.Equal(t, ErrorLevel, c.outcomes.MaxGroupSeverity(section, group))
}
