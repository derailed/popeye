package report

import (
	"testing"

	"github.com/derailed/popeye/internal/linter"
	"github.com/stretchr/testify/assert"
)

func TestTallyRollup(t *testing.T) {
	uu := []struct {
		issues linter.Issues
		e      *Tally
	}{
		{
			linter.Issues{},
			&Tally{counts: []int{0, 0, 0, 0}, score: 0, valid: false},
		},
		{
			linter.Issues{
				"a": {
					linter.NewError(linter.InfoLevel, ""),
					linter.NewError(linter.WarnLevel, ""),
				},
				"b": {
					linter.NewError(linter.ErrorLevel, ""),
				},
				"c": {},
			},
			&Tally{counts: []int{1, 1, 1, 1}, score: 50, valid: true},
		},
	}

	for _, u := range uu {
		ta := NewTally()
		ta.Rollup(u.issues)

		assert.Equal(t, u.e, ta)
	}
}

func TestTallyScore(t *testing.T) {
	uu := []struct {
		issues linter.Issues
		e      int
	}{
		{
			linter.Issues{
				"a": {
					linter.NewError(linter.InfoLevel, ""),
					linter.NewError(linter.WarnLevel, ""),
				},
				"b": {
					linter.NewError(linter.ErrorLevel, ""),
				},
				"c": {},
			},
			50,
		},
	}

	for _, u := range uu {
		ta := NewTally()
		ta.Rollup(u.issues)

		assert.Equal(t, u.e, ta.Score())
	}
}

func TestTallyWidth(t *testing.T) {
	uu := []struct {
		issues linter.Issues
		e      int
	}{
		{
			linter.Issues{
				"a": {
					linter.NewError(linter.InfoLevel, ""),
					linter.NewError(linter.WarnLevel, ""),
				},
				"b": {
					linter.NewError(linter.ErrorLevel, ""),
				},
				"c": {},
			},
			35,
		},
	}

	for _, u := range uu {
		ta := NewTally()
		ta.Rollup(u.issues)

		assert.Equal(t, u.e, ta.Width())
	}
}
