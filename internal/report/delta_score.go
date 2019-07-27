package report

import (
	"strconv"

	"github.com/derailed/popeye/internal/issues"
)

const (
	noChange = "not changed"
	better   = "improved"
	worst    = "worsened"
)

// DeltaScore tracks delta between 2 tally scores.
type DeltaScore struct {
	level   issues.Level
	s1, s2  int
	inverse bool
}

// NewDeltaScore returns a new delta score.
func NewDeltaScore(level issues.Level, s1, s2 int, inverse bool) DeltaScore {
	return DeltaScore{
		s1:      s1,
		s2:      s2,
		level:   level,
		inverse: inverse,
	}
}

func (s DeltaScore) changed() bool {
	return s.s1 != s.s2
}

func (s DeltaScore) worst() bool {
	if s.s1 == s.s2 {
		return false
	}

	return !s.better()
}

func (s DeltaScore) delta() string {
	score := s.s2 - s.s1
	if score > 0 {
		return "+" + strconv.Itoa(score)
	}

	return strconv.Itoa(score)
}

func (s DeltaScore) better() bool {
	if s.s1 == s.s2 {
		return false
	}

	if s.s2 > s.s1 {
		if s.inverse {
			return false
		}
		return true
	}

	return s.inverse
}

func (s DeltaScore) summarize() string {
	if s.s1 == s.s2 {
		return noChange
	}

	if s.s1 > s.s2 {
		if s.inverse {
			return better
		}
		return worst
	}

	if s.inverse {
		return worst
	}

	return better
}
