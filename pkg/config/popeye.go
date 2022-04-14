package config

import (
	"fmt"
	"strconv"
)

const (
	// DefaultUnderPerc indicates the default percentage for under allocation
	defaultUnderPerc = 200
	// DefaultOverPerc indicates the default percentage for over allocation
	defaultOverPerc = 50
)

type (
	// ID represents a sanitizer code indentifier.
	ID int

	// Glossary represents a collection of codes.
	Glossary map[ID]*Code

	// Code represents a sanitizer code.
	Code struct {
		Message  string `yaml:"message"`
		Severity Level  `yaml:"severity"`
	}

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

		Node       Node     `yaml:"node"`
		Pod        Pod      `yaml:"pod"`
		Codes      Glossary `yaml:"codes"`
		Registries []string `yaml:"registries"`
	}
)

// NewPopeye create a new Popeye configuration.
func NewPopeye() Popeye {
	return Popeye{
		AllocationLimits: AllocationLimits{
			CPU: Allocations{UnderPerc: defaultUnderPerc, OverPerc: defaultOverPerc},
			MEM: Allocations{UnderPerc: defaultUnderPerc, OverPerc: defaultOverPerc},
		},
		Excludes:   newExcludes(),
		Node:       newNode(),
		Pod:        newPod(),
		Registries: []string{},
	}
}

// Format hydrates a message with arguments.
func (c *Code) Format(code ID, args ...interface{}) string {
	msg := "[POP-" + strconv.Itoa(int(code)) + "] "
	if len(args) == 0 {
		return msg + c.Message
	}
	return msg + fmt.Sprintf(c.Message, args...)
}
