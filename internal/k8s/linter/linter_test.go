package linter

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLinterMaxSeverity(t *testing.T) {
	uu := []struct {
		severity Level
		issues   Issues
	}{
		{
			severity: WarnLevel,
			issues: Issues{
				NewError(InfoLevel, "blee"),
				NewError(WarnLevel, "blee"),
			},
		},
		{
			severity: InfoLevel,
			issues: Issues{
				NewError(InfoLevel, "blee"),
				NewError(InfoLevel, "blee"),
			},
		},
		{
			severity: ErrorLevel,
			issues: Issues{
				NewError(ErrorLevel, "blee"),
				NewError(InfoLevel, "blee"),
				NewError(InfoLevel, "blee"),
			},
		},
	}

	for _, u := range uu {
		l := new(Linter)
		l.addIssues(u.issues...)
		assert.Equal(t, u.severity, l.MaxSeverity())
	}
}

func TestLinterNoIssues(t *testing.T) {
	uu := []struct {
		flag   bool
		issues Issues
	}{
		{
			issues: Issues{
				NewError(InfoLevel, "blee"),
				NewError(WarnLevel, "blee"),
			},
		},
		{
			flag:   true,
			issues: Issues{},
		},
	}

	for _, u := range uu {
		l := new(Linter)
		l.addIssues(u.issues...)
		assert.Equal(t, u.flag, l.NoIssues())
	}
}
