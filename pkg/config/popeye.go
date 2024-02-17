// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Popeye

package config

import (
	"github.com/derailed/popeye/internal/rules"
)

const (
	// DefaultUnderPerc indicates the default percentage for under allocation
	defaultUnderPerc = 200
	// DefaultOverPerc indicates the default percentage for over allocation
	defaultOverPerc = 50
)

type (
	// AllocationLimits tracks limit thresholds cpu and memory thresholds.
	AllocationLimits struct {
		CPU Allocations `yaml:"cpu"`
		MEM Allocations `yaml:"memory"`
	}

	// Allocations track under/over allocation limits.
	Allocations struct {
		UnderPerc int `yaml:"underPercUtilization"`
		OverPerc  int `yanl:"overPercUtilization"`
	}

	Resources struct {
		Node Node `yaml:"node"`
		Pod  Pod  `yaml:"pod"`
	}

	// Popeye tracks Popeye configuration options.
	Popeye struct {
		// AllocationLimits tracks global resource allocations.
		AllocationLimits `yaml:"allocations"`

		// Excludes tracks linter exclusions.
		Exclusions rules.Exclusions `yaml:"excludes"`

		// Resources tracks cpu/mem limits.
		Resources Resources `yaml:"resources"`

		// Codes provides to override codes severity.
		Overrides rules.Overrides `yaml:"overrides"`

		// Registries tracks allowed docker registries.
		Registries []string `yaml:"registries"`
	}
)

// NewPopeye create a new Popeye configuration.
func NewPopeye() Popeye {
	return Popeye{
		AllocationLimits: AllocationLimits{
			CPU: Allocations{
				UnderPerc: defaultUnderPerc,
				OverPerc:  defaultOverPerc,
			},
			MEM: Allocations{
				UnderPerc: defaultUnderPerc,
				OverPerc:  defaultOverPerc,
			},
		},
		Exclusions: rules.NewExclusions(),
		Resources: Resources{
			Node: newNode(),
			Pod:  newPod(),
		},
	}
}

func (p Popeye) Match(spec rules.Spec) bool {
	return p.Exclusions.Match(spec)
}
