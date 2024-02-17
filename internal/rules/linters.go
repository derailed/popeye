// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Popeye

package rules

import (
	"fmt"
	"strings"

	"github.com/rs/zerolog/log"
)

type LinterExcludes struct {
	Codes     expressions `yaml:"codes"`
	Instances Excludes    `yaml:"instances"`
}

func (l LinterExcludes) Dump(indent string) {
	l.Codes.dump(indent)
	l.Instances.Dump(indent)
}

func (l LinterExcludes) Match(spec Spec, global bool) bool {
	if l.Instances.Match(spec, global) {
		return true
	}

	if spec.Code == ZeroCode || len(l.Codes) == 0 {
		return false
	}

	return l.Codes.match(spec.Code.String())
}

type Linters map[string]LinterExcludes

func (l Linters) Dump(indent string) {
	for k, v := range l {
		fmt.Printf("%s%s\n", indent, k)
		v.Dump(strings.Repeat(indent, 2))
	}
}

func (l Linters) isEmpty() bool {
	return len(l) == 0
}

func (l Linters) Match(spec Spec, global bool) bool {
	if l.isEmpty() {
		return false
	}

	linter, ok := l[spec.GVR.R()]
	if !ok {
		log.Debug().Msgf("No exclusions found for linter: %q", spec.GVR.R())
		return false
	}

	return linter.Match(spec, global)
}
