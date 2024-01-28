// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Popeye

package rules

import (
	"fmt"
	"strings"

	"github.com/rs/zerolog/log"
)

type Excludes []Exclude

func (ee Excludes) Dump(indent string) {
	for i, e := range ee {
		fmt.Printf("%s[%d]\n", strings.Repeat(indent, 2), i)
		e.Dump(strings.Repeat(indent, 2))
	}
}

func (ee Excludes) Match(spec Spec, global bool) bool {
	if len(ee) == 0 {
		return false
	}

	for _, e := range ee {
		if e.Match(spec, global) {
			return true
		}
	}

	return false
}

type Exclude struct {
	FQNs        expressions `yam:"FQNs"`
	Labels      keyVals     `yaml:"labels"`
	Annotations keyVals     `yaml:"annotations"`
	Codes       expressions `yaml:"codes"`
	Containers  expressions `yaml:"containers"`
}

// NewExclude returns a new instance.
func NewExclude() Exclude {
	return Exclude{
		Labels:      make(keyVals),
		Annotations: make(keyVals),
	}
}

func (e Exclude) Dump(indent string) {
	fmt.Printf("%sFQNS\n", indent)
	e.FQNs.dump(strings.Repeat(indent, 2))
	fmt.Printf("%sLABELS\n", indent)
	e.Labels.dump(strings.Repeat(indent, 2))
	fmt.Printf("%sANNOTS\n", indent)
	e.Annotations.dump(strings.Repeat(indent, 2))
	fmt.Printf("%sCODES\n", indent)
	e.Codes.dump(strings.Repeat(indent, 2))
	fmt.Printf("%sCONTAINERS\n", indent)
	e.Containers.dump(strings.Repeat(indent, 2))
}

func (e Exclude) String() string {
	return fmt.Sprintf("ns: %s ll: %s aa: %s cds: %s cos: %s", e.FQNs, e.Labels, e.Annotations, e.Codes, e.Containers)
}

func (e Exclude) isEmpty() bool {
	return e.FQNs.isEmpty() &&
		e.Labels.isEmpty() &&
		e.Annotations.isEmpty() &&
		e.Codes.isEmpty() &&
		e.Containers.isEmpty()
}

func (e Exclude) matchGlob(spec Spec) bool {
	log.Debug().Msgf("GlobalEX -- %s", spec)
	log.Debug().Msgf("  Rule: %s", e)

	var matches int
	if len(e.FQNs) > 0 && e.FQNs.match(spec.FQN) {
		log.Debug().Msgf("  match fqn: %q -- %s", spec.FQN, e.FQNs)
		matches++
	}
	if len(e.Labels) > 0 && e.Labels.match(spec.Labels) {
		log.Debug().Msgf("  match labels: %s -- %s", spec.Labels, e.Labels)
		matches++
	}
	if len(e.Annotations) > 0 && e.Annotations.match(spec.Annotations) {
		log.Debug().Msgf("  match anns: %s -- %s", spec.Annotations, e.Annotations)
		matches++
	}
	if len(e.Containers) > 0 && e.Containers.matches(spec.Containers) {
		log.Debug().Msgf("  match co: %s", e.Containers)
		matches++
	}
	if len(e.Codes) > 0 && e.Codes.match(spec.Code.String()) {
		log.Debug().Msgf("  match codes: %q -- %s", spec.Code, e.Codes)
		matches++
	}
	log.Debug().Msgf("  Matches %q (%d)", spec.FQN, matches)

	return matches > 0
}

// Match checks if a given named resource should be Excluded.
func (e Exclude) Match(spec Spec, global bool) bool {
	if spec.isEmpty() || e.isEmpty() {
		return false
	}
	if global {
		return e.matchGlob(spec)
	}

	log.Debug().Msgf("LinterEX -- %s", spec)
	log.Debug().Msgf("  Rule: %s", e)

	if !e.FQNs.match(spec.FQN) {
		log.Debug().Msgf("  fire skip fqn: %q -- %s", spec.FQN, e.FQNs)
		return false
	}

	if !e.Labels.match(spec.Labels) {
		log.Debug().Msgf("  fire skip labels: %s -- %s", spec.Labels, e.Labels)
		return false
	}

	if !e.Annotations.match(spec.Annotations) {
		log.Debug().Msgf("  fire skip anns: %s -- %s", spec.Annotations, e.Annotations)
		return false
	}

	if !e.Containers.matches(spec.Containers) {
		log.Debug().Msgf("  fire skip co: %s", e.Containers)
		return false
	}

	if spec.Code != ZeroCode && !e.Codes.match(spec.Code.String()) {
		log.Debug().Msgf("  fire skip codes: %q -- %s", spec.Code, e.Codes)
		return false
	}

	log.Debug().Msgf("  Matched! %q", spec.FQN)

	return true
}
