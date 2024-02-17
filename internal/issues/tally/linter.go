// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Popeye

package tally

import (
	"slices"

	"github.com/rs/zerolog/log"
)

// Linter tracks linters namespace tallies.
type Linter map[string]Namespace

func (l Linter) Compact() {
	for linter, v := range l {
		v.Compact()
		if len(v) == 0 {
			delete(l, linter)
		}
	}
}

func (s Linter) Dump() {
	kk := make([]string, 0, len(s))
	for k := range s {
		kk = append(kk, k)
	}
	slices.Sort(kk)
	for _, k := range kk {
		log.Debug().Msgf("%s", k)
		s[k].Dump("  ")
	}
}
