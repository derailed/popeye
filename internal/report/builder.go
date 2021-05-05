package report

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"text/template"

	"github.com/derailed/popeye/internal/client"
	"github.com/derailed/popeye/internal/issues"
	"github.com/derailed/popeye/pkg/config"
	"github.com/fvbommel/sortorder"
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

	// HTMLFormat dumps sanitizer as HTML
	HTMLFormat = "html"

	// JunitFormat dumps sanitizer as JUnit report.
	JunitFormat = "junit"

	// PrometheusFormat pushes sanitizer as Prometheus metrics.
	PrometheusFormat = "prometheus"

	// ScoreFormat pushes sanitizer as the value of the Score.
	ScoreFormat = "score"
)

// Builder represents sanitizer
type Builder struct {
	Report      Report `json:"popeye" yaml:"popeye"`
	clusterName string
}

// Report represents the output of a sanitization pass.
type Report struct {
	Score         int      `json:"score" yaml:"score"`
	Grade         string   `json:"grade" yaml:"grade"`
	Sections      Sections `json:"sanitizers,omitempty" yaml:"sanitizers,omitempty"`
	Errors        []error  `json:"errors,omitempty" yaml:"errors,omitempty"`
	sectionsCount int
	totalScore    int
}

// Sections represents a collection of sections.
type Sections []Section

// Section represents a sanitizer pass
type Section struct {
	Title    string         `json:"sanitizer" yaml:"sanitizer"`
	GVR      string         `json:"gvr" yaml:"gvr"`
	Tally    *Tally         `json:"tally" yaml:"tally"`
	Outcome  issues.Outcome `json:"issues,omitempty" yaml:"issues,omitempty"`
	singular string
}

// Len returns the list size.
func (s Sections) Len() int {
	return len(s)
}

// Swap swaps list values.
func (s Sections) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

// Less returns true if i < j.
func (s Sections) Less(i, j int) bool {
	return sortorder.NaturalLess(s[i].singular, s[j].singular)
}

// NewBuilder returns a new instance.
func NewBuilder() *Builder {
	return &Builder{}
}

// SetClusterName sets the current cluster name.
func (b *Builder) SetClusterName(s string) {
	sort.Sort(b.Report.Sections)
	b.clusterName = s
}

// ClusterName returns the cluster name.
func (b *Builder) ClusterName() string {
	return b.clusterName
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
func (b *Builder) AddSection(gvr client.GVR, singular string, o issues.Outcome, t *Tally) {
	section := Section{
		Title:    strings.ToLower(gvr.R()),
		GVR:      gvr.String(),
		singular: singular,
		Tally:    t,
		Outcome:  o,
	}
	b.Report.Sections = append(b.Report.Sections, section)
	if t.IsValid() {
		b.Report.sectionsCount++
		b.Report.totalScore += t.Score()
	}
}

// ToJunit dumps sanitizer to JUnit.
func (b *Builder) ToJunit(level config.Level) (string, error) {
	b.finalize()
	raw, err := junitMarshal(b, level)
	if err != nil {
		return "", err
	}

	return string(raw), nil
}

func (b *Builder) finalize() {
	score := b.Report.totalScore / b.Report.sectionsCount
	b.Report.Score = score
	b.Report.Grade = Grade(score)
}

// ToYAML dumps sanitizer to YAML.
func (b *Builder) ToYAML() (string, error) {
	b.finalize()
	raw, err := yaml.Marshal(b)
	if err != nil {
		return "", err
	}

	return string(raw), nil
}

// ToJSON dumps sanitizer to JSON.
func (b *Builder) ToJSON() (string, error) {
	b.finalize()
	raw, err := json.Marshal(b)
	if err != nil {
		return "", err
	}

	return string(raw), nil
}

// ToHTML dumps sanitizer to HTML.
func (b *Builder) ToHTML() (string, error) {
	b.finalize()

	fMap := template.FuncMap{
		"toEmoji": toEmoji,
		"toTitle": Titleize,
		"isRoot":  isRoot,
	}
	tpl, err := template.New("sanitize").Funcs(fMap).Parse(htmlTemplate)
	if err != nil {
		return "", err
	}

	buff := bytes.NewBufferString("")
	if err := tpl.Execute(buff, b); err != nil {
		return "", err
	}

	return buff.String(), nil
}

// ToPrometheus returns prometheus pusher.
func (b *Builder) ToPrometheus(gtwy *config.PushGateway, namespace string) *push.Pusher {
	b.finalize()
	if namespace == "" {
		namespace = "all"
	}

	return prometheusMarshal(b, gtwy, b.clusterName, namespace)
}

// ToScore dumps sanitizer to only the score value.
func (b *Builder) ToScore() (int, error) {
	b.finalize()
	return b.Report.Score, nil
}

// PrintSummary print outs summary report to screen.
func (b *Builder) PrintSummary(s *Sanitizer) {
	if b.Report.sectionsCount == 0 {
		return
	}

	b.finalize()
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
func (b *Builder) PrintClusterInfo(s *Sanitizer, clusterName string, metrics bool) {
	if clusterName == "" {
		clusterName = "n/a"
	}
	s.Open(Titleize(fmt.Sprintf("General [%s]", clusterName), -1), nil)
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

// Titleize renders a section title.
func Titleize(res string, count int) string {
	if count < 0 {
		return strings.ToUpper(res)
	}
	return strings.ToUpper(fmt.Sprintf("%s (%d scanned)", res, count))
}

func isRoot(g string) bool {
	return g == issues.Root
}

func toEmoji(level config.Level) (s string) {
	switch level {
	case config.ErrorLevel:
		s = "fas fa-bomb"
	case config.WarnLevel:
		s = "fas fa-radiation-alt"
	case config.InfoLevel:
		s = "fas fa-info-circle"
	case config.OkLevel:
		s = "far fa-check-circle"
	default:
		s = "fas fa-info-circle"
	}
	return s
}
