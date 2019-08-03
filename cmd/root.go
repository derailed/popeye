package cmd

import (
	"fmt"
	"os"
	"runtime/debug"
	"strings"

	"github.com/derailed/popeye/internal/report"
	"github.com/derailed/popeye/pkg"
	"github.com/derailed/popeye/pkg/config"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var (
	version = "dev"
	commit  = "dev"
	date    = "n/a"
	flags   = config.NewFlags()
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
	zerolog.SetGlobalLevel(zerolog.DebugLevel)

	defer func() {
		if err := recover(); err != nil {
			printMsgLogo("DOH", "X", report.ColorOrangish, report.ColorRed)
			fmt.Printf("\n\nBoom! %v\n", err)
			log.Error().Msgf("%v", err)
			log.Error().Msg(string(debug.Stack()))
			os.Exit(1)
		}
	}()

	clearScreen()
	popeye, err := pkg.NewPopeye(flags, &log.Logger)
	if err != nil {
		bomb(fmt.Sprintf("Popeye configuration load failed %v", err))
	}
	if err := popeye.Init(); err != nil {
		bomb(err.Error())
	}
	if err := popeye.Sanitize(); err != nil {
		bomb(err.Error())
	}
}

func bomb(msg string) {
	panic(fmt.Sprintf("💥 %s\n", report.Colorize(msg, report.ColorRed)))
}

func initFlags() {
	rootCmd.Flags().StringVarP(
		flags.Output,
		"out",
		"o",
		"standard",
		"Specify the output type (standard, jurassic, yaml, json)",
	)

	rootCmd.Flags().BoolVarP(
		flags.Save,
		"save",
		"",
		false,
		"Specify if you want Popeye to persist the output to a file",
	)

	rootCmd.Flags().StringVarP(
		flags.LintLevel,
		"lint",
		"l",
		"ok",
		"Specify a lint level (ok, info, warn, error)",
	)

	rootCmd.PersistentFlags().BoolVarP(
		flags.ClearScreen,
		"clear",
		"c",
		false,
		"Clears the screen before a run",
	)

	rootCmd.Flags().BoolVarP(
		flags.CheckOverAllocs,
		"over-allocs",
		"",
		false,
		"Check for cpu/memory over allocations",
	)

	rootCmd.Flags().BoolVarP(
		flags.AllNamespaces,
		"all-namespaces",
		"A",
		false,
		"Sanitize all namespaces",
	)

	rootCmd.Flags().StringVarP(
		flags.Spinach,
		"file",
		"f",
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

// ----------------------------------------------------------------------------
// Helpers...

func clearScreen() {
	if flags.ClearScreen == nil || !*flags.ClearScreen {
		return
	}
	fmt.Print("\033[H\033[2J")
}

func printSosLogo(title, logo report.Color) {
	printMsgLogo("SOS", "O", title, logo)
}

func printMsgLogo(msg, eye string, title, logo report.Color) {
	for i, s := range report.GraderLogo {
		switch i {
		case 0, 1, 2:
			s = strings.Replace(s, "o", string(msg[i]), 1)
		case 3:
			s = strings.Replace(s, "a", eye, 1)
		}

		if i < len(report.Popeye) {
			fmt.Printf(report.Colorize(report.Popeye[i], title))
			fmt.Printf(strings.Repeat(" ", 22))
		} else {
			if i == 4 {
				fmt.Printf(report.Colorize("  Biffs`em and Buffs`em!", logo))
				fmt.Printf(strings.Repeat(" ", 26))
			} else {
				fmt.Printf(strings.Repeat(" ", 50))
			}
		}
		fmt.Println(report.Colorize(s, logo))
	}
}
