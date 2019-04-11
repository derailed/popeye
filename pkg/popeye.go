package pkg

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/derailed/popeye/internal/k8s"
	"github.com/derailed/popeye/internal/linter"
	"github.com/derailed/popeye/internal/report"
	"github.com/derailed/popeye/pkg/config"
	"github.com/rs/zerolog"
)

// PopeyeLog file path to our logs.
var PopeyeLog = filepath.Join(os.TempDir(), fmt.Sprintf("popeye.log"))

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
		out          *os.File
		log          *zerolog.Logger
		flags        *k8s.Flags
	}
)

// NewPopeye returns a new sanitizer.
func NewPopeye(flags *k8s.Flags, log *zerolog.Logger, out *os.File) (*Popeye, error) {
	cfg, err := config.NewConfig(flags)
	if err != nil {
		return nil, err
	}

	f := linter.NewFilter(k8s.NewClient(flags), cfg)

	return &Popeye{loader: f, log: log, out: out, flags: flags}, nil
}

// Sanitize scans a cluster for potential issues.
func (p *Popeye) Sanitize(showHeader bool) {
	w := bufio.NewWriter(p.out)
	defer w.Flush()

	s := report.NewSanitizer(w, p.out.Fd(), p.flags.Jurassic)

	if showHeader {
		p.printHeader(s)
		p.clusterInfo(s, p.loader)
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
			s.Open(report.TitleForRes(k), nil)
			{
				s.Error("Scan failed!", err)
			}
			s.Close()
			continue
		}
		p.printReport(s, l, report.TitleForRes(k))
	}
	p.printSummary(s)
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

func (p *Popeye) printReport(s *report.Sanitizer, r Reporter, section string) {
	level := linter.Level(p.loader.LinterLevel())
	t, any := report.NewTally().Rollup(r.Issues()), false

	s.Open(section, t)
	{
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
					s.Print(linter.OkLevel, 1, res)
				}
				continue
			}
			max := r.MaxSeverity(res)
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

	if t.IsValid() {
		p.sectionCount++
		p.totalScore += t.Score()
	}
}

func (p *Popeye) printSummary(s *report.Sanitizer) {
	if p.sectionCount == 0 {
		return
	}

	s.Open("SUMMARY", nil)
	{
		score := p.totalScore / p.sectionCount
		fmt.Fprintf(s, "Your cluster score: %d -- %s\n", score, report.Grade(score))
		for _, l := range s.Badge(score) {
			fmt.Fprintf(s, "%s%s\n", strings.Repeat(" ", report.Width-20), l)
		}
	}
	s.Close()
}

func (p *Popeye) printHeader(s *report.Sanitizer) {
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

func (p *Popeye) clusterInfo(s *report.Sanitizer, l linter.Loader) {
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
