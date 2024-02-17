// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Popeye

package rules

import (
	"fmt"

	"github.com/rs/zerolog/log"
)

type Exclusions struct {
	// Excludes tracks exclusions
	Global Exclude `yaml:"global"`

	// Linters tracks exclusions
	Linters Linters `yaml:"linters"`
}

func NewExclusions() Exclusions {
	return Exclusions{
		Global:  NewExclude(),
		Linters: make(Linters),
	}
}

func (e Exclusions) Match(spec Spec) bool {
	if e.Global.Match(spec, true) {
		log.Debug().Msgf("Global exclude matched: %q::%q", spec.GVR, spec.FQN)
		return true
	}

	return e.Linters.Match(spec, false)
}

func (e Exclusions) Dump() {
	fmt.Println("Globals")
	e.Global.Dump("  ")

	fmt.Println("Linters")
	e.Linters.Dump("  ")
}
