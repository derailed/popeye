// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Popeye

package client

import (
	"fmt"
	"regexp"
	"strconv"

	"k8s.io/apimachinery/pkg/version"
)

// Revision tracks server version.
type Revision struct {
	Info         *version.Info
	Major, Minor int
}

var minorRX = regexp.MustCompile(`(\d+)\+?`)

// NewRevision returns a new instance.
func NewRevision(info *version.Info) (*Revision, error) {
	major, err := strconv.Atoi(info.Major)
	if err != nil {
		return nil, fmt.Errorf("unable to extract major %q", info.Major)
	}
	minors := minorRX.FindStringSubmatch(info.Minor)
	if len(minors) < 2 {
		return nil, fmt.Errorf("unable to extract minor %q", info.Minor)
	}
	minor, err := strconv.Atoi(minors[1])
	if err != nil {
		return nil, err
	}
	return &Revision{Info: info, Major: major, Minor: minor}, nil
}
