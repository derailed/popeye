// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Popeye

package report

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strconv"

	"github.com/derailed/popeye/internal/issues"
	"github.com/derailed/popeye/internal/lint"
	"github.com/derailed/popeye/internal/rules"
)

const targetScore = 80

// Tally tracks lint section scores.
type Tally struct {
	counts []int
	score  int
	valid  bool
}

// NewTally returns a new tally.
func NewTally() *Tally {
	return &Tally{
		counts: make([]int, 4),
	}
}

// Score returns the tally computed score.
func (t *Tally) Score() int {
	return t.score
}

// ErrCount returns the number of errors found.
func (t *Tally) ErrCount() int {
	return t.counts[3]
}

// WarnCount returns the number of warnings found.
func (t *Tally) WarnCount() int {
	return t.counts[2]
}

// IsValid checks if tally is valid.
func (t *Tally) IsValid() bool {
	return t.valid
}

// Rollup tallies up the report scores.
func (t *Tally) Rollup(o issues.Outcome) *Tally {
	if len(o) == 0 {
		t.valid, t.score = true, 100
		return t
	}

	t.valid = true
	for k := range o {
		t.counts[o.MaxSeverity(k)]++
	}
	t.computeScore()

	return t
}

// ComputeScore calculates the completed run score.
func (t *Tally) computeScore() int {
	var issues, ok int
	for i, v := range t.counts {
		if i < 2 {
			ok += v
		}
		issues += v
	}
	t.score = int(lint.ToPerc(int64(ok), int64(issues)))

	return t.score
}

// Write out a tally.
func (t *Tally) write(w io.Writer, s *ScanReport) {
	for i := len(t.counts) - 1; i >= 0; i-- {
		emoji := EmojiForLevel(rules.Level(i), s.jurassicMode)
		fmat := "%s %d "
		if s.jurassicMode {
			fmat = "%s:%d "
		}
		fmt.Fprintf(w, fmat, emoji, t.counts[i])
	}

	score, color := t.score, ColorAqua
	if score < targetScore {
		color = ColorRed
	}
	percentageSign := "٪"
	if s.jurassicMode {
		percentageSign = "%%"
	}
	fmt.Fprintf(w, "%s%s", s.Color(strconv.Itoa(score), color), percentageSign)
}

// Dump writes out tally and computes length
func (t *Tally) Dump(s *ScanReport) string {
	w := bytes.NewBufferString("")
	t.write(w, s)

	return w.String()
}

// MarshalYAML renders a tally to YAML.
func (t *Tally) MarshalYAML() (interface{}, error) {
	y := struct {
		OK    int `yaml:"ok"`
		Info  int `yaml:"info"`
		Warn  int `yaml:"warning"`
		Error int `yaml:"error"`
		Score int `yaml:"score"`
	}{
		Score: t.score,
	}

	for i, v := range t.counts {
		switch i {
		case 0:
			y.OK = v
		case 1:
			y.Info = v
		case 2:
			y.Warn = v
		case 3:
			y.Error = v
		}
	}

	return y, nil
}

// UnmarshalYAML renders a tally to YAML.
func (t *Tally) UnmarshalYAML(f func(interface{}) error) error {
	type tally struct {
		OK    int `yaml:"ok"`
		Info  int `yaml:"info"`
		Warn  int `yaml:"warning"`
		Error int `yaml:"error"`
		Score int `yaml:"score"`
	}
	var tmp tally

	if err := f(&tmp); err != nil {
		return err
	}

	t.counts = make([]int, 4)
	t.counts[0] = tmp.Error
	t.counts[1] = tmp.Warn
	t.counts[2] = tmp.Info
	t.counts[3] = tmp.OK
	t.score = tmp.Score
	t.valid = true

	return nil
}

// MarshalJSON renders a tally to JSON.
func (t *Tally) MarshalJSON() ([]byte, error) {
	y := struct {
		OK    int `json:"ok"`
		Info  int `json:"info"`
		Warn  int `json:"warning"`
		Error int `json:"error"`
		Score int `json:"score"`
	}{
		Score: t.score,
	}

	for i, v := range t.counts {
		switch i {
		case 0:
			y.OK = v
		case 1:
			y.Info = v
		case 2:
			y.Warn = v
		case 3:
			y.Error = v
		}
	}

	return json.Marshal(y)
}

// ----------------------------------------------------------------------------
// Helpers...

func toPerc(v1, v2 float64) float64 {
	if v2 == 0 {
		return 0
	}
	return (v1 / v2) * 100
}

func indexToTally(i int) string {
	switch i {
	case 1:
		return "Info"
	case 2:
		return "Warn"
	case 3:
		return "Error"
	default:
		return "OK"
	}
}
