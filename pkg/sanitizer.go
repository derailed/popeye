package pkg

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/derailed/popeye/internal/linter"
	"github.com/derailed/popeye/internal/report"
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
	// ReportBuilder represents sanitizer report.
	ReportBuilder struct {
		Report Report `yaml:"popeye"`
	}

	// Report represents the output of a sanitization pass.
	Report struct {
		Score         int       `yaml:"score"`
		Grade         string    `yaml:"grade"`
		Sections      []Section `yaml:"sanitizers"`
		sectionsCount int
		totalScore    int
		Errors        []error `yaml:"errors,omitempty"`
	}

	// Section represents a sanitizer pass
	Section struct {
		Title  string        `yaml:"sanitizer"`
		Tally  *report.Tally `yaml:"tally"`
		Issues linter.Issues `yaml:"issues,omitempty"`
	}
)

// NewReportBuilder returns a new sanitizer report.
func NewReportBuilder() *ReportBuilder {
	return &ReportBuilder{}
}

// AddError record an error associted with the report.
func (r *ReportBuilder) AddError(err error) {
	r.Report.Errors = append(r.Report.Errors, err)
}

// AddSection adds a sanitizer section to the report
func (r *ReportBuilder) AddSection(name string, issues linter.Issues, t *report.Tally) {
	n := strings.ToLower(name)
	section := Section{
		Title:  n,
		Tally:  t,
		Issues: make(linter.Issues, len(issues)),
	}

	for k, v := range issues {
		section.Issues[k] = v
	}

	r.Report.Sections = append(r.Report.Sections, section)

	if t.IsValid() {
		r.Report.sectionsCount++
		r.Report.totalScore += t.Score()
	}
}

// ToYAML dumps sanitizer to YAML.
func (r *ReportBuilder) ToYAML() (string, error) {
	score := r.Report.totalScore / r.Report.sectionsCount
	r.Report.Score = score
	r.Report.Grade = report.Grade(score)

	raw, err := yaml.Marshal(r)
	if err != nil {
		return "", err
	}

	return string(raw), nil
}

// ToJSON dumps sanitizer to JSON.
func (r *ReportBuilder) ToJSON() (string, error) {
	score := r.Report.totalScore / r.Report.sectionsCount
	r.Report.Score = score
	r.Report.Grade = report.Grade(score)

	raw, err := json.Marshal(r)
	if err != nil {
		return "", err
	}

	return string(raw), nil
}

// PrintSummary print outs summary report to screen.
func (r *ReportBuilder) PrintSummary(s *report.Sanitizer) {
	if r.Report.sectionsCount == 0 {
		return
	}

	s.Open("SUMMARY", nil)
	{
		score := r.Report.totalScore / r.Report.sectionsCount
		fmt.Fprintf(s, "Your cluster score: %d -- %s\n", score, report.Grade(score))
		for _, l := range s.Badge(score) {
			fmt.Fprintf(s, "%s%s\n", strings.Repeat(" ", report.Width-20), l)
		}
	}
	s.Close()
}

// ClusterInfo dumps cluster information to screen.
func (r *ReportBuilder) ClusterInfo(s *report.Sanitizer, l linter.Loader) {
	t := fmt.Sprintf("CLUSTER [%s]", strings.ToUpper(l.ActiveCluster()))
	s.Open(t, nil)
	{
		s.Print(linter.OkLevel, 1, "Connectivity")

		ok, err := l.ClusterHasMetrics()
		if err != nil {
			fmt.Printf("💥 %s\n", s.Color(err.Error(), report.ColorRed))
			os.Exit(1)
		}

		if ok {
			s.Print(linter.OkLevel, 1, "Metrics")
		}
	}
	s.Close()
}

// PrintHeader prints report header to screen.
func (r *ReportBuilder) PrintHeader(s *report.Sanitizer) {
	fmt.Fprintln(s)
	for i, l := range report.Logo {
		switch {
		case i < len(report.Popeye):
			fmt.Fprintf(s, s.Color(report.Popeye[i], report.ColorAqua))
			fmt.Fprintf(s, strings.Repeat(" ", 53))
		case i == 4:
			fmt.Fprintf(s, s.Color("  Biffs`em and Buffs`em!", report.ColorLighSlate))
			fmt.Fprintf(s, strings.Repeat(" ", 56))
		default:
			fmt.Fprintf(s, strings.Repeat(" ", 80))
		}
		fmt.Fprintln(s, s.Color(l, report.ColorLighSlate))
	}
	fmt.Fprintln(s, "")
}

// PrintReport prints out sanitizer report to screen
func (r *ReportBuilder) PrintReport(level linter.Level, s *report.Sanitizer) {
	for _, section := range r.Report.Sections {
		var any bool
		s.Open(report.Titleize(section.Title), section.Tally)
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
				s.Comment(s.Color("Nothing to report.", report.ColorAqua))
			}
		}
		s.Close()
	}
}
