package cmd

import (
	"fmt"
	"strings"

	"github.com/derailed/popeye/internal/report"
	"github.com/derailed/popeye/pkg"
	"github.com/spf13/cobra"
)

func versionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Prints version/build info",
		Long:  "Prints version/build information",
		Run: func(cmd *cobra.Command, args []string) {
			printVersion()
		},
	}
}

func printVersion() {
	const secFmt = "%-10s "

	printLogo()
	printTuple(secFmt, "Version", version)
	printTuple(secFmt, "Commit", commit)
	printTuple(secFmt, "Date", date)
	printTuple(secFmt, "Logs", pkg.PopeyeLog)
}

func printTuple(format, section, value string) {
	fmt.Printf(report.Colorize(fmt.Sprintf(format, section+":"), report.ColorAqua))
	fmt.Println(report.Colorize(value, report.ColorWhite))
}

func printLogo() {
	for i, s := range report.Logo {
		if i < len(report.Popeye) {
			fmt.Printf(report.Colorize(report.Popeye[i], report.ColorAqua))
			fmt.Printf(strings.Repeat(" ", 22))
		} else {
			if i == 4 {
				fmt.Printf(report.Colorize("  Biffs`em and Buffs`em!", report.ColorLighSlate))
				fmt.Printf(strings.Repeat(" ", 26))
			} else {
				fmt.Printf(strings.Repeat(" ", 50))
			}
		}
		fmt.Println(report.Colorize(s, report.ColorLighSlate))
	}
}
