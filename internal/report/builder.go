package report

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/derailed/popeye/internal"
	"github.com/derailed/popeye/internal/issues"
	"github.com/derailed/popeye/pkg/config"
	"github.com/prometheus/client_golang/prometheus/push"
	"gopkg.in/yaml.v2"
)

const (
	// DefaultFormat dumps sanitizer with color, emojis, the works.
	DefaultFormat = "standard"

	// JurassicFormat dumps sanitizer with 0 fancyness.
	JurassicFormat = "jurassic"

	// YAMLFormat dumps sanitizer as YAML.
	YAMLFormat = "yaml"

	// JSONFormat dumps sanitizer as JSON.
	JSONFormat = "json"

	// JunitFormat dumps sanitizer as JUnit report.
	JunitFormat = "junit"

	// PrometheusFormat pushes sanitizer as Prometheus metrics.
	PrometheusFormat = "prometheus"

	// ScoreFormat pushes sanitizer as the value of the Score.
	ScoreFormat = "score"
)

type (
	// Builder represents sanitizer
	Builder struct {
		Report  Report `json:"popeye" yaml:"popeye"`
		aliases *internal.Aliases
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
func NewBuilder(a *internal.Aliases) *Builder {
	return &Builder{aliases: a}
}

// HasContent checks if we actually have anything to report.
func (b *Builder) HasContent() bool {
	return b.Report.sectionsCount != 0
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

// ToJunit dumps sanitizer to JUnit.
func (b *Builder) ToJunit() (string, error) {
	b.augment()
	raw, err := junitMarshal(b)
	if err != nil {
		return "", err
	}

	return string(raw), nil
}

func (b *Builder) augment() {
	score := b.Report.totalScore / b.Report.sectionsCount
	b.Report.Score = score
	b.Report.Grade = Grade(score)

	for i, s := range b.Report.Sections {
		b.Report.Sections[i].Title = b.aliases.FromAlias(s.Title)
	}
}

// ToYAML dumps sanitizer to YAML.
func (b *Builder) ToYAML() (string, error) {
	b.augment()
	raw, err := yaml.Marshal(b)
	if err != nil {
		return "", err
	}

	return string(raw), nil
}

// ToJSON dumps sanitizer to JSON.
func (b *Builder) ToJSON() (string, error) {
	b.augment()
	raw, err := json.Marshal(b)
	if err != nil {
		return "", err
	}

	return string(raw), nil
}

// ToPrometheus returns prometheus pusher.
func (b *Builder) ToPrometheus(address *string, cluster, namespace string) *push.Pusher {
	b.augment()
	if namespace == "" {
		namespace = "all"
	}
	return prometheusMarshal(b, address, cluster, namespace)
}

// ToScore dumps sanitizer to only the score value.
func (b *Builder) ToScore() (int, error) {
	b.augment()
	return b.Report.Score, nil
}

// PrintSummary print outs summary report to screen.
func (b *Builder) PrintSummary(s *Sanitizer) {
	if b.Report.sectionsCount == 0 {
		return
	}

	b.augment()
	s.Open("SUMMARY", nil)
	{
		fmt.Fprintf(s, "Your cluster score: %d -- %s\n", b.Report.Score, b.Report.Grade)
		for _, l := range s.Badge(b.Report.Score) {
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
	s.Open(Titleize(b.aliases, fmt.Sprintf("General [%s]", name), -1), nil)
	{
		s.Print(config.OkLevel, 1, "Connectivity")
		if metrics {
			s.Print(config.OkLevel, 1, "MetricServer")
		} else {
			s.Print(config.ErrorLevel, 1, "MetricServer")
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
			fmt.Fprintf(s, "%s", s.Color(Popeye[i], ColorAqua))
			fmt.Fprintf(s, "%s", strings.Repeat(" ", 53))
		case i == 4:
			fmt.Fprintf(s, "%s", s.Color("  Biffs`em and Buffs`em!", ColorLighSlate))
			fmt.Fprintf(s, "%s", strings.Repeat(" ", 56))
		default:
			fmt.Fprintf(s, "%s", strings.Repeat(" ", 80))
		}
		fmt.Fprintln(s, s.Color(l, ColorLighSlate))
	}
	fmt.Fprintln(s, "")
}

// PrintReport prints out sanitizer report to screen
func (b *Builder) PrintReport(level config.Level, s *Sanitizer) {
	for _, section := range b.Report.Sections {
		var any bool
		s.Open(Titleize(b.aliases, section.Title, len(section.Outcome)), section.Tally)
		{
			keys := make([]string, 0, len(section.Outcome))
			for k := range section.Outcome {
				keys = append(keys, k)
			}
			sort.Strings(keys)

			for _, res := range keys {
				ii := section.Outcome[res]
				if len(ii) == 0 {
					if level <= config.OkLevel {
						any = true
						s.Print(config.OkLevel, 1, res)
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

// ----------------------------------------------------------------------------
// Helpers...

// Titleize computes a section title.
func Titleize(a *internal.Aliases, res string, count int) string {
	res = a.FromAlias(res)
	if count <= 0 || res == "general" {
		return strings.ToUpper(res)
	}

	return strings.ToUpper(fmt.Sprintf("%s (%d scanned)", a.Pluralize(res), count))
}
