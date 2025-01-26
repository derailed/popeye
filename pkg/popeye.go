// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Popeye

package pkg

import (
	"bufio"
	"bytes"
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/derailed/popeye/internal"
	"github.com/derailed/popeye/internal/cilium"
	cscrub "github.com/derailed/popeye/internal/cilium/scrub"
	"github.com/derailed/popeye/internal/client"
	"github.com/derailed/popeye/internal/db"
	"github.com/derailed/popeye/internal/db/schema"
	"github.com/derailed/popeye/internal/issues"
	"github.com/derailed/popeye/internal/report"
	"github.com/derailed/popeye/internal/rules"
	"github.com/derailed/popeye/internal/scrub"
	"github.com/derailed/popeye/pkg/config"
	"github.com/derailed/popeye/types"
	"github.com/hashicorp/go-memdb"
	"github.com/prometheus/common/expfmt"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const (
	dumpFileFMT        = "popeye-scan-%s-%d.%s"
	defaultFileMode    = 0755
	defaultInstance    = "popeye"
	defaultGtwyTimeout = 30 * time.Second
)

var (
	// LogFile the path to our logs.
	LogFile = filepath.Join(os.TempDir(), "popeye.log")

	// DumpDir track scan report directory location.
	DumpDir = dumpDir()
)

type run struct {
	outcome issues.Outcome
	gvr     types.GVR
}

// Popeye represents a kubernetes linter/linter.
type Popeye struct {
	factory      types.Factory
	db           *db.DB
	config       *config.Config
	outputTarget io.ReadWriteCloser
	log          *zerolog.Logger
	flags        *config.Flags
	builder      *report.Builder
	aliases      *internal.Aliases
	codes        *issues.Codes
}

// NewPopeye returns a new instance.
func NewPopeye(flags *config.Flags, log *zerolog.Logger) (*Popeye, error) {
	cfg, err := config.NewConfig(flags)
	if err != nil {
		return nil, err
	}

	return &Popeye{
		config:  cfg,
		log:     log,
		flags:   flags,
		builder: report.NewBuilder(),
		aliases: internal.NewAliases(),
	}, nil
}

func (p *Popeye) initDB() (*db.DB, error) {
	d, err := memdb.NewMemDB(schema.Init())
	if err != nil {
		return nil, err
	}

	return db.NewDB(d), nil
}

// Init configures popeye prior to sanitization.
func (p *Popeye) Init() error {
	if p.factory == nil {
		if err := p.initFactory(); err != nil {
			return err
		}
	}

	if err := p.aliases.Init(p.client()); err != nil {
		return err
	}
	p.aliases.Realize()

	var err error
	p.db, err = p.initDB()
	if err != nil {
		return err
	}
	if !config.IsBoolSet(p.flags.Save) {
		return p.ensureOutput()
	}
	if err := ensureDir(DumpDir, defaultFileMode); err != nil {
		return err
	}

	return p.ensureOutput()
}

// SetFactory sets the resource factory.
func (p *Popeye) SetFactory(f types.Factory) {
	p.factory = f
}

func (p *Popeye) initFactory() error {
	clt, err := client.InitConnectionOrDie(client.NewConfig(p.flags.ConfigFlags))
	if err != nil {
		return err
	}
	f := client.NewFactory(clt)
	p.factory = f

	if p.flags.StandAlone {
		return nil
	}

	ns := client.AllNamespaces
	if p.flags.ConfigFlags.Namespace != nil {
		ns = *p.flags.ConfigFlags.Namespace
	}

	f.Start(ns)
	for k, gvr := range internal.Glossary {
		if gvr == types.BlankGVR {
			log.Debug().Msgf("Skipping linter %q", k)
			continue
		}
		ok, err := clt.CanI(client.AllNamespaces, gvr, "", types.ReadAllAccess)
		if !ok || err != nil {
			return fmt.Errorf("current user does not have read access for resource %q -- %w", gvr, err)
		}
		if _, err := f.ForResource(client.AllNamespaces, gvr); err != nil {
			return err
		}
	}
	f.WaitForCacheSync()

	return nil
}

func (p *Popeye) clusterPath() string {
	return filepath.Join(
		config.SanitizeFileName(p.client().ActiveCluster()),
		config.SanitizeFileName(p.client().ActiveContext()),
	)
}

// Lint scans a cluster for potential issues.
func (p *Popeye) Lint() (int, int, error) {
	defer func() {
		switch {
		case config.IsBoolSet(p.flags.Save):
			if p.outputTarget != nil {
				p.outputTarget.Close()
			}
		case config.IsStrSet(p.flags.S3.Bucket):
			asset := filepath.Join(p.clusterPath(), p.scanFileName())
			if err := p.flags.S3.Upload(context.Background(), asset, p.fileContentType(), p.outputTarget); err != nil {
				log.Fatal().Msgf("S3 upload failed: %s", err)
			}
		}
	}()

	errCount, score, err := p.lint()
	if err != nil {
		return 0, 0, err
	}
	log.Debug().Msgf("Score [%d]", score)

	return errCount, score, p.dump(true, p.flags.Exhaust())
}

func (p *Popeye) buildCtx(ctx context.Context) context.Context {
	ctx = context.WithValue(ctx, internal.KeyOverAllocs, *p.flags.CheckOverAllocs)
	ctx = context.WithValue(ctx, internal.KeyFactory, p.factory)
	ctx = context.WithValue(ctx, internal.KeyConfig, p.config)
	if version, err := p.client().ServerVersion(); err == nil {
		ctx = context.WithValue(ctx, internal.KeyVersion, version)
	}
	ns, err := p.client().Config().CurrentNamespaceName()
	if err != nil {
		log.Warn().Msgf("Unable to determine current namespace: %v. Using `default` namespace", err)
		ns = client.DefaultNamespace
	}
	ctx = context.WithValue(ctx, internal.KeyNamespaceName, ns)

	return context.WithValue(ctx, internal.KeyNamespace, ns)
}

func (p *Popeye) validateSpinach(ss scrub.Scrubs) error {
	if p.flags.Spinach == nil || *p.flags.Spinach == "" {
		return nil
	}
	for k := range p.config.Exclusions.Linters {
		if _, ok := ss[internal.R(k)]; !ok {
			return fmt.Errorf("invalid linter name specified: %q", k)
		}
	}
	return nil
}

func (p *Popeye) lint() (int, int, error) {
	defer func(t time.Time) {
		log.Debug().Msgf("Lint %v", time.Since(t))
	}(time.Now())

	codes, err := issues.LoadCodes()
	if err != nil {
		return 0, 0, err
	}
	codes.Refine(p.config.Overrides)
	p.codes = codes

	var (
		cache    = scrub.NewCache(p.db, p.factory, p.config)
		runners  = make(map[types.GVR]scrub.Linter)
		scrubers = scrub.Scrubers()
	)

	if p.aliases.IsCiliumCluster() {
		cscrub.Inject(scrubers)
		p.aliases.Inject(cilium.Aliases)
	}
	if err := p.validateSpinach(scrubers); err != nil {
		return 0, 0, err
	}

	ctx := p.buildCtx(context.Background())
	sections, ans := p.config.Sections(), p.client().ActiveNamespace()
	nsGVR := types.NewGVR("v1/namespaces")
	for k, fn := range scrubers {
		gvr, ok := internal.Glossary[k]
		if !ok || gvr == types.BlankGVR || p.aliases.Exclude(gvr, sections) {
			continue
		}
		if gvr.String() != nsGVR.String() && client.IsNamespaced(ans) && !p.aliases.IsNamespaced(gvr) {
			continue
		}
		runners[gvr] = fn(ctx, cache, codes)
	}

	total, errCount := len(runners), 0
	if total == 0 {
		return 0, 0, fmt.Errorf("no linters matched query. check section selector")
	}
	c := make(chan run, 2)
	for gvr, r := range runners {
		ctx = context.WithValue(ctx, internal.KeyRunInfo, internal.NewRunInfo(gvr))
		go p.runLinter(ctx, gvr, r, c, cache, codes)
	}

	var score, count int
	for run := range c {
		count++
		tally := report.NewTally()
		tally.Rollup(run.outcome)
		score, errCount = score+tally.Score(), errCount+tally.ErrCount()
		p.builder.AddSection(run.gvr, p.aliases.Singular(run.gvr), run.outcome, tally)
		total--
		if total == 0 {
			close(c)
		}
	}
	if count == 0 {
		return errCount, 0, nil
	}

	return errCount, score / count, nil
}

func (p *Popeye) runLinter(ctx context.Context, gvr types.GVR, l scrub.Linter, c chan run, cache *scrub.Cache, codes *issues.Codes) {
	defer func() {
		if e := recover(); e != nil {
			BailOut(fmt.Errorf("%s", e))
		}
	}()

	if !p.aliases.IsNamespaced(gvr) {
		ctx = context.WithValue(ctx, internal.KeyNamespace, client.ClusterScope)
	}
	if err := l.Lint(ctx); err != nil {
		p.builder.AddError(err)
	}
	o := l.Outcome().Filter(rules.Level(p.config.LintLevel))
	c <- run{gvr: gvr, outcome: o}
}

func (p *Popeye) dumpJunit() error {
	res, err := p.builder.ToJunit(rules.Level(p.config.LintLevel))
	if err != nil {
		return err
	}
	if _, err := p.outputTarget.Write([]byte(xml.Header)); err != nil {
		return err
	}
	fmt.Fprintf(p.outputTarget, "%v\n", res)

	return nil
}

func (p *Popeye) dumpYAML() error {
	res, err := p.builder.ToYAML()
	if err != nil {
		return err
	}
	fmt.Fprintf(p.outputTarget, "%v\n", res)

	return nil
}

func (p *Popeye) dumpJSON() error {
	res, err := p.builder.ToJSON()
	if err != nil {
		return err
	}
	fmt.Fprintf(p.outputTarget, "%v\n", res)

	return nil
}

func (p *Popeye) dumpHTML() error {
	res, err := p.builder.ToHTML()
	if err != nil {
		return err
	}
	fmt.Fprintf(p.outputTarget, "%v\n", res)

	return nil

}
func (p *Popeye) dumpScore() error {
	res, err := p.builder.ToScore()
	if err != nil {
		return err
	}
	fmt.Fprintf(p.outputTarget, "%v\n", res)

	return nil
}

func (p *Popeye) dumpStd(header bool) error {
	var (
		w = bufio.NewWriter(p.outputTarget)
		s = report.New(w, p.flags.OutputFormat() == report.JurassicFormat)
	)

	if header {
		p.builder.PrintHeader(s)
	}
	p.builder.PrintClusterInfo(s, p.client().HasMetrics())
	p.builder.PrintReport(rules.Level(p.config.LintLevel), s)
	p.builder.PrintSummary(s)

	return w.Flush()
}

// Do implements the HTTPDoer interface to replace the standard http client push request and write to the outputTarget
func (p *Popeye) Do(req *http.Request) (*http.Response, error) {
	resp := http.Response{
		// Avoid panic when the pusher tries to close the body
		Body: io.NopCloser(bytes.NewBufferString("Dummy response from file writer")),
	}
	out, err := io.ReadAll(req.Body)
	if err != nil {
		resp.StatusCode = http.StatusInternalServerError
		return &resp, err
	}
	fmt.Fprintf(p.outputTarget, "%s\n", out)
	resp.StatusCode = http.StatusOK

	return &resp, nil
}

func (p *Popeye) client() types.Connection {
	return p.factory.Client()
}

func (p *Popeye) dumpPrometheus(ctx context.Context, asset string, persist bool) error {
	if !config.IsStrSet(p.flags.PushGateway.URL) {
		return nil
	}

	instance := defaultInstance
	if config.IsStrSet(p.flags.InClusterName) {
		instance += "-" + *p.flags.InClusterName
	}

	pusher := p.builder.ToPrometheus(
		p.flags.PushGateway,
		instance,
		p.client().ActiveNamespace(),
		asset,
		p.codes.Glossary,
	)
	// Enable saving to file
	if persist {
		pusher = pusher.Client(p)
		pusher = pusher.Format(expfmt.NewFormat(expfmt.TypeTextPlain))
	}

	return pusher.AddContext(ctx)
}

func (p *Popeye) fetchClusterName() string {
	switch {
	case config.IsStrSet(p.flags.InClusterName):
		return *p.flags.InClusterName
	case p.client().ActiveCluster() != "":
		return p.client().ActiveCluster()
	default:
		return "n/a"
	}
}

func (p *Popeye) fetchContextName() string {
	if ct := p.client().ActiveContext(); ct != "" {
		return ct
	}

	return "n/a"
}

// Dump dumps out scan report.
func (p *Popeye) dump(printHeader bool, asset string) error {
	if !p.builder.HasContent() {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), defaultGtwyTimeout)
	defer cancel()
	p.builder.SetClusterContext(p.fetchClusterName(), p.fetchContextName())
	var errs error
	switch p.flags.OutputFormat() {
	case report.JunitFormat:
		errs = errors.Join(errs, p.dumpJunit())
	case report.YAMLFormat:
		errs = errors.Join(errs, p.dumpYAML())
	case report.JSONFormat:
		errs = errors.Join(errs, p.dumpJSON())
	case report.HTMLFormat:
		errs = errors.Join(errs, p.dumpHTML())
	case report.ScoreFormat:
		errs = errors.Join(errs, p.dumpScore())
	case report.PromFormat:
		errs = errors.Join(errs, p.dumpPrometheus(ctx, asset, true))
	default:
		errs = errors.Join(errs, p.dumpStd(printHeader))
	}

	if p.flags.OutputFormat() != report.PromFormat && config.IsStrSet(p.flags.PushGateway.URL) {
		if config.IsStrSet(p.flags.S3.Bucket) {
			asset = *p.flags.S3.Bucket + "/" + filepath.Join(p.clusterPath(), p.scanFileName())
		}
		errs = errors.Join(p.dumpPrometheus(ctx, asset, false))
	}

	return errs
}

func (p *Popeye) ensureOutput() error {
	p.outputTarget = os.Stdout
	if !config.IsBoolSet(p.flags.Save) && !config.IsStrSet(p.flags.S3.Bucket) {
		return nil
	}
	if p.flags.Output == nil {
		*p.flags.Output = "standard"
	}

	var (
		f   io.ReadWriteCloser
		err error
	)
	switch {
	case config.IsBoolSet(p.flags.Save):
		dir := filepath.Join(
			DumpDir,
			p.clusterPath(),
		)
		if DumpDir != DefaultDumpDir {
			dir = DumpDir
		}

		if err := ensureDir(dir, defaultFileMode); err != nil {
			return err
		}
		file := filepath.Join(dir, config.SanitizeFileName(p.scanFileName()))
		p.flags.OutputFile = &file
		f, err = os.Create(file)
		if err != nil {
			return err
		}
		fmt.Println(file)
	case config.IsStrSet(p.flags.S3.Bucket):
		f = NopCloser(bytes.NewBufferString(""))
	}
	p.outputTarget = f

	return nil
}

func (p *Popeye) scanFileName() string {
	if config.IsStrSet(p.flags.OutputFile) {
		return *p.flags.OutputFile
	}

	ns := p.client().ActiveNamespace()
	if ns == client.BlankNamespace {
		ns = client.NamespaceAll
	}
	return fmt.Sprintf(dumpFileFMT, ns, time.Now().UnixNano(), p.fileExt())
}

func (p *Popeye) fileExt() string {
	switch *p.flags.Output {
	case "junit":
		return "xml"
	case "json", "yaml", "html":
		return *p.flags.Output
	default:
		return "txt"
	}
}

func (p *Popeye) fileContentType() string {
	switch *p.flags.Output {
	case "junit":
		// https://datatracker.ietf.org/doc/html/rfc7303#section-4.1
		return "application/xml"
	case "json":
		return "application/json"
	case "yaml":
		// https://datatracker.ietf.org/doc/html/rfc9512
		return "application/yaml"
	case "html":
		return "text/html"
	default:
		return "text/plain"
	}
}
