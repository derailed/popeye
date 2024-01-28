// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Popeye

package report

import (
	"bytes"
	"testing"

	"github.com/derailed/popeye/internal/issues"
	"github.com/derailed/popeye/internal/rules"
	"github.com/derailed/popeye/types"
	"github.com/stretchr/testify/assert"
)

func TestTallyWrite(t *testing.T) {
	uu := []struct {
		jurassic bool
		e        string
	}{
		{false, "💥 0 😱 0 🔊 0 ✅ 0 \x1b[38;5;196m0\x1b[0m٪"},
		{true, "E:0 W:0 I:0 OK:0 0%%"},
	}

	for _, u := range uu {
		ta := NewTally()
		b := bytes.NewBuffer([]byte(""))
		s := New(b, u.jurassic)
		ta.write(b, s)

		assert.Equal(t, u.e, b.String())
	}
}

func TestTallyRollup(t *testing.T) {
	uu := map[string]struct {
		o issues.Outcome
		s int
		e *Tally
	}{
		"no-issues": {
			o: issues.Outcome{},
			e: &Tally{counts: []int{0, 0, 0, 0}, score: 100, valid: true},
		},
		"plain": {
			o: issues.Outcome{
				"a": {
					issues.New(types.NewGVR("fred"), issues.Root, rules.InfoLevel, ""),
					issues.New(types.NewGVR("fred"), issues.Root, rules.WarnLevel, ""),
				},
				"b": {
					issues.New(types.NewGVR("fred"), issues.Root, rules.ErrorLevel, ""),
				},
				"c": {},
			},
			e: &Tally{counts: []int{1, 0, 1, 1}, score: 33, valid: true},
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			ta := NewTally()
			ta.Rollup(u.o)

			assert.Equal(t, u.e, ta)
		})
	}
}

func TestTallyScore(t *testing.T) {
	uu := []struct {
		o issues.Outcome
		s int
		e int
	}{
		{
			o: issues.Outcome{
				"a": {
					issues.New(types.NewGVR("fred"), issues.Root, rules.InfoLevel, ""),
					issues.New(types.NewGVR("fred"), issues.Root, rules.WarnLevel, ""),
				},
				"b": {
					issues.New(types.NewGVR("fred"), issues.Root, rules.ErrorLevel, ""),
				},
				"c": {},
			},
			e: 33,
		},
	}

	for _, u := range uu {
		ta := NewTally()
		ta.Rollup(u.o)

		assert.Equal(t, u.e, ta.Score())
	}
}

func TestTallyWidth(t *testing.T) {
	uu := []struct {
		o issues.Outcome
		s int
		e string
	}{
		{
			o: issues.Outcome{
				"a": {
					issues.New(types.NewGVR("fred"), issues.Root, rules.InfoLevel, ""),
					issues.New(types.NewGVR("fred"), issues.Root, rules.WarnLevel, ""),
				},
				"b": {
					issues.New(types.NewGVR("fred"), issues.Root, rules.ErrorLevel, ""),
				},
				"c": {},
			},
			e: "💥 1 😱 1 🔊 0 ✅ 1 \x1b[38;5;196m33\x1b[0m٪",
		},
	}

	s := new(ScanReport)
	for _, u := range uu {
		ta := NewTally()
		ta.Rollup(u.o)

		assert.Equal(t, u.e, ta.Dump(s))
	}
}

func TestToPerc(t *testing.T) {
	uu := []struct {
		v1, v2 float64
		e      float64
	}{
		{0, 0, 0},
		{100, 50, 200},
		{50, 100, 50},
	}

	for _, u := range uu {
		assert.Equal(t, u.e, toPerc(u.v1, u.v2))
	}
}

func TestMarshalJSON(t *testing.T) {
	uu := []struct {
		t *Tally
		e string
	}{
		{NewTally(), `{"ok":0,"info":0,"warning":0,"error":0,"score":0}`},
	}

	for _, u := range uu {
		s, err := u.t.MarshalJSON()
		assert.Nil(t, err)
		assert.Equal(t, u.e, string(s))
	}
}

func TestMarshalYAML(t *testing.T) {
	uu := []struct {
		t *Tally
		e interface{}
	}{
		{NewTally(), struct {
			OK    int `yaml:"ok"`
			Info  int `yaml:"info"`
			Warn  int `yaml:"warning"`
			Error int `yaml:"error"`
			Score int `yaml:"score"`
		}{
			OK:    0,
			Info:  0,
			Warn:  0,
			Error: 0,
			Score: 0,
		}},
	}

	for _, u := range uu {
		s, err := u.t.MarshalYAML()
		assert.Nil(t, err)
		assert.Equal(t, u.e, s)
	}
}
