package pkg

import (
	"bufio"
	"bytes"
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"runtime/debug"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/derailed/popeye/internal"
	"github.com/derailed/popeye/internal/client"
	"github.com/derailed/popeye/internal/issues"
	"github.com/derailed/popeye/internal/report"
	"github.com/derailed/popeye/internal/scrub"
	"github.com/derailed/popeye/pkg/config"
	"github.com/derailed/popeye/types"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const outFmt = "sanitizer_%s_%d.%s"

var (
	// LogFile the path to our logs.
	LogFile = filepath.Join(os.TempDir(), "popeye.log")
	// DumpDir indicates a directory location for sanitizer reports.
	DumpDir = dumpDir()
)

type scrubFn func(context.Context, *scrub.Cache, *issues.Codes) scrub.Sanitizer

type run struct {
	outcome issues.Outcome
	gvr     client.GVR
}

// Popeye represents a kubernetes linter/sanitizer.
type Popeye struct {
	factory      types.Factory
	config       *config.Config
	outputTarget io.ReadWriteCloser
	log          *zerolog.Logger
	flags        *config.Flags
	builder      *report.Builder
	aliases      *internal.Aliases
}

// NewPopeye returns a new instance.
func NewPopeye(flags *config.Flags, log *zerolog.Logger) (*Popeye, error) {
	cfg, err := config.NewConfig(flags)
	if err != nil {
		return nil, err
	}

	p := Popeye{
		config:  cfg,
		log:     log,
		flags:   flags,
		builder: report.NewBuilder(),
	}
	return &p, nil
}

// Init configures popeye prior to sanitization.
func (p *Popeye) Init() error {
	if p.factory == nil {
		if err := p.initFactory(); err != nil {
			return err
		}
	}
	p.aliases = internal.NewAliases()
	if err := p.aliases.Init(p.factory, p.scannedGVRs()); err != nil {
		return err
	}

	if !isSet(p.flags.Save) {
		return p.ensureOutput()
	}
	if err := ensurePath(DumpDir, 0755); err != nil {
		return err
	}

	return p.ensureOutput()
}

// SetFactory sets the resource factory.
func (p *Popeye) SetFactory(f types.Factory) {
	p.factory = f
}

func (p *Popeye) scannedGVRs() []string {
	return []string{
		"v1/limitranges",
		"v1/services",
		"v1/endpoints",
		"v1/nodes",
		"v1/namespaces",
		"v1/pods",
		"v1/configmaps",
		"v1/secrets",
		"v1/serviceaccounts",
		"v1/persistentvolumes",
		"v1/persistentvolumeclaims",
		"apps/v1/deployments",
		"apps/v1/replicasets",
		"apps/v1/daemonsets",
		"apps/v1/statefulsets",
		"policy/v1beta1/poddisruptionbudgets",
		"policy/v1beta1/podsecuritypolicies",
		"extensions/v1beta1/ingresses",
		"networking.k8s.io/v1/networkpolicies",
		"autoscaling/v1/horizontalpodautoscalers",
		"rbac.authorization.k8s.io/v1/clusterroles",
		"rbac.authorization.k8s.io/v1/clusterrolebindings",
		"rbac.authorization.k8s.io/v1/roles",
		"rbac.authorization.k8s.io/v1/rolebindings",
	}
}
func (p *Popeye) initFactory() error {
	clt := client.InitConnectionOrDie(client.NewConfig(p.flags.ConfigFlags))
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
	for _, gvr := range p.scannedGVRs() {
		ok, err := clt.CanI(client.AllNamespaces, gvr, types.ReadAllAccess)
		if !ok || err != nil {
			return fmt.Errorf("Current user does not have read access for resource %q -- %v", gvr, err)
		}
		if _, err := f.ForResource(client.AllNamespaces, gvr); err != nil {
			return err
		}
	}
	f.WaitForCacheSync()

	return nil
}

func (p *Popeye) sanitizers() map[string]scrubFn {
	return map[string]scrubFn{
		"cluster":                   scrub.NewCluster,
		"v1/configmaps":             scrub.NewConfigMap,
		"v1/namespaces":             scrub.NewNamespace,
		"v1/nodes":                  scrub.NewNode,
		"v1/pods":                   scrub.NewPod,
		"v1/persistentvolumes":      scrub.NewPersistentVolume,
		"v1/persistentvolumeclaims": scrub.NewPersistentVolumeClaim,
		"v1/secrets":                scrub.NewSecret,
		"v1/services":               scrub.NewService,
		"v1/serviceaccounts":        scrub.NewServiceAccount,
		"apps/v1/daemonsets":        scrub.NewDaemonSet,
		"apps/v1/deployments":       scrub.NewDeployment,
		"apps/v1/replicasets":       scrub.NewReplicaSet,
		"apps/v1/statefulsets":      scrub.NewStatefulSet,
		"autoscaling/v1/horizontalpodautoscalers":          scrub.NewHorizontalPodAutoscaler,
		"extensions/v1beta1/ingresses":                     scrub.NewIngress,
		"networking.k8s.io/v1/networkpolicies":             scrub.NewNetworkPolicy,
		"policy/v1beta1/poddisruptionbudgets":              scrub.NewPodDisruptionBudget,
		"policy/v1beta1/podsecuritypolicies":               scrub.NewPodSecurityPolicy,
		"rbac.authorization.k8s.io/v1/clusterroles":        scrub.NewClusterRole,
		"rbac.authorization.k8s.io/v1/clusterrolebindings": scrub.NewClusterRoleBinding,
		"rbac.authorization.k8s.io/v1/roles":               scrub.NewRole,
		"rbac.authorization.k8s.io/v1/rolebindings":        scrub.NewRoleBinding,
	}
}

// SetOutputTarget sets up a new output stream writer.
func (p *Popeye) SetOutputTarget(s io.ReadWriteCloser) {
	p.outputTarget = s
}

// Sanitize scans a cluster for potential issues.
func (p *Popeye) Sanitize() error {
	defer func() {
		switch {
		case isSet(p.flags.Save):
			if err := p.outputTarget.Close(); err != nil {
				log.Fatal().Err(err).Msg("Closing report")
			}
		case isSetStr(p.flags.S3Bucket):
			var bucket, key string
			if len(strings.Split(*p.flags.S3Bucket, "/")) > 1 {
				bucket, key = parseBucket(*p.flags.S3Bucket)
			} else {
				bucket = *p.flags.S3Bucket
			}
			// Create a single AWS session (we can re use this if we're uploading many files)
			s, err := session.NewSession(&aws.Config{
				LogLevel: aws.LogLevel(aws.LogDebugWithRequestErrors)})
			if err != nil {
				log.Fatal().Err(err).Msg("Create S3 Session")
			}
			// Create an uploader with the session and default options
			uploader := s3manager.NewUploader(s)
			// Upload input parameters
			upParams := &s3manager.UploadInput{
				Bucket: aws.String(bucket),
				Key:    aws.String(key + "/" + p.fileName()),
				Body:   p.outputTarget,
			}

			// Perform an upload.
			if _, err = uploader.Upload(upParams); err != nil {
				log.Fatal().Err(err).Msg("S3 Upload")
			}
		}
	}()

	if err := p.sanitize(); err != nil {
		return err
	}

	return p.dump(true)
}

func (p *Popeye) sanitize() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ctx = context.WithValue(ctx, internal.KeyOverAllocs, *p.flags.CheckOverAllocs)
	ctx = context.WithValue(ctx, internal.KeyFactory, p.factory)

	cache := scrub.NewCache(p.factory, p.config)
	codes, err := issues.LoadCodes()
	if err != nil {
		return err
	}
	codes.Refine(p.config.Codes)

	c := make(chan run, 2)
	var total int
	var nodeGVR = client.NewGVR("v1/nodes")
	for k, fn := range p.sanitizers() {
		gvr := client.NewGVR(k)
		if p.aliases.Exclude(gvr, p.config.Sections()) {
			continue
		}

		// Skip node sanitizer if active namespace is set.
		if gvr == nodeGVR && p.factory.Client().ActiveNamespace() != client.AllNamespaces {
			continue
		}
		total++
		ctx = context.WithValue(ctx, internal.KeyRunInfo, internal.RunInfo{Section: gvr.R(), SectionGVR: gvr})
		go p.sanitizer(ctx, gvr, fn, c, cache, codes)
	}

	if total == 0 {
		return nil
	}

	for run := range c {
		tally := report.NewTally()
		tally.Rollup(run.outcome)
		p.builder.AddSection(run.gvr, p.aliases.Singular(run.gvr), run.outcome, tally)
		total--
		if total == 0 {
			close(c)
		}
	}

	return nil
}

func (p *Popeye) sanitizer(ctx context.Context, gvr client.GVR, f scrubFn, c chan run, cache *scrub.Cache, codes *issues.Codes) {
	defer func() {
		if e := recover(); e != nil {
			log.Error().Msgf("Popeye CHOCKED! %#v", e)
			log.Error().Msgf("%v", string(debug.Stack()))
		}
	}()

	resource := f(ctx, cache, codes)
	if err := resource.Sanitize(ctx); err != nil {
		p.builder.AddError(err)
	}
	o := resource.Outcome().Filter(config.Level(p.config.LinterLevel()))
	c <- run{gvr: gvr, outcome: o}
}

func (p *Popeye) dumpJunit() error {
	res, err := p.builder.ToJunit(config.Level(p.config.LinterLevel()))
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

func (p *Popeye) dumpStd(mode, header bool) error {
	var (
		w = bufio.NewWriter(p.outputTarget)
		s = report.NewSanitizer(w, mode)
	)

	if header {
		p.builder.PrintHeader(s)
	}
	mx := p.factory.Client().HasMetrics()
	p.builder.PrintClusterInfo(s, p.factory.Client().ActiveCluster(), mx)
	p.builder.PrintReport(config.Level(p.config.LinterLevel()), s)
	p.builder.PrintSummary(s)

	return w.Flush()
}

func (p *Popeye) dumpPrometheus() error {
	pusher := p.builder.ToPrometheus(
		p.flags.PushGatewayAddress,
		p.factory.Client().ActiveNamespace(),
	)
	return pusher.Add()
}

// Dump prints out sanitizer report.
func (p *Popeye) dump(printHeader bool) error {
	if !p.builder.HasContent() {
		return errors.New("Nothing to report, check section name or permissions")
	}

	p.builder.SetClusterName(p.factory.Client().ActiveCluster())
	var err error
	switch p.flags.OutputFormat() {
	case report.JunitFormat:
		err = p.dumpJunit()
	case report.YAMLFormat:
		err = p.dumpYAML()
	case report.JSONFormat:
		err = p.dumpJSON()
	case report.HTMLFormat:
		err = p.dumpHTML()
	case report.PrometheusFormat:
		err = p.dumpPrometheus()
	case report.ScoreFormat:
		err = p.dumpScore()
	default:
		err = p.dumpStd(p.flags.OutputFormat() == report.JurassicFormat, printHeader)
	}

	return err
}

func (p *Popeye) ensureOutput() error {
	p.outputTarget = os.Stdout
	if !isSet(p.flags.Save) && !isSetStr(p.flags.S3Bucket) {
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
	case isSet(p.flags.Save):
		fPath := filepath.Join(DumpDir, p.fileName())
		f, err = os.Create(fPath)
		if err != nil {
			return err
		}
		fmt.Println(fPath)
	case isSetStr(p.flags.S3Bucket):
		f = NopWriter(bytes.NewBufferString(""))
	}
	p.outputTarget = f

	return nil
}

func (p *Popeye) fileName() string {
	return fmt.Sprintf(outFmt, p.factory.Client().ActiveCluster(), time.Now().UnixNano(), p.fileExt())
}

func (p *Popeye) fileExt() string {
	switch *p.flags.Output {
	case "json":
		return "json"
	case "junit":
		return "xml"
	case "yaml":
		return "yml"
	case "html":
		return "html"
	default:
		return "txt"
	}
}

// ----------------------------------------------------------------------------
// Helpers...

func isSet(b *bool) bool {
	return b != nil && *b
}

func isSetStr(s *string) bool {
	return s != nil && *s != ""
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

func dumpDir() string {
	if d := os.Getenv("POPEYE_REPORT_DIR"); d != "" {
		return d
	}
	return filepath.Join(os.TempDir(), "popeye")
}

func parseBucket(bucket string) (string, string) {
	if u, _ := url.Parse(bucket); u.Scheme != "" {
		return u.Host, path.Clean(u.Path)
	} else {
		uri := strings.SplitAfterN(bucket, "/", 2)
		return path.Clean(uri[0]), path.Clean(uri[1])
	}
}

type readWriteCloser struct {
	io.ReadWriter
}

// Close close read stream.
func (wC readWriteCloser) Close() error {
	return nil
}

// NopWriter fake writer.
func NopWriter(i io.ReadWriter) io.ReadWriteCloser {
	return &readWriteCloser{i}
}
