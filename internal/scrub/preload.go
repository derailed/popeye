// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Popeye

package scrub

import "github.com/derailed/popeye/internal"

type Preloads map[internal.R]LoaderFn

func (p Preloads) Merge(ll Preloads) {
	for k, v := range ll {
		if _, ok := p[k]; ok {
			continue
		}
		p[k] = v
	}
}
