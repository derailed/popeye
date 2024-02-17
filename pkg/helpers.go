// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Popeye

package pkg

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

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

	return filepath.Join(os.TempDir(), "popeye")
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
