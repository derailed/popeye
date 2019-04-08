package report

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/derailed/popeye/internal/linter"
	"github.com/stretchr/testify/assert"
)

func TestComment(t *testing.T) {
	w := bytes.NewBufferString("")

	Comment(w, "blee")

	assert.Equal(t, "  · blee\n", w.String())
}

func TestError(t *testing.T) {
	uu := []struct {
		err error
		e   string
	}{
		{
			fmt.Errorf("crapola"),
			"\n💥 \x1b[38;5;196;mblee: crapola\x1b[0m\n",
		},
		{
			fmt.Errorf(strings.Repeat("#", 200)),
			"\n💥 \x1b[38;5;196;mblee: " + strings.Repeat("#", Width-9) + "\x1b[0m\n\x1b[38;5;196;m" + strings.Repeat("#", Width-3) + "\x1b[0m\n\x1b[38;5;196;m" + strings.Repeat("#", Width-88) + "\x1b[0m\n",
		},
	}

	for _, u := range uu {
		w := bytes.NewBufferString("")
		Error(w, "blee", u.err)

		assert.Equal(t, u.e, w.String())
	}
}

func TestWrite(t *testing.T) {
	uu := []struct {
		m      string
		indent int
		e      string
	}{
		{
			"Yo mama",
			1,
			"  · \x1b[38;5;155;mYo mama\x1b[0m\x1b[38;5;250;m" + strings.Repeat(".", Width-12) + "\x1b[0m✅\n",
		},
		{
			strings.Repeat("#", Width),
			1,
			"  · \x1b[38;5;155;m" + strings.Repeat("#", Width-8) + "...\x1b[0m\x1b[38;5;250;m\x1b[0m✅\n",
		},
		{
			"Yo mama",
			2,
			"    ✅ \x1b[38;5;155;mYo mama\x1b[0m\n",
		},
	}

	for _, u := range uu {
		w := bytes.NewBufferString("")
		Write(w, linter.OkLevel, u.indent, u.m)

		assert.Equal(t, u.e, w.String())
	}
}

func TestDump(t *testing.T) {
	uu := []struct {
		issues linter.Issues
		e      string
	}{
		{
			linter.Issues{
				"fred": []linter.Issue{linter.NewError(linter.WarnLevel, "Yo Mama!")},
			},
			"    😱 \x1b[38;5;220;mYo Mama!.\x1b[0m\n",
		},
		{
			linter.Issues{
				"fred": []linter.Issue{
					linter.NewError(linter.WarnLevel, "c1||Yo Mama!"),
					linter.NewError(linter.WarnLevel, "c1||Yo!"),
				},
			},
			"    🐳 \x1b[38;5;155;mc1\x1b[0m\n      😱 \x1b[38;5;220;mYo Mama!.\x1b[0m\n      😱 \x1b[38;5;220;mYo!.\x1b[0m\n",
		},
	}

	for _, u := range uu {
		w := bytes.NewBufferString("")
		Dump(w, linter.OkLevel, u.issues["fred"]...)

		assert.Equal(t, u.e, w.String())
	}
}

func BenchmarkWrite(b *testing.B) {
	for n := 0; n < b.N; n++ {
		Write(ioutil.Discard, 1, 1, "Yo mama")
	}
}

func TestOpen(t *testing.T) {
	uu := []struct {
		issues linter.Issues
		e      string
	}{
		{
			linter.Issues{
				"fred": []linter.Issue{linter.NewError(linter.WarnLevel, "Yo Mama!")},
			},
			"\n\x1b[38;5;75;mblee\x1b[0m" + strings.Repeat(" ", 75) + "💥 0 😱 1 🔊 0 ✅ 0 \x1b[38;5;196;m0\x1b[0m٪\n\x1b[38;5;75;m" + strings.Repeat("┅", Width) + "\x1b[0m\n",
		},
	}

	for _, u := range uu {
		w := bytes.NewBufferString("")
		ta := NewTally().Rollup(u.issues)
		Open(w, "blee", ta)

		assert.Equal(t, u.e, w.String())
	}
}

func TestOpenClose(t *testing.T) {
	w := bytes.NewBufferString("")
	Open(w, "fred", nil)
	Close(w)

	assert.Equal(t, "\n\x1b[38;5;75;mfred\x1b[0m\n\x1b[38;5;75;m"+strings.Repeat("┅", Width)+"\x1b[0m\n\n", w.String())
}

func TestTruncate(t *testing.T) {
	uu := []struct {
		s string
		l int
		e string
	}{
		{"fred", 3, "..."},
		{"freddy", 5, "fr..."},
		{"fred", 10, "fred"},
	}

	for _, u := range uu {
		assert.Equal(t, u.e, truncate(u.s, u.l))
	}
}
