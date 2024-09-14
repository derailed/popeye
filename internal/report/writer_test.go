// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Popeye

package report

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/derailed/popeye/internal/issues"
	"github.com/derailed/popeye/internal/rules"
	"github.com/derailed/popeye/types"
	"github.com/stretchr/testify/assert"
)

func TestComment(t *testing.T) {
	w := bytes.NewBufferString("")
	s := New(w, false)

	s.Comment("blee")

	assert.Equal(t, "  · blee\n", w.String())
}

func TestError(t *testing.T) {
	uu := []struct {
		err error
		e   string
	}{
		{
			err: fmt.Errorf("crapola"),
			e:   "\n💥 \x1b[38;5;196mblee: crapola\x1b[0m\n",
		},
		{
			err: errors.New(strings.Repeat("#", 200)),
			e:   "\n💥 \x1b[38;5;196mblee: " + strings.Repeat("#", Width-9) + "\x1b[0m\n\x1b[38;5;196m" + strings.Repeat("#", Width-3) + "\x1b[0m\n\x1b[38;5;196m" + strings.Repeat("#", Width-88) + "\x1b[0m\n",
		},
	}

	for _, u := range uu {
		w := bytes.NewBufferString("")
		s := New(w, false)
		s.Error("blee", u.err)

		assert.Equal(t, u.e, w.String())
	}
}

func TestPrint(t *testing.T) {
	uu := []struct {
		m      string
		indent int
		e      string
	}{
		{
			"Yo mama",
			1,
			"  · \x1b[38;5;155mYo mama\x1b[0m\x1b[38;5;250m" + strings.Repeat(".", Width-12) + "\x1b[0m✅\n",
		},
		{
			strings.Repeat("#", Width-8),
			1,
			"  · \x1b[38;5;155m" + strings.Repeat("#", Width-8) + "\x1b[0m\x1b[38;5;250m...\x1b[0m✅\n",
		},
		{
			"Yo mama",
			2,
			"    ✅ \x1b[38;5;155mYo mama\x1b[0m\n",
		},
	}

	for _, u := range uu {
		w := bytes.NewBufferString("")
		s := New(w, false)
		s.Print(rules.OkLevel, u.indent, u.m)

		assert.Equal(t, u.e, w.String())
	}
}

func TestDump(t *testing.T) {
	uu := []struct {
		o issues.Outcome
		e string
	}{
		{
			issues.Outcome{
				"fred": issues.Issues{issues.New(types.NewGVR("fred"), issues.Root, rules.WarnLevel, "Yo Mama!")},
			},
			"    😱 \x1b[38;5;220mYo Mama!.\x1b[0m\n",
		},
		{
			issues.Outcome{
				"fred": issues.Issues{
					issues.New(types.NewGVR("fred"), "c1", rules.ErrorLevel, "Yo Mama!"),
					issues.New(types.NewGVR("fred"), "c1", rules.ErrorLevel, "Yo!"),
				},
			},
			"    🐳 \x1b[38;5;75mc1\x1b[0m\n      💥 \x1b[38;5;196mYo Mama!.\x1b[0m\n      💥 \x1b[38;5;196mYo!.\x1b[0m\n",
		},
	}

	for _, u := range uu {
		w := bytes.NewBufferString("")
		s := New(w, false)
		s.Dump(rules.OkLevel, u.o["fred"])

		assert.Equal(t, u.e, w.String())
	}
}

func BenchmarkPrint(b *testing.B) {
	s := New(io.Discard, false)

	b.ResetTimer()
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		s.Print(1, 1, "Yo mama")
	}
}

func TestOpen(t *testing.T) {
	uu := []struct {
		o issues.Outcome
		s int
		e string
	}{
		{
			o: issues.Outcome{
				"fred": issues.Issues{issues.New(types.NewGVR("fred"), issues.Root, rules.WarnLevel, "Yo Mama!")},
			},
			e: "\n\x1b[38;5;75mblee\x1b[0m" + strings.Repeat(" ", 75) + "💥 0 😱 1 🔊 0 ✅ 0 \x1b[38;5;196m0\x1b[0m٪\n\x1b[38;5;75m" + strings.Repeat("┅", Width+1) + "\x1b[0m\n",
		},
	}

	for _, u := range uu {
		w := bytes.NewBufferString("")
		s := New(w, false)

		ta := NewTally().Rollup(u.o)
		s.Open("blee", ta)

		assert.Equal(t, u.e, w.String())
	}
}

func TestOpenClose(t *testing.T) {
	w := bytes.NewBufferString("")
	s := New(w, false)

	s.Open("fred", nil)
	s.Close()

	assert.Equal(t, "\n\x1b[38;5;75mfred\x1b[0m\n\x1b[38;5;75m"+strings.Repeat("┅", Width+1)+"\x1b[0m\n\n", w.String())
}

func TestFormatLine(t *testing.T) {
	uu := map[string]struct {
		msg           string
		indent, width int
		e             string
	}{
		"single": {
			msg:    "fred blee",
			indent: 1,
			width:  10,
			e:      "fred blee",
		},
		"newline": {
			msg:    "fred bleeduhblablabla blee",
			indent: 1,
			width:  10,
			e:      "fred \n     bleeduhblablabla \n     blee ",
		},
	}

	for _, u := range uu {
		assert.Equal(t, u.e, formatLine(u.msg, 1, u.width))
	}
}
