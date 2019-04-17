package pkg

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/derailed/popeye/internal/k8s"
	"github.com/derailed/popeye/internal/linter"
	"github.com/derailed/popeye/internal/report"
	"github.com/derailed/popeye/pkg/config"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
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
		outputTarget *os.File
		log          *zerolog.Logger
		flags        *k8s.Flags
		builder      *report.Builder
	}
)

// NewPopeye returns a new sanitizer.
func NewPopeye(flags *k8s.Flags, log *zerolog.Logger, out *os.File) (*Popeye, error) {
	cfg, err := config.NewConfig(flags)
	if err != nil {
		return nil, err
	}

	f := linter.NewFilter(k8s.NewClient(flags), cfg)

	p := Popeye{
		loader:       f,
		log:          log,
		outputTarget: out,
		flags:        flags,
		builder:      report.NewBuilder(),
	}

	return &p, nil
}

// Dump prints out sanitizer report.
func (p *Popeye) dump(printHeader bool) {
	var jurassicMode bool

	switch p.flags.OutputFormat() {
	case report.YAMLFormat:
		res, err := p.builder.ToYAML()
		if err != nil {
			fmt.Printf("Boom! %v\n", err)
			log.Fatal().Err(err).Msg("Unable to dump YAML report")
		}
		fmt.Printf("%v\n", res)
	case report.JSONFormat:
		res, err := p.builder.ToJSON()
		if err != nil {
			fmt.Printf("Boom! %v\n", err)
			log.Fatal().Err(err).Msg("Unable to dump JSON report")
		}
		fmt.Printf("%v\n", res)
	case report.JurassicFormat:
		jurassicMode = true
		fallthrough
	default:
		w := bufio.NewWriter(p.outputTarget)
		defer w.Flush()
		s := report.NewSanitizer(w, p.outputTarget.Fd(), jurassicMode)
		if printHeader {
			p.builder.PrintHeader(s)
			p.builder.ClusterInfo(s, p.loader)
		}
		p.builder.PrintReport(linter.Level(p.loader.LinterLevel()), s)
		p.builder.PrintSummary(s)
	}
}

// Sanitize scans a cluster for potential issues.
func (p *Popeye) Sanitize() {
	p.sanitize()
	p.dump(true)
}

func linters(l linter.Loader, log *zerolog.Logger) Linters {
	return Linters{
		"cm":  linter.NewConfigMap(l, log),
		"dp":  linter.NewDeployment(l, log),
		"hpa": linter.NewHorizontalPodAutoscaler(l, log),
		"ns":  linter.NewNamespace(l, log),
		"no":  linter.NewNode(l, log),
		"pv":  linter.NewPersistentVolume(l, log),
		"pvc": linter.NewPersistentVolumeClaim(l, log),
		"po":  linter.NewPod(l, log),
		"sec": linter.NewSecret(l, log),
		"svc": linter.NewService(l, log),
		"sa":  linter.NewServiceAccount(l, log),
		"sts": linter.NewStatefulSet(l, log),
	}
}

func (p *Popeye) sanitize() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	for k, l := range linters(p.loader, p.log) {
		if !in(p.loader.Sections(), k) {
			continue
		}

		// Skip node checks if active namespace is set.
		if k == "no" && p.loader.ActiveNamespace() != "" {
			continue
		}

		if err := l.Lint(ctx); err != nil {
			p.builder.AddError(err)
			continue
		}

		tally := report.NewTally()
		tally.Rollup(l.Issues())
		p.builder.AddSection(k, l.Issues(), tally)
	}
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
