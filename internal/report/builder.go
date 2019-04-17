package report

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/derailed/popeye/internal/linter"
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
		Title  string        `json:"sanitizer" yaml:"sanitizer"`
		Tally  *Tally        `json:"tally" yaml:"tally"`
		Issues linter.Issues `json:"issues,omitempty" yaml:"issues,omitempty"`
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
func (b *Builder) AddSection(name string, issues linter.Issues, t *Tally) {
	n := strings.ToLower(name)
	section := Section{
		Title:  n,
		Tally:  t,
		Issues: make(linter.Issues, len(issues)),
	}

	for k, v := range issues {
		section.Issues[k] = v
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

// ClusterInfo dumps cluster information to screen.
func (b *Builder) ClusterInfo(s *Sanitizer, l linter.Loader) {
	t := fmt.Sprintf("CLUSTER [%s]", strings.ToUpper(l.ActiveCluster()))
	s.Open(t, nil)
	{
		s.Print(linter.OkLevel, 1, "Connectivity")

		ok, err := l.ClusterHasMetrics()
		if err != nil {
			fmt.Printf("💥 %s\n", s.Color(err.Error(), ColorRed))
			os.Exit(1)
		}

		if ok {
			s.Print(linter.OkLevel, 1, "Metrics")
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
func (b *Builder) PrintReport(level linter.Level, s *Sanitizer) {
	for _, section := range b.Report.Sections {
		var any bool
		s.Open(Titleize(section.Title), section.Tally)
		{
			keys := make([]string, 0, len(section.Issues))
			for k := range section.Issues {
				keys = append(keys, k)
			}
			sort.Strings(keys)

			for _, res := range keys {
				issues := section.Issues[res]
				if len(issues) == 0 {
					if level <= linter.OkLevel {
						any = true
						s.Print(linter.OkLevel, 1, res)
					}
					continue
				}
				max := section.Issues.MaxSeverity(res)
				if level <= max {
					any = true
					s.Print(max, 1, res)
				}
				s.Dump(level, issues...)
			}
			if !any {
				s.Comment(s.Color("Nothing to ", ColorAqua))
			}
		}
		s.Close()
	}
}
