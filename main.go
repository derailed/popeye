package main

import (
	"fmt"
	"os"

	"github.com/derailed/popeye/cmd"
	"github.com/derailed/popeye/pkg"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

func init() {
	mod := os.O_CREATE | os.O_APPEND | os.O_WRONLY
	if file, err := os.OpenFile(pkg.LogFile, mod, 0644); err == nil {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: file})
	} else {
		fmt.Printf("Unable to create Popeye log file %v. Exiting...", err)
		os.Exit(1)
	}
}

func main() {
	cmd.Execute()
}
