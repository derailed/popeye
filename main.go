package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/derailed/popeye/cmd"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

var popeye = filepath.Join(os.TempDir(), fmt.Sprintf("popeye.log"))

func init() {
	mod := os.O_CREATE | os.O_APPEND | os.O_WRONLY
	if file, err := os.OpenFile(popeye, mod, 0644); err == nil {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: file})
	} else {
		fmt.Printf("Unable to create Popeye log file %v. Exiting...", err)
		os.Exit(1)
	}
}

func main() {
	cmd.Execute()
}
