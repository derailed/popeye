package report

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/derailed/popeye/internal/client"
	"github.com/derailed/popeye/internal/issues"
	"github.com/derailed/popeye/pkg/config"
	"github.com/stretchr/testify/assert"
)

func TestComment(t *testing.T) {
	w := bytes.NewBufferString("")
	s := NewSanitizer(w, false)

	s.Comment("blee")

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
		s := NewSanitizer(w, false)
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
			"  · \x1b[38;5;155;mYo mama\x1b[0m\x1b[38;5;250;m" + strings.Repeat(".", Width-12) + "\x1b[0m✅\n",
		},
		{
			strings.Repeat("#", Width-8),
			1,
			"  · \x1b[38;5;155;m" + strings.Repeat("#", Width-8) + "\x1b[0m\x1b[38;5;250;m...\x1b[0m✅\n",
		},
		{
			"Yo mama",
			2,
			"    ✅ \x1b[38;5;155;mYo mama\x1b[0m\n",
		},
	}

	for _, u := range uu {
		w := bytes.NewBufferString("")
		s := NewSanitizer(w, false)
		s.Print(config.OkLevel, u.indent, u.m)

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
				"fred": issues.Issues{issues.New(client.NewGVR("fred"), issues.Root, config.WarnLevel, "Yo Mama!")},
			},
			"    😱 \x1b[38;5;220;mYo Mama!.\x1b[0m\n",
		},
		{
			issues.Outcome{
				"fred": issues.Issues{
					issues.New(client.NewGVR("fred"), "c1", config.ErrorLevel, "Yo Mama!"),
					issues.New(client.NewGVR("fred"), "c1", config.ErrorLevel, "Yo!"),
				},
			},
			"    🐳 \x1b[38;5;75;mc1\x1b[0m\n      💥 \x1b[38;5;196;mYo Mama!.\x1b[0m\n      💥 \x1b[38;5;196;mYo!.\x1b[0m\n",
		},
	}

	for _, u := range uu {
		w := bytes.NewBufferString("")
		s := NewSanitizer(w, false)
		s.Dump(config.OkLevel, u.o["fred"])

		assert.Equal(t, u.e, w.String())
	}
}

func BenchmarkPrint(b *testing.B) {
	s := NewSanitizer(ioutil.Discard, false)

	b.ResetTimer()
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		s.Print(1, 1, "Yo mama")
	}
}

func TestOpen(t *testing.T) {
	uu := []struct {
		o issues.Outcome
		e string
	}{
		{
			issues.Outcome{
				"fred": issues.Issues{issues.New(client.NewGVR("fred"), issues.Root, config.WarnLevel, "Yo Mama!")},
			},
			"\n\x1b[38;5;75;mblee\x1b[0m" + strings.Repeat(" ", 75) + "💥 0 😱 1 🔊 0 ✅ 0 \x1b[38;5;196;m0\x1b[0m٪\n\x1b[38;5;75;m" + strings.Repeat("┅", Width+1) + "\x1b[0m\n",
		},
	}

	for _, u := range uu {
		w := bytes.NewBufferString("")
		s := NewSanitizer(w, false)

		ta := NewTally().Rollup(u.o)
		s.Open("blee", ta)

		assert.Equal(t, u.e, w.String())
	}
}

func TestOpenClose(t *testing.T) {
	w := bytes.NewBufferString("")
	s := NewSanitizer(w, false)

	s.Open("fred", nil)
	s.Close()

	assert.Equal(t, "\n\x1b[38;5;75;mfred\x1b[0m\n\x1b[38;5;75;m"+strings.Repeat("┅", Width+1)+"\x1b[0m\n\n", w.String())
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
			e:      "fred \n    bleeduhblablabla\n    blee",
		},
	}

	for _, u := range uu {
		assert.Equal(t, u.e, formatLine(u.msg, 1, u.width))
	}
}
