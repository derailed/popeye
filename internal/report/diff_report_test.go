package report

import (
	"testing"

	"github.com/derailed/popeye/internal/issues"
	"github.com/stretchr/testify/assert"
)

func TestPodName(t *testing.T) {
	uu := map[string]struct {
		po, e string
	}{
		"match": {"default/fred-1234-4567", "fred"},
	}

	for k, u := range uu {
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, podName(u.po))
		})
	}
}

func TestDiffTallies(t *testing.T) {
	uu := map[string]struct {
		t1, t2 Tally
		e      []DeltaScore
	}{
		"same": {
			t1: Tally{counts: []int{1, 1, 1, 1}},
			t2: Tally{counts: []int{1, 1, 1, 1}},
		},
		"error": {
			t1: Tally{counts: []int{1, 1, 1, 1}},
			t2: Tally{counts: []int{2, 1, 1, 1}},
			e: []DeltaScore{
				NewDeltaScore(issues.ErrorLevel, 1, 2, true),
			},
		},
		"warn": {
			t1: Tally{counts: []int{1, 1, 1, 1}},
			t2: Tally{counts: []int{1, 5, 1, 1}},
			e: []DeltaScore{
				NewDeltaScore(issues.WarnLevel, 1, 5, true),
			},
		},
		"info": {
			t1: Tally{counts: []int{1, 1, 1, 1}},
			t2: Tally{counts: []int{1, 1, 5, 1}},
			e: []DeltaScore{
				NewDeltaScore(issues.InfoLevel, 1, 5, true),
			},
		},
		"ok": {
			t1: Tally{counts: []int{1, 1, 1, 5}},
			t2: Tally{counts: []int{1, 1, 1, 1}},
			e: []DeltaScore{
				NewDeltaScore(issues.OkLevel, 5, 1, false),
			},
		},
	}

	for k, u := range uu {
		t.Run(k, func(t *testing.T) {
			r := DiffReport{sections: make(map[string]*resourceSection)}
			s := newResourceSection()
			r.sections["fred"] = s
			r.diffTallies(s, &u.t1, &u.t2)
			assert.Equal(t, u.e, r.sections["fred"].tallies)
			assert.Equal(t, 0, len(r.errors))
		})
	}
}

func TestDiffOutcomes(t *testing.T) {
	uu := map[string]struct {
		ii1, ii2 issues.Issues
		level    issues.Level
		e        deltaIssues
	}{
		"same": {
			ii1: issues.Issues{
				{Group: "g1", Level: issues.ErrorLevel, Message: "m1"},
			},
			ii2: issues.Issues{
				{Group: "g1", Level: issues.ErrorLevel, Message: "m1"},
			},
			level: issues.ErrorLevel,
		},
		"more_errors": {
			ii1: issues.Issues{
				{Group: "g1", Level: issues.ErrorLevel, Message: "m1"},
			},
			ii2: issues.Issues{
				{Group: "g1", Level: issues.ErrorLevel, Message: "m1"},
				{Group: "g1", Level: issues.ErrorLevel, Message: "m2"},
			},
			level: issues.ErrorLevel,
			e: deltaIssues{
				newDeltaIssue(
					issues.Issue{Group: "g1", Level: issues.ErrorLevel, Message: "m2"},
					true,
				),
			},
		},
		"less_errors": {
			ii1: issues.Issues{
				{Group: "g1", Level: issues.ErrorLevel, Message: "m1"},
				{Group: "g1", Level: issues.ErrorLevel, Message: "m2"},
			},
			ii2: issues.Issues{
				{Group: "g1", Level: issues.ErrorLevel, Message: "m1"},
			},
			level: issues.ErrorLevel,
			e: deltaIssues{
				newDeltaIssue(
					issues.Issue{Group: "g1", Level: issues.ErrorLevel, Message: "m2"},
					false,
				),
			},
		},
		"more_warn": {
			ii1: issues.Issues{
				{Group: "g1", Level: issues.WarnLevel, Message: "m1"},
			},
			ii2: issues.Issues{
				{Group: "g1", Level: issues.WarnLevel, Message: "m1"},
				{Group: "g1", Level: issues.WarnLevel, Message: "m2"},
			},
			level: issues.WarnLevel,
			e: deltaIssues{
				newDeltaIssue(
					issues.Issue{Group: "g1", Level: issues.WarnLevel, Message: "m2"},
					true,
				),
			},
		},
	}

	for k, u := range uu {
		t.Run(k, func(t *testing.T) {
			r := DiffReport{sections: make(map[string]*resourceSection)}
			s := newResourceSection()
			r.sections["fred"] = s
			ii := r.surface(u.ii1, u.ii2)
			assert.Equal(t, u.e, ii)
			assert.Equal(t, 0, len(r.errors))
		})
	}
}
