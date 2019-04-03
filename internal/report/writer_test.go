package report

import (
	"bytes"
	"io/ioutil"
	"testing"
	"unicode/utf8"

	"github.com/derailed/popeye/internal/linter"
	"github.com/stretchr/testify/assert"
)

func TestOpenClose(t *testing.T) {
	w := bytes.NewBufferString("")
	Open(w, "fred")
	Close(w)
	assert.Equal(t, "\n\x1b[38;5;75;mfred\x1b[0m\n\x1b[38;5;75;m┅┅┅┅┅┅┅┅┅┅┅┅┅┅┅┅┅┅┅┅┅┅┅┅┅┅┅┅┅┅┅┅┅┅┅┅┅┅┅┅┅┅┅┅┅┅┅┅┅┅┅┅┅┅┅┅┅┅┅┅┅┅┅┅┅┅┅┅┅┅┅┅┅┅┅┅┅┅┅┅\x1b[0m\n\n\n", w.String())
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
			"  · \x1b[38;5;122;mYo mama\x1b[0m\x1b[38;5;250;m...................................................................\x1b[0m✅\n",
		},
		{
			"Yo mama",
			2,
			"      ✅ \x1b[38;5;15;mYo mama\x1b[0m\n",
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
			"      😱 \x1b[38;5;15;mYo Mama!\x1b[0m\n",
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

func TestEmojiForLevel(t *testing.T) {
	for k, v := range map[int]int{0: 1, 1: 1, 2: 1, 3: 1, 4: 1} {
		assert.Equal(t, v, utf8.RuneCountInString(emojiForLevel(linter.Level(k))))
	}
}

func TestColorForLevel(t *testing.T) {
	for k, v := range map[int]Color{0: ColorAqua, 1: ColorAqua, 2: ColorAqua, 3: ColorOrangish, 4: ColorRed} {
		assert.Equal(t, v, colorForLevel(linter.Level(k)))
	}
}
