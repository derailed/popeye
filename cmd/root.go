package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/derailed/popeye/internal/config"
	"github.com/derailed/popeye/internal/report"
	"github.com/derailed/popeye/pkg"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

var (
	version   = "dev"
	commit    = "dev"
	popConfig = config.New()
	rootCmd   *cobra.Command
	k8sFlags  *genericclioptions.ConfigFlags
)

func init() {
	rootCmd = &cobra.Command{
		Use:   "popeye",
		Short: "A Kubernetes Cluster sanitizer and linter",
		Long:  `Popeye scans your Kubernetes clusters and reports potential resource issues.`,
		Run:   doIt,
	}

	initK8sFlags()
	initPopeyeFlags()
}

func bomb(msg string) {
	fmt.Printf("💥 %s\n", report.Colorize(msg, report.ColorRed))
	os.Exit(1)
}

// Execute root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		bomb(fmt.Sprintf("Exec failed %s", err))
	}
}

// Doit runs the scans and lints pass over the specified cluster.
func doIt(cmd *cobra.Command, args []string) {
	if err := popConfig.Init(k8sFlags); err != nil {
		bomb(fmt.Sprintf("Spinach load failed %s", popConfig.Spinach))
	}

	zerolog.SetGlobalLevel(popConfig.Popeye.LogLevel)
	clearScreen()
	printHeader()
	pkg.New(popConfig).Sanitize()
}

func initPopeyeFlags() {
	rootCmd.Flags().StringVarP(
		&popConfig.LintLevel,
		"lintLevel", "l",
		popConfig.LintLevel,
		"Specify a lint level (ok, info, warn, error)",
	)

	rootCmd.Flags().BoolVarP(
		&popConfig.ClearScreen,
		"clear", "c",
		popConfig.ClearScreen,
		"Clears the screen before a run",
	)

	rootCmd.Flags().StringVarP(
		&popConfig.Spinach,
		"file", "f",
		"",
		"Use a spinach YAML configuration file",
	)

	rootCmd.Flags().StringSliceVarP(
		&popConfig.Sections,
		"sections", "s",
		popConfig.Sections,
		"Specifies which resources to include in the scan ie -s po,svc",
	)
}

func initK8sFlags() {
	k8sFlags = genericclioptions.NewConfigFlags(false)

	rootCmd.Flags().StringVar(
		k8sFlags.KubeConfig,
		"kubeconfig",
		"",
		"Path to the kubeconfig file to use for CLI requests",
	)

	rootCmd.Flags().StringVar(
		k8sFlags.Timeout,
		"request-timeout",
		"",
		"The length of time to wait before giving up on a single server request",
	)

	rootCmd.Flags().StringVar(
		k8sFlags.Context,
		"context",
		"",
		"The name of the kubeconfig context to use",
	)

	rootCmd.Flags().StringVar(
		k8sFlags.ClusterName,
		"cluster",
		"",
		"The name of the kubeconfig cluster to use",
	)

	rootCmd.Flags().StringVar(
		k8sFlags.AuthInfoName,
		"user",
		"",
		"The name of the kubeconfig user to use",
	)

	rootCmd.Flags().StringVar(
		k8sFlags.Impersonate,
		"as",
		"",
		"Username to impersonate for the operation",
	)

	rootCmd.Flags().StringArrayVar(
		k8sFlags.ImpersonateGroup,
		"as-group",
		[]string{},
		"Group to impersonate for the operation",
	)

	rootCmd.Flags().BoolVar(
		k8sFlags.Insecure,
		"insecure-skip-tls-verify",
		false,
		"If true, the server's caCertFile will not be checked for validity",
	)

	rootCmd.Flags().StringVar(
		k8sFlags.CAFile,
		"certificate-authority",
		"",
		"Path to a cert file for the certificate authority",
	)

	rootCmd.Flags().StringVar(
		k8sFlags.KeyFile,
		"client-key",
		"",
		"Path to a client key file for TLS",
	)

	rootCmd.Flags().StringVar(
		k8sFlags.CertFile,
		"client-certificate",
		"",
		"Path to a client certificate file for TLS",
	)

	rootCmd.Flags().StringVar(
		k8sFlags.BearerToken,
		"token",
		"",
		"Bearer token for authentication to the API server",
	)

	rootCmd.Flags().StringVarP(
		k8sFlags.Namespace,
		"namespace",
		"n",
		"",
		"If present, the namespace scope for this CLI request",
	)
}

// Helpers...

func clearScreen() {
	if popConfig.ClearScreen {
		fmt.Print("\033[H\033[2J")
	}
}

func printHeader() {
	fmt.Println()
	for i, s := range report.Logo {
		if i < len(report.Popeye) {
			fmt.Printf(report.Colorize(report.Popeye[i], report.ColorAqua))
			fmt.Printf(strings.Repeat(" ", 35))
		} else {
			if i == 4 {
				fmt.Printf(report.Colorize("  Biffs`em and Buffs`em!", report.ColorLighSlate))
				fmt.Printf(strings.Repeat(" ", 38))
			} else {
				fmt.Printf(strings.Repeat(" ", 62))
			}
		}
		fmt.Println(report.Colorize(s, report.ColorLighSlate))
	}
	fmt.Println("")
}
