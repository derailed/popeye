// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Popeye

package tally

import (
	"slices"
	"strings"

	"github.com/rs/zerolog/log"
)

// Namespace tracks each namespace code tally.
type Namespace map[string]Code

// Compact compacts set by removing zero entries.
func (nn Namespace) Compact() {
	for ns, v := range nn {
		v.Compact()
		if len(v) == 0 {
			delete(nn, ns)
		}
	}
}

// Merge merges 2 sets.
func (nn Namespace) Merge(t Namespace) {
	for k, v := range t {
		if v1, ok := nn[k]; ok {
			nn[k].Merge(v1)
		} else {
			nn[k] = v
		}
	}
}

// Dump for debugging.
func (s Namespace) Dump(indent string) {
	kk := make([]string, 0, len(s))
	for k := range s {
		kk = append(kk, k)
	}
	slices.Sort(kk)
	for _, k := range kk {
		log.Debug().Msgf("%s%s", indent, k)
		s[k].Dump(strings.Repeat(indent, 2))
	}
}
