// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Popeye

package pkg

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"

	"github.com/derailed/popeye/internal/report"
	"github.com/rs/zerolog/log"
)

var DefaultDumpDir = filepath.Join(os.TempDir(), "popeye")

func BailOut(err error) {
	printMsgLogo("DOH", "X", report.ColorOrangish, report.ColorRed)
	fmt.Printf("\n\nBoom! %v (see logs)\n", err)
	log.Error().Msgf("%v", err)
	log.Error().Msg(string(debug.Stack()))
	os.Exit(1)
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
			fmt.Printf("%s", report.Colorize(report.Popeye[i], title))
			fmt.Printf("%s", strings.Repeat(" ", 22))
		} else {
			if i == 4 {
				fmt.Printf("%s", report.Colorize("  Biffs`em and Buffs`em!", logo))
				fmt.Printf("%s", strings.Repeat(" ", 26))
			} else {
				fmt.Printf("%s", strings.Repeat(" ", 50))
			}
		}
		fmt.Println(report.Colorize(s, logo))
	}
}

func ensureDir(path string, mod os.FileMode) error {
	dir, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	_, err = os.Stat(dir)
	if err == nil || !os.IsNotExist(err) {
		return nil
	}
	if err = os.MkdirAll(dir, mod); err != nil {
		return fmt.Errorf("fail to create popeye scan dump dir: %w", err)
	}

	return nil
}

func dumpDir() string {
	if d := os.Getenv("POPEYE_REPORT_DIR"); d != "" {
		return d
	}

	return DefaultDumpDir
}

type readWriteCloser struct {
	io.ReadWriter
}

// Close close read stream.
func (readWriteCloser) Close() error {
	return nil
}

// NopCloser fake closer.
func NopCloser(i io.ReadWriter) io.ReadWriteCloser {
	return &readWriteCloser{i}
}
