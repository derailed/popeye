// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Popeye

package rules

import (
	"regexp"
	"strings"
)

const rxMarker = "rx:"

func rxMatch(exp, name string) (bool, error) {
	if !isRegex(exp) {
		return false, nil
	}
	rx, err := regexp.Compile(strings.Replace(exp, rxMarker, "", 1))
	if err != nil {
		return false, err
	}

	return rx.MatchString(name), nil
}

func isRegex(s string) bool {
	return strings.HasPrefix(s, rxMarker)
}
