package cmd

import (
	"fmt"
	"os"
	"runtime/debug"

	"github.com/derailed/popeye/internal/k8s"
	"github.com/derailed/popeye/internal/report"
	"github.com/derailed/popeye/pkg"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var (
	version = "dev"
	commit  = "dev"
	date    = "n/a"
	flags   = k8s.NewFlags()
	rootCmd = &cobra.Command{
		Use:   "popeye",
		Short: "A Kubernetes Cluster sanitizer and linter",
		Long:  `Popeye scans your Kubernetes clusters and reports potential resource issues.`,
		Run:   doIt,
	}
)

func init() {
	rootCmd.AddCommand(versionCmd())

	initFlags()
}

// Execute root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		bomb(fmt.Sprintf("Exec failed %s", err))
	}
}

// Doit runs the scans and lints pass over the specified cluster.
func doIt(cmd *cobra.Command, args []string) {
	defer func() {
		if err := recover(); err != nil {
			bomb(fmt.Sprintf("%v", err))
			log.Error().Msgf("%v", err)
			log.Error().Msg(string(debug.Stack()))
		}
	}()

	zerolog.SetGlobalLevel(zerolog.DebugLevel)

	clearScreen()
	popeye, err := pkg.NewPopeye(flags, &log.Logger, os.Stdout)
	if err != nil {
		bomb(fmt.Sprintf("Popeye configuration load failed %v", err))
	}
	popeye.Sanitize()
}

func bomb(msg string) {
	fmt.Printf("💥 %s\n", report.Colorize(msg, report.ColorRed))
	os.Exit(1)
}

func initFlags() {
	rootCmd.Flags().StringVarP(
		flags.Output,
		"out", "o",
		"pimpy",
		"Specify the output type (standard, jurassic, yaml, json)",
	)

	rootCmd.Flags().StringVarP(
		flags.LintLevel,
		"lint", "l",
		"ok",
		"Specify a lint level (ok, info, warn, error)",
	)

	rootCmd.Flags().BoolVarP(
		flags.ClearScreen,
		"clear", "c",
		false,
		"Clears the screen before a run",
	)

	rootCmd.Flags().StringVarP(
		flags.Spinach,
		"file", "f",
		"",
		"Use a spinach YAML configuration file",
	)

	rootCmd.Flags().StringSliceVarP(
		flags.Sections,
		"sections", "s",
		[]string{},
		"Specifies which resources to include in the scan ie -s po,svc",
	)

	rootCmd.Flags().StringVar(
		flags.KubeConfig,
		"kubeconfig",
		"",
		"Path to the kubeconfig file to use for CLI requests",
	)

	rootCmd.Flags().StringVar(
		flags.Timeout,
		"request-timeout",
		"",
		"The length of time to wait before giving up on a single server request",
	)

	rootCmd.Flags().StringVar(
		flags.Context,
		"context",
		"",
		"The name of the kubeconfig context to use",
	)

	rootCmd.Flags().StringVar(
		flags.ClusterName,
		"cluster",
		"",
		"The name of the kubeconfig cluster to use",
	)

	rootCmd.Flags().StringVar(
		flags.AuthInfoName,
		"user",
		"",
		"The name of the kubeconfig user to use",
	)

	rootCmd.Flags().StringVar(
		flags.Impersonate,
		"as",
		"",
		"Username to impersonate for the operation",
	)

	rootCmd.Flags().StringArrayVar(
		flags.ImpersonateGroup,
		"as-group",
		[]string{},
		"Group to impersonate for the operation",
	)

	rootCmd.Flags().BoolVar(
		flags.Insecure,
		"insecure-skip-tls-verify",
		false,
		"If true, the server's caCertFile will not be checked for validity",
	)

	rootCmd.Flags().StringVar(
		flags.CAFile,
		"certificate-authority",
		"",
		"Path to a cert file for the certificate authority",
	)

	rootCmd.Flags().StringVar(
		flags.KeyFile,
		"client-key",
		"",
		"Path to a client key file for TLS",
	)

	rootCmd.Flags().StringVar(
		flags.CertFile,
		"client-certificate",
		"",
		"Path to a client certificate file for TLS",
	)

	rootCmd.Flags().StringVar(
		flags.BearerToken,
		"token",
		"",
		"Bearer token for authentication to the API server",
	)

	rootCmd.Flags().StringVarP(
		flags.Namespace,
		"namespace",
		"n",
		"",
		"If present, the namespace scope for this CLI request",
	)
}

// Helpers...

func clearScreen() {
	if flags.ClearScreen == nil || !*flags.ClearScreen {
		return
	}
	fmt.Print("\033[H\033[2J")
}
