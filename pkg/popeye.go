package pkg

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
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
		loader       linter.Loader
		totalScore   int
		sectionCount int
		out          io.Writer
		log          *zerolog.Logger
	}
)

// NewPopeye returns a new sanitizer.
func NewPopeye(flags *k8s.Flags, log *zerolog.Logger, out io.Writer) (*Popeye, error) {
	cfg, err := config.NewConfig(flags)
	if err != nil {
		return nil, err
	}

	f := linter.NewFilter(k8s.NewClient(flags), cfg)

	return &Popeye{loader: f, out: out, log: log}, nil
}

// Sanitize scans a cluster for potential issues.
func (p *Popeye) Sanitize(showHeader bool) {
	if showHeader {
		p.printHeader()
		p.clusterInfo(p.loader)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	for k, l := range linters(p.loader, p.log) {
		if !in(p.loader.Sections(), k) {
			continue
		}

		// Skip no check if active namespace is set.
		if k == "no" && p.loader.ActiveNamespace() != "" {
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

func linters(l linter.Loader, log *zerolog.Logger) Linters {
	return Linters{
		"no":  linter.NewNode(l, log),
		"ns":  linter.NewNamespace(l, log),
		"po":  linter.NewPod(l, log),
		"svc": linter.NewService(l, log),
		"sa":  linter.NewSA(l, log),
		"cm":  linter.NewCM(l, log),
		"sec": linter.NewService(l, log),
	}
}

func (p *Popeye) printReport(r Reporter, section string) {
	w := bufio.NewWriter(p.out)
	defer w.Flush()

	level := linter.Level(p.loader.LinterLevel())
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
					any = true
					report.Write(w, linter.OkLevel, 1, res)
				}
				continue
			}
			max := r.MaxSeverity(res)
			if level <= max {
				any = true
				report.Write(w, max, 1, res)
			}
			report.Dump(w, level, issues...)
		}
		if !any {
			report.Comment(w, report.Colorize("Nothing to report.", report.ColorAqua))
		}
	}
	report.Close(w)

	if t.IsValid() {
		p.sectionCount++
		p.totalScore += t.Score()
	}
}

func (p *Popeye) printSummary() {
	if p.sectionCount == 0 {
		return
	}

	w := bufio.NewWriter(p.out)
	defer w.Flush()

	report.Open(w, "SUMMARY", nil)
	{
		s := p.totalScore / p.sectionCount
		fmt.Fprintf(w, "Your cluster score: %d -- %s\n", s, report.Grade(s))
		for _, l := range report.Badge(s) {
			fmt.Fprintf(w, "%s%s\n", strings.Repeat(" ", report.Width-20), l)
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
			fmt.Fprintf(w, strings.Repeat(" ", 55))
		} else {
			if i == 4 {
				fmt.Fprintf(w, report.Colorize("  Biffs`em and Buffs`em!", report.ColorLighSlate))
				fmt.Fprintf(w, strings.Repeat(" ", 58))
			} else {
				fmt.Fprintf(w, strings.Repeat(" ", 82))
			}
		}
		fmt.Fprintln(w, report.Colorize(s, report.ColorLighSlate))
	}
	fmt.Fprintln(w, "")
}

func (p *Popeye) clusterInfo(l linter.Loader) {
	w := bufio.NewWriter(p.out)
	defer w.Flush()

	t := fmt.Sprintf("CLUSTER [%s]", strings.ToUpper(l.ActiveCluster()))
	report.Open(w, t, nil)
	{
		report.Write(w, linter.OkLevel, 1, "Connectivity")

		ok, err := l.ClusterHasMetrics()
		if err != nil {
			fmt.Printf("💥 %s\n", report.Colorize(err.Error(), report.ColorRed))
			os.Exit(1)
		}

		if ok {
			report.Write(w, linter.OkLevel, 1, "Metrics")
		} else {
			report.Write(w, linter.OkLevel, 1, "Metrics")
		}
	}
	report.Close(w)
}

// ----------------------------------------------------------------------------
// Helpers...

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
