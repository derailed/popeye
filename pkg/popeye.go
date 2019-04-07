package pkg

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/derailed/popeye/internal/k8s"
	"github.com/derailed/popeye/internal/linter"
	"github.com/derailed/popeye/internal/report"
	"github.com/derailed/popeye/pkg/config"
	"github.com/rs/zerolog"
)

type (
	// Reporter obtains lint reports
	Reporter interface {
		MaxSeverity(res string) linter.Level
		Issues() linter.Issues
	}

	// Linter represents a resource linter.
	Linter interface {
		Reporter
		Lint(context.Context) error
	}

	// Linters a collection of linters.
	Linters map[string]Linter

	// Popeye a kubernetes sanitizer.
	Popeye struct {
		config       *config.Config
		totalScore   int
		sectionCount int
		out          io.Writer
		log          *zerolog.Logger
	}
)

// NewPopeye returns a new sanitizer.
func NewPopeye(c *config.Config, log *zerolog.Logger, out io.Writer) *Popeye {
	return &Popeye{config: c, out: out, log: log}
}

// Sanitize scans a cluster for potential issues.
func (p *Popeye) Sanitize(showHeader bool) {
	c := k8s.NewClient(p.config)

	if showHeader {
		p.printHeader()
		p.clusterInfo(c)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	for k, l := range linters(c, p.log) {
		if !in(p.config.Sections, k) {
			continue
		}

		// Skip no check if active namespace is set.
		if k == "no" && p.config.ActiveNamespace() != "" {
			continue
		}

		if err := l.Lint(ctx); err != nil {
			w := bufio.NewWriter(p.out)
			defer w.Flush()
			report.Open(w, report.TitleForRes(k), nil)
			{
				report.Error(w, "Scan failed!", err)
			}
			report.Close(w)
			continue
		}
		p.printReport(l, report.TitleForRes(k))
	}
	p.printSummary()
}

func linters(c *k8s.Client, log *zerolog.Logger) Linters {
	return Linters{
		"no":  linter.NewNode(c, log),
		"ns":  linter.NewNamespace(c, log),
		"po":  linter.NewPod(c, log),
		"svc": linter.NewService(c, log),
		"sa":  linter.NewSA(c, log),
		"cm":  linter.NewCM(c, log),
	}
}

func (p *Popeye) printSummary() {
	w := bufio.NewWriter(p.out)
	defer w.Flush()

	report.Open(w, "SUMMARY", nil)
	{
		s := p.totalScore / p.sectionCount
		fmt.Fprintf(w, "Your cluster score: %d -- %s\n", s, report.Grade(s))
		for _, l := range report.Badge(s) {
			fmt.Fprintf(w, "%s%s\n", strings.Repeat(" ", 60), l)
		}
	}
	report.Close(w)
}

func (p *Popeye) printReport(r Reporter, section string) {
	w := bufio.NewWriter(p.out)
	defer w.Flush()

	level := linter.Level(p.config.Popeye.LintLevel)
	t, any := report.NewTally().Rollup(r.Issues()), false
	report.Open(w, section, t)
	{
		w.Flush()
		keys := make([]string, 0, len(r.Issues()))
		for k := range r.Issues() {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, res := range keys {
			issues := r.Issues()[res]
			if len(issues) == 0 {
				if level <= linter.OkLevel {
					p.sectionCount++
					any = true
					report.Write(w, linter.OkLevel, 1, res)
				}
				continue
			}
			max := r.MaxSeverity(res)
			if level <= max {
				any = true
				p.sectionCount++
				report.Write(w, max, 1, res)
			}
			report.Dump(w, level, issues...)
		}
		if !any {
			report.Comment(w, report.Colorize("Nothing to report.", report.ColorOrangish))
		}
	}
	report.Close(w)

	if t.IsValid() {
		p.totalScore += t.Score()
	}
}

func (p *Popeye) clusterInfo(c *k8s.Client) {
	w := bufio.NewWriter(p.out)
	defer w.Flush()

	t := fmt.Sprintf("CLUSTER [%s]", strings.ToUpper(c.Config.ActiveCluster()))
	report.Open(w, t, nil)
	{
		report.Write(w, linter.OkLevel, 1, "Connectivity")

		if !c.ClusterHasMetrics() {
			report.Write(w, linter.OkLevel, 1, "Metrics")
		} else {
			report.Write(w, linter.OkLevel, 1, "Metrics")
		}
	}
	report.Close(w)
}

func (p *Popeye) printHeader() {
	w := bufio.NewWriter(p.out)
	defer w.Flush()

	fmt.Fprintln(w)
	for i, s := range report.Logo {
		if i < len(report.Popeye) {
			fmt.Fprintf(w, report.Colorize(report.Popeye[i], report.ColorAqua))
			fmt.Fprintf(w, strings.Repeat(" ", 35))
		} else {
			if i == 4 {
				fmt.Fprintf(w, report.Colorize("  Biffs`em and Buffs`em!", report.ColorLighSlate))
				fmt.Fprintf(w, strings.Repeat(" ", 38))
			} else {
				fmt.Fprintf(w, strings.Repeat(" ", 62))
			}
		}
		fmt.Fprintln(w, report.Colorize(s, report.ColorLighSlate))
	}
	fmt.Fprintln(w, "")
}

func in(list []string, member string) bool {
	if len(list) == 0 {
		return true
	}

	for _, m := range list {
		if m == member {
			return true
		}
	}

	return false
}
