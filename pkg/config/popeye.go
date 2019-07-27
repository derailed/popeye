package config

import (
	"github.com/derailed/popeye/internal/issues"
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

	// Popeye tracks Popeye configuration options.
	Popeye struct {
		AllocationLimits `yaml:"allocations"`
		Excludes         `yaml:"excludes"`

		Node  Node            `yaml:"node"`
		Pod   Pod             `yaml:"pod"`
		Codes issues.Glossary `yaml:"codes"`
	}
)

// NewPopeye create a new Popeye configuration.
func NewPopeye() Popeye {
	return Popeye{
		AllocationLimits: AllocationLimits{
			CPU: Allocations{UnderPerc: defaultUnderPerc, OverPerc: defaultOverPerc},
			MEM: Allocations{UnderPerc: defaultUnderPerc, OverPerc: defaultOverPerc},
		},
		Excludes: newExcludes(),
		Node:     newNode(),
		Pod:      newPod(),
	}
}
