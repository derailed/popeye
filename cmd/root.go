package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/derailed/popeye/internal/config"
	"github.com/derailed/popeye/internal/k8s"
	"github.com/derailed/popeye/internal/linter"
	"github.com/derailed/popeye/internal/report"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

type (
	// Reporter obtains lint reports
	Reporter interface {
		MaxSeverity(res string) linter.Level
		Issues() linter.Issues
	}

	// Linter represents a resource linter.
	Linter interface {
		Reporter
		Lint(context.Context) error
	}

	// Linters a collection of linters.
	Linters map[string]Linter
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
		Short: "A Kubernetes Cluster Linter and issues Scanner",
		Long:  `Popeye scans your Kubernetes clusters and reports potential issues.`,
		Run:   doIt,
	}

	initK8sFlags()
	initPopeyeFlags()
}

func linters(c *k8s.Client) Linters {
	return Linters{
		"NODE":      linter.NewNode(c, &log.Logger),
		"NAMESPACE": linter.NewNamespace(c, &log.Logger),
		"POD":       linter.NewPod(c, &log.Logger),
		"SERVICE":   linter.NewService(c, &log.Logger),
	}
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
	lint()
}

func lint() {
	c := k8s.NewClient(popConfig)

	clusterInfo(c)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	for k, v := range linters(c) {
		if err := v.Lint(ctx); err != nil {
			w := bufio.NewWriter(os.Stdout)
			defer w.Flush()
			report.Open(w, k)
			{
				report.Error(w, "Scan failed! %v", err)
			}
			report.Close(w)
			continue
		}
		printReport(v, k)
	}
}

func printReport(r Reporter, section string) {
	w := bufio.NewWriter(os.Stdout)
	defer w.Flush()

	level := linter.Level(popConfig.Popeye.LintLevel)
	var wrote bool
	report.Open(w, section)
	{
		for res, issues := range r.Issues() {
			if len(issues) == 0 {
				if level <= linter.OkLevel {
					wrote = true
					report.Write(w, linter.OkLevel, 1, res)
				}
				continue
			}
			max := r.MaxSeverity(res)
			if level <= max {
				wrote = true
				report.Write(w, max, 1, res)
			}
			report.Dump(w, level, issues...)
		}
		if !wrote {
			report.Comment(w, report.Colorize("Section excluded from report.", report.ColorOrangish))
		}
	}

	report.Close(w)
}

func clusterInfo(c *k8s.Client) {
	report.Open(os.Stdout, fmt.Sprintf("CLUSTER [%s]", strings.ToUpper(c.Config.ActiveCluster())))
	{
		report.Write(os.Stdout, linter.OkLevel, 1, "Connectivity")

		if !c.ClusterHasMetrics() {
			report.Write(os.Stdout, linter.OkLevel, 1, "Metrics")
		} else {
			report.Write(os.Stdout, linter.OkLevel, 1, "Metrics")
		}
	}
	report.Close(os.Stdout)
}

func initPopeyeFlags() {
	rootCmd.Flags().StringVarP(
		&popConfig.LintLevel,
		"lintLevel", "l",
		popConfig.LintLevel,
		"Specify a lint level (info, warn, error)",
	)

	rootCmd.Flags().BoolVarP(
		&popConfig.AllNamespaces,
		"all-namespaces", "",
		popConfig.AllNamespaces,
		"Includes system namespaces",
	)

	rootCmd.Flags().StringVarP(
		&popConfig.Spinach,
		"file", "f",
		"",
		"Use this spinach configuration file",
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
