package pkg

import (
	"bufio"
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/derailed/popeye/internal/issues"
	"github.com/derailed/popeye/internal/k8s"
	"github.com/derailed/popeye/internal/report"
	"github.com/derailed/popeye/internal/sanitize"
	"github.com/derailed/popeye/internal/scrub"
	"github.com/derailed/popeye/pkg/config"
	"github.com/rs/zerolog"
)

var (
	// LogFile the path to our logs.
	LogFile = filepath.Join(os.TempDir(), fmt.Sprintf("popeye.log"))
	// DumpDir indicates a directory location for sanitixer reports.
	DumpDir = filepath.Join(os.TempDir(), "popeye")
)

type (
	scrubFn func(*scrub.Cache, *issues.Codes) scrub.Sanitizer

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
func NewPopeye(flags *config.Flags, log *zerolog.Logger) (*Popeye, error) {
	cfg, err := config.NewConfig(flags)
	if err != nil {
		return nil, err
	}

	p := Popeye{
		client:  k8s.NewClient(flags),
		config:  cfg,
		log:     log,
		flags:   flags,
		builder: report.NewBuilder(),
	}

	return &p, nil
}

// Init configures popeye prior to sanitization.
func (p *Popeye) Init() error {
	if !isSet(p.flags.Save) {
		return p.ensureOutput()
	}

	if err := ensurePath(DumpDir, 0755); err != nil {
		return err
	}
	return p.ensureOutput()
}

// Sanitize scans a cluster for potential issues.
func (p *Popeye) Sanitize() error {
	defer func() {
		if p.outputTarget != os.Stdout {
			p.outputTarget.Close()
		}
	}()

	if err := p.sanitize(); err != nil {
		return err
	}
	return p.dump(true)
}

// Dump prints out sanitizer report.
func (p *Popeye) dump(printHeader bool) error {
	var jurassicMode bool

	if !p.builder.HasContent() {
		return errors.New("Nothing to report, check section name or permissions")
	}

	switch p.flags.OutputFormat() {
	case report.JunitFormat:
		res, err := p.builder.ToJunit()
		if err != nil {
			// log.().Err(err).Msg("Unable to dump Junit report")
			return err
		}
		p.outputTarget.WriteString(xml.Header)
		fmt.Fprintf(p.outputTarget, "%v\n", res)
	case report.YAMLFormat:
		res, err := p.builder.ToYAML()
		if err != nil {
			// log.Fatal().Err(err).Msg("Unable to dump YAML report")
			return err
		}
		fmt.Fprintf(p.outputTarget, "%v\n", res)
	case report.JSONFormat:
		res, err := p.builder.ToJSON()
		if err != nil {
			// log.Fatal().Err(err).Msg("Unable to dump JSON report")
			return err
		}
		fmt.Fprintf(p.outputTarget, "%v\n", res)
	case report.PrometheusFormat:
		pusher := p.builder.ToPrometheus(p.flags.PushGatewayAddress,
			p.client.ActiveCluster(), p.client.ActiveNamespace())
		return pusher.Add()
	case report.JurassicFormat:
		jurassicMode = true
		fallthrough
	default:
		w := bufio.NewWriter(p.outputTarget)
		defer w.Flush()

		s := report.NewSanitizer(w, jurassicMode)
		if printHeader {
			p.builder.PrintHeader(s)
		}
		mx, err := p.client.ClusterHasMetrics()
		if err != nil {
			mx = false
		}
		p.builder.PrintClusterInfo(s, p.client.ActiveCluster(), mx)
		p.builder.PrintReport(issues.Level(p.config.LinterLevel()), s)
		p.builder.PrintSummary(s)
	}

	return nil
}

func (p *Popeye) sanitizers() map[string]scrubFn {
	return map[string]scrubFn{
		"cl":  scrub.NewCluster,
		"cm":  scrub.NewConfigMap,
		"sec": scrub.NewSecret,
		"dp":  scrub.NewDeployment,
		"ds":  scrub.NewDaemonSet,
		"hpa": scrub.NewHorizontalPodAutoscaler,
		"ns":  scrub.NewNamespace,
		"no":  scrub.NewNode,
		"pv":  scrub.NewPersistentVolume,
		"pvc": scrub.NewPersistentVolumeClaim,
		"po":  scrub.NewPod,
		"rs":  scrub.NewReplicaSet,
		"svc": scrub.NewService,
		"sa":  scrub.NewServiceAccount,
		"sts": scrub.NewStatefulSet,
		"pdb": scrub.NewPodDisruptionBudget,
		"ing": scrub.NewIngress,
		"np":  scrub.NewNetworkPolicy,
		"psp": scrub.NewPodSecurityPolicy,
	}
}

func (p *Popeye) ensureOutput() error {
	p.outputTarget = os.Stdout
	if !isSet(p.flags.Save) {
		return nil
	}

	if p.flags.Output == nil {
		*p.flags.Output = "standard"
	}

	ext := "txt"
	switch *p.flags.Output {
	case "json":
		ext = "json"
	case "junit":
		ext = "xml"
	case "yaml":
		ext = "yml"
	}

	const outFmt = "sanitizer_%s_%d.%s"
	f := filepath.Join(DumpDir, fmt.Sprintf(outFmt, p.client.ActiveCluster(), time.Now().UnixNano(), ext))
	var err error
	if p.outputTarget, err = os.Create(f); err != nil {
		return err
	}

	fmt.Printf("Sanitizer saved to: %s\n", f)
	return nil
}

func (p *Popeye) sanitize() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ctx = context.WithValue(
		ctx,
		sanitize.PopeyeKey("OverAllocs"),
		*p.flags.CheckOverAllocs,
	)

	cache := scrub.NewCache(p.client, p.config)
	codes, err := issues.LoadCodes()
	if err != nil {
		return err
	}
	codes.Refine(p.config.Codes)
	for k, f := range p.sanitizers() {
		if !in(p.config.Sections(), k) {
			continue
		}
		// Skip node checks if active namespace is set.
		if k == "no" && p.client.ActiveNamespace() != "" {
			continue
		}
		s := f(cache, codes)
		if err := s.Sanitize(ctx); err != nil {
			p.builder.AddError(err)
			continue
		}
		tally := report.NewTally()
		tally.Rollup(s.Outcome())
		p.builder.AddSection(k, s.Outcome(), tally)
	}

	return nil
}

// ----------------------------------------------------------------------------
// Helpers...

func isSet(b *bool) bool {
	return b != nil && *b
}

func ensurePath(path string, mod os.FileMode) error {
	dir, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	_, err = os.Stat(dir)
	if err == nil || !os.IsNotExist(err) {
		return nil
	}

	if err = os.MkdirAll(dir, mod); err != nil {
		return fmt.Errorf("Fail to create popeye sanitizers dump dir: %v", err)
	}
	return nil
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
