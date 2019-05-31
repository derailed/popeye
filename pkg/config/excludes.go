package config

import (
	"regexp"
	"strings"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func init() {
	zerolog.SetGlobalLevel(zerolog.FatalLevel)
}

// RxMarker indicate exclude flag is a regular expression.
const rxMarker = "rx:"

// RegExp defined regex to check if exclude is a regex or plain string.
var regExp = regexp.MustCompile(`\A` + rxMarker)

type (
	// Exclude represents a collection of excludes items.
	// This can be a straight string match of regex using an rx: prefix.
	Exclude []string
	// Excludes represents a set of resources that should be excluded
	// from the sanitizer.
	Excludes map[string]Exclude
)

func newExcludes() Excludes {
	return Excludes{}
}

// ShouldExclude checks if a given named resource should be excluded.
func (e Excludes) ShouldExclude(res, name string) bool {
	// Not mentioned in config. Allow all
	v, ok := e[res]
	if !ok {
		return false
	}

	return v.ShouldExclude(name)
}

// ShouldExclude checks if a given named should be excluded.
func (e Exclude) ShouldExclude(name string) bool {
	for _, n := range e {
		if !isRegex(n) {
			if n == name {
				return true
			}
			continue
		}

		rx, err := regexp.Compile(`\A` + strings.Replace(n, rxMarker, "", 1))
		if err != nil {
			log.Error().Err(err).Msgf("Invalid regexp `%s found in yaml. Skipping!", n)
			continue
		}
		if rx.MatchString(name) {
			return true
		}
	}

	return false
}

// IsRegex check if rx matching is in effect.
func isRegex(f string) bool {
	return regExp.MatchString(f)
}
