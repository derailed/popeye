// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Popeye

package report

import (
	"testing"

	"github.com/derailed/popeye/internal/rules"

	"github.com/stretchr/testify/assert"
)

func TestChanged(t *testing.T) {
	uu := map[string]struct {
		old     int
		new     int
		inverse bool
		e       bool
	}{
		"same": {
			old:     10,
			new:     10,
			inverse: false,
		},
		"better": {
			old: 10,
			new: 15,
			e:   true,
		},
		"worst": {
			old: 20,
			new: 15,
			e:   true,
		},
	}

	l := rules.OkLevel
	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, NewDeltaScore(l, u.old, u.new, u.inverse).changed())
		})
	}
}

func TestBetter(t *testing.T) {
	uu := map[string]struct {
		old     int
		new     int
		inverse bool
		e       bool
	}{
		"same": {
			old: 10,
			new: 10,
		},
		"better": {
			old: 10,
			new: 15,
			e:   true,
		},
		"better_inverse": {
			old:     10,
			new:     15,
			inverse: true,
		},
		"worst": {
			old: 15,
			new: 10,
		},
		"worst_inverse": {
			old:     15,
			new:     10,
			inverse: true,
			e:       true,
		},
	}

	l := rules.OkLevel
	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, NewDeltaScore(l, u.old, u.new, u.inverse).better())
		})
	}
}

func TestWorst(t *testing.T) {
	uu := map[string]struct {
		old     int
		new     int
		inverse bool
		e       bool
	}{
		"same": {
			old: 10,
			new: 10,
		},
		"worst": {
			old: 10,
			new: 5,
			e:   true,
		},
		"worst_inverse": {
			old:     10,
			new:     5,
			inverse: true,
		},
		"better": {
			old: 15,
			new: 20,
		},
		"better_inverse": {
			old:     15,
			new:     20,
			inverse: true,
			e:       true,
		},
	}

	l := rules.OkLevel
	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, NewDeltaScore(l, u.old, u.new, u.inverse).worst())
		})
	}
}

func TestSummarize(t *testing.T) {
	uu := map[string]struct {
		old     int
		new     int
		inverse bool
		e       string
	}{
		"same": {
			old: 10,
			new: 10,
			e:   "not changed",
		},
		"worst": {
			old: 10,
			new: 5,
			e:   "worsened",
		},
		"worst_inverse": {
			old:     10,
			new:     5,
			inverse: true,
			e:       "improved",
		},
		"better": {
			old: 15,
			new: 20,
			e:   "improved",
		},
		"better_inverse": {
			old:     15,
			new:     20,
			inverse: true,
			e:       "worsened",
		},
	}

	l := rules.OkLevel
	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, NewDeltaScore(l, u.old, u.new, u.inverse).summarize())
		})
	}
}
