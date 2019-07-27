package cmd

import (
	"fmt"
	"os"
	"runtime/debug"

	"github.com/derailed/popeye/internal/report"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var cluster string

func diffCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "diff",
		Short: "Compares sanitizer reports",
		Long:  "Compares sanitizer reports",
		Run: func(cmd *cobra.Command, args []string) {
			run(cmd, args)
		},
	}

	c.Flags().StringVarP(
		&cluster,
		"cluster",
		"",
		"",
		"Specify which cluster you are targetting (required)",
	)

	return c
}

func run(cmd *cobra.Command, args []string) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("YO!!")
			printSosLogo(report.ColorOrangish, report.ColorRed)
			fmt.Printf("\n\nBoom! %v\n", err)
			fmt.Println(string(debug.Stack()))
			log.Error().Msgf("%v", err)
			log.Error().Msg(string(debug.Stack()))
			os.Exit(1)
		}
	}()

	if cluster == "" {
		panic("You must specify a cluster name")
	}

	clearScreen()
	printLogo(report.ColorAqua, report.ColorLighSlate)

	diff := report.NewDiff(os.Stdout, *flags.Output == "jurassic")
	if err := diff.Run(cluster); err != nil {
		panic(err)
	}
}
