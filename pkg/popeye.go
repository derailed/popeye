package pkg

import (
	"bufio"
	"bytes"
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
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
	"github.com/prometheus/common/expfmt"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const outFmt = "sanitizer_%s_%d.%s"

var (
	// LogFile the path to our logs.
	LogFile = filepath.Join(os.TempDir(), "popeye.log")
	// DumpDir indicates a directory location for sanitizer reports.
	DumpDir = dumpDir()
	// ErrUnknownS3BucketProtocol defines the error if we can't parse the S3 URI
	ErrUnknownS3BucketProtocol = errors.New("invalid S3 URI: hostname not valid")

	gvrs internal.GVRs
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

	return &Popeye{
		config:  cfg,
		log:     log,
		flags:   flags,
		builder: report.NewBuilder(),
	}, nil
}

// Init configures popeye prior to sanitization.
func (p *Popeye) Init() error {
	if p.factory == nil {
		if err := p.initFactory(); err != nil {
			return err
		}
	}
	rev, err := p.revision()
	if err != nil {
		return err
	}
	gvrs = p.scannedGVRs(rev)

	p.aliases = internal.NewAliases()
	if err := p.aliases.Init(p.factory, gvrs); err != nil {
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

func (p *Popeye) scannedGVRs(rev *client.Revision) internal.GVRs {
	mm := internal.GVRs{
		internal.LrGVR:  "v1/limitranges",
		internal.SvcGVR: "v1/services",
		internal.EpGVR:  "v1/endpoints",
		internal.NoGVR:  "v1/nodes",
		internal.NsGVR:  "v1/namespaces",
		internal.PoGVR:  "v1/pods",
		internal.CmGVR:  "v1/configmaps",
		internal.SecGVR: "v1/secrets",
		internal.SaGVR:  "v1/serviceaccounts",
		internal.PvGVR:  "v1/persistentvolumes",
		internal.PvcGVR: "v1/persistentvolumeclaims",
		internal.DpGVR:  "apps/v1/deployments",
		internal.RsGVR:  "apps/v1/replicasets",
		internal.DsGVR:  "apps/v1/daemonsets",
		internal.StsGVR: "apps/v1/statefulsets",
		internal.NpGVR:  "networking.k8s.io/v1/networkpolicies",
		internal.CrGVR:  "rbac.authorization.k8s.io/v1/clusterroles",
		internal.CrbGVR: "rbac.authorization.k8s.io/v1/clusterrolebindings",
		internal.RoGVR:  "rbac.authorization.k8s.io/v1/roles",
		internal.RobGVR: "rbac.authorization.k8s.io/v1/rolebindings",
		internal.IngGVR: "networking.k8s.io/v1/ingresses",
		internal.PspGVR: "policy/v1/podsecuritypolicies",
		internal.PdbGVR: "policy/v1/poddisruptionbudgets",
		internal.HpaGVR: "autoscaling/v2/horizontalpodautoscalers",
	}

	if rev.Minor < 18 {
		mm[internal.IngGVR] = "networking.k8s.io/v1beta1/ingresses"
	}
	if rev.Minor <= 24 {
		mm[internal.PspGVR] = "policy/v1beta1/podsecuritypolicies"
	}
	if rev.Minor < 21 {
		mm[internal.PdbGVR] = "policy/v1beta1/poddisruptionbudgets"
	}
	if rev.Minor < 23 {
		mm[internal.HpaGVR] = "autoscaling/v1/horizontalpodautoscalers"
	}

	return mm
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

	info, err := p.factory.Client().ServerVersion()
	if err != nil {
		return err
	}
	rev, err := client.NewRevision(info)
	if err != nil {
		return err
	}

	ns := client.AllNamespaces
	if p.flags.ConfigFlags.Namespace != nil {
		ns = *p.flags.ConfigFlags.Namespace
	}

	f.Start(ns)
	for _, gvr := range p.scannedGVRs(rev) {
		ok, err := clt.CanI(client.AllNamespaces, gvr, types.ReadAllAccess)
		if !ok || err != nil {
			return fmt.Errorf("Current user does not have read access for resource %q -- %w", gvr, err)
		}
		if _, err := f.ForResource(client.AllNamespaces, gvr); err != nil {
			return err
		}
	}
	f.WaitForCacheSync()

	return nil
}

func (p *Popeye) revision() (*client.Revision, error) {
	info, err := p.factory.Client().ServerVersion()
	if err != nil {
		return nil, err
	}

	return client.NewRevision(info)
}

func (p *Popeye) sanitizers(rev *client.Revision) map[string]scrubFn {
	mm := map[string]scrubFn{
		"cluster":             scrub.NewCluster,
		gvrs[internal.CmGVR]:  scrub.NewConfigMap,
		gvrs[internal.NsGVR]:  scrub.NewNamespace,
		gvrs[internal.NoGVR]:  scrub.NewNode,
		gvrs[internal.PoGVR]:  scrub.NewPod,
		gvrs[internal.PvGVR]:  scrub.NewPersistentVolume,
		gvrs[internal.PvcGVR]: scrub.NewPersistentVolumeClaim,
		gvrs[internal.SecGVR]: scrub.NewSecret,
		gvrs[internal.SvcGVR]: scrub.NewService,
		gvrs[internal.SaGVR]:  scrub.NewServiceAccount,
		gvrs[internal.DsGVR]:  scrub.NewDaemonSet,
		gvrs[internal.DpGVR]:  scrub.NewDeployment,
		gvrs[internal.RsGVR]:  scrub.NewReplicaSet,
		gvrs[internal.StsGVR]: scrub.NewStatefulSet,
		gvrs[internal.NpGVR]:  scrub.NewNetworkPolicy,
		gvrs[internal.IngGVR]: scrub.NewIngress,
		gvrs[internal.CrGVR]:  scrub.NewClusterRole,
		gvrs[internal.CrbGVR]: scrub.NewClusterRoleBinding,
		gvrs[internal.RoGVR]:  scrub.NewRole,
		gvrs[internal.RobGVR]: scrub.NewRoleBinding,
		gvrs[internal.PspGVR]: scrub.NewPodSecurityPolicy,
		gvrs[internal.PdbGVR]: scrub.NewPodDisruptionBudget,
		gvrs[internal.HpaGVR]: scrub.NewHorizontalPodAutoscaler,
	}

	return mm
}

// SetOutputTarget sets up a new output stream writer.
func (p *Popeye) SetOutputTarget(s io.ReadWriteCloser) {
	p.outputTarget = s
}

// Sanitize scans a cluster for potential issues.
func (p *Popeye) Sanitize() (int, int, error) {
	defer func() {
		switch {
		case isSet(p.flags.Save):
			if err := p.outputTarget.Close(); err != nil {
				log.Fatal().Err(err).Msg("Closing report")
			}
		case isSetStr(p.flags.S3Bucket):
			bucket, key, err := parseBucket(*p.flags.S3Bucket)
			if err != nil {
				log.Fatal().Err(err).Msg("Parse S3 bucket URI")
			}

			// Create a single AWS session (we can re use this if we're uploading many files)
			s, err := session.NewSession(&aws.Config{
				LogLevel: aws.LogLevel(aws.LogDebugWithRequestErrors),
				Region:   p.flags.S3Region,
				Endpoint: p.flags.S3Endpoint,
			})
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

	if err := client.Load(p.factory); err != nil {
		return 0, 0, err
	}

	errCount, score, err := p.sanitize()
	if err != nil {
		return 0, 0, err
	}

	return errCount, score, p.dump(true)
}

func (p *Popeye) sanitize() (int, int, error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ctx = context.WithValue(ctx, internal.KeyOverAllocs, *p.flags.CheckOverAllocs)
	ctx = context.WithValue(ctx, internal.KeyFactory, p.factory)
	if version, err := p.factory.Client().ServerVersion(); err == nil {
		ctx = context.WithValue(ctx, internal.KeyVersion, version)
	}

	codes, err := issues.LoadCodes()
	if err != nil {
		return 0, 0, err
	}
	codes.Refine(p.config.Codes)

	c := make(chan run, 2)
	var total, errCount int
	var nodeGVR = client.NewGVR("v1/nodes")
	cache := scrub.NewCache(p.factory, p.config)

	rev, err := p.revision()
	if err != nil {
		return 0, 0, err
	}
	for k, fn := range p.sanitizers(rev) {
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
		return 0, 0, nil
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

func (p *Popeye) sanitizer(ctx context.Context, gvr client.GVR, f scrubFn, c chan run, cache *scrub.Cache, codes *issues.Codes) {
	defer func() {
		if e := recover(); e != nil {
			log.Error().Msgf("Popeye CHOKED! %#v", e)
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
	p.builder.PrintClusterInfo(s, p.factory.Client().ActiveCluster(), p.factory.Client().HasMetrics())
	p.builder.PrintReport(config.Level(p.config.LinterLevel()), s)
	p.builder.PrintSummary(s)

	return w.Flush()
}

// Do implements the HTTPDoer interface to replace the standard http client push request and write to the outputTarget
func (p *Popeye) Do(req *http.Request) (*http.Response, error) {
	resp := http.Response{
		// Avoid panic when the pusher tries to close the body
		Body: ioutil.NopCloser(bytes.NewBufferString("Dummy response from file writer")),
	}

	out, err := ioutil.ReadAll(req.Body)
	if err != nil {
		resp.StatusCode = http.StatusInternalServerError
		return &resp, err
	}

	fmt.Fprintf(p.outputTarget, "%s\n", out)

	resp.StatusCode = http.StatusOK
	return &resp, nil
}

func (p *Popeye) dumpPrometheus() error {
	pusher := p.builder.ToPrometheus(
		p.flags.PushGateway,
		p.factory.Client().ActiveNamespace(),
	)

	// Enable saving to file
	if isSet(p.flags.Save) || isSetStr(p.flags.S3Bucket) {
		pusher.Client(p)
		pusher.Format(expfmt.FmtText)
	}

	return pusher.Add()
}

func (p *Popeye) fetchClusterName() string {
	switch {
	case p.factory.Client().ActiveCluster() != "":
		return p.factory.Client().ActiveCluster()
	case p.flags.InClusterName != nil && *p.flags.InClusterName != "":
		return *p.flags.InClusterName
	default:
		return "n/a"
	}
}

// Dump prints out sanitizer report.
func (p *Popeye) dump(printHeader bool) error {
	if !p.builder.HasContent() {
		return errors.New("Nothing to report, check section name or permissions")
	}

	p.builder.SetClusterName(p.fetchClusterName())
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
	if *p.flags.OutputFile == "" {
		return fmt.Sprintf(outFmt, p.factory.Client().ActiveCluster(), time.Now().UnixNano(), p.fileExt())
	}
	return fmt.Sprintf(*p.flags.OutputFile)
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
		return fmt.Errorf("Fail to create popeye sanitizers dump dir: %w", err)
	}
	return nil
}

func dumpDir() string {
	if d := os.Getenv("POPEYE_REPORT_DIR"); d != "" {
		return d
	}
	return filepath.Join(os.TempDir(), "popeye")
}

func parseBucket(bucketURI string) (string, string, error) {
	u, err := url.Parse(bucketURI)
	if err != nil {
		return "", "", err
	}
	switch u.Scheme {
	// s3://bucket or s3://bucket/
	case "s3":
		var key string
		if u.Path != "" {
			key = strings.Trim(u.Path, "/")
		}
		return u.Host, key, nil
	// bucket/ or bucket/path/to/key
	case "":
		tokens := strings.SplitAfterN(strings.Trim(u.Path, "/"), "/", 2)
		key, bucket := "", strings.Trim(tokens[0], "/")
		if len(tokens) > 1 {
			key = tokens[1]
		}
		return bucket, key, nil
	default:
		return "", "", ErrUnknownS3BucketProtocol
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
