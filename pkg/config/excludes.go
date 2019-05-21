package config

import (
	"regexp"
	"strings"
)

// RxMarker indicate exclude flag is a regular expression.
const rxMarker = "rx:"

// RegExp defined regex to check if exclude is a regex or plain string.
var regExp = regexp.MustCompile(`\A` + rxMarker)

// Excludes represents a lists items that should be excluded.
type Excludes []string

func (e Excludes) excluded(name string) bool {
	for _, n := range e {
		if isRegex(n) {
			n = `\A` + strings.Replace(n, rxMarker, "", 1)
			rx := regexp.MustCompile(n)
			if rx.MatchString(name) {
				return true
			}
		}
		// Fallback string matcher
		if n == name {
			return true
		}
	}

	return false
}

func isRegex(f string) bool {
	return regExp.MatchString(f)
}
