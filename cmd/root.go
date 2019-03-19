package cmd

import (
	"fmt"
	"strings"

	"github.com/derailed/popeye/internal/k8s"
	"github.com/derailed/popeye/internal/k8s/generated"
	"github.com/derailed/popeye/internal/k8s/linter"
	"github.com/derailed/popeye/internal/output"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	v1 "k8s.io/api/core/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

type popeyeConfig struct {
	logLevel      string
	lintLevel     string
	allNamespaces bool
	outputWidth   int
}

func newPopeyeConfig() popeyeConfig {
	return popeyeConfig{
		logLevel:      "debug",
		lintLevel:     "",
		allNamespaces: false,
		outputWidth:   80,
	}
}

var (
	version   = "dev"
	commit    = "dev"
	popConfig = newPopeyeConfig()
	rootCmd   *cobra.Command
	k8sConfig *genericclioptions.ConfigFlags
)

// Linter represents a resource linter.
type Linter interface {
	MaxSeverity() linter.Level
	NoIssues() bool
	Issues() linter.Issues
}

func init() {
	rootCmd = &cobra.Command{
		Use:   "Popeye",
		Short: "A Kubernetes resource linter",
		Long:  `A Kubernetes resource linter`,
		Run:   run,
	}

	initK8sFlags()
	initPopeyeFlags()
	zerolog.SetGlobalLevel(parseLogLevel(popConfig.logLevel))
}

// Execute root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Panic().Err(err)
	}
}

// run the linter based on cli args
func run(cmd *cobra.Command, args []string) {
	clearScreen()
	for _, s := range output.Logo {
		fmt.Printf(strings.Repeat(" ", 62))
		fmt.Println(output.Colorize(s, output.ColorCoolBlue))
	}
	fmt.Println("Popeye -- Biffs 'em and Buffs 'em!")
	fmt.Println("")
	lint()
}

func lint() {
	c := k8s.NewClient(k8s.NewConfig(k8sConfig))
	checkCluster(c)

	lintNode("Node", c)

	nss := getNamespacesOrDie(c)
	lintNamespace("Namespace", nss)

	nn := namespaceNames(nss)
	lintPod("Pod", c, nn)
	lintSvc("Service", c, nn)
}

func checkCluster(c *k8s.Client) {
	mx := c.ClusterHasMetrics()
	output.Write(linter.OkLevel, "Kubernetes", "Connectivity")

	if !mx {
		output.Write(linter.OkLevel, "Cluster", "Metrics")
	} else {
		output.Write(linter.OkLevel, "Cluster", "Metrics")
	}
	fmt.Println("")
}

func lintNamespace(section string, nn []v1.Namespace) {
	for _, n := range nn {
		l := linter.NewNamespace()
		l.Lint(n)
		printReport(l, section, n.Name)
	}
	fmt.Println("")
}

func lintNode(section string, c *k8s.Client) {
	var gen generated.Node
	nn, err := gen.List(c)
	if err != nil {
		log.Fatal().Err(err)
	}
	for _, n := range nn.Items {
		l := linter.NewNode()
		var mx k8s.NodeMetric
		if c.ClusterHasMetrics() {
			if mx, err = k8s.NodeMetrics(c, n); err != nil {
				log.Debug().Err(err)
			}
		}
		l.Lint(n, mx)
		printReport(l, section, n.Name)
	}
	fmt.Println("")
}

func lintPod(section string, c *k8s.Client, nn []string) {
	var gen generated.Pod
	for _, ns := range nn {
		pp, err := gen.List(c, ns)
		if err != nil {
			log.Error().Err(err)
		}

		for _, p := range pp.Items {
			mx := make(map[string]linter.PodMetric)
			if c.ClusterHasMetrics() {
				if mm, err := k8s.PodMetrics(c, ns, p.Name); err != nil {
					log.Debug().Err(err)
				} else {
					for k, v := range mm {
						mx[k] = v
					}
				}
			}
			l := linter.NewPod()
			l.Lint(p, mx)
			printReport(l, section, namespaced(ns, p.Name))
		}
	}
	fmt.Println("")
}

func lintSvc(section string, c *k8s.Client, nn []string) {
	var gen generated.Service
	for _, ns := range nn {
		ss, err := gen.List(c, ns)
		if err != nil {
			log.Error().Err(err)
		}

		for _, s := range ss.Items {
			l := linter.NewService()
			l.Lint(s)
			printReport(l, section, namespaced(ns, s.Name))
		}
	}
	fmt.Println("")
}

func printReport(l Linter, section, name string) {
	level := parseLintLevel(popConfig.lintLevel)
	if l.NoIssues() {
		if level <= linter.OkLevel {
			output.Write(linter.OkLevel, section, name)
		}
		return
	}

	max := l.MaxSeverity()
	if level <= max {
		output.Write(l.MaxSeverity(), section, name)
	}
	output.Dump(level, section, l.Issues()...)
}

var systemNS = []string{"kube-system", "kube-public"}

func isSystemNS(ns string) bool {
	for _, n := range systemNS {
		if n == ns {
			return true
		}
	}
	return false
}

func isSet(s *string) bool {
	return s != nil && *s != ""
}

func getNamespacesOrDie(c *k8s.Client) []v1.Namespace {
	var ns generated.Namespace

	if isSet(k8sConfig.Namespace) {
		n, err := ns.Get(c, *k8sConfig.Namespace)
		if err != nil {
			log.Fatal().Err(err)
		}
		return []v1.Namespace{*n}
	}

	nn, err := ns.List(c)
	if err != nil {
		log.Fatal().Err(err)
	}

	ll := make([]v1.Namespace, 0, len(nn.Items))
	for _, n := range nn.Items {
		if !popConfig.allNamespaces && isSystemNS(n.Name) {
			continue
		}
		ll = append(ll, n)
	}
	return ll
}

func namespaceNames(nn []v1.Namespace) []string {
	ll := make([]string, 0, len(nn))
	for _, n := range nn {
		ll = append(ll, n.Name)
	}
	return ll
}

func namespaced(ns, n string) string {
	return ns + "/" + n
}

func initPopeyeFlags() {
	rootCmd.Flags().StringVarP(
		&popConfig.lintLevel,
		"lintLevel", "l",
		popConfig.lintLevel,
		"Specify a lint level (info, warn, error)",
	)

	rootCmd.Flags().BoolVarP(
		&popConfig.allNamespaces,
		"all-namespaces", "",
		popConfig.allNamespaces,
		"Includes system namespaces",
	)
}

func initK8sFlags() {
	k8sConfig = genericclioptions.NewConfigFlags(false)

	rootCmd.Flags().StringVar(
		k8sConfig.KubeConfig,
		"kubeconfig",
		"",
		"Path to the kubeconfig file to use for CLI requests",
	)

	rootCmd.Flags().StringVar(
		k8sConfig.Timeout,
		"request-timeout",
		"",
		"The length of time to wait before giving up on a single server request",
	)

	rootCmd.Flags().StringVar(
		k8sConfig.Context,
		"context",
		"",
		"The name of the kubeconfig context to use",
	)

	rootCmd.Flags().StringVar(
		k8sConfig.ClusterName,
		"cluster",
		"",
		"The name of the kubeconfig cluster to use",
	)

	rootCmd.Flags().StringVar(
		k8sConfig.AuthInfoName,
		"user",
		"",
		"The name of the kubeconfig user to use",
	)

	rootCmd.Flags().StringVar(
		k8sConfig.Impersonate,
		"as",
		"",
		"Username to impersonate for the operation",
	)

	rootCmd.Flags().StringArrayVar(
		k8sConfig.ImpersonateGroup,
		"as-group",
		[]string{},
		"Group to impersonate for the operation",
	)

	rootCmd.Flags().BoolVar(
		k8sConfig.Insecure,
		"insecure-skip-tls-verify",
		false,
		"If true, the server's caCertFile will not be checked for validity",
	)

	rootCmd.Flags().StringVar(
		k8sConfig.CAFile,
		"certificate-authority",
		"",
		"Path to a cert file for the certificate authority",
	)

	rootCmd.Flags().StringVar(
		k8sConfig.KeyFile,
		"client-key",
		"",
		"Path to a client key file for TLS",
	)

	rootCmd.Flags().StringVar(
		k8sConfig.CertFile,
		"client-certificate",
		"",
		"Path to a client certificate file for TLS",
	)

	rootCmd.Flags().StringVar(
		k8sConfig.BearerToken,
		"token",
		"",
		"Bearer token for authentication to the API server",
	)

	rootCmd.Flags().StringVarP(
		k8sConfig.Namespace,
		"namespace",
		"n",
		"",
		"If present, the namespace scope for this CLI request",
	)
}

// Helpers...

func clearScreen() {
	fmt.Print("\033[H\033[2J")
}

func parseLogLevel(level string) zerolog.Level {
	switch level {
	case "debug":
		return zerolog.DebugLevel
	case "warn":
		return zerolog.WarnLevel
	case "error":
		return zerolog.ErrorLevel
	case "fatal":
		return zerolog.FatalLevel
	default:
		return zerolog.InfoLevel
	}
}

func parseLintLevel(level string) linter.Level {
	switch level {
	case "info":
		return linter.InfoLevel
	case "warn":
		return linter.WarnLevel
	case "error":
		return linter.ErrorLevel
	default:
		return linter.OkLevel
	}
}
