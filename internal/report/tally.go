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
	score  int
	valid  bool
}

// NewTally returns a new tally.
func NewTally() *Tally {
	return &Tally{counts: make([]int, 4)}
}

// Score returns the tally computed score.
func (t *Tally) Score() int {
	return t.score
}

// IsValid checks if tally is valid.
func (t *Tally) IsValid() bool {
	return t.valid
}

// Rollup tallies up the report scores.
func (t *Tally) Rollup(run linter.Issues) *Tally {
	if run == nil || len(run) == 0 {
		return t
	}

	t.valid = true
	for _, issues := range run {
		if len(issues) == 0 {
			t.counts[linter.OkLevel]++
		}
		for _, issue := range issues {
			t.counts[issue.Severity()]++
		}
	}
	t.computeScore()

	return t
}

// ComputeScore calculates the completed run score.
func (t *Tally) computeScore() int {
	var total, ok int
	for i, v := range t.counts {
		if i < 2 {
			ok += v
		}
		total += v
	}
	t.score = int(math.Round(linter.ToPerc(float64(ok), float64(total))))

	return t.score
}

// Dump prints out a tally.
func (t *Tally) Dump(w io.Writer) {
	for i := len(t.counts) - 1; i >= 0; i-- {
		emoji := EmojiForLevel(linter.Level(i))
		fmt.Fprintf(w, "%s %d ", emoji, t.counts[i])
	}

	score, color := t.score, ColorAqua
	if score < 80 {
		color = ColorRed
	}
	fmt.Fprintf(w, "%s٪", Colorize(strconv.Itoa(score), color))
}

// Width computes the tally width.
func (t *Tally) Width() int {
	w := bytes.NewBufferString("")
	t.Dump(w)

	return utf8.RuneCountInString(w.String())
}
