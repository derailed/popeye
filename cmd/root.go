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
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

const (
	defaultLogLevel = "debug"
	outputWidth     = 80
)

var (
	version   = "dev"
	commit    = "dev"
	logLevel  string
	rootCmd   *cobra.Command
	k8sConfig *genericclioptions.ConfigFlags
)

func init() {
	rootCmd = &cobra.Command{
		Use:   "Popeye",
		Short: "A Kubernetes resource linter",
		Long:  `A Kubernetes resource linter`,
		Run:   run,
	}
	rootCmd.AddCommand(genCmd)

	initK8sFlags()
	initPopeyeFlags()
	zerolog.SetGlobalLevel(parseLevel(logLevel))
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
		fmt.Println(output.Colorize(s, output.ColorBriteBlue))
	}
	fmt.Println("Popeye Biffs'em and buff'em...\n")
	conn := k8s.NewServer(k8s.NewConfig(k8sConfig))
	mx := conn.ClusterHasMetrics()
	output.Write(linter.OkLevel, "Kubernetes", "Connectivity")

	if !mx {
		output.Write(linter.OkLevel, "Cluster", "Metrics")
	} else {
		output.Write(linter.OkLevel, "Cluster", "Metrics")
	}

	lint(conn)
}

func lint(s *k8s.Server) {
	lintNS(s)
	lintPod(s)
	lintSvc(s)
}

func lintNS(s *k8s.Server) {
	section := "Namespace"
	var gen generated.Namespace
	nn, err := gen.List(s)
	if err != nil {
		log.Fatal().Err(err)
	}

	for _, n := range nn.Items {
		l := linter.NewNamespace()
		l.Lint(n)
		if l.NoIssues() {
			output.Write(linter.OkLevel, section, n.Name)
			continue
		}
		output.Write(l.MaxSeverity(), section, n.Name)
		output.Dump(section, l.Issues()...)
	}
}

func lintPod(srv *k8s.Server) {
	section := "Pod"
	var ns generated.Namespace
	nn, err := ns.List(srv)
	if err != nil {
		log.Fatal().Err(err)
	}

	var gen generated.Pod
	for _, n := range nn.Items {
		pp, err := gen.List(srv, n.Name)
		if err != nil {
			log.Error().Err(err)
		}

		for _, p := range pp.Items {
			l := linter.NewPod()
			l.Lint(p)
			if l.NoIssues() {
				output.Write(linter.OkLevel, section, n.Name+"/"+p.Name)
				continue
			}
			output.Write(l.MaxSeverity(), section, n.Name+"/"+p.Name)
			output.Dump(section, l.Issues()...)
		}
	}
}

func lintSvc(srv *k8s.Server) {
	section := "Service"
	var ns generated.Namespace
	nn, err := ns.List(srv)
	if err != nil {
		log.Fatal().Err(err)
	}

	var gen generated.Service
	for _, n := range nn.Items {
		ss, err := gen.List(srv, n.Name)
		if err != nil {
			log.Error().Err(err)
		}

		for _, s := range ss.Items {
			l := linter.NewService()
			l.Lint(s)
			if l.NoIssues() {
				output.Write(linter.OkLevel, section, n.Name+"/"+s.Name)
				continue
			}
			output.Write(l.MaxSeverity(), section, n.Name+"/"+s.Name)
			output.Dump(section, l.Issues()...)
		}
	}
}

func initPopeyeFlags() {
	rootCmd.Flags().StringVarP(
		&logLevel,
		"logLevel", "l",
		defaultLogLevel,
		"Specify a log level (info, warn, debug, error, fatal, panic, trace)",
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

func parseLevel(level string) zerolog.Level {
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
