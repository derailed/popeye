// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Popeye

package report

import (
	_ "embed"
	"slices"

	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"text/template"
	"time"

	"github.com/derailed/popeye/internal/issues"
	"github.com/derailed/popeye/internal/rules"
	"github.com/derailed/popeye/pkg/config"
	"github.com/derailed/popeye/types"
	"github.com/prometheus/client_golang/prometheus/push"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v2"
)

// Builder tracks a scan report output.
type Builder struct {
	Report      Report `json:"popeye" yaml:"popeye"`
	ClusterName string
	ContextName string
}

// NewBuilder returns a new instance.
func NewBuilder() *Builder {
	return &Builder{}
}

// SetClusterContext sets the current cluster name.
func (b *Builder) SetClusterContext(cl, ct string) {
	sort.Sort(b.Report.Sections)
	b.ClusterName, b.ContextName = cl, ct
	b.Report.Timestamp = time.Now().Format(time.RFC3339)
}

// HasContent checks if we actually have anything to report.
func (b *Builder) HasContent() bool {
	return b.Report.sectionsCount != 0
}

// AddError record an error associated with the report.
func (b *Builder) AddError(err error) {
	b.Report.Errors = append(b.Report.Errors, err)
}

// AddSection adds a linter section to the report.
func (b *Builder) AddSection(gvr types.GVR, singular string, o issues.Outcome, t *Tally) {
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

// ToJunit dumps scan to JUnit.
func (b *Builder) ToJunit(level rules.Level) (string, error) {
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

// ToYAML dumps scan to YAML.
func (b *Builder) ToYAML() (string, error) {
	b.finalize()
	raw, err := yaml.Marshal(b)
	if err != nil {
		return "", err
	}

	return string(raw), nil
}

// ToJSON dumps scan to JSON.
func (b *Builder) ToJSON() (string, error) {
	b.finalize()
	raw, err := json.Marshal(b)
	if err != nil {
		return "", err
	}

	return string(raw), nil
}

// ToHTML dumps scan to HTML.
func (b *Builder) ToHTML() (string, error) {
	b.finalize()

	fMap := template.FuncMap{
		"toEmoji": toEmoji,
		"toTitle": Titleize,
		"isRoot":  isRoot,
		"list":    b.Report.ListSections,
	}
	tpl, err := template.New("sanitize").Funcs(fMap).Parse(htmlReport)
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
func (b *Builder) ToPrometheus(gtwy *config.PushGateway, instance, ns, asset string, cc rules.Glossary) *push.Pusher {
	b.finalize()

	log.Debug().Msgf("Pushing prom metrics from instance: %q", instance)
	p := newPusher(gtwy, instance)
	if ns == "" {
		ns = "all"
	}
	b.promCollect(ns, asset, cc)

	return p
}

// ToScore dumps scan to only the score value.
func (b *Builder) ToScore() (int, error) {
	b.finalize()
	return b.Report.Score, nil
}

// PrintSummary print outs summary report to screen.
func (b *Builder) PrintSummary(s *ScanReport) {
	if b.Report.sectionsCount == 0 {
		return
	}

	b.finalize()
	s.Open("SUMMARY", nil)
	{
		fmt.Fprint(s, s.Color(fmt.Sprintf("%-19s %s (%d)\n", "Your cluster score:", b.Report.Grade, b.Report.Score), ColorAqua))
		for _, l := range s.Badge(b.Report.Score) {
			fmt.Fprintf(s, "%s%s\n", strings.Repeat(" ", Width-20), l)
		}
	}
	s.Close()
}

// PrintClusterInfo displays cluster information.
func (b *Builder) PrintClusterInfo(s *ScanReport, metrics bool) {
	cl := b.ClusterName
	if cl == "" {
		cl = "n/a"
	}
	s.Open(Titleize(fmt.Sprintf("General [%s] (%s)", cl, b.Report.Timestamp), -1), nil)
	{
		s.Print(rules.OkLevel, 1, "Connectivity")
		if metrics {
			s.Print(rules.OkLevel, 1, "MetricServer")
		} else {
			s.Print(rules.ErrorLevel, 1, "MetricServer")
		}
	}
	s.Close()
}

// PrintHeader prints report header to screen.
func (b *Builder) PrintHeader(s *ScanReport) {
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

// PrintReport prints out scan report to screen
func (b *Builder) PrintReport(level rules.Level, s *ScanReport) {
	for _, section := range b.Report.Sections {
		var any bool
		s.Open(Titleize(section.Title, len(section.Outcome)), section.Tally)
		{
			kk := make([]string, 0, len(section.Outcome))
			for k := range section.Outcome {
				kk = append(kk, k)
			}
			slices.SortFunc(kk, issues.SortKeys)

			for _, res := range kk {
				ii := section.Outcome[res]
				if len(ii) == 0 {
					if level <= rules.OkLevel {
						any = true
						s.Print(rules.OkLevel, 1, res)
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

func toEmoji(level rules.Level) (s string) {
	switch level {
	case rules.ErrorLevel:
		s = "fas fa-bomb"
	case rules.WarnLevel:
		s = "fas fa-radiation-alt"
	case rules.InfoLevel:
		s = "fas fa-info-circle"
	case rules.OkLevel:
		s = "far fa-check-circle"
	default:
		s = "fas fa-info-circle"
	}
	return s
}
