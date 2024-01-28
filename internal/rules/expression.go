// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Popeye

package rules

import (
	"fmt"
	"regexp"
	"strings"
)

type Expression string

func (e Expression) dump(indent string) {
	fmt.Printf("%s%q\n", indent, e)
}

func (e Expression) IsRX() bool {
	return strings.HasPrefix(string(e), rxMarker)
}

func (e Expression) MatchRX(s string) bool {
	rx := regexp.MustCompile(strings.Replace(string(e), rxMarker, "", 1))

	return rx.MatchString(s)
}

func (e Expression) match(s string) bool {
	if e == "" {
		return true
	}
	if e.IsRX() {
		return e.MatchRX(s)
	}

	return s == string(e)
}

type expressions []Expression

func (ee expressions) dump(indent string) {
	for _, e := range ee {
		e.dump(indent)
	}
}

func (ee expressions) isEmpty() bool {
	return len(ee) == 0
}

func (ee expressions) matches(ss []string) bool {
	if len(ee) == 0 {
		return true
	}

	for _, s := range ss {
		if ee.match(s) {
			return true
		}
	}

	return false
}

func (ee expressions) match(exp string) bool {
	if len(ee) == 0 || exp == "" {
		return true
	}
	for _, e := range ee {
		if e.match(exp) {
			return true
		}
	}

	return false
}
