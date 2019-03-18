package cmd

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"strings"
	"text/template"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

type genCfg struct {
	Version    string
	Resource   string
	Namespaced bool
}

var (
	genCmd    *cobra.Command
	genConfig genCfg
)

func init() {
	genCmd = &cobra.Command{
		Use:   "gen",
		Short: "Generates K8s artifact",
		Long:  "Generates K8s artifact",
		Run:   generate,
	}
	initGenFlags()
}

func versionResource(rev, resource string) string {
	return rev + "." + resource
}

func generators() []genCfg {
	return []genCfg{
		{"v1", "Pod", true},
		{"v1", "Service", true},
		{"v1", "Namespace", false},
	}
}

func generate(cmd *cobra.Command, args []string) {
	fmt.Println("Generating resources...")

	fMap := template.FuncMap{
		"vr":    versionResource,
		"title": strings.Title,
	}

	res, err := ioutil.ReadFile("../../assets/cluster.tpl")
	if err != nil {
		log.Fatal().Msg(err.Error())
	}
	tpl := template.New("Global").Funcs(fMap)
	if tpl, err = tpl.Parse(string(res)); err != nil {
		log.Fatal().Msg(err.Error())
	}

	nsRes, err := ioutil.ReadFile("../../assets/namespaced.tpl")
	if err != nil {
		log.Fatal().Msg(err.Error())
	}
	nsTpl := template.New("Namespaced").Funcs(fMap)
	if nsTpl, err = nsTpl.Parse(string(nsRes)); err != nil {
		log.Fatal().Msg(err.Error())
	}

	for _, g := range generators() {
		var (
			buff bytes.Buffer
			err  error
		)

		if g.Namespaced {
			err = nsTpl.Execute(&buff, g)
		} else {
			err = tpl.Execute(&buff, g)
		}
		if err != nil {
			log.Fatal().Msg(err.Error())
		}
		code, err := ioutil.ReadAll(&buff)
		if err != nil {
			log.Fatal().Msg(err.Error())
		}
		log.Info().Msgf("Generating `generated/%s.go", strings.ToLower(g.Resource))
		if err := ioutil.WriteFile(fmt.Sprintf("generated/%s.go", strings.ToLower(g.Resource)), code, 0644); err != nil {
			log.Fatal().Msg(err.Error())
		}
	}
	fmt.Println("Done!")
}

func initGenFlags() {
	genCmd.Flags().StringVarP(
		&genConfig.Version,
		"version", "v",
		"v1",
		"Specify K8s api version.",
	)

	genCmd.Flags().StringVarP(
		&genConfig.Resource,
		"resource", "r",
		"",
		"Specify K8s resource.",
	)

	genCmd.Flags().BoolVarP(
		&genConfig.Namespaced,
		"namespace", "n",
		true,
		"Specify wheither the K8s resource is namespaced.",
	)
}
