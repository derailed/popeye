package pkg

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/derailed/popeye/internal/issues"
	"github.com/derailed/popeye/internal/k8s"
	"github.com/derailed/popeye/internal/report"
	"github.com/derailed/popeye/internal/scrub"
	"github.com/derailed/popeye/pkg/config"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// PopeyeLog file path to our logs.
var PopeyeLog = filepath.Join(os.TempDir(), fmt.Sprintf("popeye.log"))

type (
	scrubFn func(*scrub.Cache) scrub.Sanitizer

	// Popeye a kubernetes sanitizer.
	Popeye struct {
		client       *k8s.Client
		config       *config.Config
		totalScore   int
		sectionCount int
		outputTarget *os.File
		log          *zerolog.Logger
		flags        *config.Flags
		builder      *report.Builder
	}
)

// NewPopeye returns a new sanitizer.
func NewPopeye(flags *config.Flags, log *zerolog.Logger, out *os.File) (*Popeye, error) {
	cfg, err := config.NewConfig(flags)
	if err != nil {
		return nil, err
	}

	p := Popeye{
		client:       k8s.NewClient(flags),
		config:       cfg,
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
		}
		p.builder.PrintReport(issues.Level(p.config.LinterLevel()), s)
		p.builder.PrintSummary(s)
	}
}

// Sanitize scans a cluster for potential issues.
func (p *Popeye) Sanitize() {
	p.sanitize()
	p.dump(true)
}

func (p *Popeye) sanitizers() map[string]scrubFn {
	return map[string]scrubFn{
		"cm":  scrub.NewConfigMap,
		"sec": scrub.NewSecret,
		"dp":  scrub.NewDeployment,
		"hpa": scrub.NewHorizontalPodAutoscaler,
		"ns":  scrub.NewNamespace,
		"no":  scrub.NewNode,
		"pv":  scrub.NewPersistentVolume,
		"pvc": scrub.NewPersistentVolumeClaim,
		"po":  scrub.NewPod,
		"svc": scrub.NewService,
		"sa":  scrub.NewServiceAccount,
		"sts": scrub.NewStatefulSet,
	}
}

func (p *Popeye) sanitize() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cache := scrub.NewCache(p.client, p.config)
	for k, f := range p.sanitizers() {
		if !in(p.config.Sections(), k) {
			continue
		}

		// Skip node checks if active namespace is set.
		if k == "no" && p.client.ActiveNamespace() != "" {
			continue
		}

		s := f(cache)
		if err := s.Sanitize(ctx); err != nil {
			p.builder.AddError(err)
			continue
		}

		tally := report.NewTally()
		tally.Rollup(s.Outcome())
		p.builder.AddSection(k, s.Outcome(), tally)
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
