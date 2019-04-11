package linter

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLinterMaxSeverity(t *testing.T) {
	uu := []struct {
		severity Level
		issues   []Issue
	}{
		{
			severity: WarnLevel,
			issues: []Issue{
				NewError(InfoLevel, "blee"),
				NewError(WarnLevel, "blee"),
			},
		},
		{
			severity: InfoLevel,
			issues: []Issue{
				NewError(InfoLevel, "blee"),
				NewError(InfoLevel, "blee"),
			},
		},
		{
			severity: ErrorLevel,
			issues: []Issue{
				NewError(ErrorLevel, "blee"),
				NewError(InfoLevel, "blee"),
				NewError(InfoLevel, "blee"),
			},
		},
	}

	for _, u := range uu {
		l := NewLinter(nil, nil)
		l.addIssues("blee", u.issues...)
		assert.Equal(t, u.severity, l.MaxSeverity("blee"))
	}
}

func TestLinterAddIssue(t *testing.T) {
	l := NewLinter(nil, nil)

	l.initIssues("fred")
	assert.True(t, l.NoIssues("fred"))

	l.addIssue("fred", InfoLevel, "blee")
	assert.False(t, l.NoIssues("fred"))
	assert.Equal(t, "blee", l.Issues()["fred"][0].Description())
}

func TestLinterAddIssuesMap(t *testing.T) {
	l := NewLinter(nil, nil)

	l.initIssues("fred")
	assert.True(t, l.NoIssues("fred"))

	l.addIssuesMap("fred", Issues{"blee": []Issue{NewError(WarnLevel, "this is hosed")}})
	assert.False(t, l.NoIssues("fred"))
	assert.Equal(t, "blee||this is hosed", l.Issues()["fred"][0].Description())
}

func TestLinterAddErrors(t *testing.T) {
	l := NewLinter(nil, nil)

	l.initIssues("fred")
	assert.True(t, l.NoIssues("fred"))

	l.addErrors("fred", fmt.Errorf("Blee"))
	assert.False(t, l.NoIssues("fred"))
	assert.Equal(t, ErrorLevel, l.MaxSeverity("fred"))
	assert.Equal(t, "Blee", l.Issues()["fred"][0].Description())
}

func TestLinterAddError(t *testing.T) {
	l := NewLinter(nil, nil)

	l.initIssues("fred")
	assert.True(t, l.NoIssues("fred"))

	l.addError("fred", fmt.Errorf("Blee"))
	assert.False(t, l.NoIssues("fred"))
	assert.Equal(t, ErrorLevel, l.MaxSeverity("fred"))
	assert.Equal(t, "Blee", l.Issues()["fred"][0].Description())
}
