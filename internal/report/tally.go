package report

import (
	"bytes"
	"fmt"
	"io"
	"math"
	"strconv"
	"unicode/utf8"

	"github.com/derailed/popeye/internal/linter"
)

// Tally tracks lint section scores.
type Tally struct {
	counts []int
}

// NewTally returns a new tally.
func NewTally() *Tally {
	return &Tally{counts: make([]int, 4)}
}

// Rollup tallies up the report scores.
func (t *Tally) Rollup(run linter.Issues) {
	for _, issues := range run {
		if len(issues) == 0 {
			t.counts[linter.OkLevel]++
		}
		for _, issue := range issues {
			t.counts[issue.Severity()]++
		}
	}
}

// Score computes the total tally score.
func (t *Tally) Score() int {
	var total, ok int
	for i, v := range t.counts {
		if i < 2 {
			ok += v
		}
		total += v
	}
	return int(math.Round(linter.ToPerc(float64(ok), float64(total))))
}

// Dump prints out a tally.
func (t *Tally) Dump(w io.Writer) {
	for i := len(t.counts) - 1; i >= 0; i-- {
		emoji := EmojiForLevel(linter.Level(i))
		fmt.Fprintf(w, "%s %d ", emoji, t.counts[i])
	}
	perc, color := t.Score(), ColorAqua
	if perc < 80 {
		color = ColorRed
	}
	fmt.Fprintf(w, "%s٪", Colorize(strconv.Itoa(perc), color))
}

// Width computes the tally width.
func (t *Tally) Width() int {
	w := bytes.NewBufferString("")
	t.Dump(w)

	return utf8.RuneCountInString(w.String())
}
