// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Popeye

package tally

import (
	"slices"
	"strconv"

	"github.com/derailed/popeye/internal/rules"
	"github.com/rs/zerolog/log"
)

// SevScore tracks per level total score.
type SevScore map[rules.Level]int

// Code tracks code issue counts.
type Code map[string]int

// Compact removes zero entries.
func (cc Code) Compact() {
	for c, v := range cc {
		if v == 0 {
			delete(cc, c)
		}
	}
}

// Rollup rollups code scores per severity.
func (cc Code) Rollup(gg rules.Glossary) SevScore {
	if len(cc) == 0 {
		return nil
	}
	ss := make(SevScore, len(cc))
	for sid, count := range cc {
		id, _ := strconv.Atoi(sid)
		c := gg[rules.ID(id)]
		ss[c.Severity] += count
	}

	return ss
}

// Merge merges two sets.
func (cc Code) Merge(cc1 Code) {
	for code, count := range cc1 {
		cc[code] += count
	}
}

// Dump for debugging.
func (cc Code) Dump(indent string) {
	kk := make([]string, 0, len(cc))
	for k := range cc {
		kk = append(kk, k)
	}
	slices.Sort(kk)
	for _, k := range kk {
		log.Debug().Msgf("%s%s: %d", indent, k, cc[k])
	}
}
