package main

import (
	"os"

	"github.com/derailed/popeye/cmd"
	// "github.com/pkg/profile"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

func init() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
}

func main() {
	// defer profile.Start(profile.TraceProfile).Stop()
	cmd.Execute()
}
