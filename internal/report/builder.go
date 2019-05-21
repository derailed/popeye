package report

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/derailed/popeye/internal/issues"
	"gopkg.in/yaml.v2"
)

const (
	// DefaultFormat dumps sanitizer with color, emojis, the works.
	DefaultFormat = "standard"

	// JurassicFormat dumps sanitizer with 0 fancyness.
	JurassicFormat = "jurassic"

	// YAMLFormat dumps sanitizer as YAML
	YAMLFormat = "yaml"

	// JSONFormat dumps sanitizer as JSON
	JSONFormat = "json"
)

type (
	// Builder represents sanitizer
	Builder struct {
		Report Report `json:"popeye" yaml:"popeye"`
	}

	// Report represents the output of a sanitization pass.
	Report struct {
		Score         int       `json:"score" yaml:"score"`
		Grade         string    `json:"grade" yaml:"grade"`
		Sections      []Section `json:"sanitizers,omitempty" yaml:"sanitizers,omitempty"`
		Errors        []error   `json:"errors,omitempty" yaml:"errors,omitempty"`
		sectionsCount int
		totalScore    int
	}

	// Section represents a sanitizer pass
	Section struct {
		Title   string         `json:"sanitizer" yaml:"sanitizer"`
		Tally   *Tally         `json:"tally" yaml:"tally"`
		Outcome issues.Outcome `json:"issues,omitempty" yaml:"issues,omitempty"`
	}
)

// NewBuilder returns a new sanitizer report.
func NewBuilder() *Builder {
	return &Builder{}
}

// AddError record an error associted with the report.
func (b *Builder) AddError(err error) {
	b.Report.Errors = append(b.Report.Errors, err)
}

// AddSection adds a sanitizer section to the report.
func (b *Builder) AddSection(name string, o issues.Outcome, t *Tally) {
	section := Section{
		Title:   strings.ToLower(name),
		Tally:   t,
		Outcome: o,
	}
	b.Report.Sections = append(b.Report.Sections, section)
	if t.IsValid() {
		b.Report.sectionsCount++
		b.Report.totalScore += t.Score()
	}
}

// ToYAML dumps sanitizer to YAML.
func (b *Builder) ToYAML() (string, error) {
	if b.Report.sectionsCount == 0 {
		return "", errors.New("Nothing to report, check permissions")
	}

	score := b.Report.totalScore / b.Report.sectionsCount
	b.Report.Score = score
	b.Report.Grade = Grade(score)

	raw, err := yaml.Marshal(b)
	if err != nil {
		return "", err
	}

	return string(raw), nil
}

// ToJSON dumps sanitizer to JSON.
func (b *Builder) ToJSON() (string, error) {
	if b.Report.sectionsCount == 0 {
		return "", errors.New("Nothing to report, check permissions")
	}
	score := b.Report.totalScore / b.Report.sectionsCount
	b.Report.Score = score
	b.Report.Grade = Grade(score)

	raw, err := json.Marshal(b)
	if err != nil {
		return "", err
	}

	return string(raw), nil
}

// PrintSummary print outs summary report to screen.
func (b *Builder) PrintSummary(s *Sanitizer) {
	if b.Report.sectionsCount == 0 {
		return
	}

	s.Open("SUMMARY", nil)
	{
		score := b.Report.totalScore / b.Report.sectionsCount
		fmt.Fprintf(s, "Your cluster score: %d -- %s\n", score, Grade(score))
		for _, l := range s.Badge(score) {
			fmt.Fprintf(s, "%s%s\n", strings.Repeat(" ", Width-20), l)
		}
	}
	s.Close()
}

// PrintClusterInfo displays cluster information.
func (b *Builder) PrintClusterInfo(s *Sanitizer, name string, metrics bool) {
	if name == "" {
		name = "n/a"
	}
	s.Open(Titleize(fmt.Sprintf("Cluster [%s]", name), -1), nil)
	{
		s.Print(issues.OkLevel, 1, "Connectivity")
		if metrics {
			s.Print(issues.OkLevel, 1, "MetricServer")
		} else {
			s.Print(issues.ErrorLevel, 1, "MetricServer")
		}
	}
	s.Close()
}

// PrintHeader prints report header to screen.
func (b *Builder) PrintHeader(s *Sanitizer) {
	fmt.Fprintln(s)
	for i, l := range Logo {
		switch {
		case i < len(Popeye):
			fmt.Fprintf(s, s.Color(Popeye[i], ColorAqua))
			fmt.Fprintf(s, strings.Repeat(" ", 53))
		case i == 4:
			fmt.Fprintf(s, s.Color("  Biffs`em and Buffs`em!", ColorLighSlate))
			fmt.Fprintf(s, strings.Repeat(" ", 56))
		default:
			fmt.Fprintf(s, strings.Repeat(" ", 80))
		}
		fmt.Fprintln(s, s.Color(l, ColorLighSlate))
	}
	fmt.Fprintln(s, "")
}

// PrintReport prints out sanitizer report to screen
func (b *Builder) PrintReport(level issues.Level, s *Sanitizer) {
	for _, section := range b.Report.Sections {
		var any bool
		s.Open(Titleize(section.Title, len(section.Outcome)), section.Tally)
		{
			keys := make([]string, 0, len(section.Outcome))
			for k := range section.Outcome {
				keys = append(keys, k)
			}
			sort.Strings(keys)

			for _, res := range keys {
				ii := section.Outcome[res]
				if len(ii) == 0 {
					if level <= issues.OkLevel {
						any = true
						s.Print(issues.OkLevel, 1, res)
					}
					continue
				}
				max := section.Outcome.MaxSeverity(res)
				if level <= max {
					any = true
					s.Print(max, 1, res)
				}
				s.Dump(level, ii)
			}
			if !any {
				s.Comment(s.Color("Nothing to report.", ColorAqua))
			}
		}
		s.Close()
	}
}
