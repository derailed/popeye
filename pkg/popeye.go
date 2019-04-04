package pkg

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/derailed/popeye/internal/config"
	"github.com/derailed/popeye/internal/k8s"
	"github.com/derailed/popeye/internal/linter"
	"github.com/derailed/popeye/internal/report"
	"github.com/rs/zerolog/log"
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
		config *config.Config
	}
)

// New returns a new sanitizer.
func New(c *config.Config) *Popeye {
	return &Popeye{config: c}
}

func linters(c *k8s.Client) Linters {
	return Linters{
		"no":  linter.NewNode(c, &log.Logger),
		"ns":  linter.NewNamespace(c, &log.Logger),
		"po":  linter.NewPod(c, &log.Logger),
		"svc": linter.NewService(c, &log.Logger),
	}
}

// Sanitize scans a cluster for potential issues.
func (p *Popeye) Sanitize() {
	c := k8s.NewClient(p.config)

	p.clusterInfo(c)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	for k, v := range linters(c) {
		if !in(p.config.Sections, k) {
			continue
		}

		if err := v.Lint(ctx); err != nil {
			w := bufio.NewWriter(os.Stdout)
			defer w.Flush()
			report.Open(w, report.TitleForRes(k), nil)
			{
				report.Error(w, "Scan failed!", err)
			}
			report.Close(w)
			continue
		}

		p.printReport(v, report.TitleForRes(k))
	}
}

func (p *Popeye) printReport(r Reporter, section string) {
	w := bufio.NewWriter(os.Stdout)
	defer w.Flush()

	level := linter.Level(p.config.Popeye.LintLevel)
	var any bool
	report.Open(w, section, r.Issues())
	{
		w.Flush()
		for res, issues := range r.Issues() {
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
			report.Comment(w, report.Colorize("Nothing to report.", report.ColorOrangish))
		}
	}
	report.Close(w)
}

func (p *Popeye) clusterInfo(c *k8s.Client) {
	t := fmt.Sprintf("CLUSTER [%s]", strings.ToUpper(c.Config.ActiveCluster()))
	report.Open(os.Stdout, t, nil)
	{
		report.Write(os.Stdout, linter.OkLevel, 1, "Connectivity")

		if !c.ClusterHasMetrics() {
			report.Write(os.Stdout, linter.OkLevel, 1, "Metrics")
		} else {
			report.Write(os.Stdout, linter.OkLevel, 1, "Metrics")
		}
	}
	report.Close(os.Stdout)
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
