package config

const (
	overDefault  = 100
	underDefault = 50
)

type (
	// AllocationLimits tracks limit thresholds cpu and memory thresholds.
	AllocationLimits struct {
		CPU Allocations `yaml:"cpu"`
		MEM Allocations `yaml:"memory"`
	}

	// Allocations track under/over allocation limits.
	Allocations struct {
		Over  int `yaml:"over"`
		Under int `yanl:"under"`
	}

	// Popeye tracks Popeye configuration options.
	Popeye struct {
		Namespace   Namespace        `yaml:"namespace"`
		Node        Node             `yaml:"node"`
		Pod         Pod              `yaml:"pod"`
		Service     Service          `yaml:"service"`
		Allocations AllocationLimits `yaml:"allocations"`
	}
)

// NewPopeye create a new Popeye configuration.
func NewPopeye() Popeye {
	return Popeye{
		Allocations: AllocationLimits{
			CPU: Allocations{Over: overDefault, Under: underDefault},
			MEM: Allocations{Over: overDefault, Under: underDefault},
		},
		Namespace: newNamespace(),
		Node:      newNode(),
		Pod:       newPod(),
		Service:   newService(),
	}
}
