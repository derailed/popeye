package pkg

import (
	"bufio"
	"bytes"
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/derailed/popeye/internal"
	"github.com/derailed/popeye/internal/client"
	"github.com/derailed/popeye/internal/issues"
	"github.com/derailed/popeye/internal/report"
	"github.com/derailed/popeye/internal/sanitize"
	"github.com/derailed/popeye/internal/scrub"
	"github.com/derailed/popeye/pkg/config"
	"github.com/derailed/popeye/types"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var (
	// LogFile the path to our logs.
	LogFile = filepath.Join(os.TempDir(), fmt.Sprintf("popeye.log"))
	// DumpDir indicates a directory location for sanitizer reports.
	DumpDir = dumpDir()
)

const outFmt = "sanitizer_%s_%d.%s"

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

func dumpDir() string {
	if d := os.Getenv("POPEYE_REPORT_DIR"); d != "" {
		return d
	}
	return filepath.Join(os.TempDir(), "popeye")
}

type (
	scrubFn func(context.Context, *scrub.Cache, *issues.Codes) scrub.Sanitizer

	// Popeye a kubernetes sanitizer.
	Popeye struct {
		factory      types.Factory
		config       *config.Config
		outputTarget io.ReadWriteCloser
		log          *zerolog.Logger
		flags        *config.Flags
		builder      *report.Builder
		aliases      *internal.Aliases
	}
)

// NewPopeye returns a new sanitizer.
func NewPopeye(flags *config.Flags, log *zerolog.Logger) (*Popeye, error) {
	cfg, err := config.NewConfig(flags)
	if err != nil {
		return nil, err
	}

	a := internal.NewAliases()
	p := Popeye{
		config:  cfg,
		log:     log,
		flags:   flags,
		aliases: a,
		builder: report.NewBuilder(a),
	}

	return &p, nil
}

// SetFactory sets the resource factory.
func (p *Popeye) SetFactory(f types.Factory) {
	p.factory = f
}

func (p *Popeye) initFactory() {
	clt := client.InitConnectionOrDie(client.NewConfig(p.flags.ConfigFlags))
	f := client.NewFactory(clt)

	ns := client.AllNamespaces
	if p.flags.ConfigFlags.Namespace != nil {
		ns = *p.flags.ConfigFlags.Namespace
	}
	f.Start(ns)
	f.ForResource(ns, "policy/v1beta1/poddisruptionbudgets")
	f.ForResource(ns, "policy/v1beta1/podsecuritypolicies")
	f.ForResource(ns, "extensions/v1beta1/ingresses")
	f.ForResource(ns, "networking.k8s.io/v1/networkpolicies")
	f.ForResource(ns, "autoscaling/v1/horizontalpodautoscalers")
	f.ForResource(ns, "apps/v1/deployments")
	f.ForResource(ns, "apps/v1/replicasets")
	f.ForResource(ns, "apps/v1/daemonsets")
	f.ForResource(ns, "apps/v1/statefulsets")
	f.ForResource(ns, "v1/limitranges")
	f.ForResource(ns, "v1/services")
	f.ForResource(ns, "v1/endpoints")
	f.ForResource(ns, "v1/nodes")
	f.ForResource(ns, "v1/namespaces")
	f.ForResource(ns, "v1/pods")
	f.ForResource(ns, "v1/configmaps")
	f.ForResource(ns, "v1/secrets")
	f.ForResource(ns, "v1/serviceaccounts")
	f.ForResource(ns, "v1/persistentvolumes")
	f.ForResource(ns, "v1/persistentvolumeclaims")
	f.ForResource(ns, "rbac.authorization.k8s.io/v1/clusterroles")
	f.ForResource(ns, "rbac.authorization.k8s.io/v1/clusterrolebindings")
	f.ForResource(ns, "rbac.authorization.k8s.io/v1/roles")
	f.ForResource(ns, "rbac.authorization.k8s.io/v1/rolebindings")
	f.WaitForCacheSync()

	p.factory = f
}

// Init configures popeye prior to sanitization.
func (p *Popeye) Init() error {
	if p.factory == nil {
		p.initFactory()
	}
	if !isSet(p.flags.Save) {
		return p.ensureOutput()
	}
	if err := ensurePath(DumpDir, 0755); err != nil {
		return err
	}

	return p.ensureOutput()
}

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
				Bucket: p.flags.S3Bucket,
				Key:    aws.String(p.fileName()),
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

func (p *Popeye) sanitizers() map[string]scrubFn {
	return map[string]scrubFn{
		"cluster":                 scrub.NewCluster,
		"configmap":               scrub.NewConfigMap,
		"secret":                  scrub.NewSecret,
		"deployment":              scrub.NewDeployment,
		"daemonset":               scrub.NewDaemonSet,
		"horizontalpodautoscaler": scrub.NewHorizontalPodAutoscaler,
		"namespace":               scrub.NewNamespace,
		"node":                    scrub.NewNode,
		"persistentvolume":        scrub.NewPersistentVolume,
		"persistentvolumeclaim":   scrub.NewPersistentVolumeClaim,
		"pod":                     scrub.NewPod,
		"replicaset":              scrub.NewReplicaSet,
		"service":                 scrub.NewService,
		"serviceaccount":          scrub.NewServiceAccount,
		"statefulset":             scrub.NewStatefulSet,
		"poddisruptionbudget":     scrub.NewPodDisruptionBudget,
		"ingress":                 scrub.NewIngress,
		"networkpolicy":           scrub.NewNetworkPolicy,
		"podsecuritypolicy":       scrub.NewPodSecurityPolicy,
		"clusterrole":             scrub.NewClusterRole,
		"clusterrolebinding":      scrub.NewClusterRoleBinding,
		"role":                    scrub.NewRole,
		"rolebinding":             scrub.NewRoleBinding,
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

func (p *Popeye) sanitize() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ctx = context.WithValue(
		ctx,
		sanitize.PopeyeKey("OverAllocs"),
		*p.flags.CheckOverAllocs,
	)

	cache := scrub.NewCache(p.factory, p.config)
	codes, err := issues.LoadCodes()
	if err != nil {
		return err
	}
	codes.Refine(p.config.Codes)
	sections := make([]string, 0, len(p.sanitizers()))
	for section := range p.sanitizers() {
		sections = append(sections, section)
	}
	sort.StringSlice(sections).Sort()
	for _, section := range sections {
		if !in(p.aliases.ToResources(p.config.Sections()), section) {
			continue
		}
		// Skip node checks if active namespace is set.
		if section == "node" && p.factory.Client().ActiveNamespace() != client.AllNamespaces {
			continue
		}

		ctx = context.WithValue(ctx, internal.KeyRun, internal.RunInfo{Section: section})
		ctx = context.WithValue(ctx, internal.KeyFactory, p.factory)
		s := p.sanitizers()[section](ctx, cache, codes)
		if err := s.Sanitize(ctx); err != nil {
			p.builder.AddError(err)
			continue
		}

		o := s.Outcome().Filter(config.Level(p.config.LinterLevel()))
		tally := report.NewTally()
		tally.Rollup(o)
		p.builder.AddSection(section, o, tally)
	}

	return nil
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
