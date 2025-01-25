// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Popeye

package cmd

import (
	"fmt"

	"github.com/derailed/popeye/internal/report"
	"github.com/derailed/popeye/pkg"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(infoCmd())
}

func infoCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "info",
		Short: "Prints Popeye info",
		Long:  "Prints Popeye information",
		Run: func(cmd *cobra.Command, args []string) {
			printInfo()
		},
	}
}

func printInfo() {
	printLogo(report.ColorAqua, report.ColorLighSlate)
	fmt.Println()
	printTuple("Logs", pkg.LogFile)
}
