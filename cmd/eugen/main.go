package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"text/template"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

type eugenConfig struct {
	Version    string
	Resource   string
	Namespaced bool
}

var (
	genCmd *cobra.Command
	genCfg eugenConfig
)

func init() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	flag.StringVar(&genCfg.Version, "v", "v1", "Specify the resource api version")
	flag.StringVar(&genCfg.Resource, "r", "", "Specify the resource name")
	flag.BoolVar(&genCfg.Namespaced, "n", true, "Specify if the resource is namespaced or not")
	flag.Parse()
}

func generators() []eugenConfig {
	return []eugenConfig{
		{"v1", "Pod", true},
		{"v1", "Service", true},
		{"v1", "Namespace", false},
	}
}

func main() {
	fmt.Println("Eugen resources...")

	gen, err := generatorFor("cmd/eugen/templates/cluster.tpl", "cluster")
	if err != nil {
		log.Fatal().Msg(err.Error())
	}

	ngen, err := generatorFor("cmd/eugen/templates/namespaced.tpl", "namespaced")
	if err != nil {
		log.Fatal().Msg(err.Error())
	}

	for _, g := range generators() {
		var (
			buff bytes.Buffer
			err  error
		)

		if g.Namespaced {
			err = ngen.Execute(&buff, g)
		} else {
			err = gen.Execute(&buff, g)
		}
		if err != nil {
			log.Fatal().Msg(err.Error())
		}

		code, err := ioutil.ReadAll(&buff)
		if err != nil {
			log.Fatal().Msg(err.Error())
		}
		log.Info().Msgf("Generating `generated/%s.go", strings.ToLower(g.Resource))
		if err := ioutil.WriteFile(fmt.Sprintf("internal/k8s/generated/%s.go", strings.ToLower(g.Resource)), code, 0644); err != nil {
			log.Fatal().Msg(err.Error())
		}
	}
	fmt.Println("Eugen is done!")
}

func versionResource(rev, resource string) string {
	return rev + "." + resource
}

func generatorFor(file, name string) (*template.Template, error) {
	fMap := template.FuncMap{
		"vr":    versionResource,
		"title": strings.Title,
	}

	res, err := ioutil.ReadFile(file)
	if err != nil {
		log.Fatal().Msg(err.Error())
		return nil, err
	}

	tpl := template.New(name).Funcs(fMap)
	if tpl, err = tpl.Parse(string(res)); err != nil {
		log.Fatal().Msg(err.Error())
	}
	return tpl, nil
}
