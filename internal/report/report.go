// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Popeye

package report

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/derailed/popeye/internal/issues"
	"github.com/derailed/popeye/internal/issues/tally"
	"github.com/fvbommel/sortorder"
)

// Report represents a popeye scan report.
type Report struct {
	Timestamp     string   `json:"report_time" yaml:"report_time"`
	Score         int      `json:"score" yaml:"score"`
	Grade         string   `json:"grade" yaml:"grade"`
	Sections      Sections `json:"sections,omitempty" yaml:"sections,omitempty"`
	Errors        Errors   `json:"errors,omitempty" yaml:"errors,omitempty"`
	sectionsCount int
	totalScore    int
}

func (r Report) ListSections() Sections {
	return r.Sections
}

// Sections represents a collection of sections.
type Sections []Section

// Section represents a linter pass
type Section struct {
	Title    string         `json:"linter" yaml:"linter"`
	GVR      string         `json:"gvr" yaml:"gvr"`
	Tally    *Tally         `json:"tally" yaml:"tally"`
	Outcome  issues.Outcome `json:"issues,omitempty" yaml:"issues,omitempty"`
	singular string
}

// Len returns the list size.
func (s Sections) Len() int {
	return len(s)
}

func (s Sections) CodeTallies() tally.Linter {
	ss := make(tally.Linter)
	for _, section := range s {
		ss[section.Title] = section.Outcome.NSTally()
	}

	return ss
}

// Swap swaps list values.
func (s Sections) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

// Less returns true if i < j.
func (s Sections) Less(i, j int) bool {
	return sortorder.NaturalLess(s[i].singular, s[j].singular)
}

type Errors []error

func (ee Errors) MarshalJSON() ([]byte, error) {
	if len(ee) == 0 {
		return nil, nil
	}
	errs := make([]string, 0, len(ee))
	for _, e := range ee {
		if e == nil {
			continue
		}
		raw, err := json.Marshal(e.Error())
		if err != nil {
			return nil, err
		}
		errs = append(errs, fmt.Sprintf(`"error": %s`, string(raw)))
	}
	s := "{" + strings.Join(errs, ",") + "}"

	return []byte(s), nil
}

func (ee Errors) MarshalYAML() (interface{}, error) {
	if len(ee) == 0 {
		return nil, nil
	}
	out := make([]string, 0, len(ee))
	for _, e := range ee {
		if e == nil || e.Error() == "" {
			continue
		}
		out = append(out, e.Error())
	}

	return out, nil
}
