// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Popeye

package rules

import (
	"fmt"
	"strings"
)

type Labels map[string]string

func (l Labels) String() string {
	if len(l) == 0 {
		return "n/a"
	}

	kk := make([]string, 0, len(l))
	for k := range l {
		kk = append(kk, k)
	}

	return strings.Join(kk, ",")
}

type keyVals map[string]expressions

func (kv keyVals) dump(indent string) {
	for k, v := range kv {
		fmt.Printf("%s%s: %s\n", indent, k, v)
	}
}

func (kv keyVals) isEmpty() bool {
	return len(kv) == 0
}

func (kv keyVals) match(ll Labels) bool {
	if len(kv) == 0 {
		return true
	}

	var matches int
	for k, ee := range kv {
		v, ok := ll[k]
		if !ok {
			continue
		}
		if ee.match(v) {
			matches++
		}
	}

	return matches > 0
}
