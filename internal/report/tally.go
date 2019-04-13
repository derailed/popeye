package report

import (
	"bytes"
	"fmt"
	"io"
	"math"
	"strconv"

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

// Write out a tally.
func (t *Tally) write(w io.Writer, s *Sanitizer) {
	for i := len(t.counts) - 1; i >= 0; i-- {
		emoji := s.EmojiForLevel(linter.Level(i))
		fmat := "%s %d "
		if s.jurassicMode {
			fmat = "%s:%d "
		}
		fmt.Fprintf(w, fmat, emoji, t.counts[i])
	}

	score, color := t.score, ColorAqua
	if score < 80 {
		color = ColorRed
	}
	percentageSign := "٪"
	if s.jurassicMode {
		percentageSign = "%%"
	}
	fmt.Fprintf(w, "%s%s", s.Color(strconv.Itoa(score), color), percentageSign)
}

// Dump writes out tally and computes length
func (t *Tally) Dump(s *Sanitizer) string {
	w := bytes.NewBufferString("")
	t.write(w, s)

	return w.String()
}
